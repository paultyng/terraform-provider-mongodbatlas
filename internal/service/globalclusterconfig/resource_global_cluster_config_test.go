package globalclusterconfig_test

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/mongodb/terraform-provider-mongodbatlas/internal/common/conversion"
	"github.com/mongodb/terraform-provider-mongodbatlas/internal/testutil/acc"
	matlas "go.mongodb.org/atlas/mongodbatlas"
)

func TestAccClusterRSGlobalCluster_basic(t *testing.T) {
	var (
		globalConfig matlas.GlobalCluster
		resourceName = "mongodbatlas_global_cluster_config.config"
		name         = fmt.Sprintf("test-acc-global-%s", acctest.RandString(10))
		orgID        = os.Getenv("MONGODB_ATLAS_ORG_ID")
		projectName  = acctest.RandomWithPrefix("test-acc")
	)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acc.PreCheckBasic(t) },
		ProtoV6ProviderFactories: acc.TestAccProviderV6Factories,
		CheckDestroy:             testAccCheckMongoDBAtlasGlobalClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccMongoDBAtlasGlobalClusterConfig(orgID, projectName, name, "false", "false", "false"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMongoDBAtlasGlobalClusterExists(resourceName, &globalConfig),
					resource.TestCheckResourceAttrSet(resourceName, "managed_namespaces.#"),
					resource.TestCheckResourceAttrSet(resourceName, "custom_zone_mappings.#"),
					resource.TestCheckResourceAttrSet(resourceName, "custom_zone_mapping.%"),
					resource.TestCheckResourceAttrSet(resourceName, "custom_zone_mapping.CA"),
					resource.TestCheckResourceAttrSet(resourceName, "project_id"),
					resource.TestCheckResourceAttr(resourceName, "cluster_name", name),
					resource.TestCheckResourceAttr(resourceName, "managed_namespaces.0.is_custom_shard_key_hashed", "false"),
					resource.TestCheckResourceAttr(resourceName, "managed_namespaces.0.is_shard_key_unique", "false"),
					testAccCheckMongoDBAtlasGlobalClusterAttributes(&globalConfig, 1),
				),
			},
			{
				Config: testAccMongoDBAtlasGlobalClusterConfig(orgID, projectName, name, "false", "true", "false"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMongoDBAtlasGlobalClusterExists(resourceName, &globalConfig),
					resource.TestCheckResourceAttr(resourceName, "managed_namespaces.0.is_custom_shard_key_hashed", "true"),
					resource.TestCheckResourceAttr(resourceName, "managed_namespaces.0.is_shard_key_unique", "false"),
					testAccCheckMongoDBAtlasGlobalClusterAttributes(&globalConfig, 1),
				),
			},
			{
				Config: testAccMongoDBAtlasGlobalClusterConfig(orgID, projectName, name, "false", "false", "true"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMongoDBAtlasGlobalClusterExists(resourceName, &globalConfig),
					resource.TestCheckResourceAttr(resourceName, "managed_namespaces.0.is_custom_shard_key_hashed", "false"),
					resource.TestCheckResourceAttr(resourceName, "managed_namespaces.0.is_shard_key_unique", "true"),
					testAccCheckMongoDBAtlasGlobalClusterAttributes(&globalConfig, 1),
				),
			},
		},
	})
}

func TestAccClusterRSGlobalCluster_WithAWSCluster(t *testing.T) {
	var (
		globalConfig matlas.GlobalCluster
		resourceName = "mongodbatlas_global_cluster_config.config"
		name         = fmt.Sprintf("test-acc-global-%s", acctest.RandString(10))
		orgID        = os.Getenv("MONGODB_ATLAS_ORG_ID")
		projectName  = acctest.RandomWithPrefix("test-acc")
	)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acc.PreCheckBasic(t) },
		ProtoV6ProviderFactories: acc.TestAccProviderV6Factories,
		CheckDestroy:             testAccCheckMongoDBAtlasGlobalClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccMongoDBAtlasGlobalClusterWithAWSClusterConfig(orgID, projectName, name, "true"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMongoDBAtlasGlobalClusterExists(resourceName, &globalConfig),
					resource.TestCheckResourceAttrSet(resourceName, "managed_namespaces.#"),
					resource.TestCheckResourceAttrSet(resourceName, "custom_zone_mappings.#"),
					resource.TestCheckResourceAttrSet(resourceName, "custom_zone_mapping.%"),
					resource.TestCheckResourceAttrSet(resourceName, "custom_zone_mapping.CA"),
					resource.TestCheckResourceAttrSet(resourceName, "project_id"),
					resource.TestCheckResourceAttr(resourceName, "cluster_name", name),
					testAccCheckMongoDBAtlasGlobalClusterAttributes(&globalConfig, 1),
				),
			},
		},
	})
}

