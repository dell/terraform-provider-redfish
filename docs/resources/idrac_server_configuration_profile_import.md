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

title: "redfish_idrac_server_configuration_profile_import resource"
linkTitle: "redfish_idrac_server_configuration_profile_import"
page_title: "redfish_idrac_server_configuration_profile_import Resource - terraform-provider-redfish"
subcategory: ""
description: |-
  Resource for managing iDRAC Server Configuration Profile Import on iDRAC Server.
---

# redfish_idrac_server_configuration_profile_import (Resource)

Resource for managing iDRAC Server Configuration Profile Import on iDRAC Server.

## Example Usage

variables.tf
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

variable "rack1" {
  type = map(object({
    user         = string
    password     = string
    endpoint     = string
    ssl_insecure = bool
  }))
}

variable "cifs_username" {
  type    = string
  default = "awesomeadmin"
}

variable "cifs_password" {
  type    = string
  default = "C00lP@ssw0rd"

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
      version = "1.3.0"
      source  = "registry.terraform.io/dell/redfish"
    }
  }
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

resource "terraform_data" "trigger_by_timestamp" {
  input = timestamp()
}

resource "redfish_idrac_server_configuration_profile_import" "share_type_local" {
  for_each = var.rack1

  redfish_server {
    user         = each.value.user
    password     = each.value.password
    endpoint     = each.value.endpoint
    ssl_insecure = each.value.ssl_insecure
  }

  share_parameters = {
    filename   = "demo_local.xml"
    target     = ["NIC"]
    share_type = "LOCAL"
  }

  lifecycle {
    replace_triggered_by = [terraform_data.trigger_by_timestamp]
  }
}

resource "redfish_idrac_server_configuration_profile_import" "share_type_nfs" {
  for_each = var.rack1

  redfish_server {
    user         = each.value.user
    password     = each.value.password
    endpoint     = each.value.endpoint
    ssl_insecure = each.value.ssl_insecure
  }

  share_parameters = {
    filename   = "demo_nfs.xml"
    target     = ["NIC"]
    share_type = "NFS"
    ip_address = "10.0.0.01"
    share_name = "/dell/terraform-idrac-nfs"
  }

  lifecycle {
    replace_triggered_by = [terraform_data.trigger_by_timestamp]
  }
}

resource "redfish_idrac_server_configuration_profile_import" "share_type_cifs" {
  for_each = var.rack1

  redfish_server {
    user         = each.value.user
    password     = each.value.password
    endpoint     = each.value.endpoint
    ssl_insecure = each.value.ssl_insecure
  }

  share_parameters = {
    filename   = "demo_cifs.xml"
    target     = ["NIC"]
    share_type = "CIFS"
    ip_address = "10.0.0.02"
    share_name = "/dell/terraform-idrac-nfs"
    username   = var.cifs_username
    password   = var.cifs_password
  }

  lifecycle {
    replace_triggered_by = [terraform_data.trigger_by_timestamp]
  }
}

resource "redfish_idrac_server_configuration_profile_import" "share_type_https" {
  for_each = var.rack1

  redfish_server {
    user         = each.value.user
    password     = each.value.password
    endpoint     = each.value.endpoint
    ssl_insecure = each.value.ssl_insecure
  }

  share_parameters = {
    filename    = "demo_https.xml"
    target      = ["NIC"]
    share_type  = "HTTPS"
    ip_address  = "10.0.0.03"
    port_number = 443
  }

  lifecycle {
    replace_triggered_by = [terraform_data.trigger_by_timestamp]
  }
}

