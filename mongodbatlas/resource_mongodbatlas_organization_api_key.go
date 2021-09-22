package mongodbatlas

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/url"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	matlas "go.mongodb.org/atlas/mongodbatlas"
)

const (
	errorOrgApiKeyCreate         = "error creating the MongoDB Organization (%s) API Key: %s"
	errorOrgApiKeyUpdate         = "error updating the MongoDB Organization (%s) API Key (%s): %s"
	errorOrgApiKeyRead           = "error reading the MongoDB Organization (%s) API Key (%s): %s"
	errorOrgApiKeyDelete         = "error deleting the MongoDB Organization (%s) API Key (%s): %s"
	errorOrgApiKeyAtLeastOneRole = "error, at least one role must be present for an API Key"
	errorOrgApiKeyInvalidCIDR    = "error creating the MongoDB Organization (%s) API Key, invalid CIDR block %s"
)

func resourceMongoDBAtlasOrganizationApiKey() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceMongoDBAtlasOrganizationApiKeyCreate,
		ReadContext:   resourceMongoDBAtlasOrganizationApiKeyRead,
		UpdateContext: resourceMongoDBAtlasOrganizationApiKeyUpdate,
		DeleteContext: resourceMongoDBAtlasOrganizationApiKeyDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceMongoDBAtlasOrganizationApiKeyImportState,
		},
		Schema: map[string]*schema.Schema{
			"org_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"api_key_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"public_key": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"private_key": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:     schema.TypeString,
				Required: true,
			},
			"roles": {
				Type:     schema.TypeSet,
				Required: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"access_list_cidr_blocks": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}
}

func resourceMongoDBAtlasOrganizationApiKeyCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*MongoDBClient).Atlas
	orgID := d.Get("org_id").(string)

	// Checking there is at least one role
	roles := expandStringListFromSetSchema(d.Get("roles").(*schema.Set))
	if len(roles) == 0 {
		return diag.FromErr(fmt.Errorf(errorOrgApiKeyAtLeastOneRole))
	}

	// Before creating the API keys, first validating cidr blocks
	accessList := expandStringListFromSetSchema(d.Get("access_list_cidr_blocks").(*schema.Set))
	invalidCIDR := validateCIDRBlocks(accessList)
	if invalidCIDR != "" {
		return diag.FromErr(fmt.Errorf(errorOrgApiKeyInvalidCIDR, orgID, invalidCIDR))
	}

	// Creating org key
	apiKey, _, err := conn.APIKeys.Create(ctx, orgID,
		&matlas.APIKeyInput{
			Desc:  d.Get("description").(string),
			Roles: roles,
		})
	if err != nil {
		return diag.FromErr(fmt.Errorf(errorOrgApiKeyCreate, orgID, err))
	}

	// Creating API key access list
	err = createAccessList(conn.AccessListAPIKeys, ctx, orgID, apiKey.ID, accessList)
	if err != nil {
		return diag.FromErr(fmt.Errorf(errorOrgApiKeyCreate, orgID, err))
	}

	d.SetId(encodeStateID(map[string]string{
		"org_id":     orgID,
		"api_key_id": apiKey.ID,
	}))

	return resourceMongoDBAtlasOrganizationApiKeyRead(ctx, d, meta)
}

