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

package models

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// NICDatasource to is struct for NIC data-source.
type NICDatasource struct {
	ID            types.String       `tfsdk:"id"`
	RedfishServer []RedfishServer    `tfsdk:"redfish_server"`
	NICFilter     *NICFilter         `tfsdk:"nic_filter"`
	NICs          []NetworkInterface `tfsdk:"network_interfaces"`
	NICAttributes types.Map          `tfsdk:"nic_attributes"`
}

// NICFilter is the tfsdk model of NICFilter.
type NICFilter struct {
	Systems []SystemFilter `tfsdk:"systems"`
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
	PhysicalPortNumber            types.String              `tfsdk:"physical_port_number"`
	SupportedEthernetCapabilities []types.String            `tfsdk:"supported_ethernet_capabilities"`
	SupportedLinkCapabilities     []SupportedLinkCapability `tfsdk:"supported_link_capabilities"`
	VendorID                      types.String              `tfsdk:"vendor_id"`
	WakeOnLANEnabled              types.Bool                `tfsdk:"wake_on_lan_enabled"`
	OemData                       NetworkPortOEM            `tfsdk:"oem"`
}

// NetworkPortOEM is the tfsdk model of NetworkPortOEM.
type NetworkPortOEM struct {
	DellNetworkTransceiver *DellNetworkTransceiver `tfsdk:"dell_network_transceiver"`
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
	AutoSpeedNegotiation  types.Bool   `tfsdk:"auto_speed_negotiation"`
	LinkNetworkTechnology types.String `tfsdk:"link_network_technology"`
	LinkSpeedMbps         types.Int64  `tfsdk:"link_speed_mbps"`
}

// NetworkDeviceFunction is the tfsdk model of NetworkDeviceFunction.
// checked with official default values
type NetworkDeviceFunction struct {
	ODataID                        types.String             `tfsdk:"odata_id"`
	ID                             types.String             `tfsdk:"id"`
	Description                    types.String             `tfsdk:"description"`
	Name                           types.String             `tfsdk:"name"`
	Ethernet                       *Ethernet                `tfsdk:"ethernet"`
	FibreChannel                   *FibreChannel            `tfsdk:"fibre_channel"`
	ISCSIBoot                      *ISCSIBoot               `tfsdk:"iscsi_boot"`
	MaxVirtualFunctions            types.Int64              `tfsdk:"max_virtual_functions"`
	NetDevFuncCapabilities         []types.String           `tfsdk:"net_dev_func_capabilities"`
	NetDevFuncType                 types.String             `tfsdk:"net_dev_func_type"`
	Status                         Status                   `tfsdk:"status"`
	PhysicalPortAssignment         types.String             `tfsdk:"physical_port_assignment"`
	AssignablePhysicalPorts        []types.String           `tfsdk:"assignable_physical_ports"`
	AssignablePhysicalNetworkPorts []types.String           `tfsdk:"assignable_physical_network_ports"`
	OemData                        NetworkDeviceFunctionOEM `tfsdk:"oem"`
}

// NetworkDeviceFunctionOEM is the tfsdk model of NetworkDeviceFunctionOEM.
type NetworkDeviceFunctionOEM struct {
	// OEM data for Dell NIC
	DellNIC             *DellNIC             `tfsdk:"dell_nic"`
	DellNICPortMetrics  *DellNICPortMetrics  `tfsdk:"dell_nic_port_metrics"`
	DellNICCapabilities *DellNICCapabilities `tfsdk:"dell_nic_capabilities"`
	// OEM data for Dell FC
	DellFC             *DellFC             `tfsdk:"dell_fc"`
	DellFCPortMetrics  *DellFCPortMetrics  `tfsdk:"dell_fc_port_metrics"`
	DellFCCapabilities *DellFCCapabilities `tfsdk:"dell_fc_port_capabilities"`
}

// DellFCCapabilities is the tfsdk model of DellFCCapabilities.
type DellFCCapabilities struct {
	ODataID                        types.String `tfsdk:"odata_id"`
	ID                             types.String `tfsdk:"id"`
	Name                           types.String `tfsdk:"name"`
	FCMaxNumberExchanges           types.Int64  `tfsdk:"fc_max_number_exchanges"`
	FCMaxNumberOutStandingCommands types.Int64  `tfsdk:"fc_max_number_out_standing_commands"`
	FeatureLicensingSupport        types.String `tfsdk:"feature_licensing_support"`
	FlexAddressingSupport          types.String `tfsdk:"flex_addressing_support"`
	OnChipThermalSensor            types.String `tfsdk:"on_chip_thermal_sensor"`
	PersistencePolicySupport       types.String `tfsdk:"persistence_policy_support"`
	UEFISupport                    types.String `tfsdk:"uefi_support"`
}

