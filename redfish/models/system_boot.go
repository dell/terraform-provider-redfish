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

package models

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// SystemBootDataSource struct for datasource
type SystemBootDataSource struct {
	RedfishServer                []RedfishServer `tfsdk:"redfish_server"`
	ID                           types.String    `tfsdk:"id"`
	SystemID                     types.String    `tfsdk:"system_id"`
	BootOrder                    types.List      `tfsdk:"boot_order"`
	BootSourceOverrideEnabled    types.String    `tfsdk:"boot_source_override_enabled"`
	BootSourceOverrideMode       types.String    `tfsdk:"boot_source_override_mode"`
	BootSourceOverrideTarget     types.String    `tfsdk:"boot_source_override_target"`
	UefiTargetBootSourceOverride types.String    `tfsdk:"uefi_target_boot_source_override"`
}
