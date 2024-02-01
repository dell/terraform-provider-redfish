package models

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// DellSystemAttributes to construct terraform schema for the system attributes resource.
type DellSystemAttributes struct {
	ID            types.String    `tfsdk:"id"`
	RedfishServer []RedfishServer `tfsdk:"redfish_server"`
	Attributes    types.Map       `tfsdk:"attributes"`
}
