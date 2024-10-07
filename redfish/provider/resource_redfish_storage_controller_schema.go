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

package provider

import (
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	redfishcommon "github.com/stmcginnis/gofish/common"
	"github.com/stmcginnis/gofish/redfish"
)

const (
	defaultStorageControllerJobTimeout    int64 = 1200
	defaultStorageControllerResetTimeout  int64 = 120
	intervalStorageControllerJobCheckTime int64 = 10
)

// StorageControllerResourceSchema defines the schema for the Storage Controller resource
func StorageControllerResourceSchema() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"id": schema.StringAttribute{
			MarkdownDescription: "ID of the storage controller resource",
			Description:         "ID of the storage controller resource",
			Computed:            true,
		},
		"storage_id": schema.StringAttribute{
			MarkdownDescription: "ID of the storage",
			Description:         "ID of the storage",
			Required:            true,
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"controller_id": schema.StringAttribute{
			MarkdownDescription: "ID of the storage controller",
			Description:         "ID of the storage controller",
			Required:            true,
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"apply_time": schema.StringAttribute{
			MarkdownDescription: "Apply time of the storage controller attributes. (Update Supported)" +
				"Accepted values: `Immediate`, `OnReset`, `AtMaintenanceWindowStart`, `InMaintenanceWindowOnReset`. " +
				"Immediate: allows the user to immediately reboot the host and apply the changes. " +
				"OnReset: allows the user to apply the changes on the next reboot of the host server." +
				"AtMaintenanceWindowStart: allows the user to apply at the start of a maintenance window as specified in `maintenance_window`." +
				"InMaintenanceWindowOnReset: allows to apply after a manual reset " +
				"but within the maintenance window as specified in `maintenance_window`.",
			Description: "Apply time of the storage controller attributes. (Update Supported)" +
				"Accepted values: `Immediate`, `OnReset`, `AtMaintenanceWindowStart`, `InMaintenanceWindowOnReset`. " +
				"Immediate: allows the user to immediately reboot the host and apply the changes. " +
				"OnReset: allows the user to apply the changes on the next reboot of the host server." +
				"AtMaintenanceWindowStart: allows the user to apply at the start of a maintenance window as specified in `maintenance_window`." +
				"InMaintenanceWindowOnReset: allows to apply after a manual reset " +
				"but within the maintenance window as specified in `maintenance_window`.",
			Required: true,
			Validators: []validator.String{
				stringvalidator.OneOf(
					string(redfishcommon.ImmediateApplyTime),
					string(redfishcommon.OnResetApplyTime),
					string(redfishcommon.AtMaintenanceWindowStartApplyTime),
					string(redfishcommon.InMaintenanceWindowOnResetApplyTime),
				),
			},
		},
		"job_timeout": schema.Int64Attribute{
			MarkdownDescription: "`job_timeout` is the time in seconds that the provider waits for the resource update job to be" +
				"completed before timing out. (Update Supported) Default value is 1200 seconds." +
				"`job_timeout` is applicable only when `apply_time` is `Immediate` or `OnReset`.",
			Description: "`job_timeout` is the time in seconds that the provider waits for the resource update job to be" +
				"completed before timing out. (Update Supported) Default value is 1200 seconds." +
				"`job_timeout` is applicable only when `apply_time` is `Immediate` or `OnReset`.",
			Optional: true,
			Computed: true,
			Default:  int64default.StaticInt64(int64(defaultStorageControllerJobTimeout)),
		},
		"reset_type": schema.StringAttribute{
			MarkdownDescription: "Reset Type. (Update Supported) " +
				"Accepted values: `ForceRestart`, `GracefulRestart`, `PowerCycle`. Default value is `ForceRestart`.",
			Description: "Reset Type. (Update Supported) " +
				"Accepted values: `ForceRestart`, `GracefulRestart`, `PowerCycle`. Default value is `ForceRestart`.",
			Optional: true,
			Computed: true,
			Default:  stringdefault.StaticString(string(redfish.ForceRestartResetType)),
			Validators: []validator.String{
				stringvalidator.OneOf([]string{
					string(redfish.ForceRestartResetType),
					string(redfish.GracefulRestartResetType),
					string(redfish.PowerCycleResetType),
				}...),
			},
		},
		"reset_timeout": schema.Int64Attribute{
			MarkdownDescription: "Reset Timeout. Default value is 120 seconds. (Update Supported)",
			Description:         "Reset Timeout. Default value is 120 seconds. (Update Supported)",
			Optional:            true,
			Computed:            true,
			Default:             int64default.StaticInt64(defaultStorageControllerResetTimeout),
		},
		"maintenance_window": schema.SingleNestedAttribute{
			Description: "This option allows you to schedule the maintenance window. (Update Supported)" +
				"This is required when `apply_time` is `AtMaintenanceWindowStart` or `InMaintenanceWindowOnReset` .",
			MarkdownDescription: "This option allows you to schedule the maintenance window. (Update Supported)" +
				"This is required when `apply_time` is `AtMaintenanceWindowStart` or `InMaintenanceWindowOnReset` .",
			Optional: true,
			Attributes: map[string]schema.Attribute{
				"start_time": schema.StringAttribute{
					Description: "The start time for the maintenance window to be scheduled. (Update Supported)" +
						"The format is YYYY-MM-DDThh:mm:ss<offset>. " +
						"<offset> is the time offset from UTC that the current timezone set in iDRAC in the format: +05:30 for IST.",
					MarkdownDescription: "The start time for the maintenance window to be scheduled. (Update Supported)" +
						"The format is YYYY-MM-DDThh:mm:ss<offset>. " +
						"<offset> is the time offset from UTC that the current timezone set in iDRAC in the format: +05:30 for IST.",
					Required:   true,
					Validators: []validator.String{stringvalidator.LengthAtLeast(1)},
				},
				"duration": schema.Int64Attribute{
					Description:         "The duration in seconds for the maintenance window. (Update Supported)",
					MarkdownDescription: "The duration in seconds for the maintenance window. (Update Supported)",
					Required:            true,
				},
			},
		},
		"system_id": schema.StringAttribute{
			MarkdownDescription: "ID of the system resource. If the value for system ID is not provided, " +
				"the resource picks the first system available from the iDRAC.",
			Description: "ID of the system resource. If the value for system ID is not provided, " +
				"the resource picks the first system available from the iDRAC.",
			Computed:   true,
			Optional:   true,
			Validators: []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"storage_controller": schema.SingleNestedAttribute{
			MarkdownDescription: "This consists of the attributes to configure the storage controller. " +
				"Please update any one out of `storage_controller` and `security` at a time.",
			Description: "This consists of the attributes to configure the storage controller. " +
				"Please update any one out of `storage_controller` and `security` at a time.",
			Optional:   true,
			Computed:   true,
			Attributes: StorageControllerInstanceSchema(),
		},
		"security": schema.SingleNestedAttribute{
			MarkdownDescription: "This consists of the attributes to configure the security of the storage controller. " +
				"Please update any one out of `security` and `storage_controller` at a time.",
			Description: "This consists of the attributes to configure the security of the storage controller. " +
				"Please update any one out of `security` and `storage_controller` at a time.",
			Optional:   true,
			Computed:   true,
			Attributes: SecuritySchema(),
		},
	}
}

// SecuritySchema is a function that returns the schema for Security.
func SecuritySchema() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"action": schema.StringAttribute{
			MarkdownDescription: "Action to create/change/delete the security key. " +
				"Accepted values: `SetControllerKey`, `ReKey`, `RemoveControllerKey`. " +
				"The `SetControllerKey` action is used to set the key on controllers and " +
				"set the controller in Local key Management (LKM) to encrypt the drives. " +
				"The `ReKey` action resets the key on the controller that support encryption of the of drives. " +
				"The `RemoveControllerKey` method erases the encryption key on controller. CAUTION: All encrypted drives shall be erased.",
			Description: "Action to create/change/delete the security key. " +
				"Accepted values: `SetControllerKey`, `ReKey`, `RemoveControllerKey`. " +
				"The `SetControllerKey` action is used to set the key on controllers and " +
				"set the controller in Local key Management (LKM) to encrypt the drives. " +
				"The `ReKey` action resets the key on the controller that support encryption of the of drives. " +
				"The `RemoveControllerKey` method erases the encryption key on controller. CAUTION: All encrypted drives shall be erased.",
			Optional: true,
			Computed: true,
			Validators: []validator.String{stringvalidator.OneOf(
				"SetControllerKey",
				"ReKey",
				"RemoveControllerKey",
			)},
		},
		"key_id": schema.StringAttribute{
			MarkdownDescription: "Key Identifier that describes the key. " +
				"The Key ID shall be maximum of 32 characters in length and should not have any spaces.",
			Description: "Key Identifier that describes the key. " +
				"The Key ID shall be maximum of 32 characters in length and should not have any spaces.",
			Optional: true,
			Computed: true,
		},
		"key": schema.StringAttribute{
			MarkdownDescription: "New controller key.",
			Description:         "New controller key.",
			Optional:            true,
			Computed:            true,
		},
		"old_key": schema.StringAttribute{
			MarkdownDescription: "Old controller key.",
			Description:         "Old controller key.",
			Optional:            true,
			Computed:            true,
		},
		"mode": schema.StringAttribute{
			MarkdownDescription: "Mode of the controller: Local Key Management(LKM)/Secure Enterprise Key Manager(SEKM). " +
				"Accepted values: `LKM`, `SEKM`.",
			Description: "Mode of the controller: Local Key Management(LKM)/Secure Enterprise Key Manager(SEKM). " +
				"Accepted values: `LKM`, `SEKM`.",
			Optional: true,
			Computed: true,
			Validators: []validator.String{stringvalidator.OneOf(
				"LKM",
				"SEKM",
			)},
		},
	}
}

