---
# Copyright (c) 2023-2024 Dell Inc., or its subsidiaries. All Rights Reserved.
#
# Licensed under the Mozilla Public License Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://mozilla.org/MPL/2.0/
#
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

title: "redfish_power resource"
linkTitle: "redfish_power"
page_title: "redfish_power Resource - terraform-provider-redfish"
subcategory: ""
description: |-
  This Terraform resource is used to configure Power attributes of the iDRAC Server. We can Read the existing power state or modify it using this resource.
---

# redfish_power (Resource)

This Terraform resource is used to configure Power attributes of the iDRAC Server. We can Read the existing power state or modify it using this resource.

## Example Usage

variables.tf
```terraform
/*
Copyright (c) 2022-2024 Dell Inc., or its subsidiaries. All Rights Reserved.

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

variable "rack1" {
  type = map(object({
    user         = string
    password     = string
    endpoint     = string
    ssl_insecure = bool
  }))
}
```

terraform.tfvars
```terraform
/*
Copyright (c) 2023 Dell Inc., or its subsidiaries. All Rights Reserved.

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

rack1 = {
  "my-server-1" = {
    user         = "admin"
    password     = "passw0rd"
    endpoint     = "https://my-server-1.myawesomecompany.org"
    ssl_insecure = true
  },
  "my-server-2" = {
    user         = "admin"
    password     = "passw0rd"
    endpoint     = "https://my-server-2.myawesomecompany.org"
    ssl_insecure = true
  },
}
```

provider.tf
```terraform
/*
Copyright (c) 2022-2024 Dell Inc., or its subsidiaries. All Rights Reserved.

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

terraform {
  required_providers {
    redfish = {
      version = "1.5.0"
      source  = "registry.terraform.io/dell/redfish"
    }
  }
}

provider "redfish" {
  # `redfish_servers` is used to align with enhancements to password management.
  # Map of server BMCs with their alias keys and respective user credentials.
  # This is required when resource/datasource's `redfish_alias` is not null
  redfish_servers = var.rack1
}
```

main.tf
```terraform
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
    # Alias name for server BMCs. The key in provider's `redfish_servers` map
    # `redfish_alias` is used to align with enhancements to password management.
    # When using redfish_alias, provider's `redfish_servers` is required.
    redfish_alias = each.key

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
```

After the successful execution of the above resource block, Power state would have been modified to the above desired value. It can be verified through state file.

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `desired_power_action` (String) Desired power setting. Applicable values are 'On','ForceOn','ForceOff','ForceRestart','GracefulRestart','GracefulShutdown','PowerCycle', 'PushPowerButton', 'Nmi'

### Optional

- `check_interval` (Number) The frequency with which to check the server's power state in seconds
- `maximum_wait_time` (Number) The maximum amount of time to wait for the server to enter the correct power state beforegiving up in seconds
- `redfish_server` (Block List) List of server BMCs and their respective user credentials (see [below for nested schema](#nestedblock--redfish_server))
- `system_id` (String) System ID of the system

### Read-Only

- `id` (String) ID of the power resource
- `power_state` (String) Desired power setting. Applicable values 'On','ForceOn','ForceOff','ForceRestart','GracefulRestart','GracefulShutdown','PowerCycle', 'PushPowerButton', 'Nmi'.

<a id="nestedblock--redfish_server"></a>
### Nested Schema for `redfish_server`

Optional:

- `endpoint` (String) Server BMC IP address or hostname
- `password` (String, Sensitive) User password for login
- `redfish_alias` (String) Alias name for server BMCs. The key in provider's `redfish_servers` map
- `ssl_insecure` (Boolean) This field indicates whether the SSL/TLS certificate must be verified or not
- `user` (String) User name for login



