package provider

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"regexp"
	"testing"
)

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
					"test1234",
					"Operator",
					true,
					"15"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("redfish_user_account.user_config", "username", "test1"),
				),
			},
			{
				Config: testAccRedfishResourceUserConfig(
					creds,
					"test1",
					"test1234",
					"None",
					false,
					"15"),
				Check: resource.ComposeTestCheckFunc(
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
					"test1234",
					"Admin",
					false,
					"15"),
				ExpectError: regexp.MustCompile(" expected role_id to be one of "),
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
					"xyz123",
					"Administrator",
					true,
					"15"),
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
					"test1234",
					"Administrator",
					true,
					"15"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("redfish_user_account.user_config", "username", "test1"),
				),
			},
			{
				Config: testAccRedfishResourceUserConfig(
					creds,
					"root",
					"test1234",
					"Administrator",
					false,
					"15"),
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
					"test1234",
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
					"test1234",
					"Administrator",
					true,
					"15"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("redfish_user_account.user_config", "username", "test1"),
				),
			},
			{
				Config: testAccRedfishResourceUserConfig(
					creds,
					"test1",
					"test1234",
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
					"test1234",
					"Administrator",
					true,
					"15"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("redfish_user_account.user_config", "username", "test1"),
				),
			},
			{
				Config: testAccRedfishResourceUserConfig(
					creds,
					"test2",
					"test1234",
					"Administrator",
					false,
					"15"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("redfish_user_account.user_config", "username", "test2"),
				),
			},
		},
	})
}

func testAccRedfishResourceUserConfig(testingInfo TestingServerCredentials,
	username string,
	password string,
	roleId string,
	enabled bool,
	userId string) string {
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
