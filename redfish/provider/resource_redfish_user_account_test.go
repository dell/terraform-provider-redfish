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

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
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

func TestAccRedfishUser_Multiple(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRedfishResourceMultipleUserConfig(
					creds),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("redfish_user_account.user_config", "users.0.username", "tom"),
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

func testAccRedfishResourceMultipleUserConfig(testingInfo TestingServerCredentials) string {
	return fmt.Sprintf(`

		resource "redfish_user_account" "user_config" {

		  redfish_server {
			user = "%s"
			password = "%s"
			endpoint = "https://%s"
			ssl_insecure = true
		  }

		  users = [
			{
			    username = "tom",
			    password = "T0mPassword123!",
			    role_id = "Operator",
			    enabled = true,
			},
			{
			    username = "dick"
			    password = "D!ckPassword123!"
			    role_id = "ReadOnly"
			    enabled = true
			},
			{
			    username = "harry"
			    password = "H@rryPassword123!"
			    role_id = "ReadOnly"
			    enabled = true
			},
		  ]
		}
		`,
		testingInfo.Username,
		testingInfo.Password,
		testingInfo.Endpoint,
	)
}
