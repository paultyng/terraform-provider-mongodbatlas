package advancedclustertpf

import (
	"github.com/mongodb/terraform-provider-mongodbatlas/internal/common/conversion"
	admin20240530 "go.mongodb.org/atlas-sdk/v20240530005/admin"
	"go.mongodb.org/atlas-sdk/v20241113004/admin"
)

func newLegacyModel20240530ReplicationSpecsAndDiskGBOnly(specs *[]admin.ReplicationSpec20240805, zoneNameNumShards map[string]int64, oldDiskGB *float64, externalIDToLegacyID map[string]string) *admin20240530.AdvancedClusterDescription {
	newDiskGB := findRegionRootDiskSize(specs)
	if oldDiskGB != nil && newDiskGB != nil && (*newDiskGB-*oldDiskGB) < 0.01 {
		newDiskGB = nil
	}
	return &admin20240530.AdvancedClusterDescription{
		DiskSizeGB:       newDiskGB,
		ReplicationSpecs: convertReplicationSpecs20240805to20240530(specs, zoneNameNumShards, externalIDToLegacyID),
	}
}

func convertReplicationSpecs20240805to20240530(replicationSpecs *[]admin.ReplicationSpec20240805, zoneNameNumShards map[string]int64, externalIDToLegacyID map[string]string) *[]admin20240530.ReplicationSpec {
	if replicationSpecs == nil {
		return nil
	}
	result := make([]admin20240530.ReplicationSpec, len(*replicationSpecs))
	for i, replicationSpec := range *replicationSpecs {
		numShards, ok := zoneNameNumShards[replicationSpec.GetZoneName()]
		if !ok {
			numShards = 1
		}
		legacyID := externalIDToLegacyID[replicationSpec.GetId()]
		result[i] = admin20240530.ReplicationSpec{
			NumShards:     conversion.Int64PtrToIntPtr(&numShards),
			Id:            conversion.StringPtr(legacyID),
			ZoneName:      replicationSpec.ZoneName,
			RegionConfigs: ConvertRegionConfigSlice20241023to20240530(replicationSpec.RegionConfigs),
		}
	}
	return &result
}
