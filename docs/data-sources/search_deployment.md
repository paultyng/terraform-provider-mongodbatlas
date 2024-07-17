# Data Source: mongodbatlas_search_deployment

`mongodbatlas_search_deployment` describes a search node deployment.

## Example Usages
```terraform
resource "mongodbatlas_project" "example" {
  name   = "project-name"
  org_id = var.org_id
}

resource "mongodbatlas_advanced_cluster" "example" {
  project_id   = mongodbatlas_project.example.id
  name         = "ClusterExample"
  cluster_type = "REPLICASET"

  replication_specs {
    region_configs {
      electable_specs {
        instance_size = "M10"
        node_count    = 3
      }
      provider_name = "AWS"
      priority      = 7
      region_name   = "US_EAST_1"
    }
  }
}

resource "mongodbatlas_search_deployment" "example" {
  project_id   = mongodbatlas_project.example.id
  cluster_name = mongodbatlas_advanced_cluster.example.name
  specs = [
    {
      instance_size = "S20_HIGHCPU_NVME"
      node_count    = 2
    }
  ]
}

data "mongodbatlas_search_deployment" "example" {
  project_id   = mongodbatlas_search_deployment.example.project_id
  cluster_name = mongodbatlas_search_deployment.example.cluster_name
}

output "mongodbatlas_search_deployment_id" {
  value = data.mongodbatlas_search_deployment.example.id
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `cluster_name` (String) Label that identifies the cluster to return the search nodes for.
- `project_id` (String) Unique 24-hexadecimal digit string that identifies your project.

### Read-Only

- `id` (String) Unique 24-hexadecimal digit string that identifies the search deployment.
- `specs` (Attributes List) List of settings that configure the search nodes for your cluster. This list is currently limited to defining a single element. (see [below for nested schema](#nestedatt--specs))
- `state_name` (String) Human-readable label that indicates the current operating condition of this search deployment.

<a id="nestedatt--specs"></a>
### Nested Schema for `specs`

Read-Only:

- `instance_size` (String) Hardware specification for the search node instance sizes. The [MongoDB Atlas API](https://www.mongodb.com/docs/atlas/reference/api-resources-spec/#tag/Atlas-Search/operation/createAtlasSearchDeployment) describes the valid values. More details can also be found in the [Search Node Documentation](https://www.mongodb.com/docs/atlas/cluster-config/multi-cloud-distribution/#search-tier).
- `node_count` (Number) Number of search nodes in the cluster.

For more information see: [MongoDB Atlas API - Search Node](https://www.mongodb.com/docs/atlas/reference/api-resources-spec/#tag/Atlas-Search/operation/createAtlasSearchDeployment) Documentation.