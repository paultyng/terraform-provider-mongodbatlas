package acc

import (
	"os"
	"strings"
	"testing"
	"time"

	"github.com/mongodb/terraform-provider-mongodbatlas/internal/common/conversion"
	"github.com/mongodb/terraform-provider-mongodbatlas/internal/config"
)

func PreCheckBasic(tb testing.TB) {
	tb.Helper()
	if os.Getenv("MONGODB_ATLAS_PUBLIC_KEY") == "" ||
		os.Getenv("MONGODB_ATLAS_PRIVATE_KEY") == "" ||
		os.Getenv("MONGODB_ATLAS_ORG_ID") == "" {
		tb.Fatal("`MONGODB_ATLAS_PUBLIC_KEY`, `MONGODB_ATLAS_PRIVATE_KEY`, and `MONGODB_ATLAS_ORG_ID` must be set for acceptance testing")
	}
}

func SkipIfAdvancedClusterV2Schema(tb testing.TB) {
	tb.Helper()
	if config.AdvancedClusterV2Schema() {
		tb.Skip("Skipping test in AdvancedClusterV2Schema as implementation is pending or test is not applicable")
	}
}

// PreCheckBasicSleep is a helper function to call SerialSleep, see its help for more info.
// Some examples of use are when the test is calling ProjectIDExecution or GetClusterInfo to create clusters.
func PreCheckBasicSleep(tb testing.TB, clusterInfo *ClusterInfo, projectID, clusterName string) func() {
	tb.Helper()
	return func() {
		PreCheckBasic(tb)
		SerialSleep(tb)
		if clusterInfo != nil {
			projectID = clusterInfo.ProjectID
			clusterName = clusterInfo.Name
		}
		tb.Logf("Time before creating cluster: %s, ProjectID: %s, Cluster name: %s", conversion.TimeToString(time.Now()), projectID, clusterName)
	}
}

// PreCheck checks common Atlas environment variables and MONGODB_ATLAS_PROJECT_ID.
// Deprecated: it should not be used as MONGODB_ATLAS_PROJECT_ID is not intended to be used in CI.
// Use PreCheckBasic instead.
func PreCheck(tb testing.TB) {
	tb.Helper()
	if os.Getenv("MONGODB_ATLAS_PUBLIC_KEY") == "" ||
		os.Getenv("MONGODB_ATLAS_PRIVATE_KEY") == "" ||
		os.Getenv("MONGODB_ATLAS_PROJECT_ID") == "" ||
		os.Getenv("MONGODB_ATLAS_ORG_ID") == "" {
		tb.Fatal("`MONGODB_ATLAS_PUBLIC_KEY`, `MONGODB_ATLAS_PRIVATE_KEY`, `MONGODB_ATLAS_PROJECT_ID` and `MONGODB_ATLAS_ORG_ID` must be set for acceptance testing")
	}
}

func PreCheckEncryptionAtRestPrivateEndpoint(tb testing.TB) {
	tb.Helper()
	if os.Getenv("MONGODB_ATLAS_PUBLIC_KEY") == "" ||
		os.Getenv("MONGODB_ATLAS_PRIVATE_KEY") == "" ||
		os.Getenv("MONGODB_ATLAS_PROJECT_EAR_PE_ID") == "" ||
		os.Getenv("MONGODB_ATLAS_ORG_ID") == "" {
		tb.Fatal("`MONGODB_ATLAS_PUBLIC_KEY`, `MONGODB_ATLAS_PRIVATE_KEY`, `MONGODB_ATLAS_PROJECT_EAR_PE_ID` and `MONGODB_ATLAS_ORG_ID` must be set for acceptance testing")
	}
}

func PreCheckCert(tb testing.TB) {
	tb.Helper()
	if os.Getenv("MONGODB_ATLAS_PUBLIC_KEY") == "" ||
		os.Getenv("MONGODB_ATLAS_PRIVATE_KEY") == "" ||
		os.Getenv("MONGODB_ATLAS_ORG_ID") == "" ||
		os.Getenv("CA_CERT") == "" {
		tb.Fatal("`CA_CERT, MONGODB_ATLAS_PUBLIC_KEY`, `MONGODB_ATLAS_PRIVATE_KEY`, and `MONGODB_ATLAS_ORG_ID` must be set for acceptance testing")
	}
}

