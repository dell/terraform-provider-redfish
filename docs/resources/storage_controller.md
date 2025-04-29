---
# Copyright (c) 2024 Dell Inc., or its subsidiaries. All Rights Reserved.
#
# Licensed under the Mozilla Public License Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://mozilla.org/MPL/2.0/
#
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

title: "redfish_storage_controller resource"
linkTitle: "redfish_storage_controller"
page_title: "redfish_storage_controller Resource - terraform-provider-redfish"
subcategory: ""
description: |-
  This Terraform resource is used to configure the storage controller. We can read the existing configurations or modify them using this resource.
---

# redfish_storage_controller (Resource)

This Terraform resource is used to configure the storage controller. We can read the existing configurations or modify them using this resource.

## Example Usage

variables.tf
```terraform
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

variable "rack1" {
  type = map(object({
    user         = string
    password     = string
    endpoint     = string
    ssl_insecure = bool
  }))
}
```

terraform.tfvars
```terraform
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

rack1 = {
  "my-server-1" = {
    user         = "admin"
    password     = "passw0rd"
    endpoint     = "https://my-server-1.myawesomecompany.org"
    ssl_insecure = true
  },
  "my-server-2" = {
    user         = "admin"
    password     = "passw0rd"
    endpoint     = "https://my-server-2.myawesomecompany.org"
    ssl_insecure = true
  },
}
```

provider.tf
```terraform
/*
Copyright (c) 2024-2025 Dell Inc., or its subsidiaries. All Rights Reserved.

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
      version = "1.6.0"
      source  = "registry.terraform.io/dell/redfish"
    }
  }
}

provider "redfish" {
  # `redfish_servers` is used to align with enhancements to password management.
  # Map of server BMCs with their alias keys and respective user credentials.
  # This is required when resource/datasource's `redfish_alias` is not null
  redfish_servers = var.rack1
}
```

main.tf
```terraform
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
    # Alias name for server BMCs. The key in provider's `redfish_servers` map
    # `redfish_alias` is used to align with enhancements to password management.
    # When using redfish_alias, provider's `redfish_servers` is required.
    redfish_alias = each.key

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
  # If server generation is lesser than 17G, accepted values: `Immediate`, `OnReset`, `AtMaintenanceWindowStart`, `InMaintenanceWindowOnReset`.
  # If server generation is 17G and above, accepted values: `Immediate`, `OnReset`.
  # When updating `controller_mode`, ensure that the `apply_time` is `OnReset`.
  # When updating `security`, ensure that the `apply_time` is `Immediate` or `OnReset`.
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
  #   start_time = "2024-10-15T22:45:00-05:00"

  #   # duration in seconds for the maintenance_window
  #   duration = 600
  # }

  # Please update any one out of `storage_controller` and `security` at a time.
  # In 17G, for `PERC H365i Front`, only the following attributes under `storage_controller` are configurable:
  #   - `consistency_check_rate_percent`
  #   - `background_initialization_rate_percent`
  # In 17G, for `PERC H965i Front`, only the following attributes under `storage_controller` are configurable:
  #   - `consistency_check_rate_percent`
  #   - `background_initialization_rate_percent`
  #   - `reconstruct_rate_percent`
  # For the above mentioned storage controllers, the other attributes under `storage_controller` need to be commented.
  storage_controller = {
    oem = {
      dell = {
        dell_storage_controller = {
          # Controller Mode. 
          # If server generation is lesser than 17G, accepted values: `RAID`, `HBA`.
          # If server generation is 17G and above, `controller_mode` need to be commented.
          # Note: In 17G and above, controller mode is a read-only property that depends upon the controller personality and hence cannot be updated.
          # In lesser than 17G, when updating `controller_mode`:
          #   - the `apply_time` should be `OnReset`
          #   - no other attributes from `storage_controller` or `security` should be updated.
          # Specifically when updating to `HBA`:
          #   - the `enhanced_auto_import_foreign_configuration_mode` attribute needs to be commented.
          #   - ensure that the security key is not present, if present first delete it using `RemoveControllerKey` action.
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
  # When updating `security`, ensure that the `apply_time` is `Immediate` or `OnReset`.
  # In lesser than 17G, when updating `controller_mode` to `HBA`, ensure that the security key is not present.
  security = {
    # Action.
    # If server generation is lesser than 17G, accepted values: `SetControllerKey`, `ReKey`, `RemoveControllerKey`.
    # If server generation is 17G and above, accepted values: `EnableSecurity`, `DisableSecurity`.
    # Note: In 17G and above, before enabling security ensure that the SEKM license is imported and SEKM/iLKM is configured.
    # action = "ReKey"

    # When `action` is set to `SetControllerKey`:
    #   - `key_id` and `key` need to be set.
    #   - `old_key` and `mode` need to be commented.
    # When `action` is set to `ReKey`:
    #   - `key_id`, `key`, `old_key` and `mode` need to be set.
    # When `action` is set to `RemoveControllerKey`:
    #   - `key_id`, `key`, `old_key` and `mode` need to be commented.
    # When `action` is set to `EnableSecurity`:
    #   - `key_id`, `key`, `old_key` and `mode` need to be commented.
    # When `action` is set to `DisableSecurity`:
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
```

