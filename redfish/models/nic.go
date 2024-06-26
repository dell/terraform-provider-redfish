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

package models

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// NICDatasource to is struct for NIC data-source.
type NICDatasource struct {
	ID            types.String       `tfsdk:"id"`
	RedfishServer []RedfishServer    `tfsdk:"redfish_server"`
	NICFilter     []SystemFilter     `tfsdk:"nic_filter"`
	NICs          []NetworkInterface `tfsdk:"network_interfaces"`
	NICAttributes types.Map          `tfsdk:"nic_attributes"`
}

// SystemFilter is the tfsdk model of SystemFilter.
type SystemFilter struct {
	SystemID        types.String           `tfsdk:"system_id"`
	NetworkAdapters []NetworkAdapterFilter `tfsdk:"network_adapters"`
}

// NetworkAdapterFilter is the tfsdk model of NetworkAdapterFilter.
type NetworkAdapterFilter struct {
	NetworkAdapterID         types.String   `tfsdk:"network_adapter_id"`
	NetworkPortIDs           []types.String `tfsdk:"network_port_ids"`
	NetworkDeviceFunctionIDs []types.String `tfsdk:"network_device_function_ids"`
}

// NetworkInterface is the tfsdk model of NetworkInterface.
// todo TTHE use gofish
type NetworkInterface struct {
	ODataID                types.String            `tfsdk:"odata_id"`
	ID                     types.String            `tfsdk:"id"`
	Description            types.String            `tfsdk:"description"`
	Name                   types.String            `tfsdk:"name"`
	Status                 Status                  `tfsdk:"status"`
	NetworkAdapter         NetworkAdapter          `tfsdk:"network_adapter"`
	NetworkPorts           []NetworkPort           `tfsdk:"network_ports"`
	NetworkDeviceFunctions []NetworkDeviceFunction `tfsdk:"network_device_functions"`
}

// NetworkAdapter is the tfsdk model of NetworkAdapter.
type NetworkAdapter struct {
	ODataID      types.String       `tfsdk:"odata_id"`
	ID           types.String       `tfsdk:"id"`
	Description  types.String       `tfsdk:"description"`
	Name         types.String       `tfsdk:"name"`
	Manufacturer types.String       `tfsdk:"manufacturer"`
	Model        types.String       `tfsdk:"model"`
	Controllers  []NetworkCollector `tfsdk:"controllers"`
	PartNumber   types.String       `tfsdk:"part_number"`
	SerialNumber types.String       `tfsdk:"serial_number"`
	Status       Status             `tfsdk:"status"`
}

// NetworkCollector - A network controller ASIC that makes up part of a network adapter.
type NetworkCollector struct {
	FirmwarePackageVersion types.String           `tfsdk:"firmware_package_version"`
	ControllerCapabilities ControllerCapabilities `tfsdk:"controller_capabilities"`
}

// ControllerCapabilities sis the tfsdk model of ControllerCapabilities.
type ControllerCapabilities struct {
	DataCenterBridging    DataCenterBridging    `tfsdk:"data_center_bridging"`
	NPAR                  NPAR                  `tfsdk:"npar"`
	NPIV                  NPIV                  `tfsdk:"npiv"`
	VirtualizationOffload VirtualizationOffload `tfsdk:"virtualization_offload"`
}

// DataCenterBridging is the tfsdk model of DataCenterBridging.
type DataCenterBridging struct {
	Capable types.Bool `tfsdk:"capable"`
}

// NPAR is the tfsdk model of NPAR.
type NPAR struct {
	NparCapable types.Bool `tfsdk:"npar_capable"`
	NparEnabled types.Bool `tfsdk:"npar_enabled"`
}

// NPIV is the tfsdk model of NPIV.
type NPIV struct {
	MaxDeviceLogins types.Int64 `tfsdk:"max_device_logins"`
	MaxPortLogins   types.Int64 `tfsdk:"max_port_logins"`
}

// VirtualizationOffload is the tfsdk model of VirtualizationOffload.
type VirtualizationOffload struct {
	SRIOV           SRIOV           `tfsdk:"sriov"`
	VirtualFunction VirtualFunction `tfsdk:"virtual_function"`
}

