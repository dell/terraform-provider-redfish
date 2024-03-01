/*
Copyright (c) 2023-2024 Dell Inc., or its subsidiaries. All Rights Reserved.

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

resource "redfish_boot_source_override" "boot" {
  for_each = var.rack1

  redfish_server {
    user         = each.value.user
    password     = each.value.password
    endpoint     = each.value.endpoint
    ssl_insecure = each.value.ssl_insecure
  }

  // boot source override parameters
  boot_source_override_enabled = "Once"
  boot_source_override_target  = "UefiTarget"
  boot_source_override_mode    = "UEFI"

  // Reset parameters to be applied after bios settings are applied
  reset_type    = "GracefulRestart"
  reset_timeout = "120"
  # // The maximum amount of time to wait for the bios job to be completed
  boot_source_job_timeout = "1200"
}
