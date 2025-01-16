package streamconnection_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/mongodb/terraform-provider-mongodbatlas/internal/service/streamconnection"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/atlas-sdk/v20241113004/admin"
)

const (
	connectionName            = "Connection"
	typeValue                 = ""
	clusterName               = "Cluster0"
	dummyProjectID            = "111111111111111111111111"
	instanceName              = "InstanceName"
	authMechanism             = "PLAIN"
	authUsername              = "user1"
	securityProtocol          = "SSL"
	bootstrapServers          = "localhost:9092,another.host:9092"
	dbRole                    = "customRole"
	dbRoleType                = "CUSTOM"
	sampleConnectionName      = "sample_stream_solar"
	networkingType            = "PUBLIC"
	privatelinkNetworkingType = "PRIVATE_LINK"
)

var connectionID = "connectionID"

var configMap = map[string]string{
	"auto.offset.reset": "earliest",
}

type sdkToTFModelTestCase struct {
	SDKResp                  *admin.StreamsConnection
	providedProjID           string
	providedInstanceName     string
	providedAuthConfig       *types.Object
	providedNetworkingConfig *types.Object
	expectedTFModel          *streamconnection.TFStreamConnectionModel
	name                     string
}

func TestStreamConnectionSDKToTFModel(t *testing.T) {
	var authConfigWithPasswordDefined = tfAuthenticationObject(t, authMechanism, authUsername, "raw password")
	var privateLinkNetworkingConfig = tfNetworkingObject(t, privatelinkNetworkingType, &connectionID)

	testCases := []sdkToTFModelTestCase{
		{
			name: "Cluster connection type SDK response",
			SDKResp: &admin.StreamsConnection{
				Name:        admin.PtrString(connectionName),
				Type:        admin.PtrString("Cluster"),
				ClusterName: admin.PtrString(clusterName),
				DbRoleToExecute: &admin.DBRoleToExecute{
					Role: admin.PtrString(dbRole),
					Type: admin.PtrString(dbRoleType),
				},
			},
			providedProjID:           dummyProjectID,
			providedInstanceName:     instanceName,
			providedAuthConfig:       nil,
			providedNetworkingConfig: nil,
			expectedTFModel: &streamconnection.TFStreamConnectionModel{
				ProjectID:       types.StringValue(dummyProjectID),
				InstanceName:    types.StringValue(instanceName),
				ConnectionName:  types.StringValue(connectionName),
				Type:            types.StringValue("Cluster"),
				ClusterName:     types.StringValue(clusterName),
				Authentication:  types.ObjectNull(streamconnection.ConnectionAuthenticationObjectType.AttrTypes),
				Config:          types.MapNull(types.StringType),
				Security:        types.ObjectNull(streamconnection.ConnectionSecurityObjectType.AttrTypes),
				DBRoleToExecute: tfDBRoleToExecuteObject(t, dbRole, dbRoleType),
				Networking:      types.ObjectNull(streamconnection.NetworkingObjectType.AttrTypes),
			},
		},
		{
			name: "Kafka connection type SDK response",
			SDKResp: &admin.StreamsConnection{
				Name: admin.PtrString(connectionName),
				Type: admin.PtrString("Kafka"),
				Authentication: &admin.StreamsKafkaAuthentication{
					Mechanism: admin.PtrString(authMechanism),
					Username:  admin.PtrString(authUsername),
				},
				BootstrapServers: admin.PtrString(bootstrapServers),
				Config:           &configMap,
				Security: &admin.StreamsKafkaSecurity{
					Protocol:                admin.PtrString(securityProtocol),
					BrokerPublicCertificate: admin.PtrString(DummyCACert),
				},
			},
			providedProjID:           dummyProjectID,
			providedInstanceName:     instanceName,
			providedAuthConfig:       &authConfigWithPasswordDefined,
			providedNetworkingConfig: nil,
			expectedTFModel: &streamconnection.TFStreamConnectionModel{
				ProjectID:        types.StringValue(dummyProjectID),
				InstanceName:     types.StringValue(instanceName),
				ConnectionName:   types.StringValue(connectionName),
				Type:             types.StringValue("Kafka"),
				Authentication:   tfAuthenticationObject(t, authMechanism, authUsername, "raw password"), // password value is obtained from config, not api resp.
				BootstrapServers: types.StringValue(bootstrapServers),
				Config:           tfConfigMap(t, configMap),
				Security:         tfSecurityObject(t, DummyCACert, securityProtocol),
				DBRoleToExecute:  types.ObjectNull(streamconnection.DBRoleToExecuteObjectType.AttrTypes),
				Networking:       types.ObjectNull(streamconnection.NetworkingObjectType.AttrTypes),
			},
		},
		{
			name: "Kafka connection type SDK response with no optional values provided",
			SDKResp: &admin.StreamsConnection{
				Name: admin.PtrString(connectionName),
				Type: admin.PtrString("Kafka"),
			},
			providedProjID:           dummyProjectID,
			providedInstanceName:     instanceName,
			providedAuthConfig:       nil,
			providedNetworkingConfig: nil,
			expectedTFModel: &streamconnection.TFStreamConnectionModel{
				ProjectID:       types.StringValue(dummyProjectID),
				InstanceName:    types.StringValue(instanceName),
				ConnectionName:  types.StringValue(connectionName),
				Type:            types.StringValue("Kafka"),
				Authentication:  types.ObjectNull(streamconnection.ConnectionAuthenticationObjectType.AttrTypes),
				Config:          types.MapNull(types.StringType),
				Security:        types.ObjectNull(streamconnection.ConnectionSecurityObjectType.AttrTypes),
				DBRoleToExecute: types.ObjectNull(streamconnection.DBRoleToExecuteObjectType.AttrTypes),
				Networking:      types.ObjectNull(streamconnection.NetworkingObjectType.AttrTypes),
			},
		},
		{
			name: "Kafka connection type with config that does not have authentication value (case of imports and data sources)",
			SDKResp: &admin.StreamsConnection{
				Name: admin.PtrString(connectionName),
				Type: admin.PtrString("Kafka"),
				Authentication: &admin.StreamsKafkaAuthentication{
					Mechanism: admin.PtrString(authMechanism),
					Username:  admin.PtrString(authUsername),
				},
				BootstrapServers: admin.PtrString(bootstrapServers),
				Config:           &configMap,
				Security: &admin.StreamsKafkaSecurity{
					Protocol:                admin.PtrString(securityProtocol),
					BrokerPublicCertificate: admin.PtrString(DummyCACert),
				},
			},
			providedProjID:           dummyProjectID,
			providedInstanceName:     instanceName,
			providedAuthConfig:       nil,
			providedNetworkingConfig: nil,
			expectedTFModel: &streamconnection.TFStreamConnectionModel{
				ProjectID:        types.StringValue(dummyProjectID),
				InstanceName:     types.StringValue(instanceName),
				ConnectionName:   types.StringValue(connectionName),
				Type:             types.StringValue("Kafka"),
				Authentication:   tfAuthenticationObjectWithNoPassword(t, authMechanism, authUsername),
				BootstrapServers: types.StringValue(bootstrapServers),
				Config:           tfConfigMap(t, configMap),
				Security:         tfSecurityObject(t, DummyCACert, securityProtocol),
				DBRoleToExecute:  types.ObjectNull(streamconnection.DBRoleToExecuteObjectType.AttrTypes),
				Networking:       types.ObjectNull(streamconnection.NetworkingObjectType.AttrTypes),
			},
		},
		{
			name: "Kafka connection type SDK response with Private link networking config",
			SDKResp: &admin.StreamsConnection{
				Name: admin.PtrString(connectionName),
				Type: admin.PtrString("Kafka"),
				Authentication: &admin.StreamsKafkaAuthentication{
					Mechanism: admin.PtrString(authMechanism),
					Username:  admin.PtrString(authUsername),
				},
				BootstrapServers: admin.PtrString(bootstrapServers),
				Config:           &configMap,
				Security: &admin.StreamsKafkaSecurity{
					Protocol:                admin.PtrString(securityProtocol),
					BrokerPublicCertificate: admin.PtrString(DummyCACert),
				},
				Networking: &admin.StreamsKafkaNetworking{
					Access: &admin.StreamsKafkaNetworkingAccess{
						Type: admin.PtrString(privatelinkNetworkingType),
					},
				},
			},
			providedProjID:           dummyProjectID,
			providedInstanceName:     instanceName,
			providedAuthConfig:       nil,
			providedNetworkingConfig: &privateLinkNetworkingConfig,
			expectedTFModel: &streamconnection.TFStreamConnectionModel{
				ProjectID:        types.StringValue(dummyProjectID),
				InstanceName:     types.StringValue(instanceName),
				ConnectionName:   types.StringValue(connectionName),
				Type:             types.StringValue("Kafka"),
				Authentication:   tfAuthenticationObjectWithNoPassword(t, authMechanism, authUsername),
				BootstrapServers: types.StringValue(bootstrapServers),
				Config:           tfConfigMap(t, configMap),
				Security:         tfSecurityObject(t, DummyCACert, securityProtocol),
				DBRoleToExecute:  types.ObjectNull(streamconnection.DBRoleToExecuteObjectType.AttrTypes),
				Networking:       tfNetworkingObject(t, privatelinkNetworkingType, &connectionID),
			},
		},
		{
			name: "Sample connection type sample_stream_solar sample",
			SDKResp: &admin.StreamsConnection{
				Name: admin.PtrString(sampleConnectionName),
				Type: admin.PtrString("Sample"),
			},
			providedProjID:       dummyProjectID,
			providedInstanceName: instanceName,
			expectedTFModel: &streamconnection.TFStreamConnectionModel{
				ProjectID:       types.StringValue(dummyProjectID),
				InstanceName:    types.StringValue(instanceName),
				ConnectionName:  types.StringValue(sampleConnectionName),
				Type:            types.StringValue("Sample"),
				Authentication:  types.ObjectNull(streamconnection.ConnectionAuthenticationObjectType.AttrTypes),
				Config:          types.MapNull(types.StringType),
				Security:        types.ObjectNull(streamconnection.ConnectionSecurityObjectType.AttrTypes),
				DBRoleToExecute: types.ObjectNull(streamconnection.DBRoleToExecuteObjectType.AttrTypes),
				Networking:      types.ObjectNull(streamconnection.NetworkingObjectType.AttrTypes),
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			resultModel, diags := streamconnection.NewTFStreamConnection(context.Background(), tc.providedProjID, tc.providedInstanceName, tc.providedAuthConfig, tc.providedNetworkingConfig, tc.SDKResp)
			if diags.HasError() {
				t.Fatalf("unexpected errors found: %s", diags.Errors()[0].Summary())
			}
			tc.expectedTFModel.ID = resultModel.ID // id is auto-generated, have no way of defining within expected model
			if !assert.Equal(t, tc.expectedTFModel, resultModel) {
				t.Fatalf("created terraform model did not match expected output")
			}
		})
	}
}

