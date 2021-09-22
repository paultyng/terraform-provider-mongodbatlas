package mongodbatlas

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceMongoDBAtlasOrganizationApiKey_basic(t *testing.T) {
	var (
		dataSourceName = "data.mongodbatlas_organization_api_key.test"
		orgID          = os.Getenv("MONGODB_ATLAS_ORG_ID")
		desc           = fmt.Sprintf("test-acc-%s", acctest.RandString(10))
		role           = "ORG_OWNER"
		accessList     = "1.1.1.1/30"
	)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceMongoDBAtlasOrganizationApiKeyConfig(orgID, desc, role, accessList),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(dataSourceName, "org_id"),
					resource.TestCheckResourceAttrSet(dataSourceName, "api_key_id"),
					resource.TestCheckResourceAttr(dataSourceName, "description", desc),
					resource.TestCheckResourceAttr(dataSourceName, "access_list_cidr_blocks.#", "1"),
					resource.TestCheckResourceAttr(dataSourceName, "roles.#", "1"),
				),
			},
		},
	})
}

func testAccDataSourceMongoDBAtlasOrganizationApiKeyConfig(orgID, desc, role, accessList string) string {
	return fmt.Sprintf(`
        resource "mongodbatlas_organization_api_key" "test" {
            org_id                  = "%s"
            description             = "%s"
            roles                   = ["%s"]
            access_list_cidr_blocks = ["%s"]
        }

        data "mongodbatlas_organization_api_key" "test" {
            org_id     = mongodbatlas_organization_api_key.test.org_id
            api_key_id = mongodbatlas_organization_api_key.test.api_key_id
        }
    `, orgID, desc, role, accessList)
}
