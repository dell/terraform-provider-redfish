/*
Copyright (c) 2020-2024 Dell Inc., or its subsidiaries. All Rights Reserved.

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

  storage_controller_id = "RAID.SL.3-1"

  volume_name = "TerraformVol"
  // This attribute is deprecated, and will be removed in a future release.
  // Plesae use the raid_type value instead
  // volume_type           = "Mirrored"

  // Sets the Raid level Options (RAID0, RAID1, RAID5, RAID6, RAID10, RAID50, RAID60)
  raid_type = "RAID0"

  // Name of the physical disk on which virtual disk should get created.
  drives = ["Physical Disk 0:1:0"]

  // Flag stating when to create virtual disk either "Immediate" or "OnReset"
  // For BOSS Drives this should be set to "OnReset" as reboot is needed for the virtual disk to be created
  settings_apply_time = "Immediate"

  // Reset parameters to be applied when upgrade is completed
  reset_type = "PowerCycle"

  reset_timeout = 100
  // The maximum amount of time to wait for the volume job to be completed

  volume_job_timeout = 1200

  // When creating on volumes on BOSS Controllers or with the encrypt field true this property is invalid. 
  //capacity_bytes        = 1073323222

  // When creating on volumes on BOSS Controllers or with the encrypt field true this property is invalid. 
  //optimum_io_size_bytes = 131072

  // Possible values are "Off", "ReadAhead", "AdaptiveReadAhead"
  read_cache_policy = "Off"

  // When creating on volumes on BOSS Controllers this property should be set to "WriteThrough"
  // Possible values are "ProtectedWriteBack", "WriteThrough", "UnprotectedWriteBack"
  write_cache_policy = "WriteThrough"

  // Possible values are "Disabled", "Enabled"
  disk_cache_policy = "Disabled"

  // Whether or not to encrypt the virtual disk, default to false
  // Once a virtual disk is set to encrypted status it cannot be changed
  // This flag is only supported on firmware levels 6 and above
  encrypted = true

  lifecycle {
    ignore_changes = [
      capacity_bytes,
      volume_type
    ]
  }
}
