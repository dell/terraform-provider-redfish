/*
Copyright (c) 2021-2025 Dell Inc., or its subsidiaries. All Rights Reserved.

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
    # Alias name for server BMCs. The key in provider's `redfish_servers` map
    # `redfish_alias` is used to align with enhancements to password management.
    # When using redfish_alias, provider's `redfish_servers` is required.
    redfish_alias = each.key

    user         = each.value.user
    password     = each.value.password
    endpoint     = each.value.endpoint
    ssl_insecure = each.value.ssl_insecure
  }

  // iDRAC attributes to be modified
  attributes = {
    "Users.3.Enable"   = "Disabled"
    "Users.3.UserName" = "mike"
    "Users.3.Password" = "test1234"
    # 17G do not supports Users.x.Privilege,
    # To Update Privileges Please use Users.x.Role configuration, valid values for this
    # (ReadOnly,Operator,Administrator instead of Privilege number)
    # "Users.3.Role"                      = "ReadOnly"

    # Only below 17G Supports Users.x.Privilege, for 17G device remove below Privilege config
    "Users.3.Privilege" = 511

    "Redfish.1.NumericDynamicSegmentsEnable" = "Disabled"
    "SysLog.1.PowerLogInterval"              = "5"
    "Time.1.Timezone"                        = "CST6CDT"
  }
}

