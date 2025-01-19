package advancedclustertpf

import (
	"context"
	"regexp"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	dsschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/mongodb/terraform-provider-mongodbatlas/internal/common/conversion"
	"github.com/mongodb/terraform-provider-mongodbatlas/internal/common/schemafunc"
)

func resourceSchema(ctx context.Context) schema.Schema {
	return schema.Schema{
		Version: 2,
		Attributes: map[string]schema.Attribute{
			"accept_data_risks_and_force_replica_set_reconfig": schema.StringAttribute{
				Optional:            true,
				PlanModifiers:       []planmodifier.String{schemafunc.PlanModifyStringUpdateOnly()},
				MarkdownDescription: "If reconfiguration is necessary to regain a primary due to a regional outage, submit this field alongside your topology reconfiguration to request a new regional outage resistant topology. Forced reconfigurations during an outage of the majority of electable nodes carry a risk of data loss if replicated writes (even majority committed writes) have not been replicated to the new primary node. MongoDB Atlas docs contain more information. To proceed with an operation which carries that risk, set **acceptDataRisksAndForceReplicaSetReconfig** to the current date.",
			},
			"backup_enabled": schema.BoolAttribute{
				Computed:            true,
				Optional:            true,
				MarkdownDescription: "Flag that indicates whether the cluster can perform backups. If set to `true`, the cluster can perform backups. You must set this value to `true` for NVMe clusters. Backup uses [Cloud Backups](https://docs.atlas.mongodb.com/backup/cloud-backup/overview/) for dedicated clusters and [Shared Cluster Backups](https://docs.atlas.mongodb.com/backup/shared-tier/overview/) for tenant clusters. If set to `false`, the cluster doesn't use backups.",
			},
			"bi_connector_config": schema.SingleNestedAttribute{
				Computed:            true,
				Optional:            true,
				MarkdownDescription: "Settings needed to configure the MongoDB Connector for Business Intelligence for this cluster.",
				Attributes: map[string]schema.Attribute{
					"enabled": schema.BoolAttribute{
						Computed:            true,
						Optional:            true,
						MarkdownDescription: "Flag that indicates whether MongoDB Connector for Business Intelligence is enabled on the specified cluster.",
					},
					"read_preference": schema.StringAttribute{
						Computed:            true,
						Optional:            true,
						MarkdownDescription: "Data source node designated for the MongoDB Connector for Business Intelligence on MongoDB Cloud. The MongoDB Connector for Business Intelligence on MongoDB Cloud reads data from the primary, secondary, or analytics node based on your read preferences. Defaults to `ANALYTICS` node, or `SECONDARY` if there are no `ANALYTICS` nodes.",
					},
				},
			},
			"cluster_type": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Configuration of nodes that comprise the cluster.",
			},
			"config_server_management_mode": schema.StringAttribute{
				Computed:            true,
				Optional:            true,
				MarkdownDescription: "Config Server Management Mode for creating or updating a sharded cluster.\n\nWhen configured as ATLAS_MANAGED, atlas may automatically switch the cluster's config server type for optimal performance and savings.\n\nWhen configured as FIXED_TO_DEDICATED, the cluster will always use a dedicated config server.",
			},
			"config_server_type": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Describes a sharded cluster's config server type.",
			},
			"connection_strings": schema.SingleNestedAttribute{
				Computed:            true,
				MarkdownDescription: "Collection of Uniform Resource Locators that point to the MongoDB database.",
				Attributes: map[string]schema.Attribute{
					"private": schema.StringAttribute{
						Computed:            true,
						MarkdownDescription: "Network peering connection strings for each interface Virtual Private Cloud (VPC) endpoint that you configured to connect to this cluster. This connection string uses the `mongodb+srv://` protocol. The resource returns this parameter once someone creates a network peering connection to this cluster. This protocol tells the application to look up the host seed list in the Domain Name System (DNS). This list synchronizes with the nodes in a cluster. If the connection string uses this Uniform Resource Identifier (URI) format, you don't need to append the seed list or change the URI if the nodes change. Use this URI format if your driver supports it. If it doesn't, use connectionStrings.private. For Amazon Web Services (AWS) clusters, this resource returns this parameter only if you enable custom DNS.",
					},
					"private_endpoint": schema.ListNestedAttribute{
						Computed:            true,
						MarkdownDescription: "List of private endpoint-aware connection strings that you can use to connect to this cluster through a private endpoint. This parameter returns only if you deployed a private endpoint to all regions to which you deployed this clusters' nodes.",
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"connection_string": schema.StringAttribute{
									Computed:            true,
									MarkdownDescription: "Private endpoint-aware connection string that uses the `mongodb://` protocol to connect to MongoDB Cloud through a private endpoint.",
								},
								"endpoints": schema.ListNestedAttribute{
									Computed:            true,
									MarkdownDescription: "List that contains the private endpoints through which you connect to MongoDB Cloud when you use **connectionStrings.privateEndpoint[n].connectionString** or **connectionStrings.privateEndpoint[n].srvConnectionString**.",
									NestedObject: schema.NestedAttributeObject{
										Attributes: map[string]schema.Attribute{
											"endpoint_id": schema.StringAttribute{
												Computed:            true,
												MarkdownDescription: "Unique string that the cloud provider uses to identify the private endpoint.",
											},
											"provider_name": schema.StringAttribute{
												Computed:            true,
												MarkdownDescription: "Cloud provider in which MongoDB Cloud deploys the private endpoint.",
											},
											"region": schema.StringAttribute{
												Computed:            true,
												MarkdownDescription: "Region where the private endpoint is deployed.",
											},
										},
									},
								},
								"srv_connection_string": schema.StringAttribute{
									Computed:            true,
									MarkdownDescription: "Private endpoint-aware connection string that uses the `mongodb+srv://` protocol to connect to MongoDB Cloud through a private endpoint. The `mongodb+srv` protocol tells the driver to look up the seed list of hosts in the Domain Name System (DNS). This list synchronizes with the nodes in a cluster. If the connection string uses this Uniform Resource Identifier (URI) format, you don't need to append the seed list or change the Uniform Resource Identifier (URI) if the nodes change. Use this Uniform Resource Identifier (URI) format if your application supports it. If it doesn't, use connectionStrings.privateEndpoint[n].connectionString.",
								},
								"srv_shard_optimized_connection_string": schema.StringAttribute{
									Computed:            true,
									MarkdownDescription: "Private endpoint-aware connection string optimized for sharded clusters that uses the `mongodb+srv://` protocol to connect to MongoDB Cloud through a private endpoint. If the connection string uses this Uniform Resource Identifier (URI) format, you don't need to change the Uniform Resource Identifier (URI) if the nodes change. Use this Uniform Resource Identifier (URI) format if your application and Atlas cluster supports it. If it doesn't, use and consult the documentation for connectionStrings.privateEndpoint[n].srvConnectionString.",
								},
								"type": schema.StringAttribute{
									Computed:            true,
									MarkdownDescription: "MongoDB process type to which your application connects. Use `MONGOD` for replica sets and `MONGOS` for sharded clusters.",
								},
							},
						},
					},
					"private_srv": schema.StringAttribute{
						Computed:            true,
						MarkdownDescription: "Network peering connection strings for each interface Virtual Private Cloud (VPC) endpoint that you configured to connect to this cluster. This connection string uses the `mongodb+srv://` protocol. The resource returns this parameter when someone creates a network peering connection to this cluster. This protocol tells the application to look up the host seed list in the Domain Name System (DNS). This list synchronizes with the nodes in a cluster. If the connection string uses this Uniform Resource Identifier (URI) format, you don't need to append the seed list or change the Uniform Resource Identifier (URI) if the nodes change. Use this Uniform Resource Identifier (URI) format if your driver supports it. If it doesn't, use `connectionStrings.private`. For Amazon Web Services (AWS) clusters, this parameter returns only if you [enable custom DNS](https://docs.atlas.mongodb.com/reference/api/aws-custom-dns-update/).",
					},
					"standard": schema.StringAttribute{
						Computed:            true,
						MarkdownDescription: "Public connection string that you can use to connect to this cluster. This connection string uses the `mongodb://` protocol.",
					},
					"standard_srv": schema.StringAttribute{
						Computed:            true,
						MarkdownDescription: "Public connection string that you can use to connect to this cluster. This connection string uses the `mongodb+srv://` protocol.",
					},
				},
			},
			"create_date": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Date and time when MongoDB Cloud created this cluster. This parameter expresses its value in ISO 8601 format in UTC.",
			},
			"encryption_at_rest_provider": schema.StringAttribute{
				Computed:            true,
				Optional:            true,
				MarkdownDescription: "Cloud service provider that manages your customer keys to provide an additional layer of encryption at rest for the cluster. To enable customer key management for encryption at rest, the cluster **replicationSpecs[n].regionConfigs[m].{type}Specs.instanceSize** setting must be `M10` or higher and `\"backupEnabled\" : false` or omitted entirely.",
			},
			"global_cluster_self_managed_sharding": schema.BoolAttribute{
				Computed:            true,
				Optional:            true,
				MarkdownDescription: "Set this field to configure the Sharding Management Mode when creating a new Global Cluster.\n\nWhen set to false, the management mode is set to Atlas-Managed Sharding. This mode fully manages the sharding of your Global Cluster and is built to provide a seamless deployment experience.\n\nWhen set to true, the management mode is set to Self-Managed Sharding. This mode leaves the management of shards in your hands and is built to provide an advanced and flexible deployment experience.\n\nThis setting cannot be changed once the cluster is deployed.",
			},
			"project_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Unique 24-hexadecimal digit string that identifies your project. Use the [/groups](#tag/Projects/operation/listProjects) endpoint to retrieve all projects to which the authenticated user has access.\n\n**NOTE**: Groups and projects are synonymous terms. Your group id is the same as your project id. For existing groups, your group/project id remains the same. The resource and corresponding endpoints use the term groups.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"cluster_id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Unique 24-hexadecimal digit string that identifies the cluster.",
			},
			"mongo_db_major_version": schema.StringAttribute{
				Computed: true,
				Optional: true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexp.MustCompile(`^([0-9]+)\.?([0-9]+)?$`), "MongoDB major version must be in the format \"8\" or \"8.0\""),
				},
				MarkdownDescription: "MongoDB major version of the cluster.\n\nOn creation: Choose from the available versions of MongoDB, or leave unspecified for the current recommended default in the MongoDB Cloud platform. The recommended version is a recent Long Term Support version. The default is not guaranteed to be the most recently released version throughout the entire release cycle. For versions available in a specific project, see the linked documentation or use the API endpoint for [project LTS versions endpoint](#tag/Projects/operation/getProjectLTSVersions).\n\n On update: Increase version only by 1 major version at a time. If the cluster is pinned to a MongoDB feature compatibility version exactly one major version below the current MongoDB version, the MongoDB version can be downgraded to the previous major version.",
			},
			"mongo_db_version": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Version of MongoDB that the cluster runs.",
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Human-readable label that identifies this cluster.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"paused": schema.BoolAttribute{
				Computed:            true,
				Optional:            true,
				MarkdownDescription: "Flag that indicates whether the cluster is paused.",
			},
			"pit_enabled": schema.BoolAttribute{
				Computed:            true,
				Optional:            true,
				MarkdownDescription: "Flag that indicates whether the cluster uses continuous cloud backups.",
			},
			"redact_client_log_data": schema.BoolAttribute{
				Computed:            true,
				Optional:            true,
				MarkdownDescription: "Enable or disable log redaction.\n\nThis setting configures the ``mongod`` or ``mongos`` to redact any document field contents from a message accompanying a given log event before logging. This prevents the program from writing potentially sensitive data stored on the database to the diagnostic log. Metadata such as error or operation codes, line numbers, and source file names are still visible in the logs.\n\nUse ``redactClientLogData`` in conjunction with Encryption at Rest and TLS/SSL (Transport Encryption) to assist compliance with regulatory requirements.\n\n*Note*: changing this setting on a cluster will trigger a rolling restart as soon as the cluster is updated.",
			},
			"replica_set_scaling_strategy": schema.StringAttribute{
				Computed:            true,
				Optional:            true,
				MarkdownDescription: "Set this field to configure the replica set scaling mode for your cluster.\n\nBy default, Atlas scales under WORKLOAD_TYPE. This mode allows Atlas to scale your analytics nodes in parallel to your operational nodes.\n\nWhen configured as SEQUENTIAL, Atlas scales all nodes sequentially. This mode is intended for steady-state workloads and applications performing latency-sensitive secondary reads.\n\nWhen configured as NODE_TYPE, Atlas scales your electable nodes in parallel with your read-only and analytics nodes. This mode is intended for large, dynamic workloads requiring frequent and timely cluster tier scaling. This is the fastest scaling strategy, but it might impact latency of workloads when performing extensive secondary reads.",
			},
			"replication_specs": schema.ListNestedAttribute{
				Required:            true,
				MarkdownDescription: "List of settings that configure your cluster regions. This array has one object per shard representing node configurations in each shard. For replica sets there is only one object representing node configurations.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							DeprecationMessage:  deprecationMsgOldSchema("id"),
							Computed:            true,
							MarkdownDescription: "Unique 24-hexadecimal digit string that identifies the replication object for a shard in a Cluster. If you include existing shard replication configurations in the request, you must specify this parameter. If you add a new shard to an existing Cluster, you may specify this parameter. The request deletes any existing shards  in the Cluster that you exclude from the request. This corresponds to Shard ID displayed in the UI.",
						},
						"container_id": schema.MapAttribute{
							ElementType:         types.StringType,
							Computed:            true,
							MarkdownDescription: "container_id", // TODO: add description
						},
						"external_id": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "external_id", // TODO: add description
						},
						"num_shards": schema.Int64Attribute{
							DeprecationMessage:  deprecationMsgOldSchema("num_shards"),
							Default:             int64default.StaticInt64(1),
							Computed:            true,
							Optional:            true,
							MarkdownDescription: "num_shards", // TODO: add description
						},
						"region_configs": schema.ListNestedAttribute{
							Required:            true,
							MarkdownDescription: "Hardware specifications for nodes set for a given region. Each **regionConfigs** object describes the region's priority in elections and the number and type of MongoDB nodes that MongoDB Cloud deploys to the region. Each **regionConfigs** object must have either an **analyticsSpecs** object, **electableSpecs** object, or **readOnlySpecs** object. Tenant clusters only require **electableSpecs. Dedicated** clusters can specify any of these specifications, but must have at least one **electableSpecs** object within a **replicationSpec**.\n\n**Example:**\n\nIf you set `\"replicationSpecs[n].regionConfigs[m].analyticsSpecs.instanceSize\" : \"M30\"`, set `\"replicationSpecs[n].regionConfigs[m].electableSpecs.instanceSize\" : `\"M30\"` if you have electable nodes and `\"replicationSpecs[n].regionConfigs[m].readOnlySpecs.instanceSize\" : `\"M30\"` if you have read-only nodes.",
							Validators: []validator.List{
								RegionSpecPriorityOrderDecreasingValidator{},
							},
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"analytics_auto_scaling": AutoScalingSchema(),
									"analytics_specs":        SpecsSchema("Hardware specifications for read-only nodes in the region. Read-only nodes can never become the primary member, but can enable local reads. If you don't specify this parameter, no read-only nodes are deployed to the region."),
									"auto_scaling":           AutoScalingSchema(),
									"backing_provider_name": schema.StringAttribute{
										Optional:            true,
										MarkdownDescription: "Cloud service provider on which MongoDB Cloud provisioned the multi-tenant cluster. The resource returns this parameter when **providerName** is `TENANT` and **electableSpecs.instanceSize** is `M0`, `M2` or `M5`.",
									},
									"electable_specs": SpecsSchema("Hardware specifications for all electable nodes deployed in the region. Electable nodes can become the primary and can enable local reads. If you don't specify this option, MongoDB Cloud deploys no electable nodes to the region."),
									"priority": schema.Int64Attribute{
										Required:            true,
										MarkdownDescription: "Precedence is given to this region when a primary election occurs. If your **regionConfigs** has only **readOnlySpecs**, **analyticsSpecs**, or both, set this value to `0`. If you have multiple **regionConfigs** objects (your cluster is multi-region or multi-cloud), they must have priorities in descending order. The highest priority is `7`.\n\n**Example:** If you have three regions, their priorities would be `7`, `6`, and `5` respectively. If you added two more regions for supporting electable nodes, the priorities of those regions would be `4` and `3` respectively.",
									},
									"provider_name": schema.StringAttribute{
										Required:            true,
										MarkdownDescription: "Cloud service provider on which MongoDB Cloud provisions the hosts. Set dedicated clusters to `AWS`, `GCP`, `AZURE` or `TENANT`.",
									},
									"read_only_specs": SpecsSchema("Hardware specifications for read-only nodes in the region. Read-only nodes can never become the primary member, but can enable local reads. If you don't specify this parameter, no read-only nodes are deployed to the region."),
									"region_name": schema.StringAttribute{
										Required:            true,
										MarkdownDescription: "Physical location of your MongoDB cluster nodes. The region you choose can affect network latency for clients accessing your databases. The region name is only returned in the response for single-region clusters. When MongoDB Cloud deploys a dedicated cluster, it checks if a VPC or VPC connection exists for that provider and region. If not, MongoDB Cloud creates them as part of the deployment. It assigns the VPC a Classless Inter-Domain Routing (CIDR) block. To limit a new VPC peering connection to one Classless Inter-Domain Routing (CIDR) block and region, create the connection first. Deploy the cluster after the connection starts. GCP Clusters and Multi-region clusters require one VPC peering connection for each region. MongoDB nodes can use only the peering connection that resides in the same region as the nodes to communicate with the peered VPC.",
									},
								},
							},
						},
						"zone_id": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Unique 24-hexadecimal digit string that identifies the zone in a Global Cluster. This value can be used to configure Global Cluster backup policies.",
						},
						"zone_name": schema.StringAttribute{
							Computed:            true,
							Optional:            true,
							MarkdownDescription: "Human-readable label that describes the zone this shard belongs to in a Global Cluster. Provide this value only if \"clusterType\" : \"GEOSHARDED\" but not \"selfManagedSharding\" : true.",
						},
					},
				},
			},
			"root_cert_type": schema.StringAttribute{
				Computed:            true,
				Optional:            true,
				MarkdownDescription: "Root Certificate Authority that MongoDB Cloud cluster uses. MongoDB Cloud supports Internet Security Research Group.",
			},
			"state_name": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Human-readable label that indicates the current operating condition of this cluster.",
			},
			"termination_protection_enabled": schema.BoolAttribute{
				Computed:            true,
				Optional:            true,
				MarkdownDescription: "Flag that indicates whether termination protection is enabled on the cluster. If set to `true`, MongoDB Cloud won't delete the cluster. If set to `false`, MongoDB Cloud will delete the cluster.",
			},
			"version_release_system": schema.StringAttribute{
				Computed:            true,
				Optional:            true,
				MarkdownDescription: "Method by which the cluster maintains the MongoDB versions. If value is `CONTINUOUS`, you must not specify **mongoDBMajorVersion**.",
			},
			"retain_backups_enabled": schema.BoolAttribute{
				Optional:            true,
				MarkdownDescription: "Flag that indicates whether to retain backup snapshots for the deleted dedicated cluster.",
			},
			"disk_size_gb": schema.Float64Attribute{
				DeprecationMessage:  deprecationMsgOldSchema("disk_size_gb"),
				Computed:            true,
				Optional:            true,
				MarkdownDescription: "Storage capacity of instance data volumes expressed in gigabytes. Increase this number to add capacity.\n\n This value must be equal for all shards and node types.\n\n This value is not configurable on M0/M2/M5 clusters.\n\n MongoDB Cloud requires this parameter if you set **replicationSpecs**.\n\n If you specify a disk size below the minimum (10 GB), this parameter defaults to the minimum disk size value. \n\n Storage charge calculations depend on whether you choose the default value or a custom value.\n\n The maximum value for disk storage cannot exceed 50 times the maximum RAM for the selected cluster. If you require more storage space, consider upgrading your cluster to a higher tier.",
			},
			"advanced_configuration": AdvancedConfigurationSchema(ctx),
			"pinned_fcv": schema.SingleNestedAttribute{
				Optional:            true,
				MarkdownDescription: "Pins the Feature Compatibility Version (FCV) to the current MongoDB version with a provided expiration date. To unpin the FCV the `pinned_fcv` attribute must be removed. This operation can take several minutes as the request processes through the MongoDB data plane. Once FCV is unpinned it will not be possible to downgrade the `mongo_db_major_version`. It is advised that updates to `pinned_fcv` are done isolated from other cluster changes. If a plan contains multiple changes, the FCV change will be applied first. If FCV is unpinned past the expiration date the `pinned_fcv` attribute must be removed. The following [knowledge hub article](https://kb.corp.mongodb.com/article/000021785/) and [FCV documentation](https://www.mongodb.com/docs/atlas/tutorial/major-version-change/#manage-feature-compatibility--fcv--during-upgrades) can be referenced for more details.",
				Attributes: map[string]schema.Attribute{
					"version": schema.StringAttribute{
						Computed:            true,
						MarkdownDescription: "Feature compatibility version of the cluster.",
					},
					"expiration_date": schema.StringAttribute{
						Required:            true,
						MarkdownDescription: "Expiration date of the fixed FCV. This value is in the ISO 8601 timestamp format (e.g. 2024-12-04T16:25:00Z). Note that this field cannot exceed 4 weeks from the pinned date.",
					},
				},
			},
			"timeouts": timeouts.Attributes(ctx, timeouts.Opts{
				Create: true,
				Update: true,
				Delete: true,
			}),
		},
		Blocks: map[string]schema.Block{
			"labels": schema.SetNestedBlock{
				MarkdownDescription: "Collection of key-value pairs between 1 to 255 characters in length that tag and categorize the cluster. The MongoDB Cloud console doesn't display your labels.\n\nCluster labels are deprecated and will be removed in a future release. We strongly recommend that you use [resource tags](https://dochub.mongodb.org/core/add-cluster-tag-atlas) instead.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"key": schema.StringAttribute{
							Required:            true,
							MarkdownDescription: "Key applied to tag and categorize this component.",
						},
						"value": schema.StringAttribute{
							Required:            true,
							MarkdownDescription: "Value set to the Key applied to tag and categorize this component.",
						},
					},
				},
			},
			"tags": schema.SetNestedBlock{
				MarkdownDescription: "List that contains key-value pairs between 1 to 255 characters in length for tagging and categorizing the cluster.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"key": schema.StringAttribute{
							Required:            true,
							MarkdownDescription: "Constant that defines the set of the tag. For example, `environment` in the `environment : production` tag.",
						},
						"value": schema.StringAttribute{
							Required:            true,
							MarkdownDescription: "Variable that belongs to the set of the tag. For example, `production` in the `environment : production` tag.",
						},
					},
				},
			},
		},
	}
}

