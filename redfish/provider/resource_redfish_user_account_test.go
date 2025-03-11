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
	"os"
	"regexp"
	"testing"

	"github.com/bytedance/mockey"
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
				log.Println("Error getting sweeper client")
				return nil
			}
			_, account, err := GetUserAccountFromID(service, userID)
			if err != nil {
				log.Println("Error getting user by ID.")
				return nil
			}
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
	version := os.Getenv("TF_TESTING_REDFISH_VERSION")
	if version == "17" {
		t.Skip("Skipping Bios Tests for 17G")
	}
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				PreConfig: func() {
					FunctionMocker = mockey.Mock(isServerGenerationSeventeenAndAbove).Return(false, nil).Build()
				},
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
				PreConfig: func() {
					if FunctionMocker != nil {
						FunctionMocker.Release()
					}
					FunctionMocker = mockey.Mock(isServerGenerationSeventeenAndAbove).Return(false, nil).Build()
				},
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
	if FunctionMocker != nil {
		FunctionMocker.Release()
	}
}

// Test to create user with invalid role-id - Negative
func TestAccRedfishUserInvalid_basic(t *testing.T) {
	version := os.Getenv("TF_TESTING_REDFISH_VERSION")
	if version == "17" {
		t.Skip("Skipping Bios Tests for 17G")
	}
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				PreConfig: func() {
					FunctionMocker = mockey.Mock(isServerGenerationSeventeenAndAbove).Return(false, nil).Build()
				},
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
	if FunctionMocker != nil {
		FunctionMocker.Release()
	}
}

// Test to create user with existing username - Negative
func TestAccRedfishUserExisting_basic(t *testing.T) {
	version := os.Getenv("TF_TESTING_REDFISH_VERSION")
	if version == "17" {
		t.Skip("Skipping Bios Tests for 17G")
	}
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				PreConfig: func() {
					FunctionMocker = mockey.Mock(isServerGenerationSeventeenAndAbove).Return(false, nil).Build()
				},
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
	if FunctionMocker != nil {
		FunctionMocker.Release()
	}
}

// Test to update username to existing username - Negative
func TestAccRedfishUserUpdateInvalid_basic(t *testing.T) {
	version := os.Getenv("TF_TESTING_REDFISH_VERSION")
	if version == "17" {
		t.Skip("Skipping Bios Tests for 17G")
	}
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				PreConfig: func() {
					FunctionMocker = mockey.Mock(isServerGenerationSeventeenAndAbove).Return(false, nil).Build()
				},
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
				PreConfig: func() {
					if FunctionMocker != nil {
						FunctionMocker.Release()
					}
					FunctionMocker = mockey.Mock(isServerGenerationSeventeenAndAbove).Return(false, nil).Build()
				},
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
	if FunctionMocker != nil {
		FunctionMocker.Release()
	}
}

// Test to create user with Invalid ID - Negative
func TestAccRedfishUserUpdateInvalidId_basic(t *testing.T) {
	version := os.Getenv("TF_TESTING_REDFISH_VERSION")
	if version == "17" {
		t.Skip("Skipping Bios Tests for 17G")
	}
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				PreConfig: func() {
					FunctionMocker = mockey.Mock(isServerGenerationSeventeenAndAbove).Return(false, nil).Build()
				},
				Config: testAccRedfishResourceUserConfig(
					creds,
					"test1",
					"Test@1234",
					"Administrator",
					true,
					"1"),
				ExpectError: regexp.MustCompile("user_id can vary between 3 to 16 only"),
			},
		},
	})
	if FunctionMocker != nil {
		FunctionMocker.Release()
	}
}