// SRIOV is the tfsdk model of SRIOV.
type SRIOV struct {
	SRIOVVEPACapable types.Bool `tfsdk:"sriov_vepa_capable"`
}

// VirtualFunction is the tfsdk model of VirtualFunction.
type VirtualFunction struct {
	DeviceMaxCount         types.Int64 `tfsdk:"device_max_count"`
	MinAssignmentGroupSize types.Int64 `tfsdk:"min_assignment_group_size"`
	NetworkPortMaxCount    types.Int64 `tfsdk:"network_port_max_count"`
}

// NetworkPort is the tfsdk model of NetworkPort.
// todo TTHE use gofish   the oem content need furthuer parse
// checked with official default values
type NetworkPort struct {
	ODataID                       types.String              `tfsdk:"odata_id"`
	ID                            types.String              `tfsdk:"id"`
	Description                   types.String              `tfsdk:"description"`
	Name                          types.String              `tfsdk:"name"`
	Status                        Status                    `tfsdk:"status"`
	ActiveLinkTechnology          types.String              `tfsdk:"active_link_technology"`
	AssociatedNetworkAddresses    []types.String            `tfsdk:"associated_network_addresses"`
	CurrentLinkSpeedMbps          types.Int64               `tfsdk:"current_link_speed_mbps"`
	EEEEnabled                    types.Bool                `tfsdk:"eee_enabled"`
	FlowControlConfiguration      types.String              `tfsdk:"flow_control_configuration"`
	FlowControlStatus             types.String              `tfsdk:"flow_control_status"`
	LinkStatus                    types.String              `tfsdk:"link_status"`
	NetDevFuncMaxBWAlloc          []NetDevFuncMaxBWAlloc    `tfsdk:"net_dev_func_max_bw_alloc"`
	NetDevFuncMinBWAlloc          []NetDevFuncMinBWAlloc    `tfsdk:"net_dev_func_min_bw_alloc"`
	PhysicalPortNumber            types.Int64               `tfsdk:"physical_port_number"`
	SupportedEthernetCapabilities []types.String            `tfsdk:"supported_ethernet_capabilities"`
	SupportedLinkCapabilities     []SupportedLinkCapability `tfsdk:"supported_link_capabilities"`
	VendorID                      types.String              `tfsdk:"vendor_id"`
	WakeOnLANEnabled              types.Bool                `tfsdk:"wake_on_lan_enabled"`
	OemData                       NetworkPortOEM            `tfsdk:"oem"`
	// rawData []byte //use this to get oem content todo TTHE
}

// NetworkPortOEM is the tfsdk model of NetworkPortOEM.
type NetworkPortOEM struct {
	DellNetworkTransceiver DellNetworkTransceiver `tfsdk:"dell_network_transceiver"`
}

// DellNetworkTransceiver is the tfsdk model of DellNetworkTransceiver.
type DellNetworkTransceiver struct {
	ODataID           types.String `tfsdk:"odata_id"`
	ID                types.String `tfsdk:"id"`
	DeviceDescription types.String `tfsdk:"device_description"`
	Name              types.String `tfsdk:"name"`
	FQDD              types.String `tfsdk:"fqdd"`
	IdentifierType    types.String `tfsdk:"identifier_type"`
	InterfaceType     types.String `tfsdk:"interface_type"`
	PartNumber        types.String `tfsdk:"part_number"`
	Revision          types.String `tfsdk:"revision"`
	SerialNumber      types.String `tfsdk:"serial_number"`
	VendorName        types.String `tfsdk:"vendor_name"`
}

// NetDevFuncMaxBWAlloc is the tfsdk model of NetDevFuncMaxBWAlloc.
type NetDevFuncMaxBWAlloc struct {
	MaxBWAllocPercent     types.Int64  `tfsdk:"max_bw_alloc_percent"`
	NetworkDeviceFunction types.String `tfsdk:"network_device_function"`
}

// NetDevFuncMinBWAlloc is the tfsdk model of NetDevFuncMinBWAlloc.
type NetDevFuncMinBWAlloc struct {
	MinBWAllocPercent     types.Int64  `tfsdk:"min_bw_alloc_percent"`
	NetworkDeviceFunction types.String `tfsdk:"network_device_function"`
}

