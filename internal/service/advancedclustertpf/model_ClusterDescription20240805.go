package advancedclustertpf

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/mongodb/terraform-provider-mongodbatlas/internal/common/conversion"
	"go.mongodb.org/atlas-sdk/v20241113004/admin"
)

const (
	errorZoneNameNotSet          = "zoneName is required for legacy schema"
	errorNumShardsNotSet         = "numShards not set for zoneName %s"
	errorReplicationSpecIDNotSet = "replicationSpecID not set for zoneName %s"
)

type ExtraAPIInfo struct {
	ZoneNameNumShards          map[string]int64
	ZoneNameReplicationSpecIDs map[string]string
	ContainerIDs               map[string]string
	UsingLegacySchema          bool
	ForceLegacySchemaFailed    bool
}

func NewTFModel(ctx context.Context, input *admin.ClusterDescription20240805, timeout timeouts.Value, diags *diag.Diagnostics, apiInfo ExtraAPIInfo) *TFModel {
	biConnector := NewBiConnectorConfigObjType(ctx, input.BiConnector, diags)
	connectionStrings := NewConnectionStringsObjType(ctx, input.ConnectionStrings, diags)
	labels := NewLabelsObjType(ctx, input.Labels, diags)
	replicationSpecs := NewReplicationSpecsObjType(ctx, input.ReplicationSpecs, diags, &apiInfo)
	tags := NewTagsObjType(ctx, input.Tags, diags)
	pinnedFCV := NewPinnedFCVObjType(ctx, input, diags)
	if diags.HasError() {
		return nil
	}
	return &TFModel{
		AcceptDataRisksAndForceReplicaSetReconfig: types.StringPointerValue(conversion.TimePtrToStringPtr(input.AcceptDataRisksAndForceReplicaSetReconfig)),
		BackupEnabled:                    types.BoolValue(conversion.SafeValue(input.BackupEnabled)),
		BiConnectorConfig:                biConnector,
		ClusterType:                      types.StringValue(conversion.SafeValue(input.ClusterType)),
		ConfigServerManagementMode:       types.StringValue(conversion.SafeValue(input.ConfigServerManagementMode)),
		ConfigServerType:                 types.StringValue(conversion.SafeValue(input.ConfigServerType)),
		ConnectionStrings:                connectionStrings,
		CreateDate:                       types.StringValue(conversion.SafeValue(conversion.TimePtrToStringPtr(input.CreateDate))),
		DiskSizeGB:                       types.Float64PointerValue(findRegionRootDiskSize(input.ReplicationSpecs)),
		EncryptionAtRestProvider:         types.StringValue(conversion.SafeValue(input.EncryptionAtRestProvider)),
		GlobalClusterSelfManagedSharding: types.BoolValue(conversion.SafeValue(input.GlobalClusterSelfManagedSharding)),
		ProjectID:                        types.StringValue(conversion.SafeValue(input.GroupId)),
		ClusterID:                        types.StringValue(conversion.SafeValue(input.Id)),
		Labels:                           labels,
		MongoDBMajorVersion:              types.StringValue(conversion.SafeValue(input.MongoDBMajorVersion)),
		MongoDBVersion:                   types.StringValue(conversion.SafeValue(input.MongoDBVersion)),
		Name:                             types.StringValue(conversion.SafeValue(input.Name)),
		Paused:                           types.BoolValue(conversion.SafeValue(input.Paused)),
		PitEnabled:                       types.BoolValue(conversion.SafeValue(input.PitEnabled)),
		RedactClientLogData:              types.BoolValue(conversion.SafeValue(input.RedactClientLogData)),
		ReplicaSetScalingStrategy:        types.StringValue(conversion.SafeValue(input.ReplicaSetScalingStrategy)),
		ReplicationSpecs:                 replicationSpecs,
		RootCertType:                     types.StringValue(conversion.SafeValue(input.RootCertType)),
		StateName:                        types.StringValue(conversion.SafeValue(input.StateName)),
		Tags:                             tags,
		TerminationProtectionEnabled:     types.BoolValue(conversion.SafeValue(input.TerminationProtectionEnabled)),
		VersionReleaseSystem:             types.StringValue(conversion.SafeValue(input.VersionReleaseSystem)),
		PinnedFCV:                        pinnedFCV,
		Timeouts:                         timeout,
	}
}

