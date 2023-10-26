package models

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// BiosDatasource to construct terraform schema for the idrac attributes resource.
type BiosDatasource struct {
	ID            types.String  `tfsdk:"id"`
	OdataID       types.String  `tfsdk:"odata_id"`
	RedfishServer RedfishServer `tfsdk:"redfish_server"`
	Attributes    types.Map     `tfsdk:"attributes"`
}
