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
	"context"
	"terraform-provider-redfish/gofish/dell"
	"terraform-provider-redfish/redfish/models"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

// nolint: revive
func parseSecurityAttributesIntoState(ctx context.Context, storageControllerExtended *dell.StorageControllerExtended, state *models.StorageControllerResource) diag.Diagnostics {
	var diags diag.Diagnostics

	objectAsOptions := basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true}

	var stateSecurityAttributes models.SecurityAttributes
	diags = state.Security.As(ctx, &stateSecurityAttributes, objectAsOptions)
	if diags.HasError() {
		return diags
	}

	stateSecurityAttributes.KeyID = types.StringValue(storageControllerExtended.Oem.Dell.DellStorageController.KeyID)

	if storageControllerExtended.Oem.Dell.DellStorageController.EncryptionMode == "None" {
		stateSecurityAttributes.Mode = types.StringValue("")
	} else if storageControllerExtended.Oem.Dell.DellStorageController.EncryptionMode == "LocalKeyManagement" {
		stateSecurityAttributes.Mode = types.StringValue("LKM")
	} else if storageControllerExtended.Oem.Dell.DellStorageController.EncryptionMode == "SecureEnterpriseKeyManager" {
		stateSecurityAttributes.Mode = types.StringValue("SEKM")
	}

	if stateSecurityAttributes.Action.IsNull() || stateSecurityAttributes.Action.IsUnknown() {
		stateSecurityAttributes.Action = types.StringValue("")
	}
	if stateSecurityAttributes.Key.IsNull() || stateSecurityAttributes.Key.IsUnknown() {
		stateSecurityAttributes.Key = types.StringValue("")
	}
	if stateSecurityAttributes.OldKey.IsNull() || stateSecurityAttributes.OldKey.IsUnknown() {
		stateSecurityAttributes.OldKey = types.StringValue("")
	}

	newSecurityAttributesItemMap := map[string]attr.Value{
		"action":  stateSecurityAttributes.Action,
		"key_id":  stateSecurityAttributes.KeyID,
		"key":     stateSecurityAttributes.Key,
		"old_key": stateSecurityAttributes.OldKey,
		"mode":    stateSecurityAttributes.Mode,
	}

	state.Security, diags = types.ObjectValue(getSecurityAttributesModelType(), newSecurityAttributesItemMap)
	return diags
}

// nolint: revive
func parseStorageControllerExtendedIntoState(ctx context.Context, storageControllerExtended *dell.StorageControllerExtended, state *models.StorageControllerResource, isPlan bool) diag.Diagnostics {
	var diags diag.Diagnostics

	objectAsOptions := basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true}

	var oldStorageControllerAttributes models.StorageControllerAttributes
	diags = state.StorageController.As(ctx, &oldStorageControllerAttributes, objectAsOptions)
	if diags.HasError() {
		return diags
	}

	controllerRatesObj, diags := getControllerRatesObjectValue(storageControllerExtended, oldStorageControllerAttributes, ctx, isPlan)
	if diags.HasError() {
		return diags
	}
	oemAttributesObj, diags := getOEMAttributesObjectValue(storageControllerExtended, oldStorageControllerAttributes, ctx, isPlan)
	if diags.HasError() {
		return diags
	}

	newStorageControllerAttributesItemMap := map[string]attr.Value{
		"controller_rates": controllerRatesObj,
		"oem":              oemAttributesObj,
	}

	state.StorageController, diags = types.ObjectValue(getStorageControllerAttributesModelType(), newStorageControllerAttributesItemMap)
	return diags
}

// nolint: revive
func getControllerRatesObjectValue(storageControllerExtended *dell.StorageControllerExtended, oldState models.StorageControllerAttributes, ctx context.Context, isPlan bool) (basetypes.ObjectValue, diag.Diagnostics) {
	controllerRatesItemMap := map[string]attr.Value{
		"consistency_check_rate_percent": types.Int64Value(int64(storageControllerExtended.ControllerRates.ConsistencyCheckRatePercent)),
		"rebuild_rate_percent":           types.Int64Value(int64(storageControllerExtended.ControllerRates.RebuildRatePercent)),
	}

	if isPlan {
		if !oldState.ControllerRates.IsNull() && !oldState.ControllerRates.IsUnknown() {
			objectAsOptions := basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true}
			var oldAttributes models.ControllerRates
			diags := oldState.ControllerRates.As(ctx, &oldAttributes, objectAsOptions)
			if diags.HasError() {
				return types.ObjectValue(getControllerRatesModelType(), controllerRatesItemMap)
			}

			if !oldAttributes.ConsistencyCheckRatePercent.IsNull() && !oldAttributes.ConsistencyCheckRatePercent.IsUnknown() {
				controllerRatesItemMap["consistency_check_rate_percent"] = oldAttributes.ConsistencyCheckRatePercent
			}
			if !oldAttributes.RebuildRatePercent.IsNull() && !oldAttributes.RebuildRatePercent.IsUnknown() {
				controllerRatesItemMap["rebuild_rate_percent"] = oldAttributes.RebuildRatePercent
			}
		}
	}

	return types.ObjectValue(getControllerRatesModelType(), controllerRatesItemMap)
}

