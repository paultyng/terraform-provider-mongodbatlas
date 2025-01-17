package streamprivatelinkendpoint

import (
	"context"
	"errors"
	"net/http"
	"regexp"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/mongodb/terraform-provider-mongodbatlas/internal/common/conversion"
	"github.com/mongodb/terraform-provider-mongodbatlas/internal/common/retrystrategy"
	"github.com/mongodb/terraform-provider-mongodbatlas/internal/config"
)

const (
	resourceName                     = "stream_privatelink_endpoint"
	warnUnsupportedOperation         = "Operation not supported"
	FailedStatusErrorMessageSummary  = "Private endpoint is in a failed status"
	NonEmptyErrorMessageFieldSummary = "Something went wrong. Please review the `status` field of this resource"
)

var _ resource.ResourceWithConfigure = &rs{}
var _ resource.ResourceWithImportState = &rs{}

func Resource() resource.Resource {
	return &rs{
		RSCommon: config.RSCommon{
			ResourceName: resourceName,
		},
	}
}

type rs struct {
	config.RSCommon
}

func (r *rs) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = ResourceSchema(ctx)
	conversion.UpdateSchemaDescription(&resp.Schema)
}

func (r *rs) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan TFModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	streamPrivatelinkEndpointReq, diags := NewAtlasReq(ctx, &plan)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	projectID := plan.ProjectId.ValueString()

	connV2 := r.Client.AtlasV2
	streamsPrivateLinkConnection, _, err := connV2.StreamsApi.CreatePrivateLinkConnection(ctx, projectID, streamPrivatelinkEndpointReq).Execute()
	if err != nil {
		resp.Diagnostics.AddError("error creating resource", err.Error())
		return
	}

	finalResp, err := waitStateTransition(ctx, projectID, *streamsPrivateLinkConnection.Id, connV2.StreamsApi)
	if err != nil {
		if finalResp != nil { // delete the resource that has been created but fails to reach desired state
			if _, _, err := connV2.StreamsApi.DeletePrivateLinkConnection(ctx, projectID, finalResp.GetId()).Execute(); err != nil {
				resp.Diagnostics.AddError("error deleting resource after failed creation", err.Error())
				return
			}
			_, err := WaitDeleteStateTransition(ctx, projectID, *finalResp.Id, connV2.StreamsApi)
			if err != nil {
				resp.Diagnostics.AddError("error waiting for state transition in deletion after a failed creation", err.Error())
				return
			}
		}
		resp.Diagnostics.AddError("error when waiting for status transition in creation", err.Error())
		return
	}

	newStreamPrivatelinkEndpointModel, diags := NewTFModel(ctx, projectID, finalResp)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	if plan.DnsSubDomain.IsNull() && len(newStreamPrivatelinkEndpointModel.DnsSubDomain.Elements()) == 0 {
		newStreamPrivatelinkEndpointModel.DnsSubDomain = types.ListNull(types.StringType)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, newStreamPrivatelinkEndpointModel)...)
}

func (r *rs) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state TFModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := state.ProjectId.ValueString()
	connectionID := state.Id.ValueString()

	connV2 := r.Client.AtlasV2
	streamsPrivateLinkConnection, apiResp, err := connV2.StreamsApi.GetPrivateLinkConnection(ctx, projectID, connectionID).Execute()
	if err != nil {
		if apiResp != nil && apiResp.StatusCode == http.StatusNotFound {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("error fetching resource", err.Error())
		return
	}

	newStreamPrivatelinkEndpointModel, diags := NewTFModel(ctx, projectID, streamsPrivateLinkConnection)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	if state.DnsSubDomain.IsNull() && len(newStreamPrivatelinkEndpointModel.DnsSubDomain.Elements()) == 0 {
		newStreamPrivatelinkEndpointModel.DnsSubDomain = types.ListNull(types.StringType)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, newStreamPrivatelinkEndpointModel)...)
}

func (r *rs) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError(warnUnsupportedOperation, "Updating the private endpoint for streams is not supported. To modify your infrastructure, please delete the existing mongodbatlas_stream_privatelink_endpoint resource and create a new one with the necessary updates")
}

func (r *rs) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state *TFModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := state.ProjectId.ValueString()
	connectionID := state.Id.ValueString()

	connV2 := r.Client.AtlasV2
	if _, _, err := connV2.StreamsApi.DeletePrivateLinkConnection(ctx, projectID, connectionID).Execute(); err != nil {
		resp.Diagnostics.AddError("error deleting resource", err.Error())
		return
	}

	model, err := WaitDeleteStateTransition(ctx, projectID, connectionID, connV2.StreamsApi)
	if err != nil {
		resp.Diagnostics.AddError("error waiting for state transition", err.Error())
		return
	}

	if model.GetState() == retrystrategy.RetryStrategyFailedState {
		resp.Diagnostics.Append(diag.NewErrorDiagnostic(FailedStatusErrorMessageSummary, ""))
	}
}

func (r *rs) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	projectID, connectionID, err := splitStreamPrivatelinkEndpointImportID(req.ID)
	if err != nil {
		resp.Diagnostics.AddError("error splitting import ID", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("project_id"), projectID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), connectionID)...)
}

func splitStreamPrivatelinkEndpointImportID(id string) (projectID, connectionID string, err error) {
	re := regexp.MustCompile(`(?s)^([0-9a-fA-F]{24})-([0-9a-fA-F]{24})$`)
	parts := re.FindStringSubmatch(id)

	if len(parts) != 3 {
		err = errors.New("import format error: to import a stream private link endpoint, use the format {project_id}-{connection_id}")
		return
	}

	projectID = parts[1]
	connectionID = parts[2]
	return
}