After the successful execution of the above resource block, the storage controller would have been configured. More details can be verified through state file.

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `apply_time` (String) Apply time of the storage controller attributes. (Update Supported) If server generation is lesser than 17G, accepted values: `Immediate`, `OnReset`, `AtMaintenanceWindowStart`, `InMaintenanceWindowOnReset`. If server generation is 17G and above, accepted values: `Immediate`, `OnReset`. Immediate: allows the user to immediately reboot the host and apply the changes. OnReset: allows the user to apply the changes on the next reboot of the host server. AtMaintenanceWindowStart: allows the user to apply at the start of a maintenance window as specified in `maintenance_window`. InMaintenanceWindowOnReset: allows to apply after a manual reset but within the maintenance window as specified in `maintenance_window`. When updating `controller_mode`, ensure that the `apply_time` is `OnReset`. When updating `security`, ensure that the `apply_time` is `Immediate` or `OnReset`.
- `controller_id` (String) ID of the storage controller
- `storage_id` (String) ID of the storage

### Optional

- `job_timeout` (Number) `job_timeout` is the time in seconds that the provider waits for the resource update job to becompleted before timing out. (Update Supported) Default value is 1200 seconds.`job_timeout` is applicable only when `apply_time` is `Immediate` or `OnReset`.
- `maintenance_window` (Attributes) This option allows you to schedule the maintenance window. (Update Supported)This is required when `apply_time` is `AtMaintenanceWindowStart` or `InMaintenanceWindowOnReset` . (see [below for nested schema](#nestedatt--maintenance_window))
- `redfish_server` (Block List) List of server BMCs and their respective user credentials (see [below for nested schema](#nestedblock--redfish_server))
- `reset_timeout` (Number) Reset Timeout. Default value is 120 seconds. (Update Supported)
- `reset_type` (String) Reset Type. (Update Supported) Accepted values: `ForceRestart`, `GracefulRestart`, `PowerCycle`. Default value is `ForceRestart`.
- `security` (Attributes) This consists of the attributes to configure the security of the storage controller. Please update any one out of `security` and `storage_controller` at a time. When updating `security`, ensure that the `apply_time` is `Immediate` or `OnReset`. When updating `controller_mode` to `HBA`, ensure that the security key is not present. (see [below for nested schema](#nestedatt--security))
- `storage_controller` (Attributes) This consists of the attributes to configure the storage controller. Please update any one out of `storage_controller` and `security` at a time. In 17G, for `PERC H365i Front`, only the following attributes under `storage_controller` are configurable: `consistency_check_rate_percent`, `background_initialization_rate_percent`. In 17G, for `PERC H965i Front`, only the following attributes under `storage_controller` are configurable: `consistency_check_rate_percent`, `background_initialization_rate_percent`, `reconstruct_rate_percent`. (see [below for nested schema](#nestedatt--storage_controller))
- `system_id` (String) ID of the system resource. If the value for system ID is not provided, the resource picks the first system available from the iDRAC.

### Read-Only

- `id` (String) ID of the storage controller resource

<a id="nestedatt--maintenance_window"></a>
### Nested Schema for `maintenance_window`

Required:

- `duration` (Number) The duration in seconds for the maintenance window. (Update Supported)
- `start_time` (String) The start time for the maintenance window to be scheduled. (Update Supported)The format is YYYY-MM-DDThh:mm:ss<offset>. <offset> is the time offset from UTC that the current timezone set in iDRAC in the format: +05:30 for IST.


<a id="nestedblock--redfish_server"></a>
### Nested Schema for `redfish_server`

Optional:

- `endpoint` (String) Server BMC IP address or hostname
- `password` (String, Sensitive) User password for login
- `redfish_alias` (String) Alias name for server BMCs. The key in provider's `redfish_servers` map
- `ssl_insecure` (Boolean) This field indicates whether the SSL/TLS certificate must be verified or not
- `user` (String) User name for login


<a id="nestedatt--security"></a>
### Nested Schema for `security`

Optional:

- `action` (String) Action to create/change/delete the security key, if server generation is lesser than 17G. Accepted values: `SetControllerKey`, `ReKey`, `RemoveControllerKey`. Action to enable/disable the security, if server generation is 17G and above. Accepted values: `EnableSecurity`, `DisableSecurity`. Note: In 17G and above, before enabling security ensure that the SEKM license is imported and SEKM/iLKM is configured. In lesser than 17G, the `SetControllerKey` action is used to set the key on controllers and set the controller in Local key Management (LKM) to encrypt the drives. In lesser than 17G, the `ReKey` action resets the key on the controller that support encryption of the of drives. In lesser than 17G, the `RemoveControllerKey` method erases the encryption key on controller. CAUTION: All encrypted drives shall be erased. In 17G and above, the `EnableSecurity` action is used to enable the security. In 17G and above, the `DisableSecurity` action is used to disable the security.
- `key` (String) New controller key.
- `key_id` (String) Key Identifier that describes the key. The Key ID shall be maximum of 32 characters in length and should not have any spaces.
- `mode` (String) Encryption mode of the controller: Local Key Management(LKM)/Secure Enterprise Key Manager(SEKM), if server generation is lesser than 17G. If server generation is lesser than 17G, the accepted values are: `LKM`, `SEKM`. Encryption mode of the controller: Enabled/Disabled, if server generation is 17G and above. If server generation is 17G and above, it will be set to `Enabled`, if SEKM license is imported, SEKM/iLKM is configured and `EnableSecurity` action has been performed successfully. It will be set to `Disabled`, if SEKM license is not imported or SEKM/iLKM is not configured or `EnableSecurity` action has not yet been performed or `DisableSecurity` action has been performed successfully.
- `old_key` (String) Old controller key.


<a id="nestedatt--storage_controller"></a>
### Nested Schema for `storage_controller`

Optional:

- `controller_rates` (Attributes) This type describes the various controller rates used for processes such as volume rebuild or consistency checks. (see [below for nested schema](#nestedatt--storage_controller--controller_rates))
- `oem` (Attributes) The OEM extension to the StorageController resource. (see [below for nested schema](#nestedatt--storage_controller--oem))

<a id="nestedatt--storage_controller--controller_rates"></a>
### Nested Schema for `storage_controller.controller_rates`

Optional:

- `consistency_check_rate_percent` (Number) This property describes the controller rate for consistency check
- `rebuild_rate_percent` (Number) This property describes the controller rate for volume rebuild


<a id="nestedatt--storage_controller--oem"></a>
### Nested Schema for `storage_controller.oem`

Optional:

- `dell` (Attributes) Dell (see [below for nested schema](#nestedatt--storage_controller--oem--dell))

<a id="nestedatt--storage_controller--oem--dell"></a>
### Nested Schema for `storage_controller.oem.dell`

Optional:

- `dell_storage_controller` (Attributes) Dell Storage Controller (see [below for nested schema](#nestedatt--storage_controller--oem--dell--dell_storage_controller))

<a id="nestedatt--storage_controller--oem--dell--dell_storage_controller"></a>
### Nested Schema for `storage_controller.oem.dell.dell_storage_controller`

Optional:

- `background_initialization_rate_percent` (Number) Background Initialization Rate Percent
- `check_consistency_mode` (String) Check Consistency Mode. Accepted values: `Normal`, `StopOnError`.
- `controller_mode` (String) Controller Mode. Accepted values: `RAID`, `HBA` if server generation is lesser than 17G. If server generation is 17G and above, `EnhancedHBA` is another value it supports. However, in 17G and above, ensure the controller mode attribute is commented. Note: In 17G and above, controller mode is a read-only property that depends upon the controller personality and hence cannot be updated. If server generation is lesser than 17G, when updating `controller_mode`, the `apply_time` should be `OnReset` and no other attributes from `storage_controller` or `security` should be updated. Specifically, when updating `controller_mode` to `HBA`, the `enhanced_auto_import_foreign_configuration_mode` attribute needs to be commented and also ensure that the security key is not present, if present first delete it using `RemoveControllerKey` action.
- `copyback_mode` (String) Copyback Mode. Accepted values: `On`, `OnWithSMART`, `Off`.
- `enhanced_auto_import_foreign_configuration_mode` (String) Enhanced Auto Import Foreign Configuration Mode. Accepted values: `Disabled`, `Enabled`. When updating `controller_mode` to `HBA`, this attribute needs to be commented.
- `load_balance_mode` (String) Load Balance Mode. Accepted values: `Automatic`, `Disabled`.
- `patrol_read_mode` (String) Patrol Read Mode. Accepted values: `Disabled`, `Automatic`, `Manual`.
- `patrol_read_unconfigured_area_mode` (String) Patrol Read Unconfigured Area Mode. Accepted values: `Disabled`, `Enabled`.
- `reconstruct_rate_percent` (Number) Reconstruct Rate Percent

## Import

Import is supported using the following syntax:

```shell
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

# system_id is optional. If system_id is not provided, the resource picks the first one from system resources returned by the iDRAC.
terraform import redfish_storage_controller.storage_controller_example '{"storage_id":"<storage_id>","controller_id":"<controller_id>","username":"<username>","password":"<password>","endpoint":"<endpoint>","ssl_insecure":<true/false>}'

# terraform import with system_id
terraform import redfish_storage_controller.storage_controller_example '{"system_id":"<system_id>","storage_id":"<storage_id>","controller_id":"<controller_id>","username":"<username>","password":"<password>","endpoint":"<endpoint>","ssl_insecure":<true/false>}'

# terraform import with redfish_alias. When using redfish_alias, provider's `redfish_servers` is required.
# redfish_alias is used to align with enhancements to password management.
terraform import redfish_storage_controller.storage_controller_example '{"storage_id":"<storage_id>","controller_id":"<controller_id>","redfish_alias":"<redfish_alias>"}'
```

1. This will import the Storage Controller configuration into your Terraform state.
2. After successful import, you can run terraform state list to ensure the resource has been imported successfully.
3. Now, you can fill in the resource block with the appropriate arguments and settings that match the imported resource's real-world configuration.
4. Execute terraform plan to see if your configuration and the imported resource are in sync. Make adjustments if needed.
5. Finally, execute terraform apply to bring the resource fully under Terraform's management.
6. Now, the resource which was not part of terraform became part of Terraform managed infrastructure.
