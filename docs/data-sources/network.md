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

title: "redfish_network data source"
linkTitle: "redfish_network"
page_title: "redfish_network Data Source - terraform-provider-redfish"
subcategory: ""
description: |-
  This Terraform datasource is used to query existing network interface cards(NIC) configuration including network adapters, network ports, network device functions and their OEM attributes. The information fetched from this block can be further used for resource block.
---

# redfish_network (Data Source)

This Terraform datasource is used to query existing network interface cards(NIC) configuration including network adapters, network ports, network device functions and their OEM attributes. The information fetched from this block can be further used for resource block.

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
```

After the successful execution of the above data block, we can see the output in the state file.

<!-- schema generated by tfplugindocs -->
## Schema

### Optional

- `nic_filter` (Block, Optional) NIC filter for systems, nework adapters, network ports and network device functions (see [below for nested schema](#nestedblock--nic_filter))
- `redfish_server` (Block List) List of server BMCs and their respective user credentials (see [below for nested schema](#nestedblock--redfish_server))

### Read-Only

- `id` (String) ID of the network interface cards data-source
- `network_interfaces` (Attributes List) List of network interface cards(NIC) fetched. (see [below for nested schema](#nestedatt--network_interfaces))
- `nic_attributes` (Map of String) nic.* attributes in Dell iDRAC attributes.

<a id="nestedblock--nic_filter"></a>
### Nested Schema for `nic_filter`

Optional:

- `systems` (Attributes List) Filter for systems, nework adapters, network ports and network device functions (see [below for nested schema](#nestedatt--nic_filter--systems))

<a id="nestedatt--nic_filter--systems"></a>
### Nested Schema for `nic_filter.systems`

Required:

- `system_id` (String) Filter for systems

Optional:

- `network_adapters` (Attributes List) Filter for nework adapters, network ports and network device functions (see [below for nested schema](#nestedatt--nic_filter--systems--network_adapters))

<a id="nestedatt--nic_filter--systems--network_adapters"></a>
### Nested Schema for `nic_filter.systems.network_adapters`

Required:

- `network_adapter_id` (String) Filter for network adapters

Optional:

- `network_device_function_ids` (Set of String) Filter for network device functions
- `network_port_ids` (Set of String) Filter for network ports




<a id="nestedblock--redfish_server"></a>
### Nested Schema for `redfish_server`

Optional:

- `endpoint` (String) Server BMC IP address or hostname
- `password` (String, Sensitive) User password for login
- `redfish_alias` (String) Alias name for server BMCs. The key in provider's `redfish_servers` map
- `ssl_insecure` (Boolean) This field indicates whether the SSL/TLS certificate must be verified or not
- `user` (String) User name for login


<a id="nestedatt--network_interfaces"></a>
### Nested Schema for `network_interfaces`

Read-Only:

- `description` (String) Description of the NIC data-source
- `id` (String) ID of the NIC data-source
- `name` (String) Name of the NIC data-source
- `network_adapter` (Attributes) Network adapter fetched (see [below for nested schema](#nestedatt--network_interfaces--network_adapter))
- `network_device_functions` (Attributes List) List of network device functions fetched (see [below for nested schema](#nestedatt--network_interfaces--network_device_functions))
- `network_ports` (Attributes List) List of network ports fetched (see [below for nested schema](#nestedatt--network_interfaces--network_ports))
- `odata_id` (String) OData ID for the NIC instance
- `status` (Attributes) The status and health of a resource and its children (see [below for nested schema](#nestedatt--network_interfaces--status))

<a id="nestedatt--network_interfaces--network_adapter"></a>
### Nested Schema for `network_interfaces.network_adapter`

Read-Only:

- `controllers` (Attributes List) A network controller ASIC that makes up part of a network adapter (see [below for nested schema](#nestedatt--network_interfaces--network_adapter--controllers))
- `description` (String) Description of the network adapter
- `id` (String) ID of the network adapter
- `manufacturer` (String) The manufacturer or OEM of this network adapter
- `model` (String) The model string for this network adapter
- `name` (String) Name of the network adapter
- `odata_id` (String) OData ID for the network adapter
- `part_number` (String) Part number for this network adapter
- `serial_number` (String) The serial number for this network adapter
- `status` (Attributes) The status and health of a resource and its children (see [below for nested schema](#nestedatt--network_interfaces--network_adapter--status))

<a id="nestedatt--network_interfaces--network_adapter--controllers"></a>
### Nested Schema for `network_interfaces.network_adapter.controllers`

Read-Only:

- `controller_capabilities` (Attributes) The capabilities of this controller (see [below for nested schema](#nestedatt--network_interfaces--network_adapter--controllers--controller_capabilities))
- `firmware_package_version` (String) The version of the user-facing firmware package

<a id="nestedatt--network_interfaces--network_adapter--controllers--controller_capabilities"></a>
### Nested Schema for `network_interfaces.network_adapter.controllers.controller_capabilities`

Read-Only:

- `data_center_bridging` (Attributes) Data center bridging (DCB) for capabilities of a controller (see [below for nested schema](#nestedatt--network_interfaces--network_adapter--controllers--controller_capabilities--data_center_bridging))
- `npar` (Attributes) NIC Partitioning capability, status, and configuration for a controller (see [below for nested schema](#nestedatt--network_interfaces--network_adapter--controllers--controller_capabilities--npar))
- `npiv` (Attributes) N_Port ID Virtualization (NPIV) capabilities for a controller (see [below for nested schema](#nestedatt--network_interfaces--network_adapter--controllers--controller_capabilities--npiv))
- `virtualization_offload` (Attributes) A Virtualization offload capability of a controller (see [below for nested schema](#nestedatt--network_interfaces--network_adapter--controllers--controller_capabilities--virtualization_offload))

<a id="nestedatt--network_interfaces--network_adapter--controllers--controller_capabilities--data_center_bridging"></a>
### Nested Schema for `network_interfaces.network_adapter.controllers.controller_capabilities.data_center_bridging`

Read-Only:

- `capable` (Boolean) An indication of whether this controller is capable of data center bridging (DCB)


<a id="nestedatt--network_interfaces--network_adapter--controllers--controller_capabilities--npar"></a>
### Nested Schema for `network_interfaces.network_adapter.controllers.controller_capabilities.npar`

Read-Only:

- `npar_capable` (Boolean) An indication of whether the controller supports NIC function partitioning
- `npar_enabled` (Boolean) An indication of whether NIC function partitioning is active on this controller.


<a id="nestedatt--network_interfaces--network_adapter--controllers--controller_capabilities--npiv"></a>
### Nested Schema for `network_interfaces.network_adapter.controllers.controller_capabilities.npiv`

Read-Only:

- `max_device_logins` (Number) The maximum number of N_Port ID Virtualization (NPIV) logins allowed simultaneously from all ports on this controller
- `max_port_logins` (Number) The maximum number of N_Port ID Virtualization (NPIV) logins allowed per physical port on this controller


<a id="nestedatt--network_interfaces--network_adapter--controllers--controller_capabilities--virtualization_offload"></a>
### Nested Schema for `network_interfaces.network_adapter.controllers.controller_capabilities.virtualization_offload`

Read-Only:

- `sriov` (Attributes) Single-root input/output virtualization (SR-IOV) capabilities (see [below for nested schema](#nestedatt--network_interfaces--network_adapter--controllers--controller_capabilities--virtualization_offload--sriov))
- `virtual_function` (Attributes) A virtual function of a controller (see [below for nested schema](#nestedatt--network_interfaces--network_adapter--controllers--controller_capabilities--virtualization_offload--virtual_function))

<a id="nestedatt--network_interfaces--network_adapter--controllers--controller_capabilities--virtualization_offload--sriov"></a>
### Nested Schema for `network_interfaces.network_adapter.controllers.controller_capabilities.virtualization_offload.sriov`

Read-Only:

- `sriov_vepa_capable` (Boolean) An indication of whether this controller supports single root input/output virtualization (SR-IOV)in Virtual Ethernet Port Aggregator (VEPA) mode


<a id="nestedatt--network_interfaces--network_adapter--controllers--controller_capabilities--virtualization_offload--virtual_function"></a>
### Nested Schema for `network_interfaces.network_adapter.controllers.controller_capabilities.virtualization_offload.virtual_function`

Read-Only:

- `device_max_count` (Number) The maximum number of virtual functions supported by this controller
- `min_assignment_group_size` (Number) The minimum number of virtual functions that can be allocated or moved between physical functions for this controller
- `network_port_max_count` (Number) The maximum number of virtual functions supported per network port for this controller





<a id="nestedatt--network_interfaces--network_adapter--status"></a>
### Nested Schema for `network_interfaces.network_adapter.status`

Read-Only:

- `health` (String) health
- `health_rollup` (String) health rollup
- `state` (String) state of the storage controller



<a id="nestedatt--network_interfaces--network_device_functions"></a>
### Nested Schema for `network_interfaces.network_device_functions`

Read-Only:

- `assignable_physical_network_ports` (List of String) A reference to assignable physical network ports to this function
- `assignable_physical_ports` (List of String) A reference to assignable physical ports to this function
- `description` (String) description of the network device function
- `ethernet` (Attributes) This type describes Ethernet capabilities, status, and configuration for a network device function (see [below for nested schema](#nestedatt--network_interfaces--network_device_functions--ethernet))
- `fibre_channel` (Attributes) This type describes Fibre Channel capabilities, status, and configuration for a network device function (see [below for nested schema](#nestedatt--network_interfaces--network_device_functions--fibre_channel))
- `id` (String) ID of the network device function
- `iscsi_boot` (Attributes) The iSCSI boot capabilities, status, and configuration for a network device function (see [below for nested schema](#nestedatt--network_interfaces--network_device_functions--iscsi_boot))
- `max_virtual_functions` (Number) The number of virtual functions that are available for this network device function
- `name` (String) name of the network device function
- `net_dev_func_capabilities` (List of String) An array of capabilities for this network device function
- `net_dev_func_type` (String) The configured capability of this network device function
- `odata_id` (String) OData ID for the network device function
- `oem` (Attributes) The OEM extension for this network network function (see [below for nested schema](#nestedatt--network_interfaces--network_device_functions--oem))
- `physical_port_assignment` (String) A reference to a physical port assignment to this function
- `status` (Attributes) status of the network device function (see [below for nested schema](#nestedatt--network_interfaces--network_device_functions--status))

<a id="nestedatt--network_interfaces--network_device_functions--ethernet"></a>
### Nested Schema for `network_interfaces.network_device_functions.ethernet`

Read-Only:

- `mac_address` (String) The currently configured MAC address
- `mtu_size` (Number) The maximum transmission unit (MTU) configured for this network device function
- `permanent_mac_address` (String) The permanent MAC address assigned to this function
- `vlan` (Attributes) The attributes of a VLAN (see [below for nested schema](#nestedatt--network_interfaces--network_device_functions--ethernet--vlan))

<a id="nestedatt--network_interfaces--network_device_functions--ethernet--vlan"></a>
### Nested Schema for `network_interfaces.network_device_functions.ethernet.vlan`

Read-Only:

- `vlan_enabled` (Boolean) An indication of whether the VLAN is enabled
- `vlan_id` (Number) The vlan id of the network device function



<a id="nestedatt--network_interfaces--network_device_functions--fibre_channel"></a>
### Nested Schema for `network_interfaces.network_device_functions.fibre_channel`

Read-Only:

- `allow_fip_vlan_discovery` (Boolean) An indication of whether the FCoE Initialization Protocol (FIP) populates the FCoE VLAN ID
- `boot_targets` (Attributes List) A Fibre Channel boot target configured for a network device function (see [below for nested schema](#nestedatt--network_interfaces--network_device_functions--fibre_channel--boot_targets))
- `fcoe_active_vlan_id` (Number) The active FCoE VLAN ID
- `fcoe_local_vlan_id` (Number) The locally configured FCoE VLAN ID
- `fibre_channel_id` (String) The Fibre Channel ID that the switch assigns for this interface
- `permanent_wwnn` (String) The permanent World Wide Node Name (WWNN) address assigned to this function
- `permanent_wwpn` (String) The permanent World Wide Port Name (WWPN) address assigned to this function
- `wwn_source` (String) The configuration source of the World Wide Names (WWN) for this World Wide Node Name (WWNN) and World Wide Port Name (WWPN) connection
- `wwnn` (String) The currently configured World Wide Node Name (WWNN) address of this function
- `wwpn` (String) The currently configured World Wide Port Name (WWPN) address of this function

<a id="nestedatt--network_interfaces--network_device_functions--fibre_channel--boot_targets"></a>
### Nested Schema for `network_interfaces.network_device_functions.fibre_channel.boot_targets`

Read-Only:

- `boot_priority` (Number) The relative priority for this entry in the boot targets array
- `lun_id` (String) The logical unit number (LUN) ID from which to boot on the device to which the corresponding WWPN refers
- `wwpn` (String) The World Wide Port Name (WWPN) from which to boot



<a id="nestedatt--network_interfaces--network_device_functions--iscsi_boot"></a>
### Nested Schema for `network_interfaces.network_device_functions.iscsi_boot`

Read-Only:

- `authentication_method` (String) The iSCSI boot authentication method for this network device function
- `chap_secret` (String, Sensitive) The shared secret for CHAP authentication
- `chap_username` (String) The user name for CHAP authentication
- `initiator_default_gateway` (String) The IPv6 or IPv4 iSCSI boot default gateway
- `initiator_ip_address` (String) The IPv6 or IPv4 address of the iSCSI initiator
- `initiator_name` (String) The iSCSI initiator name
- `initiator_netmask` (String) The IPv6 or IPv4 netmask of the iSCSI boot initiator
- `ip_address_type` (String) The type of IP address being populated in the iSCSIBoot IP address fields
- `ip_mask_dns_via_dhcp` (Boolean) An indication of whether the iSCSI boot initiator uses DHCP to obtain the initiator name, IP address, and netmask
- `mutual_chap_secret` (String, Sensitive) The CHAP secret for two-way CHAP authentication
- `mutual_chap_username` (String) The CHAP user name for two-way CHAP authentication
- `primary_dns` (String) The IPv6 or IPv4 address of the primary DNS server for the iSCSI boot initiator
- `primary_lun` (Number) The logical unit number (LUN) for the primary iSCSI boot target
- `primary_target_ip_address` (String) The IPv4 or IPv6 address for the primary iSCSI boot target
- `primary_target_name` (String) The name of the iSCSI primary boot target
- `primary_target_tcp_port` (Number) The TCP port for the primary iSCSI boot target
- `primary_vlan_enable` (Boolean) An indication of whether the primary VLAN is enabled
- `primary_vlan_id` (Number) The 802.1q VLAN ID to use for iSCSI boot from the primary target
- `router_advertisement_enabled` (Boolean) An indication of whether IPv6 router advertisement is enabled for the iSCSI boot target
- `secondary_dns` (String) The IPv6 or IPv4 address of the secondary DNS server for the iSCSI boot initiator
- `secondary_lun` (Number) The logical unit number (LUN) for the secondary iSCSI boot target
- `secondary_target_ip_address` (String) The IPv4 or IPv6 address for the secondary iSCSI boot target
- `secondary_target_name` (String) The name of the iSCSI secondary boot target
- `secondary_target_tcp_port` (Number) The TCP port for the secondary iSCSI boot target
- `secondary_vlan_enable` (Boolean) An indication of whether the secondary VLAN is enabled
- `secondary_vlan_id` (Number) The 802.1q VLAN ID to use for iSCSI boot from the secondary target
- `target_info_via_dhcp` (Boolean) An indication of whether the iSCSI boot target name, LUN, IP address, and netmask should be obtained from DHCP


<a id="nestedatt--network_interfaces--network_device_functions--oem"></a>
### Nested Schema for `network_interfaces.network_device_functions.oem`

Read-Only:

- `dell_fc` (Attributes) The OEM extension of Dell FC for this network device function (see [below for nested schema](#nestedatt--network_interfaces--network_device_functions--oem--dell_fc))
- `dell_fc_port_capabilities` (Attributes) The OEM extension of Dell FC capabilities for this network device function (see [below for nested schema](#nestedatt--network_interfaces--network_device_functions--oem--dell_fc_port_capabilities))
- `dell_fc_port_metrics` (Attributes) The OEM extension of Dell FC port metrics for this network device function (see [below for nested schema](#nestedatt--network_interfaces--network_device_functions--oem--dell_fc_port_metrics))
- `dell_nic` (Attributes) The OEM extension of Dell NIC for this network device function (see [below for nested schema](#nestedatt--network_interfaces--network_device_functions--oem--dell_nic))
- `dell_nic_capabilities` (Attributes) The OEM extension of Dell NIC capabilities for this network device function (see [below for nested schema](#nestedatt--network_interfaces--network_device_functions--oem--dell_nic_capabilities))
- `dell_nic_port_metrics` (Attributes) The OEM extension of Dell NIC port metrics for this network device function (see [below for nested schema](#nestedatt--network_interfaces--network_device_functions--oem--dell_nic_port_metrics))

<a id="nestedatt--network_interfaces--network_device_functions--oem--dell_fc"></a>
### Nested Schema for `network_interfaces.network_device_functions.oem.dell_fc`

Read-Only:

- `bus` (Number) This property represents the bus number of the PCI device
- `cable_length_metres` (Number) This property represents the cable length of Small Form Factor pluggable(SFP) Transceiver for the DellFC. Note: This property will be deprecated in Poweredge systems with model YX5X and iDRAC firmware version 4.20.20.20 or later
- `chip_type` (String) This property represents the chip type
- `device` (Number) This property represents the device number of the PCI device
- `device_description` (String) A string that contains the friendly Fully Qualified Device Description - a property that describes the device and its location
- `device_name` (String) This property represents FC HBA device name
- `efi_version` (String) This property represents the EFI version on the device
- `fabric_login_retry_count` (Number) This property represents the Fabric Login Retry Count
- `fabric_login_timeout` (Number) This property represents the Fabric Login Timeout in milliseconds
- `family_version` (String) This property represents the firmware version
- `fc_os_driver_version` (String) This property represents the FCOS OS Driver version for the DellFC
- `fc_tape_enable` (String) This property represents the FC Tape state
- `fcoe_os_driver_version` (String) This property represents the FCOE OS Driver version
- `frame_payload_size` (String) This property represents the frame payload size
- `function` (Number) This property represents the function number of the PCI device
- `hard_zone_address` (Number) This property represents the Hard Zone Address
- `hard_zone_enable` (String) This property represents the Hard Zone state
- `id` (String) ID of DellFC
- `identifier_type` (String) This property represents the type of Small Form Factor pluggable(SFP) Transceiver for the DellFC. Note: This property is deprecated and not supported in iDRAC firmware version 4.40.00.00 or later versions
- `iscsi_os_driver_version` (String) This property represents the ISCSI OS Driver version
- `lan_driver_version` (String) This property represents the LAN Driver version
- `link_down_timeout` (Number) This property represents the Link Down Timeout in milliseconds
- `loop_reset_delay` (Number) This property represents the Loop Reset Delay in seconds
- `name` (String) Name of DellFC
- `odata_id` (String) OData ID of DellFC for the network device function
- `part_number` (String) The part number assigned by the organization that is responsible for producing or manufacturing the FC device
- `port_down_retry_count` (Number) This property represents the Port Down Retry Count
- `port_down_timeout` (Number) This property represents the Port Down Timeout in milliseconds
- `port_login_retry_count` (Number) This property represents the Port Login Retry Count
- `port_login_timeout` (Number) This property represents the Port Login Timeout in milliseconds
- `product_name` (String) This property represents the Product Name
- `rdma_os_driver_version` (String) This property represents the RDMA OS Driver version
- `revision` (String) This property represents the revision number of the Small Form Factor pluggable(SFP) Transceiver.Note: This property is deprecated and not supported in iDRAC firmware version 4.40.00.00 or later versions
- `second_fc_target_lun` (Number) This property represents the Second FC Target LUN
- `second_fc_target_wwpn` (String) This property represents the Second FC Target World-Wide Port Name
- `serial_number` (String) A manufacturer-allocated number used to identify the FC device
- `transceiver_part_number` (String) The part number assigned by the organization that is responsible for producing or manufacturing the Small Form Factor pluggable(SFP) Transceivers. Note: This property is deprecated and not supported in iDRAC firmware version 4.40.00.00 or later versions
- `transceiver_serial_number` (String) A manufacturer-allocated number used to identify the Small Form Factor pluggable(SFP) TransceiverNote: This property is deprecated and not supported in iDRAC firmware version 4.40.00.00 or later versions
- `transceiver_vendor_name` (String) This property represents the vendor name of Small Form Factor pluggable(SFP) Transceiver for the DellFC. Note: This property is deprecated and not supported in iDRAC firmware version 4.40.00.00 or later versions
- `vendor_name` (String) This property represents the Vendor Name


<a id="nestedatt--network_interfaces--network_device_functions--oem--dell_fc_port_capabilities"></a>
### Nested Schema for `network_interfaces.network_device_functions.oem.dell_fc_port_capabilities`

Read-Only:

- `fc_max_number_exchanges` (Number) This property represents the maximum number of exchanges
- `fc_max_number_out_standing_commands` (Number) This property represents the maximum number of outstanding commands across all connections
- `feature_licensing_support` (String) The property provides details of the FC's feature licensing support
- `flex_addressing_support` (String) The property provides detail of the FC's port's flex addressing support
- `id` (String) ID of the DellFCCapabilities
- `name` (String) Name of the DellFCCapabilities
- `odata_id` (String) OData ID of DellFCCapabilities for the network device function
- `on_chip_thermal_sensor` (String) The property provides details of the FC's on-chip thermal sensor support
- `persistence_policy_support` (String) This property specifies if the card supports persistence policy
- `uefi_support` (String) The property provides details of the FC's port's UEFI support


<a id="nestedatt--network_interfaces--network_device_functions--oem--dell_fc_port_metrics"></a>
### Nested Schema for `network_interfaces.network_device_functions.oem.dell_fc_port_metrics`

Read-Only:

- `fc_invalid_crcs` (Number) This property represents invalid CRCs
- `fc_link_failures` (Number) This property represents link failures
- `fc_loss_of_signals` (Number) This property represents loss of signals
- `fc_rx_kb_count` (Number) This property represents the KB count received
- `fc_rx_sequences` (Number) This property represents the FC sequences received
- `fc_rx_total_frames` (Number) This property represents the total FC frames received
- `fc_tx_kb_count` (Number) This property represents the KB count transmitted
- `fc_tx_sequences` (Number) This property represents the FC sequences transmitted
- `fc_tx_total_frames` (Number) This property represents the total FC frames transmitted
- `id` (String) ID of the DellFCPortMetrics
- `name` (String) Name of the DellFCPortMetrics
- `odata_id` (String) OData ID of DellFCPortMetrics for the network device function
- `os_driver_state` (String) This property indicates the OS driver states for the DellFCPortMetrics
- `port_status` (String) This property represents port status for the DellFCPortMetrics
- `rx_input_power_mw` (Number) Indicates the RX input power value of Small Form Factor pluggable(SFP) Transceiver for the DellFCPortMetrics. Note: This property is deprecated and not supported in iDRAC firmware version 4.40.00.00 or later versions
- `rx_input_power_status` (String) Indicates the status of Rx Input Power value limits for the DellFCPortMetrics. Note: This property is deprecated and not supported in iDRAC firmware version 4.40.00.00 or later versions
- `temperature_celsius` (Number) Indicates the temperature value of Small Form Factor pluggable(SFP) Transceiver for the DellFCPortMetrics. Note: This property is deprecated and not supported in iDRAC firmware version 4.40.00.00 or later versions
- `temperature_status` (String) Indicates the status of Temperature value limits for the DellFCPortMetrics. Note: This property is deprecated and not supported in iDRAC firmware version 4.40.00.00 or later versions
- `tx_bias_current_mw` (Number) Indicates the TX Bias current value of Small Form Factor pluggable(SFP) Transceiver for the DellFCPortMetrics. Note: This property is deprecated and not supported in iDRAC firmware version 4.40.00.00 or later versions
- `tx_bias_current_status` (String) Indicates the status of Tx Bias Current value limits for the DellFCPortMetrics. Note: This property is deprecated and not supported in iDRAC firmware version 4.40.00.00 or later versions
- `tx_output_power_mw` (Number) Indicates the TX output power value of Small Form Factor pluggable(SFP) Transceiver for the DellFCPortMetrics. Note: This property is deprecated and not supported in iDRAC firmware version 4.40.00.00 or later versions
- `tx_output_power_status` (String) Indicates the status of Tx Output Power value limits for the DellFCPortMetrics. Note: This property is deprecated and not supported in iDRAC firmware version 4.40.00.00 or later versions
- `voltage_status` (String) Indicates the status of voltage value limits for the DellFCPortMetrics. Note: This property is deprecated and not supported in iDRAC firmware version 4.40.00.00 or later versions
- `voltage_value_volts` (Number) Indicates the voltage value of Small Form Factor pluggable(SFP) Transceiver for the DellFCPortMetrics. Note: This property is deprecated and not supported in iDRAC firmware version 4.40.00.00 or later versions


<a id="nestedatt--network_interfaces--network_device_functions--oem--dell_nic"></a>
### Nested Schema for `network_interfaces.network_device_functions.oem.dell_nic`

Read-Only:

- `bus_number` (Number) The bus number where this PCI device resides
- `cable_length_metres` (Number) This property represents the cable length of Small Form Factor pluggable(SFP) Transceiver for the DellNIC. Note: This property will be deprecated in Poweredge systems with model YX5X and iDRAC firmware version 4.20.20.20 or later
- `controller_bios_version` (String) This property represents the firmware version of Controller BIOS
- `data_bus_width` (String) This property represents the data-bus width of the NIC PCI device
- `device_description` (String) A string that contains the friendly Fully Qualified Device Description (FQDD), which is a property that describes the device and its location
- `efi_version` (String) This property represents the firmware version of EFI
- `family_version` (String) Represents family version of firmware
- `fc_os_driver_version` (String) This property represents the FCOS OS Driver version for the DellNIC
- `fcoe_offload_mode` (String) This property indicates if Fibre Channel over Ethernet (FCoE) personality is enabled or disabled on current partition in a Converged Network Adaptor device
- `fqdd` (String) A string that contains the Fully Qualified Device Description (FQDD) for the DellNIC
- `id` (String) ID of DellNIC
- `identifier_type` (String) This property represents the type of Small Form Factor pluggable(SFP) Transceiver for the DellNIC. Note: This property is deprecated and not supported in iDRAC firmware version 4.40.00.00 or later versions
- `instance_id` (String) A unique identifier for the instance
- `iscsi_offload_mode` (String) This property indicates if Internet Small Computer System Interface (iSCSI) personality is enabled or disabled on current partition in a Converged Network Adaptor device
- `last_system_inventory_time` (String) This property represents the time when System Inventory Collection On Reboot (CSIOR) was last performed or the object was last updated on iDRAC. The value is represented in the format yyyymmddHHMMSS
- `last_update_time` (String) This property represents the time when the data was last updated. The value is represented in the format yyyymmddHHMMSS
- `link_duplex` (String) This property indicates whether the Link is full-duplex or half-duplex
- `media_type` (String) The property shall represent the drive media type
- `name` (String) name of DellNIC
- `nic_mode` (String) Represents if network interface card personality is enabled or disabled on current partition in a Converged Network Adaptor device
- `odata_id` (String) OData ID of DellNIC for the network device function
- `part_number` (String) The part number assigned by the organization that is responsible for producing or manufacturing the NIC device
- `pci_device_id` (String) This property contains a value assigned by the device manufacturer used to identify the type of device
- `pci_sub_device_id` (String) Represents PCI sub device ID
- `pci_sub_vendor_id` (String) This property represents the subsystem vendor ID. ID information is reported from a PCIDevice through protocol-specific requests
- `pci_vendor_id` (String) This property represents the register that contains a value assigned by the PCI SIG used to identify the manufacturer of the device
- `permanent_fcoe_emac_address` (String) PermanentFCOEMACAddress defines the network address that is hardcoded into a port for FCoE
- `permanent_iscsi_emac_address` (String) PermanentAddress defines the network address that is hardcoded into a port for iSCSI. This 'hardcoded' address can be changed using a firmware upgrade or a software configuration. When this change is made, the field should be updated at the same time. PermanentAddress should be left blank if no 'hardcoded' address exists for the NetworkAdapter.
- `product_name` (String) A string containing the product name
- `protocol` (String) Supported Protocol Types
- `revision` (String) This property represents the revision number of the Small Form Factor pluggable(SFP) Transceiver. Note: This property is deprecated and not supported in iDRAC firmware version 4.40.00.00 or later versions
- `serial_number` (String) A manufacturer-allocated number used to identify the NIC device
- `slot_length` (String) This property represents the represents the slot length of the NIC PCI device
- `slot_type` (String) This property indicates the slot type of the NIC PCI device
- `snapi_state` (String) This property represents the SNAPI state
- `snapi_support` (String) This property represents the SNAPI support
- `transceiver_part_number` (String) The part number assigned by the organization that is responsible for producing or SFP Transceivers(manufacturing the Small Form Factor pluggable). Note: This property is deprecated and not supported in iDRAC firmware version 4.40.00.00 or later versions
- `transceiver_serial_number` (String) A manufacturer-allocated number used to identify the Small Form Factor pluggable(SFP) TransceiverNote: This property is deprecated and not supported in iDRAC firmware version 4.40.00.00 or later versions
- `transceiver_vendor_name` (String) This property represents the vendor name of Small Form Factor pluggable(SFP) Transceiver for the DellNIC.Note: This property is deprecated and not supported in iDRAC firmware version 4.40.00.00 or later versions
- `vendor_name` (String) This property represents the vendor name
- `vpi_support` (String) This property represents the VPI support


<a id="nestedatt--network_interfaces--network_device_functions--oem--dell_nic_capabilities"></a>
### Nested Schema for `network_interfaces.network_device_functions.oem.dell_nic_capabilities`

Read-Only:

- `bpe_support` (String) This property represents Bridge Port Extension (BPE) for the ports of the NIC
- `congestion_notification` (String) This property represents congestion notification support for a NIC port
- `dcb_exchange_protocol` (String) This property represents DCB Exchange protocol support for a NIC port
- `ets` (String) This property represents Enhanced Transmission Selection support for a NIC port
- `evb_modes_support` (String) This property represents EVB Edge Virtual Bridging) mode support for the ports of the NIC. Possible values are 0 Unknown, 2 Supported, 3 Not Supported
- `fcoe_boot_support` (String) The property shall represent FCoE boot support for a NIC port
- `fcoe_max_ios_per_session` (Number) This property represents the maximum number of I/Os per connection supported by the NIC
- `fcoe_max_npiv_per_port` (Number) This property represents the maximum number of NPIV per port supported by the DellNICCapabilities
- `fcoe_max_number_exchanges` (Number) This property represents the maximum number of exchanges for the NIC
- `fcoe_max_number_logins` (Number) This property represents the maximum logins per port for the NIC
- `fcoe_max_number_of_fc_targets` (Number) This property represents the maximum number of FCoE targets supported by the NIC
- `fcoe_max_number_outstanding_commands` (Number) This property represents the maximum number of outstanding commands supported across all connections for the NIC
- `fcoe_offload_support` (String) The property shall represent FCoE offload support for the NIC
- `feature_licensing_support` (String) This property represents feature licensing support for the NIC
- `flex_addressing_support` (String) The property shall represent flex adddressing support for a NIC port
- `id` (String) ID of DellNICCapabilities
- `ipsec_offload_support` (String) This property represents IPSec offload support for a NIC port
- `iscsi_boot_support` (String) The property shall represent iSCSI boot support for a NIC port
- `iscsi_offload_support` (String) The property shall represent iSCSI offload support for a NIC port
- `mac_sec_support` (String) This property represents secure MAC support for a NIC port
- `name` (String) Name of DellNICCapabilities
- `nic_partitioning_support` (String) This property represents partitioning support for the NIC
- `nw_management_pass_through` (String) This property represents network management passthrough support for a NIC port
- `odata_id` (String) OData ID of DellNICCapabilities for the network device function
- `on_chip_thermal_sensor` (String) This property represents on-chip thermal sensor support for the NIC
- `open_flow_support` (String) This property represents open-flow support for a NIC port
- `os_bmc_management_pass_through` (String) This property represents OS-inband to BMC-out-of-band management passthrough support for a NIC port
- `partition_wol_support` (String) This property represents Wake-On-LAN support for a NIC partition
- `persistence_policy_support` (String) This property specifies whether the card supports persistence policy
- `priority_flow_control` (String) This property represents priority flow-control support for a NIC port
- `pxe_boot_support` (String) The property shall represent PXE boot support for a NIC port
- `rdma_support` (String) This property represents RDMA support for a NIC port
- `remote_phy` (String) This property represents remote PHY support for a NIC port
- `tcp_chimney_support` (String) This property represents TCP Chimney support for a NIC port
- `tcp_offload_engine_support` (String) This property represents the support of TCP Offload Engine for a NIC port
- `uefi_support` (String) This property represents UEFI support for a NIC port
- `veb` (String) This property provides details about the VEB (Virtual Ethernet Bridging) - single channel support for the ports of the NIC
- `veb_vepa_multi_channel` (String) This property provides details about the Virtual Ethernet Bridging and Virtual Ethernet Port Aggregator (VEB-VEPA) multichannel support for the ports of the NIC
- `veb_vepa_single_channel` (String) This property provides details about the VEB-VEPA (Virtual Ethernet Bridging and Virtual Ethernet Port Aggregator) - single channel support for the ports of the NIC
- `virtual_link_control` (String) This property represents virtual link-control support for a NIC partition


<a id="nestedatt--network_interfaces--network_device_functions--oem--dell_nic_port_metrics"></a>
### Nested Schema for `network_interfaces.network_device_functions.oem.dell_nic_port_metrics`

Read-Only:

- `discarded_pkts` (Number) Indicates the total number of discarded packets
- `fc_crc_error_count` (Number) Indicates the number of FC frames with CRC errors
- `fcoe_link_failures` (Number) Indicates the number of FCoE/FIP login failures
- `fcoe_pkt_rx_count` (Number) Indicates the number of good (FCS valid) packets received with the active FCoE MAC address of the partition
- `fcoe_pkt_tx_count` (Number) Indicates the number of good (FCS valid) packets transmitted that passed L2 filtering by a specific MAC address
- `fcoe_rx_pkt_dropped_count` (Number) Indicates the number of receive packets with FCS errors
- `fqdd` (String) A string that contains the Fully Qualified Device Description (FQDD) for the DellNICPortMetrics
- `id` (String) ID of DellNICPortMetrics
- `lan_fcs_rx_errors` (Number) Indicates the Lan FCS receive Errors
- `lan_unicast_pkt_rx_count` (Number) Indicates the total number of Lan Unicast Packets Received
- `lan_unicast_pkt_tx_count` (Number) Indicates the total number of Lan Unicast Packets Transmitted
- `name` (String) Name of DellNICPortMetrics
- `odata_id` (String) OData ID of DellNICPortMetrics for the network device function
- `os_driver_state` (String) Indicates operating system driver states
- `partition_link_status` (String) Indicates whether the partition link is up or down for the DellFCPortMetrics
- `partition_os_driver_state` (String) Indicates operating system driver states of the partitions for the DellFCPortMetrics
- `rdma_rx_total_bytes` (Number) Indicates the total number of RDMA bytes received
- `rdma_rx_total_packets` (Number) Indicates the total number of RDMA packets received
- `rdma_total_protection_errors` (Number) Indicates the total number of RDMA Protection errors
- `rdma_total_protocol_errors` (Number) Indicates the total number of RDMA Protocol errors
- `rdma_tx_total_bytes` (Number) Indicates the total number of RDMA bytes transmitted
- `rdma_tx_total_packets` (Number) Indicates the total number of RDMA packets transmitted
- `rdma_tx_total_read_req_pkts` (Number) Indicates the total number of RDMA ReadRequest packets transmitted
- `rdma_tx_total_send_pkts` (Number) Indicates the total number of RDMA Send packets transmitted
- `rdma_tx_total_write_pkts` (Number) Indicates the total number of RDMA Write packets transmitted
- `rx_broadcast` (Number) Indicates the total number of good broadcast packets received
- `rx_bytes` (Number) Indicates the total number of bytes received, including host and remote management pass through traffic. Remote management passthrough received traffic is applicable to LOMs only
- `rx_error_pkt_alignment_errors` (Number) Indicates the total number of packets received with alignment errors
- `rx_error_pkt_fcs_errors` (Number) Indicates the total number of packets received with FCS errors
- `rx_false_carrier_detection` (Number) Indicates the total number of false carrier errors received from PHY
- `rx_input_power_mw` (Number) Indicates the RX input power value of Small Form Factor pluggable(SFP) Transceiver for the DellNICPortMetrics. Note: This property is deprecated and not supported in iDRAC firmware version 4.40.00.00 or later versions
- `rx_input_power_status` (String) Indicates the status of Rx Input Power value limits for the DellFCPortMetrics
- `rx_jabber_pkt` (Number) Indicates the total number of frames that are too long
- `rx_mutlicast_packets` (Number) Indicates the total number of good multicast packets received
- `rx_pause_xoff_frames` (Number) Indicates the flow control frames from the network to pause transmission
- `rx_pause_xon_frames` (Number) Indicates the flow control frames from the network to resume transmission
- `rx_runt_pkt` (Number) Indicates the total number of frames that are too short (< 64 bytes)
- `rx_unicast_packets` (Number) Indicates the total number of good unicast packets received
- `start_statistic_time` (String) Indicates the measurement time for the first NIC statistics. The property is used with the StatisticTime property to calculate the duration over which the NIC statistics are gathered
- `statistic_time` (String) Indicates the most recent measurement time for NIC statistics. The property is used with the StatisticStartTime property to calculate the duration over which the NIC statistics are gathered
- `temperature_celsius` (Number) Indicates the temperature value of Small Form Factor pluggable(SFP) Transceiver for the DellNICPortMetrics. Note: This property is deprecated and not supported in iDRAC firmware version 4.40.00.00 or later versions
- `temperature_status` (String) Indicates the status of Temperature value limits for the DellNICPortMetrics.
- `tx_bias_current_ma` (Number) Indicates the TX Bias current value of Small Form Factor pluggable(SFP) Transceiver for the DellNICPortMetrics. Note: This property is deprecated and not supported in iDRAC firmware version 4.40.00.00 or later versions
- `tx_bias_current_status` (String) Indicates the status of Tx Bias Current value limits for the DellNICPortMetrics. Note: This property is deprecated and not supported in iDRAC firmware version 4.40.00.00 or later versions
- `tx_broadcast` (Number) Indicates the total number of good broadcast packets transmitted
- `tx_bytes` (Number) Indicates the total number of bytes transmitted, including host and remote management passthrough traffic. Remote management passthrough transmitted traffic is applicable to LOMs only
- `tx_error_pkt_excessive_collision` (Number) Indicates the number of times a single transmitted packet encountered more than 15 collisions
- `tx_error_pkt_late_collision` (Number) Indicates the number of collisions that occurred after one slot time (defined by IEEE 802.3)
- `tx_error_pkt_multiple_collision` (Number) Indicates the number of times that a transmitted packet encountered 2-15 collisions
- `tx_error_pkt_single_collision` (Number) Indicates the number of times that a successfully transmitted packet encountered a single collision
- `tx_mutlicast_packets` (Number) Indicates the total number of good multicast packets transmitted
- `tx_output_power_mw` (Number) Indicates the TX output power value of Small Form Factor pluggable(SFP) Transceiver for the DellNICPortMetrics. Note: This property is deprecated and not supported in iDRAC firmware version 4.40.00.00 or later versions
- `tx_output_power_status` (String) Indicates the status of Tx Output Power value limits for the DellNICPortMetrics.. Note: This property is deprecated and not supported in iDRAC firmware version 4.40.00.00 or later versions
- `tx_pause_xoff_frames` (Number) Indicates the number of XOFF packets transmitted to the network
- `tx_pause_xon_frames` (Number) Indicates the number of XON packets transmitted to the network
- `tx_unicast_packets` (Number) Indicates the total number of good unicast packets transmitted for the DellFCPortMetrics
- `voltage_status` (String) Indicates the status of voltage value limits for the DellNICPortMetrics. Note: This property is deprecated and not supported in iDRAC firmware version 4.40.00.00 or later versions
- `voltage_value_volts` (Number) Indicates the voltage value of Small Form Factor pluggable(SFP) Transceiver for the DellNICPortMetrics. Note: This property is deprecated and not supported in iDRAC firmware version 4.40.00.00 or later versions



<a id="nestedatt--network_interfaces--network_device_functions--status"></a>
### Nested Schema for `network_interfaces.network_device_functions.status`

Read-Only:

- `health` (String) health
- `health_rollup` (String) health rollup
- `state` (String) state of the storage controller



<a id="nestedatt--network_interfaces--network_ports"></a>
### Nested Schema for `network_interfaces.network_ports`

Read-Only:

- `active_link_technology` (String) Network port active link technology
- `associated_network_addresses` (List of String) An array of configured MAC or WWN network addresses that are associated with this network port, including the programmed address of the lowest numbered network device function, the configured but not active address, if applicable, the address for hardware port teaming, or other network addresses
- `current_link_speed_mbps` (Number) Network port current link speed
- `description` (String) description of the network port
- `eee_enabled` (Boolean) An indication of whether IEEE 802.3az Energy-Efficient Ethernet (EEE) is enabled for this network port
- `flow_control_configuration` (String) The locally configured 802.3x flow control setting for this network port
- `flow_control_status` (String) The 802.3x flow control behavior negotiated with the link partner for this network port (Ethernet-only)
- `id` (String) ID of the network port
- `link_status` (String) The status of the link between this port and its link partner
- `name` (String) name of the network port
- `net_dev_func_max_bw_alloc` (Attributes List) A maximum bandwidth allocation percentage for a network device functions associated a port (see [below for nested schema](#nestedatt--network_interfaces--network_ports--net_dev_func_max_bw_alloc))
- `net_dev_func_min_bw_alloc` (Attributes List) A minimum bandwidth allocation percentage for a network device functions associated a port (see [below for nested schema](#nestedatt--network_interfaces--network_ports--net_dev_func_min_bw_alloc))
- `odata_id` (String) OData ID for the network port
- `oem` (Attributes) The OEM extension for this network port (see [below for nested schema](#nestedatt--network_interfaces--network_ports--oem))
- `physical_port_number` (String) The physical port number label for this port
- `status` (Attributes) status of the network port (see [below for nested schema](#nestedatt--network_interfaces--network_ports--status))
- `supported_ethernet_capabilities` (List of String) The set of Ethernet capabilities that this port supports.
- `supported_link_capabilities` (Attributes List) The link capabilities of an associated port (see [below for nested schema](#nestedatt--network_interfaces--network_ports--supported_link_capabilities))
- `vendor_id` (String) The vendor Identification for this port
- `wake_on_lan_enabled` (Boolean) An indication of whether Wake on LAN (WoL) is enabled for this network port

<a id="nestedatt--network_interfaces--network_ports--net_dev_func_max_bw_alloc"></a>
### Nested Schema for `network_interfaces.network_ports.net_dev_func_max_bw_alloc`

Read-Only:

- `max_bw_alloc_percent` (Number) The maximum bandwidth allocation percentage allocated to the corresponding network device function instance
- `network_device_function` (String) List of network device functions for NetDevFuncMaxBWAlloc associated with this port


<a id="nestedatt--network_interfaces--network_ports--net_dev_func_min_bw_alloc"></a>
### Nested Schema for `network_interfaces.network_ports.net_dev_func_min_bw_alloc`

Read-Only:

- `min_bw_alloc_percent` (Number) The minimum bandwidth allocation percentage allocated to the corresponding network device function instance
- `network_device_function` (String) List of network device functions for NetDevFuncMinBWAlloc associated with this port


<a id="nestedatt--network_interfaces--network_ports--oem"></a>
### Nested Schema for `network_interfaces.network_ports.oem`

Read-Only:

- `dell_network_transceiver` (Attributes) Dell Network Transceiver (see [below for nested schema](#nestedatt--network_interfaces--network_ports--oem--dell_network_transceiver))

<a id="nestedatt--network_interfaces--network_ports--oem--dell_network_transceiver"></a>
### Nested Schema for `network_interfaces.network_ports.oem.dell_network_transceiver`

Read-Only:

- `device_description` (String) A string that contains the friendly Fully Qualified Device Description (FQDD), which is a property that describes the device and its location
- `fqdd` (String) A string that contains the Fully Qualified Device Description (FQDD) for the DellNetworkTransceiver
- `id` (String) The unique identifier for this resource within the collection of similar resources
- `identifier_type` (String) This property represents the type of Small Form Factor pluggable(SFP) Transceiver for the DellNetworkTransceiver
- `interface_type` (String) This property represents the interface type of Small Form Factor pluggable(SFP) Transceiver
- `name` (String) The name of the resource or array member
- `odata_id` (String) The unique identifier for a resource
- `part_number` (String) The part number assigned by the organization that is responsible for producing or SFP(manufacturing the Small Form Factor pluggable) Transceivers
- `revision` (String) This property represents the revision number of the Small Form Factor pluggable(SFP) Transceiver
- `serial_number` (String) A manufacturer-allocated number used to identify the Small Form Factor pluggable(SFP) Transceiver
- `vendor_name` (String) This property represents the vendor name of Small Form Factor pluggable(SFP) Transceiver for the object.



<a id="nestedatt--network_interfaces--network_ports--status"></a>
### Nested Schema for `network_interfaces.network_ports.status`

Read-Only:

- `health` (String) health
- `health_rollup` (String) health rollup
- `state` (String) state of the storage controller


<a id="nestedatt--network_interfaces--network_ports--supported_link_capabilities"></a>
### Nested Schema for `network_interfaces.network_ports.supported_link_capabilities`

Read-Only:

- `auto_speed_negotiation` (Boolean) An indication of whether the port is capable of autonegotiating speed
- `link_network_technology` (String) The link network technology capabilities of this port
- `link_speed_mbps` (Number) The speed of the link in Mbit/s when this link network technology is active



<a id="nestedatt--network_interfaces--status"></a>
### Nested Schema for `network_interfaces.status`

Read-Only:

- `health` (String) health
- `health_rollup` (String) health rollup
- `state` (String) state of the storage controller

