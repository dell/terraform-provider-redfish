---
# Copyright (c) 2023 Dell Inc., or its subsidiaries. All Rights Reserved.

# Licensed under the Mozilla Public License Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at

#     http://mozilla.org/MPL/2.0/


# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
page_title: "Using IPv6"
title: "Using IPv6"
linkTitle: "Using IPv6"
---

The Redfish API works over REST, and hence it works the same over IPv4 or IPv6.
In order to run this provider on an iDRAC over IPv6, all one needs to do is provide the IPv6 address in the endpoint.

## Example

```terraform
resource "redfish_simple_update" "update" {
  for_each = var.rack1

  redfish_server {
    user         = "root"
    password     = "passw0rd"
    endpoint     = "https:://[2001:db8:a::123]"
  }

  // The network protocols and image for firmware update
  transfer_protocol     = "HTTP"
  target_firmware_image = "/home/mikeletux/Downloads/BIOS_FXC54_WN64_1.15.0.EXE"
  // Reset parameters to be applied when upgrade is completed
  reset_type    = "ForceRestart"
  reset_timeout = 120 // If not set, by default will be 120s
  // The maximum amount of time to wait for the simple update job to be completed
  simple_update_job_timeout = 1200 // If not set, by default will be 1200s
}
```
