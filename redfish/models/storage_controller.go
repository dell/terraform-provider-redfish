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
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// StorageControllerResource is struct for StorageController resource.
type StorageControllerResource struct {
	RedfishServer []RedfishServer `tfsdk:"redfish_server"`
	ID            types.String    `tfsdk:"id"`
	// Required params
	StorageID    types.String `tfsdk:"storage_id"`
	ControllerID types.String `tfsdk:"controller_id"`
	ApplyTime    types.String `tfsdk:"apply_time"`
	// Optional params
	JobTimeout        types.Int64        `tfsdk:"job_timeout"`
	ResetType         types.String       `tfsdk:"reset_type"`
	ResetTimeout      types.Int64        `tfsdk:"reset_timeout"`
	MaintenanceWindow *MaintenanceWindow `tfsdk:"maintenance_window"`
	SystemID          types.String       `tfsdk:"system_id"`
	StorageController types.Object       `tfsdk:"storage_controller"`
	Security          types.Object       `tfsdk:"security"`
}

// SecurityAttributes is the struct for security.
type SecurityAttributes struct {
	Action types.String `tfsdk:"action"`
	KeyID  types.String `tfsdk:"key_id"`
	Key    types.String `tfsdk:"key"`
	OldKey types.String `tfsdk:"old_key"`
	Mode   types.String `tfsdk:"mode"`
}

// StorageControllerAttributes is the struct for storage controller attributes.
type StorageControllerAttributes struct {
	ControllerRates types.Object `tfsdk:"controller_rates"`
	Oem             types.Object `tfsdk:"oem"`
}

// OEMAttributes is the struct for OEM Attributes.
type OEMAttributes struct {
	Dell types.Object `tfsdk:"dell"`
}

// DellAttributes is the struct for Dell Attributes.
type DellAttributes struct {
	DellStorageController types.Object `tfsdk:"dell_storage_controller"`
}

// DellStorageControllerAttributes is the struct for Dell Storage Controller Attributes.
type DellStorageControllerAttributes struct {
	ControllerMode                             types.String `tfsdk:"controller_mode"`
	CheckConsistencyMode                       types.String `tfsdk:"check_consistency_mode"`
	CopybackMode                               types.String `tfsdk:"copyback_mode"`
	LoadBalanceMode                            types.String `tfsdk:"load_balance_mode"`
	EnhancedAutoImportForeignConfigurationMode types.String `tfsdk:"enhanced_auto_import_foreign_configuration_mode"`
	PatrolReadUnconfiguredAreaMode             types.String `tfsdk:"patrol_read_unconfigured_area_mode"`
	PatrolReadMode                             types.String `tfsdk:"patrol_read_mode"`
	BackgroundInitializationRatePercent        types.Int64  `tfsdk:"background_initialization_rate_percent"`
	ReconstructRatePercent                     types.Int64  `tfsdk:"reconstruct_rate_percent"`
}

// StorageControllerDatasource is struct for StorageController data-source.
type StorageControllerDatasource struct {
	ID                      types.String             `tfsdk:"id"`
	RedfishServer           []RedfishServer          `tfsdk:"redfish_server"`
	StorageControllerFilter *StorageControllerFilter `tfsdk:"storage_controller_filter"`
	StorageControllers      []StorageController      `tfsdk:"storage_controllers"`
}

// StorageControllerFilter is the tfsdk model of StorageControllerFilter.
type StorageControllerFilter struct {
	Systems []SystemsFilter `tfsdk:"systems"`
}

// SystemsFilter is the tfsdk model of SystemsFilter.
type SystemsFilter struct {
	SystemID types.String     `tfsdk:"system_id"`
	Storages []StoragesFilter `tfsdk:"storages"`
}

// StoragesFilter is the tfsdk model of StoragesFilter.
type StoragesFilter struct {
	StorageID     types.String   `tfsdk:"storage_id"`
	ControllerIDs []types.String `tfsdk:"controller_ids"`
}

