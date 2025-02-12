---
# Copyright (c) 2024 Dell Inc., or its subsidiaries. All Rights Reserved.
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

title: "redfish_directory_service_auth_provider resource"
linkTitle: "redfish_directory_service_auth_provider"
page_title: "redfish_directory_service_auth_provider Resource - terraform-provider-redfish"
subcategory: ""
description: |-
  This Terraform resource is used to configure Directory Service Auth Provider Active Directory and LDAP Service We can Read the existing configurations or modify them using this resource.
---

# redfish_directory_service_auth_provider (Resource)

This Terraform resource is used to configure Directory Service Auth Provider Active Directory and LDAP Service We can Read the existing configurations or modify them using this resource.

## Example Usage

variables.tf
```terraform
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

terraform {
  required_providers {
    redfish = {
      version = "1.5.0"
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

# kerberod file is not supported by 17G and this configuration works only for below 17G
data "local_file" "kerberos" {
  # this is the path to the kerberos keytab file that we want to upload.
  # this file must be base64 encoded format
  filename = "/root/directoryservice/new/terraform-provider-redfish/test-data/kerberos_file.txt"
}

# redfish_directory_service_auth_provider Terraform resource is used to configure Directory Service Auth Provider Active Directory and LDAP Service
# Available action: Create, Update (Active Directory, LDAP)
# Active Directory (Create, Update): remote_role_mapping, service_addresses, service_enabled,authentication, active_directory_attributes
# LDAP (Create, Update): remote_role_mapping, service_addresses, service_enabled,ldap_service, ldap_attributes
resource "redfish_directory_service_auth_provider" "ds_auth" {
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

  #Note: `active_directory` is mutually inclusive with `active_directory_attributes`.
  #Note: `ldap` is mutually inclusive with `ldap_attributes`.
  #Note: `active_directory` is mutually exclusive with `ldap`.
  #Note: `active_directory_attributes` is mutually exclusive with `ldap_attributes`. 
  active_directory = {
    directory = {
      # remote_role_mapping = [
      #     {
      #         local_role = "Administrator",
      #         remote_group = "idracgroup"
      #     }
      # ],
      # To Update service addresses for 17G please provide configuration in active_directory_attributes
      # This configuration will be working once the issue for 17G get resolved 
      # service_addresses = [
      #     "yulanadhost11.yulan.pie.lab.emc.com"
      #  ],
      service_enabled = true,
      # authentication configuration works for below 17G server, 17G server do not support kerberos
      authentication = {
        kerberos_key_tab_file = data.local_file.kerberos.content
      }
    }
  }

  active_directory_attributes = {
    "ActiveDirectory.1.AuthTimeout"          = "120",
    "ActiveDirectory.1.CertValidationEnable" = "Enabled",
    "ActiveDirectory.1.DCLookupEnable"       = "Enabled",

    # DomainController can be configured when DCLookupEnable is Disabled
    # To update service addresses for 17G please provide below configuration
    #"ActiveDirectory.1.DomainController1"= "yulanadhost1.yulan.pie.lab.emc.com",
    #"ActiveDirectory.1.DomainController2"= "yulanadhost2.yulan.pie.lab.emc.com",
    #"ActiveDirectory.1.DomainController3"= "yulanadhost3.yulan.pie.lab.emc.com",

    # RacName and RacDomain can be configured when Schema is Extended Schema which is supported by below 17G server
    "ActiveDirectory.1.RacDomain" = "test",
    "ActiveDirectory.1.RacName"   = "test",

    # SSOEnable configuration can be done for below 17G server and it's not supported by 17G server
    # if SSOEnable is Enabled make sure ActiveDirectory Service is enabled and valid kerberos_key_tab_file is provided
    "ActiveDirectory.1.SSOEnable" = "Disabled",

    # Schema can be Extended Schema or Standard Schema
    # Schema configuration can be done for below 17G, and this configuration is not supported by 17G 
    "ActiveDirectory.1.Schema" = "Extended Schema",
    "UserDomain.1.Name"        = "yulan.pie.lab.emc.com",

    # DCLookupByUserDomain must be configured when DCLookupEnable is enabled 
    "ActiveDirectory.1.DCLookupByUserDomain" : "Enabled",

    # DCLookupDomainName must be configured when DCLookupByUserDomain is Disabled and DCLookupEnable is Enabled
    #"ActiveDirectory.1.DCLookupDomainName"="test", 

    #"ActiveDirectory.1.GCLookupEnable" = "Disabled"

    # for 17G below configuration can be performed without schema configuration
    # at least any one from GlobalCatalog1,GlobalCatalog2,GlobalCatalog3 must be configured when Schema is Standard and GCLookupEnable is Disabled
    # "ActiveDirectory.1.GlobalCatalog1" = "yulanadhost11.yulan.pie.lab.emc.com",
    # "ActiveDirectory.1.GlobalCatalog2" = "yulanadhost11.yulan.pie.lab.emc.com",
    # "ActiveDirectory.1.GlobalCatalog3" = "yulanadhost11.yulan.pie.lab.emc.com", 

    # GCRootDomain can be configured when GCLookupEnable is Enabled  
    #"ActiveDirectory.1.GCRootDomain" = "test"     
  }



  #    ldap = {
  #		directory = {
  #			remote_role_mapping = [
  #				{
  #					local_role = "Administrator",
  #					remote_group = "cn = idracgroup,cn = users,dc = yulan,dc = pie,dc = lab,dc = emc,dc = com"
  #				}        
  #			],
  #   To Update LDAP service addresses for 17G please provide configuration in ldap_attributes
  #			service_addresses = [
  #				"yulanadhost12.yulan.pie.lab.emc.com"
  #			],
  #			service_enabled = false
  #		},
  #		ldap_service = {
  #			search_settings = {
  #			    base_distinguished_names = [
  #				  "dc = yulan,dc = pie,dc = lab,dc = emc,dc = com"
  #				],
  #				group_name_attribute = "name",
  #				user_name_attribute = "member"
  #			}
  #		}
  #	}
  #		
  #	ldap_attributes = {
  #	  "LDAP.1.GroupAttributeIsDN" = "Enabled"
  #	  "LDAP.1.Port" = "636",
  #	  "LDAP.1.BindDN" = "cn = adtester,cn = users,dc = yulan,dc = pie,dc = lab,dc = emc,dc = com",
  #	  "LDAP.1.BindPassword" = "",
  #	  "LDAP.1.SearchFilter" = "(objectclass = *)",
  #   To Update LDAP service addresses for 17G please provide below configuration
  #   "LDAP.1.Server": "yulanadhost12.yulan.pie.lab.emc.com",
  #	  
  #		  }

}
```

