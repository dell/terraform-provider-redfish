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

title: "redfish_boot_order resource"
linkTitle: "redfish_boot_order"
page_title: "redfish_boot_order Resource - terraform-provider-redfish"
subcategory: ""
description: |-
  This Terraform resource is used to configure Boot Order and enable/disable Boot Options of the iDRAC Server. We can Read the existing configurations or modify them using this resource.
---

# redfish_boot_order (Resource)

This Terraform resource is used to configure Boot Order and enable/disable Boot Options of the iDRAC Server. We can Read the existing configurations or modify them using this resource.

~> **Note:** `boot_order` and `boot_options` are mutually exclusive.

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
      version = "1.2.0"
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

resource "redfish_boot_order" "boot" {
  for_each = var.rack1

  redfish_server {
    user         = each.value.user
    password     = each.value.password
    endpoint     = each.value.endpoint
    ssl_insecure = each.value.ssl_insecure
  }
  // sets the boot devices in the required boot order sequences
  boot_order = ["Boot0001", "Boot0000", "Boot0002", "Boot0003"]

  // Options to enable or disable the boot device. Uncomment the same and comment the boot_order to use this.
  // boot_options = [{boot_option_reference= "Boot0000", boot_option_enabled= false}]

  // Reset parameters to be applied after bios settings are applied
  reset_type    = "ForceRestart"
  reset_timeout = "120"
  // The maximum amount of time to wait for the bios job to be completed
  boot_order_job_timeout = "1200"
}
```

After the successful execution of the above resource block, the boot order would have been configured. More details can be verified through state file.

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `reset_type` (String) Reset type allows to choose the type of restart to apply when firmware upgrade is scheduled. Possible values are: "ForceRestart", "GracefulRestart" or "PowerCycle"

### Optional

- `boot_options` (Attributes List) Options to enable or disable the boot device. (see [below for nested schema](#nestedatt--boot_options))
- `boot_order` (List of String) sets the boot devices in the required boot order sequences.
- `boot_order_job_timeout` (Number) Time in seconds that the provider waits for the BootSource override job to be completed before timing out.
- `redfish_server` (Block List) List of server BMCs and their respective user credentials (see [below for nested schema](#nestedblock--redfish_server))
- `reset_timeout` (Number) Time in seconds that the provider waits for the server to be reset before timing out.

### Read-Only

- `id` (String) ID of the Boot Order Resource

<a id="nestedatt--boot_options"></a>
### Nested Schema for `boot_options`

Required:

- `boot_option_enabled` (Boolean) Enable or disable the boot device.

Optional:

- `boot_option_reference` (String) FQDD of the boot device.


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

# The synatx is:
# terraform import redfish_boot_order.boot "{\"username\":\"<username>\",\"password\":\"<password>\",\"endpoint\":\"<endpoint>\",\"ssl_insecure\":<true/false>}"

terraform import redfish_boot_order.boot '{"username":"admin","password":"passw0rd","endpoint":"https://my-server-1.myawesomecompany.org","ssl_insecure":true}'
```

1. This will import the storage volume instance with specified ID into your Terraform state.
2. After successful import, you can run terraform state list to ensure the resource has been imported successfully.
3. Now, you can fill in the resource block with the appropriate arguments and settings that match the imported resource's real-world configuration.
4. Execute terraform plan to see if your configuration and the imported resource are in sync. Make adjustments if needed.
5. Finally, execute terraform apply to bring the resource fully under Terraform's management.
6. Now, the resource which was not part of terraform became part of Terraform managed infrastructure.
