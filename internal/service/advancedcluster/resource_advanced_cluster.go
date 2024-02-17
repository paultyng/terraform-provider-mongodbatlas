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

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/mongodb/terraform-provider-mongodbatlas/internal/common/constant"
	"github.com/mongodb/terraform-provider-mongodbatlas/internal/common/conversion"
	"github.com/mongodb/terraform-provider-mongodbatlas/internal/config"
	"github.com/spf13/cast"
	matlas "go.mongodb.org/atlas/mongodbatlas"
)

const (
	errorCreate       = "error creating advanced cluster: %s"
	errorRead         = "error reading  advanced cluster (%s): %s"
	errorDelete       = "error deleting advanced cluster (%s): %s"
	errorUpdate       = "error updating advanced cluster (%s): %s"
	errorConfigUpdate = "error updating advanced cluster configuration options (%s): %s"
	errorConfigRead   = "error reading advanced cluster configuration options (%s): %s"
)

type acCtxKey string

var upgradeRequestCtxKey acCtxKey = "upgradeRequest"

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
				Type:     schema.TypeFloat,
				Optional: true,
				Computed: true,
			},
			"encryption_at_rest_provider": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"labels": {
				Type:       schema.TypeSet,
				Optional:   true,
				Set:        HashFunctionForKeyValuePair,
				Deprecated: fmt.Sprintf(constant.DeprecationParamByDateWithReplacement, "September 2024", "tags"),
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
							Type:     schema.TypeString,
							Computed: true,
						},
						"num_shards": {
							Type:         schema.TypeInt,
							Optional:     true,
							Default:      1,
							ValidateFunc: validation.IntBetween(1, 50),
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
										ValidateDiagFunc: StringIsUppercase(),
									},
									"read_only_specs": schemaSpecs(),
									"region_name": {
										Type:             schema.TypeString,
										Required:         true,
										ValidateDiagFunc: StringIsUppercase(),
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
				// Set: replicationSpecsHashSet,
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
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"disk_iops": {
					Type:     schema.TypeInt,
					Optional: true,
					Computed: true,
				},
				"ebs_volume_type": {
					Type:     schema.TypeString,
					Optional: true,
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

	conn := meta.(*config.MongoDBClient).Atlas

	projectID := d.Get("project_id").(string)

	request := &matlas.AdvancedCluster{
		Name:             d.Get("name").(string),
		ClusterType:      cast.ToString(d.Get("cluster_type")),
		ReplicationSpecs: expandAdvancedReplicationSpecs(d.Get("replication_specs").([]any)),
	}

	if v, ok := d.GetOk("backup_enabled"); ok {
		request.BackupEnabled = conversion.Pointer(v.(bool))
	}
	if _, ok := d.GetOk("bi_connector_config"); ok {
		biConnector, err := expandBiConnectorConfig(d)
		if err != nil {
			return diag.FromErr(fmt.Errorf(errorCreate, err))
		}
		request.BiConnector = biConnector
	}
	if v, ok := d.GetOk("disk_size_gb"); ok {
		request.DiskSizeGB = conversion.Pointer(v.(float64))
	}
	if v, ok := d.GetOk("encryption_at_rest_provider"); ok {
		request.EncryptionAtRestProvider = v.(string)
	}

	if _, ok := d.GetOk("labels"); ok && ContainsLabelOrKey(expandLabelSliceFromSetSchema(d), defaultLabel) {
		return diag.FromErr(fmt.Errorf("you should not set `Infrastructure Tool` label, it is used for internal purposes"))
	}
	request.Labels = append(expandLabelSliceFromSetSchema(d), defaultLabel)

	if _, ok := d.GetOk("tags"); ok {
		request.Tags = expandTagSliceFromSetSchema(d)
	}
	if v, ok := d.GetOk("mongo_db_major_version"); ok {
		request.MongoDBMajorVersion = FormatMongoDBMajorVersion(v.(string))
	}
	if v, ok := d.GetOk("pit_enabled"); ok {
		request.PitEnabled = conversion.Pointer(v.(bool))
	}
	if v, ok := d.GetOk("root_cert_type"); ok {
		request.RootCertType = v.(string)
	}
	if v, ok := d.GetOk("termination_protection_enabled"); ok {
		request.TerminationProtectionEnabled = conversion.Pointer(v.(bool))
	}
	if v, ok := d.GetOk("version_release_system"); ok {
		request.VersionReleaseSystem = v.(string)
	}

	// We need to validate the oplog_size_mb attr of the advanced configuration option to show the error
	// before that the cluster is created
	if oplogSizeMB, ok := d.GetOkExists("advanced_configuration.0.oplog_size_mb"); ok {
		if cast.ToInt64(oplogSizeMB) <= 0 {
			return diag.FromErr(fmt.Errorf("`advanced_configuration.oplog_size_mb` cannot be <= 0"))
		}
	}

	cluster, _, err := conn.AdvancedClusters.Create(ctx, projectID, request)
	if err != nil {
		return diag.FromErr(fmt.Errorf(errorCreate, err))
	}

	timeout := d.Timeout(schema.TimeoutCreate)
	stateConf := &retry.StateChangeConf{
		Pending:    []string{"CREATING", "UPDATING", "REPAIRING", "REPEATING", "PENDING"},
		Target:     []string{"IDLE"},
		Refresh:    resourceRefreshFunc(ctx, d.Get("name").(string), projectID, conn),
		Timeout:    timeout,
		MinTimeout: 1 * time.Minute,
		Delay:      3 * time.Minute,
	}

	// Wait, catching any errors
	_, err = stateConf.WaitForStateContext(ctx)
	if err != nil {
		return diag.FromErr(fmt.Errorf(errorCreate, err))
	}

	/*
		So far, the cluster has created correctly, so we need to set up
		the advanced configuration option to attach it
	*/
	ac, ok := d.GetOk("advanced_configuration")
	if aclist, ok1 := ac.([]any); ok1 && len(aclist) > 0 {
		advancedConfReq := expandProcessArgs(d, aclist[0].(map[string]any))

		if ok {
			_, _, err := conn.Clusters.UpdateProcessArgs(ctx, projectID, cluster.Name, advancedConfReq)
			if err != nil {
				return diag.FromErr(fmt.Errorf(errorConfigUpdate, cluster.Name, err))
			}
		}
	}

	// To pause a cluster
	if v := d.Get("paused").(bool); v {
		request = &matlas.AdvancedCluster{
			Paused: conversion.Pointer(v),
		}

		_, _, err = updateAdvancedCluster(ctx, conn, request, projectID, d.Get("name").(string), timeout)
		if err != nil {
			return diag.FromErr(fmt.Errorf(errorUpdate, d.Get("name").(string), err))
		}
	}

	d.SetId(conversion.EncodeStateID(map[string]string{
		"cluster_id":   cluster.ID,
		"project_id":   projectID,
		"cluster_name": cluster.Name,
	}))

	return resourceRead(ctx, d, meta)
}

func resourceRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	connV2 := meta.(*config.MongoDBClient).AtlasV2
	ids := conversion.DecodeStateID(d.Id())
	projectID := ids["project_id"]
	clusterName := ids["cluster_name"]

	cluster, resp, err := connV2.ClustersApi.GetCluster(ctx, projectID, clusterName).Execute()
	if err != nil {
		if resp != nil && resp.StatusCode == http.StatusNotFound {
			d.SetId("")
			return nil
		}

		return diag.FromErr(fmt.Errorf(errorRead, clusterName, err))
	}

	if err := d.Set("cluster_id", cluster.GetId()); err != nil {
		return diag.FromErr(fmt.Errorf(ErrorClusterAdvancedSetting, "cluster_id", clusterName, err))
	}

	if err := d.Set("backup_enabled", cluster.GetBackupEnabled()); err != nil {
		return diag.FromErr(fmt.Errorf(ErrorClusterAdvancedSetting, "backup_enabled", clusterName, err))
	}

	if err := d.Set("bi_connector_config", flattenBiConnectorConfig(cluster.GetBiConnector())); err != nil {
		return diag.FromErr(fmt.Errorf(ErrorClusterAdvancedSetting, "bi_connector_config", clusterName, err))
	}

	if err := d.Set("cluster_type", cluster.GetClusterType()); err != nil {
		return diag.FromErr(fmt.Errorf(ErrorClusterAdvancedSetting, "cluster_type", clusterName, err))
	}

	if err := d.Set("connection_strings", flattenConnectionStrings(cluster.GetConnectionStrings())); err != nil {
		return diag.FromErr(fmt.Errorf(ErrorClusterAdvancedSetting, "connection_strings", clusterName, err))
	}

	if err := d.Set("create_date", conversion.TimePtrToStringPtr(cluster.CreateDate)); err != nil {
		return diag.FromErr(fmt.Errorf(ErrorClusterAdvancedSetting, "create_date", clusterName, err))
	}

	if err := d.Set("disk_size_gb", cluster.GetDiskSizeGB()); err != nil {
		return diag.FromErr(fmt.Errorf(ErrorClusterAdvancedSetting, "disk_size_gb", clusterName, err))
	}

	if err := d.Set("encryption_at_rest_provider", cluster.GetEncryptionAtRestProvider()); err != nil {
		return diag.FromErr(fmt.Errorf(ErrorClusterAdvancedSetting, "encryption_at_rest_provider", clusterName, err))
	}

	if err := d.Set("labels", flattenLabels(cluster.GetLabels())); err != nil {
		return diag.FromErr(fmt.Errorf(ErrorClusterAdvancedSetting, "labels", clusterName, err))
	}

	if err := d.Set("tags", conversion.FlattenTags(cluster.GetTags())); err != nil {
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

	replicationSpecs, err := flattenAdvancedReplicationSpecs(ctx, cluster.GetReplicationSpecs(), d.Get("replication_specs").([]any), d, connV2)
	if err != nil {
		return diag.FromErr(fmt.Errorf(ErrorClusterAdvancedSetting, "replication_specs", clusterName, err))
	}

	if err := d.Set("replication_specs", replicationSpecs); err != nil {
		return diag.FromErr(fmt.Errorf(ErrorClusterAdvancedSetting, "replication_specs", clusterName, err))
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

	if err := d.Set("accept_data_risks_and_force_replica_set_reconfig", conversion.TimePtrToStringPtr(cluster.AcceptDataRisksAndForceReplicaSetReconfig)); err != nil {
		return diag.FromErr(fmt.Errorf(ErrorClusterAdvancedSetting, "accept_data_risks_and_force_replica_set_reconfig", clusterName, err))
	}

	processArgs, _, err := connV2.ClustersApi.GetClusterAdvancedConfiguration(ctx, projectID, clusterName).Execute()
	if err != nil {
		return diag.FromErr(fmt.Errorf(errorConfigRead, clusterName, err))
	}

	if err := d.Set("advanced_configuration", flattenProcessArgs(processArgs)); err != nil {
		return diag.FromErr(fmt.Errorf(ErrorClusterAdvancedSetting, "advanced_configuration", clusterName, err))
	}
	return nil
}

func resourceUpdateOrUpgrade(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	if upgradeRequest := getUpgradeRequest(d); upgradeRequest != nil {
		upgradeCtx := context.WithValue(ctx, upgradeRequestCtxKey, upgradeRequest)
		return resourceMongoDBAtlasAdvancedClusterUpgrade(upgradeCtx, d, meta)
	}

	return resourceMongoDBAtlasAdvancedClusterUpdate(ctx, d, meta)
}

func resourceMongoDBAtlasAdvancedClusterUpgrade(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*config.MongoDBClient).Atlas
	ids := conversion.DecodeStateID(d.Id())
	projectID := ids["project_id"]
	clusterName := ids["cluster_name"]

	upgradeRequest := ctx.Value(upgradeRequestCtxKey).(*matlas.Cluster)

	if upgradeRequest == nil {
		return diag.FromErr(fmt.Errorf("upgrade called without %s in ctx", string(upgradeRequestCtxKey)))
	}

	upgradeResponse, _, err := UpgradeCluster(ctx, conn, upgradeRequest, projectID, clusterName, d.Timeout(schema.TimeoutUpdate))

	if err != nil {
		return diag.FromErr(fmt.Errorf(errorUpdate, clusterName, err))
	}

	d.SetId(conversion.EncodeStateID(map[string]string{
		"cluster_id":   upgradeResponse.ID,
		"project_id":   projectID,
		"cluster_name": clusterName,
	}))

	return resourceRead(ctx, d, meta)
}

func resourceMongoDBAtlasAdvancedClusterUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	// Get client connection.
	conn := meta.(*config.MongoDBClient).Atlas
	ids := conversion.DecodeStateID(d.Id())
	projectID := ids["project_id"]
	clusterName := ids["cluster_name"]

	cluster := new(matlas.AdvancedCluster)
	clusterChangeDetect := new(matlas.AdvancedCluster)

	if d.HasChange("backup_enabled") {
		cluster.BackupEnabled = conversion.Pointer(d.Get("backup_enabled").(bool))
	}

	if d.HasChange("bi_connector_config") {
		cluster.BiConnector, _ = expandBiConnectorConfig(d)
	}

	if d.HasChange("cluster_type") {
		cluster.ClusterType = d.Get("cluster_type").(string)
	}

	if d.HasChange("disk_size_gb") {
		cluster.DiskSizeGB = conversion.Pointer(d.Get("disk_size_gb").(float64))
	}

	if d.HasChange("encryption_at_rest_provider") {
		cluster.EncryptionAtRestProvider = d.Get("encryption_at_rest_provider").(string)
	}

	if d.HasChange("labels") {
		if ContainsLabelOrKey(expandLabelSliceFromSetSchema(d), defaultLabel) {
			return diag.FromErr(fmt.Errorf("you should not set `Infrastructure Tool` label, it is used for internal purposes"))
		}

		cluster.Labels = append(expandLabelSliceFromSetSchema(d), defaultLabel)
	}

	if d.HasChange("tags") {
		cluster.Tags = expandTagSliceFromSetSchema(d)
	}

	if d.HasChange("mongo_db_major_version") {
		cluster.MongoDBMajorVersion = FormatMongoDBMajorVersion(d.Get("mongo_db_major_version"))
	}

	if d.HasChange("pit_enabled") {
		cluster.PitEnabled = conversion.Pointer(d.Get("pit_enabled").(bool))
	}

	if d.HasChange("replication_specs") {
		cluster.ReplicationSpecs = expandAdvancedReplicationSpecs(d.Get("replication_specs").([]any))
	}

	if d.HasChange("root_cert_type") {
		cluster.RootCertType = d.Get("root_cert_type").(string)
	}

	if d.HasChange("termination_protection_enabled") {
		cluster.TerminationProtectionEnabled = conversion.Pointer(d.Get("termination_protection_enabled").(bool))
	}

	if d.HasChange("version_release_system") {
		cluster.VersionReleaseSystem = d.Get("version_release_system").(string)
	}

	if d.HasChange("accept_data_risks_and_force_replica_set_reconfig") {
		cluster.AcceptDataRisksAndForceReplicaSetReconfig = d.Get("accept_data_risks_and_force_replica_set_reconfig").(string)
	}

	if d.HasChange("paused") && !d.Get("paused").(bool) {
		cluster.Paused = conversion.Pointer(d.Get("paused").(bool))
	}

	timeout := d.Timeout(schema.TimeoutUpdate)

	if d.HasChange("advanced_configuration") {
		ac := d.Get("advanced_configuration")
		if aclist, ok := ac.([]any); ok && len(aclist) > 0 {
			advancedConfReq := expandProcessArgs(d, aclist[0].(map[string]any))
			if !reflect.DeepEqual(advancedConfReq, matlas.ProcessArgs{}) {
				_, _, err := conn.Clusters.UpdateProcessArgs(ctx, projectID, clusterName, advancedConfReq)
				if err != nil {
					return diag.FromErr(fmt.Errorf(errorConfigUpdate, clusterName, err))
				}
			}
		}
	}

	// Has changes
	if !reflect.DeepEqual(cluster, clusterChangeDetect) {
		err := retry.RetryContext(ctx, timeout, func() *retry.RetryError {
			_, resp, err := updateAdvancedCluster(ctx, conn, cluster, projectID, clusterName, timeout)
			if err != nil {
				if resp == nil || resp.StatusCode == 400 {
					return retry.NonRetryableError(fmt.Errorf(errorUpdate, clusterName, err))
				}
				return retry.RetryableError(fmt.Errorf(errorUpdate, clusterName, err))
			}
			return nil
		})
		if err != nil {
			return diag.FromErr(fmt.Errorf(errorUpdate, clusterName, err))
		}
	}

	if d.Get("paused").(bool) {
		clusterRequest := &matlas.AdvancedCluster{
			Paused: conversion.Pointer(true),
		}

		_, _, err := updateAdvancedCluster(ctx, conn, clusterRequest, projectID, clusterName, timeout)
		if err != nil {
			return diag.FromErr(fmt.Errorf(errorUpdate, clusterName, err))
		}
	}

	return resourceRead(ctx, d, meta)
}

func resourceDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	// Get client connection.
	conn := meta.(*config.MongoDBClient).Atlas
	ids := conversion.DecodeStateID(d.Id())
	projectID := ids["project_id"]
	clusterName := ids["cluster_name"]

	var options *matlas.DeleteAdvanceClusterOptions
	if v, ok := d.GetOkExists("retain_backups_enabled"); ok {
		options = &matlas.DeleteAdvanceClusterOptions{
			RetainBackups: conversion.Pointer(v.(bool)),
		}
	}

	_, err := conn.AdvancedClusters.Delete(ctx, projectID, clusterName, options)
	if err != nil {
		return diag.FromErr(fmt.Errorf(errorDelete, clusterName, err))
	}

	log.Println("[INFO] Waiting for MongoDB ClusterAdvanced to be destroyed")

	stateConf := &retry.StateChangeConf{
		Pending:    []string{"IDLE", "CREATING", "UPDATING", "REPAIRING", "DELETING"},
		Target:     []string{"DELETED"},
		Refresh:    resourceRefreshFunc(ctx, clusterName, projectID, conn),
		Timeout:    d.Timeout(schema.TimeoutDelete),
		MinTimeout: 30 * time.Second,
		Delay:      1 * time.Minute, // Wait 30 secs before starting
	}

	// Wait, catching any errors
	_, err = stateConf.WaitForStateContext(ctx)
	if err != nil {
		return diag.FromErr(fmt.Errorf(errorDelete, clusterName, err))
	}

	return nil
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

