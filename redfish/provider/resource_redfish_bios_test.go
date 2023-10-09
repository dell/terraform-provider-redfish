package provider

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"regexp"
	"testing"
)

// redfish.Power represents a concrete Go type that represents an API resource
func TestAccRedfishBios_basic(t *testing.T) {

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRedfishResourceBiosConfigOn(
					creds),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("redfish_bios.bios", "attributes.NumLock", "On"),
				),
			},
		},
	})
}

func TestAccRedfishBios_NumLockOff(t *testing.T) {

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRedfishResourceBiosConfigOff(
					creds),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("redfish_bios.bios", "attributes.NumLock", "Off"),
				),
			},
		},
	})
}

func TestAccRedfishBios_basic_InvalidSettings(t *testing.T) {

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRedfishResourceBiosConfigInvalidSettingsApplyTime(
					creds),
				ExpectError: regexp.MustCompile(" expected settings_apply_time to be one of "),
			},
		},
	})
}

func TestAccRedfishBios_basic_InvalidAttributes(t *testing.T) {

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRedfishResourceBiosConfigInvalidAttributes(
					creds),
				ExpectError: regexp.MustCompile(" expected settings_apply_time to be one of "),
			},
		},
	})
}

func testAccRedfishResourceBiosConfigOn(testingInfo TestingServerCredentials) string {
	return fmt.Sprintf(`

		resource "redfish_bios" "bios"  {
		
		  redfish_server {
			user = "%s"
			password = "%s"
			endpoint = "https://%s"
			ssl_insecure = true
		  }

		  attributes = {
			"NumLock" = "On"
		  }
		  reset_type = "ForceRestart"
		}
		`,
		testingInfo.Username,
		testingInfo.Password,
		testingInfo.Endpoint,
	)
}

func testAccRedfishResourceBiosConfigOff(testingInfo TestingServerCredentials) string {
	return fmt.Sprintf(`

		resource "redfish_bios" "bios"  {
		
		  redfish_server {
			user = "%s"
			password = "%s"
			endpoint = "https://%s"
			ssl_insecure = true
		  }

		  attributes = {
			"NumLock" = "Off"
		  }
		  reset_type = "ForceRestart"
   		  bios_job_timeout = 1200
		  reset_timeout = 120
		}
		`,
		testingInfo.Username,
		testingInfo.Password,
		testingInfo.Endpoint,
	)
}

func testAccRedfishResourceBiosConfigInvalidSettingsApplyTime(testingInfo TestingServerCredentials) string {
	return fmt.Sprintf(`

		resource "redfish_bios" "bios"  {
		
		  redfish_server {
			user = "%s"
			password = "%s"
			endpoint = "https://%s"
			ssl_insecure = true
		  }

		  attributes = {
		  }
		  settings_apply_time = "random"
		}
		`,
		testingInfo.Username,
		testingInfo.Password,
		testingInfo.Endpoint,
	)
}

func testAccRedfishResourceBiosConfigInvalidAttributes(testingInfo TestingServerCredentials) string {
	return fmt.Sprintf(`

		resource "redfish_bios" "bios"  {
		
		  redfish_server {
			user = "%s"
			password = "%s"
			endpoint = "https://%s"
			ssl_insecure = true
		  }

		  attributes = {
		  }
		  settings_apply_time = "ForceRestart"
		}
		`,
		testingInfo.Username,
		testingInfo.Password,
		testingInfo.Endpoint,
	)
}
