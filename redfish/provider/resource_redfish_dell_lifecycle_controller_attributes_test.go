package provider

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccRedfishLCAttributesBasic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRedfishResourceLCAttributesConfig(creds),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("redfish_dell_lc_attributes.lc", "attributes.LCAttributes.1.IgnoreCertWarning", "On"),
					resource.TestCheckResourceAttr("redfish_dell_lc_attributes.lc", "attributes.LCAttributes.1.CollectSystemInventoryOnRestart", "Disabled"),
				),
			},
		},
	})
}

func TestAccRedfishLCAttributesInvalidAttribute(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRedfishResourceLCConfigInvalid(
					creds),
				ExpectError: regexp.MustCompile("there was an issue when creating/updating LC attributes"),
			},
		},
	})
}

func TestAccRedfishLCAttributeImport(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `resource "redfish_dell_lc_attributes" "lc" {
				}`,
				ResourceName:  "redfish_dell_lc_attributes.lc",
				ImportState:   true,
				ImportStateId: "{\"username\":\"" + creds.Username + "\",\"password\":\"" + creds.Password + "\",\"endpoint\":\"https://" + creds.Endpoint + "\",\"ssl_insecure\":true}",
				ExpectError:   nil,
			},
		},
	})
}

func TestAccRedfishLCAttributeImportByFilter(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `resource "redfish_dell_lc_attributes" "lc" {
				}`,
				ResourceName:  "redfish_dell_lc_attributes.lc",
				ImportState:   true,
				ImportStateId: "{\"attributes\":[\"LCAttributes.1.CollectSystemInventoryOnRestart\"],\"username\":\"" + creds.Username + "\",\"password\":\"" + creds.Password + "\",\"endpoint\":\"https://" + creds.Endpoint + "\",\"ssl_insecure\":true}",
				ExpectError:   nil,
			},
		},
	})
}

func testAccRedfishResourceLCAttributesConfig(testingInfo TestingServerCredentials) string {
	return fmt.Sprintf(`
	resource "redfish_dell_lc_attributes" "lc" {
		redfish_server {
		  user         = "%s"
		  password     = "%s"
		  endpoint     = "https://%s"
		  ssl_insecure = true
		}
	  
		attributes = {
			"LCAttributes.1.CollectSystemInventoryOnRestart" = "Disabled"
			"LCAttributes.1.IgnoreCertWarning" = "On"
		}
	  }
	  `,
		testingInfo.Username,
		testingInfo.Password,
		testingInfo.Endpoint,
	)
}

func testAccRedfishResourceLCConfigInvalid(testingInfo TestingServerCredentials) string {
	return fmt.Sprintf(`
	resource "redfish_dell_lc_attributes" "lc" {
		redfish_server {
		  user         = "%s"
		  password     = "%s"
		  endpoint     = "https://%s"
		  ssl_insecure = true
		}
	  
		attributes = {
			"LCAttributes.1.CollectSystemInventoryOnRestart" = "Disabled",
			"LCAttributes.1.IgnoreCertWarning" = "On",
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
