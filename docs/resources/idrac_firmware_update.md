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

title: "redfish_idrac_firmware_update resource"
linkTitle: "redfish_idrac_firmware_update"
page_title: "redfish_idrac_firmware_update Resource - terraform-provider-redfish"
subcategory: ""
description: |-
  This Terraform resource is used to Update firmware of the iDRAC Server based on a catalog.
---

# redfish_idrac_firmware_update (Resource)

This Terraform resource is used to Update firmware of the iDRAC Server based on a catalog.

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

// Resource to manage lifecycle for redfish_idrac_firmware_update
resource "terraform_data" "always_run" {
  input = timestamp()
}

resource "redfish_idrac_firmware_update" "update" {
  //  This resource is supposed to be replaced each time we run terraform apply.
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

  // by default, the resource uses the first system
  # system_id = "System.Embedded.1"

  // This will allow terraform create process to trigger each time we run terraform apply.
  lifecycle {
    replace_triggered_by = [
      terraform_data.always_run
    ]
  }

}
```

After the successful execution of the above resource block, iDRAC firmware attributes configuration would have been altered. It can be verified through state file.

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `ip_address` (String) IP address for the remote share.
- `share_type` (String) Type of the Network Share.

### Optional

- `apply_update` (Boolean) If ApplyUpdate is set to true, the updatable packages from Catalog XML are staged. If it is set to False, no updates are applied but the list of updatable packages can be seen in the UpdateList.Default is true.
- `catalog_file_name` (String) Name of the catalog file on the repository. Default is Catalog.xml.
- `ignore_cert_warning` (String) Specifies if certificate warning should be ignored when HTTPS is used. If ignore_cert_warning is On,warnings are ignored. Default is On.
- `mount_point` (String) The local directory where the share should be mounted.
- `proxy_password` (String) The password for the proxy server.
- `proxy_port` (Number) The Port for the proxy server.Default is set to 80.
- `proxy_server` (String) The IP address of the proxy server.This IP will not be validated. The download job will be created even forinvalid proxy_server.Please check the results of the job for error details.This is required when proxy_support is ParametersProxy.
- `proxy_support` (String) Specifies if a proxy should be used. Default is Off. This option is only used for HTTP, HTTPS, and FTP shares.
- `proxy_type` (String) The proxy type of the proxy server. Default is (HTTP).
- `proxy_username` (String) The user name for the proxy server.
- `reboot_needed` (Boolean) This property indicates if a reboot should be performed. True indicates that the system (host) is rebooted duringthe update process. False indicates that the updates take effect after the system is rebooted the next time.Default is true.
- `redfish_server` (Block List) List of server BMCs and their respective user credentials (see [below for nested schema](#nestedblock--redfish_server))
- `share_name` (String) Name of the CIFS share or full path to the NFS share. Optional for HTTP/HTTPS share (if supported)this may be treated as the path of the directory containing the file.
- `share_password` (String) Network share user password. This option is mandatory for CIFS Network Share.
- `share_user` (String) Network share user in the format 'user@domain' or 'domain\user' if user is part of a domain else 'user'.This option is mandatory for CIFS Network Share.
- `system_id` (String) System ID of the system

### Read-Only

- `id` (String) ID of the iDRAC Firmware Update Resource.
- `update_list` (Attributes List) List of properties of the update list. (see [below for nested schema](#nestedatt--update_list))

<a id="nestedblock--redfish_server"></a>
### Nested Schema for `redfish_server`

Optional:

- `endpoint` (String) Server BMC IP address or hostname
- `password` (String, Sensitive) User password for login
- `redfish_alias` (String) Alias name for server BMCs. The key in provider's `redfish_servers` map
- `ssl_insecure` (Boolean) This field indicates whether the SSL/TLS certificate must be verified or not
- `user` (String) User name for login


<a id="nestedatt--update_list"></a>
### Nested Schema for `update_list`

Read-Only:

- `criticality` (String) Criticality of the package update.
- `current_package_version` (String) Current version of the package.
- `display_name` (String) Display name of the package.
- `job_id` (String) ID of the job if it's triggered.
- `job_message` (String) Message from the job if it's triggered.
- `job_status` (String) Status of the job if it's triggered.
- `package_name` (String) Name of the package to be updated.
- `reboot_type` (String) Reboot type of the package update.
- `target_package_version` (String) Target version of the package.