func dataSourceSchema(ctx context.Context) dsschema.Schema {
	return conversion.DataSourceSchemaFromResource(resourceSchema(ctx), &conversion.DataSourceSchemaRequest{
		RequiredFields:  []string{"project_id", "name"},
		OverridenFields: dataSourceOverridenFields(),
	})
}

func pluralDataSourceSchema(ctx context.Context) dsschema.Schema {
	return conversion.PluralDataSourceSchemaFromResource(resourceSchema(ctx), &conversion.PluralDataSourceSchemaRequest{
		RequiredFields: []string{"project_id"},
		OverridenRootFields: map[string]dsschema.Attribute{
			"use_replication_spec_per_shard": dsschema.BoolAttribute{ // TODO: added as in current resource
				Optional:            true,
				MarkdownDescription: "use_replication_spec_per_shard", // TODO: add documentation
			},
			"include_deleted_with_retained_backups": dsschema.BoolAttribute{ // TODO: not in current resource, decide if keep
				Optional:            true,
				MarkdownDescription: "Flag that indicates whether to return Clusters with retain backups.",
			},
		},
		OverridenFields: dataSourceOverridenFields(),
	})
}

func dataSourceOverridenFields() map[string]dsschema.Attribute {
	return map[string]dsschema.Attribute{
		"use_replication_spec_per_shard": dsschema.BoolAttribute{ // TODO: added as in current resource
			Optional:            true,
			MarkdownDescription: "use_replication_spec_per_shard", // TODO: add documentation
		},
		"accept_data_risks_and_force_replica_set_reconfig": nil,
	}
}