// SupportedLinkCapability is the tfsdk model of SupportedLinkCapability.
type SupportedLinkCapability struct {
	AutoSpeedNegotiation  types.Bool    `tfsdk:"auto_speed_negotiation"`
	CapableLinkSpeedMbps  []types.Int64 `tfsdk:"capable_link_speed_mbps"` // todo not in same of actual api, may remove this param
	LinkNetworkTechnology types.String  `tfsdk:"link_network_technology"`
	LinkSpeedMbps         types.Int64   `tfsdk:"link_speed_mbps"` //todo TTHE need to get from raw data
}

// NetworkDeviceFunction is the tfsdk model of NetworkDeviceFunction.
// checked with official default values
type NetworkDeviceFunction struct {
	ODataID                types.String   `tfsdk:"odata_id"`
	ID                     types.String   `tfsdk:"id"`
	Description            types.String   `tfsdk:"description"`
	Name                   types.String   `tfsdk:"name"`
	Ethernet               Ethernet       `tfsdk:"ethernet"`
	FibreChannel           FibreChannel   `tfsdk:"fibre_channel"`
	ISCSIBoot              ISCSIBoot      `tfsdk:"iscsi_boot"`
	MaxVirtualFunctions    types.Int64    `tfsdk:"max_virtual_functions"`
	NetDevFuncCapabilities []types.String `tfsdk:"net_dev_func_capabilities"`
	NetDevFuncType         types.String   `tfsdk:"net_dev_func_type"`
	Status                 Status         `tfsdk:"status"`

	AssignablePhysicalPortsCount   types.Int64              `tfsdk:"assignable_physical_ports_count"`
	PhysicalPortAssignment         types.String             `tfsdk:"physical_port_assignment"`          //todo TTHE this should get from raw data, this could be in the link or it has direct value
	AssignablePhysicalPorts        []types.String           `tfsdk:"assignable_physical_ports"`         //todo TTHE  this should get from raw data
	AssignablePhysicalNetworkPorts []types.String           `tfsdk:"assignable_physical_network_ports"` //todo TTHE  this should get from raw data
	OemData                        NetworkDeviceFunctionOEM `tfsdk:"oem"`

	// rawData []byte
	// OEM {} actual response `tfsdk:"oem"` //todo TTHE ome has alot params
	// ony 7.x has, but also find defintion in official api guide
	/*{
		"Oem": {
		  "Dell": {
			"DellNICPortMetrics": {
			  "RXInputPowerStatus": null,
			  "RXInputPowermW": null,
			  "TXBiasCurrentStatus": null,
			  "TXBiasCurrentmA": null,
			  "TXOutputPowerStatus": null,
			  "TXOutputPowermW": null,
			  "TemperatureCelsius": null,
			  "TemperatureStatus": null,
			  "VoltageStatus": null,
			  "VoltageValueVolts": null,
			}
		  }
		},
	  }*/
}

// NetworkDeviceFunctionOEM is the tfsdk model of NetworkDeviceFunctionOEM.
type NetworkDeviceFunctionOEM struct {
	DellNIC             DellNIC             `tfsdk:"dell_nic"`
	DellNICCapabilities DellNICCapabilities `tfsdk:"dell_nic_port_metrics"`
	DellNICPortMetrics  DellNICPortMetrics  `tfsdk:"dell_nic_port_metrics"`
}