type paginatedConnectionsSDKToTFModelTestCase struct {
	SDKResp         *admin.PaginatedApiStreamsConnection
	providedConfig  *streamconnection.TFStreamConnectionsDSModel
	expectedTFModel *streamconnection.TFStreamConnectionsDSModel
	name            string
}

func TestStreamConnectionsSDKToTFModel(t *testing.T) {
	testCases := []paginatedConnectionsSDKToTFModelTestCase{
		{
			name: "Complete SDK response with configured page options",
			SDKResp: &admin.PaginatedApiStreamsConnection{
				Results: &[]admin.StreamsConnection{
					{
						Name: admin.PtrString(connectionName),
						Type: admin.PtrString("Kafka"),
						Authentication: &admin.StreamsKafkaAuthentication{
							Mechanism: admin.PtrString(authMechanism),
							Username:  admin.PtrString(authUsername),
						},
						BootstrapServers: admin.PtrString(bootstrapServers),
						Config:           &configMap,
						Security: &admin.StreamsKafkaSecurity{
							Protocol:                admin.PtrString(securityProtocol),
							BrokerPublicCertificate: admin.PtrString(DummyCACert),
						},
						Networking: &admin.StreamsKafkaNetworking{
							Access: &admin.StreamsKafkaNetworkingAccess{
								Type: admin.PtrString(networkingType),
							},
						},
					},
					{
						Name:        admin.PtrString(connectionName),
						Type:        admin.PtrString("Cluster"),
						ClusterName: admin.PtrString(clusterName),
						DbRoleToExecute: &admin.DBRoleToExecute{
							Role: admin.PtrString(dbRole),
							Type: admin.PtrString(dbRoleType),
						},
					},
					{
						Name: admin.PtrString(sampleConnectionName),
						Type: admin.PtrString("Sample"),
					},
				},
				TotalCount: admin.PtrInt(3),
			},
			providedConfig: &streamconnection.TFStreamConnectionsDSModel{
				ProjectID:    types.StringValue(dummyProjectID),
				InstanceName: types.StringValue(instanceName),
				PageNum:      types.Int64Value(1),
				ItemsPerPage: types.Int64Value(3),
			},
			expectedTFModel: &streamconnection.TFStreamConnectionsDSModel{
				ProjectID:    types.StringValue(dummyProjectID),
				InstanceName: types.StringValue(instanceName),
				PageNum:      types.Int64Value(1),
				ItemsPerPage: types.Int64Value(3),
				TotalCount:   types.Int64Value(3),
				Results: []streamconnection.TFStreamConnectionModel{
					{
						ID:               types.StringValue(fmt.Sprintf("%s-%s-%s", instanceName, dummyProjectID, connectionName)),
						ProjectID:        types.StringValue(dummyProjectID),
						InstanceName:     types.StringValue(instanceName),
						ConnectionName:   types.StringValue(connectionName),
						Type:             types.StringValue("Kafka"),
						Authentication:   tfAuthenticationObjectWithNoPassword(t, authMechanism, authUsername),
						BootstrapServers: types.StringValue(bootstrapServers),
						Config:           tfConfigMap(t, configMap),
						Security:         tfSecurityObject(t, DummyCACert, securityProtocol),
						DBRoleToExecute:  types.ObjectNull(streamconnection.DBRoleToExecuteObjectType.AttrTypes),
						Networking:       tfNetworkingObject(t, networkingType, nil),
					},
					{
						ID:              types.StringValue(fmt.Sprintf("%s-%s-%s", instanceName, dummyProjectID, connectionName)),
						ProjectID:       types.StringValue(dummyProjectID),
						InstanceName:    types.StringValue(instanceName),
						ConnectionName:  types.StringValue(connectionName),
						Type:            types.StringValue("Cluster"),
						ClusterName:     types.StringValue(clusterName),
						Authentication:  types.ObjectNull(streamconnection.ConnectionAuthenticationObjectType.AttrTypes),
						Config:          types.MapNull(types.StringType),
						Security:        types.ObjectNull(streamconnection.ConnectionSecurityObjectType.AttrTypes),
						DBRoleToExecute: tfDBRoleToExecuteObject(t, dbRole, dbRoleType),
						Networking:      types.ObjectNull(streamconnection.NetworkingObjectType.AttrTypes),
					},
					{
						ID:              types.StringValue(fmt.Sprintf("%s-%s-%s", instanceName, dummyProjectID, sampleConnectionName)),
						ProjectID:       types.StringValue(dummyProjectID),
						InstanceName:    types.StringValue(instanceName),
						ConnectionName:  types.StringValue(sampleConnectionName),
						Type:            types.StringValue("Sample"),
						ClusterName:     types.StringNull(),
						Authentication:  types.ObjectNull(streamconnection.ConnectionAuthenticationObjectType.AttrTypes),
						Config:          types.MapNull(types.StringType),
						Security:        types.ObjectNull(streamconnection.ConnectionSecurityObjectType.AttrTypes),
						DBRoleToExecute: types.ObjectNull(streamconnection.DBRoleToExecuteObjectType.AttrTypes),
						Networking:      types.ObjectNull(streamconnection.NetworkingObjectType.AttrTypes),
					},
				},
			},
		},
		{
			name: "Without defining page options",
			SDKResp: &admin.PaginatedApiStreamsConnection{
				Results:    &[]admin.StreamsConnection{},
				TotalCount: admin.PtrInt(0),
			},
			providedConfig: &streamconnection.TFStreamConnectionsDSModel{
				ProjectID:    types.StringValue(dummyProjectID),
				InstanceName: types.StringValue(instanceName),
			},
			expectedTFModel: &streamconnection.TFStreamConnectionsDSModel{
				ProjectID:    types.StringValue(dummyProjectID),
				InstanceName: types.StringValue(instanceName),
				PageNum:      types.Int64Null(),
				ItemsPerPage: types.Int64Null(),
				TotalCount:   types.Int64Value(0),
				Results:      []streamconnection.TFStreamConnectionModel{},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			resultModel, diags := streamconnection.NewTFStreamConnections(context.Background(), tc.providedConfig, tc.SDKResp)
			if diags.HasError() {
				t.Fatalf("unexpected errors found: %s", diags.Errors()[0].Summary())
			}
			tc.expectedTFModel.ID = resultModel.ID // id is auto-generated, have no way of defining within expected model
			if !assert.Equal(t, tc.expectedTFModel, resultModel) {
				t.Fatalf("created terraform model did not match expected output")
			}
		})
	}
}

