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

title: "redfish_dell_idrac_attributes resource"
linkTitle: "redfish_dell_idrac_attributes"
page_title: "redfish_dell_idrac_attributes Resource - terraform-provider-redfish"
subcategory: ""
description: |-
  Resource for managing DellIdracAttributes on OpenManage Enterprise.
---

# redfish_dell_idrac_attributes (Resource)

Resource for managing DellIdracAttributes on OpenManage Enterprise.

This Terraform resource is used to configure iDRAC attributes of the iDRAC Server. We can Read the existing configurations or modify them using this resource.
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
    user          = string
    password      = string
    endpoint      = string
    validate_cert = bool
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
    user          = "admin"
    password      = "passw0rd"
    endpoint      = "https://my-server-1.myawesomecompany.org"
    validate_cert = false
  },
  "my-server-2" = {
    user          = "admin"
    password      = "passw0rd"
    endpoint      = "https://my-server-2.myawesomecompany.org"
    validate_cert = false
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

resource "redfish_dell_idrac_attributes" "idrac" {
  for_each = var.rack1

  redfish_server = {
    user          = each.value.user
    password      = each.value.password
    endpoint      = each.value.endpoint
    validate_cert = each.value.validate_cert
  }

  // iDRAC attributes to be modified
  attributes = {
    "Users.3.Enable"                         = "Disabled"
    "Users.3.UserName"                       = "mike"
    "Users.3.Password"                       = "test1234"
    "Users.3.Privilege"                      = 511
    "Redfish.1.NumericDynamicSegmentsEnable" = "Disabled"
    "SysLog.1.PowerLogInterval"              = "5"
    "Time.1.Timezone"                        = "CST6CDT"
  }
}
```

After the successful execution of the above resource block, iDRAC attributes configuration would have got altered. It can be verified through state file.

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `attributes` (Map of String) iDRAC attributes. To check allowed attributes please either use the datasource for dell idrac attributes or query /redfish/v1/Managers/iDRAC.Embedded.1/Oem/Dell/DellAttributes/iDRAC.Embedded.1. To get allowed values for those attributes, check /redfish/v1/Registries/ManagerAttributeRegistry/ManagerAttributeRegistry.v1_0_0.json from a Redfish Instance
- `redfish_server` (Attributes) Redfish Server (see [below for nested schema](#nestedatt--redfish_server))

### Read-Only

- `id` (String) ID of the iDRAC attributes resource

<a id="nestedatt--redfish_server"></a>
### Nested Schema for `redfish_server`

Required:

- `endpoint` (String) Server BMC IP address or hostname

Optional:

- `password` (String, Sensitive) User password for login
- `user` (String) User name for login
- `validate_cert` (Boolean) This field indicates whether the SSL/TLS certificate must be verified or not