func resourceMongoDBAtlasOrganizationApiKeyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {

	conn := meta.(*MongoDBClient).Atlas

	ids := decodeStateID(d.Id())
	orgID := ids["org_id"]
	apiKeyID := ids["api_key_id"]

	apiKey, resp, err := conn.APIKeys.Get(context.Background(), orgID, apiKeyID)

	if err != nil {
		if resp != nil && resp.StatusCode == http.StatusNotFound {
			d.SetId("")
			return nil
		}

		return diag.FromErr(fmt.Errorf(errorOrgApiKeyRead, orgID, apiKeyID, err))
	}

	if err := d.Set("api_key_id", apiKey.ID); err != nil {
		return diag.FromErr(fmt.Errorf(errorOrgApiKeyRead, orgID, apiKeyID, err))
	}

	if err := d.Set("public_key", apiKey.PublicKey); err != nil {
		return diag.FromErr(fmt.Errorf(errorOrgApiKeyRead, orgID, apiKeyID, err))
	}

	if err := d.Set("private_key", apiKey.PrivateKey); err != nil {
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
	accessList := []string{}
	for i := range apiKeyAccessList.Results {
		accessList = append(accessList, apiKeyAccessList.Results[i].CidrBlock)
	}

	if err := d.Set("access_list_cidr_blocks", accessList); err != nil {
		return diag.FromErr(fmt.Errorf(errorOrgApiKeyRead, orgID, apiKeyID, err))
	}

	return nil
}

func resourceMongoDBAtlasOrganizationApiKeyUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*MongoDBClient).Atlas

	ids := decodeStateID(d.Id())
	orgID := ids["org_id"]
	apiKeyID := ids["api_key_id"]

	// description
	if d.HasChange("description") {
		_, _, err := conn.APIKeys.Update(ctx, orgID, apiKeyID,
			&matlas.APIKeyInput{
				Desc: d.Get("description").(string),
			})
		if err != nil {
			return diag.FromErr(fmt.Errorf(errorOrgApiKeyUpdate, orgID, apiKeyID, err))
		}
	}
	// roles
	if d.HasChange("roles") {
		roles := expandStringListFromSetSchema(d.Get("roles").(*schema.Set))

		// Checking there is at least one role
		if len(roles) == 0 {
			return diag.FromErr(fmt.Errorf(errorOrgApiKeyAtLeastOneRole))
		}

		_, _, err := conn.APIKeys.Update(ctx, orgID, apiKeyID,
			&matlas.APIKeyInput{
				Roles: roles,
			})
		if err != nil {
			return diag.FromErr(fmt.Errorf(errorOrgApiKeyUpdate, orgID, apiKeyID, err))
		}
	}

	// access_list_cidr_blocks
	// As of 8/20/2021, access list update (PATCH) is not supported in Atlas API
	// https://docs.atlas.mongodb.com/reference/api/apiKeys/#organization-api-key-access-list-endpoints
	//
	// So, deleting the existing ones and create the new ones
	if d.HasChange("access_list_cidr_blocks") {
		// First, validating new CIDR blocks
		accessList := expandStringListFromSetSchema(d.Get("access_list_cidr_blocks").(*schema.Set))
		invalidCIDR := validateCIDRBlocks(accessList)
		if invalidCIDR != "" {
			return diag.FromErr(fmt.Errorf(errorOrgApiKeyInvalidCIDR, orgID, invalidCIDR))
		}

		// Deleting
		err := deleteAccessListIPs(conn.AccessListAPIKeys, ctx, orgID, apiKeyID)
		if err != nil {
			return diag.FromErr(fmt.Errorf(errorOrgApiKeyDelete, orgID, apiKeyID, err))
		}
		// Recreating
		err = createAccessList(conn.AccessListAPIKeys, ctx, orgID, apiKeyID, accessList)
		if err != nil {
			return diag.FromErr(fmt.Errorf(errorOrgApiKeyCreate, orgID, err))
		}
	}

	return resourceMongoDBAtlasOrganizationApiKeyRead(ctx, d, meta)
}

func resourceMongoDBAtlasOrganizationApiKeyDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*MongoDBClient).Atlas

	ids := decodeStateID(d.Id())
	orgID := ids["org_id"]
	apiKeyID := ids["api_key_id"]

	// Deleting access list ips
	err := deleteAccessListIPs(conn.AccessListAPIKeys, ctx, orgID, apiKeyID)
	if err != nil {
		return diag.FromErr(fmt.Errorf(errorOrgApiKeyDelete, orgID, apiKeyID, err))
	}

	// Deleting API keys
	_, err = conn.APIKeys.Delete(ctx, orgID, apiKeyID)
	if err != nil {
		return diag.FromErr(fmt.Errorf(errorOrgApiKeyDelete, orgID, apiKeyID, err))
	}

	return nil
}

