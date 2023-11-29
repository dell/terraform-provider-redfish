---
# Copyright (c) 2023 Dell Inc., or its subsidiaries. All Rights Reserved.
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

title: "redfish_storage_volume resource"
linkTitle: "redfish_storage_volume"
page_title: "redfish_storage_volume Resource - terraform-provider-redfish"
subcategory: ""
description: |-
  Resource for managing storage volume.
---

# redfish_storage_volume (Resource)

This Terraform resource is used to configure virtual disks on the iDRAC Server. We can Create, Read, Update, Delete the virtual disks using this resource.
Resource for managing storage volume.

~> **Note:** `capacity_bytes` and `volume_type` attributes cannot be updated.
## Example Usage

variables.tf
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

terraform {
  required_providers {
    redfish = {
      version = "1.1.0"
      source  = "registry.terraform.io/dell/redfish"
    }
  }
}
```

main.tf
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

resource "redfish_storage_volume" "volume" {
  for_each = var.rack1

  redfish_server {
    user         = each.value.user
    password     = each.value.password
    endpoint     = each.value.endpoint
    ssl_insecure = each.value.ssl_insecure
  }

  storage_controller_id = "RAID.Integrated.1-1"
  volume_name           = "TerraformVol"
  volume_type           = "NonRedundant"
  // Name of the physical disk on which virtual disk should get created.
  drives = ["Solid State Disk 0:0:1"]
  // Flag stating when to create virtual disk either "Immediate" or "OnReset"
  settings_apply_time = "Immediate"
  // Reset parameters to be applied when upgrade is completed
  reset_type    = "PowerCycle"
  reset_timeout = 100
  // The maximum amount of time to wait for the volume job to be completed
  volume_job_timeout    = 1200
  capacity_bytes        = 1073323222
  optimum_io_size_bytes = 131072
  read_cache_policy     = "AdaptiveReadAhead"
  write_cache_policy    = "UnprotectedWriteBack"
  disk_cache_policy     = "Disabled"

  lifecycle {
    ignore_changes = [
      capacity_bytes,
      volume_type
    ]
  }
}
```

After the successful execution of the above resource block, virtual disk would have been created. It can be verified through state file.

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `drives` (List of String) Drives
- `storage_controller_id` (String) Storage Controller ID
- `volume_name` (String) Volume Name
- `volume_type` (String) Volume Type

### Optional

- `capacity_bytes` (Number) Capacity Bytes
- `disk_cache_policy` (String) Disk Cache Policy
- `optimum_io_size_bytes` (Number) Optimum Io Size Bytes
- `read_cache_policy` (String) Read Cache Policy
- `redfish_server` (Block List) List of server BMCs and their respective user credentials (see [below for nested schema](#nestedblock--redfish_server))
- `reset_timeout` (Number) Reset Timeout
- `reset_type` (String) Reset Type
- `settings_apply_time` (String) Settings Apply Time
- `volume_job_timeout` (Number) Volume Job Timeout
- `write_cache_policy` (String) Write Cache Policy

### Read-Only

- `id` (String) ID of the storage volume resource

<a id="nestedblock--redfish_server"></a>
### Nested Schema for `redfish_server`

Required:

- `endpoint` (String) Server BMC IP address or hostname

Optional:

- `password` (String, Sensitive) User password for login
- `ssl_insecure` (Boolean) This field indicates whether the SSL/TLS certificate must be verified or not
- `user` (String) User name for login

## Import

Import is supported using the following syntax:

```shell
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

terraform import redfish_storage_volume.volume "{\"id\":\"<odata id of the volume>\",\"username\":\"<username>\",\"password\":\"<password>\",\"endpoint\":\"<endpoint>\",\"ssl_insecure\":<true/false>}"
```

1. This will import the storage volume instance with specified ID into your Terraform state.
2. After successful import, you can run terraform state list to ensure the resource has been imported successfully.
3. Now, you can fill in the resource block with the appropriate arguments and settings that match the imported resource's real-world configuration.
4. Execute terraform plan to see if your configuration and the imported resource are in sync. Make adjustments if needed.
5. Finally, execute terraform apply to bring the resource fully under Terraform's management.
6. Now, the resource which was not part of terraform became part of Terraform managed infrastructure.
