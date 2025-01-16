package streamconnection

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/mongodb/terraform-provider-mongodbatlas/internal/common/conversion"
	"go.mongodb.org/atlas-sdk/v20241113004/admin"
)

func NewStreamConnectionReq(ctx context.Context, plan *TFStreamConnectionModel) (*admin.StreamsConnection, diag.Diagnostics) {
	streamConnection := admin.StreamsConnection{
		Name:             plan.ConnectionName.ValueStringPointer(),
		Type:             plan.Type.ValueStringPointer(),
		ClusterName:      plan.ClusterName.ValueStringPointer(),
		BootstrapServers: plan.BootstrapServers.ValueStringPointer(),
	}
	if !plan.Authentication.IsNull() {
		authenticationModel := &TFConnectionAuthenticationModel{}
		if diags := plan.Authentication.As(ctx, authenticationModel, basetypes.ObjectAsOptions{}); diags.HasError() {
			return nil, diags
		}
		streamConnection.Authentication = &admin.StreamsKafkaAuthentication{
			Mechanism: authenticationModel.Mechanism.ValueStringPointer(),
			Password:  authenticationModel.Password.ValueStringPointer(),
			Username:  authenticationModel.Username.ValueStringPointer(),
		}
	}
	if !plan.Security.IsNull() {
		securityModel := &TFConnectionSecurityModel{}
		if diags := plan.Security.As(ctx, securityModel, basetypes.ObjectAsOptions{}); diags.HasError() {
			return nil, diags
		}
		streamConnection.Security = &admin.StreamsKafkaSecurity{
			BrokerPublicCertificate: securityModel.BrokerPublicCertificate.ValueStringPointer(),
			Protocol:                securityModel.Protocol.ValueStringPointer(),
		}
	}

	if !plan.Config.IsNull() {
		configMap := &map[string]string{}
		if diags := plan.Config.ElementsAs(ctx, configMap, true); diags.HasError() {
			return nil, diags
		}
		streamConnection.Config = configMap
	}

	if !plan.DBRoleToExecute.IsNull() {
		dbRoleToExecuteModel := &TFDbRoleToExecuteModel{}
		if diags := plan.DBRoleToExecute.As(ctx, dbRoleToExecuteModel, basetypes.ObjectAsOptions{}); diags.HasError() {
			return nil, diags
		}
		streamConnection.DbRoleToExecute = &admin.DBRoleToExecute{
			Role: dbRoleToExecuteModel.Role.ValueStringPointer(),
			Type: dbRoleToExecuteModel.Type.ValueStringPointer(),
		}
	}

	if !plan.Networking.IsNull() && !plan.Networking.IsUnknown() {
		networkingModel := &TFNetworkingModel{}
		if diags := plan.Networking.As(ctx, networkingModel, basetypes.ObjectAsOptions{}); diags.HasError() {
			return nil, diags
		}
		networkingAccessModel := &TFNetworkingAccessModel{}
		if diags := networkingModel.Access.As(ctx, networkingAccessModel, basetypes.ObjectAsOptions{}); diags.HasError() {
			return nil, diags
		}
		streamConnection.Networking = &admin.StreamsKafkaNetworking{
			Access: &admin.StreamsKafkaNetworkingAccess{
				Type:         networkingAccessModel.Type.ValueStringPointer(),
				ConnectionId: networkingAccessModel.ConnectionID.ValueStringPointer(),
			},
		}
	}

	return &streamConnection, nil
}

