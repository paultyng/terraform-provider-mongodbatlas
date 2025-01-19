package advancedcluster

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"hash/crc32"
	"log"
	"slices"
	"strconv"
	"strings"

	admin20240530 "go.mongodb.org/atlas-sdk/v20240530005/admin"
	"go.mongodb.org/atlas-sdk/v20241113004/admin"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/spf13/cast"

	"github.com/mongodb/terraform-provider-mongodbatlas/internal/common/constant"
	"github.com/mongodb/terraform-provider-mongodbatlas/internal/common/conversion"
	"github.com/mongodb/terraform-provider-mongodbatlas/internal/service/advancedclustertpf"
)

const minVersionForChangeStreamOptions = 6.0
const minVersionForDefaultMaxTimeMS = 8.0

type OldShardConfigMeta struct {
	ID       string
	NumShard int
}

var (
	DSTagsSchema = schema.Schema{
		Type:     schema.TypeSet,
		Computed: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"key": {
					Type:     schema.TypeString,
					Computed: true,
				},
				"value": {
					Type:     schema.TypeString,
					Computed: true,
				},
			},
		},
	}
	RSTagsSchema = schema.Schema{
		Type:     schema.TypeSet,
		Optional: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"key": {
					Type:     schema.TypeString,
					Required: true,
				},
				"value": {
					Type:     schema.TypeString,
					Required: true,
				},
			},
		},
	}
)

func SchemaAdvancedConfigDS() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Computed: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"default_read_concern": {
					Type:       schema.TypeString,
					Computed:   true,
					Deprecated: DeprecationMsgOldSchema,
				},
				"default_write_concern": {
					Type:     schema.TypeString,
					Computed: true,
				},
				"fail_index_key_too_long": {
					Type:       schema.TypeBool,
					Computed:   true,
					Deprecated: DeprecationMsgOldSchema,
				},
				"javascript_enabled": {
					Type:     schema.TypeBool,
					Computed: true,
				},
				"minimum_enabled_tls_protocol": {
					Type:     schema.TypeString,
					Computed: true,
				},
				"no_table_scan": {
					Type:     schema.TypeBool,
					Computed: true,
				},
				"oplog_size_mb": {
					Type:     schema.TypeInt,
					Computed: true,
				},
				"sample_size_bi_connector": {
					Type:     schema.TypeInt,
					Computed: true,
				},
				"sample_refresh_interval_bi_connector": {
					Type:     schema.TypeInt,
					Computed: true,
				},
				"oplog_min_retention_hours": {
					Type:     schema.TypeFloat,
					Computed: true,
				},
				"transaction_lifetime_limit_seconds": {
					Type:     schema.TypeInt,
					Computed: true,
				},
				"change_stream_options_pre_and_post_images_expire_after_seconds": {
					Type:     schema.TypeInt,
					Computed: true,
				},
				"default_max_time_ms": {
					Type:     schema.TypeInt,
					Computed: true,
				},
				"tls_cipher_config_mode": {
					Type:     schema.TypeString,
					Computed: true,
				},
				"custom_openssl_cipher_config_tls12": {
					Type:     schema.TypeSet,
					Computed: true,
					Elem: &schema.Schema{
						Type: schema.TypeString,
					},
				},
			},
		},
	}
}

func SchemaConnectionStrings() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Computed: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"standard": {
					Type:     schema.TypeString,
					Computed: true,
				},
				"standard_srv": {
					Type:     schema.TypeString,
					Computed: true,
				},
				"private": {
					Type:     schema.TypeString,
					Computed: true,
				},
				"private_srv": {
					Type:     schema.TypeString,
					Computed: true,
				},
				"private_endpoint": {
					Type:     schema.TypeList,
					Computed: true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"connection_string": {
								Type:     schema.TypeString,
								Computed: true,
							},
							"endpoints": {
								Type:     schema.TypeList,
								Computed: true,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"endpoint_id": {
											Type:     schema.TypeString,
											Computed: true,
										},
										"provider_name": {
											Type:     schema.TypeString,
											Computed: true,
										},
										"region": {
											Type:     schema.TypeString,
											Computed: true,
										},
									},
								},
							},
							"srv_connection_string": {
								Type:     schema.TypeString,
								Computed: true,
							},
							"srv_shard_optimized_connection_string": {
								Type:     schema.TypeString,
								Computed: true,
							},
							"type": {
								Type:     schema.TypeString,
								Computed: true,
							},
						},
					},
				},
			},
		},
	}
}

func SchemaAdvancedConfig() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		Computed: true,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"default_read_concern": {
					Type:       schema.TypeString,
					Optional:   true,
					Computed:   true,
					Deprecated: DeprecationMsgOldSchema,
				},
				"default_write_concern": {
					Type:     schema.TypeString,
					Optional: true,
					Computed: true,
				},
				"fail_index_key_too_long": {
					Type:       schema.TypeBool,
					Optional:   true,
					Computed:   true,
					Deprecated: DeprecationMsgOldSchema,
				},
				"javascript_enabled": {
					Type:     schema.TypeBool,
					Optional: true,
					Computed: true,
				},
				"minimum_enabled_tls_protocol": {
					Type:     schema.TypeString,
					Optional: true,
					Computed: true,
				},
				"no_table_scan": {
					Type:     schema.TypeBool,
					Optional: true,
					Computed: true,
				},
				"oplog_size_mb": {
					Type:     schema.TypeInt,
					Optional: true,
					Computed: true,
				},
				"oplog_min_retention_hours": {
					Type:     schema.TypeFloat,
					Optional: true,
				},
				"sample_size_bi_connector": {
					Type:     schema.TypeInt,
					Optional: true,
					Computed: true,
				},
				"sample_refresh_interval_bi_connector": {
					Type:     schema.TypeInt,
					Optional: true,
					Computed: true,
				},
				"transaction_lifetime_limit_seconds": {
					Type:     schema.TypeInt,
					Optional: true,
					Computed: true,
				},
				"change_stream_options_pre_and_post_images_expire_after_seconds": {
					Type:     schema.TypeInt,
					Optional: true,
					Default:  -1,
				},
				"default_max_time_ms": {
					Type:     schema.TypeInt,
					Optional: true,
				},
				"custom_openssl_cipher_config_tls12": {
					Type:     schema.TypeSet,
					Optional: true,
					Elem: &schema.Schema{
						Type: schema.TypeString,
					},
				},
				"tls_cipher_config_mode": {
					Type:     schema.TypeString,
					Optional: true,
					Computed: true,
				},
			},
		},
	}
}

