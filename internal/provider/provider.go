package provider

import (
	"context"
	"log"
	"os"
	"regexp"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-mux/tf5to6server"
	"github.com/hashicorp/terraform-plugin-mux/tf6muxserver"

	"github.com/mongodb/terraform-provider-mongodbatlas/internal/common/conversion"
	"github.com/mongodb/terraform-provider-mongodbatlas/internal/common/validate"
	"github.com/mongodb/terraform-provider-mongodbatlas/internal/config"
	"github.com/mongodb/terraform-provider-mongodbatlas/internal/service/advancedclustertpf"
	"github.com/mongodb/terraform-provider-mongodbatlas/internal/service/alertconfiguration"
	"github.com/mongodb/terraform-provider-mongodbatlas/internal/service/atlasuser"
	"github.com/mongodb/terraform-provider-mongodbatlas/internal/service/controlplaneipaddresses"
	"github.com/mongodb/terraform-provider-mongodbatlas/internal/service/databaseuser"
	"github.com/mongodb/terraform-provider-mongodbatlas/internal/service/encryptionatrest"
	"github.com/mongodb/terraform-provider-mongodbatlas/internal/service/encryptionatrestprivateendpoint"
	"github.com/mongodb/terraform-provider-mongodbatlas/internal/service/flexcluster"
	"github.com/mongodb/terraform-provider-mongodbatlas/internal/service/mongodbemployeeaccessgrant"
	"github.com/mongodb/terraform-provider-mongodbatlas/internal/service/project"
	"github.com/mongodb/terraform-provider-mongodbatlas/internal/service/projectipaccesslist"
	"github.com/mongodb/terraform-provider-mongodbatlas/internal/service/projectipaddresses"
	"github.com/mongodb/terraform-provider-mongodbatlas/internal/service/pushbasedlogexport"
	"github.com/mongodb/terraform-provider-mongodbatlas/internal/service/resourcepolicy"
	"github.com/mongodb/terraform-provider-mongodbatlas/internal/service/searchdeployment"
	"github.com/mongodb/terraform-provider-mongodbatlas/internal/service/streamconnection"
	"github.com/mongodb/terraform-provider-mongodbatlas/internal/service/streaminstance"
	"github.com/mongodb/terraform-provider-mongodbatlas/internal/service/streamprivatelinkendpoint"
	"github.com/mongodb/terraform-provider-mongodbatlas/internal/service/streamprocessor"
	"github.com/mongodb/terraform-provider-mongodbatlas/version"
)

var (
	providerEnablePreview, _ = strconv.ParseBool(os.Getenv("MONGODB_ATLAS_ENABLE_PREVIEW"))
)

const (
	MongodbGovCloudURL    = "https://cloud.mongodbgov.com"
	MongodbGovCloudQAURL  = "https://cloud-qa.mongodbgov.com"
	MongodbGovCloudDevURL = "https://cloud-dev.mongodbgov.com"
	ProviderConfigError   = "error in configuring the provider."
	MissingAuthAttrError  = "either Atlas Programmatic API Keys or AWS Secrets Manager attributes must be set"
)

type MongodbtlasProvider struct {
	proxyPort *int
}

type tfMongodbAtlasProviderModel struct {
	AssumeRole           types.List   `tfsdk:"assume_role"`
	PublicKey            types.String `tfsdk:"public_key"`
	PrivateKey           types.String `tfsdk:"private_key"`
	BaseURL              types.String `tfsdk:"base_url"`
	RealmBaseURL         types.String `tfsdk:"realm_base_url"`
	SecretName           types.String `tfsdk:"secret_name"`
	Region               types.String `tfsdk:"region"`
	StsEndpoint          types.String `tfsdk:"sts_endpoint"`
	AwsAccessKeyID       types.String `tfsdk:"aws_access_key_id"`
	AwsSecretAccessKeyID types.String `tfsdk:"aws_secret_access_key"`
	AwsSessionToken      types.String `tfsdk:"aws_session_token"`
	IsMongodbGovCloud    types.Bool   `tfsdk:"is_mongodbgov_cloud"`
}

