package redfish

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"testing"
)

// redfish.Power represents a concrete Go type that represents an API resource
func TestAccRedfishBiosDataSource_basic(t *testing.T) {

	resource.Test(t, resource.TestCase{
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccRedfishDataSourceBiosConfig(creds),
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
