package models

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// DellIdracAttributes to construct terraform schema for the idrac attributes resource.
type DellIdracAttributes struct {
	ID            types.String    `tfsdk:"id"`
	RedfishServer []RedfishServer `tfsdk:"redfish_server"`
	Attributes    types.Map       `tfsdk:"attributes"`
}
