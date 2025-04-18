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

package provider

import (
	"fmt"
	"os"
	"regexp"
	"terraform-provider-redfish/gofish/dell"
	"testing"

	"github.com/bytedance/mockey"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stmcginnis/gofish/redfish"
)

var nicParams, fcParams testingNICInputs

type testingNICInputs struct {
	TestingServerCredentials
	NetworkDeviceFunctionID string
	NetworkAdapterID        string
	SystemID                string
}

func init() {
	nicParams = testingNICInputs{
		TestingServerCredentials: creds,
		NetworkDeviceFunctionID:  os.Getenv("NETWORK_DEVICE_FUNCTION_ID_1"),
		NetworkAdapterID:         os.Getenv("NETWORK_ADAPTER_ID_1"),
		SystemID:                 "System.Embedded.1",
	}

	fcParams = testingNICInputs{
		TestingServerCredentials: creds,
		NetworkDeviceFunctionID:  os.Getenv("NETWORK_DEVICE_FUNCTION_ID_2"),
		NetworkAdapterID:         os.Getenv("NETWORK_ADAPTER_ID_2"),
		SystemID:                 "System.Embedded.1",
	}
}
func TestAccRedfishNICAttributesBasic(t *testing.T) {
	terraformResourceName := "redfish_network_adapter.nic"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// error create with both `network_attributes` and `oem_network_attributes`
			{
				Config:      testAccRedfishResourceNICAttributesConfig(nicParams),
				ExpectError: regexp.MustCompile("Error when creating both of `network_attributes` and `oem_network_attributes`"),
			},
			// create with `oem_network_attributes` only
			{
				Config: testAccRedfishResourceNICAttributesConfigNetworkAttrs(nicParams),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(terraformResourceName, "network_attributes.ethernet.mac_address", "E4:43:4B:17:E0:A9"),
					resource.TestCheckResourceAttr(terraformResourceName, "network_attributes.ethernet.mtu_size", "100"),
					resource.TestCheckResourceAttr(terraformResourceName, "network_attributes.ethernet.vlan.vlan_id", "100")),
			},
			// error update ids
			{
				Config:      testAccRedfishResourceNICAttributesConfigUpdateNetAttrs(fcParams),
				ExpectError: regexp.MustCompile("Error when updating with invalid input"),
			},
			// add `network_attributes`
			{
				Config: testAccRedfishResourceNICAttributesConfig(nicParams),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(terraformResourceName, "oem_network_attributes.attributes.WakeOnLan", "Disabled"),
				),
			},
			// update `oem_network_attributes`
			{
				Config: testAccRedfishResourceNICAttributesConfigUpdateNetAttrs(nicParams),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(terraformResourceName, "network_attributes.ethernet.mac_address", "E4:43:4B:17:E0:A0"),
					resource.TestCheckResourceAttr(terraformResourceName, "network_attributes.ethernet.mtu_size", "1000"),
					resource.TestCheckResourceAttr(terraformResourceName, "network_attributes.ethernet.mtu_size", "1000"),
				),
			},
		},
	})
}

