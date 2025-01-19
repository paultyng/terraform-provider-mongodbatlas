package advancedcluster

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"reflect"
	"regexp"
	"strings"
	"time"

	admin20240530 "go.mongodb.org/atlas-sdk/v20240530005/admin"
	admin20240805 "go.mongodb.org/atlas-sdk/v20240805005/admin"
	"go.mongodb.org/atlas-sdk/v20241113004/admin"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/spf13/cast"

	"github.com/mongodb/terraform-provider-mongodbatlas/internal/common/constant"
	"github.com/mongodb/terraform-provider-mongodbatlas/internal/common/conversion"
	"github.com/mongodb/terraform-provider-mongodbatlas/internal/common/validate"
	"github.com/mongodb/terraform-provider-mongodbatlas/internal/config"
	"github.com/mongodb/terraform-provider-mongodbatlas/internal/service/advancedclustertpf"
)

const (
	errorCreate                    = "error creating advanced cluster: %s"
	errorRead                      = "error reading  advanced cluster (%s): %s"
	errorDelete                    = "error deleting advanced cluster (%s): %s"
	errorUpdate                    = "error updating advanced cluster (%s): %s"
	errorConfigUpdate              = "error updating advanced cluster configuration options (%s): %s"
	errorConfigRead                = "error reading advanced cluster configuration options (%s): %s"
	ErrorClusterSetting            = "error setting `%s` for MongoDB Cluster (%s): %s"
	ErrorAdvancedConfRead          = "error reading Advanced Configuration Option %s for MongoDB Cluster (%s): %s"
	ErrorClusterAdvancedSetting    = "error setting `%s` for MongoDB ClusterAdvanced (%s): %s"
	ErrorAdvancedClusterListStatus = "error awaiting MongoDB ClusterAdvanced List IDLE: %s"
	ErrorOperationNotPermitted     = "error operation not permitted"
	ErrorDefaultMaxTimeMinVersion  = "default_max_time_ms can not be set for mongo_db_major_version lower than 8.0"
	DeprecationOldSchemaAction     = "Please refer to our examples, documentation, and 1.18.0 migration guide for more details at https://registry.terraform.io/providers/mongodb/mongodbatlas/latest/docs/guides/1.18.0-upgrade-guide.html.markdown"
	V20240530                      = "(v20240530)"
)

var DeprecationMsgOldSchema = fmt.Sprintf("%s %s", constant.DeprecationParam, DeprecationOldSchemaAction)

func Resource() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceCreate,
		ReadWithoutTimeout:   resourceRead,
		UpdateWithoutTimeout: resourceUpdateOrUpgrade,
		DeleteWithoutTimeout: resourceDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceImport,
		},
		SchemaVersion: 1,
		StateUpgraders: []schema.StateUpgrader{
			{
				Type:    ResourceV0().CoreConfigSchema().ImpliedType(),
				Upgrade: resourceStateUpgradeV0,
				Version: 0,
			},
		},
		Schema: map[string]*schema.Schema{
			"project_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"cluster_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"backup_enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
			},
			"retain_backups_enabled": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Flag that indicates whether to retain backup snapshots for the deleted dedicated cluster",
			},
			"bi_connector_config": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"enabled": {
							Type:     schema.TypeBool,
							Optional: true,
							Computed: true,
						},
						"read_preference": {
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
						},
					},
				},
			},
			"cluster_type": {
				Type:     schema.TypeString,
				Required: true,
			},
			"connection_strings": SchemaConnectionStrings(),
			"create_date": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"disk_size_gb": {
				Type:       schema.TypeFloat,
				Optional:   true,
				Computed:   true,
				Deprecated: DeprecationMsgOldSchema,
			},
			"encryption_at_rest_provider": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"labels": {
				Type:     schema.TypeSet,
				Optional: true,
				Set:      HashFunctionForKeyValuePair,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"key": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"value": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
			"tags": &RSTagsSchema,
			"mongo_db_major_version": {
				Type:      schema.TypeString,
				Optional:  true,
				Computed:  true,
				StateFunc: FormatMongoDBMajorVersion,
			},
			"mongo_db_version": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"paused": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
			},
			"pit_enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
			},
			"replication_specs": {
				Type:     schema.TypeList,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:       schema.TypeString,
							Computed:   true,
							Deprecated: DeprecationMsgOldSchema,
						},
						"zone_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"external_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"num_shards": {
							Type:         schema.TypeInt,
							Optional:     true,
							Default:      1,
							ValidateFunc: validation.IntBetween(1, 50),
							Deprecated:   DeprecationMsgOldSchema,
						},
						"region_configs": {
							Type:     schema.TypeList,
							Required: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"analytics_specs": schemaSpecs(),
									"auto_scaling": {
										Type:     schema.TypeList,
										MaxItems: 1,
										Optional: true,
										Computed: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"disk_gb_enabled": {
													Type:     schema.TypeBool,
													Optional: true,
													Computed: true,
												},
												"compute_enabled": {
													Type:     schema.TypeBool,
													Optional: true,
													Computed: true,
												},
												"compute_scale_down_enabled": {
													Type:     schema.TypeBool,
													Optional: true,
													Computed: true,
												},
												"compute_min_instance_size": {
													Type:     schema.TypeString,
													Optional: true,
													Computed: true,
												},
												"compute_max_instance_size": {
													Type:     schema.TypeString,
													Optional: true,
													Computed: true,
												},
											},
										},
									},
									"analytics_auto_scaling": {
										Type:     schema.TypeList,
										MaxItems: 1,
										Optional: true,
										Computed: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"disk_gb_enabled": {
													Type:     schema.TypeBool,
													Optional: true,
													Computed: true,
												},
												"compute_enabled": {
													Type:     schema.TypeBool,
													Optional: true,
													Computed: true,
												},
												"compute_scale_down_enabled": {
													Type:     schema.TypeBool,
													Optional: true,
													Computed: true,
												},
												"compute_min_instance_size": {
													Type:     schema.TypeString,
													Optional: true,
													Computed: true,
												},
												"compute_max_instance_size": {
													Type:     schema.TypeString,
													Optional: true,
													Computed: true,
												},
											},
										},
									},
									"backing_provider_name": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"electable_specs": schemaSpecs(),
									"priority": {
										Type:     schema.TypeInt,
										Required: true,
									},
									"provider_name": {
										Type:             schema.TypeString,
										Required:         true,
										ValidateDiagFunc: validate.StringIsUppercase(),
									},
									"read_only_specs": schemaSpecs(),
									"region_name": {
										Type:             schema.TypeString,
										Required:         true,
										ValidateDiagFunc: validate.StringIsUppercase(),
									},
								},
							},
						},
						"container_id": {
							Type: schema.TypeMap,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
							Computed: true,
						},
						"zone_name": {
							Type:     schema.TypeString,
							Optional: true,
							Default:  "ZoneName managed by Terraform",
						},
					},
				},
			},
			"root_cert_type": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"state_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"termination_protection_enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
			},
			"version_release_system": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringInSlice([]string{"LTS", "CONTINUOUS"}, false),
			},
			"advanced_configuration": SchemaAdvancedConfig(),
			"accept_data_risks_and_force_replica_set_reconfig": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Submit this field alongside your topology reconfiguration to request a new regional outage resistant topology",
			},
			"global_cluster_self_managed_sharding": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
			},
			"replica_set_scaling_strategy": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"redact_client_log_data": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
			},
			"config_server_management_mode": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"config_server_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"pinned_fcv": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"version": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"expiration_date": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(3 * time.Hour),
			Update: schema.DefaultTimeout(3 * time.Hour),
			Delete: schema.DefaultTimeout(3 * time.Hour),
		},
	}
}