type tfAssumeRoleModel struct {
	PolicyARNs        types.Set    `tfsdk:"policy_arns"`
	TransitiveTagKeys types.Set    `tfsdk:"transitive_tag_keys"`
	Tags              types.Map    `tfsdk:"tags"`
	Duration          types.String `tfsdk:"duration"`
	ExternalID        types.String `tfsdk:"external_id"`
	Policy            types.String `tfsdk:"policy"`
	RoleARN           types.String `tfsdk:"role_arn"`
	SessionName       types.String `tfsdk:"session_name"`
	SourceIdentity    types.String `tfsdk:"source_identity"`
}

var AssumeRoleType = types.ObjectType{AttrTypes: map[string]attr.Type{
	"policy_arns":         types.SetType{ElemType: types.StringType},
	"transitive_tag_keys": types.SetType{ElemType: types.StringType},
	"tags":                types.MapType{ElemType: types.StringType},
	"duration":            types.StringType,
	"external_id":         types.StringType,
	"policy":              types.StringType,
	"role_arn":            types.StringType,
	"session_name":        types.StringType,
	"source_identity":     types.StringType,
}}

func (p *MongodbtlasProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "mongodbatlas"
	resp.Version = version.ProviderVersion
}

func (p *MongodbtlasProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Blocks: map[string]schema.Block{
			"assume_role": fwAssumeRoleSchema,
		},
		Attributes: map[string]schema.Attribute{
			"public_key": schema.StringAttribute{
				Optional:    true,
				Description: "MongoDB Atlas Programmatic Public Key",
			},
			"private_key": schema.StringAttribute{
				Optional:    true,
				Description: "MongoDB Atlas Programmatic Private Key",
				Sensitive:   true,
			},
			"base_url": schema.StringAttribute{
				Optional:    true,
				Description: "MongoDB Atlas Base URL",
			},
			"realm_base_url": schema.StringAttribute{
				Optional:    true,
				Description: "MongoDB Realm Base URL",
			},
			"is_mongodbgov_cloud": schema.BoolAttribute{
				Optional:    true,
				Description: "MongoDB Atlas Base URL default to gov",
			},
			"secret_name": schema.StringAttribute{
				Optional:    true,
				Description: "Name of secret stored in AWS Secret Manager.",
			},
			"region": schema.StringAttribute{
				Optional:    true,
				Description: "Region where secret is stored as part of AWS Secret Manager.",
			},
			"sts_endpoint": schema.StringAttribute{
				Optional:    true,
				Description: "AWS Security Token Service endpoint. Required for cross-AWS region or cross-AWS account secrets.",
			},
			"aws_access_key_id": schema.StringAttribute{
				Optional:    true,
				Description: "AWS API Access Key.",
			},
			"aws_secret_access_key": schema.StringAttribute{
				Optional:    true,
				Description: "AWS API Access Secret Key.",
			},
			"aws_session_token": schema.StringAttribute{
				Optional:    true,
				Description: "AWS Security Token Service provided session token.",
			},
		},
	}
}

var fwAssumeRoleSchema = schema.ListNestedBlock{
	Validators: []validator.List{listvalidator.SizeAtMost(1)},
	NestedObject: schema.NestedBlockObject{
		Attributes: map[string]schema.Attribute{
			"duration": schema.StringAttribute{
				Optional:    true,
				Description: "The duration, between 15 minutes and 12 hours, of the role session. Valid time units are ns, us (or µs), ms, s, h, or m.",
				Validators: []validator.String{
					validate.ValidDurationBetween(15, 12*60),
				},
			},
			"external_id": schema.StringAttribute{
				Optional:    true,
				Description: "A unique identifier that might be required when you assume a role in another account.",
				Validators: []validator.String{
					stringvalidator.LengthBetween(2, 1224),
					stringvalidator.RegexMatches(regexp.MustCompile(`[\w+=,.@:/\-]*`), ""),
				},
			},
			"policy": schema.StringAttribute{
				Optional:    true,
				Description: "IAM Policy JSON describing further restricting permissions for the IAM Role being assumed.",
				Validators: []validator.String{
					validate.StringIsJSON(),
				},
			},
			"policy_arns": schema.SetAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Description: "Amazon Resource Names (ARNs) of IAM Policies describing further restricting permissions for the IAM Role being assumed.",
			},
			"role_arn": schema.StringAttribute{
				Optional:    true,
				Description: "Amazon Resource Name (ARN) of an IAM Role to assume prior to making API calls.",
			},
			"session_name": schema.StringAttribute{
				Optional:    true,
				Description: "An identifier for the assumed role session.",
				Validators: []validator.String{
					stringvalidator.LengthBetween(2, 64),
					stringvalidator.RegexMatches(regexp.MustCompile(`[\w+=,.@\-]*`), ""),
				},
			},
			"source_identity": schema.StringAttribute{
				Optional:    true,
				Description: "Source identity specified by the principal assuming the role.",
				Validators: []validator.String{
					stringvalidator.LengthBetween(2, 64),
					stringvalidator.RegexMatches(regexp.MustCompile(`[\w+=,.@\-]*`), ""),
				},
			},
			"tags": schema.MapAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Description: "Assume role session tags.",
			},
			"transitive_tag_keys": schema.SetAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Description: "Assume role session tag keys to pass to any subsequent sessions.",
			},
		},
	},
}

