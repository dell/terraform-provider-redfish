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
	"fmt"
	"strings"
	"terraform-provider-redfish/redfish/models"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/stmcginnis/gofish"
	"github.com/stmcginnis/gofish/redfish"
)

// nolint: gocyclo, gocognit, revive
func getStorageControllerInstance(service *gofish.Service, systemID, storageID, controllerID string) (*redfish.ComputerSystem, *redfish.StorageController, error) {
	// get system by id, if system id is empty, use the first one.
	system, err := getSystemResource(service, systemID)
	if err != nil {
		return nil, nil, err
	}

	// get storage by id
	storage, err := getStorageUsingID(system, storageID)
	if err != nil {
		return system, nil, err
	}

	// get controller by id
	storageController, err := getStorageControllerUsingID(storage, controllerID)
	if err != nil {
		return system, nil, err
	}

	return system, storageController, nil
}

func getStorageUsingID(system *redfish.ComputerSystem, storageID string) (*redfish.Storage, error) {
	storageList, err := system.Storage()
	if err != nil {
		return nil, err
	}

	for _, storage := range storageList {
		if storage.ID == storageID {
			return storage, nil
		}
	}

	return nil, fmt.Errorf("couldn't find the storage: %s", storageID)
}

func getStorageControllerUsingID(storage *redfish.Storage, controllerID string) (*redfish.StorageController, error) {
	storageControllerList, err := storage.Controllers()
	if err != nil {
		return nil, err
	}

	for _, storageController := range storageControllerList {
		if storageController.ID == controllerID {
			return storageController, nil
		}
	}

	return nil, fmt.Errorf("couldn't find the storage controller: %s", controllerID)
}

// nolint: gocyclo, gocognit, revive
func securityAttributesChanged(ctx context.Context, plan, state *models.StorageControllerResource) bool {
	if plan.Security.IsNull() || plan.Security.IsUnknown() {
		return false
	}

	objectAsOptions := basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true}

	var planAttributes models.SecurityAttributes
	var stateAttributes models.SecurityAttributes

	diags := plan.Security.As(ctx, &planAttributes, objectAsOptions)
	if diags.HasError() {
		tflog.Debug(ctx, "resource_redfish_storage_controller: securityAttributesChanged: plan.Security.As: error")
		return false
	}

	diags = state.Security.As(ctx, &stateAttributes, objectAsOptions)
	if diags.HasError() {
		tflog.Debug(ctx, "resource_redfish_storage_controller: securityAttributesChanged: state.Security.As: error")
		return false
	}

	if !planAttributes.Action.IsNull() && !planAttributes.Action.IsUnknown() && planAttributes.Action.ValueString() != "" &&
		planAttributes.Action.ValueString() != stateAttributes.Action.ValueString() {
		return true
	}
	if !planAttributes.KeyID.IsNull() && !planAttributes.KeyID.IsUnknown() && planAttributes.KeyID.ValueString() != "" &&
		planAttributes.KeyID.ValueString() != stateAttributes.KeyID.ValueString() {
		return true
	}
	if !planAttributes.Key.IsNull() && !planAttributes.Key.IsUnknown() && planAttributes.Key.ValueString() != "" &&
		planAttributes.Key.ValueString() != stateAttributes.Key.ValueString() {
		return true
	}
	if !planAttributes.OldKey.IsNull() && !planAttributes.OldKey.IsUnknown() && planAttributes.OldKey.ValueString() != "" &&
		planAttributes.OldKey.ValueString() != stateAttributes.OldKey.ValueString() {
		return true
	}
	if !planAttributes.Mode.IsNull() && !planAttributes.Mode.IsUnknown() && planAttributes.Mode.ValueString() != "" &&
		planAttributes.Mode.ValueString() != stateAttributes.Mode.ValueString() {
		return true
	}

	return false
}

// nolint: gocyclo, gocognit, revive
func storageControllerAttributesChanged(ctx context.Context, plan, state *models.StorageControllerResource, checkForOnlyControllerModeChange bool) bool {
	if plan.StorageController.IsNull() || plan.StorageController.IsUnknown() {
		return false
	}

	objectAsOptions := basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true}

	var planAttributes models.StorageControllerAttributes
	var stateAttributes models.StorageControllerAttributes

	diags := plan.StorageController.As(ctx, &planAttributes, objectAsOptions)
	if diags.HasError() {
		tflog.Debug(ctx, "resource_redfish_storage_controller: storageControllerAttributesChanged: plan.StorageController.As: error")
		return false
	}

	diags = state.StorageController.As(ctx, &stateAttributes, objectAsOptions)
	if diags.HasError() {
		tflog.Debug(ctx, "resource_redfish_storage_controller: storageControllerAttributesChanged: state.StorageController.As: error")
		return false
	}

	if checkForOnlyControllerModeChange {
		return oemAttributesChanged(ctx, &planAttributes, &stateAttributes, checkForOnlyControllerModeChange)
	}

	return controllerRatesChanged(ctx, &planAttributes, &stateAttributes) ||
		oemAttributesChanged(ctx, &planAttributes, &stateAttributes, checkForOnlyControllerModeChange)
}

