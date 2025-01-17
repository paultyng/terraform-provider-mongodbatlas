package streamprivatelinkendpoint_test

import (
	"os"
	"testing"

	"github.com/mongodb/terraform-provider-mongodbatlas/internal/testutil/acc"
)

func TestMain(m *testing.M) {
	cleanup := acc.SetupSharedResources()
	exitCode := m.Run()
	cleanup()
	os.Exit(exitCode)
}
