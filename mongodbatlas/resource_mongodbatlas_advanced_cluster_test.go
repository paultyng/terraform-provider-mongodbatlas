package mongodbatlas

import (
	"context"
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	matlas "go.mongodb.org/atlas/mongodbatlas"
)

func TestAccResourceMongoDBAtlasAdvancedCluster_basicTenant(t *testing.T) {
	var (
		cluster      matlas.AdvancedCluster
		resourceName = "mongodbatlas_advanced_cluster.test"
		projectID    = os.Getenv("MONGODB_ATLAS_PROJECT_ID")
		rName        = acctest.RandomWithPrefix("test-acc")
		rNameUpdated = acctest.RandomWithPrefix("test-acc")
	)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckMongoDBAtlasAdvancedClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccMongoDBAtlasAdvancedClusterConfigTenant(projectID, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMongoDBAtlasAdvancedClusterExists(resourceName, &cluster),
					testAccCheckMongoDBAtlasAdvancedClusterAttributes(&cluster, rName),
					resource.TestCheckResourceAttrSet(resourceName, "project_id"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttrSet(resourceName, "replication_specs.#"),
					resource.TestCheckResourceAttrSet(resourceName, "replication_specs.0.region_configs.#"),
				),
			},
			{
				Config: testAccMongoDBAtlasAdvancedClusterConfigTenant(projectID, rNameUpdated),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMongoDBAtlasAdvancedClusterExists(resourceName, &cluster),
					testAccCheckMongoDBAtlasAdvancedClusterAttributes(&cluster, rNameUpdated),
					resource.TestCheckResourceAttrSet(resourceName, "project_id"),
					resource.TestCheckResourceAttr(resourceName, "name", rNameUpdated),
					resource.TestCheckResourceAttrSet(resourceName, "replication_specs.#"),
					resource.TestCheckResourceAttrSet(resourceName, "replication_specs.0.region_configs.#"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateIdFunc: testAccCheckMongoDBAtlasClusterImportStateIDFunc(resourceName),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccResourceMongoDBAtlasAdvancedCluster_multicloud(t *testing.T) {
	var (
		cluster      matlas.AdvancedCluster
		resourceName = "mongodbatlas_advanced_cluster.test"
		projectID    = os.Getenv("MONGODB_ATLAS_PROJECT_ID")
		rName        = acctest.RandomWithPrefix("test-acc")
		rNameUpdated = acctest.RandomWithPrefix("test-acc")
	)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckMongoDBAtlasAdvancedClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccMongoDBAtlasAdvancedClusterConfigMultiCloud(projectID, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMongoDBAtlasAdvancedClusterExists(resourceName, &cluster),
					testAccCheckMongoDBAtlasAdvancedClusterAttributes(&cluster, rName),
					resource.TestCheckResourceAttrSet(resourceName, "project_id"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttrSet(resourceName, "replication_specs.#"),
					resource.TestCheckResourceAttrSet(resourceName, "replication_specs.0.region_configs.#"),
				),
			},
			{
				Config: testAccMongoDBAtlasAdvancedClusterConfigMultiCloud(projectID, rNameUpdated),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMongoDBAtlasAdvancedClusterExists(resourceName, &cluster),
					testAccCheckMongoDBAtlasAdvancedClusterAttributes(&cluster, rNameUpdated),
					resource.TestCheckResourceAttrSet(resourceName, "project_id"),
					resource.TestCheckResourceAttr(resourceName, "name", rNameUpdated),
					resource.TestCheckResourceAttrSet(resourceName, "replication_specs.#"),
					resource.TestCheckResourceAttrSet(resourceName, "replication_specs.0.region_configs.#"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportStateIdFunc:       testAccCheckMongoDBAtlasClusterImportStateIDFunc(resourceName),
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"replication_specs"},
			},
		},
	})
}

func testAccCheckMongoDBAtlasAdvancedClusterExists(resourceName string, cluster *matlas.AdvancedCluster) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*MongoDBClient).Atlas

		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no ID is set")
		}

		ids := decodeStateID(rs.Primary.ID)

		log.Printf("[DEBUG] projectID: %s, name %s", ids["project_id"], ids["cluster_name"])

		if clusterResp, _, err := conn.AdvancedClusters.Get(context.Background(), ids["project_id"], ids["cluster_name"]); err == nil {
			*cluster = *clusterResp
			return nil
		}

		return fmt.Errorf("cluster(%s:%s) does not exist", rs.Primary.Attributes["project_id"], rs.Primary.ID)
	}
}

func testAccCheckMongoDBAtlasAdvancedClusterAttributes(cluster *matlas.AdvancedCluster, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if cluster.Name != name {
			return fmt.Errorf("bad name: %s", cluster.Name)
		}

		return nil
	}
}

func testAccCheckMongoDBAtlasAdvancedClusterDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*MongoDBClient).Atlas

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "mongodbatlas_advanced_cluster" {
			continue
		}

		// Try to find the cluster
		_, _, err := conn.AdvancedClusters.Get(context.Background(), rs.Primary.Attributes["project_id"], rs.Primary.Attributes["cluster_name"])

		if err == nil {
			return fmt.Errorf("cluster (%s:%s) still exists", rs.Primary.Attributes["cluster_name"], rs.Primary.ID)
		}
	}

	return nil
}

func testAccMongoDBAtlasAdvancedClusterConfigTenant(projectID, name string) string {
	return fmt.Sprintf(`
resource "mongodbatlas_advanced_cluster" "test" {
  project_id   = %[1]q
  name         = %[2]q
  cluster_type = "REPLICASET"

  replication_specs {
    region_configs {
      electable_specs {
        instance_size = "M5"
      }
      provider_name         = "TENANT"
      backing_provider_name = "AWS"
      region_name           = "US_EAST_1"
      priority              = 7
    }
  }
}
	`, projectID, name)
}

func testAccMongoDBAtlasAdvancedClusterConfigMultiCloud(projectID, name string) string {
	return fmt.Sprintf(`
resource "mongodbatlas_advanced_cluster" "test" {
  project_id   = %[1]q
  name         = %[2]q
  cluster_type = "REPLICASET"

  replication_specs {
    region_configs {
      electable_specs {
        instance_size = "M10"
        node_count    = 3
      }
      analytics_specs {
        instance_size = "M10"
        node_count    = 1
      }
      provider_name = "AWS"
      priority      = 7
      region_name   = "US_EAST_1"
    }
    region_configs {
      electable_specs {
        instance_size = "M10"
        node_count    = 2
      }
      provider_name = "GCP"
      priority      = 6
      region_name   = "NORTH_AMERICA_NORTHEAST_1"
    }
  }
}
	`, projectID, name)
}