func controllerRatesChanged(ctx context.Context, plan, state *models.StorageControllerAttributes) bool {
	if plan.ControllerRates.IsNull() || plan.ControllerRates.IsUnknown() {
		return false
	}

	objectAsOptions := basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true}

	var planAttributes models.ControllerRates
	var stateAttributes models.ControllerRates

	diags := plan.ControllerRates.As(ctx, &planAttributes, objectAsOptions)
	if diags.HasError() {
		tflog.Debug(ctx, "resource_redfish_storage_controller: controllerRatesChanged: plan.ControllerRates.As: error")
		return false
	}

	diags = state.ControllerRates.As(ctx, &stateAttributes, objectAsOptions)
	if diags.HasError() {
		tflog.Debug(ctx, "resource_redfish_storage_controller: controllerRatesChanged: state.ControllerRates.As: error")
		return false
	}

	if !planAttributes.ConsistencyCheckRatePercent.IsNull() && !planAttributes.ConsistencyCheckRatePercent.IsUnknown() &&
		planAttributes.ConsistencyCheckRatePercent.ValueInt64() != stateAttributes.ConsistencyCheckRatePercent.ValueInt64() {
		return true
	}

	if !planAttributes.RebuildRatePercent.IsNull() && !planAttributes.RebuildRatePercent.IsUnknown() &&
		planAttributes.RebuildRatePercent.ValueInt64() != stateAttributes.RebuildRatePercent.ValueInt64() {
		return true
	}

	return false
}

func oemAttributesChanged(ctx context.Context, plan, state *models.StorageControllerAttributes, checkForOnlyControllerModeChange bool) bool {
	if plan.Oem.IsNull() || plan.Oem.IsUnknown() {
		return false
	}

	objectAsOptions := basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true}

	var planAttributes models.OEMAttributes
	var stateAttributes models.OEMAttributes

	diags := plan.Oem.As(ctx, &planAttributes, objectAsOptions)
	if diags.HasError() {
		tflog.Debug(ctx, "resource_redfish_storage_controller: oemAttributesChanged: plan.Oem.As: error")
		return false
	}

	diags = state.Oem.As(ctx, &stateAttributes, objectAsOptions)
	if diags.HasError() {
		tflog.Debug(ctx, "resource_redfish_storage_controller: oemAttributesChanged: state.Oem.As: error")
		return false
	}

	return dellAttributesChanged(ctx, &planAttributes, &stateAttributes, checkForOnlyControllerModeChange)
}

func dellAttributesChanged(ctx context.Context, plan, state *models.OEMAttributes, checkForOnlyControllerModeChange bool) bool {
	if plan.Dell.IsNull() || plan.Dell.IsUnknown() {
		return false
	}

	objectAsOptions := basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true}

	var planAttributes models.DellAttributes
	var stateAttributes models.DellAttributes

	diags := plan.Dell.As(ctx, &planAttributes, objectAsOptions)
	if diags.HasError() {
		tflog.Debug(ctx, "resource_redfish_storage_controller: dellAttributesChanged: plan.Dell.As: error")
		return false
	}

	diags = state.Dell.As(ctx, &stateAttributes, objectAsOptions)
	if diags.HasError() {
		tflog.Debug(ctx, "resource_redfish_storage_controller: dellAttributesChanged: state.Dell.As: error")
		return false
	}

	return dellStorageControllerAttributesChanged(ctx, &planAttributes, &stateAttributes, checkForOnlyControllerModeChange)
}

