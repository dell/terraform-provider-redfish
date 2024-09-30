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
	"os"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

// redfish.Power represents a concrete Go type that represents an API resource
func TestAccRedfishDirectoryServiceAuthProviderDatasource_basic(t *testing.T) {
	os.Setenv("TF_ACC", "1")
	dsAuthProviderDatasourceName := "data.redfish_directory_service_auth_provider.ds_auth"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRedfishDataSourceNICConfig(creds),
			},
			{
				Config: testAccRedfishDatasourceDirectoryServiceAuthProviderConfig(creds),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(dsAuthProviderDatasourceName, "directory_service_auth_provider.id", "RemoteAccountService"),
					resource.TestMatchResourceAttr(dsAuthProviderDatasourceName, "directory_service_auth_provider.odata_id", regexp.MustCompile(`.*AccountService`)),
				),
			},
		},
	})
}

func testAccRedfishDatasourceDirectoryServiceAuthProviderConfig(testingInfo TestingServerCredentials) string {
	return fmt.Sprintf(`
		data "redfish_directory_service_auth_provider" "ds_auth" {		
		  redfish_server {
			user = "%s"
			password = "%s"
			endpoint = "https://%s"
			ssl_insecure = true
		  }
		}
		`,
		testingInfo.Username,
		testingInfo.Password,
		testingInfo.Endpoint,
	)
}
