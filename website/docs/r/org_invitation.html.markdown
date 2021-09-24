---
layout: "mongodbatlas"
page_title: "MongoDB Atlas: org_invitation"
sidebar_current: "docs-mongodbatlas-organization-invitation"
description: |-
    Provides a Atlas Organisation Invitation resource.
---

# mongodbatlas_org_invitation

`mongodbatlas_org_invitation` provides a Organisation Invitation resource. This represents an invitation for an Atlas User within an Atlas Organisation.

Each invitation for an Atlas user has a set of roles that provide access to an organization.

The roles that can be utilised can be found in the [MongoDB Documentation](https://docs.atlas.mongodb.com/reference/user-roles/#organization-roles), which map to:

* ORG_OWNER
* ORG_GROUP_CREATOR
* ORG_BILLING_ADMIN
* ORG_READ_ONLY
* ORG_MEMBER

-> **NOTE:** Groups and projects are synonymous terms. You may find group_id in the official documentation.

~> **IMPORTANT:** All arguments including the password will be stored in the raw state as plain-text. [Read more about sensitive data in state.](https://www.terraform.io/docs/state/sensitive-data.html)

## Example Usages

```hcl
resource "mongodbatlas_org_invitation" "test" {
  username    = "test-acc-username"
  org_id      = "<ORG-ID>"
  roles       = [ "GROUP_DATA_ACCESS_READ_WRITE" ]
}
```

```hcl
resource "mongodbatlas_org_invitation" "test" {
  username    = "test-acc-username"
  org_id      = "<ORG-ID>"
  roles       = [ "GROUP_READ_ONLY", "GROUP_DATA_ACCESS_READ_ONLY" ]
}
```

## Argument Reference

* `project_id` - (Required) The unique ID for the project to create the database user.
* `username` - (Required) The Atlas user's email address.
* `roles` - (Required) 	List of user’s Atlas roles. The available options are:
  * ORG_OWNER
  * ORG_GROUP_CREATOR
  * ORG_BILLING_ADMIN
  * ORG_READ_ONLY
  * ORG_MEMBER

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The database user's name.
* `created_at` - The date and time the invitation was created
* `expires_at` - The date and time that the invitation will expire
* `invitation_id` - The identify of the invitation in Atlas

## Import

Organisations Invitations can be imported using organization ID and username (email address), in the format `org_id`-`username`, e.g.

```
$ terraform import mongodbatlas_org_invitation.my_user 1112222b3bf99403840e8934-my_user
```
