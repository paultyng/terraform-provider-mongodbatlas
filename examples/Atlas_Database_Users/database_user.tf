# DATABASE USER
resource "mongodbatlas_database_user" "user1" {
  username           = var.user[0]
  password           = var.password[0]
  project_id         = mongodbatlas_project.project1.id
  auth_database_name = "admin"

  roles {
    role_name     = "readWrite"
    database_name = var.database_name[0]
  }
  labels {
    key   = "Name"
    value = "DB User1"
  }

  scopes {
    name = mongodbatlas_cluster.cluster.name
    type = "CLUSTER"
  }
}
output "user1" {
  value = mongodbatlas_database_user.user1.username
}
# DATA LAKE USER
resource "mongodbatlas_database_user" "user2" {
  username           = var.user[1]
  password           = var.password[1]
  project_id         = mongodbatlas_project.project1.id
  auth_database_name = "admin"

  roles {
    role_name     = "readWrite"
    database_name = var.database_name[1]
  }
  labels {
    key   = "Name"
    value = "DB User2"
  }

  scopes {
    name = var.data_lake
    type = "DATA_LAKE"
  }
}
output "user2" {
  value = mongodbatlas_database_user.user2.username
}
