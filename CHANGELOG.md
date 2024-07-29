<!--
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
-->

All notable changes to this project will be documented in this file.
This project adheres to [Semantic Versioning](http://semver.org/spec/v2.0.0.html)

# v1.4.0 (July, 2024)
## Release Summary
The release supports resources and data sources mentioned in the Features section for RedFish.
## Features

### Resources
* Server NIC Resource

### DataSources
* Server NIC DataSource

### Others
* Added session based authentication for the provider.
* System ID is added for few resources and datasources.

## Enhancements
N/A

## Bug Fixes
N/A

# v1.3.0 (May, 2024)
## Release Summary
The release supports resources and data sources mentioned in the Features section for RedFish.
## Features

### Resources
* Server Configuration Profile Resource
* User Account Password Resource

### Others
* Added a common [example](docs/guides/using_export_and_import_scp.md) for Server Configuration Profile Import and Export.

## Enhancements
N/A

## Bug Fixes
N/A

# v1.2.0 (March, 2024)
## Release Summary
The release supports resources and data sources mentioned in the Features section for RedFish.
## Features

### Resources
* Lifecycle Controller Attributes Resource
* System Attributes Resource
* iDRAC Firmware Update Resource

### Others
* Added Import functionality for Bios, iDRAC Attributes, Virtual Media and Boot Order resources.

## Enhancements
* Added storage controller details to storage datasource
* Enhancements in storage resource
    * Added `encrypted` field. It requires firmware versions 6 or greater and self encrypting disk(s).
    * Added `raid_type` field for setting the RAID level. This replaces the deprecated `volume_type` attribute.
    * Added support for creating storage volumes on BOSS controllers. Volumes on BOSS controllers must take up the whole disk. The fields `capacity_bytes` and `optimum_io_size_bytes` must not be set for these volumes.

## Bug Fixes
N/A

# v1.1.0 (December, 2023)
## Release Summary
The release supports resources and data sources mentioned in the Features section for RedFish.
## Features

### Resources
* Certificate Resource
* Boot Order Resource
* Boot Source Override Resource
* Manager Reset

### Others
N/A

## Enhancements
N/A

## Bug Fixes
N/A

# v1.0.0 (September, 2023)
## Release Summary
The release supports resources and data sources mentioned in the Features section for RedFish.
## Features

### Resources
* `redfish_bios` for configuring BIOS attributes.
* `redfish_dell_idrac_attributes` for confugiring Dell iDRAC attributes.
* `redfish_power` for managing the power state of PowerEdge server.
* `redfish_simple_update` for managing firmware updates on redfish instance.
* `redfish_storage_volume` for managing storage volumes.
* `redfish_user_account` for managing users.
* `redfish_virtual_media` for attaching/detaching virtual media.

### Data Sources:
* `redfish_bios` for reading bios attributes.
* `redfish_dell_idrac_attributes` for reading Dell iDRAC attributes.
* `redfish_firmware_inventory` for reading firware inventory details.
* `redfish_storage` for reading storage volume details.
* `redfish_system_boot` for reading system boot details.
* `redfish_virtual_media` for reading virtual media details.

### Notes:
* `write_protected` attribute of virtual media can only be configured as “true”.
* `capacity_bytes` and `volume_type` attributes of the storage volume cannot be updated.

# [v0.2.0]
#### Major Changes
- [resource_redfish_virtual_media](https://github.com/dell/terraform-provider-redfish/blob/v0.2.0/redfish/resource_redfish_virtual_media.go) - Redfish provider resource for provisionining a server BMC Virtual Media resources for e.g. insert or remove an ISO or USB image as a virtual media device
- [data_source_redfish_virtual_media](https://github.com/dell/terraform-provider-redfish/blob/v0.2.0/redfish/data_source_redfish_virtual_media.go) - data source for server BMC Virtual Media resource
- [resource_redfish_power](https://github.com/dell/terraform-provider-redfish/blob/v0.2.0/redfish/resource_redfish_power.go) - Power cycle a server such as On, Off, GracefulRestart, ForceRestart, PowerCycle etc.
- [resource_simple_update](https://github.com/dell/terraform-provider-redfish/blob/v0.2.0/redfish/resource_simple_update.go) - Redfish provider resource for update a component firmware version on the server
- [data_source_redfish_firmware_inventory](https://github.com/dell/terraform-provider-redfish/blob/v0.2.0/redfish/data_source_redfish_firmware_inventory.go) - Redfish provider data source for getting the components' firmware version
- [data_source_redfish_storage](https://github.com/dell/terraform-provider-redfish/blob/v0.2.0/redfish/data_source_redfish_storage.go) - Redfish data source for storage volumes

#### Bug fixes & Enhancements
- [#31](https://github.com/dell/terraform-provider-redfish/pull/31) - Add support for OnReset operations (previously only supported immediate)
- [#19](https://github.com/dell/terraform-provider-redfish/issues/19) - Bios update is not idempotent
- [#27](https://github.com/dell/terraform-provider-redfish/pull/27) - Add bios change job wait using StateChangeConf method
