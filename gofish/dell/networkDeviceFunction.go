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

package dell

import (
	"encoding/json"

	"github.com/stmcginnis/gofish/redfish"
)

// NetworkDeviceFunctionOEM hold the json model of OEM information regarding Dell NetworkDeviceFunctionOEM.
type NetworkDeviceFunctionOEM struct {
	Dell NetworkDeviceFunctionOEMNode
}

// NetworkDeviceFunctionOEMNode hold the json model of OEM data of Dell node.
type NetworkDeviceFunctionOEMNode struct {
	// OEM data for Dell NIC
	DellNIC             NIC
	DellNICPortMetrics  NICPortMetrics
	DellNICCapabilities NICCapabilities
	// OEM data for Dell FC
	DellFC             FC
	DellFCPortMetrics  FCPortMetrics
	DellFCCapabilities FCCapabilities
}

// NIC hold the json model of Dell OEM data for Dell NIC.
type NIC struct {
	Entity
	DeviceDescription        string
	BusNumber                int
	ControllerBIOSVersion    string
	DataBusWidth             string
	EFIVersion               string
	FCoEOffloadMode          string
	FCOSDriverVersion        string
	FQDD                     string
	FamilyVersion            string
	InstanceID               string
	LastSystemInventoryTime  string
	LinkDuplex               string
	LastUpdateTime           string
	MediaType                string
	NICMode                  string
	PCIDeviceID              string
	PCIVendorID              string
	PartNumber               string
	PCISubDeviceID           string
	PCISubVendorID           string
	ProductName              string
	Protocol                 string
	SNAPIState               string
	SerialNumber             string
	SNAPISupport             string
	SlotLength               string
	SlotType                 string
	VPISupport               string
	VendorName               string
	ISCSIOffloadMode         string
	IdentifierType           string
	TransceiverVendorName    string
	CableLengthMetres        int
	PermanentFCOEMACAddress  string
	PermanentiSCSIMACAddress string
	Revision                 string
	TransceiverPartNumber    string
	TransceiverSerialNumber  string
}

// FC hold the json model of Dell OEM data for Dell FC.
type FC struct {
	Entity
	Bus                     int
	CableLengthMetres       int
	ChipType                string
	Device                  int
	DeviceDescription       string
	DeviceName              string
	EFIVersion              string
	FCTapeEnable            string
	FCOSDriverVersion       string
	FCoEOSDriverVersion     string
	FabricLoginRetryCount   int
	FabricLoginTimeout      int
	FamilyVersion           string
	FramePayloadSize        string
	Function                int
	HardZoneAddress         int
	HardZoneEnable          string
	IdentifierType          string
	ISCSIOSDriverVersion    string
	LanDriverVersion        string
	LinkDownTimeout         int
	LoopResetDelay          int
	PartNumber              string
	PortDownRetryCount      int
	PortDownTimeout         int
	PortLoginRetryCount     int
	PortLoginTimeout        int
	ProductName             string
	RDMAOSDriverVersion     string
	Revision                string
	SecondFCTargetLUN       int
	SecondFCTargetWWPN      string
	SerialNumber            string
	TransceiverPartNumber   string
	TransceiverSerialNumber string
	TransceiverVendorName   string
	VendorName              string
}

// NICPortMetrics holds the json model of Dell OEM data for DellNICPortMetrics.
type NICPortMetrics struct {
	Entity
	DiscardedPkts                int
	FQDD                         string
	OSDriverState                string
	RxBytes                      int
	RxBroadcast                  int
	RxErrorPktAlignmentErrors    int
	RxErrorPktFCSErrors          int
	RxJabberPkt                  int
	RxMutlicastPackets           int
	RxPauseXOFFFrames            int
	RxPauseXONFrames             int
	RxRuntPkt                    int
	RxUnicastPackets             int
	TxBroadcast                  int
	TxBytes                      int
	TxErrorPktExcessiveCollision int
	TxErrorPktLateCollision      int
	TxErrorPktMultipleCollision  int
	TxErrorPktSingleCollision    int
	TxMutlicastPackets           int
	TxPauseXOFFFrames            int
	TxPauseXONFrames             int
	StartStatisticTime           string
	StatisticTime                string
	FCCRCErrorCount              int
	FCOELinkFailures             int
	FCOEPktRxCount               int
	FCOEPktTxCount               int
	FCOERxPktDroppedCount        int
	LanFCSRxErrors               int
	LanUnicastPktRXCount         int
	LanUnicastPktTXCount         int
	RDMARxTotalBytes             int
	RDMARxTotalPackets           int
	RDMATotalProtectionErrors    int
	RDMATotalProtocolErrors      int
	RDMATxTotalBytes             int
	RDMATxTotalPackets           int
	RDMATxTotalReadReqPkts       int
	RDMATxTotalSendPkts          int
	RDMATxTotalWritePkts         int
	TxUnicastPackets             int
	PartitionLinkStatus          string
	PartitionOSDriverState       string
	RXInputPowerStatus           string
	RxFalseCarrierDetection      int
	TXBiasCurrentStatus          string
	TXOutputPowerStatus          string
	TemperatureStatus            string
	VoltageStatus                string
	RXInputPowermW               float64
	TXBiasCurrentmA              float64
	TXOutputPowermW              float64
	TemperatureCelsius           float64
	VoltageValueVolts            float64
}