// nolint: gocyclo, gocognit, revive
func dellStorageControllerAttributesChanged(ctx context.Context, plan, state *models.DellAttributes, checkForOnlyControllerModeChange bool) bool {
	if plan.DellStorageController.IsNull() || plan.DellStorageController.IsUnknown() {
		return false
	}

	objectAsOptions := basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true}

	var planAttributes models.DellStorageControllerAttributes
	var stateAttributes models.DellStorageControllerAttributes

	diags := plan.DellStorageController.As(ctx, &planAttributes, objectAsOptions)
	if diags.HasError() {
		tflog.Debug(ctx, "resource_redfish_storage_controller: dellStorageControllerAttributesChanged: plan.DellStorageController.As: error")
		return false
	}

	diags = state.DellStorageController.As(ctx, &stateAttributes, objectAsOptions)
	if diags.HasError() {
		tflog.Debug(ctx, "resource_redfish_storage_controller: dellStorageControllerAttributesChanged: state.DellStorageController.As: error")
		return false
	}

	if checkForOnlyControllerModeChange {
		if !planAttributes.ControllerMode.IsNull() && !planAttributes.ControllerMode.IsUnknown() &&
			planAttributes.ControllerMode.ValueString() != "" &&
			planAttributes.ControllerMode.ValueString() != stateAttributes.ControllerMode.ValueString() {
			return true
		}

		return false
	}

	if !planAttributes.CheckConsistencyMode.IsNull() && !planAttributes.CheckConsistencyMode.IsUnknown() &&
		planAttributes.CheckConsistencyMode.ValueString() != "" &&
		planAttributes.CheckConsistencyMode.ValueString() != stateAttributes.CheckConsistencyMode.ValueString() {
		return true
	}
	if !planAttributes.CopybackMode.IsNull() && !planAttributes.CopybackMode.IsUnknown() &&
		planAttributes.CopybackMode.ValueString() != "" &&
		planAttributes.CopybackMode.ValueString() != stateAttributes.CopybackMode.ValueString() {
		return true
	}
	if !planAttributes.LoadBalanceMode.IsNull() && !planAttributes.LoadBalanceMode.IsUnknown() &&
		planAttributes.LoadBalanceMode.ValueString() != "" &&
		planAttributes.LoadBalanceMode.ValueString() != stateAttributes.LoadBalanceMode.ValueString() {
		return true
	}
	if !planAttributes.EnhancedAutoImportForeignConfigurationMode.IsNull() && !planAttributes.EnhancedAutoImportForeignConfigurationMode.IsUnknown() &&
		planAttributes.EnhancedAutoImportForeignConfigurationMode.ValueString() != "" &&
		planAttributes.EnhancedAutoImportForeignConfigurationMode.ValueString() != stateAttributes.EnhancedAutoImportForeignConfigurationMode.ValueString() {
		return true
	}
	if !planAttributes.PatrolReadUnconfiguredAreaMode.IsNull() && !planAttributes.PatrolReadUnconfiguredAreaMode.IsUnknown() &&
		planAttributes.PatrolReadUnconfiguredAreaMode.ValueString() != "" &&
		planAttributes.PatrolReadUnconfiguredAreaMode.ValueString() != stateAttributes.PatrolReadUnconfiguredAreaMode.ValueString() {
		return true
	}
	if !planAttributes.PatrolReadMode.IsNull() && !planAttributes.PatrolReadMode.IsUnknown() &&
		planAttributes.PatrolReadMode.ValueString() != "" &&
		planAttributes.PatrolReadMode.ValueString() != stateAttributes.PatrolReadMode.ValueString() {
		return true
	}
	if !planAttributes.BackgroundInitializationRatePercent.IsNull() && !planAttributes.BackgroundInitializationRatePercent.IsUnknown() &&
		planAttributes.BackgroundInitializationRatePercent.ValueInt64() != stateAttributes.BackgroundInitializationRatePercent.ValueInt64() {
		return true
	}
	if !planAttributes.ReconstructRatePercent.IsNull() && !planAttributes.ReconstructRatePercent.IsUnknown() &&
		planAttributes.ReconstructRatePercent.ValueInt64() != stateAttributes.ReconstructRatePercent.ValueInt64() {
		return true
	}

	return false
}

