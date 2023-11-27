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
	RedfishServer                []RedfishServer `tfsdk:"redfish_server"`
}
