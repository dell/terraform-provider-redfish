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

func dependsOnUser() string {
	return `depends_on = [redfish_user_account.user_config]`
}

// Test to create and update redfish user - Positive
func TestAccRedfishUserPassword_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
				%s
				%s
				`,
					testAccRedfishResourceUserConfig(creds, "test", "Test@123", "Administrator", true, "15"), testAccRedfishResourceUserPasswordConfig(creds, "test", "Test@123", "Test@1234", dependsOnUser())),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("redfish_user_account.user_config", "username", "test"),
					resource.TestCheckResourceAttr("redfish_user_account_password.user", "new_password", "Test@1234"),
				),
			},
			{
				Config: fmt.Sprintf(`
				%s
				`,
					testAccRedfishResourceUserPasswordConfig(creds, "test1", "Test@1234", "Test@1235", "")),
				ExpectError: regexp.MustCompile(ServiceErrorMsg),
			},
			{
				Config: fmt.Sprintf(`
				%s
				`,
					testAccRedfishResourceUserPasswordConfig(creds, "", "xyz", "xyz@123", "")),
				ExpectError: regexp.MustCompile(ServiceErrorMsg),
			},
		},
	})
}

func testAccRedfishResourceUserPasswordConfig(
	testingInfo TestingServerCredentials,
	username string,
	old_password string,
	new_password string,
	depends string,
) string {
	return fmt.Sprintf(`
		
		resource "redfish_user_account_password" "user" {
			username     = "%s"
			endpoint     = "https://%s"
			ssl_insecure = true
			old_password     = "%s"
			new_password     = "%s"
			%s
		}
		`,
		username,
		testingInfo.Endpoint,
		old_password,
		new_password,
		depends,
	)
}
