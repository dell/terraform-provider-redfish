/*
Copyright (c) 2025 Dell Inc., or its subsidiaries. All Rights Reserved.

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

	"github.com/stmcginnis/gofish/redfish"
)

// SettingsObject contains OdataID
type SettingsObject struct {
	OdataID string `json:"@odata.id"`
}

// ComputerSystemExtended contains gofish ComputerSystems and Settings
type ComputerSystemExtended struct {
	*redfish.ComputerSystem
	Settings *SettingsObject
}

// ComputerSystems returns Redfish Settings Pointer
func ComputerSystems(comSys *redfish.ComputerSystem) (*ComputerSystemExtended, error) {
	dellSystem := &ComputerSystemExtended{
		ComputerSystem: comSys,
		Settings:       &SettingsObject{},
	}
	rawDataBytes, err := GetRawDataBytesByFieldName(comSys, "RawData")
	if err != nil {
		return dellSystem, err
	}
	if settingsRawData, found := GetNodeFromRawDataHavingDotBytes(rawDataBytes, "@Redfish.Settings"); found == nil {
		if settingsObjectRawData, found := GetNodeFromRawDataBytes(settingsRawData, "SettingsObject"); found == nil {
			var settings *SettingsObject
			if err = json.Unmarshal(settingsObjectRawData, &settings); err == nil {
				dellSystem.Settings = settings
			}
		}
	}

	return dellSystem, nil
}

// GetRawDataBytesByFieldName extracts the rawDataFieldName field from a gofish struct.
func GetRawDataBytesByFieldName(source interface{}, rawDataFieldName string) ([]byte, error) {
	destinationValue := reflect.ValueOf(source)
	destinationType := reflect.TypeOf(source)

	if destinationValue.Kind() == reflect.Ptr {
		destinationValue = destinationValue.Elem()
		destinationType = destinationType.Elem()
	}

	if destinationValue.Kind() != reflect.Struct {
		return nil, fmt.Errorf("source is not a struct")
	}
	destinationValue = destinationValue.FieldByName(rawDataFieldName)
	destinationTye, found := destinationType.FieldByName(rawDataFieldName)
	if !found || destinationTye.Type != reflect.TypeOf([]byte{}) || !destinationValue.IsValid() {
		return nil, fmt.Errorf("source contains no rawData field or rawData not of type []byte")
	}

	return destinationValue.Bytes(), nil
}

// GetNodeFromRawDataHavingDotBytes extracts the node with the given name from the rawData field in a pointer to a struct.
func GetNodeFromRawDataHavingDotBytes(rawDataBytes []byte, nodeName string) (json.RawMessage, error) {
	var jsonNodes map[string]json.RawMessage
	// jsonNodeSplit := "."
	err := json.Unmarshal(rawDataBytes, &jsonNodes)
	if err != nil {
		return nil, err
	}

	for key, value := range jsonNodes {
		if key == nodeName {
			return value, nil
		}
	}

	return nil, fmt.Errorf("node:%s not found in rawData", nodeName)
}
