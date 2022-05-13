terraform {
  required_providers {
    redfish = {
      version = "~> 0.2.0"
      source  = "registry.terraform.io/dell/redfish"
    }
  }
}

provider "redfish" {}

data "redfish_firmware_inventory" "inventory" {
  for_each = var.rack1

  redfish_server {
    user         = each.value.user
    password     = each.value.password
    endpoint     = each.value.endpoint
    ssl_insecure = each.value.ssl_insecure
  }
}

output "firmware_inventory" {
  value = data.redfish_firmware_inventory.inventory
}
