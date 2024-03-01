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
	"strings"
	"testing"

	"github.com/stmcginnis/gofish/redfish"
)

var oemLinksBody = `
{
	"Dell": {
		"@odata.type": "#DellOem.v1_0_0.DellOemLinks",
		"DellAttributes": [
			{
				"@odata.id": "/redfish/v1/Managers/iDRAC.Embedded.1/Oem/Dell/DellAttributes/iDRAC.Embedded.1"
			},
			{
				"@odata.id": "/redfish/v1/Managers/iDRAC.Embedded.1/Oem/Dell/DellAttributes/System.Embedded.1"
			},
			{
				"@odata.id": "/redfish/v1/Managers/iDRAC.Embedded.1/Oem/Dell/DellAttributes/LifecycleController.Embedded.1"
			}
		],
		"DellAttributes@odata.count": 3,
		"DellJobService": {
			"@odata.id": "/redfish/v1/Managers/iDRAC.Embedded.1/Oem/Dell/DellJobService"
		},
		"DellLCService": {
			"@odata.id": "/redfish/v1/Managers/iDRAC.Embedded.1/Oem/Dell/DellLCService"
		},
		"DellLicensableDeviceCollection": {
			"@odata.id": "/redfish/v1/Managers/iDRAC.Embedded.1/Oem/Dell/DellLicensableDevices"
		},
		"DellLicenseCollection": {
			"@odata.id": "/redfish/v1/Managers/iDRAC.Embedded.1/Oem/Dell/DellLicenses"
		},
		"DellLicenseManagementService": {
			"@odata.id": "/redfish/v1/Managers/iDRAC.Embedded.1/Oem/Dell/DellLicenseManagementService"
		},
		"DellOpaqueManagementDataCollection": {
			"@odata.id": "/redfish/v1/Managers/iDRAC.Embedded.1/Oem/Dell/DellOpaqueManagementData"
		},
		"DellPersistentStorageService": {
			"@odata.id": "/redfish/v1/Managers/iDRAC.Embedded.1/Oem/Dell/DellPersistentStorageService"
		},
		"DellSwitchConnectionCollection": {
			"@odata.id": "/redfish/v1/Systems/System.Embedded.1/NetworkPorts/Oem/Dell/DellSwitchConnections"
		},
		"DellSwitchConnectionService": {
			"@odata.id": "/redfish/v1/Systems/System.Embedded.1/Oem/Dell/DellSwitchConnectionService"
		},
		"DellSystemManagementService": {
			"@odata.id": "/redfish/v1/Systems/System.Embedded.1/Oem/Dell/DellSystemManagementService"
		},
		"DellSystemQuickSyncCollection": {
			"@odata.id": "/redfish/v1/Managers/iDRAC.Embedded.1/Oem/Dell/DellSystemQuickSync"
		},
		"DellTimeService": {
			"@odata.id": "/redfish/v1/Managers/iDRAC.Embedded.1/Oem/Dell/DellTimeService"
		},
		"DellUSBDeviceCollection": {
			"@odata.id": "/redfish/v1/Managers/iDRAC.Embedded.1/Oem/Dell/DellUSBDevices"
		},
		"DelliDRACCardService": {
			"@odata.id": "/redfish/v1/Managers/iDRAC.Embedded.1/Oem/Dell/DelliDRACCardService"
		},
		"DellvFlashCollection": {
			"@odata.id": "/redfish/v1/Managers/iDRAC.Embedded.1/Oem/Dell/DellvFlash"
		},
		"Jobs": {
			"@odata.id": "/redfish/v1/Managers/iDRAC.Embedded.1/Oem/Dell/Jobs"
		}
	}
}
`
var oemDataBody = `
		{
			"Dell": {
				"DelliDRACCard": {
					"@odata.context": "/redfish/v1/$metadata#DelliDRACCard.DelliDRACCard",
					"@odata.id": "/redfish/v1/Managers/iDRAC.Embedded.1/Oem/Dell/DelliDRACCard/iDRAC.Embedded.1-1_0x23_IDRACinfo",
					"@odata.type": "#DelliDRACCard.v1_1_0.DelliDRACCard",
					"Description": "An instance of DelliDRACCard will have data specific to the Integrated Dell Remote Access Controller (iDRAC) in the managed system.",
					"IPMIVersion": "2.0",
					"Id": "iDRAC.Embedded.1-1_0x23_IDRACinfo",
					"LastSystemInventoryTime": "2021-06-08T09:12:53+00:00",
					"LastUpdateTime": "2021-06-08T14:44:15+00:00",
					"Name": "DelliDRACCard",
					"URLString": "https://10.0.41.190:443"
				}
			}
		}
`

