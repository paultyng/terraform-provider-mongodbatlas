package mongodbatlas

import (
	"context"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceMongoDBAtlasAdvancedCluster() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceMongoDBAtlasAdvancedClusterRead,
		Schema: map[string]*schema.Schema{
			"project_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"backup_enabled": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"bi_connector": {
				Type:     schema.TypeList,
				Computed: true,
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
				Computed: true,
			},
			"connection_strings": clusterConnectionStringsSchema(),
			"create_date": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"disk_size_gb": {
				Type:     schema.TypeFloat,
				Computed: true,
			},
			"encryption_at_rest_provider": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"labels": {
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
			},
			"mongo_db_major_version": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"mongo_db_version": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
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
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"num_shards": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"region_configs": {
							Type:     schema.TypeSet,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"analytics_specs": advancedClusterRegionConfigsSpecsSchema(),
									"auto_scaling": {
										Type:     schema.TypeList,
										Computed: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"disk_gb_enabled": {
													Type:     schema.TypeBool,
													Computed: true,
												},
												"compute_enabled": {
													Type:     schema.TypeBool,
													Computed: true,
												},
												"compute_scale_down_enabled": {
													Type:     schema.TypeBool,
													Computed: true,
												},
												"compute_min_instance_size": {
													Type:     schema.TypeString,
													Computed: true,
												},
												"compute_max_instance_size": {
													Type:     schema.TypeString,
													Computed: true,
												},
											},
										},
									},
									"backing_provider_name": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"electable_specs": advancedClusterRegionConfigsSpecsSchema(),
									"priority": {
										Type:     schema.TypeInt,
										Computed: true,
									},
									"provider_name": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"read_only_specs": advancedClusterRegionConfigsSpecsSchema(),
									"region_name": {
										Type:     schema.TypeString,
										Computed: true,
									},
								},
							},
						},
						"zone_name": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
				Set: replicationSpecsHashSet,
			},
			"root_cert_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"state_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"version_release_system": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceMongoDBAtlasAdvancedClusterRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// Get client connection.
	conn := meta.(*MongoDBClient).Atlas
	projectID := d.Get("project_id").(string)
	clusterName := d.Get("name").(string)

	cluster, resp, err := conn.AdvancedClusters.Get(ctx, projectID, clusterName)
	if err != nil {
		if resp != nil && resp.StatusCode == http.StatusNotFound {
			return nil
		}

		return diag.FromErr(fmt.Errorf(errorClusterAdvancedRead, clusterName, err))
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

	if err := d.Set("replication_specs", flattenAdvancedReplicationSpecs(cluster.ReplicationSpecs, nil)); err != nil {
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

	d.SetId(cluster.ID)

	return nil
}
