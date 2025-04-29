/*
Copyright (c) 2023-2025 Dell Inc., or its subsidiaries. All Rights Reserved.

Licensed under the Mozilla Public License Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://mozilla.org/MPL/2.0/


Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package models

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// VirtualMedia struct
type VirtualMedia struct {
	VirtualMediaID       types.String    `tfsdk:"id"`
	RedfishServer        []RedfishServer `tfsdk:"redfish_server"`
	Image                types.String    `tfsdk:"image"`
	Inserted             types.Bool      `tfsdk:"inserted"`
	TransferMethod       types.String    `tfsdk:"transfer_method"`
	TransferProtocolType types.String    `tfsdk:"transfer_protocol_type"`
	WriteProtected       types.Bool      `tfsdk:"write_protected"`
	SystemID             types.String    `tfsdk:"system_id"`
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
