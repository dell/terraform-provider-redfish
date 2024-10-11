/*
Copyright (c) 2023-2024 Dell Inc., or its subsidiaries. All Rights Reserved.

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

# import all idrac attributes
terraform import redfish_dell_idrac_attributes.idrac '{"username":"<user>","password":"<password>","endpoint":"<endpoint>","ssl_insecure":<true/false>}'

# import list of idrac attributes
terraform import redfish_dell_idrac_attributes.idrac '{"username":"<user>","password":"<password>","endpoint":"<endpoint>","ssl_insecure":<true/false>, "attributes":["Users.2.UserName"]}'

# terraform import with redfish_alias. When using redfish_alias, provider's `redfish_servers` is required.
# redfish_alias is used to align with enhancements to password management.
terraform import redfish_dell_idrac_attributes.idrac '{"redfish_alias":"<redfish_alias>"}'
