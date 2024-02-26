package provider

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

// Test to create manager reset resource with invalid reset type- Negative
func TestAccRedfishManagerReset_Invalid_ResetType_Negative(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccRedfishResourceManagerResetConfig(creds, "iDRAC.Embedded.1", "On"),
				ExpectError: regexp.MustCompile("Invalid Attribute Value Match"),
			},
		},
	})
}

// Test to create manager reset resource with invalid manager id- Negative
func TestAccRedfishManagerReset_Invalid_ManagerID_Negative(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccRedfishResourceManagerResetConfig(creds, "iDRAC.Embedded.0", "GracefulRestart"),
				ExpectError: regexp.MustCompile("Invalid Manager ID provided"),
			},
		},
	})
}

// Test to update manager reset resource with invalid maanger id- Negative
func TestAccRedfishManagerReset_Update_Negative(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRedfishResourceManagerResetConfig(creds, "iDRAC.Embedded.1", "GracefulRestart"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("redfish_manager_reset.manager_reset", "id", "iDRAC.Embedded.1"),
				),
			},
			{
				Config:      testAccRedfishResourceManagerResetConfig(creds, "iDRAC.Embedded", "GracefulRestart"),
				ExpectError: regexp.MustCompile("Invalid Manager ID provided"),
			},
		},
	})
}

// Test to perform manager reset
func TestAccRedfishManagerReset_Create(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRedfishResourceManagerResetConfig(creds, "iDRAC.Embedded.1", "GracefulRestart"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("redfish_manager_reset.manager_reset", "id", "iDRAC.Embedded.1"),
				),
			},
		},
	})
}

func testAccRedfishResourceManagerResetConfig(testingInfo TestingServerCredentials,
	managerID string,
	resetType string,
) string {
	return fmt.Sprintf(`
		
	resource "redfish_manager_reset" "manager_reset" {
		redfish_server {
		  user         = "%s"
		  password     = "%s"
		  endpoint     = "https://%s"
		  ssl_insecure = true
		}
	  
		id = "%s"
		reset_type = "%s"
	}
		`,
		testingInfo.Username,
		testingInfo.Password,
		testingInfo.Endpoint,
		managerID,
		resetType,
	)
}
