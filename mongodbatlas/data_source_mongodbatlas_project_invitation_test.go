package mongodbatlas

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceMongoDBAtlasProjectInvitation_basic(t *testing.T) {
	var (
		dataSourceName = "mongodbatlas_project_invitations.test"
		projectID      = os.Getenv("MONGODB_ATLAS_PROJECT_ID")
		name           = fmt.Sprintf("test-acc-%s@mongodb.com", acctest.RandString(10))
		initialRole    = []string{"GROUP_OWNER"}
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckMongoDBAtlasProjectInvitationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceMongoDBAtlasProjectInvitationConfig(projectID, name, initialRole),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(dataSourceName, "project_id"),
					resource.TestCheckResourceAttrSet(dataSourceName, "username"),
					resource.TestCheckResourceAttrSet(dataSourceName, "invitation_id"),
					resource.TestCheckResourceAttr(dataSourceName, "username", name),
					resource.TestCheckResourceAttr(dataSourceName, "roles.#", "1"),
				),
			},
		},
	})
}

func testAccDataSourceMongoDBAtlasProjectInvitationConfig(projectID, username string, roles []string) string {
	return fmt.Sprintf(`
		resource "mongodbatlas_project_invitation" "test" {
			project_id	= "%s"
			username  	= "%s"
			roles  			= %s
		}

		data "mongodbatlas_project_invitation" "test" {
			project_id = mongodbatlas_project_invitation.test.projectID
			username 	 = mongodbatlas_project_invitation.test.username
		}`, projectID, username,
		strings.Join(roles, ","),
	)
}
