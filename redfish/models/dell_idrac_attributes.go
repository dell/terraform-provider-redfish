package models

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type DellIdracAttributes struct {
	ID            types.String  `tfsdk:"id"`
	RedfishServer RedfishServer `tfsdk:"redfish_server"`
	Attributes    types.Map     `tfsdk:"attributes"`
}
