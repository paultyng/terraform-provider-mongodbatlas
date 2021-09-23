package mongodbatlas

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
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/mwielbut/pointy"
	"github.com/spf13/cast"
	matlas "go.mongodb.org/atlas/mongodbatlas"
)

const (
	errorClusterAdvancedCreate  = "error creating MongoDB ClusterAdvanced: %s"
	errorClusterAdvancedRead    = "error reading MongoDB ClusterAdvanced (%s): %s"
	errorClusterAdvancedDelete  = "error deleting MongoDB ClusterAdvanced (%s): %s"
	errorClusterAdvancedUpdate  = "error updating MongoDB ClusterAdvanced (%s): %s"
	errorClusterAdvancedSetting = "error setting `%s` for MongoDB ClusterAdvanced (%s): %s"
)

func resourceMongoDBAtlasAdvancedCluster() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceMongoDBAtlasAdvancedClusterCreate,
		ReadWithoutTimeout:   resourceMongoDBAtlasAdvancedClusterRead,
		UpdateWithoutTimeout: resourceMongoDBAtlasAdvancedClusterUpdate,
		DeleteWithoutTimeout: resourceMongoDBAtlasAdvancedClusterDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceMongoDBAtlasAdvancedClusterImportState,
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
			"bi_connector": {
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
			"connection_strings": clusterConnectionStringsSchema(),
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
				Type:     schema.TypeSet,
				Optional: true,
				Set: func(v interface{}) int {
					var buf bytes.Buffer
					m := v.(map[string]interface{})
					buf.WriteString(m["key"].(string))
					buf.WriteString(m["value"].(string))
					return HashCodeString(buf.String())
				},
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"key": {
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
						},
						"value": {
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
						},
					},
				},
			},
			"mongo_db_major_version": {
				Type:      schema.TypeString,
				Optional:  true,
				Computed:  true,
				StateFunc: formatMongoDBMajorVersion,
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
				Computed: true,
			},
			"pit_enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
			},
			"replication_specs": {
				Type:     schema.TypeSet,
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
							Type:     schema.TypeSet,
							Required: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"analytics_specs": advancedClusterRegionConfigsSpecsSchema(),
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
									"backing_provider_name": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"electable_specs": advancedClusterRegionConfigsSpecsSchema(),
									"priority": {
										Type:     schema.TypeInt,
										Required: true,
									},
									"provider_name": {
										Type:     schema.TypeString,
										Required: true,
									},
									"read_only_specs": advancedClusterRegionConfigsSpecsSchema(),
									"region_name": {
										Type:     schema.TypeString,
										Required: true,
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
				Set: replicationSpecsHashSet,
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
			"version_release_system": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "LTS",
				ValidateFunc: validation.StringInSlice([]string{"LTS", "CONTINUOUS"}, false),
			},
		},
	}
}

func advancedClusterRegionConfigsSpecsSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		MaxItems: 1,
		Optional: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"disk_iops": {
					Type:     schema.TypeInt,
					Optional: true,
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

func resourceMongoDBAtlasAdvancedClusterCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// Get client connection.
	conn := meta.(*MongoDBClient).Atlas

	projectID := d.Get("project_id").(string)

	request := &matlas.AdvancedCluster{
		Name:             d.Get("name").(string),
		ClusterType:      cast.ToString(d.Get("cluster_type")),
		ReplicationSpecs: expandAdvancedReplicationSpecs(d.Get("replication_specs").(*schema.Set).List()),
	}

	if v, ok := d.GetOk("backup_enabled"); ok {
		request.BackupEnabled = pointy.Bool(v.(bool))
	}
	if _, ok := d.GetOk("bi_connector"); ok {
		biConnector, err := expandBiConnectorConfig(d)
		if err != nil {
			return diag.FromErr(fmt.Errorf(errorClusterAdvancedCreate, err))
		}
		request.BiConnector = biConnector
	}
	if v, ok := d.GetOk("disk_size_gb"); ok {
		request.DiskSizeGB = pointy.Float64(v.(float64))
	}
	if v, ok := d.GetOk("encryption_at_rest_provider"); ok {
		request.EncryptionAtRestProvider = v.(string)
	}

	if _, ok := d.GetOk("labels"); ok {
		if containsLabelOrKey(expandLabelSliceFromSetSchema(d), defaultLabel) {
			return diag.FromErr(fmt.Errorf("you should not set `Infrastructure Tool` label, it is used for internal purposes"))
		}
		request.Labels = append(expandLabelSliceFromSetSchema(d), defaultLabel)
	}

	if v, ok := d.GetOk("mongo_db_major_version"); ok {
		request.MongoDBMajorVersion = formatMongoDBMajorVersion(v.(string))
	}
	if v, ok := d.GetOk("pit_enabled"); ok {
		request.PitEnabled = pointy.Bool(v.(bool))
	}
	if v, ok := d.GetOk("root_cert_type"); ok {
		request.RootCertType = v.(string)
	}
	if v, ok := d.GetOk("version_release_system"); ok {
		request.VersionReleaseSystem = v.(string)
	}

	cluster, _, err := conn.AdvancedClusters.Create(ctx, projectID, request)
	if err != nil {
		return diag.FromErr(fmt.Errorf(errorClusterAdvancedCreate, err))
	}

	stateConf := &resource.StateChangeConf{
		Pending:    []string{"CREATING", "UPDATING", "REPAIRING", "REPEATING", "PENDING"},
		Target:     []string{"IDLE"},
		Refresh:    resourceClusterAdvancedRefreshFunc(ctx, d.Get("name").(string), projectID, conn),
		Timeout:    3 * time.Hour,
		MinTimeout: 1 * time.Minute,
		Delay:      3 * time.Minute,
	}

	// Wait, catching any errors
	_, err = stateConf.WaitForStateContext(ctx)
	if err != nil {
		return diag.FromErr(fmt.Errorf(errorClusterAdvancedCreate, err))
	}

	d.SetId(encodeStateID(map[string]string{
		"cluster_id":   cluster.ID,
		"project_id":   projectID,
		"cluster_name": cluster.Name,
	}))

	return resourceMongoDBAtlasAdvancedClusterRead(ctx, d, meta)
}

