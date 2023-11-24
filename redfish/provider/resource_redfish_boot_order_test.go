package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

// test redfish Boot Order
func TestAccRedfishBootOrder_basic(t *testing.T) {

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				// Config: testAccRedfishResourceBootOrder(creds, `["NIC.Integrated.1-1-1","HardDisk.List.1-1"]`),
				Config: testAccRedfishResourceBootOrder(creds, `["Boot0003","Boot0004","Boot0005"]`),
			},
		},
	})
}

func TestAccRedfishBootOrderOptions_basic(t *testing.T) {

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				// Config: testAccRedfishResourceBootOptions(creds, os.Getenv("TF_TESTING_BOOT_OPTION_REFERENCE"), true),
				Config: testAccRedfishResourceBootOptions(creds),
			},
			{
				Config: testAccRedfishResourceBootOptions_update(creds),
			},
		},
	})
}

func testAccRedfishResourceBootOrder(testingInfo TestingServerCredentials, bootOrder string) string {
	return fmt.Sprintf(`

	resource "redfish_boot_order" "boot" {
		redfish_server {
			user = "%s"
			password = "%s"
			endpoint = "https://%s"
			ssl_insecure = true
		}
	   
		reset_type="ForceRestart"
		boot_order=%s
	}	  
	`,
		testingInfo.Username,
		testingInfo.Password,
		testingInfo.Endpoint,
		bootOrder,
	)
}

func testAccRedfishResourceBootOptions(testingInfo TestingServerCredentials) string {
	return fmt.Sprintf(`

	resource "redfish_boot_order" "boot" {
		redfish_server {
			user = "%s"
			password = "%s"
			endpoint = "https://%s"
			ssl_insecure = true
		}
	   
		reset_type="ForceRestart"   
		boot_options = [{boot_option_reference= "Boot0003", boot_option_enabled=true}]
	}	  
	`,
		testingInfo.Username,
		testingInfo.Password,
		testingInfo.Endpoint,
	)
}

func testAccRedfishResourceBootOptions_update(testingInfo TestingServerCredentials) string {
	return fmt.Sprintf(`

	resource "redfish_boot_order" "boot" {
		redfish_server {
			user = "%s"
			password = "%s"
			endpoint = "https://%s"
			ssl_insecure = true
		}
	   
		reset_type="ForceRestart"   
		boot_options = [{boot_option_reference= "Boot0003", boot_option_enabled=false}]
	}	  
	`,
		testingInfo.Username,
		testingInfo.Password,
		testingInfo.Endpoint,
	)
}
