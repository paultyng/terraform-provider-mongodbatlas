package conversion_test

import (
	"testing"

	"github.com/mongodb/terraform-provider-mongodbatlas/internal/common/conversion"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/atlas-sdk/v20241023002/admin"
)

func TestJsonPatchReplicationSpecs(t *testing.T) {
	var (
		idGlobal                    = "id_root"
		idReplicationSpec1          = "id_replicationSpec1"
		idReplicationSpec2          = "id_replicationSpec2"
		replicationSpec1ZoneNameOld = "replicationSpec1_zoneName_old"
		replicationSpec1ZoneNameNew = "replicationSpec1_zoneName_new"
		replicationSpec1ZoneID      = "replicationSpec1_zoneId"
		replicationSpec2ZoneID      = "replicationSpec2_zoneId"
		replicationSpec2ZoneName    = "replicationSpec2_zoneName"
		rootName                    = "my-cluster"
		rootNameUpdated             = "my-cluster-updated"
		state                       = admin.ClusterDescription20240805{
			Id:   &idGlobal,
			Name: &rootName,
			ReplicationSpecs: &[]admin.ReplicationSpec20240805{
				{
					Id:       &idReplicationSpec1,
					ZoneId:   &replicationSpec1ZoneID,
					ZoneName: &replicationSpec1ZoneNameOld,
				},
			},
		}
		planOptionalUpdated = admin.ClusterDescription20240805{
			Name: &rootName,
			ReplicationSpecs: &[]admin.ReplicationSpec20240805{
				{
					ZoneName: &replicationSpec1ZoneNameNew,
				},
			},
		}
		planNewListEntry = admin.ClusterDescription20240805{
			ReplicationSpecs: &[]admin.ReplicationSpec20240805{
				{
					ZoneName: &replicationSpec1ZoneNameOld,
				},
				{
					ZoneName: &replicationSpec2ZoneName,
				},
			},
		}
		planNameDifferentAndEnableBackup = admin.ClusterDescription20240805{
			Name:          &rootNameUpdated,
			BackupEnabled: conversion.Pointer(true),
		}
		planNoChanges = admin.ClusterDescription20240805{
			ReplicationSpecs: &[]admin.ReplicationSpec20240805{
				{
					ZoneName: &replicationSpec1ZoneNameOld,
				},
			},
		}
		testCases = map[string]struct {
			state         *admin.ClusterDescription20240805
			plan          *admin.ClusterDescription20240805
			patchExpected *admin.ClusterDescription20240805
			noChanges     bool
		}{
			"ComputedValues from the state are added to plan and unchanged attributes are not included": {
				state: &state,
				plan:  &planOptionalUpdated,
				patchExpected: &admin.ClusterDescription20240805{
					ReplicationSpecs: &[]admin.ReplicationSpec20240805{
						{
							Id:       &idReplicationSpec1,
							ZoneId:   &replicationSpec1ZoneID,
							ZoneName: &replicationSpec1ZoneNameNew,
						},
					},
				},
			},
			"New list entry added should be included": {
				state: &state,
				plan:  &planNewListEntry,
				patchExpected: &admin.ClusterDescription20240805{
					ReplicationSpecs: &[]admin.ReplicationSpec20240805{
						{
							Id:       &idReplicationSpec1,
							ZoneId:   &replicationSpec1ZoneID,
							ZoneName: &replicationSpec1ZoneNameOld,
						},
						{
							ZoneName: &replicationSpec2ZoneName,
						},
					},
				},
			},
			"Removed list entry should be included": {
				state: &admin.ClusterDescription20240805{
					ReplicationSpecs: &[]admin.ReplicationSpec20240805{
						{
							Id:       &idReplicationSpec1,
							ZoneId:   &replicationSpec1ZoneID,
							ZoneName: &replicationSpec1ZoneNameOld,
						},
						{
							Id:       &idReplicationSpec2,
							ZoneName: &replicationSpec2ZoneName,
							ZoneId:   &replicationSpec2ZoneID,
						},
					},
				},
				plan: &admin.ClusterDescription20240805{
					ReplicationSpecs: &[]admin.ReplicationSpec20240805{
						{
							Id:       &idReplicationSpec1,
							ZoneId:   &replicationSpec1ZoneID,
							ZoneName: &replicationSpec1ZoneNameOld,
						},
					},
				},
				patchExpected: &admin.ClusterDescription20240805{
					ReplicationSpecs: &[]admin.ReplicationSpec20240805{
						{
							Id:       &idReplicationSpec1,
							ZoneId:   &replicationSpec1ZoneID,
							ZoneName: &replicationSpec1ZoneNameOld,
						},
					},
				},
			},
			"Region Config changes are included in patch": {
				state: &admin.ClusterDescription20240805{
					ReplicationSpecs: &[]admin.ReplicationSpec20240805{
						{
							Id: &idReplicationSpec1,
							RegionConfigs: &[]admin.CloudRegionConfig20240805{
								{
									Priority: conversion.Pointer(1),
								},
							},
						},
					},
				},
				plan: &admin.ClusterDescription20240805{
					ReplicationSpecs: &[]admin.ReplicationSpec20240805{
						{
							Id: &idReplicationSpec1,
							RegionConfigs: &[]admin.CloudRegionConfig20240805{
								{
									Priority: conversion.Pointer(1),
								},
								{
									Priority: conversion.Pointer(2),
								},
							},
						},
					},
				},
				patchExpected: &admin.ClusterDescription20240805{
					ReplicationSpecs: &[]admin.ReplicationSpec20240805{
						{
							Id: &idReplicationSpec1,
							RegionConfigs: &[]admin.CloudRegionConfig20240805{
								{
									Priority: conversion.Pointer(1),
								},
								{
									Priority: conversion.Pointer(2),
								},
							},
						},
					},
				},
			},
			"Name change and backup enabled added": {
				state: &state,
				plan:  &planNameDifferentAndEnableBackup,
				patchExpected: &admin.ClusterDescription20240805{
					Name:          &rootNameUpdated,
					BackupEnabled: conversion.Pointer(true),
				},
			},
			"No Changes when only computed attributes are not in plan": {
				state:         &state,
				plan:          &planNoChanges,
				noChanges:     true,
				patchExpected: &admin.ClusterDescription20240805{},
			},
		}
	)
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			if name == "Removed list entry should be included" {
				t.Log("This test case is expected to fail due to the current implementation")
			}
			patchReq := &admin.ClusterDescription20240805{}
			noChanges, err := conversion.PatchPayloadNoChanges(tc.state, tc.plan, patchReq)
			require.NoError(t, err)
			assert.Equal(t, tc.noChanges, noChanges)
			assert.Equal(t, tc.patchExpected, patchReq)
		})
	}
}

