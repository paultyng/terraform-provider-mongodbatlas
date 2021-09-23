package mongodbatlas

import (
	"context"
	"fmt"
	"log"
	"sort"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	matlas "go.mongodb.org/atlas/mongodbatlas"
)

/*
	A cloud provider access authorization
*/

func resourceMongoDBAtlasCloudProviderAccessAuthorization() *schema.Resource {
	return &schema.Resource{
		ReadContext:   resourceMongoDBAtlasCloudProviderAccessAuthorizationRead,
		CreateContext: resourceMongoDBAtlasCloudProviderAccessAuthorizationCreate,
		UpdateContext: resourceMongoDBAtlasCloudProviderAccessAuthorizationUpdate,
		DeleteContext: resourceMongoDBAtlasCloudProviderAccessAuthorizationPlaceHolder,

		Schema: map[string]*schema.Schema{
			"project_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"role_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"aws": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"iam_assumed_role_arn": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
			"feature_usages": {
				Type:     schema.TypeList,
				Elem:     featureUsagesSchema(),
				Computed: true,
			},
			"authorized_date": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
		SchemaVersion: 1,
		StateUpgraders: []schema.StateUpgrader{
			{
				Type:    resourceMongoDBAtlasCloudProviderAccessAuthorizationResourceV0().CoreConfigSchema().ImpliedType(),
				Upgrade: resourceMongoDBAtlasCloudProviderAccessAuthorizationStateUpgradeV0,
				Version: 0,
			},
		},
	}
}

func resourceMongoDBAtlasCloudProviderAccessAuthorizationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// sadly there is no just get API
	conn := meta.(*MongoDBClient).Atlas
	ids := decodeStateID(d.Id())

	roleID := ids["id"] // atlas ID
	projectID := ids["project_id"]

	targetRole, err := FindRole(ctx, conn, projectID, roleID)
	if err != nil {
		reset := strings.Contains(err.Error(), "404") && !d.IsNewResource()
		if reset {
			d.SetId("")
			return nil
		}

		return diag.FromErr(err)
	}

	if targetRole == nil {
		return diag.FromErr(fmt.Errorf(errorGetRead, "cloud provider access role not found in mongodbatlas, please create it first"))
	}

	roleSchema := roleToSchemaAuthorization(targetRole)

	for key, val := range roleSchema {
		if err := d.Set(key, val); err != nil {
			return diag.FromErr(fmt.Errorf(errorGetRead, err))
		}
	}

	// If not authorize , then request the authorization
	if targetRole.AuthorizedDate == "" && !d.IsNewResource() {
		d.SetId("")
		return nil
	}

	return nil
}

func resourceMongoDBAtlasCloudProviderAccessAuthorizationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*MongoDBClient).Atlas

	projectID := d.Get("project_id").(string)
	roleID := d.Get("role_id").(string)

	// validation
	targetRole, err := FindRole(ctx, conn, projectID, roleID)

	if err != nil {
		return diag.FromErr(err)
	}

	if targetRole == nil {
		return diag.FromErr(fmt.Errorf(errorGetRead, "cloud provider access role not found in mongodbatlas, please create it first"))
	}

	// once multiple providers added, modify this section
	roleAWS, ok := d.GetOk("aws")

	if !ok {
		return diag.FromErr(fmt.Errorf("error CloudProviderAccessAuthorization missing iam_assumed_role_arn"))
	}

	iamRole := roleAWS.([]interface{})[0].(map[string]interface{})["iam_assumed_role_arn"]

	req := &matlas.CloudProviderAuthorizationRequest{
		ProviderName:      targetRole.ProviderName,
		IAMAssumedRoleARN: iamRole.(string),
	}

	var role *matlas.AWSIAMRole

	// aws takes time to update , in case of single path
	for i := 0; i < 3; i++ {
		role, _, err = conn.CloudProviderAccess.AuthorizeRole(ctx, projectID, roleID, req)
		if err != nil && strings.Contains(err.Error(), "CANNOT_ASSUME_ROLE") {
			log.Printf("warning issue performing authorize: %s \n", err.Error())
			log.Println("retrying ")
			time.Sleep(10 * time.Second)
			continue
		}
		break
	}

	if err != nil {
		return diag.FromErr(fmt.Errorf("error cloud provider access authorization %s", err))
	}

	authSchema := roleToSchemaAuthorization(role)

	// role id
	d.SetId(encodeStateID(map[string]string{
		"id":         role.RoleID,
		"project_id": projectID,
	}))

	for key, val := range authSchema {
		if err := d.Set(key, val); err != nil {
			return diag.FromErr(fmt.Errorf(errorCloudProviderAccessCreate, err))
		}
	}

	return nil
}

