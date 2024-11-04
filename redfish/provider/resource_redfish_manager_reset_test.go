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

// Test to create manager reset resource with invalid reset type- Negative
func TestAccRedfishManagerReset_Invalid_ResetType_Negative(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccRedfishResourceManagerResetConfig(creds, "iDRAC.Embedded.1", "On"),
				ExpectError: regexp.MustCompile("Invalid Attribute Value Match"),
			},
		},
	})
}

// Test to create manager reset resource with invalid manager id- Negative
func TestAccRedfishManagerReset_Invalid_ManagerID_Negative(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccRedfishResourceManagerResetConfig(creds, "iDRAC.Embedded.0", "GracefulRestart"),
				ExpectError: regexp.MustCompile("invalid Manager ID provided"),
			},
		},
	})
}

// Test to update manager reset resource with invalid maanger id- Negative
func TestAccRedfishManagerReset_Update_Negative(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRedfishResourceManagerResetConfig(creds, "iDRAC.Embedded.1", "GracefulRestart"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("redfish_manager_reset.manager_reset", "id", "iDRAC.Embedded.1"),
				),
			},
			{
				Config:      testAccRedfishResourceManagerResetConfig(creds, "iDRAC.Embedded", "GracefulRestart"),
				ExpectError: regexp.MustCompile("invalid Manager ID provided"),
			},
		},
	})
}

// Test to perform manager reset
func TestAccRedfishManagerReset_Create(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRedfishResourceManagerResetConfig(creds, "iDRAC.Embedded.1", "GracefulRestart"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("redfish_manager_reset.manager_reset", "id", "iDRAC.Embedded.1"),
				),
			},
		},
	})
}

func testAccRedfishResourceManagerResetConfig(testingInfo TestingServerCredentials,
	managerID string,
	resetType string,
) string {
	return fmt.Sprintf(`
		
	resource "redfish_manager_reset" "manager_reset" {
		redfish_server {
		  user         = "%s"
		  password     = "%s"
		  endpoint     = "%s"
		  ssl_insecure = true
		}
	  
		id = "%s"
		reset_type = "%s"
	}
		`,
		testingInfo.Username,
		testingInfo.Password,
		testingInfo.Endpoint,
		managerID,
		resetType,
	)
}
