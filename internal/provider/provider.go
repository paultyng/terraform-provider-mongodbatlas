package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"regexp"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/credentials/stscreds"
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/mongodb/terraform-provider-mongodbatlas/internal/config"
	"github.com/mongodb/terraform-provider-mongodbatlas/mongodbatlas"
	"github.com/mongodb/terraform-provider-mongodbatlas/mongodbatlas/util"
	"github.com/mwielbut/pointy"
)

var (
	ProviderEnableBeta, _ = strconv.ParseBool(os.Getenv("MONGODB_ATLAS_ENABLE_BETA"))
)

type SecretData struct {
	PublicKey  string `json:"public_key"`
	PrivateKey string `json:"private_key"`
}

// NewSdkV2Provider returns the provider to be use by the code.
func NewSdkV2Provider() *schema.Provider {
	provider := &schema.Provider{
		Schema: map[string]*schema.Schema{
			"public_key": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "MongoDB Atlas Programmatic Public Key",
			},
			"private_key": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "MongoDB Atlas Programmatic Private Key",
				Sensitive:   true,
			},
			"base_url": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "MongoDB Atlas Base URL",
			},
			"realm_base_url": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "MongoDB Realm Base URL",
			},
			"is_mongodbgov_cloud": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "MongoDB Atlas Base URL default to gov",
			},
			"assume_role": assumeRoleSchema(),
			"secret_name": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Name of secret stored in AWS Secret Manager.",
			},
			"region": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Region where secret is stored as part of AWS Secret Manager.",
			},
			"sts_endpoint": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "AWS Security Token Service endpoint. Required for cross-AWS region or cross-AWS account secrets.",
			},
			"aws_access_key_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "AWS API Access Key.",
			},
			"aws_secret_access_key": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "AWS API Access Secret Key.",
			},
			"aws_session_token": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "AWS Security Token Service provided session token.",
			},
		},
		DataSourcesMap:       getDataSourcesMap(),
		ResourcesMap:         getResourcesMap(),
		ConfigureContextFunc: providerConfigure,
	}
	addBetaFeatures(provider)
	return provider
}

