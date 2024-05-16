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

// BiosDatasource to is struct for bios data-source
type BiosDatasource struct {
	ID            types.String    `tfsdk:"id"`
	OdataID       types.String    `tfsdk:"odata_id"`
	RedfishServer []RedfishServer `tfsdk:"redfish_server"`
	Attributes    types.Map       `tfsdk:"attributes"`
	BootOptions   types.List      `tfsdk:"boot_options"`
}

// Bios is struct to create schema for bios resource
type Bios struct {
	ID                types.String    `tfsdk:"id"`
	Attributes        types.Map       `tfsdk:"attributes"`
	RedfishServer     []RedfishServer `tfsdk:"redfish_server"`
	SettingsApplyTime types.String    `tfsdk:"settings_apply_time"`
	ResetType         types.String    `tfsdk:"reset_type"`
	ResetTimeout      types.Int64     `tfsdk:"reset_timeout"`
	JobTimeout        types.Int64     `tfsdk:"bios_job_timeout"`
}