// Test to update user-id - Negative
func TestAccRedfishUserUpdateId_basic(t *testing.T) {
	version := os.Getenv("TF_TESTING_REDFISH_VERSION")
	if version == "17" {
		t.Skip("Skipping Bios Tests for 17G")
	}
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				PreConfig: func() {
					FunctionMocker = mockey.Mock(isServerGenerationSeventeenAndAbove).Return(false, nil).Build()
				},
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
				PreConfig: func() {
					if FunctionMocker != nil {
						FunctionMocker.Release()
					}
					FunctionMocker = mockey.Mock(isServerGenerationSeventeenAndAbove).Return(false, nil).Build()
				},
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
	if FunctionMocker != nil {
		FunctionMocker.Release()
	}
}

// Test to update username - positive
func TestAccRedfishUserUpdateUser_basic(t *testing.T) {
	version := os.Getenv("TF_TESTING_REDFISH_VERSION")
	if version == "17" {
		t.Skip("Skipping Bios Tests for 17G")
	}
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				PreConfig: func() {
					FunctionMocker = mockey.Mock(isServerGenerationSeventeenAndAbove).Return(false, nil).Build()
				},
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
				PreConfig: func() {
					if FunctionMocker != nil {
						FunctionMocker.Release()
					}
					FunctionMocker = mockey.Mock(isServerGenerationSeventeenAndAbove).Return(false, nil).Build()
				},
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
	if FunctionMocker != nil {
		FunctionMocker.Release()
	}
}

// Test to import user - positive
func TestAccRedfishUserImportUser_basic(t *testing.T) {
	version := os.Getenv("TF_TESTING_REDFISH_VERSION")
	if version == "17" {
		t.Skip("Skipping Bios Tests for 17G")
	}
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				PreConfig: func() {
					FunctionMocker = mockey.Mock(isServerGenerationSeventeenAndAbove).Return(false, nil).Build()
				},
				Config: testAccRedfishResourceUserConfig(
					creds,
					"test1",
					"Test@1234",
					"None",
					false,
					userID),
				ResourceName:  "redfish_user_account.user_config",
				ImportState:   true,
				ImportStateId: "{\"id\":\"3\",\"username\":\"" + creds.Username + "\",\"password\":\"" + creds.Password + "\",\"endpoint\":\"" + creds.Endpoint + "\",\"ssl_insecure\":true}",
				ExpectError:   nil,
			},
		},
	})
	if FunctionMocker != nil {
		FunctionMocker.Release()
	}
}

// Test to import user - negative
func TestAccRedfishUserImportUser_invalid(t *testing.T) {
	version := os.Getenv("TF_TESTING_REDFISH_VERSION")
	if version == "17" {
		t.Skip("Skipping Bios Tests for 17G")
	}
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				PreConfig: func() {
					FunctionMocker = mockey.Mock(isServerGenerationSeventeenAndAbove).Return(false, nil).Build()
				},
				Config: testAccRedfishResourceUserConfig(
					creds,
					"test1",
					"Test@1234",
					"None",
					false,
					userID),
				ResourceName:  "redfish_user_account.user_config",
				ImportState:   true,
				ImportStateId: "{\"id\":\"invalid\",\"username\":\"" + creds.Username + "\",\"password\":\"" + creds.Password + "\",\"endpoint\":\"" + creds.Endpoint + "\",\"ssl_insecure\":true}",
				ExpectError:   regexp.MustCompile("Error when retrieving accounts"),
			},
		},
	})
	if FunctionMocker != nil {
		FunctionMocker.Release()
	}
}

// validation tests - Negative
func TestAccRedfishUserValidation_basic(t *testing.T) {
	version := os.Getenv("TF_TESTING_REDFISH_VERSION")
	if version == "17" {
		t.Skip("Skipping Bios Tests for 17G")
	}
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

// Test to create and update redfish user - Positive
func TestAccRedfishUser17G_basic(t *testing.T) {
	version := os.Getenv("TF_TESTING_REDFISH_VERSION")
	if version != "17" {
		t.Skip("Skipping Bios Tests for below 17G")
	}
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				PreConfig: func() {
					FunctionMocker = mockey.Mock(isServerGenerationSeventeenAndAbove).Return(true, nil).Build()
				},
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
			{ // Update user with ReadOnly role
				Config: testAccRedfishResourceUserConfig(
					creds,
					"test1",
					"Test@1234",
					"ReadOnly",
					false,
					userID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("redfish_user_account.user_config", "username", "test1"),
				),
			},
		},
	})
	if FunctionMocker != nil {
		FunctionMocker.Release()
	}
}

