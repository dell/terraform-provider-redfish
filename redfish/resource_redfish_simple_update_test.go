package redfish

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"os"
	"regexp"
	"testing"
)

// Test to create and update Simple update - Positive
func TestAccRedfishSimpleUpdate_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		Providers: testAccProviders,
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
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccRedfishResourceUpdateConfig(
					creds,
					"HTTP",
					os.Getenv("TF_TESTING_FIRMWARE_IMAGE_INVALID")),
				ExpectError: regexp.MustCompile("couldn't open FW file to upload - error when opening"),
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
	imagePath string) string {
	return fmt.Sprintf(`
		
		resource "redfish_simple_update" "update" {
		
		  redfish_server {
			user = "%s"
			password = "%s"
			endpoint = "https://%s"
			ssl_insecure = true
		  }

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
