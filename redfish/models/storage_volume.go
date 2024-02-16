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
}