func TestAccRedfishUser17GWithoutId_basic(t *testing.T) {
	version := os.Getenv("TF_TESTING_REDFISH_VERSION")
	if version != "17" {
		t.Skip("Skipping Bios Tests for below 17G")
	}
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{ // Create user with ReadOnly role without user id
				PreConfig: func() {
					FunctionMocker = mockey.Mock(isServerGenerationSeventeenAndAbove).Return(true, nil).Build()
				},
				Config: testAccRedfishResourceUserConfigWithoutId(
					creds,
					"test2",
					"Test@123",
					"Operator",
					false),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("redfish_user_account.user_config", "username", "test2"),
				),
			},
		},
	})
	if FunctionMocker != nil {
		FunctionMocker.Release()
	}
}

// Test to create user with invalid role-id - Negative
func TestAccRedfishUserInvalid17G_basic(t *testing.T) {
	version := os.Getenv("TF_TESTING_REDFISH_VERSION")
	if version != "17" {
		t.Skip("Skipping Bios Tests for below 17G")
	}
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				PreConfig: func() {
					FunctionMocker = mockey.Mock(isServerGenerationSeventeenAndAbove).Return(true, nil).Build()
				},
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
	if FunctionMocker != nil {
		FunctionMocker.Release()
	}
}

// Test to create user with existing username - Negative
func TestAccRedfishUserExisting17G_basic(t *testing.T) {
	version := os.Getenv("TF_TESTING_REDFISH_VERSION")
	if version != "17" {
		t.Skip("Skipping Bios Tests for below 17G")
	}
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				PreConfig: func() {
					FunctionMocker = mockey.Mock(isServerGenerationSeventeenAndAbove).Return(true, nil).Build()
				},
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
	if FunctionMocker != nil {
		FunctionMocker.Release()
	}
}

// Test to update username to existing username - Negative
func TestAccRedfishUserUpdateInvalid17G_basic(t *testing.T) {
	version := os.Getenv("TF_TESTING_REDFISH_VERSION")
	if version != "17" {
		t.Skip("Skipping Bios Tests for below 17G")
	}
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
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
	if FunctionMocker != nil {
		FunctionMocker.Release()
	}
}

// Test to create user with Invalid ID - Negative
func TestAccRedfishUserUpdateInvalidId17G_basic(t *testing.T) {
	version := os.Getenv("TF_TESTING_REDFISH_VERSION")
	if version != "17" {
		t.Skip("Skipping Bios Tests for below 17G")
	}
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				PreConfig: func() {
					FunctionMocker = mockey.Mock(isServerGenerationSeventeenAndAbove).Return(true, nil).Build()
				},
				Config: testAccRedfishResourceUserConfig(
					creds,
					"newtest1",
					"Test@1234",
					"Administrator",
					true,
					"1"),
				ExpectError: regexp.MustCompile("user_id can vary between 3 to 31 only"),
			},
		},
	})
	if FunctionMocker != nil {
		FunctionMocker.Release()
	}
}