func getDataSourcesMap() map[string]*schema.Resource {
	dataSourcesMap := map[string]*schema.Resource{
		"mongodbatlas_advanced_cluster":                  mongodbatlas.DataSourceMongoDBAtlasAdvancedCluster(),
		"mongodbatlas_advanced_clusters":                 mongodbatlas.DataSourceMongoDBAtlasAdvancedClusters(),
		"mongodbatlas_custom_db_role":                    mongodbatlas.DataSourceMongoDBAtlasCustomDBRole(),
		"mongodbatlas_custom_db_roles":                   mongodbatlas.DataSourceMongoDBAtlasCustomDBRoles(),
		"mongodbatlas_api_key":                           mongodbatlas.DataSourceMongoDBAtlasAPIKey(),
		"mongodbatlas_api_keys":                          mongodbatlas.DataSourceMongoDBAtlasAPIKeys(),
		"mongodbatlas_access_list_api_key":               mongodbatlas.DataSourceMongoDBAtlasAccessListAPIKey(),
		"mongodbatlas_access_list_api_keys":              mongodbatlas.DataSourceMongoDBAtlasAccessListAPIKeys(),
		"mongodbatlas_project_api_key":                   mongodbatlas.DataSourceMongoDBAtlasProjectAPIKey(),
		"mongodbatlas_project_api_keys":                  mongodbatlas.DataSourceMongoDBAtlasProjectAPIKeys(),
		"mongodbatlas_roles_org_id":                      mongodbatlas.DataSourceMongoDBAtlasOrgID(),
		"mongodbatlas_cluster":                           mongodbatlas.DataSourceMongoDBAtlasCluster(),
		"mongodbatlas_clusters":                          mongodbatlas.DataSourceMongoDBAtlasClusters(),
		"mongodbatlas_network_container":                 mongodbatlas.DataSourceMongoDBAtlasNetworkContainer(),
		"mongodbatlas_network_containers":                mongodbatlas.DataSourceMongoDBAtlasNetworkContainers(),
		"mongodbatlas_network_peering":                   mongodbatlas.DataSourceMongoDBAtlasNetworkPeering(),
		"mongodbatlas_network_peerings":                  mongodbatlas.DataSourceMongoDBAtlasNetworkPeerings(),
		"mongodbatlas_maintenance_window":                mongodbatlas.DataSourceMongoDBAtlasMaintenanceWindow(),
		"mongodbatlas_auditing":                          mongodbatlas.DataSourceMongoDBAtlasAuditing(),
		"mongodbatlas_team":                              mongodbatlas.DataSourceMongoDBAtlasTeam(),
		"mongodbatlas_teams":                             mongodbatlas.DataSourceMongoDBAtlasTeam(),
		"mongodbatlas_global_cluster_config":             mongodbatlas.DataSourceMongoDBAtlasGlobalCluster(),
		"mongodbatlas_x509_authentication_database_user": mongodbatlas.DataSourceMongoDBAtlasX509AuthDBUser(),
		"mongodbatlas_private_endpoint_regional_mode":    mongodbatlas.DataSourceMongoDBAtlasPrivateEndpointRegionalMode(),
		"mongodbatlas_privatelink_endpoint_service_data_federation_online_archive":  mongodbatlas.DataSourceMongoDBAtlasPrivatelinkEndpointServiceDataFederationOnlineArchive(),
		"mongodbatlas_privatelink_endpoint_service_data_federation_online_archives": mongodbatlas.DataSourceMongoDBAtlasPrivatelinkEndpointServiceDataFederationOnlineArchives(),
		"mongodbatlas_privatelink_endpoint":                                         mongodbatlas.DataSourceMongoDBAtlasPrivateLinkEndpoint(),
		"mongodbatlas_privatelink_endpoint_service":                                 mongodbatlas.DataSourceMongoDBAtlasPrivateEndpointServiceLink(),
		"mongodbatlas_privatelink_endpoint_service_serverless":                      mongodbatlas.DataSourceMongoDBAtlasPrivateLinkEndpointServerless(),
		"mongodbatlas_privatelink_endpoints_service_serverless":                     mongodbatlas.DataSourceMongoDBAtlasPrivateLinkEndpointsServiceServerless(),
		"mongodbatlas_cloud_backup_schedule":                                        mongodbatlas.DataSourceMongoDBAtlasCloudBackupSchedule(),
		"mongodbatlas_third_party_integration":                                      mongodbatlas.DataSourceMongoDBAtlasThirdPartyIntegration(),
		"mongodbatlas_third_party_integrations":                                     mongodbatlas.DataSourceMongoDBAtlasThirdPartyIntegrations(),
		"mongodbatlas_cloud_provider_access":                                        mongodbatlas.DataSourceMongoDBAtlasCloudProviderAccessList(),
		"mongodbatlas_cloud_provider_access_setup":                                  mongodbatlas.DataSourceMongoDBAtlasCloudProviderAccessSetup(),
		"mongodbatlas_custom_dns_configuration_cluster_aws":                         mongodbatlas.DataSourceMongoDBAtlasCustomDNSConfigurationAWS(),
		"mongodbatlas_online_archive":                                               mongodbatlas.DataSourceMongoDBAtlasOnlineArchive(),
		"mongodbatlas_online_archives":                                              mongodbatlas.DataSourceMongoDBAtlasOnlineArchives(),
		"mongodbatlas_ldap_configuration":                                           mongodbatlas.DataSourceMongoDBAtlasLDAPConfiguration(),
		"mongodbatlas_ldap_verify":                                                  mongodbatlas.DataSourceMongoDBAtlasLDAPVerify(),
		"mongodbatlas_search_index":                                                 mongodbatlas.DataSourceMongoDBAtlasSearchIndex(),
		"mongodbatlas_search_indexes":                                               mongodbatlas.DataSourceMongoDBAtlasSearchIndexes(),
		"mongodbatlas_data_lake_pipeline_run":                                       mongodbatlas.DataSourceMongoDBAtlasDataLakePipelineRun(),
		"mongodbatlas_data_lake_pipeline_runs":                                      mongodbatlas.DataSourceMongoDBAtlasDataLakePipelineRuns(),
		"mongodbatlas_data_lake_pipeline":                                           mongodbatlas.DataSourceMongoDBAtlasDataLakePipeline(),
		"mongodbatlas_data_lake_pipelines":                                          mongodbatlas.DataSourceMongoDBAtlasDataLakePipelines(),
		"mongodbatlas_event_trigger":                                                mongodbatlas.DataSourceMongoDBAtlasEventTrigger(),
		"mongodbatlas_event_triggers":                                               mongodbatlas.DataSourceMongoDBAtlasEventTriggers(),
		"mongodbatlas_project_invitation":                                           mongodbatlas.DataSourceMongoDBAtlasProjectInvitation(),
		"mongodbatlas_org_invitation":                                               mongodbatlas.DataSourceMongoDBAtlasOrgInvitation(),
		"mongodbatlas_organization":                                                 mongodbatlas.DataSourceMongoDBAtlasOrganization(),
		"mongodbatlas_organizations":                                                mongodbatlas.DataSourceMongoDBAtlasOrganizations(),
		"mongodbatlas_cloud_backup_snapshot":                                        mongodbatlas.DataSourceMongoDBAtlasCloudBackupSnapshot(),
		"mongodbatlas_cloud_backup_snapshots":                                       mongodbatlas.DataSourceMongoDBAtlasCloudBackupSnapshots(),
		"mongodbatlas_backup_compliance_policy":                                     mongodbatlas.DataSourceMongoDBAtlasBackupCompliancePolicy(),
		"mongodbatlas_cloud_backup_snapshot_restore_job":                            mongodbatlas.DataSourceMongoDBAtlasCloudBackupSnapshotRestoreJob(),
		"mongodbatlas_cloud_backup_snapshot_restore_jobs":                           mongodbatlas.DataSourceMongoDBAtlasCloudBackupSnapshotRestoreJobs(),
		"mongodbatlas_cloud_backup_snapshot_export_bucket":                          mongodbatlas.DatasourceMongoDBAtlasCloudBackupSnapshotExportBucket(),
		"mongodbatlas_cloud_backup_snapshot_export_buckets":                         mongodbatlas.DatasourceMongoDBAtlasCloudBackupSnapshotExportBuckets(),
		"mongodbatlas_cloud_backup_snapshot_export_job":                             mongodbatlas.DatasourceMongoDBAtlasCloudBackupSnapshotExportJob(),
		"mongodbatlas_cloud_backup_snapshot_export_jobs":                            mongodbatlas.DatasourceMongoDBAtlasCloudBackupSnapshotExportJobs(),
		"mongodbatlas_federated_settings":                                           mongodbatlas.DataSourceMongoDBAtlasFederatedSettings(),
		"mongodbatlas_federated_settings_identity_provider":                         mongodbatlas.DataSourceMongoDBAtlasFederatedSettingsIdentityProvider(),
		"mongodbatlas_federated_settings_identity_providers":                        mongodbatlas.DataSourceMongoDBAtlasFederatedSettingsIdentityProviders(),
		"mongodbatlas_federated_settings_org_config":                                mongodbatlas.DataSourceMongoDBAtlasFederatedSettingsOrganizationConfig(),
		"mongodbatlas_federated_settings_org_configs":                               mongodbatlas.DataSourceMongoDBAtlasFederatedSettingsOrganizationConfigs(),
		"mongodbatlas_federated_settings_org_role_mapping":                          mongodbatlas.DataSourceMongoDBAtlasFederatedSettingsOrganizationRoleMapping(),
		"mongodbatlas_federated_settings_org_role_mappings":                         mongodbatlas.DataSourceMongoDBAtlasFederatedSettingsOrganizationRoleMappings(),
		"mongodbatlas_federated_database_instance":                                  mongodbatlas.DataSourceMongoDBAtlasFederatedDatabaseInstance(),
		"mongodbatlas_federated_database_instances":                                 mongodbatlas.DataSourceMongoDBAtlasFederatedDatabaseInstances(),
		"mongodbatlas_federated_query_limit":                                        mongodbatlas.DataSourceMongoDBAtlasFederatedDatabaseQueryLimit(),
		"mongodbatlas_federated_query_limits":                                       mongodbatlas.DataSourceMongoDBAtlasFederatedDatabaseQueryLimits(),
		"mongodbatlas_serverless_instance":                                          mongodbatlas.DataSourceMongoDBAtlasServerlessInstance(),
		"mongodbatlas_serverless_instances":                                         mongodbatlas.DataSourceMongoDBAtlasServerlessInstances(),
		"mongodbatlas_cluster_outage_simulation":                                    mongodbatlas.DataSourceMongoDBAtlasClusterOutageSimulation(),
		"mongodbatlas_shared_tier_restore_job":                                      mongodbatlas.DataSourceMongoDBAtlasCloudSharedTierRestoreJob(),
		"mongodbatlas_shared_tier_restore_jobs":                                     mongodbatlas.DataSourceMongoDBAtlasCloudSharedTierRestoreJobs(),
		"mongodbatlas_shared_tier_snapshot":                                         mongodbatlas.DataSourceMongoDBAtlasSharedTierSnapshot(),
		"mongodbatlas_shared_tier_snapshots":                                        mongodbatlas.DataSourceMongoDBAtlasSharedTierSnapshots(),
	}
	return dataSourcesMap
}

