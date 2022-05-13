terraform {
  required_providers {
    redfish = {
      version = "~> 0.2.0"
      source  = "registry.terraform.io/dell/redfish"
    }
  }
}

// For servers without a uniquely defined username/password these values will be used
provider "redfish" {
  user     = "root"
  password = "password"
}

resource "redfish_power" "system_power" {
  for_each = var.rack1

  redfish_server {
    user         = each.value.user
    password     = each.value.password
    endpoint     = each.value.endpoint
    ssl_insecure = each.value.ssl_insecure
  }

  // The valid options are defined below.
  // Taken from the Redfish specification at: https://redfish.dmtf.org/schemas/DSP2046_2019.4.html
  /*
  | string           | Description                                                                             |
  |------------------|-----------------------------------------------------------------------------------------|
  | ForceOff         | Turn off the unit immediately (non-graceful shutdown).                                  |
  | ForceOn          | Turn on the unit immediately.                                                           |
  | ForceRestart     | Shut down immediately and non-gracefully and restart the system.                        |
  | GracefulShutdown | Shut down gracefully and power off.                                                     |
  | On               | Turn on the unit.                                                                       |
  | PowerCycle       | Power cycle the unit.                                                                   |
  */
  desired_power_action = "ForceRestart"

  // The maximum amount of time to wait for the server to enter the correct power state before
  // giving up in seconds
  maximum_wait_time = 120

  // The frequency with which to check the server's power state in seconds
  check_interval = 10
}

output "current_power_state" {
  value = redfish_power.system_power
}
