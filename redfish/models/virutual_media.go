package models

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type VirtualMedia struct {
	VirtualMediaID       types.String  `tfsdk:"id"`
	RedfishServer        RedfishServer `tfsdk:"redfish_server"`
	Image                types.String  `tfsdk:"image"`
	Inserted             types.Bool    `tfsdk:"inserted"`
	TransferMethod       types.String  `tfsdk:"transfer_method"`
	TransferProtocolType types.String  `tfsdk:"transfer_protocol_type"`
	WriteProtected       types.Bool    `tfsdk:"write_protected"`
}

// type TransferMethodType struct {
// 	Stream types.String `tfsdk:"stream"`
// 	Upload types.String `tfsdk:"upload"`
// }

// type TransferProtocolType struct {
// 	CIFSTransferProtocolType  types.String `tfsdk:"CIFS"`
// 	FTPTransferProtocolType   types.String `tfsdk:"FTP"`
// 	SFTPTransferProtocolType  types.String `tfsdk:"SFTP"`
// 	HTTPTransferProtocolType  types.String `tfsdk:"HTTP"`
// 	HTTPSTransferProtocolType types.String `tfsdk:"HTTPS"`
// 	NFSTransferProtocolType   types.String `tfsdk:"NFS"`
// 	SCPTransferProtocolType   types.String `tfsdk:"SCP"`
// 	TFTPTransferProtocolType  types.String `tfsdk:"TFTP"`
// 	OEMTransferProtocolType   types.String `tfsdk:"OEM"`
// }

// TransferMethod is how the data is transferred.
// type TransferMethod string

// const (

// 	// StreamTransferMethod Stream image file data from the source URI.
// 	StreamTransferMethod TransferMethod = "Stream"
// 	// UploadTransferMethod Upload the entire image file from the source URI
// 	// to the service.
// 	UploadTransferMethod TransferMethod = "Upload"
// )
