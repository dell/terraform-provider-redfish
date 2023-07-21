package redfish

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccRedfishVirtualMedia_fetch(t *testing.T) {
	resource.Test(t, resource.TestCase{
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccRedfishDatasourceVirtualMediaConfig(creds),
			},
		},
	})
}

func testAccRedfishDatasourceVirtualMediaConfig(testingInfo TestingServerCredentials) string {
	return fmt.Sprintf(`
	data "redfish_virtual_media" "vm" {
	  
		redfish_server {
		  user         = "%s"
		  password     = "%s"
		  endpoint     = "https://%s"
		  ssl_insecure = true
		}
	  }
	  
	  output "virtual_media" {
		 value = data.redfish_virtual_media.vm
		 sensitive = true
	  }
	`,
		testingInfo.Username,
		testingInfo.Password,
		testingInfo.Endpoint,
	)
}
