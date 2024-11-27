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

title: "redfish_directory_service_auth_provider_certificate data source"
linkTitle: "redfish_directory_service_auth_provider_certificate"
page_title: "redfish_directory_service_auth_provider_certificate Data Source - terraform-provider-redfish"
subcategory: ""
description: |-
  This Terraform datasource is used to query existing Directory Service auth provider Certificate. The information fetched from this block can be further used for resource block.
---

# redfish_directory_service_auth_provider_certificate (Data Source)

This Terraform datasource is used to query existing Directory Service auth provider Certificate. The information fetched from this block can be further used for resource block.

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

data "redfish_directory_service_auth_provider_certificate" "ds_auth_certificate" {
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

  certificate_filter {
    certificate_provider_type = "LDAP"
    # certificate_id            = "SecurityCertificate.5"
  }

  # security_certificate can be viewed if server has datacenter license
}

output "directory_service_auth_provider_certificate" {
  value     = data.redfish_directory_service_auth_provider_certificate.ds_auth_certificate
  sensitive = true
}
```

After the successful execution of the above data block, we can see the output in the state file.

<!-- schema generated by tfplugindocs -->
## Schema

### Optional

- `certificate_filter` (Block, Optional) Certificate filter for Directory Service Auth Provider (see [below for nested schema](#nestedblock--certificate_filter))
- `redfish_server` (Block List) List of server BMCs and their respective user credentials (see [below for nested schema](#nestedblock--redfish_server))

### Read-Only

- `directory_service_auth_provider_certificate` (Attributes) Directory Service Auth Provider Certificate Details. (see [below for nested schema](#nestedatt--directory_service_auth_provider_certificate))
- `id` (String) ID of the Directory Service Auth Provider Certificate data-source

<a id="nestedblock--certificate_filter"></a>
### Nested Schema for `certificate_filter`

Required:

- `certificate_provider_type` (String) Filter for CertificateProviderType

Optional:

- `certificate_id` (String) CertificateId


<a id="nestedblock--redfish_server"></a>
### Nested Schema for `redfish_server`

Optional:

- `endpoint` (String) Server BMC IP address or hostname
- `password` (String, Sensitive) User password for login
- `redfish_alias` (String) Alias name for server BMCs. The key in provider's `redfish_servers` map
- `ssl_insecure` (Boolean) This field indicates whether the SSL/TLS certificate must be verified or not
- `user` (String) User name for login


<a id="nestedatt--directory_service_auth_provider_certificate"></a>
### Nested Schema for `directory_service_auth_provider_certificate`

Read-Only:

- `directory_service_certificate` (Attributes) Directory Service Certificate Details. (see [below for nested schema](#nestedatt--directory_service_auth_provider_certificate--directory_service_certificate))
- `security_certificate` (Map of String) SecurityCertificate attributes in Dell iDRAC attributes.

<a id="nestedatt--directory_service_auth_provider_certificate--directory_service_certificate"></a>
### Nested Schema for `directory_service_auth_provider_certificate.directory_service_certificate`

Read-Only:

- `certificate_usage_types` (List of String) The types or purposes for this certificate
- `description` (String) Description of the Certificate
- `issuer` (Attributes) The issuer of the certificate (see [below for nested schema](#nestedatt--directory_service_auth_provider_certificate--directory_service_certificate--issuer))
- `name` (String) Name of the Certificate
- `odata_id` (String) OData ID for the Certificate
- `serial_number` (String) The serial number of the certificate
- `subject` (Attributes) The subject of the certificate (see [below for nested schema](#nestedatt--directory_service_auth_provider_certificate--directory_service_certificate--subject))
- `valid_not_after` (String) The date when the certificate is no longer valid
- `valid_not_before` (String) The date when the certificate becomes valid

<a id="nestedatt--directory_service_auth_provider_certificate--directory_service_certificate--issuer"></a>
### Nested Schema for `directory_service_auth_provider_certificate.directory_service_certificate.issuer`

Read-Only:

- `city` (String) The city or locality of the organization of the entity
- `common_name` (String) The common name of the entity
- `country` (String) The country of the organization of the entity
- `email` (String) The email address of the contact within the organization of the entity
- `organization` (String) The name of the organization of the entity
- `organizational_unit` (String) The name of the unit or division of the organization of the entity
- `state` (String) The state, province, or region of the organization of the entity


<a id="nestedatt--directory_service_auth_provider_certificate--directory_service_certificate--subject"></a>
### Nested Schema for `directory_service_auth_provider_certificate.directory_service_certificate.subject`

Read-Only:

- `city` (String) The city or locality of the organization of the entity
- `common_name` (String) The common name of the entity
- `country` (String) The country of the organization of the entity
- `email` (String) The email address of the contact within the organization of the entity
- `organization` (String) The name of the organization of the entity
- `organizational_unit` (String) The name of the unit or division of the organization of the entity
- `state` (String) The state, province, or region of the organization of the entity

