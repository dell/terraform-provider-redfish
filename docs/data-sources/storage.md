---
# Copyright (c) 2023-2024 Dell Inc., or its subsidiaries. All Rights Reserved.
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

title: "redfish_storage data source"
linkTitle: "redfish_storage"
page_title: "redfish_storage Data Source - terraform-provider-redfish"
subcategory: ""
description: |-
  This Terraform datasource is used to query existing storage details from iDRAC. The information fetched from this block can be further used for resource block.
---

# redfish_storage (Data Source)

This Terraform datasource is used to query existing storage details from iDRAC. The information fetched from this block can be further used for resource block.

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

data "redfish_storage" "storage" {
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

  // by default, the data source uses the first system 
  # system_id = "System.Embedded.1"
}

output "storage_volume" {
  value     = data.redfish_storage.storage
  sensitive = true
}
```

After the successful execution of the above data block, we can see the output in the state file.

<!-- schema generated by tfplugindocs -->
## Schema

### Optional

- `controller_ids` (List of String) List of IDs of the storage controllers to be fetched.
- `controller_names` (List of String) List of names of the storage controller to be fetched.
- `redfish_server` (Block List) List of server BMCs and their respective user credentials (see [below for nested schema](#nestedblock--redfish_server))
- `system_id` (String) System ID of the system

### Read-Only

- `id` (String) ID of the storage data-source
- `storage` (Attributes List) List of storage controllers fetched. (see [below for nested schema](#nestedatt--storage))

<a id="nestedblock--redfish_server"></a>
### Nested Schema for `redfish_server`

Optional:

- `endpoint` (String) Server BMC IP address or hostname
- `password` (String, Sensitive) User password for login
- `redfish_alias` (String) Alias name for server BMCs. The key in provider's `redfish_servers` map
- `ssl_insecure` (Boolean) This field indicates whether the SSL/TLS certificate must be verified or not
- `user` (String) User name for login


<a id="nestedatt--storage"></a>
### Nested Schema for `storage`

Read-Only:

- `description` (String) description of the storage
- `drive_ids` (List of String) IDs of drives on the storage
- `drives` (List of String) Names of drives on the storage. They are in same order as in `drive_ids`, ie. `drives[i]` will be the name of the drive whose ID is given by `drive_ids[i].`
- `name` (String) name of the storage
- `oem` (Attributes) oem attributes of storage controller (see [below for nested schema](#nestedatt--storage--oem))
- `status` (Attributes) status of the storage (see [below for nested schema](#nestedatt--storage--status))
- `storage_controller_id` (String) storage controller id
- `storage_controllers` (Attributes List) storage controllers list (see [below for nested schema](#nestedatt--storage--storage_controllers))

<a id="nestedatt--storage--oem"></a>
### Nested Schema for `storage.oem`

Read-Only:

- `dell` (Attributes) dell attributes (see [below for nested schema](#nestedatt--storage--oem--dell))

<a id="nestedatt--storage--oem--dell"></a>
### Nested Schema for `storage.oem.dell`

Read-Only:

- `dell_controller` (Attributes) dell controller (see [below for nested schema](#nestedatt--storage--oem--dell--dell_controller))
- `dell_controller_battery` (Attributes) dell controller battery (see [below for nested schema](#nestedatt--storage--oem--dell--dell_controller_battery))

<a id="nestedatt--storage--oem--dell--dell_controller"></a>
### Nested Schema for `storage.oem.dell.dell_controller`

Read-Only:

- `alarm_state` (String) alarm state
- `auto_config_behavior` (String) auto config behavior
- `boot_virtual_disk_fqdd` (String) boot virtual disk fqdd
- `cache_size_in_mb` (Number) cache size in mb
- `cachecade_capability` (String) cachecade capability
- `connector_count` (Number) connector count
- `controller_description` (String) description of the controller
- `controller_firmware_version` (String) controller firmware version
- `controller_id` (String) id of controller
- `controller_name` (String) controller name
- `current_controller_mode` (String) current controller mode
- `device` (String) device
- `device_card_data_bus_width` (String) device card data bus width
- `device_card_slot_length` (String) device card slot length
- `device_card_slot_type` (String) device card slot type
- `driver_version` (String) driver version
- `encryption_capability` (String) encryption capability
- `encryption_mode` (String) encryption mode
- `key_id` (String) key id
- `last_system_inventory_time` (String) last system inventory time
- `last_update_time` (String) last update time
- `max_available_pci_link_speed` (String) max available pci link speed
- `max_possible_pci_link_speed` (String) max possible pci link speed
- `patrol_read_state` (String) patrol read state
- `pci_slot` (String) pci slot
- `persistent_hotspare` (String) persistent hotspare
- `realtime_capability` (String) realtime capability
- `rollup_status` (String) rollup status
- `sas_address` (String) sas address
- `security_status` (String) security status
- `shared_slot_assignment_allowed` (String) shared slot assignment allowed
- `sliced_vd_capability` (String) sliced vd capability
- `support_controller_boot_mode` (String) support controller boot mode
- `support_enhanced_auto_foreign_import` (String) support enhanced auto foreign import
- `support_raid_10_uneven_spans` (String) support raid 10 uneven spans
- `supports_lk_mto_sekm_transition` (String) supports lk mto sekm transition
- `t_10_pi_capability` (String) t 10 pi capability


<a id="nestedatt--storage--oem--dell--dell_controller_battery"></a>
### Nested Schema for `storage.oem.dell.dell_controller_battery`

Read-Only:

- `controller_battery_description` (String) description of the controller battery
- `controller_battery_id` (String) id of controller battery
- `controller_battery_name` (String) controller battery name
- `fqdd` (String) fqdd
- `primary_status` (String) primary_status
- `raid_state` (String) raid state




<a id="nestedatt--storage--status"></a>
### Nested Schema for `storage.status`

Read-Only:

- `health` (String) health
- `health_rollup` (String) health rollup
- `state` (String) state of the storage controller


<a id="nestedatt--storage--storage_controllers"></a>
### Nested Schema for `storage.storage_controllers`

Read-Only:

- `cache_summary` (Attributes) cache summary (see [below for nested schema](#nestedatt--storage--storage_controllers--cache_summary))
- `firmware_version` (String) firmware version
- `manufacturer` (String) manufacturer
- `model` (String) model
- `name` (String) name of the storage controller
- `speed_gbps` (Number) speed gbps
- `status` (Attributes) status of the storage controller (see [below for nested schema](#nestedatt--storage--storage_controllers--status))
- `supported_controller_protocols` (List of String) supported controller protocols
- `supported_device_protocols` (List of String) supported device protocols
- `supported_raid_types` (List of String) supported raid types

<a id="nestedatt--storage--storage_controllers--cache_summary"></a>
### Nested Schema for `storage.storage_controllers.cache_summary`

Read-Only:

- `total_cache_size_mi_b` (Number)


<a id="nestedatt--storage--storage_controllers--status"></a>
### Nested Schema for `storage.storage_controllers.status`

Read-Only:

- `health` (String) health
- `health_rollup` (String) health rollup
- `state` (String) state of the storage controller