// nolint: revive
func getOEMAttributesObjectValue(storageControllerExtended *dell.StorageControllerExtended, oldState models.StorageControllerAttributes, ctx context.Context, isPlan bool) (basetypes.ObjectValue, diag.Diagnostics) {
	emptyObj := types.ObjectNull(getOEMAttributesModelType())

	dellAttributesObj, diags := getDellAttributesObjectValue(storageControllerExtended, oldState, ctx, isPlan)
	if diags.HasError() {
		return emptyObj, diags
	}

	oemAttributesItemMap := map[string]attr.Value{
		"dell": dellAttributesObj,
	}

	return types.ObjectValue(getOEMAttributesModelType(), oemAttributesItemMap)
}

// nolint: revive
func getDellAttributesObjectValue(storageControllerExtended *dell.StorageControllerExtended, oldState models.StorageControllerAttributes, ctx context.Context, isPlan bool) (basetypes.ObjectValue, diag.Diagnostics) {
	emptyObj := types.ObjectNull(getDellAttributesModelType())

	dellStorageControllerAttributesObj, diags := getDellStorageControllerAttributesObjectValue(storageControllerExtended, oldState, ctx, isPlan)
	if diags.HasError() {
		return emptyObj, diags
	}

	dellAttributesItemMap := map[string]attr.Value{
		"dell_storage_controller": dellStorageControllerAttributesObj,
	}

	return types.ObjectValue(getDellAttributesModelType(), dellAttributesItemMap)
}

// nolint: revive
func getDellStorageControllerAttributesObjectValue(storageControllerExtended *dell.StorageControllerExtended, oldState models.StorageControllerAttributes, ctx context.Context, isPlan bool) (basetypes.ObjectValue, diag.Diagnostics) {
	dellStorageControllerAttributesItemMap := map[string]attr.Value{
		"controller_mode":        types.StringValue(storageControllerExtended.Oem.Dell.DellStorageController.ControllerMode),
		"check_consistency_mode": types.StringValue(storageControllerExtended.Oem.Dell.DellStorageController.CheckConsistencyMode),
		"copyback_mode":          types.StringValue(storageControllerExtended.Oem.Dell.DellStorageController.CopybackMode),
		"load_balance_mode":      types.StringValue(storageControllerExtended.Oem.Dell.DellStorageController.LoadBalanceMode),
		"enhanced_auto_import_foreign_configuration_mode": types.StringValue(storageControllerExtended.Oem.Dell.DellStorageController.EnhancedAutoImportForeignConfigurationMode),
		"patrol_read_unconfigured_area_mode":              types.StringValue(storageControllerExtended.Oem.Dell.DellStorageController.PatrolReadUnconfiguredAreaMode),
		"patrol_read_mode":                                types.StringValue(storageControllerExtended.Oem.Dell.DellStorageController.PatrolReadMode),
		"background_initialization_rate_percent":          types.Int64Value(storageControllerExtended.Oem.Dell.DellStorageController.BackgroundInitializationRatePercent),
		"reconstruct_rate_percent":                        types.Int64Value(storageControllerExtended.Oem.Dell.DellStorageController.ReconstructRatePercent),
	}

	if isPlan {
		if !oldState.Oem.IsNull() && !oldState.Oem.IsUnknown() {
			objectAsOptions := basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true}

			var oldOemAttributes models.OEMAttributes
			diags := oldState.Oem.As(ctx, &oldOemAttributes, objectAsOptions)
			if diags.HasError() {
				return types.ObjectValue(getDellStorageControllerAttributesModelType(), dellStorageControllerAttributesItemMap)
			}

			if !oldOemAttributes.Dell.IsNull() && !oldOemAttributes.Dell.IsUnknown() {
				var oldDellAttributes models.DellAttributes
				diags := oldOemAttributes.Dell.As(ctx, &oldDellAttributes, objectAsOptions)
				if diags.HasError() {
					return types.ObjectValue(getDellStorageControllerAttributesModelType(), dellStorageControllerAttributesItemMap)
				}

				if !oldDellAttributes.DellStorageController.IsNull() && !oldDellAttributes.DellStorageController.IsUnknown() {
					var oldDellStorageControllerAttributes models.DellStorageControllerAttributes
					diags := oldDellAttributes.DellStorageController.As(ctx, &oldDellStorageControllerAttributes, objectAsOptions)
					if diags.HasError() {
						return types.ObjectValue(getDellStorageControllerAttributesModelType(), dellStorageControllerAttributesItemMap)
					}

					if !oldDellStorageControllerAttributes.ControllerMode.IsNull() && !oldDellStorageControllerAttributes.ControllerMode.IsUnknown() {
						dellStorageControllerAttributesItemMap["controller_mode"] = oldDellStorageControllerAttributes.ControllerMode
					}
					if !oldDellStorageControllerAttributes.CheckConsistencyMode.IsNull() && !oldDellStorageControllerAttributes.CheckConsistencyMode.IsUnknown() {
						dellStorageControllerAttributesItemMap["check_consistency_mode"] = oldDellStorageControllerAttributes.CheckConsistencyMode
					}
					if !oldDellStorageControllerAttributes.CopybackMode.IsNull() && !oldDellStorageControllerAttributes.CopybackMode.IsUnknown() {
						dellStorageControllerAttributesItemMap["copyback_mode"] = oldDellStorageControllerAttributes.CopybackMode
					}
					if !oldDellStorageControllerAttributes.LoadBalanceMode.IsNull() && !oldDellStorageControllerAttributes.LoadBalanceMode.IsUnknown() {
						dellStorageControllerAttributesItemMap["load_balance_mode"] = oldDellStorageControllerAttributes.LoadBalanceMode
					}
					if !oldDellStorageControllerAttributes.EnhancedAutoImportForeignConfigurationMode.IsNull() && !oldDellStorageControllerAttributes.EnhancedAutoImportForeignConfigurationMode.IsUnknown() {
						dellStorageControllerAttributesItemMap["enhanced_auto_import_foreign_configuration_mode"] = oldDellStorageControllerAttributes.EnhancedAutoImportForeignConfigurationMode
					}
					if !oldDellStorageControllerAttributes.PatrolReadUnconfiguredAreaMode.IsNull() && !oldDellStorageControllerAttributes.PatrolReadUnconfiguredAreaMode.IsUnknown() {
						dellStorageControllerAttributesItemMap["patrol_read_unconfigured_area_mode"] = oldDellStorageControllerAttributes.PatrolReadUnconfiguredAreaMode
					}
					if !oldDellStorageControllerAttributes.PatrolReadMode.IsNull() && !oldDellStorageControllerAttributes.PatrolReadMode.IsUnknown() {
						dellStorageControllerAttributesItemMap["patrol_read_mode"] = oldDellStorageControllerAttributes.PatrolReadMode
					}
					if !oldDellStorageControllerAttributes.BackgroundInitializationRatePercent.IsNull() && !oldDellStorageControllerAttributes.BackgroundInitializationRatePercent.IsUnknown() {
						dellStorageControllerAttributesItemMap["background_initialization_rate_percent"] = oldDellStorageControllerAttributes.BackgroundInitializationRatePercent
					}
					if !oldDellStorageControllerAttributes.ReconstructRatePercent.IsNull() && !oldDellStorageControllerAttributes.ReconstructRatePercent.IsUnknown() {
						dellStorageControllerAttributesItemMap["reconstruct_rate_percent"] = oldDellStorageControllerAttributes.ReconstructRatePercent
					}
				}
			}
		}
	}

	return types.ObjectValue(getDellStorageControllerAttributesModelType(), dellStorageControllerAttributesItemMap)
}

