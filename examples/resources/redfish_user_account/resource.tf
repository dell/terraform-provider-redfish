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

resource "redfish_user_account" "rr" {
  for_each = var.rack1

  redfish_server {
    user         = each.value.user
    password     = each.value.password
    endpoint     = each.value.endpoint
    ssl_insecure = true
  }

  // user details for creating/modifying a user
  # user_id  = "4"
  #  username = "test"
  #   password = "T0mPassword123!"
  #   role_id  = "Operator"
  #   // to set user as active or inactive
  #   enabled = true

  users = [
    {
      # user_id="9"
      username = "tom",
      password = "T0mPassword123!",
      role_id  = "Operator",
      enabled  = true,
    },
    {
      # user_id="10"
      username = "dick"
      password = "D!ckPassword123!"
      role_id  = "ReadOnly"
      enabled  = true
    },
    {
      # user_id="11"
      username = "harry"
      password = "H@rryPassword123!"
      role_id  = "ReadOnly"
      enabled  = true
    },
  ]
}
