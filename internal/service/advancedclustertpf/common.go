package advancedclustertpf

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/mongodb/terraform-provider-mongodbatlas/internal/common/constant"
	"github.com/mongodb/terraform-provider-mongodbatlas/internal/common/conversion"
	"github.com/spf13/cast"
	"go.mongodb.org/atlas-sdk/v20241113004/admin"
)

const (
	IgnoreLabelKey = "Infrastructure Tool"
)

var (
	ErrIgnoreLabel = fmt.Errorf("you should not set `%s` label, it is used for internal purposes", IgnoreLabelKey)
)

func FormatMongoDBMajorVersion(version string) string {
	if strings.Contains(version, ".") {
		return version
	}
	return fmt.Sprintf("%.1f", cast.ToFloat32(version))
}

func AddIDsToReplicationSpecs(replicationSpecs []admin.ReplicationSpec20240805, zoneToReplicationSpecsIDs map[string][]string) []admin.ReplicationSpec20240805 {
	for zoneName, availableIDs := range zoneToReplicationSpecsIDs {
		indexOfIDToUse := 0
		for i := range replicationSpecs {
			if indexOfIDToUse >= len(availableIDs) {
				break // all available ids for this zone have been used
			}
			if replicationSpecs[i].GetZoneName() == zoneName {
				newID := availableIDs[indexOfIDToUse]
				indexOfIDToUse++
				replicationSpecs[i].Id = &newID
			}
		}
	}
	return replicationSpecs
}

func GetAdvancedClusterContainerID(containers []admin.CloudProviderContainer, cluster *admin.CloudRegionConfig20240805) string {
	for i, container := range containers {
		gpc := cluster.GetProviderName() == constant.GCP
		azure := container.GetProviderName() == cluster.GetProviderName() && container.GetRegion() == cluster.GetRegionName()
		aws := container.GetRegionName() == cluster.GetRegionName()
		if gpc || azure || aws {
			return containers[i].GetId()
		}
	}
	return ""
}

func PinFCV(ctx context.Context, api admin.ClustersApi, projectID, clusterName, expirationDateStr string) error {
	expirationTime, ok := conversion.StringToTime(expirationDateStr)
	if !ok {
		return fmt.Errorf("expiration_date format is incorrect: %s", expirationDateStr)
	}
	req := admin.PinFCV{
		ExpirationDate: &expirationTime,
	}
	if _, _, err := api.PinFeatureCompatibilityVersion(ctx, projectID, clusterName, &req).Execute(); err != nil {
		return err
	}
	return nil
}

func GenerateFCVPinningWarningForRead(fcvPresentInState bool, apiRespFCVExpirationDate *time.Time) []diag.Diagnostic {
	pinIsActive := apiRespFCVExpirationDate != nil
	if fcvPresentInState && !pinIsActive { // pin is not active but present in state (and potentially in config file)
		warning := diag.NewWarningDiagnostic(
			"FCV pin is no longer active",
			"Please remove `pinned_fcv` from the configuration and apply changes to avoid re-pinning the FCV. Warning can be ignored if `pinned_fcv` block has been removed from the configuration.")
		return []diag.Diagnostic{warning}
	}
	if fcvPresentInState && pinIsActive {
		if time.Now().After(*apiRespFCVExpirationDate) { // pin is active, present in state, but its expiration date has passed
			warning := diag.NewWarningDiagnostic(
				"FCV pin expiration date has expired",
				"During the next maintenance window FCV will be unpinned. FCV expiration date can be extended, or `pinned_fcv` block can be removed to trigger the unpin immediately.")
			return []diag.Diagnostic{warning}
		}
	}
	return nil
}