func TestAccRedfishNICAttributesFC(t *testing.T) {
	terraformResourceName := "redfish_network_adapter.nic"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// error create without `maintenance_window` when `apply_time` is `AtMaintenanceWindowStart`
			{
				Config:      testAccRedfishResourceFCConfigWithoutMW(fcParams),
				ExpectError: regexp.MustCompile("Input param is not valid"),
			},
			// error create with outdated `maintenance_window`
			{
				Config:      testAccRedfishResourceFCConfigOutDatedMW(fcParams),
				ExpectError: regexp.MustCompile("there was an issue when creating/updating network attributes"),
			},
			// create with `network_attributes` only for FC
			{
				Config: testAccRedfishResourceFCConfigNetworkAttrs(fcParams),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(terraformResourceName, "network_attributes.fibre_channel.wwnn", "20:00:F4:E9:D4:56:10:AB"),
					resource.TestCheckResourceAttr(terraformResourceName, "network_attributes.fibre_channel.boot_targets.0.lun_id", "2"),
				),
			},
			// error update `oem_network_attributes` for FC with outdated `maintenance_window`
			{
				Config:      testAccRedfishResourceFCConfigUpdateOutDatedMW(fcParams),
				ExpectError: regexp.MustCompile("there was an issue when creating/updating ome network attributes"),
			},
			// add `oem_network_attributes` for FC
			{
				Config: testAccRedfishResourceFCConfig(fcParams),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(terraformResourceName, "network_attributes.fibre_channel.wwnn", "20:00:F4:E9:D4:56:10:AB"),
					resource.TestCheckResourceAttr(terraformResourceName, "network_attributes.fibre_channel.boot_targets.0.lun_id", "2"),
					resource.TestCheckResourceAttr(terraformResourceName, "oem_network_attributes.attributes.PortLoginTimeout", "4000"),
				),
			},
			// update `network_attributes` for FC
			{
				Config: testAccRedfishResourceFCConfigUpdateNetAttrs(fcParams),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(terraformResourceName, "network_attributes.fibre_channel.wwnn", "20:00:F4:E9:D4:56:10:CD"),
					resource.TestCheckResourceAttr(terraformResourceName, "network_attributes.fibre_channel.boot_targets.0.lun_id", "1"),
					resource.TestCheckResourceAttr(terraformResourceName, "oem_network_attributes.attributes.PortLoginTimeout", "4000"),
				),
			},
		},
	})
}

func TestAccRedfishNICAttributesISCSI(t *testing.T) {
	terraformResourceName := "redfish_network_adapter.nic"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// create with `network_attributes` only for ISCSI
			{
				Config: testAccRedfishResourceNICAttributesIscsiConfig(nicParams),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(terraformResourceName, "network_attributes.iscsi_boot.authentication_method", "CHAP"),
				),
			},
			// update `network_attributes` for ISCSI
			{
				Config: testAccRedfishResourceNICAttributesIscsiConfigUpdate(nicParams),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(terraformResourceName, "network_attributes.iscsi_boot.authentication_method", "None"),
				),
			},
		},
	})
}

func TestAccRedfishNICAttributesImport(t *testing.T) {
	importReqID := fmt.Sprintf("{\"system_id\":\"%s\",\"network_adapter_id\":\"%s\",\"network_device_function_id\":\"%s\",\"username\":\"%s\",\"password\":\"%s\",\"endpoint\":\"%s\",\"ssl_insecure\":true}",
		nicParams.SystemID, nicParams.NetworkAdapterID, nicParams.NetworkDeviceFunctionID, nicParams.Username, nicParams.Password, nicParams.Endpoint)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `resource "redfish_network_adapter" "nic" {
				}`,
				ResourceName:  "redfish_network_adapter.nic",
				ImportState:   true,
				ImportStateId: importReqID,
				ExpectError:   nil,
			},
		},
	})
}