func getSecurityAttributesModelType() map[string]attr.Type {
	return map[string]attr.Type{
		"action":  types.StringType,
		"key_id":  types.StringType,
		"key":     types.StringType,
		"old_key": types.StringType,
		"mode":    types.StringType,
	}
}

func getStorageControllerAttributesModelType() map[string]attr.Type {
	return map[string]attr.Type{
		"controller_rates": types.ObjectType{
			AttrTypes: getControllerRatesModelType(),
		},
		"oem": types.ObjectType{
			AttrTypes: getOEMAttributesModelType(),
		},
	}
}

func getControllerRatesModelType() map[string]attr.Type {
	return map[string]attr.Type{
		"consistency_check_rate_percent": types.Int64Type,
		"rebuild_rate_percent":           types.Int64Type,
	}
}

func getOEMAttributesModelType() map[string]attr.Type {
	return map[string]attr.Type{
		"dell": types.ObjectType{
			AttrTypes: getDellAttributesModelType(),
		},
	}
}

func getDellAttributesModelType() map[string]attr.Type {
	return map[string]attr.Type{
		"dell_storage_controller": types.ObjectType{
			AttrTypes: getDellStorageControllerAttributesModelType(),
		},
	}
}

func getDellStorageControllerAttributesModelType() map[string]attr.Type {
	return map[string]attr.Type{
		"controller_mode":        types.StringType,
		"check_consistency_mode": types.StringType,
		"copyback_mode":          types.StringType,
		"load_balance_mode":      types.StringType,
		"enhanced_auto_import_foreign_configuration_mode": types.StringType,
		"patrol_read_unconfigured_area_mode":              types.StringType,
		"patrol_read_mode":                                types.StringType,
		"background_initialization_rate_percent":          types.Int64Type,
		"reconstruct_rate_percent":                        types.Int64Type,
	}
}
