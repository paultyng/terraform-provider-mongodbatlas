package mongodbatlas

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"regexp"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	matlas "go.mongodb.org/atlas/mongodbatlas"
)

func resourceMongoDBAtlasOrgInvitation() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceMongoDBAtlasOrgInvitationCreate,
		ReadContext:   resourceMongoDBAtlasOrgInvitationRead,
		DeleteContext: resourceMongoDBAtlasOrgInvitationDelete,
		UpdateContext: resourceMongoDBAtlasOrgInvitationUpdate,
		Importer: &schema.ResourceImporter{
			StateContext: resourceMongoDBAtlasOrgInvitationImportState,
		},
		Schema: map[string]*schema.Schema{
			"org_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"username": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"invitation_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"expires_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"created_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"roles": {
				Type:     schema.TypeSet,
				Required: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
					ValidateFunc: validation.StringInSlice([]string{
						"ORG_OWNER",
						"ORG_GROUP_CREATOR",
						"ORG_BILLING_ADMIN",
						"ORG_READ_ONLY",
						"ORG_MEMBER",
					}, false),
				},
			},
		},
	}
}

func resourceMongoDBAtlasOrgInvitationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// Get client connection.
	conn := meta.(*MongoDBClient).Atlas
	ids := decodeStateID(d.Id())
	orgID := ids["org_id"]
	username := ids["username"]
	invitationID := ids["invitation_id"]

	orgInvitation, _, err := conn.Organizations.Invitation(ctx, orgID, invitationID)
	if err != nil {
		// case 404
		// deleted in the backend case
		var target *matlas.ErrorResponse
		if errors.As(err, &target) && target.ErrorCode == "NOT_FOUND" {
			d.SetId("")
			return nil
		}

		return diag.FromErr(fmt.Errorf("error getting Organization Invitation information: %w", err))
	}

	if err := d.Set("username", orgInvitation.Username); err != nil {
		return diag.FromErr(fmt.Errorf("error getting `username` for Organization Invitation (%s): %w", d.Id(), err))
	}

	if err := d.Set("org_id", orgInvitation.GroupID); err != nil {
		return diag.FromErr(fmt.Errorf("error getting `username` for Organization Invitation (%s): %w", d.Id(), err))
	}

	if err := d.Set("invitation_id", orgInvitation.ID); err != nil {
		return diag.FromErr(fmt.Errorf("error getting `invitation_id` for Organization Invitation (%s): %w", d.Id(), err))
	}

	if err := d.Set("expires_at", orgInvitation.ExpiresAt); err != nil {
		return diag.FromErr(fmt.Errorf("error getting `expires_at` for Organization Invitation (%s): %w", d.Id(), err))
	}

	if err := d.Set("created_at", orgInvitation.CreatedAt); err != nil {
		return diag.FromErr(fmt.Errorf("error getting `created_at` for Organization Invitation (%s): %w", d.Id(), err))
	}

	if err := d.Set("roles", orgInvitation.Roles); err != nil {
		return diag.FromErr(fmt.Errorf("error getting `roles` for Organization Invitation (%s): %w", d.Id(), err))
	}

	d.SetId(encodeStateID(map[string]string{
		"username":      username,
		"org_id":        orgID,
		"invitation_id": invitationID,
	}))

	return nil
}

func resourceMongoDBAtlasOrgInvitationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// Get client connection.
	conn := meta.(*MongoDBClient).Atlas
	orgID := d.Get("org_id").(string)

	invitationReq := &matlas.Invitation{
		Roles:    expandStringListFromSetSchema(d.Get("roles").(*schema.Set)),
		Username: d.Get("username").(string),
	}

	invitationRes, resp, err := conn.Organizations.InviteUser(ctx, orgID, invitationReq)
	if err != nil {
		// case 404
		// deleted in the backend case
		if resp != nil && resp.StatusCode == http.StatusNotFound {
			d.SetId("")
			return nil
		}
		return diag.FromErr(fmt.Errorf("error creating Organization invitation for user %s: %w", d.Get("username").(string), err))
	}

	d.SetId(encodeStateID(map[string]string{
		"username":      invitationRes.Username,
		"org_id":        invitationRes.OrgID,
		"invitation_id": invitationRes.ID,
	}))

	return resourceMongoDBAtlasOrgInvitationRead(ctx, d, meta)
}

func resourceMongoDBAtlasOrgInvitationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*MongoDBClient).Atlas
	ids := decodeStateID(d.Id())
	orgID := ids["org_id"]
	username := ids["username"]
	invitationID := ids["invitation_id"]

	_, err := conn.Organizations.DeleteInvitation(ctx, orgID, invitationID)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error deleting Organization invitation for user %s: %w", username, err))
	}

	return nil
}

func resourceMongoDBAtlasOrgInvitationUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*MongoDBClient).Atlas
	ids := decodeStateID(d.Id())
	orgID := ids["org_id"]
	username := ids["username"]
	invitationID := ids["invitation_id"]

	invitationReq := &matlas.Invitation{
		Roles: expandStringListFromSetSchema(d.Get("roles").(*schema.Set)),
	}

	_, _, err := conn.Organizations.UpdateInvitationByID(ctx, orgID, invitationID, invitationReq)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error updating Organization invitation for user %s: for %w", username, err))
	}

	return resourceMongoDBAtlasOrgInvitationRead(ctx, d, meta)
}

func resourceMongoDBAtlasOrgInvitationImportState(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	conn := meta.(*MongoDBClient).Atlas
	orgID, username, err := splitOrgInvitationImportID(d.Id())
	if err != nil {
		return nil, err
	}

	orgInvitations, _, err := conn.Organizations.Invitations(ctx, orgID, nil)
	if err != nil {
		return nil, fmt.Errorf("couldn't import Organization invitations, error: %w", err)
	}

	for _, orgInvitation := range orgInvitations {
		if orgInvitation.Username != username {
			continue
		}

		if err := d.Set("username", orgInvitation.Username); err != nil {
			return nil, fmt.Errorf("error getting `username` for Organization Invitation (%s): %w", username, err)
		}
		if err := d.Set("org_id", orgInvitation.GroupID); err != nil {
			return nil, fmt.Errorf("error getting `org_id` for Organization Invitation (%s): %w", username, err)
		}
		if err := d.Set("invitation_id", orgInvitation.ID); err != nil {
			return nil, fmt.Errorf("error getting `invitation_id` for Organization Invitation (%s): %w", username, err)
		}
		d.SetId(encodeStateID(map[string]string{
			"username":      username,
			"org_id":        orgID,
			"invitation_id": orgInvitation.ID,
		}))
		return []*schema.ResourceData{d}, nil
	}

	return nil, fmt.Errorf("could not import Organization Invitation for %s", d.Id())
}

func splitOrgInvitationImportID(id string) (orgID, username string, err error) {
	var re = regexp.MustCompile(`(?s)^([0-9a-fA-F]{24})-(.*)$`)
	parts := re.FindStringSubmatch(id)

	if len(parts) != 3 {
		err = fmt.Errorf("import format error: to import a Organization Invitation, use the format {org_id}-{username}")
		return
	}

	orgID = parts[1]
	username = parts[2]

	return
}