func resourceMongoDBAtlasAdvancedClusterRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// Get client connection.
	conn := meta.(*MongoDBClient).Atlas
	ids := decodeStateID(d.Id())
	projectID := ids["project_id"]
	clusterName := ids["cluster_name"]

	cluster, resp, err := conn.AdvancedClusters.Get(ctx, projectID, clusterName)
	if err != nil {
		if resp != nil && resp.StatusCode == http.StatusNotFound {
			d.SetId("")
			return nil
		}

		return diag.FromErr(fmt.Errorf(errorClusterAdvancedRead, clusterName, err))
	}

	log.Printf("[DEBUG] GET ClusterAdvanced %+v", cluster)

	if err := d.Set("cluster_id", cluster.ID); err != nil {
		return diag.FromErr(fmt.Errorf(errorClusterAdvancedSetting, "cluster_id", clusterName, err))
	}

	if err := d.Set("backup_enabled", cluster.BackupEnabled); err != nil {
		return diag.FromErr(fmt.Errorf(errorClusterAdvancedSetting, "backup_enabled", clusterName, err))
	}

	if err := d.Set("bi_connector", flattenBiConnectorConfig(cluster.BiConnector)); err != nil {
		return diag.FromErr(fmt.Errorf(errorClusterAdvancedSetting, "bi_connector", clusterName, err))
	}

	if err := d.Set("cluster_type", cluster.ClusterType); err != nil {
		return diag.FromErr(fmt.Errorf(errorClusterAdvancedSetting, "cluster_type", clusterName, err))
	}

	if err := d.Set("connection_strings", flattenConnectionStrings(cluster.ConnectionStrings)); err != nil {
		return diag.FromErr(fmt.Errorf(errorClusterAdvancedSetting, "connection_strings", clusterName, err))
	}

	if err := d.Set("create_date", cluster.CreateDate); err != nil {
		return diag.FromErr(fmt.Errorf(errorClusterAdvancedSetting, "create_date", clusterName, err))
	}

	if err := d.Set("disk_size_gb", cluster.DiskSizeGB); err != nil {
		return diag.FromErr(fmt.Errorf(errorClusterAdvancedSetting, "disk_size_gb", clusterName, err))
	}

	if err := d.Set("encryption_at_rest_provider", cluster.EncryptionAtRestProvider); err != nil {
		return diag.FromErr(fmt.Errorf(errorClusterAdvancedSetting, "encryption_at_rest_provider", clusterName, err))
	}

	if err := d.Set("labels", flattenLabels(removeLabel(cluster.Labels, defaultLabel))); err != nil {
		return diag.FromErr(fmt.Errorf(errorClusterAdvancedSetting, "labels", clusterName, err))
	}

	if err := d.Set("mongo_db_major_version", cluster.MongoDBMajorVersion); err != nil {
		return diag.FromErr(fmt.Errorf(errorClusterAdvancedSetting, "mongo_db_major_version", clusterName, err))
	}

	if err := d.Set("mongo_db_version", cluster.MongoDBVersion); err != nil {
		return diag.FromErr(fmt.Errorf(errorClusterAdvancedSetting, "mongo_db_version", clusterName, err))
	}

	if err := d.Set("name", cluster.Name); err != nil {
		return diag.FromErr(fmt.Errorf(errorClusterAdvancedSetting, "name", clusterName, err))
	}

	if err := d.Set("paused", cluster.Paused); err != nil {
		return diag.FromErr(fmt.Errorf(errorClusterAdvancedSetting, "paused", clusterName, err))
	}

	if err := d.Set("pit_enabled", cluster.PitEnabled); err != nil {
		return diag.FromErr(fmt.Errorf(errorClusterAdvancedSetting, "pit_enabled", clusterName, err))
	}

	if err := d.Set("replication_specs", flattenAdvancedReplicationSpecs(cluster.ReplicationSpecs, d.Get("replication_specs").(*schema.Set).List())); err != nil {
		return diag.FromErr(fmt.Errorf(errorClusterAdvancedSetting, "replication_specs", clusterName, err))
	}

	if err := d.Set("root_cert_type", cluster.RootCertType); err != nil {
		return diag.FromErr(fmt.Errorf(errorClusterAdvancedSetting, "state_name", clusterName, err))
	}

	if err := d.Set("state_name", cluster.StateName); err != nil {
		return diag.FromErr(fmt.Errorf(errorClusterAdvancedSetting, "state_name", clusterName, err))
	}

	if err := d.Set("version_release_system", cluster.VersionReleaseSystem); err != nil {
		return diag.FromErr(fmt.Errorf(errorClusterAdvancedSetting, "version_release_system", clusterName, err))
	}

	return nil
}

