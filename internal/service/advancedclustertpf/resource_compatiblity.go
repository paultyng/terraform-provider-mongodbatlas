package advancedclustertpf

import (
	"context"
	"fmt"
	"reflect"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/mongodb/terraform-provider-mongodbatlas/internal/common/conversion"
	"github.com/mongodb/terraform-provider-mongodbatlas/internal/config"
	admin20240530 "go.mongodb.org/atlas-sdk/v20240530005/admin"
	"go.mongodb.org/atlas-sdk/v20241113004/admin"
)

func overrideAttributesWithPrevStateValue(modelIn, modelOut *TFModel) {
	beforeVersion := conversion.NilForUnknown(modelIn.MongoDBMajorVersion, modelIn.MongoDBMajorVersion.ValueStringPointer())
	if beforeVersion != nil && !modelIn.MongoDBMajorVersion.Equal(modelOut.MongoDBMajorVersion) {
		modelOut.MongoDBMajorVersion = types.StringPointerValue(beforeVersion)
	}
	retainBackups := conversion.NilForUnknown(modelIn.RetainBackupsEnabled, modelIn.RetainBackupsEnabled.ValueBoolPointer())
	if retainBackups != nil && !modelIn.RetainBackupsEnabled.Equal(modelOut.RetainBackupsEnabled) {
		modelOut.RetainBackupsEnabled = types.BoolPointerValue(retainBackups)
	}
}

func findNumShardsUpdates(ctx context.Context, state, plan *TFModel, diags *diag.Diagnostics) map[string]int64 {
	if !usingLegacyShardingConfig(ctx, plan.ReplicationSpecs, diags) {
		return nil
	}
	stateCounts := numShardsMap(ctx, state.ReplicationSpecs, diags)
	planCounts := numShardsMap(ctx, plan.ReplicationSpecs, diags)
	if diags.HasError() {
		return nil
	}
	if reflect.DeepEqual(stateCounts, planCounts) {
		return nil
	}
	return planCounts
}

func resolveAPIInfo(ctx context.Context, diags *diag.Diagnostics, client *config.MongoDBClient, plan *TFModel, clusterLatest *admin.ClusterDescription20240805, forceLegacySchema bool) *ExtraAPIInfo {
	var (
		api20240530             = client.AtlasV220240530.ClustersApi
		projectID               = plan.ProjectID.ValueString()
		clusterName             = plan.Name.ValueString()
		forceLegacySchemaFailed = false
	)
	clusterRespOld, _, err := api20240530.GetCluster(ctx, projectID, clusterName).Execute()
	if err != nil {
		if admin20240530.IsErrorCode(err, "ASYMMETRIC_SHARD_UNSUPPORTED") {
			forceLegacySchemaFailed = forceLegacySchema
		} else {
			diags.AddError(errorReadLegacy20240530, defaultAPIErrorDetails(clusterName, err))
			return nil
		}
	}
	containerIDs, err := resolveContainerIDs(ctx, projectID, clusterLatest, client.AtlasV2.NetworkPeeringApi)
	if err != nil {
		diags.AddError(errorResolveContainerIDs, fmt.Sprintf("cluster name = %s, error details: %s", clusterName, err.Error()))
		return nil
	}
	return &ExtraAPIInfo{
		ContainerIDs:               containerIDs,
		ZoneNameReplicationSpecIDs: replicationSpecIDsFromOldAPI(clusterRespOld),
		ForceLegacySchemaFailed:    forceLegacySchemaFailed,
		ZoneNameNumShards:          numShardsMapFromOldAPI(clusterRespOld),
		UsingLegacySchema:          forceLegacySchema || usingLegacyShardingConfig(ctx, plan.ReplicationSpecs, diags),
	}
}

// instead of using `num_shards` explode the replication specs, and set disk_size_gb
func normalizeFromTFModel(ctx context.Context, model *TFModel, diags *diag.Diagnostics, shoudlExplodeNumShards bool) *admin.ClusterDescription20240805 {
	latestModel := NewAtlasReq(ctx, model, diags)
	if diags.HasError() {
		return nil
	}
	counts := numShardsCounts(ctx, model.ReplicationSpecs, diags)
	if diags.HasError() {
		return nil
	}
	usingLegacySchema := isNumShardsGreaterThanOne(counts)
	if usingLegacySchema && shoudlExplodeNumShards {
		explodeNumShards(latestModel, counts)
	}
	diskSize := normalizeDiskSize(model, latestModel, diags)
	if diags.HasError() {
		return nil
	}
	if diskSize != nil {
		setDiskSize(latestModel, diskSize)
	}
	return latestModel
}

func normalizeDiskSize(model *TFModel, latestModel *admin.ClusterDescription20240805, diags *diag.Diagnostics) *float64 {
	rootDiskSize := conversion.NilForUnknown(model.DiskSizeGB, model.DiskSizeGB.ValueFloat64Pointer())
	regionRootDiskSize := findRegionRootDiskSize(latestModel.ReplicationSpecs)
	if rootDiskSize != nil && regionRootDiskSize != nil && (*regionRootDiskSize-*rootDiskSize) > 0.01 {
		errMsg := "disk_size_gb @ root != disk_size_gb @ region (%.2f!=%.2f)"
		diags.AddError(errMsg, errMsg)
		return nil
	}
	// Prefer regionRootDiskSize over rootDiskSize
	if regionRootDiskSize != nil {
		return regionRootDiskSize
	}
	return rootDiskSize
}

