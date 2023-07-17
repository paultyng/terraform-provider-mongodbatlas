package mongodbatlas

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	atlasSDK "go.mongodb.org/atlas-sdk/v20230201002/admin"
)

// This datasource does not have a resource: we tested it manually
func dataSourceMongoDBAtlasSharedTierSnapshots() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceMongoDBAtlasSharedTierSnapshotsRead,
		Schema: map[string]*schema.Schema{
			"project_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"cluster_name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"results": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"snapshot_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"mongo_db_version": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"status": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"expiration": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"start_time": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"finish_time": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"scheduled_time": {
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

func dataSourceMongoDBAtlasSharedTierSnapshotsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*MongoDBClient).AtlasV2

	projectID := d.Get("project_id").(string)
	clusterName := d.Get("cluster_name").(string)
	snapshots, _, err := conn.SharedTierSnapshotsApi.ListSharedClusterBackups(ctx, projectID, clusterName).Execute()
	if err != nil {
		return diag.FromErr(fmt.Errorf("error getting shard-tier snapshots for cluster '%s': %w", clusterName, err))
	}

	if err := d.Set("results", flattenShardTierSnapshots(snapshots.Results)); err != nil {
		return diag.FromErr(fmt.Errorf("error setting `results`: %w", err))
	}

	if err := d.Set("total_count", snapshots.TotalCount); err != nil {
		return diag.FromErr(fmt.Errorf("error setting `total_count`: %w", err))
	}

	return nil
}

func flattenShardTierSnapshots(shardTierSnapshots []atlasSDK.BackupTenantSnapshot) []map[string]interface{} {
	if len(shardTierSnapshots) == 0 {
		return nil
	}

	results := make([]map[string]interface{}, len(shardTierSnapshots))
	for k, shardTierSnapshot := range shardTierSnapshots {
		results[k] = map[string]interface{}{
			"snapshot_id":    shardTierSnapshot.Id,
			"start_time":     shardTierSnapshot.StartTime,
			"finish_time":    shardTierSnapshot.FinishTime,
			"scheduled_time": shardTierSnapshot.ScheduledTime,
			"expiration":     shardTierSnapshot.Expiration,
			"mongod_version": shardTierSnapshot.MongoDBVersion,
			"status":         shardTierSnapshot.Status,
		}
	}

	return results
}