type tfToSDKCreateModelTestCase struct {
	tfModel        *streamconnection.TFStreamConnectionModel
	expectedSDKReq *admin.StreamsConnection
	name           string
}

func TestStreamInstanceTFToSDKCreateModel(t *testing.T) {
	testCases := []tfToSDKCreateModelTestCase{
		{
			name: "Cluster type complete TF state",
			tfModel: &streamconnection.TFStreamConnectionModel{
				ProjectID:       types.StringValue(dummyProjectID),
				InstanceName:    types.StringValue(instanceName),
				ConnectionName:  types.StringValue(connectionName),
				Type:            types.StringValue("Cluster"),
				ClusterName:     types.StringValue(clusterName),
				DBRoleToExecute: tfDBRoleToExecuteObject(t, dbRole, dbRoleType),
			},
			expectedSDKReq: &admin.StreamsConnection{
				Name:        admin.PtrString(connectionName),
				Type:        admin.PtrString("Cluster"),
				ClusterName: admin.PtrString(clusterName),
				DbRoleToExecute: &admin.DBRoleToExecute{
					Role: admin.PtrString(dbRole),
					Type: admin.PtrString(dbRoleType),
				},
			},
		},
		{
			name: "Kafka type complete TF state",
			tfModel: &streamconnection.TFStreamConnectionModel{
				ProjectID:        types.StringValue(dummyProjectID),
				InstanceName:     types.StringValue(instanceName),
				ConnectionName:   types.StringValue(connectionName),
				Type:             types.StringValue("Kafka"),
				Authentication:   tfAuthenticationObject(t, authMechanism, authUsername, "raw password"),
				BootstrapServers: types.StringValue(bootstrapServers),
				Config:           tfConfigMap(t, configMap),
				Security:         tfSecurityObject(t, DummyCACert, securityProtocol),
			},
			expectedSDKReq: &admin.StreamsConnection{
				Name: admin.PtrString(connectionName),
				Type: admin.PtrString("Kafka"),
				Authentication: &admin.StreamsKafkaAuthentication{
					Mechanism: admin.PtrString(authMechanism),
					Username:  admin.PtrString(authUsername),
					Password:  admin.PtrString("raw password"),
				},
				BootstrapServers: admin.PtrString(bootstrapServers),
				Config:           &configMap,
				Security: &admin.StreamsKafkaSecurity{
					Protocol:                admin.PtrString(securityProtocol),
					BrokerPublicCertificate: admin.PtrString(DummyCACert),
				},
			},
		},
		{
			name: "Kafka type TF state with no optional attributes",
			tfModel: &streamconnection.TFStreamConnectionModel{
				ProjectID:      types.StringValue(dummyProjectID),
				InstanceName:   types.StringValue(instanceName),
				ConnectionName: types.StringValue(connectionName),
				Type:           types.StringValue("Kafka"),
			},
			expectedSDKReq: &admin.StreamsConnection{
				Name: admin.PtrString(connectionName),
				Type: admin.PtrString("Kafka"),
			},
		},
		{
			name: "Sample type TF state",
			tfModel: &streamconnection.TFStreamConnectionModel{
				ProjectID:      types.StringValue(dummyProjectID),
				InstanceName:   types.StringValue(instanceName),
				ConnectionName: types.StringValue(connectionName),
				Type:           types.StringValue("Sample"),
			},
			expectedSDKReq: &admin.StreamsConnection{
				Name: admin.PtrString(connectionName),
				Type: admin.PtrString("Sample"),
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			apiReqResult, diags := streamconnection.NewStreamConnectionReq(context.Background(), tc.tfModel)
			if diags.HasError() {
				t.Errorf("unexpected errors found: %s", diags.Errors()[0].Summary())
			}
			if !assert.Equal(t, tc.expectedSDKReq, apiReqResult) {
				t.Errorf("created sdk model did not match expected output")
			}
		})
	}
}