func explodeNumShards(req *admin.ClusterDescription20240805, counts []int64) {
	specs := req.GetReplicationSpecs()
	newSpecs := []admin.ReplicationSpec20240805{}
	for i, spec := range specs {
		newSpecs = append(newSpecs, spec)
		for range counts[i] - 1 {
			newSpecs = append(newSpecs, *repSpecNoIDs(spec))
		}
	}
	req.ReplicationSpecs = &newSpecs
}

func repSpecNoIDs(repspec admin.ReplicationSpec20240805) *admin.ReplicationSpec20240805 {
	repspec.Id = nil
	repspec.ZoneId = nil
	return &repspec
}

func numShardsCounts(ctx context.Context, input types.List, diags *diag.Diagnostics) []int64 {
	elements := make([]TFReplicationSpecsModel, len(input.Elements()))
	if len(elements) == 0 {
		return nil
	}
	if localDiags := input.ElementsAs(ctx, &elements, false); len(localDiags) > 0 {
		diags.Append(localDiags...)
		return nil
	}
	counts := make([]int64, len(elements))
	for i := range elements {
		item := &elements[i]
		counts[i] = item.NumShards.ValueInt64()
	}
	return counts
}

func usingLegacyShardingConfig(ctx context.Context, input types.List, diags *diag.Diagnostics) bool {
	counts := numShardsCounts(ctx, input, diags)
	if diags.HasError() {
		return false
	}
	return isNumShardsGreaterThanOne(counts)
}

func numShardsMap(ctx context.Context, input types.List, diags *diag.Diagnostics) map[string]int64 {
	elements := make([]TFReplicationSpecsModel, len(input.Elements()))
	if len(elements) == 0 {
		return nil
	}
	if localDiags := input.ElementsAs(ctx, &elements, false); len(localDiags) > 0 {
		diags.Append(localDiags...)
		return nil
	}
	counts := map[string]int64{}
	for i := range elements {
		e := elements[i]
		zoneName := resolveZoneNameOrUseDefault(&e)
		counts[zoneName] = e.NumShards.ValueInt64()
	}
	return counts
}

func numShardsMapFromOldAPI(clusterRespOld *admin20240530.AdvancedClusterDescription) map[string]int64 {
	ret := make(map[string]int64)
	for i := range clusterRespOld.GetReplicationSpecs() {
		spec := &clusterRespOld.GetReplicationSpecs()[i]
		ret[spec.GetZoneName()] = int64(spec.GetNumShards())
	}
	return ret
}

func isNumShardsGreaterThanOne(counts []int64) bool {
	for _, count := range counts {
		if count > 1 {
			return true
		}
	}
	return false
}

func setDiskSize(req *admin.ClusterDescription20240805, size *float64) {
	for i, spec := range req.GetReplicationSpecs() {
		for j := range spec.GetRegionConfigs() {
			actualConfig := req.GetReplicationSpecs()[i].GetRegionConfigs()[j]
			analyticsSpecs := actualConfig.AnalyticsSpecs
			if analyticsSpecs != nil {
				analyticsSpecs.DiskSizeGB = size
			}
			electable := actualConfig.ElectableSpecs
			if electable != nil {
				electable.DiskSizeGB = size
			}
			readonly := actualConfig.ReadOnlySpecs
			if readonly != nil {
				readonly.DiskSizeGB = size
			}
		}
	}
}

func findRegionRootDiskSize(specs *[]admin.ReplicationSpec20240805) *float64 {
	if specs == nil {
		return nil
	}
	for _, spec := range *specs {
		for _, regionConfig := range spec.GetRegionConfigs() {
			analyticsSpecs := regionConfig.AnalyticsSpecs
			if analyticsSpecs != nil && analyticsSpecs.DiskSizeGB != nil {
				return analyticsSpecs.DiskSizeGB
			}
			electable := regionConfig.ElectableSpecs
			if electable != nil && electable.DiskSizeGB != nil {
				return electable.DiskSizeGB
			}
			readonly := regionConfig.ReadOnlySpecs
			if readonly != nil && readonly.DiskSizeGB != nil {
				return readonly.DiskSizeGB
			}
		}
	}
	return nil
}

func externalIDToLegacyID(ctx context.Context, input types.List, diags *diag.Diagnostics) map[string]string {
	elements := make([]TFReplicationSpecsModel, len(input.Elements()))
	if localDiags := input.ElementsAs(ctx, &elements, false); len(localDiags) > 0 {
		diags.Append(localDiags...)
		return nil
	}
	idsMapped := map[string]string{}
	for i := range elements {
		e := elements[i]
		externalID := e.ExternalId.ValueString()
		legacyID := e.Id.ValueString()
		if externalID != "" && legacyID != "" {
			idsMapped[externalID] = legacyID
		}
	}
	return idsMapped
}
