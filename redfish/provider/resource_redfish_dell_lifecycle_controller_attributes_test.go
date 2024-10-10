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

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccRedfishLCAttributesBasic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRedfishResourceLCAttributesConfig(creds),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("redfish_dell_lc_attributes.lc", "attributes.LCAttributes.1.IgnoreCertWarning", "On"),
					resource.TestCheckResourceAttr("redfish_dell_lc_attributes.lc", "attributes.LCAttributes.1.CollectSystemInventoryOnRestart", "Disabled"),
				),
			},
			{
				Config: `resource "redfish_dell_lc_attributes" "lc" {
				}`,
				ResourceName:  "redfish_dell_lc_attributes.lc",
				ImportState:   true,
				ImportStateId: "{\"attributes\":[\"LCAttributes.1.CollectSystemInventoryOnRestart\"],\"username\":\"" + creds.Username + "\",\"password\":\"" + creds.Password + "\",\"endpoint\":\"https://" + creds.Endpoint + "\",\"ssl_insecure\":true}",
				ExpectError:   nil,
			},
		},
	})
}

func TestAccRedfishLCAttributesInvalidAttribute(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRedfishResourceLCConfigInvalid(
					creds),
				ExpectError: regexp.MustCompile("there was an issue when creating/updating LC attributes"),
			},
		},
	})
}

func TestAccRedfishLCAttributesUpdate(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRedfishResourceLCAttributesConfig(creds),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("redfish_dell_lc_attributes.lc", "attributes.LCAttributes.1.IgnoreCertWarning", "On"),
					resource.TestCheckResourceAttr("redfish_dell_lc_attributes.lc", "attributes.LCAttributes.1.CollectSystemInventoryOnRestart", "Disabled"),
				),
			},
			{
				Config: testAccRedfishResourceLCAttributesUpdateConfig(creds),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("redfish_dell_lc_attributes.lc", "attributes.LCAttributes.1.IgnoreCertWarning", "Off"),
					resource.TestCheckResourceAttr("redfish_dell_lc_attributes.lc", "attributes.LCAttributes.1.CollectSystemInventoryOnRestart", "Enabled"),
				),
			},
		},
	})
}

func TestAccRedfishLCAttributeImport(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `resource "redfish_dell_lc_attributes" "lc" {
				}`,
				ResourceName:  "redfish_dell_lc_attributes.lc",
				ImportState:   true,
				ImportStateId: "{\"username\":\"" + creds.Username + "\",\"password\":\"" + creds.Password + "\",\"endpoint\":\"https://" + creds.Endpoint + "\",\"ssl_insecure\":true}",
				ExpectError:   nil,
			},
		},
	})
}

func testAccRedfishResourceLCAttributesConfig(testingInfo TestingServerCredentials) string {
	return fmt.Sprintf(`
	resource "redfish_dell_lc_attributes" "lc" {
		redfish_server {
		  user         = "%s"
		  password     = "%s"
		  endpoint     = "https://%s"
		  ssl_insecure = true
		}
	  
		attributes = {
			"LCAttributes.1.CollectSystemInventoryOnRestart" = "Disabled"
			"LCAttributes.1.IgnoreCertWarning" = "On"
		}
	  }
	  `,
		testingInfo.Username,
		testingInfo.Password,
		testingInfo.Endpoint,
	)
}

func testAccRedfishResourceLCAttributesUpdateConfig(testingInfo TestingServerCredentials) string {
	return fmt.Sprintf(`
	resource "redfish_dell_lc_attributes" "lc" {
		redfish_server {
		  user         = "%s"
		  password     = "%s"
		  endpoint     = "https://%s"
		  ssl_insecure = true
		}

		attributes = {
			"LCAttributes.1.CollectSystemInventoryOnRestart" = "Enabled"
			"LCAttributes.1.IgnoreCertWarning" = "Off"
		}
	  }
	  `,
		testingInfo.Username,
		testingInfo.Password,
		testingInfo.Endpoint,
	)
}

func testAccRedfishResourceLCConfigInvalid(testingInfo TestingServerCredentials) string {
	return fmt.Sprintf(`
	resource "redfish_dell_lc_attributes" "lc" {
		redfish_server {
		  user         = "%s"
		  password     = "%s"
		  endpoint     = "https://%s"
		  ssl_insecure = true
		}
	  
		attributes = {
			"LCAttributes.1.CollectSystemInventoryOnRestart" = "Disabled",
			"LCAttributes.1.IgnoreCertWarning" = "On",
		  	"SysLog.1.PowerLogInterval" = 5,
		  	"InvalidAttribute" 		  = "invalid",
		}
	  }
	  `,
		testingInfo.Username,
		testingInfo.Password,
		testingInfo.Endpoint,
	)
}
