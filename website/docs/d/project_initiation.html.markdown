---
layout: "mongodbatlas"
page_title: "MongoDB Atlas: project_invitation"
sidebar_current: "docs-mongodbatlas-project-invitation"
description: |-
    Provides a Atlas Project Invitation resource.
---

# mongodbatlas_project_invitation

`mongodbatlas_project_invitation` describes a Project Invitation resource. This represents an invitation for an Atlas User within an Atlas Project.

Each invitation for an Atlas user has a set of roles that provide access to a project in an organization.

-> **NOTE:** Groups and projects are synonymous terms. You may find group_id in the official documentation.

~> **IMPORTANT:** All arguments including the password will be stored in the raw state as plain-text. [Read more about sensitive data in state.](https://www.terraform.io/docs/state/sensitive-data.html)

## Example Usages

```hcl
resource "mongodbatlas_project_invitation" "test" {
  username    = "test-acc-username"
  project_id  = "<PROJECT-ID>"
  roles       = [ "GROUP_DATA_ACCESS_READ_WRITE" ]
}

data "mongodbatlas_project_invitation" "test" {
  project_id = mongodbatlas_project_invitation.test.project_id
  username   = mongodbatlas_project_invitation.test.username
}
```

## Argument Reference

* `project_id` - (Required) The unique ID for the project to create the database user.
* `username` - (Required) The Atlas user's email address.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The database user's name.
* `created_at` - The date and time the invitation was created
* `expires_at` - The date and time that the invitation will expire
* `invitation_id` - The identify of the invitation in Atlas
* `roles` - List of user’s roles within the Atlas project. The available options are:
  * GROUP_OWNER
  * GROUP_CLUSTER_MANAGER
  * GROUP_READ_ONLY
  * GROUP_DATA_ACCESS_ADMIN
  * GROUP_DATA_ACCESS_READ_WRITE
  * GROUP_DATA_ACCESS_READ_ONLY

See [MongoDB Atlas API](https://docs.atlas.mongodb.com/reference/user-roles/#project-roles) Documentation for more information.