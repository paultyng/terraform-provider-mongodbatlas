package project_test

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/mongodb/terraform-provider-mongodbatlas/internal/common/conversion"
	"github.com/mongodb/terraform-provider-mongodbatlas/internal/service/project"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/atlas-sdk/v20231115003/admin"
)

const (
	limitName           = "limitName"
	limitValue          = int64(64)
	limitCurrentUsage   = int64(64)
	limitDefaultLimit   = int64(32)
	limitMaximumLimit   = int64(16)
	projectID           = "projectId"
	projectName         = "projectName"
	projectOrgID        = "orgId"
	projectClusterCount = int64(1)
	clusterCount        = 1
)

var (
	roles              = []string{"GROUP_DATA_ACCESS_READ_ONLY", "GROUP_CLUSTER_MANAGER"}
	roleList, _        = types.ListValueFrom(context.Background(), types.StringType, roles)
	ipAddresses        = []string{"13.13.13.13"}
	ipAddressesList, _ = types.ListValueFrom(context.Background(), types.StringType, ipAddresses)
	teamRolesSDK       = []admin.TeamRole{
		{
			TeamId:    conversion.StringPtr("teamId"),
			RoleNames: conversion.NonEmptyToPtr(roles),
		},
	}
	teamsDSTF = []*project.TfTeamDSModel{
		{
			TeamID:    types.StringValue("teamId"),
			RoleNames: roleList,
		},
	}
	limitsSDK = []admin.DataFederationLimit{
		{
			Name:         limitName,
			Value:        limitValue,
			CurrentUsage: admin.PtrInt64(limitCurrentUsage),
			DefaultLimit: admin.PtrInt64(limitDefaultLimit),
			MaximumLimit: admin.PtrInt64(limitMaximumLimit),
		},
	}
	limitsTF = []*project.TfLimitModel{
		{
			Name:         types.StringValue(limitName),
			Value:        types.Int64Value(limitValue),
			CurrentUsage: types.Int64Value(limitCurrentUsage),
			DefaultLimit: types.Int64Value(limitDefaultLimit),
			MaximumLimit: types.Int64Value(limitMaximumLimit),
		},
	}
	ipAddressesTF = project.TFIPAddressesModel{
		Services: project.TFServicesModel{
			Clusters: []project.TFClusterIPsModel{
				{
					Inbound:     ipAddressesList,
					Outbound:    ipAddressesList,
					ClusterName: types.StringValue("Cluster0"),
				},
			},
		},
	}
	projectSDK = admin.Group{
		Id:           admin.PtrString(projectID),
		Name:         projectName,
		OrgId:        projectOrgID,
		ClusterCount: int64(clusterCount),
	}
	projectSettingsSDK = admin.GroupSettings{
		IsCollectDatabaseSpecificsStatisticsEnabled: admin.PtrBool(true),
		IsDataExplorerEnabled:                       admin.PtrBool(true),
		IsExtendedStorageSizesEnabled:               admin.PtrBool(true),
		IsPerformanceAdvisorEnabled:                 admin.PtrBool(true),
		IsRealtimePerformancePanelEnabled:           admin.PtrBool(true),
		IsSchemaAdvisorEnabled:                      admin.PtrBool(true),
	}
	projectIPAddressesSDK = admin.GroupIPAddresses{
		GroupId: admin.PtrString(projectID),
		Services: &admin.GroupService{
			Clusters: &[]admin.ClusterIPAddresses{
				{
					Inbound:     &[]string{"13.13.13.13"},
					Outbound:    &[]string{"13.13.13.13"},
					ClusterName: admin.PtrString("Cluster0"),
				},
			},
		},
	}
)

func TestTeamsDataSourceSDKToTFModel(t *testing.T) {
	testCases := []struct {
		name              string
		paginatedTeamRole *admin.PaginatedTeamRole
		expectedTFModel   []*project.TfTeamDSModel
	}{
		{
			name: "TeamRole",
			paginatedTeamRole: &admin.PaginatedTeamRole{
				TotalCount: conversion.IntPtr(0),
			}, // not setting explicitly expected result because we expect it to be nil
		},
		{
			name: "Complete TeamRole",
			paginatedTeamRole: &admin.PaginatedTeamRole{
				Results:    conversion.NonEmptyToPtr(teamRolesSDK),
				TotalCount: conversion.IntPtr(1),
			},
			expectedTFModel: teamsDSTF,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			resultModel := project.NewTFTeamsDataSourceModel(context.Background(), tc.paginatedTeamRole)
			if !assert.Equal(t, tc.expectedTFModel, resultModel) {
				t.Errorf("created terraform model did not match expected output")
			}
		})
	}
}

func TestLimitsDataSourceSDKToTFModel(t *testing.T) {
	testCases := []struct {
		name                 string
		dataFederationLimits []admin.DataFederationLimit
		expectedTFModel      []*project.TfLimitModel
	}{
		{
			name:                 "Limit",
			dataFederationLimits: limitsSDK,
			expectedTFModel:      limitsTF,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			resultModel := project.NewTFLimitsDataSourceModel(context.Background(), tc.dataFederationLimits)
			if !assert.Equal(t, tc.expectedTFModel, resultModel) {
				t.Errorf("created terraform model did not match expected output")
			}
		})
	}
}