func HashFunctionForKeyValuePair(v any) int {
	var buf bytes.Buffer
	m := v.(map[string]any)
	buf.WriteString(m["key"].(string))
	buf.WriteString(m["value"].(string))
	return HashCodeString(buf.String())
}

// HashCodeString hashes a string to a unique hashcode.
//
// crc32 returns a uint32, but for our use we need
// and non negative integer. Here we cast to an integer
// and invert it if the result is negative.
func HashCodeString(s string) int {
	v := int(crc32.ChecksumIEEE([]byte(s)))
	if v >= 0 {
		return v
	}
	if -v >= 0 {
		return -v
	}
	// v == MinInt
	return 0
}

func IsSharedTier(instanceSize string) bool {
	return instanceSize == "M0" || instanceSize == "M2" || instanceSize == "M5"
}

// GetDiskSizeGBFromReplicationSpec obtains the diskSizeGB value by looking into the electable spec of the first replication spec.
// Independent storage size scaling is not supported (CLOUDP-201331), meaning all electable/analytics/readOnly configs in all replication specs are the same.
func GetDiskSizeGBFromReplicationSpec(cluster *admin.ClusterDescription20240805) float64 {
	specs := cluster.GetReplicationSpecs()
	if len(specs) < 1 {
		return 0
	}
	configs := specs[0].GetRegionConfigs()
	if len(configs) < 1 {
		return 0
	}
	return configs[0].ElectableSpecs.GetDiskSizeGB()
}

func UpgradeRefreshFunc(ctx context.Context, name, projectID string, client admin.ClustersApi) retry.StateRefreshFunc {
	return func() (any, string, error) {
		cluster, resp, err := client.GetCluster(ctx, projectID, name).Execute()

		if err != nil && strings.Contains(err.Error(), "reset by peer") {
			return nil, "REPEATING", nil
		}

		if err != nil && cluster == nil && resp == nil {
			return nil, "", err
		} else if err != nil {
			if resp.StatusCode == 404 {
				return "", "DELETED", nil
			}
			if resp.StatusCode == 503 {
				return "", "PENDING", nil
			}
			return nil, "", err
		}

		state := cluster.GetStateName()
		return cluster, state, nil
	}
}

func ResourceClusterListAdvancedRefreshFunc(ctx context.Context, projectID string, clustersAPI admin.ClustersApi) retry.StateRefreshFunc {
	return func() (any, string, error) {
		clusters, resp, err := clustersAPI.ListClusters(ctx, projectID).Execute()

		if err != nil && strings.Contains(err.Error(), "reset by peer") {
			return nil, "REPEATING", nil
		}

		if err != nil && clusters == nil && resp == nil {
			return nil, "", err
		}

		if err != nil {
			if resp.StatusCode == 404 {
				return "", "DELETED", nil
			}
			if resp.StatusCode == 503 {
				return "", "PENDING", nil
			}
			return nil, "", err
		}

		for i := range clusters.GetResults() {
			cluster := clusters.GetResults()[i]
			if cluster.GetStateName() != "IDLE" {
				return cluster, "PENDING", nil
			}
		}
		return clusters, "IDLE", nil
	}
}

func FormatMongoDBMajorVersion(val any) string {
	return advancedclustertpf.FormatMongoDBMajorVersion(val.(string))
}

func flattenLabels(l []admin.ComponentLabel) []map[string]string {
	labels := make([]map[string]string, 0, len(l))
	for _, item := range l {
		if item.GetKey() == advancedclustertpf.IgnoreLabelKey {
			continue
		}
		labels = append(labels, map[string]string{
			"key":   item.GetKey(),
			"value": item.GetValue(),
		})
	}
	return labels
}

func flattenTags(tags *[]admin.ResourceTag) []map[string]string {
	tagSlice := *tags
	ret := make([]map[string]string, len(tagSlice))
	for i, tag := range tagSlice {
		ret[i] = map[string]string{
			"key":   tag.GetKey(),
			"value": tag.GetValue(),
		}
	}
	return ret
}

// CheckRegionConfigsPriorityOrder will be deleted in CLOUDP-275825
func CheckRegionConfigsPriorityOrder(regionConfigs []admin.ReplicationSpec20240805) error {
	for _, spec := range regionConfigs {
		configs := spec.GetRegionConfigs()
		for i := range len(configs) - 1 {
			if configs[i].GetPriority() < configs[i+1].GetPriority() {
				return errors.New("priority values in region_configs must be in descending order")
			}
		}
	}
	return nil
}

// CheckRegionConfigsPriorityOrderOld will be deleted in CLOUDP-275825
func CheckRegionConfigsPriorityOrderOld(regionConfigs []admin20240530.ReplicationSpec) error {
	for _, spec := range regionConfigs {
		configs := spec.GetRegionConfigs()
		for i := range len(configs) - 1 {
			if configs[i].GetPriority() < configs[i+1].GetPriority() {
				return errors.New("priority values in region_configs must be in descending order")
			}
		}
	}
	return nil
}

func FlattenPinnedFCV(cluster *admin.ClusterDescription20240805) []map[string]string {
	if cluster.FeatureCompatibilityVersionExpirationDate == nil { // pinned_fcv is defined in state only if featureCompatibilityVersionExpirationDate is present in cluster response
		return nil
	}
	nestedObj := map[string]string{}
	nestedObj["version"] = cluster.GetFeatureCompatibilityVersion()
	nestedObj["expiration_date"] = conversion.TimeToString(cluster.GetFeatureCompatibilityVersionExpirationDate())
	return []map[string]string{nestedObj}
}

