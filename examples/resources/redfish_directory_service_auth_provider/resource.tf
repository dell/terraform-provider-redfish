/*
Copyright (c) 2024 Dell Inc., or its subsidiaries. All Rights Reserved.

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

data "local_file" "kerberos" {
  # this is the path to the certificate that we want to upload.
  filename = "C:\\Users\\Sapana_Gupta\\TerraForm Workspace\\Terraform_Redfish_GofishUpdate\\terraform-provider-redfish\\encoded_kerberos_file.txt"
}


resource "redfish_directory_service_auth_provider" "ds_auth" {
  for_each = var.rack1

  redfish_server {
    user         = each.value.user
    password     = each.value.password
    endpoint     = each.value.endpoint
    ssl_insecure = each.value.ssl_insecure
  }

  active_directory = {
    directory = {
       # remote_role_mapping = [
       #     {
       #         local_role = "None",
       #         remote_group = "idracgroup"
       #     }
       # ],
       # service_addresses = [
       #     "yulanadhost11.yulan.pie.lab.emc.com"
       #  ],
       # service_enabled = true
    },
        authentication = {
            #kerberos_key_tab_file = data.local_file.kerberos.content
        }
}

active_directory_attributes = {
  "ActiveDirectory.1.SSOEnable" = "Disabled",
  #"UserDomain.1.Name" = "yulan.pie.lab.emc.com",
  "ActiveDirectory.1.AuthTimeout" = "120",
  "ActiveDirectory.1.DCLookupEnable" = "Disabled",
  "ActiveDirectory.1.Schema" = "Extended Schema",
  "ActiveDirectory.1.GCLookupEnable" = "Disabled",
  "ActiveDirectory.1.GlobalCatalog1" = "yulanadhost11.yulan.pie.lab.emc.com",
  "ActiveDirectory.1.GlobalCatalog2" = "",
  "ActiveDirectory.1.GlobalCatalog3" = "",
  "ADGroup.1.Domain" = "yulan.pie.lab.emc.com",
  "ActiveDirectory.1.RacName" = "test",
  "ActiveDirectory.1.RacDomain" = "test"             
}

ldap = {
    directory = {
        remote_role_mapping = [
            {
                local_role = "Administrator",
                remote_group = "cn = idracgroup,cn = users,dc = yulan,dc = pie,dc = lab,dc = emc,dc = com"
            },
            {
                local_role = "Administrator",
                remote_group = "cn = Admins,ou = Groups,dc = example,dc = org"
            },
            {
                local_role = "Operator",
                remote_group = "cn = PowerUsers,ou = Groups,dc = example,dc = org"
            }        
        ],
        service_addresses = [
            "yulanadhost12.yulan.pie.lab.emc.com"
        ],
        service_enabled = false
    },
    ldap_service = {
        search_settings = {
            base_distinguished_names = [
                  "dc = yulan,dc = pie,dc = lab,dc = emc,dc = com"
            ],
            group_name_attribute = "name",
            user_name_attribute = "member"
        }
    }
}

 ldap_attributes = {
  "LDAP.1.GroupAttributeIsDN" = "Enabled"
  "LDAP.1.Port" = "636",
  "LDAP.1.BindDN" = "cn = adtester,cn = users,dc = yulan,dc = pie,dc = lab,dc = emc,dc = com",
  "LDAP.1.BindPassword" = "",
  "LDAP.1.SearchFilter" = "(objectclass = *)"
  }

}
