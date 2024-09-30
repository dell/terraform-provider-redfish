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

	"github.com/stmcginnis/gofish/common"
	"github.com/stmcginnis/gofish/redfish"
)

// StorageControllerExtended contains gofish storage controller as well as OEM, Description, Assembly and PCIeFunctions data.
type StorageControllerExtended struct {
	*redfish.StorageController
	Oem           StorageControllerOEM
	Description   string
	Assembly      *redfish.Assembly
	PCIeFunctions []string
}

// StorageControllerOEM contains the OEM data.
type StorageControllerOEM struct {
	Dell StorageControllerOEMDell
}

// StorageControllerOEMDell contains the Dell data.
type StorageControllerOEMDell struct {
	DellStorageController DellStorageController
}

// DellStorageController contains the Dell storage controller data.
// nolint: revive
type DellStorageController struct {
	AlarmState                                 string
	AutoConfigBehavior                         string
	BackgroundInitializationRatePercent        int64
	BatteryLearnMode                           string
	BootVirtualDiskFQDD                        string
	CacheSizeInMB                              int64
	CachecadeCapability                        string
	CheckConsistencyMode                       string
	ConnectorCount                             int64
	ControllerBootMode                         string
	ControllerFirmwareVersion                  string
	ControllerMode                             string
	CopybackMode                               string
	CurrentControllerMode                      string
	Device                                     string
	DeviceCardDataBusWidth                     string
	DeviceCardSlotLength                       string
	DeviceCardSlotType                         string
	DriverVersion                              string
	EncryptionCapability                       string
	EncryptionMode                             string
	EnhancedAutoImportForeignConfigurationMode string
	KeyID                                      string
	LastSystemInventoryTime                    string
	LastUpdateTime                             string
	LoadBalanceMode                            string
	MaxAvailablePCILinkSpeed                   string
	MaxDrivesInSpanCount                       int64
	MaxPossiblePCILinkSpeed                    string
	MaxSpansInVolumeCount                      int64
	MaxSupportedVolumesCount                   int64
	PCISlot                                    string
	PatrolReadIterationsCount                  int64
	PatrolReadMode                             string
	PatrolReadRatePercent                      int64
	PatrolReadState                            string
	PatrolReadUnconfiguredAreaMode             string
	PersistentHotspare                         string
	PersistentHotspareMode                     string
	RAIDMode                                   string
	RealtimeCapability                         string
	ReconstructRatePercent                     int64
	RollupStatus                               string
	SASAddress                                 string
	SecurityStatus                             string
	SharedSlotAssignmentAllowed                string
	SlicedVDCapability                         string
	SpindownIdleTimeSeconds                    int64
	SupportControllerBootMode                  string
	SupportEnhancedAutoForeignImport           string
	SupportRAID10UnevenSpans                   string
	SupportedInitializationTypes               []string
	SupportsLKMtoSEKMTransition                string
	T10PICapability                            string
}

// StorageController given redfish.StorageController, returns dell.StorageControllerExtended.
// This is a wrapper that extracts and parses OEM, Description, Assembly and PCIeFunctions data.
func StorageController(storageController *redfish.StorageController) (*StorageControllerExtended, error) {
	storageControllerExtended := &StorageControllerExtended{
		StorageController: storageController,
		Oem:               StorageControllerOEM{},
		Description:       "",
		Assembly:          &redfish.Assembly{},
		PCIeFunctions:     []string{},
	}

	rawDataBytes, err := GetRawDataBytes(storageController)
	if err != nil {
		return storageControllerExtended, err
	}

	if oemRawData, found := GetNodeFromRawDataBytes(rawDataBytes, "Oem"); found == nil {
		var oemData StorageControllerOEM
		if err = json.Unmarshal(oemRawData, &oemData); err == nil {
			storageControllerExtended.Oem = oemData
		}
	}

	if descriptionRawData, found := GetNodeFromRawDataBytes(rawDataBytes, "Description"); found == nil {
		var description string
		if err = json.Unmarshal(descriptionRawData, &description); err == nil {
			storageControllerExtended.Description = description
		}
	}

	if assemblyRawData, found := GetNodeFromRawDataBytes(rawDataBytes, "Assembly"); found == nil {
		var assembly *redfish.Assembly
		if err = json.Unmarshal(assemblyRawData, &assembly); err == nil {
			storageControllerExtended.Assembly = assembly
		}
	}

	type links struct {
		PCIeFunctions common.Links
	}

	var t struct {
		Links links
	}

	err = json.Unmarshal(rawDataBytes, &t)
	if err != nil {
		return storageControllerExtended, err
	}

	storageControllerExtended.PCIeFunctions = t.Links.PCIeFunctions.ToStrings()

	return storageControllerExtended, nil
}