func resourceMongoDBAtlasAdvancedClusterUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// Get client connection.
	conn := meta.(*MongoDBClient).Atlas
	ids := decodeStateID(d.Id())
	projectID := ids["project_id"]
	clusterName := ids["cluster_name"]

	cluster := new(matlas.AdvancedCluster)

	if d.HasChange("backup_enabled") {
		cluster.BackupEnabled = pointy.Bool(d.Get("backup_enabled").(bool))
	}

	if d.HasChange("bi_connector") {
		cluster.BiConnector, _ = expandBiConnectorConfig(d)
	}

	if d.HasChange("cluster_type") {
		cluster.ClusterType = d.Get("cluster_type").(string)
	}

	if d.HasChange("disk_size_gb") {
		cluster.DiskSizeGB = pointy.Float64(d.Get("disk_size_gb").(float64))
	}

	if d.HasChange("encryption_at_rest_provider") {
		cluster.EncryptionAtRestProvider = d.Get("encryption_at_rest_provider").(string)
	}

	if d.HasChange("labels") {
		if containsLabelOrKey(expandLabelSliceFromSetSchema(d), defaultLabel) {
			return diag.FromErr(fmt.Errorf("you should not set `Infrastructure Tool` label, it is used for internal purposes"))
		}

		cluster.Labels = append(expandLabelSliceFromSetSchema(d), defaultLabel)
	}

	if d.HasChange("mongo_db_major_version") {
		cluster.MongoDBMajorVersion = formatMongoDBMajorVersion(d.Get("mongo_db_major_version"))
	}

	if d.HasChange("pit_enabled") {
		cluster.PitEnabled = pointy.Bool(d.Get("pit_enabled").(bool))
	}

	if d.HasChange("replication_specs") {
		cluster.ReplicationSpecs = expandAdvancedReplicationSpecs(d.Get("replication_specs").(*schema.Set).List())
	}

	if d.HasChange("root_cert_type") {
		cluster.RootCertType = d.Get("root_cert_type").(string)
	}

	if d.HasChange("version_release_system") {
		cluster.VersionReleaseSystem = d.Get("version_release_system").(string)
	}

	// Has changes
	if !reflect.DeepEqual(cluster, matlas.AdvancedCluster{}) {
		_, _, err := conn.AdvancedClusters.Update(ctx, projectID, clusterName, cluster)
		if err != nil {
			return diag.FromErr(fmt.Errorf(errorClusterAdvancedUpdate, clusterName, err))
		}
	}

	stateConf := &resource.StateChangeConf{
		Pending:    []string{"CREATING", "UPDATING", "REPAIRING"},
		Target:     []string{"IDLE"},
		Refresh:    resourceClusterAdvancedRefreshFunc(ctx, clusterName, projectID, conn),
		Timeout:    3 * time.Hour,
		MinTimeout: 30 * time.Second,
		Delay:      1 * time.Minute,
	}

	// Wait, catching any errors
	_, err := stateConf.WaitForStateContext(ctx)
	if err != nil {
		return diag.FromErr(fmt.Errorf(errorClusterAdvancedUpdate, clusterName, err))
	}

	return resourceMongoDBAtlasAdvancedClusterRead(ctx, d, meta)
}

func resourceMongoDBAtlasAdvancedClusterDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// Get client connection.
	conn := meta.(*MongoDBClient).Atlas
	ids := decodeStateID(d.Id())
	projectID := ids["project_id"]
	clusterName := ids["cluster_name"]

	_, err := conn.AdvancedClusters.Delete(ctx, projectID, clusterName)

	if err != nil {
		return diag.FromErr(fmt.Errorf(errorClusterAdvancedDelete, clusterName, err))
	}

	log.Println("[INFO] Waiting for MongoDB ClusterAdvanced to be destroyed")

	stateConf := &resource.StateChangeConf{
		Pending:    []string{"IDLE", "CREATING", "UPDATING", "REPAIRING", "DELETING"},
		Target:     []string{"DELETED"},
		Refresh:    resourceClusterAdvancedRefreshFunc(ctx, clusterName, projectID, conn),
		Timeout:    3 * time.Hour,
		MinTimeout: 30 * time.Second,
		Delay:      1 * time.Minute, // Wait 30 secs before starting
	}

	// Wait, catching any errors
	_, err = stateConf.WaitForStateContext(ctx)
	if err != nil {
		return diag.FromErr(fmt.Errorf(errorClusterAdvancedDelete, clusterName, err))
	}

	return nil
}