func TestAccClusterRSGlobalCluster_importBasic(t *testing.T) {
	var (
		resourceName = "mongodbatlas_global_cluster_config.config"
		name         = fmt.Sprintf("test-acc-global-%s", acctest.RandString(10))
		orgID        = os.Getenv("MONGODB_ATLAS_ORG_ID")
		projectName  = acctest.RandomWithPrefix("test-acc")
	)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acc.PreCheckBasic(t) },
		ProtoV6ProviderFactories: acc.TestAccProviderV6Factories,
		CheckDestroy:             testAccCheckMongoDBAtlasGlobalClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccMongoDBAtlasGlobalClusterConfig(orgID, projectName, name, "false", "false", "false"),
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

func TestAccClusterRSGlobalCluster_database(t *testing.T) {
	var (
		globalConfig matlas.GlobalCluster
		resourceName = "mongodbatlas_global_cluster_config.test"
		name         = acctest.RandomWithPrefix("test-acc-global")
		orgID        = os.Getenv("MONGODB_ATLAS_ORG_ID")
		projectName  = acctest.RandomWithPrefix("test-acc")
	)

	customZone := `
  custom_zone_mappings {
    location = "US"
    zone     = "US"
  }
  custom_zone_mappings {
    location = "IE"
    zone     = "EU"
  }
  custom_zone_mappings {
    location = "DE"
    zone     = "DE"
  }`
	customZoneUpdated := `
  custom_zone_mappings {
    location = "US"
    zone     = "US"
  }
  custom_zone_mappings {
    location = "IE"
    zone     = "EU"
  }
  custom_zone_mappings {
    location = "DE"
    zone     = "DE"
  }
  custom_zone_mappings {
    location = "JP"
    zone     = "JP"
  }`

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acc.PreCheckBasic(t) },
		ProtoV6ProviderFactories: acc.TestAccProviderV6Factories,
		CheckDestroy:             testAccCheckMongoDBAtlasGlobalClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccMongoDBAtlasGlobalClusterWithDBConfig(orgID, projectName, name, "false", customZone),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMongoDBAtlasGlobalClusterExists(resourceName, &globalConfig),
					resource.TestCheckResourceAttrSet(resourceName, "managed_namespaces.#"),
					resource.TestCheckResourceAttrSet(resourceName, "custom_zone_mappings.#"),
					resource.TestCheckResourceAttrSet(resourceName, "custom_zone_mapping.%"),
					resource.TestCheckResourceAttrSet(resourceName, "custom_zone_mapping.US"),
					resource.TestCheckResourceAttrSet(resourceName, "custom_zone_mapping.IE"),
					resource.TestCheckResourceAttrSet(resourceName, "custom_zone_mapping.DE"),
					resource.TestCheckResourceAttrSet(resourceName, "project_id"),
					resource.TestCheckResourceAttr(resourceName, "cluster_name", name),
					testAccCheckMongoDBAtlasGlobalClusterAttributes(&globalConfig, 5),
				),
			},
			{
				Config: testAccMongoDBAtlasGlobalClusterWithDBConfig(orgID, projectName, name, "false", customZoneUpdated),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMongoDBAtlasGlobalClusterExists(resourceName, &globalConfig),
					resource.TestCheckResourceAttrSet(resourceName, "managed_namespaces.#"),
					resource.TestCheckResourceAttrSet(resourceName, "custom_zone_mappings.#"),
					resource.TestCheckResourceAttrSet(resourceName, "custom_zone_mapping.%"),
					resource.TestCheckResourceAttrSet(resourceName, "custom_zone_mapping.US"),
					resource.TestCheckResourceAttrSet(resourceName, "custom_zone_mapping.IE"),
					resource.TestCheckResourceAttrSet(resourceName, "custom_zone_mapping.DE"),
					resource.TestCheckResourceAttrSet(resourceName, "custom_zone_mapping.JP"),
					resource.TestCheckResourceAttrSet(resourceName, "project_id"),
					resource.TestCheckResourceAttr(resourceName, "cluster_name", name),
					testAccCheckMongoDBAtlasGlobalClusterAttributes(&globalConfig, 5),
				),
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
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("not found: %s", resourceName)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("no ID is set")
		}
		ids := conversion.DecodeStateID(rs.Primary.ID)
		globalConfigResp, _, err := acc.Conn().GlobalClusters.Get(context.Background(), ids["project_id"], ids["cluster_name"])
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
		ids := conversion.DecodeStateID(rs.Primary.ID)
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
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "mongodbatlas_global_cluster_config" {
			continue
		}

		// Try to find the cluster
		globalConfig, _, err := acc.Conn().GlobalClusters.Get(context.Background(), rs.Primary.Attributes["project_id"], rs.Primary.Attributes["cluster_name"])
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

