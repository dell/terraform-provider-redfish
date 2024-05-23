---
# Copyright (c) 2024 Dell Inc., or its subsidiaries. All Rights Reserved.

# Licensed under the Mozilla Public License Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at

#     http://mozilla.org/MPL/2.0/


# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
page_title: "Using Export and Import Server Configuration Profile"
title: "Using Export and Import Server Configuration Profile"
linkTitle: "Using Export and Import Server Configuration Profile"
---

The guide provides a terraform configuration of using the Redfish Provider to export and import server configuration profiles. It sets default values for sharing a file using HTTPS, exports the server configuration profile using Redfish, waits for 10 seconds, and then imports the server configuration profile using Redfish.

## Variables
The configuration defines several variables to configure the share parameters for different sharing types.

### Share Type - Local

```terraform
variable "local_setting" {
  default = {
    filename   = "demo_local.xml"
    target     = ["NIC"]
    share_type = "LOCAL"
  }
}
```


The local_setting variable is defined with a default value that specifies the settings for sharing a file locally.

Here's a breakdown of the configuration:

local_setting: This variable is defined with a default value that specifies the settings for sharing a file locally.
default: This block specifies the default value for the local_setting variable.

filename: This specifies the name of the file to be shared. In this example, the file is named "demo_local.xml".

target: This specifies the target devices or systems to which the file will be shared. In this example, the target is "NIC".

share_type: This specifies the type of sharing to be used. In this example, the sharing type is "LOCAL".

Overall, this configuration sets the default values for sharing a file locally using the local_setting variable.


### Share Type - NFS

```terraform
variable "nfs_setting" {
  default = {
    filename   = "demo_nfs.xml"
    target     = ["NIC"]
    share_type = "NFS"
    ip_address = "10.0.0.1"
    share_name = "/nfs/terraform"
  }
}
```

The nfs_setting variable is defined with a default value that specifies the settings for sharing a file using NFS (Network File System).

Here's a breakdown of the configuration:

nfs_setting: This variable is defined with a default value that specifies the settings for sharing a file using NFS.
default: This block specifies the default value for the nfs_setting variable.

filename: This specifies the name of the file to be shared. In this example, the file is named "demo_nfs.xml".

target: This specifies the target devices or systems to which the file will be shared. In this example, the target is "NIC".

share_type: This specifies the type of sharing to be used. In this example, the sharing type is "NFS".

ip_address: This specifies the IP address of the NFS server. In this example, the IP address is "10.0.0.1".

share_name: This specifies the name of the NFS share. In this example, the share name is "/nfs/terraform".

Overall, this configuration sets the default values for sharing a file using NFS using the nfs_setting variable.


### Share Type - CIFS

```terraform
variable "cifs_username" {
    type = string 
    default = "awesomeadmin"
}

variable "cifs_password" {
    type = string 
    default = "Pa$$w0rd"
}

variable "cifs_setting" {
  default = {
    filename   = "demo.xml"
    target     = ["NIC"]
    share_type = "CIFS"
    ip_address = "10.0.0.2"
    share_name = "/cifs/terraform"
    username   = var.cifs_username
    password   = var.cifs_password
  }
}
```

The cifs_setting variable is defined with a default value that specifies the settings for sharing a file using CIFS (Common Internet File System).

Here's a breakdown of the configuration:

cifs_username: This variable is defined with a type of string and a default value of "awesomeadmin". It represents the username for accessing the CIFS share.

cifs_password: This variable is defined with a type of string and a default value of "Pa$$w0rd". It represents the password for accessing the CIFS share.

cifs_setting: This variable is defined with a default value that specifies the settings for sharing a file using CIFS.
default: This block specifies the default value for the cifs_setting variable.

filename: This specifies the name of the file to be shared. In this example, the file is named "demo.xml".

target: This specifies the target devices or systems to which the file will be shared. In this example, the target is "NIC".

share_type: This specifies the type of sharing to be used. In this example, the sharing type is "CIFS".

ip_address: This specifies the IP address of the CIFS server. In this example, the IP address is "10.0.0.2".

share_name: This specifies the name of the CIFS share. In this example, the share name is "/cifs/terraform".

username: This specifies the username for accessing the CIFS share. It is set to the value of the cifs_username variable.

password: This specifies the password for accessing the CIFS share. It is set to the value of the cifs_password variable.

Overall, this configuration sets the default values for sharing a file using CIFS using the cifs_setting variable. It also uses the cifs_username and cifs_password variables to store the username and password for accessing the CIFS share.


### Share Type - HTTPS

```terraform
variable "https_setting" {
  default = {
    filename    = "demo_https.xml"
    target      = ["NIC"]
    share_type  = "HTTPS"
    ip_address  = "10.0.0.3"
    port_number = 443
  }
}
```

