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

var simpleUpdateBody = `{
    "@odata.context": "/redfish/v1/$metadata#UpdateService.UpdateService",
    "@odata.id": "/redfish/v1/UpdateService",
    "@odata.type": "#UpdateService.v1_6_0.UpdateService",
    "Actions": {
        "#UpdateService.SimpleUpdate": {
            "TransferProtocol@Redfish.AllowableValues": [
                "HTTP"
            ],
            "target": "/redfish/v1/UpdateService/Actions/UpdateService.SimpleUpdate"
        },
        "Oem": {
            "DellUpdateService.v1_0_0#DellUpdateService.Install": {
                "InstallUpon@Redfish.AllowableValues": [
                    "Now",
                    "NowAndReboot",
                    "NextReboot"
                ],
                "target": "/redfish/v1/UpdateService/Actions/Oem/DellUpdateService.Install"
            }
        }
    },
    "Description": "Represents the properties for the Update Service",
    "FirmwareInventory": {
        "@odata.id": "/redfish/v1/UpdateService/FirmwareInventory"
    },
    "HttpPushUri": "/redfish/v1/UpdateService/FirmwareInventory",
    "Id": "UpdateService",
    "Name": "Update Service",
    "ServiceEnabled": true,
    "Status": {
        "Health": "OK",
        "State": "Enabled"
    }
}`

func TestDellUpdateService(t *testing.T) {
	t.Run("Test redfish values", func(t *testing.T) {
		dellUpdateService := getDellUpdateService(t)

		assertField(t, dellUpdateService.UpdateServiceTarget, "/redfish/v1/UpdateService/Actions/UpdateService.SimpleUpdate")
		assertField(t, dellUpdateService.HTTPPushURI, "/redfish/v1/UpdateService/FirmwareInventory")
		assertArray(t, dellUpdateService.TransferProtocol, []string{"HTTP"})
	})
	t.Run("Check Dell values", func(t *testing.T) {
		dellUpdateService := getDellUpdateService(t)

		assertField(t, dellUpdateService.Actions.DellUpdateServiceTarget, "/redfish/v1/UpdateService/Actions/Oem/DellUpdateService.Install")
		assertArray(t, dellUpdateService.Actions.DellUpdateServiceInstallUpon, []string{"Now", "NowAndReboot", "NextReboot"})
	})
}

func getDellUpdateService(t testing.TB) *UpdateServiceExtended {
	t.Helper()

	var result redfish.UpdateService

	err := json.NewDecoder(strings.NewReader(simpleUpdateBody)).Decode(&result)
	if err != nil {
		t.Errorf("Error decoding simpleUpdate JSON - %s", err)
	}

	dellUpdateService, err := UpdateService(&result)
	if err != nil {
		t.Errorf("Error decoding Dell simpleUpdate JSON - %s", err)
	}

	return dellUpdateService
}