// DellNIC is the tfsdk model of DellNIC.
type DellNIC struct {
	ODataID                  types.String `tfsdk:"odata_id"`
	ID                       types.String `tfsdk:"id"`
	DeviceDescription        types.String `tfsdk:"device_description"`
	Name                     types.String `tfsdk:"name"`
	BusNumber                types.Int64  `tfsdk:"bus_number"`
	ControllerBIOSVersion    types.String `tfsdk:"controller_bios_version"`
	DataBusWidth             types.String `tfsdk:"data_bus_width"`
	EFIVersion               types.String `tfsdk:"efi_version"`
	FCoEOffloadMode          types.String `tfsdk:"fcoe_offload_mode"`
	FQDD                     types.String `tfsdk:"fqdd"`
	FamilyVersion            types.String `tfsdk:"family_version"`
	InstanceID               types.String `tfsdk:"instance_id"`
	LastSystemInventoryTime  types.String `tfsdk:"last_system_inventory_time"`
	LinkDuplex               types.String `tfsdk:"link_duplex"`
	LastUpdateTime           types.String `tfsdk:"last_update_time"`
	MediaType                types.String `tfsdk:"media_type"`
	NICMode                  types.String `tfsdk:"nic_mode"`
	PCIDeviceID              types.String `tfsdk:"pci_device_id"`
	PCIVendorID              types.Int64  `tfsdk:"pci_vendor_id"`
	PartNumber               types.String `tfsdk:"part_number"`
	PCISubDeviceID           types.String `tfsdk:"pci_sub_device_id"`
	PCISubVendorID           types.Int64  `tfsdk:"pci_sub_vendor_id"`
	ProductName              types.String `tfsdk:"product_name"`
	Protocol                 types.String `tfsdk:"protocol"`
	SNAPIState               types.String `tfsdk:"snapi_state"`
	SerialNumber             types.String `tfsdk:"serial_number"`
	SNAPISupport             types.String `tfsdk:"snapi_support"`
	SlotLength               types.Int64  `tfsdk:"slot_length"`
	SlotType                 types.String `tfsdk:"slot_type"`
	VPISupport               types.String `tfsdk:"vpi_support"`
	VendorName               types.String `tfsdk:"vendor_name"`
	ISCSIOffloadMode         types.String `tfsdk:"iscsi_offload_mode"`
	IdentifierType           types.String `tfsdk:"identifier_type"`
	TransceiverVendorName    types.String `tfsdk:"transceiver_vendor_name"`
	CableLengthMetres        types.Int64  `tfsdk:"cable_length_metres"`
	PermanentFCOEMACAddress  types.String `tfsdk:"permanent_fcoe_emac_address"`
	PermanentiSCSIMACAddress types.String `tfsdk:"permanent_iscsi_emac_address"`
	Revision                 types.String `tfsdk:"revision"`
	TransceiverPartNumber    types.String `tfsdk:"transceiver_part_number"`
	TransceiverSerialNumber  types.String `tfsdk:"transceiver_serial_number"`
}

// DellNICCapabilities is the tfsdk model of DellNICCapabilities.
type DellNICCapabilities struct {
	//todo TTHE
	// "DellNICCapabilities": {
	//     "@odata.context": "/redfish/v1/$metadata#DellNICCapabilities.DellNICCapabilities",
	//     "@odata.id": "/redfish/v1/Chassis/System.Embedded.1/NetworkAdapters/NIC.Embedded.1/NetworkDeviceFunctions/NIC.Embedded.1-1-1/Oem/Dell/DellNICCapabilities/NIC.Embedded.1-1-1",
	//     "@odata.type": "#DellNICCapabilities.v1_2_0.DellNICCapabilities",
	//     "BPESupport": "NotSupported",
	//     "CongestionNotification": "NotSupported",
	//     "DCBExchangeProtocol": "NotSupported",
	//     "ETS": "NotSupported",
	//     "EVBModesSupport": "NotSupported",
	//     "FCoEBootSupport": "NotSupported",
	//     "FCoEMaxIOsPerSession": 0,
	//     "FCoEMaxNPIVPerPort": 0,
	//     "FCoEMaxNumberExchanges": 0,
	//     "FCoEMaxNumberLogins": 0,
	//     "FCoEMaxNumberOfFCTargets": 0,
	//     "FCoEMaxNumberOutStandingCommands": 0,
	//     "FCoEOffloadSupport": "NotSupported",
	//     "FeatureLicensingSupport": "NotSupported",
	//     "FlexAddressingSupport": "Supported",
	//     "IPSecOffloadSupport": "NotSupported",
	//     "Id": "NIC.Embedded.1-1-1",
	//     "MACSecSupport": "NotSupported",
	//     "NWManagementPassThrough": "Supported",
	//     "Name": "DellNICCapabilities",
	//     "NicPartitioningSupport": "NotSupported",
	//     "OSBMCManagementPassThrough": "Supported",
	//     "OnChipThermalSensor": "Supported",
	//     "OpenFlowSupport": "NotSupported",
	//     "PXEBootSupport": "Supported",
	//     "PartitionWOLSupport": "NotSupported",
	//     "PersistencePolicySupport": "Supported",
	//     "PriorityFlowControl": "NotSupported",
	//     "RDMASupport": "NotSupported",
	//     "RemotePHY": "NotSupported",
	//     "TCPChimneySupport": "NotSupported",
	//     "TCPOffloadEngineSupport": "NotSupported",
	//     "VEB": "NotSupported",
	//     "VEBVEPAMultiChannel": "NotSupported",
	//     "VEBVEPASingleChannel": "NotSupported",
	//     "VirtualLinkControl": "NotSupported",
	//     "iSCSIBootSupport": "NotSupported",
	//     "iSCSIOffloadSupport": "NotSupported",
	//     "uEFISupport": "Supported"
	//   },
}