func getStorageControllerAttributeInfo(ctx context.Context, plan *models.StorageControllerResource, attributeName string) (bool, string) {
	objectAsOptions := basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true}

	var planStorageControllerAttributes models.StorageControllerAttributes
	diags := plan.StorageController.As(ctx, &planStorageControllerAttributes, objectAsOptions)
	if diags.HasError() {
		tflog.Debug(ctx, "resource_redfish_storage_controller: getStorageControllerAttributeInfo: plan.StorageController.As: error")
		return true, ""
	}

	var planOEMAttributes models.OEMAttributes
	diags = planStorageControllerAttributes.Oem.As(ctx, &planOEMAttributes, objectAsOptions)
	if diags.HasError() {
		tflog.Debug(ctx, "resource_redfish_storage_controller: getStorageControllerAttributeInfo: plan.Oem.As: error")
		return true, ""
	}

	var planDellAttributes models.DellAttributes
	diags = planOEMAttributes.Dell.As(ctx, &planDellAttributes, objectAsOptions)
	if diags.HasError() {
		tflog.Debug(ctx, "resource_redfish_storage_controller: getStorageControllerAttributeInfo: plan.Dell.As: error")
		return true, ""
	}

	var planDellStorageControllerAttributes models.DellStorageControllerAttributes
	diags = planDellAttributes.DellStorageController.As(ctx, &planDellStorageControllerAttributes, objectAsOptions)
	if diags.HasError() {
		tflog.Debug(ctx, "resource_redfish_storage_controller: getStorageControllerAttributeInfo: plan.DellStorageController.As: error")
		return true, ""
	}

	if attributeName == "ControllerMode" {
		if !planDellStorageControllerAttributes.ControllerMode.IsNull() &&
			!planDellStorageControllerAttributes.ControllerMode.IsUnknown() &&
			planDellStorageControllerAttributes.ControllerMode.ValueString() != "" {
			return false, planDellStorageControllerAttributes.ControllerMode.ValueString()
		}

		return true, ""
	} else if attributeName == "EnhancedAutoImportForeignConfigurationMode" {
		if !planDellStorageControllerAttributes.EnhancedAutoImportForeignConfigurationMode.IsNull() &&
			!planDellStorageControllerAttributes.EnhancedAutoImportForeignConfigurationMode.IsUnknown() &&
			planDellStorageControllerAttributes.EnhancedAutoImportForeignConfigurationMode.ValueString() != "" {
			return false, planDellStorageControllerAttributes.EnhancedAutoImportForeignConfigurationMode.ValueString()
		}

		return true, ""
	}

	return true, ""
}

// nolint: gocyclo, gocognit, revive
func updateStorageControllerAttributes(ctx context.Context, service *gofish.Service, plan, state *models.StorageControllerResource) (jobURL string, diags diag.Diagnostics) {
	objectAsOptions := basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true}

	var planAttributes models.StorageControllerAttributes
	diags = plan.StorageController.As(ctx, &planAttributes, objectAsOptions)
	if diags.HasError() {
		return
	}

	var stateAttributes models.StorageControllerAttributes
	diags = state.StorageController.As(ctx, &stateAttributes, objectAsOptions)
	if diags.HasError() {
		return
	}

	patchBody := make(map[string]interface{})
	patchBody[patchBodySettingsApplyTime] = map[string]interface{}{
		patchBodyApplyTime: plan.ApplyTime.ValueString(),
	}
	if strings.Contains(plan.ApplyTime.ValueString(), "Maintenance") {
		patchBody[patchBodySettingsApplyTime] = map[string]interface{}{
			patchBodyApplyTime:                   plan.ApplyTime.ValueString(),
			"MaintenanceWindowStartTime":         plan.MaintenanceWindow.StartTime.ValueString(),
			"MaintenanceWindowDurationInSeconds": plan.MaintenanceWindow.Duration.ValueInt64(),
		}
	}

	if !planAttributes.ControllerRates.IsNull() && !planAttributes.ControllerRates.IsUnknown() {
		controllerRatesPatchBody, diags := getControllerRatesPatchBody(ctx, &planAttributes, &stateAttributes)
		if diags.HasError() {
			return "", diags
		}
		if len(controllerRatesPatchBody) != 0 {
			patchBody["ControllerRates"] = controllerRatesPatchBody
		}
	}

	if !planAttributes.Oem.IsNull() && !planAttributes.Oem.IsUnknown() {
		patchBody["Oem"], diags = getOemPatchBody(ctx, &planAttributes, &stateAttributes)
		if diags.HasError() {
			return "", diags
		}
	}

	// get storage controller by using system id, storage id, controller id
	_, storageController, err := getStorageControllerInstance(
		service,
		plan.SystemID.ValueString(),
		plan.StorageID.ValueString(),
		plan.ControllerID.ValueString(),
	)
	if err != nil {
		diags.AddError("Error when retrieving storage controller", err.Error())
		return "", diags
	}

	url := storageController.ODataID + "/Settings"

	resp, err := service.GetClient().Patch(url, patchBody)
	if err != nil {
		diags.AddError("Patch request to IDRAC failed", err.Error())
		return
	}
	defer resp.Body.Close()

	location, err := resp.Location()
	if err != nil {
		diags.AddError("Getting location failed after patch request to IDRAC", err.Error())
		return
	}

	return location.EscapedPath(), diags
}

