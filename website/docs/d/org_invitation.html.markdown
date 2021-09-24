---
layout: "mongodbatlas"
page_title: "MongoDB Atlas: org_invitation"
sidebar_current: "docs-mongodbatlas-organization-invitation"
description: |-
    Provides a Atlas Organisation Invitation resource.
---

# mongodbatlas_database_user

`mongodbatlas_org_invitation` describes a Organisation Invitation resource. This represents an invitation for an Atlas User within an Atlas Organisation.

Each invitation has a set of roles for an Atlas user within an organization.

-> **NOTE:** Groups and projects are synonymous terms. You may find group_id in the official documentation.

## Example Usage

```hcl
resource "mongodbatlas_org_invitation" "test" {
  username    = "test-acc-username"
  org_id      = "<ORG-ID>"
  roles       = [ "GROUP_DATA_ACCESS_READ_WRITE" ]
}

data "mongodbatlas_org_user" "test" {
  org_id     = mongodbatlas_org_user.test.org_id
  username   = mongodbatlas_org_user.test.username
}
```

## Argument Reference

* `org_id` - (Required) The unique ID for the project to create the database user.
* `username` - (Required) The Atlas user's email address.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The database user's name.
* `created_at` - The date and time the invitation was created
* `expires_at` - The date and time that the invitation will expire
* `invitation_id` - The identify of the invitation in Atlas
* `roles` - List of user’s Atlas roles. The available options are:
  * ORG_OWNER
  * ORG_GROUP_CREATOR
  * ORG_BILLING_ADMIN
  * ORG_READ_ONLY
  * ORG_MEMBER

See [MongoDB Atlas API](https://docs.atlas.mongodb.com/reference/api/organization-get-one-invitation/) Documentation for more information.