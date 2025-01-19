package cluster

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"reflect"
	"regexp"
	"time"

	admin20240530 "go.mongodb.org/atlas-sdk/v20240530005/admin"
	"go.mongodb.org/atlas-sdk/v20241113004/admin"
	matlas "go.mongodb.org/atlas/mongodbatlas"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/spf13/cast"

	"github.com/mongodb/terraform-provider-mongodbatlas/internal/common/conversion"
	"github.com/mongodb/terraform-provider-mongodbatlas/internal/common/validate"
	"github.com/mongodb/terraform-provider-mongodbatlas/internal/config"
	"github.com/mongodb/terraform-provider-mongodbatlas/internal/service/advancedcluster"
	"github.com/mongodb/terraform-provider-mongodbatlas/internal/service/advancedclustertpf"
)

const (
	errorClusterCreate            = "error creating MongoDB Cluster: %s"
	errorClusterRead              = "error reading MongoDB Cluster (%s): %s"
	errorClusterDelete            = "error deleting MongoDB Cluster (%s): %s"
	errorClusterUpdate            = "error updating MongoDB Cluster (%s): %s"
	errorAdvancedConfUpdate       = "error updating Advanced Configuration Option %s for MongoDB Cluster (%s): %s"
	ErrorSnapshotBackupPolicyRead = "error getting a Cloud Provider Snapshot Backup Policy for the cluster(%s): %s"
)

var defaultLabel = matlas.Label{Key: advancedclustertpf.IgnoreLabelKey, Value: "MongoDB Atlas Terraform Provider"}

func Resource() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceCreate,
		ReadWithoutTimeout:   resourceRead,
		UpdateWithoutTimeout: resourceUpdate,
		DeleteWithoutTimeout: resourceDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceImport,
		},
		SchemaVersion: 1,
		StateUpgraders: []schema.StateUpgrader{
			{
				Type:    ResourceClusterResourceV0().CoreConfigSchema().ImpliedType(),
				Upgrade: resourceMongoDBAtlasClusterStateUpgradeV0,
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
			"auto_scaling_disk_gb_enabled": {
				Type:     schema.TypeBool,
				Computed: true,
				Optional: true,
			},
			"auto_scaling_compute_enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
			},
			"auto_scaling_compute_scale_down_enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
			},
			"backup_enabled": {
				Type:        schema.TypeBool,
				Optional:    true,
				Computed:    true,
				Description: "Clusters running MongoDB FCV 4.2 or later and any new Atlas clusters of any type do not support this parameter",
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
				Optional: true,
				Computed: true,
			},
			"connection_strings": advancedcluster.SchemaConnectionStrings(),
			"disk_size_gb": {
				Type:     schema.TypeFloat,
				Optional: true,
				Computed: true,
			},
			"encryption_at_rest_provider": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"mongo_db_major_version": {
				Type:      schema.TypeString,
				Optional:  true,
				Computed:  true,
				StateFunc: advancedcluster.FormatMongoDBMajorVersion,
			},
			"num_shards": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
			},
			"cloud_backup": {
				Type:          schema.TypeBool,
				Optional:      true,
				Computed:      true,
				ConflictsWith: []string{"backup_enabled"},
			},
			"provider_instance_size_name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"provider_name": {
				Type:             schema.TypeString,
				Required:         true,
				ValidateDiagFunc: validate.StringIsUppercase(),
			},
			"pit_enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
			},
			"backing_provider_name": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"provider_disk_iops": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
			},
			"provider_disk_type_name": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"provider_encrypt_ebs_volume": {
				Type:       schema.TypeBool,
				Optional:   true,
				Deprecated: "All EBS volumes are encrypted by default, the option to disable encryption has been removed",
				Computed:   true,
			},
			"provider_encrypt_ebs_volume_flag": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"provider_region_name": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"provider_volume_type": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"provider_auto_scaling_compute_min_instance_size": {
				Type:             schema.TypeString,
				Optional:         true,
				Computed:         true,
				DiffSuppressFunc: isEqualProviderAutoScalingMinInstanceSize,
			},
			"provider_auto_scaling_compute_max_instance_size": {
				Type:             schema.TypeString,
				Optional:         true,
				Computed:         true,
				DiffSuppressFunc: isEqualProviderAutoScalingMaxInstanceSize,
			},
			"replication_factor": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
			},
			"replication_specs": {
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
						},
						"num_shards": {
							Type:     schema.TypeInt,
							Required: true,
						},
						"regions_config": {
							Type:     schema.TypeSet,
							Optional: true,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"region_name": {
										Type:             schema.TypeString,
										Required:         true,
										ValidateDiagFunc: validate.StringIsUppercase(),
									},
									"electable_nodes": {
										Type:     schema.TypeInt,
										Optional: true,
										Computed: true,
									},
									"priority": {
										Type:     schema.TypeInt,
										Optional: true,
										Computed: true,
									},
									"read_only_nodes": {
										Type:     schema.TypeInt,
										Optional: true,
										Default:  0,
									},
									"analytics_nodes": {
										Type:     schema.TypeInt,
										Optional: true,
										Default:  0,
									},
								},
							},
						},
						"zone_name": {
							Type:     schema.TypeString,
							Optional: true,
							Default:  "ZoneName managed by Terraform",
						},
					},
				},
				Set: func(v any) int {
					var buf bytes.Buffer
					m := v.(map[string]any)
					buf.WriteString(fmt.Sprintf("%d", m["num_shards"].(int)))
					buf.WriteString(m["zone_name"].(string))
					buf.WriteString(fmt.Sprintf("%+v", m["regions_config"].(*schema.Set)))
					return advancedcluster.HashCodeString(buf.String())
				},
			},
			"mongo_db_version": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"mongo_uri": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"mongo_uri_updated": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"mongo_uri_with_options": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"paused": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
			},
			"srv_address": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"state_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"advanced_configuration": advancedcluster.SchemaAdvancedConfig(),
			"labels": {
				Type:     schema.TypeSet,
				Optional: true,
				Set:      advancedcluster.HashFunctionForKeyValuePair,
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
			"tags":                   &advancedcluster.RSTagsSchema,
			"snapshot_backup_policy": computedCloudProviderSnapshotBackupPolicySchema(),
			"termination_protection_enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
			},
			"container_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"version_release_system": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringInSlice([]string{"LTS", "CONTINUOUS"}, false),
			},
			"accept_data_risks_and_force_replica_set_reconfig": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "Submit this field alongside your topology reconfiguration to request a new regional outage resistant topology",
			},
			"redact_client_log_data": {
				Type:     schema.TypeBool,
				Optional: true,
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
		CustomizeDiff: resourceClusterCustomizeDiff,
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(3 * time.Hour),
			Update: schema.DefaultTimeout(3 * time.Hour),
			Delete: schema.DefaultTimeout(3 * time.Hour),
		},
	}
}

func resourceCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	if v, ok := d.GetOk("accept_data_risks_and_force_replica_set_reconfig"); ok {
		if v.(string) != "" {
			return diag.FromErr(fmt.Errorf("accept_data_risks_and_force_replica_set_reconfig can not be set in creation, only in update"))
		}
	}

	var (
		conn             = meta.(*config.MongoDBClient).Atlas
		connV2           = meta.(*config.MongoDBClient).AtlasV2
		connV220240530   = meta.(*config.MongoDBClient).AtlasV220240530
		projectID        = d.Get("project_id").(string)
		clusterName      = d.Get("name").(string)
		providerName     = d.Get("provider_name").(string)
		computeEnabled   = d.Get("auto_scaling_compute_enabled").(bool)
		scaleDownEnabled = d.Get("auto_scaling_compute_scale_down_enabled").(bool)
		minInstanceSize  = d.Get("provider_auto_scaling_compute_min_instance_size").(string)
		maxInstanceSize  = d.Get("provider_auto_scaling_compute_max_instance_size").(string)
	)

	if scaleDownEnabled && !computeEnabled {
		return diag.FromErr(fmt.Errorf("`auto_scaling_compute_scale_down_enabled` must be set when `auto_scaling_compute_enabled` is set"))
	}

	if computeEnabled && maxInstanceSize == "" {
		return diag.FromErr(fmt.Errorf("`provider_auto_scaling_compute_max_instance_size` must be set when `auto_scaling_compute_enabled` is set"))
	}

	if scaleDownEnabled && minInstanceSize == "" {
		return diag.FromErr(fmt.Errorf("`provider_auto_scaling_compute_min_instance_size` must be set when `auto_scaling_compute_scale_down_enabled` is set"))
	}

	autoScaling := &matlas.AutoScaling{
		DiskGBEnabled: conversion.Pointer(d.Get("auto_scaling_disk_gb_enabled").(bool)),
		Compute: &matlas.Compute{
			Enabled:          &computeEnabled,
			ScaleDownEnabled: &scaleDownEnabled,
		},
	}

	// validate cluster_type conditional
	if _, ok := d.GetOk("replication_specs"); ok {
		if _, ok1 := d.GetOk("cluster_type"); !ok1 {
			return diag.FromErr(fmt.Errorf("`cluster_type` should be set when `replication_specs` is set"))
		}
	}

	if providerName != "AWS" {
		if _, ok := d.GetOk("provider_disk_iops"); ok {
			return diag.Errorf("`provider_disk_iops` shouldn't be set when provider name is `GCP` or `AZURE`")
		}

		if _, ok := d.GetOk("provider_volume_type"); ok {
			return diag.FromErr(fmt.Errorf("`provider_volume_type` shouldn't be set when provider name is `GCP` or `AZURE`"))
		}
	}

	if providerName != "AZURE" {
		if _, ok := d.GetOk("provider_disk_type_name"); ok {
			return diag.FromErr(fmt.Errorf("`provider_disk_type_name` shouldn't be set when provider name is `GCP` or `AWS`"))
		}
	}

	if providerName == "AZURE" {
		if _, ok := d.GetOk("disk_size_gb"); ok {
			return diag.FromErr(fmt.Errorf("`disk_size_gb` cannot be used with Azure clusters"))
		}
	}

	tenantDisksize := conversion.Pointer[float64](0.0)
	if providerName == "TENANT" {
		autoScaling = nil

		if instanceSizeName, ok := d.GetOk("provider_instance_size_name"); ok {
			if instanceSizeName == "M2" {
				if diskSizeGB, ok := d.GetOk("disk_size_gb"); ok {
					if cast.ToFloat64(diskSizeGB) != 2 {
						return diag.FromErr(fmt.Errorf("`disk_size_gb` must be 2 for M2 shared tier"))
					}
				}
			}
			if instanceSizeName == "M5" {
				if diskSizeGB, ok := d.GetOk("disk_size_gb"); ok {
					if cast.ToFloat64(diskSizeGB) != 5 {
						return diag.FromErr(fmt.Errorf("`disk_size_gb` must be 5 for M5 shared tier"))
					}
				}
			}
		}
	}

	// We need to validate the oplog_size_mb attr of the advanced configuration option to show the error
	// before that the cluster is created
	if oplogSizeMB, ok := d.GetOkExists("advanced_configuration.0.oplog_size_mb"); ok {
		if cast.ToInt64(oplogSizeMB) <= 0 {
			return diag.FromErr(fmt.Errorf("`advanced_configuration.oplog_size_mb` cannot be <= 0"))
		}
	}

	providerSettings, err := expandProviderSetting(d)
	if err != nil {
		return diag.FromErr(fmt.Errorf(errorClusterCreate, err))
	}

	replicationSpecs, err := expandReplicationSpecs(d)
	if err != nil {
		return diag.FromErr(fmt.Errorf(errorClusterCreate, err))
	}

	clusterType := cast.ToString(d.Get("cluster_type"))
	err = ValidateProviderRegionName(clusterType, providerSettings.RegionName, replicationSpecs)
	if err != nil {
		return diag.FromErr(fmt.Errorf(errorClusterCreate, err))
	}

	clusterRequest := &matlas.Cluster{
		Name:                     clusterName,
		EncryptionAtRestProvider: d.Get("encryption_at_rest_provider").(string),
		ClusterType:              clusterType,
		BackupEnabled:            conversion.Pointer(d.Get("backup_enabled").(bool)),
		PitEnabled:               conversion.Pointer(d.Get("pit_enabled").(bool)),
		AutoScaling:              autoScaling,
		ProviderSettings:         providerSettings,
		ReplicationSpecs:         replicationSpecs,
	}
	if v, ok := d.GetOk("cloud_backup"); ok {
		clusterRequest.ProviderBackupEnabled = conversion.Pointer(v.(bool))
	}

	if _, ok := d.GetOk("bi_connector_config"); ok {
		biConnector, err := expandBiConnectorConfig(d)
		if err != nil {
			return diag.FromErr(fmt.Errorf(errorClusterCreate, err))
		}
		clusterRequest.BiConnector = biConnector
	}

	if containsLabelOrKey(expandLabelSliceFromSetSchema(d), defaultLabel) {
		return diag.FromErr(advancedclustertpf.ErrIgnoreLabel)
	}

	clusterRequest.Labels = append(expandLabelSliceFromSetSchema(d), defaultLabel)

	if _, ok := d.GetOk("tags"); ok {
		tagsSlice := expandTagSliceFromSetSchema(d)
		clusterRequest.Tags = &tagsSlice
	}

	if v, ok := d.GetOk("disk_size_gb"); ok {
		clusterRequest.DiskSizeGB = conversion.Pointer(v.(float64))
	}
	if cast.ToFloat64(tenantDisksize) != 0 {
		clusterRequest.DiskSizeGB = tenantDisksize
	}
	if v, ok := d.GetOk("mongo_db_major_version"); ok {
		clusterRequest.MongoDBMajorVersion = advancedcluster.FormatMongoDBMajorVersion(v.(string))
	}

	if r, ok := d.GetOk("replication_factor"); ok {
		clusterRequest.ReplicationFactor = conversion.Pointer(cast.ToInt64(r))
	}

	if n, ok := d.GetOk("num_shards"); ok {
		clusterRequest.NumShards = conversion.Pointer(cast.ToInt64(n))
	}

	if v, ok := d.GetOk("termination_protection_enabled"); ok {
		clusterRequest.TerminationProtectionEnabled = conversion.Pointer(v.(bool))
	}

	if v, ok := d.GetOk("version_release_system"); ok {
		clusterRequest.VersionReleaseSystem = v.(string)
	}

	cluster, _, err := conn.Clusters.Create(ctx, projectID, clusterRequest)
	if err != nil {
		return diag.FromErr(fmt.Errorf(errorClusterCreate, err))
	}

	timeout := d.Timeout(schema.TimeoutCreate)
	stateConf := advancedcluster.CreateStateChangeConfig(ctx, connV2, projectID, clusterName, timeout)
	if _, err = stateConf.WaitForStateContext(ctx); err != nil {
		return diag.FromErr(fmt.Errorf(errorClusterCreate, err))
	}

	/*
		So far, the cluster has created correctly, so we need to set up
		the advanced configuration option to attach it
	*/
	ac, ok := d.GetOk("advanced_configuration")
	if aclist, ok1 := ac.([]any); ok1 && len(aclist) > 0 {
		params20240530, params := expandProcessArgs(d, aclist[0].(map[string]any), &clusterRequest.MongoDBMajorVersion)

		if ok {
			_, _, err = connV220240530.ClustersApi.UpdateClusterAdvancedConfiguration(ctx, projectID, cluster.Name, &params20240530).Execute()
			if err != nil {
				return diag.FromErr(fmt.Errorf(errorAdvancedConfUpdate, advancedcluster.V20240530, cluster.Name, err))
			}
			_, _, err = connV2.ClustersApi.UpdateClusterAdvancedConfiguration(ctx, projectID, cluster.Name, &params).Execute()
			if err != nil {
				return diag.FromErr(fmt.Errorf(errorAdvancedConfUpdate, "", cluster.Name, err))
			}
		}
	}

	// To pause a cluster
	if v := d.Get("paused").(bool); v {
		clusterRequest = &matlas.Cluster{
			Paused: conversion.Pointer(v),
		}
		_, _, err = updateCluster(ctx, conn, connV2, clusterRequest, projectID, clusterName, timeout)
		if err != nil {
			return diag.FromErr(fmt.Errorf(errorClusterUpdate, clusterName, err))
		}
	}

	if v, ok := d.GetOk("redact_client_log_data"); ok {
		if err := newAtlasUpdate(ctx, d.Timeout(schema.TimeoutCreate), connV2, projectID, clusterName, v.(bool)); err != nil {
			return diag.FromErr(fmt.Errorf(errorClusterUpdate, clusterName, err))
		}
	}

	if pinnedFCVBlock, _ := d.Get("pinned_fcv").([]any); len(pinnedFCVBlock) > 0 {
		nestedObj := pinnedFCVBlock[0].(map[string]any)
		expDateStr := cast.ToString(nestedObj["expiration_date"])
		if err := advancedclustertpf.PinFCV(ctx, connV2.ClustersApi, projectID, clusterName, expDateStr); err != nil {
			return diag.FromErr(fmt.Errorf(errorClusterUpdate, clusterName, err))
		}
		stateConf := advancedcluster.CreateStateChangeConfig(ctx, connV2, projectID, clusterName, timeout)
		if _, err = stateConf.WaitForStateContext(ctx); err != nil {
			return diag.FromErr(fmt.Errorf(errorClusterUpdate, clusterName, err))
		}
	}

	d.SetId(conversion.EncodeStateID(map[string]string{
		"cluster_id":    cluster.ID,
		"project_id":    projectID,
		"cluster_name":  clusterName,
		"provider_name": providerName,
	}))

	return resourceRead(ctx, d, meta)
}

func resourceRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*config.MongoDBClient).Atlas
	connV2 := meta.(*config.MongoDBClient).AtlasV2
	connV220240530 := meta.(*config.MongoDBClient).AtlasV220240530
	ids := conversion.DecodeStateID(d.Id())
	projectID := ids["project_id"]
	clusterName := ids["cluster_name"]
	providerName := ids["provider_name"]

	cluster, resp, err := conn.Clusters.Get(ctx, projectID, clusterName)
	if err != nil {
		if resp != nil && resp.StatusCode == http.StatusNotFound {
			d.SetId("")
			return nil
		}

		return diag.FromErr(fmt.Errorf(errorClusterRead, clusterName, err))
	}

	log.Printf("[DEBUG] GET Cluster %+v", cluster)

	if err := d.Set("cluster_id", cluster.ID); err != nil {
		return diag.FromErr(fmt.Errorf(advancedcluster.ErrorClusterSetting, "cluster_id", clusterName, err))
	}

	if err := d.Set("auto_scaling_compute_enabled", cluster.AutoScaling.Compute.Enabled); err != nil {
		return diag.FromErr(fmt.Errorf(advancedcluster.ErrorClusterSetting, "auto_scaling_compute_enabled", clusterName, err))
	}

	if err := d.Set("auto_scaling_compute_scale_down_enabled", cluster.AutoScaling.Compute.ScaleDownEnabled); err != nil {
		return diag.FromErr(fmt.Errorf(advancedcluster.ErrorClusterSetting, "auto_scaling_compute_scale_down_enabled", clusterName, err))
	}

	if err := d.Set("provider_auto_scaling_compute_min_instance_size", cluster.ProviderSettings.AutoScaling.Compute.MinInstanceSize); err != nil {
		return diag.FromErr(fmt.Errorf(advancedcluster.ErrorClusterSetting, "provider_auto_scaling_compute_min_instance_size", clusterName, err))
	}

	if err := d.Set("provider_auto_scaling_compute_max_instance_size", cluster.ProviderSettings.AutoScaling.Compute.MaxInstanceSize); err != nil {
		return diag.FromErr(fmt.Errorf(advancedcluster.ErrorClusterSetting, "provider_auto_scaling_compute_max_instance_size", clusterName, err))
	}

	if err := d.Set("backup_enabled", cluster.BackupEnabled); err != nil {
		return diag.FromErr(fmt.Errorf(advancedcluster.ErrorClusterSetting, "backup_enabled", clusterName, err))
	}

	if _, ok := d.GetOk("cloud_backup"); ok {
		if err := d.Set("cloud_backup", cluster.ProviderBackupEnabled); err != nil {
			return diag.FromErr(fmt.Errorf(advancedcluster.ErrorClusterSetting, "cloud_backup", clusterName, err))
		}
	}

	if err := d.Set("cluster_type", cluster.ClusterType); err != nil {
		return diag.FromErr(fmt.Errorf(advancedcluster.ErrorClusterSetting, "cluster_type", clusterName, err))
	}

	if err := d.Set("connection_strings", flattenConnectionStrings(cluster.ConnectionStrings)); err != nil {
		return diag.FromErr(fmt.Errorf(advancedcluster.ErrorClusterSetting, "connection_strings", clusterName, err))
	}

	if err := d.Set("disk_size_gb", cluster.DiskSizeGB); err != nil {
		return diag.FromErr(fmt.Errorf(advancedcluster.ErrorClusterSetting, "disk_size_gb", clusterName, err))
	}

	if err := d.Set("encryption_at_rest_provider", cluster.EncryptionAtRestProvider); err != nil {
		return diag.FromErr(fmt.Errorf(advancedcluster.ErrorClusterSetting, "encryption_at_rest_provider", clusterName, err))
	}

	// Avoid Global Cluster issues. (NumShards is not present in Global Clusters)
	if cluster.NumShards != nil {
		if err := d.Set("num_shards", cluster.NumShards); err != nil {
			return diag.FromErr(fmt.Errorf(advancedcluster.ErrorClusterSetting, "num_shards", clusterName, err))
		}
	}

	if err := d.Set("mongo_db_version", cluster.MongoDBVersion); err != nil {
		return diag.FromErr(fmt.Errorf(advancedcluster.ErrorClusterSetting, "mongo_db_version", clusterName, err))
	}

	if err := d.Set("mongo_uri", cluster.MongoURI); err != nil {
		return diag.FromErr(fmt.Errorf(advancedcluster.ErrorClusterSetting, "mongo_uri", clusterName, err))
	}

	if err := d.Set("mongo_uri_updated", cluster.MongoURIUpdated); err != nil {
		return diag.FromErr(fmt.Errorf(advancedcluster.ErrorClusterSetting, "mongo_uri_updated", clusterName, err))
	}

	if err := d.Set("mongo_uri_with_options", cluster.MongoURIWithOptions); err != nil {
		return diag.FromErr(fmt.Errorf(advancedcluster.ErrorClusterSetting, "mongo_uri_with_options", clusterName, err))
	}

	if err := d.Set("pit_enabled", cluster.PitEnabled); err != nil {
		return diag.FromErr(fmt.Errorf(advancedcluster.ErrorClusterSetting, "pit_enabled", clusterName, err))
	}

	if err := d.Set("paused", cluster.Paused); err != nil {
		return diag.FromErr(fmt.Errorf(advancedcluster.ErrorClusterSetting, "paused", clusterName, err))
	}

	if err := d.Set("srv_address", cluster.SrvAddress); err != nil {
		return diag.FromErr(fmt.Errorf(advancedcluster.ErrorClusterSetting, "srv_address", clusterName, err))
	}

	if err := d.Set("state_name", cluster.StateName); err != nil {
		return diag.FromErr(fmt.Errorf(advancedcluster.ErrorClusterSetting, "state_name", clusterName, err))
	}

	if err := d.Set("termination_protection_enabled", cluster.TerminationProtectionEnabled); err != nil {
		return diag.FromErr(fmt.Errorf(advancedcluster.ErrorClusterSetting, "termination_protection_enabled", clusterName, err))
	}

	if err := d.Set("bi_connector_config", flattenBiConnectorConfig(cluster.BiConnector)); err != nil {
		return diag.FromErr(fmt.Errorf(advancedcluster.ErrorClusterSetting, "bi_connector_config", clusterName, err))
	}

	if cluster.ProviderSettings != nil {
		flattenProviderSettings(d, cluster.ProviderSettings, clusterName)
	}

	if err := d.Set("replication_specs", flattenReplicationSpecs(cluster.ReplicationSpecs)); err != nil {
		return diag.FromErr(fmt.Errorf(advancedcluster.ErrorClusterSetting, "replication_specs", clusterName, err))
	}

	if err := d.Set("replication_factor", cluster.ReplicationFactor); err != nil {
		return diag.FromErr(fmt.Errorf(advancedcluster.ErrorClusterSetting, "replication_factor", clusterName, err))
	}

	if err := d.Set("labels", flattenLabels(removeLabel(cluster.Labels, defaultLabel))); err != nil {
		return diag.FromErr(fmt.Errorf(advancedcluster.ErrorClusterSetting, "labels", clusterName, err))
	}

	if err := d.Set("tags", flattenTags(cluster.Tags)); err != nil {
		return diag.FromErr(fmt.Errorf(advancedcluster.ErrorClusterSetting, "tags", clusterName, err))
	}

	if err := d.Set("version_release_system", cluster.VersionReleaseSystem); err != nil {
		return diag.FromErr(fmt.Errorf(advancedcluster.ErrorClusterSetting, "version_release_system", clusterName, err))
	}

	if err := d.Set("accept_data_risks_and_force_replica_set_reconfig", cluster.AcceptDataRisksAndForceReplicaSetReconfig); err != nil {
		return diag.FromErr(fmt.Errorf(advancedcluster.ErrorClusterSetting, "accept_data_risks_and_force_replica_set_reconfig", clusterName, err))
	}

	if providerName != "TENANT" {
		containers, _, err := conn.Containers.List(ctx, projectID,
			&matlas.ContainersListOptions{ProviderName: providerName})
		if err != nil {
			return diag.FromErr(fmt.Errorf(errorClusterRead, clusterName, err))
		}

		if err := d.Set("container_id", getContainerID(containers, cluster)); err != nil {
			return diag.FromErr(fmt.Errorf(advancedcluster.ErrorClusterSetting, "container_id", clusterName, err))
		}

		if err := d.Set("auto_scaling_disk_gb_enabled", cluster.AutoScaling.DiskGBEnabled); err != nil {
			return diag.FromErr(fmt.Errorf(advancedcluster.ErrorClusterSetting, "auto_scaling_disk_gb_enabled", clusterName, err))
		}
	}

	/*
		Get the advaced configuration options and set up to the terraform state
	*/
	processArgs20240530, _, err := connV220240530.ClustersApi.GetClusterAdvancedConfiguration(ctx, projectID, clusterName).Execute()
	if err != nil {
		return diag.FromErr(fmt.Errorf(advancedcluster.ErrorAdvancedConfRead, advancedcluster.V20240530, clusterName, err))
	}
	processArgs, _, err := connV2.ClustersApi.GetClusterAdvancedConfiguration(ctx, projectID, clusterName).Execute()
	if err != nil {
		return diag.FromErr(fmt.Errorf(advancedcluster.ErrorAdvancedConfRead, "", clusterName, err))
	}

	if err := d.Set("advanced_configuration", flattenProcessArgs(processArgs20240530, processArgs)); err != nil {
		return diag.FromErr(fmt.Errorf(advancedcluster.ErrorClusterSetting, "advanced_configuration", clusterName, err))
	}

	// Get the snapshot policy and set the data
	snapshotBackupPolicy, err := flattenCloudProviderSnapshotBackupPolicy(ctx, d, conn, projectID, clusterName)
	if err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("snapshot_backup_policy", snapshotBackupPolicy); err != nil {
		return diag.FromErr(err)
	}

	latestClusterModel, err := newAtlasGet(ctx, connV2, projectID, clusterName)
	if err != nil {
		return diag.FromErr(fmt.Errorf(errorClusterRead, clusterName, err))
	}
	if err := d.Set("redact_client_log_data", latestClusterModel.GetRedactClientLogData()); err != nil {
		return diag.FromErr(fmt.Errorf(advancedcluster.ErrorClusterSetting, "redact_client_log_data", clusterName, err))
	}

	if err := d.Set("mongo_db_major_version", latestClusterModel.MongoDBMajorVersion); err != nil { // uses 2024-08-05 or above as it has fix for correct value when FCV is active
		return diag.FromErr(fmt.Errorf(advancedcluster.ErrorClusterSetting, "mongo_db_major_version", clusterName, err))
	}

	warning := advancedcluster.WarningIfFCVExpiredOrUnpinnedExternally(d, latestClusterModel) // has to be called before pinned_fcv value is updated in ResourceData to know prior state value

	if err := d.Set("pinned_fcv", advancedcluster.FlattenPinnedFCV(latestClusterModel)); err != nil {
		return diag.FromErr(fmt.Errorf(advancedcluster.ErrorClusterSetting, "pinned_fcv", clusterName, err))
	}

	return warning
}

func resourceUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var (
		conn                = meta.(*config.MongoDBClient).Atlas
		connV2              = meta.(*config.MongoDBClient).AtlasV2
		connV220240530      = meta.(*config.MongoDBClient).AtlasV220240530
		ids                 = conversion.DecodeStateID(d.Id())
		projectID           = ids["project_id"]
		clusterName         = ids["cluster_name"]
		timeout             = d.Timeout(schema.TimeoutUpdate)
		cluster             = new(matlas.Cluster)
		clusterChangeDetect = &matlas.Cluster{
			AutoScaling: &matlas.AutoScaling{
				Compute: &matlas.Compute{},
			},
		}
	)

	// FCV update is intentionally handled before other cluster updates, and will wait for cluster to reach IDLE state before continuing
	if diags := advancedcluster.HandlePinnedFCVUpdate(ctx, connV2, projectID, clusterName, d, timeout); diags != nil {
		return diags
	}

	if d.HasChange("name") {
		cluster.Name, _ = d.Get("name").(string)
	}

	if d.HasChange("bi_connector_config") {
		cluster.BiConnector, _ = expandBiConnectorConfig(d)
	}

	// If at least one of the provider settings argument has changed, expand all provider settings
	if d.HasChange("provider_disk_iops") ||
		d.HasChange("backing_provider_name") ||
		d.HasChange("provider_disk_type_name") ||
		d.HasChange("provider_instance_size_name") ||
		d.HasChange("provider_name") ||
		d.HasChange("provider_region_name") ||
		d.HasChange("provider_volume_type") ||
		d.HasChange("provider_auto_scaling_compute_min_instance_size") ||
		d.HasChange("provider_auto_scaling_compute_max_instance_size") {
		var err error
		cluster.ProviderSettings, err = expandProviderSetting(d)
		if err != nil {
			return diag.FromErr(fmt.Errorf(errorClusterUpdate, clusterName, err))
		}
	}

	replicationSpecs, err := expandReplicationSpecs(d)
	if err != nil {
		return diag.FromErr(fmt.Errorf(errorClusterUpdate, clusterName, err))
	}
	if d.HasChange("replication_specs") {
		cluster.ReplicationSpecs = replicationSpecs
	}

	if v, ok := d.GetOk("provider_region_name"); ok {
		err = ValidateProviderRegionName(d.Get("cluster_type").(string), v.(string), replicationSpecs)
		// we swallow the error here as the user may not always be able to 'unset' provider_region_name value in the state,
		// We then ensure ProviderSettings.RegionName is not set in case of a multi-region cluster, refer https://jira.mongodb.org/browse/HELP-51429
		if err != nil {
			if cluster.ProviderSettings != nil {
				cluster.ProviderSettings.RegionName = ""
			}
		}
	}

	cluster.AutoScaling = &matlas.AutoScaling{Compute: &matlas.Compute{}}

	if d.HasChange("auto_scaling_disk_gb_enabled") {
		cluster.AutoScaling.DiskGBEnabled = conversion.Pointer(d.Get("auto_scaling_disk_gb_enabled").(bool))
	}

	if d.HasChange("auto_scaling_compute_enabled") {
		cluster.AutoScaling.Compute.Enabled = conversion.Pointer(d.Get("auto_scaling_compute_enabled").(bool))
	}

	if d.HasChange("auto_scaling_compute_scale_down_enabled") {
		cluster.AutoScaling.Compute.ScaleDownEnabled = conversion.Pointer(d.Get("auto_scaling_compute_scale_down_enabled").(bool))
	}

	if d.HasChange("encryption_at_rest_provider") {
		cluster.EncryptionAtRestProvider = d.Get("encryption_at_rest_provider").(string)
	}

	if d.HasChange("mongo_db_major_version") {
		cluster.MongoDBMajorVersion = advancedcluster.FormatMongoDBMajorVersion(d.Get("mongo_db_major_version"))
	}

	if d.HasChange("cluster_type") {
		cluster.ClusterType = d.Get("cluster_type").(string)
	}

	if d.HasChange("backup_enabled") {
		cluster.BackupEnabled = conversion.Pointer(d.Get("backup_enabled").(bool))
	}

	if d.HasChange("disk_size_gb") {
		cluster.DiskSizeGB = conversion.Pointer(d.Get("disk_size_gb").(float64))
	}

	if d.HasChange("cloud_backup") {
		cluster.ProviderBackupEnabled = conversion.Pointer(d.Get("cloud_backup").(bool))
	}

	if d.HasChange("pit_enabled") {
		cluster.PitEnabled = conversion.Pointer(d.Get("pit_enabled").(bool))
	}

	if d.HasChange("replication_factor") {
		cluster.ReplicationFactor = conversion.Pointer(cast.ToInt64(d.Get("replication_factor")))
	}

	if d.HasChange("num_shards") {
		cluster.NumShards = conversion.Pointer(cast.ToInt64(d.Get("num_shards")))
	}

	if d.HasChange("version_release_system") {
		cluster.VersionReleaseSystem = d.Get("version_release_system").(string)
	}

	if d.HasChange("accept_data_risks_and_force_replica_set_reconfig") {
		cluster.AcceptDataRisksAndForceReplicaSetReconfig = d.Get("accept_data_risks_and_force_replica_set_reconfig").(string)
	}

	if d.HasChange("termination_protection_enabled") {
		cluster.TerminationProtectionEnabled = conversion.Pointer(d.Get("termination_protection_enabled").(bool))
	}

	if d.HasChange("labels") {
		if containsLabelOrKey(expandLabelSliceFromSetSchema(d), defaultLabel) {
			return diag.FromErr(advancedclustertpf.ErrIgnoreLabel)
		}

		cluster.Labels = append(expandLabelSliceFromSetSchema(d), defaultLabel)
	}

	if d.HasChange("tags") {
		tagsSlice := expandTagSliceFromSetSchema(d)
		cluster.Tags = &tagsSlice
	}

	// when Provider instance type changes this argument must be passed explicitly in patch request
	if d.HasChange("provider_instance_size_name") {
		if _, ok := d.GetOk("cloud_backup"); ok {
			cluster.ProviderBackupEnabled = conversion.Pointer(d.Get("cloud_backup").(bool))
		}
	}

	if d.HasChange("paused") && !d.Get("paused").(bool) {
		cluster.Paused = conversion.Pointer(d.Get("paused").(bool))
	}

	/*
		Check if advaced configuration option has a changes to update it
	*/
	if d.HasChange("advanced_configuration") {
		var mongoDBMajorVersion string
		if v, ok := d.GetOk("mongo_db_major_version"); ok {
			mongoDBMajorVersion = v.(string)
		}

		ac := d.Get("advanced_configuration")
		if aclist, ok1 := ac.([]any); ok1 && len(aclist) > 0 {
			params20240530, params := expandProcessArgs(d, aclist[0].(map[string]any), &mongoDBMajorVersion)
			if !reflect.DeepEqual(params20240530, admin20240530.ClusterDescriptionProcessArgs{}) {
				_, _, err := connV220240530.ClustersApi.UpdateClusterAdvancedConfiguration(ctx, projectID, clusterName, &params20240530).Execute()
				if err != nil {
					return diag.FromErr(fmt.Errorf(errorAdvancedConfUpdate, advancedcluster.V20240530, clusterName, err))
				}
			}
			if !reflect.DeepEqual(params, admin.ClusterDescriptionProcessArgs20240805{}) {
				_, _, err = connV2.ClustersApi.UpdateClusterAdvancedConfiguration(ctx, projectID, clusterName, &params).Execute()
				if err != nil {
					return diag.FromErr(fmt.Errorf(errorAdvancedConfUpdate, "", clusterName, err))
				}
			}
		}
	}

	if isUpgradeRequired(d) {
		updatedCluster, _, err := upgradeCluster(ctx, conn, connV2, cluster, projectID, clusterName, timeout)

		if err != nil {
			return diag.FromErr(fmt.Errorf(errorClusterUpdate, clusterName, err))
		}

		d.SetId(conversion.EncodeStateID(map[string]string{
			"cluster_id":    updatedCluster.ID,
			"project_id":    projectID,
			"cluster_name":  updatedCluster.Name,
			"provider_name": updatedCluster.ProviderSettings.ProviderName,
		}))
	} else if !reflect.DeepEqual(cluster, clusterChangeDetect) {
		err := retry.RetryContext(ctx, timeout, func() *retry.RetryError {
			_, _, err := updateCluster(ctx, conn, connV2, cluster, projectID, clusterName, timeout)

			if didErrOnPausedCluster(err) {
				clusterRequest := &matlas.Cluster{
					Paused: conversion.Pointer(false),
				}

				_, _, err = updateCluster(ctx, conn, connV2, clusterRequest, projectID, clusterName, timeout)
			}

			if err != nil {
				return retry.NonRetryableError(fmt.Errorf(errorClusterUpdate, clusterName, err))
			}

			return nil
		})

		if err != nil {
			return diag.FromErr(fmt.Errorf(errorClusterUpdate, clusterName, err))
		}
	}

	if d.Get("paused").(bool) && !advancedcluster.IsSharedTier(d.Get("provider_instance_size_name").(string)) {
		clusterRequest := &matlas.Cluster{
			Paused: conversion.Pointer(true),
		}

		_, _, err := updateCluster(ctx, conn, connV2, clusterRequest, projectID, clusterName, timeout)
		if err != nil {
			return diag.FromErr(fmt.Errorf(errorClusterUpdate, clusterName, err))
		}
	}

	if d.HasChange("redact_client_log_data") {
		redactClientLogData := d.Get("redact_client_log_data").(bool)
		if err := newAtlasUpdate(ctx, d.Timeout(schema.TimeoutUpdate), connV2, projectID, clusterName, redactClientLogData); err != nil {
			return diag.FromErr(fmt.Errorf(errorClusterUpdate, clusterName, err))
		}
	}

	return resourceRead(ctx, d, meta)
}

