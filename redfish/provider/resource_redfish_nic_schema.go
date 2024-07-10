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

package provider

import (
	"github.com/hashicorp/terraform-plugin-framework-validators/boolvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/mapvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/objectvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	redfishcommon "github.com/stmcginnis/gofish/common"
	"github.com/stmcginnis/gofish/redfish"
)

const (
	defaultNICJobTimeout                int64 = 1200
	intervalNICJobCheckTime             int64 = 10
	defaultNICResetTimeout              int64 = 120
	fieldDescriptionNetDevFuncID              = "ID of the network device function"
	errMessageInvalidInput                    = "input params are not valid"
	noteMessageUpdateOneAttrsOnly             = "Please update one of network_attributes or oem_network_attributes at a time."
	noteMessageUpdateAttrsExclusive           = "Note: `oem_network_attributes` is mutually exclusive with `network_attributes`. "
	patchBodySettingsApplyTime                = "@Redfish.SettingsApplyTime"
	patchBodyApplyTime                        = "ApplyTime"
	fieldNameClearPending                     = "clear_pending"
	fieldNameAttributes                       = "attributes"
	fieldNameWWPN                             = "wwpn"
	fieldNameWWNN                             = "wwnn"
	fieldNameWWNNSource                       = "wwn_source"
	fieldNameBootPriority                     = "boot_priority"
	fieldNameLunID                            = "lun_id"
	fieldNameFibreChannel                     = "fibre_channel"
	fieldNameNetDevFuncType                   = "net_dev_func_type"
	fieldNameEthernet                         = "ethernet"
	fieldNameIscsiBoot                        = "iscsi_boot"
	fieldNameHealth                           = "health"
	fieldNameMACAddress                       = "mac_address"
	fieldNameMTUSize                          = "mtu_size"
	fieldNameVLAN                             = "vlan"
	fieldNameVLANID                           = "vlan_id"
	fieldNameVLANEnabled                      = "vlan_enabled"
	fieldNameAllowFipVlanDiscovery            = "allow_fip_vlan_discovery" // nolint: gosec
	fieldNameBootTargets                      = "boot_targets"
	fieldNameFcoeLocalVlanID                  = "fcoe_local_vlan_id"
	fieldNameAuthenticationMethod             = "authentication_method"
	fieldNameChapSec                          = "chap_secret"
	fieldNameChapUsername                     = "chap_username"
	fieldNameIPAddressType                    = "ip_address_type"
	fieldNameIPMaskDNSViaDHCP                 = "ip_mask_dns_via_dhcp"
	fieldNameInitiatorDefaultGateway          = "initiator_default_gateway"
	fieldNameInitiatorIPAddress               = "initiator_ip_address"
	fieldNameInitiatorName                    = "initiator_name"
	fieldNameInitiatorNetmask                 = "initiator_netmask"
	fieldNameMutualChapSec                    = "mutual_chap_secret"
	fieldNameMutualChapUsername               = "mutual_chap_username"
	fieldNamePrimaryDNS                       = "primary_dns"
	fieldNamePrimaryLun                       = "primary_lun"
	fieldNamePrimaryTargetIPAddress           = "primary_target_ip_address"
	fieldNamePrimaryTargetName                = "primary_target_name"
	fieldNamePrimaryTargetTCPPort             = "primary_target_tcp_port"
	fieldNamePrimaryVLANEnable                = "primary_vlan_enable"
	fieldNamePrimaryVLANID                    = "primary_vlan_id"
	fieldNameRouterAdvertisementEnabled       = "router_advertisement_enabled"
	fieldNameSecondaryDNS                     = "secondary_dns"
	fieldNameSecondaryLun                     = "secondary_lun"
	fieldNameSecondaryTargetIPAddress         = "secondary_target_ip_address"
	fieldNameSecondaryTargetName              = "secondary_target_name"
	fieldNameSecondaryTargetTCPPort           = "secondary_target_tcp_port"
	fieldNameSecondaryVLANEnable              = "secondary_vlan_enable"
	fieldNameSecondaryVLANID                  = "secondary_vlan_id"
	fieldNameTargetInfoViaDHCP                = "target_info_via_dhcp"
)