func (p *MongodbtlasProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data tfMongodbAtlasProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	data = setDefaultValuesWithValidations(ctx, &data, resp)
	if resp.Diagnostics.HasError() {
		return
	}

	cfg := config.Config{
		PublicKey:        data.PublicKey.ValueString(),
		PrivateKey:       data.PrivateKey.ValueString(),
		BaseURL:          data.BaseURL.ValueString(),
		RealmBaseURL:     data.RealmBaseURL.ValueString(),
		TerraformVersion: req.TerraformVersion,
		ProxyPort:        p.proxyPort,
	}

	var assumeRoles []tfAssumeRoleModel
	data.AssumeRole.ElementsAs(ctx, &assumeRoles, true)
	awsRoleDefined := len(assumeRoles) > 0
	if awsRoleDefined {
		cfg.AssumeRole = parseTfModel(ctx, &assumeRoles[0])
		secret := data.SecretName.ValueString()
		region := conversion.MongoDBRegionToAWSRegion(data.Region.ValueString())
		awsAccessKeyID := data.AwsAccessKeyID.ValueString()
		awsSecretAccessKey := data.AwsSecretAccessKeyID.ValueString()
		awsSessionToken := data.AwsSessionToken.ValueString()
		endpoint := data.StsEndpoint.ValueString()
		var err error
		cfg, err = configureCredentialsSTS(&cfg, secret, region, awsAccessKeyID, awsSecretAccessKey, awsSessionToken, endpoint)
		if err != nil {
			resp.Diagnostics.AddError("failed to configure credentials STS", err.Error())
			return
		}
	}

	client, err := cfg.NewClient(ctx)

	if err != nil {
		resp.Diagnostics.AddError(
			"failed to initialize a new client",
			err.Error(),
		)
		return
	}

	resp.DataSourceData = client
	resp.ResourceData = client
}

// parseTfModel extracts the values from tfAssumeRoleModel creating a new instance of our internal model AssumeRole used in Config
func parseTfModel(ctx context.Context, tfAssumeRoleModel *tfAssumeRoleModel) *config.AssumeRole {
	assumeRole := config.AssumeRole{}

	if !tfAssumeRoleModel.Duration.IsNull() {
		duration, _ := time.ParseDuration(tfAssumeRoleModel.Duration.ValueString())
		assumeRole.Duration = duration
	}

	assumeRole.ExternalID = tfAssumeRoleModel.ExternalID.ValueString()
	assumeRole.Policy = tfAssumeRoleModel.Policy.ValueString()

	if !tfAssumeRoleModel.PolicyARNs.IsNull() {
		var policiesARNs []string
		tfAssumeRoleModel.PolicyARNs.ElementsAs(ctx, &policiesARNs, true)
		assumeRole.PolicyARNs = policiesARNs
	}

	assumeRole.RoleARN = tfAssumeRoleModel.RoleARN.ValueString()
	assumeRole.SessionName = tfAssumeRoleModel.SessionName.ValueString()
	assumeRole.SourceIdentity = tfAssumeRoleModel.SourceIdentity.ValueString()

	if !tfAssumeRoleModel.TransitiveTagKeys.IsNull() {
		var transitiveTagKeys []string
		tfAssumeRoleModel.TransitiveTagKeys.ElementsAs(ctx, &transitiveTagKeys, true)
		assumeRole.TransitiveTagKeys = transitiveTagKeys
	}

	return &assumeRole
}

