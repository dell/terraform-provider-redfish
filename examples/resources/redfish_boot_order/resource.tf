/*
Copyright (c) 2023-2025 Dell Inc., or its subsidiaries. All Rights Reserved.

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

resource "redfish_boot_order" "boot" {
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
  // sets the boot devices in the required boot order sequences
  boot_order = ["Boot0001", "Boot0000", "Boot0002", "Boot0003"]

  // Options to enable or disable the boot device. Uncomment the same and comment the boot_order to use this.
  // boot_options = [{boot_option_reference= "Boot0000", boot_option_enabled= false}]

  /* Reset parameters to be applied after bios settings are applied
     list of possible value:
      [ ForceRestart, GracefulRestart, PowerCycle]
  */
  reset_type    = "ForceRestart"
  reset_timeout = "120"
  // The maximum amount of time to wait for the bios job to be completed
  boot_order_job_timeout = "1200"

  // by default, the resource uses the first system
  # system_id = "System.Embedded.1"
}
