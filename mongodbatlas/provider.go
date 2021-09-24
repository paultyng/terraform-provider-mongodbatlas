package mongodbatlas

import (
	"context"
	"encoding/base64"
	"fmt"
	"hash/crc32"
	"log"
	"reflect"
	"sort"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/spf13/cast"
	matlas "go.mongodb.org/atlas/mongodbatlas"
)

// Provider returns the provider to be use by the code.
func Provider() *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"public_key": {
				Type:     schema.TypeString,
				Required: true,
				DefaultFunc: schema.MultiEnvDefaultFunc([]string{
					"MONGODB_ATLAS_PUBLIC_KEY",
					"MCLI_PUBLIC_API_KEY",
				}, ""),
				Description: "MongoDB Atlas Programmatic Public Key",
			},
			"private_key": {
				Type:     schema.TypeString,
				Required: true,
				DefaultFunc: schema.MultiEnvDefaultFunc([]string{
					"MONGODB_ATLAS_PRIVATE_KEY",
					"MCLI_PRIVATE_API_KEY",
				}, ""),
				Description: "MongoDB Atlas Programmatic Private Key",
				Sensitive:   true,
			},
			"base_url": {
				Type:     schema.TypeString,
				Optional: true,
				DefaultFunc: schema.MultiEnvDefaultFunc([]string{
					"MONGODB_ATLAS_BASE_URL",
					"MCLI_OPS_MANAGER_URL",
				}, ""),
				Description: "MongoDB Atlas Base URL",
			},
			"realm_base_url": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("MONGODB_REALM_BASE_URL", ""),
				Description: "MongoDB Realm Base URL",
			},
		},

		DataSourcesMap: map[string]*schema.Resource{
			"mongodbatlas_custom_db_role":                        dataSourceMongoDBAtlasCustomDBRole(),
			"mongodbatlas_custom_db_roles":                       dataSourceMongoDBAtlasCustomDBRoles(),
			"mongodbatlas_database_user":                         dataSourceMongoDBAtlasDatabaseUser(),
			"mongodbatlas_database_users":                        dataSourceMongoDBAtlasDatabaseUsers(),
			"mongodbatlas_project":                               dataSourceMongoDBAtlasProject(),
			"mongodbatlas_projects":                              dataSourceMongoDBAtlasProjects(),
			"mongodbatlas_cluster":                               dataSourceMongoDBAtlasCluster(),
			"mongodbatlas_clusters":                              dataSourceMongoDBAtlasClusters(),
			"mongodbatlas_cloud_provider_snapshot":               dataSourceMongoDBAtlasCloudProviderSnapshot(),
			"mongodbatlas_cloud_provider_snapshots":              dataSourceMongoDBAtlasCloudProviderSnapshots(),
			"mongodbatlas_network_container":                     dataSourceMongoDBAtlasNetworkContainer(),
			"mongodbatlas_network_containers":                    dataSourceMongoDBAtlasNetworkContainers(),
			"mongodbatlas_network_peering":                       dataSourceMongoDBAtlasNetworkPeering(),
			"mongodbatlas_network_peerings":                      dataSourceMongoDBAtlasNetworkPeerings(),
			"mongodbatlas_cloud_provider_snapshot_restore_job":   dataSourceMongoDBAtlasCloudProviderSnapshotRestoreJob(),
			"mongodbatlas_cloud_provider_snapshot_restore_jobs":  dataSourceMongoDBAtlasCloudProviderSnapshotRestoreJobs(),
			"mongodbatlas_maintenance_window":                    dataSourceMongoDBAtlasMaintenanceWindow(),
			"mongodbatlas_auditing":                              dataSourceMongoDBAtlasAuditing(),
			"mongodbatlas_team":                                  dataSourceMongoDBAtlasTeam(),
			"mongodbatlas_teams":                                 dataSourceMongoDBAtlasTeam(),
			"mongodbatlas_global_cluster_config":                 dataSourceMongoDBAtlasGlobalCluster(),
			"mongodbatlas_alert_configuration":                   dataSourceMongoDBAtlasAlertConfiguration(),
			"mongodbatlas_x509_authentication_database_user":     dataSourceMongoDBAtlasX509AuthDBUser(),
			"mongodbatlas_privatelink_endpoint":                  dataSourceMongoDBAtlasPrivateLinkEndpoint(),
			"mongodbatlas_privatelink_endpoint_service":          dataSourceMongoDBAtlasPrivateEndpointServiceLink(),
			"mongodbatlas_cloud_provider_snapshot_backup_policy": dataSourceMongoDBAtlasCloudProviderSnapshotBackupPolicy(),
			"mongodbatlas_cloud_backup_schedule":                 dataSourceMongoDBAtlasCloudBackupSchedule(),
			"mongodbatlas_third_party_integrations":              dataSourceMongoDBAtlasThirdPartyIntegrations(),
			"mongodbatlas_third_party_integration":               dataSourceMongoDBAtlasThirdPartyIntegration(),
			"mongodbatlas_project_ip_access_list":                dataSourceMongoDBAtlasProjectIPAccessList(),
			"mongodbatlas_cloud_provider_access":                 dataSourceMongoDBAtlasCloudProviderAccessList(),
			"mongodbatlas_cloud_provider_access_setup":           dataSourceMongoDBAtlasCloudProviderAccessSetup(),
			"mongodbatlas_custom_dns_configuration_cluster_aws":  dataSourceMongoDBAtlasCustomDNSConfigurationAWS(),
			"mongodbatlas_online_archive":                        dataSourceMongoDBAtlasOnlineArchive(),
			"mongodbatlas_online_archives":                       dataSourceMongoDBAtlasOnlineArchives(),
			"mongodbatlas_ldap_configuration":                    dataSourceMongoDBAtlasLDAPConfiguration(),
			"mongodbatlas_ldap_verify":                           dataSourceMongoDBAtlasLDAPVerify(),
			"mongodbatlas_search_index":                          dataSourceMongoDBAtlasSearchIndex(),
			"mongodbatlas_search_indexes":                        dataSourceMongoDBAtlasSearchIndexes(),
			"mongodbatlas_data_lake":                             dataSourceMongoDBAtlasDataLake(),
			"mongodbatlas_data_lakes":                            dataSourceMongoDBAtlasDataLakes(),
			"mongodbatlas_event_trigger":                         dataSourceMongoDBAtlasEventTrigger(),
			"mongodbatlas_event_triggers":                        dataSourceMongoDBAtlasEventTriggers(),
			"mongodbatlas_project_invitation":                    dataSourceMongoDBAtlasProjectInvitation(),
			"mongodbatlas_org_invitation":                        dataSourceMongoDBAtlasOrgInvitation(),
		},

		ResourcesMap: map[string]*schema.Resource{
			"mongodbatlas_custom_db_role":                        resourceMongoDBAtlasCustomDBRole(),
			"mongodbatlas_database_user":                         resourceMongoDBAtlasDatabaseUser(),
			"mongodbatlas_project":                               resourceMongoDBAtlasProject(),
			"mongodbatlas_cluster":                               resourceMongoDBAtlasCluster(),
			"mongodbatlas_cloud_provider_snapshot":               resourceMongoDBAtlasCloudProviderSnapshot(),
			"mongodbatlas_network_container":                     resourceMongoDBAtlasNetworkContainer(),
			"mongodbatlas_cloud_provider_snapshot_restore_job":   resourceMongoDBAtlasCloudProviderSnapshotRestoreJob(),
			"mongodbatlas_network_peering":                       resourceMongoDBAtlasNetworkPeering(),
			"mongodbatlas_encryption_at_rest":                    resourceMongoDBAtlasEncryptionAtRest(),
			"mongodbatlas_private_ip_mode":                       resourceMongoDBAtlasPrivateIPMode(),
			"mongodbatlas_maintenance_window":                    resourceMongoDBAtlasMaintenanceWindow(),
			"mongodbatlas_auditing":                              resourceMongoDBAtlasAuditing(),
			"mongodbatlas_team":                                  resourceMongoDBAtlasTeam(),
			"mongodbatlas_teams":                                 resourceMongoDBAtlasTeam(),
			"mongodbatlas_global_cluster_config":                 resourceMongoDBAtlasGlobalCluster(),
			"mongodbatlas_alert_configuration":                   resourceMongoDBAtlasAlertConfiguration(),
			"mongodbatlas_x509_authentication_database_user":     resourceMongoDBAtlasX509AuthDBUser(),
			"mongodbatlas_privatelink_endpoint":                  resourceMongoDBAtlasPrivateLinkEndpoint(),
			"mongodbatlas_privatelink_endpoint_service":          resourceMongoDBAtlasPrivateEndpointServiceLink(),
			"mongodbatlas_cloud_provider_snapshot_backup_policy": resourceMongoDBAtlasCloudProviderSnapshotBackupPolicy(),
			"mongodbatlas_third_party_integration":               resourceMongoDBAtlasThirdPartyIntegration(),
			"mongodbatlas_project_ip_access_list":                resourceMongoDBAtlasProjectIPAccessList(),
			"mongodbatlas_cloud_provider_access":                 resourceMongoDBAtlasCloudProviderAccess(),
			"mongodbatlas_online_archive":                        resourceMongoDBAtlasOnlineArchive(),
			"mongodbatlas_custom_dns_configuration_cluster_aws":  resourceMongoDBAtlasCustomDNSConfiguration(),
			"mongodbatlas_ldap_configuration":                    resourceMongoDBAtlasLDAPConfiguration(),
			"mongodbatlas_ldap_verify":                           resourceMongoDBAtlasLDAPVerify(),
			"mongodbatlas_cloud_provider_access_setup":           resourceMongoDBAtlasCloudProviderAccessSetup(),
			"mongodbatlas_cloud_provider_access_authorization":   resourceMongoDBAtlasCloudProviderAccessAuthorization(),
			"mongodbatlas_search_index":                          resourceMongoDBAtlasSearchIndex(),
			"mongodbatlas_data_lake":                             resourceMongoDBAtlasDataLake(),
			"mongodbatlas_event_trigger":                         resourceMongoDBAtlasEventTriggers(),
			"mongodbatlas_cloud_backup_schedule":                 resourceMongoDBAtlasCloudBackupSchedule(),
			"mongodbatlas_project_invitation":                    resourceMongoDBAtlasProjectInvitation(),
			"mongodbatlas_org_invitation":                        resourceMongoDBAtlasOrgInvitation(),
		},

		ConfigureContextFunc: providerConfigure,
	}
}