func AutoScalingSchema() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		Computed:            true,
		Optional:            true,
		MarkdownDescription: "Options that determine how this cluster handles resource scaling.",
		Attributes: map[string]schema.Attribute{
			"compute_enabled": schema.BoolAttribute{
				Computed:            true,
				Optional:            true,
				MarkdownDescription: "Flag that indicates whether someone enabled instance size auto-scaling.\n\n- Set to `true` to enable instance size auto-scaling. If enabled, you must specify a value for **replicationSpecs[n].regionConfigs[m].autoScaling.compute.maxInstanceSize**.\n- Set to `false` to disable instance size automatic scaling.",
			},
			"compute_max_instance_size": schema.StringAttribute{
				Computed:            true,
				Optional:            true,
				MarkdownDescription: "Minimum instance size to which your cluster can automatically scale. MongoDB Cloud requires this parameter if `\"replicationSpecs[n].regionConfigs[m].autoScaling.compute.scaleDownEnabled\" : true`.",
			},
			"compute_min_instance_size": schema.StringAttribute{
				Computed:            true,
				Optional:            true,
				MarkdownDescription: "Minimum instance size to which your cluster can automatically scale. MongoDB Cloud requires this parameter if `\"replicationSpecs[n].regionConfigs[m].autoScaling.compute.scaleDownEnabled\" : true`.",
			},
			"compute_scale_down_enabled": schema.BoolAttribute{
				Computed:            true,
				Optional:            true,
				MarkdownDescription: "Flag that indicates whether the instance size may scale down. MongoDB Cloud requires this parameter if `\"replicationSpecs[n].regionConfigs[m].autoScaling.compute.enabled\" : true`. If you enable this option, specify a value for **replicationSpecs[n].regionConfigs[m].autoScaling.compute.minInstanceSize**.",
			},
			"disk_gb_enabled": schema.BoolAttribute{
				Computed:            true,
				Optional:            true,
				MarkdownDescription: "Flag that indicates whether this cluster enables disk auto-scaling. The maximum memory allowed for the selected cluster tier and the oplog size can limit storage auto-scaling.",
			},
		},
	}
}

