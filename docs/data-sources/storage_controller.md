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

title: "redfish_storage_controller data source"
linkTitle: "redfish_storage_controller"
page_title: "redfish_storage_controller Data Source - terraform-provider-redfish"
subcategory: ""
description: |-
  This Terraform datasource is used to query existing storage controller configuration. The information fetched from this block can be further used for resource block.
---

# redfish_storage_controller (Data Source)

This Terraform datasource is used to query existing storage controller configuration. The information fetched from this block can be further used for resource block.

## Example Usage

variables.tf
```terraform
/*
Copyright (c) 2021-2024 Dell Inc., or its subsidiaries. All Rights Reserved.

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
Copyright (c) 2022-2024 Dell Inc., or its subsidiaries. All Rights Reserved.

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

data "redfish_storage_controller" "storage_controller_example" {
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

  storage_controller_filter {
    systems = [
      {
        system_id = "System.Embedded.1"
        storages = [
          {
            storage_id     = "RAID.Integrated.1-1"
            controller_ids = ["RAID.Integrated.1-1"]
          }
        ]
      }
    ]
  }

}

output "storage_controller_example" {
  value     = data.redfish_storage_controller.storage_controller_example
  sensitive = true
}
```

After the successful execution of the above data block, we can see the output in the state file.

<!-- schema generated by tfplugindocs -->
## Schema

### Optional

