---
# Copyright (c) 2024 Dell Inc., or its subsidiaries. All Rights Reserved.

# Licensed under the Mozilla Public License Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at

#     http://mozilla.org/MPL/2.0/


# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
page_title: "Enhancement for password managemant"
title: "Enhancement for password managemant"
linkTitle: "Enhancement for password managemant"
---
Enhancements to password management
The guide provides a terraform configuration of using `redfish_alias` to enhance password managemant. 
The purpose of this enhancement is that when the user password changes, we only need to update the password value in the locals variable, and we no longer need to manually edit the state files to change the old root password to new password.
All we need to do is introduce `redfish_servers` to the provider, while introducing `redfish_alias` to resource/datasource's `redfish_server`.

## Example

```terraform
provider "redfish" {
  # Add `redfish_servers` to provider. This is required when resource/datasource's `redfish_alias` is not null
  redfish_servers  = var.rack1
}

resource "redfish_user_account" "rr" {
  for_each = var.rack1

  redfish_server {
    # Add `redfish_alias` to resource/datasource
    redfish_alias = each.key
  }

  user_id  = "4"
  username = "test"
  password = "Test@123"
  role_id  = "Operator"
  enabled = true
}
```

## Example for Import
```terraform
# terraform import with redfish_alias. When using redfish_alias, provider's `redfish_servers` is required.
# redfish_alias is used to align with enhancements to password management.
terraform import redfish_user_account.rr "{\"id\":\"<id>\",\"redfish_alias\":\"<redfish_alias>\"}"
```