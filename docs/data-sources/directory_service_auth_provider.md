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

title: "redfish_directory_service_auth_provider data source"
linkTitle: "redfish_directory_service_auth_provider"
page_title: "redfish_directory_service_auth_provider Data Source - terraform-provider-redfish"
subcategory: ""
description: |-
  This Terraform datasource is used to query existing Directory Service auth provider. The information fetched from this block can be further used for resource block.
---

# redfish_directory_service_auth_provider (Data Source)

This Terraform datasource is used to query existing Directory Service auth provider. The information fetched from this block can be further used for resource block.

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
Copyright (c) 2024-2025 Dell Inc., or its subsidiaries. All Rights Reserved.

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
      version = "1.6.0"
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

data "redfish_directory_service_auth_provider" "ds_auth" {
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
}

output "directory_service_auth_provider" {
  value     = data.redfish_directory_service_auth_provider.ds_auth
  sensitive = true
}
```

After the successful execution of the above data block, we can see the output in the state file.

<!-- schema generated by tfplugindocs -->
## Schema

### Optional

- `redfish_server` (Block List) List of server BMCs and their respective user credentials (see [below for nested schema](#nestedblock--redfish_server))

### Read-Only

- `active_directory_attributes` (Map of String) ActiveDirectory.* attributes in Dell iDRAC attributes.
- `directory_service_auth_provider` (Attributes) Directory Service Auth Provider Attributes. (see [below for nested schema](#nestedatt--directory_service_auth_provider))
- `id` (String) ID of the Directory Service Auth Provider data-source
- `ldap_attributes` (Map of String) LDAP.* attributes in Dell iDRAC attributes.

<a id="nestedblock--redfish_server"></a>
### Nested Schema for `redfish_server`

Optional:

- `endpoint` (String) Server BMC IP address or hostname
- `password` (String, Sensitive) User password for login
- `redfish_alias` (String) Alias name for server BMCs. The key in provider's `redfish_servers` map
- `ssl_insecure` (Boolean) This field indicates whether the SSL/TLS certificate must be verified or not
- `user` (String) User name for login


<a id="nestedatt--directory_service_auth_provider"></a>
### Nested Schema for `directory_service_auth_provider`

Read-Only:

- `account_lockout_counter_reset_after` (Number) Account Lockout Counter Reset After
- `account_lockout_duration` (Number) Account Lockout Duration
- `account_lockout_threshold` (Number) Account Lockout Threshold
- `accounts` (String) Accounts is a link to a Resource Collection of type ManagerAccountCollection.
- `active_directory` (Attributes) Active Directory (see [below for nested schema](#nestedatt--directory_service_auth_provider--active_directory))
- `additional_external_account_providers` (String) AdditionalExternalAccountProviders is the additional external account providers that this Account Service uses.
- `auth_failure_logging_threshold` (Number) Auth Failure Logging Threshold
- `description` (String) Description of the Account Service
- `id` (String) ID of the Account Service
- `ldap` (Attributes) LDAP (see [below for nested schema](#nestedatt--directory_service_auth_provider--ldap))
- `local_account_auth` (String) Local Account Auth
- `max_password_length` (Number) Maximum Length of the Password
- `min_password_length` (Number) Minimum Length of the Password
- `name` (String) Name of the Account Service.
- `odata_id` (String) OData ID for the Account Service instance
- `password_expiration_days` (Number) Password Expiration Days
- `privilege_map` (String) Privilege Map
- `roles` (String) roles is a link to a Resource Collection of type RoleCollection.
- `service_enabled` (Boolean) ServiceEnabled indicate whether the Accountr Service is enabled.
- `status` (Attributes) Status is any status or health properties of the Resource. (see [below for nested schema](#nestedatt--directory_service_auth_provider--status))
- `supported_account_types` (List of String) SupportedAccountTypes is an array of the account types supported by the service.
- `supported_oem_account_types` (List of String) SupportedOEMAccountTypes is an array of the OEM account types supported by the service.

<a id="nestedatt--directory_service_auth_provider--active_directory"></a>
### Nested Schema for `directory_service_auth_provider.active_directory`

Read-Only:

- `directory` (Attributes) Directory for Active Directory . (see [below for nested schema](#nestedatt--directory_service_auth_provider--active_directory--directory))

<a id="nestedatt--directory_service_auth_provider--active_directory--directory"></a>
### Nested Schema for `directory_service_auth_provider.active_directory.directory`

Read-Only:

- `account_provider_type` (String) AccountProviderType is the type of external account provider to which this service connects.
- `authentication` (Attributes) Authentication information for the account provider. (see [below for nested schema](#nestedatt--directory_service_auth_provider--active_directory--directory--authentication))
- `certificates` (String) Certificates is a link to a resource collection of type CertificateCollection that contains certificates the external account provider uses.
- `remote_role_mapping` (Attributes List) Mapping rules that are used to convert the account providers account information to the local Redfish role (see [below for nested schema](#nestedatt--directory_service_auth_provider--active_directory--directory--remote_role_mapping))
- `service_addresses` (List of String) ServiceAddresses of the account providers
- `service_enabled` (Boolean) ServiceEnabled indicate whether this service is enabled.

<a id="nestedatt--directory_service_auth_provider--active_directory--directory--authentication"></a>
### Nested Schema for `directory_service_auth_provider.active_directory.directory.authentication`

Read-Only:

- `authentication_type` (String) AuthenticationType is used to connect to the account provider


<a id="nestedatt--directory_service_auth_provider--active_directory--directory--remote_role_mapping"></a>
### Nested Schema for `directory_service_auth_provider.active_directory.directory.remote_role_mapping`

Read-Only:

- `local_role` (String) Role Assigned to the Group.
- `remote_group` (String) Name of the remote group.




<a id="nestedatt--directory_service_auth_provider--ldap"></a>
### Nested Schema for `directory_service_auth_provider.ldap`

Read-Only:

- `directory` (Attributes) Directory for LDAP. (see [below for nested schema](#nestedatt--directory_service_auth_provider--ldap--directory))
- `ldap_service` (Attributes) LDAPService is any additional mapping information needed to parse a generic LDAP service. (see [below for nested schema](#nestedatt--directory_service_auth_provider--ldap--ldap_service))

<a id="nestedatt--directory_service_auth_provider--ldap--directory"></a>
### Nested Schema for `directory_service_auth_provider.ldap.directory`

Read-Only:

- `account_provider_type` (String) AccountProviderType is the type of external account provider to which this service connects.
- `authentication` (Attributes) Authentication information for the account provider. (see [below for nested schema](#nestedatt--directory_service_auth_provider--ldap--directory--authentication))
- `certificates` (String) Certificates is a link to a resource collection of type CertificateCollection that contains certificates the external account provider uses.
- `remote_role_mapping` (Attributes List) Mapping rules that are used to convert the account providers account information to the local Redfish role (see [below for nested schema](#nestedatt--directory_service_auth_provider--ldap--directory--remote_role_mapping))
- `service_addresses` (List of String) ServiceAddresses of the account providers
- `service_enabled` (Boolean) ServiceEnabled indicate whether this service is enabled.

<a id="nestedatt--directory_service_auth_provider--ldap--directory--authentication"></a>
### Nested Schema for `directory_service_auth_provider.ldap.directory.authentication`

Read-Only:

- `authentication_type` (String) AuthenticationType is used to connect to the account provider


<a id="nestedatt--directory_service_auth_provider--ldap--directory--remote_role_mapping"></a>
### Nested Schema for `directory_service_auth_provider.ldap.directory.remote_role_mapping`

Read-Only:

- `local_role` (String) Role Assigned to the Group.
- `remote_group` (String) Name of the remote group.



<a id="nestedatt--directory_service_auth_provider--ldap--ldap_service"></a>
### Nested Schema for `directory_service_auth_provider.ldap.ldap_service`

Read-Only:

- `search_settings` (Attributes) SearchSettings is the required settings to search an external LDAP service. (see [below for nested schema](#nestedatt--directory_service_auth_provider--ldap--ldap_service--search_settings))

<a id="nestedatt--directory_service_auth_provider--ldap--ldap_service--search_settings"></a>
### Nested Schema for `directory_service_auth_provider.ldap.ldap_service.search_settings`

Read-Only:

- `base_distinguished_names` (List of String) BaseDistinguishedNames is an array of base distinguished names to use to search an external LDAP service.
- `group_name_attribute` (String) GroupNameAttribute is the attribute name that contains the LDAP group name.
- `user_name_attribute` (String) UsernameAttribute is the attribute name that contains the LDAP user name.




<a id="nestedatt--directory_service_auth_provider--status"></a>
### Nested Schema for `directory_service_auth_provider.status`

Read-Only:

- `health` (String) health
- `health_rollup` (String) health rollup
- `state` (String) state of the storage controller