func TestJsonPatchAdvancedConfig(t *testing.T) {
	var (
		state = admin.ClusterDescriptionProcessArgs20240805{
			JavascriptEnabled: conversion.Pointer(true),
		}
		testCases = map[string]struct {
			state         *admin.ClusterDescriptionProcessArgs20240805
			plan          *admin.ClusterDescriptionProcessArgs20240805
			patchExpected *admin.ClusterDescriptionProcessArgs20240805
			noChanges     bool
		}{
			"JavascriptEnabled is set to false": {
				state: &state,
				plan: &admin.ClusterDescriptionProcessArgs20240805{
					JavascriptEnabled: conversion.Pointer(false),
				},
				patchExpected: &admin.ClusterDescriptionProcessArgs20240805{
					JavascriptEnabled: conversion.Pointer(false),
				},
			},
			"JavascriptEnabled is set to null leads to no changes": {
				state:         &state,
				plan:          &admin.ClusterDescriptionProcessArgs20240805{},
				patchExpected: &admin.ClusterDescriptionProcessArgs20240805{},
				noChanges:     true,
			},
			"JavascriptEnabled state equals plan leads to no changes": {
				state:         &state,
				plan:          &state,
				patchExpected: &admin.ClusterDescriptionProcessArgs20240805{},
				noChanges:     true,
			},
			"Adding NoTableScan changes the plan payload and but doesn't include old value of JavascriptEnabled": {
				state: &state,
				plan: &admin.ClusterDescriptionProcessArgs20240805{
					NoTableScan: conversion.Pointer(true),
				},
				patchExpected: &admin.ClusterDescriptionProcessArgs20240805{
					NoTableScan: conversion.Pointer(true),
				},
			},
			"Nil plan should return no changes": {
				state:         &state,
				plan:          nil,
				patchExpected: &admin.ClusterDescriptionProcessArgs20240805{},
				noChanges:     true,
			},
		}
	)
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			patchReq := &admin.ClusterDescriptionProcessArgs20240805{}
			noChanges, err := conversion.PatchPayloadNoChanges(tc.state, tc.plan, patchReq)
			require.NoError(t, err)
			assert.Equal(t, tc.noChanges, noChanges)
			assert.Equal(t, tc.patchExpected, patchReq)
		})
	}
}
