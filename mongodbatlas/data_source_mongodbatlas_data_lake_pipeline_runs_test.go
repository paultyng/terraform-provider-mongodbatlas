package mongodbatlas

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccBackupDSDataLakePipelineRuns_basic(t *testing.T) {
	// testCheckDataLakePipelineRuns(t)
	var (
		dataSourceName = "data.mongodbatlas_data_lake_pipeline_runs.test"
		// projectID      = os.Getenv("MONGODB_ATLAS_PROJECT_ID")
		// pipelineName   = os.Getenv("MONGODB_ATLAS_DATA_LAKE_PIPELINE_NAME")

		projectID    = "63f4d4a47baeac59406dc131"
		pipelineName = "sample_guides.planets"
		// runID        = "6467558d70fc1a140034adf0"
	)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccMongoDBAtlasDataLakeDataSourcePipelineRunsConfig(projectID, pipelineName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(dataSourceName, "project_id"),
					resource.TestCheckResourceAttr(dataSourceName, "name", pipelineName),
					resource.TestCheckResourceAttrSet(dataSourceName, "results.#"),
				),
			},
		},
	})
}

func testAccMongoDBAtlasDataLakeDataSourcePipelineRunsConfig(projectID, pipelineName string) string {
	return fmt.Sprintf(`

data "mongodbatlas_data_lake_pipeline_runs" "test" {
  project_id           = %[1]q
  name                 = %[2]q
}
	`, projectID, pipelineName)
}