func tfAuthenticationObject(t *testing.T, mechanism, username, password string) types.Object {
	t.Helper()
	auth, diags := types.ObjectValueFrom(context.Background(), streamconnection.ConnectionAuthenticationObjectType.AttrTypes, streamconnection.TFConnectionAuthenticationModel{
		Mechanism: types.StringValue(mechanism),
		Username:  types.StringValue(username),
		Password:  types.StringValue(password),
	})
	if diags.HasError() {
		t.Errorf("failed to create terraform data model: %s", diags.Errors()[0].Summary())
	}
	return auth
}

func tfAuthenticationObjectWithNoPassword(t *testing.T, mechanism, username string) types.Object {
	t.Helper()
	auth, diags := types.ObjectValueFrom(context.Background(), streamconnection.ConnectionAuthenticationObjectType.AttrTypes, streamconnection.TFConnectionAuthenticationModel{
		Mechanism: types.StringValue(mechanism),
		Username:  types.StringValue(username),
	})
	if diags.HasError() {
		t.Errorf("failed to create terraform data model: %s", diags.Errors()[0].Summary())
	}
	return auth
}

func tfSecurityObject(t *testing.T, brokerPublicCertificate, protocol string) types.Object {
	t.Helper()
	auth, diags := types.ObjectValueFrom(context.Background(), streamconnection.ConnectionSecurityObjectType.AttrTypes, streamconnection.TFConnectionSecurityModel{
		BrokerPublicCertificate: types.StringValue(brokerPublicCertificate),
		Protocol:                types.StringValue(protocol),
	})
	if diags.HasError() {
		t.Errorf("failed to create terraform data model: %s", diags.Errors()[0].Summary())
	}
	return auth
}

