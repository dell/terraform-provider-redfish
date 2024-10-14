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

# redfish_storage_controller is used to configure the storage controller
resource "redfish_storage_controller" "storage_controller_example" {
  for_each = var.rack1

  redfish_server {
    user         = each.value.user
    password     = each.value.password
    endpoint     = each.value.endpoint
    ssl_insecure = each.value.ssl_insecure
  }

  # Required params for creating

  # Storage ID.
  storage_id = "RAID.Integrated.1-1"
  # Controller ID.
  controller_id = "RAID.Integrated.1-1"

  # Apply Time. Required for creating and updating.
  # Accepted values: `Immediate`, `OnReset`, `AtMaintenanceWindowStart`, `InMaintenanceWindowOnReset`
  apply_time = "Immediate"

  # System ID. Optional for creating.
  # ID of the system resource. If `system_id` is not provided, the first system available from the iDRAC will be used.
  # system_id = "System.Embedded.1"

  # Optional params for creating and updating

  # Reset Type. 
  # Accepted values: `ForceRestart`, `GracefulRestart`, `PowerCycle`. 
  # Default value is `ForceRestart`.
  # reset_type = "ForceRestart"

  # Reset Timeout. 
  # Default value is 120 seconds.
  # reset_timeout = 120

  # Job Timeout.
  # It is applicable only when `apply_time` is `Immediate` or `OnReset`.
  # Default value is 1200 seconds.
  # job_timeout = 1200

  # Maintenance Window.
  # It is required when `apply_time` is `AtMaintenanceWindowStart` or `InMaintenanceWindowOnReset`.
  # maintenance_window = {
  #   # The start time for the maintenance window to be scheduled. Format is YYYY-MM-DDThh:mm:ss<offset>.
  #   # <offset> is the time offset from UTC that the current timezone set in iDRAC in the format: +05:30 for IST.
  #   start_time = "2024-06-30T05:15:40-05:00"

  #   # duration in seconds for the maintenance_window
  #   duration = 600
  # }

  # Please update any one out of `storage_controller` and `security` at a time.
  storage_controller = {
    oem = {
      dell = {
        dell_storage_controller = {
          # Controller Mode. 
          # Accepted values: `RAID`, `HBA`.
          # When updating `controller_mode`:
          #   - the `apply_time` should be `OnReset` or `InMaintenanceWindowOnReset`
          #   - no other attributes from `storage_controller` or `security` should be updated.
          # Specifically when updating to `HBA`:
          #   - the `enhanced_auto_import_foreign_configuration_mode` attribute needs to be commented.
          # controller_mode = "RAID"

          # Check Consistency Mode. 
          # Accepted values: `Normal`, `StopOnError`.
          # check_consistency_mode = "Normal"

          # Copyback Mode.
          # Accepted values: `On`, `OnWithSMART`, `Off`.
          # copyback_mode = "On"

          # Load Balance Mode. 
          # Accepted values: `Automatic`, `Disabled`.
          # load_balance_mode = "Disabled"

          # Enhanced Auto Import Foreign Configuration Mode.
          # Accepted values: `Disabled`, `Enabled`.
          # When updating `controller_mode` to `HBA`, this attribute needs to be commented.
          # enhanced_auto_import_foreign_configuration_mode = "Disabled"

          # Patrol Read Unconfigured Area Mode.
          # Accepted values: `Disabled`, `Enabled`.
          # patrol_read_unconfigured_area_mode = "Enabled"

          # Patrol Read Mode.
          # Accepted values: `Disabled`, `Automatic`, `Manual`.
          # patrol_read_mode = "Automatic"

          # Background Initialization Rate.
          # background_initialization_rate_percent = 30
          # Reconstruct Rate.
          # reconstruct_rate_percent = 30
        }
      }
    }

    controller_rates = {
      # Consistency Check Rate.
      # consistency_check_rate_percent = 30
      # Rebuild Rate.
      # rebuild_rate_percent = 30
    }

  }

  # Please update any one out of `security` and `storage_controller` at a time.
  security = {
    # Action.
    # Accepted values: `SetControllerKey`, `ReKey`, `RemoveControllerKey`.
    # action = "ReKey"

    # When `action` is set to `SetControllerKey`:
    #   - `key_id` and `key` need to be set.
    #   - `old_key` and `mode` need to be commented.
    # When `action` is set to `ReKey`:
    #   - `key_id`, `key`, `old_key` and `mode` need to be set.
    # When `action` is set to `RemoveControllerKey`:
    #   - `key_id`, `key`, `old_key` and `mode` need to be commented.

    # Key ID.
    # key_id = "testkey"

    # Key.
    # key = "Test123##"
    # Old Key.
    # old_key = "Test123###"

    # Mode.
    # Accepted values: `LKM`, `SEKM`.
    # mode = "LKM"
  }

}