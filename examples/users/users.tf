terraform {
  required_providers {
    redfish = {
      version = "~> 0.2.0"
      source  = "registry.terraform.io/dell/redfish"
    }
  }
}

provider "redfish" {
  //user = "root"
  //password = "calvin"
}

resource "redfish_user_account" "rr" {
  for_each = var.rack1

  redfish_server {
    user         = each.value.user
    password     = each.value.password
    endpoint     = each.value.endpoint
    ssl_insecure = each.value.ssl_insecure
  }

  username = "mike"
  password = "test1234"
  role_id  = "Operator"
  enabled  = true
}