func resourceMongoDBAtlasAdvancedClusterImportState(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	conn := meta.(*MongoDBClient).Atlas

	projectID, name, err := splitSClusterAdvancedImportID(d.Id())
	if err != nil {
		return nil, err
	}

	u, _, err := conn.AdvancedClusters.Get(ctx, *projectID, *name)
	if err != nil {
		return nil, fmt.Errorf("couldn't import cluster %s in project %s, error: %s", *name, *projectID, err)
	}

	if err := d.Set("project_id", u.GroupID); err != nil {
		log.Printf(errorClusterAdvancedSetting, "project_id", u.ID, err)
	}

	if err := d.Set("name", u.Name); err != nil {
		log.Printf(errorClusterAdvancedSetting, "name", u.ID, err)
	}

	d.SetId(encodeStateID(map[string]string{
		"cluster_id":   u.ID,
		"project_id":   *projectID,
		"cluster_name": u.Name,
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

func expandAdvancedReplicationSpec(tfMap map[string]interface{}) *matlas.AdvancedReplicationSpec {
	if tfMap == nil {
		return nil
	}

	apiObject := &matlas.AdvancedReplicationSpec{
		NumShards:     tfMap["num_shards"].(int),
		ZoneName:      tfMap["zone_name"].(string),
		RegionConfigs: expandRegionConfigs(tfMap["region_configs"].(*schema.Set).List()),
	}

	return apiObject
}

func expandAdvancedReplicationSpecs(tfList []interface{}) []*matlas.AdvancedReplicationSpec {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []*matlas.AdvancedReplicationSpec

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})

		if !ok {
			continue
		}

		apiObject := expandAdvancedReplicationSpec(tfMap)

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func expandRegionConfig(tfMap map[string]interface{}) *matlas.AdvancedRegionConfig {
	if tfMap == nil {
		return nil
	}

	providerName := tfMap["provider_name"].(string)
	apiObject := &matlas.AdvancedRegionConfig{
		Priority:     pointy.Int(cast.ToInt(tfMap["priority"])),
		ProviderName: providerName,
		RegionName:   tfMap["region_name"].(string),
	}

	if v, ok := tfMap["analytics_specs"]; ok && len(v.([]interface{})) > 0 {
		apiObject.AnalyticsSpecs = expandRegionConfigSpec(v.([]interface{}), providerName)
	}
	if v, ok := tfMap["electable_specs"]; ok && len(v.([]interface{})) > 0 {
		apiObject.ElectableSpecs = expandRegionConfigSpec(v.([]interface{}), providerName)
	}
	if v, ok := tfMap["read_only_specs"]; ok && len(v.([]interface{})) > 0 {
		apiObject.ReadOnlySpecs = expandRegionConfigSpec(v.([]interface{}), providerName)
	}
	if v, ok := tfMap["auto_scaling"]; ok && len(v.([]interface{})) > 0 {
		apiObject.AutoScaling = expandRegionConfigAutoScaling(v.([]interface{}))
	}
	if v, ok := tfMap["backing_provider_name"]; ok {
		apiObject.BackingProviderName = v.(string)
	}

	return apiObject
}

func expandRegionConfigs(tfList []interface{}) []*matlas.AdvancedRegionConfig {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []*matlas.AdvancedRegionConfig

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})

		if !ok {
			continue
		}

		apiObject := expandRegionConfig(tfMap)

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func expandRegionConfigSpec(tfList []interface{}, providerName string) *matlas.Specs {
	if tfList == nil && len(tfList) > 0 {
		return nil
	}

	tfMap, _ := tfList[0].(map[string]interface{})

	apiObject := &matlas.Specs{}

	if providerName == "AWS" {
		if v, ok := tfMap["disk_iops"]; ok && v.(int) > 0 {
			apiObject.DiskIOPS = pointy.Int64(cast.ToInt64(v.(int)))
		}
		if v, ok := tfMap["ebs_volume_type"]; ok {
			apiObject.EbsVolumeType = v.(string)
		}
	}
	if v, ok := tfMap["instance_size"]; ok {
		apiObject.InstanceSize = v.(string)
	}
	if v, ok := tfMap["node_count"]; ok {
		apiObject.NodeCount = pointy.Int(v.(int))
	}

	return apiObject
}

func expandRegionConfigAutoScaling(tfList []interface{}) *matlas.AdvancedAutoScaling {
	if tfList == nil && len(tfList) > 0 {
		return nil
	}

	tfMap, _ := tfList[0].(map[string]interface{})

	apiObject := &matlas.AdvancedAutoScaling{}
	apiObject2 := &matlas.DiskGB{}
	apiObject3 := &matlas.Compute{}

	if v, ok := tfMap["disk_gb_enabled"]; ok {
		apiObject2.Enabled = pointy.Bool(v.(bool))
	}
	if v, ok := tfMap["compute_enabled"]; ok {
		apiObject3.Enabled = pointy.Bool(v.(bool))
	}
	if v, ok := tfMap["compute_scale_down_enabled"]; ok {
		apiObject3.ScaleDownEnabled = pointy.Bool(v.(bool))
	}
	if v, ok := tfMap["compute_min_instance_size"]; ok {
		apiObject3.MinInstanceSize = v.(string)
	}
	if v, ok := tfMap["compute_max_instance_size"]; ok {
		apiObject3.MaxInstanceSize = v.(string)
	}

	return apiObject
}

func flattenAdvancedReplicationSpec(apiObject *matlas.AdvancedReplicationSpec, tfMapObject map[string]interface{}) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	tfMap["num_shards"] = apiObject.NumShards
	tfMap["id"] = apiObject.ID
	if tfMapObject != nil {
		tfMap["region_configs"] = flattenAdvancedReplicationSpecRegionConfigs(apiObject.RegionConfigs, tfMapObject["region_configs"].(*schema.Set).List())
	} else {
		tfMap["region_configs"] = flattenAdvancedReplicationSpecRegionConfigs(apiObject.RegionConfigs, nil)
	}
	tfMap["zone_name"] = apiObject.ZoneName

	return tfMap
}

func flattenAdvancedReplicationSpecs(apiObjects []*matlas.AdvancedReplicationSpec, tfMapObjects []interface{}) []map[string]interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []map[string]interface{}

	for i, apiObject := range apiObjects {
		if apiObject == nil {
			continue
		}

		if len(tfMapObjects) > 0 {
			tfMapObject := tfMapObjects[i].(map[string]interface{})
			tfList = append(tfList, flattenAdvancedReplicationSpec(apiObject, tfMapObject))
		} else {
			tfList = append(tfList, flattenAdvancedReplicationSpec(apiObject, nil))
		}
	}

	return tfList
}

