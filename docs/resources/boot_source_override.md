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

title: "redfish_boot_source_override resource"
linkTitle: "redfish_boot_source_override"
page_title: "redfish_boot_source_override Resource - terraform-provider-redfish"
subcategory: ""
description: |-
  This Terraform resource is used to configure Boot sources of the iDRAC Server.
---

# redfish_boot_source_override (Resource)

This Terraform resource is used to configure Boot sources of the iDRAC Server.

~> **Note:** If the state in `boot_source_override_enabled` is set `once` or `continuous`, the value is reset to disabled after the `boot_source_override_target` actions have completed successfully.

~> **Note:** Changes to these options do not alter the BIOS persistent boot order configuration.

## Example Usage

variables.tf
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
      version = "1.4.0"
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

resource "redfish_boot_source_override" "boot" {
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

  // boot source override parameters
  boot_source_override_enabled = "Once"
  /* list of possible boot source override targets : 
      [ None, Pxe, Floppy, Cd, Usb, Hdd, 
        BiosSetup, Utilities, Diags, UefiShell,UefiTarget
        SDCard, UefiHttp, RemoteDrive, UefiBootNext]
  */
  boot_source_override_target = "UefiTarget"
  boot_source_override_mode   = "UEFI"

  /* Reset parameters to be applied after bios settings are applied
     list of possible value:
      [ ForceRestart, GracefulRestart, PowerCycle]
  */
  reset_type    = "GracefulRestart"
  reset_timeout = "120"
  # // The maximum amount of time to wait for the bios job to be completed
  boot_source_job_timeout = "1200"

  // by default, the resource uses the first system
  # system_id = "System.Embedded.1"
}
```

After the successful execution of the above resource block, the boot source overrides would have been configured. More details can be verified through state file.

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `reset_type` (String) Reset type allows to choose the type of restart to apply when firmware upgrade is scheduled. Possible values are: "ForceRestart", "GracefulRestart" or "PowerCycle"

### Optional

- `boot_source_job_timeout` (Number) Time in seconds that the provider waits for the BootSource override job to be completed before timing out.
- `boot_source_override_enabled` (String) The state of the Boot Source Override feature.
- `boot_source_override_mode` (String) The BIOS boot mode to be used when boot source is booted from.
- `boot_source_override_target` (String) The boot source override target device to use during the next boot instead of the normal boot device.
- `redfish_server` (Block List) List of server BMCs and their respective user credentials (see [below for nested schema](#nestedblock--redfish_server))
- `reset_timeout` (Number) Time in seconds that the provider waits for the server to be reset before timing out.
- `system_id` (String) System ID of the system
- `uefi_target_boot_source_override` (String) The UEFI device path of the device from which to boot when boot_source_override_target is UefiTarget

### Read-Only

- `id` (String) ID of the Boot Source Override Resource

<a id="nestedblock--redfish_server"></a>
### Nested Schema for `redfish_server`

Optional:

- `endpoint` (String) Server BMC IP address or hostname
- `password` (String, Sensitive) User password for login
- `redfish_alias` (String) Alias name for server BMCs. The key in provider's `redfish_servers` map
- `ssl_insecure` (Boolean) This field indicates whether the SSL/TLS certificate must be verified or not
- `user` (String) User name for login