func getResourcesMap() map[string]*schema.Resource {
	resourcesMap := map[string]*schema.Resource{
		"mongodbatlas_advanced_cluster":                  mongodbatlas.ResourceMongoDBAtlasAdvancedCluster(),
		"mongodbatlas_api_key":                           mongodbatlas.ResourceMongoDBAtlasAPIKey(),
		"mongodbatlas_access_list_api_key":               mongodbatlas.ResourceMongoDBAtlasAccessListAPIKey(),
		"mongodbatlas_project_api_key":                   mongodbatlas.ResourceMongoDBAtlasProjectAPIKey(),
		"mongodbatlas_custom_db_role":                    mongodbatlas.ResourceMongoDBAtlasCustomDBRole(),
		"mongodbatlas_cluster":                           mongodbatlas.ResourceMongoDBAtlasCluster(),
		"mongodbatlas_network_container":                 mongodbatlas.ResourceMongoDBAtlasNetworkContainer(),
		"mongodbatlas_network_peering":                   mongodbatlas.ResourceMongoDBAtlasNetworkPeering(),
		"mongodbatlas_maintenance_window":                mongodbatlas.ResourceMongoDBAtlasMaintenanceWindow(),
		"mongodbatlas_auditing":                          mongodbatlas.ResourceMongoDBAtlasAuditing(),
		"mongodbatlas_team":                              mongodbatlas.ResourceMongoDBAtlasTeam(),
		"mongodbatlas_teams":                             mongodbatlas.ResourceMongoDBAtlasTeam(),
		"mongodbatlas_global_cluster_config":             mongodbatlas.ResourceMongoDBAtlasGlobalCluster(),
		"mongodbatlas_x509_authentication_database_user": mongodbatlas.ResourceMongoDBAtlasX509AuthDBUser(),
		"mongodbatlas_private_endpoint_regional_mode":    mongodbatlas.ResourceMongoDBAtlasPrivateEndpointRegionalMode(),
		"mongodbatlas_privatelink_endpoint_service_data_federation_online_archive": mongodbatlas.ResourceMongoDBAtlasPrivatelinkEndpointServiceDataFederationOnlineArchive(),
		"mongodbatlas_privatelink_endpoint":                                        mongodbatlas.ResourceMongoDBAtlasPrivateLinkEndpoint(),
		"mongodbatlas_privatelink_endpoint_serverless":                             mongodbatlas.ResourceMongoDBAtlasPrivateLinkEndpointServerless(),
		"mongodbatlas_privatelink_endpoint_service":                                mongodbatlas.ResourceMongoDBAtlasPrivateEndpointServiceLink(),
		"mongodbatlas_privatelink_endpoint_service_serverless":                     mongodbatlas.ResourceMongoDBAtlasPrivateLinkEndpointServiceServerless(),
		"mongodbatlas_third_party_integration":                                     mongodbatlas.ResourceMongoDBAtlasThirdPartyIntegration(),
		"mongodbatlas_cloud_provider_access":                                       mongodbatlas.ResourceMongoDBAtlasCloudProviderAccess(),
		"mongodbatlas_online_archive":                                              mongodbatlas.ResourceMongoDBAtlasOnlineArchive(),
		"mongodbatlas_custom_dns_configuration_cluster_aws":                        mongodbatlas.ResourceMongoDBAtlasCustomDNSConfiguration(),
		"mongodbatlas_ldap_configuration":                                          mongodbatlas.ResourceMongoDBAtlasLDAPConfiguration(),
		"mongodbatlas_ldap_verify":                                                 mongodbatlas.ResourceMongoDBAtlasLDAPVerify(),
		"mongodbatlas_cloud_provider_access_setup":                                 mongodbatlas.ResourceMongoDBAtlasCloudProviderAccessSetup(),
		"mongodbatlas_cloud_provider_access_authorization":                         mongodbatlas.ResourceMongoDBAtlasCloudProviderAccessAuthorization(),
		"mongodbatlas_search_index":                                                mongodbatlas.ResourceMongoDBAtlasSearchIndex(),
		"mongodbatlas_data_lake_pipeline":                                          mongodbatlas.ResourceMongoDBAtlasDataLakePipeline(),
		"mongodbatlas_event_trigger":                                               mongodbatlas.ResourceMongoDBAtlasEventTriggers(),
		"mongodbatlas_cloud_backup_schedule":                                       mongodbatlas.ResourceMongoDBAtlasCloudBackupSchedule(),
		"mongodbatlas_project_invitation":                                          mongodbatlas.ResourceMongoDBAtlasProjectInvitation(),
		"mongodbatlas_org_invitation":                                              mongodbatlas.ResourceMongoDBAtlasOrgInvitation(),
		"mongodbatlas_organization":                                                mongodbatlas.ResourceMongoDBAtlasOrganization(),
		"mongodbatlas_cloud_backup_snapshot":                                       mongodbatlas.ResourceMongoDBAtlasCloudBackupSnapshot(),
		"mongodbatlas_backup_compliance_policy":                                    mongodbatlas.ResourceMongoDBAtlasBackupCompliancePolicy(),
		"mongodbatlas_cloud_backup_snapshot_restore_job":                           mongodbatlas.ResourceMongoDBAtlasCloudBackupSnapshotRestoreJob(),
		"mongodbatlas_cloud_backup_snapshot_export_bucket":                         mongodbatlas.ResourceMongoDBAtlasCloudBackupSnapshotExportBucket(),
		"mongodbatlas_cloud_backup_snapshot_export_job":                            mongodbatlas.ResourceMongoDBAtlasCloudBackupSnapshotExportJob(),
		"mongodbatlas_federated_settings_org_config":                               mongodbatlas.ResourceMongoDBAtlasFederatedSettingsOrganizationConfig(),
		"mongodbatlas_federated_settings_org_role_mapping":                         mongodbatlas.ResourceMongoDBAtlasFederatedSettingsOrganizationRoleMapping(),
		"mongodbatlas_federated_settings_identity_provider":                        mongodbatlas.ResourceMongoDBAtlasFederatedSettingsIdentityProvider(),
		"mongodbatlas_federated_database_instance":                                 mongodbatlas.ResourceMongoDBAtlasFederatedDatabaseInstance(),
		"mongodbatlas_federated_query_limit":                                       mongodbatlas.ResourceMongoDBAtlasFederatedDatabaseQueryLimit(),
		"mongodbatlas_serverless_instance":                                         mongodbatlas.ResourceMongoDBAtlasServerlessInstance(),
		"mongodbatlas_cluster_outage_simulation":                                   mongodbatlas.ResourceMongoDBAtlasClusterOutageSimulation(),
	}
	return resourcesMap
}

