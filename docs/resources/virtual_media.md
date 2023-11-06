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

title: "redfish_virtual_media resource"
linkTitle: "redfish_virtual_media"
page_title: "redfish_virtual_media Resource - terraform-provider-redfish"
subcategory: ""
description: |-
  Resource for managing virtual media.
---

# redfish_virtual_media (Resource)

Resource for managing virtual media.

~> **Note:** `write_protected` attribute can only be configured as `true`.
## Example Usage

This Terraform resource is used to configure virtual media on the iDRAC Server. We can Read, Attach, Detach the virtual media or Modify the attached image using this resource.
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

provider "redfish" {
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

resource "redfish_virtual_media" "vm" {
  for_each = var.rack1

  redfish_server = {
    user         = each.value.user
    password     = each.value.password
    endpoint     = each.value.endpoint
    ssl_insecure = each.value.ssl_insecure
  }
  // Image to be attached to virtual media
  # image           = "http://inuxlib.com/pub/redhat/RHEL8/8.8/BaseOS/x86_64/os/images/efiboot.img"
  image           = "http://linuxlib.us.dell.com/pub/redhat/RHEL8/8.8/BaseOS/x86_64/os/images/efiboot.img"
  transfer_method = "Stream"
  // Network protocol used to fetch the image
  transfer_protocol_type = "HTTP"
  write_protected        = true
}
```

After the successful execution of the above resource block, virutal media would have been attached with specified image. More details can be verified through state file.

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `image` (String) The URI of the remote media to attach to the virtual media

### Optional

- `redfish_server` (Block List) List of server BMCs and their respective user credentials (see [below for nested schema](#nestedblock--redfish_server))
- `transfer_method` (String) Indicates how the data is transferred
- `transfer_protocol_type` (String) The protocol used to transfer.
- `write_protected` (Boolean) Indicates whether the remote device media prevents writing to that media.

### Read-Only

- `id` (String) ID of the virtual media resource
- `inserted` (Boolean) Describes whether virtual media is attached or detached

<a id="nestedblock--redfish_server"></a>
### Nested Schema for `redfish_server`

Required:

- `endpoint` (String) Server BMC IP address or hostname

Optional:

- `password` (String, Sensitive) User password for login
- `user` (String) User name for login
- `validate_cert` (Boolean) This field indicates whether the SSL/TLS certificate must be verified or not