func SpecsSchema(markdownDescription string) schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		Computed:            true,
		Optional:            true,
		MarkdownDescription: markdownDescription,
		Attributes: map[string]schema.Attribute{
			"disk_iops": schema.Int64Attribute{
				Computed:            true,
				Optional:            true,
				MarkdownDescription: "Target throughput desired for storage attached to your Azure-provisioned cluster. Change this parameter if you:\n\n- set `\"replicationSpecs[n].regionConfigs[m].providerName\" : \"Azure\"`.\n- set `\"replicationSpecs[n].regionConfigs[m].electableSpecs.instanceSize\" : \"M40\"` or greater not including `Mxx_NVME` tiers.\n\nThe maximum input/output operations per second (IOPS) depend on the selected **.instanceSize** and **.diskSizeGB**.\nThis parameter defaults to the cluster tier's standard IOPS value.\nChanging this value impacts cluster cost.",
			},
			"disk_size_gb": schema.Float64Attribute{
				Computed:            true,
				Optional:            true,
				MarkdownDescription: "Storage capacity of instance data volumes expressed in gigabytes. Increase this number to add capacity.\n\n This value must be equal for all shards and node types.\n\n This value is not configurable on M0/M2/M5 clusters.\n\n MongoDB Cloud requires this parameter if you set **replicationSpecs**.\n\n If you specify a disk size below the minimum (10 GB), this parameter defaults to the minimum disk size value. \n\n Storage charge calculations depend on whether you choose the default value or a custom value.\n\n The maximum value for disk storage cannot exceed 50 times the maximum RAM for the selected cluster. If you require more storage space, consider upgrading your cluster to a higher tier.",
			},
			"ebs_volume_type": schema.StringAttribute{
				Computed:            true,
				Optional:            true,
				MarkdownDescription: "Type of storage you want to attach to your AWS-provisioned cluster.\n\n- `STANDARD` volume types can't exceed the default input/output operations per second (IOPS) rate for the selected volume size. \n\n- `PROVISIONED` volume types must fall within the allowable IOPS range for the selected volume size. You must set this value to (`PROVISIONED`) for NVMe clusters.",
			},
			"instance_size": schema.StringAttribute{
				Computed:            true,
				Optional:            true,
				MarkdownDescription: "Hardware specification for the instance sizes in this region in this shard. Each instance size has a default storage and memory capacity. Electable nodes and read-only nodes (known as \"base nodes\") within a single shard must use the same instance size. Analytics nodes can scale independently from base nodes within a shard. Both base nodes and analytics nodes can scale independently from their equivalents in other shards.",
			},
			"node_count": schema.Int64Attribute{
				Computed:            true,
				Optional:            true,
				MarkdownDescription: "Number of nodes of the given type for MongoDB Cloud to deploy to the region.",
			},
		},
	}
}

