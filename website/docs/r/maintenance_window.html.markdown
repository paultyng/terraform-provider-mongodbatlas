---
layout: "mongodbatlas"
page_title: "MongoDB Atlas: maintenance_window"
sidebar_current: "docs-mongodbatlas-resource-maintenance_window"
description: |-
    Provides an Maintenance Window resource.
---

# mongodbatlas_maintenance_window

`mongodbatlas_maintenance_window` provides a resource to schedule a maintenance window for your MongoDB Atlas Project and/or set to defer a scheduled maintenance up to two times.

-> **NOTE:** Groups and projects are synonymous terms. You may find `groupId` in the official documentation.

## Maintenance Window Considerations:
- Urgent Maintenance Activities Cannot Wait: Urgent maintenance activities such as security patches cannot wait for your chosen window. Atlas will start those maintenance activities when needed.

Once maintenance is scheduled for your cluster, you cannot change your maintenance window until the current maintenance efforts have completed.
- Maintenance Requires Replica Set Elections: Atlas performs maintenance the same way as the manual maintenance procedure. This requires at least one replica set election during the maintenance window per replica set.
- Maintenance Starts As Close to the Hour As Possible: Maintenance always begins as close to the scheduled hour as possible, but in-progress cluster updates or expected system issues could delay the start time.


## Example Usage

```hcl
  resource "mongodbatlas_maintenance_window" "test" {
    project_id  = "<your-project-id>"
    day_of_week = 3
    hour_of_day = 4
  }

```

```hcl
  resource "mongodbatlas_maintenance_window" "test" {
    project_id = "<your-project-id>"
    defer      = true
  }
```

## Argument Reference

* `project_id` - The unique identifier of the project for the Maintenance Window.
* `day_of_week` - Day of the week when you would like the maintenance window to start as a 1-based integer: S=1, M=2, T=3, W=4, T=5, F=6, S=7.
* `hour_of_day` - Hour of the day when you would like the maintenance window to start. This parameter uses the 24-hour clock, where midnight is 0, noon is 12 (Time zone is UTC).
* `start_asap` - Flag indicating whether project maintenance has been directed to start immediately. If you request that maintenance begin immediately, this field returns true from the time the request was made until the time the maintenance event completes.
* `number_of_deferrals` - Number of times the current maintenance event for this project has been deferred, you can set a maximum of 2 deferrals.
* `defer` - Defer maintenance for the given project for one week.
* `auto_defer` - Automatically defers any maintenance for the given project for one week when a maintenance event is available.
* `auto_defer_once_enabled` - Flag that indicates whether you want to defer all maintenance windows one week they would be triggered.

-> **NOTE:** The `start_asap` attribute can't be used because of breaks the Terraform flow, but you can enable via API.

## Import

Maintenance Window entries can be imported using project project_id, in the format `PROJECTID`, e.g.

```
$ terraform import mongodbatlas_maintenance_window.test 5d0f1f73cf09a29120e173cf
```

For more information see: [MongoDB Atlas API Reference.](https://docs.atlas.mongodb.com/reference/api/maintenance-windows/)
