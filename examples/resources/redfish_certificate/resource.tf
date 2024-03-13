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