After the successful execution of the above resource block, the directory service auth provider would have been configured. More details can be verified through state file.

<!-- schema generated by tfplugindocs -->
## Schema

### Optional

- `active_directory` (Attributes) Active DirectoryNote: `active_directory` is mutually inclusive with `active_directory_attributes`. , Note: `active_directory` is mutually exclusive with `ldap`. (see [below for nested schema](#nestedatt--active_directory))
- `active_directory_attributes` (Map of String) ActiveDirectory.* attributes in Dell iDRAC attributes.Note: `active_directory` is mutually inclusive with `active_directory_attributes`. , Note: `active_directory_attributes` is mutually exclusive with `ldap_attributes`.
- `ldap` (Attributes) LDAPNote: `ldap` is mutually inclusive with `ldap_attributes`. , Note: `active_directory` is mutually exclusive with `ldap`. (see [below for nested schema](#nestedatt--ldap))
- `ldap_attributes` (Map of String) LDAP.* attributes in Dell iDRAC attributes.Note: `ldap` is mutually inclusive with `ldap_attributes`. , Note: `active_directory_attributes` is mutually exclusive with `ldap_attributes`.
- `redfish_server` (Block List) List of server BMCs and their respective user credentials (see [below for nested schema](#nestedblock--redfish_server))

### Read-Only

- `id` (String) ID of the Directory Service Auth Provider resource

<a id="nestedatt--active_directory"></a>
### Nested Schema for `active_directory`

Optional:

- `authentication` (Attributes) Authentication information for the account provider. (see [below for nested schema](#nestedatt--active_directory--authentication))
- `directory` (Attributes) Directory for Active Directory . (see [below for nested schema](#nestedatt--active_directory--directory))

<a id="nestedatt--active_directory--authentication"></a>
### Nested Schema for `active_directory.authentication`

Optional:

- `kerberos_key_tab_file` (String) KerberosKeytab is a Base64-encoded version of the Kerberos keytab for this Service


<a id="nestedatt--active_directory--directory"></a>
### Nested Schema for `active_directory.directory`

Optional:

- `remote_role_mapping` (Attributes List) Mapping rules that are used to convert the account providers account information to the local Redfish role (see [below for nested schema](#nestedatt--active_directory--directory--remote_role_mapping))
- `service_addresses` (List of String) ServiceAddresses of the account providers
- `service_enabled` (Boolean) ServiceEnabled indicate whether this service is enabled.

<a id="nestedatt--active_directory--directory--remote_role_mapping"></a>
### Nested Schema for `active_directory.directory.remote_role_mapping`

Optional:

- `local_role` (String) Role Assigned to the Group.
- `remote_group` (String) Name of the remote group.




<a id="nestedatt--ldap"></a>
### Nested Schema for `ldap`

Optional:

- `directory` (Attributes) Directory for LDAP. (see [below for nested schema](#nestedatt--ldap--directory))
- `ldap_service` (Attributes) LDAPService is any additional mapping information needed to parse a generic LDAP service. (see [below for nested schema](#nestedatt--ldap--ldap_service))

<a id="nestedatt--ldap--directory"></a>
### Nested Schema for `ldap.directory`

Optional:

- `remote_role_mapping` (Attributes List) Mapping rules that are used to convert the account providers account information to the local Redfish role (see [below for nested schema](#nestedatt--ldap--directory--remote_role_mapping))
- `service_addresses` (List of String) ServiceAddresses of the account providers
- `service_enabled` (Boolean) ServiceEnabled indicate whether this service is enabled.

<a id="nestedatt--ldap--directory--remote_role_mapping"></a>
### Nested Schema for `ldap.directory.remote_role_mapping`

Optional:

- `local_role` (String) Role Assigned to the Group.
- `remote_group` (String) Name of the remote group.



<a id="nestedatt--ldap--ldap_service"></a>
### Nested Schema for `ldap.ldap_service`

Optional:

- `search_settings` (Attributes) SearchSettings is the required settings to search an external LDAP service. (see [below for nested schema](#nestedatt--ldap--ldap_service--search_settings))

<a id="nestedatt--ldap--ldap_service--search_settings"></a>
### Nested Schema for `ldap.ldap_service.search_settings`

Optional:

- `base_distinguished_names` (List of String) BaseDistinguishedNames is an array of base distinguished names to use to search an external LDAP service.
- `group_name_attribute` (String) GroupNameAttribute is the attribute name that contains the LDAP group name.
- `user_name_attribute` (String) UsernameAttribute is the attribute name that contains the LDAP user name.




<a id="nestedblock--redfish_server"></a>
### Nested Schema for `redfish_server`

Optional:

- `endpoint` (String) Server BMC IP address or hostname
- `password` (String, Sensitive) User password for login
- `redfish_alias` (String) Alias name for server BMCs. The key in provider's `redfish_servers` map
- `ssl_insecure` (Boolean) This field indicates whether the SSL/TLS certificate must be verified or not
- `user` (String) User name for login

## Import

Import is supported using the following syntax:

```shell
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

terraform import redfish_directory_service_auth_provider.ds_auth '{"username":"<username>","password":"<password>","endpoint":"<endpoint>","ssl_insecure":<true/false>}'

# terraform import with redfish_alias. When using redfish_alias, provider's `redfish_servers` is required.
# redfish_alias is used to align with enhancements to password management.
terraform import redfish_directory_service_auth_provider.ds_auth '{"redfish_alias":"<redfish_alias>"}'
```

1. This will import the Directory Service Auth Provider configuration into your Terraform state.
2. After successful import, you can run terraform state list to ensure the resource has been imported successfully.
3. Now, you can fill in the resource block with the appropriate arguments and settings that match the imported resource's real-world configuration.
4. Execute terraform plan to see if your configuration and the imported resource are in sync. Make adjustments if needed.
5. Finally, execute terraform apply to bring the resource fully under Terraform's management.
6. Now, the resource which was not part of terraform became part of Terraform managed infrastructure.
