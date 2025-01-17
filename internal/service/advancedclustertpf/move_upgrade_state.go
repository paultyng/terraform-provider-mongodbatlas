package advancedclustertpf

import (
	"context"
	"fmt"
	"math/big"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/mongodb/terraform-provider-mongodbatlas/internal/common/conversion"
	"go.mongodb.org/atlas-sdk/v20241113004/admin"
)

// MoveState is used with moved block to upgrade from cluster to adv_cluster
func (r *rs) MoveState(context.Context) []resource.StateMover {
	return []resource.StateMover{{StateMover: stateMover}}
}

// UpgradeState is used to upgrade from adv_cluster schema v1 (SDKv2) to v2 (TPF)
func (r *rs) UpgradeState(ctx context.Context) map[int64]resource.StateUpgrader {
	return map[int64]resource.StateUpgrader{
		1: {StateUpgrader: stateUpgraderFromV1},
	}
}

func stateMover(ctx context.Context, req resource.MoveStateRequest, resp *resource.MoveStateResponse) {
	if req.SourceTypeName != "mongodbatlas_cluster" || !strings.HasSuffix(req.SourceProviderAddress, "/mongodbatlas") {
		return
	}
	setStateResponse(ctx, &resp.Diagnostics, req.SourceRawState, &resp.TargetState)
}

func stateUpgraderFromV1(ctx context.Context, req resource.UpgradeStateRequest, resp *resource.UpgradeStateResponse) {
	setStateResponse(ctx, &resp.Diagnostics, req.RawState, &resp.State)
}

// Minimum attributes needed from source schema. Read will fill in the rest
var stateAttrs = map[string]tftypes.Type{
	"project_id":             tftypes.String, // project_id and name to identify the cluster
	"name":                   tftypes.String,
	"retain_backups_enabled": tftypes.Bool,   // TF specific so can't be got in Read
	"mongo_db_major_version": tftypes.String, // Has special logic in overrideAttributesWithPrevStateValue that needs the previous state
	"timeouts": tftypes.Object{ // TF specific so can't be got in Read
		AttributeTypes: map[string]tftypes.Type{
			"create": tftypes.String,
			"update": tftypes.String,
			"delete": tftypes.String,
		},
	},
	"replication_specs": tftypes.List{ // Needed to check if some num_shards are > 1 so we need to force legacy schema
		ElementType: tftypes.Object{
			AttributeTypes: map[string]tftypes.Type{
				"num_shards": tftypes.Number,
			},
		},
	},
}

func setStateResponse(ctx context.Context, diags *diag.Diagnostics, stateIn *tfprotov6.RawState, stateOut *tfsdk.State) {
	rawStateValue, err := stateIn.UnmarshalWithOpts(tftypes.Object{
		AttributeTypes: stateAttrs,
	}, tfprotov6.UnmarshalOpts{ValueFromJSONOpts: tftypes.ValueFromJSONOpts{IgnoreUndefinedAttributes: true}})
	if err != nil {
		diags.AddError("Unable to Unmarshal state", err.Error())
		return
	}
	var stateObj map[string]tftypes.Value
	if err := rawStateValue.As(&stateObj); err != nil {
		diags.AddError("Unable to Parse state", err.Error())
		return
	}
	projectID, name := getProjectIDNameFromStateObj(diags, stateObj)
	if diags.HasError() {
		return
	}
	model := NewTFModel(ctx, &admin.ClusterDescription20240805{
		GroupId: projectID,
		Name:    name,
	}, getTimeoutFromStateObj(stateObj), diags, ExtraAPIInfo{})
	if diags.HasError() {
		return
	}
	AddAdvancedConfig(ctx, model, nil, nil, diags)
	if diags.HasError() {
		return
	}
	setOptionalModelAttrs(stateObj, model)
	diags.Append(stateOut.Set(ctx, model)...)
}

