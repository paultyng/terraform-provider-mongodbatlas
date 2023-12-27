package networkpeering_test

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/mongodb/terraform-provider-mongodbatlas/internal/common/conversion"
	"github.com/mongodb/terraform-provider-mongodbatlas/internal/config"
	"github.com/mongodb/terraform-provider-mongodbatlas/internal/testutil/acc"
	matlas "go.mongodb.org/atlas/mongodbatlas"
)

func TestAccNetworkRSNetworkPeering_basicAWS(t *testing.T) {
	var (
		peer                 matlas.Peer
		resourceName         = "mongodbatlas_network_peering.test"
		dataSourceName       = "data.mongodbatlas_network_peering.test"
		pluralDataSourceName = "data.mongodbatlas_network_peerings.test"
		orgID                = os.Getenv("MONGODB_ATLAS_ORG_ID")
		vpcID                = os.Getenv("AWS_VPC_ID")
		vpcCIDRBlock         = os.Getenv("AWS_VPC_CIDR_BLOCK")
		awsAccountID         = os.Getenv("AWS_ACCOUNT_ID")
		awsRegion            = os.Getenv("AWS_REGION")
		providerName         = "AWS"
		projectName          = acctest.RandomWithPrefix("test-acc-project-aws")
	)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acc.PreCheckBasic(t); acc.PreCheckPeeringEnvAWS(t) },
		ProtoV6ProviderFactories: acc.TestAccProviderV6Factories,
		CheckDestroy:             acc.CheckDestroyNetworkPeering,
		Steps: []resource.TestStep{
			{
				Config: testAccMongoDBAtlasNetworkPeeringConfigAWS(orgID, projectName, providerName, vpcID, awsAccountID, vpcCIDRBlock, awsRegion),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMongoDBAtlasNetworkPeeringExists(resourceName, &peer),
					resource.TestCheckResourceAttrSet(resourceName, "project_id"),
					resource.TestCheckResourceAttrSet(resourceName, "container_id"),
					resource.TestCheckResourceAttr(resourceName, "provider_name", providerName),
					resource.TestCheckResourceAttr(resourceName, "vpc_id", vpcID),
					resource.TestCheckResourceAttr(resourceName, "aws_account_id", awsAccountID),

					resource.TestCheckResourceAttrSet(dataSourceName, "project_id"),
					resource.TestCheckResourceAttrSet(dataSourceName, "container_id"),
					resource.TestCheckResourceAttr(dataSourceName, "provider_name", providerName),
					resource.TestCheckResourceAttr(dataSourceName, "vpc_id", vpcID),
					resource.TestCheckResourceAttr(dataSourceName, "aws_account_id", awsAccountID),

					resource.TestCheckResourceAttrSet(pluralDataSourceName, "results.#"),
					resource.TestCheckResourceAttrSet(pluralDataSourceName, "results.0.provider_name"),
					resource.TestCheckResourceAttrSet(pluralDataSourceName, "results.0.vpc_id"),
					resource.TestCheckResourceAttrSet(pluralDataSourceName, "results.0.aws_account_id"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportStateIdFunc:       testAccCheckMongoDBAtlasNetworkPeeringImportStateIDFunc(resourceName),
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"accepter_region_name", "container_id"},
			},
		},
	})
}

