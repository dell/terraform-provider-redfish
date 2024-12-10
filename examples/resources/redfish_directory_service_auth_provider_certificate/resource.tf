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

data "local_file" "ds_certificate" {
  # this is the path to the certificate that we want to upload.
  filename = "/root/certificate/new/terraform-provider-redfish/test-data/valid-ds-cert.txt"
}

resource "redfish_directory_service_auth_provider_certificate" "ds_auth_certificate" {
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

  # RSA_CA_CERT certificate resource can be created/modified only if server have datacenter license
  # certificate type can be PEM or RSA_CA_CERT
  certificate_type   = "RSA_CA_CERT1"
  certificate_string = data.local_file.ds_certificate.content
}