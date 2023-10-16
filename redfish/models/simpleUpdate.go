package models

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type SimpleUpdateRes struct {
	Id            types.String  `tfsdk:"id"`
	RedfishServer RedfishServer `tfsdk:"redfish_server"`
	Protocol      types.String  `tfsdk:"transfer_protocol"`
	Image         types.String  `tfsdk:"target_firmware_image"`
	ResetType     types.String  `tfsdk:"reset_type"`
	ResetTimeout  types.Int64   `tfsdk:"reset_timeout"`
	JobTimeout    types.Int64   `tfsdk:"simple_update_job_timeout"`
	SoftwareId    types.String  `tfsdk:"software_id"`
	Version       types.String  `tfsdk:"version"`
}