func TestAccRedfishNICAttributes_CreateMockErr(t *testing.T) {
	var funcMocker1 *mockey.Mocker
	settObj := dell.Entity{
		ODataID: "",
	}
	dellRes := &dell.NetworkDeviceFunctionExtended{
		SettingsObject: settObj,
	}
	settObj1 := dell.Entity{
		ODataID: "testOdataId",
	}
	dellRes1 := &dell.NetworkDeviceFunctionExtended{
		SettingsObject: settObj1,
	}

	fibreChannel := redfish.FibreChannel{
		PermanentWWNN: "",
		PermanentWWPN: "",
	}
	etherNet := redfish.Ethernet{
		PermanentMACAddress: "",
	}
	boot := redfish.ISCSIBoot{
		AuthenticationMethod: "",
		IPAddressType:        "",
	}
	netDevFunc := &redfish.NetworkDeviceFunction{
		FibreChannel: fibreChannel,
		Ethernet:     etherNet,
		ISCSIBoot:    boot,
	}
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// create with `network_attributes` only for FC with newcofig mock error
			{
				PreConfig: func() {
					FunctionMocker = mockey.Mock(NewConfig).Return(nil, fmt.Errorf("mock error")).Build()
				},
				Config:      testAccRedfishResourceFCConfigNetworkAttrs(fcParams),
				ExpectError: regexp.MustCompile(`.*mock error*.`),
			},
			//  create with `network_attributes` only for FC with getSystemResource mock error
			{
				PreConfig: func() {
					if FunctionMocker != nil {
						FunctionMocker.Release()
					}
					FunctionMocker = mockey.Mock(getSystemResource).Return(nil, fmt.Errorf("mock error")).Build()
				},
				Config:      testAccRedfishResourceFCConfigNetworkAttrs(fcParams),
				ExpectError: regexp.MustCompile(`.*mock error*.`),
			},
			{
				PreConfig: func() {
					if FunctionMocker != nil {
						FunctionMocker.Release()
					}
					FunctionMocker = mockey.Mock(getNetworkDeviceFunction).Return(nil, nil, fmt.Errorf("mock error")).Build()
				},
				Config:      testAccRedfishResourceFCConfigNetworkAttrs(fcParams),
				ExpectError: regexp.MustCompile(`.*mock error*.`),
			},
			{
				PreConfig: func() {
					if FunctionMocker != nil {
						FunctionMocker.Release()
					}
					FunctionMocker = mockey.Mock(dell.NetworkDeviceFunction).Return(dellRes, nil).Build()
				},
				Config:      testAccRedfishResourceFCConfigNetworkAttrs(fcParams),
				ExpectError: regexp.MustCompile(`.*there was an issue when creating/updating network attributes: *.`),
			},
			{
				PreConfig: func() {
					if FunctionMocker != nil {
						FunctionMocker.Release()
					}
					funcMocker1 = mockey.Mock(dell.NetworkDeviceFunction).Return(dellRes1, nil).Build()
					FunctionMocker = mockey.Mock(getNetworkDeviceFunction).Return(nil, netDevFunc, nil).Build()
				},
				Config:      testAccRedfishResourceFCConfigNetworkAttrs(fcParams),
				ExpectError: regexp.MustCompile(`.*there was an issue when creating/updating network attributes: *.`),
			},
			{
				PreConfig: func() {
					if FunctionMocker != nil {
						FunctionMocker.Release()
					}
					FunctionMocker = mockey.Mock(getNetworkDeviceFunction).Return(nil, netDevFunc, nil).Build()
				},
				Config:      testAccRedfishResourceNICAttributesConfigNetworkAttrs(fcParams),
				ExpectError: regexp.MustCompile(`.*there was an issue when creating/updating network attributes: *.`),
			},
			{
				PreConfig: func() {
					if FunctionMocker != nil {
						FunctionMocker.Release()
					}
					FunctionMocker = mockey.Mock(getNetworkDeviceFunction).Return(nil, netDevFunc, nil).Build()
				},
				Config:      testAccRedfishResourceNICAttributesIscsiConfig(fcParams),
				ExpectError: regexp.MustCompile(`.*there was an issue when creating/updating network attributes: *.`),
			},
		},
	})
	if funcMocker1 != nil {
		funcMocker1.Release()
	}
	if FunctionMocker != nil {
		FunctionMocker.Release()
	}

}