func TestAccNetworkRSNetworkPeering_basicAzure(t *testing.T) {
	acc.SkipTestExtCred(t)
	var (
		peer              matlas.Peer
		resourceName      = "mongodbatlas_network_peering.test"
		projectID         = os.Getenv("MONGODB_ATLAS_PROJECT_ID")
		directoryID       = os.Getenv("AZURE_DIRECTORY_ID")
		subscriptionID    = os.Getenv("AZURE_SUBSCRIPTION_ID")
		resourceGroupName = os.Getenv("AZURE_RESOURCE_GROUP_NAME")
		vNetName          = os.Getenv("AZURE_VNET_NAME")
		providerName      = "AZURE"
	)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acc.PreCheck(t); acc.PreCheckPeeringEnvAzure(t) },
		ProtoV6ProviderFactories: acc.TestAccProviderV6Factories,
		CheckDestroy:             acc.CheckDestroyNetworkPeering,
		Steps: []resource.TestStep{
			{
				Config: testAccMongoDBAtlasNetworkPeeringConfigAzure(projectID, providerName, directoryID, subscriptionID, resourceGroupName, vNetName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMongoDBAtlasNetworkPeeringExists(resourceName, &peer),
					resource.TestCheckResourceAttrSet(resourceName, "project_id"),
					resource.TestCheckResourceAttrSet(resourceName, "container_id"),
					resource.TestCheckResourceAttr(resourceName, "provider_name", providerName),
					resource.TestCheckResourceAttr(resourceName, "vnet_name", vNetName),
					resource.TestCheckResourceAttr(resourceName, "azure_directory_id", directoryID),
				),
			},
			{
				ResourceName:            resourceName,
				ImportStateIdFunc:       testAccCheckMongoDBAtlasNetworkPeeringImportStateIDFunc(resourceName),
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"container_id"},
			},
		},
	})
}

func TestAccNetworkRSNetworkPeering_basicGCP(t *testing.T) {
	acc.SkipTestExtCred(t)
	var (
		peer         matlas.Peer
		resourceName = "mongodbatlas_network_peering.test"
		projectID    = os.Getenv("MONGODB_ATLAS_PROJECT_ID")
		providerName = "GCP"
		gcpProjectID = os.Getenv("GCP_PROJECT_ID")
		networkName  = fmt.Sprintf("test-acc-name-%s", acctest.RandString(5))
	)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acc.PreCheck(t); acc.PreCheckPeeringEnvGCP(t) },
		ProtoV6ProviderFactories: acc.TestAccProviderV6Factories,
		CheckDestroy:             acc.CheckDestroyNetworkPeering,
		Steps: []resource.TestStep{
			{
				Config: testAccMongoDBAtlasNetworkPeeringConfigGCP(projectID, providerName, gcpProjectID, networkName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMongoDBAtlasNetworkPeeringExists(resourceName, &peer),
					resource.TestCheckResourceAttrSet(resourceName, "project_id"),
					resource.TestCheckResourceAttrSet(resourceName, "container_id"),
					resource.TestCheckResourceAttrSet(resourceName, "network_name"),

					resource.TestCheckResourceAttr(resourceName, "provider_name", providerName),
					resource.TestCheckResourceAttr(resourceName, "gcp_project_id", gcpProjectID),
					resource.TestCheckResourceAttr(resourceName, "network_name", networkName),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateIdFunc: testAccCheckMongoDBAtlasNetworkPeeringImportStateIDFunc(resourceName),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccNetworkRSNetworkPeering_AWSDifferentRegionName(t *testing.T) {
	acc.SkipTestExtCred(t)
	var (
		peer                  matlas.Peer
		resourcePeerName      = "mongodbatlas_network_peering.diff_region"
		resourceContainerName = "mongodbatlas_network_container.test"
		projectID             = os.Getenv("MONGODB_ATLAS_PROJECT_ID")
		vpcID                 = os.Getenv("AWS_VPC_ID")
		vpcCIDRBlock          = os.Getenv("AWS_VPC_CIDR_BLOCK")
		awsAccountID          = os.Getenv("AWS_ACCOUNT_ID")
		containerRegion       = "US_WEST_2"
		peerRegion            = strings.ToLower(strings.ReplaceAll(os.Getenv("AWS_REGION"), "_", "-"))
		providerName          = "AWS"
	)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acc.PreCheck(t)
			acc.PreCheckPeeringEnvAWS(t)
			func() {
				if strings.EqualFold(containerRegion, peerRegion) {
					t.Fatalf("the `AWS_REGION` (%s) must be different region than %s", peerRegion, containerRegion)
				}
			}()
		},
		ProtoV6ProviderFactories: acc.TestAccProviderV6Factories,
		CheckDestroy:             acc.CheckDestroyNetworkPeering,
		Steps: []resource.TestStep{
			{
				Config: testAccMongoDBAtlasNetworkPeeringConfigAWSWithDifferentRegionName(projectID, providerName, containerRegion, peerRegion, vpcCIDRBlock, vpcID, awsAccountID),
				Check: resource.ComposeTestCheckFunc(
					// Peering test
					testAccCheckMongoDBAtlasNetworkPeeringExists(resourcePeerName, &peer),
					resource.TestCheckResourceAttrSet(resourcePeerName, "accepter_region_name"),
					resource.TestCheckResourceAttrSet(resourcePeerName, "project_id"),
					resource.TestCheckResourceAttrSet(resourcePeerName, "container_id"),
					resource.TestCheckResourceAttrSet(resourcePeerName, "provider_name"),
					resource.TestCheckResourceAttrSet(resourcePeerName, "route_table_cidr_block"),
					resource.TestCheckResourceAttrSet(resourcePeerName, "vpc_id"),
					resource.TestCheckResourceAttrSet(resourcePeerName, "aws_account_id"),

					resource.TestCheckResourceAttr(resourcePeerName, "accepter_region_name", peerRegion),
					resource.TestCheckResourceAttr(resourcePeerName, "project_id", projectID),
					resource.TestCheckResourceAttr(resourcePeerName, "provider_name", providerName),
					resource.TestCheckResourceAttr(resourcePeerName, "route_table_cidr_block", vpcCIDRBlock),
					resource.TestCheckResourceAttr(resourcePeerName, "vpc_id", vpcID),
					resource.TestCheckResourceAttr(resourcePeerName, "aws_account_id", awsAccountID),

					// Container test
					resource.TestCheckResourceAttrSet(resourceContainerName, "project_id"),
					resource.TestCheckResourceAttrSet(resourceContainerName, "atlas_cidr_block"),
					resource.TestCheckResourceAttrSet(resourceContainerName, "provider_name"),
					resource.TestCheckResourceAttrSet(resourceContainerName, "region_name"),

					resource.TestCheckResourceAttr(resourceContainerName, "project_id", projectID),
					resource.TestCheckResourceAttr(resourceContainerName, "atlas_cidr_block", "192.168.200.0/21"),
					resource.TestCheckResourceAttr(resourceContainerName, "provider_name", providerName),
					resource.TestCheckResourceAttr(resourceContainerName, "region_name", containerRegion),
				),
			},
		},
	})
}