// PreCheckPreviewFlag is used for resources not yet in GA
func PreCheckPreviewFlag(tb testing.TB) {
	tb.Helper()
	if os.Getenv("MONGODB_ATLAS_ENABLE_PREVIEW") == "" {
		tb.Fatal("`MONGODB_ATLAS_ENABLE_PREVIEW` must be set for running this acceptance test")
	}
}

func PreCheckCloudProviderAccessAzure(tb testing.TB) {
	tb.Helper()
	PreCheckBasic(tb)
	if os.Getenv("AZURE_ATLAS_APP_ID") == "" ||
		os.Getenv("AZURE_SERVICE_PRINCIPAL_ID") == "" ||
		os.Getenv("AZURE_TENANT_ID") == "" {
		tb.Fatal("`AZURE_ATLAS_APP_ID`, `AZURE_SERVICE_PRINCIPAL_ID`, and `AZURE_TENANT_ID` must be set for acceptance testing")
	}
}

func PreCheckBasicOwnerID(tb testing.TB) {
	tb.Helper()
	PreCheckBasic(tb)
	if os.Getenv("MONGODB_ATLAS_PROJECT_OWNER_ID") == "" {
		tb.Fatal("`MONGODB_ATLAS_PROJECT_OWNER_ID` must be set ")
	}
}

func PreCheckAtlasUsername(tb testing.TB) {
	tb.Helper()
	PreCheckBasic(tb)
	if os.Getenv("MONGODB_ATLAS_USERNAME") == "" {
		tb.Fatal("`MONGODB_ATLAS_USERNAME` must be set ")
	}
}

func PreCheckAtlasUsernames(tb testing.TB) {
	tb.Helper()
	PreCheckAtlasUsername(tb)
	if os.Getenv("MONGODB_ATLAS_USERNAME_2") == "" {
		tb.Fatal("`MONGODB_ATLAS_USERNAME_2` must be set")
	}
}

func PreCheckProjectTeamsIDsWithMinCount(tb testing.TB, minTeamsCount int) {
	tb.Helper()
	envVar := os.Getenv("MONGODB_ATLAS_TEAMS_IDS")
	if envVar == "" {
		tb.Fatal("`MONGODB_ATLAS_TEAMS_IDS` must be set for Projects acceptance testing")
		return
	}
	teamsIDs := strings.Split(envVar, ",")
	if count := len(teamsIDs); count < minTeamsCount {
		tb.Fatalf("`MONGODB_ATLAS_TEAMS_IDS` must have at least %d team ids for this acceptance testing, has %d", minTeamsCount, count)
	}
}

func GetProjectTeamsIDsWithPos(pos int) string {
	envVar := os.Getenv("MONGODB_ATLAS_TEAMS_IDS")
	teamsIDs := strings.Split(envVar, ",")
	count := len(teamsIDs)
	if envVar == "" || pos >= count {
		return ""
	}
	return teamsIDs[pos]
}

func PreCheckGovBasic(tb testing.TB) {
	tb.Helper()
	if os.Getenv("MONGODB_ATLAS_GOV_PUBLIC_KEY") == "" ||
		os.Getenv("MONGODB_ATLAS_GOV_PRIVATE_KEY") == "" ||
		os.Getenv("MONGODB_ATLAS_GOV_BASE_URL") == "" ||
		os.Getenv("MONGODB_ATLAS_GOV_ORG_ID") == "" {
		tb.Fatal("`MONGODB_ATLAS_GOV_BASE_URL`, `MONGODB_ATLAS_GOV_PUBLIC_KEY`, `MONGODB_ATLAS_GOV_PRIVATE_KEY`and `MONGODB_ATLAS_GOV_ORG_ID` must be set for acceptance testing")
	}
}
func PreCheckPublicKey2(tb testing.TB) {
	tb.Helper()
	if os.Getenv("MONGODB_ATLAS_PUBLIC_KEY_READ_ONLY") == "" ||
		os.Getenv("MONGODB_ATLAS_PRIVATE_KEY_READ_ONLY") == "" {
		tb.Fatal("`MONGODB_ATLAS_PUBLIC_KEY_READ_ONLY`, `MONGODB_ATLAS_PRIVATE_KEY_READ_ONLY` must be set for acceptance testing")
	}
}

func PreCheckGPCEnv(tb testing.TB) {
	tb.Helper()
	if os.Getenv("GCP_SERVICE_ACCOUNT_KEY") == "" || os.Getenv("GCP_KEY_VERSION_RESOURCE_ID") == "" {
		tb.Fatal("`GCP_SERVICE_ACCOUNT_KEY` and `GCP_KEY_VERSION_RESOURCE_ID` must be set for acceptance testing")
	}
}

