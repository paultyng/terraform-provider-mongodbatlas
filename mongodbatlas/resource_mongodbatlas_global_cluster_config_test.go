package mongodbatlas

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	matlas "go.mongodb.org/atlas/mongodbatlas"
)

func TestAccResourceMongoDBAtlasGlobalCluster_basic(t *testing.T) {
	var (
		globalConfig matlas.GlobalCluster
		resourceName = "mongodbatlas_global_cluster_config.config"
		projectID    = os.Getenv("MONGODB_ATLAS_PROJECT_ID")
		name         = fmt.Sprintf("test-acc-global-%s", acctest.RandString(10))
	)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckMongoDBAtlasGlobalClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccMongoDBAtlasGlobalClusterConfig(projectID, name, "false"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMongoDBAtlasGlobalClusterExists(resourceName, &globalConfig),
					resource.TestCheckResourceAttrSet(resourceName, "managed_namespaces.#"),
					resource.TestCheckResourceAttrSet(resourceName, "custom_zone_mappings.#"),
					resource.TestCheckResourceAttrSet(resourceName, "custom_zone_mapping.%"),
					resource.TestCheckResourceAttrSet(resourceName, "custom_zone_mapping.CA"),
					resource.TestCheckResourceAttr(resourceName, "cluster_name", name),
					resource.TestCheckResourceAttr(resourceName, "project_id", projectID),
					testAccCheckMongoDBAtlasGlobalClusterAttributes(&globalConfig, 1),
				),
			},
		},
	})
}

func TestAccResourceMongoDBAtlasGlobalCluster_WithAWSCluster(t *testing.T) {
	var (
		globalConfig matlas.GlobalCluster
		resourceName = "mongodbatlas_global_cluster_config.config"
		projectID    = os.Getenv("MONGODB_ATLAS_PROJECT_ID")
		name         = fmt.Sprintf("test-acc-global-%s", acctest.RandString(10))
	)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckMongoDBAtlasGlobalClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccMongoDBAtlasGlobalClusterWithAWSClusterConfig(projectID, name, "true"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMongoDBAtlasGlobalClusterExists(resourceName, &globalConfig),
					resource.TestCheckResourceAttrSet(resourceName, "managed_namespaces.#"),
					resource.TestCheckResourceAttrSet(resourceName, "custom_zone_mappings.#"),
					resource.TestCheckResourceAttrSet(resourceName, "custom_zone_mapping.%"),
					resource.TestCheckResourceAttrSet(resourceName, "custom_zone_mapping.CA"),
					resource.TestCheckResourceAttr(resourceName, "cluster_name", name),
					resource.TestCheckResourceAttr(resourceName, "project_id", projectID),
					testAccCheckMongoDBAtlasGlobalClusterAttributes(&globalConfig, 1),
				),
			},
		},
	})
}

func TestAccResourceMongoDBAtlasGlobalCluster_importBasic(t *testing.T) {
	var (
		resourceName = "mongodbatlas_global_cluster_config.config"
		projectID    = os.Getenv("MONGODB_ATLAS_PROJECT_ID")
		name         = fmt.Sprintf("test-acc-global-%s", acctest.RandString(10))
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckMongoDBAtlasProjectIPAccessListDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccMongoDBAtlasGlobalClusterConfig(projectID, name, "false"),
			},
			{
				ResourceName:            resourceName,
				ImportStateIdFunc:       testAccCheckMongoDBAtlasGlobalClusterImportStateIDFunc(resourceName),
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"custom_zone_mappings"},
			},
		},
	})
}

func testAccCheckMongoDBAtlasGlobalClusterExists(resourceName string, globalConfig *matlas.GlobalCluster) resource.TestCheckFunc {
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

		globalConfigResp, _, err := conn.GlobalClusters.Get(context.Background(), ids["project_id"], ids["cluster_name"])
		if err == nil {
			*globalConfig = *globalConfigResp

			if len(globalConfig.CustomZoneMapping) > 0 || len(globalConfig.ManagedNamespaces) > 0 {
				return nil
			}
		}

		return fmt.Errorf("global config for cluster(%s) does not exist", ids["cluster_name"])
	}
}