func flattenConnectionStrings(str admin.ClusterConnectionStrings) []map[string]any {
	return []map[string]any{
		{
			"standard":         str.GetStandard(),
			"standard_srv":     str.GetStandardSrv(),
			"private":          str.GetPrivate(),
			"private_srv":      str.GetPrivateSrv(),
			"private_endpoint": flattenPrivateEndpoint(str.GetPrivateEndpoint()),
		},
	}
}

func flattenPrivateEndpoint(privateEndpoints []admin.ClusterDescriptionConnectionStringsPrivateEndpoint) []map[string]any {
	endpoints := make([]map[string]any, 0, len(privateEndpoints))
	for _, endpoint := range privateEndpoints {
		endpoints = append(endpoints, map[string]any{
			"connection_string":                     endpoint.GetConnectionString(),
			"srv_connection_string":                 endpoint.GetSrvConnectionString(),
			"srv_shard_optimized_connection_string": endpoint.GetSrvShardOptimizedConnectionString(),
			"type":                                  endpoint.GetType(),
			"endpoints":                             flattenEndpoints(endpoint.GetEndpoints()),
		})
	}
	return endpoints
}

func flattenEndpoints(listEndpoints []admin.ClusterDescriptionConnectionStringsPrivateEndpointEndpoint) []map[string]any {
	endpoints := make([]map[string]any, 0, len(listEndpoints))
	for _, endpoint := range listEndpoints {
		endpoints = append(endpoints, map[string]any{
			"region":        endpoint.GetRegion(),
			"provider_name": endpoint.GetProviderName(),
			"endpoint_id":   endpoint.GetEndpointId(),
		})
	}
	return endpoints
}

func flattenBiConnectorConfig(biConnector *admin.BiConnector) []map[string]any {
	return []map[string]any{
		{
			"enabled":         biConnector.GetEnabled(),
			"read_preference": biConnector.GetReadPreference(),
		},
	}
}

func expandBiConnectorConfig(d *schema.ResourceData) *admin.BiConnector {
	if v, ok := d.GetOk("bi_connector_config"); ok {
		if biConn := v.([]any); len(biConn) > 0 {
			biConnMap := biConn[0].(map[string]any)
			return &admin.BiConnector{
				Enabled:        conversion.Pointer(cast.ToBool(biConnMap["enabled"])),
				ReadPreference: conversion.StringPtr(cast.ToString(biConnMap["read_preference"])),
			}
		}
	}
	return nil
}

func flattenProcessArgs(p20240530 *admin20240530.ClusterDescriptionProcessArgs, p *admin.ClusterDescriptionProcessArgs20240805) []map[string]any {
	if p20240530 == nil {
		return nil
	}
	flattenedProcessArgs := []map[string]any{
		{
			"default_read_concern":                 p20240530.GetDefaultReadConcern(),
			"default_write_concern":                p20240530.GetDefaultWriteConcern(),
			"fail_index_key_too_long":              p20240530.GetFailIndexKeyTooLong(),
			"javascript_enabled":                   p20240530.GetJavascriptEnabled(),
			"minimum_enabled_tls_protocol":         p20240530.GetMinimumEnabledTlsProtocol(),
			"no_table_scan":                        p20240530.GetNoTableScan(),
			"oplog_size_mb":                        p20240530.GetOplogSizeMB(),
			"oplog_min_retention_hours":            p20240530.GetOplogMinRetentionHours(),
			"sample_size_bi_connector":             p20240530.GetSampleSizeBIConnector(),
			"sample_refresh_interval_bi_connector": p20240530.GetSampleRefreshIntervalBIConnector(),
			"transaction_lifetime_limit_seconds":   p20240530.GetTransactionLifetimeLimitSeconds(),
		},
	}
	if p != nil {
		if v := p.ChangeStreamOptionsPreAndPostImagesExpireAfterSeconds; v == nil {
			flattenedProcessArgs[0]["change_stream_options_pre_and_post_images_expire_after_seconds"] = -1 // default in schema, otherwise user gets drift detection
		} else {
			flattenedProcessArgs[0]["change_stream_options_pre_and_post_images_expire_after_seconds"] = p.GetChangeStreamOptionsPreAndPostImagesExpireAfterSeconds()
		}

		if v := p.DefaultMaxTimeMS; v != nil {
			flattenedProcessArgs[0]["default_max_time_ms"] = p.GetDefaultMaxTimeMS()
		}
		flattenedProcessArgs[0]["tls_cipher_config_mode"] = p.GetTlsCipherConfigMode()
		flattenedProcessArgs[0]["custom_openssl_cipher_config_tls12"] = p.GetCustomOpensslCipherConfigTls12()
	}

	return flattenedProcessArgs
}

func FlattenAdvancedReplicationSpecsOldShardingConfig(ctx context.Context, apiObjects []admin.ReplicationSpec20240805, zoneNameToOldReplicationSpecMeta map[string]OldShardConfigMeta, tfMapObjects []any,
	d *schema.ResourceData, connV2 *admin.APIClient) ([]map[string]any, error) {
	replicationSpecFlattener := func(ctx context.Context, sdkModel *admin.ReplicationSpec20240805, tfModel map[string]any, resourceData *schema.ResourceData, client *admin.APIClient) (map[string]any, error) {
		return flattenAdvancedReplicationSpecOldShardingConfig(ctx, sdkModel, zoneNameToOldReplicationSpecMeta, tfModel, resourceData, connV2)
	}
	compressedAPIObjects := compressAPIObjectList(apiObjects)
	return flattenAdvancedReplicationSpecsLogic(ctx, compressedAPIObjects, tfMapObjects, d,
		doesAdvancedReplicationSpecMatchAPIOldShardConfig, replicationSpecFlattener, connV2)
}

func flattenAdvancedReplicationSpecs(ctx context.Context, apiObjects []admin.ReplicationSpec20240805, zoneNameToOldReplicationSpecMeta map[string]OldShardConfigMeta, tfMapObjects []any,
	d *schema.ResourceData, connV2 *admin.APIClient) ([]map[string]any, error) {
	// for flattening new model we need information of replication spec ids associated to old API to avoid breaking changes for users referencing replication_specs.*.id
	replicationSpecFlattener := func(ctx context.Context, sdkModel *admin.ReplicationSpec20240805, tfModel map[string]any, resourceData *schema.ResourceData, client *admin.APIClient) (map[string]any, error) {
		return flattenAdvancedReplicationSpec(ctx, sdkModel, zoneNameToOldReplicationSpecMeta, tfModel, resourceData, connV2)
	}
	return flattenAdvancedReplicationSpecsLogic(ctx, apiObjects, tfMapObjects, d,
		doesAdvancedReplicationSpecMatchAPI, replicationSpecFlattener, connV2)
}

