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
)

var sampleJSONAttributeRegistry = `
{
    "@odata.context": "/redfish/v1/$metadata#DellAttributeRegistry.DellAttributeRegistry",
    "@odata.id": "/redfish/v1/Registries/ManagerAttributeRegistry/ManagerAttributeRegistry.v1_0_0.json",
    "@odata.type": "#DellAttributeRegistry.v1_1_0.DellAttributeRegistry",
    "Description": "This registry defines a representation of OEM Attribute instances",
    "Id": "OEMAttributeRegistry",
    "Language": "en",
    "Name": "AttributeRegistry",
    "OwningEntity": "Dell",
    "RegistryEntries": {
        "Attributes": [
            {
                "AttributeName": "LCAttributes.1.AutoBackup",
                "CurrentValue": null,
                "DefaultValue": "0",
                "DisplayName": "Automatic Backup Feature",
                "DisplayOrder": 23,
                "HelpText": "Enables or disables the automatic backup scheduler.",
                "Hidden": false,
                "Id": "LifecycleController.Embedded.1#LCAttributes.1#AutoBackup",
                "MenuPath": "./LifecycleController.Embedded.1/LCAttributes",
                "Readonly": false,
                "Regex": "",
                "Type": "Enumeration",
                "Value": [
                    {
                        "ValueDisplayName": "Disabled",
                        "ValueName": "0"
                    },
                    {
                        "ValueDisplayName": "Enabled",
                        "ValueName": "1"
                    }
                ],
                "WriteOnly": false
            },
            {
                "AttributeName": "OpenIDConnectServer.12.RegistrationDetails",
                "CurrentValue": null,
                "DefaultValue": "",
                "DisplayName": "Credentials needed to register on server",
                "DisplayOrder": 3,
                "HelpText": "Registration details",
                "Hidden": false,
                "Id": "System.Embedded.1#OpenIDConnectServer.12#RegistrationDetails",
                "MaxLength": 1024,
                "MenuPath": "./System.Embedded.1/OpenIDConnectServer",
                "MinLength": 0,
                "Readonly": false,
                "Regex": "",
                "Type": "Password",
                "WriteOnly": true
            },
            {
                "AttributeName": "OpenIDConnectServer.12.Name",
                "CurrentValue": null,
                "DefaultValue": "",
                "DisplayName": "Server Name",
                "DisplayOrder": 7,
                "HelpText": "Server name",
                "Hidden": false,
                "Id": "System.Embedded.1#OpenIDConnectServer.12#Name",
                "MaxLength": 128,
                "MenuPath": "./System.Embedded.1/OpenIDConnectServer",
                "MinLength": 0,
                "Readonly": false,
                "Regex": "",
                "Type": "String",
                "WriteOnly": false
            },
			{
                "AttributeName": "LCD.1.ChassisIdentifyDuration",
                "CurrentValue": null,
                "DefaultValue": 0,
                "DisplayName": "Chassis Identify Duration",
                "DisplayOrder": 1438,
                "HelpText": "Enable/Disable chassis Identify. A value of -1 = force on indefinitely; 0 = 0ff; &gt; 0 number of seconds chassis identify is on",
                "Hidden": false,
                "Id": "System.Embedded.1#LCD.1#ChassisIdentifyDuration",
                "LowerBound": -1,
                "MenuPath": "./System.Embedded.1/LCD",
                "Readonly": false,
                "Regex": "",
                "Type": "Integer",
                "UpperBound": 2592000,
                "WriteOnly": false
            },
            {
                "AttributeName": "PCIeSlotLFM.3.MaxLFM",
                "CurrentValue": null,
                "DefaultValue": 0,
                "DisplayName": "Maximum LFM",
                "DisplayOrder": 5,
                "HelpText": "Estimated airflow delivered to the slot at full fan speed",
                "Hidden": false,
                "Id": "System.Embedded.1#PCIeSlotLFM.3#MaxLFM",
                "LowerBound": 0,
                "MenuPath": "./System.Embedded.1/PCIeSlotLFM",
                "Readonly": true,
                "Regex": "",
                "Type": "Integer",
                "UpperBound": 65536,
                "WriteOnly": false
            }
        ],
        "Dependencies": [
            {
                "Dependency": {
                    "MapFrom": [
                        {
                            "MapFromAttribute": "QuickSync.n.InactivityTimerEnable",
                            "MapFromCondition": "EQU",
                            "MapFromProperty": "CurrentValue",
                            "MapFromValue": "Enabled"
                        }
                    ],
                    "MapToAttribute": "QuickSync.n.InactivityTimeout",
                    "MapToProperty": "Readonly",
                    "MapToValue": true
                },
                "DependencyFor": "QuickSync.n.InactivityTimeout",
                "Type": "Map"
            },
            {
                "Dependency": {
                    "MapFrom": [
                        {
                            "MapFromAttribute": "System.Embedded.n#ThermalSettings.n.AirExhaustTempSupport",
                            "MapFromCondition": "EQU",
                            "MapFromProperty": "CurrentValue",
                            "MapFromValue": "Supported"
                        }
                    ],
                    "MapToAttribute": "ThermalSettings.n.AirExhaustTemp",
                    "MapToProperty": "Readonly",
                    "MapToValue": true
                },
                "DependencyFor": "ThermalSettings.n.AirExhaustTemp",
                "Type": "Map"
            }
        ],
        "Menus": [
            {
                "DisplayName": "LifecycleController.Embedded.1",
                "DisplayOrder": 1,
                "Hidden": false,
                "MenuName": "LifecycleController.Embedded.1",
                "MenuPath": "./LifecycleController.Embedded.1",
                "Readonly": false
            },
            {
                "DisplayName": "LCAttributes",
                "DisplayOrder": 1,
                "Hidden": false,
                "MenuName": "LCAttributes",
                "MenuPath": "./LifecycleController.Embedded.1/LCAttributes",
                "Readonly": false
            }
        ]
    },
    "RegistryPrefix": "iDRAC",
    "RegistryVersion": "1.0.0",
    "SupportedSystems": [
        {
            "FirmwareVersion": "4.40.10.00",
            "ProductName": "Integrated Dell Remote Access Controller",
            "SystemId": "14G"
        }
    ]
}
`