// NICResourceSchema defines the schema for the resource.
func NICResourceSchema() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		NICComponmentSchemaID: schema.StringAttribute{
			MarkdownDescription: "ID of the network interface cards resource",
			Description:         "ID of the network interface cards resource",
			Computed:            true,
		},
		"network_adapter_id": schema.StringAttribute{
			MarkdownDescription: "ID of the network adapter",
			Description:         "ID of the network adapter",
			Required:            true,
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"network_device_function_id": schema.StringAttribute{
			MarkdownDescription: fieldDescriptionNetDevFuncID,
			Description:         fieldDescriptionNetDevFuncID,
			Required:            true,
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"apply_time": schema.StringAttribute{
			MarkdownDescription: "Apply time of the `network_attributes` and `oem_network_attributes`. (Update Supported)" +
				"Accepted values: `Immediate`, `OnReset`, `AtMaintenanceWindowStart`, `InMaintenanceWindowOnReset`. " +
				"Immediate: allows the user to immediately reboot the host and apply the changes. " +
				"This is only applicable for `oem_network_attributes`." +
				"OnReset: allows the user to apply the changes on the next reboot of the host server." +
				"AtMaintenanceWindowStart: allows the user to apply at the start of a maintenance window as specified in `maintenance_window`." +
				"InMaintenanceWindowOnReset: allows to apply after a manual reset but within the maintenance window as specified in " +
				"`maintenance_window`.",
			Description: "Apply time of the `network_attributes` and `oem_network_attributes`. (Update Supported)" +
				"Accepted values: `Immediate`, `OnReset`, `AtMaintenanceWindowStart`, `InMaintenanceWindowOnReset`. " +
				"Immediate: allows the user to immediately reboot the host and apply the changes. " +
				"This is only applicable for `oem_network_attributes`." +
				"OnReset: allows the user to apply the changes on the next reboot of the host server." +
				"AtMaintenanceWindowStart: allows the user to apply at the start of a maintenance window as specified in `maintenance_window`." +
				"InMaintenanceWindowOnReset: allows to apply after a manual reset but within the maintenance window as specified in " +
				"`maintenance_window`.",
			Required: true,
			Validators: []validator.String{
				stringvalidator.OneOf(
					string(redfishcommon.ImmediateApplyTime),
					string(redfishcommon.OnResetApplyTime),
					string(redfishcommon.AtMaintenanceWindowStartApplyTime),
					string(redfishcommon.InMaintenanceWindowOnResetApplyTime),
				),
			},
		},
		"reset_timeout": schema.Int64Attribute{
			MarkdownDescription: "Reset Timeout. Default value is 120 seconds. (Update Supported)",
			Description:         "Reset Timeout. Default value is 120 seconds. (Update Supported)",
			Optional:            true,
			Computed:            true,
			Default:             int64default.StaticInt64(defaultNICResetTimeout),
		},
		"reset_type": schema.StringAttribute{
			MarkdownDescription: "Reset Type. (Update Supported) " +
				"Accepted values: `ForceRestart`, `GracefulRestart`, `PowerCycle`. Default value is `ForceRestart`.",
			Description: "Reset Type. (Update Supported) " +
				"Accepted values: `ForceRestart`, `GracefulRestart`, `PowerCycle`. Default value is `ForceRestart`.",
			Optional: true,
			Computed: true,
			Default:  stringdefault.StaticString(string(redfish.ForceRestartResetType)),
			Validators: []validator.String{
				stringvalidator.OneOf([]string{
					string(redfish.ForceRestartResetType),
					string(redfish.GracefulRestartResetType),
					string(redfish.PowerCycleResetType),
				}...),
			},
		},
		"system_id": schema.StringAttribute{
			MarkdownDescription: "ID of the system resource. If the value for system ID is not provided, " +
				"the resource picks the first system available from the iDRAC.",
			Description: "ID of the system resource. If the value for system ID is not provided, " +
				"the resource picks the first system available from the iDRAC.",
			Computed:   true,
			Optional:   true,
			Validators: []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"job_timeout": schema.Int64Attribute{
			MarkdownDescription: "`job_timeout` is the time in seconds that the provider waits for the resource update job to be" +
				"completed before timing out. (Update Supported) Default value is 1200 seconds." +
				"`job_timeout` is applicable only when `apply_time` is `Immediate` or `OnReset`.",
			Description: "`job_timeout` is the time in seconds that the provider waits for the resource update job to be" +
				"completed before timing out. (Update Supported) Default value is 1200 seconds." +
				"`job_timeout` is applicable only when `apply_time` is `Immediate` or `OnReset`.",
			Optional: true,
			Computed: true,
			Default:  int64default.StaticInt64(int64(defaultNICJobTimeout)),
		},
		"maintenance_window": schema.SingleNestedAttribute{
			Description: "This option allows you to schedule the maintenance window. (Update Supported)" +
				"This is required when `apply_time` is `AtMaintenanceWindowStart` or `InMaintenanceWindowOnReset` .",
			MarkdownDescription: "This option allows you to schedule the maintenance window. (Update Supported)" +
				"This is required when `apply_time` is `AtMaintenanceWindowStart` or `InMaintenanceWindowOnReset` .",
			Optional: true,
			Attributes: map[string]schema.Attribute{
				"start_time": schema.StringAttribute{
					Description: "The start time for the maintenance window to be scheduled. (Update Supported)" +
						"The format is YYYY-MM-DDThh:mm:ss<offset>. " +
						"<offset> is the time offset from UTC that the current timezone set in iDRAC in the format: +05:30 for IST.",
					MarkdownDescription: "The start time for the maintenance window to be scheduled. (Update Supported)" +
						"The format is YYYY-MM-DDThh:mm:ss<offset>. " +
						"<offset> is the time offset from UTC that the current timezone set in iDRAC in the format: +05:30 for IST.",
					Required:   true,
					Validators: []validator.String{stringvalidator.LengthAtLeast(1)},
				},
				"duration": schema.Int64Attribute{
					Description:         "The duration in seconds for the maintenance window. (Update Supported)",
					MarkdownDescription: "The duration in seconds for the maintenance window. (Update Supported)",
					Required:            true,
				},
			},
		},
		"oem_network_attributes": schema.SingleNestedAttribute{
			Description: "oem_network_attributes to configure dell network attributes and clear pending action. (Update Supported) " +
				noteMessageUpdateAttrsExclusive +
				noteMessageUpdateOneAttrsOnly,
			MarkdownDescription: "oem_network_attributes to configure dell network attributes and clear pending action. (Update Supported) " +
				noteMessageUpdateAttrsExclusive +
				noteMessageUpdateOneAttrsOnly,
			Optional:   true,
			Computed:   true,
			Validators: []validator.Object{objectvalidator.AtLeastOneOf(path.MatchRoot("network_attributes"))},
			Attributes: map[string]schema.Attribute{
				NICComponmentSchemaOdataID: schema.StringAttribute{
					Computed:            true,
					MarkdownDescription: "OData ID for the network_attributes",
					Description:         "OData ID for the network_attributes",
				},
				NICComponmentSchemaID: schema.StringAttribute{
					Computed:            true,
					MarkdownDescription: "ID of the network_attributes",
					Description:         "ID of the network_attributes",
				},
				NICComponmentSchemaName: schema.StringAttribute{
					Computed:            true,
					MarkdownDescription: "name of the network_attributes",
					Description:         "name of the network_attributes",
				},
				NICComponmentSchemaDescription: schema.StringAttribute{
					Computed:            true,
					MarkdownDescription: "description of network_attributes",
					Description:         "description of the network_attributes",
				},
				"attribute_registry": schema.StringAttribute{
					Computed:            true,
					MarkdownDescription: "registry of the network_attributes",
					Description:         "registry of the network_attributes",
				},
				fieldNameClearPending: schema.BoolAttribute{
					Description: "This parameter allows you to clear all the pending OEM network attributes changes. (Update Supported)" +
						"`false`: does not perform any operation. `true`:  discards any pending changes to network attributes, " +
						"or if a job is in scheduled state, removes the job." +
						" `apply_time` value will be ignored and will not have any impact for `clear_pending` operation.",
					MarkdownDescription: "This parameter allows you to clear all the pending OEM network attributes changes. (Update Supported)" +
						"`false`: does not perform any operation. `true`:  discards any pending changes to network attributes, " +
						"or if a job is in scheduled state, removes the job." +
						" `apply_time` value will be ignored and will not have any impact for `clear_pending` operation.",
					Optional: true,
					Validators: []validator.Bool{boolvalidator.AtLeastOneOf(
						path.MatchRelative().AtParent().AtName(fieldNameAttributes),
					)},
				},
				fieldNameAttributes: schema.MapAttribute{
					MarkdownDescription: "dell network attributes. (Update Supported) " +
						"To check allowed attributes please either use the datasource for dell network attributes: data.redfish_network or query " +
						"/redfish/v1/Chassis/System.Embedded.1/NetworkAdapters/NIC.Integrated.1/NetworkDeviceFunctions/NIC.Integrated.1-3-1/" +
						"Oem/Dell/DellNetworkAttributes/NIC.Integrated.1-3-1 to get attributes for NIC. " +
						"To get allowed values for those attributes, check " +
						"/redfish/v1/Registries/NetworkAttributesRegistry_{network_device_function_id}/" +
						"NetworkAttributesRegistry_{network_device_function_id}.json from a Redfish Instance",
					Description: "dell network attributes. (Update Supported) " +
						"To check allowed attributes please either use the datasource for dell network attributes: data.redfish_network or query " +
						"/redfish/v1/Chassis/System.Embedded.1/NetworkAdapters/NIC.Integrated.1/NetworkDeviceFunctions/NIC.Integrated.1-3-1/" +
						"Oem/Dell/DellNetworkAttributes/NIC.Integrated.1-3-1 to get attributes for NIC. " +
						"To get allowed values for those attributes, check " +
						"/redfish/v1/Registries/NetworkAttributesRegistry_{network_device_function_id}/" +
						"NetworkAttributesRegistry_{network_device_function_id}.json from a Redfish Instance",
					ElementType: types.StringType,
					Validators: []validator.Map{
						mapvalidator.AtLeastOneOf(path.MatchRelative().AtParent().AtName(fieldNameClearPending)),
					},
					Optional: true,
					Computed: true,
				},
			},
		},
		"network_attributes": schema.SingleNestedAttribute{
			Description: "Dictionary of network attributes and value for network device function. (Update Supported)" +
				"To check allowed attributes please either use the datasource for dell nic attributes: data.redfish_network or query " +
				"/redfish/v1/Systems/System.Embedded.1/NetworkAdapters/{NetworkAdapterID}/NetworkDeviceFunctions/" +
				"{NetworkDeviceFunctionID}/Settings. " +
				noteMessageUpdateAttrsExclusive +
				noteMessageUpdateOneAttrsOnly +
				"NOTE: Updating network_attributes property may result with an error stating the property is Read-only. " +
				"This may occur if Patch method is performed to change the property to the state that the property is already in or " +
				"because there is dependency of attribute values. For example, if CHAP is disabled, MutualChap becomes a Read-only attribute.",
			MarkdownDescription: "Dictionary of network attributes and value for network device function. (Update Supported)" +
				"To check allowed attributes please either use the datasource for dell nic attributes: data.redfish_network or query " +
				"/redfish/v1/Systems/System.Embedded.1/NetworkAdapters/{NetworkAdapterID}/NetworkDeviceFunctions/" +
				"{NetworkDeviceFunctionID}/Settings. " +
				noteMessageUpdateAttrsExclusive +
				noteMessageUpdateOneAttrsOnly +
				"NOTE: Updating network_attributes property may result with an error stating the property is Read-only. " +
				"This may occur if Patch method is performed to change the property to the state that the property is already in or " +
				"because there is dependency of attribute values. For example, if CHAP is disabled, MutualChap becomes a Read-only attribute.",
			Optional:   true,
			Computed:   true,
			Validators: []validator.Object{objectvalidator.AtLeastOneOf(path.MatchRoot("oem_network_attributes"))},
			Attributes: map[string]schema.Attribute{
				NICComponmentSchemaOdataID: schema.StringAttribute{
					Computed:            true,
					MarkdownDescription: "OData ID for the network device function",
					Description:         "OData ID for the network device function",
				},
				NICComponmentSchemaID: schema.StringAttribute{
					Computed:            true,
					MarkdownDescription: fieldDescriptionNetDevFuncID,
					Description:         fieldDescriptionNetDevFuncID,
				},
				NICComponmentSchemaName: schema.StringAttribute{
					Computed:            true,
					MarkdownDescription: "name of the network device function",
					Description:         "name of the network device function",
				},
				NICComponmentSchemaDescription: schema.StringAttribute{
					Computed:            true,
					MarkdownDescription: "description of the network device function",
					Description:         "description of the network device function",
				},
				NICComponmentSchemaStatus: schema.SingleNestedAttribute{
					MarkdownDescription: "status of the network device function",
					Description:         "status of the network device function",
					Computed:            true,
					Attributes:          NetworkStatusResourceSchema(),
				},
				fieldNameEthernet: schema.SingleNestedAttribute{
					MarkdownDescription: "This type describes Ethernet capabilities, status, and configuration " +
						"for a network device function.  (Update Supported)",
					Description: "This type describes Ethernet capabilities, status, and configuration " +
						"for a network device function.  (Update Supported)",
					Computed:   true,
					Optional:   true,
					Attributes: EthernetResourceSchema(),
					Validators: []validator.Object{
						objectvalidator.ConflictsWith(path.MatchRelative().AtParent().AtName(fieldNameFibreChannel)),
						objectvalidator.AtLeastOneOf(
							path.MatchRelative().AtParent().AtName(fieldNameFibreChannel),
							path.MatchRelative().AtParent().AtName(fieldNameIscsiBoot),
							path.MatchRelative().AtParent().AtName(fieldNameNetDevFuncType),
						),
					},
				},
				fieldNameFibreChannel: schema.SingleNestedAttribute{
					MarkdownDescription: "This type describes Fibre Channel capabilities, status, and configuration " +
						"for a network device function. (Update Supported)",
					Description: "This type describes Fibre Channel capabilities, status, and configuration " +
						"for a network device function. (Update Supported)",
					Computed:   true,
					Optional:   true,
					Attributes: FibreChannelResourceSchema(),
					Validators: []validator.Object{
						objectvalidator.ConflictsWith(path.MatchRelative().AtParent().AtName(fieldNameEthernet)),
						objectvalidator.AtLeastOneOf(
							path.MatchRelative().AtParent().AtName(fieldNameEthernet),
							path.MatchRelative().AtParent().AtName(fieldNameIscsiBoot),
							path.MatchRelative().AtParent().AtName(fieldNameNetDevFuncType),
						),
					},
				},
				fieldNameIscsiBoot: schema.SingleNestedAttribute{
					MarkdownDescription: "The iSCSI boot capabilities, status, and configuration for a network device function. (Update Supported)",
					Description:         "The iSCSI boot capabilities, status, and configuration for a network device function. (Update Supported)",
					Computed:            true,
					Optional:            true,
					Attributes:          ISCSIBootResourceSchema(),
					Validators: []validator.Object{
						objectvalidator.ConflictsWith(path.MatchRelative().AtParent().AtName(fieldNameFibreChannel)),
						objectvalidator.AtLeastOneOf(
							path.MatchRelative().AtParent().AtName(fieldNameEthernet),
							path.MatchRelative().AtParent().AtName(fieldNameFibreChannel),
							path.MatchRelative().AtParent().AtName(fieldNameNetDevFuncType),
						),
					},
				},
				"max_virtual_functions": schema.Int64Attribute{
					Computed:            true,
					MarkdownDescription: "The number of virtual functions that are available for this network device function",
					Description:         "The number of virtual functions that are available for this network device function",
				},
				"net_dev_func_capabilities": schema.ListAttribute{
					ElementType:         types.StringType,
					Computed:            true,
					MarkdownDescription: "An array of capabilities for this network device function",
					Description:         "An array of capabilities for this network device function",
				},
				fieldNameNetDevFuncType: schema.StringAttribute{
					Computed: true,
					Optional: true,
					MarkdownDescription: "The configured capability of this network device function. (Update Supported)" +
						"Accepted values: `Disabled`, `Ethernet`, `FibreChannel`, `iSCSI`, `FibreChannelOverEthernet`, `InfiniBand`.",
					Description: "The configured capability of this network device function. (Update Supported)",
					Validators: []validator.String{stringvalidator.OneOf(
						"Disabled", "Ethernet", "FibreChannel", "iSCSI", "FibreChannelOverEthernet", "InfiniBand",
					)},
				},
				"physical_port_assignment": schema.StringAttribute{
					Computed:            true,
					MarkdownDescription: "A reference to a physical port assignment to this function",
					Description:         "A reference to a physical port assignment to this function",
				},
				"assignable_physical_ports": schema.ListAttribute{
					ElementType:         types.StringType,
					Computed:            true,
					MarkdownDescription: "A reference to assignable physical ports to this function",
					Description:         "A reference to assignable physical ports to this function",
				},
				"assignable_physical_network_ports": schema.ListAttribute{
					ElementType:         types.StringType,
					Computed:            true,
					MarkdownDescription: "A reference to assignable physical network ports to this function",
					Description:         "A reference to assignable physical network ports to this function",
				},
			},
		},
	}
}

// NetworkStatusResourceSchema is a function that returns the schema for Status
func NetworkStatusResourceSchema() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		fieldNameHealth: schema.StringAttribute{
			MarkdownDescription: fieldNameHealth,
			Description:         fieldNameHealth,
			Computed:            true,
		},
		"health_rollup": schema.StringAttribute{
			MarkdownDescription: "health rollup",
			Description:         "health rollup",
			Computed:            true,
		},
		"state": schema.StringAttribute{
			MarkdownDescription: "state of the storage controller",
			Description:         "state of the storage controller",
			Computed:            true,
		},
	}
}