// StorageControllerInstanceSchema is a function that returns the schema for Storage Controller Instance.
func StorageControllerInstanceSchema() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"controller_rates": schema.SingleNestedAttribute{
			MarkdownDescription: "This type describes the various controller rates used for processes such as volume rebuild or consistency checks.",
			Description:         "This type describes the various controller rates used for processes such as volume rebuild or consistency checks.",
			Optional:            true,
			Computed:            true,
			Attributes:          ControllerRatesResourceSchema(),
		},
		"oem": schema.SingleNestedAttribute{
			MarkdownDescription: "The OEM extension to the StorageController resource.",
			Description:         "The OEM extension to the StorageController resource.",
			Optional:            true,
			Computed:            true,
			Attributes:          StorageControllerOEMResourceSchema(),
		},
	}
}

// ControllerRatesResourceSchema is a function that returns the schema for Controller Rates.
func ControllerRatesResourceSchema() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"consistency_check_rate_percent": schema.Int64Attribute{
			MarkdownDescription: "This property describes the controller rate for consistency check",
			Description:         "This property describes the controller rate for consistency check",
			Optional:            true,
			Computed:            true,
		},
		"rebuild_rate_percent": schema.Int64Attribute{
			MarkdownDescription: "This property describes the controller rate for volume rebuild",
			Description:         "This property describes the controller rate for volume rebuild",
			Optional:            true,
			Computed:            true,
		},
	}
}

