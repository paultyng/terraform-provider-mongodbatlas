data "mongodbatlas_project" "test" {
  name = var.project_name
}

resource "mongodbatlas_alert_configuration" "test" {
  project_id = data.mongodbatlas_project.test.id
  event_type = "OUTSIDE_METRIC_THRESHOLD"
  enabled    = true

  notification {
    type_name     = "GROUP"
    interval_min  = 5
    delay_min     = 0
    sms_enabled   = false
    email_enabled = true
  }

  matcher {
    field_name = "HOSTNAME_AND_PORT"
    operator   = "EQUALS"
    value      = "SECONDARY"
  }

  metric_threshold = {
    metric_name = "ASSERT_REGULAR"
    operator    = "LESS_THAN"
    threshold   = 99.0
    units       = "RAW"
    mode        = "AVERAGE"
  }
}

# tflint-ignore: terraform_unused_declarations
data "mongodbatlas_alert_configuration" "test" {
  project_id             = mongodbatlas_alert_configuration.test.project_id
  alert_configuration_id = mongodbatlas_alert_configuration.test.id
}
