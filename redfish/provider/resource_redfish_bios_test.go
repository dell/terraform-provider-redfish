package provider

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

// test redfish bios settings
func TestAccRedfishBios_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRedfishResourceBiosConfigOn(
					creds),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("redfish_bios.bios", "attributes.NumLock", "On"),
				),
			},
			{
				Config: testAccRedfishResourceBiosConfigOff(
					creds),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("redfish_bios.bios", "attributes.NumLock", "Off"),
				),
			},
		},
	})
}

func TestAccRedfishBios_InvalidSettings(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRedfishResourceBiosConfigInvalidSettingsApplyTime(
					creds),
				ExpectError: regexp.MustCompile("Attribute settings_apply_time value must be one of"),
			},
		},
	})
}

func TestAccRedfishBios_InvalidAttributes(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRedfishResourceBiosConfigInvalidAttributes(
					creds),
				ExpectError: regexp.MustCompile("Attribute settings_apply_time value must be one of"),
			},
		},
	})
}

// Test to import bios - positive
func TestAccRedfishBios_Import(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRedfishResourceBiosConfigOn(
					creds),
				ResourceName:  "redfish_bios.bios",
				ImportState:   true,
				ImportStateId: "{\"username\":\"" + creds.Username + "\",\"password\":\"" + creds.Password + "\",\"endpoint\":\"https://" + creds.Endpoint + "\",\"ssl_insecure\":true}",
				ExpectError:   nil,
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
			"AcPwrRcvryUserDelay" = 70
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
