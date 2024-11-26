package acc

import (
	"testing"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func ConvertAdvancedClusterToTPF(def string) string {
	if !IsTPFAdvancedCluster() {
		return def
	}
	return "invalid resource"
}

func AssertEqualHCL(t *testing.T, expected, actual string, msgAndArgs ...interface{}) {
	t.Helper()
	assert.Equal(t, CanonicalHCL(t, expected), CanonicalHCL(t, actual), msgAndArgs...)
}

func CanonicalHCL(t *testing.T, def string) string {
	t.Helper()
	parse, diags := hclwrite.ParseConfig([]byte(def), "", hcl.Pos{Line: 1, Column: 1})
	require.False(t, diags.HasErrors())
	return string(parse.Bytes())
}
