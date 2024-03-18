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

# Odata ID of all available volumes on a storage controller can be fetched by running the following GET request on the iDRAC
# "/redfish/v1/Systems/System.Embedded.1/Storage/<storage controller ID>/Volumes/"
# Eg. redfish/v1/Systems/System.Embedded.1/Storage/RAID.Integrated.1-1/Volumes/
# The ID of any storage controller on the iDRAC, in turn, can be fetched using the "redfish_storage" data source

terraform import redfish_storage_volume.volume "{\"id\":\"<odata id of the volume>\",\"username\":\"<username>\",\"password\":\"<password>\",\"endpoint\":\"<endpoint>\",\"ssl_insecure\":<true/false>}"