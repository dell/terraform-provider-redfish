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
	"encoding/json"
	"strings"
	"terraform-provider-redfish/common"
	"terraform-provider-redfish/gofish/dell"
	"terraform-provider-redfish/redfish/models"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/stmcginnis/gofish"
	redfishcommon "github.com/stmcginnis/gofish/common"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource = &RedfishStorageControllerResource{}
)

// NewRedfishStorageControllerResource is a helper function to simplify the provider implementation.
func NewRedfishStorageControllerResource() resource.Resource {
	return &RedfishStorageControllerResource{}
}

// RedfishStorageControllerResource is the resource implementation.
type RedfishStorageControllerResource struct {
	p *redfishProvider
}

// Configure implements resource.ResourceWithConfigure
func (r *RedfishStorageControllerResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	r.p = req.ProviderData.(*redfishProvider)
}

// Metadata returns the resource type name.
func (*RedfishStorageControllerResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "storage_controller"
}

// Schema defines the schema for the resource.
func (*RedfishStorageControllerResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "This Terraform resource is used to configure the storage controller. " +
			"We can read the existing configurations or modify them using this resource.",
		Description: "This Terraform resource is used to configure the storage controller. " +
			"We can read the existing configurations or modify them using this resource.",
		Attributes: StorageControllerResourceSchema(),
		Blocks:     RedfishServerResourceBlockMap(),
	}
}

// Create creates the resource and sets the initial Terraform state.
func (r *RedfishStorageControllerResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	tflog.Trace(ctx, "resource_RedfishStorageController create: started")

	var plan, emptyState models.StorageControllerResource
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	api, err := NewConfig(r.p, &plan.RedfishServer)
	if err != nil {
		resp.Diagnostics.AddError("service error", err.Error())
		return
	}
	service := api.Service
	defer api.Logout()

	// update
	diags = updateRedfishStorageController(ctx, service, &emptyState, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// read
	diags = readRedfishStorageController(ctx, service, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// save into the state
	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)

	tflog.Trace(ctx, "resource_RedfishStorageController create: finished")
}

// Read refreshes the Terraform state with the latest data.
func (r *RedfishStorageControllerResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	tflog.Trace(ctx, "resource_RedfishStorageController read: started")

	var state models.StorageControllerResource
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	api, err := NewConfig(r.p, &state.RedfishServer)
	if err != nil {
		resp.Diagnostics.AddError("service error", err.Error())
		return
	}
	service := api.Service
	defer api.Logout()

	diags = readRedfishStorageController(ctx, service, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Trace(ctx, "resource_RedfishStorageController read: finished reading state")
	// Save into State
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)

	tflog.Trace(ctx, "resource_RedfishStorageController read: finished")
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *RedfishStorageControllerResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	tflog.Trace(ctx, "resource_RedfishStorageController update: started")

	var state, plan models.StorageControllerResource
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	api, err := NewConfig(r.p, &plan.RedfishServer)
	if err != nil {
		resp.Diagnostics.AddError("service error", err.Error())
		return
	}
	service := api.Service
	defer api.Logout()

	// checks
	if plan.ControllerID.ValueString() != state.ControllerID.ValueString() {
		resp.Diagnostics.AddError("Error when updating with invalid input", "may not change resource `controller_id`")
	}
	if plan.StorageID.ValueString() != state.StorageID.ValueString() {
		resp.Diagnostics.AddError("Error when updating with invalid input", "may not change resource `storage_id`")
	}
	if plan.SystemID.ValueString() != "" && plan.SystemID.ValueString() != state.SystemID.ValueString() {
		resp.Diagnostics.AddError("Error when updating with invalid input", "may not change resource `system_id`")
	}
	if resp.Diagnostics.HasError() {
		return
	}

	// update
	diags = updateRedfishStorageController(ctx, service, &state, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// read
	diags = readRedfishStorageController(ctx, service, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// save into state
	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)

	tflog.Trace(ctx, "resource_RedfishStorageController update: finished")
}

// Delete deletes the resource and removes the Terraform state on success.
func (*RedfishStorageControllerResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	tflog.Trace(ctx, "resource_RedfishStorageController delete: started")
	// Get State Data
	var state models.StorageControllerResource
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.State.RemoveResource(ctx)
	tflog.Trace(ctx, "resource_RedfishStorageController delete: finished")
}

// ImportState import state for existing storage controller
func (*RedfishStorageControllerResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	type creds struct {
		Username     string `json:"username"`
		Password     string `json:"password"`
		Endpoint     string `json:"endpoint"`
		SslInsecure  bool   `json:"ssl_insecure"`
		SystemID     string `json:"system_id"`
		StorageID    string `json:"storage_id"`
		ControllerID string `json:"controller_id"`
		RedfishAlias string `json:"redfish_alias"`
	}

	var c creds
	err := json.Unmarshal([]byte(req.ID), &c)
	if err != nil {
		resp.Diagnostics.AddError("Error while unmarshalling id", err.Error())
	}

	server := models.RedfishServer{
		User:         types.StringValue(c.Username),
		Password:     types.StringValue(c.Password),
		Endpoint:     types.StringValue(c.Endpoint),
		SslInsecure:  types.BoolValue(c.SslInsecure),
		RedfishAlias: types.StringValue(c.RedfishAlias),
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), "importId")...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("redfish_server"), []models.RedfishServer{server})...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("system_id"), c.SystemID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("storage_id"), c.StorageID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("controller_id"), c.ControllerID)...)
}