// EthernetResourceSchema is a function that returns the schema for ethernet.
func EthernetResourceSchema() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		fieldNameMACAddress: schema.StringAttribute{
			Computed:            true,
			Optional:            true,
			MarkdownDescription: "The currently configured MAC address. (Update Supported)",
			Description:         "The currently configured MAC address. (Update Supported)",
		},
		fieldNameMTUSize: schema.Int64Attribute{
			Computed:            true,
			Optional:            true,
			MarkdownDescription: "The maximum transmission unit (MTU) configured for this network device function. (Update Supported)",
			Description:         "The maximum transmission unit (MTU) configured for this network device function. (Update Supported)",
		},
		"permanent_mac_address": schema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "The permanent MAC address assigned to this function",
			Description:         "The permanent MAC address assigned to this function",
		},
		fieldNameVLAN: schema.SingleNestedAttribute{
			Computed:            true,
			Optional:            true,
			MarkdownDescription: "The attributes of a VLAN. (Update Supported)",
			Description:         "The attributes of a VLAN. (Update Supported)",
			Attributes:          VLANResourceSchema(),
		},
	}
}

// VLANResourceSchema is a function that returns the schema for vlan.
func VLANResourceSchema() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		fieldNameVLANID: schema.Int64Attribute{
			Computed:            true,
			Optional:            true,
			MarkdownDescription: "The vlan id of the network device function. (Update Supported)",
			Description:         "The vlan id of the network device function. (Update Supported)",
		},
		fieldNameVLANEnabled: schema.BoolAttribute{
			Computed:            true,
			Optional:            true,
			MarkdownDescription: "An indication of whether the VLAN is enabled. (Update Supported)",
			Description:         "An indication of whether the VLAN is enabled. (Update Supported)",
		},
	}
}

