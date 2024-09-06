package employeeaccessgrant

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"go.mongodb.org/atlas-sdk/v20240805003/admin"
)

// TODO: `ctx` parameter and `diags` return value can be removed if tf schema has no complex data types (e.g., schema.ListAttribute, schema.SetAttribute)
func NewTFEmployeeAccessGrant(ctx context.Context, apiResp *admin.EmployeeAccessGrant) (*TFEmployeeAccessGrantModel, diag.Diagnostics) {
	// complexAttr, diagnostics := types.ListValueFrom(ctx, InnerObjectType, newTFComplexAttrModel(apiResp.ComplexAttr))
	// if diagnostics.HasError() {
	// 	return nil, diagnostics
	// }
	return &TFEmployeeAccessGrantModel{}, nil
}

// TODO: If SDK defined different models for create and update separate functions will need to be defined.
// TODO: `ctx` parameter and `diags` in return value can be removed if tf schema has no complex data types (e.g., schema.ListAttribute, schema.SetAttribute)
func NewEmployeeAccessGrantReq(ctx context.Context, plan *TFEmployeeAccessGrantModel) (*admin.EmployeeAccessGrant, diag.Diagnostics) {
	// var tfList []complexArgumentData
	// resp.Diagnostics.Append(plan.ComplexArgument.ElementsAs(ctx, &tfList, false)...)
	// if resp.Diagnostics.HasError() {
	// 	return nil, diagnostics
	// }
	return &admin.EmployeeAccessGrant{}, nil
}
