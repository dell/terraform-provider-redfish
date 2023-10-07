package redfish

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
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
				ExpectError: regexp.MustCompile(" Could not find a ComputerSystem resource with resource ID"),
			},
		},
	})
}

func testAccRedfishDatasourceSystemBootConfig(testingInfo TestingServerCredentials, id string) string {
	return fmt.Sprintf(`
	data "redfish_system_boot" "system_boot" {
		resource_id = "%s"
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
