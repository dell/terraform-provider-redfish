terraform {
  required_providers {
    redfish = {
      version = "~> 0.2.0"
      source  = "registry.terraform.io/dell/redfish"
    }
  }
}

provider "redfish" {}

resource "redfish_virtual_media" "vm" {
  for_each = var.rack1

  redfish_server {
    user         = each.value.user
    password     = each.value.password
    endpoint     = each.value.endpoint
    ssl_insecure = each.value.ssl_insecure
  }
  image                  = "http://linuxlib.us.dell.com/pub/redhat/RHEL8/8.8/BaseOS/x86_64/iso/RHEL-8.8.0-20230411.3-x86_64-boot.iso"
  transfer_method        = "Stream"
  transfer_protocol_type = "HTTP"
  write_protected        = true
}
