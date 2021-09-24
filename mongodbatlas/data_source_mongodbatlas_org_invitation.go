package mongodbatlas

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceMongoDBAtlasOrgInvitation() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceMongoDBAtlasOrgInvitationRead,
		Schema: map[string]*schema.Schema{
			"org_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"username": {
				Type:     schema.TypeString,
				Required: true,
			},
			"invitation_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"expires_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"created_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"inviter_username": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"roles": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}
}

func dataSourceMongoDBAtlasOrgInvitationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// Get client connection.
	conn := meta.(*MongoDBClient).Atlas
	orgID := d.Get("org_id").(string)
	username := d.Get("username").(string)
	invitationID := d.Get("invitation_id").(string)

	orgInvitation, _, err := conn.Organizations.Invitation(ctx, orgID, invitationID)
	if err != nil {
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

	if err := d.Set("inviter_username", orgInvitation.InviterUsername); err != nil {
		return diag.FromErr(fmt.Errorf("error getting `inviter_username` for Organization Invitation (%s): %w", d.Id(), err))
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