func tfConfigMap(t *testing.T, config map[string]string) types.Map {
	t.Helper()
	mapValue, diags := types.MapValueFrom(context.Background(), types.StringType, config)
	if diags.HasError() {
		t.Errorf("failed to create terraform data model: %s", diags.Errors()[0].Summary())
	}
	return mapValue
}

func tfDBRoleToExecuteObject(t *testing.T, role, roleType string) types.Object {
	t.Helper()
	auth, diags := types.ObjectValueFrom(context.Background(), streamconnection.DBRoleToExecuteObjectType.AttrTypes, streamconnection.TFDbRoleToExecuteModel{
		Role: types.StringValue(role),
		Type: types.StringValue(roleType),
	})
	if diags.HasError() {
		t.Errorf("failed to create terraform data model: %s", diags.Errors()[0].Summary())
	}
	return auth
}

func tfNetworkingObject(t *testing.T, networkingType string, connectionID *string) types.Object {
	t.Helper()
	networkingAccessModel, diags := types.ObjectValueFrom(context.Background(), streamconnection.NetworkingAccessObjectType.AttrTypes, streamconnection.TFNetworkingAccessModel{
		Type:         types.StringValue(networkingType),
		ConnectionID: types.StringPointerValue(connectionID),
	})
	if diags.HasError() {
		t.Errorf("failed to create terraform data model: %s", diags.Errors()[0].Summary())
	}
	networking, diags := types.ObjectValueFrom(context.Background(), streamconnection.NetworkingObjectType.AttrTypes, streamconnection.TFNetworkingModel{
		Access: networkingAccessModel,
	})
	if diags.HasError() {
		t.Errorf("failed to create terraform data model: %s", diags.Errors()[0].Summary())
	}
	return networking
}