func IsMultiRegionCluster(repSpecs []matlas.ReplicationSpec) bool {
	if len(repSpecs) > 1 {
		return true
	}

	for i := range repSpecs {
		if len(repSpecs[i].RegionsConfig) > 1 {
			return true
		}
	}
	return false
}

func ValidateProviderRegionName(clusterType, providerRegionName string, repSpecs []matlas.ReplicationSpec) error {
	if conversion.IsStringPresent(&providerRegionName) && (clusterType == "GEOSHARDED" || IsMultiRegionCluster(repSpecs)) {
		return fmt.Errorf("provider_region_name attribute must be set ONLY for single-region clusters")
	}

	return nil
}

func didErrOnPausedCluster(err error) bool {
	if err == nil {
		return false
	}

	var target *matlas.ErrorResponse

	return errors.As(err, &target) && target.ErrorCode == "CANNOT_UPDATE_PAUSED_CLUSTER"
}

func resourceDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*config.MongoDBClient).Atlas
	connV2 := meta.(*config.MongoDBClient).AtlasV2
	ids := conversion.DecodeStateID(d.Id())
	projectID := ids["project_id"]
	clusterName := ids["cluster_name"]

	var options *matlas.DeleteAdvanceClusterOptions
	if v, ok := d.GetOkExists("retain_backups_enabled"); ok {
		options = &matlas.DeleteAdvanceClusterOptions{
			RetainBackups: conversion.Pointer(v.(bool)),
		}
	}

	_, err := conn.Clusters.Delete(ctx, projectID, clusterName, options)
	if err != nil {
		return diag.FromErr(fmt.Errorf(errorClusterDelete, clusterName, err))
	}

	stateConf := advancedcluster.DeleteStateChangeConfig(ctx, connV2, projectID, clusterName, d.Timeout(schema.TimeoutDelete))
	if _, err = stateConf.WaitForStateContext(ctx); err != nil {
		return diag.FromErr(fmt.Errorf(errorClusterDelete, clusterName, err))
	}

	return nil
}