// FibreChannelResourceSchema is a function that returns the schema for fibre channel.
func FibreChannelResourceSchema() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		fieldNameAllowFipVlanDiscovery: schema.BoolAttribute{
			Computed:            true,
			Optional:            true,
			MarkdownDescription: "An indication of whether the FCoE Initialization Protocol (FIP) populates the FCoE VLAN ID. (Update Supported)",
			Description:         "An indication of whether the FCoE Initialization Protocol (FIP) populates the FCoE VLAN ID. (Update Supported)",
		},
		fieldNameBootTargets: schema.ListNestedAttribute{
			Description:         "A Fibre Channel boot target configured for a network device function. (Update Supported)",
			MarkdownDescription: "A Fibre Channel boot target configured for a network device function. (Update Supported)",
			Computed:            true,
			Optional:            true,
			NestedObject: schema.NestedAttributeObject{
				Attributes: BootTargetResourceSchema(),
			},
		},
		"fcoe_active_vlan_id": schema.Int64Attribute{
			Computed:            true,
			MarkdownDescription: "The active FCoE VLAN ID",
			Description:         "The active FCoE VLAN ID",
		},
		fieldNameFcoeLocalVlanID: schema.Int64Attribute{
			Computed:            true,
			Optional:            true,
			MarkdownDescription: "The locally configured FCoE VLAN ID. (Update Supported)",
			Description:         "The locally configured FCoE VLAN ID. (Update Supported)",
		},
		"permanent_wwnn": schema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "The permanent World Wide Node Name (WWNN) address assigned to this function",
			Description:         "The permanent World Wide Node Name (WWNN) address assigned to this function",
		},
		"permanent_wwpn": schema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "The permanent World Wide Port Name (WWPN) address assigned to this function",
			Description:         "The permanent World Wide Port Name (WWPN) address assigned to this function",
		},
		fieldNameWWNN: schema.StringAttribute{
			Computed:            true,
			Optional:            true,
			MarkdownDescription: "The currently configured World Wide Node Name (WWNN) address of this function. (Update Supported)",
			Description:         "The currently configured World Wide Node Name (WWNN) address of this function. (Update Supported)",
		},
		fieldNameWWNNSource: schema.StringAttribute{
			Computed: true,
			Optional: true,
			MarkdownDescription: "The configuration source of the World Wide Names (WWN) for this World Wide Node Name (WWNN) and " +
				"World Wide Port Name (WWPN) connection. (Update Supported). Accepted values: `ConfiguredLocally`, `ProvidedByFabric`.",
			Description: "The configuration source of the World Wide Names (WWN) for this World Wide Node Name (WWNN) and " +
				"World Wide Port Name (WWPN) connection. (Update Supported). Accepted values: `ConfiguredLocally`, `ProvidedByFabric`.",
			Validators: []validator.String{stringvalidator.OneOf("ConfiguredLocally", "ProvidedByFabric")},
		},
		fieldNameWWPN: schema.StringAttribute{
			Computed:            true,
			Optional:            true,
			MarkdownDescription: "The currently configured World Wide Port Name (WWPN) address of this function. (Update Supported)",
			Description:         "The currently configured World Wide Port Name (WWPN) address of this function. (Update Supported)",
		},
		"fibre_channel_id": schema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "The Fibre Channel ID that the switch assigns for this interface",
			Description:         "The Fibre Channel ID that the switch assigns for this interface",
		},
	}
}

