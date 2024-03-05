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

// Resource to manage lifecycle for redfish_idrac_firmware_update
resource "terraform_data" "always_run" {
  input = timestamp()
}

resource "redfish_idrac_firmware_update" "update" {
  //  This resource is supposed to be replaced each time we run terraform apply.
  for_each = var.rack1

  redfish_server {
    user         = each.value.user
    password     = each.value.password
    endpoint     = each.value.endpoint
    ssl_insecure = each.value.ssl_insecure
  }
  # Required
  // IP address for the remote share.
  ip_address = "xx.xx.xx.xx"
  // Type of the Network Share.
  share_type = "HTTP" # "CIFS" | "NFS" | "HTTP" | "HTTPS" | "FTP" | "TFTP"

  // These two fields should are set to true by default. It will check the repository for any updates that are available for the server and apply those updates.
  // If you do not want to apply the updates and just want to get the details for the updates available,you can set these fields to false.
  # apply_update = true  # Default is true
  # reboot_needed = true # Default is true

  # Optional fields(Based on  share Type)

  // Name of the CIFS share or full path to the NFS share. Optional for HTTP/HTTPS share (if supported), this may be treated as the path of the directory containing the file.
  # share_name = "idrac-firmware"

  // Name of the catalog file on the repository. Default is Catalog.xml.
  # catalog_file_name = "5_10_50_00_A00_Catalog.xml"

  // Username and Password for the remote share. They must be provided for CIFS.
  #  share_user = "username"
  #  share_password = "password"

  # Proxy Settings
  # proxy_support = "ParametersProxy" # "ParametersProxy" | "Off" , Default is "Off"
  # proxy_server = "xx.xx.xx.xx"
  # proxy_port = 80

  // This will allow terraform create process to trigger each time we run terraform apply.
  lifecycle {
    replace_triggered_by = [
      terraform_data.always_run
    ]
  }

}