func PreCheckPeeringEnvAWS(tb testing.TB) {
	tb.Helper()
	if os.Getenv("AWS_ACCOUNT_ID") == "" ||
		os.Getenv("AWS_VPC_ID") == "" ||
		os.Getenv("AWS_VPC_CIDR_BLOCK") == "" ||
		os.Getenv("AWS_REGION") == "" {
		tb.Fatal("`AWS_ACCOUNT_ID`, `AWS_VPC_ID`, `AWS_VPC_CIDR_BLOCK` and `AWS_VPC_ID` must be set for  network peering acceptance testing")
	}
	PreCheckBasic(tb)
}

func PreCheckPeeringEnvAzure(tb testing.TB) {
	tb.Helper()
	if os.Getenv("AZURE_DIRECTORY_ID") == "" ||
		os.Getenv("AZURE_SUBSCRIPTION_ID") == "" ||
		os.Getenv("AZURE_VNET_NAME") == "" ||
		os.Getenv("AZURE_RESOURCE_GROUP_NAME") == "" {
		tb.Fatal("`AZURE_DIRECTORY_ID`, `AZURE_SUBSCRIPTION_ID`, `AZURE_VNET_NAME` and `AZURE_RESOURCE_GROUP_NAME` must be set for  network peering acceptance testing")
	}
}

func PreCheckEncryptionAtRestEnvAzure(tb testing.TB) {
	tb.Helper()
	if os.Getenv("AZURE_CLIENT_ID") == "" ||
		os.Getenv("AZURE_SUBSCRIPTION_ID") == "" ||
		os.Getenv("AZURE_RESOURCE_GROUP_NAME") == "" ||
		os.Getenv("AZURE_APP_SECRET") == "" ||
		os.Getenv("AZURE_KEY_VAULT_NAME") == "" ||
		os.Getenv("AZURE_KEY_IDENTIFIER") == "" ||
		os.Getenv("AZURE_TENANT_ID") == "" {
		tb.Fatal(`'AZURE_CLIENT_ID', 'AZURE_SUBSCRIPTION_ID',
		'AZURE_RESOURCE_GROUP_NAME', 'AZURE_APP_SECRET', 'AZURE_KEY_VAULT_NAME', 'AZURE_KEY_IDENTIFIER', and 'AZURE_TENANT_ID' must be set for Encryption At Rest acceptance testing`)
	}
}

func PreCheckEncryptionAtRestEnvAzureWithUpdate(tb testing.TB) {
	tb.Helper()
	PreCheckEncryptionAtRestEnvAzure(tb)

	if os.Getenv("AZURE_KEY_VAULT_NAME_UPDATED") == "" ||
		os.Getenv("AZURE_KEY_IDENTIFIER_UPDATED") == "" {
		tb.Fatal(`'AZURE_CLIENT_ID', 'AZURE_SUBSCRIPTION_ID',
		'AZURE_RESOURCE_GROUP_NAME', 'AZURE_APP_SECRET',
		, 'AZURE_KEY_VAULT_NAME', 'AZURE_KEY_IDENTIFIER', 'AZURE_KEY_VAULT_NAME_UPDATED',
		'AZURE_KEY_IDENTIFIER_UPDATED', and 'AZURE_TENANT_ID' must be set for Encryption At Rest acceptance testing`)
	}
}

func PreCheckPeeringEnvGCP(tb testing.TB) {
	tb.Helper()
	if os.Getenv("GCP_PROJECT_ID") == "" ||
		os.Getenv("GCP_CLUSTER_REGION_NAME") == "" ||
		os.Getenv("GCP_REGION_NAME") == "" ||
		os.Getenv("GOOGLE_CLOUD_KEYFILE_JSON") == "" {
		tb.Fatal("`GCP_PROJECT_ID`,`GOOGLE_CLOUD_KEYFILE_JSON`, `GCP_CLUSTER_REGION_NAME`, `and GCP_REGION_NAME` must be set for network peering acceptance testing")
	}
}

func PreCheckAwsEnvBasic(tb testing.TB) {
	tb.Helper()
	if os.Getenv("AWS_ACCESS_KEY_ID") == "" ||
		os.Getenv("AWS_SECRET_ACCESS_KEY") == "" {
		tb.Fatal("`AWS_ACCESS_KEY_ID` and `AWS_SECRET_ACCESS_KEY` must be set for acceptance testing")
	}
}

