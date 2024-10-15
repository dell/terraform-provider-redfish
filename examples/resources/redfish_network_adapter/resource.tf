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