func providerConfigure(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
	config := Config{
		PublicKey:    d.Get("public_key").(string),
		PrivateKey:   d.Get("private_key").(string),
		BaseURL:      d.Get("base_url").(string),
		RealmBaseURL: d.Get("realm_base_url").(string),
	}

	return config.NewClient(ctx)
}

func encodeStateID(values map[string]string) string {
	encode := func(e string) string { return base64.StdEncoding.EncodeToString([]byte(e)) }
	encodedValues := make([]string, 0)

	// sort to make sure the same encoding is returned in case of same input
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}

	sort.Strings(keys)

	for _, key := range keys {
		encodedValues = append(encodedValues, fmt.Sprintf("%s:%s", encode(key), encode(values[key])))
	}

	return strings.Join(encodedValues, "-")
}

func decodeStateID(stateID string) map[string]string {
	decode := func(d string) string {
		decodedString, err := base64.StdEncoding.DecodeString(d)
		if err != nil {
			log.Printf("[WARN] error decoding state ID: %s", err)
		}

		return string(decodedString)
	}
	decodedValues := make(map[string]string)
	encodedValues := strings.Split(stateID, "-")

	for _, value := range encodedValues {
		keyValue := strings.Split(value, ":")
		decodedValues[decode(keyValue[0])] = decode(keyValue[1])
	}

	return decodedValues
}

