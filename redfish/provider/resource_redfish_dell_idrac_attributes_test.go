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

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccRedfishIDRACAttributesBasic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRedfishResourceIDracAttributesConfig(
					creds, "avengers"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("redfish_dell_idrac_attributes.idrac", "attributes.Users.3.Enable", "Disabled"),
					resource.TestCheckResourceAttr("redfish_dell_idrac_attributes.idrac", "attributes.Time.1.Timezone", "CST6CDT"),
				),
			},
			{
				Config: testAccRedfishResourceIDracAttributesConfig(
					creds, "ironman"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("redfish_dell_idrac_attributes.idrac", "attributes.Users.3.Enable", "Disabled"),
					resource.TestCheckResourceAttr("redfish_dell_idrac_attributes.idrac", "attributes.Time.1.Timezone", "CST6CDT"),
				),
			},
		},
	})
}

func TestAccRedfishIDRACAttributesInvalidAttribute(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRedfishResourceIDracAttributesConfigInvalid(
					creds),
				ExpectError: regexp.MustCompile("there was an issue when creating/updating idrac attributes"),
			},
		},
	})
}

func TestAccRedfishIDRACAttributeImport(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `resource "redfish_dell_idrac_attributes" "idrac" {
				}`,
				ResourceName:  "redfish_dell_idrac_attributes.idrac",
				ImportState:   true,
				ImportStateId: "{\"username\":\"" + creds.Username + "\",\"password\":\"" + creds.Password + "\",\"endpoint\":\"https://" + creds.Endpoint + "\",\"ssl_insecure\":true}",
				ExpectError:   nil,
			},
		},
	})
}

func TestAccRedfishIDRACAttributeImportByFilter(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `resource "redfish_dell_idrac_attributes" "idrac" {
				}`,
				ResourceName:  "redfish_dell_idrac_attributes.idrac",
				ImportState:   true,
				ImportStateId: "{\"attributes\":[\"Users.2.UserName\"],\"username\":\"" + creds.Username + "\",\"password\":\"" + creds.Password + "\",\"endpoint\":\"https://" + creds.Endpoint + "\",\"ssl_insecure\":true}",
				ExpectError:   nil,
			},
		},
	})
}

func testAccRedfishResourceIDracAttributesConfig(testingInfo TestingServerCredentials, username string) string {
	return fmt.Sprintf(`
	resource "redfish_dell_idrac_attributes" "idrac" {
		redfish_server {
		  user         = "%s"
		  password     = "%s"
		  endpoint     = "https://%s"
		  ssl_insecure = true
		}
	  
		attributes = {
		  "Users.3.Enable"    		  = "Disabled"
		  "Users.3.UserName"  		  = "%s"
		  "Users.3.Password"  		  = "test1234"
		  "Users.3.Privilege" 		  = 511
		  "Time.1.Timezone"   		  = "CST6CDT",
		  "SysLog.1.PowerLogInterval" = 5,
		}
	  }
	  `,
		testingInfo.Username,
		testingInfo.Password,
		testingInfo.Endpoint,
		username,
	)
}

func testAccRedfishResourceIDracAttributesConfigInvalid(testingInfo TestingServerCredentials) string {
	return fmt.Sprintf(`
	resource "redfish_dell_idrac_attributes" "idrac" {
		redfish_server {
		  user         = "%s"
		  password     = "%s"
		  endpoint     = "https://%s"
		  ssl_insecure = true
		}
	  
		attributes = {
		  "Users.3.Enable"            = "Disabled"
		  "Users.3.UserName"          = "mike"
		  "Users.3.Password"          = "test1234"
		  "Users.3.Privilege"         = 511
		  "Time.1.Timezone"			  = "CST6CDT",
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
