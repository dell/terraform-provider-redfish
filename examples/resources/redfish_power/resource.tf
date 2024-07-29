/*
Copyright (c) 2021-2024 Dell Inc., or its subsidiaries. All Rights Reserved.

Licensed under the Mozilla Public License Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://mozilla.org/MPL/2.0/


Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

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
  | GracefulRestart  | Shut down gracefully and restart the system .                                           |
  | PushPowerButton  | Alters the power state of the system. If the system is Off, it powers On and vice-versa |
  | Nmi              | Turns the unit on in troubleshooting mode.                                              |
  */

  desired_power_action = "ForceRestart"

  // The maximum amount of time to wait for the server to enter the correct power state before
  // giving up in seconds
  maximum_wait_time = 120

  // The frequency with which to check the server's power state in seconds
  check_interval = 10

  // by default, the resource uses the first system
  # system_id = "System.Embedded.1"
}

output "current_power_state" {
  value = redfish_power.system_power
}
