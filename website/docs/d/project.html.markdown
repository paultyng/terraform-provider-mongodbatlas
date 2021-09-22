---
layout: "mongodbatlas"
page_title: "MongoDB Atlas: project"
sidebar_current: "docs-mongodbatlas-datasource-project"
description: |-
    Describes a Project.
---

# mongodbatlas_project

`mongodbatlas_project` describes a MongoDB Atlas Project. This represents a project that has been created.

-> **NOTE:** Groups and projects are synonymous terms. You may find group_id in the official documentation.

## Example Usage

### Using project_id attribute to query
```hcl
resource "mongodbatlas_project" "test" {
  name   = "project-name"
  org_id = "<ORG_ID>"

  teams {
    team_id    = "5e0fa8c99ccf641c722fe645"
    role_names = ["GROUP_OWNER"]

  }
  teams {
    team_id    = "5e1dd7b4f2a30ba80a70cd4rw"
    role_names = ["GROUP_READ_ONLY", "GROUP_DATA_ACCESS_READ_WRITE"]
  }
}

data "mongodbatlas_project" "test" {
  project_id = "${mongodbatlas_project.test.id}"
}
```

### Using name attribute to query
```hcl
resource "mongodbatlas_project" "test" {
  name   = "project-name"
  org_id = "<ORG_ID>"

  teams {
    team_id    = "5e0fa8c99ccf641c722fe645"
    role_names = ["GROUP_OWNER"]

  }
  teams {
    team_id    = "5e1dd7b4f2a30ba80a70cd4rw"
    role_names = ["GROUP_READ_ONLY", "GROUP_DATA_ACCESS_READ_WRITE"]
  }
  api_keys {
    api_key_id = "61003b299dda8d54a9d7d10c"
    role_names = ["GROUP_READ_ONLY"]
  }
}

data "mongodbatlas_project" "test" {
  name = mongodbatlas_project.test.name
}
```

## Argument Reference

* `project_id` - (Optional) The unique ID for the project.
* `name` - (Optional) The unique ID for the project.

~> **IMPORTANT:** Either `project_id` or `name` must be configurated.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `name` - The name of the project you want to create. (Cannot be changed via this Provider after creation.)
* `org_id` - The ID of the organization you want to create the project within.
*`cluster_count` - The number of Atlas clusters deployed in the project.
*`created` - The ISO-8601-formatted timestamp of when Atlas created the project.
* `teams.#.team_id` - The unique identifier of the team you want to associate with the project. The team and project must share the same parent organization.
* `teams.#.role_names` - Each string in the array represents a project role assigned to the team. Every user associated with the team inherits these roles.
The following are valid roles:
  * `GROUP_OWNER`
  * `GROUP_READ_ONLY`
  * `GROUP_DATA_ACCESS_ADMIN`
  * `GROUP_DATA_ACCESS_READ_WRITE`
  * `GROUP_DATA_ACCESS_READ_ONLY`
  * `GROUP_CLUSTER_MANAGER`
* `api_keys.#.api_key_id` - The unique identifier of the api key you want to associate with the project. The api key and project must share the same parent organization.
* `api_keys.#.role_names` - Each string in the array represents a project role assigned to the api key.
The following are valid roles:
  * `GROUP_OWNER`
  * `GROUP_READ_ONLY`
  * `GROUP_DATA_ACCESS_ADMIN`
  * `GROUP_DATA_ACCESS_READ_WRITE`
  * `GROUP_DATA_ACCESS_READ_ONLY`
  * `GROUP_CLUSTER_MANAGER`

See [MongoDB Atlas API - Project](https://docs.atlas.mongodb.com/reference/api/project-get-one/) - [and MongoDB Atlas API - Teams](https://docs.atlas.mongodb.com/reference/api/project-get-teams/) Documentation for more information.