func NewBiConnectorConfigObjType(ctx context.Context, input *admin.BiConnector, diags *diag.Diagnostics) types.Object {
	if input == nil {
		return types.ObjectNull(BiConnectorConfigObjType.AttrTypes)
	}
	tfModel := TFBiConnectorModel{
		Enabled:        types.BoolValue(conversion.SafeValue(input.Enabled)),
		ReadPreference: types.StringValue(conversion.SafeValue(input.ReadPreference)),
	}
	objType, diagsLocal := types.ObjectValueFrom(ctx, BiConnectorConfigObjType.AttrTypes, tfModel)
	diags.Append(diagsLocal...)
	return objType
}

func NewConnectionStringsObjType(ctx context.Context, input *admin.ClusterConnectionStrings, diags *diag.Diagnostics) types.Object {
	if input == nil {
		return types.ObjectNull(ConnectionStringsObjType.AttrTypes)
	}
	privateEndpoint := NewPrivateEndpointObjType(ctx, input.PrivateEndpoint, diags)
	tfModel := TFConnectionStringsModel{
		Private:         types.StringValue(conversion.SafeValue(input.Private)),
		PrivateEndpoint: privateEndpoint,
		PrivateSrv:      types.StringValue(conversion.SafeValue(input.PrivateSrv)),
		Standard:        types.StringValue(conversion.SafeValue(input.Standard)),
		StandardSrv:     types.StringValue(conversion.SafeValue(input.StandardSrv)),
	}
	objType, diagsLocal := types.ObjectValueFrom(ctx, ConnectionStringsObjType.AttrTypes, tfModel)
	diags.Append(diagsLocal...)
	return objType
}

func NewLabelsObjType(ctx context.Context, input *[]admin.ComponentLabel, diags *diag.Diagnostics) types.Set {
	if input == nil {
		return types.SetNull(LabelsObjType)
	}
	tfModels := make([]TFLabelsModel, 0, len(*input))
	for _, item := range *input {
		key := conversion.SafeValue(item.Key)
		value := conversion.SafeValue(item.Value)
		if key == IgnoreLabelKey {
			continue
		}
		tfModels = append(tfModels, TFLabelsModel{Key: types.StringValue(key), Value: types.StringValue(value)})
	}
	setType, diagsLocal := types.SetValueFrom(ctx, LabelsObjType, tfModels)
	diags.Append(diagsLocal...)
	return setType
}

func NewReplicationSpecsObjType(ctx context.Context, input *[]admin.ReplicationSpec20240805, diags *diag.Diagnostics, apiInfo *ExtraAPIInfo) types.List {
	if input == nil {
		return types.ListNull(ReplicationSpecsObjType)
	}
	var tfModels *[]TFReplicationSpecsModel
	if apiInfo.UsingLegacySchema {
		tfModels = convertReplicationSpecsLegacy(ctx, input, diags, apiInfo)
	} else {
		tfModels = convertReplicationSpecs(ctx, input, diags, apiInfo)
	}
	if diags.HasError() {
		return types.ListNull(ReplicationSpecsObjType)
	}
	listType, diagsLocal := types.ListValueFrom(ctx, ReplicationSpecsObjType, *tfModels)
	diags.Append(diagsLocal...)
	return listType
}

func NewPinnedFCVObjType(ctx context.Context, cluster *admin.ClusterDescription20240805, diags *diag.Diagnostics) types.Object {
	if cluster.FeatureCompatibilityVersionExpirationDate == nil {
		return types.ObjectNull(PinnedFCVObjType.AttrTypes)
	}
	tfModel := TFPinnedFCVModel{
		Version:        types.StringValue(cluster.GetFeatureCompatibilityVersion()),
		ExpirationDate: types.StringValue(conversion.TimeToString(cluster.GetFeatureCompatibilityVersionExpirationDate())),
	}
	objType, diagsLocal := types.ObjectValueFrom(ctx, PinnedFCVObjType.AttrTypes, tfModel)
	diags.Append(diagsLocal...)
	return objType
}

