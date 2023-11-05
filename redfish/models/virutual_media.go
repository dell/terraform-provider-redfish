package models

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// VirtualMedia struct
type VirtualMedia struct {
	VirtualMediaID       types.String  `tfsdk:"id"`
	RedfishServer        RedfishServer `tfsdk:"redfish_server"`
	Image                types.String  `tfsdk:"image"`
	Inserted             types.Bool    `tfsdk:"inserted"`
	TransferMethod       types.String  `tfsdk:"transfer_method"`
	TransferProtocolType types.String  `tfsdk:"transfer_protocol_type"`
	WriteProtected       types.Bool    `tfsdk:"write_protected"`
}

// VirtualMediaDataSource struct for datasource
type VirtualMediaDataSource struct {
	ID               types.String       `tfsdk:"id"`
	RedfishServer    []RedfishServer    `tfsdk:"redfish_server"`
	VirtualMediaData []VirtualMediaData `tfsdk:"virtual_media"`
}

// VirtualMediaData to get odata / id of virtual media
type VirtualMediaData struct {
	OdataId types.String `tfsdk:"odata_id"`
	Id      types.String `tfsdk:"id"`
}