func testAccCheckMongoDBAtlasNetworkPeeringImportStateIDFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("not found: %s", resourceName)
		}

		ids := conversion.DecodeStateID(rs.Primary.ID)

		return fmt.Sprintf("%s-%s-%s", ids["project_id"], ids["peer_id"], ids["provider_name"]), nil
	}
}

func testAccCheckMongoDBAtlasNetworkPeeringExists(resourceName string, peer *matlas.Peer) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acc.TestAccProviderSdkV2.Meta().(*config.MongoDBClient).Atlas

		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no ID is set")
		}

		ids := conversion.DecodeStateID(rs.Primary.ID)
		log.Printf("[DEBUG] projectID: %s", ids["project_id"])

		if peerResp, _, err := conn.Peers.Get(context.Background(), ids["project_id"], ids["peer_id"]); err == nil {
			*peer = *peerResp
			peer.ProviderName = ids["provider_name"]

			return nil
		}

		return fmt.Errorf("peer(%s:%s) does not exist", rs.Primary.Attributes["project_id"], rs.Primary.Attributes["peer_id"])
	}
}

func testAccMongoDBAtlasNetworkPeeringConfigAWS(orgID, projectName, providerName, vpcID, awsAccountID, vpcCIDRBlock, awsRegion string) string {
	return fmt.Sprintf(`
		resource "mongodbatlas_project" "my_project" {
			name   = %[2]q
			org_id = %[1]q
		}
		resource "mongodbatlas_network_container" "test" {
			project_id   		  = mongodbatlas_project.my_project.id
			atlas_cidr_block  = "192.168.208.0/21"
			provider_name		  = "%[3]s"
			region_name			  = "%[7]s"
		}

		resource "mongodbatlas_network_peering" "test" {
			accepter_region_name	  = lower(replace("%[7]s", "_", "-"))
			project_id    			    = mongodbatlas_project.my_project.id
			container_id            = mongodbatlas_network_container.test.id
			provider_name           = "%[3]s"
			route_table_cidr_block  = "%[6]s"
			vpc_id					        = "%[4]s"
			aws_account_id	        = "%[5]s"
		}

		data "mongodbatlas_network_peering" "test" {
			project_id = mongodbatlas_project.my_project.id
			peering_id = mongodbatlas_network_peering.test.peer_id
		}

		data "mongodbatlas_network_peerings" "test" {
			project_id = mongodbatlas_network_peering.test.project_id
		}
	`, orgID, projectName, providerName, vpcID, awsAccountID, vpcCIDRBlock, awsRegion)
}

