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
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

// test redfish Boot Order
func TestAccRedfishBootSourceOverride_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRedfishResourceBootSourceLegacyconfig(creds),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("redfish_boot_source_override.boot", "boot_source_override_mode", "Legacy"),
				),
			},
		},
	})
}

func TestAccRedfishBootSourceOverride_updated(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{

				Config: testAccRedfishResourceBootSourceResetType(creds),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("redfish_boot_source_override.boot", "boot_source_override_mode", "UEFI"),
				),
			},
			{
				PreConfig: func() {
					time.Sleep(120 * time.Second)
				},
				Config: testAccRedfishResourceBootSourceUEFIconfig(creds),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("redfish_boot_source_override.boot", "boot_source_override_mode", "UEFI"),
				),
			},
		},
	})
}

func testAccRedfishResourceBootSourceLegacyconfig(testingInfo TestingServerCredentials) string {
	return fmt.Sprintf(`

	resource "redfish_boot_source_override" "boot" {
		redfish_server {
			user = "%s"
			password = "%s"
			endpoint = "https://%s"
			ssl_insecure = true
		}
	    system_id = "System.Embedded.1"
		boot_source_override_enabled = "Once"
		boot_source_override_target = "Pxe"
		boot_source_override_mode = "Legacy"
		reset_type    = "GracefulRestart"
	}	  
	`,
		testingInfo.Username,
		testingInfo.Password,
		testingInfo.Endpoint,
	)
}

func testAccRedfishResourceBootSourceUEFIconfig(testingInfo TestingServerCredentials) string {
	return fmt.Sprintf(`

	resource "redfish_boot_source_override" "boot" {
	  
		redfish_server {
			user = "%s"
			password = "%s"
			endpoint = "https://%s"
			ssl_insecure = true
		}
	   
		boot_source_override_enabled = "Once"
		boot_source_override_target = "UefiTarget"
		boot_source_override_mode = "UEFI"
		reset_type    = "GracefulRestart"
	}	  
	`,
		testingInfo.Username,
		testingInfo.Password,
		testingInfo.Endpoint,
	)
}

func testAccRedfishResourceBootSourceResetType(testingInfo TestingServerCredentials) string {
	return fmt.Sprintf(`

	resource "redfish_boot_source_override" "boot" {
	  
		redfish_server {
			user = "%s"
			password = "%s"
			endpoint = "https://%s"
			ssl_insecure = true
		}
	   
		boot_source_override_enabled = "Once"
		boot_source_override_target = "UefiTarget"
		boot_source_override_mode = "UEFI"
		reset_type    = "ForceRestart"
	}	  
	`,
		testingInfo.Username,
		testingInfo.Password,
		testingInfo.Endpoint,
	)
}
