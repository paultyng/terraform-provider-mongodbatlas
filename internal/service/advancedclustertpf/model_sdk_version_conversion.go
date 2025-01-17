package advancedclustertpf

import (
	admin20240530 "go.mongodb.org/atlas-sdk/v20240530005/admin"
	admin20240805 "go.mongodb.org/atlas-sdk/v20240805005/admin"
	"go.mongodb.org/atlas-sdk/v20241113004/admin"
)

// Conversions from one SDK model version to another are used to avoid duplicating our flatten/expand conversion functions.
// - These functions must not contain any business logic.
// - All will be removed once we rely on a single API version.

func ConvertClusterDescription20241023to20240805(clusterDescription *admin.ClusterDescription20240805) *admin20240805.ClusterDescription20240805 {
	return &admin20240805.ClusterDescription20240805{
		Name:                             clusterDescription.Name,
		ClusterType:                      clusterDescription.ClusterType,
		ReplicationSpecs:                 convertReplicationSpecs20241023to20240805(clusterDescription.ReplicationSpecs),
		BackupEnabled:                    clusterDescription.BackupEnabled,
		BiConnector:                      convertBiConnector20241023to20240805(clusterDescription.BiConnector),
		EncryptionAtRestProvider:         clusterDescription.EncryptionAtRestProvider,
		Labels:                           convertLabels20241023to20240805(clusterDescription.Labels),
		Tags:                             convertTag20241023to20240805(clusterDescription.Tags),
		MongoDBMajorVersion:              clusterDescription.MongoDBMajorVersion,
		PitEnabled:                       clusterDescription.PitEnabled,
		RootCertType:                     clusterDescription.RootCertType,
		TerminationProtectionEnabled:     clusterDescription.TerminationProtectionEnabled,
		VersionReleaseSystem:             clusterDescription.VersionReleaseSystem,
		GlobalClusterSelfManagedSharding: clusterDescription.GlobalClusterSelfManagedSharding,
		ReplicaSetScalingStrategy:        clusterDescription.ReplicaSetScalingStrategy,
		RedactClientLogData:              clusterDescription.RedactClientLogData,
		ConfigServerManagementMode:       clusterDescription.ConfigServerManagementMode,
	}
}

func convertReplicationSpecs20241023to20240805(replicationSpecs *[]admin.ReplicationSpec20240805) *[]admin20240805.ReplicationSpec20240805 {
	if replicationSpecs == nil {
		return nil
	}
	result := make([]admin20240805.ReplicationSpec20240805, len(*replicationSpecs))
	for i, replicationSpec := range *replicationSpecs {
		result[i] = admin20240805.ReplicationSpec20240805{
			Id:            replicationSpec.Id,
			ZoneName:      replicationSpec.ZoneName,
			ZoneId:        replicationSpec.ZoneId,
			RegionConfigs: convertCloudRegionConfig20241023to20240805(replicationSpec.RegionConfigs),
		}
	}
	return &result
}

func convertCloudRegionConfig20241023to20240805(cloudRegionConfig *[]admin.CloudRegionConfig20240805) *[]admin20240805.CloudRegionConfig20240805 {
	if cloudRegionConfig == nil {
		return nil
	}
	result := make([]admin20240805.CloudRegionConfig20240805, len(*cloudRegionConfig))
	for i, regionConfig := range *cloudRegionConfig {
		result[i] = admin20240805.CloudRegionConfig20240805{
			ProviderName:         regionConfig.ProviderName,
			RegionName:           regionConfig.RegionName,
			BackingProviderName:  regionConfig.BackingProviderName,
			Priority:             regionConfig.Priority,
			ElectableSpecs:       convertHardwareSpec20241023to20240805(regionConfig.ElectableSpecs),
			ReadOnlySpecs:        convertDedicatedHardwareSpec20241023to20240805(regionConfig.ReadOnlySpecs),
			AnalyticsSpecs:       convertDedicatedHardwareSpec20241023to20240805(regionConfig.AnalyticsSpecs),
			AutoScaling:          convertAdvancedAutoScalingSettings20241023to20240805(regionConfig.AutoScaling),
			AnalyticsAutoScaling: convertAdvancedAutoScalingSettings20241023to20240805(regionConfig.AnalyticsAutoScaling),
		}
	}
	return &result
}