// compressAPIObjectList returns an array of ReplicationSpec20240805. The input array is reduced from all shards to only one shard per zoneName
func compressAPIObjectList(apiObjects []admin.ReplicationSpec20240805) []admin.ReplicationSpec20240805 {
	var compressedAPIObjectList []admin.ReplicationSpec20240805
	wasZoneNameUsed := populateZoneNameMap(apiObjects)
	for _, apiObject := range apiObjects {
		if !wasZoneNameUsed[apiObject.GetZoneName()] {
			compressedAPIObjectList = append(compressedAPIObjectList, apiObject)
			wasZoneNameUsed[apiObject.GetZoneName()] = true
		}
	}
	return compressedAPIObjectList
}

// populateZoneNameMap returns a map of zoneNames and initializes all keys to false.
func populateZoneNameMap(apiObjects []admin.ReplicationSpec20240805) map[string]bool {
	zoneNames := make(map[string]bool)
	for _, apiObject := range apiObjects {
		if _, exists := zoneNames[apiObject.GetZoneName()]; !exists {
			zoneNames[apiObject.GetZoneName()] = false
		}
	}
	return zoneNames
}

type ReplicationSpecSDKModel interface {
	admin20240530.ReplicationSpec | admin.ReplicationSpec20240805
}

func flattenAdvancedReplicationSpecsLogic[T ReplicationSpecSDKModel](
	ctx context.Context, apiObjects []T, tfMapObjects []any, d *schema.ResourceData,
	tfModelWithSDKMatcher func(map[string]any, *T) bool,
	flattenRepSpec func(context.Context, *T, map[string]any, *schema.ResourceData, *admin.APIClient) (map[string]any, error),
	connV2 *admin.APIClient) ([]map[string]any, error) {
	if len(apiObjects) == 0 {
		return nil, nil
	}

	tfList := make([]map[string]any, len(apiObjects))
	wasAPIObjectUsed := make([]bool, len(apiObjects))

	for i := range len(tfList) {
		var tfMapObject map[string]any

		if len(tfMapObjects) > i {
			tfMapObject = tfMapObjects[i].(map[string]any)
		}
		for j := 0; j < len(apiObjects); j++ {
			if wasAPIObjectUsed[j] || !tfModelWithSDKMatcher(tfMapObject, &apiObjects[j]) {
				continue
			}

			advancedReplicationSpec, err := flattenRepSpec(ctx, &apiObjects[j], tfMapObject, d, connV2)

			if err != nil {
				return nil, err
			}

			tfList[i] = advancedReplicationSpec
			wasAPIObjectUsed[j] = true
			break
		}
	}

	for i, tfo := range tfList {
		var tfMapObject map[string]any

		if tfo != nil {
			continue
		}

		if len(tfMapObjects) > i {
			tfMapObject = tfMapObjects[i].(map[string]any)
		}

		j := slices.IndexFunc(wasAPIObjectUsed, func(isUsed bool) bool { return !isUsed })
		advancedReplicationSpec, err := flattenRepSpec(ctx, &apiObjects[j], tfMapObject, d, connV2)

		if err != nil {
			return nil, err
		}

		tfList[i] = advancedReplicationSpec
		wasAPIObjectUsed[j] = true
	}

	return tfList, nil
}

func doesAdvancedReplicationSpecMatchAPIOldShardConfig(tfObject map[string]any, apiObject *admin.ReplicationSpec20240805) bool {
	return tfObject["zone_name"] == apiObject.GetZoneName()
}

func doesAdvancedReplicationSpecMatchAPI(tfObject map[string]any, apiObject *admin.ReplicationSpec20240805) bool {
	return tfObject["external_id"] == apiObject.GetId()
}

func flattenAdvancedReplicationSpecRegionConfigs(ctx context.Context, apiObjects []admin.CloudRegionConfig20240805, tfMapObjects []any,
	d *schema.ResourceData, connV2 *admin.APIClient) (tfResult []map[string]any, containersIDs map[string]string, err error) {
	if len(apiObjects) == 0 {
		return nil, nil, nil
	}

	var tfList []map[string]any
	containerIDs := make(map[string]string)

	for i := range apiObjects {
		apiObject := apiObjects[i]
		if len(tfMapObjects) > i {
			tfMapObject := tfMapObjects[i].(map[string]any)
			tfList = append(tfList, flattenAdvancedReplicationSpecRegionConfig(&apiObject, tfMapObject))
		} else {
			tfList = append(tfList, flattenAdvancedReplicationSpecRegionConfig(&apiObject, nil))
		}

		if apiObject.GetProviderName() != "TENANT" {
			params := &admin.ListPeeringContainerByCloudProviderApiParams{
				GroupId:      d.Get("project_id").(string),
				ProviderName: apiObject.ProviderName,
			}
			containers, _, err := connV2.NetworkPeeringApi.ListPeeringContainerByCloudProviderWithParams(ctx, params).Execute()
			if err != nil {
				return nil, nil, err
			}
			if result := advancedclustertpf.GetAdvancedClusterContainerID(containers.GetResults(), &apiObject); result != "" {
				// Will print as "providerName:regionName" = "containerId" in terraform show
				containerIDs[fmt.Sprintf("%s:%s", apiObject.GetProviderName(), apiObject.GetRegionName())] = result
			}
		}
	}
	return tfList, containerIDs, nil
}