func flattenAdvancedReplicationSpecRegionConfig(apiObject *matlas.AdvancedRegionConfig, tfMapObject map[string]interface{}) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if tfMapObject != nil {
		if v, ok := tfMapObject["analytics_specs"]; ok && len(v.([]interface{})) > 0 {
			tfMap["analytics_specs"] = flattenAdvancedReplicationSpecRegionConfigSpec(apiObject.AnalyticsSpecs, apiObject.ProviderName, tfMapObject["analytics_specs"].([]interface{}))
		}
		if v, ok := tfMapObject["electable_specs"]; ok && len(v.([]interface{})) > 0 {
			tfMap["electable_specs"] = flattenAdvancedReplicationSpecRegionConfigSpec(apiObject.ElectableSpecs, apiObject.ProviderName, tfMapObject["electable_specs"].([]interface{}))
		}
		if v, ok := tfMapObject["read_only_specs"]; ok && len(v.([]interface{})) > 0 {
			tfMap["read_only_specs"] = flattenAdvancedReplicationSpecRegionConfigSpec(apiObject.ReadOnlySpecs, apiObject.ProviderName, tfMapObject["read_only_specs"].([]interface{}))
		}
		if v, ok := tfMapObject["auto_scaling"]; ok && len(v.([]interface{})) > 0 {
			tfMap["auto_scaling"] = flattenAdvancedReplicationSpecAutoScaling(apiObject.AutoScaling)
		}
	} else {
		tfMap["analytics_specs"] = flattenAdvancedReplicationSpecRegionConfigSpec(apiObject.AnalyticsSpecs, apiObject.ProviderName, nil)
		tfMap["electable_specs"] = flattenAdvancedReplicationSpecRegionConfigSpec(apiObject.ElectableSpecs, apiObject.ProviderName, nil)
		tfMap["read_only_specs"] = flattenAdvancedReplicationSpecRegionConfigSpec(apiObject.ReadOnlySpecs, apiObject.ProviderName, nil)
		tfMap["auto_scaling"] = flattenAdvancedReplicationSpecAutoScaling(apiObject.AutoScaling)
	}

	tfMap["region_name"] = apiObject.RegionName
	tfMap["provider_name"] = apiObject.ProviderName
	tfMap["backing_provider_name"] = apiObject.BackingProviderName
	tfMap["priority"] = apiObject.Priority

	return tfMap
}