func expandAdvancedReplicationSpec(tfMap map[string]any) *matlas.AdvancedReplicationSpec {
	if tfMap == nil {
		return nil
	}

	apiObject := &matlas.AdvancedReplicationSpec{
		NumShards:     tfMap["num_shards"].(int),
		ZoneName:      tfMap["zone_name"].(string),
		RegionConfigs: expandRegionConfigs(tfMap["region_configs"].([]any)),
	}

	if tfMap["id"].(string) != "" {
		apiObject.ID = tfMap["id"].(string)
	}

	return apiObject
}

func expandAdvancedReplicationSpecs(tfList []any) []*matlas.AdvancedReplicationSpec {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []*matlas.AdvancedReplicationSpec

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]any)

		if !ok {
			continue
		}

		apiObject := expandAdvancedReplicationSpec(tfMap)

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func expandRegionConfig(tfMap map[string]any) *matlas.AdvancedRegionConfig {
	if tfMap == nil {
		return nil
	}

	providerName := tfMap["provider_name"].(string)
	apiObject := &matlas.AdvancedRegionConfig{
		Priority:     conversion.Pointer(cast.ToInt(tfMap["priority"])),
		ProviderName: providerName,
		RegionName:   tfMap["region_name"].(string),
	}

	if v, ok := tfMap["analytics_specs"]; ok && len(v.([]any)) > 0 {
		apiObject.AnalyticsSpecs = expandRegionConfigSpec(v.([]any), providerName)
	}
	if v, ok := tfMap["electable_specs"]; ok && len(v.([]any)) > 0 {
		apiObject.ElectableSpecs = expandRegionConfigSpec(v.([]any), providerName)
	}
	if v, ok := tfMap["read_only_specs"]; ok && len(v.([]any)) > 0 {
		apiObject.ReadOnlySpecs = expandRegionConfigSpec(v.([]any), providerName)
	}
	if v, ok := tfMap["auto_scaling"]; ok && len(v.([]any)) > 0 {
		apiObject.AutoScaling = expandRegionConfigAutoScaling(v.([]any))
	}
	if v, ok := tfMap["analytics_auto_scaling"]; ok && len(v.([]any)) > 0 {
		apiObject.AnalyticsAutoScaling = expandRegionConfigAutoScaling(v.([]any))
	}
	if v, ok := tfMap["backing_provider_name"]; ok {
		apiObject.BackingProviderName = v.(string)
	}

	return apiObject
}