func valRegion(reg interface{}, opt ...string) (string, error) {
	region, err := cast.ToStringE(reg)
	if err != nil {
		return "", err
	}

	if region == "" {
		return "", fmt.Errorf("region must be set")
	}

	/*
		We need to check if the option will be similar to network_peering word
		 (this comes in from the same resource) because network_pering resource
		 has not the standard region name pattern "US_EAST_1",
		 instead it needs the following one: "us-east-1".
	*/
	if len(opt) > 0 && strings.EqualFold("network_peering", opt[0]) {
		return strings.ToLower(strings.ReplaceAll(region, "_", "-")), nil
	}

	return strings.ReplaceAll(region, "-", "_"), nil
}

func flattenLabels(l []matlas.Label) []map[string]interface{} {
	labels := make([]map[string]interface{}, len(l))
	for i, v := range l {
		labels[i] = map[string]interface{}{
			"key":   v.Key,
			"value": v.Value,
		}
	}

	return labels
}

func expandLabelSliceFromSetSchema(d *schema.ResourceData) []matlas.Label {
	list := d.Get("labels").(*schema.Set)
	res := make([]matlas.Label, list.Len())

	for i, val := range list.List() {
		v := val.(map[string]interface{})
		res[i] = matlas.Label{
			Key:   v["key"].(string),
			Value: v["value"].(string),
		}
	}

	return res
}

func containsLabelOrKey(list []matlas.Label, item matlas.Label) bool {
	for _, v := range list {
		if reflect.DeepEqual(v, item) || v.Key == item.Key {
			return true
		}
	}

	return false
}

func removeLabel(list []matlas.Label, item matlas.Label) []matlas.Label {
	var pos int

	for _, v := range list {
		if reflect.DeepEqual(v, item) {
			list = append(list[:pos], list[pos+1:]...)

			if pos > 0 {
				pos--
			}

			continue
		}
		pos++
	}

	return list
}

func expandStringList(list []interface{}) (res []string) {
	for _, v := range list {
		res = append(res, v.(string))
	}

	return
}

func getEncodedID(stateID, keyPosition string) string {
	id := ""
	if !hasMultipleValues(stateID) {
		return stateID
	}

	decoded := decodeStateID(stateID)
	id = decoded[keyPosition]

	return id
}

func hasMultipleValues(value string) bool {
	if strings.Contains(value, "-") && strings.Contains(value, ":") {
		return true
	}

	return false
}

// HashCodeString hashes a string to a unique hashcode.
//
// crc32 returns a uint32, but for our use we need
// and non negative integer. Here we cast to an integer
// and invert it if the result is negative.
func HashCodeString(s string) int {
	v := int(crc32.ChecksumIEEE([]byte(s)))
	if v >= 0 {
		return v
	}
	if -v >= 0 {
		return -v
	}
	// v == MinInt
	return 0
}