// DellNICPortMetrics is the tfsdk model of DellNICPortMetrics.
type DellNICPortMetrics struct {
	//todo TTHE
	// "DellNICPortMetrics": {
	//     "@odata.context": "/redfish/v1/$metadata#DellNICPortMetrics.DellNICPortMetrics",
	//     "@odata.id": "/redfish/v1/Chassis/System.Embedded.1/NetworkAdapters/NIC.Embedded.1/NetworkDeviceFunctions/NIC.Embedded.1-1-1/Oem/Dell/DellNICPortMetrics/NIC.Embedded.1-1-1",
	//     "@odata.type": "#DellNICPortMetrics.v1_1_1.DellNICPortMetrics",
	//     "DiscardedPkts": 0,
	//     "FCCRCErrorCount": null,
	//     "FCOELinkFailures": null,
	//     "FCOEPktRxCount": null,
	//     "FCOEPktTxCount": null,
	//     "FCOERxPktDroppedCount": null,
	//     "FQDD": "NIC.Embedded.1-1-1",
	//     "Id": "NIC.Embedded.1-1-1",
	//     "LanFCSRxErrors": null,
	//     "LanUnicastPktRXCount": null,
	//     "LanUnicastPktTXCount": null,
	//     "Name": "DellNICPortMetrics",
	//     "OSDriverState": "Non-operational",
	//     "PartitionLinkStatus": null,
	//     "PartitionOSDriverState": null,
	//     "RDMARxTotalBytes": null,
	//     "RDMARxTotalPackets": null,
	//     "RDMATotalProtectionErrors": null,
	//     "RDMATotalProtocolErrors": null,
	//     "RDMATxTotalBytes": null,
	//     "RDMATxTotalPackets": null,
	//     "RDMATxTotalReadReqPkts": null,
	//     "RDMATxTotalSendPkts": null,
	//     "RDMATxTotalWritePkts": null,
	//     "RXInputPowerStatus": null,
	//     "RXInputPowermW": null,
	//     "RxBroadcast": 221,
	//     "RxBytes": 179330,
	//     "RxErrorPktAlignmentErrors": 0,
	//     "RxErrorPktFCSErrors": 0,
	//     "RxFalseCarrierDetection": null,
	//     "RxJabberPkt": 0,
	//     "RxMutlicastPackets": 2051,
	//     "RxPauseXOFFFrames": 0,
	//     "RxPauseXONFrames": 0,
	//     "RxRuntPkt": 0,
	//     "RxUnicastPackets": 0,
	//     "StartStatisticTime": "2023-12-28T01:28:52-06:00",
	//     "StatisticTime": "2023-12-28T01:40:26-06:00",
	//     "TXBiasCurrentStatus": null,
	//     "TXBiasCurrentmA": null,
	//     "TXOutputPowerStatus": null,
	//     "TXOutputPowermW": null,
	//     "TemperatureCelsius": null,
	//     "TemperatureStatus": null,
	//     "TxBroadcast": 0,
	//     "TxBytes": 0,
	//     "TxErrorPktExcessiveCollision": 0,
	//     "TxErrorPktLateCollision": 0,
	//     "TxErrorPktMultipleCollision": 0,
	//     "TxErrorPktSingleCollision": 0,
	//     "TxMutlicastPackets": 0,
	//     "TxPauseXOFFFrames": 0,
	//     "TxPauseXONFrames": 0,
	//     "TxUnicastPackets": 0,
	//     "VoltageStatus": null,
	//     "VoltageValueVolts": null
	//   }
}