func PreCheckAwsEnv(tb testing.TB) {
	tb.Helper()
	if os.Getenv("AWS_ACCESS_KEY_ID") == "" ||
		os.Getenv("AWS_SECRET_ACCESS_KEY") == "" ||
		os.Getenv("AWS_CUSTOMER_MASTER_KEY_ID") == "" {
		tb.Fatal("`AWS_ACCESS_KEY_ID`, `AWS_VPC_ID`, `AWS_SECRET_ACCESS_KEY` and `AWS_CUSTOMER_MASTER_KEY_ID` must be set for acceptance testing")
	}
}

func PreCheckAwsRegionCases(tb testing.TB) {
	tb.Helper()
	if os.Getenv("AWS_REGION_UPPERCASE") == "" ||
		os.Getenv("AWS_REGION_LOWERCASE") == "" {
		tb.Fatal("`AWS_REGION_UPPERCASE`, `AWS_REGION_LOWERCASE` must be set for acceptance testing")
	}
}

func PreCheckAwsEnvPrivateLinkEndpointService(tb testing.TB) {
	tb.Helper()
	if os.Getenv("AWS_ACCESS_KEY_ID") == "" ||
		os.Getenv("AWS_SECRET_ACCESS_KEY") == "" ||
		os.Getenv("AWS_VPC_ID") == "" {
		tb.Fatal("`AWS_ACCESS_KEY_ID`, `AWS_VPC_ID` and `AWS_SECRET_ACCESS_KEY` must be set for acceptance testing")
	}
}

func PreCheckRegularCredsAreEmpty(tb testing.TB) {
	tb.Helper()
	if os.Getenv("MONGODB_ATLAS_PUBLIC_KEY") != "" || os.Getenv("MONGODB_ATLAS_PRIVATE_KEY") != "" {
		tb.Fatal(`"MONGODB_ATLAS_PUBLIC_KEY" and "MONGODB_ATLAS_PRIVATE_KEY" are defined in this test and they should not.`)
	}
}

func PreCheckSTSAssumeRole(tb testing.TB) {
	tb.Helper()
	if os.Getenv("AWS_REGION") == "" {
		tb.Fatal(`'AWS_REGION' must be set for acceptance testing with STS Assume Role.`)
	}
	if os.Getenv("STS_ENDPOINT") == "" {
		tb.Fatal(`'STS_ENDPOINT' must be set for acceptance testing with STS Assume Role.`)
	}
	if os.Getenv("ASSUME_ROLE_ARN") == "" {
		tb.Fatal(`'ASSUME_ROLE_ARN' must be set for acceptance testing with STS Assume Role.`)
	}
	if os.Getenv("AWS_ACCESS_KEY_ID") == "" {
		tb.Fatal(`'AWS_ACCESS_KEY_ID' must be set for acceptance testing with STS Assume Role.`)
	}
	if os.Getenv("AWS_SECRET_ACCESS_KEY") == "" {
		tb.Fatal(`'AWS_SECRET_ACCESS_KEY' must be set for acceptance testing with STS Assume Role.`)
	}
	if os.Getenv("AWS_SESSION_TOKEN") == "" {
		tb.Fatal(`'AWS_SESSION_TOKEN' must be set for acceptance testing with STS Assume Role.`)
	}
	if os.Getenv("SECRET_NAME") == "" {
		tb.Fatal(`'SECRET_NAME' must be set for acceptance testing with STS Assume Role.`)
	}
}

func PreCheckDataLakePipelineRun(tb testing.TB) {
	tb.Helper()
	if os.Getenv("MONGODB_ATLAS_DATA_LAKE_PIPELINE_RUN_ID") == "" {
		tb.Fatal("`MONGODB_ATLAS_DATA_LAKE_PIPELINE_RUN_ID` must be set for Projects acceptance testing")
	}
	PreCheckDataLakePipelineRuns(tb)
}

func PreCheckDataLakePipelineRuns(tb testing.TB) {
	tb.Helper()
	if os.Getenv("MONGODB_ATLAS_DATA_LAKE_PIPELINE_NAME") == "" {
		tb.Fatal("`MONGODB_ATLAS_DATA_LAKE_PIPELINE_NAME` must be set for Projects acceptance testing")
	}
	PreCheck(tb)
}

