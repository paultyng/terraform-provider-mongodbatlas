package encryptionatrest_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"

	"github.com/mongodb/terraform-provider-mongodbatlas/internal/common/conversion"
	"github.com/mongodb/terraform-provider-mongodbatlas/internal/testutil/acc"
	"github.com/mongodb/terraform-provider-mongodbatlas/internal/testutil/mig"
	"go.mongodb.org/atlas-sdk/v20240805003/admin"
)

func TestMigEncryptionAtRest_basicAWS(t *testing.T) {
	acc.SkipTestForCI(t) // needs AWS configuration

	var (
		resourceName = "mongodbatlas_encryption_at_rest.test"
		projectID    = os.Getenv("MONGODB_ATLAS_PROJECT_ID")

		awsKms = admin.AWSKMSConfiguration{
			Enabled:             conversion.Pointer(true),
			CustomerMasterKeyID: conversion.StringPtr(os.Getenv("AWS_CUSTOMER_MASTER_KEY_ID")),
			Region:              conversion.StringPtr(conversion.AWSRegionToMongoDBRegion(os.Getenv("AWS_REGION"))),
			RoleId:              conversion.StringPtr(os.Getenv("AWS_ROLE_ID")),
		}
	)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { mig.PreCheck(t); acc.PreCheckAwsEnv(t) },
		CheckDestroy: testAccCheckMongoDBAtlasEncryptionAtRestDestroy,
		Steps: []resource.TestStep{
			{
				ExternalProviders: mig.ExternalProviders(),
				Config:            testAccMongoDBAtlasEncryptionAtRestConfigAwsKms(projectID, &awsKms),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckMongoDBAtlasEncryptionAtRestExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "project_id", projectID),
					resource.TestCheckResourceAttr(resourceName, "aws_kms_config.0.enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "aws_kms_config.0.region", awsKms.GetRegion()),
					resource.TestCheckResourceAttr(resourceName, "aws_kms_config.0.role_id", awsKms.GetRoleId()),
				),
			},
			{
				ProtoV6ProviderFactories: acc.TestAccProviderV6Factories,
				Config:                   testAccMongoDBAtlasEncryptionAtRestConfigAwsKms(projectID, &awsKms),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						acc.DebugPlan(),
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
		},
	})
}

func TestMigEncryptionAtRest_withRole_basicAWS(t *testing.T) {
	acc.SkipTestForCI(t) // needs AWS configuration

	var (
		resourceName = "mongodbatlas_encryption_at_rest.test"
		projectID    = os.Getenv("MONGODB_ATLAS_PROJECT_ID")

		awsIAMRoleName       = acc.RandomIAMRole()
		awsIAMRolePolicyName = fmt.Sprintf("%s-policy", awsIAMRoleName)
		awsKeyName           = acc.RandomName()

		awsKms = admin.AWSKMSConfiguration{
			Enabled:             conversion.Pointer(true),
			Region:              conversion.StringPtr(conversion.AWSRegionToMongoDBRegion(os.Getenv("AWS_REGION"))),
			CustomerMasterKeyID: conversion.StringPtr(os.Getenv("AWS_CUSTOMER_MASTER_KEY_ID")),
		}
	)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { mig.PreCheck(t); acc.PreCheckAwsEnv(t) },
		CheckDestroy: testAccCheckMongoDBAtlasEncryptionAtRestDestroy,
		Steps: []resource.TestStep{
			{
				ExternalProviders: mig.ExternalProvidersWithAWS(),
				Config:            testAccMongoDBAtlasEncryptionAtRestConfigAwsKmsWithRole(projectID, awsIAMRoleName, awsIAMRolePolicyName, awsKeyName, &awsKms),
			},
			{
				ExternalProviders:        acc.ExternalProvidersOnlyAWS(),
				ProtoV6ProviderFactories: acc.TestAccProviderV6Factories,
				Config:                   testAccMongoDBAtlasEncryptionAtRestConfigAwsKmsWithRole(projectID, awsIAMRoleName, awsIAMRolePolicyName, awsKeyName, &awsKms),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckMongoDBAtlasEncryptionAtRestExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "project_id", projectID),
					resource.TestCheckResourceAttr(resourceName, "aws_kms_config.0.enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "aws_kms_config.0.region", awsKms.GetRegion()),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						acc.DebugPlan(),
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
		},
	})
}

