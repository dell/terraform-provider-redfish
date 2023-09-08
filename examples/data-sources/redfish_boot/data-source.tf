/*
Copyright (c) 2023 Dell Inc., or its subsidiaries. All Rights Reserved.

Licensed under the Mozilla Public License Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://mozilla.org/MPL/2.0/


Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

terraform {
  required_providers {
    redfish = {
      source  = "registry.terraform.io/dell/redfish"
      version = "~> 1.0.0"
    }
  }
}

data "redfish_system_boot" "system_boot" {
  for_each = var.rack1

  redfish_server {
    user         = each.value.user
    password     = each.value.password
    endpoint     = each.value.endpoint
    ssl_insecure = each.value.ssl_insecure
  }

  // resource_id is an optional argument. By default, the data source uses
  // the first ComputerSystem resource present in the ComputerSystem collection
  resource_id = "System.Embedded.1"
}

output "system_boot" {
  value     = data.redfish_system_boot.system_boot
  sensitive = true
}
