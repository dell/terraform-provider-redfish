All notable changes to this project will be documented in this file.
This project adheres to [Semantic Versioning](http://semver.org/spec/v2.0.0.html)

# [v0.2.0]
#### Major Changes
- [resource_redfish_virtual_media](https://github.com/dell/terraform-provider-redfish/blob/master/redfish/resource_redfish_virtual_media.go) - Redfish provider resource for provisionining a server BMC Virtual Media resources for e.g. insert or remove an ISO or USB image as a virtual media device
- [data_source_redfish_virtual_media](https://github.com/dell/terraform-provider-redfish/blob/master/redfish/data_source_redfish_virtual_media.go) - data source for server BMC Virtual Media resource
- [resource_redfish_power](https://github.com/dell/terraform-provider-redfish/blob/master/redfish/resource_redfish_power.go) - Power cycle a server such as On, Off, GracefulRestart, ForceRestart, PowerCycle etc.
- [resource_simple_update](https://github.com/dell/terraform-provider-redfish/blob/master/redfish/resource_simple_update.go) - Redfish provider resource for update a component firmware version on the server
- [data_source_redfish_firmware_inventory](https://github.com/dell/terraform-provider-redfish/blob/master/redfish/data_source_redfish_firmware_inventory.go) - Redfish provider data source for getting the components' firmware version
- [data_source_redfish_storage](https://github.com/dell/terraform-provider-redfish/blob/master/redfish/data_source_redfish_storage.go) - Redfish data source for storage volumes

#### Bug fixes & Enhancements
- [#31](https://github.com/dell/terraform-provider-redfish/pull/31) - Add support for OnReset operations (previously only supported immediate)
- [#19](https://github.com/dell/terraform-provider-redfish/issues/19) - Bios update is not idempotent
- [#27](https://github.com/dell/terraform-provider-redfish/pull/27) - Add bios change job wait using StateChangeConf method
