---
layout: "mongodbatlas"
page_title: "MongoDB Atlas: maintenance_window"
sidebar_current: "docs-mongodbatlas-datasource-maintenance_window"
description: |-
    Provides a Maintenance Window Datasource.
---

# mongodbatlas_maintenance_window

`mongodbatlas_maintenance_window` provides a Maintenance Window entry datasource. Gets information regarding the configured maintenance window for a MongoDB Atlas project.

-> **NOTE:** Groups and projects are synonymous terms. You may find `groupId` in the official documentation.

## Examples Usage

```hcl
resource "mongodbatlas_maintenance_window" "test" {
  project_id  = "<your-project-id>"
  day_of_week = 3
  hour_of_day = 4
  auto_defer_once_enabled = true
}

data "mongodbatlas_maintenance_window" "test" {
  project_id = mongodbatlas_maintenance_window.test.id
}
```


```hcl
resource "mongodbatlas_maintenance_window" "test" {
  project_id  = "<your-project-id>"
  start_asap  = true 
}

data "mongodbatlas_maintenance_window" "test" {
  project_id = mongodbatlas_maintenance_window.test.id
}
```

## Argument Reference

* `project_id` - The unique identifier of the project for the Maintenance Window.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `day_of_week` - Day of the week when you would like the maintenance window to start as a 1-based integer: S=1, M=2, T=3, W=4, T=5, F=6, S=7.
* `hour_of_day` - Hour of the day when you would like the maintenance window to start. This parameter uses the 24-hour clock, where midnight is 0, noon is 12  (Time zone is UTC).
* `start_asap` - Flag indicating whether project maintenance has been directed to start immediately. If you request that maintenance begin immediately, this field returns true from the time the request was made until the time the maintenance event completes.
* `number_of_deferrals` - Number of times the current maintenance event for this project has been deferred, you can set a maximum of 2 deferrals.
* `auto_defer_once_enabled` - Flag that indicates whether you want to defer all maintenance windows one week they would be triggered.
For more information see: [MongoDB Atlas API Reference.](https://docs.atlas.mongodb.com/reference/api/maintenance-windows/)