func AdvancedConfigurationSchema(ctx context.Context) schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		Computed:            true,
		Optional:            true,
		MarkdownDescription: "advanced_configuration", // TODO: add description
		Attributes: map[string]schema.Attribute{
			"change_stream_options_pre_and_post_images_expire_after_seconds": schema.Int64Attribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "The minimum pre- and post-image retention time in seconds.",
				Default:             int64default.StaticInt64(-1), // in case the user removes the value, we should set it to -1, a special value used by the backend to use its default behavior
				PlanModifiers: []planmodifier.Int64{
					PlanMustUseMongoDBVersion(7.0, EqualOrHigher),
				},
			},
			"default_write_concern": schema.StringAttribute{
				Computed:            true,
				Optional:            true,
				MarkdownDescription: "Default level of acknowledgment requested from MongoDB for write operations when none is specified by the driver.",
			},
			"javascript_enabled": schema.BoolAttribute{
				Computed:            true,
				Optional:            true,
				MarkdownDescription: "Flag that indicates whether the cluster allows execution of operations that perform server-side executions of JavaScript. When using 8.0+, we recommend disabling server-side JavaScript and using operators of aggregation pipeline as more performant alternative.",
			},
			"minimum_enabled_tls_protocol": schema.StringAttribute{
				Computed:            true,
				Optional:            true,
				MarkdownDescription: "Minimum Transport Layer Security (TLS) version that the cluster accepts for incoming connections. Clusters using TLS 1.0 or 1.1 should consider setting TLS 1.2 as the minimum TLS protocol version.",
			},
			"no_table_scan": schema.BoolAttribute{
				Computed:            true,
				Optional:            true,
				MarkdownDescription: "Flag that indicates whether the cluster disables executing any query that requires a collection scan to return results.",
			},
			"oplog_min_retention_hours": schema.Float64Attribute{
				Computed:            true,
				Optional:            true,
				MarkdownDescription: "Minimum retention window for cluster's oplog expressed in hours. A value of null indicates that the cluster uses the default minimum oplog window that MongoDB Cloud calculates.",
			},
			"oplog_size_mb": schema.Int64Attribute{
				Optional: true,
				Computed: true,
				Validators: []validator.Int64{
					int64validator.AtLeast(0),
				},
				MarkdownDescription: "Storage limit of cluster's oplog expressed in megabytes. A value of null indicates that the cluster uses the default oplog size that MongoDB Cloud calculates.",
			},
			"sample_refresh_interval_bi_connector": schema.Int64Attribute{
				Computed:            true,
				Optional:            true,
				MarkdownDescription: "Interval in seconds at which the mongosqld process re-samples data to create its relational schema.",
			},
			"sample_size_bi_connector": schema.Int64Attribute{
				Computed:            true,
				Optional:            true,
				MarkdownDescription: "Number of documents per database to sample when gathering schema information.",
			},
			"transaction_lifetime_limit_seconds": schema.Int64Attribute{
				Computed:            true,
				Optional:            true,
				MarkdownDescription: "Lifetime, in seconds, of multi-document transactions. Atlas considers the transactions that exceed this limit as expired and so aborts them through a periodic cleanup process.",
			},
			"default_read_concern": schema.StringAttribute{
				DeprecationMessage: deprecationMsgOldSchema("default_read_concern"),

				Computed:            true,
				Optional:            true,
				MarkdownDescription: "default_read_concern", // TODO: add description
			},
			"fail_index_key_too_long": schema.BoolAttribute{
				DeprecationMessage:  deprecationMsgOldSchema("fail_index_key_too_long"),
				Computed:            true,
				Optional:            true,
				MarkdownDescription: "fail_index_key_too_long", // TODO: add description
			},
			"default_max_time_ms": schema.Int64Attribute{
				Optional:            true,
				MarkdownDescription: "Default time limit in milliseconds for individual read operations to complete. This parameter is supported only for MongoDB version 8.0 and above.",
				PlanModifiers: []planmodifier.Int64{
					PlanMustUseMongoDBVersion(8.0, EqualOrHigher),
				},
			},
			"custom_openssl_cipher_config_tls12": schema.SetAttribute{
				Optional:            true,
				ElementType:         types.StringType,
				MarkdownDescription: "The custom OpenSSL cipher suite list for TLS 1.2. This field is only valid when `tls_cipher_config_mode` is set to `CUSTOM`.",
			},
			"tls_cipher_config_mode": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "The TLS cipher suite configuration mode. Valid values include `CUSTOM` or `DEFAULT`. The `DEFAULT` mode uses the default cipher suites. The `CUSTOM` mode allows you to specify custom cipher suites for both TLS 1.2 and TLS 1.3. To unset, this should be set back to `DEFAULT`.",
			},
		},
	}
}