// DellFCPortMetrics is the tfsdk model of DellFCPortMetrics.
type DellFCPortMetrics struct {
	ODataID             types.String `tfsdk:"odata_id"`
	ID                  types.String `tfsdk:"id"`
	Name                types.String `tfsdk:"name"`
	FCInvalidCRCs       types.Int64  `tfsdk:"fc_invalid_crcs"`
	FCLinkFailures      types.Int64  `tfsdk:"fc_link_failures"`
	FCLossOfSignals     types.Int64  `tfsdk:"fc_loss_of_signals"`
	FCRxKBCount         types.Int64  `tfsdk:"fc_rx_kb_count"`
	FCRxSequences       types.Int64  `tfsdk:"fc_rx_sequences"`
	FCRxTotalFrames     types.Int64  `tfsdk:"fc_rx_total_frames"`
	FCTxKBCount         types.Int64  `tfsdk:"fc_tx_kb_count"`
	FCTxSequences       types.Int64  `tfsdk:"fc_tx_sequences"`
	FCTxTotalFrames     types.Int64  `tfsdk:"fc_tx_total_frames"`
	OSDriverState       types.String `tfsdk:"os_driver_state"`
	PortStatus          types.String `tfsdk:"port_status"`
	RXInputPowerStatus  types.String `tfsdk:"rx_input_power_status"`
	RXInputPowermW      types.Number `tfsdk:"rx_input_power_mw"`
	TXBiasCurrentStatus types.String `tfsdk:"tx_bias_current_status"`
	TXBiasCurrentmW     types.Number `tfsdk:"tx_bias_current_mw"`
	TXOutputPowerStatus types.String `tfsdk:"tx_output_power_status"`
	TXOutputPowermW     types.Number `tfsdk:"tx_output_power_mw"`
	TemperatureStatus   types.String `tfsdk:"temperature_status"`
	TemperatureCelsius  types.Number `tfsdk:"temperature_celsius"`
	VoltageStatus       types.String `tfsdk:"voltage_status"`
	VoltageValueVolts   types.Number `tfsdk:"voltage_value_volts"`
}