func convertReplicationSpecs(ctx context.Context, input *[]admin.ReplicationSpec20240805, diags *diag.Diagnostics, apiInfo *ExtraAPIInfo) *[]TFReplicationSpecsModel {
	tfModels := make([]TFReplicationSpecsModel, len(*input))
	for i, item := range *input {
		regionConfigs := NewRegionConfigsObjType(ctx, item.RegionConfigs, diags)
		zoneName := item.GetZoneName()
		if zoneName == "" {
			diags.AddError(errorZoneNameNotSet, errorZoneNameNotSet)
			return &tfModels
		}
		legacyID := apiInfo.ZoneNameReplicationSpecIDs[zoneName]
		containerIDs := selectContainerIDs(&item, apiInfo.ContainerIDs)
		tfModels[i] = TFReplicationSpecsModel{
			Id:            types.StringValue(legacyID),
			ExternalId:    types.StringValue(conversion.SafeValue(item.Id)),
			NumShards:     types.Int64Value(1),
			ContainerId:   conversion.ToTFMapOfString(ctx, diags, &containerIDs),
			RegionConfigs: regionConfigs,
			ZoneId:        types.StringValue(conversion.SafeValue(item.ZoneId)),
			ZoneName:      types.StringValue(conversion.SafeValue(item.ZoneName)),
		}
	}
	return &tfModels
}

func selectContainerIDs(spec *admin.ReplicationSpec20240805, allIDs map[string]string) map[string]string {
	containerIDs := map[string]string{}
	regions := spec.GetRegionConfigs()
	for i := range regions {
		regionConfig := regions[i]
		providerName := regionConfig.GetProviderName()
		key := containerIDKey(providerName, regionConfig.GetRegionName())
		value := allIDs[key]
		// Should be no hard failure if not found, as it is not required for TENANT, error responsibility in resolveContainerIDs
		if value == "" {
			continue
		}
		containerIDs[key] = value
	}
	return containerIDs
}

func convertReplicationSpecsLegacy(ctx context.Context, input *[]admin.ReplicationSpec20240805, diags *diag.Diagnostics, apiInfo *ExtraAPIInfo) *[]TFReplicationSpecsModel {
	tfModels := []TFReplicationSpecsModel{}
	tfModelsSkipIndexes := []int{}
	for i, item := range *input {
		if slices.Contains(tfModelsSkipIndexes, i) {
			continue
		}
		regionConfigs := NewRegionConfigsObjType(ctx, item.RegionConfigs, diags)
		zoneName := item.GetZoneName()
		if zoneName == "" {
			diags.AddError(errorZoneNameNotSet, errorZoneNameNotSet)
			return &tfModels
		}
		numShards, ok := apiInfo.ZoneNameNumShards[zoneName]
		errMsg := []string{}
		if !ok {
			errMsg = append(errMsg, fmt.Sprintf(errorNumShardsNotSet, zoneName))
		}
		legacyID, ok := apiInfo.ZoneNameReplicationSpecIDs[zoneName]
		if !ok {
			errMsg = append(errMsg, fmt.Sprintf(errorReplicationSpecIDNotSet, zoneName))
		}
		if len(errMsg) > 0 {
			diags.AddError("replicationSpecsLegacySchema", strings.Join(errMsg, ", "))
			return &tfModels
		}
		if numShards > 1 {
			for j := 1; j < int(numShards); j++ {
				tfModelsSkipIndexes = append(tfModelsSkipIndexes, i+j)
			}
		}
		containerIDs := selectContainerIDs(&item, apiInfo.ContainerIDs)
		tfModels = append(tfModels, TFReplicationSpecsModel{
			ContainerId:   conversion.ToTFMapOfString(ctx, diags, &containerIDs),
			ExternalId:    types.StringValue(conversion.SafeValue(item.Id)),
			Id:            types.StringValue(legacyID),
			RegionConfigs: regionConfigs,
			NumShards:     types.Int64Value(numShards),
			ZoneId:        types.StringValue(conversion.SafeValue(item.ZoneId)),
			ZoneName:      types.StringValue(conversion.SafeValue(item.ZoneName)),
		})
	}
	return &tfModels
}

