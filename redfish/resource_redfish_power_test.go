package redfish

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"testing"
)

// redfish.Power represents a concrete Go type that represents an API resource
func TestAccRedfishPower_basic(t *testing.T) {

	resource.Test(t, resource.TestCase{
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccRedfishResourcePowerConfig(
					creds,
					"On",
					120,
					10),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("redfish_power.system_power", "power_state", "On"),
				),
			},
			{
				Config: testAccRedfishResourcePowerConfig(
					creds,
					"ForceOn",
					120,
					10),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("redfish_power.system_power", "power_state", "On"),
				),
			},
			{
				Config: testAccRedfishResourcePowerConfig(
					creds,
					"ForceOff",
					120,
					10),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("redfish_power.system_power", "power_state", "Off"),
				),
			},
			{
				Config: testAccRedfishResourcePowerConfig(
					creds,
					"ForceRestart",
					120,
					10),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("redfish_power.system_power", "power_state", "Reset_On"),
				),
				ExpectNonEmptyPlan: true,
			},
			{
				Config: testAccRedfishResourcePowerConfig(
					creds,
					"GracefulShutdown",
					120,
					10),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("redfish_power.system_power", "power_state", "Off"),
				),
			},
			{
				Config: testAccRedfishResourcePowerConfig(
					creds,
					"PowerCycle",
					120,
					10),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("redfish_power.system_power", "power_state", "Reset_On"),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccRedfishResourcePowerConfig(testingInfo TestingServerCredentials,
	desiredPowerAction string,
	maximumWaitTime int,
	checkInterval int) string {
	return fmt.Sprintf(`
		
		resource "redfish_power" "system_power" {
		
		  redfish_server {
			user = "%s"
			password = "%s"
			endpoint = "https://%s"
			ssl_insecure = true
		  }

		  desired_power_action = "%s"
		  maximum_wait_time = %d
		  check_interval = %d
		}
		
		output "current_power_state" {
		  value = redfish_power.system_power
          sensitive = true
		}
		`,
		testingInfo.Username,
		testingInfo.Password,
		testingInfo.Endpoint,
		desiredPowerAction,
		maximumWaitTime,
		checkInterval,
	)
}