func getControllerRatesPatchBody(ctx context.Context, plan, state *models.StorageControllerAttributes) (map[string]interface{}, diag.Diagnostics) {
	objectAsOptions := basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true}

	var planAttributes models.ControllerRates
	diags := plan.ControllerRates.As(ctx, &planAttributes, objectAsOptions)
	if diags.HasError() {
		return nil, diags
	}

	var stateAttributes models.ControllerRates
	diags = state.ControllerRates.As(ctx, &stateAttributes, objectAsOptions)
	if diags.HasError() {
		return nil, diags
	}

	controllerRatesInfo := make(map[string]interface{})

	if !planAttributes.ConsistencyCheckRatePercent.IsNull() && !planAttributes.ConsistencyCheckRatePercent.IsUnknown() &&
		planAttributes.ConsistencyCheckRatePercent.ValueInt64() != stateAttributes.ConsistencyCheckRatePercent.ValueInt64() {
		controllerRatesInfo["ConsistencyCheckRatePercent"] = planAttributes.ConsistencyCheckRatePercent.ValueInt64()
	}
	if !planAttributes.RebuildRatePercent.IsNull() && !planAttributes.RebuildRatePercent.IsUnknown() &&
		planAttributes.RebuildRatePercent.ValueInt64() != stateAttributes.RebuildRatePercent.ValueInt64() {
		controllerRatesInfo["RebuildRatePercent"] = planAttributes.RebuildRatePercent.ValueInt64()
	}

	return controllerRatesInfo, diags
}

func getOemPatchBody(ctx context.Context, plan, state *models.StorageControllerAttributes) (map[string]interface{}, diag.Diagnostics) {
	objectAsOptions := basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true}

	var planAttributes models.OEMAttributes
	diags := plan.Oem.As(ctx, &planAttributes, objectAsOptions)
	if diags.HasError() {
		return nil, diags
	}

	var stateAttributes models.OEMAttributes
	diags = state.Oem.As(ctx, &stateAttributes, objectAsOptions)
	if diags.HasError() {
		return nil, diags
	}

	omeInfo := make(map[string]interface{})

	if !planAttributes.Dell.IsNull() && !planAttributes.Dell.IsUnknown() {
		omeInfo["Dell"], diags = getDellPatchBody(ctx, &planAttributes, &stateAttributes)
		if diags.HasError() {
			return nil, diags
		}
	}

	return omeInfo, diags
}

func getDellPatchBody(ctx context.Context, plan, state *models.OEMAttributes) (map[string]interface{}, diag.Diagnostics) {
	objectAsOptions := basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true}

	var planAttributes models.DellAttributes
	diags := plan.Dell.As(ctx, &planAttributes, objectAsOptions)
	if diags.HasError() {
		return nil, diags
	}

	var stateAttributes models.DellAttributes
	diags = state.Dell.As(ctx, &stateAttributes, objectAsOptions)
	if diags.HasError() {
		return nil, diags
	}

	dellInfo := make(map[string]interface{})

	if !planAttributes.DellStorageController.IsNull() && !planAttributes.DellStorageController.IsUnknown() {
		dellInfo["DellStorageController"], diags = getDellStorageControllerPatchBody(ctx, &planAttributes, &stateAttributes)
		if diags.HasError() {
			return nil, diags
		}
	}

	return dellInfo, diags
}