func schemaSpecs() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		MaxItems: 1,
		Optional: true,
		Computed: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"disk_size_gb": {
					Type:     schema.TypeFloat,
					Optional: true,
					Computed: true,
				},
				"disk_iops": {
					Type:     schema.TypeInt,
					Optional: true,
					Computed: true,
				},
				"ebs_volume_type": {
					Type:     schema.TypeString,
					Optional: true,
					Computed: true,
				},
				"instance_size": {
					Type:     schema.TypeString,
					Required: true,
				},
				"node_count": {
					Type:     schema.TypeInt,
					Optional: true,
				},
			},
		},
	}
}

func resourceCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	if v, ok := d.GetOk("accept_data_risks_and_force_replica_set_reconfig"); ok {
		if v.(string) != "" {
			return diag.FromErr(fmt.Errorf("accept_data_risks_and_force_replica_set_reconfig can not be set in creation, only in update"))
		}
	}
	connV220240530 := meta.(*config.MongoDBClient).AtlasV220240530
	connV220240805 := meta.(*config.MongoDBClient).AtlasV220240805
	connV2 := meta.(*config.MongoDBClient).AtlasV2
	projectID := d.Get("project_id").(string)

	var rootDiskSizeGB *float64
	if v, ok := d.GetOk("disk_size_gb"); ok {
		rootDiskSizeGB = conversion.Pointer(v.(float64))
	}

	params := &admin.ClusterDescription20240805{
		Name:             conversion.StringPtr(cast.ToString(d.Get("name"))),
		ClusterType:      conversion.StringPtr(cast.ToString(d.Get("cluster_type"))),
		ReplicationSpecs: expandAdvancedReplicationSpecs(d.Get("replication_specs").([]any), rootDiskSizeGB),
	}

	if v, ok := d.GetOk("backup_enabled"); ok {
		params.BackupEnabled = conversion.Pointer(v.(bool))
	}
	if _, ok := d.GetOk("bi_connector_config"); ok {
		params.BiConnector = expandBiConnectorConfig(d)
	}

	if v, ok := d.GetOk("encryption_at_rest_provider"); ok {
		params.EncryptionAtRestProvider = conversion.StringPtr(v.(string))
	}

	if _, ok := d.GetOk("labels"); ok {
		labels, err := expandLabelSliceFromSetSchema(d)
		if err != nil {
			return err
		}
		params.Labels = &labels
	}

	if _, ok := d.GetOk("tags"); ok {
		params.Tags = conversion.ExpandTagsFromSetSchema(d)
	}
	if v, ok := d.GetOk("mongo_db_major_version"); ok {
		params.MongoDBMajorVersion = conversion.StringPtr(FormatMongoDBMajorVersion(v.(string)))
	}
	if v, ok := d.GetOk("pit_enabled"); ok {
		params.PitEnabled = conversion.Pointer(v.(bool))
	}
	if v, ok := d.GetOk("root_cert_type"); ok {
		params.RootCertType = conversion.StringPtr(v.(string))
	}
	if v, ok := d.GetOk("termination_protection_enabled"); ok {
		params.TerminationProtectionEnabled = conversion.Pointer(v.(bool))
	}
	if v, ok := d.GetOk("version_release_system"); ok {
		params.VersionReleaseSystem = conversion.StringPtr(v.(string))
	}
	if v, ok := d.GetOk("global_cluster_self_managed_sharding"); ok {
		params.GlobalClusterSelfManagedSharding = conversion.Pointer(v.(bool))
	}
	if v, ok := d.GetOk("replica_set_scaling_strategy"); ok {
		params.ReplicaSetScalingStrategy = conversion.StringPtr(v.(string))
	}
	if v, ok := d.GetOk("redact_client_log_data"); ok {
		params.RedactClientLogData = conversion.Pointer(v.(bool))
	}
	if v, ok := d.GetOk("config_server_management_mode"); ok {
		params.ConfigServerManagementMode = conversion.StringPtr(v.(string))
	}

	// Validate advanced configuration params to show the error before the cluster is created.
	if oplogSizeMB, ok := d.GetOkExists("advanced_configuration.0.oplog_size_mb"); ok {
		if cast.ToInt64(oplogSizeMB) < 0 {
			return diag.FromErr(fmt.Errorf("`advanced_configuration.oplog_size_mb` cannot be < 0"))
		}
	}
	if _, ok := d.GetOkExists("advanced_configuration.0.default_max_time_ms"); ok {
		if !IsDefaultMaxTimeMinRequiredMajorVersion(params.MongoDBMajorVersion) {
			return diag.FromErr(errors.New(ErrorDefaultMaxTimeMinVersion))
		}
	}

	if err := CheckRegionConfigsPriorityOrder(params.GetReplicationSpecs()); err != nil {
		return diag.FromErr(err)
	}

	var clusterName string
	var clusterID string
	var err error
	// With old sharding config we call older API (2024-08-05) to avoid cluster having asymmetric autoscaling mode. Old sharding config can only represent symmetric clusters.
	if isUsingOldShardingConfiguration(d) {
		var cluster20240805 *admin20240805.ClusterDescription20240805
		cluster20240805, _, err = connV220240805.ClustersApi.CreateCluster(ctx, projectID, ConvertClusterDescription20241023to20240805(params)).Execute()
		if err != nil {
			return diag.FromErr(fmt.Errorf(errorCreate, err))
		}
		clusterName = cluster20240805.GetName()
		clusterID = cluster20240805.GetId()
	} else {
		var cluster *admin.ClusterDescription20240805
		cluster, _, err = connV2.ClustersApi.CreateCluster(ctx, projectID, params).Execute()
		if err != nil {
			return diag.FromErr(fmt.Errorf(errorCreate, err))
		}
		clusterName = cluster.GetName()
		clusterID = cluster.GetId()
	}

	timeout := d.Timeout(schema.TimeoutCreate)
	stateConf := CreateStateChangeConfig(ctx, connV2, projectID, d.Get("name").(string), timeout)
	_, err = stateConf.WaitForStateContext(ctx)
	if err != nil {
		return diag.FromErr(fmt.Errorf(errorCreate, err))
	}

	if ac, ok := d.GetOk("advanced_configuration"); ok {
		if aclist, ok := ac.([]any); ok && len(aclist) > 0 {
			params20240530, params := expandProcessArgs(d, aclist[0].(map[string]any), params.MongoDBMajorVersion)
			_, _, err := connV220240530.ClustersApi.UpdateClusterAdvancedConfiguration(ctx, projectID, clusterName, &params20240530).Execute()
			if err != nil {
				return diag.FromErr(fmt.Errorf(errorConfigUpdate, clusterName, err))
			}
			_, _, err = connV2.ClustersApi.UpdateClusterAdvancedConfiguration(ctx, projectID, clusterName, &params).Execute()
			if err != nil {
				return diag.FromErr(fmt.Errorf(errorConfigUpdate, clusterName, err))
			}
		}
	}

	var waitForChanges bool
	if v := d.Get("paused").(bool); v {
		request := &admin.ClusterDescription20240805{
			Paused: conversion.Pointer(v),
		}
		// can call latest API (2024-10-23 or newer) as replications specs (with nested autoscaling property) is not specified
		if _, _, err := connV2.ClustersApi.UpdateCluster(ctx, projectID, d.Get("name").(string), request).Execute(); err != nil {
			return diag.FromErr(fmt.Errorf(errorUpdate, d.Get("name").(string), err))
		}
		waitForChanges = true
	}

	if pinnedFCVBlock, _ := d.Get("pinned_fcv").([]any); len(pinnedFCVBlock) > 0 {
		nestedObj := pinnedFCVBlock[0].(map[string]any)
		expDateStr := cast.ToString(nestedObj["expiration_date"])
		if err := advancedclustertpf.PinFCV(ctx, connV2.ClustersApi, projectID, clusterName, expDateStr); err != nil {
			return diag.FromErr(fmt.Errorf(errorUpdate, clusterName, err))
		}
		waitForChanges = true
	}

	if waitForChanges {
		if err = waitForUpdateToFinish(ctx, connV2, projectID, d.Get("name").(string), timeout); err != nil {
			return diag.FromErr(fmt.Errorf(errorUpdate, d.Get("name").(string), err))
		}
	}

	d.SetId(conversion.EncodeStateID(map[string]string{
		"cluster_id":   clusterID,
		"project_id":   projectID,
		"cluster_name": clusterName,
	}))

	return resourceRead(ctx, d, meta)
}