type TFModel struct {
	DiskSizeGB                                types.Float64  `tfsdk:"disk_size_gb"`
	Labels                                    types.Set      `tfsdk:"labels"`
	ReplicationSpecs                          types.List     `tfsdk:"replication_specs"`
	Tags                                      types.Set      `tfsdk:"tags"`
	StateName                                 types.String   `tfsdk:"state_name"`
	ConnectionStrings                         types.Object   `tfsdk:"connection_strings"`
	CreateDate                                types.String   `tfsdk:"create_date"`
	AcceptDataRisksAndForceReplicaSetReconfig types.String   `tfsdk:"accept_data_risks_and_force_replica_set_reconfig"`
	EncryptionAtRestProvider                  types.String   `tfsdk:"encryption_at_rest_provider"`
	Timeouts                                  timeouts.Value `tfsdk:"timeouts"`
	ProjectID                                 types.String   `tfsdk:"project_id"`
	ClusterID                                 types.String   `tfsdk:"cluster_id"`
	ConfigServerManagementMode                types.String   `tfsdk:"config_server_management_mode"`
	MongoDBMajorVersion                       types.String   `tfsdk:"mongo_db_major_version"`
	MongoDBVersion                            types.String   `tfsdk:"mongo_db_version"`
	Name                                      types.String   `tfsdk:"name"`
	VersionReleaseSystem                      types.String   `tfsdk:"version_release_system"`
	BiConnectorConfig                         types.Object   `tfsdk:"bi_connector_config"`
	ConfigServerType                          types.String   `tfsdk:"config_server_type"`
	ReplicaSetScalingStrategy                 types.String   `tfsdk:"replica_set_scaling_strategy"`
	ClusterType                               types.String   `tfsdk:"cluster_type"`
	RootCertType                              types.String   `tfsdk:"root_cert_type"`
	AdvancedConfiguration                     types.Object   `tfsdk:"advanced_configuration"`
	PinnedFCV                                 types.Object   `tfsdk:"pinned_fcv"`
	TerminationProtectionEnabled              types.Bool     `tfsdk:"termination_protection_enabled"`
	Paused                                    types.Bool     `tfsdk:"paused"`
	RetainBackupsEnabled                      types.Bool     `tfsdk:"retain_backups_enabled"`
	BackupEnabled                             types.Bool     `tfsdk:"backup_enabled"`
	GlobalClusterSelfManagedSharding          types.Bool     `tfsdk:"global_cluster_self_managed_sharding"`
	RedactClientLogData                       types.Bool     `tfsdk:"redact_client_log_data"`
	PitEnabled                                types.Bool     `tfsdk:"pit_enabled"`
}