func testAccMongoDBAtlasNetworkPeeringConfigAzure(projectID, providerName, directoryID, subscriptionID, resourceGroupName, vNetName string) string {
	return fmt.Sprintf(`
		resource "mongodbatlas_network_container" "test" {
			project_id   		  = "%[1]s"
			atlas_cidr_block  = "192.168.208.0/21"
			provider_name		  = "%[2]s"
			region    			  = "US_EAST_2"
		}

		resource "mongodbatlas_network_peering" "test" {
			project_id   		      = "%[1]s"
			container_id          = mongodbatlas_network_container.test.container_id
			provider_name         = "%[2]s"
			azure_directory_id    = "%[3]s"
			azure_subscription_id = "%[4]s"
			resource_group_name   = "%[5]s"
			vnet_name	            = "%[6]s"
		}
	`, projectID, providerName, directoryID, subscriptionID, resourceGroupName, vNetName)
}

func testAccMongoDBAtlasNetworkPeeringConfigGCP(projectID, providerName, gcpProjectID, networkName string) string {
	return fmt.Sprintf(`
		resource "mongodbatlas_network_container" "test" {
			project_id       = "%[1]s"
			atlas_cidr_block = "192.168.192.0/18"
			provider_name    = "%[2]s"
		}

		resource "mongodbatlas_network_peering" "test" {
			project_id     = "%[1]s"
			container_id   = mongodbatlas_network_container.test.container_id
			provider_name  = "%[2]s"
			gcp_project_id = "%[3]s"
			network_name   = "%[4]s"
		}
	`, projectID, providerName, gcpProjectID, networkName)
}

func testAccMongoDBAtlasNetworkPeeringConfigAWSWithDifferentRegionName(projectID, providerName, containerRegion, peerRegion, vpcCIDRBlock, vpcID, awsAccountID string) string {
	return fmt.Sprintf(`
		resource "mongodbatlas_network_container" "test" {
			project_id   		  = "%[1]s"
			atlas_cidr_block  = "192.168.200.0/21"
			provider_name		  = "%[2]s"
			region_name			  = "%[3]s"
		}

		resource "mongodbatlas_network_peering" "diff_region" {
			accepter_region_name	  = "%[4]s"
			project_id    			    = "%[1]s"
			container_id            = mongodbatlas_network_container.test.container_id
			provider_name           = "%[2]s"
			route_table_cidr_block  = "%[5]s"
			vpc_id					        = "%[6]s"
			aws_account_id	        = "%[7]s"
		}
	`, projectID, providerName, containerRegion, peerRegion, vpcCIDRBlock, vpcID, awsAccountID)
}
