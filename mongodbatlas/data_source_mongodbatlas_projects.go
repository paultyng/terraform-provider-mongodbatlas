package mongodbatlas

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	matlas "go.mongodb.org/atlas/mongodbatlas"
)

func dataSourceMongoDBAtlasProjects() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceMongoDBAtlasProjectsRead,
		Schema: map[string]*schema.Schema{
			"page_num": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"items_per_page": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"results": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"org_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"name": {
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
					},
				},
			},
			"total_count": {
				Type:     schema.TypeInt,
				Computed: true,
			},
		},
	}
}

func dataSourceMongoDBAtlasProjectsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// Get client connection.
	conn := meta.(*MongoDBClient).Atlas
	options := &matlas.ListOptions{
		PageNum:      d.Get("page_num").(int),
		ItemsPerPage: d.Get("items_per_page").(int),
	}

	projects, _, err := conn.Projects.GetAllProjects(ctx, options)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error getting projects information: %s", err))
	}

	if err := d.Set("results", flattenProjects(ctx, conn, projects.Results)); err != nil {
		return diag.FromErr(fmt.Errorf("error setting `results`: %s", err))
	}

	if err := d.Set("total_count", projects.TotalCount); err != nil {
		return diag.FromErr(fmt.Errorf("error setting `name`: %s", err))
	}

	d.SetId(resource.UniqueId())

	return nil
}

func flattenProjects(ctx context.Context, conn *matlas.Client, projects []*matlas.Project) []map[string]interface{} {
	var results []map[string]interface{}

	if len(projects) > 0 {
		results = make([]map[string]interface{}, len(projects))

		for k, project := range projects {
			teams, _, err := conn.Projects.GetProjectTeamsAssigned(ctx, project.ID)
			if err != nil {
				fmt.Printf("[WARN] error getting project's teams assigned (%s): %s", project.ID, err)
			}

			apiKeys, err := getProjectApiKeys(conn, ctx, project.OrgID, project.ID)
			if err != nil {
				fmt.Printf("[WARN] error getting project's api keys (%s): %s", project.ID, err)
			}
			results[k] = map[string]interface{}{
				"id":            project.ID,
				"org_id":        project.OrgID,
				"name":          project.Name,
				"cluster_count": project.ClusterCount,
				"created":       project.Created,
				"teams":         flattenTeams(teams),
				"api_keys":      flattenAPIKeys(apiKeys),
			}
		}
	}

	return results
}
