terraform {
  required_providers {
    redfish = {
      version = "~> 1.0.0"
      source  = "registry.terraform.io/dell/redfish"
    }
  }
}

data "redfish_virtual_media" "vm" {
  for_each = var.rack1

  redfish_server {
    user         = each.value.user
    password     = each.value.password
    endpoint     = each.value.endpoint
    ssl_insecure = each.value.ssl_insecure
  }
}

output "virtual_media" {
  value = data.redfish_virtual_media.vm
  sensitive = true
}