func resourceMongoDBAtlasCloudProviderAccessAuthorizationUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// sadly there is no just get API
	conn := meta.(*MongoDBClient).Atlas
	ids := decodeStateID(d.Id())

	roleID := ids["id"] // atlas ID
	projectID := ids["project_id"]

	targetRole, err := FindRole(ctx, conn, projectID, roleID)

	if err != nil {
		return diag.FromErr(err)
	}

	if targetRole == nil {
		return diag.FromErr(fmt.Errorf(errorGetRead, "cloud provider access role not found in mongodbatlas, please create it first"))
	}

	if d.HasChange("aws") {
		roleAWS, ok := d.GetOk("aws")

		if !ok {
			return diag.FromErr(fmt.Errorf("error CloudProviderAccessAuthorization missing iam_assumed_role_arn"))
		}

		iamRole := roleAWS.([]interface{})[0].(map[string]interface{})["iam_assumed_role_arn"]

		req := &matlas.CloudProviderAuthorizationRequest{
			ProviderName:      targetRole.ProviderName,
			IAMAssumedRoleARN: iamRole.(string),
		}

		var role *matlas.AWSIAMRole

		// aws takes time to update , in case of single path
		for i := 0; i < 3; i++ {
			role, _, err = conn.CloudProviderAccess.AuthorizeRole(ctx, projectID, roleID, req)
			if err != nil && strings.Contains(err.Error(), "CANNOT_ASSUME_ROLE") {
				log.Printf("warning issue performing authorize: %s \n", err.Error())
				log.Println("retrying ")
				time.Sleep(10 * time.Second)
				continue
			}
			break
		}

		if err != nil {
			return diag.FromErr(fmt.Errorf("error cloud provider access authorization update %s", err))
		}

		authSchema := roleToSchemaAuthorization(role)

		// role id
		d.SetId(encodeStateID(map[string]string{
			"id":         role.RoleID,
			"project_id": projectID,
		}))

		for key, val := range authSchema {
			if err := d.Set(key, val); err != nil {
				return diag.FromErr(fmt.Errorf(errorCloudProviderAccessCreate, err))
			}
		}
	}

	return nil
}

func resourceMongoDBAtlasCloudProviderAccessAuthorizationPlaceHolder(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return nil
}

func roleToSchemaAuthorization(role *matlas.AWSIAMRole) map[string]interface{} {
	out := map[string]interface{}{
		"role_id": role.RoleID,
		"aws": []interface{}{map[string]interface{}{
			"iam_assumed_role_arn": role.IAMAssumedRoleARN,
		}},
		"authorized_date": role.AuthorizedDate,
	}

	// features
	features := make([]map[string]interface{}, 0, len(role.FeatureUsages))

	for _, featureUsage := range role.FeatureUsages {
		features = append(features, featureToSchema(featureUsage))
	}

	out["feature_usages"] = features

	return out
}

func FindRole(ctx context.Context, conn *matlas.Client, projectID, roleID string) (targetRole *matlas.AWSIAMRole, err error) {
	roles, _, err := conn.CloudProviderAccess.ListRoles(ctx, projectID)

	if err != nil {
		return nil, fmt.Errorf(errorGetRead, err)
	}

	sort.Slice(roles.AWSIAMRoles,
		func(i, j int) bool { return roles.AWSIAMRoles[i].RoleID < roles.AWSIAMRoles[j].RoleID })

	index := sort.Search(len(roles.AWSIAMRoles), func(i int) bool { return roles.AWSIAMRoles[i].RoleID >= roleID })

	if index < len(roles.AWSIAMRoles) && roles.AWSIAMRoles[index].RoleID == roleID {
		targetRole = &(roles.AWSIAMRoles[index])
	}

	return
}

func resourceMongoDBAtlasCloudProviderAccessAuthorizationResourceV0() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"project_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"role_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"aws": {
				Type:     schema.TypeMap,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"iam_assumed_role_arn": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
			"feature_usages": {
				Type:     schema.TypeList,
				Elem:     featureUsagesSchema(),
				Computed: true,
			},
			"authorized_date": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceMongoDBAtlasCloudProviderAccessAuthorizationStateUpgradeV0(ctx context.Context, rawState map[string]interface{}, meta interface{}) (map[string]interface{}, error) {
	rawState["aws"] = []interface{}{}

	return rawState, nil
}
