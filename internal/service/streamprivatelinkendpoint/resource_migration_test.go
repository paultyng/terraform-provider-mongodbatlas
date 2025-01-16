package streamprivatelinkendpoint_test

import (
	"testing"

	"github.com/mongodb/terraform-provider-mongodbatlas/internal/testutil/acc"
	"github.com/mongodb/terraform-provider-mongodbatlas/internal/testutil/mig"
)

func TestMigStreamPrivatelinkEndpoint_basic(t *testing.T) {
	acc.SkipTestForCI(t)                // needs confluent cloud resources
	mig.SkipIfVersionBelow(t, "1.25.0") // when resource 1st released
	mig.CreateTestAndRunUseExternalProviderNonParallel(t, basicTestCase(t), mig.ExternalProvidersWithConfluent(), nil)
}
