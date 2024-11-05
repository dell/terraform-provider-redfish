/*
Copyright (c) 2023-2024 Dell Inc., or its subsidiaries. All Rights Reserved.

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

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// test redfish bios settings
func TestAccRedfishBios_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRedfishResourceBiosConfigOn(
					creds),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("redfish_bios.bios", "attributes.NumLock", "On"),
				),
			},
			{
				Config: testAccRedfishResourceBiosConfigOff(
					creds),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("redfish_bios.bios", "attributes.NumLock", "Off"),
				),
			},
		},
	})
}

func TestAccRedfishBios_InvalidSettings(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRedfishResourceBiosConfigInvalidSettingsApplyTime(
					creds),
				ExpectError: regexp.MustCompile("Attribute settings_apply_time value must be one of"),
			},
		},
	})
}

func TestAccRedfishBios_InvalidAttributes(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRedfishResourceBiosConfigInvalidAttributes(
					creds),
				ExpectError: regexp.MustCompile("Attribute settings_apply_time value must be one of"),
			},
		},
	})
}

// Test to import bios - positive
func TestAccRedfishBios_Import(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRedfishResourceBiosConfigOn(
					creds),
				ResourceName:  "redfish_bios.bios",
				ImportState:   true,
				ImportStateId: "{\"username\":\"" + creds.Username + "\",\"password\":\"" + creds.Password + "\",\"endpoint\":\"" + creds.Endpoint + "\",\"ssl_insecure\":true}",
				ExpectError:   nil,
			},
		},
	})
}

func TestAccRedfishBios_ImportSystemID(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRedfishResourceBiosConfigOn(
					creds),
				ResourceName:  "redfish_bios.bios",
				ImportState:   true,
				ImportStateId: "{\"username\":\"" + creds.Username + "\",\"password\":\"" + creds.Password + "\",\"endpoint\":\"" + creds.Endpoint + "\",\"ssl_insecure\":true,\"system_id\":\"System.Embedded.1\"}",
				ExpectError:   nil,
			},
		},
	})
}

func testAccRedfishResourceBiosConfigOn(testingInfo TestingServerCredentials) string {
	return fmt.Sprintf(`

		resource "redfish_bios" "bios"  {
		
		  redfish_server {
			user = "%s"
			password = "%s"
			endpoint = "%s"
			ssl_insecure = true
		  }

		  attributes = {
			"NumLock" = "On"
		  }
		  reset_type = "ForceRestart"
		//   system_id = "System.Embedded.1"
		}
		`,
		testingInfo.Username,
		testingInfo.Password,
		testingInfo.Endpoint,
	)
}

func testAccRedfishResourceBiosConfigOff(testingInfo TestingServerCredentials) string {
	return fmt.Sprintf(`

		resource "redfish_bios" "bios"  {
		
		  redfish_server {
			user = "%s"
			password = "%s"
			endpoint = "%s"
			ssl_insecure = true
		  }

		  attributes = {
			"NumLock" = "Off"
			"AcPwrRcvryUserDelay" = 70
		  }
		  reset_type = "ForceRestart"
   		  bios_job_timeout = 1200
		  reset_timeout = 120
		}
		`,
		testingInfo.Username,
		testingInfo.Password,
		testingInfo.Endpoint,
	)
}

func testAccRedfishResourceBiosConfigInvalidSettingsApplyTime(testingInfo TestingServerCredentials) string {
	return fmt.Sprintf(`

		resource "redfish_bios" "bios"  {
		
		  redfish_server {
			user = "%s"
			password = "%s"
			endpoint = "%s"
			ssl_insecure = true
		  }

		  attributes = {
		  }
		  settings_apply_time = "random"
		}
		`,
		testingInfo.Username,
		testingInfo.Password,
		testingInfo.Endpoint,
	)
}

func testAccRedfishResourceBiosConfigInvalidAttributes(testingInfo TestingServerCredentials) string {
	return fmt.Sprintf(`

		resource "redfish_bios" "bios"  {
		
		  redfish_server {
			user = "%s"
			password = "%s"
			endpoint = "%s"
			ssl_insecure = true
		  }

		  attributes = {
		  }
		  settings_apply_time = "ForceRestart"
		}
		`,
		testingInfo.Username,
		testingInfo.Password,
		testingInfo.Endpoint,
	)
}