func deleteAccessListIPs(apiKeysResource matlas.AccessListAPIKeysService, ctx context.Context, orgID, apiKeyID string) error {
	apiKeyAccessList, _, err := apiKeysResource.List(ctx, orgID, apiKeyID, nil)
	if err != nil {
		return err
	}

	for i := range apiKeyAccessList.Results {
		// Atlas API stores /32 as regular IPs without the CIDR notation
		// detecting them and stripping the "/32" part so they can be deleted
		toDelete := apiKeyAccessList.Results[i].CidrBlock
		if toDelete[len(toDelete)-3:] == "/32" {
			toDelete = toDelete[:len(toDelete)-3]
		}
		_, err = apiKeysResource.Delete(ctx, orgID, apiKeyID, url.QueryEscape(toDelete))
		if err != nil {
			return err
		}
	}
	return nil
}

func validateCIDRBlocks(accessList []string) string {
	for i := range accessList {
		// Checking address is a cidr block
		cidrBlock, _, _ := net.ParseCIDR(accessList[i])
		if cidrBlock == nil {
			return accessList[i]
		}
	}
	return ""
}

func createAccessList(apiKeysResource matlas.AccessListAPIKeysService, ctx context.Context, orgID, apiKeyID string, accessList []string) error {
	keyAccessList := []*matlas.AccessListAPIKeysReq{}
	for i := range accessList {
		keyAccessList = append(keyAccessList, &matlas.AccessListAPIKeysReq{
			CidrBlock: accessList[i],
		})
	}

	_, _, err := apiKeysResource.Create(ctx, orgID, apiKeyID, keyAccessList)
	if err != nil {
		return err
	}
	return nil
}

func resourceMongoDBAtlasOrganizationApiKeyImportState(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	conn := meta.(*MongoDBClient).Atlas

	parts := strings.SplitN(d.Id(), "-", 2)
	if len(parts) != 2 {
		return nil, errors.New("import format error: to import an API key, use the format {org_id}-{api_public_key}")
	}

	orgID := parts[0]
	apiPublicKey := parts[1]

	apiKeyList, _, err := conn.APIKeys.List(ctx, orgID, nil)
	if err != nil {
		return nil, fmt.Errorf("couldn't import API key (%s) in organization (%s), error: %s", apiPublicKey, orgID, err)
	}

	var apiKey matlas.APIKey
	for i := range apiKeyList {
		if apiKeyList[i].PublicKey == apiPublicKey {
			apiKey = apiKeyList[i]
		}
	}
	if apiKey.ID == "" {
		return nil, fmt.Errorf("couldn't find API key (%s) in organization (%s), error: %s", apiPublicKey, orgID, err)
	}

	if err := d.Set("org_id", orgID); err != nil {
		log.Printf("[WARN] Error setting org_id for (%s): %s", orgID, err)
	}

	if err := d.Set("api_key_id", apiKey.ID); err != nil {
		log.Printf("[WARN] Error setting api_key_id for (%s): %s", apiPublicKey, err)
	}

	if err := d.Set("public_key", apiPublicKey); err != nil {
		log.Printf("[WARN] Error setting api_key_id for (%s): %s", apiPublicKey, err)
	}

	if err := d.Set("private_key", apiKey.PrivateKey); err != nil {
		log.Printf("[WARN] Error setting api_key_id for (%s): %s", apiPublicKey, err)
	}

	d.SetId(encodeStateID(map[string]string{
		"org_id":     orgID,
		"api_key_id": apiKey.ID,
	}))

	return []*schema.ResourceData{d}, nil
}