func expandRegionConfigs(tfList []any) []*matlas.AdvancedRegionConfig {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []*matlas.AdvancedRegionConfig

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]any)

		if !ok {
			continue
		}

		apiObject := expandRegionConfig(tfMap)

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func expandRegionConfigSpec(tfList []any, providerName string) *matlas.Specs {
	if tfList == nil && len(tfList) > 0 {
		return nil
	}

	tfMap, _ := tfList[0].(map[string]any)

	apiObject := &matlas.Specs{}

	if providerName == "AWS" {
		if v, ok := tfMap["disk_iops"]; ok && v.(int) > 0 {
			apiObject.DiskIOPS = conversion.Pointer(cast.ToInt64(v.(int)))
		}
		if v, ok := tfMap["ebs_volume_type"]; ok {
			apiObject.EbsVolumeType = v.(string)
		}
	}
	if v, ok := tfMap["instance_size"]; ok {
		apiObject.InstanceSize = v.(string)
	}
	if v, ok := tfMap["node_count"]; ok {
		apiObject.NodeCount = conversion.Pointer(v.(int))
	}

	return apiObject
}

func expandRegionConfigAutoScaling(tfList []any) *matlas.AdvancedAutoScaling {
	if tfList == nil && len(tfList) > 0 {
		return nil
	}

	tfMap, _ := tfList[0].(map[string]any)

	advancedAutoScaling := &matlas.AdvancedAutoScaling{}
	diskGB := &matlas.DiskGB{}
	compute := &matlas.Compute{}

	if v, ok := tfMap["disk_gb_enabled"]; ok {
		diskGB.Enabled = conversion.Pointer(v.(bool))
	}
	if v, ok := tfMap["compute_enabled"]; ok {
		compute.Enabled = conversion.Pointer(v.(bool))
	}
	if v, ok := tfMap["compute_scale_down_enabled"]; ok {
		compute.ScaleDownEnabled = conversion.Pointer(v.(bool))
	}
	if v, ok := tfMap["compute_min_instance_size"]; ok {
		value := compute.ScaleDownEnabled
		if *value {
			compute.MinInstanceSize = v.(string)
		}
	}
	if v, ok := tfMap["compute_max_instance_size"]; ok {
		value := compute.Enabled
		if *value {
			compute.MaxInstanceSize = v.(string)
		}
	}

	advancedAutoScaling.DiskGB = diskGB
	advancedAutoScaling.Compute = compute

	return advancedAutoScaling
}

