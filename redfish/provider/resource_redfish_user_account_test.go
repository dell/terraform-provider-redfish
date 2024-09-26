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
	"log"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

const userID = "15"

func init() {
	resource.AddTestSweepers("redfish_user_account", &resource.Sweeper{
		Name: "redfish_user_account",
		F: func(region string) error {
			log.Println("Sweepers for user")
			service, err := getSweeperClient(region)
			if err != nil {
				log.Println("Error getting sweeper client ", err.Error())
				return nil
			}
			_, account, _ := GetUserAccountFromID(service, userID)

			if account != nil { // user exists so we need to remove it
				// PATCH call to remove username.
				payload := make(map[string]interface{})
				payload["UserName"] = ""
				payload["Enable"] = "false"
				payload["RoleId"] = "None"
				_, err = service.GetClient().Patch(account.ODataID, payload)
				if err != nil {
					log.Println("failed to sweep dangling user.")
					return nil
				}

			}
			return nil
		},
	})
}

// Test to create and update redfish user - Positive
func TestAccRedfishUser_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRedfishResourceUserConfig(
					creds,
					"test1",
					"Test@1234",
					"Operator",
					true,
					userID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("redfish_user_account.user_config", "username", "test1"),
				),
			},
			{
				Config: testAccRedfishResourceUserConfig(
					creds,
					"test1",
					"Test@1234",
					"None",
					false,
					userID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("redfish_user_account.user_config", "username", "test1"),
				),
			},
		},
	})
}

// Test to create user with invalid role-id - Negative
func TestAccRedfishUserInvalid_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRedfishResourceUserConfig(
					creds,
					"test1",
					"Test@1234",
					"Admin",
					false,
					userID),
				ExpectError: regexp.MustCompile("Invalid Attribute Value Match"),
			},
		},
	})
}

// Test to create user with existing username - Negative
func TestAccRedfishUserExisting_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRedfishResourceUserConfig(
					creds,
					"root",
					"Xyz@123",
					"Administrator",
					true,
					userID),
				ExpectError: regexp.MustCompile("user root already exists against ID 2"),
			},
		},
	})
}

// Test to update username to existing username - Negative
func TestAccRedfishUserUpdateInvalid_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRedfishResourceUserConfig(
					creds,
					"test1",
					"Test@1234",
					"Administrator",
					true,
					userID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("redfish_user_account.user_config", "username", "test1"),
				),
			},
			{
				Config: testAccRedfishResourceUserConfig(
					creds,
					"root",
					"Test@1234",
					"Administrator",
					false,
					userID),
				ExpectError: regexp.MustCompile("user root already exists"),
			},
		},
	})
}

// Test to create user with Invalid ID - Negative
func TestAccRedfishUserUpdateInvalidId_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRedfishResourceUserConfig(
					creds,
					"test1",
					"Test@1234",
					"Administrator",
					true,
					"1"),
				ExpectError: regexp.MustCompile("User_id can vary between 3 to 16 only"),
			},
		},
	})
}

// Test to update user-id - Negative
func TestAccRedfishUserUpdateId_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRedfishResourceUserConfig(
					creds,
					"test1",
					"Test@1234",
					"Administrator",
					true,
					userID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("redfish_user_account.user_config", "username", "test1"),
				),
			},
			{
				Config: testAccRedfishResourceUserConfig(
					creds,
					"test1",
					"Test@1234",
					"Administrator",
					false,
					"1"),
				ExpectError: regexp.MustCompile("user_id cannot be updated"),
			},
		},
	})
}

// Test to update username - positive
func TestAccRedfishUserUpdateUser_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRedfishResourceUserConfig(
					creds,
					"test1",
					"Test@1234",
					"Administrator",
					true,
					userID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("redfish_user_account.user_config", "username", "test1"),
				),
			},
			{
				Config: testAccRedfishResourceUserConfig(
					creds,
					"test2",
					"Test@1234",
					"Administrator",
					false,
					userID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("redfish_user_account.user_config", "username", "test2"),
				),
			},
		},
	})
}

// Test to import user - positive
func TestAccRedfishUserImportUser_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRedfishResourceUserConfig(
					creds,
					"test1",
					"Test@1234",
					"None",
					false,
					userID),
				ResourceName:  "redfish_user_account.user_config",
				ImportState:   true,
				ImportStateId: "{\"id\":\"3\",\"username\":\"" + creds.Username + "\",\"password\":\"" + creds.Password + "\",\"endpoint\":\"https://" + creds.Endpoint + "\",\"ssl_insecure\":true}",
				ExpectError:   nil,
			},
		},
	})
}

// Test to import user - negative
func TestAccRedfishUserImportUser_invalid(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRedfishResourceUserConfig(
					creds,
					"test1",
					"Test@1234",
					"None",
					false,
					userID),
				ResourceName:  "redfish_user_account.user_config",
				ImportState:   true,
				ImportStateId: "{\"id\":\"invalid\",\"username\":\"" + creds.Username + "\",\"password\":\"" + creds.Password + "\",\"endpoint\":\"https://" + creds.Endpoint + "\",\"ssl_insecure\":true}",
				ExpectError:   regexp.MustCompile("Error when retrieving accounts"),
			},
		},
	})
}

