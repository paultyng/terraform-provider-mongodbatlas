---
layout: "mongodbatlas"
page_title: "MongoDB Atlas: organization_api_key"
sidebar_current: "docs-mongodbatlas-datasource-organization-api-key"
description: |-
    Describes an Organization API key.
---

# mongodbatlas_organization_api_key

`mongodbatlas_organization_api_key` describes an Organization API key.

-> **NOTE:** Groups and projects are synonymous terms. You may find group_id in the official documentation.

## Example Usage

```hcl
resource "mongodbatlas_organization_api_key" "test" {
  org_id                  = "<ORGANIZATION-ID>"
  description             = "My organization API key description"
  roles                   = ["ORG_GROUP_CREATOR", "ORG_READ_ONLY"]
  access_list_cidr_blocks = ["1.1.1.1/28", "10.10.10.10/32"]
}

data "mongodbatlas_organization_api_key" "test" {
	org_id     = mongodbatlas_organization_api_key.test.org_id
	api_key_id = mongodbatlas_organization_api_key.test.api_key_id
}

```

```hcl
resource "mongodbatlas_organization_api_key" "test2" {
  org_id                  = "<ORGANIZATION-ID>"
  description             = "My organization API key description"
  roles                   = ["ORG_MEMBER"]
}

data "mongodbatlas_organization_api_key" "test2" {
	org_id     = mongodbatlas_organization_api_key.test2.org_id
	api_key_id = mongodbatlas_organization_api_key.test2.api_key_id
}
```


## Argument Reference

* `org_id` - (Required) The unique identifier for the organization you want to associate the organization API key with.
* `description` - (Required) API key description
* `roles` - (Required) List of organization roles, at least one required. See possible values [here](https://docs.atlas.mongodb.com/reference/api/apiKeys-orgs-create-one/)
* `access_list_cidr_blocks` - (Optional) API key CIDR block access list. If you only want to add a single ip, then add it with block /32


## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` -	The Terraform's unique identifier used internally for state management.
* `api_key_id` -  The unique identifier for the organization API key
* `public_key` -  The public key string that for the API key

See detailed information for arguments and attributes: [MongoDB API apiKeys](https://docs.atlas.mongodb.com/reference/api/apiKeys/)
