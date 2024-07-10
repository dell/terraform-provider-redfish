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

# system_id is optional. If system_id is not provided, the resource picks the first one from system resources returned by the iDRAC.
terraform import redfish_network_adapter.nic '{"network_adapter_id":"<network_adapter_id>","network_device_function_id":"<network_device_function_id>","username":"<user>","password":"<password>","endpoint":"<endpoint>","ssl_insecure":<true/false>}'

# terraform import with system_id.
terraform import redfish_network_adapter.nic '{"system_id":"<system_id>","network_adapter_id":"<network_adapter_id>","network_device_function_id":"<network_device_function_id>","username":"<user>","password":"<password>","endpoint":"<endpoint>","ssl_insecure":<true/false>}'