// FCPortMetrics holds the json model of Dell OEM data for DellFCPortMetrics.
type FCPortMetrics struct {
	Entity
	FCInvalidCRCs       int
	FCLinkFailures      int
	FCLossOfSignals     int
	FCRxKBCount         int
	FCRxSequences       int
	FCRxTotalFrames     int
	FCTxKBCount         int
	FCTxSequences       int
	FCTxTotalFrames     int
	OSDriverState       string
	PortStatus          string
	RXInputPowerStatus  string
	RXInputPowermW      float64
	TXBiasCurrentStatus string
	TXBiasCurrentmW     float64
	TXOutputPowerStatus string
	TXOutputPowermW     float64
	TemperatureStatus   string
	TemperatureCelsius  float64
	VoltageStatus       string
	VoltageValueVolts   float64
}

// NICCapabilities holds the json model of Dell OEM data for DellNICCapabilities.
type NICCapabilities struct {
	Entity
	BPESupport                       string
	CongestionNotification           string
	DCBExchangeProtocol              string
	ETS                              string
	EVBModesSupport                  string
	FCoEBootSupport                  string
	FCoEMaxIOsPerSession             int
	FCoEMaxNPIVPerPort               int
	FCoEMaxNumberExchanges           int
	FCoEMaxNumberLogins              int
	FCoEMaxNumberOfFCTargets         int
	FCoEMaxNumberOutStandingCommands int
	FCoEOffloadSupport               string
	FeatureLicensingSupport          string
	FlexAddressingSupport            string
	IPSecOffloadSupport              string
	MACSecSupport                    string
	NWManagementPassThrough          string
	NicPartitioningSupport           string
	OSBMCManagementPassThrough       string
	OnChipThermalSensor              string
	OpenFlowSupport                  string
	PXEBootSupport                   string
	PartitionWOLSupport              string
	PersistencePolicySupport         string
	PriorityFlowControl              string
	RDMASupport                      string
	RemotePHY                        string
	TCPChimneySupport                string
	TCPOffloadEngineSupport          string
	VEB                              string
	VEBVEPAMultiChannel              string
	VEBVEPASingleChannel             string
	VirtualLinkControl               string
	ISCSIBootSupport                 string
	ISCSIOffloadSupport              string
	UEFISupport                      string
}

// FCCapabilities holds the json model of Dell OEM data for DellFCCapabilities.
type FCCapabilities struct {
	Entity
	FCMaxNumberExchanges           int
	FCMaxNumberOutStandingCommands int
	FeatureLicensingSupport        string
	FlexAddressingSupport          string
	OnChipThermalSensor            string
	PersistencePolicySupport       string
	UEFISupport                    string
}

// NetworkDeviceFunctionExtended contains gofish NetworkDeviceFunction data,
// as well as Dell OEM data, DellEthernet, PhysicalPortAssignment, AssignablePhysicalPorts and AssignablePhysicalNetworkPorts.
type NetworkDeviceFunctionExtended struct {
	*redfish.NetworkDeviceFunction
	OemData                            NetworkDeviceFunctionOEM
	DellAssignablePhysicalPorts        []Entity
	DellAssignablePhysicalNetworkPorts []Entity
	DellPhysicalPortAssignment         Entity
	DellEthernet                       Ethernet
	DellNetworkAttributes              Entity
	SettingsObject                     Entity
}

// NICSettings holds the NICSettings entity.
type NICSettings struct {
	Settings NICSettingsObject `json:"@Redfish.Settings"`
}

// NICSettingsObject holds the SettingsObject entity.
type NICSettingsObject struct {
	SettingsObject Entity `json:"SettingsObject"`
}

// Ethernet is the json model of Ethernet including VLAN data.
type Ethernet struct {
	MACAddress          string
	MTUSize             int
	PermanentMACAddress string
	VLAN                VLAN
}

// VLAN is the json model of VLAN
type VLAN struct {
	VLANEnabled bool
	VLANID      int
}