func TestAccRedfishNICAttributesOME_CreateMockErr(t *testing.T) {
	settObj := dell.Entity{
		ODataID: "",
	}
	dellRes := &dell.NetworkDeviceFunctionExtended{
		SettingsObject: settObj,
	}
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// create with `network_attributes` only for FC with newcofig mock error
			{
				PreConfig: func() {
					if FunctionMocker != nil {
						FunctionMocker.Release()
					}
					FunctionMocker = mockey.Mock(getNetworkDeviceFunction).Return(nil, nil, fmt.Errorf("mock error")).Build()
				},
				Config:      testAccRedfishResourceFCConfig_Mocky(fcParams),
				ExpectError: regexp.MustCompile(`.*mock error*.`),
			},
			{
				PreConfig: func() {
					if FunctionMocker != nil {
						FunctionMocker.Release()
					}
					FunctionMocker = mockey.Mock(dell.NetworkDeviceFunction).Return(dellRes, nil).Build()
				},
				Config:      testAccRedfishResourceFCConfig_Mocky(fcParams),
				ExpectError: regexp.MustCompile(`.*there was an issue when creating/updating ome network attributes*.`),
			},
		},
	})
	if FunctionMocker != nil {
		FunctionMocker.Release()
	}
}

func TestAccRedfishNICAttributes_UpdateMockErr(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRedfishResourceFCConfigNetworkAttrs(fcParams),
			},
			// Update with `network_attributes` only for FC with newcofig mock error
			{
				PreConfig: func() {
					FunctionMocker = mockey.Mock(NewConfig).Return(nil, fmt.Errorf("mock error")).Build()
				},
				Config:      testAccRedfishResourceFCConfig(fcParams),
				ExpectError: regexp.MustCompile(`.*mock error*.`),
			},
			//  Update with `network_attributes` only for FC with getSystemResource mock error
			{
				PreConfig: func() {
					if FunctionMocker != nil {
						FunctionMocker.Release()
					}
					FunctionMocker = mockey.Mock(getSystemResource).Return(nil, fmt.Errorf("mock error")).Build()
				},
				Config:      testAccRedfishResourceFCConfig(fcParams),
				ExpectError: regexp.MustCompile(`.*mock error*.`),
			},
		},
	})
	if FunctionMocker != nil {
		FunctionMocker.Release()
	}
}

func TestAccRedfishNICAttributes_ReadMockErr(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRedfishResourceFCConfigNetworkAttrs(fcParams),
			},
			// Read `network_attributes` only for FC with newcofig mock error
			{
				PreConfig: func() {
					FunctionMocker = mockey.Mock(NewConfig).Return(nil, fmt.Errorf("mock error")).Build()
				},
				Config:      testAccRedfishResourceFCConfigNetworkAttrs(fcParams),
				ExpectError: regexp.MustCompile(`.*mock error*.`),
			},
			//  Read with `network_attributes` only for FC with getSystemResource mock error
			{
				PreConfig: func() {
					if FunctionMocker != nil {
						FunctionMocker.Release()
					}
					FunctionMocker = mockey.Mock(getSystemResource).Return(nil, fmt.Errorf("mock error")).Build()
				},
				Config:      testAccRedfishResourceFCConfigNetworkAttrs(fcParams),
				ExpectError: regexp.MustCompile(`.*mock error*.`),
			},
		},
	})
	if FunctionMocker != nil {
		FunctionMocker.Release()
	}
}

func testAccRedfishResourceNICAttributesConfigNetworkAttrs(testingInfo testingNICInputs) string {
	return fmt.Sprintf(`
	resource "redfish_network_adapter" "nic" {
	  redfish_server {
		user         = "%s"
		password     = "%s"
		endpoint     = "%s"
		ssl_insecure = true
	  }
	  system_id = "%s"
	  network_adapter_id         = "%s"
	  network_device_function_id = "%s"
	  apply_time = "OnReset"
	  job_timeout = 1200

	  network_attributes = {
		ethernet = {
			mac_address = "E4:43:4B:17:E0:A9"
			mtu_size    = 100
			vlan = {
				vlan_id      = 100
				vlan_enabled = true
			}
		}
	  }
	}
	  `,
		testingInfo.Username,
		testingInfo.PasswordNIC,
		testingInfo.EndpointNIC,
		testingInfo.SystemID,
		testingInfo.NetworkAdapterID,
		testingInfo.NetworkDeviceFunctionID,
	)
}

