package mongodbatlas

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/mongodb/terraform-provider-mongodbatlas/internal/testutil/acc"
)

func testAccPreCheckBasic(tb testing.TB) {
	acc.PreCheckBasic(tb)
}

func testAccPreCheck(tb testing.TB) {
	acc.PreCheck(tb)
}

func SkipTest(tb testing.TB) {
	acc.SkipTest(tb)
}

func SkipIfTFAccNotDefined(tb testing.TB) {
	acc.SkipIfTFAccNotDefined(tb)
}

func SkipTestForCI(tb testing.TB) {
	acc.SkipTestForCI(tb)
}

// TODO INITIALIZE OR LINK TO INTERNAL ************
// TODO INITIALIZE OR LINK TO INTERNAL ************
var testAccProviderV6Factories map[string]func() (tfprotov6.ProviderServer, error)
var testAccProviderSdkV2 *schema.Provider
var testMongoDBClient any
