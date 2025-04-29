/*
Copyright (c) 2023-2025 Dell Inc., or its subsidiaries. All Rights Reserved.

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

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// Test to create and update Simple update - Positive
func TestAccRedfishSimpleUpdate_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRedfishResourceUpdateConfig(
					creds,
					"HTTP",
					os.Getenv("TF_TESTING_FIRMWARE_IMAGE_LOCAL")),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("redfish_simple_update.update", "transfer_protocol", "HTTP"),
				),
			},
			{
				Config: testAccRedfishResourceUpdateConfig(
					creds,
					"HTTP",
					os.Getenv("TF_TESTING_FIRMWARE_IMAGE_HTTP")),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("redfish_simple_update.update", "transfer_protocol", "HTTP"),
				),
			},
			{
				Config: testAccRedfishResourceUpdateConfig(
					creds,
					"NFS",
					os.Getenv("TF_TESTING_FIRMWARE_IMAGE_NFS")),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("redfish_simple_update.update", "transfer_protocol", "NFS"),
				),
			},
		},
	})
}

// Test to update with invalid path and protocol - Negative
func TestAccRedfishSimpleUpdate_InvalidProto(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRedfishResourceUpdateConfig(
					creds,
					"HTTP",
					os.Getenv("TF_TESTING_FIRMWARE_IMAGE_INVALID")),
				ExpectError: regexp.MustCompile("please check the image path, download failed"),
			},
			{
				Config: testAccRedfishResourceUpdateConfig(
					creds,
					"FTP",
					os.Getenv("TF_TESTING_FIRMWARE_IMAGE_HTTP")),
				ExpectError: regexp.MustCompile("this transfer protocol is not available in this redfish instance"),
			},
			{
				Config: testAccRedfishResourceUpdateConfig(
					creds,
					"CIFS",
					os.Getenv("TF_TESTING_FIRMWARE_IMAGE_HTTP")),
				ExpectError: regexp.MustCompile("Transfer protocol not available in this implementation"),
			},
		},
	})
}

func testAccRedfishResourceUpdateConfig(testingInfo TestingServerCredentials,
	transferProtocol string,
	imagePath string,
) string {
	return fmt.Sprintf(`
		
		resource "redfish_simple_update" "update" {
		
		  redfish_server {
			user = "%s"
			password = "%s"
			endpoint = "%s"
			ssl_insecure = true
		  }
		  system_id = "System.Embedded.1"
		  transfer_protocol     = "%s"
		  target_firmware_image = "%s"
		  reset_type  = "ForceRestart"
		}
		`,
		testingInfo.Username,
		testingInfo.Password,
		testingInfo.Endpoint,
		transferProtocol,
		imagePath,
	)
}
