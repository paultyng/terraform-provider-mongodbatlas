package datalakepipeline_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/mongodb/terraform-provider-mongodbatlas/internal/testutil/acc"
	matlas "go.mongodb.org/atlas/mongodbatlas"
)

func TestAccDataLakeDS_basic(t *testing.T) {
	var (
		pipeline     matlas.DataLakePipeline
		resourceName = "mongodbatlas_data_lake_pipeline.test"
		clusterName  = acctest.RandomWithPrefix("test-acc-index")
		name         = acctest.RandomWithPrefix("test-acc-index")
		orgID        = os.Getenv("MONGODB_ATLAS_ORG_ID")
		projectName  = acctest.RandomWithPrefix("test-acc")
	)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acc.PreCheckBasic(t) },
		ProtoV6ProviderFactories: acc.TestAccProviderV6Factories,
		CheckDestroy:             checkDestroy,
		Steps: []resource.TestStep{
			{
				Config: configDS(orgID, projectName, clusterName, name),
				Check: resource.ComposeTestCheckFunc(
					checkExists(resourceName, &pipeline),
					resource.TestCheckResourceAttrSet(resourceName, "project_id"),
					resource.TestCheckResourceAttr(resourceName, "name", name),
					resource.TestCheckResourceAttr(resourceName, "state", "ACTIVE"),
				),
			},
		},
	})
}

func configDS(orgID, projectName, clusterName, pipelineName string) string {
	return fmt.Sprintf(`
		resource "mongodbatlas_project" "project" {
			org_id = %[1]q
			name   = %[2]q
		}

		resource "mongodbatlas_advanced_cluster" "aws_conf" {
			project_id   = mongodbatlas_project.project.id
			name         = %[3]q
			cluster_type = "REPLICASET"
		
			replication_specs {
			region_configs {
				electable_specs {
				instance_size = "M10"
				node_count    = 3
				}
				provider_name = "AWS"
				priority      = 7
				region_name   = "EU_WEST_1"
			}
			}
			backup_enabled               = true
		}

		resource "mongodbatlas_data_lake_pipeline" "test" {
			project_id       = mongodbatlas_project.project.id
			name			 = %[4]q
			sink {
				type = "DLS"
				partition_fields {
						field_name = "access"
						order = 0
				}
			}	
	
			source {
				type = "ON_DEMAND_CPS"
				cluster_name = mongodbatlas_advanced_cluster.aws_conf.name
				database_name = "sample_airbnb"
				collection_name = "listingsAndReviews"
			}

			transformations {
				field = "test"
				type =  "EXCLUDE"
			}
		}

		data "mongodbatlas_data_lake_pipeline" "testDataSource" {
			project_id       = mongodbatlas_data_lake_pipeline.test.project_id
			name			 = mongodbatlas_data_lake_pipeline.test.name	
		}
	`, orgID, projectName, clusterName, pipelineName)
}
