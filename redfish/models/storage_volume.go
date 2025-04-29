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

import "github.com/hashicorp/terraform-plugin-framework/types"

// RedfishStorageVolume is struct for storage volume resource
type RedfishStorageVolume struct {
	CapacityBytes       types.Int64     `tfsdk:"capacity_bytes"`
	DiskCachePolicy     types.String    `tfsdk:"disk_cache_policy"`
	RaidType            types.String    `tfsdk:"raid_type"`
	Drives              types.List      `tfsdk:"drives"`
	ID                  types.String    `tfsdk:"id"`
	RedfishServer       []RedfishServer `tfsdk:"redfish_server"`
	OptimumIoSizeBytes  types.Int64     `tfsdk:"optimum_io_size_bytes"`
	ReadCachePolicy     types.String    `tfsdk:"read_cache_policy"`
	ResetTimeout        types.Int64     `tfsdk:"reset_timeout"`
	ResetType           types.String    `tfsdk:"reset_type"`
	SettingsApplyTime   types.String    `tfsdk:"settings_apply_time"`
	StorageControllerID types.String    `tfsdk:"storage_controller_id"`
	VolumeJobTimeout    types.Int64     `tfsdk:"volume_job_timeout"`
	VolumeName          types.String    `tfsdk:"volume_name"`
	VolumeType          types.String    `tfsdk:"volume_type"`
	WriteCachePolicy    types.String    `tfsdk:"write_cache_policy"`
	Encrypted           types.Bool      `tfsdk:"encrypted"`
	SystemID            types.String    `tfsdk:"system_id"`
}
