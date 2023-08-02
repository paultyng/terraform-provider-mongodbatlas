package mongodbatlas

import (
	"context"
	"errors"
	"log"
	"strings"

	"go.mongodb.org/atlas-sdk/v20230201002/admin"
	matlas "go.mongodb.org/atlas/mongodbatlas"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

type apiKey struct {
	id    string
	roles []string
}

func dataSourceMongoDBAtlasProject() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceMongoDBAtlasProjectRead,
		Schema: map[string]*schema.Schema{
			"project_id": {
				Type:          schema.TypeString,
				Optional:      true,
				ConflictsWith: []string{"name"},
			},
			"name": {
				Type:          schema.TypeString,
				Optional:      true,
				ConflictsWith: []string{"project_id"},
			},
			"org_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"cluster_count": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"created": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"teams": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"team_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"role_names": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
					},
				},
			},
			"api_keys": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"api_key_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"role_names": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
					},
				},
			},
			"is_collect_database_specifics_statistics_enabled": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"is_data_explorer_enabled": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"is_extended_storage_sizes_enabled": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"is_performance_advisor_enabled": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"is_realtime_performance_panel_enabled": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"is_schema_advisor_enabled": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"region_usage_restrictions": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"limits": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"value": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"current_usage": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"default_limit": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"maximum_limit": {
							Type:     schema.TypeInt,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func getProjectAPIKeys(ctx context.Context, conn *matlas.Client, orgID, projectID string) ([]*apiKey, error) {
	apiKeys, _, err := conn.ProjectAPIKeys.List(ctx, projectID, &matlas.ListOptions{})
	if err != nil {
		return nil, err
	}
	var keys []*apiKey
	for _, key := range apiKeys {
		id := key.ID

		var roles []string
		for _, role := range key.Roles {
			// ProjectAPIKeys.List returns all API keys of the Project, including the org and project roles
			// For more details: https://docs.atlas.mongodb.com/reference/api/projectApiKeys/get-all-apiKeys-in-one-project/
			if !strings.HasPrefix(role.RoleName, "ORG_") && role.GroupID == projectID {
				roles = append(roles, role.RoleName)
			}
		}
		keys = append(keys, &apiKey{id, roles})
	}

	return keys, nil
}

func dataSourceMongoDBAtlasProjectRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// Get client connection.
	conn := meta.(*MongoDBClient).Atlas
	connV2 := meta.(*MongoDBClient).AtlasV2

	projectID, projectIDOk := d.GetOk("project_id")
	name, nameOk := d.GetOk("name")

	if !projectIDOk && !nameOk {
		return diag.FromErr(errors.New("either project_id or name must be configured"))
	}

	var (
		err     error
		project *matlas.Project
	)

	if projectIDOk {
		project, _, err = conn.Projects.GetOneProject(ctx, projectID.(string))
	} else {
		project, _, err = conn.Projects.GetOneProjectByName(ctx, name.(string))
	}

	if err != nil {
		return diag.Errorf(errorProjectRead, projectID, err)
	}

	teams, _, err := conn.Projects.GetProjectTeamsAssigned(ctx, project.ID)
	if err != nil {
		return diag.Errorf("error getting project's teams assigned (%s): %s", projectID, err)
	}

	apiKeys, err := getProjectAPIKeys(ctx, conn, project.OrgID, project.ID)
	if err != nil {
		var target *matlas.ErrorResponse
		if errors.As(err, &target) && target.ErrorCode != "USER_UNAUTHORIZED" {
			return diag.Errorf("error getting project's api keys (%s): %s", projectID, err)
		}
		log.Println("[WARN] `api_keys` will be empty because the user has no permissions to read the api keys endpoint")
	}

	limits, _, err := connV2.ProjectsApi.ListProjectLimits(ctx, project.ID).Execute()
	if err != nil {
		return diag.Errorf("error getting project's limits (%s): %s", projectID, err)
	}

	projectSettings, _, err := conn.Projects.GetProjectSettings(ctx, project.ID)
	if err != nil {
		return diag.Errorf("error getting project's settings assigned (%s): %s", projectID, err)
	}

	if err := d.Set("name", project.Name); err != nil {
		return diag.Errorf(errorProjectSetting, `name`, project.Name, err)
	}

	if err := d.Set("org_id", project.OrgID); err != nil {
		return diag.Errorf(errorProjectSetting, `org_id`, project.ID, err)
	}

	if err := d.Set("cluster_count", project.ClusterCount); err != nil {
		return diag.Errorf(errorProjectSetting, `clusterCount`, project.ID, err)
	}

	if err := d.Set("created", project.Created); err != nil {
		return diag.Errorf(errorProjectSetting, `created`, project.ID, err)
	}

	if err := d.Set("teams", flattenTeams(teams)); err != nil {
		return diag.Errorf(errorProjectSetting, `teams`, project.ID, err)
	}

	if err := d.Set("api_keys", flattenAPIKeys(apiKeys)); err != nil {
		return diag.Errorf(errorProjectSetting, `api_keys`, project.ID, err)
	}

	if err := d.Set("limits", flattenLimits(limits)); err != nil {
		return diag.Errorf(errorProjectSetting, `limits`, projectID, err)
	}

	if err := d.Set("is_collect_database_specifics_statistics_enabled", projectSettings.IsCollectDatabaseSpecificsStatisticsEnabled); err != nil {
		return diag.Errorf(errorProjectSetting, `is_collect_database_specifics_statistics_enabled`, project.ID, err)
	}
	if err := d.Set("is_data_explorer_enabled", projectSettings.IsDataExplorerEnabled); err != nil {
		return diag.Errorf(errorProjectSetting, `is_data_explorer_enabled`, project.ID, err)
	}
	if err := d.Set("is_extended_storage_sizes_enabled", projectSettings.IsExtendedStorageSizesEnabled); err != nil {
		return diag.Errorf(errorProjectSetting, `is_extended_storage_sizes_enabled`, project.ID, err)
	}
	if err := d.Set("is_performance_advisor_enabled", projectSettings.IsPerformanceAdvisorEnabled); err != nil {
		return diag.Errorf(errorProjectSetting, `is_performance_advisor_enabled`, project.ID, err)
	}
	if err := d.Set("is_realtime_performance_panel_enabled", projectSettings.IsRealtimePerformancePanelEnabled); err != nil {
		return diag.Errorf(errorProjectSetting, `is_realtime_performance_panel_enabled`, project.ID, err)
	}
	if err := d.Set("is_schema_advisor_enabled", projectSettings.IsSchemaAdvisorEnabled); err != nil {
		return diag.Errorf(errorProjectSetting, `is_schema_advisor_enabled`, project.ID, err)
	}
	if err := d.Set("region_usage_restrictions", project.RegionUsageRestrictions); err != nil {
		return diag.Errorf(errorProjectSetting, `region_usage_restrictions`, project.ID, err)
	}
	d.SetId(project.ID)

	return nil
}

func flattenTeams(ta *matlas.TeamsAssigned) []map[string]interface{} {
	teams := ta.Results
	res := make([]map[string]interface{}, len(teams))

	for i, team := range teams {
		res[i] = map[string]interface{}{
			"team_id":    team.TeamID,
			"role_names": team.RoleNames,
		}
	}

	return res
}

func flattenAPIKeys(keys []*apiKey) []map[string]interface{} {
	res := make([]map[string]interface{}, len(keys))

	for i, key := range keys {
		res[i] = map[string]interface{}{
			"api_key_id": key.id,
			"role_names": key.roles,
		}
	}

	return res
}

func flattenLimits(limits []admin.DataFederationLimit) []map[string]interface{} {
	res := make([]map[string]interface{}, len(limits))

	for i, limit := range limits {
		res[i] = map[string]interface{}{
			"name":  limit.Name,
			"value": limit.Value,
		}
		if limit.CurrentUsage != nil {
			res[i]["current_usage"] = *limit.CurrentUsage
		}
		if limit.DefaultLimit != nil {
			res[i]["default_limit"] = *limit.DefaultLimit
		}
		if limit.MaximumLimit != nil {
			res[i]["maximum_limit"] = *limit.MaximumLimit
		}
	}

	return res
}