func PreCheckLDAP(tb testing.TB) {
	tb.Helper()
	if os.Getenv("MONGODB_ATLAS_LDAP_HOSTNAME") == "" ||
		os.Getenv("MONGODB_ATLAS_LDAP_USERNAME") == "" ||
		os.Getenv("MONGODB_ATLAS_LDAP_PASSWORD") == "" ||
		os.Getenv("MONGODB_ATLAS_LDAP_PORT") == "" {
		tb.Fatal("`MONGODB_ATLAS_LDAP_HOSTNAME`, `MONGODB_ATLAS_LDAP_USERNAME`, `MONGODB_ATLAS_LDAP_PASSWORD` and `MONGODB_ATLAS_LDAP_PORT` must be set for ldap configuration/verify acceptance testing")
	}
	PreCheckBasic(tb)
}

func PreCheckLDAPCert(tb testing.TB) {
	tb.Helper()
	if os.Getenv("MONGODB_ATLAS_LDAP_CA_CERTIFICATE") == "" {
		tb.Fatal("`MONGODB_ATLAS_LDAP_CA_CERTIFICATE` must be set")
	}
	PreCheckLDAP(tb)
}

func PreCheckFederatedSettings(tb testing.TB) {
	tb.Helper()
	if os.Getenv("MONGODB_ATLAS_FEDERATED_ORG_ID") == "" ||
		os.Getenv("MONGODB_ATLAS_FEDERATION_SETTINGS_ID") == "" {
		tb.Fatal("`MONGODB_ATLAS_FEDERATED_ORG_ID` and `MONGODB_ATLAS_FEDERATION_SETTINGS_ID` must be set for federated settings/verify acceptance testing")
	}
}

func PreCheckFederatedSettingsIdentityProvider(tb testing.TB) {
	tb.Helper()
	if os.Getenv("MONGODB_ATLAS_FEDERATED_ORG_ID") == "" ||
		os.Getenv("MONGODB_ATLAS_FEDERATED_SETTINGS_ASSOCIATED_DOMAIN") == "" ||
		os.Getenv("MONGODB_ATLAS_FEDERATION_SETTINGS_ID") == "" {
		tb.Fatal("`MONGODB_ATLAS_FEDERATED_SETTINGS_ASSOCIATED_DOMAIN`, MONGODB_ATLAS_FEDERATED_ORG_ID`, and `MONGODB_ATLAS_FEDERATION_SETTINGS_ID` must be set for federated settings acceptance testing")
	}
}

func PreCheckPrivateEndpoint(tb testing.TB) {
	tb.Helper()
	if os.Getenv("MONGODB_ATLAS_PRIVATE_ENDPOINT_ID") == "" ||
		os.Getenv("MONGODB_ATLAS_PRIVATE_ENDPOINT_DNS_NAME") == "" {
		tb.Fatal("`MONGODB_ATLAS_PRIVATE_ENDPOINT_ID` and `MONGODB_ATLAS_PRIVATE_ENDPOINT_DNS_NAME`must be set for Private Endpoint Service Data Federation and Online Archive acceptance testing")
	}
	PreCheckBasic(tb)
}

func PreCheckS3Bucket(tb testing.TB) {
	tb.Helper()
	if os.Getenv("AWS_S3_BUCKET") == "" {
		tb.Fatal("`AWS_S3_BUCKET` must be set ")
	}
}

func PreCheckAzureExportBucket(tb testing.TB) {
	tb.Helper()
	if os.Getenv("AZURE_SERVICE_URL") == "" ||
		os.Getenv("AZURE_BLOB_STORAGE_CONTAINER_NAME") == "" {
		tb.Fatal("`AZURE_SERVICE_URL` and `AZURE_SERVICE_URL`must be set for Cloud Backup Snapshot Export Bucket acceptance testing")
	}
}

func PreCheckConfluentAWSPrivatelink(tb testing.TB) {
	tb.Helper()
	if os.Getenv("CONFLUENT_CLOUD_API_KEY") == "" ||
		os.Getenv("CONFLUENT_CLOUD_API_SECRET") == "" ||
		os.Getenv("CONFLUENT_CLOUD_NETWORK_ID") == "" ||
		os.Getenv("CONFLUENT_CLOUD_PRIVATELINK_ACCESS_ID") == "" {
		tb.Fatal("`CONFLUENT_CLOUD_API_KEY`, `CONFLUENT_CLOUD_API_SECRET`, `CONFLUENT_CLOUD_NETWORK_ID`, and `CONFLUENT_CLOUD_PRIVATELINK_ACCESS_ID` must be set for Cloud Backup Snapshot Export Bucket acceptance testing")
	}
}
