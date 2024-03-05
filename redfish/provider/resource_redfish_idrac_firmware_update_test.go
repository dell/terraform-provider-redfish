package provider

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

// test redfish bios settings
func TestAccRedfishIdracFirmwareUpdateResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRedfishIdracFirmwareUpdateCreate(
					creds),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("redfish_idrac_firmware_update.update", "apply_update", "false"),
					resource.TestCheckResourceAttr("redfish_idrac_firmware_update.update", "ip_address", "downloads.dell.com"),
				),
			},
			{
				Config: testAccRedfishIdracFirmwareUpdateReapply(
					creds),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("redfish_idrac_firmware_update.update2", "apply_update", "false"),
					resource.TestCheckResourceAttr("redfish_idrac_firmware_update.update2", "ip_address", "10.225.104.166"),
				),
			},
		},
	})
}

func TestAccRedfishIdracFirmwareUpdateResourceFail(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRedfishIdracFirmwareUpdateCreateError(
					creds),
				ExpectError: regexp.MustCompile(`.*The argument "ip_address" is required*.`),
			},
			{
				Config: testAccRedfishIdracFirmwareUpdateCreateError2(
					creds),
				ExpectError: regexp.MustCompile(`.*The argument "share_type" is required*.`),
			},
		},
	})
}

func testAccRedfishIdracFirmwareUpdateCreate(testingInfo TestingServerCredentials) string {
	return fmt.Sprintf(`

	resource "redfish_idrac_firmware_update" "update" {
	  
		redfish_server {
		  user         = "%s"
		  password     = "%s"
		  endpoint     = "https://%s"
		  ssl_insecure = true
		}
	  
		ip_address = "downloads.dell.com"
		share_type = "HTTP"
		// These two fields should are set to true by default. It will check the repository for any updates that are available for the server and apply those updates.
  		// If you do not want to apply the updates and just want to get the details for the updates available, set these fields to false.
		apply_update = false
		reboot_needed = false
	  }
		`,
		testingInfo.Username,
		testingInfo.Password,
		testingInfo.Endpoint,
	)
}

func testAccRedfishIdracFirmwareUpdateCreateError(testingInfo TestingServerCredentials) string {
	return fmt.Sprintf(`

	resource "redfish_idrac_firmware_update" "update" {
	  
		redfish_server {
		  user         = "%s"
		  password     = "%s"
		  endpoint     = "https://%s"
		  ssl_insecure = true
		}
	  
		share_type = "HTTP"
	  }
		`,
		testingInfo.Username,
		testingInfo.Password,
		testingInfo.Endpoint,
	)
}

func testAccRedfishIdracFirmwareUpdateCreateError2(testingInfo TestingServerCredentials) string {
	return fmt.Sprintf(`

	resource "redfish_idrac_firmware_update" "update" {
	  
		redfish_server {
		  user         = "%s"
		  password     = "%s"
		  endpoint     = "https://%s"
		  ssl_insecure = true
		}
		ip_address = "downloads.dell.com"
	  }
		`,
		testingInfo.Username,
		testingInfo.Password,
		testingInfo.Endpoint,
	)
}

func testAccRedfishIdracFirmwareUpdateReapply(testingInfo TestingServerCredentials) string {
	return fmt.Sprintf(`

	resource "redfish_idrac_firmware_update" "update2" {
	  
		redfish_server {
		  user         = "%s"
		  password     = "%s"
		  endpoint     = "https://%s"
		  ssl_insecure = true
		}
	  
		ip_address = "10.225.104.166"
		share_type = "HTTP"
		apply_update = false
		reboot_needed = false
	  }
		`,
		testingInfo.Username,
		testingInfo.Password,
		testingInfo.Endpoint,
	)
}