// nolint: gocyclo, gocognit, revive
func getDellStorageControllerPatchBody(ctx context.Context, plan, state *models.DellAttributes) (map[string]interface{}, diag.Diagnostics) {
	objectAsOptions := basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true}

	var planAttributes models.DellStorageControllerAttributes
	diags := plan.DellStorageController.As(ctx, &planAttributes, objectAsOptions)
	if diags.HasError() {
		return nil, diags
	}

	var stateAttributes models.DellStorageControllerAttributes
	diags = state.DellStorageController.As(ctx, &stateAttributes, objectAsOptions)
	if diags.HasError() {
		return nil, diags
	}

	dellStorageControllerInfo := make(map[string]interface{})

	// modes
	if !planAttributes.ControllerMode.IsNull() && !planAttributes.ControllerMode.IsUnknown() &&
		planAttributes.ControllerMode.ValueString() != "" &&
		planAttributes.ControllerMode.ValueString() != stateAttributes.ControllerMode.ValueString() {
		dellStorageControllerInfo["ControllerMode"] = planAttributes.ControllerMode.ValueString()
	}
	if !planAttributes.CheckConsistencyMode.IsNull() && !planAttributes.CheckConsistencyMode.IsUnknown() &&
		planAttributes.CheckConsistencyMode.ValueString() != "" &&
		planAttributes.CheckConsistencyMode.ValueString() != stateAttributes.CheckConsistencyMode.ValueString() {
		dellStorageControllerInfo["CheckConsistencyMode"] = planAttributes.CheckConsistencyMode.ValueString()
	}
	if !planAttributes.CopybackMode.IsNull() && !planAttributes.CopybackMode.IsUnknown() &&
		planAttributes.CopybackMode.ValueString() != "" &&
		planAttributes.CopybackMode.ValueString() != stateAttributes.CopybackMode.ValueString() {
		dellStorageControllerInfo["CopybackMode"] = planAttributes.CopybackMode.ValueString()
	}
	if !planAttributes.LoadBalanceMode.IsNull() && !planAttributes.LoadBalanceMode.IsUnknown() &&
		planAttributes.LoadBalanceMode.ValueString() != "" &&
		planAttributes.LoadBalanceMode.ValueString() != stateAttributes.LoadBalanceMode.ValueString() {
		dellStorageControllerInfo["LoadBalanceMode"] = planAttributes.LoadBalanceMode.ValueString()
	}
	if !planAttributes.EnhancedAutoImportForeignConfigurationMode.IsNull() && !planAttributes.EnhancedAutoImportForeignConfigurationMode.IsUnknown() &&
		planAttributes.EnhancedAutoImportForeignConfigurationMode.ValueString() != "" &&
		planAttributes.EnhancedAutoImportForeignConfigurationMode.ValueString() != stateAttributes.EnhancedAutoImportForeignConfigurationMode.ValueString() {
		dellStorageControllerInfo["EnhancedAutoImportForeignConfigurationMode"] = planAttributes.EnhancedAutoImportForeignConfigurationMode.ValueString()
	}

	// rates
	if !planAttributes.PatrolReadUnconfiguredAreaMode.IsNull() && !planAttributes.PatrolReadUnconfiguredAreaMode.IsUnknown() &&
		planAttributes.PatrolReadUnconfiguredAreaMode.ValueString() != "" &&
		planAttributes.PatrolReadUnconfiguredAreaMode.ValueString() != stateAttributes.PatrolReadUnconfiguredAreaMode.ValueString() {
		dellStorageControllerInfo["PatrolReadUnconfiguredAreaMode"] = planAttributes.PatrolReadUnconfiguredAreaMode.ValueString()
	}
	if !planAttributes.PatrolReadMode.IsNull() && !planAttributes.PatrolReadMode.IsUnknown() &&
		planAttributes.PatrolReadMode.ValueString() != "" &&
		planAttributes.PatrolReadMode.ValueString() != stateAttributes.PatrolReadMode.ValueString() {
		dellStorageControllerInfo["PatrolReadMode"] = planAttributes.PatrolReadMode.ValueString()
	}
	if !planAttributes.BackgroundInitializationRatePercent.IsNull() && !planAttributes.BackgroundInitializationRatePercent.IsUnknown() &&
		planAttributes.BackgroundInitializationRatePercent.ValueInt64() != stateAttributes.BackgroundInitializationRatePercent.ValueInt64() {
		dellStorageControllerInfo["BackgroundInitializationRatePercent"] = planAttributes.BackgroundInitializationRatePercent.ValueInt64()
	}
	if !planAttributes.ReconstructRatePercent.IsNull() && !planAttributes.ReconstructRatePercent.IsUnknown() &&
		planAttributes.ReconstructRatePercent.ValueInt64() != stateAttributes.ReconstructRatePercent.ValueInt64() {
		dellStorageControllerInfo["ReconstructRatePercent"] = planAttributes.ReconstructRatePercent.ValueInt64()
	}

	return dellStorageControllerInfo, diags
}