func flattenAdvancedReplicationSpecRegionConfig(apiObject *admin.CloudRegionConfig20240805, tfMapObject map[string]any) map[string]any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}
	if tfMapObject != nil {
		if v, ok := tfMapObject["analytics_specs"]; ok && len(v.([]any)) > 0 {
			tfMap["analytics_specs"] = flattenAdvancedReplicationSpecRegionConfigSpec(apiObject.AnalyticsSpecs, apiObject.GetProviderName(), tfMapObject["analytics_specs"].([]any))
		}
		if v, ok := tfMapObject["electable_specs"]; ok && len(v.([]any)) > 0 {
			tfMap["electable_specs"] = flattenAdvancedReplicationSpecRegionConfigSpec(hwSpecToDedicatedHwSpec(apiObject.ElectableSpecs), apiObject.GetProviderName(), tfMapObject["electable_specs"].([]any))
		}
		if v, ok := tfMapObject["read_only_specs"]; ok && len(v.([]any)) > 0 {
			tfMap["read_only_specs"] = flattenAdvancedReplicationSpecRegionConfigSpec(apiObject.ReadOnlySpecs, apiObject.GetProviderName(), tfMapObject["read_only_specs"].([]any))
		}
		if v, ok := tfMapObject["auto_scaling"]; ok && len(v.([]any)) > 0 {
			tfMap["auto_scaling"] = flattenAdvancedReplicationSpecAutoScaling(apiObject.AutoScaling)
		}
		if v, ok := tfMapObject["analytics_auto_scaling"]; ok && len(v.([]any)) > 0 {
			tfMap["analytics_auto_scaling"] = flattenAdvancedReplicationSpecAutoScaling(apiObject.AnalyticsAutoScaling)
		}
	} else {
		tfMap["analytics_specs"] = flattenAdvancedReplicationSpecRegionConfigSpec(apiObject.AnalyticsSpecs, apiObject.GetProviderName(), nil)
		tfMap["electable_specs"] = flattenAdvancedReplicationSpecRegionConfigSpec(hwSpecToDedicatedHwSpec(apiObject.ElectableSpecs), apiObject.GetProviderName(), nil)
		tfMap["read_only_specs"] = flattenAdvancedReplicationSpecRegionConfigSpec(apiObject.ReadOnlySpecs, apiObject.GetProviderName(), nil)
		tfMap["auto_scaling"] = flattenAdvancedReplicationSpecAutoScaling(apiObject.AutoScaling)
		tfMap["analytics_auto_scaling"] = flattenAdvancedReplicationSpecAutoScaling(apiObject.AnalyticsAutoScaling)
	}

	tfMap["region_name"] = apiObject.GetRegionName()
	tfMap["provider_name"] = apiObject.GetProviderName()
	tfMap["backing_provider_name"] = apiObject.GetBackingProviderName()
	tfMap["priority"] = apiObject.GetPriority()

	return tfMap
}

func hwSpecToDedicatedHwSpec(apiObject *admin.HardwareSpec20240805) *admin.DedicatedHardwareSpec20240805 {
	if apiObject == nil {
		return nil
	}
	return &admin.DedicatedHardwareSpec20240805{
		NodeCount:     apiObject.NodeCount,
		DiskIOPS:      apiObject.DiskIOPS,
		EbsVolumeType: apiObject.EbsVolumeType,
		InstanceSize:  apiObject.InstanceSize,
		DiskSizeGB:    apiObject.DiskSizeGB,
	}
}

func dedicatedHwSpecToHwSpec(apiObject *admin.DedicatedHardwareSpec20240805) *admin.HardwareSpec20240805 {
	if apiObject == nil {
		return nil
	}
	return &admin.HardwareSpec20240805{
		DiskSizeGB:    apiObject.DiskSizeGB,
		NodeCount:     apiObject.NodeCount,
		DiskIOPS:      apiObject.DiskIOPS,
		EbsVolumeType: apiObject.EbsVolumeType,
		InstanceSize:  apiObject.InstanceSize,
	}
}

func flattenAdvancedReplicationSpecRegionConfigSpec(apiObject *admin.DedicatedHardwareSpec20240805, providerName string, tfMapObjects []any) []map[string]any {
	if apiObject == nil {
		return nil
	}
	var tfList []map[string]any

	tfMap := map[string]any{}

	if len(tfMapObjects) > 0 {
		tfMapObject := tfMapObjects[0].(map[string]any)

		if providerName == constant.AWS || providerName == constant.AZURE {
			if cast.ToInt64(apiObject.GetDiskIOPS()) > 0 {
				tfMap["disk_iops"] = apiObject.GetDiskIOPS()
			}
		}
		if providerName == constant.AWS {
			if v, ok := tfMapObject["ebs_volume_type"]; ok && v.(string) != "" {
				tfMap["ebs_volume_type"] = apiObject.GetEbsVolumeType()
			}
		}
		if _, ok := tfMapObject["disk_size_gb"]; ok {
			tfMap["disk_size_gb"] = apiObject.GetDiskSizeGB()
		}
		if _, ok := tfMapObject["node_count"]; ok {
			tfMap["node_count"] = apiObject.GetNodeCount()
		}
		if v, ok := tfMapObject["instance_size"]; ok && v.(string) != "" {
			tfMap["instance_size"] = apiObject.GetInstanceSize()
			tfList = append(tfList, tfMap)
		}
	} else {
		tfMap["disk_size_gb"] = apiObject.GetDiskSizeGB()
		tfMap["disk_iops"] = apiObject.GetDiskIOPS()
		tfMap["ebs_volume_type"] = apiObject.GetEbsVolumeType()
		tfMap["node_count"] = apiObject.GetNodeCount()
		tfMap["instance_size"] = apiObject.GetInstanceSize()
		tfList = append(tfList, tfMap)
	}
	return tfList
}

func flattenAdvancedReplicationSpecAutoScaling(apiObject *admin.AdvancedAutoScalingSettings) []map[string]any {
	if apiObject == nil {
		return nil
	}
	var tfList []map[string]any
	tfMap := map[string]any{}
	if apiObject.DiskGB != nil {
		tfMap["disk_gb_enabled"] = apiObject.DiskGB.GetEnabled()
	}
	if apiObject.Compute != nil {
		tfMap["compute_enabled"] = apiObject.Compute.GetEnabled()
		tfMap["compute_scale_down_enabled"] = apiObject.Compute.GetScaleDownEnabled()
		tfMap["compute_min_instance_size"] = apiObject.Compute.GetMinInstanceSize()
		tfMap["compute_max_instance_size"] = apiObject.Compute.GetMaxInstanceSize()
	}
	tfList = append(tfList, tfMap)
	return tfList
}