// Test to update user-id - Negative
func TestAccRedfishUserUpdateId17G_basic(t *testing.T) {
	version := os.Getenv("TF_TESTING_REDFISH_VERSION")
	if version != "17" {
		t.Skip("Skipping Bios Tests for below 17G")
	}
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{ // Create User
				PreConfig: func() {
					FunctionMocker = mockey.Mock(isServerGenerationSeventeenAndAbove).Return(true, nil).Build()
				},
				Config: testAccRedfishResourceUserConfig(
					creds,
					"test1",
					"Test@1234",
					"ReadOnly",
					false,
					userID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("redfish_user_account.user_config", "username", "test1"),
				),
			},
			{ // Update User
				Config: testAccRedfishResourceUserConfig(
					creds,
					"test1",
					"Test@1234",
					"ReadOnly",
					false,
					"1"),
				ExpectError: regexp.MustCompile("user_id cannot be updated"),
			},
		},
	})
	if FunctionMocker != nil {
		FunctionMocker.Release()
	}
}

// Test to update username - positive
func TestAccRedfishUserUpdateUser17G_basic(t *testing.T) {
	version := os.Getenv("TF_TESTING_REDFISH_VERSION")
	if version != "17" {
		t.Skip("Skipping Bios Tests for below 17G")
	}
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				PreConfig: func() {
					FunctionMocker = mockey.Mock(isServerGenerationSeventeenAndAbove).Return(true, nil).Build()
				},
				Config: testAccRedfishResourceUserConfig(
					creds,
					"test1",
					"Test@1234",
					"ReadOnly",
					false,
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
					"ReadOnly",
					false,
					userID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("redfish_user_account.user_config", "username", "test2"),
				),
			},
		},
	})
	if FunctionMocker != nil {
		FunctionMocker.Release()
	}
}

// Test to import user - positive
func TestAccRedfishUserImportUser17G_basic(t *testing.T) {
	version := os.Getenv("TF_TESTING_REDFISH_VERSION")
	if version != "17" {
		t.Skip("Skipping Bios Tests for below 17G")
	}
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				PreConfig: func() {
					FunctionMocker = mockey.Mock(isServerGenerationSeventeenAndAbove).Return(true, nil).Build()
				},
				Config: testAccRedfishResourceUserConfig(
					creds,
					"test2",
					"Test@1234",
					"ReadOnly",
					false,
					userID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("redfish_user_account.user_config", "username", "test2"),
				),
			},
			{
				PreConfig: func() {
					if FunctionMocker != nil {
						FunctionMocker.Release()
					}
					FunctionMocker = mockey.Mock(isServerGenerationSeventeenAndAbove).Return(true, nil).Build()
				},
				Config: testAccRedfishResourceUserConfig(
					creds,
					"test2",
					"Test@1234",
					"None",
					false,
					userID),
				ResourceName:  "redfish_user_account.user_config",
				ImportState:   true,
				ImportStateId: "{\"id\":\"15\",\"username\":\"" + creds.Username + "\",\"password\":\"" + creds.Password + "\",\"endpoint\":\"" + creds.Endpoint + "\",\"ssl_insecure\":true}",
				ExpectError:   nil,
			},
		},
	})
	if FunctionMocker != nil {
		FunctionMocker.Release()
	}
}

// Test to import user - negative
func TestAccRedfishUserImportUser17G_invalid(t *testing.T) {
	version := os.Getenv("TF_TESTING_REDFISH_VERSION")
	if version != "17" {
		t.Skip("Skipping Bios Tests for below 17G")
	}
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				PreConfig: func() {
					FunctionMocker = mockey.Mock(isServerGenerationSeventeenAndAbove).Return(true, nil).Build()
				},
				Config: testAccRedfishResourceUserConfig(
					creds,
					"test2",
					"Test@1234",
					"ReadOnly",
					false,
					userID),
				ResourceName:  "redfish_user_account.user_config",
				ImportState:   true,
				ImportStateId: "{\"id\":\"invalid\",\"username\":\"" + creds.Username + "\",\"password\":\"" + creds.Password + "\",\"endpoint\":\"" + creds.Endpoint + "\",\"ssl_insecure\":true}",
				ExpectError:   regexp.MustCompile("Error when retrieving accounts"),
			},
		},
	})
	if FunctionMocker != nil {
		FunctionMocker.Release()
	}
}