func resourceImport(ctx context.Context, d *schema.ResourceData, meta any) ([]*schema.ResourceData, error) {
	conn := meta.(*config.MongoDBClient).Atlas

	projectID, name, err := splitSClusterImportID(d.Id())
	if err != nil {
		return nil, err
	}

	u, _, err := conn.Clusters.Get(ctx, *projectID, *name)
	if err != nil {
		return nil, fmt.Errorf("couldn't import cluster %s in project %s, error: %s", *name, *projectID, err)
	}

	if err := d.Set("project_id", u.GroupID); err != nil {
		log.Printf(advancedcluster.ErrorClusterSetting, "project_id", u.ID, err)
	}

	if err := d.Set("name", u.Name); err != nil {
		log.Printf(advancedcluster.ErrorClusterSetting, "name", u.ID, err)
	}

	if err := d.Set("cloud_backup", u.ProviderBackupEnabled); err != nil {
		return nil, fmt.Errorf("couldn't import cluster backup configuration %s in project %s, error: %s", *name, *projectID, err)
	}

	d.SetId(conversion.EncodeStateID(map[string]string{
		"cluster_id":    u.ID,
		"project_id":    *projectID,
		"cluster_name":  u.Name,
		"provider_name": u.ProviderSettings.ProviderName,
	}))

	return []*schema.ResourceData{d}, nil
}

