---
# Copyright (c) 2023-2024 Dell Inc., or its subsidiaries. All Rights Reserved.
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

title: "redfish_certificate resource"
linkTitle: "redfish_certificate"
page_title: "redfish_certificate Resource - terraform-provider-redfish"
subcategory: ""
description: |-
  Resource for import the ssl certificate to iDRAC, on the basis of input parameter Type. After importing the certificate, the iDRAC will automatically restart.
---

# redfish_certificate (Resource)

Resource for import the ssl certificate to iDRAC, on the basis of input parameter Type. After importing the certificate, the iDRAC will automatically restart.

~> **Note:** By default, the iDRAC comes with a self-signed certificate for its web server. If user wants to replace with her own server certificate (signed by Trusted CA). We support two kinds of SSL certificates (1) Server certificate (2) Custom certificate 

~> **Note:** Server Certificate: Steps:- (1) Generate the CSR from iDrac. (2) Create the certificate using CSR and sign with trusted CA. (3) The certificate should be signed with hashing algorithm equivalent to sha256

~> **Note:** Custom Certificate: Steps:- (1) An externally created custom certificate which can be imported into the iDRAC. (2) Convert the external custom certificate into PKCS#12 format and should be encoded via base64. The converion will require passphrase which should be provided in 'passphrase' attribute."



## Example Usage

variables.tf
```terraform
/*
Copyright (c) 2021-2024 Dell Inc., or its subsidiaries. All Rights Reserved.

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
Copyright (c) 2022-2024 Dell Inc., or its subsidiaries. All Rights Reserved.

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
      version = "1.4.0"
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
Copyright (c) 2021-2024 Dell Inc., or its subsidiaries. All Rights Reserved.

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

data "local_file" "cert" {
  # this is the path to the certificate that we want to upload.
  filename = "/root/certificate/new/terraform-provider-redfish/test-data/valid-cert.txt"
}

resource "redfish_certificate" "cert" {
  for_each = var.rack1

  redfish_server {
    user         = each.value.user
    password     = each.value.password
    endpoint     = each.value.endpoint
    ssl_insecure = each.value.ssl_insecure
  }

  /* Type of the certificate to be imported
   List of possible values: [CustomCertificate, Server]
  */
  certificate_type        = "CustomCertificate"
  passphrase              = "12345"
  ssl_certificate_content = data.local_file.cert.content
}
```

After the successful execution of the above resource block, the iDRAC web server would have been configured with the provided SSL certificate. More details can be verified through state file.

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `certificate_type` (String) Type of the certificate to be imported.
- `ssl_certificate_content` (String) SSLCertificate File require content of certificate 
				supported certificate type: 
				"CustomCertificate" - The certificate must be converted pkcs#12 format to encoded in Base64 and entire Base64 Content is required. The passphrase that was used to convert the certificate to pkcs#12 format must also be provided in "passphrase" attribute. "Server" - Certificate Content is required. Note - The certificate should be signed with hashing algorithm equivalent to sha256.

### Optional

- `passphrase` (String) A passphrase for certificate file. Note: This is optional parameter for CSC certificate, and not required for Server and CA certificates.
- `redfish_server` (Block List) List of server BMCs and their respective user credentials (see [below for nested schema](#nestedblock--redfish_server))

### Read-Only

- `id` (String) ID

<a id="nestedblock--redfish_server"></a>
### Nested Schema for `redfish_server`

Optional:

- `endpoint` (String) Server BMC IP address or hostname
- `password` (String, Sensitive) User password for login
- `redfish_alias` (String) Alias name for server BMCs. The key in provider's `redfish_servers` map
- `ssl_insecure` (Boolean) This field indicates whether the SSL/TLS certificate must be verified or not
- `user` (String) User name for login