// validation tests - Negative
func TestAccRedfishUserValidation17G_basic(t *testing.T) {
	version := os.Getenv("TF_TESTING_REDFISH_VERSION")
	if version != "17" {
		t.Skip("Skipping Bios Tests for below 17G")
	}
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
					"ReadOnly",
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
	version := os.Getenv("TF_TESTING_REDFISH_VERSION")
	if version == "17" {
		t.Skip("Skipping Bios Tests for 17G")
	}
	serverAlias := "my-server-1"
	testUser, testUserPass, testUserRole := "testAlias", "Test@1234", "Administrator"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRedfishProviderWithServersConfig(serverAlias, creds.Username, creds.Password, creds.Endpoint) +
					testAccRedfishResourceUserConfig_alias(serverAlias, testUser, testUserPass, testUserRole, true, userID) +
					testAccRedfishResourceUserConfig_power_alias(serverAlias, "ForceRestart"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("redfish_user_account.user_config", "username", testUser),
				),
			},
		},
	})
}

func TestAccRedfishUserCreateCfgError(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				PreConfig: func() {
					FunctionMocker = mockey.Mock(NewConfig).Return(nil, fmt.Errorf("mock error")).Build()
				},
				Config: testAccRedfishResourceUserConfig(
					creds,
					"testerr",
					"TestPass@1234",
					"Operator",
					false,
					userID),
				ExpectError: regexp.MustCompile(`.*mock error*.`),
			},
			{
				PreConfig: func() {
					if FunctionMocker != nil {
						FunctionMocker.Release()
					}
					FunctionMocker = mockey.Mock(isServerGenerationSeventeenAndAbove).Return(nil, fmt.Errorf("Error retrieving the server generation")).Build()
				},
				Config: testAccRedfishResourceUserConfig(
					creds,
					"testerr",
					"TestPass@1234",
					"Operator",
					false,
					userID),
				ExpectError: regexp.MustCompile(`.*Error retrieving the server generation*.`),
			},
			{
				PreConfig: func() {
					if FunctionMocker != nil {
						FunctionMocker.Release()
					}
					FunctionMocker = mockey.Mock(GetAccountList).Return(nil, fmt.Errorf("Error when retrieving account list")).Build()
				},
				Config: testAccRedfishResourceUserConfig(
					creds,
					"testerr",
					"TestPass@1234",
					"Operator",
					false,
					userID),
				ExpectError: regexp.MustCompile(`.*Error when retrieving account list*.`),
			},
		},
	})
	if FunctionMocker != nil {
		FunctionMocker.Release()
	}
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
			endpoint = "%s"
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

func testAccRedfishResourceUserConfigWithoutId(testingInfo TestingServerCredentials,
	username string,
	password string,
	roleId string,
	enabled bool,
) string {
	return fmt.Sprintf(`

		resource "redfish_user_account" "user_config" {

		  redfish_server {
			user = "%s"
			password = "%s"
			endpoint = "%s"
			ssl_insecure = true
		  }

		  username = "%s"
		  password = "%s"
		  role_id = "%s"
		  enabled = %t
		}
		`,
		testingInfo.Username,
		testingInfo.Password,
		testingInfo.Endpoint,
		username,
		password,
		roleId,
		enabled,
	)
}

func testAccRedfishProviderWithServersConfig(serverAlias, username, password, endpoint string) string {
	return fmt.Sprintf(`
		locals {
			rack1 = {
				"%s" = {
					user         = "%s"
					password     = "%s"
					endpoint = "%s"
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

func testAccRedfishResourceUserConfig_alias(alias string, username string, password string, roleId string, enabled bool, userId string) string {
	return fmt.Sprintf(`

		resource "redfish_user_account" "user_config" {
		
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
		alias,
		username,
		password,
		roleId,
		enabled,
		userId,
	)
}

func testAccRedfishResourceUserConfig_power_alias(alias, powerAction string) string {
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
