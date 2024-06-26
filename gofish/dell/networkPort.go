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
	"fmt"
	"reflect"
	"strings"

	"github.com/stmcginnis/gofish/redfish"
)

// NetworkPortOEM hold the json model of OEM information regarding Dell NetworkPortOEM.
type NetworkPortOEM struct {
	Dell NetworkPortOEMNode
}

// NetworkPortOEMNode hold the json model of Dell OEM data.
type NetworkPortOEMNode struct {
	DellNetworkTransceiver NetworkTransceiver
}

// NetworkTransceiver is the json model of NetworkTransceiver.
type NetworkTransceiver struct {
	Entity
	DeviceDescription string
	FQDD              string
	IdentifierType    string
	InterfaceType     string
	PartNumber        string
	Revision          string
	SerialNumber      string
	VendorName        string
}

// NetworkPortExtended contains gofish NetworkPort data, as well as Dell OEM data and SupportedLinkCapabilities.
type NetworkPortExtended struct {
	*redfish.NetworkPort
	OemData                           NetworkPortOEM
	SupportedLinkCapabilitiesExtended []SupportedLinkCapabilityExtended
}

// SupportedLinkCapabilityExtended contains gofish SupportedLinkCapability data, as well as LinkSpeedMbps.
type SupportedLinkCapabilityExtended struct {
	AutoSpeedNegotiation  bool
	LinkSpeedMbps         int
	LinkNetworkTechnology string
}

// NetworkPort returns a Dell.NetworkPort pointer given a redfish.NetworkPort pointer from Gofish.
// This is the wrapper that extracts and parses Dell NetworkPort OEM data and SupportedLinkCapabilities.
func NetworkPort(networkPort *redfish.NetworkPort) (*NetworkPortExtended, error) {
	dellNetworkPort := &NetworkPortExtended{
		NetworkPort:                       networkPort,
		OemData:                           NetworkPortOEM{},
		SupportedLinkCapabilitiesExtended: []SupportedLinkCapabilityExtended{},
	}

	rawDataBytes, err := GetRawDataBytes(networkPort)
	if err != nil {
		return dellNetworkPort, err
	}

	if oemRawData, found := GetNodeFromRawDataBytes(rawDataBytes, "Oem"); found == nil {
		var oemData NetworkPortOEM
		if err = json.Unmarshal(oemRawData, &oemData); err == nil {
			dellNetworkPort.OemData = oemData
		}
	}

	if supportedLinkCapabilityRawData, found := GetNodeFromRawDataBytes(rawDataBytes, "SupportedLinkCapabilities"); found == nil {
		var supportedLinkCapabilityExtendedList []SupportedLinkCapabilityExtended
		if err = json.Unmarshal(supportedLinkCapabilityRawData, &supportedLinkCapabilityExtendedList); err == nil {
			dellNetworkPort.SupportedLinkCapabilitiesExtended = supportedLinkCapabilityExtendedList
		}
	}

	return dellNetworkPort, nil
}

// GetRawDataBytes extracts the rawData field from a gofish struct.
func GetRawDataBytes(source interface{}) ([]byte, error) {
	destinationValue := reflect.ValueOf(source)
	destinationType := reflect.TypeOf(source)

	if destinationValue.Kind() == reflect.Ptr {
		destinationValue = destinationValue.Elem()
		destinationType = destinationType.Elem()
	}

	if destinationValue.Kind() != reflect.Struct {
		return nil, fmt.Errorf("source is not a struct")
	}
	rawDataFieldName := "rawData"
	destinationValue = destinationValue.FieldByName(rawDataFieldName)
	destinationTye, found := destinationType.FieldByName(rawDataFieldName)
	if !found || destinationTye.Type != reflect.TypeOf([]byte{}) || !destinationValue.IsValid() {
		return nil, fmt.Errorf("source contains no rawData field or rawData not of type []byte")
	}

	return destinationValue.Bytes(), nil
}

// GetNodeFromRawDataBytes extracts the node with the given name from the rawData field in a pointer to a struct.
func GetNodeFromRawDataBytes(rawDataBytes []byte, nodeName string) (json.RawMessage, error) {
	var jsonNodes map[string]json.RawMessage
	jsonNodeSplit := "."
	err := json.Unmarshal(rawDataBytes, &jsonNodes)
	if err != nil {
		return nil, err
	}

	nodeNames := strings.Split(nodeName, jsonNodeSplit)
	topNodeName := nodeNames[0]
	var secondaryNodeName string
	if len(nodeNames) > 1 {
		secondaryNodeName = strings.TrimPrefix(nodeName, topNodeName+jsonNodeSplit)
	}
	for key, value := range jsonNodes {
		if key == topNodeName {
			if len(secondaryNodeName) > 0 {
				return GetNodeFromRawDataBytes(value, secondaryNodeName)
			}
			return value, nil
		}
	}

	return nil, fmt.Errorf("node:%s not found in rawData", nodeName)
}