func CreateStateChangeConfig(ctx context.Context, connV2 *admin.APIClient, projectID, name string, timeout time.Duration) retry.StateChangeConf {
	return retry.StateChangeConf{
		Pending:    []string{"CREATING", "UPDATING", "REPAIRING", "REPEATING", "PENDING"},
		Target:     []string{"IDLE"},
		Refresh:    resourceRefreshFunc(ctx, name, projectID, connV2),
		Timeout:    timeout,
		MinTimeout: 1 * time.Minute,
		Delay:      3 * time.Minute,
	}
}

func resourceRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	connV220240530 := meta.(*config.MongoDBClient).AtlasV220240530
	connV2 := meta.(*config.MongoDBClient).AtlasV2
	ids := conversion.DecodeStateID(d.Id())
	projectID := ids["project_id"]
	clusterName := ids["cluster_name"]

	var replicationSpecs []map[string]any

	cluster, resp, err := connV2.ClustersApi.GetCluster(ctx, projectID, clusterName).Execute()
	if err != nil {
		if resp != nil && resp.StatusCode == http.StatusNotFound {
			d.SetId("")
			return nil
		}
		return diag.FromErr(fmt.Errorf(errorRead, clusterName, err))
	}

	zoneNameToOldReplicationSpecMeta, err := GetReplicationSpecAttributesFromOldAPI(ctx, projectID, clusterName, connV220240530.ClustersApi)
	if err != nil {
		if apiError, ok := admin20240530.AsError(err); !ok {
			return diag.FromErr(err)
		} else if apiError.GetErrorCode() != "ASYMMETRIC_SHARD_UNSUPPORTED" || (apiError.GetErrorCode() == "ASYMMETRIC_SHARD_UNSUPPORTED" && isUsingOldShardingConfiguration(d)) {
			return diag.FromErr(fmt.Errorf(errorRead, clusterName, err))
		}
	}
	// if config uses old sharding configuration we call latest API but group replications specs from the same zone and define num_shards attribute
	if isUsingOldShardingConfiguration(d) {
		replicationSpecs, err = FlattenAdvancedReplicationSpecsOldShardingConfig(ctx, cluster.GetReplicationSpecs(), zoneNameToOldReplicationSpecMeta, d.Get("replication_specs").([]any), d, connV2)
		if err != nil {
			return diag.FromErr(fmt.Errorf(ErrorClusterAdvancedSetting, "replication_specs", clusterName, err))
		}
	} else {
		replicationSpecs, err = flattenAdvancedReplicationSpecs(ctx, cluster.GetReplicationSpecs(), zoneNameToOldReplicationSpecMeta, d.Get("replication_specs").([]any), d, connV2)
		if err != nil {
			return diag.FromErr(fmt.Errorf(ErrorClusterAdvancedSetting, "replication_specs", clusterName, err))
		}
	}

	warning := WarningIfFCVExpiredOrUnpinnedExternally(d, cluster) // has to be called before pinned_fcv value is updated in ResourceData to know prior state value
	diags := setRootFields(d, cluster, true)
	if diags.HasError() {
		return diags
	}

	if err := d.Set("replication_specs", replicationSpecs); err != nil {
		return diag.FromErr(fmt.Errorf(ErrorClusterAdvancedSetting, "replication_specs", clusterName, err))
	}

	processArgs20240530, _, err := connV220240530.ClustersApi.GetClusterAdvancedConfiguration(ctx, projectID, clusterName).Execute()
	if err != nil {
		return diag.FromErr(fmt.Errorf(errorConfigRead, clusterName, err))
	}
	processArgs, _, err := connV2.ClustersApi.GetClusterAdvancedConfiguration(ctx, projectID, clusterName).Execute()
	if err != nil {
		return diag.FromErr(fmt.Errorf(errorConfigRead, clusterName, err))
	}

	if err := d.Set("advanced_configuration", flattenProcessArgs(processArgs20240530, processArgs)); err != nil {
		return diag.FromErr(fmt.Errorf(ErrorClusterAdvancedSetting, "advanced_configuration", clusterName, err))
	}

	return warning
}