func convertAdvancedAutoScalingSettings20241023to20240805(advancedAutoScalingSettings *admin.AdvancedAutoScalingSettings) *admin20240805.AdvancedAutoScalingSettings {
	if advancedAutoScalingSettings == nil {
		return nil
	}
	return &admin20240805.AdvancedAutoScalingSettings{
		Compute: convertAdvancedComputeAutoScaling20241023to20240805(advancedAutoScalingSettings.Compute),
		DiskGB:  convertDiskGBAutoScaling20241023to20240805(advancedAutoScalingSettings.DiskGB),
	}
}

func convertDiskGBAutoScaling20241023to20240805(diskGBAutoScaling *admin.DiskGBAutoScaling) *admin20240805.DiskGBAutoScaling {
	if diskGBAutoScaling == nil {
		return nil
	}
	return &admin20240805.DiskGBAutoScaling{
		Enabled: diskGBAutoScaling.Enabled,
	}
}

func convertAdvancedComputeAutoScaling20241023to20240805(advancedComputeAutoScaling *admin.AdvancedComputeAutoScaling) *admin20240805.AdvancedComputeAutoScaling {
	if advancedComputeAutoScaling == nil {
		return nil
	}
	return &admin20240805.AdvancedComputeAutoScaling{
		Enabled:          advancedComputeAutoScaling.Enabled,
		MaxInstanceSize:  advancedComputeAutoScaling.MaxInstanceSize,
		MinInstanceSize:  advancedComputeAutoScaling.MinInstanceSize,
		ScaleDownEnabled: advancedComputeAutoScaling.ScaleDownEnabled,
	}
}

func convertHardwareSpec20241023to20240805(hardwareSpec *admin.HardwareSpec20240805) *admin20240805.HardwareSpec20240805 {
	if hardwareSpec == nil {
		return nil
	}
	return &admin20240805.HardwareSpec20240805{
		DiskSizeGB:    hardwareSpec.DiskSizeGB,
		NodeCount:     hardwareSpec.NodeCount,
		DiskIOPS:      hardwareSpec.DiskIOPS,
		EbsVolumeType: hardwareSpec.EbsVolumeType,
		InstanceSize:  hardwareSpec.InstanceSize,
	}
}

func convertDedicatedHardwareSpec20241023to20240805(hardwareSpec *admin.DedicatedHardwareSpec20240805) *admin20240805.DedicatedHardwareSpec20240805 {
	if hardwareSpec == nil {
		return nil
	}
	return &admin20240805.DedicatedHardwareSpec20240805{
		DiskSizeGB:    hardwareSpec.DiskSizeGB,
		NodeCount:     hardwareSpec.NodeCount,
		DiskIOPS:      hardwareSpec.DiskIOPS,
		EbsVolumeType: hardwareSpec.EbsVolumeType,
		InstanceSize:  hardwareSpec.InstanceSize,
	}
}

func convertBiConnector20241023to20240805(biConnector *admin.BiConnector) *admin20240805.BiConnector {
	if biConnector == nil {
		return nil
	}
	return &admin20240805.BiConnector{
		ReadPreference: biConnector.ReadPreference,
		Enabled:        biConnector.Enabled,
	}
}

func convertLabels20241023to20240805(labels *[]admin.ComponentLabel) *[]admin20240805.ComponentLabel {
	if labels == nil {
		return nil
	}
	result := make([]admin20240805.ComponentLabel, len(*labels))
	for i, label := range *labels {
		result[i] = admin20240805.ComponentLabel{
			Key:   label.Key,
			Value: label.Value,
		}
	}
	return &result
}

