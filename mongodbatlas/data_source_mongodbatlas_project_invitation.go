package mongodbatlas

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceMongoDBAtlasProjectInvitation() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceMongoDBAtlasProjectInvitationRead,
		Schema: map[string]*schema.Schema{
			"project_id": {
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

func dataSourceMongoDBAtlasProjectInvitationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// Get client connection.
	conn := meta.(*MongoDBClient).Atlas
	projectID := d.Get("project_id").(string)
	username := d.Get("username").(string)
	invitationID := d.Get("invitation_id").(string)

	projectInvitation, _, err := conn.Projects.Invitation(ctx, projectID, invitationID)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error getting Project Invitation information: %w", err))
	}

	if err := d.Set("username", projectInvitation.Username); err != nil {
		return diag.FromErr(fmt.Errorf("error getting `username` for Project Invitation (%s): %w", d.Id(), err))
	}

	if err := d.Set("project_id", projectInvitation.GroupID); err != nil {
		return diag.FromErr(fmt.Errorf("error getting `username` for Project Invitation (%s): %w", d.Id(), err))
	}

	if err := d.Set("invitation_id", projectInvitation.ID); err != nil {
		return diag.FromErr(fmt.Errorf("error getting `invitation_id` for Project Invitation (%s): %w", d.Id(), err))
	}

	if err := d.Set("expires_at", projectInvitation.ExpiresAt); err != nil {
		return diag.FromErr(fmt.Errorf("error getting `expires_at` for Project Invitation (%s): %w", d.Id(), err))
	}

	if err := d.Set("created_at", projectInvitation.CreatedAt); err != nil {
		return diag.FromErr(fmt.Errorf("error getting `created_at` for Project Invitation (%s): %w", d.Id(), err))
	}

	if err := d.Set("inviter_username", projectInvitation.InviterUsername); err != nil {
		return diag.FromErr(fmt.Errorf("error getting `inviter_username` for Project Invitation (%s): %w", d.Id(), err))
	}

	if err := d.Set("roles", projectInvitation.Roles); err != nil {
		return diag.FromErr(fmt.Errorf("error getting `roles` for Project Invitation (%s): %s", d.Id(), err))
	}

	d.SetId(encodeStateID(map[string]string{
		"username":      username,
		"project_id":    projectID,
		"invitation_id": invitationID,
	}))

	return nil
}
