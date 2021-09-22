package mongodbatlas

import (
	"context"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	matlas "go.mongodb.org/atlas/mongodbatlas"
)

func TestAccResourceMongoDBAtlasOrganizationApiKeys_basic(t *testing.T) {
	var (
		apiKey           matlas.APIKey
		resourceName     = "mongodbatlas_organization_api_key.test"
		orgID            = os.Getenv("MONGODB_ATLAS_ORG_ID")
		desc             = fmt.Sprintf("test-acc-%s", acctest.RandString(10))
		descUpdate       = fmt.Sprintf("test-update-acc-%s", acctest.RandString(10))
		roles            = []string{"ORG_OWNER"}
		rolesUpdate      = []string{"ORG_OWNER", "ORG_READ_ONLY"}
		accessList       = []string{"1.1.1.1/30", "10.10.10.10/32"}
		accessListUpdate = []string{"1.1.1.1/30"}
	)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccMongoDBAtlasOrganizationApiKeyConfig(orgID, desc, roles, accessList),
				Check: resource.ComposeTestCheckFunc(
					testAccMongoDBAtlasOrganizationApiKeyExists(resourceName, &apiKey),
					resource.TestCheckResourceAttrSet(resourceName, "org_id"),
					resource.TestCheckResourceAttr(resourceName, "description", desc),
					resource.TestCheckResourceAttr(resourceName, "roles.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "access_list_cidr_blocks.#", "2"),
				),
			},
			{
				Config: testAccMongoDBAtlasOrganizationApiKeyConfig(orgID, descUpdate, roles, accessList),
				Check: resource.ComposeTestCheckFunc(
					testAccMongoDBAtlasOrganizationApiKeyExists(resourceName, &apiKey),
					resource.TestCheckResourceAttrSet(resourceName, "org_id"),
					resource.TestCheckResourceAttr(resourceName, "description", descUpdate),
					resource.TestCheckResourceAttr(resourceName, "roles.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "access_list_cidr_blocks.#", "2"),
				),
			},
			{
				Config: testAccMongoDBAtlasOrganizationApiKeyConfig(orgID, descUpdate, rolesUpdate, accessList),
				Check: resource.ComposeTestCheckFunc(
					testAccMongoDBAtlasOrganizationApiKeyExists(resourceName, &apiKey),
					resource.TestCheckResourceAttrSet(resourceName, "org_id"),
					resource.TestCheckResourceAttr(resourceName, "description", descUpdate),
					resource.TestCheckResourceAttr(resourceName, "roles.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "access_list_cidr_blocks.#", "2"),
				),
			},
			{
				Config: testAccMongoDBAtlasOrganizationApiKeyConfig(orgID, descUpdate, rolesUpdate, accessListUpdate),
				Check: resource.ComposeTestCheckFunc(
					testAccMongoDBAtlasOrganizationApiKeyExists(resourceName, &apiKey),
					resource.TestCheckResourceAttrSet(resourceName, "org_id"),
					resource.TestCheckResourceAttr(resourceName, "description", descUpdate),
					resource.TestCheckResourceAttr(resourceName, "roles.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "access_list_cidr_blocks.#", "1"),
				),
			},
		},
	})
}

func TestAccResourceMongoDBAtlasOrganizationApiKeys_InvalidAccessList(t *testing.T) {
	var (
		orgID      = os.Getenv("MONGODB_ATLAS_ORG_ID")
		desc       = fmt.Sprintf("test-acc-%s", acctest.RandString(10))
		roles      = []string{"ORG_OWNER"}
		accessList = []string{"1.1.1.1/30", "10.10.10.10"}
	)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccMongoDBAtlasOrganizationApiKeyConfig(orgID, desc, roles, accessList),
				ExpectError: regexp.MustCompile(".*error creating the MongoDB Organization .* API Key, invalid CIDR block.*"),
			},
		},
	})
}