func addBetaFeatures(provider *schema.Provider) {
	if ProviderEnableBeta {
		return
	}
}

func providerConfigure(ctx context.Context, d *schema.ResourceData) (any, diag.Diagnostics) {
	diagnostics := setDefaultsAndValidations(d)
	if diagnostics.HasError() {
		return nil, diagnostics
	}

	cfg := config.Config{
		PublicKey:    d.Get("public_key").(string),
		PrivateKey:   d.Get("private_key").(string),
		BaseURL:      d.Get("base_url").(string),
		RealmBaseURL: d.Get("realm_base_url").(string),
	}

	assumeRoleValue, ok := d.GetOk("assume_role")
	awsRoleDefined := ok && len(assumeRoleValue.([]any)) > 0 && assumeRoleValue.([]any)[0] != nil
	if awsRoleDefined {
		cfg.AssumeRole = expandAssumeRole(assumeRoleValue.([]any)[0].(map[string]any))
		secret := d.Get("secret_name").(string)
		region := util.MongoDBRegionToAWSRegion(d.Get("region").(string))
		awsAccessKeyID := d.Get("aws_access_key_id").(string)
		awsSecretAccessKey := d.Get("aws_secret_access_key").(string)
		awsSessionToken := d.Get("aws_session_token").(string)
		endpoint := d.Get("sts_endpoint").(string)
		var err error
		cfg, err = configureCredentialsSTS(cfg, secret, region, awsAccessKeyID, awsSecretAccessKey, awsSessionToken, endpoint)
		if err != nil {
			return nil, append(diagnostics, diag.FromErr(err)...)
		}
	}

	client, err := cfg.NewClient(ctx)
	if err != nil {
		return nil, append(diagnostics, diag.FromErr(err)...)
	}
	return client, diagnostics
}