// DellFC is the tfsdk model of DellFC.
type DellFC struct {
	ODataID                 types.String `tfsdk:"odata_id"`
	ID                      types.String `tfsdk:"id"`
	Name                    types.String `tfsdk:"name"`
	Bus                     types.Int64  `tfsdk:"bus"`
	CableLengthMetres       types.Int64  `tfsdk:"cable_length_metres"`
	ChipType                types.String `tfsdk:"chip_type"`
	Device                  types.Int64  `tfsdk:"device"`
	DeviceDescription       types.String `tfsdk:"device_description"`
	DeviceName              types.String `tfsdk:"device_name"`
	EFIVersion              types.String `tfsdk:"efi_version"`
	FCTapeEnable            types.String `tfsdk:"fc_tape_enable"`
	FCOSDriverVersion       types.String `tfsdk:"fc_os_driver_version"`
	FCoEOSDriverVersion     types.String `tfsdk:"fcoe_os_driver_version"`
	FabricLoginRetryCount   types.Int64  `tfsdk:"fabric_login_retry_count"`
	FabricLoginTimeout      types.Int64  `tfsdk:"fabric_login_timeout"`
	FamilyVersion           types.String `tfsdk:"family_version"`
	FramePayloadSize        types.String `tfsdk:"frame_payload_size"`
	Function                types.Int64  `tfsdk:"function"`
	HardZoneAddress         types.Int64  `tfsdk:"hard_zone_address"`
	HardZoneEnable          types.String `tfsdk:"hard_zone_enable"`
	IdentifierType          types.String `tfsdk:"identifier_type"`
	ISCSIOSDriverVersion    types.String `tfsdk:"iscsi_os_driver_version"`
	LanDriverVersion        types.String `tfsdk:"lan_driver_version"`
	LinkDownTimeout         types.Int64  `tfsdk:"link_down_timeout"`
	LoopResetDelay          types.Int64  `tfsdk:"loop_reset_delay"`
	PartNumber              types.String `tfsdk:"part_number"`
	PortDownRetryCount      types.Int64  `tfsdk:"port_down_retry_count"`
	PortDownTimeout         types.Int64  `tfsdk:"port_down_timeout"`
	PortLoginRetryCount     types.Int64  `tfsdk:"port_login_retry_count"`
	PortLoginTimeout        types.Int64  `tfsdk:"port_login_timeout"`
	ProductName             types.String `tfsdk:"product_name"`
	RDMAOSDriverVersion     types.String `tfsdk:"rdma_os_driver_version"`
	Revision                types.String `tfsdk:"revision"`
	SecondFCTargetLUN       types.Int64  `tfsdk:"second_fc_target_lun"`
	SecondFCTargetWWPN      types.String `tfsdk:"second_fc_target_wwpn"`
	SerialNumber            types.String `tfsdk:"serial_number"`
	TransceiverPartNumber   types.String `tfsdk:"transceiver_part_number"`
	TransceiverSerialNumber types.String `tfsdk:"transceiver_serial_number"`
	TransceiverVendorName   types.String `tfsdk:"transceiver_vendor_name"`
	VendorName              types.String `tfsdk:"vendor_name"`
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
	FCOSDriverVersion        types.String `tfsdk:"fc_os_driver_version"`
	FQDD                     types.String `tfsdk:"fqdd"`
	FamilyVersion            types.String `tfsdk:"family_version"`
	InstanceID               types.String `tfsdk:"instance_id"`
	LastSystemInventoryTime  types.String `tfsdk:"last_system_inventory_time"`
	LinkDuplex               types.String `tfsdk:"link_duplex"`
	LastUpdateTime           types.String `tfsdk:"last_update_time"`
	MediaType                types.String `tfsdk:"media_type"`
	NICMode                  types.String `tfsdk:"nic_mode"`
	PCIDeviceID              types.String `tfsdk:"pci_device_id"`
	PCIVendorID              types.String `tfsdk:"pci_vendor_id"`
	PartNumber               types.String `tfsdk:"part_number"`
	PCISubDeviceID           types.String `tfsdk:"pci_sub_device_id"`
	PCISubVendorID           types.String `tfsdk:"pci_sub_vendor_id"`
	ProductName              types.String `tfsdk:"product_name"`
	Protocol                 types.String `tfsdk:"protocol"`
	SNAPIState               types.String `tfsdk:"snapi_state"`
	SerialNumber             types.String `tfsdk:"serial_number"`
	SNAPISupport             types.String `tfsdk:"snapi_support"`
	SlotLength               types.String `tfsdk:"slot_length"`
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
	ODataID                          types.String `tfsdk:"odata_id"`
	ID                               types.String `tfsdk:"id"`
	Name                             types.String `tfsdk:"name"`
	BPESupport                       types.String `tfsdk:"bpe_support"`
	CongestionNotification           types.String `tfsdk:"congestion_notification"`
	DCBExchangeProtocol              types.String `tfsdk:"dcb_exchange_protocol"`
	ETS                              types.String `tfsdk:"ets"`
	EVBModesSupport                  types.String `tfsdk:"evb_modes_support"`
	FCoEBootSupport                  types.String `tfsdk:"fcoe_boot_support"`
	FCoEMaxIOsPerSession             types.Int64  `tfsdk:"fcoe_max_ios_per_session"`
	FCoEMaxNPIVPerPort               types.Int64  `tfsdk:"fcoe_max_npiv_per_port"`
	FCoEMaxNumberExchanges           types.Int64  `tfsdk:"fcoe_max_number_exchanges"`
	FCoEMaxNumberLogins              types.Int64  `tfsdk:"fcoe_max_number_logins"`
	FCoEMaxNumberOfFCTargets         types.Int64  `tfsdk:"fcoe_max_number_of_fc_targets"`
	FCoEMaxNumberOutStandingCommands types.Int64  `tfsdk:"fcoe_max_number_outstanding_commands"`
	FCoEOffloadSupport               types.String `tfsdk:"fcoe_offload_support"`
	FeatureLicensingSupport          types.String `tfsdk:"feature_licensing_support"`
	FlexAddressingSupport            types.String `tfsdk:"flex_addressing_support"`
	IPSecOffloadSupport              types.String `tfsdk:"ipsec_offload_support"`
	MACSecSupport                    types.String `tfsdk:"mac_sec_support"`
	NWManagementPassThrough          types.String `tfsdk:"nw_management_pass_through"`
	NicPartitioningSupport           types.String `tfsdk:"nic_partitioning_support"`
	OSBMCManagementPassThrough       types.String `tfsdk:"os_bmc_management_pass_through"`
	OnChipThermalSensor              types.String `tfsdk:"on_chip_thermal_sensor"`
	OpenFlowSupport                  types.String `tfsdk:"open_flow_support"`
	PXEBootSupport                   types.String `tfsdk:"pxe_boot_support"`
	PartitionWOLSupport              types.String `tfsdk:"partition_wol_support"`
	PersistencePolicySupport         types.String `tfsdk:"persistence_policy_support"`
	PriorityFlowControl              types.String `tfsdk:"priority_flow_control"`
	RDMASupport                      types.String `tfsdk:"rdma_support"`
	RemotePHY                        types.String `tfsdk:"remote_phy"`
	TCPChimneySupport                types.String `tfsdk:"tcp_chimney_support"`
	TCPOffloadEngineSupport          types.String `tfsdk:"tcp_offload_engine_support"`
	VEB                              types.String `tfsdk:"veb"`
	VEBVEPAMultiChannel              types.String `tfsdk:"veb_vepa_multi_channel"`
	VEBVEPASingleChannel             types.String `tfsdk:"veb_vepa_single_channel"`
	VirtualLinkControl               types.String `tfsdk:"virtual_link_control"`
	ISCSIBootSupport                 types.String `tfsdk:"iscsi_boot_support"`
	ISCSIOffloadSupport              types.String `tfsdk:"iscsi_offload_support"`
	UEFISupport                      types.String `tfsdk:"uefi_support"`
}

