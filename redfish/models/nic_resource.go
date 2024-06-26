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

// NICResource is struct for NIC resource.
type NICResource struct {
	RedfishServer []RedfishServer `tfsdk:"redfish_server"`

	// Required params
	NetworkAdapterID        types.String `tfsdk:"network_adapter_id"`
	NetworkDeviceFunctionID types.String `tfsdk:"network_device_function_id"`
	ApplyTime               types.String `tfsdk:"apply_time"`
	// Optional params
	SystemID               types.String      `tfsdk:"system_id"`
	Networktributes        types.Map         `tfsdk:"network_attributes"`
	OemNetworkAttributes   types.Map         `tfsdk:"oem_network_attributes"`
	OemNetworkClearPending types.Bool        `tfsdk:"oem_network_clear_pending"`
	JobTimeout             types.Int64       `tfsdk:"job_timeout"` //"default": 1200 TTHE
	MaintenanceWindow      MaintenanceWindow `tfsdk:"maintenance_window"`
}

// MaintenanceWindow is struct for maintenance window.
type MaintenanceWindow struct {
	StartTime types.String `tfsdk:"start_time"` // required
	Duration  types.Int64  `tfsdk:"duration"`   // required TTHE
}