func splitSClusterImportID(id string) (projectID, clusterName *string, err error) {
	var re = regexp.MustCompile(`(?s)^([0-9a-fA-F]{24})-(.*)$`)
	parts := re.FindStringSubmatch(id)

	if len(parts) != 3 {
		err = errors.New("import format error: to import a cluster, use the format {project_id}-{name}")
		return
	}

	projectID = &parts[1]
	clusterName = &parts[2]

	return
}

func getInstanceSizeToInt(instanceSize string) int {
	regex := regexp.MustCompile(`\d+`)
	num := regex.FindString(instanceSize)

	return cast.ToInt(num) // if the string is empty it always return 0
}

func isUpgradeRequired(d *schema.ResourceData) bool {
	currentSize, updatedSize := d.GetChange("provider_instance_size_name")

	return currentSize != updatedSize && advancedcluster.IsSharedTier(currentSize.(string))
}

func resourceClusterCustomizeDiff(ctx context.Context, d *schema.ResourceDiff, meta any) error {
	var err error
	currentProvider, updatedProvider := d.GetChange("provider_name")

	willProviderChange := currentProvider != updatedProvider
	willLeaveTenant := willProviderChange && currentProvider == "TENANT"

	if willLeaveTenant {
		err = d.SetNewComputed("backing_provider_name")
	} else if willProviderChange {
		err = d.ForceNew("provider_name")
	}

	return err
}

