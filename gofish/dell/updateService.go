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

package dell

import (
	"encoding/json"

	"github.com/stmcginnis/gofish/redfish"
)

// UpdateServiceExtended struct extends the gofish UpdateService and includes Dell OEM actions
type UpdateServiceExtended struct {
	*redfish.UpdateService
	// Actions will hold all UpdateService Dell OEM actions
	Actions             UpdateServiceActions
	SimpleUpdateActions SimpleUpdateActions

	FirmwareInventory *redfish.SoftwareInventory
}

// UpdateServiceActions contains Dell OEM actions
type UpdateServiceActions struct {
	// DellUpdateServiceTarget is the URL to be targetted for Dell's update
	DellUpdateServiceTarget string
	// DellUpdateServiceInstallUpon are the installing times
	DellUpdateServiceInstallUpon []string
}

// SimpleUpdate contains SimpleUpdate
type SimpleUpdate struct {
	AllowableValues []string `json:"TransferProtocol@Redfish.AllowableValues"`
	Target          string
}

// SimpleUpdateActions contains SimpleUpdate
type SimpleUpdateActions struct {
	SimpleUpdate SimpleUpdate `json:"#UpdateService.SimpleUpdate"`
}

// UnmarshalJSON unmarshals Dell update service object from raw JSON
func (u *UpdateServiceActions) UnmarshalJSON(data []byte) error {
	type DellUpdateService struct {
		InstallUpon []string `json:"InstallUpon@Redfish.AllowableValues"`
		Target      string
	}
	var t struct {
		DellUpdateService DellUpdateService `json:"DellUpdateService.v1_0_0#DellUpdateService.Install"`
	}

	err := json.Unmarshal(data, &t)
	if err != nil {
		return err
	}

	u.DellUpdateServiceTarget = t.DellUpdateService.Target
	u.DellUpdateServiceInstallUpon = t.DellUpdateService.InstallUpon

	return nil
}

// UpdateService returns a Dell.UpdateServiceExtended pointer given a redfish.UpdateService pointer from gofish library
// This is the wrapper that extracts and parses Dell UpdateService OEM actions
func UpdateService(updateService *redfish.UpdateService) (*UpdateServiceExtended, error) {
	dellUpdate := UpdateServiceExtended{
		UpdateService:       updateService,
		SimpleUpdateActions: SimpleUpdateActions{},
		FirmwareInventory:   &redfish.SoftwareInventory{},
	}
	var oemUpdateService UpdateServiceActions

	err := json.Unmarshal(dellUpdate.OemActions, &oemUpdateService)
	if err != nil {
		return nil, err
	}
	dellUpdate.Actions = oemUpdateService

	rawDataBytes, err := GetRawDataBytes(updateService)
	if err != nil {
		return &dellUpdate, err
	}
	if updateActionRawData, found := GetNodeFromRawDataBytes(rawDataBytes, "Actions"); found == nil {
		var simpleUpdateData SimpleUpdateActions
		if err = json.Unmarshal(updateActionRawData, &simpleUpdateData); err == nil {
			dellUpdate.SimpleUpdateActions = simpleUpdateData
		}
	}

	if inventoryRawData, found := GetNodeFromRawDataBytes(rawDataBytes, "FirmwareInventory"); found == nil {
		var inventoryData *redfish.SoftwareInventory
		if err = json.Unmarshal(inventoryRawData, &inventoryData); err == nil {
			dellUpdate.FirmwareInventory = inventoryData
		}
	}

	return &dellUpdate, nil
}
