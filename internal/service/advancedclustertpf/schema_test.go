package advancedclustertpf_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/mongodb/terraform-provider-mongodbatlas/internal/testutil/acc"
)

func TestAdvancedCluster_ValidationErrors(t *testing.T) {
	const (
		projectID   = "111111111111111111111111"
		clusterName = "test"
	)
	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: acc.TestAccProviderV6Factories,
		Steps: []resource.TestStep{
			{
				Config:      acc.ConvertAdvancedClusterToSchemaV2(t, true, invalidRegionConfigsPriorities),
				ExpectError: regexp.MustCompile("priority values in region_configs must be in descending order"),
			},
			{
				Config:      configBasic(projectID, clusterName, "mongo_db_major_version = \"8a\""),
				ExpectError: regexp.MustCompile("Invalid Attribute Value Match"),
			},
			{
				Config:      configBasic(projectID, clusterName, "advanced_configuration = {oplog_size_mb = -1}"),
				ExpectError: regexp.MustCompile("Invalid Attribute Value"),
			},
		},
	})
}

func TestAdvancedCluster_PlanModifierErrors(t *testing.T) {
	const (
		projectID   = "111111111111111111111111"
		clusterName = "test"
	)
	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: acc.TestAccProviderV6Factories,
		Steps: []resource.TestStep{
			{
				Config:      configBasic(projectID, clusterName, "advanced_configuration = { change_stream_options_pre_and_post_images_expire_after_seconds = 100 }\nmongo_db_major_version=\"6\""),
				ExpectError: regexp.MustCompile("`advanced_configuration.change_stream_options_pre_and_post_images_expire_after_seconds` can only be configured if the mongo_db_major_version is 7.0 or higher"),
			},
			{
				Config:      configBasic(projectID, clusterName, "advanced_configuration = { default_max_time_ms = 100 }\nmongo_db_major_version=\"6\""),
				ExpectError: regexp.MustCompile("`advanced_configuration.default_max_time_ms` can only be configured if the mongo_db_major_version is 8.0 or higher"),
			},
			{
				Config:      configBasic(projectID, clusterName, "advanced_configuration = { fail_index_key_too_long = true }"),
				ExpectError: regexp.MustCompile("`advanced_configuration.fail_index_key_too_long` can only be configured if the mongo_db_major_version is 4.4 or lower"),
			},
			{
				Config:      configBasic(projectID, clusterName, "accept_data_risks_and_force_replica_set_reconfig = \"2006-01-02T15:04:05Z\""),
				ExpectError: regexp.MustCompile("Update only attribute set on create: accept_data_risks_and_force_replica_set_reconfig"),
			},
		},
	})
}

func TestAdvancedCluster_PlanModifierValid(t *testing.T) {
	var (
		projectID   = "111111111111111111111111"
		clusterName = "test"
	)
	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: acc.TestAccProviderV6Factories,
		Steps: []resource.TestStep{
			{
				Config:             configBasic(projectID, clusterName, "advanced_configuration = { change_stream_options_pre_and_post_images_expire_after_seconds = 100 }\nmongo_db_major_version=\"7\""),
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			},
			{
				Config:             configBasic(projectID, clusterName, "advanced_configuration = { change_stream_options_pre_and_post_images_expire_after_seconds = 100 }\nmongo_db_major_version=\"7.0\""),
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			},
			{
				Config:             configBasic(projectID, clusterName, "advanced_configuration = { change_stream_options_pre_and_post_images_expire_after_seconds = 100 }"), // mongo_db_major_version is not set should also work
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			},
			{
				Config:             configBasic(projectID, clusterName, "advanced_configuration = { fail_index_key_too_long = true }\nmongo_db_major_version=\"4\""),
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func configBasic(projectID, clusterName, extra string) string {
	return fmt.Sprintf(`
		resource "mongodbatlas_advanced_cluster" "test" {
			timeouts = {
				create = "2000s"
			}
			project_id = %[1]q
			name = %[2]q
			cluster_type = "REPLICASET"
			replication_specs = [{
				region_configs = [{
					priority        = 7
					provider_name = "AWS"
					region_name     = "US_EAST_1"
					auto_scaling = {
						compute_scale_down_enabled = false # necessary to have similar SDKv2 request
						compute_enabled = false # necessary to have similar SDKv2 request
						disk_gb_enabled = true
					}
					electable_specs = {
						node_count = 3
						instance_size = "M10"
						disk_size_gb = 10
					}
				}]
			}]
			%[3]s
		}
	`, projectID, clusterName, extra)
}

var invalidRegionConfigsPriorities = `
resource "mongodbatlas_advanced_cluster" "test" {
	project_id     = "111111111111111111111111"
	name           = "test-acc-tf-c-2670522663699021050"
	cluster_type   = "REPLICASET"
	backup_enabled = false

	replication_specs {
		region_configs {
			provider_name = "AWS"
			priority      = 6
			region_name   = "US_WEST_2"
			electable_specs {
				node_count    = 1
				instance_size = "M10"
			}
		}
		region_configs {
			provider_name = "AWS"
			priority      = 7
			region_name   = "US_EAST_1"
			electable_specs {
				node_count    = 2
				instance_size = "M10"
			}
		}
	}
}
`