- `redfish_server` (Block List) List of server BMCs and their respective user credentials (see [below for nested schema](#nestedblock--redfish_server))
- `storage_controller_filter` (Block, Optional) Storage Controller filter for systems, storages and controllers (see [below for nested schema](#nestedblock--storage_controller_filter))

### Read-Only

- `id` (String) ID of the storage controller data-source.
- `storage_controllers` (Attributes List) List of storage controllers fetched. (see [below for nested schema](#nestedatt--storage_controllers))

<a id="nestedblock--redfish_server"></a>
### Nested Schema for `redfish_server`

Optional:

- `endpoint` (String) Server BMC IP address or hostname
- `password` (String, Sensitive) User password for login
- `redfish_alias` (String) Alias name for server BMCs. The key in provider's `redfish_servers` map
- `ssl_insecure` (Boolean) This field indicates whether the SSL/TLS certificate must be verified or not
- `user` (String) User name for login


<a id="nestedblock--storage_controller_filter"></a>
### Nested Schema for `storage_controller_filter`

Optional:

- `systems` (Attributes List) Filter for systems, storages and storage controllers (see [below for nested schema](#nestedatt--storage_controller_filter--systems))

<a id="nestedatt--storage_controller_filter--systems"></a>
### Nested Schema for `storage_controller_filter.systems`

Required:

- `system_id` (String) Filter for systems

Optional:

- `storages` (Attributes List) Filter for storages and storage controllers (see [below for nested schema](#nestedatt--storage_controller_filter--systems--storages))

<a id="nestedatt--storage_controller_filter--systems--storages"></a>
### Nested Schema for `storage_controller_filter.systems.storages`

Required:

- `storage_id` (String) Filter for storages

Optional:

- `controller_ids` (Set of String) Filter for storage controllers




<a id="nestedatt--storage_controllers"></a>
### Nested Schema for `storage_controllers`

Read-Only:

- `assembly` (Attributes) A reference to a resource. (see [below for nested schema](#nestedatt--storage_controllers--assembly))
- `cache_summary` (Attributes) This type describes the cache memory of the storage controller in general detail. (see [below for nested schema](#nestedatt--storage_controllers--cache_summary))
- `controller_rates` (Attributes) This type describes the various controller rates used for processes such as volume rebuild or consistency checks. (see [below for nested schema](#nestedatt--storage_controllers--controller_rates))
- `description` (String) The description of this resource. Used for commonality in the schema definitions.
- `firmware_version` (String) The firmware version of this storage controller.
- `id` (String) The unique identifier for this resource within the collection of similar resources.
- `identifiers` (Attributes List) Any additional identifiers for a resource. (see [below for nested schema](#nestedatt--storage_controllers--identifiers))
- `links` (Attributes) The links to other resources that are related to this resource. (see [below for nested schema](#nestedatt--storage_controllers--links))
- `manufacturer` (String) The manufacturer of this storage controller.
- `model` (String) The model number for the storage controller.
- `name` (String) The name of the resource or array member.
- `odata_id` (String) The unique identifier for a resource.
- `oem` (Attributes) The OEM extension to the StorageController resource. (see [below for nested schema](#nestedatt--storage_controllers--oem))
- `speed_gbps` (Number) The maximum speed of the storage controller's device interface.
- `status` (Attributes) The status and health of a resource and its children. (see [below for nested schema](#nestedatt--storage_controllers--status))
- `supported_controller_protocols` (List of String) The supported set of protocols for communicating to this storage controller.
- `supported_device_protocols` (List of String) The protocols that the storage controller can use to communicate with attached devices.
- `supported_raid_types` (List of String) The set of RAID types supported by the storage controller.

<a id="nestedatt--storage_controllers--assembly"></a>
### Nested Schema for `storage_controllers.assembly`

Read-Only:

- `odata_id` (String) The link to the assembly associated with this storage controller.


<a id="nestedatt--storage_controllers--cache_summary"></a>
### Nested Schema for `storage_controllers.cache_summary`

Read-Only:

- `total_cache_size_mi_b` (Number)


<a id="nestedatt--storage_controllers--controller_rates"></a>
### Nested Schema for `storage_controllers.controller_rates`

Read-Only:

- `consistency_check_rate_percent` (Number) This property describes the controller rate for consistency check
- `rebuild_rate_percent` (Number) This property describes the controller rate for volume rebuild


<a id="nestedatt--storage_controllers--identifiers"></a>
### Nested Schema for `storage_controllers.identifiers`

Read-Only:

- `durable_name` (String) This property describes the durable name for the storage controller.
- `durable_name_format` (String) This property describes the durable name format for the storage controller.


<a id="nestedatt--storage_controllers--links"></a>
### Nested Schema for `storage_controllers.links`

Read-Only:

- `pcie_functions` (Attributes List) PCIeFunctions (see [below for nested schema](#nestedatt--storage_controllers--links--pcie_functions))

<a id="nestedatt--storage_controllers--links--pcie_functions"></a>
### Nested Schema for `storage_controllers.links.pcie_functions`

Read-Only:

- `odata_id` (String) The link to the PCIeFunctions



<a id="nestedatt--storage_controllers--oem"></a>
### Nested Schema for `storage_controllers.oem`

Read-Only:

- `dell` (Attributes) Dell (see [below for nested schema](#nestedatt--storage_controllers--oem--dell))

<a id="nestedatt--storage_controllers--oem--dell"></a>
### Nested Schema for `storage_controllers.oem.dell`

Read-Only:

- `dell_storage_controller` (Attributes) Dell Storage Controller (see [below for nested schema](#nestedatt--storage_controllers--oem--dell--dell_storage_controller))

<a id="nestedatt--storage_controllers--oem--dell--dell_storage_controller"></a>
### Nested Schema for `storage_controllers.oem.dell.dell_storage_controller`

Read-Only:

- `alarm_state` (String) Alarm State
- `auto_config_behavior` (String) Auto Config Behavior
- `background_initialization_rate_percent` (Number) Background Initialization Rate Percent
- `battery_learn_mode` (String) Battery Learn Mode
- `boot_virtual_disk_fqdd` (String) Boot Virtual Disk FQDD
- `cache_size_in_mb` (Number) Cache Size In MB
- `cachecade_capability` (String) Cachecade Capability
- `check_consistency_mode` (String) Check Consistency Mode
- `connector_count` (Number) Connector Count
- `controller_boot_mode` (String) Controller Boot Mode
- `controller_firmware_version` (String) Controller Firmware Version
- `controller_mode` (String) Controller Mode
- `copyback_mode` (String) Copyback Mode
- `current_controller_mode` (String) Current Controller Mode
- `device` (String) Device
- `device_card_data_bus_width` (String) Device Card Data Bus Width
- `device_card_slot_length` (String) Device Card Slot Length
- `device_card_slot_type` (String) Device Card Slot Type
- `driver_version` (String) Driver Version
- `encryption_capability` (String) Encryption Capability
- `encryption_mode` (String) Encryption Mode
- `enhanced_auto_import_foreign_configuration_mode` (String) Enhanced Auto Import Foreign Configuration Mode
- `key_id` (String) Key ID
- `last_system_inventory_time` (String) Last System Inventory Time
- `last_update_time` (String) Last Update Time
- `load_balance_mode` (String) Load Balance Mode
- `max_available_pci_link_speed` (String) Max Available PCI Link Speed
- `max_drives_in_span_count` (Number) Max Drives In Span Count
- `max_possible_pci_link_speed` (String) Max Possible PCI Link Speed
- `max_spans_in_volume_count` (Number) Max Spans In Volume Count
- `max_supported_volumes_count` (Number) Max Supported Volumes Count
- `patrol_read_iterations_count` (Number) Patrol Read Iterations Count
- `patrol_read_mode` (String) Patrol Read Mode
- `patrol_read_rate_percent` (Number) Patrol Read Rate Percent
- `patrol_read_state` (String) Patrol Read State
- `patrol_read_unconfigured_area_mode` (String) Patrol Read Unconfigured Area Mode
- `pci_slot` (String) PCI Slot
- `persistent_hotspare` (String) Persistent Hotspare
- `persistent_hotspare_mode` (String) Persistent Hotspare Mode
- `raid_mode` (String) RAID Mode
- `real_time_capability` (String) Realtime Capability
- `reconstruct_rate_percent` (Number) Reconstruct Rate Percent
- `rollup_status` (String) Rollup Status
- `sas_address` (String) SAS Address
- `security_status` (String) Security Status
- `shared_slot_assignment_allowed` (String) Shared Slot Assignment Allowed
- `sliced_vd_capability` (String) Sliced VD Capability
- `spindown_idle_time_seconds` (Number) Spindown Idle Time Seconds
- `support_controller_boot_mode` (String) Support Controller Boot Mode
- `support_enhanced_auto_foreign_import` (String) Support Enhanced Auto Foreign Import
- `support_raid10_uneven_spans` (String) Support RAID10 Uneven Spans
- `supported_initialization_types` (List of String) Supported Initialization Types
- `supports_lkm_to_sekm_transition` (String) Supports LKM to SEKM Transition
- `t10_pi_capability` (String) T10 PI Capability




<a id="nestedatt--storage_controllers--status"></a>
### Nested Schema for `storage_controllers.status`

Read-Only:

- `health` (String) health
- `health_rollup` (String) health rollup
- `state` (String) state of the storage controller

