package models

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// SystemBootDataSource struct for datasource
type SystemBootDataSource struct {
	RedfishServer                []RedfishServer `tfsdk:"redfish_server"`
	ResourceID                   types.String    `tfsdk:"id"`
	BootOrder                    types.List      `tfsdk:"boot_order"`
	BootSourceOverrideEnabled    types.String    `tfsdk:"boot_source_override_enabled"`
	BootSourceOverrideMode       types.String    `tfsdk:"boot_source_override_mode"`
	BootSourceOverrideTarget     types.String    `tfsdk:"boot_source_override_target"`
	UefiTargetBootSourceOverride types.String    `tfsdk:"uefi_target_boot_source_override"`
}
