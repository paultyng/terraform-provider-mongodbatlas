//nolint:gocritic
package encryptionatrestprivateendpoint

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"

	"github.com/mongodb/terraform-provider-mongodbatlas/internal/config"
)

var _ datasource.DataSource = &encryptionAtRestPrivateEndpointsDS{}
var _ datasource.DataSourceWithConfigure = &encryptionAtRestPrivateEndpointsDS{}

func PluralDataSource() datasource.DataSource {
	return &encryptionAtRestPrivateEndpointsDS{
		DSCommon: config.DSCommon{
			DataSourceName: fmt.Sprintf("%ss", encryptionAtRestPrivateEndpointName),
		},
	}
}

type encryptionAtRestPrivateEndpointsDS struct {
	config.DSCommon
}

func (d *encryptionAtRestPrivateEndpointsDS) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = PluralDataSourceSchema(ctx)
}

func (d *encryptionAtRestPrivateEndpointsDS) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var encryptionAtRestPrivateEndpointsConfig TFEncryptionAtRestPrivateEndpointsDSModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &encryptionAtRestPrivateEndpointsConfig)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// TODO: make get request to obtain list of results

	// connV2 := d.Client.AtlasV2
	// TODO: GetEncryptionAtRestPrivateEndpointsForCloudProvider returns paginated results
	// v, resp, err := connV2.EncryptionAtRestUsingCustomerKeyManagementApi.GetEncryptionAtRestPrivateEndpointsForCloudProvider(ctx, "", "").Execute()
	// if err != nil {
	//	resp.Diagnostics.AddError("error fetching results", err.Error())
	//	return
	//}

	// TODO: process response into new terraform state
	//  newEncryptionAtRestPrivateEndpointsModel := NewTFEarPrivateEndpoints("","",conversion.IntPtr(99), v.GetResults())
	// if diags.HasError() {
	// 	resp.Diagnostics.Append(diags...)
	// 	return
	// }
	// resp.Diagnostics.Append(resp.State.Set(ctx, NewTFEarPrivateEndpoints(ctx))...)
}