// StorageController is the tfsdk model of StorageController.
type StorageController struct {
	ODataID                      types.String         `tfsdk:"odata_id"`
	ID                           types.String         `tfsdk:"id"`
	Description                  types.String         `tfsdk:"description"`
	Name                         types.String         `tfsdk:"name"`
	Assembly                     Assembly             `tfsdk:"assembly"`
	CacheSummary                 CacheSummary         `tfsdk:"cache_summary"`
	ControllerRates              ControllerRates      `tfsdk:"controller_rates"`
	FirmwareVersion              types.String         `tfsdk:"firmware_version"`
	Identifiers                  []Identifier         `tfsdk:"identifiers"`
	Links                        Links                `tfsdk:"links"`
	Manufacturer                 types.String         `tfsdk:"manufacturer"`
	Model                        types.String         `tfsdk:"model"`
	Oem                          StorageControllerOEM `tfsdk:"oem"`
	SpeedGbps                    types.Float64        `tfsdk:"speed_gbps"`
	Status                       Status               `tfsdk:"status"`
	SupportedControllerProtocols []types.String       `tfsdk:"supported_controller_protocols"`
	SupportedDeviceProtocols     []types.String       `tfsdk:"supported_device_protocols"`
	SupportedRAIDTypes           []types.String       `tfsdk:"supported_raid_types"`
}

// Assembly is the tfsdk model of Assembly.
type Assembly struct {
	ODataID types.String `tfsdk:"odata_id"`
}

// ControllerRates is the tfsdk model of ControllerRates.
type ControllerRates struct {
	ConsistencyCheckRatePercent types.Int64 `tfsdk:"consistency_check_rate_percent"`
	RebuildRatePercent          types.Int64 `tfsdk:"rebuild_rate_percent"`
}

// Identifier is the tfsdk model of Identifier.
type Identifier struct {
	DurableName       types.String `tfsdk:"durable_name"`
	DurableNameFormat types.String `tfsdk:"durable_name_format"`
}

// Links is the tfsdk model of Links.
type Links struct {
	PCIeFunctions []PCIeFunction `tfsdk:"pcie_functions"`
}

// PCIeFunction is the tfsdk model of PCIeFunction.
type PCIeFunction struct {
	ODataID types.String `tfsdk:"odata_id"`
}

// StorageControllerOEM is the tfsdk model of StorageControllerOEM.
type StorageControllerOEM struct {
	Dell StorageControllerOEMDell `tfsdk:"dell"`
}

// StorageControllerOEMDell is the tfsdk model of StorageControllerOEMDell.
type StorageControllerOEMDell struct {
	DellStorageController DellStorageController `tfsdk:"dell_storage_controller"`
}

