/*
Copyright (c) 2023-2024 Dell Inc., or its subsidiaries. All Rights Reserved.

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

import "github.com/hashicorp/terraform-plugin-framework/types"

// SCPExport to provide payload for server configuration profile
type SCPExport struct {
	ExportFormat    string          `json:"ExportFormat"`
	ExportUse       string          `json:"ExportUse"`
	IncludeInExport []string        `json:"IncludeInExport"`
	ShareParameters ShareParameters `json:"ShareParameters"`
}

// SCPImport defines the SCP import resource.
type SCPImport struct {
	HostPowerState  string          `json:"HostPowerState"`
	ImportBuffer    string          `json:"ImportBuffer,omitempty"`
	ShutdownType    string          `json:"ShutdownType,omitempty"`
	TimeToWait      int64           `json:"TimeToWait,omitempty"`
	ShareParameters ShareParameters `json:"ShareParameters"`
}

// ShareParameters to provide configuration for local/network share type
type ShareParameters struct {
	FileName                 string      `json:"FileName"`
	Target                   interface{} `json:"Target,omitempty"`
	IPAddress                string      `json:"IPAddress,omitempty"`
	IgnoreCertificateWarning string      `json:"IgnoreCertificateWarning,omitempty"`
	Password                 string      `json:"Password,omitempty"`
	PortNumber               string      `json:"PortNumber,omitempty"`
	ProxyPassword            string      `json:"ProxyPassword,omitempty"`
	ProxyPort                string      `json:"ProxyPort,omitempty"`
	ProxyServer              string      `json:"ProxyServer,omitempty"`
	ProxySupport             string      `json:"ProxySupport,omitempty"`
	ProxyType                string      `json:"ProxyType,omitempty"`
	ProxyUserName            string      `json:"ProxyUserName,omitempty"`
	ShareName                string      `json:"ShareName,omitempty"`
	ShareType                string      `json:"ShareType,omitempty"`
	Username                 string      `json:"Username,omitempty"`
	Workgroup                string      `json:"Workgroup,omitempty"`
}

// RedfishScpImport defines the SCP import resource.
type RedfishScpImport struct {
	ID             types.String    `tfsdk:"id"`
	RedfishServer  []RedfishServer `tfsdk:"redfish_server"`
	HostPowerState types.String    `tfsdk:"host_power_state"`
	ImportBuffer   types.String    `tfsdk:"import_buffer"`
	ShutdownType   types.String    `tfsdk:"shutdown_type"`
	TimeToWait     types.Int64     `tfsdk:"time_to_wait"`
	// ShareParameters Object of type TFShareParameters
	ShareParameters types.Object `tfsdk:"share_parameters"`
}

// TFRedfishScpExport is the tfsdk model of ScpExport
type TFRedfishScpExport struct {
	ID              types.String    `tfsdk:"id"`
	RedfishServer   []RedfishServer `tfsdk:"redfish_server"`
	FileContent     types.String    `tfsdk:"file_content"`
	ExportFormat    types.String    `tfsdk:"export_format"`
	ExportUse       types.String    `tfsdk:"export_use"`
	IncludeInExport types.List      `tfsdk:"include_in_export"`
	// ShareParameters Object of type TFShareParameters
	ShareParameters types.Object `tfsdk:"share_parameters"`
}

// TFShareParameters to provide configuration for local/network share type
type TFShareParameters struct {
	FileName                 types.String `tfsdk:"filename"`
	IPAddress                types.String `tfsdk:"ip_address"`
	IgnoreCertificateWarning types.Bool   `tfsdk:"ignore_certificate_warning"`
	Password                 types.String `tfsdk:"password"`
	PortNumber               types.Int64  `tfsdk:"port_number"`
	ProxyPassword            types.String `tfsdk:"proxy_password"`
	ProxyPort                types.Int64  `tfsdk:"proxy_port"`
	ProxyServer              types.String `tfsdk:"proxy_server"`
	ProxySupport             types.Bool   `tfsdk:"proxy_support"`
	ProxyType                types.String `tfsdk:"proxy_type"`
	ProxyUserName            types.String `tfsdk:"proxy_username"`
	ShareName                types.String `tfsdk:"share_name"`
	ShareType                types.String `tfsdk:"share_type"`
	Target                   types.List   `tfsdk:"target"`
	Username                 types.String `tfsdk:"username"`
}