func testAccCheckMongoDBAtlasGlobalClusterImportStateIDFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("not found: %s", resourceName)
		}

		ids := decodeStateID(rs.Primary.ID)

		return fmt.Sprintf("%s-%s", ids["project_id"], ids["cluster_name"]), nil
	}
}

func testAccCheckMongoDBAtlasGlobalClusterAttributes(globalCluster *matlas.GlobalCluster, managedNamespacesCount int) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if len(globalCluster.ManagedNamespaces) != managedNamespacesCount {
			return fmt.Errorf("bad managed namespaces: %v", globalCluster.ManagedNamespaces)
		}

		return nil
	}
}

func testAccCheckMongoDBAtlasGlobalClusterDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*MongoDBClient).Atlas

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "mongodbatlas_global_cluster_config" {
			continue
		}

		// Try to find the cluster
		globalConfig, _, err := conn.GlobalClusters.Get(context.Background(), rs.Primary.Attributes["project_id"], rs.Primary.Attributes["cluster_name"])
		if err != nil {
			if strings.Contains(err.Error(), fmt.Sprintf("No cluster named %s exists in group %s", rs.Primary.Attributes["cluster_name"], rs.Primary.Attributes["project_id"])) {
				return nil
			}

			return err
		}

		if len(globalConfig.CustomZoneMapping) > 0 || len(globalConfig.ManagedNamespaces) > 0 {
			return fmt.Errorf("global cluster configuration for cluster(%s) still exists", rs.Primary.Attributes["cluster_name"])
		}
	}

	return nil
}

func testAccMongoDBAtlasGlobalClusterConfig(projectID, name, backupEnabled string) string {
	return fmt.Sprintf(`
		resource "mongodbatlas_cluster" "test" {
			project_id              = "%s"
			name                    = "%s"
			disk_size_gb            = 80
			backup_enabled          = "%s"
			provider_backup_enabled = true
			cluster_type            = "GEOSHARDED"

			// Provider Settings "block"
			provider_name               = "AWS"
 			provider_instance_size_name = "M30"

			replication_specs {
				zone_name  = "Zone 1"
				num_shards = 1
				regions_config {
					region_name     = "US_EAST_1"
					electable_nodes = 3
					priority        = 7
					read_only_nodes = 0
				}
			}

			replication_specs {
				zone_name  = "Zone 2"
				num_shards = 1
				regions_config {
					region_name     = "US_EAST_2"
					electable_nodes = 3
					priority        = 7
					read_only_nodes = 0
				}
			}
		}

		resource "mongodbatlas_global_cluster_config" "config" {
			project_id   = mongodbatlas_cluster.test.project_id
			cluster_name = mongodbatlas_cluster.test.name

			managed_namespaces {
				db               = "mydata"
				collection       = "publishers"
				custom_shard_key = "city"
			}

			custom_zone_mappings {
				location = "CA"
				zone     = "Zone 1"
			}
		}
	`, projectID, name, backupEnabled)
}

func testAccMongoDBAtlasGlobalClusterWithAWSClusterConfig(projectID, name, backupEnabled string) string {
	return fmt.Sprintf(`
		resource "mongodbatlas_cluster" "test" {
			project_id              = "%s"
			name                    = "%s"
			disk_size_gb            = 80
			provider_backup_enabled = %s
			cluster_type            = "GEOSHARDED"

			// Provider Settings "block"
			provider_name               = "AWS"
 			provider_instance_size_name = "M30"

			replication_specs {
				zone_name  = "Zone 1"
				num_shards = 1
				regions_config {
					region_name     = "US_EAST_1"
					electable_nodes = 3
					priority        = 7
					read_only_nodes = 0
				}
			}
		}

		resource "mongodbatlas_global_cluster_config" "config" {
			project_id   = mongodbatlas_cluster.test.project_id
			cluster_name = mongodbatlas_cluster.test.name

			managed_namespaces {
				db               = "mydata"
				collection       = "publishers"
				custom_shard_key = "city"
			}

			custom_zone_mappings {
				location = "CA"
				zone     = "Zone 1"
			}
		}
	`, projectID, name, backupEnabled)
}