// DellNICPortMetrics is the tfsdk model of DellNICPortMetrics.
type DellNICPortMetrics struct {
	ODataID                      types.String `tfsdk:"odata_id"`
	ID                           types.String `tfsdk:"id"`
	Name                         types.String `tfsdk:"name"`
	DiscardedPkts                types.Int64  `tfsdk:"discarded_pkts"`
	FQDD                         types.String `tfsdk:"fqdd"`
	OSDriverState                types.String `tfsdk:"os_driver_state"`
	RxBytes                      types.Int64  `tfsdk:"rx_bytes"`
	RxBroadcast                  types.Int64  `tfsdk:"rx_broadcast"`
	RxErrorPktAlignmentErrors    types.Int64  `tfsdk:"rx_error_pkt_alignment_errors"`
	RxErrorPktFCSErrors          types.Int64  `tfsdk:"rx_error_pkt_fcs_errors"`
	RxJabberPkt                  types.Int64  `tfsdk:"rx_jabber_pkt"`
	RxMutlicastPackets           types.Int64  `tfsdk:"rx_mutlicast_packets"`
	RxPauseXOFFFrames            types.Int64  `tfsdk:"rx_pause_xoff_frames"`
	RxPauseXONFrames             types.Int64  `tfsdk:"rx_pause_xon_frames"`
	RxRuntPkt                    types.Int64  `tfsdk:"rx_runt_pkt"`
	RxUnicastPackets             types.Int64  `tfsdk:"rx_unicast_packets"`
	TxBroadcast                  types.Int64  `tfsdk:"tx_broadcast"`
	TxBytes                      types.Int64  `tfsdk:"tx_bytes"`
	TxErrorPktExcessiveCollision types.Int64  `tfsdk:"tx_error_pkt_excessive_collision"`
	TxErrorPktLateCollision      types.Int64  `tfsdk:"tx_error_pkt_late_collision"`
	TxErrorPktMultipleCollision  types.Int64  `tfsdk:"tx_error_pkt_multiple_collision"`
	TxErrorPktSingleCollision    types.Int64  `tfsdk:"tx_error_pkt_single_collision"`
	TxMutlicastPackets           types.Int64  `tfsdk:"tx_mutlicast_packets"`
	TxPauseXOFFFrames            types.Int64  `tfsdk:"tx_pause_xoff_frames"`
	TxPauseXONFrames             types.Int64  `tfsdk:"tx_pause_xon_frames"`
	StartStatisticTime           types.String `tfsdk:"start_statistic_time"`
	StatisticTime                types.String `tfsdk:"statistic_time"`
	FCCRCErrorCount              types.Int64  `tfsdk:"fc_crc_error_count"`
	FCOELinkFailures             types.Int64  `tfsdk:"fcoe_link_failures"`
	FCOEPktRxCount               types.Int64  `tfsdk:"fcoe_pkt_rx_count"`
	FCOEPktTxCount               types.Int64  `tfsdk:"fcoe_pkt_tx_count"`
	FCOERxPktDroppedCount        types.Int64  `tfsdk:"fcoe_rx_pkt_dropped_count"`
	LanFCSRxErrors               types.Int64  `tfsdk:"lan_fcs_rx_errors"`
	LanUnicastPktRXCount         types.Int64  `tfsdk:"lan_unicast_pkt_rx_count"`
	LanUnicastPktTXCount         types.Int64  `tfsdk:"lan_unicast_pkt_tx_count"`
	RDMARxTotalBytes             types.Int64  `tfsdk:"rdma_rx_total_bytes"`
	RDMARxTotalPackets           types.Int64  `tfsdk:"rdma_rx_total_packets"`
	RDMATotalProtectionErrors    types.Int64  `tfsdk:"rdma_total_protection_errors"`
	RDMATotalProtocolErrors      types.Int64  `tfsdk:"rdma_total_protocol_errors"`
	RDMATxTotalBytes             types.Int64  `tfsdk:"rdma_tx_total_bytes"`
	RDMATxTotalPackets           types.Int64  `tfsdk:"rdma_tx_total_packets"`
	RDMATxTotalReadReqPkts       types.Int64  `tfsdk:"rdma_tx_total_read_req_pkts"`
	RDMATxTotalSendPkts          types.Int64  `tfsdk:"rdma_tx_total_send_pkts"`
	RDMATxTotalWritePkts         types.Int64  `tfsdk:"rdma_tx_total_write_pkts"`
	TxUnicastPackets             types.Int64  `tfsdk:"tx_unicast_packets"`
	PartitionLinkStatus          types.String `tfsdk:"partition_link_status"`
	PartitionOSDriverState       types.String `tfsdk:"partition_os_driver_state"`
	RXInputPowerStatus           types.String `tfsdk:"rx_input_power_status"`
	RxFalseCarrierDetection      types.Int64  `tfsdk:"rx_false_carrier_detection"`
	TXBiasCurrentStatus          types.String `tfsdk:"tx_bias_current_status"`
	TXOutputPowerStatus          types.String `tfsdk:"tx_output_power_status"`
	TemperatureStatus            types.String `tfsdk:"temperature_status"`
	VoltageStatus                types.String `tfsdk:"voltage_status"`
	RXInputPowermW               types.Number `tfsdk:"rx_input_power_mw"`
	TXBiasCurrentmA              types.Number `tfsdk:"tx_bias_current_ma"`
	TXOutputPowermW              types.Number `tfsdk:"tx_output_power_mw"`
	TemperatureCelsius           types.Number `tfsdk:"temperature_celsius"`
	VoltageValueVolts            types.Number `tfsdk:"voltage_value_volts"`
}

