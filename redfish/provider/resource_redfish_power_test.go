package provider

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"os"
	"regexp"
	"testing"
)

// redfish.Power represents a concrete Go type that represents an API resource
func TestAccRedfishPowerT1(t *testing.T) {

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRedfishResourcePowerConfig(creds, "On", 120, 10),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("redfish_power.system_power", "power_state", "On"),
				),
			},
			{
				Config: testAccRedfishResourcePowerConfig(creds, "ForceOn", 120, 10),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("redfish_power.system_power", "power_state", "On"),
				),
			},
			{
				Config: testAccRedfishResourcePowerConfig(creds, "GracefulShutdown", 120, 10),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("redfish_power.system_power", "power_state", "Off"),
				),
			},
			{
				Config: testAccRedfishResourcePowerConfig(creds, "On", 120, 10),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("redfish_power.system_power", "power_state", "On"),
				),
			},
			{
				Config: testAccRedfishResourcePowerConfig(creds, "ForceOff", 120, 10),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("redfish_power.system_power", "power_state", "Off"),
				),
			},
			{
				Config: testAccRedfishResourcePowerConfig(creds, "ForceRestart", 120, 10),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("redfish_power.system_power", "power_state", "Reset_On"),
				),
			},
			{
				Config: testAccRedfishResourcePowerConfig(creds, "PowerCycle", 120, 10),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("redfish_power.system_power", "power_state", "Reset_On"),
				),
			},
			{
				Config: testAccRedfishResourcePowerConfig(creds, "PushPowerButton", 120, 10),
			},
			{
				Config: testAccRedfishResourcePowerConfig(creds, "PushPowerButton", 125, 12),
			},
		},
	})
}

// redfish.Power represents a concrete Go type that represents an API resource
func TestAccRedfishPower_Invalid(t *testing.T) {
	os.Setenv("TF_ACC", "1")
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRedfishResourcePowerConfig1(
					creds,
					"nil"),
				ExpectError: regexp.MustCompile("desired_power_action value must be one of"),
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
		
		  redfish_server = {
			user = "%s"
			password = "%s"
			endpoint = "https://%s"
			validate_cert = false
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

func testAccRedfishResourcePowerConfig1(testingInfo TestingServerCredentials,
	desiredPowerAction string) string {
	return fmt.Sprintf(`

		resource "redfish_power" "system_power" {

			redfish_server = {
				user = "%s"
				password = "%s"
				endpoint = "https://%s"
				validate_cert = false
			  }

		  desired_power_action = "%s"
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
	)
}
