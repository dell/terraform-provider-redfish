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
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
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
		NetworkDeviceFunctionID:  "NIC.Integrated.1-4-1",
		NetworkAdapterID:         "NIC.Integrated.1",
		SystemID:                 "System.Embedded.1",
	}

	fcParams = testingNICInputs{
		TestingServerCredentials: creds,
		NetworkDeviceFunctionID:  "FC.Slot.1-2",
		NetworkAdapterID:         "FC.Slot.1",
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

func TestAccRedfishNICAttributesImport(t *testing.T) {
	importReqID := fmt.Sprintf("{\"system_id\":\"%s\",\"network_adapter_id\":\"%s\",\"network_device_function_id\":\"%s\",\"username\":\"%s\",\"password\":\"%s\",\"endpoint\":\"https://%s\",\"ssl_insecure\":true}",
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

func testAccRedfishResourceNICAttributesConfigNetworkAttrs(testingInfo testingNICInputs) string {
	return fmt.Sprintf(`
	resource "redfish_network_adapter" "nic" {
	  redfish_server {
		user         = "%s"
		password     = "%s"
		endpoint     = "https://%s"
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
		testingInfo.Password,
		testingInfo.Endpoint,
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
		endpoint     = "https://%s"
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
		testingInfo.Password,
		testingInfo.Endpoint,
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
		endpoint     = "https://%s"
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
		testingInfo.Password,
		testingInfo.Endpoint,
		testingInfo.SystemID,
		testingInfo.NetworkAdapterID,
		testingInfo.NetworkDeviceFunctionID,
	)
}