func convertTag20241023to20240805(tags *[]admin.ResourceTag) *[]admin20240805.ResourceTag {
	if tags == nil {
		return nil
	}
	result := make([]admin20240805.ResourceTag, len(*tags))
	for i, tag := range *tags {
		result[i] = admin20240805.ResourceTag{
			Key:   tag.Key,
			Value: tag.Value,
		}
	}
	return &result
}

func ConvertRegionConfigSlice20241023to20240530(slice *[]admin.CloudRegionConfig20240805) *[]admin20240530.CloudRegionConfig {
	if slice == nil {
		return nil
	}
	cloudRegionSlice := *slice
	results := make([]admin20240530.CloudRegionConfig, len(cloudRegionSlice))
	for i := range len(cloudRegionSlice) {
		cloudRegion := cloudRegionSlice[i]
		results[i] = admin20240530.CloudRegionConfig{
			ElectableSpecs:       convertHardwareSpec20241023to20240530(cloudRegion.ElectableSpecs),
			Priority:             cloudRegion.Priority,
			ProviderName:         cloudRegion.ProviderName,
			RegionName:           cloudRegion.RegionName,
			AnalyticsAutoScaling: convertAdvancedAutoScalingSettings20241023to20240530(cloudRegion.AnalyticsAutoScaling),
			AnalyticsSpecs:       convertDedicatedHardwareSpec20241023to20240530(cloudRegion.AnalyticsSpecs),
			AutoScaling:          convertAdvancedAutoScalingSettings20241023to20240530(cloudRegion.AutoScaling),
			ReadOnlySpecs:        convertDedicatedHardwareSpec20241023to20240530(cloudRegion.ReadOnlySpecs),
			BackingProviderName:  cloudRegion.BackingProviderName,
		}
	}
	return &results
}

func convertHardwareSpec20241023to20240530(hwspec *admin.HardwareSpec20240805) *admin20240530.HardwareSpec {
	if hwspec == nil {
		return nil
	}
	return &admin20240530.HardwareSpec{
		DiskIOPS:      hwspec.DiskIOPS,
		EbsVolumeType: hwspec.EbsVolumeType,
		InstanceSize:  hwspec.InstanceSize,
		NodeCount:     hwspec.NodeCount,
	}
}

func convertAdvancedAutoScalingSettings20241023to20240530(settings *admin.AdvancedAutoScalingSettings) *admin20240530.AdvancedAutoScalingSettings {
	if settings == nil {
		return nil
	}
	return &admin20240530.AdvancedAutoScalingSettings{
		Compute: convertAdvancedComputeAutoScaling20241023to20240530(settings.Compute),
		DiskGB:  convertDiskGBAutoScaling20241023to20240530(settings.DiskGB),
	}
}

func convertAdvancedComputeAutoScaling20241023to20240530(settings *admin.AdvancedComputeAutoScaling) *admin20240530.AdvancedComputeAutoScaling {
	if settings == nil {
		return nil
	}
	return &admin20240530.AdvancedComputeAutoScaling{
		Enabled:          settings.Enabled,
		MaxInstanceSize:  settings.MaxInstanceSize,
		MinInstanceSize:  settings.MinInstanceSize,
		ScaleDownEnabled: settings.ScaleDownEnabled,
	}
}

func convertDiskGBAutoScaling20241023to20240530(settings *admin.DiskGBAutoScaling) *admin20240530.DiskGBAutoScaling {
	if settings == nil {
		return nil
	}
	return &admin20240530.DiskGBAutoScaling{
		Enabled: settings.Enabled,
	}
}

func convertDedicatedHardwareSpec20241023to20240530(spec *admin.DedicatedHardwareSpec20240805) *admin20240530.DedicatedHardwareSpec {
	if spec == nil {
		return nil
	}
	return &admin20240530.DedicatedHardwareSpec{
		NodeCount:     spec.NodeCount,
		DiskIOPS:      spec.DiskIOPS,
		EbsVolumeType: spec.EbsVolumeType,
		InstanceSize:  spec.InstanceSize,
	}
}
