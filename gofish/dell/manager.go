package dell

import (
	"encoding/json"

	"github.com/stmcginnis/gofish/common"
	"github.com/stmcginnis/gofish/redfish"
)

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

func (m *ManagerActions) UnmarshalJSON(data []byte) error {
	var tempActions struct {
		ExportSystemConfiguration struct {
			Target          string
			ExportFormat    []string `json:"ExportFormat@Redfish.AllowableValues"`
			ExportUse       []string `json:"ExportUse@Redfish.AllowableValues"`
			IncludeInExport []string `json:"IncludeInExport@Redfish.AllowableValues"`
			ShareParameters ShareParameters
		} `json:"#OemManager.v1_2_0.OemManager#OemManager.ExportSystemConfiguration"`
		ImportSystemConfiguration struct {
			Target                    string
			HostPowerState            []string `json:"HostPowerState@Redfish.AllowableValues"`
			ImportSystemConfiguration []string `json:"ImportSystemConfiguration@Redfish.AllowableValues"`
			ShutdownType              []string `json:"ShutdownType@Redfish.AllowableValues"`
			ShareParameters           ShareParameters
		} `json:"#OemManager.v1_2_0.OemManager#OemManager.ImportSystemConfiguration"`
		ImportSystemConfigurationPreview struct {
			Target                           string
			ImportSystemConfigurationPreview []string `json:"ImportSystemConfigurationPreview@Redfish.AllowableValues"`
			ShareParameters                  ShareParameters
		} `json:"#OemManager.v1_2_0.OemManager#OemManager.ImportSystemConfigurationPreview"`
		ResetToDefaults struct {
			Target    string
			ResetType []string `json:"ResetType@Redfish.AllowableValues"`
		} `json:"DellManager.v1_0_0#DellManager.ResetToDefaults"`
	}

	err := json.Unmarshal(data, &tempActions)
	if err != nil {
		return err
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

func (m *managerLinks) UnmarshalJSON(data []byte) error {
	type temp managerLinks
	var tempLink struct {
		Dell struct {
			temp
		}
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

func (m *ManagerOEM) UnmarshalJSON(data []byte) error {
	type temp ManagerOEM
	var tempOEM struct {
		Dell struct {
			temp
		}
	}

	err := json.Unmarshal(data, &tempOEM)
	if err != nil {
		return err
	}

	*m = ManagerOEM(tempOEM.Dell.temp)
	return nil
}

// Manager contains gofish Manager data, as well as Dell OEM actions, links and data
type Manager struct {
	*redfish.Manager
	// Actions will hold all Manager Dell OEM actions
	Actions ManagerActions
	links   managerLinks
	// OemData will hold all Manager Dell OEM data
	OemData ManagerOEM
}

// DellManager returns a Dell.Manager pointer given a redfish.Manager pointer from Gofish
// This is the wrapper that extracts and parses Dell Manager OEM actions, links and data.
func DellManager(manager *redfish.Manager) (*Manager, error) {
	dellManager := &Manager{Manager: manager, Actions: ManagerActions{}, links: managerLinks{}, OemData: ManagerOEM{}}
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
func (m *Manager) DellAttributes() ([]*DellAttributes, error) {
	return ListReferenceDellAttributes(m.Client, m.links.DellAttributes)
}
