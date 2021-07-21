terraform {
  required_providers {
    redfish = {
      source = "dellemc/redfish"
      version = "~> 0.2.0"
    }
  }
}

provider "redfish" {
    //user = "admin"
    //password = "passw0rd"
}

data "redfish_system_boot" "system_boot" {
  for_each = var.rack1

  redfish_server {
    user = each.value.user
    password = each.value.password
    endpoint = each.value.endpoint
    ssl_insecure = each.value.ssl_insecure
  }

  // resource_id is an optional argument. By default, the data source uses
  // the first ComputerSystem resource present in the ComputerSystem collection
  resource_id = "System.Embedded.1"
}

output "system_boot" {
  value = data.redfish_system_boot.system_boot
  sensitive = true
}