resource "redfish_idrac_server_configuration_profile_import" "share_type_http" {
  for_each = var.rack1

  redfish_server {
    user         = each.value.user
    password     = each.value.password
    endpoint     = each.value.endpoint
    ssl_insecure = each.value.ssl_insecure
  }

  share_parameters = {
    filename      = "demo_http.xml"
    target        = ["NIC"]
    share_type    = "HTTP"
    ip_address    = "10.0.0.04"
    port_number   = 80
    proxy_support = true
    proxy_server  = "10.0.0.05"
    proxy_port    = 5000
  }

  lifecycle {
    replace_triggered_by = [terraform_data.trigger_by_timestamp]
  }
}
```

After the successful execution of the above resource block, Server Configuration Profile will be imported from share type.

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `share_parameters` (Attributes) Share Parameters (see [below for nested schema](#nestedatt--share_parameters))

### Optional

- `host_power_state` (String) Host Power State. This attribute allows you to specify the power state of the host when the
				iDRAC is performing the import operation. Accepted values are: "On" or "Off". If this attribute is not specified
				or is set to "On", the host is powered on before the import operation. If it is set to "Off", the host is powered
				off before the import operation. Note that the host will be powered back on after the import is completed.
- `import_buffer` (String) Buffer content to perform Import.This is only required for localstore and is not applicable for CIFS/NFS style Import. If the import buffer is empty, then it will perform the import from the source path specified in share parameters.
- `redfish_server` (Block List) List of server BMCs and their respective user credentials (see [below for nested schema](#nestedblock--redfish_server))
- `shutdown_type` (String) Shutdown Type. This attribute specifies the type of shutdown that should be performed before importing the server configuration profile. Accepted values are: "Graceful" (default), "Forced", or "NoReboot". If set to "Graceful", the server will be gracefully shut down before the import. If set to "Forced", the server will be forcefully shut down before the import. If set to "NoReboot", the server will not be restarted after the import. Note that if the server is powered off before the import operation, it will not be powered back on after the import is completed. If the server is powered on before the import operation, it will be powered off during the import process if this attribute is set to "Forced" or "NoReboot", and will be powered back on after the import is completed if this attribute is set to "Graceful" or "NoReboot".
- `time_to_wait` (Number) Time To Wait (in seconds) - specifies the time to wait for the server configuration profile
				to be imported. This is useful for ensuring that the server is powered off before the import operation, and for waiting
				for the import to complete before powering the server back on. The default value is 1200 seconds (or 20 minutes), but can
				be set to a lower value of 300 seconds (or 5 minutes) upto a max value of 3600 seconds (or 60 minutes) if desired. Note
				that this attribute only applies if the server is powered on before the import operation, or if the server is powered off
				before the import operation and the shutdown type is set to "Graceful" or "NoReboot". The minimum value is 300 seconds, and
				the maximum value is 3600 seconds (or 1 hour). If the server is powered off before the import operation and the shutdown
				type is set to "Forced" or "NoReboot", the import operation will occur immediately and the server will not be powered
				back on after the import is completed.

### Read-Only

- `id` (String) ID of the Import SCP resource

<a id="nestedatt--share_parameters"></a>
### Nested Schema for `share_parameters`

Required:

- `filename` (String) File Name - The name of the server configuration profile file to import. This is the name of the file that was previously exported using the Server Configuration Profile Export operation. This file is typically in the xml/json format
- `share_type` (String) Share Type - The type of share being used to import the Server Configuration Profile file.

Optional:

- `ignore_certificate_warning` (Boolean) Ignore Certificate Warning
- `ip_address` (String) IPAddress - The IP address of the target export server.
- `password` (String, Sensitive) Password - The password for the share server user account. This password is required if the share type is set to "CIFS". It is required only if the share type is set to "CIFS". It is not required if the share type is set to "NFS".
- `port_number` (Number) Port Number - The port number used to communicate with the share server. The default value is 80.
- `proxy_password` (String, Sensitive) The password for the proxy server. This is required if the proxy_support parameter is set to `true`. It is used for authenticating the proxy server credentials.
- `proxy_port` (Number) The port number used by the proxy server. 
			This parameter is optional. 
			If not provided, the default port number (80) is used for the communication with the proxy server.
- `proxy_server` (String) The IP address or hostname of the proxy server.
			 This is the server that acts as a bridge between the iDRAC and the Server Configuration Profile share server. 
			 It is used to communicate with the Server Configuration Profile share server 
			 in order to import the Server Configuration Profile. If the Server Configuration Profile share server
			  is not accessible from the iDRAC directly, then a proxy server must be used in order to establish the connection. 
			  This parameter is optional. 
			  If it is not provided, the Server Configuration Profile import operation
			   will attempt to connect to the Server Configuration Profile share server directly.
- `proxy_support` (Boolean) Proxy Support - Specifies whether or not to use a proxy server for the import operation. If `true`, import operation will use a proxy server for communication with the export server. If `false`, import operation will not use a proxy server for communication with the export server. Default value is `false`.
- `proxy_type` (String) The type of proxy server to be used. The default is "HTTP". If set to "SOCKS4", a SOCKS4 proxy server must be specified. If set to "HTTP", an HTTP proxy server must be specified. If not specified, the Server Configuration Profile import operation will attempt to connect to the Server Configuration Profile share server directly.
- `proxy_username` (String) The username to be used when connecting to the proxy server.
- `share_name` (String) Share Name - The name of the directory or share on the server 
			that contains the Server Configuration Profile file to import.
- `target` (List of String) Filter configuration by target
- `username` (String) Username - The username to use when authenticating with the server
			 that contains the Server Configuration Profile file being imported.


<a id="nestedblock--redfish_server"></a>
### Nested Schema for `redfish_server`

Required:

- `endpoint` (String) Server BMC IP address or hostname

Optional:

- `password` (String, Sensitive) User password for login
- `ssl_insecure` (Boolean) This field indicates whether the SSL/TLS certificate must be verified or not
- `user` (String) User name for login