func testAccRedfishResourceNICAttributesConfig(testingInfo testingNICInputs) string {
	return fmt.Sprintf(`
	resource "redfish_network_adapter" "nic" {
	  redfish_server {
		user         = "%s"
		password     = "%s"
		endpoint     = "%s"
		ssl_insecure = true
	  }
	  system_id = "%s"
	  network_adapter_id         = "%s"
	  network_device_function_id = "%s"
	  apply_time = "OnReset"
	  job_timeout = 1200

	  network_attributes = {
		ethernet = {
			mac_address = "E4:43:4B:17:E0:A9"
			mtu_size    = 100
			vlan = {
				vlan_id      = 100
				vlan_enabled = true
			}
		}
	  }

	  oem_network_attributes = {
	  	clear_pending = false
	  	attributes = {
	  		"WakeOnLan" = "Disabled"
	  	}
  	  }
	}
	  `,
		testingInfo.Username,
		testingInfo.PasswordNIC,
		testingInfo.EndpointNIC,
		testingInfo.SystemID,
		testingInfo.NetworkAdapterID,
		testingInfo.NetworkDeviceFunctionID,
	)
}

func testAccRedfishResourceNICAttributesConfigUpdateNetAttrs(testingInfo testingNICInputs) string {
	return fmt.Sprintf(`
	resource "redfish_network_adapter" "nic" {
	  redfish_server {
		user         = "%s"
		password     = "%s"
		endpoint     = "%s"
		ssl_insecure = true
	  }
	  system_id = "%s"
	  network_adapter_id         = "%s"
	  network_device_function_id = "%s"
	  apply_time = "OnReset"
	  job_timeout = 1200

	  network_attributes = {
		ethernet = {
			mac_address = "E4:43:4B:17:E0:A0"
			mtu_size    = 1000
			vlan = {
				vlan_id      = 1000
				vlan_enabled = true
			}
		}
	  }

	  oem_network_attributes = {
	  	clear_pending = false
	  	attributes = {
	  		"WakeOnLan" = "Disabled"
	  	}
  	  }
	}
	  `,
		testingInfo.Username,
		testingInfo.PasswordNIC,
		testingInfo.EndpointNIC,
		testingInfo.SystemID,
		testingInfo.NetworkAdapterID,
		testingInfo.NetworkDeviceFunctionID,
	)
}

func testAccRedfishResourceFCConfigWithoutMW(testingInfo testingNICInputs) string {
	return fmt.Sprintf(`
	resource "redfish_network_adapter" "nic" {
	  redfish_server {
		user         = "%s"
		password     = "%s"
		endpoint     = "%s"
		ssl_insecure = true
	  }
	  system_id = "%s"
	  network_adapter_id         = "%s"
	  network_device_function_id = "%s"
	  apply_time = "AtMaintenanceWindowStart"
	  job_timeout = 1200

	  network_attributes = {
    	fibre_channel = {
      	  wwnn    = "20:00:F4:E9:D4:56:10:AB"
      	  boot_targets = [
        	{
          		lun_id        = "2"
        	}
      	  ]
    	}
	  }
	}
	  `,
		testingInfo.Username,
		testingInfo.PasswordNIC,
		testingInfo.EndpointNIC,
		testingInfo.SystemID,
		testingInfo.NetworkAdapterID,
		testingInfo.NetworkDeviceFunctionID,
	)
}

func testAccRedfishResourceFCConfigOutDatedMW(testingInfo testingNICInputs) string {
	return fmt.Sprintf(`
	resource "redfish_network_adapter" "nic" {
	  redfish_server {
		user         = "%s"
		password     = "%s"
		endpoint     = "%s"
		ssl_insecure = true
	  }
	  system_id = "%s"
	  network_adapter_id         = "%s"
	  network_device_function_id = "%s"
	  apply_time = "AtMaintenanceWindowStart"
	  job_timeout = 1200
	  maintenance_window = {
	    start_time = "2024-06-30T05:15:40-05:00"
		duration = 600
	  }

	  network_attributes = {
    	fibre_channel = {
      	  wwnn    = "20:00:F4:E9:D4:56:10:AB"
      	  boot_targets = [
        	{
          		lun_id        = "2"
        	}
      	  ]
    	}
	  }
	}
	  `,
		testingInfo.Username,
		testingInfo.PasswordNIC,
		testingInfo.EndpointNIC,
		testingInfo.SystemID,
		testingInfo.NetworkAdapterID,
		testingInfo.NetworkDeviceFunctionID,
	)
}

