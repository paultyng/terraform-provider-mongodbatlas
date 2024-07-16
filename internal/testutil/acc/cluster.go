package acc

import (
	"fmt"
	"os"
	"testing"

	"github.com/mongodb/terraform-provider-mongodbatlas/internal/common/constant"
	"github.com/mongodb/terraform-provider-mongodbatlas/internal/common/conversion"
	"go.mongodb.org/atlas-sdk/v20240530002/admin"
)

type ClusterRequest struct {
	Tags                   map[string]string
	ProjectID              string
	ResourceSuffix         string
	AdvancedConfiguration  map[string]any
	ResourceDependencyName string
	ClusterName            string
	MongoDBMajorVersion    string
	ReplicationSpecs       []ReplicationSpecRequest
	DiskSizeGb             int
	CloudBackup            bool
	Geosharded             bool
	RetainBackupsEnabled   bool
	PitEnabled             bool
}

func (r *ClusterRequest) AddDefaults() {
	if r.ResourceSuffix == "" {
		r.ResourceSuffix = defaultClusterResourceSuffix
	}
	if len(r.ReplicationSpecs) == 0 {
		r.ReplicationSpecs = []ReplicationSpecRequest{{}}
	}
	if r.ClusterName == "" {
		r.ClusterName = RandomClusterName()
	}
}

func (r *ClusterRequest) ClusterType() string {
	if r.Geosharded {
		return "GEOSHARDED"
	}
	return "REPLICASET"
}

type ClusterInfo struct {
	ProjectIDStr     string
	ProjectID        string
	Name             string
	ResourceName     string
	TerraformNameRef string
	TerraformStr     string
}

const defaultClusterResourceSuffix = "cluster_info"

// GetClusterInfo is used to obtain a project and cluster configuration resource.
// When `MONGODB_ATLAS_CLUSTER_NAME` and `MONGODB_ATLAS_PROJECT_ID` are defined, creation of resources is avoided. This is useful for local execution but not intended for CI executions.
// Clusters will be created in project ProjectIDExecution.
func GetClusterInfo(tb testing.TB, req *ClusterRequest) ClusterInfo {
	tb.Helper()
	if req == nil {
		req = new(ClusterRequest)
	}
	hclCreator := ClusterResourceHcl
	if req.ProjectID == "" {
		if ExistingClusterUsed() {
			projectID, clusterName := existingProjectIDClusterName()
			req.ProjectID = projectID
			req.ClusterName = clusterName
			hclCreator = ClusterDatasourceHcl
		} else {
			req.ProjectID = ProjectIDExecution(tb)
		}
	}
	clusterTerraformStr, clusterName, clusterResourceName, err := hclCreator(req)
	if err != nil {
		tb.Error(err)
	}
	return ClusterInfo{
		ProjectIDStr:     fmt.Sprintf("%q", req.ProjectID),
		ProjectID:        req.ProjectID,
		Name:             clusterName,
		TerraformNameRef: fmt.Sprintf("%s.name", clusterResourceName),
		ResourceName:     clusterResourceName,
		TerraformStr:     clusterTerraformStr,
	}
}

func ExistingClusterUsed() bool {
	projectID, clusterName := existingProjectIDClusterName()
	return clusterName != "" && projectID != ""
}

func existingProjectIDClusterName() (projectID, clusterName string) {
	return os.Getenv("MONGODB_ATLAS_PROJECT_ID"), os.Getenv("MONGODB_ATLAS_CLUSTER_NAME")
}

type ReplicationSpecRequest struct {
	ZoneName                 string
	Region                   string
	InstanceSize             string
	ProviderName             string
	EbsVolumeType            string
	ExtraRegionConfigs       []ReplicationSpecRequest
	NodeCount                int
	NodeCountReadOnly        int
	Priority                 int
	AutoScalingDiskGbEnabled bool
}

func (r *ReplicationSpecRequest) AddDefaults() {
	if r.Priority == 0 {
		r.Priority = 7
	}
	if r.NodeCount == 0 {
		r.NodeCount = 3
	}
	if r.ZoneName == "" {
		r.ZoneName = "Zone 1"
	}
	if r.Region == "" {
		r.Region = "US_WEST_2"
	}
	if r.InstanceSize == "" {
		r.InstanceSize = "M10"
	}
	if r.ProviderName == "" {
		r.ProviderName = constant.AWS
	}
}

func (r *ReplicationSpecRequest) AllRegionConfigs() []admin.CloudRegionConfig {
	config := CloudRegionConfig(*r)
	configs := []admin.CloudRegionConfig{config}
	for i := range r.ExtraRegionConfigs {
		extra := r.ExtraRegionConfigs[i]
		configs = append(configs, CloudRegionConfig(extra))
	}
	return configs
}

func ReplicationSpec(req *ReplicationSpecRequest) admin.ReplicationSpec {
	if req == nil {
		req = new(ReplicationSpecRequest)
	}
	req.AddDefaults()
	defaultNumShards := 1
	regionConfigs := req.AllRegionConfigs()
	return admin.ReplicationSpec{
		NumShards:     &defaultNumShards,
		ZoneName:      &req.ZoneName,
		RegionConfigs: &regionConfigs,
	}
}

func CloudRegionConfig(req ReplicationSpecRequest) admin.CloudRegionConfig {
	req.AddDefaults()
	var readOnly admin.DedicatedHardwareSpec
	if req.NodeCountReadOnly != 0 {
		readOnly = admin.DedicatedHardwareSpec{
			NodeCount:    &req.NodeCountReadOnly,
			InstanceSize: &req.InstanceSize,
		}
	}
	return admin.CloudRegionConfig{
		RegionName:   &req.Region,
		Priority:     &req.Priority,
		ProviderName: &req.ProviderName,
		ElectableSpecs: &admin.HardwareSpec{
			InstanceSize:  &req.InstanceSize,
			NodeCount:     &req.NodeCount,
			EbsVolumeType: conversion.StringPtr(req.EbsVolumeType),
		},
		ReadOnlySpecs: &readOnly,
		AutoScaling: &admin.AdvancedAutoScalingSettings{
			DiskGB: &admin.DiskGBAutoScaling{Enabled: &req.AutoScalingDiskGbEnabled},
		},
	}
}