// Ethernet is the tfsdk model of Ethernet.
type Ethernet struct {
	MACAddress          types.String `tfsdk:"mac_address"`
	MTUSize             types.Int64  `tfsdk:"mtu_size"`
	PermanentMACAddress types.String `tfsdk:"permanent_mac_address"`
	VLAN                VLAN         `tfsdk:"vlan"`
}

// VLAN is the tfsdk model of VLAN
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
	FibreChannelId        types.String `tfsdk:"fibre_channel_id"`
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
	MutualCHAPSecret           types.String `tfsdk:"mutual_chap_secret"`   // only for 6.x 7.x
	MutualCHAPUsername         types.String `tfsdk:"mutual_chap_username"` // only for 6.x 7.x
	PrimaryDNS                 types.String `tfsdk:"primary_dns"`
	PrimaryLUN                 types.Int64  `tfsdk:"primary_lun"`
	PrimaryTargetIPAddress     types.String `tfsdk:"primary_target_ip_address"`
	PrimaryTargetName          types.String `tfsdk:"primary_target_name"`
	PrimaryTargetTCPPort       types.Int64  `tfsdk:"primary_target_tcp_port"`
	PrimaryVLANEnable          types.Bool   `tfsdk:"primary_vlan_enable"`
	PrimaryVLANId              types.Int64  `tfsdk:"primary_vlan_id"`
	RouterAdvertisementEnabled types.Bool   `tfsdk:"router_advertisement_enabled"` // only for 6.x 7.x
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