func testAccRedfishResourceFCConfigNetworkAttrs(testingInfo testingNICInputs) string {
	return fmt.Sprintf(`
	resource "redfish_network_adapter" "nic" {
	  redfish_server {
		user         = "%s"
		password     = "%s"
		endpoint     = "%s"
		ssl_insecure = true
	  }
	  system_id = "%s"
	  network_adapter_id         = "%s"
	  network_device_function_id = "%s"
	  apply_time = "OnReset"
	  job_timeout = 1200

	  network_attributes = {
    	fibre_channel = {
      	  wwnn    = "20:00:F4:E9:D4:56:10:AB"
      	  boot_targets = [
        	{
          		lun_id        = "2"
        	}
      	  ]
    	}
	  }
	}
	  `,
		testingInfo.Username,
		testingInfo.PasswordNIC,
		testingInfo.EndpointNIC,
		testingInfo.SystemID,
		testingInfo.NetworkAdapterID,
		testingInfo.NetworkDeviceFunctionID,
	)
}

func testAccRedfishResourceFCConfigUpdateOutDatedMW(testingInfo testingNICInputs) string {
	return fmt.Sprintf(`
	resource "redfish_network_adapter" "nic" {
	  redfish_server {
		user         = "%s"
		password     = "%s"
		endpoint     = "%s"
		ssl_insecure = true
	  }
	  system_id = "%s"
	  network_adapter_id         = "%s"
	  network_device_function_id = "%s"
	  apply_time = "AtMaintenanceWindowStart"
	  job_timeout = 1200
	  maintenance_window = {
	    start_time = "2024-06-30T05:15:40-05:00"
		duration = 600
	  }

	  network_attributes = {
    	fibre_channel = {
      	  wwnn    = "20:00:F4:E9:D4:56:10:AB"
      	  boot_targets = [
        	{
          		lun_id        = "2"
        	}
      	  ]
    	}
	  }

	  oem_network_attributes = {
	  	clear_pending = true
	  	attributes = {
	  		"PortLoginTimeout" = "4000"
	  	}
  	  }
	}
	  `,
		testingInfo.Username,
		testingInfo.PasswordNIC,
		testingInfo.EndpointNIC,
		testingInfo.SystemID,
		testingInfo.NetworkAdapterID,
		testingInfo.NetworkDeviceFunctionID,
	)
}

func testAccRedfishResourceFCConfig(testingInfo testingNICInputs) string {
	return fmt.Sprintf(`
	resource "redfish_network_adapter" "nic" {
	  redfish_server {
		user         = "%s"
		password     = "%s"
		endpoint     = "%s"
		ssl_insecure = true
	  }
	  system_id = "%s"
	  network_adapter_id         = "%s"
	  network_device_function_id = "%s"
	  apply_time = "OnReset"
	  job_timeout = 1200

	  network_attributes = {
    	fibre_channel = {
      	  wwnn    = "20:00:F4:E9:D4:56:10:AB"
      	  boot_targets = [
        	{
          		lun_id        = "2"
        	}
      	  ]
    	}
	  }

	  oem_network_attributes = {
	  	clear_pending = true
	  	attributes = {
	  		"PortLoginTimeout" = "4000"
	  	}
  	  }
	}
	  `,
		testingInfo.Username,
		testingInfo.PasswordNIC,
		testingInfo.EndpointNIC,
		testingInfo.SystemID,
		testingInfo.NetworkAdapterID,
		testingInfo.NetworkDeviceFunctionID,
	)
}

