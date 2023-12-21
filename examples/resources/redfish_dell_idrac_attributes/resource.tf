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

  redfish_server {
    user         = each.value.user
    password     = each.value.password
    endpoint     = each.value.endpoint
    ssl_insecure = each.value.ssl_insecure
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

