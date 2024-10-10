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

func TestAccRedfishSystemBoot_fetch(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRedfishDatasourceSystemBootConfig(creds, "System.Embedded.1"),
			},
			{
				Config: testAccRedfishDatasourceSystemBootConfigBasic(creds),
			},
		},
	})
}

func TestAccRedfishSystemBoot_fetchInvalidID(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccRedfishDatasourceSystemBootConfig(creds, "invalid-id"),
				ExpectError: regexp.MustCompile("Error fetching computer system"),
			},
		},
	})
}

func testAccRedfishDatasourceSystemBootConfig(testingInfo TestingServerCredentials, id string) string {
	return fmt.Sprintf(`
	data "redfish_system_boot" "system_boot" {
		system_id = "%s"
		redfish_server {
		  user         = "%s"
		  password     = "%s"
		  endpoint     = "https://%s"
		  ssl_insecure = true
		}
	  }	  
	`,
		id,
		testingInfo.Username,
		testingInfo.Password,
		testingInfo.Endpoint,
	)
}

func testAccRedfishDatasourceSystemBootConfigBasic(testingInfo TestingServerCredentials) string {
	return fmt.Sprintf(`
	data "redfish_system_boot" "system_boot" {
		redfish_server {
		  user         = "%s"
		  password     = "%s"
		  endpoint     = "https://%s"
		  ssl_insecure = true
		}
	  }
	`,
		testingInfo.Username,
		testingInfo.Password,
		testingInfo.Endpoint,
	)
}
