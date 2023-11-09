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