func setDefaultsAndValidations(d *schema.ResourceData) diag.Diagnostics {
	diagnostics := []diag.Diagnostic{}

	mongodbgovCloud := pointy.Bool(d.Get("is_mongodbgov_cloud").(bool))
	if *mongodbgovCloud {
		if err := d.Set("base_url", config.MongodbGovCloudURL); err != nil {
			return append(diagnostics, diag.FromErr(err)...)
		}
	}

	if err := setValueFromConfigOrEnv(d, "base_url", []string{
		"MONGODB_ATLAS_BASE_URL",
		"MCLI_OPS_MANAGER_URL",
	}); err != nil {
		return append(diagnostics, diag.FromErr(err)...)
	}

	awsRoleDefined := false
	assumeRoles := d.Get("assume_role").([]any)
	if len(assumeRoles) == 0 {
		roleArn := MultiEnvDefaultFunc([]string{
			"ASSUME_ROLE_ARN",
			"TF_VAR_ASSUME_ROLE_ARN",
		}, "").(string)
		if roleArn != "" {
			awsRoleDefined = true
			if err := d.Set("assume_role", []map[string]any{{"role_arn": roleArn}}); err != nil {
				return append(diagnostics, diag.FromErr(err)...)
			}
		}
	} else {
		awsRoleDefined = true
	}

	if err := setValueFromConfigOrEnv(d, "public_key", []string{
		"MONGODB_ATLAS_PUBLIC_KEY",
		"MCLI_PUBLIC_API_KEY",
	}); err != nil {
		return append(diagnostics, diag.FromErr(err)...)
	}
	if d.Get("public_key").(string) == "" && !awsRoleDefined {
		diagnostics = append(diagnostics, diag.Diagnostic{Severity: diag.Warning, Summary: config.MissingAuthAttrError})
	}

	if err := setValueFromConfigOrEnv(d, "private_key", []string{
		"MONGODB_ATLAS_PRIVATE_KEY",
		"MCLI_PRIVATE_API_KEY",
	}); err != nil {
		return append(diagnostics, diag.FromErr(err)...)
	}

	if d.Get("private_key").(string) == "" && !awsRoleDefined {
		diagnostics = append(diagnostics, diag.Diagnostic{Severity: diag.Warning, Summary: config.MissingAuthAttrError})
	}

	if err := setValueFromConfigOrEnv(d, "realm_base_url", []string{
		"MONGODB_REALM_BASE_URL",
	}); err != nil {
		return append(diagnostics, diag.FromErr(err)...)
	}

	if err := setValueFromConfigOrEnv(d, "region", []string{
		"AWS_REGION",
		"TF_VAR_AWS_REGION",
	}); err != nil {
		return append(diagnostics, diag.FromErr(err)...)
	}

	if err := setValueFromConfigOrEnv(d, "sts_endpoint", []string{
		"STS_ENDPOINT",
		"TF_VAR_STS_ENDPOINT",
	}); err != nil {
		return append(diagnostics, diag.FromErr(err)...)
	}

	if err := setValueFromConfigOrEnv(d, "aws_access_key_id", []string{
		"AWS_ACCESS_KEY_ID",
		"TF_VAR_AWS_ACCESS_KEY_ID",
	}); err != nil {
		return append(diagnostics, diag.FromErr(err)...)
	}

	if err := setValueFromConfigOrEnv(d, "aws_secret_access_key", []string{
		"AWS_SECRET_ACCESS_KEY",
		"TF_VAR_AWS_SECRET_ACCESS_KEY",
	}); err != nil {
		return append(diagnostics, diag.FromErr(err)...)
	}

	if err := setValueFromConfigOrEnv(d, "secret_name", []string{
		"SECRET_NAME",
		"TF_VAR_SECRET_NAME",
	}); err != nil {
		return append(diagnostics, diag.FromErr(err)...)
	}

	if err := setValueFromConfigOrEnv(d, "aws_session_token", []string{
		"AWS_SESSION_TOKEN",
		"TF_VAR_AWS_SESSION_TOKEN",
	}); err != nil {
		return append(diagnostics, diag.FromErr(err)...)
	}

	return diagnostics
}

