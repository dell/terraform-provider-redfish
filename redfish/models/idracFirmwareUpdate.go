/*
Copyright (c) 2024 Dell Inc., or its subsidiaries. All Rights Reserved.

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
	"encoding/xml"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

// IdracFirmwareUpdate model for IdracFirmwareUpdateResource
type IdracFirmwareUpdate struct {
	Id                       types.String    `tfsdk:"id"`
	RedfishServer            []RedfishServer `tfsdk:"redfish_server"`
	ShareType                types.String    `tfsdk:"share_type"`
	IPAddress                types.String    `tfsdk:"ip_address"`
	ShareName                types.String    `tfsdk:"share_name"`
	CatalogFileName          types.String    `tfsdk:"catalog_file_name"`
	IgnoreCertificateWarning types.String    `tfsdk:"ignore_cert_warning"`
	ShareUser                types.String    `tfsdk:"share_user"`
	SharePassword            types.String    `tfsdk:"share_password"`
	ProxySupport             types.String    `tfsdk:"proxy_support"`
	ProxyServer              types.String    `tfsdk:"proxy_server"`
	ProxyPort                types.Int64     `tfsdk:"proxy_port"`
	ProxyUsername            types.String    `tfsdk:"proxy_username"`
	ProxyPassword            types.String    `tfsdk:"proxy_password"`
	ProxyType                types.String    `tfsdk:"proxy_type"`
	MountPoint               types.String    `tfsdk:"mount_point"`
	ApplyUpdate              types.Bool      `tfsdk:"apply_update"`
	RebootNeeded             types.Bool      `tfsdk:"reboot_needed"`
	SystemID                 types.String    `tfsdk:"system_id"`
	UpdateList               types.List      `tfsdk:"update_list"`
}

// UpdateListProperty model for UpdateList Property
type UpdateListProperty struct {
	Name  types.String `tfsdk:"name"`
	Value types.String `tfsdk:"value"`
}

// UpdateList model for UpdateList
type UpdateList struct {
	ClassName  types.String         `tfsdk:"class_name"`
	Properties []UpdateListProperty `tfsdk:"properties"`
}

// GetPackageListResponse model for GetPackageListResponse
type GetPackageListResponse struct {
	PackageList string `json:"PackageList"`
}

// Property model for Property
type Property struct {
	Name  string `xml:"NAME,attr"`
	Value string `xml:"VALUE"`
}

// InstanceName model for InstanceName
type InstanceName struct {
	ClassName      string          `xml:"CLASSNAME,attr"`
	Properties     []Property      `xml:"PROPERTY"`
	PropertyArrays []PropertyArray `xml:"PROPERTY.ARRAY"`
}

// PropertyArray model for PropertyArray
type PropertyArray struct {
	Name   string   `xml:"NAME,attr"`
	Values []string `xml:"VALUE.ARRAY>VALUE"`
}

// CIM model for CIM
type CIM struct {
	XMLName   xml.Name       `xml:"CIM"`
	Instances []InstanceName `xml:"MESSAGE>SIMPLEREQ>VALUE.NAMEDINSTANCE>INSTANCENAME"`
}