// TFModelDS differs from TFModel: removes timeouts, accept_data_risks_and_force_replica_set_reconfig; adds use_replication_spec_per_shard.
type TFModelDS struct {
	DiskSizeGB                       types.Float64 `tfsdk:"disk_size_gb"`
	Labels                           types.Set     `tfsdk:"labels"`
	ReplicationSpecs                 types.List    `tfsdk:"replication_specs"`
	Tags                             types.Set     `tfsdk:"tags"`
	ReplicaSetScalingStrategy        types.String  `tfsdk:"replica_set_scaling_strategy"`
	Name                             types.String  `tfsdk:"name"`
	AdvancedConfiguration            types.Object  `tfsdk:"advanced_configuration"`
	BiConnectorConfig                types.Object  `tfsdk:"bi_connector_config"`
	RootCertType                     types.String  `tfsdk:"root_cert_type"`
	ClusterType                      types.String  `tfsdk:"cluster_type"`
	MongoDBMajorVersion              types.String  `tfsdk:"mongo_db_major_version"`
	ConfigServerType                 types.String  `tfsdk:"config_server_type"`
	VersionReleaseSystem             types.String  `tfsdk:"version_release_system"`
	ConnectionStrings                types.Object  `tfsdk:"connection_strings"`
	StateName                        types.String  `tfsdk:"state_name"`
	MongoDBVersion                   types.String  `tfsdk:"mongo_db_version"`
	CreateDate                       types.String  `tfsdk:"create_date"`
	EncryptionAtRestProvider         types.String  `tfsdk:"encryption_at_rest_provider"`
	ProjectID                        types.String  `tfsdk:"project_id"`
	ClusterID                        types.String  `tfsdk:"cluster_id"`
	ConfigServerManagementMode       types.String  `tfsdk:"config_server_management_mode"`
	PinnedFCV                        types.Object  `tfsdk:"pinned_fcv"`
	UseReplicationSpecPerShard       types.Bool    `tfsdk:"use_replication_spec_per_shard"`
	RedactClientLogData              types.Bool    `tfsdk:"redact_client_log_data"`
	GlobalClusterSelfManagedSharding types.Bool    `tfsdk:"global_cluster_self_managed_sharding"`
	BackupEnabled                    types.Bool    `tfsdk:"backup_enabled"`
	RetainBackupsEnabled             types.Bool    `tfsdk:"retain_backups_enabled"`
	Paused                           types.Bool    `tfsdk:"paused"`
	TerminationProtectionEnabled     types.Bool    `tfsdk:"termination_protection_enabled"`
	PitEnabled                       types.Bool    `tfsdk:"pit_enabled"`
}

type TFModelPluralDS struct {
	ProjectID                         types.String `tfsdk:"project_id"`
	Results                           []*TFModelDS `tfsdk:"results"`
	UseReplicationSpecPerShard        types.Bool   `tfsdk:"use_replication_spec_per_shard"`        // TODO: added as in current resource
	IncludeDeletedWithRetainedBackups types.Bool   `tfsdk:"include_deleted_with_retained_backups"` // TODO: not in current resource, decide if keep
}

type TFBiConnectorModel struct {
	ReadPreference types.String `tfsdk:"read_preference"`
	Enabled        types.Bool   `tfsdk:"enabled"`
}

var BiConnectorConfigObjType = types.ObjectType{AttrTypes: map[string]attr.Type{
	"enabled":         types.BoolType,
	"read_preference": types.StringType,
}}

type TFConnectionStringsModel struct {
	Private         types.String `tfsdk:"private"`
	PrivateEndpoint types.List   `tfsdk:"private_endpoint"`
	PrivateSrv      types.String `tfsdk:"private_srv"`
	Standard        types.String `tfsdk:"standard"`
	StandardSrv     types.String `tfsdk:"standard_srv"`
}

var ConnectionStringsObjType = types.ObjectType{AttrTypes: map[string]attr.Type{
	"private":          types.StringType,
	"private_endpoint": types.ListType{ElemType: PrivateEndpointObjType},
	"private_srv":      types.StringType,
	"standard":         types.StringType,
	"standard_srv":     types.StringType,
}}

type TFPrivateEndpointModel struct {
	ConnectionString                  types.String `tfsdk:"connection_string"`
	Endpoints                         types.List   `tfsdk:"endpoints"`
	SrvConnectionString               types.String `tfsdk:"srv_connection_string"`
	SrvShardOptimizedConnectionString types.String `tfsdk:"srv_shard_optimized_connection_string"`
	Type                              types.String `tfsdk:"type"`
}

var PrivateEndpointObjType = types.ObjectType{AttrTypes: map[string]attr.Type{
	"connection_string":                     types.StringType,
	"endpoints":                             types.ListType{ElemType: EndpointsObjType},
	"srv_connection_string":                 types.StringType,
	"srv_shard_optimized_connection_string": types.StringType,
	"type":                                  types.StringType,
}}

type TFEndpointsModel struct {
	EndpointId   types.String `tfsdk:"endpoint_id"`
	ProviderName types.String `tfsdk:"provider_name"`
	Region       types.String `tfsdk:"region"`
}

var EndpointsObjType = types.ObjectType{AttrTypes: map[string]attr.Type{
	"endpoint_id":   types.StringType,
	"provider_name": types.StringType,
	"region":        types.StringType,
}}

type TFLabelsModel struct {
	Key   types.String `tfsdk:"key"`
	Value types.String `tfsdk:"value"`
}

var LabelsObjType = types.ObjectType{AttrTypes: map[string]attr.Type{
	"key":   types.StringType,
	"value": types.StringType,
}}

type TFReplicationSpecsModel struct {
	RegionConfigs types.List   `tfsdk:"region_configs"`
	ContainerId   types.Map    `tfsdk:"container_id"`
	Id            types.String `tfsdk:"id"`
	ExternalId    types.String `tfsdk:"external_id"`
	ZoneId        types.String `tfsdk:"zone_id"`
	ZoneName      types.String `tfsdk:"zone_name"`
	NumShards     types.Int64  `tfsdk:"num_shards"`
}

var ReplicationSpecsObjType = types.ObjectType{AttrTypes: map[string]attr.Type{
	"id":             types.StringType,
	"container_id":   types.MapType{ElemType: types.StringType},
	"external_id":    types.StringType,
	"num_shards":     types.Int64Type,
	"region_configs": types.ListType{ElemType: RegionConfigsObjType},
	"zone_id":        types.StringType,
	"zone_name":      types.StringType,
}}

