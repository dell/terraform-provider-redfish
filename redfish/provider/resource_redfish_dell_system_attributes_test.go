package provider

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccRedfishSystemAttributesBasic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRedfishResourceSystemAttributesConfig(creds),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("redfish_dell_system_attributes.system", "attributes.SupportInfo.1.Outsourced", "Yes"),
					resource.TestCheckResourceAttr("redfish_dell_system_attributes.system", "attributes.ServerPwr.1.PSPFCEnabled", "Disabled"),
				),
			},
		},
	})
}

func TestAccRedfishSystemAttributesInvalidAttribute(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRedfishResourceSystemConfigInvalid(
					creds),
				ExpectError: regexp.MustCompile("there was an issue when creating/updating System attributes"),
			},
		},
	})
}

func TestAccRedfishSystemAttributesUpdate(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRedfishResourceSystemAttributesConfig(creds),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("redfish_dell_system_attributes.system", "attributes.SupportInfo.1.Outsourced", "Yes"),
					resource.TestCheckResourceAttr("redfish_dell_system_attributes.system", "attributes.ServerPwr.1.PSPFCEnabled", "Disabled"),
				),
			},
			{
				Config: testAccRedfishResourceSystemAttributesUpdateConfig(creds),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("redfish_dell_system_attributes.system", "attributes.SupportInfo.1.Outsourced", "No"),
					resource.TestCheckResourceAttr("redfish_dell_system_attributes.system", "attributes.ServerPwr.1.PSPFCEnabled", "Enabled"),
				),
			},
		},
	})
}

func TestAccRedfishSystemAttributesImport(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `resource "redfish_dell_system_attributes" "system" {
				}`,
				ResourceName:  "redfish_dell_system_attributes.system",
				ImportState:   true,
				ImportStateId: "{\"username\":\"" + creds.Username + "\",\"password\":\"" + creds.Password + "\",\"endpoint\":\"https://" + creds.Endpoint + "\",\"ssl_insecure\":true}",
				ExpectError:   nil,
			},
		},
	})
}

func testAccRedfishResourceSystemAttributesConfig(testingInfo TestingServerCredentials) string {
	return fmt.Sprintf(`
	resource "redfish_dell_system_attributes" "system" {
		redfish_server {
		  user         = "%s"
		  password     = "%s"
		  endpoint     = "https://%s"
		  ssl_insecure = true
		}
	  
		attributes = {
			"ServerPwr.1.PSPFCEnabled" = "Disabled"
			"SupportInfo.1.Outsourced" = "Yes"
		}
	  }
	  `,
		testingInfo.Username,
		testingInfo.Password,
		testingInfo.Endpoint,
	)
}

func testAccRedfishResourceSystemAttributesUpdateConfig(testingInfo TestingServerCredentials) string {
	return fmt.Sprintf(`
	resource "redfish_dell_system_attributes" "system" {
		redfish_server {
		  user         = "%s"
		  password     = "%s"
		  endpoint     = "https://%s"
		  ssl_insecure = true
		}

		attributes = {
			"ServerPwr.1.PSPFCEnabled" = "Enabled"
			"SupportInfo.1.Outsourced" = "No"
		}
	  }
	  `,
		testingInfo.Username,
		testingInfo.Password,
		testingInfo.Endpoint,
	)
}

func testAccRedfishResourceSystemConfigInvalid(testingInfo TestingServerCredentials) string {
	return fmt.Sprintf(`
	resource "redfish_dell_system_attributes" "system" {
		redfish_server {
		  user         = "%s"
		  password     = "%s"
		  endpoint     = "https://%s"
		  ssl_insecure = true
		}
	  
		attributes = {
			"ServerPwr.1.PSPFCEnabled" = "Disabled",
			"SupportInfo.1.Outsourced" = "Yes",
		  	"SysLog.1.PowerLogInterval" = 5,
		  	"InvalidAttribute" 		  = "invalid",
		}
	  }
	  `,
		testingInfo.Username,
		testingInfo.Password,
		testingInfo.Endpoint,
	)
}