func TestAccResourceMongoDBAtlasOrganizationApiKeys_InvalidRoles(t *testing.T) {
	var (
		orgID      = os.Getenv("MONGODB_ATLAS_ORG_ID")
		desc       = fmt.Sprintf("test-acc-%s", acctest.RandString(10))
		roles      = []string{}
		accessList = []string{"1.1.1.1/30", "10.10.10.10/20"}
	)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccMongoDBAtlasOrganizationApiKeyConfig(orgID, desc, roles, accessList),
				ExpectError: regexp.MustCompile(".*at least one role must be present for an API Key.*"),
			},
		},
	})
}

func TestAccResourceMongoDBAtlasOrganizationApiKeys_EmptyAccessList(t *testing.T) {
	var (
		apiKey       matlas.APIKey
		resourceName = "mongodbatlas_organization_api_key.test"
		orgID        = os.Getenv("MONGODB_ATLAS_ORG_ID")
		desc         = fmt.Sprintf("test-acc-%s", acctest.RandString(10))
		roles        = []string{"ORG_OWNER"}
		accessList   = []string{}
	)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccMongoDBAtlasOrganizationApiKeyConfig(orgID, desc, roles, accessList),
				Check: resource.ComposeTestCheckFunc(
					testAccMongoDBAtlasOrganizationApiKeyExists(resourceName, &apiKey),
					resource.TestCheckResourceAttrSet(resourceName, "org_id"),
					resource.TestCheckResourceAttr(resourceName, "description", desc),
					resource.TestCheckResourceAttr(resourceName, "roles.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "access_list_cidr_blocks.#", "0"),
				),
			},
		},
	})
}

func TestAccResourceMongoDBAtlasOrganizationApiKeys_importBasic(t *testing.T) {
	var (
		apiKey       matlas.APIKey
		resourceName = "mongodbatlas_organization_api_key.test"
		orgID        = os.Getenv("MONGODB_ATLAS_ORG_ID")
		desc         = fmt.Sprintf("test-acc-%s", acctest.RandString(10))
		roles        = []string{"ORG_OWNER"}
		accessList   = []string{}
	)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccMongoDBAtlasOrganizationApiKeyConfig(orgID, desc, roles, accessList),
				Check: resource.ComposeTestCheckFunc(
					testAccMongoDBAtlasOrganizationApiKeyExists(resourceName, &apiKey),
					resource.TestCheckResourceAttrSet(resourceName, "org_id"),
					resource.TestCheckResourceAttr(resourceName, "description", desc),
					resource.TestCheckResourceAttr(resourceName, "roles.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "access_list_cidr_blocks.#", "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateIdFunc: testAccCheckMongoDBAtlasOrganizationApiKeyStateIDFunc(resourceName),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckMongoDBAtlasOrganizationApiKeyStateIDFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("not found: %s", resourceName)
		}

		return fmt.Sprintf("%s-%s", rs.Primary.Attributes["org_id"], rs.Primary.Attributes["public_key"]), nil
	}
}

func testAccMongoDBAtlasOrganizationApiKeyExists(resourceName string, apiKey *matlas.APIKey) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*MongoDBClient).Atlas

		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("not found: %s", resourceName)
		}

		ids := decodeStateID(rs.Primary.ID)

		orgID := ids["org_id"]
		api_key_id := ids["api_key_id"]

		if orgID == "" && api_key_id == "" {
			return fmt.Errorf("no ID is set")
		}

		log.Printf("[DEBUG] orgID: %s", orgID)
		log.Printf("[DEBUG] apiKeyID: %s", api_key_id)

		apiKeyResp, _, err := conn.APIKeys.Get(context.Background(), orgID, api_key_id)
		if err == nil {
			*apiKey = *apiKeyResp
			return nil
		}

		return fmt.Errorf("API key(%s) does not exist", api_key_id)
	}
}

func testAccMongoDBAtlasOrganizationApiKeyConfig(orgID, desc string, roles, accessList []string) string {
	return fmt.Sprintf(`
        resource "mongodbatlas_organization_api_key" "test" {
            org_id                  = "%s"
            description             = "%s"
            roles                   = %s
            access_list_cidr_blocks = %s
        }
    `, orgID, desc, strings.ReplaceAll(fmt.Sprintf("%+q", roles), " ", ","),
		strings.ReplaceAll(fmt.Sprintf("%+q", accessList), " ", ","),
	)
}