type TFRegionConfigsModel struct {
	AnalyticsAutoScaling types.Object `tfsdk:"analytics_auto_scaling"`
	AnalyticsSpecs       types.Object `tfsdk:"analytics_specs"`
	AutoScaling          types.Object `tfsdk:"auto_scaling"`
	BackingProviderName  types.String `tfsdk:"backing_provider_name"`
	ElectableSpecs       types.Object `tfsdk:"electable_specs"`
	ProviderName         types.String `tfsdk:"provider_name"`
	ReadOnlySpecs        types.Object `tfsdk:"read_only_specs"`
	RegionName           types.String `tfsdk:"region_name"`
	Priority             types.Int64  `tfsdk:"priority"`
}

var RegionConfigsObjType = types.ObjectType{AttrTypes: map[string]attr.Type{
	"analytics_auto_scaling": AutoScalingObjType,
	"analytics_specs":        SpecsObjType,
	"auto_scaling":           AutoScalingObjType,
	"backing_provider_name":  types.StringType,
	"electable_specs":        SpecsObjType,
	"priority":               types.Int64Type,
	"provider_name":          types.StringType,
	"read_only_specs":        SpecsObjType,
	"region_name":            types.StringType,
}}

type TFAutoScalingModel struct {
	ComputeMaxInstanceSize  types.String `tfsdk:"compute_max_instance_size"`
	ComputeMinInstanceSize  types.String `tfsdk:"compute_min_instance_size"`
	ComputeEnabled          types.Bool   `tfsdk:"compute_enabled"`
	ComputeScaleDownEnabled types.Bool   `tfsdk:"compute_scale_down_enabled"`
	DiskGBEnabled           types.Bool   `tfsdk:"disk_gb_enabled"`
}

var AutoScalingObjType = types.ObjectType{AttrTypes: map[string]attr.Type{
	"compute_enabled":            types.BoolType,
	"compute_max_instance_size":  types.StringType,
	"compute_min_instance_size":  types.StringType,
	"compute_scale_down_enabled": types.BoolType,
	"disk_gb_enabled":            types.BoolType,
}}

type TFSpecsModel struct {
	DiskSizeGb    types.Float64 `tfsdk:"disk_size_gb"`
	EbsVolumeType types.String  `tfsdk:"ebs_volume_type"`
	InstanceSize  types.String  `tfsdk:"instance_size"`
	DiskIops      types.Int64   `tfsdk:"disk_iops"`
	NodeCount     types.Int64   `tfsdk:"node_count"`
}

var SpecsObjType = types.ObjectType{AttrTypes: map[string]attr.Type{
	"disk_iops":       types.Int64Type,
	"disk_size_gb":    types.Float64Type,
	"ebs_volume_type": types.StringType,
	"instance_size":   types.StringType,
	"node_count":      types.Int64Type,
}}

type TFTagsModel struct {
	Key   types.String `tfsdk:"key"`
	Value types.String `tfsdk:"value"`
}

var TagsObjType = types.ObjectType{AttrTypes: map[string]attr.Type{
	"key":   types.StringType,
	"value": types.StringType,
}}

type TFAdvancedConfigurationModel struct {
	OplogMinRetentionHours                                types.Float64 `tfsdk:"oplog_min_retention_hours"`
	CustomOpensslCipherConfigTls12                        types.Set     `tfsdk:"custom_openssl_cipher_config_tls12"`
	MinimumEnabledTlsProtocol                             types.String  `tfsdk:"minimum_enabled_tls_protocol"`
	DefaultWriteConcern                                   types.String  `tfsdk:"default_write_concern"`
	DefaultReadConcern                                    types.String  `tfsdk:"default_read_concern"`
	TlsCipherConfigMode                                   types.String  `tfsdk:"tls_cipher_config_mode"`
	SampleRefreshIntervalBiconnector                      types.Int64   `tfsdk:"sample_refresh_interval_bi_connector"`
	SampleSizeBiconnector                                 types.Int64   `tfsdk:"sample_size_bi_connector"`
	TransactionLifetimeLimitSeconds                       types.Int64   `tfsdk:"transaction_lifetime_limit_seconds"`
	DefaultMaxTimeMS                                      types.Int64   `tfsdk:"default_max_time_ms"`
	OplogSizeMb                                           types.Int64   `tfsdk:"oplog_size_mb"`
	ChangeStreamOptionsPreAndPostImagesExpireAfterSeconds types.Int64   `tfsdk:"change_stream_options_pre_and_post_images_expire_after_seconds"`
	JavascriptEnabled                                     types.Bool    `tfsdk:"javascript_enabled"`
	NoTableScan                                           types.Bool    `tfsdk:"no_table_scan"`
	FailIndexKeyTooLong                                   types.Bool    `tfsdk:"fail_index_key_too_long"`
}

var AdvancedConfigurationObjType = types.ObjectType{AttrTypes: map[string]attr.Type{
	"change_stream_options_pre_and_post_images_expire_after_seconds": types.Int64Type,
	"default_read_concern":                 types.StringType,
	"default_write_concern":                types.StringType,
	"fail_index_key_too_long":              types.BoolType,
	"javascript_enabled":                   types.BoolType,
	"minimum_enabled_tls_protocol":         types.StringType,
	"no_table_scan":                        types.BoolType,
	"oplog_min_retention_hours":            types.Float64Type,
	"oplog_size_mb":                        types.Int64Type,
	"sample_refresh_interval_bi_connector": types.Int64Type,
	"sample_size_bi_connector":             types.Int64Type,
	"transaction_lifetime_limit_seconds":   types.Int64Type,
	"default_max_time_ms":                  types.Int64Type,
	"tls_cipher_config_mode":               types.StringType,
	"custom_openssl_cipher_config_tls12":   types.SetType{ElemType: types.StringType},
}}

type TFPinnedFCVModel struct {
	Version        types.String `tfsdk:"version"`
	ExpirationDate types.String `tfsdk:"expiration_date"`
}

var PinnedFCVObjType = types.ObjectType{AttrTypes: map[string]attr.Type{
	"version":         types.StringType,
	"expiration_date": types.StringType,
}}