func NewTagsObjType(ctx context.Context, input *[]admin.ResourceTag, diags *diag.Diagnostics) types.Set {
	if input == nil {
		// API Response not consistent, even when not set in POST/PATCH `[]` is returned instead of null
		return types.SetValueMust(TagsObjType, nil)
	}
	tfModels := make([]TFTagsModel, len(*input))
	for i, item := range *input {
		tfModels[i] = TFTagsModel{
			Key:   types.StringValue(item.Key),
			Value: types.StringValue(item.Value),
		}
	}
	setType, diagsLocal := types.SetValueFrom(ctx, TagsObjType, tfModels)
	diags.Append(diagsLocal...)
	return setType
}

func NewPrivateEndpointObjType(ctx context.Context, input *[]admin.ClusterDescriptionConnectionStringsPrivateEndpoint, diags *diag.Diagnostics) types.List {
	if input == nil {
		return types.ListNull(PrivateEndpointObjType)
	}
	tfModels := make([]TFPrivateEndpointModel, len(*input))
	for i, item := range *input {
		endpoints := NewEndpointsObjType(ctx, item.Endpoints, diags)
		tfModels[i] = TFPrivateEndpointModel{
			ConnectionString:                  types.StringValue(conversion.SafeValue(item.ConnectionString)),
			Endpoints:                         endpoints,
			SrvConnectionString:               types.StringValue(conversion.SafeValue(item.SrvConnectionString)),
			SrvShardOptimizedConnectionString: types.StringValue(conversion.SafeValue(item.SrvShardOptimizedConnectionString)),
			Type:                              types.StringValue(conversion.SafeValue(item.Type)),
		}
	}
	listType, diagsLocal := types.ListValueFrom(ctx, PrivateEndpointObjType, tfModels)
	diags.Append(diagsLocal...)
	return listType
}

func NewRegionConfigsObjType(ctx context.Context, input *[]admin.CloudRegionConfig20240805, diags *diag.Diagnostics) types.List {
	if input == nil {
		return types.ListNull(RegionConfigsObjType)
	}
	tfModels := make([]TFRegionConfigsModel, len(*input))
	for i, item := range *input {
		analyticsAutoScaling := NewAutoScalingObjType(ctx, item.AnalyticsAutoScaling, diags)
		analyticsSpecs := NewSpecsObjType(ctx, item.AnalyticsSpecs, diags)
		autoScaling := NewAutoScalingObjType(ctx, item.AutoScaling, diags)
		electableSpecs := NewSpecsFromHwObjType(ctx, item.ElectableSpecs, diags)
		readOnlySpecs := NewSpecsObjType(ctx, item.ReadOnlySpecs, diags)
		tfModels[i] = TFRegionConfigsModel{
			AnalyticsAutoScaling: analyticsAutoScaling,
			AnalyticsSpecs:       analyticsSpecs,
			AutoScaling:          autoScaling,
			BackingProviderName:  types.StringPointerValue(item.BackingProviderName),
			ElectableSpecs:       electableSpecs,
			Priority:             types.Int64PointerValue(conversion.IntPtrToInt64Ptr(item.Priority)),
			ProviderName:         types.StringValue(conversion.SafeValue(item.ProviderName)),
			ReadOnlySpecs:        readOnlySpecs,
			RegionName:           types.StringValue(conversion.SafeValue(item.RegionName)),
		}
	}
	listType, diagsLocal := types.ListValueFrom(ctx, RegionConfigsObjType, tfModels)
	diags.Append(diagsLocal...)
	return listType
}