func TestProjectDataSourceSDKToTFModel(t *testing.T) {
	testCases := []struct {
		name                 string
		project              *admin.Group
		teams                *admin.PaginatedTeamRole
		projectSettings      *admin.GroupSettings
		ipAddresses          *admin.GroupIPAddresses
		dataFederationLimits []admin.DataFederationLimit
		expectedTFModel      project.TfProjectDSModel
	}{
		{
			name:    "Project",
			project: &projectSDK,
			teams: &admin.PaginatedTeamRole{
				Results:    conversion.NonEmptyToPtr(teamRolesSDK),
				TotalCount: conversion.IntPtr(1),
			},
			projectSettings:      &projectSettingsSDK,
			ipAddresses:          &projectIPAddressesSDK,
			dataFederationLimits: limitsSDK,
			expectedTFModel: project.TfProjectDSModel{

				ID:           types.StringValue(projectID),
				ProjectID:    types.StringValue(projectID),
				Name:         types.StringValue(projectName),
				OrgID:        types.StringValue(projectOrgID),
				ClusterCount: types.Int64Value(clusterCount),
				IsCollectDatabaseSpecificsStatisticsEnabled: types.BoolValue(true),
				IsDataExplorerEnabled:                       types.BoolValue(true),
				IsExtendedStorageSizesEnabled:               types.BoolValue(true),
				IsPerformanceAdvisorEnabled:                 types.BoolValue(true),
				IsRealtimePerformancePanelEnabled:           types.BoolValue(true),
				IsSchemaAdvisorEnabled:                      types.BoolValue(true),
				Teams:                                       teamsDSTF,
				Limits:                                      limitsTF,
				IPAddresses:                                 &ipAddressesTF,
				Created:                                     types.StringValue("0001-01-01T00:00:00Z"),
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			resultModel := project.NewTFProjectDataSourceModel(context.Background(), tc.project, tc.teams, tc.projectSettings, tc.dataFederationLimits, tc.ipAddresses)
			if !assert.Equal(t, tc.expectedTFModel, resultModel) {
				t.Errorf("created terraform model did not match expected output")
			}
		})
	}
}

func TestTeamRoleListTFtoSDK(t *testing.T) {
	var rolesSet, _ = types.SetValueFrom(context.Background(), types.StringType, roles)
	teamsTF := []project.TfTeamModel{
		{
			TeamID:    types.StringValue("teamId"),
			RoleNames: rolesSet,
		},
	}
	testCases := []struct {
		name           string
		expectedResult *[]admin.TeamRole
		teamRolesTF    []project.TfTeamModel
	}{
		{
			name:           "Team roles",
			teamRolesTF:    teamsTF,
			expectedResult: &teamRolesSDK,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			resultModel := project.NewTeamRoleList(context.Background(), tc.teamRolesTF)
			if !assert.Equal(t, tc.expectedResult, resultModel) {
				t.Errorf("created terraform model did not match expected output")
			}
		})
	}
}

func TestTeamModelMapTF(t *testing.T) {
	teams := []project.TfTeamModel{
		{
			TeamID: types.StringValue("id1"),
		},
		{
			TeamID: types.StringValue("id2"),
		},
	}
	testCases := []struct {
		name           string
		expectedResult map[types.String]project.TfTeamModel
		teamRolesTF    []project.TfTeamModel
	}{
		{
			name:        "Team roles",
			teamRolesTF: teams,
			expectedResult: map[types.String]project.TfTeamModel{
				types.StringValue("id1"): teams[0],
				types.StringValue("id2"): teams[1],
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			resultModel := project.NewTfTeamModelMap(tc.teamRolesTF)
			if !assert.Equal(t, tc.expectedResult, resultModel) {
				t.Errorf("created terraform model did not match expected output")
			}
		})
	}
}

func TestLimitModelMapTF(t *testing.T) {
	limits := []project.TfLimitModel{
		{
			Name: types.StringValue("limit1"),
		},
		{
			Name: types.StringValue("limit2"),
		},
	}
	testCases := []struct {
		name           string
		expectedResult map[types.String]project.TfLimitModel
		limitsTF       []project.TfLimitModel
	}{
		{
			name:     "Limits",
			limitsTF: limits,
			expectedResult: map[types.String]project.TfLimitModel{
				types.StringValue("limit1"): limits[0],
				types.StringValue("limit2"): limits[1],
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			resultModel := project.NewTfLimitModelMap(tc.limitsTF)
			if !assert.Equal(t, tc.expectedResult, resultModel) {
				t.Errorf("created terraform model did not match expected output")
			}
		})
	}
}

func TestIPAddressesModelToTF(t *testing.T) {
	testCases := []struct {
		name           string
		sdkModel       *admin.GroupIPAddresses
		expectedResult project.TFIPAddressesModel
	}{
		{
			name:     "No response",
			sdkModel: nil,
			expectedResult: project.TFIPAddressesModel{
				Services: project.TFServicesModel{
					Clusters: []project.TFClusterIPsModel{},
				},
			},
		},
		{
			name: "Empty response when no clusters are created",
			sdkModel: &admin.GroupIPAddresses{
				GroupId: admin.PtrString(projectID),
				Services: &admin.GroupService{
					Clusters: &[]admin.ClusterIPAddresses{},
				},
			},
			expectedResult: project.TFIPAddressesModel{
				Services: project.TFServicesModel{
					Clusters: []project.TFClusterIPsModel{},
				},
			},
		},
		{
			name:           "Full response",
			sdkModel:       &projectIPAddressesSDK,
			expectedResult: ipAddressesTF,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			resultModel := project.NewTFIPAddressesModel(context.Background(), tc.sdkModel)
			if !assert.Equal(t, tc.expectedResult, resultModel) {
				t.Errorf("created terraform model did not match expected output")
			}
		})
	}
}
