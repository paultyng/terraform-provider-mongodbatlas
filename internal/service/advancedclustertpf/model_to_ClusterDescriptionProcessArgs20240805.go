package advancedclustertpf

import (
	"context"

	"go.mongodb.org/atlas-sdk/v20241113004/admin"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"

	"github.com/mongodb/terraform-provider-mongodbatlas/internal/common/conversion"
)

func NewAtlasReqAdvancedConfiguration(ctx context.Context, objInput *types.Object, diags *diag.Diagnostics) *admin.ClusterDescriptionProcessArgs20240805 {
	var resp *admin.ClusterDescriptionProcessArgs20240805
	if objInput == nil || objInput.IsUnknown() || objInput.IsNull() {
		return resp
	}
	input := &TFAdvancedConfigurationModel{}
	if localDiags := objInput.As(ctx, input, basetypes.ObjectAsOptions{}); len(localDiags) > 0 {
		diags.Append(localDiags...)
		return resp
	}
	return &admin.ClusterDescriptionProcessArgs20240805{
		ChangeStreamOptionsPreAndPostImagesExpireAfterSeconds: conversion.NilForUnknown(input.ChangeStreamOptionsPreAndPostImagesExpireAfterSeconds, conversion.Int64PtrToIntPtr(input.ChangeStreamOptionsPreAndPostImagesExpireAfterSeconds.ValueInt64Pointer())),
		DefaultWriteConcern:              conversion.NilForUnknown(input.DefaultWriteConcern, input.DefaultWriteConcern.ValueStringPointer()),
		JavascriptEnabled:                conversion.NilForUnknown(input.JavascriptEnabled, input.JavascriptEnabled.ValueBoolPointer()),
		MinimumEnabledTlsProtocol:        conversion.NilForUnknown(input.MinimumEnabledTlsProtocol, input.MinimumEnabledTlsProtocol.ValueStringPointer()),
		NoTableScan:                      conversion.NilForUnknown(input.NoTableScan, input.NoTableScan.ValueBoolPointer()),
		OplogMinRetentionHours:           conversion.NilForUnknown(input.OplogMinRetentionHours, input.OplogMinRetentionHours.ValueFloat64Pointer()),
		OplogSizeMB:                      conversion.NilForUnknown(input.OplogSizeMb, conversion.Int64PtrToIntPtr(input.OplogSizeMb.ValueInt64Pointer())),
		SampleRefreshIntervalBIConnector: conversion.NilForUnknown(input.SampleRefreshIntervalBiconnector, conversion.Int64PtrToIntPtr(input.SampleRefreshIntervalBiconnector.ValueInt64Pointer())),
		SampleSizeBIConnector:            conversion.NilForUnknown(input.SampleSizeBiconnector, conversion.Int64PtrToIntPtr(input.SampleSizeBiconnector.ValueInt64Pointer())),
		TransactionLifetimeLimitSeconds:  conversion.NilForUnknown(input.TransactionLifetimeLimitSeconds, input.TransactionLifetimeLimitSeconds.ValueInt64Pointer()),
		DefaultMaxTimeMS:                 conversion.NilForUnknown(input.DefaultMaxTimeMS, conversion.Int64PtrToIntPtr(input.DefaultMaxTimeMS.ValueInt64Pointer())),
		TlsCipherConfigMode:              conversion.NilForUnknown(input.TlsCipherConfigMode, input.TlsCipherConfigMode.ValueStringPointer()),
		CustomOpensslCipherConfigTls12:   conversion.Pointer(conversion.TypesSetToString(ctx, input.CustomOpensslCipherConfigTls12)),
	}
}
