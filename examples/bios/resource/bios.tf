terraform {
  required_providers {
    redfish = {
      version = "1.0.0"
      source  = "registry.terraform.io/dell/redfish"
    }
  }
}

resource "redfish_bios" "bios" {
  for_each = var.rack1

  redfish_server {
    user         = each.value.user
    password     = each.value.password
    endpoint     = each.value.endpoint
    ssl_insecure = each.value.ssl_insecure
  }

  attributes = {
    "NumLock" = "On"
  }
  reset_type          = "ForceRestart"
  reset_timeout       = "120"
  bios_job_timeout    = "1200"
}