package mongodbatlas

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccResourceMongoDBAtlasCloudProviderAccessAuthorization_basic(t *testing.T) {
	SkipTestExtCred(t)
	var (
		projectID       = os.Getenv("MONGODB_ATLAS_PROJECT_ID")
		policyName      = acctest.RandomWithPrefix("tf-acc")
		roleName        = acctest.RandomWithPrefix("tf-acc")
		roleNameUpdated = acctest.RandomWithPrefix("tf-acc")
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		// same as regular cloud provider access resource
		CheckDestroy: testAccCheckMongoDBAtlasProviderAccessDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccMongoDBAtlasCloudProviderAccessAuthorizationConfig(projectID, policyName, roleName),
			},
			{
				Config: testAccMongoDBAtlasCloudProviderAccessAuthorizationConfig(projectID, policyName, roleNameUpdated),
			},
		},
	},
	)
}

func testAccMongoDBAtlasCloudProviderAccessAuthorizationConfig(projectID, roleName, policyName string) string {
	return fmt.Sprintf(`
resource "aws_iam_role_policy" "test_policy" {
  name = %[2]q
  role = aws_iam_role.test_role.id

  policy = <<-EOF
  {
    "Version": "2012-10-17",
    "Statement": [
      {
        "Effect": "Allow",
		"Action": "*",
		"Resource": "*"
      }
    ]
  }
  EOF
}

resource "aws_iam_role" "test_role" {
  name = %[3]q

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "AWS": "${mongodbatlas_cloud_provider_access_setup.setup_only.aws_config.0.atlas_aws_account_arn}"
      },
      "Action": "sts:AssumeRole",
      "Condition": {
        "StringEquals": {
          "sts:ExternalId": "${mongodbatlas_cloud_provider_access_setup.setup_only.aws_config.0.atlas_assumed_role_external_id}"
        }
      }
    }
  ]
}
EOF
}


resource "mongodbatlas_cloud_provider_access_setup" "setup_only" {
  project_id    = %[1]q
  provider_name = "AWS"
}

resource "mongodbatlas_cloud_provider_access_authorization" "auth_role" {
  project_id = %[1]q
  role_id    = mongodbatlas_cloud_provider_access_setup.setup_only.role_id

  aws {
    iam_assumed_role_arn = aws_iam_role.test_role.arn
  }
}
	`, projectID, policyName, roleName)
}
