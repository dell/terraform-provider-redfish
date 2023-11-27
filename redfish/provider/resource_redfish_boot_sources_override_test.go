package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

// test redfish Boot Order
func TestAccRedfishBootSourceOverride_basic(t *testing.T) {

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRedfishResourceBootSourceLegacyconfig(creds),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("redfish_boot_source_override.boot", "boot_source_override_mode", "Legacy"),
				),
			},
		},
	})
}

func TestAccRedfishBootSourceOverride_updated(t *testing.T) {

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRedfishResourceBootSourceUEFIconfig(creds),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("redfish_boot_source_override.boot", "boot_source_override_mode", "UEFI"),
				),
			},
			{
				Config: testAccRedfishResourceBootSourceResetType(creds),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("redfish_boot_source_override.boot", "boot_source_override_mode", "UEFI"),
				),
			},
		},
	})
}

func testAccRedfishResourceBootSourceLegacyconfig(testingInfo TestingServerCredentials) string {
	return fmt.Sprintf(`

	resource "redfish_boot_source_override" "boot" {
		redfish_server {
			user = "%s"
			password = "%s"
			endpoint = "https://%s"
			ssl_insecure = true
		}
	   
		boot_source_override_enabled = "Once"
		boot_source_override_target = "Pxe"
		boot_source_override_mode = "Legacy"
		reset_type    = "GracefulRestart"
	}	  
	`,
		testingInfo.Username,
		testingInfo.Password,
		testingInfo.Endpoint,
	)
}

func testAccRedfishResourceBootSourceUEFIconfig(testingInfo TestingServerCredentials) string {
	return fmt.Sprintf(`

	resource "redfish_boot_source_override" "boot" {
	  
		redfish_server {
			user = "%s"
			password = "%s"
			endpoint = "https://%s"
			ssl_insecure = true
		}
	   
		boot_source_override_enabled = "Once"
		boot_source_override_target = "UefiTarget"
		boot_source_override_mode = "UEFI"
		reset_type    = "GracefulRestart"
	}	  
	`,
		testingInfo.Username,
		testingInfo.Password,
		testingInfo.Endpoint,
	)
}

func testAccRedfishResourceBootSourceResetType(testingInfo TestingServerCredentials) string {
	return fmt.Sprintf(`

	resource "redfish_boot_source_override" "boot" {
	  
		redfish_server {
			user = "%s"
			password = "%s"
			endpoint = "https://%s"
			ssl_insecure = true
		}
	   
		boot_source_override_enabled = "Once"
		boot_source_override_target = "UefiTarget"
		boot_source_override_mode = "UEFI"
		reset_type    = "ForceRestart"
	}	  
	`,
		testingInfo.Username,
		testingInfo.Password,
		testingInfo.Endpoint,
	)
}
