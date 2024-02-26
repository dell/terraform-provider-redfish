package provider

import (
	"fmt"
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
