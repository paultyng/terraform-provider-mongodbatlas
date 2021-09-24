---
layout: "mongodbatlas"
page_title: "MongoDB Atlas: project_invitation"
sidebar_current: "docs-mongodbatlas-resource-project-invitation"
description: |-
    Provides a Atlas Project Invitation resource.
---

# mongodbatlas_project_invitation

`mongodbatlas_project_invitation` provides a Project Invitation resource. This represents an invitation for an Atlas User within an Atlas Project.

Each invitation for an Atlas user has a set of roles that provide access to a project in an organization.

The roles that can be utilised can be found in the [MongoDB Documentation](https://docs.atlas.mongodb.com/reference/user-roles/#project-roles), which map to:

* GROUP_OWNER
* GROUP_CLUSTER_MANAGER
* GROUP_READ_ONLY
* GROUP_DATA_ACCESS_ADMIN
* GROUP_DATA_ACCESS_READ_WRITE
* GROUP_DATA_ACCESS_READ_ONLY

-> **NOTE:** Groups and projects are synonymous terms. You may find group_id in the official documentation.

~> **IMPORTANT:** All arguments including the password will be stored in the raw state as plain-text. [Read more about sensitive data in state.](https://www.terraform.io/docs/state/sensitive-data.html)

## Example Usages

```hcl
resource "mongodbatlas_project_invitation" "test" {
  username    = "test-acc-username"
  project_id  = "<PROJECT-ID>"
  roles       = [ "GROUP_DATA_ACCESS_READ_WRITE" ]
}
```

```hcl
resource "mongodbatlas_project_invitation" "test" {
  username    = "test-acc-username"
  project_id  = "<PROJECT-ID>"
  roles       = [ "GROUP_READ_ONLY", "GROUP_DATA_ACCESS_READ_ONLY" ]
}
```

## Argument Reference

* `project_id` - (Required) The unique ID for the project to create the database user.
* `username` - (Required) The Atlas user's email address.
* `roles` - (Required) 	List of user’s roles within the Atlas project. The available options are:
  * GROUP_OWNER
  * GROUP_CLUSTER_MANAGER
  * GROUP_READ_ONLY
  * GROUP_DATA_ACCESS_ADMIN
  * GROUP_DATA_ACCESS_READ_WRITE
  * GROUP_DATA_ACCESS_READ_ONLY

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The database user's name.
* `created_at` - The date and time the invitation was created
* `expires_at` - The date and time that the invitation will expire
* `invitation_id` - The identify of the invitation in Atlas

## Import

Projects Invitations can be imported using project ID and username (email address), in the format `project_id`-`username`, e.g.

```
$ terraform import mongodbatlas_project_invitation.my_user 1112222b3bf99403840e8934-my_user
```
