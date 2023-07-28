package redfish

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"testing"
)

// Test case for Firmware DataSource
func TestAccRedfishFirmwareDataSource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccRedfishDataSourceFirmwareConfig(creds),
			},
		},
	})
}

func testAccRedfishDataSourceFirmwareConfig(testingInfo TestingServerCredentials) string {
	return fmt.Sprintf(`
		
		data "redfish_firmware_inventory" "inventory" {
		
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
