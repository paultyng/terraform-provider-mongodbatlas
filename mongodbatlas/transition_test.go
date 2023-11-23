package mongodbatlas_test

import (
	"os"
	"testing"

	"github.com/mongodb/terraform-provider-mongodbatlas/internal/config"
	"github.com/mongodb/terraform-provider-mongodbatlas/internal/testutil/acc"
)

type MongoDBClient = config.MongoDBClient

const (
	errorGetRead = "error reading cloud provider access %s"
)

func decodeStateID(stateID string) map[string]string {
	return config.DecodeStateID(stateID)
}

func encodeStateID(values map[string]string) string {
	return config.EncodeStateID(values)
}

func testAccPreCheckBasic(tb testing.TB) {
	acc.PreCheckBasic(tb)
}

func testCheckAwsEnv(tb testing.TB) {
	acc.PreCheckAwsEnv(tb)
}

func testAccPreCheck(tb testing.TB) {
	acc.PreCheck(tb)
}

func testCheckPeeringEnvGCP(tb testing.TB) {
	acc.PreCheckPeeringEnvGCP(tb)
}

func testCheckPeeringEnvAzure(tb testing.TB) {
	acc.PreCheckPeeringEnvAzure(tb)
}

func testAccPreCheckCloudProviderAccessAzure(tb testing.TB) {
	acc.PreCheckCloudProviderAccessAzure(tb)
}

func SkipTest(tb testing.TB) {
	acc.SkipTest(tb)
}

func SkipTestForCI(tb testing.TB) {
	acc.SkipTestForCI(tb)
}

func SkipTestExtCred(tb testing.TB) {
	acc.SkipTestExtCred(tb)
}

func testCheckDataLakePipelineRun(tb testing.TB) {
	if os.Getenv("MONGODB_ATLAS_DATA_LAKE_PIPELINE_RUN_ID") == "" {
		tb.Skip("`MONGODB_ATLAS_DATA_LAKE_PIPELINE_RUN_ID` must be set for Projects acceptance testing")
	}
	testCheckDataLakePipelineRuns(tb)
}

func testCheckDataLakePipelineRuns(tb testing.TB) {
	if os.Getenv("MONGODB_ATLAS_DATA_LAKE_PIPELINE_NAME") == "" {
		tb.Skip("`MONGODB_ATLAS_DATA_LAKE_PIPELINE_NAME` must be set for Projects acceptance testing")
	}
}

func testCheckLDAP(tb testing.TB) {
	if os.Getenv("MONGODB_ATLAS_LDAP_HOSTNAME") == "" ||
		os.Getenv("MONGODB_ATLAS_LDAP_USERNAME") == "" ||
		os.Getenv("MONGODB_ATLAS_LDAP_PASSWORD") == "" ||
		os.Getenv("MONGODB_ATLAS_LDAP_PORT") == "" {
		tb.Fatal("`MONGODB_ATLAS_LDAP_HOSTNAME`, `MONGODB_ATLAS_LDAP_USERNAME`, `MONGODB_ATLAS_LDAP_PASSWORD` and `MONGODB_ATLAS_LDAP_PORT` must be set for ldap configuration/verify acceptance testing")
	}
}

func testCheckFederatedSettings(tb testing.TB) {
	if os.Getenv("MONGODB_ATLAS_FEDERATED_PROJECT_ID") == "" ||
		os.Getenv("MONGODB_ATLAS_FEDERATION_SETTINGS_ID") == "" ||
		os.Getenv("MONGODB_ATLAS_FEDERATED_ORG_ID") == "" {
		tb.Fatal("`MONGODB_ATLAS_FEDERATED_PROJECT_ID`, `MONGODB_ATLAS_FEDERATED_ORG_ID` and `MONGODB_ATLAS_FEDERATION_SETTINGS_ID` must be set for federated settings/verify acceptance testing")
	}
}

func testCheckPrivateEndpointServiceDataFederationOnlineArchiveRun(tb testing.TB) {
	if os.Getenv("MONGODB_ATLAS_PRIVATE_ENDPOINT_ID") == "" {
		tb.Skip("`MONGODB_ATLAS_PRIVATE_ENDPOINT_ID` must be set for Private Endpoint Service Data Federation and Online Archive acceptance testing")
	}
}

func testAccPreCheckSearchIndex(tb testing.TB) {
	if os.Getenv("MONGODB_ATLAS_PUBLIC_KEY") == "" ||
		os.Getenv("MONGODB_ATLAS_PRIVATE_KEY") == "" ||
		os.Getenv("MONGODB_ATLAS_ORG_ID") == "" ||
		os.Getenv("MONGODB_ATLAS_PROJECT_ID") == "" {
		tb.Fatal("`MONGODB_ATLAS_PUBLIC_KEY`, `MONGODB_ATLAS_PRIVATE_KEY`,  `MONGODB_ATLAS_ORG_ID`, and `MONGODB_ATLAS_PROJECT_ID` must be set for acceptance testing")
	}
}

func testCheckPeeringEnvAWS(tb testing.TB) {
	acc.PreCheckPeeringEnvAWS(tb)
}

// TODO INITIALIZE OR LINK TO INTERNAL ************
// TODO INITIALIZE OR LINK TO INTERNAL ************

/*
type MongoDBClient = config.MongoDBClient

var acc.TestAccProviderV6Factories map[string]func() (tfprotov6.ProviderServer, error)
var acc.TestAccProviderSdkV2 *schema.Provider
var testMongoDBClient any
*/
