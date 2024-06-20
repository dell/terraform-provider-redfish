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

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// SimpleUpdateRes is struct for simple update resource
type SimpleUpdateRes struct {
	Id            types.String    `tfsdk:"id"`
	RedfishServer []RedfishServer `tfsdk:"redfish_server"`
	Protocol      types.String    `tfsdk:"transfer_protocol"`
	Image         types.String    `tfsdk:"target_firmware_image"`
	ResetType     types.String    `tfsdk:"reset_type"`
	ResetTimeout  types.Int64     `tfsdk:"reset_timeout"`
	JobTimeout    types.Int64     `tfsdk:"simple_update_job_timeout"`
	SoftwareId    types.String    `tfsdk:"software_id"`
	Version       types.String    `tfsdk:"version"`
	SystemID      types.String    `tfsdk:"system_id"`
}
