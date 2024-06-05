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
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

// redfish.Power represents a concrete Go type that represents an API resource
func TestAccRedfishBiosDataSource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRedfishDataSourceBiosConfig(creds),
			},
			{
				Config: testAccRedfishDataSourceBootOptions(creds, os.Getenv("TF_TESTING_BOOT_OPTION_REFERENCE"), true, os.Getenv("TF_TESTING_DISPLAY_NAME"), os.Getenv("TF_TESTING_ID"), os.Getenv("TF_TESTING_NAME"), os.Getenv("TF_TESTING_UEFI_DEVICE_PATH")),
			},
		},
	})
}

func testAccRedfishDataSourceBiosConfig(testingInfo TestingServerCredentials) string {
	return fmt.Sprintf(`
		
		data "redfish_bios" "bios" {		
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

func testAccRedfishDataSourceBootOptions(testingInfo TestingServerCredentials, bootOptionReference string, bootOptionEnabled bool, displayName string, id string, name string, uefiDevicePath string) string {
	return fmt.Sprintf(`
		
		data "redfish_bios" "bios" {		
		  redfish_server {
			user = "%s"
			password = "%s"
			endpoint = "https://%s"
			ssl_insecure = true
		  }
		  boot_options = [{boot_option_reference="%s", boot_option_enabled=%t", display_name="%s", id="%s", name="%s", uefi_device_path="%s"}]
		}
		`,
		testingInfo.Username,
		testingInfo.Password,
		testingInfo.Endpoint,
		bootOptionReference,
		bootOptionEnabled,
		displayName,
		id,
		name,
		uefiDevicePath,
	)
}
