package models

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// FirmwareInventory struct is created using this
type FirmwareInventory struct {
	ID            types.String    `tfsdk:"id"`
	OdataID       types.String    `tfsdk:"odata_id"`
	RedfishServer []RedfishServer `tfsdk:"redfish_server"`
	Inventory     []Inventory     `tfsdk:"inventory"`
}

// Inventory struct is created which is used in firmware inventory
type Inventory struct {
	EntityName types.String `tfsdk:"entity_name"`
	EntityId   types.String `tfsdk:"entity_id"`
	Version    types.String `tfsdk:"version"`
}