// GetReplicationSpecAttributesFromOldAPI returns the id and num shard values of replication specs coming from old API. This is used to populate replication_specs.*.id and replication_specs.*.num_shard attributes for old sharding confirgurations.
// In the old API (2023-02-01), each replications spec has a 1:1 relation with each zone, so ids and num shards are stored in a struct oldShardConfigMeta and are returned in a map from zoneName to oldShardConfigMeta.
func GetReplicationSpecAttributesFromOldAPI(ctx context.Context, projectID, clusterName string, client20240530 admin20240530.ClustersApi) (map[string]OldShardConfigMeta, error) {
	clusterOldAPI, _, err := client20240530.GetCluster(ctx, projectID, clusterName).Execute()
	if err != nil {
		return nil, err
	}
	specs := clusterOldAPI.GetReplicationSpecs()
	result := make(map[string]OldShardConfigMeta, len(specs))
	for _, spec := range specs {
		result[spec.GetZoneName()] = OldShardConfigMeta{spec.GetId(), spec.GetNumShards()}
	}
	return result, nil
}

func setRootFields(d *schema.ResourceData, cluster *admin.ClusterDescription20240805, isResourceSchema bool) diag.Diagnostics {
	clusterName := *cluster.Name

	if isResourceSchema {
		if err := d.Set("cluster_id", cluster.GetId()); err != nil {
			return diag.FromErr(fmt.Errorf(ErrorClusterAdvancedSetting, "cluster_id", clusterName, err))
		}

		if err := d.Set("accept_data_risks_and_force_replica_set_reconfig", conversion.TimePtrToStringPtr(cluster.AcceptDataRisksAndForceReplicaSetReconfig)); err != nil {
			return diag.FromErr(fmt.Errorf(ErrorClusterAdvancedSetting, "accept_data_risks_and_force_replica_set_reconfig", clusterName, err))
		}
	}

	if err := d.Set("backup_enabled", cluster.GetBackupEnabled()); err != nil {
		return diag.FromErr(fmt.Errorf(ErrorClusterAdvancedSetting, "backup_enabled", clusterName, err))
	}

	if err := d.Set("bi_connector_config", flattenBiConnectorConfig(cluster.BiConnector)); err != nil {
		return diag.FromErr(fmt.Errorf(ErrorClusterAdvancedSetting, "bi_connector_config", clusterName, err))
	}

	if err := d.Set("cluster_type", cluster.GetClusterType()); err != nil {
		return diag.FromErr(fmt.Errorf(ErrorClusterAdvancedSetting, "cluster_type", clusterName, err))
	}

	if err := d.Set("connection_strings", flattenConnectionStrings(*cluster.ConnectionStrings)); err != nil {
		return diag.FromErr(fmt.Errorf(ErrorClusterAdvancedSetting, "connection_strings", clusterName, err))
	}

	if err := d.Set("create_date", conversion.TimePtrToStringPtr(cluster.CreateDate)); err != nil {
		return diag.FromErr(fmt.Errorf(ErrorClusterAdvancedSetting, "create_date", clusterName, err))
	}

	// root disk_size_gb defined for backwards compatibility avoiding breaking changes
	if err := d.Set("disk_size_gb", GetDiskSizeGBFromReplicationSpec(cluster)); err != nil {
		return diag.FromErr(fmt.Errorf(ErrorClusterAdvancedSetting, "disk_size_gb", clusterName, err))
	}

	if err := d.Set("encryption_at_rest_provider", cluster.GetEncryptionAtRestProvider()); err != nil {
		return diag.FromErr(fmt.Errorf(ErrorClusterAdvancedSetting, "encryption_at_rest_provider", clusterName, err))
	}

	if err := d.Set("labels", flattenLabels(cluster.GetLabels())); err != nil {
		return diag.FromErr(fmt.Errorf(ErrorClusterAdvancedSetting, "labels", clusterName, err))
	}

	if err := d.Set("tags", flattenTags(cluster.Tags)); err != nil {
		return diag.FromErr(fmt.Errorf(ErrorClusterAdvancedSetting, "tags", clusterName, err))
	}

	if err := d.Set("mongo_db_major_version", cluster.GetMongoDBMajorVersion()); err != nil {
		return diag.FromErr(fmt.Errorf(ErrorClusterAdvancedSetting, "mongo_db_major_version", clusterName, err))
	}

	if err := d.Set("mongo_db_version", cluster.GetMongoDBVersion()); err != nil {
		return diag.FromErr(fmt.Errorf(ErrorClusterAdvancedSetting, "mongo_db_version", clusterName, err))
	}

	if err := d.Set("name", cluster.GetName()); err != nil {
		return diag.FromErr(fmt.Errorf(ErrorClusterAdvancedSetting, "name", clusterName, err))
	}

	if err := d.Set("paused", cluster.GetPaused()); err != nil {
		return diag.FromErr(fmt.Errorf(ErrorClusterAdvancedSetting, "paused", clusterName, err))
	}

	if err := d.Set("pit_enabled", cluster.GetPitEnabled()); err != nil {
		return diag.FromErr(fmt.Errorf(ErrorClusterAdvancedSetting, "pit_enabled", clusterName, err))
	}

	if err := d.Set("root_cert_type", cluster.GetRootCertType()); err != nil {
		return diag.FromErr(fmt.Errorf(ErrorClusterAdvancedSetting, "state_name", clusterName, err))
	}

	if err := d.Set("state_name", cluster.GetStateName()); err != nil {
		return diag.FromErr(fmt.Errorf(ErrorClusterAdvancedSetting, "state_name", clusterName, err))
	}

	if err := d.Set("termination_protection_enabled", cluster.GetTerminationProtectionEnabled()); err != nil {
		return diag.FromErr(fmt.Errorf(ErrorClusterAdvancedSetting, "termination_protection_enabled", clusterName, err))
	}

	if err := d.Set("version_release_system", cluster.GetVersionReleaseSystem()); err != nil {
		return diag.FromErr(fmt.Errorf(ErrorClusterAdvancedSetting, "version_release_system", clusterName, err))
	}

	if err := d.Set("global_cluster_self_managed_sharding", cluster.GetGlobalClusterSelfManagedSharding()); err != nil {
		return diag.FromErr(fmt.Errorf(ErrorClusterAdvancedSetting, "global_cluster_self_managed_sharding", clusterName, err))
	}

	if err := d.Set("replica_set_scaling_strategy", cluster.GetReplicaSetScalingStrategy()); err != nil {
		return diag.FromErr(fmt.Errorf(ErrorClusterAdvancedSetting, "replica_set_scaling_strategy", clusterName, err))
	}
	if err := d.Set("redact_client_log_data", cluster.GetRedactClientLogData()); err != nil {
		return diag.FromErr(fmt.Errorf(ErrorClusterAdvancedSetting, "redact_client_log_data", clusterName, err))
	}

	if err := d.Set("config_server_type", cluster.GetConfigServerType()); err != nil {
		return diag.FromErr(fmt.Errorf(ErrorClusterAdvancedSetting, "config_server_type", clusterName, err))
	}

	if err := d.Set("config_server_management_mode", cluster.GetConfigServerManagementMode()); err != nil {
		return diag.FromErr(fmt.Errorf(ErrorClusterAdvancedSetting, "config_server_management_mode", clusterName, err))
	}

	if err := d.Set("pinned_fcv", FlattenPinnedFCV(cluster)); err != nil {
		return diag.FromErr(fmt.Errorf(ErrorClusterAdvancedSetting, "pinned_fcv", clusterName, err))
	}

	return nil
}