func getAttrFromStateObj[T any](rawState map[string]tftypes.Value, attrName string) *T {
	var ret *T
	if err := rawState[attrName].As(&ret); err != nil {
		return nil
	}
	return ret
}

func getProjectIDNameFromStateObj(diags *diag.Diagnostics, stateObj map[string]tftypes.Value) (projectID, name *string) {
	projectID = getAttrFromStateObj[string](stateObj, "project_id")
	name = getAttrFromStateObj[string](stateObj, "name")
	if !conversion.IsStringPresent(projectID) || !conversion.IsStringPresent(name) {
		diags.AddError("Unable to read project_id or name from state", fmt.Sprintf("project_id: %s, name: %s",
			conversion.SafeString(projectID), conversion.SafeString(name)))
		return
	}
	return projectID, name
}

func getTimeoutFromStateObj(stateObj map[string]tftypes.Value) timeouts.Value {
	attrTypes := map[string]attr.Type{
		"create": types.StringType,
		"update": types.StringType,
		"delete": types.StringType,
	}
	nullObj := timeouts.Value{Object: types.ObjectNull(attrTypes)}
	timeoutState := getAttrFromStateObj[map[string]tftypes.Value](stateObj, "timeouts")
	if timeoutState == nil {
		return nullObj
	}
	timeoutMap := make(map[string]attr.Value)
	for action := range attrTypes {
		actionTimeout := getAttrFromStateObj[string](*timeoutState, action)
		if actionTimeout == nil {
			timeoutMap[action] = types.StringNull()
		} else {
			timeoutMap[action] = types.StringPointerValue(actionTimeout)
		}
	}
	obj, d := types.ObjectValue(attrTypes, timeoutMap)
	if d.HasError() {
		return nullObj
	}
	return timeouts.Value{Object: obj}
}

func setOptionalModelAttrs(stateObj map[string]tftypes.Value, model *TFModel) {
	if retainBackupsEnabled := getAttrFromStateObj[bool](stateObj, "retain_backups_enabled"); retainBackupsEnabled != nil {
		model.RetainBackupsEnabled = types.BoolPointerValue(retainBackupsEnabled)
	}
	if mongoDBMajorVersion := getAttrFromStateObj[string](stateObj, "mongo_db_major_version"); mongoDBMajorVersion != nil {
		model.MongoDBMajorVersion = types.StringPointerValue(mongoDBMajorVersion)
	}
	if isLegacySchemaState(stateObj) {
		sendLegacySchemaRequestToRead(model)
	}
}

func isLegacySchemaState(stateObj map[string]tftypes.Value) bool {
	one := big.NewFloat(1.0)
	specsVal := getAttrFromStateObj[[]tftypes.Value](stateObj, "replication_specs")
	if specsVal == nil {
		return false
	}
	for _, specVal := range *specsVal {
		var specObj map[string]tftypes.Value
		if err := specVal.As(&specObj); err != nil {
			return false
		}
		numShardsVal := specObj["num_shards"]
		var numShards *big.Float
		if err := numShardsVal.As(&numShards); err != nil || numShards == nil {
			return false
		}
		if numShards.Cmp(one) > 0 { // legacy schema if numShards > 1
			return true
		}
	}
	return false
}

// sendLegacySchemaRequestToRead sets ClusterID to a special value so Read can know whether it must use legacy schema.
// private state can't be used here because it's not available in Move Upgrader.
// ClusterID is computed (not optional) so the value will be overridden in Read and the special value won't ever appear in the state file.
func sendLegacySchemaRequestToRead(model *TFModel) {
	model.ClusterID = types.StringValue("forceLegacySchema")
}

// receivedLegacySchemaRequestInRead checks if Read has to use the legacy schema because a State Move or Upgrader happened just before.
func receivedLegacySchemaRequestInRead(model *TFModel) bool {
	return model.ClusterID.ValueString() == "forceLegacySchema"
}
