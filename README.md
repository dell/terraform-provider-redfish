<!--
Copyright (c) 2020-2024 Dell Inc., or its subsidiaries. All Rights Reserved.

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
# Terraform Provider for RedFish

[![Contributor Covenant](https://img.shields.io/badge/Contributor%20Covenant-v2.0%20adopted-ff69b4.svg)](about/CODE_OF_CONDUCT.md)
[![License](https://img.shields.io/badge/License-MPL_2.0-blue.svg)](LICENSE)

The Terraform Provider for RedFish allows data center and IT administrators to use Hashicorp Terraform to automate and orchestrate the provisioning and management of PowerEdge servers.

The Terraform Provider can be used to manage server power cycles, IDRAC attributes, BIOS attributes, virtual media, storage volumes, user support, and firmware updates on the server.

## Table of contents

* [Code of Conduct](https://github.com/dell/dell-terraform-providers/blob/main/docs/CODE_OF_CONDUCT.md)
* [Maintainer Guide](https://github.com/dell/dell-terraform-providers/blob/main/docs/MAINTAINER_GUIDE.md)
* [Committer Guide](https://github.com/dell/dell-terraform-providers/blob/main/docs/COMMITTER_GUIDE.md)
* [Contributing Guide](https://github.com/dell/dell-terraform-providers/blob/main/docs/CONTRIBUTING.md)
* [List of Adopters](https://github.com/dell/dell-terraform-providers/blob/main/docs/ADOPTERS.md)
* [Support](#support)
* [Security](https://github.com/dell/dell-terraform-providers/blob/main/docs/SECURITY.md)
* [License](#license)
* [Prerequisites](#prerequisites)
* [List of DataSources in Terraform Provider for RedFish](#list-of-datasources-in-terraform-provider-for-redfish)
* [List of Resources in Terraform Provider for RedFish](#list-of-resources-in-terraform-provider-for-redfish)
* [Releasing, Maintenance and Deprecation](#releasing-maintenance-and-deprecation)
* [Documentation](#documentation)
* [New to Terraform?](#new-to-terraform)

## Support
For any Terraform Provider for RedFish issues, questions or feedback, please follow our [support process](https://github.com/dell/dell-terraform-providers/blob/main/docs/SUPPORT.md)

## License
The Terraform Provider for RedFish is released and licensed under the MPL-2.0 license. See [LICENSE](LICENSE) for the full terms.

## Prerequisites

| **Terraform Provider** | **iDRAC9 Firmware Version** | **OS** | **Terraform** | **Golang** |
|---------------------|-----------------------|-------|--------------------|--------------------------|
| v1.5.0 | 5.x <br> 6.x <br> 7.x | ubuntu22.04 <br> rhel9.x | 1.8.x <br> 1.9.x | 1.22

## List of DataSources in Terraform Provider for RedFish
  * [Bios](docs/data-sources/bios.md)
  * [iDRAC Attributes](docs/data-sources/dell_idrac_attributes.md)
  * [Firmware Inventory](docs/data-sources/firmware_inventory.md)
  * [Storage](docs/data-sources/storage.md)
  * [System Boot](docs/data-sources/system_boot.md)
  * [Virtual Media](docs/data-sources/virtual_media.md)
  * [Server NIC](docs/data-sources/network.md)
  * [Storage Controller](docs/data-sources/storage_controller.md)
  * [Directory Service Auth Provider](docs/data-sources/directory_service_auth_provider.md)

## List of Resources in Terraform Provider for RedFish
  * [Bios](docs/resources/bios.md)
  * [iDRAC Attributes](docs/resources/dell_idrac_attributes.md)
  * [Lifecycle Controller Attributes](docs/resources/dell_lc_attributes.md)
  * [System Attributes](docs/resources/dell_system_attributes.md)
  * [Power](docs/resources/power.md)
  * [Simple Update](docs/resources/simple_update.md)
  * [Storage Volume](docs/resources/storage_volume.md)
  * [User Account](docs/resources/user_account.md)
  * [Virtual Media](docs/resources/virtual_media.md)
  * [Manager reset](docs/resources/manager_reset.md)
  * [Boot Order](docs/resources/boot_order.md)
  * [Boot Source Override](docs/resources/boot_source_override.md)
  * [Certificate](docs/resources/certificate.md)
  * [iDRAC Firmware Update](docs/resources/idrac_firmware_update.md)
  * [User Account Password](docs/resources/user_account_password.md)
  * [Server Configuration Profile Export](docs/resources/idrac_server_configuration_profile_export.md)
  * [Server Configuration Profile Import](docs/resources/idrac_server_configuration_profile_import.md)
  * [Server NIC](docs/resources/network_adapter.md)
  * [Storage Controller](docs/resources/storage_controller.md)
  * [Directory Service Auth Provider](docs/resources/directory_service_auth_provider.md)

## Installation and execution of Terraform Provider for RedFish
The installation and execution steps of Terraform Provider for Dell RedFish can be found [here](about/INSTALLATION.md).

## Releasing, Maintenance and Deprecation

Terraform Provider for RedFish follows [Semantic Versioning](https://semver.org/).

New versions will be release regularly if significant changes (bug fix or new feature) are made in the provider.

Released code versions are located on tags in the form of "vx.y.z" where x.y.z corresponds to the version number.

## Documentation

For more detailed information, please refer to [Dell Terraform Providers Documentation](https://dell.github.io/terraform-docs/).

## Terraform Redfish Modules

**Check the following links for the terraform-modules repository and registry**
[Terraform Redfish Modules Github](https://github.com/dell/terraform-redfish-modules)
[Terraform Redfish Modules Registry](https://registry.terraform.io/modules/dell/modules/redfish/latest)

## New to Terraform?
**Here are some helpful links to get you started if you are new to terraform before using our provider:**

- Intro to Terraform: https://developer.hashicorp.com/terraform/intro 
- Providers: https://developer.hashicorp.com/terraform/language/providers 
- Resources: https://developer.hashicorp.com/terraform/language/resources
- Datasources: https://developer.hashicorp.com/terraform/language/data-sources
- Import: https://developer.hashicorp.com/terraform/language/import
- Variables: https://developer.hashicorp.com/terraform/language/values/variables
- Modules: https://developer.hashicorp.com/terraform/language/modules
- State: https://developer.hashicorp.com/terraform/language/state