func setValueFromConfigOrEnv(d *schema.ResourceData, attrName string, envVars []string) error {
	var val = d.Get(attrName).(string)
	if val == "" {
		val = MultiEnvDefaultFunc(envVars, "").(string)
	}
	return d.Set(attrName, val)
}

func MultiEnvDefaultFunc(ks []string, def any) any {
	for _, k := range ks {
		if v := os.Getenv(k); v != "" {
			return v
		}
	}
	return def
}

func configureCredentialsSTS(cfg config.Config, secret, region, awsAccessKeyID, awsSecretAccessKey, awsSessionToken, endpoint string) (config.Config, error) {
	ep, err := endpoints.GetSTSRegionalEndpoint("regional")
	if err != nil {
		log.Printf("GetSTSRegionalEndpoint error: %s", err)
		return cfg, err
	}

	defaultResolver := endpoints.DefaultResolver()
	stsCustResolverFn := func(service, region string, optFns ...func(*endpoints.Options)) (endpoints.ResolvedEndpoint, error) {
		if service == endpoints.StsServiceID {
			if endpoint == "" {
				return endpoints.ResolvedEndpoint{
					URL:           config.EndPointSTSDefault,
					SigningRegion: region,
				}, nil
			}
			return endpoints.ResolvedEndpoint{
				URL:           endpoint,
				SigningRegion: region,
			}, nil
		}

		return defaultResolver.EndpointFor(service, region, optFns...)
	}

	configAWS := aws.Config{
		Region:              aws.String(region),
		Credentials:         credentials.NewStaticCredentials(awsAccessKeyID, awsSecretAccessKey, awsSessionToken),
		STSRegionalEndpoint: ep,
		EndpointResolver:    endpoints.ResolverFunc(stsCustResolverFn),
	}

	sess := session.Must(session.NewSession(&configAWS))

	creds := stscreds.NewCredentials(sess, cfg.AssumeRole.RoleARN)

	_, err = sess.Config.Credentials.Get()
	if err != nil {
		log.Printf("Session get credentials error: %s", err)
		return cfg, err
	}
	_, err = creds.Get()
	if err != nil {
		log.Printf("STS get credentials error: %s", err)
		return cfg, err
	}
	secretString, err := secretsManagerGetSecretValue(sess, &aws.Config{Credentials: creds, Region: aws.String(region)}, secret)
	if err != nil {
		log.Printf("Get Secrets error: %s", err)
		return cfg, err
	}

	var secretData SecretData
	err = json.Unmarshal([]byte(secretString), &secretData)
	if err != nil {
		return cfg, err
	}
	if secretData.PrivateKey == "" {
		return cfg, fmt.Errorf("secret missing value for credential PrivateKey")
	}

	if secretData.PublicKey == "" {
		return cfg, fmt.Errorf("secret missing value for credential PublicKey")
	}

	cfg.PublicKey = secretData.PublicKey
	cfg.PrivateKey = secretData.PrivateKey
	return cfg, nil
}