func NewEndpointsObjType(ctx context.Context, input *[]admin.ClusterDescriptionConnectionStringsPrivateEndpointEndpoint, diags *diag.Diagnostics) types.List {
	if input == nil {
		return types.ListNull(EndpointsObjType)
	}
	tfModels := make([]TFEndpointsModel, len(*input))
	for i, item := range *input {
		tfModels[i] = TFEndpointsModel{
			EndpointId:   types.StringValue(conversion.SafeValue(item.EndpointId)),
			ProviderName: types.StringValue(conversion.SafeValue(item.ProviderName)),
			Region:       types.StringValue(conversion.SafeValue(item.Region)),
		}
	}
	listType, diagsLocal := types.ListValueFrom(ctx, EndpointsObjType, tfModels)
	diags.Append(diagsLocal...)
	return listType
}

func NewSpecsObjType(ctx context.Context, input *admin.DedicatedHardwareSpec20240805, diags *diag.Diagnostics) types.Object {
	if input == nil {
		return types.ObjectNull(SpecsObjType.AttrTypes)
	}
	tfModel := TFSpecsModel{
		DiskIops:      types.Int64PointerValue(conversion.IntPtrToInt64Ptr(input.DiskIOPS)),
		DiskSizeGb:    types.Float64PointerValue(input.DiskSizeGB),
		EbsVolumeType: types.StringValue(conversion.SafeValue(input.EbsVolumeType)),
		InstanceSize:  types.StringValue(conversion.SafeValue(input.InstanceSize)),
		NodeCount:     types.Int64PointerValue(conversion.IntPtrToInt64Ptr(input.NodeCount)),
	}
	objType, diagsLocal := types.ObjectValueFrom(ctx, SpecsObjType.AttrTypes, tfModel)
	diags.Append(diagsLocal...)
	return objType
}

func NewSpecsFromHwObjType(ctx context.Context, input *admin.HardwareSpec20240805, diags *diag.Diagnostics) types.Object {
	if input == nil {
		return types.ObjectNull(SpecsObjType.AttrTypes)
	}
	tfModel := TFSpecsModel{
		DiskIops:      types.Int64PointerValue(conversion.IntPtrToInt64Ptr(input.DiskIOPS)),
		DiskSizeGb:    types.Float64PointerValue(input.DiskSizeGB),
		EbsVolumeType: types.StringValue(conversion.SafeValue(input.EbsVolumeType)),
		InstanceSize:  types.StringValue(conversion.SafeValue(input.InstanceSize)),
		NodeCount:     types.Int64PointerValue(conversion.IntPtrToInt64Ptr(input.NodeCount)),
	}
	objType, diagsLocal := types.ObjectValueFrom(ctx, SpecsObjType.AttrTypes, tfModel)
	diags.Append(diagsLocal...)
	return objType
}

func NewAutoScalingObjType(ctx context.Context, input *admin.AdvancedAutoScalingSettings, diags *diag.Diagnostics) types.Object {
	if input == nil {
		return types.ObjectNull(AutoScalingObjType.AttrTypes)
	}
	compute := input.Compute
	tfModel := TFAutoScalingModel{}
	if compute != nil {
		tfModel.ComputeMaxInstanceSize = types.StringValue(conversion.SafeValue(compute.MaxInstanceSize))
		tfModel.ComputeMinInstanceSize = types.StringValue(conversion.SafeValue(compute.MinInstanceSize))
		tfModel.ComputeEnabled = types.BoolValue(conversion.SafeValue(compute.Enabled))
		tfModel.ComputeScaleDownEnabled = types.BoolValue(conversion.SafeValue(compute.ScaleDownEnabled))
	}
	diskGB := input.DiskGB
	if diskGB != nil {
		tfModel.DiskGBEnabled = types.BoolValue(conversion.SafeValue(diskGB.Enabled))
	}
	objType, diagsLocal := types.ObjectValueFrom(ctx, AutoScalingObjType.AttrTypes, tfModel)
	diags.Append(diagsLocal...)
	return objType
}