var oemActions = `
{
	"#OemManager.v1_2_0.OemManager#OemManager.ExportSystemConfiguration": {
		"ExportFormat@Redfish.AllowableValues": [
			"XML",
			"JSON"
		],
		"ExportUse@Redfish.AllowableValues": [
			"Default",
			"Clone",
			"Replace"
		],
		"IncludeInExport@Redfish.AllowableValues": [
			"Default",
			"IncludeReadOnly",
			"IncludePasswordHashValues",
			"IncludeCustomTelemetry"
		],
		"ShareParameters": {
			"IgnoreCertificateWarning@Redfish.AllowableValues": [
				"Disabled",
				"Enabled"
			],
			"ProxySupport@Redfish.AllowableValues": [
				"Disabled",
				"EnabledProxyDefault",
				"Enabled"
			],
			"ProxyType@Redfish.AllowableValues": [
				"HTTP",
				"SOCKS4"
			],
			"ShareType@Redfish.AllowableValues": [
				"LOCAL",
				"NFS",
				"CIFS",
				"HTTP",
				"HTTPS"
			],
			"Target@Redfish.AllowableValues": [
				"ALL",
				"IDRAC",
				"BIOS",
				"NIC",
				"RAID"
			]
		},
		"target": "/redfish/v1/Managers/iDRAC.Embedded.1/Actions/Oem/EID_674_Manager.ExportSystemConfiguration"
	},
	"#OemManager.v1_2_0.OemManager#OemManager.ImportSystemConfiguration": {
		"HostPowerState@Redfish.AllowableValues": [
			"On",
			"Off"
		],
		"ImportSystemConfiguration@Redfish.AllowableValues": [
			"TimeToWait",
			"ImportBuffer"
		],
		"ShareParameters": {
			"IgnoreCertificateWarning@Redfish.AllowableValues": [
				"Disabled",
				"Enabled"
			],
			"ProxySupport@Redfish.AllowableValues": [
				"Disabled",
				"EnabledProxyDefault",
				"Enabled"
			],
			"ProxyType@Redfish.AllowableValues": [
				"HTTP",
				"SOCKS4"
			],
			"ShareType@Redfish.AllowableValues": [
				"LOCAL",
				"NFS",
				"CIFS",
				"HTTP",
				"HTTPS"
			],
			"Target@Redfish.AllowableValues": [
				"ALL",
				"IDRAC",
				"BIOS",
				"NIC",
				"RAID"
			]
		},
		"ShutdownType@Redfish.AllowableValues": [
			"Graceful",
			"Forced",
			"NoReboot"
		],
		"target": "/redfish/v1/Managers/iDRAC.Embedded.1/Actions/Oem/EID_674_Manager.ImportSystemConfiguration"
	},
	"#OemManager.v1_2_0.OemManager#OemManager.ImportSystemConfigurationPreview": {
		"ImportSystemConfigurationPreview@Redfish.AllowableValues": [
			"ImportBuffer"
		],
		"ShareParameters": {
			"IgnoreCertificateWarning@Redfish.AllowableValues": [
				"Disabled",
				"Enabled"
			],
			"ProxySupport@Redfish.AllowableValues": [
				"Disabled",
				"EnabledProxyDefault",
				"Enabled"
			],
			"ProxyType@Redfish.AllowableValues": [
				"HTTP",
				"SOCKS4"
			],
			"ShareType@Redfish.AllowableValues": [
				"LOCAL",
				"NFS",
				"CIFS",
				"HTTP",
				"HTTPS"
			],
			"Target@Redfish.AllowableValues": [
				"ALL"
			]
		},
		"target": "/redfish/v1/Managers/iDRAC.Embedded.1/Actions/Oem/EID_674_Manager.ImportSystemConfigurationPreview"
	},
	"DellManager.v1_0_0#DellManager.ResetToDefaults": {
		"ResetType@Redfish.AllowableValues": [
			"All",
			"ResetAllWithRootDefaults",
			"Default"
		],
		"target": "/redfish/v1/Managers/iDRAC.Embedded.1/Actions/Oem/DellManager.ResetToDefaults"
	}
}
`
var managerBody = `{
		"@Redfish.Copyright": "Copyright 2014-2019 DMTF. All rights reserved.",
		"@odata.context": "/redfish/v1/$metadata#Manager.Manager",
		"@odata.id": "/redfish/v1/Managers/BMC-1",
		"@odata.type": "#Manager.v1_1_0.Manager",
		"Id": "BMC-1",
		"Name": "Manager",
		"ManagerType": "BMC",
		"Description": "BMC",
		"AutoDSTEnabled": true,
		"ServiceEntryPointUUID": "92384634-2938-2342-8820-489239905423",
		"UUID": "00000000-0000-0000-0000-000000000000",
		"Model": "Joo Janta 200",
		"DateTime": "2015-03-13T04:14:33+06:00",
		"DateTimeLocalOffset": "+06:00",
		"PowerState": "On",
		"Status": {
			"State": "Enabled",
			"Health": "OK"
		},
		"GraphicalConsole": {
			"ServiceEnabled": true,
			"MaxConcurrentSessions": 2,
			"ConnectTypesSupported": [
				"KVMIP"
			]
		},
		"SerialConsole": {
			"ServiceEnabled": true,
			"MaxConcurrentSessions": 1,
			"ConnectTypesSupported": [
				"Telnet",
				"SSH",
				"IPMI"
			]
		},
		"CommandShell": {
			"ServiceEnabled": true,
			"MaxConcurrentSessions": 4,
			"ConnectTypesSupported": [
				"Telnet",
				"SSH"
			]
		},
		"FirmwareVersion": "1.00",
		"RemoteAccountService": {
			"@odata.id": "/redfish/v1/Managers/AccountService"
		},
		"RemoteRedfishServiceUri": "http://example.com/",
		"NetworkProtocol": {
			"@odata.id": "/redfish/v1/Managers/BMC-1/NetworkProtocol"
		},
		"HostInterfaces": {
			"@odata.id": "/redfish/v1/Managers/BMC-1/HostInterfaces"
		},
		"EthernetInterfaces": {
			"@odata.id": "/redfish/v1/Managers/BMC-1/EthernetInterfaces"
		},
		"SerialInterfaces": {
			"@odata.id": "/redfish/v1/Managers/BMC-1/SerialInterfaces"
		},
		"LogServices": {
			"@odata.id": "/redfish/v1/Managers/BMC-1/LogServices"
		},
		"VirtualMedia": {
			"@odata.id": "/redfish/v1/Managers/BMC-1/VM1"
		},
		"Links": {
			"ManagerForServers": [
				{
					"@odata.id": "/redfish/v1/Systems/System-1"
				}
			],
			"ManagerForChassis": [
				{
					"@odata.id": "/redfish/v1/Chassis/Chassis-1"
				}
			],
			"ManagerInChassis": {
				"@odata.id": "/redfish/v1/Chassis/Chassis-1"
			},
			"Oem":
` + oemLinksBody +
	`		},
		"Actions": {
			"#Manager.Reset": {
				"target": "/redfish/v1/Managers/BMC-1/Actions/Manager.Reset",
				"ResetType@Redfish.AllowableValues": [
					"ForceRestart",
					"GracefulRestart"
				]
			},
			"Oem":
` + oemActions +
	`	},
		"Oem":
` + oemDataBody +
	`	}`