func secretsManagerGetSecretValue(sess *session.Session, creds *aws.Config, secret string) (string, error) {
	svc := secretsmanager.New(sess, creds)
	input := &secretsmanager.GetSecretValueInput{
		SecretId:     aws.String(secret),
		VersionStage: aws.String("AWSCURRENT"),
	}

	result, err := svc.GetSecretValue(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case secretsmanager.ErrCodeResourceNotFoundException:
				log.Println(secretsmanager.ErrCodeResourceNotFoundException, aerr.Error())
			case secretsmanager.ErrCodeInvalidParameterException:
				log.Println(secretsmanager.ErrCodeInvalidParameterException, aerr.Error())
			case secretsmanager.ErrCodeInvalidRequestException:
				log.Println(secretsmanager.ErrCodeInvalidRequestException, aerr.Error())
			case secretsmanager.ErrCodeDecryptionFailure:
				log.Println(secretsmanager.ErrCodeDecryptionFailure, aerr.Error())
			case secretsmanager.ErrCodeInternalServiceError:
				log.Println(secretsmanager.ErrCodeInternalServiceError, aerr.Error())
			default:
				log.Println(aerr.Error())
			}
		} else {
			log.Println(err.Error())
		}
		return "", err
	}

	return *result.SecretString, err
}