func setDefaultValuesWithValidations(ctx context.Context, data *tfMongodbAtlasProviderModel, resp *provider.ConfigureResponse) tfMongodbAtlasProviderModel {
	if mongodbgovCloud := data.IsMongodbGovCloud.ValueBool(); mongodbgovCloud {
		if !isGovBaseURLConfiguredForProvider(data) {
			data.BaseURL = types.StringValue(MongodbGovCloudURL)
		}
	}
	if data.BaseURL.ValueString() == "" {
		data.BaseURL = types.StringValue(MultiEnvDefaultFunc([]string{
			"MONGODB_ATLAS_BASE_URL",
			"MCLI_OPS_MANAGER_URL",
		}, "").(string))
	}

	awsRoleDefined := false
	if len(data.AssumeRole.Elements()) == 0 {
		assumeRoleArn := MultiEnvDefaultFunc([]string{
			"ASSUME_ROLE_ARN",
			"TF_VAR_ASSUME_ROLE_ARN",
		}, "").(string)
		if assumeRoleArn != "" {
			awsRoleDefined = true
			var diags diag.Diagnostics
			data.AssumeRole, diags = types.ListValueFrom(ctx, AssumeRoleType, []tfAssumeRoleModel{
				{
					Tags:              types.MapNull(types.StringType),
					PolicyARNs:        types.SetNull(types.StringType),
					TransitiveTagKeys: types.SetNull(types.StringType),
					RoleARN:           types.StringValue(assumeRoleArn),
				},
			})
			if diags.HasError() {
				resp.Diagnostics.Append(diags...)
			}
		}
	} else {
		awsRoleDefined = true
	}

	if data.PublicKey.ValueString() == "" {
		data.PublicKey = types.StringValue(MultiEnvDefaultFunc([]string{
			"MONGODB_ATLAS_PUBLIC_KEY",
			"MCLI_PUBLIC_API_KEY",
		}, "").(string))
		if data.PublicKey.ValueString() == "" && !awsRoleDefined {
			resp.Diagnostics.AddWarning(ProviderConfigError, MissingAuthAttrError)
		}
	}

	if data.PrivateKey.ValueString() == "" {
		data.PrivateKey = types.StringValue(MultiEnvDefaultFunc([]string{
			"MONGODB_ATLAS_PRIVATE_KEY",
			"MCLI_PRIVATE_API_KEY",
		}, "").(string))
		if data.PrivateKey.ValueString() == "" && !awsRoleDefined {
			resp.Diagnostics.AddWarning(ProviderConfigError, MissingAuthAttrError)
		}
	}

	if data.RealmBaseURL.ValueString() == "" {
		data.RealmBaseURL = types.StringValue(MultiEnvDefaultFunc([]string{
			"MONGODB_REALM_BASE_URL",
		}, "").(string))
	}

	if data.Region.ValueString() == "" {
		data.Region = types.StringValue(MultiEnvDefaultFunc([]string{
			"AWS_REGION",
			"TF_VAR_AWS_REGION",
		}, "").(string))
	}

	if data.StsEndpoint.ValueString() == "" {
		data.StsEndpoint = types.StringValue(MultiEnvDefaultFunc([]string{
			"STS_ENDPOINT",
			"TF_VAR_STS_ENDPOINT",
		}, "").(string))
	}

	if data.AwsAccessKeyID.ValueString() == "" {
		data.AwsAccessKeyID = types.StringValue(MultiEnvDefaultFunc([]string{
			"AWS_ACCESS_KEY_ID",
			"TF_VAR_AWS_ACCESS_KEY_ID",
		}, "").(string))
	}

	if data.AwsSecretAccessKeyID.ValueString() == "" {
		data.AwsSecretAccessKeyID = types.StringValue(MultiEnvDefaultFunc([]string{
			"AWS_SECRET_ACCESS_KEY",
			"TF_VAR_AWS_SECRET_ACCESS_KEY",
		}, "").(string))
	}

	if data.AwsSessionToken.ValueString() == "" {
		data.AwsSessionToken = types.StringValue(MultiEnvDefaultFunc([]string{
			"AWS_SESSION_TOKEN",
			"TF_VAR_AWS_SESSION_TOKEN",
		}, "").(string))
	}

	if data.SecretName.ValueString() == "" {
		data.SecretName = types.StringValue(MultiEnvDefaultFunc([]string{
			"SECRET_NAME",
			"TF_VAR_SECRET_NAME",
		}, "").(string))
	}

	return *data
}

