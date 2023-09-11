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

resource "redfish_storage_volume" "volume" {
  for_each = var.rack1

  redfish_server {
    user         = each.value.user
    password     = each.value.password
    endpoint     = each.value.endpoint
    ssl_insecure = each.value.ssl_insecure
  }

  storage_controller_id = "RAID.Integrated.1-1"
  volume_name           = "TerraformVol"
  volume_type           = "NonRedundant"
  drives                = ["Solid State Disk 0:0:1"]
  settings_apply_time   = "Immediate"
  reset_type            = "PowerCycle"
  reset_timeout         = 100
  volume_job_timeout    = 1200
  capacity_bytes        = 1073323222
  optimum_io_size_bytes = 131072
  read_cache_policy     = "AdaptiveReadAhead"
  write_cache_policy    = "UnprotectedWriteBack"
  disk_cache_policy     = "Disabled"

  lifecycle {
    ignore_changes = [
      capacity_bytes,
      volume_type
    ]
  }
}
