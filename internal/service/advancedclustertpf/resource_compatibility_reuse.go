package advancedclustertpf

import (
	"context"
	"fmt"
	"strconv"

	"github.com/mongodb/terraform-provider-mongodbatlas/internal/common/constant"
	"github.com/mongodb/terraform-provider-mongodbatlas/internal/common/conversion"
	admin20240530 "go.mongodb.org/atlas-sdk/v20240530005/admin"
	"go.mongodb.org/atlas-sdk/v20241113004/admin"
)

type MajorVersionOperator int

const (
	EqualOrHigher MajorVersionOperator = iota
	Higher
	EqualOrLower
	Lower
)

func MajorVersionCompatible(input *string, version float64, operator MajorVersionOperator) *bool {
	if !conversion.IsStringPresent(input) {
		return nil
	}
	value, err := strconv.ParseFloat(*input, 64)
	if err != nil {
		return nil
	}
	var result bool
	switch operator {
	case EqualOrHigher:
		result = value >= version
	case Higher:
		result = value > version
	case EqualOrLower:
		result = value <= version
	case Lower:
		result = value < version
	default:
		return nil
	}
	return &result
}

func containerIDKey(providerName, regionName string) string {
	return fmt.Sprintf("%s:%s", providerName, regionName)
}

// based on flattenAdvancedReplicationSpecRegionConfigs in model_advanced_cluster.go
func resolveContainerIDs(ctx context.Context, projectID string, cluster *admin.ClusterDescription20240805, api admin.NetworkPeeringApi) (map[string]string, error) {
	containerIDs := map[string]string{}
	responseCache := map[string]*admin.PaginatedCloudProviderContainer{}
	for _, spec := range cluster.GetReplicationSpecs() {
		for _, regionConfig := range spec.GetRegionConfigs() {
			providerName := regionConfig.GetProviderName()
			if providerName == constant.TENANT {
				continue
			}
			params := &admin.ListPeeringContainerByCloudProviderApiParams{
				GroupId:      projectID,
				ProviderName: &providerName,
			}
			key := containerIDKey(providerName, regionConfig.GetRegionName())
			if _, ok := containerIDs[key]; ok {
				continue
			}
			var containersResponse *admin.PaginatedCloudProviderContainer
			var err error
			if response, ok := responseCache[providerName]; ok {
				containersResponse = response
			} else {
				containersResponse, _, err = api.ListPeeringContainerByCloudProviderWithParams(ctx, params).Execute()
				if err != nil {
					return nil, err
				}
				responseCache[providerName] = containersResponse
			}
			if results := GetAdvancedClusterContainerID(containersResponse.GetResults(), &regionConfig); results != "" {
				containerIDs[key] = results
			} else {
				return nil, fmt.Errorf("container id not found for %s", key)
			}
		}
	}
	return containerIDs, nil
}

func replicationSpecIDsFromOldAPI(clusterRespOld *admin20240530.AdvancedClusterDescription) map[string]string {
	specs := clusterRespOld.GetReplicationSpecs()
	zoneNameSpecIDs := make(map[string]string, len(specs))
	for _, spec := range specs {
		zoneNameSpecIDs[spec.GetZoneName()] = spec.GetId()
	}
	return zoneNameSpecIDs
}

// copied from advancedcluster/resource_update_logic.go
func populateIDValuesUsingNewAPI(ctx context.Context, projectID, clusterName string, connV2ClusterAPI admin.ClustersApi, replicationSpecs *[]admin.ReplicationSpec20240805) (*[]admin.ReplicationSpec20240805, error) {
	if replicationSpecs == nil || len(*replicationSpecs) == 0 {
		return replicationSpecs, nil
	}
	cluster, _, err := connV2ClusterAPI.GetCluster(ctx, projectID, clusterName).Execute()
	if err != nil {
		return nil, err
	}

	zoneToReplicationSpecsIDs := groupIDsByZone(cluster.GetReplicationSpecs())
	result := AddIDsToReplicationSpecs(*replicationSpecs, zoneToReplicationSpecsIDs)
	return &result, nil
}

// copied from advancedcluster/resource_update_logic.go
func groupIDsByZone(specs []admin.ReplicationSpec20240805) map[string][]string {
	result := make(map[string][]string)
	for _, spec := range specs {
		result[spec.GetZoneName()] = append(result[spec.GetZoneName()], spec.GetId())
	}
	return result
}