func (p *MongodbtlasProvider) DataSources(context.Context) []func() datasource.DataSource {
	dataSources := []func() datasource.DataSource{
		project.DataSource,
		project.PluralDataSource,
		databaseuser.DataSource,
		databaseuser.PluralDataSource,
		alertconfiguration.DataSource,
		alertconfiguration.PluralDataSource,
		projectipaccesslist.DataSource,
		atlasuser.DataSource,
		atlasuser.PluralDataSource,
		searchdeployment.DataSource,
		pushbasedlogexport.DataSource,
		streaminstance.DataSource,
		streaminstance.PluralDataSource,
		streamconnection.DataSource,
		streamconnection.PluralDataSource,
		controlplaneipaddresses.DataSource,
		projectipaddresses.DataSource,
		streamprocessor.DataSource,
		streamprocessor.PluralDataSource,
		encryptionatrest.DataSource,
		encryptionatrestprivateendpoint.DataSource,
		encryptionatrestprivateendpoint.PluralDataSource,
		mongodbemployeeaccessgrant.DataSource,
		streamprivatelinkendpoint.DataSource,
		streamprivatelinkendpoint.PluralDataSource,
	}
	if config.AdvancedClusterV2Schema() {
		dataSources = append(dataSources, advancedclustertpf.DataSource, advancedclustertpf.PluralDataSource)
	}
	previewDataSources := []func() datasource.DataSource{
		resourcepolicy.DataSource,
		resourcepolicy.PluralDataSource,
		flexcluster.DataSource,
		flexcluster.PluralDataSource,
	} // Data sources not yet in GA
	if providerEnablePreview {
		dataSources = append(dataSources, previewDataSources...)
	}
	return dataSources
}

func (p *MongodbtlasProvider) Resources(context.Context) []func() resource.Resource {
	resources := []func() resource.Resource{
		project.Resource,
		encryptionatrest.Resource,
		databaseuser.Resource,
		alertconfiguration.Resource,
		projectipaccesslist.Resource,
		searchdeployment.Resource,
		pushbasedlogexport.Resource,
		streaminstance.Resource,
		streamconnection.Resource,
		streamprocessor.Resource,
		encryptionatrestprivateendpoint.Resource,
		mongodbemployeeaccessgrant.Resource,
		streamprivatelinkendpoint.Resource,
	}
	if config.AdvancedClusterV2Schema() {
		resources = append(resources, advancedclustertpf.Resource)
	}
	previewResources := []func() resource.Resource{
		resourcepolicy.Resource,
		flexcluster.Resource,
	} // Resources not yet in GA
	if providerEnablePreview {
		resources = append(resources, previewResources...)
	}
	return resources
}

func NewFrameworkProvider(proxyPort *int) provider.Provider {
	return &MongodbtlasProvider{
		proxyPort: proxyPort,
	}
}

func MuxProviderFactory() func() tfprotov6.ProviderServer {
	return muxProviderFactory(nil)
}

func MuxProviderFactoryForTesting(proxyPort *int) func() tfprotov6.ProviderServer {
	return muxProviderFactory(proxyPort)
}

func muxProviderFactory(proxyPort *int) func() tfprotov6.ProviderServer {
	v2Provider := NewSdkV2Provider(proxyPort)
	newProvider := NewFrameworkProvider(proxyPort)
	ctx := context.Background()
	upgradedSdkProvider, err := tf5to6server.UpgradeServer(ctx, v2Provider.GRPCProvider)
	if err != nil {
		log.Fatal(err)
	}
	muxServer, err := tf6muxserver.NewMuxServer(ctx,
		func() tfprotov6.ProviderServer { return upgradedSdkProvider },
		providerserver.NewProtocol6(newProvider),
	)
	if err != nil {
		log.Fatal(err)
	}
	return muxServer.ProviderServer
}

func MultiEnvDefaultFunc(ks []string, def any) any {
	for _, k := range ks {
		if v := os.Getenv(k); v != "" {
			return v
		}
	}
	return def
}

func isGovBaseURLConfigured(baseURL string) bool {
	if baseURL == "" {
		baseURL = MultiEnvDefaultFunc([]string{
			"MONGODB_ATLAS_BASE_URL",
			"MCLI_OPS_MANAGER_URL",
		}, "").(string)
	}
	return baseURL == MongodbGovCloudDevURL || baseURL == MongodbGovCloudQAURL
}

func isGovBaseURLConfiguredForProvider(data *tfMongodbAtlasProviderModel) bool {
	return isGovBaseURLConfigured(data.BaseURL.ValueString())
}
