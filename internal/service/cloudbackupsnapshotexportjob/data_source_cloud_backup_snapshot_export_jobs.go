package cloudbackupsnapshotexportjob

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mongodb/terraform-provider-mongodbatlas/internal/config"
	"go.mongodb.org/atlas-sdk/v20240530002/admin"
)

func PluralDataSource() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceMongoDBAtlasCloudBackupSnapshotsExportJobsRead,
		Schema: map[string]*schema.Schema{
			"project_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"cluster_name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"page_num": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"items_per_page": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"results": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"export_job_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"snapshot_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"custom_data": {
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
								}},
						},
						"components": {
							Type:     schema.TypeSet,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"export_id": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"replica_set_name": {
										Type:     schema.TypeString,
										Computed: true,
									},
								},
							},
						},
						"created_at": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"err_msg": { //TODO: this is no longer returned by the API, could be removed
							Type:     schema.TypeString,
							Computed: true,
						},
						"export_bucket_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"export_status_exported_collections": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"export_status_total_collections": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"finished_at": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"prefix": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"state": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"total_count": {
				Type:     schema.TypeInt,
				Computed: true,
			},
		},
	}
}

func dataSourceMongoDBAtlasCloudBackupSnapshotsExportJobsRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	connV2 := meta.(*config.MongoDBClient).AtlasV2

	projectID := d.Get("project_id").(string)
	clusterName := d.Get("cluster_name").(string)

	jobs, _, err := connV2.CloudBackupsApi.ListBackupExportJobs(ctx, projectID, clusterName).Execute()
	if err != nil {
		return diag.Errorf("error getting CloudProviderSnapshotExportJobs information: %s", err)
	}

	if err := d.Set("results", flattenCloudBackupSnapshotExportJobs(jobs.Results)); err != nil {
		return diag.Errorf("error setting `results`: %s", err)
	}

	if err := d.Set("total_count", jobs.TotalCount); err != nil {
		return diag.Errorf("error setting `total_count`: %s", err)
	}

	d.SetId(id.UniqueId())

	return nil
}

func flattenCloudBackupSnapshotExportJobs(jobs *[]admin.DiskBackupExportJob) []map[string]any {
	var results []map[string]any

	if len(*jobs) == 0 {
		return results
	}

	results = make([]map[string]any, len(*jobs))

	for k, job := range *jobs {
		results[k] = map[string]any{
			"export_job_id":                      job.Id,
			"created_at":                         job.CreatedAt,
			"components":                         flattenExportJobsComponents(job.Components),
			"custom_data":                        flattenExportJobsCustomData(job.CustomData),
			"export_bucket_id":                   job.ExportBucketId,
			"export_status_exported_collections": job.ExportStatus.ExportedCollections,
			"export_status_total_collections":    job.ExportStatus.TotalCollections,
			"finished_at":                        job.FinishedAt,
			"prefix":                             job.Prefix,
			"snapshot_id":                        job.SnapshotId,
			"state":                              job.State,
		}
	}

	return results
}
