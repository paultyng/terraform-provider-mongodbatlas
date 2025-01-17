package advancedclustertpf_test

import (
	"testing"

	"github.com/mongodb/terraform-provider-mongodbatlas/internal/common/conversion"
	"github.com/mongodb/terraform-provider-mongodbatlas/internal/service/advancedclustertpf"
	"github.com/stretchr/testify/assert"
	admin20240805 "go.mongodb.org/atlas-sdk/v20240805005/admin"
	"go.mongodb.org/atlas-sdk/v20241113004/admin"
)

func TestConvertClusterDescription20241023to20240805(t *testing.T) {
	var (
		clusterName                = "clusterName"
		clusterType                = "REPLICASET"
		earProvider                = "AWS"
		booleanValue               = true
		mongoDBMajorVersion        = "7.0"
		rootCertType               = "rootCertType"
		replicaSetScalingStrategy  = "WORKLOAD_TYPE"
		configServerManagementMode = "ATLAS_MANAGED"
		readPreference             = "primary"
		zoneName                   = "z1"
		id                         = "id1"
		regionConfigProvider       = "AWS"
		region                     = "EU_WEST_1"
		priority                   = 7
		instanceSize               = "M10"
		nodeCount                  = 3
		diskSizeGB                 = 30.3
		ebsVolumeType              = "STANDARD"
		diskIOPS                   = 100
	)
	testCases := []struct {
		input          *admin.ClusterDescription20240805
		expectedOutput *admin20240805.ClusterDescription20240805
		name           string
	}{
		{
			name: "Converts cluster description from 20241023 to 20240805",
			input: &admin.ClusterDescription20240805{
				Name:        conversion.StringPtr(clusterName),
				ClusterType: conversion.StringPtr(clusterType),
				ReplicationSpecs: &[]admin.ReplicationSpec20240805{
					{
						Id:       conversion.StringPtr(id),
						ZoneName: conversion.StringPtr(zoneName),
						RegionConfigs: &[]admin.CloudRegionConfig20240805{
							{
								ProviderName:        conversion.StringPtr(regionConfigProvider),
								RegionName:          conversion.StringPtr(region),
								BackingProviderName: conversion.StringPtr(regionConfigProvider),
								Priority:            conversion.IntPtr(priority),
								AnalyticsSpecs: &admin.DedicatedHardwareSpec20240805{
									InstanceSize:  conversion.StringPtr(instanceSize),
									NodeCount:     conversion.IntPtr(nodeCount),
									DiskSizeGB:    conversion.Pointer(diskSizeGB),
									EbsVolumeType: conversion.StringPtr(ebsVolumeType),
									DiskIOPS:      conversion.IntPtr(diskIOPS),
								},
								ElectableSpecs: &admin.HardwareSpec20240805{
									InstanceSize:  conversion.StringPtr(instanceSize),
									NodeCount:     conversion.IntPtr(nodeCount),
									DiskSizeGB:    conversion.Pointer(diskSizeGB),
									EbsVolumeType: conversion.StringPtr(ebsVolumeType),
									DiskIOPS:      conversion.IntPtr(diskIOPS),
								},
								AutoScaling: &admin.AdvancedAutoScalingSettings{
									Compute: &admin.AdvancedComputeAutoScaling{
										Enabled:          conversion.Pointer(booleanValue),
										MaxInstanceSize:  conversion.Pointer(instanceSize),
										MinInstanceSize:  conversion.Pointer(instanceSize),
										ScaleDownEnabled: conversion.Pointer(booleanValue),
									},
									DiskGB: &admin.DiskGBAutoScaling{
										Enabled: conversion.Pointer(booleanValue),
									},
								},
							},
						},
					},
				},
				BackupEnabled: conversion.Pointer(booleanValue),
				BiConnector: &admin.BiConnector{
					Enabled:        conversion.Pointer(booleanValue),
					ReadPreference: conversion.StringPtr(readPreference),
				},
				EncryptionAtRestProvider: conversion.StringPtr(earProvider),
				Labels: &[]admin.ComponentLabel{
					{Key: conversion.StringPtr("key1"), Value: conversion.StringPtr("value1")},
					{Key: conversion.StringPtr("key2"), Value: conversion.StringPtr("value2")},
				},
				Tags: &[]admin.ResourceTag{
					{Key: "key1", Value: "value1"},
					{Key: "key2", Value: "value2"},
				},
				MongoDBMajorVersion:              conversion.StringPtr(mongoDBMajorVersion),
				PitEnabled:                       conversion.Pointer(booleanValue),
				RootCertType:                     conversion.StringPtr(rootCertType),
				TerminationProtectionEnabled:     conversion.Pointer(booleanValue),
				VersionReleaseSystem:             conversion.StringPtr(""),
				GlobalClusterSelfManagedSharding: conversion.Pointer(booleanValue),
				ReplicaSetScalingStrategy:        conversion.StringPtr(replicaSetScalingStrategy),
				RedactClientLogData:              conversion.Pointer(booleanValue),
				ConfigServerManagementMode:       conversion.StringPtr(configServerManagementMode),
			},
			expectedOutput: &admin20240805.ClusterDescription20240805{
				Name:        conversion.StringPtr(clusterName),
				ClusterType: conversion.StringPtr(clusterType),
				ReplicationSpecs: &[]admin20240805.ReplicationSpec20240805{
					{
						Id:       conversion.StringPtr(id),
						ZoneName: conversion.StringPtr(zoneName),
						RegionConfigs: &[]admin20240805.CloudRegionConfig20240805{
							{
								ProviderName:        conversion.StringPtr(regionConfigProvider),
								RegionName:          conversion.StringPtr(region),
								BackingProviderName: conversion.StringPtr(regionConfigProvider),
								Priority:            conversion.IntPtr(priority),
								AnalyticsSpecs: &admin20240805.DedicatedHardwareSpec20240805{
									InstanceSize:  conversion.StringPtr(instanceSize),
									NodeCount:     conversion.IntPtr(nodeCount),
									DiskSizeGB:    conversion.Pointer(diskSizeGB),
									EbsVolumeType: conversion.StringPtr(ebsVolumeType),
									DiskIOPS:      conversion.IntPtr(diskIOPS),
								},
								ElectableSpecs: &admin20240805.HardwareSpec20240805{
									InstanceSize:  conversion.StringPtr(instanceSize),
									NodeCount:     conversion.IntPtr(nodeCount),
									DiskSizeGB:    conversion.Pointer(diskSizeGB),
									EbsVolumeType: conversion.StringPtr(ebsVolumeType),
									DiskIOPS:      conversion.IntPtr(diskIOPS),
								},
								AutoScaling: &admin20240805.AdvancedAutoScalingSettings{
									Compute: &admin20240805.AdvancedComputeAutoScaling{
										Enabled:          conversion.Pointer(booleanValue),
										MaxInstanceSize:  conversion.Pointer(instanceSize),
										MinInstanceSize:  conversion.Pointer(instanceSize),
										ScaleDownEnabled: conversion.Pointer(booleanValue),
									},
									DiskGB: &admin20240805.DiskGBAutoScaling{
										Enabled: conversion.Pointer(booleanValue),
									},
								},
							},
						},
					},
				},
				BackupEnabled: conversion.Pointer(booleanValue),
				BiConnector: &admin20240805.BiConnector{
					Enabled:        conversion.Pointer(booleanValue),
					ReadPreference: conversion.StringPtr(readPreference),
				},
				EncryptionAtRestProvider: conversion.StringPtr(earProvider),
				Labels: &[]admin20240805.ComponentLabel{
					{Key: conversion.StringPtr("key1"), Value: conversion.StringPtr("value1")},
					{Key: conversion.StringPtr("key2"), Value: conversion.StringPtr("value2")},
				},
				Tags: &[]admin20240805.ResourceTag{
					{Key: "key1", Value: "value1"},
					{Key: "key2", Value: "value2"},
				},
				MongoDBMajorVersion:              conversion.StringPtr(mongoDBMajorVersion),
				PitEnabled:                       conversion.Pointer(booleanValue),
				RootCertType:                     conversion.StringPtr(rootCertType),
				TerminationProtectionEnabled:     conversion.Pointer(booleanValue),
				VersionReleaseSystem:             conversion.StringPtr(""),
				GlobalClusterSelfManagedSharding: conversion.Pointer(booleanValue),
				ReplicaSetScalingStrategy:        conversion.StringPtr(replicaSetScalingStrategy),
				RedactClientLogData:              conversion.Pointer(booleanValue),
				ConfigServerManagementMode:       conversion.StringPtr(configServerManagementMode),
			},
		},
		{
			name:  "Converts cluster description from 20241023 to 20240805 with nil values",
			input: &admin.ClusterDescription20240805{},
			expectedOutput: &admin20240805.ClusterDescription20240805{
				ReplicationSpecs: nil,
				BiConnector:      nil,
				Labels:           nil,
				Tags:             nil,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := advancedclustertpf.ConvertClusterDescription20241023to20240805(tc.input)
			assert.Equal(t, tc.expectedOutput, result)
		})
	}
}
