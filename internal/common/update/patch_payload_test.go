package update_test

import (
	"testing"

	"github.com/mongodb/terraform-provider-mongodbatlas/internal/common/conversion"
	"github.com/mongodb/terraform-provider-mongodbatlas/internal/common/update"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/atlas-sdk/v20241113003/admin"
)

func TestPatchReplicationSpecs(t *testing.T) {
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
		stateReplicationSpecs       = []admin.ReplicationSpec20240805{
			{
				Id:       &idReplicationSpec1,
				ZoneId:   &replicationSpec1ZoneID,
				ZoneName: &replicationSpec1ZoneNameOld,
			},
		}
		state = admin.ClusterDescription20240805{
			Id:               &idGlobal,
			Name:             &rootName,
			ReplicationSpecs: &stateReplicationSpecs,
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
			options       []update.PatchOptions
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
				patchExpected: nil,
			},
			"Forced changes when forceUpdateAttr set": {
				state: &state,
				plan:  &planNoChanges,
				patchExpected: &admin.ClusterDescription20240805{
					ReplicationSpecs: &stateReplicationSpecs,
				},
				options: []update.PatchOptions{
					{ForceUpdateAttr: []string{"replicationSpecs"}},
				},
			},
			"Empty array should return no changes": {
				state: &admin.ClusterDescription20240805{
					Labels: &[]admin.ComponentLabel{},
				},
				plan: &admin.ClusterDescription20240805{
					Labels: &[]admin.ComponentLabel{},
				},
				patchExpected: nil,
			},
			"diskSizeGb ignored in state": {
				state:         clusterDescriptionDiskSizeNodeCount(50.0, 3, conversion.Pointer(50.0), 0, conversion.Pointer(3500)),
				plan:          clusterDescriptionDiskSizeNodeCount(55.0, 3, nil, 0, nil),
				patchExpected: clusterDescriptionDiskSizeNodeCount(55.0, 3, nil, 0, conversion.Pointer(3500)),
				options: []update.PatchOptions{
					{
						IgnoreInStateSuffix: []string{"diskSizeGB"},
					},
				},
			},
			"regionConfigs ignored in state but diskIOPS included": {
				state:         clusterDescriptionDiskSizeNodeCount(50.0, 3, conversion.Pointer(50.0), 0, conversion.Pointer(3500)),
				plan:          clusterDescriptionDiskSizeNodeCount(55.0, 3, nil, 0, nil),
				patchExpected: clusterDescriptionDiskSizeNodeCount(55.0, 3, nil, 0, conversion.Pointer(3500)),
				options: []update.PatchOptions{
					{
						IgnoreInStatePrefix:  []string{"regionConfigs"},
						IncludeInStateSuffix: []string{"diskIOPS"},
					},
				},
			},
		}
	)
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			patchReq, err := update.PatchPayload(tc.state, tc.plan, tc.options...)
			require.NoError(t, err)
			assert.Equal(t, tc.patchExpected, patchReq)
		})
	}
}

func TestPatchAdvancedConfig(t *testing.T) {
	var (
		state = admin.ClusterDescriptionProcessArgs20240805{
			JavascriptEnabled: conversion.Pointer(true),
		}
		testCases = map[string]struct {
			state         *admin.ClusterDescriptionProcessArgs20240805
			plan          *admin.ClusterDescriptionProcessArgs20240805
			patchExpected *admin.ClusterDescriptionProcessArgs20240805
			options       []update.PatchOptions
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
				patchExpected: nil,
			},
			"JavascriptEnabled state equals plan leads to no changes": {
				state:         &state,
				plan:          &state,
				patchExpected: nil,
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
				patchExpected: nil,
			},
		}
	)
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			patchReq, err := update.PatchPayload(tc.state, tc.plan, tc.options...)
			require.NoError(t, err)
			assert.Equal(t, tc.patchExpected, patchReq)
		})
	}
}

func TestIsEmpty(t *testing.T) {
	assert.True(t, update.IsZeroValues(&admin.ClusterDescription20240805{}))
	var myVar admin.ClusterDescription20240805
	assert.True(t, update.IsZeroValues(&myVar))
	assert.False(t, update.IsZeroValues(&admin.ClusterDescription20240805{Name: conversion.Pointer("my-cluster")}))
}

func clusterDescriptionDiskSizeNodeCount(diskSizeGBElectable float64, nodeCountElectable int, diskSizeGBReadOnly *float64, nodeCountReadOnly int, diskIopsState *int) *admin.ClusterDescription20240805 {
	return &admin.ClusterDescription20240805{
		ReplicationSpecs: &[]admin.ReplicationSpec20240805{
			{
				RegionConfigs: &[]admin.CloudRegionConfig20240805{
					{
						ElectableSpecs: &admin.HardwareSpec20240805{
							NodeCount:  &nodeCountElectable,
							DiskSizeGB: &diskSizeGBElectable,
							DiskIOPS:   diskIopsState,
						},
						ReadOnlySpecs: &admin.DedicatedHardwareSpec20240805{
							NodeCount:  &nodeCountReadOnly,
							DiskSizeGB: diskSizeGBReadOnly,
							DiskIOPS:   diskIopsState,
						},
					},
				},
			},
		},
	}
}
