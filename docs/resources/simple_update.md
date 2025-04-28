---
# Copyright (c) 2023-2025 Dell Inc., or its subsidiaries. All Rights Reserved.
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

title: "redfish_simple_update resource"
linkTitle: "redfish_simple_update"
page_title: "redfish_simple_update Resource - terraform-provider-redfish"
subcategory: ""
description: |-
  This Terraform resource is used to Update firmware of the iDRAC Server. We can Read the existing firmware version or update the same using this resource.
---

# redfish_simple_update (Resource)

This Terraform resource is used to Update firmware of the iDRAC Server. We can Read the existing firmware version or update the same using this resource.

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
      version = "1.6.0"
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
Copyright (c) 2023-2025 Dell Inc., or its subsidiaries. All Rights Reserved.

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
    # Alias name for server BMCs. The key in provider's `redfish_servers` map
    # `redfish_alias` is used to align with enhancements to password management.
    # When using redfish_alias, provider's `redfish_servers` is required.
    redfish_alias = each.key

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

  // by default, the resource uses the first system
  # system_id = "System.Embedded.1"
}
```

After the successful execution of the above resource block, firmware would have been updated. It can be verified through state file.

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `reset_type` (String) Reset type allows to choose the type of restart to apply when firmware upgrade is scheduled. Possible values are: "ForceRestart", "GracefulRestart" or "PowerCycle"
- `target_firmware_image` (String) Target firmware image used for firmware update on the redfish instance. Make sure you place your firmware packages in the same folder as the module and set it as follows: "${path.module}/BIOS_FXC54_WN64_1.15.0.EXE"
- `transfer_protocol` (String) The network protocol that the Update Service uses to retrieve the software image file located at the URI provided in ImageURI, if the URI does not contain a scheme. Accepted values: CIFS, FTP, SFTP, HTTP, HTTPS, NSF, SCP, TFTP, OEM, NFS. Currently only HTTP, HTTPS and NFS are supported with local file path or HTTP(s)/NFS link.

### Optional

- `redfish_server` (Block List) List of server BMCs and their respective user credentials (see [below for nested schema](#nestedblock--redfish_server))
- `reset_timeout` (Number) Time in seconds that the provider waits for the server to be reset before timing out.
- `simple_update_job_timeout` (Number) Time in seconds that the provider waits for the simple update job to be completed before timing out.
- `system_id` (String) System ID of the system

### Read-Only

- `id` (String) ID of the simple update resource
- `software_id` (String) Software ID from the firmware package uploaded
- `version` (String) Software version from the firmware package uploaded

<a id="nestedblock--redfish_server"></a>
### Nested Schema for `redfish_server`

Optional:

- `endpoint` (String) Server BMC IP address or hostname
- `password` (String, Sensitive) User password for login
- `redfish_alias` (String) Alias name for server BMCs. The key in provider's `redfish_servers` map
- `ssl_insecure` (Boolean) This field indicates whether the SSL/TLS certificate must be verified or not
- `user` (String) User name for login