func expandProcessArgs(d *schema.ResourceData, p map[string]any, mongodbMajorVersion *string) (admin20240530.ClusterDescriptionProcessArgs, admin.ClusterDescriptionProcessArgs20240805) {
	res20240530 := admin20240530.ClusterDescriptionProcessArgs{}
	res := admin.ClusterDescriptionProcessArgs20240805{}

	if _, ok := d.GetOkExists("advanced_configuration.0.default_read_concern"); ok {
		res20240530.DefaultReadConcern = conversion.StringPtr(cast.ToString(p["default_read_concern"]))
	}

	if _, ok := d.GetOkExists("advanced_configuration.0.default_write_concern"); ok {
		res20240530.DefaultWriteConcern = conversion.StringPtr(cast.ToString(p["default_write_concern"]))
	}

	if _, ok := d.GetOkExists("advanced_configuration.0.fail_index_key_too_long"); ok {
		res20240530.FailIndexKeyTooLong = conversion.Pointer(cast.ToBool(p["fail_index_key_too_long"]))
	}

	if _, ok := d.GetOkExists("advanced_configuration.0.javascript_enabled"); ok {
		res20240530.JavascriptEnabled = conversion.Pointer(cast.ToBool(p["javascript_enabled"]))
	}

	if _, ok := d.GetOkExists("advanced_configuration.0.minimum_enabled_tls_protocol"); ok {
		res20240530.MinimumEnabledTlsProtocol = conversion.StringPtr(cast.ToString(p["minimum_enabled_tls_protocol"]))
	}

	if _, ok := d.GetOkExists("advanced_configuration.0.no_table_scan"); ok {
		res20240530.NoTableScan = conversion.Pointer(cast.ToBool(p["no_table_scan"]))
	}

	if _, ok := d.GetOkExists("advanced_configuration.0.sample_size_bi_connector"); ok {
		res20240530.SampleSizeBIConnector = conversion.Pointer(cast.ToInt(p["sample_size_bi_connector"]))
	}

	if _, ok := d.GetOkExists("advanced_configuration.0.sample_refresh_interval_bi_connector"); ok {
		res20240530.SampleRefreshIntervalBIConnector = conversion.Pointer(cast.ToInt(p["sample_refresh_interval_bi_connector"]))
	}

	if _, ok := d.GetOkExists("advanced_configuration.0.oplog_size_mb"); ok {
		if sizeMB := cast.ToInt64(p["oplog_size_mb"]); sizeMB != 0 {
			res20240530.OplogSizeMB = conversion.Pointer(cast.ToInt(p["oplog_size_mb"]))
		} else {
			log.Printf(ErrorClusterSetting, `oplog_size_mb`, "", cast.ToString(sizeMB))
		}
	}

	if _, ok := d.GetOkExists("advanced_configuration.0.oplog_min_retention_hours"); ok {
		if minRetentionHours := cast.ToFloat64(p["oplog_min_retention_hours"]); minRetentionHours >= 0 {
			res20240530.OplogMinRetentionHours = conversion.Pointer(cast.ToFloat64(p["oplog_min_retention_hours"]))
		} else {
			log.Printf(ErrorClusterSetting, `oplog_min_retention_hours`, "", cast.ToString(minRetentionHours))
		}
	}

	if _, ok := d.GetOkExists("advanced_configuration.0.transaction_lifetime_limit_seconds"); ok {
		if transactionLifetimeLimitSeconds := cast.ToInt64(p["transaction_lifetime_limit_seconds"]); transactionLifetimeLimitSeconds > 0 {
			res20240530.TransactionLifetimeLimitSeconds = conversion.Pointer(cast.ToInt64(p["transaction_lifetime_limit_seconds"]))
		} else {
			log.Printf(ErrorClusterSetting, `transaction_lifetime_limit_seconds`, "", cast.ToString(transactionLifetimeLimitSeconds))
		}
	}

	if _, ok := d.GetOkExists("advanced_configuration.0.change_stream_options_pre_and_post_images_expire_after_seconds"); ok && IsChangeStreamOptionsMinRequiredMajorVersion(mongodbMajorVersion) {
		tmp := p["change_stream_options_pre_and_post_images_expire_after_seconds"]
		tmpInt := cast.ToInt(tmp)

		res.ChangeStreamOptionsPreAndPostImagesExpireAfterSeconds = conversion.IntPtr(tmpInt)
	}

	if _, ok := d.GetOkExists("advanced_configuration.0.default_max_time_ms"); ok {
		if IsDefaultMaxTimeMinRequiredMajorVersion(mongodbMajorVersion) {
			res.DefaultMaxTimeMS = conversion.Pointer(cast.ToInt(p["default_max_time_ms"]))
		} else {
			log.Print(ErrorDefaultMaxTimeMinVersion)
		}
	}

	if _, ok := d.GetOkExists("advanced_configuration.0.tls_cipher_config_mode"); ok {
		res.TlsCipherConfigMode = conversion.StringPtr(cast.ToString(p["tls_cipher_config_mode"]))
	}

	if _, ok := d.GetOkExists("advanced_configuration.0.custom_openssl_cipher_config_tls12"); ok {
		tmp := conversion.ExpandStringListFromSetSchema(d.Get("advanced_configuration.0.custom_openssl_cipher_config_tls12").(*schema.Set))
		res.CustomOpensslCipherConfigTls12 = &tmp
	}
	return res20240530, res
}

func isMinRequiredMajorVersion(input *string, minVersion float64) bool {
	if input == nil || *input == "" {
		return true
	}
	parts := strings.SplitN(*input, ".", 2)
	if len(parts) == 0 {
		return false
	}

	value, err := strconv.ParseFloat(parts[0], 64)
	if err != nil {
		return false
	}

	return value >= minVersion
}

func IsChangeStreamOptionsMinRequiredMajorVersion(input *string) bool {
	return isMinRequiredMajorVersion(input, minVersionForChangeStreamOptions)
}