// StorageControllerOEMResourceSchema is a function that returns the schema for Storage Controller OEM.
func StorageControllerOEMResourceSchema() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"dell": schema.SingleNestedAttribute{
			MarkdownDescription: "Dell",
			Description:         "Dell",
			Optional:            true,
			Computed:            true,
			Attributes:          StorageControllerOEMDellResourceSchema(),
		},
	}
}

// StorageControllerOEMDellResourceSchema is a function that returns the schema for Storage Controller OEM Dell.
func StorageControllerOEMDellResourceSchema() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"dell_storage_controller": schema.SingleNestedAttribute{
			MarkdownDescription: "Dell Storage Controller",
			Description:         "Dell Storage Controller",
			Optional:            true,
			Computed:            true,
			Attributes:          DellStorageControllerResourceSchema(),
		},
	}
}

// nolint: revive
// DellStorageControllerResourceSchema is a function that returns the schema for Dell Storage Controller
func DellStorageControllerResourceSchema() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"controller_mode": schema.StringAttribute{
			MarkdownDescription: "Controller Mode. Accepted values: `RAID`, `HBA`. " +
				"When updating `controller_mode`, the `apply_time` should be `OnReset` or `InMaintenanceWindowOnReset` and " +
				"no other attributes from `storage_controller` or `security` should be updated. " +
				"Specifically, when updating `controller_mode` to `HBA`, the `enhanced_auto_import_foreign_configuration_mode` attribute needs to be commented.",
			Description: "Controller Mode. Accepted values: `RAID`, `HBA`. " +
				"When updating `controller_mode`, the `apply_time` should be `OnReset` or `InMaintenanceWindowOnReset` and " +
				"no other attributes from `storage_controller` or `security` should be updated. " +
				"Specifically, when updating `controller_mode` to `HBA`, the `enhanced_auto_import_foreign_configuration_mode` attribute needs to be commented.",
			Optional: true,
			Computed: true,
			Validators: []validator.String{stringvalidator.OneOf(
				"RAID",
				"HBA",
			)},
		},
		"check_consistency_mode": schema.StringAttribute{
			MarkdownDescription: "Check Consistency Mode. Accepted values: `Normal`, `StopOnError`.",
			Description:         "Check Consistency Mode. Accepted values: `Normal`, `StopOnError`.",
			Optional:            true,
			Computed:            true,
			Validators: []validator.String{stringvalidator.OneOf(
				"Normal",
				"StopOnError",
			)},
		},
		"copyback_mode": schema.StringAttribute{
			MarkdownDescription: "Copyback Mode. Accepted values: `On`, `OnWithSMART`, `Off`.",
			Description:         "Copyback Mode. Accepted values: `On`, `OnWithSMART`, `Off`.",
			Optional:            true,
			Computed:            true,
			Validators: []validator.String{stringvalidator.OneOf(
				"On",
				"OnWithSMART",
				"Off",
			)},
		},
		"load_balance_mode": schema.StringAttribute{
			MarkdownDescription: "Load Balance Mode. Accepted values: `Automatic`, `Disabled`.",
			Description:         "Load Balance Mode. Accepted values: `Automatic`, `Disabled`.",
			Optional:            true,
			Computed:            true,
			Validators: []validator.String{stringvalidator.OneOf(
				"Automatic",
				"Disabled",
			)},
		},
		"enhanced_auto_import_foreign_configuration_mode": schema.StringAttribute{
			MarkdownDescription: "Enhanced Auto Import Foreign Configuration Mode. Accepted values: `Disabled`, `Enabled`. " +
				"When updating `controller_mode` to `HBA`, this attribute needs to be commented.",
			Description: "Enhanced Auto Import Foreign Configuration Mode. Accepted values: `Disabled`, `Enabled`. " +
				"When updating `controller_mode` to `HBA`, this attribute needs to be commented.",
			Optional: true,
			Computed: true,
			Validators: []validator.String{stringvalidator.OneOf(
				"Disabled",
				"Enabled",
			)},
		},
		"patrol_read_unconfigured_area_mode": schema.StringAttribute{
			MarkdownDescription: "Patrol Read Unconfigured Area Mode. Accepted values: `Disabled`, `Enabled`.",
			Description:         "Patrol Read Unconfigured Area Mode. Accepted values: `Disabled`, `Enabled`.",
			Optional:            true,
			Computed:            true,
			Validators: []validator.String{stringvalidator.OneOf(
				"Disabled",
				"Enabled",
			)},
		},
		"patrol_read_mode": schema.StringAttribute{
			MarkdownDescription: "Patrol Read Mode. Accepted values: `Disabled`, `Automatic`, `Manual`.",
			Description:         "Patrol Read Mode. Accepted values: `Disabled`, `Automatic`, `Manual`.",
			Optional:            true,
			Computed:            true,
			Validators: []validator.String{stringvalidator.OneOf(
				"Disabled",
				"Automatic",
				"Manual",
			)},
		},
		"background_initialization_rate_percent": schema.Int64Attribute{
			MarkdownDescription: "Background Initialization Rate Percent",
			Description:         "Background Initialization Rate Percent",
			Optional:            true,
			Computed:            true,
		},
		"reconstruct_rate_percent": schema.Int64Attribute{
			MarkdownDescription: "Reconstruct Rate Percent",
			Description:         "Reconstruct Rate Percent",
			Optional:            true,
			Computed:            true,
		},
	}
}
