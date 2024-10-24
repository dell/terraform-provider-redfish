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

title: "redfish_network_adapter resource"
linkTitle: "redfish_network_adapter"
page_title: "redfish_network_adapter Resource - terraform-provider-redfish"
subcategory: ""
description: |-
  This Terraform resource is used to configure the port and partition network attributes on the network interface cards(NIC). We can Read the existing configurations or modify them using this resource.
---

# redfish_network_adapter (Resource)

This Terraform resource is used to configure the port and partition network attributes on the network interface cards(NIC). We can Read the existing configurations or modify them using this resource.

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
      version = "1.5.0"
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

# redfish_network_adapter is used to configure the port and partition network attributes on the network interface cards(NIC).
resource "redfish_network_adapter" "nic" {
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
  network_adapter_id         = "FC.Slot.1"
  network_device_function_id = "FC.Slot.1-2"

  # Required apply_time for creating and updating
  # Accepted values: `Immediate`, `OnReset`, `AtMaintenanceWindowStart`, `InMaintenanceWindowOnReset`
  # `Immediate` is only applicable for `oem_network_attributes`
  apply_time = "OnReset"

  # Optional system_id for creating
  # ID of the system resource. If `system_id` is not provided, the first system available from the iDRAC will be used.
  system_id = "System.Embedded.1"


  # Optional params for creating and updating

  # Reset Type. Accepted values: `ForceRestart`, `GracefulRestart`, `PowerCycle`. 
  # Default value is `ForceRestart`
  reset_type = "ForceRestart"
  # Reset Timeout. Default value is 120 seconds.
  reset_timeout = 120
  # `job_timeout` is applicable only when `apply_time` is `Immediate` or `OnReset`.
  # Default value is 1200 seconds.
  job_timeout = 1200
  # maintenance_window is required when `apply_time` is `AtMaintenanceWindowStart` or `InMaintenanceWindowOnReset`
  maintenance_window = {
    # The start time for the maintenance window to be scheduled. Format is YYYY-MM-DDThh:mm:ss<offset>.
    # <offset> is the time offset from UTC that the current timezone set in iDRAC in the format: +05:30 for IST.
    start_time = "2024-06-30T05:15:40-05:00"
    # duration in seconds for the maintenance_window
    duration = 600
  }

  # # Dictionary of network attributes and value for network device function.
  # # NOTE: `oem_network_attributes` is mutually exclusive with `network_attributes`. Please update one of `network_attributes` or `oem_network_attributes` at a time.
  # #  NOTE: Updating network_attributes property may result with an error stating the property is Read-only.
  # # This may occur if Patch method is performed to change the property to the state that the property is already in or because there is dependency of attribute values. 
  # # For example, if CHAP is disabled, MutualChap becomes a Read-only attribute.
  # network_attributes = {
  #   # The configured capability of this network device function.
  #   # Accepted values: `Disabled`, `Ethernet`, `FibreChannel`, `iSCSI`, `FibreChannelOverEthernet`, `InfiniBand`
  #   net_dev_func_type = "Ethernet"

  #   ethernet = {
  #     mac_address = "E4:43:4B:17:E0:A8"
  #     mtu_size    = 1000
  #     vlan = {
  #       vlan_id      = 100
  #       vlan_enabled = true
  #     }
  #   }

  #   # # `fibre_channel` is exclusive with `ethernet` and `iscsi_boot`
  #   # fibre_channel = {
  #   #   allow_fip_vlan_discovery = true
  #   #   fcoe_local_vlan_id       = 10
  #   #   # wwn_source is used for wwnn and wwpn connection. Accepted values: `ConfiguredLocally`, `ProvidedByFabric`
  #   #   wwn_source               = "ProvidedByFabric"
  #   #   wwnn                     = "20:00:F4:E9:D4:56:10:BF"
  #   #   wwpn                     = "21:00:F4:E9:D4:56:10:BF"
  #   #   boot_targets = [
  #   #     {
  #   #       boot_priority = 0
  #   #       lun_id        = "2"
  #   #       wwpn          = "00:00:00:00:00:00:00:00"
  #   #     }
  #   #   ]
  #   # }

  #   iscsi_boot = {
  #     # The iSCSI boot authentication method for this network device function. Accepted values: `None`, `CHAP`, `MutualCHAP`
  #     authentication_method        = "None"
  #     chap_secret                  = "secret"
  #     chap_username                = "username"
  #     # The type of IP address being populated in the iSCSIBoot IP address fields. Accepted values: `IPv4`, `IPv6`
  #     ip_address_type              = "IPv4"
  #     ip_mask_dns_via_dhcp         = true
  #     initiator_default_gateway    = "0.0.0.0"
  #     initiator_ip_address         = "0.0.0.0"
  #     initiator_name               = "iqn.1995-05.com.broadcom.iscsiboot"
  #     initiator_netmask            = "0.0.0.0"
  #     primary_dns                  = "0.0.0.0"
  #     primary_lun                  = 0
  #     primary_target_ip_address    = "0.0.0.0"
  #     primary_target_name          = "targetName"
  #     primary_target_tcp_port      = 3260
  #     primary_vlan_enable          = true
  #     primary_vlan_id              = 10
  #     secondary_dns                = "0.0.0.0"
  #     secondary_lun                = 0
  #     secondary_target_ip_address  = "0.0.0.0"
  #     secondary_target_name        = "targetName"
  #     secondary_target_tcp_port    = 3260
  #     secondary_vlan_enable        = false
  #     secondary_vlan_id            = 20
  #     target_info_via_dhcp         = false
  #     mutual_chap_secret           = "secret"  
  #     mutual_chap_username         = "username"  
  #     router_advertisement_enabled = false      
  #   }
  # }

  # oem_network_attributes to configure dell network attributes and clear pending action.
  # Note: `oem_network_attributes` is mutually exclusive with `network_attributes`.
  # Please update one of `network_attributes` or `oem_network_attributes` at a time.
  oem_network_attributes = {
    # `clear_pending` allows you to clear all the pending OEM network attributes changes.
    # `apply_time` value will be ignored and will not have any impact for `clear_pending` operation.
    clear_pending = false
    # dell network attributes. To check allowed attributes please use the datasource for dell network attributes: data.redfish_network. 
    # To get allowed values, please check /redfish/v1/Registries/NetworkAttributesRegistry_{network_device_function_id}/NetworkAttributesRegistry_{network_device_function_id}.json from a Redfish Instance
    attributes = {
      "WakeOnLan" = "Enabled"
    }
  }
}
```

After the successful execution of the above resource block, the server nic would have been configured. More details can be verified through state file.

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `apply_time` (String) Apply time of the `network_attributes` and `oem_network_attributes`. (Update Supported)Accepted values: `Immediate`, `OnReset`, `AtMaintenanceWindowStart`, `InMaintenanceWindowOnReset`. Immediate: allows the user to immediately reboot the host and apply the changes. This is only applicable for `oem_network_attributes`.OnReset: allows the user to apply the changes on the next reboot of the host server.AtMaintenanceWindowStart: allows the user to apply at the start of a maintenance window as specified in `maintenance_window`.InMaintenanceWindowOnReset: allows to apply after a manual reset but within the maintenance window as specified in `maintenance_window`.
- `network_adapter_id` (String) ID of the network adapter
- `network_device_function_id` (String) ID of the network device function

### Optional

- `job_timeout` (Number) `job_timeout` is the time in seconds that the provider waits for the resource update job to becompleted before timing out. (Update Supported) Default value is 1200 seconds.`job_timeout` is applicable only when `apply_time` is `Immediate` or `OnReset`.
- `maintenance_window` (Attributes) This option allows you to schedule the maintenance window. (Update Supported)This is required when `apply_time` is `AtMaintenanceWindowStart` or `InMaintenanceWindowOnReset` . (see [below for nested schema](#nestedatt--maintenance_window))
- `network_attributes` (Attributes) Dictionary of network attributes and value for network device function. (Update Supported)To check allowed attributes please either use the datasource for dell nic attributes: data.redfish_network or query /redfish/v1/Systems/System.Embedded.1/NetworkAdapters/{NetworkAdapterID}/NetworkDeviceFunctions/{NetworkDeviceFunctionID}/Settings. Note: `oem_network_attributes` is mutually exclusive with `network_attributes`. Please update one of network_attributes or oem_network_attributes at a time.NOTE: Updating network_attributes property may result with an error stating the property is Read-only. This may occur if Patch method is performed to change the property to the state that the property is already in or because there is dependency of attribute values. For example, if CHAP is disabled, MutualChap becomes a Read-only attribute. (see [below for nested schema](#nestedatt--network_attributes))
- `oem_network_attributes` (Attributes) oem_network_attributes to configure dell network attributes and clear pending action. (Update Supported) Note: `oem_network_attributes` is mutually exclusive with `network_attributes`. Please update one of network_attributes or oem_network_attributes at a time. (see [below for nested schema](#nestedatt--oem_network_attributes))
- `redfish_server` (Block List) List of server BMCs and their respective user credentials (see [below for nested schema](#nestedblock--redfish_server))
- `reset_timeout` (Number) Reset Timeout. Default value is 120 seconds. (Update Supported)
- `reset_type` (String) Reset Type. (Update Supported) Accepted values: `ForceRestart`, `GracefulRestart`, `PowerCycle`. Default value is `ForceRestart`.
- `system_id` (String) ID of the system resource. If the value for system ID is not provided, the resource picks the first system available from the iDRAC.

### Read-Only

- `id` (String) ID of the network interface cards resource

<a id="nestedatt--maintenance_window"></a>
### Nested Schema for `maintenance_window`

Required:

- `duration` (Number) The duration in seconds for the maintenance window. (Update Supported)
- `start_time` (String) The start time for the maintenance window to be scheduled. (Update Supported)The format is YYYY-MM-DDThh:mm:ss<offset>. <offset> is the time offset from UTC that the current timezone set in iDRAC in the format: +05:30 for IST.


<a id="nestedatt--network_attributes"></a>
### Nested Schema for `network_attributes`

Optional:

- `ethernet` (Attributes) This type describes Ethernet capabilities, status, and configuration for a network device function.  (Update Supported) (see [below for nested schema](#nestedatt--network_attributes--ethernet))
- `fibre_channel` (Attributes) This type describes Fibre Channel capabilities, status, and configuration for a network device function. (Update Supported) (see [below for nested schema](#nestedatt--network_attributes--fibre_channel))
- `iscsi_boot` (Attributes) The iSCSI boot capabilities, status, and configuration for a network device function. (Update Supported) (see [below for nested schema](#nestedatt--network_attributes--iscsi_boot))
- `net_dev_func_type` (String) The configured capability of this network device function. (Update Supported)Accepted values: `Disabled`, `Ethernet`, `FibreChannel`, `iSCSI`, `FibreChannelOverEthernet`, `InfiniBand`.

Read-Only:

- `assignable_physical_network_ports` (List of String) A reference to assignable physical network ports to this function
- `assignable_physical_ports` (List of String) A reference to assignable physical ports to this function
- `description` (String) description of the network device function
- `id` (String) ID of the network device function
- `max_virtual_functions` (Number) The number of virtual functions that are available for this network device function
- `name` (String) name of the network device function
- `net_dev_func_capabilities` (List of String) An array of capabilities for this network device function
- `odata_id` (String) OData ID for the network device function
- `physical_port_assignment` (String) A reference to a physical port assignment to this function
- `status` (Attributes) status of the network device function (see [below for nested schema](#nestedatt--network_attributes--status))

<a id="nestedatt--network_attributes--ethernet"></a>
### Nested Schema for `network_attributes.ethernet`

Optional:

- `mac_address` (String) The currently configured MAC address. (Update Supported)
- `mtu_size` (Number) The maximum transmission unit (MTU) configured for this network device function. (Update Supported)
- `vlan` (Attributes) The attributes of a VLAN. (Update Supported) (see [below for nested schema](#nestedatt--network_attributes--ethernet--vlan))

Read-Only:

- `permanent_mac_address` (String) The permanent MAC address assigned to this function

<a id="nestedatt--network_attributes--ethernet--vlan"></a>
### Nested Schema for `network_attributes.ethernet.vlan`

Optional:

- `vlan_enabled` (Boolean) An indication of whether the VLAN is enabled. (Update Supported)
- `vlan_id` (Number) The vlan id of the network device function. (Update Supported)



<a id="nestedatt--network_attributes--fibre_channel"></a>
### Nested Schema for `network_attributes.fibre_channel`

Optional:

- `allow_fip_vlan_discovery` (Boolean) An indication of whether the FCoE Initialization Protocol (FIP) populates the FCoE VLAN ID. (Update Supported)
- `boot_targets` (Attributes List) A Fibre Channel boot target configured for a network device function. (Update Supported) (see [below for nested schema](#nestedatt--network_attributes--fibre_channel--boot_targets))
- `fcoe_local_vlan_id` (Number) The locally configured FCoE VLAN ID. (Update Supported)
- `wwn_source` (String) The configuration source of the World Wide Names (WWN) for this World Wide Node Name (WWNN) and World Wide Port Name (WWPN) connection. (Update Supported). Accepted values: `ConfiguredLocally`, `ProvidedByFabric`.
- `wwnn` (String) The currently configured World Wide Node Name (WWNN) address of this function. (Update Supported)
- `wwpn` (String) The currently configured World Wide Port Name (WWPN) address of this function. (Update Supported)

Read-Only:

- `fcoe_active_vlan_id` (Number) The active FCoE VLAN ID
- `fibre_channel_id` (String) The Fibre Channel ID that the switch assigns for this interface
- `permanent_wwnn` (String) The permanent World Wide Node Name (WWNN) address assigned to this function
- `permanent_wwpn` (String) The permanent World Wide Port Name (WWPN) address assigned to this function

<a id="nestedatt--network_attributes--fibre_channel--boot_targets"></a>
### Nested Schema for `network_attributes.fibre_channel.boot_targets`

Optional:

- `boot_priority` (Number) The relative priority for this entry in the boot targets array. (Update Supported)
- `lun_id` (String) The logical unit number (LUN) ID from which to boot on the device to which the corresponding WWPN refers. (Update Supported)
- `wwpn` (String) The World Wide Port Name (WWPN) from which to boot. (Update Supported)



<a id="nestedatt--network_attributes--iscsi_boot"></a>
### Nested Schema for `network_attributes.iscsi_boot`

Optional:

- `authentication_method` (String) The iSCSI boot authentication method for this network device function. (Update Supported)Accepted values: `None`, `CHAP`, `MutualCHAP`.
- `chap_secret` (String, Sensitive) The shared secret for CHAP authentication. (Update Supported)
- `chap_username` (String) The user name for CHAP authentication. (Update Supported)
- `initiator_default_gateway` (String) The IPv6 or IPv4 iSCSI boot default gateway. (Update Supported)
- `initiator_ip_address` (String) The IPv6 or IPv4 address of the iSCSI initiator. (Update Supported)
- `initiator_name` (String) The iSCSI initiator name. (Update Supported)
- `initiator_netmask` (String) The IPv6 or IPv4 netmask of the iSCSI boot initiator. (Update Supported)
- `ip_address_type` (String) The type of IP address being populated in the iSCSIBoot IP address fields. (Update Supported) Accepted values: `IPv4`, `IPv6`.
- `ip_mask_dns_via_dhcp` (Boolean) An indication of whether the iSCSI boot initiator uses DHCP to obtain the initiator name, IP address, and netmask. (Update Supported)
- `mutual_chap_secret` (String, Sensitive) The CHAP secret for two-way CHAP authentication. (Update Supported)
- `mutual_chap_username` (String) The CHAP user name for two-way CHAP authentication. (Update Supported)
- `primary_dns` (String) The IPv6 or IPv4 address of the primary DNS server for the iSCSI boot initiator. (Update Supported)
- `primary_lun` (Number) The logical unit number (LUN) for the primary iSCSI boot target. (Update Supported)
- `primary_target_ip_address` (String) The IPv4 or IPv6 address for the primary iSCSI boot target. (Update Supported)
- `primary_target_name` (String) The name of the iSCSI primary boot target. (Update Supported)
- `primary_target_tcp_port` (Number) The TCP port for the primary iSCSI boot target. (Update Supported)
- `primary_vlan_enable` (Boolean) An indication of whether the primary VLAN is enabled. (Update Supported)
- `primary_vlan_id` (Number) The 802.1q VLAN ID to use for iSCSI boot from the primary target. (Update Supported)
- `router_advertisement_enabled` (Boolean) An indication of whether IPv6 router advertisement is enabled for the iSCSI boot target. (Update Supported)
- `secondary_dns` (String) The IPv6 or IPv4 address of the secondary DNS server for the iSCSI boot initiator. (Update Supported)
- `secondary_lun` (Number) The logical unit number (LUN) for the secondary iSCSI boot target. (Update Supported)
- `secondary_target_ip_address` (String) The IPv4 or IPv6 address for the secondary iSCSI boot target. (Update Supported)
- `secondary_target_name` (String) The name of the iSCSI secondary boot target. (Update Supported)
- `secondary_target_tcp_port` (Number) The TCP port for the secondary iSCSI boot target. (Update Supported)
- `secondary_vlan_enable` (Boolean) An indication of whether the secondary VLAN is enabled. (Update Supported)
- `secondary_vlan_id` (Number) The 802.1q VLAN ID to use for iSCSI boot from the secondary target. (Update Supported)
- `target_info_via_dhcp` (Boolean) An indication of whether the iSCSI boot target name, LUN, IP address, and netmask should be obtained from DHCP. (Update Supported)


<a id="nestedatt--network_attributes--status"></a>
### Nested Schema for `network_attributes.status`

Read-Only:

- `health` (String) health
- `health_rollup` (String) health rollup
- `state` (String) state of the storage controller



<a id="nestedatt--oem_network_attributes"></a>
### Nested Schema for `oem_network_attributes`

Optional:

- `attributes` (Map of String) dell network attributes. (Update Supported) To check allowed attributes please either use the datasource for dell network attributes: data.redfish_network or query /redfish/v1/Chassis/System.Embedded.1/NetworkAdapters/NIC.Integrated.1/NetworkDeviceFunctions/NIC.Integrated.1-3-1/Oem/Dell/DellNetworkAttributes/NIC.Integrated.1-3-1 to get attributes for NIC. To get allowed values for those attributes, check /redfish/v1/Registries/NetworkAttributesRegistry_{network_device_function_id}/NetworkAttributesRegistry_{network_device_function_id}.json from a Redfish Instance
- `clear_pending` (Boolean) This parameter allows you to clear all the pending OEM network attributes changes. (Update Supported)`false`: does not perform any operation. `true`:  discards any pending changes to network attributes, or if a job is in scheduled state, removes the job. `apply_time` value will be ignored and will not have any impact for `clear_pending` operation.

Read-Only:

- `attribute_registry` (String) registry of the network_attributes
- `description` (String) description of network_attributes
- `id` (String) ID of the network_attributes
- `name` (String) name of the network_attributes
- `odata_id` (String) OData ID for the network_attributes


<a id="nestedblock--redfish_server"></a>
### Nested Schema for `redfish_server`

Optional:

- `endpoint` (String) Server BMC IP address or hostname
- `password` (String, Sensitive) User password for login
- `redfish_alias` (String) Alias name for server BMCs. The key in provider's `redfish_servers` map
- `ssl_insecure` (Boolean) This field indicates whether the SSL/TLS certificate must be verified or not
- `user` (String) User name for login

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
terraform import redfish_network_adapter.nic '{"network_adapter_id":"<network_adapter_id>","network_device_function_id":"<network_device_function_id>","username":"<user>","password":"<password>","endpoint":"<endpoint>","ssl_insecure":<true/false>}'

# terraform import with system_id.
terraform import redfish_network_adapter.nic '{"system_id":"<system_id>","network_adapter_id":"<network_adapter_id>","network_device_function_id":"<network_device_function_id>","username":"<user>","password":"<password>","endpoint":"<endpoint>","ssl_insecure":<true/false>}'

# terraform import with redfish_alias. When using redfish_alias, provider's `redfish_servers` is required.
# redfish_alias is used to align with enhancements to password management.
terraform import redfish_network_adapter.nic '{"network_adapter_id":"<network_adapter_id>","network_device_function_id":"<network_device_function_id>","redfish_alias":"<redfish_alias>"}'
```

1. This will import the Sever NIC configuration into your Terraform state.
2. After successful import, you can run terraform state list to ensure the resource has been imported successfully.
3. Now, you can fill in the resource block with the appropriate arguments and settings that match the imported resource's real-world configuration.
4. Execute terraform plan to see if your configuration and the imported resource are in sync. Make adjustments if needed.
5. Finally, execute terraform apply to bring the resource fully under Terraform's management.
6. Now, the resource which was not part of terraform became part of Terraform managed infrastructure.