func WarningIfFCVExpiredOrUnpinnedExternally(d *schema.ResourceData, cluster *admin.ClusterDescription20240805) diag.Diagnostics {
	pinnedFCVBlock, _ := d.Get("pinned_fcv").([]any)
	fcvPresentInState := len(pinnedFCVBlock) > 0
	diagsTpf := advancedclustertpf.GenerateFCVPinningWarningForRead(fcvPresentInState, cluster.FeatureCompatibilityVersionExpirationDate)
	return conversion.FromTPFDiagsToSDKV2Diags(diagsTpf)
}

// isUsingOldShardingConfiguration is identified if at least one replication spec defines num_shards > 1. This legacy form is from 2023-02-01 API and can only represent symmetric sharded clusters.
func isUsingOldShardingConfiguration(d *schema.ResourceData) bool {
	tfList := d.Get("replication_specs").([]any)
	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]any)
		if !ok || tfMap == nil {
			continue
		}
		numShards := tfMap["num_shards"].(int)
		if numShards > 1 {
			return true
		}
	}
	return false
}

func resourceUpdateOrUpgrade(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	if upgradeRequest := getUpgradeRequest(d); upgradeRequest != nil {
		return resourceUpgrade(ctx, upgradeRequest, d, meta)
	}
	return resourceUpdate(ctx, d, meta)
}

func resourceUpgrade(ctx context.Context, upgradeRequest *admin.LegacyAtlasTenantClusterUpgradeRequest, d *schema.ResourceData, meta any) diag.Diagnostics {
	connV2 := meta.(*config.MongoDBClient).AtlasV2
	ids := conversion.DecodeStateID(d.Id())
	projectID := ids["project_id"]
	clusterName := ids["cluster_name"]

	upgradeResponse, _, err := upgradeCluster(ctx, connV2, upgradeRequest, projectID, clusterName, d.Timeout(schema.TimeoutUpdate))

	if err != nil {
		return diag.FromErr(fmt.Errorf(errorUpdate, clusterName, err))
	}

	d.SetId(conversion.EncodeStateID(map[string]string{
		"cluster_id":   upgradeResponse.GetId(),
		"project_id":   projectID,
		"cluster_name": clusterName,
	}))

	return resourceRead(ctx, d, meta)
}

func resourceUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	connV220240530 := meta.(*config.MongoDBClient).AtlasV220240530
	connV2 := meta.(*config.MongoDBClient).AtlasV2
	ids := conversion.DecodeStateID(d.Id())
	projectID := ids["project_id"]
	clusterName := ids["cluster_name"]

	if v, err := isUpdateAllowed(d); !v {
		return diag.FromErr(fmt.Errorf("%s: %s", ErrorOperationNotPermitted, err))
	}

	timeout := d.Timeout(schema.TimeoutUpdate)

	// FCV update is intentionally handled before other cluster updates, and will wait for cluster to reach IDLE state before continuing
	if diags := HandlePinnedFCVUpdate(ctx, connV2, projectID, clusterName, d, timeout); diags != nil {
		return diags
	}

	// With old sharding config we call older API (2023-02-01) to avoid cluster having asymmetric autoscaling mode. Old sharding config can only represent symmetric clusters.
	if isUsingOldShardingConfiguration(d) {
		req, diags := updateRequestOldAPI(d, clusterName)
		if diags != nil {
			return diags
		}
		clusterChangeDetect := new(admin20240530.AdvancedClusterDescription)
		var waitOnUpdate bool
		if !reflect.DeepEqual(req, clusterChangeDetect) {
			if err := CheckRegionConfigsPriorityOrderOld(req.GetReplicationSpecs()); err != nil {
				return diag.FromErr(err)
			}
			if _, _, err := connV220240530.ClustersApi.UpdateCluster(ctx, projectID, clusterName, req).Execute(); err != nil {
				return diag.FromErr(fmt.Errorf(errorUpdate, clusterName, err))
			}
			waitOnUpdate = true
		}
		if d.HasChange("replica_set_scaling_strategy") || d.HasChange("redact_client_log_data") || d.HasChange("config_server_management_mode") {
			request := new(admin.ClusterDescription20240805)
			if d.HasChange("replica_set_scaling_strategy") {
				request.ReplicaSetScalingStrategy = conversion.Pointer(d.Get("replica_set_scaling_strategy").(string))
			}
			if d.HasChange("redact_client_log_data") {
				request.RedactClientLogData = conversion.Pointer(d.Get("redact_client_log_data").(bool))
			}
			if d.HasChange("config_server_management_mode") {
				request.ConfigServerManagementMode = conversion.StringPtr(d.Get("config_server_management_mode").(string))
			}
			// can call latest API (2024-10-23 or newer) as replications specs (with nested autoscaling property) is not specified
			if _, _, err := connV2.ClustersApi.UpdateCluster(ctx, projectID, clusterName, request).Execute(); err != nil {
				return diag.FromErr(fmt.Errorf(errorUpdate, clusterName, err))
			}
			waitOnUpdate = true
		}
		if waitOnUpdate {
			if err := waitForUpdateToFinish(ctx, connV2, projectID, clusterName, timeout); err != nil {
				return diag.FromErr(fmt.Errorf(errorUpdate, clusterName, err))
			}
		}
	} else {
		req, diags := updateRequest(ctx, d, projectID, clusterName, connV2)
		if diags != nil {
			return diags
		}
		clusterChangeDetect := new(admin.ClusterDescription20240805)
		if !reflect.DeepEqual(req, clusterChangeDetect) {
			if err := CheckRegionConfigsPriorityOrder(req.GetReplicationSpecs()); err != nil {
				return diag.FromErr(err)
			}
			if _, _, err := connV2.ClustersApi.UpdateCluster(ctx, projectID, clusterName, req).Execute(); err != nil {
				return diag.FromErr(fmt.Errorf(errorUpdate, clusterName, err))
			}
			if err := waitForUpdateToFinish(ctx, connV2, projectID, clusterName, timeout); err != nil {
				return diag.FromErr(fmt.Errorf(errorUpdate, clusterName, err))
			}
		}
	}

	if d.HasChange("advanced_configuration") {
		var mongoDBMajorVersion string
		if v, ok := d.GetOk("mongo_db_major_version"); ok {
			mongoDBMajorVersion = v.(string)
		}

		ac := d.Get("advanced_configuration")
		if aclist, ok := ac.([]any); ok && len(aclist) > 0 {
			params20240530, params := expandProcessArgs(d, aclist[0].(map[string]any), &mongoDBMajorVersion)
			if !reflect.DeepEqual(params20240530, admin20240530.ClusterDescriptionProcessArgs{}) {
				_, _, err := connV220240530.ClustersApi.UpdateClusterAdvancedConfiguration(ctx, projectID, clusterName, &params20240530).Execute()
				if err != nil {
					return diag.FromErr(fmt.Errorf(errorConfigUpdate, clusterName, err))
				}
				if err := waitForUpdateToFinish(ctx, connV2, projectID, clusterName, timeout); err != nil {
					return diag.FromErr(fmt.Errorf(errorUpdate, clusterName, err))
				}
			}
			if !reflect.DeepEqual(params, admin.ClusterDescriptionProcessArgs20240805{}) {
				_, _, err := connV2.ClustersApi.UpdateClusterAdvancedConfiguration(ctx, projectID, clusterName, &params).Execute()
				if err != nil {
					return diag.FromErr(fmt.Errorf(errorConfigUpdate, clusterName, err))
				}
				if err := waitForUpdateToFinish(ctx, connV2, projectID, clusterName, timeout); err != nil {
					return diag.FromErr(fmt.Errorf(errorUpdate, clusterName, err))
				}
			}
		}
	}

	if d.Get("paused").(bool) {
		clusterRequest := &admin.ClusterDescription20240805{
			Paused: conversion.Pointer(true),
		}
		if _, _, err := connV2.ClustersApi.UpdateCluster(ctx, projectID, clusterName, clusterRequest).Execute(); err != nil {
			return diag.FromErr(fmt.Errorf(errorUpdate, clusterName, err))
		}
		if err := waitForUpdateToFinish(ctx, connV2, projectID, clusterName, timeout); err != nil {
			return diag.FromErr(fmt.Errorf(errorUpdate, clusterName, err))
		}
	}

	return resourceRead(ctx, d, meta)
}