// assumeRoleSchema From aws provider.go
func assumeRoleSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"duration": {
					Type:         schema.TypeString,
					Optional:     true,
					Description:  "The duration, between 15 minutes and 12 hours, of the role session. Valid time units are ns, us (or µs), ms, s, h, or m.",
					ValidateFunc: validAssumeRoleDuration,
				},
				"external_id": {
					Type:        schema.TypeString,
					Optional:    true,
					Description: "A unique identifier that might be required when you assume a role in another account.",
					ValidateFunc: validation.All(
						validation.StringLenBetween(2, 1224),
						validation.StringMatch(regexp.MustCompile(`[\w+=,.@:/\-]*`), ""),
					),
				},
				"policy": {
					Type:         schema.TypeString,
					Optional:     true,
					Description:  "IAM Policy JSON describing further restricting permissions for the IAM Role being assumed.",
					ValidateFunc: validation.StringIsJSON,
				},
				"policy_arns": {
					Type:        schema.TypeSet,
					Optional:    true,
					Description: "Amazon Resource Names (ARNs) of IAM Policies describing further restricting permissions for the IAM Role being assumed.",
					Elem: &schema.Schema{
						Type: schema.TypeString,
					},
				},
				"role_arn": {
					Type:        schema.TypeString,
					Optional:    true,
					Description: "Amazon Resource Name (ARN) of an IAM Role to assume prior to making API calls.",
				},
				"session_name": {
					Type:         schema.TypeString,
					Optional:     true,
					Description:  "An identifier for the assumed role session.",
					ValidateFunc: validAssumeRoleSessionName,
				},
				"source_identity": {
					Type:         schema.TypeString,
					Optional:     true,
					Description:  "Source identity specified by the principal assuming the role.",
					ValidateFunc: validAssumeRoleSourceIdentity,
				},
				"tags": {
					Type:        schema.TypeMap,
					Optional:    true,
					Description: "Assume role session tags.",
					Elem:        &schema.Schema{Type: schema.TypeString},
				},
				"transitive_tag_keys": {
					Type:        schema.TypeSet,
					Optional:    true,
					Description: "Assume role session tag keys to pass to any subsequent sessions.",
					Elem:        &schema.Schema{Type: schema.TypeString},
				},
			},
		},
	}
}

var validAssumeRoleSessionName = validation.All(
	validation.StringLenBetween(2, 64),
	validation.StringMatch(regexp.MustCompile(`[\w+=,.@\-]*`), ""),
)

var validAssumeRoleSourceIdentity = validation.All(
	validation.StringLenBetween(2, 64),
	validation.StringMatch(regexp.MustCompile(`[\w+=,.@\-]*`), ""),
)

// validAssumeRoleDuration validates a string can be parsed as a valid time.Duration
// and is within a minimum of 15 minutes and maximum of 12 hours
func validAssumeRoleDuration(v any, k string) (ws []string, errorResults []error) {
	duration, err := time.ParseDuration(v.(string))

	if err != nil {
		errorResults = append(errorResults, fmt.Errorf("%q cannot be parsed as a duration: %w", k, err))
		return
	}

	if duration.Minutes() < 15 || duration.Hours() > 12 {
		errorResults = append(errorResults, fmt.Errorf("duration %q must be between 15 minutes (15m) and 12 hours (12h), inclusive", k))
	}

	return
}

func expandAssumeRole(tfMap map[string]any) *config.AssumeRole {
	if tfMap == nil {
		return nil
	}

	assumeRole := config.AssumeRole{}

	if v, ok := tfMap["duration"].(string); ok && v != "" {
		duration, _ := time.ParseDuration(v)
		assumeRole.Duration = duration
	}

	if v, ok := tfMap["external_id"].(string); ok && v != "" {
		assumeRole.ExternalID = v
	}

	if v, ok := tfMap["policy"].(string); ok && v != "" {
		assumeRole.Policy = v
	}

	if v, ok := tfMap["policy_arns"].(*schema.Set); ok && v.Len() > 0 {
		assumeRole.PolicyARNs = config.ExpandStringList(v.List())
	}

	if v, ok := tfMap["role_arn"].(string); ok && v != "" {
		assumeRole.RoleARN = v
	}

	if v, ok := tfMap["session_name"].(string); ok && v != "" {
		assumeRole.SessionName = v
	}

	if v, ok := tfMap["source_identity"].(string); ok && v != "" {
		assumeRole.SourceIdentity = v
	}

	if v, ok := tfMap["transitive_tag_keys"].(*schema.Set); ok && v.Len() > 0 {
		assumeRole.TransitiveTagKeys = config.ExpandStringList(v.List())
	}

	return &assumeRole
}
