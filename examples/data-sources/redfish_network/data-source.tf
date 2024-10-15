/*
Copyright (c) 2024 Dell Inc., or its subsidiaries. All Rights Reserved.

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

data "redfish_network" "nic_example" {
  for_each = var.rack1

  redfish_server {
    # Alias name for server BMCs. The key in provider's `redfish_servers` map
    # `redfish_alias` is used to align with enhancements to password management.
    # When using redfish_alias, provider's `redfish_servers` is required.
    redfish_alias = each.key

    user         = each.value.user
    password     = each.value.password
    endpoint     = each.value.endpoint
    ssl_insecure = each.value.ssl_insecure
  }

  nic_filter {
    systems = [
      {
        system_id = "System.Embedded.1"
        network_adapters = [
          {
            network_adapter_id          = "FC.Slot.1"
            network_port_ids            = ["FC.Slot.1-2"]
            network_device_function_ids = ["FC.Slot.1-2"]
          },
          {
            network_adapter_id          = "NIC.Integrated.1"
            network_port_ids            = ["NIC.Integrated.1-1", "NIC.Integrated.1-2"]
            network_device_function_ids = ["NIC.Integrated.1-3-1", "NIC.Integrated.1-2-1"]
        }]
    }]
  }


}

output "nic_example" {
  value     = data.redfish_network.nic_example
  sensitive = true
}