// DellStorageController is the tfsdk model of DellStorageController.
type DellStorageController struct {
	AlarmState                                 types.String   `tfsdk:"alarm_state"`
	AutoConfigBehavior                         types.String   `tfsdk:"auto_config_behavior"`
	BackgroundInitializationRatePercent        types.Int64    `tfsdk:"background_initialization_rate_percent"`
	BatteryLearnMode                           types.String   `tfsdk:"battery_learn_mode"`
	BootVirtualDiskFQDD                        types.String   `tfsdk:"boot_virtual_disk_fqdd"`
	CacheSizeInMB                              types.Int64    `tfsdk:"cache_size_in_mb"`
	CachecadeCapability                        types.String   `tfsdk:"cachecade_capability"`
	CheckConsistencyMode                       types.String   `tfsdk:"check_consistency_mode"`
	ConnectorCount                             types.Int64    `tfsdk:"connector_count"`
	ControllerBootMode                         types.String   `tfsdk:"controller_boot_mode"`
	ControllerFirmwareVersion                  types.String   `tfsdk:"controller_firmware_version"`
	ControllerMode                             types.String   `tfsdk:"controller_mode"`
	CopybackMode                               types.String   `tfsdk:"copyback_mode"`
	CurrentControllerMode                      types.String   `tfsdk:"current_controller_mode"`
	Device                                     types.String   `tfsdk:"device"`
	DeviceCardDataBusWidth                     types.String   `tfsdk:"device_card_data_bus_width"`
	DeviceCardSlotLength                       types.String   `tfsdk:"device_card_slot_length"`
	DeviceCardSlotType                         types.String   `tfsdk:"device_card_slot_type"`
	DriverVersion                              types.String   `tfsdk:"driver_version"`
	EncryptionCapability                       types.String   `tfsdk:"encryption_capability"`
	EncryptionMode                             types.String   `tfsdk:"encryption_mode"`
	EnhancedAutoImportForeignConfigurationMode types.String   `tfsdk:"enhanced_auto_import_foreign_configuration_mode"`
	KeyID                                      types.String   `tfsdk:"key_id"`
	LastSystemInventoryTime                    types.String   `tfsdk:"last_system_inventory_time"`
	LastUpdateTime                             types.String   `tfsdk:"last_update_time"`
	LoadBalanceMode                            types.String   `tfsdk:"load_balance_mode"`
	MaxAvailablePCILinkSpeed                   types.String   `tfsdk:"max_available_pci_link_speed"`
	MaxDrivesInSpanCount                       types.Int64    `tfsdk:"max_drives_in_span_count"`
	MaxPossiblePCILinkSpeed                    types.String   `tfsdk:"max_possible_pci_link_speed"`
	MaxSpansInVolumeCount                      types.Int64    `tfsdk:"max_spans_in_volume_count"`
	MaxSupportedVolumesCount                   types.Int64    `tfsdk:"max_supported_volumes_count"`
	PCISlot                                    types.String   `tfsdk:"pci_slot"`
	PatrolReadIterationsCount                  types.Int64    `tfsdk:"patrol_read_iterations_count"`
	PatrolReadMode                             types.String   `tfsdk:"patrol_read_mode"`
	PatrolReadRatePercent                      types.Int64    `tfsdk:"patrol_read_rate_percent"`
	PatrolReadState                            types.String   `tfsdk:"patrol_read_state"`
	PatrolReadUnconfiguredAreaMode             types.String   `tfsdk:"patrol_read_unconfigured_area_mode"`
	PersistentHotspare                         types.String   `tfsdk:"persistent_hotspare"`
	PersistentHotspareMode                     types.String   `tfsdk:"persistent_hotspare_mode"`
	RAIDMode                                   types.String   `tfsdk:"raid_mode"`
	RealtimeCapability                         types.String   `tfsdk:"real_time_capability"`
	ReconstructRatePercent                     types.Int64    `tfsdk:"reconstruct_rate_percent"`
	RollupStatus                               types.String   `tfsdk:"rollup_status"`
	SASAddress                                 types.String   `tfsdk:"sas_address"`
	SecurityStatus                             types.String   `tfsdk:"security_status"`
	SharedSlotAssignmentAllowed                types.String   `tfsdk:"shared_slot_assignment_allowed"`
	SlicedVDCapability                         types.String   `tfsdk:"sliced_vd_capability"`
	SpindownIdleTimeSeconds                    types.Int64    `tfsdk:"spindown_idle_time_seconds"`
	SupportControllerBootMode                  types.String   `tfsdk:"support_controller_boot_mode"`
	SupportEnhancedAutoForeignImport           types.String   `tfsdk:"support_enhanced_auto_foreign_import"`
	SupportRAID10UnevenSpans                   types.String   `tfsdk:"support_raid10_uneven_spans"`
	SupportedInitializationTypes               []types.String `tfsdk:"supported_initialization_types"`
	SupportsLKMtoSEKMTransition                types.String   `tfsdk:"supports_lkm_to_sekm_transition"`
	T10PICapability                            types.String   `tfsdk:"t10_pi_capability"`
}