func testAccMongoDBAtlasGlobalClusterConfig(orgID, projectName, name, backupEnabled, isCustomShard, isShardKeyUnique string) string {
	return fmt.Sprintf(`
		resource "mongodbatlas_project" "project" {
			org_id = %[1]q
			name   = %[2]q
		}
		
		resource "mongodbatlas_cluster" "test" {
			project_id              = mongodbatlas_project.project.id
			name                    = %[3]q
			disk_size_gb            = 80
			backup_enabled          = %[4]q
			cloud_backup            = true
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
				db               		   = "mydata"
				collection       		   = "publishers"
				custom_shard_key		   = "city"
				is_custom_shard_key_hashed = %[5]q
				is_shard_key_unique 	   = %[6]q
			}

			custom_zone_mappings {
				location = "CA"
				zone     = "Zone 1"
			}
		}
	`, orgID, projectName, name, backupEnabled, isCustomShard, isShardKeyUnique)
}

func testAccMongoDBAtlasGlobalClusterWithAWSClusterConfig(orgID, projectName, name, backupEnabled string) string {
	return fmt.Sprintf(`

		resource "mongodbatlas_project" "project" {
			org_id = %[1]q
			name   = %[2]q
		}

		resource "mongodbatlas_cluster" "test" {
			project_id              = mongodbatlas_project.project.id
			name                    = %[3]q
			disk_size_gb            = 80
			cloud_backup            = %[4]s
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
	`, orgID, projectName, name, backupEnabled)
}

func testAccMongoDBAtlasGlobalClusterWithDBConfig(orgID, projectName, name, backupEnabled, zones string) string {
	return fmt.Sprintf(`

resource "mongodbatlas_project" "project" {
	org_id = %[1]q
	name   = %[2]q
}

resource "mongodbatlas_database_user" "test" {
  username           = "horizonv2-sg"
  password           = "password testing something"
  project_id         = mongodbatlas_project.project.id
  auth_database_name = "admin"

  roles {
    role_name     = "readWrite"
    database_name = "horizonv2-sg"
  }
}

resource "mongodbatlas_cluster" "test" {
  project_id   = mongodbatlas_project.project.id
  name         = %[3]q
  disk_size_gb = 80
  cloud_backup = %[4]s
  cluster_type = "GEOSHARDED"

  // Provider Settings "block"
  provider_name               = "AWS"
  provider_instance_size_name = "M30"

  replication_specs {
    zone_name  = "US"
    num_shards = 1
    regions_config {
      region_name     = "US_EAST_1"
      electable_nodes = 3
      priority        = 7
      read_only_nodes = 0
    }
  }
  replication_specs {
    zone_name  = "EU"
    num_shards = 1
    regions_config {
      region_name     = "EU_WEST_1"
      electable_nodes = 3
      priority        = 7
      read_only_nodes = 0
    }
  }
  replication_specs {
    zone_name  = "DE"
    num_shards = 1
    regions_config {
      region_name     = "EU_NORTH_1"
      electable_nodes = 3
      priority        = 7
      read_only_nodes = 0
    }
  }
  replication_specs {
    zone_name  = "JP"
    num_shards = 1
    regions_config {
      region_name     = "AP_NORTHEAST_1"
      electable_nodes = 3
      priority        = 7
      read_only_nodes = 0
    }
  }
}

resource "mongodbatlas_global_cluster_config" "test" {
  project_id = mongodbatlas_cluster.test.project_id
  cluster_name = mongodbatlas_cluster.test.name

  managed_namespaces {
    db               = "horizonv2-sg"
    collection       = "entitlements.entitlement"
    custom_shard_key = "orgId"
  }
  managed_namespaces {
    db               = "horizonv2-sg"
    collection       = "entitlements.homesitemapping"
    custom_shard_key = "orgId"
  }
  managed_namespaces {
    db               = "horizonv2-sg"
    collection       = "entitlements.site"
    custom_shard_key = "orgId"
  }
  managed_namespaces {
    db               = "horizonv2-sg"
    collection       = "entitlements.userDesktopMapping"
    custom_shard_key = "orgId"
  }
  managed_namespaces {
    db               = "horizonv2-sg"
    collection       = "session"
    custom_shard_key = "orgId"
  }
  %[5]s
}
	`, orgID, projectName, name, backupEnabled, zones)
}