func HandlePinnedFCVUpdate(ctx context.Context, connV2 *admin.APIClient, projectID, clusterName string, d *schema.ResourceData, timeout time.Duration) diag.Diagnostics {
	if d.HasChange("pinned_fcv") {
		pinnedFCVBlock, _ := d.Get("pinned_fcv").([]any)
		isFCVPresentInConfig := len(pinnedFCVBlock) > 0
		if isFCVPresentInConfig {
			// pinned_fcv has been defined or updated expiration date
			nestedObj := pinnedFCVBlock[0].(map[string]any)
			expDateStr := cast.ToString(nestedObj["expiration_date"])
			if err := advancedclustertpf.PinFCV(ctx, connV2.ClustersApi, projectID, clusterName, expDateStr); err != nil {
				return diag.FromErr(fmt.Errorf(errorUpdate, clusterName, err))
			}
		} else {
			// pinned_fcv has been removed from the config so unpin method is called
			if _, _, err := connV2.ClustersApi.UnpinFeatureCompatibilityVersion(ctx, projectID, clusterName).Execute(); err != nil {
				return diag.FromErr(fmt.Errorf(errorUpdate, clusterName, err))
			}
		}
		// ensures cluster is in IDLE state before continuing with other changes
		if err := waitForUpdateToFinish(ctx, connV2, projectID, clusterName, timeout); err != nil {
			return diag.FromErr(fmt.Errorf(errorUpdate, clusterName, err))
		}
	}
	return nil
}

func updateRequest(ctx context.Context, d *schema.ResourceData, projectID, clusterName string, connV2 *admin.APIClient) (*admin.ClusterDescription20240805, diag.Diagnostics) {
	cluster := new(admin.ClusterDescription20240805)

	if d.HasChange("replication_specs") || d.HasChange("disk_size_gb") {
		var updatedDiskSizeGB *float64
		if d.HasChange("disk_size_gb") {
			updatedDiskSizeGB = conversion.Pointer(d.Get("disk_size_gb").(float64))
		}
		updatedReplicationSpecs := expandAdvancedReplicationSpecs(d.Get("replication_specs").([]any), updatedDiskSizeGB)

		// case where sharding schema is transitioning from legacy to new structure (external_id is not present in the state so no ids are are currently present)
		if noIDsPopulatedInReplicationSpecs(updatedReplicationSpecs) {
			// ids need to be populated to avoid error in the update request
			specsWithIDs, diags := populateIDValuesUsingNewAPI(ctx, projectID, clusterName, connV2.ClustersApi, updatedReplicationSpecs)
			if diags != nil {
				return nil, diags
			}
			updatedReplicationSpecs = specsWithIDs
		}
		SyncAutoScalingConfigs(updatedReplicationSpecs)
		cluster.ReplicationSpecs = updatedReplicationSpecs
	}

	if d.HasChange("backup_enabled") {
		cluster.BackupEnabled = conversion.Pointer(d.Get("backup_enabled").(bool))
	}

	if d.HasChange("bi_connector_config") {
		cluster.BiConnector = expandBiConnectorConfig(d)
	}

	if d.HasChange("cluster_type") {
		cluster.ClusterType = conversion.StringPtr(d.Get("cluster_type").(string))
	}

	if d.HasChange("encryption_at_rest_provider") {
		cluster.EncryptionAtRestProvider = conversion.StringPtr(d.Get("encryption_at_rest_provider").(string))
	}

	if d.HasChange("labels") {
		labels, err := expandLabelSliceFromSetSchema(d)
		if err != nil {
			return nil, err
		}
		cluster.Labels = &labels
	}

	if d.HasChange("tags") {
		cluster.Tags = conversion.ExpandTagsFromSetSchema(d)
	}

	if d.HasChange("mongo_db_major_version") {
		cluster.MongoDBMajorVersion = conversion.StringPtr(FormatMongoDBMajorVersion(d.Get("mongo_db_major_version")))
	}

	if d.HasChange("pit_enabled") {
		cluster.PitEnabled = conversion.Pointer(d.Get("pit_enabled").(bool))
	}

	if d.HasChange("root_cert_type") {
		cluster.RootCertType = conversion.StringPtr(d.Get("root_cert_type").(string))
	}

	if d.HasChange("termination_protection_enabled") {
		cluster.TerminationProtectionEnabled = conversion.Pointer(d.Get("termination_protection_enabled").(bool))
	}

	if d.HasChange("version_release_system") {
		cluster.VersionReleaseSystem = conversion.StringPtr(d.Get("version_release_system").(string))
	}

	if d.HasChange("global_cluster_self_managed_sharding") {
		cluster.GlobalClusterSelfManagedSharding = conversion.Pointer(d.Get("global_cluster_self_managed_sharding").(bool))
	}

	if d.HasChange("accept_data_risks_and_force_replica_set_reconfig") {
		if strTime := d.Get("accept_data_risks_and_force_replica_set_reconfig").(string); strTime != "" {
			t, ok := conversion.StringToTime(strTime)
			if !ok {
				return nil, diag.FromErr(fmt.Errorf(errorUpdate, clusterName, "accept_data_risks_and_force_replica_set_reconfig time format is incorrect"))
			}
			cluster.AcceptDataRisksAndForceReplicaSetReconfig = &t
		}
	}

	if d.HasChange("paused") && !d.Get("paused").(bool) {
		cluster.Paused = conversion.Pointer(d.Get("paused").(bool))
	}

	if d.HasChange("replica_set_scaling_strategy") {
		cluster.ReplicaSetScalingStrategy = conversion.Pointer(d.Get("replica_set_scaling_strategy").(string))
	}
	if d.HasChange("redact_client_log_data") {
		cluster.RedactClientLogData = conversion.Pointer(d.Get("redact_client_log_data").(bool))
	}
	if d.HasChange("config_server_management_mode") {
		cluster.ConfigServerManagementMode = conversion.StringPtr(d.Get("config_server_management_mode").(string))
	}

	return cluster, nil
}

