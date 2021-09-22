package mongodbatlas

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceMongoDBAtlasOrganizationApiKey() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceMongoDBAtlasOrganizationApiKeyRead,
		Schema: map[string]*schema.Schema{
			"org_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"api_key_id": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
			},
			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"roles": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"access_list_cidr_blocks": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}
}

func dataSourceMongoDBAtlasOrganizationApiKeyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*MongoDBClient).Atlas
	orgID := d.Get("org_id").(string)
	apiKeyID := d.Get("api_key_id").(string)

	apiKey, _, err := conn.APIKeys.Get(ctx, orgID, apiKeyID)
	if err != nil {
		return diag.FromErr(fmt.Errorf(errorOrgApiKeyRead, orgID, apiKeyID, err))
	}

	if err := d.Set("api_key_id", apiKey.ID); err != nil {
		return diag.FromErr(fmt.Errorf(errorOrgApiKeyRead, orgID, apiKeyID, err))
	}

	apiKeyAccessList, _, err := conn.AccessListAPIKeys.List(ctx, orgID, apiKeyID, nil)
	if err != nil {
		return diag.FromErr(fmt.Errorf(errorOrgApiKeyRead, orgID, apiKeyID, err))
	}

	// description
	if err := d.Set("description", apiKey.Desc); err != nil {
		return diag.FromErr(fmt.Errorf(errorOrgApiKeyRead, orgID, apiKeyID, err))
	}

	// roles
	roles := []string{}
	for i := range apiKey.Roles {
		roles = append(roles, apiKey.Roles[i].RoleName)
	}

	if err := d.Set("roles", roles); err != nil {
		return diag.FromErr(fmt.Errorf(errorOrgApiKeyRead, orgID, apiKeyID, err))
	}

	// access_list_cidr_blocks
	ip_addresses := []string{}
	for i := range apiKeyAccessList.Results {
		ip_addresses = append(ip_addresses, apiKeyAccessList.Results[i].CidrBlock)
	}

	if err := d.Set("access_list_cidr_blocks", ip_addresses); err != nil {
		return diag.FromErr(fmt.Errorf(errorOrgApiKeyRead, orgID, apiKeyID, err))
	}

	d.SetId(encodeStateID(map[string]string{
		"org_id":     orgID,
		"api_key_id": apiKeyID,
	}))

	return nil
}