// Ethernet is the tfsdk model of Ethernet.
type Ethernet struct {
	MACAddress          types.String `tfsdk:"mac_address"`
	MTUSize             types.Int64  `tfsdk:"mtu_size"`
	PermanentMACAddress types.String `tfsdk:"permanent_mac_address"`
	VLAN                VLAN         `tfsdk:"vlan"`
}

// VLAN is the tfsdk model of VLAN
// todo TTHE need to get from raw data
type VLAN struct {
	VLANEnabled types.Bool  `tfsdk:"vlan_enabled"`
	VLANID      types.Int64 `tfsdk:"vlan_id"`
}

// FibreChannel is the tfsdk model of FibreChannel.
type FibreChannel struct {
	AllowFIPVLANDiscovery types.Bool   `tfsdk:"allow_fip_vlan_discovery"`
	BootTargets           []BootTarget `tfsdk:"boot_targets"`
	FCoEActiveVLANId      types.Int64  `tfsdk:"fcoe_active_vlan_id"`
	FCoELocalVLANId       types.Int64  `tfsdk:"fcoe_local_vlan_id"`
	PermanentWWNN         types.String `tfsdk:"permanent_wwnn"`
	PermanentWWPN         types.String `tfsdk:"permanent_wwpn"`
	WWNN                  types.String `tfsdk:"wwnn"`
	WWNSource             types.String `tfsdk:"wwn_source"`
	WWPN                  types.String `tfsdk:"wwpn"`
}

// ISCSIBoot is the tfsdk model of ISCSIBoot.
type ISCSIBoot struct {
	AuthenticationMethod       types.String `tfsdk:"authentication_method"`
	CHAPSecret                 types.String `tfsdk:"chap_secret"`
	CHAPUsername               types.String `tfsdk:"chap_username"`
	IPAddressType              types.String `tfsdk:"ip_address_type"`
	IPMaskDNSViaDHCP           types.Bool   `tfsdk:"ip_mask_dns_via_dhcp"`
	InitiatorDefaultGateway    types.String `tfsdk:"initiator_default_gateway"`
	InitiatorIPAddress         types.String `tfsdk:"initiator_ip_address"`
	InitiatorName              types.String `tfsdk:"initiator_name"`
	InitiatorNetmask           types.String `tfsdk:"initiator_netmask"`
	MutualCHAPSecret           types.String `tfsdk:"mutual_chap_secret"`   // ony for 6.x 7.x
	MutualCHAPUsername         types.String `tfsdk:"mutual_chap_username"` // ony for 6.x 7.x
	PrimaryDNS                 types.String `tfsdk:"primary_dns"`
	PrimaryLUN                 types.Int64  `tfsdk:"primary_lun"`
	PrimaryTargetIPAddress     types.String `tfsdk:"primary_target_ip_address"`
	PrimaryTargetName          types.String `tfsdk:"primary_target_name"`
	PrimaryTargetTCPPort       types.Int64  `tfsdk:"primary_target_tcp_port"`
	PrimaryVLANEnable          types.Bool   `tfsdk:"primary_vlan_enable"`
	PrimaryVLANId              types.Int64  `tfsdk:"primary_vlan_id"`
	RouterAdvertisementEnabled types.Bool   `tfsdk:"router_advertisement_enabled"` // ony for 6.x 7.x
	SecondaryDNS               types.String `tfsdk:"secondary_dns"`
	SecondaryLUN               types.Int64  `tfsdk:"secondary_lun"`
	SecondaryTargetIPAddress   types.String `tfsdk:"secondary_target_ip_address"`
	SecondaryTargetName        types.String `tfsdk:"secondary_target_name"`
	SecondaryTargetTCPPort     types.Int64  `tfsdk:"secondary_target_tcp_port"`
	SecondaryVLANEnable        types.Bool   `tfsdk:"secondary_vlan_enable"`
	SecondaryVLANId            types.Int64  `tfsdk:"secondary_vlan_id"`
	TargetInfoViaDHCP          types.Bool   `tfsdk:"target_info_via_dhcp"`
}

// BootTarget is the tfsdk model of BootTarget.
type BootTarget struct {
	BootPriority types.Int64  `tfsdk:"boot_priority"`
	LUNID        types.String `tfsdk:"lun_id"`
	WWPN         types.String `tfsdk:"wwpn"`
}

//todo TTHE cheek gofish params - const - enum
