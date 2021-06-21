terraform {
  required_providers {
    redfish = {
      source = "dell/redfish"
    }
  }
}

provider "redfish" {}

resource "redfish_dell_idrac_attributes" "idrac" {
  for_each = var.rack1

  redfish_server {
    user = each.value.user
    password = each.value.password
    endpoint = each.value.endpoint
    ssl_insecure = each.value.ssl_insecure
  }

  attributes = {
    "Users.3.Enable" = "Disabled"
    "Users.3.UserName" = "mike"
    "Users.3.Password" = "test1234"
    "Users.3.Privilege" = 511
    "TelemetryFanSensor.1.ReportInterval" = 60
  }
}