func TestMigEncryptionAtRest_basicAzure(t *testing.T) {
	acc.SkipTestForCI(t) // needs Azure configuration

	var (
		resourceName = "mongodbatlas_encryption_at_rest.test"
		projectID    = os.Getenv("MONGODB_ATLAS_PROJECT_ID")

		azureKeyVault = admin.AzureKeyVault{
			Enabled:           conversion.Pointer(true),
			ClientID:          conversion.StringPtr(os.Getenv("AZURE_CLIENT_ID")),
			AzureEnvironment:  conversion.StringPtr("AZURE"),
			SubscriptionID:    conversion.StringPtr(os.Getenv("AZURE_SUBSCRIPTION_ID")),
			ResourceGroupName: conversion.StringPtr(os.Getenv("AZURE_RESOURCE_GROUP_NAME")),
			KeyVaultName:      conversion.StringPtr(os.Getenv("AZURE_KEY_VAULT_NAME")),
			KeyIdentifier:     conversion.StringPtr(os.Getenv("AZURE_KEY_IDENTIFIER")),
			Secret:            conversion.StringPtr(os.Getenv("AZURE_SECRET")),
			TenantID:          conversion.StringPtr(os.Getenv("AZURE_TENANT_ID")),
		}
	)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { mig.PreCheck(t); acc.PreCheckEncryptionAtRestEnvAzure(t) },
		CheckDestroy: testAccCheckMongoDBAtlasEncryptionAtRestDestroy,
		Steps: []resource.TestStep{
			{
				ExternalProviders: mig.ExternalProviders(),
				Config:            acc.ConfigEARAzureKeyVault(projectID, &azureKeyVault, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckMongoDBAtlasEncryptionAtRestExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "project_id", projectID),
					resource.TestCheckResourceAttr(resourceName, "azure_key_vault_config.0.enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "azure_key_vault_config.0.azure_environment", azureKeyVault.GetAzureEnvironment()),
					resource.TestCheckResourceAttr(resourceName, "azure_key_vault_config.0.resource_group_name", azureKeyVault.GetResourceGroupName()),
					resource.TestCheckResourceAttr(resourceName, "azure_key_vault_config.0.key_vault_name", azureKeyVault.GetKeyVaultName()),
				),
			},
			{
				ProtoV6ProviderFactories: acc.TestAccProviderV6Factories,
				Config:                   acc.ConfigEARAzureKeyVault(projectID, &azureKeyVault, false),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						acc.DebugPlan(),
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
		},
	})
}

func TestMigEncryptionAtRest_basicGCP(t *testing.T) {
	acc.SkipTestForCI(t) // needs GCP configuration

	var (
		resourceName = "mongodbatlas_encryption_at_rest.test"
		projectID    = os.Getenv("MONGODB_ATLAS_PROJECT_ID")

		googleCloudKms = admin.GoogleCloudKMS{
			Enabled:              conversion.Pointer(true),
			ServiceAccountKey:    conversion.StringPtr(os.Getenv("GCP_SERVICE_ACCOUNT_KEY")),
			KeyVersionResourceID: conversion.StringPtr(os.Getenv("GCP_KEY_VERSION_RESOURCE_ID")),
		}
	)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { mig.PreCheck(t); acc.PreCheckGPCEnv(t) },
		CheckDestroy: testAccCheckMongoDBAtlasEncryptionAtRestDestroy,
		Steps: []resource.TestStep{
			{
				ExternalProviders: mig.ExternalProviders(),
				Config:            testAccMongoDBAtlasEncryptionAtRestConfigGoogleCloudKms(projectID, &googleCloudKms),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckMongoDBAtlasEncryptionAtRestExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "project_id", projectID),
					resource.TestCheckResourceAttr(resourceName, "google_cloud_kms_config.0.enabled", "true"),
				),
			},
			{
				ProtoV6ProviderFactories: acc.TestAccProviderV6Factories,
				Config:                   testAccMongoDBAtlasEncryptionAtRestConfigGoogleCloudKms(projectID, &googleCloudKms),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						acc.DebugPlan(),
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
		},
	})
}

func TestMigEncryptionAtRest_basicAWS_from_v1_11_0(t *testing.T) {
	acc.SkipTestForCI(t) // needs AWS configuration

	var (
		resourceName = "mongodbatlas_encryption_at_rest.test"
		projectID    = os.Getenv("MONGODB_ATLAS_PROJECT_ID")

		awsKms = admin.AWSKMSConfiguration{
			Enabled:             conversion.Pointer(true),
			AccessKeyID:         conversion.StringPtr(os.Getenv("AWS_ACCESS_KEY_ID")),
			SecretAccessKey:     conversion.StringPtr(os.Getenv("AWS_SECRET_ACCESS_KEY")),
			CustomerMasterKeyID: conversion.StringPtr(os.Getenv("AWS_CUSTOMER_MASTER_KEY_ID")),
			Region:              conversion.StringPtr(conversion.AWSRegionToMongoDBRegion(os.Getenv("AWS_REGION"))),
			RoleId:              conversion.StringPtr(os.Getenv("AWS_ROLE_ID")),
		}
	)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acc.PreCheck(t); acc.PreCheckAwsEnv(t) },
		CheckDestroy: testAccCheckMongoDBAtlasEncryptionAtRestDestroy,
		Steps: []resource.TestStep{
			{
				ExternalProviders: acc.ExternalProvidersWithAWS("1.11.0"),
				Config:            testAccMongoDBAtlasEncryptionAtRestConfigAwsKms(projectID, &awsKms),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckMongoDBAtlasEncryptionAtRestExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "project_id", projectID),
					resource.TestCheckResourceAttr(resourceName, "aws_kms_config.0.enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "aws_kms_config.0.region", awsKms.GetRegion()),
					resource.TestCheckResourceAttr(resourceName, "aws_kms_config.0.role_id", awsKms.GetRoleId()),
				),
			},
			{
				ProtoV6ProviderFactories: acc.TestAccProviderV6Factories,
				Config:                   testAccMongoDBAtlasEncryptionAtRestConfigAwsKms(projectID, &awsKms),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						acc.DebugPlan(),
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
		},
	})
}