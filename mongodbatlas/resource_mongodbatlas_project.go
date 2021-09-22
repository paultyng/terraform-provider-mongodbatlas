package mongodbatlas

import (
	"context"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	matlas "go.mongodb.org/atlas/mongodbatlas"
)

const (
	errorProjectCreate  = "error creating Project: %s"
	errorProjectRead    = "error getting project(%s): %s"
	errorProjectDelete  = "error deleting project (%s): %s"
	errorProjectSetting = "error setting `%s` for project (%s): %s"
)

func resourceMongoDBAtlasProject() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceMongoDBAtlasProjectCreate,
		ReadContext:   resourceMongoDBAtlasProjectRead,
		UpdateContext: resourceMongoDBAtlasProjectUpdate,
		DeleteContext: resourceMongoDBAtlasProjectDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"org_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
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
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"team_id": {
							Type:     schema.TypeString,
							Required: true,
						},
						"role_names": {
							Type:     schema.TypeSet,
							Required: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
					},
				},
			},
			"api_keys": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"api_key_id": {
							Type:     schema.TypeString,
							Required: true,
						},
						"role_names": {
							Type:     schema.TypeSet,
							Required: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
					},
				},
			},
		},
	}
}

func resourceMongoDBAtlasProjectCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// Get client connection.
	conn := meta.(*MongoDBClient).Atlas

	projectReq := &matlas.Project{
		OrgID: d.Get("org_id").(string),
		Name:  d.Get("name").(string),
	}

	project, _, err := conn.Projects.Create(ctx, projectReq, nil)
	if err != nil {
		return diag.FromErr(fmt.Errorf(errorProjectCreate, err))
	}

	// Check if teams were set, if so we need to add the teams into the project
	if teams, ok := d.GetOk("teams"); ok {
		// adding the teams into the project
		_, _, err := conn.Projects.AddTeamsToProject(ctx, project.ID, expandTeamsSet(teams.(*schema.Set)))
		if err != nil {
			return diag.FromErr(fmt.Errorf("error adding teams into the project: %s", err))
		}
	}

	// Check if api keys were set, if so we need to add keys into the project
	if apiKeys, ok := d.GetOk("api_keys"); ok {
		// assign api keys to the project
		for _, apiKey := range expandAPIKeysSet(apiKeys.(*schema.Set)) {
			_, err := conn.ProjectAPIKeys.Assign(ctx, project.ID, apiKey.id, &matlas.AssignAPIKey{
				Roles: apiKey.roles,
			})
			if err != nil {
				return diag.FromErr(fmt.Errorf("error assigning api keys to the project: %s", err))
			}
		}
	}

	d.SetId(project.ID)

	return resourceMongoDBAtlasProjectRead(ctx, d, meta)
}

func resourceMongoDBAtlasProjectRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*MongoDBClient).Atlas
	projectID := d.Id()

	projectRes, resp, err := conn.Projects.GetOneProject(context.Background(), projectID)
	if err != nil {
		if resp != nil && resp.StatusCode == http.StatusNotFound {
			d.SetId("")
			return nil
		}

		return diag.FromErr(fmt.Errorf(errorProjectRead, projectID, err))
	}

	teams, _, err := conn.Projects.GetProjectTeamsAssigned(ctx, projectID)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error getting project's teams assigned (%s): %s", projectID, err))
	}

	apiKeys, err := getProjectApiKeys(conn, ctx, projectRes.OrgID, projectRes.ID)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error getting project's api keys (%s): %s", projectID, err))
	}

	if err := d.Set("name", projectRes.Name); err != nil {
		return diag.FromErr(fmt.Errorf(errorProjectSetting, `name`, projectID, err))
	}

	if err := d.Set("org_id", projectRes.OrgID); err != nil {
		return diag.FromErr(fmt.Errorf(errorProjectSetting, `org_id`, projectID, err))
	}

	if err := d.Set("cluster_count", projectRes.ClusterCount); err != nil {
		return diag.FromErr(fmt.Errorf(errorProjectSetting, `clusterCount`, projectID, err))
	}

	if err := d.Set("created", projectRes.Created); err != nil {
		return diag.FromErr(fmt.Errorf(errorProjectSetting, `created`, projectID, err))
	}

	if err := d.Set("teams", flattenTeams(teams)); err != nil {
		return diag.FromErr(fmt.Errorf(errorProjectSetting, `created`, projectID, err))
	}

	if err := d.Set("api_keys", flattenAPIKeys(apiKeys)); err != nil {
		return diag.FromErr(fmt.Errorf(errorProjectSetting, `api_keys`, projectID, err))
	}

	return nil
}

func resourceMongoDBAtlasProjectUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*MongoDBClient).Atlas
	projectID := d.Id()

	if d.HasChange("teams") {
		// get the current teams and the new teams with changes
		newTeams, changedTeams, removedTeams := getStateTeams(d)

		// adding new teams into the project
		if len(newTeams) > 0 {
			_, _, err := conn.Projects.AddTeamsToProject(ctx, projectID, expandTeamsList(newTeams))
			if err != nil {
				return diag.FromErr(fmt.Errorf("error adding teams into the project(%s): %s", projectID, err))
			}
		}

		// Removing teams from the project
		for _, team := range removedTeams {
			teamID := team.(map[string]interface{})["team_id"].(string)

			_, err := conn.Teams.RemoveTeamFromProject(ctx, projectID, teamID)
			if err != nil {
				return diag.FromErr(fmt.Errorf("error removing team(%s) from the project(%s): %s", teamID, projectID, err))
			}
		}

		// Updating the role names for a team
		for _, t := range changedTeams {
			team := t.(map[string]interface{})

			_, _, err := conn.Teams.UpdateTeamRoles(ctx, projectID, team["team_id"].(string),
				&matlas.TeamUpdateRoles{
					RoleNames: expandStringList(team["role_names"].(*schema.Set).List()),
				},
			)
			if err != nil {
				return diag.FromErr(fmt.Errorf("error updating role names for the team(%s): %s", team["team_id"], err))
			}
		}
	}

	if d.HasChange("api_keys") {
		// get the current api_keys and the new api_keys with changes
		newAPIKeys, changedAPIKeys, removedAPIKeys := getStateAPIKeys(d)

		// adding new api_keys into the project
		if len(newAPIKeys) > 0 {
			for _, apiKey := range expandAPIKeysList(newAPIKeys) {
				_, err := conn.ProjectAPIKeys.Assign(ctx, projectID, apiKey.id, &matlas.AssignAPIKey{
					Roles: apiKey.roles,
				})
				if err != nil {
					return diag.FromErr(fmt.Errorf("error assigning api_keys into the project(%s): %s", projectID, err))
				}
			}
		}

		// Removing api_keys from the project
		for _, apiKey := range removedAPIKeys {
			apiKeyID := apiKey.(map[string]interface{})["api_key_id"].(string)
			_, err := conn.ProjectAPIKeys.Unassign(ctx, projectID, apiKeyID)
			if err != nil {
				return diag.FromErr(fmt.Errorf("error removing api_key(%s) from the project(%s): %s", apiKeyID, projectID, err))
			}
		}

		// Updating the role names for the api_key
		for _, apiKey := range expandAPIKeysList(changedAPIKeys) {
			_, err := conn.ProjectAPIKeys.Assign(ctx, projectID, apiKey.id, &matlas.AssignAPIKey{
				Roles: apiKey.roles,
			})
			if err != nil {
				return diag.FromErr(fmt.Errorf("error updating role names for the api_key(%s): %s", apiKey, err))
			}
		}
	}

	return resourceMongoDBAtlasProjectRead(ctx, d, meta)
}

func resourceMongoDBAtlasProjectDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*MongoDBClient).Atlas
	projectID := d.Id()

	_, err := conn.Projects.Delete(ctx, projectID)
	if err != nil {
		return diag.FromErr(fmt.Errorf(errorProjectDelete, projectID, err))
	}

	return nil
}