// validation tests - Negative
func TestAccRedfishUserValidation_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRedfishResourceUserConfig(
					creds,
					"test1234567890123456",
					"Test@1234",
					"Administrator",
					false,
					userID),
				ExpectError: regexp.MustCompile("Invalid Attribute Value Length"),
			},
			{
				Config: testAccRedfishResourceUserConfig(
					creds,
					"test1",
					"T@1",
					"Administrator",
					false,
					userID),
				ExpectError: regexp.MustCompile("Attribute password string length must be between 4 and 40"),
			},
			{
				Config: testAccRedfishResourceUserConfig(
					creds,
					"test1",
					"test123",
					"Administrator",
					true,
					userID),
				ExpectError: regexp.MustCompile("Password validation failed"),
			},
			{
				Config: testAccRedfishResourceUserConfig(
					creds,
					"test1",
					"Test@1234",
					"Administrator",
					false,
					"2"),
				ExpectError: regexp.MustCompile("User ID already exists"),
			},
			{
				Config: testAccRedfishResourceUserConfig(
					creds,
					"test2",
					"Test@1234",
					"Administrator",
					false,
					userID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("redfish_user_account.user_config", "username", "test2"),
				),
			},
			{
				Config: testAccRedfishResourceUserConfig(
					creds,
					"test1",
					"test123",
					"Administrator",
					true,
					userID),
				ExpectError: regexp.MustCompile("Password validation failed"),
			},
		},
	})
}

func TestAccRedfishUserPassword_alias(t *testing.T) {
	serverAlias := "my-server-1"
	testUser, testUserPass, testUserRole := "testAlias", "Test@1234", "Administrator"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				// prepare test user
				Config: testAccRedfishProviderWithServersConfig(serverAlias, creds.Username, creds.Password, creds.Endpoint) +
					testAccRedfishResourceUserConfig(creds, testUser, testUserPass, testUserRole, true, userID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("redfish_user_account.user_config", "username", testUser),
				),
			},
			{
				// use `testUser` as provder creds
				Config: testAccRedfishProviderWithServersConfig(serverAlias, testUser, testUserPass, creds.Endpoint) +
					testAccRedfishResourceUserConfig(creds, testUser, testUserPass, testUserRole, true, userID) +
					testAccRedfishResourceUserImportConfig_alias(serverAlias, testUser, testUserPass, testUserRole, true, userID) +
					testAccRedfishResourceUserConfig_power(serverAlias, "GracefulRestart"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("redfish_user_account.user_config", "username", testUser),
					resource.TestCheckResourceAttr("redfish_user_account.user_creds", "username", testUser),
					resource.TestCheckResourceAttr("redfish_power.system_power", "desired_power_action", "GracefulRestart"),
				),
			},
			// Invalidation password test: update `testUser` password, but provider still use old passed.
			// TODO: skip, always failed due to post-apply refresh
			// {
			// 	Config: testAccRedfishProviderWithServersConfig(serverAlias, testUser, testUserPass, creds.Endpoint) +
			// 		testAccRedfishResourceUserConfig(creds, testUser, testUserPass, testUserRole, true, userID) +
			// 		testAccRedfishResourceUserImportConfig_alias(serverAlias, testUser, "NewTest@1234", testUserRole, true, userID) +
			// 		testAccRedfishResourceUserConfig_power(serverAlias, "GracefulRestart"),
			// 	Check: resource.ComposeAggregateTestCheckFunc(
			// 		resource.TestCheckResourceAttr("redfish_user_account.user_config", "username", testUser),
			// 		resource.TestCheckResourceAttr("redfish_user_account.user_creds", "username", testUser),
			// 		resource.TestCheckResourceAttr("redfish_power.system_power", "desired_power_action", "GracefulRestart"),
			// 	),
			// },
		},
	})
}
func testAccRedfishResourceUserConfig(testingInfo TestingServerCredentials,
	username string,
	password string,
	roleId string,
	enabled bool,
	userId string,
) string {
	return fmt.Sprintf(`
		
		resource "redfish_user_account" "user_config" {
		
		  redfish_server {
			user = "%s"
			password = "%s"
			endpoint = "https://%s"
			ssl_insecure = true
		  }

		  username = "%s"
		  password = "%s"
		  role_id = "%s"
		  enabled = %t
		  user_id = %s
		}
		`,
		testingInfo.Username,
		testingInfo.Password,
		testingInfo.Endpoint,
		username,
		password,
		roleId,
		enabled,
		userId,
	)
}

func testAccRedfishProviderWithServersConfig(serverAlias, username, password, endpoint string) string {
	return fmt.Sprintf(`
		locals {
			rack1 = {
				"%s" = {
					user         = "%s"
					password     = "%s"
					endpoint = "https://%s"
					ssl_insecure = true
				},
			}
		}
		provider "redfish" {
			redfish_servers = local.rack1
		}
	`,
		serverAlias,
		username,
		password,
		endpoint)
}

func testAccRedfishResourceUserImportConfig_alias(alias string, username string, password string, roleId string, enabled bool, userId string) string {
	return fmt.Sprintf(`

		import {
			to = redfish_user_account.user_creds
			id = jsonencode({
				id = "%s"
				redfish_alias = "%s"
			})
		  }

		resource "redfish_user_account" "user_creds" {
		
		  redfish_server {
			redfish_alias = "%s"
		  }

		  username = "%s"
		  password = "%s"
		  role_id = "%s"
		  enabled = %t
		  user_id = "%s"
		}
		`,
		userId,
		alias,
		alias,
		username,
		password,
		roleId,
		enabled,
		userId,
	)
}

func testAccRedfishResourceUserConfig_power(alias, powerAction string) string {
	return fmt.Sprintf(`
		
	resource "redfish_power" "system_power" {
	
	  redfish_server {
		redfish_alias = "%s"
	  }
	  system_id = "System.Embedded.1"
	  desired_power_action = "%s"
	  maximum_wait_time = 120
	  check_interval = 12
	}
	`,
		alias,
		powerAction,
	)
}
