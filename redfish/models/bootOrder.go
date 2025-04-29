/*
Copyright (c) 2023-2025 Dell Inc., or its subsidiaries. All Rights Reserved.

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

// BootOrder is strut for configuring boot order and boot options
type BootOrder struct {
	ID            types.String    `tfsdk:"id"`
	BootOptions   types.List      `tfsdk:"boot_options"`
	ResetType     types.String    `tfsdk:"reset_type"`
	ResetTimeout  types.Int64     `tfsdk:"reset_timeout"`
	JobTimeout    types.Int64     `tfsdk:"boot_order_job_timeout"`
	BootOrder     types.List      `tfsdk:"boot_order"`
	RedfishServer []RedfishServer `tfsdk:"redfish_server"`
	SystemID      types.String    `tfsdk:"system_id"`
}

// BootOptions is strut for configuring boot options
type BootOptions struct {
	BootOptionReference types.String `tfsdk:"boot_option_reference"`
	BootOptionEnabled   types.Bool   `tfsdk:"boot_option_enabled"`
}

// BootSourceOverride is struct for configuring boot source override options
type BootSourceOverride struct {
	ID                           types.String    `tfsdk:"id"`
	BootSourceOverrideMode       types.String    `tfsdk:"boot_source_override_mode"`
	BootSourceOverrideEnabled    types.String    `tfsdk:"boot_source_override_enabled"`
	BootSourceOverrideTarget     types.String    `tfsdk:"boot_source_override_target"`
	ResetType                    types.String    `tfsdk:"reset_type"`
	ResetTimeout                 types.Int64     `tfsdk:"reset_timeout"`
	JobTimeout                   types.Int64     `tfsdk:"boot_source_job_timeout"`
	UefiTargetBootSourceOverride types.String    `tfsdk:"uefi_target_boot_source_override"`
	SystemID                     types.String    `tfsdk:"system_id"`
	RedfishServer                []RedfishServer `tfsdk:"redfish_server"`
}