func expandTeamsSet(teams *schema.Set) []*matlas.ProjectTeam {
	res := make([]*matlas.ProjectTeam, teams.Len())

	for i, value := range teams.List() {
		v := value.(map[string]interface{})
		res[i] = &matlas.ProjectTeam{
			TeamID:    v["team_id"].(string),
			RoleNames: expandStringList(v["role_names"].(*schema.Set).List()),
		}
	}

	return res
}

func expandAPIKeysSet(apiKeys *schema.Set) []*apiKey {
	res := make([]*apiKey, apiKeys.Len())

	for i, value := range apiKeys.List() {
		v := value.(map[string]interface{})
		res[i] = &apiKey{
			id:    v["api_key_id"].(string),
			roles: expandStringList(v["role_names"].(*schema.Set).List()),
		}
	}

	return res
}

func expandTeamsList(teams []interface{}) []*matlas.ProjectTeam {
	res := make([]*matlas.ProjectTeam, len(teams))

	for i, value := range teams {
		v := value.(map[string]interface{})
		res[i] = &matlas.ProjectTeam{
			TeamID:    v["team_id"].(string),
			RoleNames: expandStringList(v["role_names"].(*schema.Set).List()),
		}
	}

	return res
}

func expandAPIKeysList(apiKeys []interface{}) []*apiKey {
	res := make([]*apiKey, len(apiKeys))

	for i, value := range apiKeys {
		v := value.(map[string]interface{})
		res[i] = &apiKey{
			id:    v["api_key_id"].(string),
			roles: expandStringList(v["role_names"].(*schema.Set).List()),
		}
	}

	return res
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

func getStateTeams(d *schema.ResourceData) (newTeams, changedTeams, removedTeams []interface{}) {
	currentTeams, changes := d.GetChange("teams")

	rTeams := currentTeams.(*schema.Set).Difference(changes.(*schema.Set))
	nTeams := changes.(*schema.Set).Difference(currentTeams.(*schema.Set))
	changedTeams = make([]interface{}, 0)

	for _, changed := range nTeams.List() {
		for _, removed := range rTeams.List() {
			if changed.(map[string]interface{})["team_id"] == removed.(map[string]interface{})["team_id"] {
				rTeams.Remove(removed)
			}
		}

		for _, current := range currentTeams.(*schema.Set).List() {
			if changed.(map[string]interface{})["team_id"] == current.(map[string]interface{})["team_id"] {
				changedTeams = append(changedTeams, changed.(map[string]interface{}))
				nTeams.Remove(changed)
			}
		}
	}

	newTeams = nTeams.List()
	removedTeams = rTeams.List()

	return
}

func getStateAPIKeys(d *schema.ResourceData) (newAPIKeys, changedAPIKeys, removedAPIKeys []interface{}) {
	currentAPIKeys, changes := d.GetChange("api_keys")

	rAPIKeys := currentAPIKeys.(*schema.Set).Difference(changes.(*schema.Set))
	nAPIKeys := changes.(*schema.Set).Difference(currentAPIKeys.(*schema.Set))
	changedAPIKeys = make([]interface{}, 0)

	for _, changed := range nAPIKeys.List() {
		for _, removed := range rAPIKeys.List() {
			if changed.(map[string]interface{})["api_key_id"] == removed.(map[string]interface{})["api_key_id"] {
				rAPIKeys.Remove(removed)
			}
		}

		for _, current := range currentAPIKeys.(*schema.Set).List() {
			if changed.(map[string]interface{})["api_key_id"] == current.(map[string]interface{})["api_key_id"] {
				changedAPIKeys = append(changedAPIKeys, changed.(map[string]interface{}))
				nAPIKeys.Remove(changed)
			}
		}
	}

	newAPIKeys = nAPIKeys.List()
	removedAPIKeys = rAPIKeys.List()

	return
}