func updateRequestOldAPI(d *schema.ResourceData, clusterName string) (*admin20240530.AdvancedClusterDescription, diag.Diagnostics) {
	cluster := new(admin20240530.AdvancedClusterDescription)

	if d.HasChange("replication_specs") {
		cluster.ReplicationSpecs = expandAdvancedReplicationSpecsOldSDK(d.Get("replication_specs").([]any))
	}

	if d.HasChange("disk_size_gb") {
		cluster.DiskSizeGB = conversion.Pointer(d.Get("disk_size_gb").(float64))
	}

	if changedValue := obtainChangeForDiskSizeGBInFirstRegion(d); changedValue != nil {
		cluster.DiskSizeGB = changedValue
	}

	if d.HasChange("backup_enabled") {
		cluster.BackupEnabled = conversion.Pointer(d.Get("backup_enabled").(bool))
	}

	if d.HasChange("bi_connector_config") {
		cluster.BiConnector = convertBiConnectToOldSDK(expandBiConnectorConfig(d))
	}

	if d.HasChange("cluster_type") {
		cluster.ClusterType = conversion.StringPtr(d.Get("cluster_type").(string))
	}

	if d.HasChange("encryption_at_rest_provider") {
		cluster.EncryptionAtRestProvider = conversion.StringPtr(d.Get("encryption_at_rest_provider").(string))
	}

	if d.HasChange("labels") {
		labels, err := convertLabelSliceToOldSDK(expandLabelSliceFromSetSchema(d))
		if err != nil {
			return nil, err
		}
		cluster.Labels = &labels
	}

	if d.HasChange("tags") {
		cluster.Tags = convertTagsPtrToOldSDK(conversion.ExpandTagsFromSetSchema(d))
	}

	if d.HasChange("mongo_db_major_version") {
		cluster.MongoDBMajorVersion = conversion.StringPtr(FormatMongoDBMajorVersion(d.Get("mongo_db_major_version")))
	}

	if d.HasChange("pit_enabled") {
		cluster.PitEnabled = conversion.Pointer(d.Get("pit_enabled").(bool))
	}

	if d.HasChange("root_cert_type") {
		cluster.RootCertType = conversion.StringPtr(d.Get("root_cert_type").(string))
	}

	if d.HasChange("termination_protection_enabled") {
		cluster.TerminationProtectionEnabled = conversion.Pointer(d.Get("termination_protection_enabled").(bool))
	}

	if d.HasChange("version_release_system") {
		cluster.VersionReleaseSystem = conversion.StringPtr(d.Get("version_release_system").(string))
	}

	if d.HasChange("global_cluster_self_managed_sharding") {
		cluster.GlobalClusterSelfManagedSharding = conversion.Pointer(d.Get("global_cluster_self_managed_sharding").(bool))
	}

	if d.HasChange("accept_data_risks_and_force_replica_set_reconfig") {
		if strTime := d.Get("accept_data_risks_and_force_replica_set_reconfig").(string); strTime != "" {
			t, ok := conversion.StringToTime(strTime)
			if !ok {
				return nil, diag.FromErr(fmt.Errorf(errorUpdate, clusterName, "accept_data_risks_and_force_replica_set_reconfig time format is incorrect"))
			}
			cluster.AcceptDataRisksAndForceReplicaSetReconfig = &t
		}
	}

	if d.HasChange("paused") && !d.Get("paused").(bool) {
		cluster.Paused = conversion.Pointer(d.Get("paused").(bool))
	}
	return cluster, nil
}

func isUpdateAllowed(d *schema.ResourceData) (bool, error) {
	cs, us := d.GetChange("replication_specs")
	currentSpecs, updatedSpecs := cs.([]any), us.([]any)

	isNewSchemaCompatible := checkNewSchemaCompatibility(currentSpecs)

	for _, specRaw := range updatedSpecs {
		if specMap, ok := specRaw.(map[string]any); ok && specMap != nil {
			numShards, _ := specMap["num_shards"].(int)
			if numShards > 1 && isNewSchemaCompatible {
				return false, fmt.Errorf("cannot increase num_shards to > 1 under the current configuration. New shards can be defined by adding new replication spec objects; %s", DeprecationOldSchemaAction)
			}
		}
	}
	return true, nil
}

func checkNewSchemaCompatibility(specs []any) bool {
	for _, specRaw := range specs {
		if specMap, ok := specRaw.(map[string]any); ok && specMap != nil {
			numShards, _ := specMap["num_shards"].(int)
			if numShards >= 2 {
				return false
			}
		}
	}
	return true
}

// When legacy schema structure is used we invoke the old API for updates. This API sends diskSizeGB at root level.
// This function is used to detect if changes are made in the inner spec levels. It assumes that all disk_size_gb values at the inner spec level have the same value, so it looks into first region config.
func obtainChangeForDiskSizeGBInFirstRegion(d *schema.ResourceData) *float64 {
	electableLocation := "replication_specs.0.region_configs.0.electable_specs.0.disk_size_gb"
	readOnlyLocation := "replication_specs.0.region_configs.0.read_only_specs.0.disk_size_gb"
	analyticsLocation := "replication_specs.0.region_configs.0.analytics_specs.0.disk_size_gb"
	if d.HasChange(electableLocation) {
		return admin.PtrFloat64(d.Get(electableLocation).(float64))
	}
	if d.HasChange(readOnlyLocation) {
		return admin.PtrFloat64(d.Get(readOnlyLocation).(float64))
	}
	if d.HasChange(analyticsLocation) {
		return admin.PtrFloat64(d.Get(analyticsLocation).(float64))
	}
	return nil
}

func resourceDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	connV2 := meta.(*config.MongoDBClient).AtlasV2
	ids := conversion.DecodeStateID(d.Id())
	projectID := ids["project_id"]
	clusterName := ids["cluster_name"]

	params := &admin.DeleteClusterApiParams{
		GroupId:     projectID,
		ClusterName: clusterName,
	}
	if v, ok := d.GetOkExists("retain_backups_enabled"); ok {
		params.RetainBackups = conversion.Pointer(v.(bool))
	}

	_, err := connV2.ClustersApi.DeleteClusterWithParams(ctx, params).Execute()
	if err != nil {
		return diag.FromErr(fmt.Errorf(errorDelete, clusterName, err))
	}

	log.Println("[INFO] Waiting for MongoDB ClusterAdvanced to be destroyed")

	stateConf := DeleteStateChangeConfig(ctx, connV2, projectID, clusterName, d.Timeout(schema.TimeoutDelete))
	// Wait, catching any errors
	_, err = stateConf.WaitForStateContext(ctx)
	if err != nil {
		return diag.FromErr(fmt.Errorf(errorDelete, clusterName, err))
	}

	return nil
}

func DeleteStateChangeConfig(ctx context.Context, connV2 *admin.APIClient, projectID, name string, timeout time.Duration) retry.StateChangeConf {
	return retry.StateChangeConf{
		Pending:    []string{"IDLE", "CREATING", "UPDATING", "REPAIRING", "DELETING"},
		Target:     []string{"DELETED"},
		Refresh:    resourceRefreshFunc(ctx, name, projectID, connV2),
		Timeout:    timeout,
		MinTimeout: 30 * time.Second,
		Delay:      1 * time.Minute,
	}
}

func resourceImport(ctx context.Context, d *schema.ResourceData, meta any) ([]*schema.ResourceData, error) {
	connV2 := meta.(*config.MongoDBClient).AtlasV2

	projectID, name, err := splitSClusterAdvancedImportID(d.Id())
	if err != nil {
		return nil, err
	}

	cluster, _, err := connV2.ClustersApi.GetCluster(ctx, *projectID, *name).Execute()
	if err != nil {
		return nil, fmt.Errorf("couldn't import cluster %s in project %s, error: %s", *name, *projectID, err)
	}

	if err := d.Set("project_id", cluster.GetGroupId()); err != nil {
		log.Printf(ErrorClusterAdvancedSetting, "project_id", cluster.GetId(), err)
	}

	if err := d.Set("name", cluster.GetName()); err != nil {
		log.Printf(ErrorClusterAdvancedSetting, "name", cluster.GetId(), err)
	}

	d.SetId(conversion.EncodeStateID(map[string]string{
		"cluster_id":   cluster.GetId(),
		"project_id":   *projectID,
		"cluster_name": cluster.GetName(),
	}))

	return []*schema.ResourceData{d}, nil
}

func upgradeCluster(ctx context.Context, connV2 *admin.APIClient, request *admin.LegacyAtlasTenantClusterUpgradeRequest, projectID, name string, timeout time.Duration) (*admin.LegacyAtlasCluster, *http.Response, error) {
	request.Name = name

	cluster, resp, err := connV2.ClustersApi.UpgradeSharedCluster(ctx, projectID, request).Execute()
	if err != nil {
		return nil, nil, err
	}

	stateConf := &retry.StateChangeConf{
		Pending:    []string{"CREATING", "UPDATING", "REPAIRING"},
		Target:     []string{"IDLE"},
		Refresh:    UpgradeRefreshFunc(ctx, name, projectID, connV2.ClustersApi),
		Timeout:    timeout,
		MinTimeout: 30 * time.Second,
		Delay:      1 * time.Minute,
	}

	// Wait, catching any errors
	_, err = stateConf.WaitForStateContext(ctx)
	if err != nil {
		return nil, nil, err
	}

	return cluster, resp, nil
}

func splitSClusterAdvancedImportID(id string) (projectID, clusterName *string, err error) {
	var re = regexp.MustCompile(`(?s)^([0-9a-fA-F]{24})-(.*)$`)
	parts := re.FindStringSubmatch(id)

	if len(parts) != 3 {
		err = errors.New("import format error: to import a advanced cluster, use the format {project_id}-{name}")
		return
	}

	projectID = &parts[1]
	clusterName = &parts[2]

	return
}

func resourceRefreshFunc(ctx context.Context, name, projectID string, connV2 *admin.APIClient) retry.StateRefreshFunc {
	return func() (any, string, error) {
		cluster, resp, err := connV2.ClustersApi.GetCluster(ctx, projectID, name).Execute()
		if err != nil && strings.Contains(err.Error(), "reset by peer") {
			return nil, "REPEATING", nil
		}

		if err != nil && cluster == nil && resp == nil {
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

		state := cluster.GetStateName()
		return cluster, state, nil
	}
}

func replicationSpecsHashSet(v any) int {
	var buf bytes.Buffer
	m := v.(map[string]any)
	buf.WriteString(fmt.Sprintf("%d", m["num_shards"].(int)))
	buf.WriteString(fmt.Sprintf("%+v", m["region_configs"].(*schema.Set)))
	buf.WriteString(m["zone_name"].(string))
	return schema.HashString(buf.String())
}

func getUpgradeRequest(d *schema.ResourceData) *admin.LegacyAtlasTenantClusterUpgradeRequest {
	if !d.HasChange("replication_specs") {
		return nil
	}

	cs, us := d.GetChange("replication_specs")
	currentSpecs := expandAdvancedReplicationSpecsOldSDK(cs.([]any))
	updatedSpecs := expandAdvancedReplicationSpecsOldSDK(us.([]any))

	if currentSpecs == nil || updatedSpecs == nil || len(*currentSpecs) != 1 || len(*updatedSpecs) != 1 || len((*currentSpecs)[0].GetRegionConfigs()) != 1 || len((*updatedSpecs)[0].GetRegionConfigs()) != 1 {
		return nil
	}

	currentRegion := (*currentSpecs)[0].GetRegionConfigs()[0]
	updatedRegion := (*updatedSpecs)[0].GetRegionConfigs()[0]
	currentSize := conversion.SafeString(currentRegion.ElectableSpecs.InstanceSize)

	if currentRegion.ElectableSpecs.InstanceSize == updatedRegion.ElectableSpecs.InstanceSize || !IsSharedTier(currentSize) {
		return nil
	}

	return &admin.LegacyAtlasTenantClusterUpgradeRequest{
		ProviderSettings: &admin.ClusterProviderSettings{
			ProviderName:     updatedRegion.GetProviderName(),
			InstanceSizeName: updatedRegion.ElectableSpecs.InstanceSize,
			RegionName:       updatedRegion.RegionName,
		},
	}
}

func waitForUpdateToFinish(ctx context.Context, connV2 *admin.APIClient, projectID, name string, timeout time.Duration) error {
	stateConf := &retry.StateChangeConf{
		Pending:    []string{"CREATING", "UPDATING", "REPAIRING"},
		Target:     []string{"IDLE"},
		Refresh:    resourceRefreshFunc(ctx, name, projectID, connV2),
		Timeout:    timeout,
		MinTimeout: 30 * time.Second,
		Delay:      1 * time.Minute,
	}

	_, err := stateConf.WaitForStateContext(ctx)
	return err
}