// BootTargetResourceSchema is a function that returns the schema for boot target.
func BootTargetResourceSchema() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		fieldNameBootPriority: schema.Int64Attribute{
			Computed:            true,
			Optional:            true,
			MarkdownDescription: "The relative priority for this entry in the boot targets array. (Update Supported)",
			Description:         "The relative priority for this entry in the boot targets array. (Update Supported)",
		},
		fieldNameLunID: schema.StringAttribute{
			Computed: true,
			Optional: true,
			MarkdownDescription: "The logical unit number (LUN) ID from which to boot on the device to " +
				"which the corresponding WWPN refers. (Update Supported)",
			Description: "The logical unit number (LUN) ID from which to boot on the device to " +
				"which the corresponding WWPN refers. (Update Supported)",
			Validators: []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		fieldNameWWPN: schema.StringAttribute{
			Computed:            true,
			Optional:            true,
			MarkdownDescription: "The World Wide Port Name (WWPN) from which to boot. (Update Supported)",
			Description:         "The World Wide Port Name (WWPN) from which to boot. (Update Supported)",
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
		},
	}
}

// ISCSIBootResourceSchema is a function that returns the schema for iscsi boot.
func ISCSIBootResourceSchema() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		fieldNameAuthenticationMethod: schema.StringAttribute{
			Computed: true,
			Optional: true,
			MarkdownDescription: "The iSCSI boot authentication method for this network device function. (Update Supported)" +
				"Accepted values: `None`, `CHAP`, `MutualCHAP`.",
			Description: "The iSCSI boot authentication method for this network device function. (Update Supported)",
			Validators:  []validator.String{stringvalidator.OneOf("None", "CHAP", "MutualCHAP")},
		},
		fieldNameChapSec: schema.StringAttribute{
			Computed:            true,
			Optional:            true,
			MarkdownDescription: "The shared secret for CHAP authentication. (Update Supported)",
			Description:         "The shared secret for CHAP authentication. (Update Supported)",
			Sensitive:           true,
		},
		fieldNameChapUsername: schema.StringAttribute{
			Computed:            true,
			Optional:            true,
			MarkdownDescription: "The user name for CHAP authentication. (Update Supported)",
			Description:         "The user name for CHAP authentication. (Update Supported)",
		},
		fieldNameIPAddressType: schema.StringAttribute{
			Computed: true,
			Optional: true,
			MarkdownDescription: "The type of IP address being populated in the iSCSIBoot IP address fields. (Update Supported) " +
				"Accepted values: `IPv4`, `IPv6`.",
			Description: "The type of IP address being populated in the iSCSIBoot IP address fields. (Update Supported) " +
				"Accepted values: `IPv4`, `IPv6`.",
			Validators: []validator.String{stringvalidator.OneOf("IPv4", "IPv6")},
		},
		fieldNameIPMaskDNSViaDHCP: schema.BoolAttribute{
			Computed: true,
			Optional: true,
			MarkdownDescription: "An indication of whether the iSCSI boot initiator uses DHCP to obtain the initiator name, IP address, " +
				"and netmask. (Update Supported)",
			Description: "An indication of whether the iSCSI boot initiator uses DHCP to obtain the initiator name, IP address, " +
				"and netmask. (Update Supported)",
		},
		fieldNameInitiatorDefaultGateway: schema.StringAttribute{
			Computed:            true,
			Optional:            true,
			MarkdownDescription: "The IPv6 or IPv4 iSCSI boot default gateway. (Update Supported)",
			Description:         "The IPv6 or IPv4 iSCSI boot default gateway. (Update Supported)",
		},
		fieldNameInitiatorIPAddress: schema.StringAttribute{
			Computed:            true,
			Optional:            true,
			MarkdownDescription: "The IPv6 or IPv4 address of the iSCSI initiator. (Update Supported)",
			Description:         "The IPv6 or IPv4 address of the iSCSI initiator. (Update Supported)",
		},
		fieldNameInitiatorName: schema.StringAttribute{
			Computed:            true,
			Optional:            true,
			MarkdownDescription: "The iSCSI initiator name. (Update Supported)",
			Description:         "The iSCSI initiator name. (Update Supported)",
		},
		fieldNameInitiatorNetmask: schema.StringAttribute{
			Computed:            true,
			Optional:            true,
			MarkdownDescription: "The IPv6 or IPv4 netmask of the iSCSI boot initiator. (Update Supported)",
			Description:         "The IPv6 or IPv4 netmask of the iSCSI boot initiator. (Update Supported)",
		},
		fieldNameMutualChapSec: schema.StringAttribute{
			Computed:            true,
			Optional:            true,
			MarkdownDescription: "The CHAP secret for two-way CHAP authentication. (Update Supported)",
			Description:         "The CHAP secret for two-way CHAP authentication. (Update Supported)",
			Sensitive:           true,
		},
		fieldNameMutualChapUsername: schema.StringAttribute{
			Computed:            true,
			Optional:            true,
			MarkdownDescription: "The CHAP user name for two-way CHAP authentication. (Update Supported)",
			Description:         "The CHAP user name for two-way CHAP authentication. (Update Supported)",
		},
		fieldNamePrimaryDNS: schema.StringAttribute{
			Computed:            true,
			Optional:            true,
			MarkdownDescription: "The IPv6 or IPv4 address of the primary DNS server for the iSCSI boot initiator. (Update Supported)",
			Description:         "The IPv6 or IPv4 address of the primary DNS server for the iSCSI boot initiator. (Update Supported)",
		},
		fieldNamePrimaryLun: schema.Int64Attribute{
			Computed:            true,
			Optional:            true,
			MarkdownDescription: "The logical unit number (LUN) for the primary iSCSI boot target. (Update Supported)",
			Description:         "The logical unit number (LUN) for the primary iSCSI boot target. (Update Supported)",
		},
		fieldNamePrimaryTargetIPAddress: schema.StringAttribute{
			Computed:            true,
			Optional:            true,
			MarkdownDescription: "The IPv4 or IPv6 address for the primary iSCSI boot target. (Update Supported)",
			Description:         "The IPv4 or IPv6 address for the primary iSCSI boot target. (Update Supported)",
		},
		fieldNamePrimaryTargetName: schema.StringAttribute{
			Computed:            true,
			Optional:            true,
			MarkdownDescription: "The name of the iSCSI primary boot target. (Update Supported)",
			Description:         "The name of the iSCSI primary boot target. (Update Supported)",
		},
		fieldNamePrimaryTargetTCPPort: schema.Int64Attribute{
			Computed:            true,
			Optional:            true,
			MarkdownDescription: "The TCP port for the primary iSCSI boot target. (Update Supported)",
			Description:         "The TCP port for the primary iSCSI boot target. (Update Supported)",
		},
		fieldNamePrimaryVLANEnable: schema.BoolAttribute{
			Computed:            true,
			Optional:            true,
			MarkdownDescription: "An indication of whether the primary VLAN is enabled. (Update Supported)",
			Description:         "An indication of whether the primary VLAN is enabled. (Update Supported)",
		},
		fieldNamePrimaryVLANID: schema.Int64Attribute{
			Computed:            true,
			Optional:            true,
			MarkdownDescription: "The 802.1q VLAN ID to use for iSCSI boot from the primary target. (Update Supported)",
			Description:         "The 802.1q VLAN ID to use for iSCSI boot from the primary target. (Update Supported)",
		},
		fieldNameRouterAdvertisementEnabled: schema.BoolAttribute{
			Computed:            true,
			Optional:            true,
			MarkdownDescription: "An indication of whether IPv6 router advertisement is enabled for the iSCSI boot target. (Update Supported)",
			Description:         "An indication of whether IPv6 router advertisement is enabled for the iSCSI boot target. (Update Supported)",
		},
		fieldNameSecondaryDNS: schema.StringAttribute{
			Computed:            true,
			Optional:            true,
			MarkdownDescription: "The IPv6 or IPv4 address of the secondary DNS server for the iSCSI boot initiator. (Update Supported)",
			Description:         "The IPv6 or IPv4 address of the secondary DNS server for the iSCSI boot initiator. (Update Supported)",
		},
		fieldNameSecondaryLun: schema.Int64Attribute{
			Computed:            true,
			Optional:            true,
			MarkdownDescription: "The logical unit number (LUN) for the secondary iSCSI boot target. (Update Supported)",
			Description:         "The logical unit number (LUN) for the secondary iSCSI boot target. (Update Supported)",
		},
		fieldNameSecondaryTargetIPAddress: schema.StringAttribute{
			Computed:            true,
			Optional:            true,
			MarkdownDescription: "The IPv4 or IPv6 address for the secondary iSCSI boot target. (Update Supported)",
			Description:         "The IPv4 or IPv6 address for the secondary iSCSI boot target. (Update Supported)",
		},
		fieldNameSecondaryTargetName: schema.StringAttribute{
			Computed:            true,
			Optional:            true,
			MarkdownDescription: "The name of the iSCSI secondary boot target. (Update Supported)",
			Description:         "The name of the iSCSI secondary boot target. (Update Supported)",
		},
		fieldNameSecondaryTargetTCPPort: schema.Int64Attribute{
			Computed:            true,
			Optional:            true,
			MarkdownDescription: "The TCP port for the secondary iSCSI boot target. (Update Supported)",
			Description:         "The TCP port for the secondary iSCSI boot target. (Update Supported)",
		},
		fieldNameSecondaryVLANEnable: schema.BoolAttribute{
			Computed:            true,
			Optional:            true,
			MarkdownDescription: "An indication of whether the secondary VLAN is enabled. (Update Supported)",
			Description:         "An indication of whether the secondary VLAN is enabled. (Update Supported)",
		},
		fieldNameSecondaryVLANID: schema.Int64Attribute{
			Computed:            true,
			Optional:            true,
			MarkdownDescription: "The 802.1q VLAN ID to use for iSCSI boot from the secondary target. (Update Supported)",
			Description:         "The 802.1q VLAN ID to use for iSCSI boot from the secondary target. (Update Supported)",
		},
		fieldNameTargetInfoViaDHCP: schema.BoolAttribute{
			Computed: true,
			Optional: true,
			MarkdownDescription: "An indication of whether the iSCSI boot target name, LUN, IP address, " +
				"and netmask should be obtained from DHCP. (Update Supported)",
			Description: "An indication of whether the iSCSI boot target name, LUN, IP address, " +
				"and netmask should be obtained from DHCP. (Update Supported)",
		},
	}
}