func testAccRedfishResourceFCConfig_Mocky(testingInfo testingNICInputs) string {
	return fmt.Sprintf(`
	resource "redfish_network_adapter" "nic" {
	  redfish_server {
		user         = "%s"
		password     = "%s"
		endpoint     = "%s"
		ssl_insecure = true
	  }
	  system_id = "%s"
	  network_adapter_id         = "%s"
	  network_device_function_id = "%s"
	  apply_time = "OnReset"
	  job_timeout = 1200
	  oem_network_attributes = {
	  	clear_pending = true
	  	attributes = {
	  		"PortLoginTimeout" = "4000"
	  	}
  	  }
	}
	  `,
		testingInfo.Username,
		testingInfo.PasswordNIC,
		testingInfo.EndpointNIC,
		testingInfo.SystemID,
		testingInfo.NetworkAdapterID,
		testingInfo.NetworkDeviceFunctionID,
	)
}

func testAccRedfishResourceFCConfigUpdateNetAttrs(testingInfo testingNICInputs) string {
	return fmt.Sprintf(`
	resource "redfish_network_adapter" "nic" {
	  redfish_server {
		user         = "%s"
		password     = "%s"
		endpoint     = "%s"
		ssl_insecure = true
	  }
	  system_id = "%s"
	  network_adapter_id         = "%s"
	  network_device_function_id = "%s"
	  apply_time = "OnReset"
	  job_timeout = 1200

	  network_attributes = {
    	fibre_channel = {
      	  wwnn    = "20:00:F4:E9:D4:56:10:CD"
      	  boot_targets = [
        	{
          		lun_id        = "1"
        	}
      	  ]
    	}
	  }

	  oem_network_attributes = {
	  	clear_pending = false
	  	attributes = {
	  		"PortLoginTimeout" = "4000"
	  	}
  	  }
	}
	  `,
		testingInfo.Username,
		testingInfo.PasswordNIC,
		testingInfo.EndpointNIC,
		testingInfo.SystemID,
		testingInfo.NetworkAdapterID,
		testingInfo.NetworkDeviceFunctionID,
	)
}

func testAccRedfishResourceNICAttributesIscsiConfig(testingInfo testingNICInputs) string {
	return fmt.Sprintf(`
	resource "redfish_network_adapter" "nic" {
	  redfish_server {
		user         = "%s"
		password     = "%s"
		endpoint     = "%s"
		ssl_insecure = true
	  }
	  system_id = "%s"
	  network_adapter_id         = "%s"
	  network_device_function_id = "%s"
	  apply_time = "OnReset"
	  job_timeout = 1200

	  network_attributes = {
	    iscsi_boot = {
      	  authentication_method  = "CHAP"
    	}
	  }
	}
	  `,
		testingInfo.Username,
		testingInfo.PasswordNIC,
		testingInfo.EndpointNIC,
		testingInfo.SystemID,
		testingInfo.NetworkAdapterID,
		testingInfo.NetworkDeviceFunctionID,
	)
}

func testAccRedfishResourceNICAttributesIscsiConfigUpdate(testingInfo testingNICInputs) string {
	return fmt.Sprintf(`
	resource "redfish_network_adapter" "nic" {
	  redfish_server {
		user         = "%s"
		password     = "%s"
		endpoint     = "%s"
		ssl_insecure = true
	  }
	  system_id = "%s"
	  network_adapter_id         = "%s"
	  network_device_function_id = "%s"
	  apply_time = "OnReset"
	  job_timeout = 1200

	  network_attributes = {
	    iscsi_boot = {
      	  authentication_method  = "None"
    	}
	  }
	}
	  `,
		testingInfo.Username,
		testingInfo.PasswordNIC,
		testingInfo.EndpointNIC,
		testingInfo.SystemID,
		testingInfo.NetworkAdapterID,
		testingInfo.NetworkDeviceFunctionID,
	)
}