func IsDefaultMaxTimeMinRequiredMajorVersion(input *string) bool {
	return isMinRequiredMajorVersion(input, minVersionForDefaultMaxTimeMS)
}

func expandLabelSliceFromSetSchema(d *schema.ResourceData) ([]admin.ComponentLabel, diag.Diagnostics) {
	list := d.Get("labels").(*schema.Set)
	res := make([]admin.ComponentLabel, list.Len())
	for i, val := range list.List() {
		v := val.(map[string]any)
		key := v["key"].(string)
		if key == advancedclustertpf.IgnoreLabelKey {
			return nil, diag.FromErr(advancedclustertpf.ErrIgnoreLabel)
		}
		res[i] = admin.ComponentLabel{
			Key:   conversion.StringPtr(key),
			Value: conversion.StringPtr(v["value"].(string)),
		}
	}
	return res, nil
}

func expandAdvancedReplicationSpecs(tfList []any, rootDiskSizeGB *float64) *[]admin.ReplicationSpec20240805 {
	var apiObjects []admin.ReplicationSpec20240805
	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]any)
		if !ok || tfMap == nil {
			continue
		}
		apiObject := expandAdvancedReplicationSpec(tfMap, rootDiskSizeGB)
		apiObjects = append(apiObjects, *apiObject)

		// handles adding additional replication spec objects if legacy num_shards attribute is being used and greater than 1
		numShards := tfMap["num_shards"].(int)
		for range numShards - 1 {
			apiObjects = append(apiObjects, *apiObject)
		}
	}
	if apiObjects == nil {
		return nil
	}
	return &apiObjects
}

func expandAdvancedReplicationSpecsOldSDK(tfList []any) *[]admin20240530.ReplicationSpec {
	var apiObjects []admin20240530.ReplicationSpec
	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]any)
		if !ok || tfMap == nil {
			continue
		}
		apiObject := expandAdvancedReplicationSpecOldSDK(tfMap)
		apiObjects = append(apiObjects, *apiObject)
	}
	if apiObjects == nil {
		return nil
	}
	return &apiObjects
}

func expandAdvancedReplicationSpec(tfMap map[string]any, rootDiskSizeGB *float64) *admin.ReplicationSpec20240805 {
	apiObject := &admin.ReplicationSpec20240805{
		ZoneName:      conversion.StringPtr(tfMap["zone_name"].(string)),
		RegionConfigs: expandRegionConfigs(tfMap["region_configs"].([]any), rootDiskSizeGB),
	}
	if tfMap["external_id"].(string) != "" {
		apiObject.Id = conversion.StringPtr(tfMap["external_id"].(string))
	}
	return apiObject
}

func expandAdvancedReplicationSpecOldSDK(tfMap map[string]any) *admin20240530.ReplicationSpec {
	apiObject := &admin20240530.ReplicationSpec{
		NumShards:     conversion.Pointer(tfMap["num_shards"].(int)),
		ZoneName:      conversion.StringPtr(tfMap["zone_name"].(string)),
		RegionConfigs: convertRegionConfigSliceToOldSDK(expandRegionConfigs(tfMap["region_configs"].([]any), nil)),
	}
	if tfMap["id"].(string) != "" {
		apiObject.Id = conversion.StringPtr(tfMap["id"].(string))
	}
	return apiObject
}

func expandRegionConfigs(tfList []any, rootDiskSizeGB *float64) *[]admin.CloudRegionConfig20240805 {
	var apiObjects []admin.CloudRegionConfig20240805
	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]any)
		if !ok || tfMap == nil {
			continue
		}
		apiObject := expandRegionConfig(tfMap, rootDiskSizeGB)
		apiObjects = append(apiObjects, *apiObject)
	}
	if apiObjects == nil {
		return nil
	}
	return &apiObjects
}

func expandRegionConfig(tfMap map[string]any, rootDiskSizeGB *float64) *admin.CloudRegionConfig20240805 {
	providerName := tfMap["provider_name"].(string)
	apiObject := &admin.CloudRegionConfig20240805{
		Priority:     conversion.Pointer(cast.ToInt(tfMap["priority"])),
		ProviderName: conversion.StringPtr(providerName),
		RegionName:   conversion.StringPtr(tfMap["region_name"].(string)),
	}

	if v, ok := tfMap["analytics_specs"]; ok && len(v.([]any)) > 0 {
		apiObject.AnalyticsSpecs = expandRegionConfigSpec(v.([]any), providerName, rootDiskSizeGB)
	}
	if v, ok := tfMap["electable_specs"]; ok && len(v.([]any)) > 0 {
		apiObject.ElectableSpecs = dedicatedHwSpecToHwSpec(expandRegionConfigSpec(v.([]any), providerName, rootDiskSizeGB))
	}
	if v, ok := tfMap["read_only_specs"]; ok && len(v.([]any)) > 0 {
		apiObject.ReadOnlySpecs = expandRegionConfigSpec(v.([]any), providerName, rootDiskSizeGB)
	}
	if v, ok := tfMap["auto_scaling"]; ok && len(v.([]any)) > 0 {
		apiObject.AutoScaling = expandRegionConfigAutoScaling(v.([]any))
	}
	if v, ok := tfMap["analytics_auto_scaling"]; ok && len(v.([]any)) > 0 {
		apiObject.AnalyticsAutoScaling = expandRegionConfigAutoScaling(v.([]any))
	}
	if v, ok := tfMap["backing_provider_name"]; ok {
		apiObject.BackingProviderName = conversion.StringPtr(v.(string))
	}
	return apiObject
}