The https_setting variable is defined with a default value that specifies the settings for sharing a file using HTTPS (Hypertext Transfer Protocol Secure).

Here's a breakdown of the configuration:

https_setting: This variable is defined with a default value that specifies the settings for sharing a file using HTTPS.
default: This block specifies the default value for the https_setting variable.
filename: This specifies the name of the file to be shared. In this example, the file is named "demo_https.xml".

target: This specifies the target devices or systems to which the file will be shared. In this example, the target is "NIC".

share_type: This specifies the type of sharing to be used. In this example, the sharing type is "HTTPS".

ip_address: This specifies the IP address of the HTTPS server. In this example, the IP address is "10.0.0.3".

port_number: This specifies the port number to be used for the HTTPS connection. In this example, the port number is 443.

Overall, this configuration sets the default values for sharing a file using HTTPS using the https_setting variable. It specifies the IP address and port number for the HTTPS connection.


### Share Type - HTTP

```terraform
variable "http_setting" {
  default = {
    filename      = "demo_http.xml"
    target        = ["NIC"]
    share_type    = "HTTP"
    ip_address    = "10.0.0.4"
    port_number   = 80
    proxy_support = true
    proxy_server  = "10.0.0.5"
    proxy_port    = 45127
  }
}
```

The https_setting variable is defined with a default value that specifies the settings for sharing a file using HTTPS (Hypertext Transfer Protocol Secure).

Here's a breakdown of the configuration:

https_setting: This variable is defined with a default value that specifies the settings for sharing a file using HTTPS.
default: This block specifies the default value for the https_setting variable.

filename: This specifies the name of the file to be shared. In this example, the file is named "demo_https.xml".

target: This specifies the target devices or systems to which the file will be shared. In this example, the target is "NIC".

share_type: This specifies the type of sharing to be used. In this example, the sharing type is "HTTPS".

ip_address: This specifies the IP address of the HTTPS server. In this example, the IP address is "10.0.0.3".

port_number: This specifies the port number to be used for the HTTPS connection. In this example, the port number is 443.

Overall, this configuration sets the default values for sharing a file using HTTPS using the https_setting variable. It specifies the IP address and port number for the HTTPS connection.



## Import and Export Server Configuration Profile 

```terraform
resource "terraform_data" "trigger_by_timestamp" {
  input = timestamp()
}

resource "time_sleep" "wait_10_seconds" {
  depends_on      = [redfish_idrac_server_configuration_profile_export.golden-config]
  create_duration = "10s"
}

resource "redfish_idrac_server_configuration_profile_export" "golden-config" {
  redfish_server {
    user         = var.user
    password     = var.password
    endpoint     = var.endpoint
    ssl_insecure = false
  }
  share_parameters = var.http_setting // change the variable here
  lifecycle {
    replace_triggered_by = [terraform_data.trigger_by_timestamp]
  }
}

resource "redfish_idrac_server_configuration_profile_import" "servers" {
  redfish_server {
    user         = var.user
    password     = var.password
    endpoint     = var.endpoint
    ssl_insecure = true
  }

  import_buffer    = lookup(redfish_idrac_server_configuration_profile_export.golden-config.share_parameters, "share_type") == "LOCAL" ? base64decode(redfish_idrac_server_configuration_profile_export.golden-config.file_content) : null

  share_parameters = redfish_idrac_server_configuration_profile_export.golden-config.share_parameters

  depends_on       = [time_sleep.wait_10_seconds, redfish_idrac_server_configuration_profile_export.golden-config]

  lifecycle {
    replace_triggered_by = [terraform_data.trigger_by_timestamp]
  }
}
```

The provided code snippet defines a Terraform configuration that exports and imports a server configuration profile using the Redfish API.

Here's a breakdown of the configuration:

terraform_data.trigger_by_timestamp: This resource generates a timestamp value and assigns it to the input attribute.

time_sleep.wait_10_seconds: This resource waits for 10 seconds after the redfish_idrac_server_configuration_profile_export.golden-config resource completes.

redfish_idrac_server_configuration_profile_export.golden-config: This resource exports the server configuration profile using the Redfish API. It specifies the Redfish server details (user, password, endpoint, and SSL insecure flag) and the share parameters using the var.<share_type> variable. 

The lifecycle block specifies that the resource should be replaced when the trigger_by_timestamp value changes.

redfish_idrac_server_configuration_profile_import.servers: This resource imports the server configuration profile using the Redfish API. It specifies the Redfish server details and the import buffer, which is the content of the exported configuration profile. 

The lifecycle block specifies that the resource should be replaced when the trigger_by_timestamp value changes. The depends_on block ensures that the import resource waits for the export resource and the 10-second delay before executing.
