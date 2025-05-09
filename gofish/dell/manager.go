/*
Copyright (c) 2021-2025 Dell Inc., or its subsidiaries. All Rights Reserved.

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

// ShareParameters struct is used to represent common shared parameters
type ShareParameters struct {
	IgnoreCertificateWarning []string `json:"IgnoreCertificateWarning@Redfish.AllowableValues"`
	ProxySupport             []string `json:"ProxySupport@Redfish.AllowableValues"`
	ProxyType                []string `json:"ProxyType@Redfish.AllowableValues"`
	ShareType                []string `json:"ShareType@Redfish.AllowableValues"`
	Target                   []string `json:"Target@Redfish.AllowableValues"`
}

// ManagerActions stores the OEM Manager actions from Dell
type ManagerActions struct {
	// ExportSystemConfiguration
	ExportSystemConfigurationTarget          string
	ExportSystemConfigurationExportFormat    []string
	ExportSystemConfigurationExportUse       []string
	ExportSystemConfigurationIncludeInExport []string
	ExportSystemConfigurationShareParameters ShareParameters

	// ImportSystemConfiguration
	ImportSystemConfigurationTarget                    string
	ImportSystemConfigurationHostPowerState            []string
	ImportSystemConfigurationImportSystemConfiguration []string
	ImportSystemConfigurationShutdownType              []string
	ImportSystemConfigurationShareParameters           ShareParameters

	// ImportSystemConfigurationPreview
	ImportSystemConfigurationPreviewTarget          string
	ImportSystemConfigurationPreview                []string
	ImportSystemConfigurationPreviewShareParameters ShareParameters

	// ResetToDefaults
	ResetToDefaultsTarget    string
	ResetToDefaultsResetType []string
}

// UnmarshalJSON unmarshals Manager Actions oject from the raw JSON
func (m *ManagerActions) UnmarshalJSON(data []byte) error {
	type ExportSystemConfiguration struct {
		Target          string
		ExportFormat    []string `json:"ExportFormat@Redfish.AllowableValues"`
		ExportUse       []string `json:"ExportUse@Redfish.AllowableValues"`
		IncludeInExport []string `json:"IncludeInExport@Redfish.AllowableValues"`
		ShareParameters ShareParameters
	}

	type ResetToDefaults struct {
		Target    string
		ResetType []string `json:"ResetType@Redfish.AllowableValues"`
	}

	type ImportSystemConfigurationPreview struct {
		Target                           string
		ImportSystemConfigurationPreview []string `json:"ImportSystemConfigurationPreview@Redfish.AllowableValues"`
		ShareParameters                  ShareParameters
	}

	type ImportSystemConfiguration struct {
		Target                    string
		HostPowerState            []string `json:"HostPowerState@Redfish.AllowableValues"`
		ImportSystemConfiguration []string `json:"ImportSystemConfiguration@Redfish.AllowableValues"`
		ShutdownType              []string `json:"ShutdownType@Redfish.AllowableValues"`
		ShareParameters           ShareParameters
	}

	var tempActions struct {
		ExportSystemConfiguration   ExportSystemConfiguration `json:"#OemManager.ExportSystemConfiguration"`
		ImportSystemConfiguration   ImportSystemConfiguration `json:"#OemManager.ImportSystemConfiguration"`
		ExportSystemConfigurationV5 ExportSystemConfiguration `json:"#OemManager.v1_4_0.OemManager#OemManager.ExportSystemConfiguration"`
		ImportSystemConfigurationV5 ImportSystemConfiguration `json:"#OemManager.v1_4_0.OemManager#OemManager.ImportSystemConfiguration"`
		//revive:disable-next-line:line-length-limit
		ImportSystemConfigurationPreview ImportSystemConfigurationPreview `json:"#OemManager.v1_2_0.OemManager#OemManager.ImportSystemConfigurationPreview"`
		ResetToDefaults                  ResetToDefaults                  `json:"DellManager.v1_0_0#DellManager.ResetToDefaults"`
	}

	err := json.Unmarshal(data, &tempActions)
	if err != nil {
		return err
	}

	if tempActions.ExportSystemConfigurationV5.Target != "" {
		tempActions.ExportSystemConfiguration = tempActions.ExportSystemConfigurationV5
	}
	if tempActions.ImportSystemConfigurationV5.Target != "" {
		tempActions.ImportSystemConfiguration = tempActions.ImportSystemConfigurationV5
	}

	// Fill actions
	m.ExportSystemConfigurationTarget = tempActions.ExportSystemConfiguration.Target
	m.ExportSystemConfigurationExportFormat = tempActions.ExportSystemConfiguration.ExportFormat
	m.ExportSystemConfigurationExportUse = tempActions.ExportSystemConfiguration.ExportUse
	m.ExportSystemConfigurationIncludeInExport = tempActions.ExportSystemConfiguration.IncludeInExport
	m.ExportSystemConfigurationShareParameters = tempActions.ExportSystemConfiguration.ShareParameters

	m.ImportSystemConfigurationTarget = tempActions.ImportSystemConfiguration.Target
	m.ImportSystemConfigurationHostPowerState = tempActions.ImportSystemConfiguration.HostPowerState
	m.ImportSystemConfigurationImportSystemConfiguration = tempActions.ImportSystemConfiguration.ImportSystemConfiguration
	m.ImportSystemConfigurationShutdownType = tempActions.ImportSystemConfiguration.ShutdownType
	m.ImportSystemConfigurationShareParameters = tempActions.ImportSystemConfiguration.ShareParameters

	m.ImportSystemConfigurationPreviewTarget = tempActions.ImportSystemConfigurationPreview.Target
	m.ImportSystemConfigurationPreview = tempActions.ImportSystemConfigurationPreview.ImportSystemConfigurationPreview
	m.ImportSystemConfigurationPreviewShareParameters = tempActions.ImportSystemConfigurationPreview.ShareParameters

	m.ResetToDefaultsTarget = tempActions.ResetToDefaults.Target
	m.ResetToDefaultsResetType = tempActions.ResetToDefaults.ResetType

	return nil
}

type managerLinks struct {
	DellAttributes                     common.Links
	DellJobService                     common.Link
	DellLCService                      common.Link
	DellLicensableDeviceCollection     common.Link
	DellLicenseCollection              common.Link
	DellLicenseManagementService       common.Link
	DellOpaqueManagementDataCollection common.Link
	DellPersistentStorageService       common.Link
	DellSwitchConnectionCollection     common.Link
	DellSwitchConnectionService        common.Link
	DellSystemManagementService        common.Link
	DellSystemQuickSyncCollection      common.Link
	DellTimeService                    common.Link
	DellUSBDeviceCollection            common.Link
	DelliDRACCardService               common.Link
	DellvFlashCollection               common.Link
	Jobs                               common.Link
}

// UnmarshalJSON unmarshals Manager Links object from the raw JSON
func (m *managerLinks) UnmarshalJSON(data []byte) error {
	type temp managerLinks
	type Dell struct {
		temp
	}
	var tempLink struct {
		Dell Dell
	}

	err := json.Unmarshal(data, &tempLink)
	if err != nil {
		return err
	}

	*m = managerLinks(tempLink.Dell.temp)

	return nil
}

// DelliDRACCard stores OEM data about Dell iDRAC
type DelliDRACCard struct {
	Entity
	IPMIVersion             string
	LastSystemInventoryTime string
	LastUpdateTime          string
	URLString               string
}

// ManagerOEM hold OEM information regarding Dell Manager (iDRAC)
type ManagerOEM struct {
	DelliDRACCard DelliDRACCard
}

// UnmarshalJSON unmrshals Manager OEM object from the raw JSON
func (m *ManagerOEM) UnmarshalJSON(data []byte) error {
	type temp ManagerOEM
	type Dell struct {
		temp
	}
	var tempOEM struct {
		Dell Dell
	}

	err := json.Unmarshal(data, &tempOEM)
	if err != nil {
		return err
	}

	*m = ManagerOEM(tempOEM.Dell.temp)
	return nil
}

// ManagerExtended contains gofish Manager data, as well as Dell OEM actions, links and data
type ManagerExtended struct {
	*redfish.Manager
	// Actions will hold all Manager Dell OEM actions
	Actions ManagerActions
	links   managerLinks
	// OemData will hold all Manager Dell OEM data
	OemData ManagerOEM
}

// Manager returns a Dell.Manager pointer given a redfish.Manager pointer from Gofish
// This is the wrapper that extracts and parses Dell Manager OEM actions, links and data.
func Manager(manager *redfish.Manager) (*ManagerExtended, error) {
	dellManager := &ManagerExtended{Manager: manager, Actions: ManagerActions{}, links: managerLinks{}, OemData: ManagerOEM{}}
	var actions ManagerActions
	var links managerLinks
	var oemData ManagerOEM

	err := json.Unmarshal(dellManager.OemActions, &actions)
	if err != nil {
		return nil, err
	}
	dellManager.Actions = actions

	err = json.Unmarshal(dellManager.OemLinks, &links)
	if err != nil {
		return nil, err
	}
	dellManager.links = links

	err = json.Unmarshal(dellManager.Oem, &oemData)
	if err != nil {
		return nil, err
	}
	dellManager.OemData = oemData

	return dellManager, nil
}

// DellAttributes return an slice with all configurable dell attributes
func (m *ManagerExtended) DellAttributes() ([]*Attributes, error) {
	return ListReferenceDellAttributes(m.GetClient(), m.links.DellAttributes)
}