func NewTFStreamConnection(ctx context.Context, projID, instanceName string, currAuthConfig, currNetworkingConfig *types.Object, apiResp *admin.StreamsConnection) (*TFStreamConnectionModel, diag.Diagnostics) {
	rID := fmt.Sprintf("%s-%s-%s", instanceName, projID, conversion.SafeString(apiResp.Name))
	connectionModel := TFStreamConnectionModel{
		ID:               types.StringValue(rID),
		ProjectID:        types.StringValue(projID),
		InstanceName:     types.StringValue(instanceName),
		ConnectionName:   types.StringPointerValue(apiResp.Name),
		Type:             types.StringPointerValue(apiResp.Type),
		ClusterName:      types.StringPointerValue(apiResp.ClusterName),
		BootstrapServers: types.StringPointerValue(apiResp.BootstrapServers),
	}

	authModel, diags := newTFConnectionAuthenticationModel(ctx, currAuthConfig, apiResp.Authentication)
	if diags.HasError() {
		return nil, diags
	}
	connectionModel.Authentication = *authModel

	connectionModel.Config = types.MapNull(types.StringType)
	if apiResp.Config != nil {
		mapValue, diags := types.MapValueFrom(ctx, types.StringType, apiResp.Config)
		if diags.HasError() {
			return nil, diags
		}
		connectionModel.Config = mapValue
	}

	connectionModel.Security = types.ObjectNull(ConnectionSecurityObjectType.AttrTypes)
	if apiResp.Security != nil {
		securityModel, diags := types.ObjectValueFrom(ctx, ConnectionSecurityObjectType.AttrTypes, TFConnectionSecurityModel{
			BrokerPublicCertificate: types.StringPointerValue(apiResp.Security.BrokerPublicCertificate),
			Protocol:                types.StringPointerValue(apiResp.Security.Protocol),
		})
		if diags.HasError() {
			return nil, diags
		}
		connectionModel.Security = securityModel
	}

	connectionModel.DBRoleToExecute = types.ObjectNull(DBRoleToExecuteObjectType.AttrTypes)
	if apiResp.DbRoleToExecute != nil {
		dbRoleToExecuteModel, diags := types.ObjectValueFrom(ctx, DBRoleToExecuteObjectType.AttrTypes, TFDbRoleToExecuteModel{
			Role: types.StringPointerValue(apiResp.DbRoleToExecute.Role),
			Type: types.StringPointerValue(apiResp.DbRoleToExecute.Type),
		})
		if diags.HasError() {
			return nil, diags
		}
		connectionModel.DBRoleToExecute = dbRoleToExecuteModel
	}

	connectionModel.Networking = types.ObjectNull(NetworkingObjectType.AttrTypes)
	if apiResp.Networking != nil {
		connectionID := types.StringNull()
		if currNetworkingConfig != nil && !currNetworkingConfig.IsNull() && !currNetworkingConfig.IsUnknown() { // if config is available (create & update of resource) connectionID value from the config is set in new state
			configNetworkingModel := &TFNetworkingModel{}
			if diags := currNetworkingConfig.As(ctx, configNetworkingModel, basetypes.ObjectAsOptions{}); diags.HasError() {
				return nil, diags
			}
			configNetworkingAccessModel := &TFNetworkingAccessModel{}
			if diags := configNetworkingModel.Access.As(ctx, configNetworkingAccessModel, basetypes.ObjectAsOptions{}); diags.HasError() {
				return nil, diags
			}
			connectionID = configNetworkingAccessModel.ConnectionID
		}
		networkingAccessModel, diags := types.ObjectValueFrom(ctx, NetworkingAccessObjectType.AttrTypes, TFNetworkingAccessModel{
			Type:         types.StringPointerValue(apiResp.Networking.Access.Type),
			ConnectionID: connectionID,
		})
		if diags.HasError() {
			return nil, diags
		}
		networkingModel, diags := types.ObjectValueFrom(ctx, NetworkingObjectType.AttrTypes, TFNetworkingModel{
			Access: networkingAccessModel,
		})
		if diags.HasError() {
			return nil, diags
		}
		connectionModel.Networking = networkingModel
	}

	return &connectionModel, nil
}

func newTFConnectionAuthenticationModel(ctx context.Context, currAuthConfig *types.Object, authResp *admin.StreamsKafkaAuthentication) (*types.Object, diag.Diagnostics) {
	if authResp != nil {
		resultAuthModel := TFConnectionAuthenticationModel{
			Mechanism: types.StringPointerValue(authResp.Mechanism),
			Username:  types.StringPointerValue(authResp.Username),
		}

		if currAuthConfig != nil && !currAuthConfig.IsNull() { // if config is available (create & update of resource) password value is set in new state
			configAuthModel := &TFConnectionAuthenticationModel{}
			if diags := currAuthConfig.As(ctx, configAuthModel, basetypes.ObjectAsOptions{}); diags.HasError() {
				return nil, diags
			}
			resultAuthModel.Password = configAuthModel.Password
		}

		resultObject, diags := types.ObjectValueFrom(ctx, ConnectionAuthenticationObjectType.AttrTypes, resultAuthModel)
		if diags.HasError() {
			return nil, diags
		}
		return &resultObject, nil
	}
	nullValue := types.ObjectNull(ConnectionAuthenticationObjectType.AttrTypes)
	return &nullValue, nil
}

func NewTFStreamConnections(ctx context.Context,
	streamConnectionsConfig *TFStreamConnectionsDSModel,
	paginatedResult *admin.PaginatedApiStreamsConnection) (*TFStreamConnectionsDSModel, diag.Diagnostics) {
	input := paginatedResult.GetResults()
	results := make([]TFStreamConnectionModel, len(input))
	for i := range input {
		projectID := streamConnectionsConfig.ProjectID.ValueString()
		instanceName := streamConnectionsConfig.InstanceName.ValueString()
		connectionModel, diags := NewTFStreamConnection(ctx, projectID, instanceName, nil, nil, &input[i])
		if diags.HasError() {
			return nil, diags
		}
		results[i] = *connectionModel
	}
	return &TFStreamConnectionsDSModel{
		ID:           types.StringValue(id.UniqueId()),
		ProjectID:    streamConnectionsConfig.ProjectID,
		InstanceName: streamConnectionsConfig.InstanceName,
		Results:      results,
		PageNum:      streamConnectionsConfig.PageNum,
		ItemsPerPage: streamConnectionsConfig.ItemsPerPage,
		TotalCount:   types.Int64PointerValue(conversion.IntPtrToInt64Ptr(paginatedResult.TotalCount)),
	}, nil
}
