package advancedclustertpf_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
	"github.com/mongodb/terraform-provider-mongodbatlas/internal/testutil/acc"
)

func TestAccAdvancedCluster_moveBasic(t *testing.T) {
	var (
		projectID   = acc.ProjectIDExecution(t)
		clusterName = acc.RandomClusterName()
	)
	resource.ParallelTest(t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_8_0),
		},
		ProtoV6ProviderFactories: acc.TestAccProviderV6Factories,
		CheckDestroy:             acc.CheckDestroyCluster,
		Steps: []resource.TestStep{
			{
				Config: configMoveFirst(projectID, clusterName),
			},
			{
				Config: configMoveSecond(projectID, clusterName),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PostApplyPreRefresh: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
		},
	})
}

func TestAccAdvancedCluster_moveInvalid(t *testing.T) {
	var (
		projectID   = acc.ProjectIDExecution(t)
		clusterName = acc.RandomClusterName()
	)
	resource.ParallelTest(t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_8_0),
		},
		ProtoV6ProviderFactories: acc.TestAccProviderV6Factories,
		CheckDestroy:             acc.CheckDestroyCluster,
		Steps: []resource.TestStep{
			{
				Config: configMoveFirstInvalid(projectID, clusterName),
			},
			{
				Config:      configMoveSecondInvalid(projectID, clusterName),
				ExpectError: regexp.MustCompile("Unable to Move Resource State"),
			},
			{
				Config: configMoveFirstInvalid(projectID, clusterName),
			},
		},
	})
}

func configMoveFirst(projectID, clusterName string) string {
	return fmt.Sprintf(`
		resource "mongodbatlas_cluster" "old" {
			project_id = %[1]q
			name = %[2]q
			disk_size_gb = 10
			cluster_type                = "REPLICASET"
			provider_name               = "AWS"
			provider_instance_size_name = "M10"
			replication_specs {
				num_shards = 1
				regions_config {
					region_name     = "US_EAST_1"
					electable_nodes = 3
					priority        = 7
				}
			}
		}
	`, projectID, clusterName)
}

func configMoveFirstInvalid(projectID, clusterName string) string {
	return fmt.Sprintf(`
		resource "mongodbatlas_database_user" "old" {
			project_id         = %[1]q
			username           = %[2]q # use cluster name as username
			password           = "test-acc-password"
			auth_database_name = "admin"
			roles {
				role_name     = "atlasAdmin"
				database_name = "admin"
			}
		}
	`, projectID, clusterName)
}

func configMoveSecond(projectID, clusterName string) string {
	return `
		moved {
			from = mongodbatlas_cluster.old
			to   = mongodbatlas_advanced_cluster.test
		}
	` + configBasic(projectID, clusterName, "")
}

func configMoveSecondInvalid(projectID, clusterName string) string {
	return `
		moved {
			from = mongodbatlas_database_user.old
			to   = mongodbatlas_advanced_cluster.test
		}
	` + configBasic(projectID, clusterName, "")
}
