terraform {
  required_providers {
    redfish = {
      version = "~> 1.0.0"
      source  = "registry.terraform.io/dell/redfish"
    }
  }
}

resource "redfish_user_account" "rr" {
  for_each = var.rack1

  redfish_server {
    user         = each.value.user
    password     = each.value.password
    endpoint     = each.value.endpoint
    ssl_insecure = each.value.ssl_insecure
  }

  user_id = "4"
  username = "test"
  password = "Test@123"
  role_id  = "Operator"
  enabled  = true
}
