package models

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// BiosDatasource to is struct for bios data-source
type BiosDatasource struct {
	ID            types.String    `tfsdk:"id"`
	OdataID       types.String    `tfsdk:"odata_id"`
	RedfishServer []RedfishServer `tfsdk:"redfish_server"`
	Attributes    types.Map       `tfsdk:"attributes"`
}

// Bios is struct to create schema for bios resource
type Bios struct {
	ID                types.String    `tfsdk:"id"`
	Attributes        types.Map       `tfsdk:"attributes"`
	RedfishServer     []RedfishServer `tfsdk:"redfish_server"`
	SettingsApplyTime types.String    `tfsdk:"settings_apply_time"`
	ResetType         types.String    `tfsdk:"reset_type"`
	ResetTimeout      types.Int64     `tfsdk:"reset_timeout"`
	JobTimeout        types.Int64     `tfsdk:"bios_job_timeout"`
}