// NetworkDeviceFunction returns a Dell.NetworkDeviceFunction pointer given a redfish.NetworkDeviceFunction pointer from Gofish.
// This is the wrapper that extracts and parses Dell NetworkDeviceFunction OEM data,
// as well as DellAssignablePhysicalPorts, DellAssignablePhysicalNetworkPorts, DellPhysicalPortAssignment and DellEthernet.
// nolint: gocyclo,revive
func NetworkDeviceFunction(deviceFunction *redfish.NetworkDeviceFunction) (*NetworkDeviceFunctionExtended, error) {
	dellNetworkDeviceFunction := &NetworkDeviceFunctionExtended{
		NetworkDeviceFunction: deviceFunction,
		OemData:               NetworkDeviceFunctionOEM{Dell: NetworkDeviceFunctionOEMNode{}},
	}

	rawDataBytes, err := GetRawDataBytes(deviceFunction)
	if err != nil {
		return dellNetworkDeviceFunction, err
	}
	desiredJSONNodes := []string{
		"Oem.Dell.DellNIC", "Oem.Dell.DellNICPortMetrics", "Oem.Dell.DellNICCapabilities",
		"Oem.Dell.DellFC", "Oem.Dell.DellFCPortMetrics", "Oem.Dell.DellFCCapabilities",
		"Ethernet", "AssignablePhysicalPorts", "AssignablePhysicalNetworkPorts",
		"Links.Oem.Dell.DellNetworkAttributes",
	}
	for _, node := range desiredJSONNodes {
		nodeRawData, found := GetNodeFromRawDataBytes(rawDataBytes, node)
		if found != nil {
			continue
		}

		switch node {
		case "Ethernet":
			var ethernet Ethernet
			if err = json.Unmarshal(nodeRawData, &ethernet); err == nil {
				dellNetworkDeviceFunction.DellEthernet = ethernet
			}
		case "AssignablePhysicalPorts":
			var assignablePhysicalPorts []Entity
			if err = json.Unmarshal(nodeRawData, &assignablePhysicalPorts); err == nil {
				dellNetworkDeviceFunction.DellAssignablePhysicalPorts = assignablePhysicalPorts
			}
		case "AssignablePhysicalNetworkPorts":
			var assignablePhysicalNetworkPorts []Entity
			if err = json.Unmarshal(nodeRawData, &assignablePhysicalNetworkPorts); err == nil {
				dellNetworkDeviceFunction.DellAssignablePhysicalNetworkPorts = assignablePhysicalNetworkPorts
			}
		case "Oem.Dell.DellNIC":
			var oemNode NIC
			if err = json.Unmarshal(nodeRawData, &oemNode); err == nil {
				dellNetworkDeviceFunction.OemData.Dell.DellNIC = oemNode
			}
		case "Oem.Dell.DellNICPortMetrics":
			var oemNode NICPortMetrics
			if err = json.Unmarshal(nodeRawData, &oemNode); err == nil {
				dellNetworkDeviceFunction.OemData.Dell.DellNICPortMetrics = oemNode
			}
		case "Oem.Dell.DellNICCapabilities":
			var oemNode NICCapabilities
			if err = json.Unmarshal(nodeRawData, &oemNode); err == nil {
				dellNetworkDeviceFunction.OemData.Dell.DellNICCapabilities = oemNode
			}
		case "Oem.Dell.DellFC":
			var oemNode FC
			if err = json.Unmarshal(nodeRawData, &oemNode); err == nil {
				dellNetworkDeviceFunction.OemData.Dell.DellFC = oemNode
			}
		case "Oem.Dell.DellFCPortMetrics":
			var oemNode FCPortMetrics
			if err = json.Unmarshal(nodeRawData, &oemNode); err == nil {
				dellNetworkDeviceFunction.OemData.Dell.DellFCPortMetrics = oemNode
			}
		case "Oem.Dell.DellFCCapabilities":
			var oemNode FCCapabilities
			if err = json.Unmarshal(nodeRawData, &oemNode); err == nil {
				dellNetworkDeviceFunction.OemData.Dell.DellFCCapabilities = oemNode
			}
		case "Links.Oem.Dell.DellNetworkAttributes":
			var dellNetworkAttributes Entity
			if err = json.Unmarshal(nodeRawData, &dellNetworkAttributes); err == nil {
				dellNetworkDeviceFunction.DellNetworkAttributes = dellNetworkAttributes
			}
		}
	}
	var settings NICSettings
	if err = json.Unmarshal(rawDataBytes, &settings); err == nil {
		dellNetworkDeviceFunction.SettingsObject = settings.Settings.SettingsObject
	}
	return dellNetworkDeviceFunction, nil
}
