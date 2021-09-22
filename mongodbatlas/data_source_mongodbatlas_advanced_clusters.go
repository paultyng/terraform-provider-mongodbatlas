package mongodbatlas

import (
	"context"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	matlas "go.mongodb.org/atlas/mongodbatlas"
)

func dataSourceMongoDBAtlasAdvancedClusters() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceMongoDBAtlasAdvancedClustersRead,
		Schema: map[string]*schema.Schema{
			"project_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"results": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
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
							Computed: true,
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
				},
			},
		},
	}
}

func dataSourceMongoDBAtlasAdvancedClustersRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// Get client connection.
	conn := meta.(*MongoDBClient).Atlas
	projectID := d.Get("project_id").(string)
	d.SetId(resource.UniqueId())

	clusters, resp, err := conn.AdvancedClusters.List(ctx, projectID, nil)
	if err != nil {
		if resp != nil && resp.StatusCode == http.StatusNotFound {
			return nil
		}

		return diag.FromErr(fmt.Errorf("error reading advanced cluster list for project(%s): %s", projectID, err))
	}

	if err := d.Set("results", flattenAdvancedClusters(ctx, d, conn, clusters.Results)); err != nil {
		return diag.FromErr(fmt.Errorf(errorClusterAdvancedSetting, "results", d.Id(), err))
	}

	return nil
}

func flattenAdvancedClusters(ctx context.Context, d *schema.ResourceData, conn *matlas.Client, clusters []*matlas.AdvancedCluster) []map[string]interface{} {
	results := make([]map[string]interface{}, 0)

	for i := range clusters {
		result := map[string]interface{}{
			"backup_enabled":              clusters[i].BackupEnabled,
			"bi_connector":                flattenBiConnectorConfig(clusters[i].BiConnector),
			"cluster_type":                clusters[i].ClusterType,
			"create_date":                 clusters[i].CreateDate,
			"connection_strings":          flattenConnectionStrings(clusters[i].ConnectionStrings),
			"disk_size_gb":                clusters[i].DiskSizeGB,
			"encryption_at_rest_provider": clusters[i].EncryptionAtRestProvider,
			"labels":                      flattenLabels(clusters[i].Labels),
			"mongo_db_major_version":      clusters[i].MongoDBMajorVersion,
			"mongo_db_version":            clusters[i].MongoDBVersion,
			"name":                        clusters[i].Name,
			"paused":                      clusters[i].Paused,
			"pit_enabled":                 clusters[i].PitEnabled,
			"replication_specs":           flattenAdvancedReplicationSpecs(clusters[i].ReplicationSpecs, nil),
			"root_cert_type":              clusters[i].RootCertType,
			"state_name":                  clusters[i].StateName,
			"version_release_system":      clusters[i].VersionReleaseSystem,
		}
		results = append(results, result)
	}

	return results
}