func expandRegionConfigSpec(tfList []any, providerName string, rootDiskSizeGB *float64) *admin.DedicatedHardwareSpec20240805 {
	tfMap, _ := tfList[0].(map[string]any)
	apiObject := new(admin.DedicatedHardwareSpec20240805)
	if providerName == constant.AWS || providerName == constant.AZURE {
		if v, ok := tfMap["disk_iops"]; ok && v.(int) > 0 {
			apiObject.DiskIOPS = conversion.Pointer(v.(int))
		}
	}
	if providerName == constant.AWS {
		if v, ok := tfMap["ebs_volume_type"]; ok {
			apiObject.EbsVolumeType = conversion.StringPtr(v.(string))
		}
	}
	if v, ok := tfMap["instance_size"]; ok {
		apiObject.InstanceSize = conversion.StringPtr(v.(string))
	}
	if v, ok := tfMap["node_count"]; ok {
		apiObject.NodeCount = conversion.Pointer(v.(int))
	}

	if v, ok := tfMap["disk_size_gb"]; ok && v.(float64) != 0 {
		apiObject.DiskSizeGB = conversion.Pointer(v.(float64))
	}

	// value defined in root is set if it is defined in the create, or value has changed in the update.
	if rootDiskSizeGB != nil {
		apiObject.DiskSizeGB = rootDiskSizeGB
	}

	return apiObject
}

func expandRegionConfigAutoScaling(tfList []any) *admin.AdvancedAutoScalingSettings {
	tfMap, _ := tfList[0].(map[string]any)
	settings := admin.AdvancedAutoScalingSettings{
		DiskGB:  new(admin.DiskGBAutoScaling),
		Compute: new(admin.AdvancedComputeAutoScaling),
	}

	if v, ok := tfMap["disk_gb_enabled"]; ok {
		settings.DiskGB.Enabled = conversion.Pointer(v.(bool))
	}
	if v, ok := tfMap["compute_enabled"]; ok {
		settings.Compute.Enabled = conversion.Pointer(v.(bool))
	}
	if v, ok := tfMap["compute_scale_down_enabled"]; ok {
		settings.Compute.ScaleDownEnabled = conversion.Pointer(v.(bool))
	}
	if v, ok := tfMap["compute_min_instance_size"]; ok {
		value := settings.Compute.ScaleDownEnabled
		if *value {
			settings.Compute.MinInstanceSize = conversion.StringPtr(v.(string))
		}
	}
	if v, ok := tfMap["compute_max_instance_size"]; ok {
		value := settings.Compute.Enabled
		if *value {
			settings.Compute.MaxInstanceSize = conversion.StringPtr(v.(string))
		}
	}
	return &settings
}

func flattenAdvancedReplicationSpecsDS(ctx context.Context, apiRepSpecs []admin.ReplicationSpec20240805, zoneNameToOldReplicationSpecMeta map[string]OldShardConfigMeta, d *schema.ResourceData, connV2 *admin.APIClient) ([]map[string]any, error) {
	if len(apiRepSpecs) == 0 {
		return nil, nil
	}

	tfList := make([]map[string]any, len(apiRepSpecs))

	for i, apiRepSpec := range apiRepSpecs {
		tfReplicationSpec, err := flattenAdvancedReplicationSpec(ctx, &apiRepSpec, zoneNameToOldReplicationSpecMeta, nil, d, connV2)
		if err != nil {
			return nil, err
		}
		tfList[i] = tfReplicationSpec
	}
	return tfList, nil
}

func flattenAdvancedReplicationSpec(ctx context.Context, apiObject *admin.ReplicationSpec20240805, zoneNameToOldReplicationSpecMeta map[string]OldShardConfigMeta, tfMapObject map[string]any,
	d *schema.ResourceData, connV2 *admin.APIClient) (map[string]any, error) {
	if apiObject == nil {
		return nil, nil
	}

	tfMap := map[string]any{}
	tfMap["external_id"] = apiObject.GetId()

	if oldShardConfig, ok := zoneNameToOldReplicationSpecMeta[apiObject.GetZoneName()]; ok {
		tfMap["id"] = oldShardConfig.ID // replicationSpecs.*.id stores value associated to old cluster API (2023-02-01)
	}

	// define num_shards for backwards compatibility as this attribute has default value of 1.
	tfMap["num_shards"] = 1

	if tfMapObject != nil {
		object, containerIDs, err := flattenAdvancedReplicationSpecRegionConfigs(ctx, apiObject.GetRegionConfigs(), tfMapObject["region_configs"].([]any), d, connV2)
		if err != nil {
			return nil, err
		}
		tfMap["region_configs"] = object
		tfMap["container_id"] = containerIDs
	} else {
		object, containerIDs, err := flattenAdvancedReplicationSpecRegionConfigs(ctx, apiObject.GetRegionConfigs(), nil, d, connV2)
		if err != nil {
			return nil, err
		}
		tfMap["region_configs"] = object
		tfMap["container_id"] = containerIDs
	}
	tfMap["zone_name"] = apiObject.GetZoneName()
	tfMap["zone_id"] = apiObject.GetZoneId()

	return tfMap, nil
}

func flattenAdvancedReplicationSpecOldShardingConfig(ctx context.Context, apiObject *admin.ReplicationSpec20240805, zoneNameToOldShardConfigMeta map[string]OldShardConfigMeta, tfMapObject map[string]any,
	d *schema.ResourceData, connV2 *admin.APIClient) (map[string]any, error) {
	if apiObject == nil {
		return nil, nil
	}

	tfMap := map[string]any{}
	if oldShardConfigData, ok := zoneNameToOldShardConfigMeta[apiObject.GetZoneName()]; ok {
		tfMap["num_shards"] = oldShardConfigData.NumShard
		tfMap["id"] = oldShardConfigData.ID
	}
	if tfMapObject != nil {
		object, containerIDs, err := flattenAdvancedReplicationSpecRegionConfigs(ctx, apiObject.GetRegionConfigs(), tfMapObject["region_configs"].([]any), d, connV2)
		if err != nil {
			return nil, err
		}
		tfMap["region_configs"] = object
		tfMap["container_id"] = containerIDs
	} else {
		object, containerIDs, err := flattenAdvancedReplicationSpecRegionConfigs(ctx, apiObject.GetRegionConfigs(), nil, d, connV2)
		if err != nil {
			return nil, err
		}
		tfMap["region_configs"] = object
		tfMap["container_id"] = containerIDs
	}
	tfMap["zone_name"] = apiObject.GetZoneName()
	tfMap["zone_id"] = apiObject.GetZoneId()

	return tfMap, nil
}
