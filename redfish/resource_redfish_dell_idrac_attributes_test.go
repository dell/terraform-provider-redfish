package redfish

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccRedfishIDRACAttributes_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccRedfishResourceIDracAttributesConfig(
					creds),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("redfish_dell_idrac_attributes.idrac", "attributes.Users.3.Enable", "Disabled"),
					resource.TestCheckResourceAttr("redfish_dell_idrac_attributes.idrac", "attributes.Time.1.Timezone", "CST6CDT"),
				),
			},
		},
	})
}

func TestAccRedfishIDRACAttributes_invalidAttribute(t *testing.T) {
	resource.Test(t, resource.TestCase{
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccRedfishResourceIDracAttributesConfigInvalid(
					creds),
				ExpectError: regexp.MustCompile(" there was an issue when creating/updating idrac attributes"),
			},
		},
	})
}

func testAccRedfishResourceIDracAttributesConfig(testingInfo TestingServerCredentials) string {
	return fmt.Sprintf(`
	resource "redfish_dell_idrac_attributes" "idrac" {
		redfish_server {
		  user         = "%s"
		  password     = "%s"
		  endpoint     = "%s"
		  ssl_insecure = true
		}
	  
		attributes = {
		  "Users.3.Enable"    		  = "Disabled"
		  "Users.3.UserName"  		  = "mike"
		  "Users.3.Password"  		  = "test1234"
		  "Users.3.Privilege" 		  = 511
		  "Time.1.Timezone"   		  = "CST6CDT",
		  "SysLog.1.PowerLogInterval" = 5,
		}
	  }
	  `,
		testingInfo.Username,
		testingInfo.Password,
		testingInfo.Endpoint,
	)
}

func testAccRedfishResourceIDracAttributesConfigInvalid(testingInfo TestingServerCredentials) string {
	return fmt.Sprintf(`
	resource "redfish_dell_idrac_attributes" "idrac" {
		redfish_server {
		  user         = "%s"
		  password     = "%s"
		  endpoint     = "%s"
		  ssl_insecure = true
		}
	  
		attributes = {
		  "Users.3.Enable"            = "Disabled"
		  "Users.3.UserName"          = "mike"
		  "Users.3.Password"          = "test1234"
		  "Users.3.Privilege"         = 511
		  "Time.1.Timezone"			  = "CST6CDT",
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
