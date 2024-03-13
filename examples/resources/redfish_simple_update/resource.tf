/*
Copyright (c) 2023-2024 Dell Inc., or its subsidiaries. All Rights Reserved.

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

resource "redfish_simple_update" "update" {
  for_each = var.rack1

  redfish_server {
    user         = each.value.user
    password     = each.value.password
    endpoint     = each.value.endpoint
    ssl_insecure = each.value.ssl_insecure
  }

  /* The network protocols and image for firmware update
   List of possible value: [
    CIFS, FTP, SFTP, HTTP, HTTPS, 
    NSF, SCP, TFTP, OEM, NFS
   ]
   Currently only HTTP, HTTPS and NFS are supported with local file path or HTTP(s)/NFS link
  */
  transfer_protocol     = "HTTP"
  target_firmware_image = "/home/mikeletux/Downloads/BIOS_FXC54_WN64_1.15.0.EXE"
  
  /* Reset parameters to be applied when upgrade is completed
     list of possible value:
      [ ForceRestart, GracefulRestart, PowerCycle]
  */
  reset_type    = "ForceRestart"
  reset_timeout = 120 // If not set, by default will be 120s
  // The maximum amount of time to wait for the simple update job to be completed
  simple_update_job_timeout = 1200 // If not set, by default will be 1200s
}
