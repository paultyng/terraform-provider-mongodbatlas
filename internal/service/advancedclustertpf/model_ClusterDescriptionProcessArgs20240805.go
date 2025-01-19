package advancedclustertpf

import (
	"context"

	admin20240530 "go.mongodb.org/atlas-sdk/v20240530005/admin"
	"go.mongodb.org/atlas-sdk/v20241113004/admin"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/mongodb/terraform-provider-mongodbatlas/internal/common/conversion"
)

func AddAdvancedConfig(ctx context.Context, tfModel *TFModel, input *admin.ClusterDescriptionProcessArgs20240805, inputLegacy *admin20240530.ClusterDescriptionProcessArgs, diags *diag.Diagnostics) {
	var advancedConfig TFAdvancedConfigurationModel
	if input != nil && inputLegacy != nil {
		// Using the new API as source of Truth, only use `inputLegacy` for fields not in `input`
		changeStreamOptionsPreAndPostImagesExpireAfterSeconds := input.ChangeStreamOptionsPreAndPostImagesExpireAfterSeconds
		if changeStreamOptionsPreAndPostImagesExpireAfterSeconds == nil {
			// special behavior using -1 when it is unset by the user
			changeStreamOptionsPreAndPostImagesExpireAfterSeconds = conversion.Pointer(-1)
		}
		// When MongoDBMajorVersion is not 4.4 or lower, the API response for fail_index_key_too_long will always be null, to ensure no consistency issues, we need to match the config
		failIndexKeyTooLong := inputLegacy.GetFailIndexKeyTooLong()
		if tfModel != nil {
			stateConfig := tfModel.AdvancedConfiguration
			stateConfigSDK := NewAtlasReqAdvancedConfigurationLegacy(ctx, &stateConfig, diags)
			if diags.HasError() {
				return
			}
			if stateConfigSDK != nil && stateConfigSDK.GetFailIndexKeyTooLong() != failIndexKeyTooLong {
				failIndexKeyTooLong = stateConfigSDK.GetFailIndexKeyTooLong()
			}
		}
		advancedConfig = TFAdvancedConfigurationModel{
			ChangeStreamOptionsPreAndPostImagesExpireAfterSeconds: types.Int64PointerValue(conversion.IntPtrToInt64Ptr(changeStreamOptionsPreAndPostImagesExpireAfterSeconds)),
			DefaultWriteConcern:              types.StringValue(conversion.SafeValue(input.DefaultWriteConcern)),
			DefaultReadConcern:               types.StringValue(conversion.SafeValue(inputLegacy.DefaultReadConcern)),
			FailIndexKeyTooLong:              types.BoolValue(failIndexKeyTooLong),
			JavascriptEnabled:                types.BoolValue(conversion.SafeValue(input.JavascriptEnabled)),
			MinimumEnabledTlsProtocol:        types.StringValue(conversion.SafeValue(input.MinimumEnabledTlsProtocol)),
			NoTableScan:                      types.BoolValue(conversion.SafeValue(input.NoTableScan)),
			OplogMinRetentionHours:           types.Float64Value(conversion.SafeValue(input.OplogMinRetentionHours)),
			OplogSizeMb:                      types.Int64Value(conversion.SafeValue(conversion.IntPtrToInt64Ptr(input.OplogSizeMB))),
			SampleSizeBiconnector:            types.Int64Value(conversion.SafeValue(conversion.IntPtrToInt64Ptr(input.SampleSizeBIConnector))),
			SampleRefreshIntervalBiconnector: types.Int64Value(conversion.SafeValue(conversion.IntPtrToInt64Ptr(input.SampleRefreshIntervalBIConnector))),
			TransactionLifetimeLimitSeconds:  types.Int64Value(conversion.SafeValue(input.TransactionLifetimeLimitSeconds)),
			DefaultMaxTimeMS:                 types.Int64PointerValue(conversion.IntPtrToInt64Ptr(input.DefaultMaxTimeMS)),
			TlsCipherConfigMode:              types.StringValue(conversion.SafeValue(input.TlsCipherConfigMode)),
			CustomOpensslCipherConfigTls12:   customOpensslCipherConfigTLS12(ctx, input, diags),
		}
	} else {
		advancedConfig.CustomOpensslCipherConfigTls12 = types.SetNull(types.StringType) // required to handle move state
	}

	objType, diagsLocal := types.ObjectValueFrom(ctx, AdvancedConfigurationObjType.AttrTypes, advancedConfig)
	diags.Append(diagsLocal...)
	tfModel.AdvancedConfiguration = objType
}

func customOpensslCipherConfigTLS12(ctx context.Context, processArgs *admin.ClusterDescriptionProcessArgs20240805, diags *diag.Diagnostics) types.Set {
	if processArgs != nil && len(*processArgs.CustomOpensslCipherConfigTls12) == 0 {
		return types.SetNull(types.StringType)
	}

	customOpensslCipherConfigTLS12, d := types.SetValueFrom(ctx, types.StringType, processArgs.CustomOpensslCipherConfigTls12)
	diags.Append(d...)
	return customOpensslCipherConfigTLS12
}