// nolint: gocyclo, gocognit, revive
func updateSecurityAttributes(ctx context.Context, service *gofish.Service, plan, state *models.StorageControllerResource) (string, diag.Diagnostics) {
	var diags diag.Diagnostics

	isGenerationSeventeenAndAbove, err := isServerGenerationSeventeenAndAbove(service)
	if err != nil {
		diags.AddError("Error retrieving the server generation", err.Error())
		return "", diags
	}

	objectAsOptions := basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true}

	var planAttributes models.SecurityAttributes
	diags = plan.Security.As(ctx, &planAttributes, objectAsOptions)
	if diags.HasError() {
		return "", diags
	}

	var stateAttributes models.SecurityAttributes
	diags = state.Security.As(ctx, &stateAttributes, objectAsOptions)
	if diags.HasError() {
		return "", diags
	}

	if planAttributes.Action.IsNull() || planAttributes.Action.IsUnknown() {
		diags.AddError("Security updates will not be applied since the `action` is not specified.", "`action` is either null or unknown.")
		return "", diags
	}
	securityAction := planAttributes.Action.ValueString()

	// Create the request body using security action
	postBody := make(map[string]interface{})

	if securityAction == "SetControllerKey" {
		if isGenerationSeventeenAndAbove {
			diags.AddError("If server generation is 17G and above, the action `SetControllerKey` is not supported.",
				"In lesser than 17G, the action `SetControllerKey` is supported.")
			return "", diags
		}

		if planAttributes.KeyID.IsNull() || planAttributes.KeyID.IsUnknown() {
			diags.AddError("With `action` set to `SetControllerKey`, the `key_id` needs to be set.", "`key_id` is not set.")
			return "", diags
		}

		if planAttributes.Key.IsNull() || planAttributes.Key.IsUnknown() {
			diags.AddError("With `action` set to `SetControllerKey`, the `key` needs to be set.", "`key` is not set.")
			return "", diags
		}

		if !planAttributes.OldKey.IsNull() && !planAttributes.OldKey.IsUnknown() {
			diags.AddError("With `action` set to `SetControllerKey`, the `old_key` needs to be commented.", "`old_key` is not commented.")
			return "", diags
		}

		if !planAttributes.Mode.IsNull() && !planAttributes.Mode.IsUnknown() {
			diags.AddError("With `action` set to `SetControllerKey`, the `mode` needs to be commented.", "`mode` is not commented.")
			return "", diags
		}

		postBody["Keyid"] = planAttributes.KeyID.ValueString()
		postBody["Key"] = planAttributes.Key.ValueString()
		postBody["TargetFQDD"] = plan.ControllerID.ValueString()
	} else if securityAction == "ReKey" {
		if isGenerationSeventeenAndAbove {
			diags.AddError("If server generation is 17G and above, the action `ReKey` is not supported.",
				"In lesser than 17G, the action `ReKey` is supported.")
			return "", diags
		}

		if planAttributes.KeyID.IsNull() || planAttributes.KeyID.IsUnknown() {
			diags.AddError("With `action` set to `ReKey`, the `key_id` needs to be set.", "`key_id` is not set.")
			return "", diags
		}

		if planAttributes.Key.IsNull() || planAttributes.Key.IsUnknown() {
			diags.AddError("With `action` set to `ReKey`, the `key` needs to be set.", "`key` is not set.")
			return "", diags
		}

		if planAttributes.OldKey.IsNull() || planAttributes.OldKey.IsUnknown() {
			diags.AddError("With `action` set to `ReKey`, the `old_key` needs to be set.", "`old_key` is not set.")
			return "", diags
		}

		if planAttributes.Mode.IsNull() || planAttributes.Mode.IsUnknown() {
			diags.AddError("With `action` set to `ReKey`, the `mode` needs to be set.", "`mode` is not set.")
			return "", diags
		}

		postBody["Keyid"] = planAttributes.KeyID.ValueString()
		postBody["Mode"] = planAttributes.Mode.ValueString()
		postBody["NewKey"] = planAttributes.Key.ValueString()
		postBody["OldKey"] = planAttributes.OldKey.ValueString()
		postBody["TargetFQDD"] = plan.ControllerID.ValueString()
	} else if securityAction == "RemoveControllerKey" {
		if isGenerationSeventeenAndAbove {
			diags.AddError("If server generation is 17G and above, the action `RemoveControllerKey` is not supported.",
				"In lesser than 17G, the action `RemoveControllerKey` is supported.")
			return "", diags
		}

		if !planAttributes.KeyID.IsNull() && !planAttributes.KeyID.IsUnknown() {
			diags.AddError("With `action` set to `RemoveControllerKey`, the `key_id` needs to be commented.", "`key_id` is not commented.")
			return "", diags
		}

		if !planAttributes.Key.IsNull() && !planAttributes.Key.IsUnknown() {
			diags.AddError("With `action` set to `RemoveControllerKey`, the `key` needs to be commented.", "`key` is not commented.")
			return "", diags
		}

		if !planAttributes.OldKey.IsNull() && !planAttributes.OldKey.IsUnknown() {
			diags.AddError("With `action` set to `RemoveControllerKey`, the `old_key` needs to be commented.", "`old_key` is not commented.")
			return "", diags
		}

		if !planAttributes.Mode.IsNull() && !planAttributes.Mode.IsUnknown() {
			diags.AddError("With `action` set to `RemoveControllerKey`, the `mode` needs to be commented.", "`mode` is not commented.")
			return "", diags
		}

		postBody["TargetFQDD"] = plan.ControllerID.ValueString()
	} else if securityAction == "EnableSecurity" {
		if !isGenerationSeventeenAndAbove {
			diags.AddError("If server generation is lesser than 17G, the action `EnableSecurity` is not supported.",
				"In 17G and above, the action `EnableSecurity` is supported.")
			return "", diags
		}

		if !planAttributes.KeyID.IsNull() && !planAttributes.KeyID.IsUnknown() {
			diags.AddError("With `action` set to `EnableSecurity`, the `key_id` needs to be commented.", "`key_id` is not commented.")
			return "", diags
		}

		if !planAttributes.Key.IsNull() && !planAttributes.Key.IsUnknown() {
			diags.AddError("With `action` set to `EnableSecurity`, the `key` needs to be commented.", "`key` is not commented.")
			return "", diags
		}

		if !planAttributes.OldKey.IsNull() && !planAttributes.OldKey.IsUnknown() {
			diags.AddError("With `action` set to `EnableSecurity`, the `old_key` needs to be commented.", "`old_key` is not commented.")
			return "", diags
		}

		if !planAttributes.Mode.IsNull() && !planAttributes.Mode.IsUnknown() {
			diags.AddError("With `action` set to `EnableSecurity`, the `mode` needs to be commented.", "`mode` is not commented.")
			return "", diags
		}

		postBody["TargetFQDD"] = plan.ControllerID.ValueString()
	} else if securityAction == "DisableSecurity" {
		if !isGenerationSeventeenAndAbove {
			diags.AddError("If server generation is lesser than 17G, the action `DisableSecurity` is not supported.",
				"In 17G and above, the action `DisableSecurity` is supported.")
			return "", diags
		}

		if !planAttributes.KeyID.IsNull() && !planAttributes.KeyID.IsUnknown() {
			diags.AddError("With `action` set to `DisableSecurity`, the `key_id` needs to be commented.", "`key_id` is not commented.")
			return "", diags
		}

		if !planAttributes.Key.IsNull() && !planAttributes.Key.IsUnknown() {
			diags.AddError("With `action` set to `DisableSecurity`, the `key` needs to be commented.", "`key` is not commented.")
			return "", diags
		}

		if !planAttributes.OldKey.IsNull() && !planAttributes.OldKey.IsUnknown() {
			diags.AddError("With `action` set to `DisableSecurity`, the `old_key` needs to be commented.", "`old_key` is not commented.")
			return "", diags
		}

		if !planAttributes.Mode.IsNull() && !planAttributes.Mode.IsUnknown() {
			diags.AddError("With `action` set to `DisableSecurity`, the `mode` needs to be commented.", "`mode` is not commented.")
			return "", diags
		}

		postBody["ControllerFQDD"] = plan.ControllerID.ValueString()
	}

	// Create the url using security action
	// get storage controller by using system id, storage id, controller id
	system, storageController, err := getStorageControllerInstance(
		service,
		plan.SystemID.ValueString(),
		plan.StorageID.ValueString(),
		plan.ControllerID.ValueString(),
	)
	if err != nil {
		diags.AddError("Error when retrieving storage controller", err.Error())
		return "", diags
	}

	storageControllerODataID := storageController.ODataID
	systemID := system.ID
	pathToAppend := "/Oem/Dell/DellRaidService/Actions/DellRaidService." + securityAction

	idx := strings.Index(storageControllerODataID, systemID)
	url := storageControllerODataID[:idx+len(systemID)] + pathToAppend

	// Make the POST call using the url and the request body
	resp, err := service.GetClient().Post(url, postBody)
	if err != nil {
		diags.AddError("Post request to IDRAC failed", err.Error())
		return "", diags
	}
	defer resp.Body.Close()

	location, err := resp.Location()
	if err != nil {
		diags.AddError("Getting location failed after post request to IDRAC", err.Error())
		return "", diags
	}

	return location.EscapedPath(), diags
}