func TestDellAttributesRegistry(t *testing.T) {
	var registry ManagerAttributeRegistry
	err := json.NewDecoder(strings.NewReader(sampleJSONAttributeRegistry)).Decode(&registry)
	if err != nil {
		t.Error("could not decode ManagerAttributeRegistry JSON")
	}

	t.Run("Check attributes are parsed accordingly", func(t *testing.T) {
		assertField(t, registry.ID, "OEMAttributeRegistry")
		assertField(t, registry.Name, "AttributeRegistry")
		assertField(t, registry.OwningEntity, "Dell")
		assertField(t, registry.RegistryPrefix, "iDRAC")

		// Check Attributes (0 -> Enumeration, 1 -> Password, 2 -> String, 3 -> Integer, 4 -> Integer(readonly))
		assertField(t, registry.Attributes[0].AttributeName, "LCAttributes.1.AutoBackup")
		assertField(t, registry.Attributes[0].Type, "Enumeration")
		assertBool(t, registry.Attributes[0].Readonly, false)
		assertBool(t, registry.Attributes[0].WriteOnly, false)
		assertField(t, registry.Attributes[0].Value[0].ValueDisplayName, "Disabled")
		assertField(t, registry.Attributes[0].Value[1].ValueDisplayName, "Enabled")

		assertField(t, registry.Attributes[1].AttributeName, "OpenIDConnectServer.12.RegistrationDetails")
		assertField(t, registry.Attributes[1].Type, "Password")
		assertBool(t, registry.Attributes[1].Readonly, false)
		assertBool(t, registry.Attributes[1].WriteOnly, true)
		assertInt(t, registry.Attributes[1].MinLength, 0)
		assertInt(t, registry.Attributes[1].MaxLength, 1024)

		assertField(t, registry.Attributes[2].AttributeName, "OpenIDConnectServer.12.Name")
		assertField(t, registry.Attributes[2].Type, "String")
		assertBool(t, registry.Attributes[2].Readonly, false)
		assertBool(t, registry.Attributes[2].WriteOnly, false)
		assertInt(t, registry.Attributes[2].MinLength, 0)
		assertInt(t, registry.Attributes[2].MaxLength, 128)

		assertField(t, registry.Attributes[3].AttributeName, "LCD.1.ChassisIdentifyDuration")
		assertField(t, registry.Attributes[3].Type, "Integer")
		assertBool(t, registry.Attributes[3].Readonly, false)
		assertBool(t, registry.Attributes[3].WriteOnly, false)
		assertInt(t, registry.Attributes[3].LowerBound, -1)
		assertInt(t, registry.Attributes[3].UpperBound, 2592000)
	})

	t.Run("Test CheckAttribute method", func(t *testing.T) {
		assertCheckAttribute(t, true, registry.CheckAttribute("LCAttributes.1.AutoBackup", 0.0))         // Float, must fail
		assertCheckAttribute(t, true, registry.CheckAttribute("LCAttributes.1.AutoBackup", 1))           // Int, type is Enum, must fail
		assertCheckAttribute(t, false, registry.CheckAttribute("LCAttributes.1.AutoBackup", "Disabled")) // String within enum, must pass
		assertCheckAttribute(t, false, registry.CheckAttribute("LCAttributes.1.AutoBackup", "Enabled"))  // String within enum ,must pass
		assertCheckAttribute(t, true, registry.CheckAttribute("LCAttributes.1.AutoBackup", "Madeup"))    // String out of enum, must fail

		assertCheckAttribute(t, true, registry.CheckAttribute("OpenIDConnectServer.12.RegistrationDetails", 45))                        // Int must fail
		assertCheckAttribute(t, false, registry.CheckAttribute("OpenIDConnectServer.12.RegistrationDetails", "32141ff"))                // String complian with password, must pass
		assertCheckAttribute(t, true, registry.CheckAttribute("OpenIDConnectServer.12.RegistrationDetails", strings.Repeat("a", 1025))) // String non compliant with length, must fail

		assertCheckAttribute(t, true, registry.CheckAttribute("OpenIDConnectServer.12.Name", 45))                       // Int must fail
		assertCheckAttribute(t, false, registry.CheckAttribute("OpenIDConnectServer.12.Name", "32141ff"))               // String complian with password, must pass
		assertCheckAttribute(t, true, registry.CheckAttribute("OpenIDConnectServer.12.Name", strings.Repeat("a", 129))) // String non compliant with length, must fail

		assertCheckAttribute(t, true, registry.CheckAttribute("LCD.1.ChassisIdentifyDuration", "test"))   // String must fail
		assertCheckAttribute(t, false, registry.CheckAttribute("LCD.1.ChassisIdentifyDuration", -1))      // Int compliant with bounds, must pass
		assertCheckAttribute(t, false, registry.CheckAttribute("LCD.1.ChassisIdentifyDuration", 2592000)) // Int compliant with bounds, must pass
		assertCheckAttribute(t, true, registry.CheckAttribute("LCD.1.ChassisIdentifyDuration", 2592001))  // Int not compliant with bounds, must fail
		assertCheckAttribute(t, true, registry.CheckAttribute("LCD.1.ChassisIdentifyDuration", -2))       // Int not compliant with bounds, must fail

		assertCheckAttribute(t, true, registry.CheckAttribute("PCIeSlotLFM.3.MaxLFM", "test")) // String must fail
		assertCheckAttribute(t, true, registry.CheckAttribute("PCIeSlotLFM.3.MaxLFM", 2))      // Int compliant but property readonly, must fail

		assertCheckAttribute(t, true, registry.CheckAttribute("non.existent.property", "test")) // property doesn't exist. Must fail
	})

	t.Run("Test GetAttributeType func", func(t *testing.T) {
		assertGetAttributeType(t, &registry, "LCAttributes.1.AutoBackup", "string")
		assertGetAttributeType(t, &registry, "OpenIDConnectServer.12.RegistrationDetails", "string")
		assertGetAttributeType(t, &registry, "OpenIDConnectServer.12.Name", "string")
		assertGetAttributeType(t, &registry, "LCD.1.ChassisIdentifyDuration", "int")
		assertGetAttributeType(t, &registry, "PCIeSlotLFM.3.MaxLFM", "int")
	})
}

func assertCheckAttribute(t testing.TB, hasError bool, err error) {
	t.Helper()
	if hasError && err == nil {
		t.Errorf("expected to have an error but no error was returned")
	}
	if !hasError && err != nil {
		t.Errorf("not expected to return an error but it returned one - %s", err)
	}
	if err != nil {
		t.Logf("[INFO] error returned was %s", err)
	}
}

func assertGetAttributeType(t testing.TB, registry *ManagerAttributeRegistry, attrName, want string) {
	attrType, err := registry.GetAttributeType(attrName)
	if err != nil {
		t.Errorf("error. Couldn't determine the type.")
	}
	assertField(t, attrType, want)
}
