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
func parseStorageControllerExtendedIntoState(ctx context.Context, storageControllerExtended *dell.StorageControllerExtended, state *models.StorageControllerResource) diag.Diagnostics {
	var diags diag.Diagnostics

	objectAsOptions := basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true}

	var oldStorageControllerAttributes models.StorageControllerAttributes
	diags = state.StorageController.As(ctx, &oldStorageControllerAttributes, objectAsOptions)
	if diags.HasError() {
		return diags
	}

	controllerRatesObj, diags := getControllerRatesObjectValue(storageControllerExtended)
	if diags.HasError() {
		return diags
	}
	oemAttributesObj, diags := getOEMAttributesObjectValue(storageControllerExtended)
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

func getControllerRatesObjectValue(storageControllerExtended *dell.StorageControllerExtended) (basetypes.ObjectValue, diag.Diagnostics) {
	controllerRatesItemMap := map[string]attr.Value{
		"consistency_check_rate_percent": types.Int64Value(int64(storageControllerExtended.ControllerRates.ConsistencyCheckRatePercent)),
		"rebuild_rate_percent":           types.Int64Value(int64(storageControllerExtended.ControllerRates.RebuildRatePercent)),
	}
	return types.ObjectValue(getControllerRatesModelType(), controllerRatesItemMap)
}

func getOEMAttributesObjectValue(storageControllerExtended *dell.StorageControllerExtended) (basetypes.ObjectValue, diag.Diagnostics) {
	emptyObj := types.ObjectNull(getOEMAttributesModelType())

	dellAttributesObj, diags := getDellAttributesObjectValue(storageControllerExtended)
	if diags.HasError() {
		return emptyObj, diags
	}

	oemAttributesItemMap := map[string]attr.Value{
		"dell": dellAttributesObj,
	}

	return types.ObjectValue(getOEMAttributesModelType(), oemAttributesItemMap)
}

func getDellAttributesObjectValue(storageControllerExtended *dell.StorageControllerExtended) (basetypes.ObjectValue, diag.Diagnostics) {
	emptyObj := types.ObjectNull(getDellAttributesModelType())

	dellStorageControllerAttributesObj, diags := getDellStorageControllerAttributesObjectValue(storageControllerExtended)
	if diags.HasError() {
		return emptyObj, diags
	}

	dellAttributesItemMap := map[string]attr.Value{
		"dell_storage_controller": dellStorageControllerAttributesObj,
	}

	return types.ObjectValue(getDellAttributesModelType(), dellAttributesItemMap)
}

// nolint: revive
func getDellStorageControllerAttributesObjectValue(storageControllerExtended *dell.StorageControllerExtended) (basetypes.ObjectValue, diag.Diagnostics) {
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
