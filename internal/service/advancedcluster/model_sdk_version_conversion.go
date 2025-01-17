package advancedcluster

import (
	admin20240530 "go.mongodb.org/atlas-sdk/v20240530005/admin"
	"go.mongodb.org/atlas-sdk/v20241113004/admin"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
)

// Conversions from one SDK model version to another are used to avoid duplicating our flatten/expand conversion functions.
// - These functions must not contain any business logic.
// - All will be removed once we rely on a single API version.

func convertTagsPtrToOldSDK(tags *[]admin.ResourceTag) *[]admin20240530.ResourceTag {
	if tags == nil {
		return nil
	}
	tagsSlice := *tags
	results := make([]admin20240530.ResourceTag, len(tagsSlice))
	for i := range len(tagsSlice) {
		tag := tagsSlice[i]
		results[i] = admin20240530.ResourceTag{
			Key:   tag.Key,
			Value: tag.Value,
		}
	}
	return &results
}

func convertBiConnectToOldSDK(biconnector *admin.BiConnector) *admin20240530.BiConnector {
	if biconnector == nil {
		return nil
	}
	return &admin20240530.BiConnector{
		Enabled:        biconnector.Enabled,
		ReadPreference: biconnector.ReadPreference,
	}
}

func convertLabelSliceToOldSDK(slice []admin.ComponentLabel, err diag.Diagnostics) ([]admin20240530.ComponentLabel, diag.Diagnostics) {
	if err != nil {
		return nil, err
	}
	results := make([]admin20240530.ComponentLabel, len(slice))
	for i := range len(slice) {
		label := slice[i]
		results[i] = admin20240530.ComponentLabel{
			Key:   label.Key,
			Value: label.Value,
		}
	}
	return results, nil
}