func readRedfishStorageController(ctx context.Context, service *gofish.Service, state *models.StorageControllerResource) diag.Diagnostics {
	var diags diag.Diagnostics

	// get storage controller by using system id, storage id, controller id
	system, storageController, err := getStorageControllerInstance(
		service,
		state.SystemID.ValueString(),
		state.StorageID.ValueString(),
		state.ControllerID.ValueString(),
	)
	if err != nil {
		diags.AddError("Error when retrieving storage controller", err.Error())
		return diags
	}

	storageControllerExtended, err := dell.StorageController(storageController)
	if err != nil {
		diags.AddError("Error when retrieving storage controller extended", err.Error())
		return diags
	}

	diags = parseStorageControllerExtendedIntoState(ctx, storageControllerExtended, state)
	if diags.HasError() {
		return diags
	}

	diags = parseSecurityAttributesIntoState(ctx, storageControllerExtended, state)
	if diags.HasError() {
		return diags
	}

	state.ID = types.StringValue("redfish_storage_controller_resource")
	state.SystemID = types.StringValue(system.ID)
	return diags
}

// nolint: gocyclo, gocognit, revive
func updateRedfishStorageController(ctx context.Context, service *gofish.Service, state, plan *models.StorageControllerResource) diag.Diagnostics {
	var diags diag.Diagnostics

	applyTime := plan.ApplyTime.ValueString()
	resetType := plan.ResetType.ValueString()
	resetTimeout := plan.ResetTimeout.ValueInt64()
	jobTimeout := plan.JobTimeout.ValueInt64()

	jobWait := true
	if applyTime == string(redfishcommon.AtMaintenanceWindowStartApplyTime) ||
		applyTime == string(redfishcommon.InMaintenanceWindowOnResetApplyTime) {
		if plan.MaintenanceWindow == nil || plan.MaintenanceWindow.StartTime.IsUnknown() {
			diags.AddError("Input param is not valid",
				"Please set `maintenance_window` when `apply_time` is `AtMaintenanceWindowStart` or `InMaintenanceWindowOnReset`")
			return diags
		}
		// when apply_time is AtMaintenanceWindowStart or InMaintenanceWindowOnReset, skip wait for job to finish
		jobWait = false
	}

	// Lock the mutex to avoid race conditions with other resources
	redfishMutexKV.Lock(plan.RedfishServer[0].Endpoint.ValueString())
	defer redfishMutexKV.Unlock(plan.RedfishServer[0].Endpoint.ValueString())

	// OnReset case
	if applyTime == string(redfishcommon.OnResetApplyTime) {
		// Reboot the server
		pOp := powerOperator{ctx, service, plan.SystemID.ValueString()}
		_, err := pOp.PowerOperation(resetType, resetTimeout, intervalStorageControllerJobCheckTime)
		if err != nil {
			diags.AddError(RedfishJobErrorMsg, err.Error())
			return diags
		}
	}

	isControllerModeAttributeChanged := storageControllerAttributesChanged(ctx, plan, state, true)
	isAnyOtherStorageControllerAttributeChanged := storageControllerAttributesChanged(ctx, plan, state, false)
	isAnySecurityAttributeChanged := securityAttributesChanged(ctx, plan, state)

	var jobURL string
	if isControllerModeAttributeChanged {
		if isAnyOtherStorageControllerAttributeChanged || isAnySecurityAttributeChanged {
			diags.AddError("While updating `controller_mode`, no other property should be changed.",
				"Along with `controller_mode`, some other property is changed.")
			return diags
		}

		_, controllerModeVal := getStorageControllerAttributeInfo(ctx, plan, "ControllerMode")
		isEnhancedAutoImportForeignConfigurationModeUnknown, enhancedAutoImportForeignConfigurationModeVal := getStorageControllerAttributeInfo(ctx, plan, "EnhancedAutoImportForeignConfigurationMode")

		if (controllerModeVal == "HBA") && !isEnhancedAutoImportForeignConfigurationModeUnknown && (enhancedAutoImportForeignConfigurationModeVal == "Enabled") {
			diags.AddError("Either with `controller_mode` attribute set to `RAID`, set `enhanced_auto_import_foreign_configuration_mode` attribute to `Disabled` first "+
				"or now that the `controller_mode` attribute is set to `HBA`, ensure `enhanced_auto_import_foreign_configuration_mode` attribute is commented.",
				"The `enhanced_auto_import_foreign_configuration_mode` gets `Disabled` in the `HBA` controller mode.")
			return diags
		}

		if !(applyTime == string(redfishcommon.OnResetApplyTime) || applyTime == string(redfishcommon.InMaintenanceWindowOnResetApplyTime)) {
			diags.AddError("While updating `controller_mode`, the `apply_time` should be `OnReset` or `InMaintenanceWindowOnReset`.",
				"The `apply_time` is not `OnReset` or `InMaintenanceWindowOnReset`.")
			return diags
		}

		jobURL, diags = updateStorageControllerAttributes(ctx, service, plan, state)
		if diags.HasError() {
			return diags
		}
	} else {
		if isAnySecurityAttributeChanged && isAnyOtherStorageControllerAttributeChanged {
			diags.AddError("Attributes of both `security` and `storage_controller` were changed.",
				"At a time, update the attributes of any one out of `security` and `storage_controller`.")
			return diags
		}

		if isAnySecurityAttributeChanged {
			jobURL, diags = updateSecurityAttributes(ctx, service, plan, state)
			if diags.HasError() {
				return diags
			}
		}

		if isAnyOtherStorageControllerAttributeChanged {
			jobURL, diags = updateStorageControllerAttributes(ctx, service, plan, state)
			if diags.HasError() {
				return diags
			}
		}
	}

	if !isControllerModeAttributeChanged && !isAnyOtherStorageControllerAttributeChanged && !isAnySecurityAttributeChanged {
		jobWait = false
		tflog.Trace(ctx, "No attributes changed. Skip update for Storage Controller.")
	}

	var err error
	if jobWait && jobURL != "" {
		// jobURL could contain Jobs or Tasks
		if strings.Contains(jobURL, "Job") {
			err = common.WaitForJobToFinish(service, jobURL, intervalStorageControllerJobCheckTime, jobTimeout)
		} else {
			err = common.WaitForTaskToFinish(service, jobURL, intervalStorageControllerJobCheckTime, jobTimeout)
		}
		if err != nil {
			diags.AddError(RedfishJobErrorMsg, err.Error())
			return diags
		}
	}

	if isControllerModeAttributeChanged {
		// controller mode changes take additional time to reflect.
		time.Sleep(240 * time.Second)
	}

	time.Sleep(60 * time.Second)
	tflog.Trace(ctx, "Job has been completed")

	return diags
}

// nolint: gocyclo, gocognit, revive
func updateSecurityAttributes(ctx context.Context, service *gofish.Service, plan, state *models.StorageControllerResource) (string, diag.Diagnostics) {
	var diags diag.Diagnostics

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

	postBody := make(map[string]interface{})
	postBody["TargetFQDD"] = plan.ControllerID.ValueString()

	if securityAction == "SetControllerKey" {
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
	} else if securityAction == "ReKey" {
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
	} else if securityAction == "RemoveControllerKey" {
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
	}

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
