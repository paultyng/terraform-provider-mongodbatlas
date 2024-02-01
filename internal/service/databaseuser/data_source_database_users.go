package databaseuser

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/mongodb/terraform-provider-mongodbatlas/internal/common/constant"
	"github.com/mongodb/terraform-provider-mongodbatlas/internal/common/conversion"
	"github.com/mongodb/terraform-provider-mongodbatlas/internal/common/dsschema"
	"github.com/mongodb/terraform-provider-mongodbatlas/internal/config"
	"go.mongodb.org/atlas-sdk/v20231115005/admin"
)

const (
	databaseUsersDSName = "database_users"
)

type DatabaseUsersDS struct {
	config.DSCommon
}

func PluralDataSource() datasource.DataSource {
	return &DatabaseUsersDS{
		DSCommon: config.DSCommon{
			DataSourceName: databaseUsersDSName,
		},
	}
}

var _ datasource.DataSource = &DatabaseUsersDS{}
var _ datasource.DataSourceWithConfigure = &DatabaseUsersDS{}

type TfDatabaseUsersDSModel struct {
	ID           types.String             `tfsdk:"id"`
	ProjectID    types.String             `tfsdk:"project_id"`
	Results      []*TfDatabaseUserDSModel `tfsdk:"results"`
	PageNum      types.Int64              `tfsdk:"page_num"`
	ItemsPerPage types.Int64              `tfsdk:"items_per_page"`
	TotalCount   types.Int64              `tfsdk:"total_count"`
}

func (d *DatabaseUsersDS) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = dsschema.PaginatedDSSchema(
		map[string]schema.Attribute{
			"project_id": schema.StringAttribute{
				Required: true,
			},
		},
		map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"project_id": schema.StringAttribute{
				Computed: true,
			},
			"auth_database_name": schema.StringAttribute{
				Computed: true,
			},
			"username": schema.StringAttribute{
				Computed: true,
			},
			"password": schema.StringAttribute{
				Computed:           true,
				Sensitive:          true,
				DeprecationMessage: fmt.Sprintf(constant.DeprecationParamByVersion, "1.16.0"),
			},
			"x509_type": schema.StringAttribute{
				Computed: true,
			},
			"oidc_auth_type": schema.StringAttribute{
				Computed: true,
			},
			"ldap_auth_type": schema.StringAttribute{
				Computed: true,
			},
			"aws_iam_type": schema.StringAttribute{
				Computed: true,
			},
			"roles": schema.SetNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"collection_name": schema.StringAttribute{
							Computed: true,
						},
						"database_name": schema.StringAttribute{
							Computed: true,
						},
						"role_name": schema.StringAttribute{
							Computed: true,
						},
					},
				},
			},
			"labels": schema.SetNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"key": schema.StringAttribute{
							Computed: true,
						},
						"value": schema.StringAttribute{
							Computed: true,
						},
					},
				},
			},
			"scopes": schema.SetNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Computed: true,
						},
						"type": schema.StringAttribute{
							Computed: true,
						},
					},
				},
			},
		})
}

func (d *DatabaseUsersDS) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var databaseUsersModel *TfDatabaseUsersDSModel
	var err error
	resp.Diagnostics.Append(req.Config.Get(ctx, &databaseUsersModel)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := databaseUsersModel.ProjectID.ValueString()
	itemsPerPage := databaseUsersModel.ItemsPerPage.ValueInt64Pointer()
	pageNum := databaseUsersModel.PageNum.ValueInt64Pointer()
	connV2 := d.Client.AtlasV2
	paginatedResp, _, err := connV2.DatabaseUsersApi.ListDatabaseUsersWithParams(ctx, &admin.ListDatabaseUsersApiParams{
		GroupId:      projectID,
		ItemsPerPage: conversion.Int64PtrToIntPtr(itemsPerPage),
		PageNum:      conversion.Int64PtrToIntPtr(pageNum),
	}).Execute()
	if err != nil {
		resp.Diagnostics.AddError("error getting database user information", err.Error())
		return
	}

	dbUserModel, diagnostic := NewTFDatabaseUsersModel(ctx, databaseUsersModel, paginatedResp)
	resp.Diagnostics.Append(diagnostic...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &dbUserModel)...)
	if resp.Diagnostics.HasError() {
		return
	}
}