func TestDellManager(t *testing.T) {
	// Get gofish manager
	var result redfish.Manager
	err := json.NewDecoder(strings.NewReader(managerBody)).Decode(&result)
	if err != nil {
		t.Fatalf("couldn't decode redfish.Manager mocked json")
	}

	// Get Dell manager
	dellManager, err := Manager(&result)
	if err != nil {
		t.Fatalf("couldn't decode dell.Manager mocked json")
	}

	t.Run("Test Dell OEM actions", func(t *testing.T) {
		assertField(t, dellManager.Actions.ExportSystemConfigurationTarget, "/redfish/v1/Managers/iDRAC.Embedded.1/Actions/Oem/EID_674_Manager.ExportSystemConfiguration")

		assertField(t, dellManager.Actions.ImportSystemConfigurationTarget, "/redfish/v1/Managers/iDRAC.Embedded.1/Actions/Oem/EID_674_Manager.ImportSystemConfiguration")

		assertField(t, dellManager.Actions.ImportSystemConfigurationPreviewTarget, "/redfish/v1/Managers/iDRAC.Embedded.1/Actions/Oem/EID_674_Manager.ImportSystemConfigurationPreview")

		assertField(t, dellManager.Actions.ResetToDefaultsTarget, "/redfish/v1/Managers/iDRAC.Embedded.1/Actions/Oem/DellManager.ResetToDefaults")
	})

	t.Run("Test Dell Links OEM", func(t *testing.T) {
		assertLinkArray(t, dellManager.links.DellAttributes, []string{
			"/redfish/v1/Managers/iDRAC.Embedded.1/Oem/Dell/DellAttributes/iDRAC.Embedded.1",
			"/redfish/v1/Managers/iDRAC.Embedded.1/Oem/Dell/DellAttributes/System.Embedded.1",
			"/redfish/v1/Managers/iDRAC.Embedded.1/Oem/Dell/DellAttributes/LifecycleController.Embedded.1",
		})
		assertLink(t, dellManager.links.DellJobService, "/redfish/v1/Managers/iDRAC.Embedded.1/Oem/Dell/DellJobService")
		assertLink(t, dellManager.links.DellLCService, "/redfish/v1/Managers/iDRAC.Embedded.1/Oem/Dell/DellLCService")
		assertLink(t, dellManager.links.DellLicensableDeviceCollection, "/redfish/v1/Managers/iDRAC.Embedded.1/Oem/Dell/DellLicensableDevices")
		assertLink(t, dellManager.links.DellLicenseCollection, "/redfish/v1/Managers/iDRAC.Embedded.1/Oem/Dell/DellLicenses")
		assertLink(t, dellManager.links.DellLicenseManagementService, "/redfish/v1/Managers/iDRAC.Embedded.1/Oem/Dell/DellLicenseManagementService")
		assertLink(t, dellManager.links.DellOpaqueManagementDataCollection, "/redfish/v1/Managers/iDRAC.Embedded.1/Oem/Dell/DellOpaqueManagementData")
		assertLink(t, dellManager.links.DellPersistentStorageService, "/redfish/v1/Managers/iDRAC.Embedded.1/Oem/Dell/DellPersistentStorageService")
		assertLink(t, dellManager.links.DellSwitchConnectionCollection, "/redfish/v1/Systems/System.Embedded.1/NetworkPorts/Oem/Dell/DellSwitchConnections")
		assertLink(t, dellManager.links.DellSwitchConnectionService, "/redfish/v1/Systems/System.Embedded.1/Oem/Dell/DellSwitchConnectionService")
		assertLink(t, dellManager.links.DellSystemManagementService, "/redfish/v1/Systems/System.Embedded.1/Oem/Dell/DellSystemManagementService")
		assertLink(t, dellManager.links.DellSystemQuickSyncCollection, "/redfish/v1/Managers/iDRAC.Embedded.1/Oem/Dell/DellSystemQuickSync")
		assertLink(t, dellManager.links.DellTimeService, "/redfish/v1/Managers/iDRAC.Embedded.1/Oem/Dell/DellTimeService")
		assertLink(t, dellManager.links.DellUSBDeviceCollection, "/redfish/v1/Managers/iDRAC.Embedded.1/Oem/Dell/DellUSBDevices")
		assertLink(t, dellManager.links.DelliDRACCardService, "/redfish/v1/Managers/iDRAC.Embedded.1/Oem/Dell/DelliDRACCardService")
		assertLink(t, dellManager.links.DellvFlashCollection, "/redfish/v1/Managers/iDRAC.Embedded.1/Oem/Dell/DellvFlash")
		assertLink(t, dellManager.links.Jobs, "/redfish/v1/Managers/iDRAC.Embedded.1/Oem/Dell/Jobs")
	})

	t.Run("Test Dell OEM field", func(t *testing.T) {
		assertField(t, dellManager.OemData.DelliDRACCard.ODataContext, "/redfish/v1/$metadata#DelliDRACCard.DelliDRACCard")
		assertField(t, dellManager.OemData.DelliDRACCard.ODataID, "/redfish/v1/Managers/iDRAC.Embedded.1/Oem/Dell/DelliDRACCard/iDRAC.Embedded.1-1_0x23_IDRACinfo")
		assertField(t, dellManager.OemData.DelliDRACCard.ODataType, "#DelliDRACCard.v1_1_0.DelliDRACCard")
		assertField(t, dellManager.OemData.DelliDRACCard.Description, "An instance of DelliDRACCard will have data specific to the Integrated Dell Remote Access Controller (iDRAC) in the managed system.")
		assertField(t, dellManager.OemData.DelliDRACCard.IPMIVersion, "2.0")
		assertField(t, dellManager.OemData.DelliDRACCard.ID, "iDRAC.Embedded.1-1_0x23_IDRACinfo")
		assertField(t, dellManager.OemData.DelliDRACCard.LastSystemInventoryTime, "2021-06-08T09:12:53+00:00")
		assertField(t, dellManager.OemData.DelliDRACCard.LastUpdateTime, "2021-06-08T14:44:15+00:00")
		assertField(t, dellManager.OemData.DelliDRACCard.Name, "DelliDRACCard")
		assertField(t, dellManager.OemData.DelliDRACCard.URLString, "https://10.0.41.190:443")
	})
}