func resourceRefreshFunc(ctx context.Context, name, projectID string, client *matlas.Client) retry.StateRefreshFunc {
	return func() (any, string, error) {
		c, resp, err := client.AdvancedClusters.Get(ctx, projectID, name)

		if err != nil && strings.Contains(err.Error(), "reset by peer") {
			return nil, "REPEATING", nil
		}

		if err != nil && c == nil && resp == nil {
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

		if c.StateName != "" {
			log.Printf("[DEBUG] status for MongoDB cluster: %s: %s", name, c.StateName)
		}

		return c, c.StateName, nil
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

func getUpgradeRequest(d *schema.ResourceData) *matlas.Cluster {
	if !d.HasChange("replication_specs") {
		return nil
	}

	cs, us := d.GetChange("replication_specs")
	currentSpecs := expandAdvancedReplicationSpecs(cs.([]any))
	updatedSpecs := expandAdvancedReplicationSpecs(us.([]any))

	if len(currentSpecs) != 1 || len(updatedSpecs) != 1 || len(currentSpecs[0].RegionConfigs) != 1 || len(updatedSpecs[0].RegionConfigs) != 1 {
		return nil
	}

	currentRegion := currentSpecs[0].RegionConfigs[0]
	updatedRegion := updatedSpecs[0].RegionConfigs[0]
	currentSize := currentRegion.ElectableSpecs.InstanceSize

	if currentRegion.ElectableSpecs.InstanceSize == updatedRegion.ElectableSpecs.InstanceSize || !IsSharedTier(currentSize) {
		return nil
	}

	return &matlas.Cluster{
		ProviderSettings: &matlas.ProviderSettings{
			ProviderName:     updatedRegion.ProviderName,
			InstanceSizeName: updatedRegion.ElectableSpecs.InstanceSize,
			RegionName:       updatedRegion.RegionName,
		},
	}
}

func updateAdvancedCluster(
	ctx context.Context,
	conn *matlas.Client,
	request *matlas.AdvancedCluster,
	projectID, name string,
	timeout time.Duration,
) (*matlas.AdvancedCluster, *matlas.Response, error) {
	cluster, resp, err := conn.AdvancedClusters.Update(ctx, projectID, name, request)
	if err != nil {
		return nil, nil, err
	}

	stateConf := &retry.StateChangeConf{
		Pending:    []string{"CREATING", "UPDATING", "REPAIRING"},
		Target:     []string{"IDLE"},
		Refresh:    resourceRefreshFunc(ctx, name, projectID, conn),
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
