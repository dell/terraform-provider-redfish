terraform {
  required_providers {
    redfish = {
      version = "~> 0.2.0"
      source  = "registry.terraform.io/dell/redfish"
    }
  }
}

provider "redfish" {}

resource "redfish_storage_volume" "volume" {
  for_each = var.rack1

  redfish_server {
    user         = each.value.user
    password     = each.value.password
    endpoint     = each.value.endpoint
    ssl_insecure = each.value.ssl_insecure
  }

  storage_controller_id = "RAID.Integrated.1-1"
  volume_name           = "TerraformVol"
  volume_type           = "NonRedundant"
  drives                = ["Solid State Disk 0:0:1"]
  settings_apply_time   = "Immediate"
  reset_type = "PowerCycle"
  reset_timeout = 100
  volume_job_timeout = 1200
  capacity_bytes = 1073323222
  optimum_io_size_bytes = 131072
  read_cache_policy = "AdaptiveReadAhead"
  write_cache_policy = "UnprotectedWriteBack"
  disk_cache_policy = "Disabled"

  lifecycle {
    ignore_changes = [
      capacity_bytes,
      volume_type
    ]
  }
}