func flattenAdvancedReplicationSpecRegionConfigs(apiObjects []*matlas.AdvancedRegionConfig, tfMapObjects []interface{}) []map[string]interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []map[string]interface{}

	for i, apiObject := range apiObjects {
		if apiObject == nil {
			continue
		}

		if len(tfMapObjects) > 0 {
			tfMapObject := tfMapObjects[i].(map[string]interface{})
			tfList = append(tfList, flattenAdvancedReplicationSpecRegionConfig(apiObject, tfMapObject))
		} else {
			tfList = append(tfList, flattenAdvancedReplicationSpecRegionConfig(apiObject, nil))
		}
	}

	return tfList
}

func flattenAdvancedReplicationSpecRegionConfigSpec(apiObject *matlas.Specs, providerName string, tfMapObjects []interface{}) []map[string]interface{} {
	if apiObject == nil {
		return nil
	}
	var tfList []map[string]interface{}

	tfMap := map[string]interface{}{}

	if len(tfMapObjects) > 0 {
		checkObject := false
		checkObject = true
		tfMapObject := tfMapObjects[0].(map[string]interface{})

		if providerName == "AWS" {
			if checkObject {
				if cast.ToInt64(apiObject.DiskIOPS) > 0 {
					if v, ok := tfMapObject["disk_iops"]; ok && v.(int) > 0 {
						tfMap["disk_iops"] = apiObject.DiskIOPS
					}
				}
				if v, ok := tfMapObject["ebs_volume_type"]; ok && v.(string) != "" {
					tfMap["ebs_volume_type"] = apiObject.EbsVolumeType
				}
			}
		}
		tfMap["node_count"] = apiObject.NodeCount
		if checkObject {
			if v, ok := tfMapObject["instance_size"]; ok && v.(string) != "" {
				tfMap["instance_size"] = apiObject.InstanceSize
				tfList = append(tfList, tfMap)
			}
		}
	} else {
		tfMap["disk_iops"] = apiObject.DiskIOPS
		tfMap["ebs_volume_type"] = apiObject.EbsVolumeType
		tfMap["node_count"] = apiObject.NodeCount
		tfMap["instance_size"] = apiObject.InstanceSize
		tfList = append(tfList, tfMap)
	}

	return tfList
}

func flattenAdvancedReplicationSpecAutoScaling(apiObject *matlas.AdvancedAutoScaling) []map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	var tfList []map[string]interface{}

	tfMap := map[string]interface{}{}
	if apiObject.DiskGB != nil {
		tfMap["disk_gb_enabled"] = apiObject.DiskGB.Enabled
	}
	if apiObject.Compute != nil {
		tfMap["compute_enabled"] = apiObject.Compute.Enabled
		tfMap["compute_scale_down_enabled"] = apiObject.Compute.ScaleDownEnabled
		tfMap["compute_min_instance_size"] = apiObject.Compute.MinInstanceSize
		tfMap["compute_max_instance_size"] = apiObject.Compute.MaxInstanceSize
	}

	tfList = append(tfList, tfMap)

	return tfList
}

func resourceClusterAdvancedRefreshFunc(ctx context.Context, name, projectID string, client *matlas.Client) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		c, resp, err := client.AdvancedClusters.Get(ctx, projectID, name)

		if err != nil && strings.Contains(err.Error(), "reset by peer") {
			return nil, "REPEATING", nil
		}

		if err != nil && c == nil && resp == nil {
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

		if c.StateName != "" {
			log.Printf("[DEBUG] status for MongoDB cluster: %s: %s", name, c.StateName)
		}

		return c, c.StateName, nil
	}
}

func replicationSpecsHashSet(v interface{}) int {
	var buf bytes.Buffer
	m := v.(map[string]interface{})
	buf.WriteString(fmt.Sprintf("%d", m["num_shards"].(int)))
	buf.WriteString(fmt.Sprintf("%+v", m["region_configs"].(*schema.Set)))
	buf.WriteString(m["zone_name"].(string))
	return HashCodeString(buf.String())
}