func getContainerID(containers []matlas.Container, cluster *matlas.Cluster) string {
	if len(containers) != 0 {
		for i := range containers {
			if cluster.ProviderSettings.ProviderName == "GCP" {
				return containers[i].ID
			}

			if containers[i].ProviderName == cluster.ProviderSettings.ProviderName &&
				containers[i].Region == cluster.ProviderSettings.RegionName || // For Azure
				containers[i].RegionName == cluster.ProviderSettings.RegionName { // For AWS
				return containers[i].ID
			}
		}
	}

	return ""
}

func isEqualProviderAutoScalingMinInstanceSize(k, old, newStr string, d *schema.ResourceData) bool {
	canScaleDown, scaleDownOK := d.GetOk("auto_scaling_compute_scale_down_enabled")
	canScaleUp, scaleUpOk := d.GetOk("auto_scaling_compute_enabled")

	if !scaleDownOK || !scaleUpOk {
		return true // if the return is true, it means that both values are the same and there's nothing to do
	}

	if canScaleUp.(bool) && canScaleDown.(bool) {
		if old != newStr {
			return false
		}
	}
	return true
}

func isEqualProviderAutoScalingMaxInstanceSize(k, old, newStr string, d *schema.ResourceData) bool {
	canScaleUp, _ := d.GetOk("auto_scaling_compute_enabled")
	if canScaleUp != nil && canScaleUp.(bool) {
		if old != newStr {
			return false
		}
	}
	return true
}

func updateCluster(ctx context.Context, conn *matlas.Client, connV2 *admin.APIClient, request *matlas.Cluster, projectID, name string, timeout time.Duration) (*matlas.Cluster, *matlas.Response, error) {
	cluster, resp, err := conn.Clusters.Update(ctx, projectID, name, request)
	if err != nil {
		return nil, nil, err
	}

	stateConf := advancedcluster.CreateStateChangeConfig(ctx, connV2, projectID, name, timeout)
	if _, err = stateConf.WaitForStateContext(ctx); err != nil {
		return nil, nil, err
	}

	return cluster, resp, nil
}

func computedCloudProviderSnapshotBackupPolicySchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Computed: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"cluster_id": {
					Type:     schema.TypeString,
					Computed: true,
				},
				"cluster_name": {
					Type:     schema.TypeString,
					Computed: true,
				},
				"next_snapshot": {
					Type:     schema.TypeString,
					Computed: true,
				},
				"reference_hour_of_day": {
					Type:     schema.TypeInt,
					Computed: true,
				},
				"reference_minute_of_hour": {
					Type:     schema.TypeInt,
					Computed: true,
				},
				"restore_window_days": {
					Type:     schema.TypeInt,
					Computed: true,
				},
				"update_snapshots": {
					Type:     schema.TypeBool,
					Computed: true,
				},
				"policies": {
					Type:     schema.TypeList,
					Computed: true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"id": {
								Type:     schema.TypeString,
								Computed: true,
							},
							"policy_item": {
								Type:     schema.TypeList,
								Computed: true,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"id": {
											Type:     schema.TypeString,
											Computed: true,
										},
										"frequency_interval": {
											Type:     schema.TypeInt,
											Computed: true,
										},
										"frequency_type": {
											Type:     schema.TypeString,
											Computed: true,
										},
										"retention_unit": {
											Type:     schema.TypeString,
											Computed: true,
										},
										"retention_value": {
											Type:     schema.TypeInt,
											Computed: true,
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func upgradeCluster(ctx context.Context, conn *matlas.Client, connV2 *admin.APIClient, request *matlas.Cluster, projectID, name string, timeout time.Duration) (*matlas.Cluster, *matlas.Response, error) {
	request.Name = name

	cluster, resp, err := conn.Clusters.Upgrade(ctx, projectID, request)
	if err != nil {
		return nil, nil, err
	}

	stateConf := advancedcluster.CreateStateChangeConfig(ctx, connV2, projectID, name, timeout)
	if _, err = stateConf.WaitForStateContext(ctx); err != nil {
		return nil, nil, err
	}

	return cluster, resp, nil
}
