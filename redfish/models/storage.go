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

// StorageDatasource is struct for storage data-source
type StorageDatasource struct {
	ID              types.String    `tfsdk:"id"`
	RedfishServer   []RedfishServer `tfsdk:"redfish_server"`
	Storages        []Storage       `tfsdk:"storage"`
	ControllerIDs   types.List      `tfsdk:"controller_ids"`
	ControllerNames types.List      `tfsdk:"controller_names"`
	SystemID        types.String    `tfsdk:"system_id"`
}

// Storage is the tfsdk model of Storage
type Storage struct {
	ID                 types.String         `tfsdk:"storage_controller_id"`
	Drives             []types.String       `tfsdk:"drives"`
	DriveIDs           []types.String       `tfsdk:"drive_ids"`
	Description        types.String         `tfsdk:"description"`
	Name               types.String         `tfsdk:"name"`
	Oem                Oem                  `tfsdk:"oem"`
	Status             Status               `tfsdk:"status"`
	StorageControllers []StorageControllers `tfsdk:"storage_controllers"`
}

// DellController is the tfsdk model of DellController
type DellController struct {
	AlarmState                       types.String `tfsdk:"alarm_state"`
	AutoConfigBehavior               types.String `tfsdk:"auto_config_behavior"`
	BootVirtualDiskFQDD              types.String `tfsdk:"boot_virtual_disk_fqdd"`
	CacheSizeInMB                    types.Int64  `tfsdk:"cache_size_in_mb"`
	CachecadeCapability              types.String `tfsdk:"cachecade_capability"`
	ConnectorCount                   types.Int64  `tfsdk:"connector_count"`
	ControllerFirmwareVersion        types.String `tfsdk:"controller_firmware_version"`
	CurrentControllerMode            types.String `tfsdk:"current_controller_mode"`
	Description                      types.String `tfsdk:"controller_description"`
	Device                           types.String `tfsdk:"device"`
	DeviceCardDataBusWidth           types.String `tfsdk:"device_card_data_bus_width"`
	DeviceCardSlotLength             types.String `tfsdk:"device_card_slot_length"`
	DeviceCardSlotType               types.String `tfsdk:"device_card_slot_type"`
	DriverVersion                    types.String `tfsdk:"driver_version"`
	EncryptionCapability             types.String `tfsdk:"encryption_capability"`
	EncryptionMode                   types.String `tfsdk:"encryption_mode"`
	ID                               types.String `tfsdk:"controller_id"`
	KeyID                            types.String `tfsdk:"key_id"`
	LastSystemInventoryTime          types.String `tfsdk:"last_system_inventory_time"`
	LastUpdateTime                   types.String `tfsdk:"last_update_time"`
	MaxAvailablePCILinkSpeed         types.String `tfsdk:"max_available_pci_link_speed"`
	MaxPossiblePCILinkSpeed          types.String `tfsdk:"max_possible_pci_link_speed"`
	Name                             types.String `tfsdk:"controller_name"`
	PCISlot                          types.String `tfsdk:"pci_slot"`
	PatrolReadState                  types.String `tfsdk:"patrol_read_state"`
	PersistentHotspare               types.String `tfsdk:"persistent_hotspare"`
	RealtimeCapability               types.String `tfsdk:"realtime_capability"`
	RollupStatus                     types.String `tfsdk:"rollup_status"`
	SASAddress                       types.String `tfsdk:"sas_address"`
	SecurityStatus                   types.String `tfsdk:"security_status"`
	SharedSlotAssignmentAllowed      types.String `tfsdk:"shared_slot_assignment_allowed"`
	SlicedVDCapability               types.String `tfsdk:"sliced_vd_capability"`
	SupportControllerBootMode        types.String `tfsdk:"support_controller_boot_mode"`
	SupportEnhancedAutoForeignImport types.String `tfsdk:"support_enhanced_auto_foreign_import"`
	SupportRAID10UnevenSpans         types.String `tfsdk:"support_raid_10_uneven_spans"`
	SupportsLKMtoSEKMTransition      types.String `tfsdk:"supports_lk_mto_sekm_transition"`
	T10PICapability                  types.String `tfsdk:"t_10_pi_capability"`
}

// DellControllerBattery is the tfsdk model of DellControllerBattery
type DellControllerBattery struct {
	Description   types.String `tfsdk:"controller_battery_description"`
	Fqdd          types.String `tfsdk:"fqdd"`
	ID            types.String `tfsdk:"controller_battery_id"`
	Name          types.String `tfsdk:"controller_battery_name"`
	PrimaryStatus types.String `tfsdk:"primary_status"`
	RAIDState     types.String `tfsdk:"raid_state"`
}

// Dell is the tfsdk model of Dell
type Dell struct {
	DellController        DellController        `tfsdk:"dell_controller"`
	DellControllerBattery DellControllerBattery `tfsdk:"dell_controller_battery"`
}

// Oem is the tfsdk model of Oem
type Oem struct {
	Dell Dell `tfsdk:"dell"`
}

// Status is the tfsdk model of Status
type Status struct {
	Health       types.String `tfsdk:"health"`
	HealthRollup types.String `tfsdk:"health_rollup"`
	State        types.String `tfsdk:"state"`
}

// CacheSummary is the tfsdk model of CacheSummary
type CacheSummary struct {
	TotalCacheSizeMiB types.Int64 `tfsdk:"total_cache_size_mi_b"`
}

// StorageControllers is the tfsdk model of StorageControllers
type StorageControllers struct {
	CacheSummary                 CacheSummary   `tfsdk:"cache_summary"`
	FirmwareVersion              types.String   `tfsdk:"firmware_version"`
	Manufacturer                 types.String   `tfsdk:"manufacturer"`
	Model                        types.String   `tfsdk:"model"`
	Name                         types.String   `tfsdk:"name"`
	SpeedGbps                    types.Int64    `tfsdk:"speed_gbps"`
	Status                       Status         `tfsdk:"status"`
	SupportedControllerProtocols []types.String `tfsdk:"supported_controller_protocols"`
	SupportedDeviceProtocols     []types.String `tfsdk:"supported_device_protocols"`
	SupportedRAIDTypes           []types.String `tfsdk:"supported_raid_types"`
}
