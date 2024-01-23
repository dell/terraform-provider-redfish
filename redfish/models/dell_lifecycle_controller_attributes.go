package models

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// DellLCAttributes to construct terraform schema for the lifecycle controller attributes resource.
type DellLCAttributes struct {
	ID            types.String    `tfsdk:"id"`
	RedfishServer []RedfishServer `tfsdk:"redfish_server"`
	Attributes    types.Map       `tfsdk:"attributes"`
}
