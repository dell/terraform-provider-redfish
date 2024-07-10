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
	"errors"
	"fmt"
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
	_ resource.Resource = &RedfishNICResource{}
)

// NewRedfishNICResource is a helper function to simplify the provider implementation.
func NewRedfishNICResource() resource.Resource {
	return &RedfishNICResource{}
}

// RedfishNICResource is the resource implementation.
type RedfishNICResource struct {
	p *redfishProvider
}

// Configure implements resource.ResourceWithConfigure
func (r *RedfishNICResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	r.p = req.ProviderData.(*redfishProvider)
}

// Metadata returns the resource type name.
func (*RedfishNICResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "network_adapter"
}

// Schema defines the schema for the resource.
func (*RedfishNICResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "This Terraform resource is used to configure the port and partition network attributes on " +
			"the network interface cards(NIC). We can Read the existing configurations or modify them using this resource.",
		Description: "This Terraform resource is used to configure the port and partition network attributes on " +
			"the network interface cards(NIC). We can Read the existing configurations or modify them using this resource.",
		Attributes: NICResourceSchema(),
		Blocks:     RedfishServerResourceBlockMap(),
	}
}

// Create creates the resource and sets the initial Terraform state.
func (r *RedfishNICResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	tflog.Trace(ctx, "resource_RedfishNIC create : Started")
	// Get Plan Data
	var plan, emptyState models.NICResource
	diags := req.Plan.Get(ctx, &plan)
	if resp.Diagnostics.Append(diags...); resp.Diagnostics.HasError() {
		return
	}

	api, err := NewConfig(r.p, &plan.RedfishServer)
	if err != nil {
		resp.Diagnostics.AddError(ServiceErrorMsg, err.Error())
		return
	}
	service := api.Service
	defer api.Logout()

	if networkAttributesChanged(ctx, &plan, &emptyState) && oemNetworkAttributesChanged(ctx, &plan, &emptyState) {
		resp.Diagnostics.AddError("Error when creating both of `network_attributes` and `oem_network_attributes`",
			noteMessageUpdateOneAttrsOnly)
		return
	}
	diags = updateRedfishNIC(ctx, service, &emptyState, &plan)
	if resp.Diagnostics.Append(diags...); resp.Diagnostics.HasError() {
		return
	}
	tflog.Trace(ctx, "resource_RedfishNIC create: updating remote settings finished, saving ...")

	diags = readRedfishNIC(ctx, service, &plan)
	if resp.Diagnostics.Append(diags...); resp.Diagnostics.HasError() {
		return
	}
	tflog.Trace(ctx, "resource_RedfishNIC create: updating state finished, saving ...")
	// Save into State
	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	tflog.Trace(ctx, "resource_RedfishNIC create: finish")
}

// Read refreshes the Terraform state with the latest data.
func (r *RedfishNICResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	tflog.Trace(ctx, "resource_RedfishNIC read: started")
	var state models.NICResource
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

	diags = readRedfishNIC(ctx, service, &state)
	if resp.Diagnostics.Append(diags...); resp.Diagnostics.HasError() {
		return
	}

	tflog.Trace(ctx, "resource_RedfishNIC read: finished reading state")
	// Save into State
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	tflog.Trace(ctx, "resource_RedfishNIC read: finished")
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *RedfishNICResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Get state Data
	tflog.Trace(ctx, "resource_RedfishNIC update: started")
	var state, plan models.NICResource
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get plan Data
	diags = req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	api, err := NewConfig(r.p, &plan.RedfishServer)
	if err != nil {
		resp.Diagnostics.AddError(ServiceErrorMsg, err.Error())
		return
	}
	service := api.Service
	defer api.Logout()

	if networkAttributesChanged(ctx, &plan, &state) && oemNetworkAttributesChanged(ctx, &plan, &state) {
		resp.Diagnostics.AddError("Error when updating both of `network_attributes` and `oem_network_attributes`",
			noteMessageUpdateOneAttrsOnly)
	}
	if plan.NetworkAdapterID.ValueString() != state.NetworkAdapterID.ValueString() {
		resp.Diagnostics.AddError("Error when updating with invalid input", "may not change resource `network_adapter_id`")
	}
	if plan.NetworkDeviceFunctionID.ValueString() != state.NetworkDeviceFunctionID.ValueString() {
		resp.Diagnostics.AddError("Error when updating with invalid input", "may not change resource `network_device_function_id`")
	}
	if plan.SystemID.ValueString() != "" && plan.SystemID.ValueString() != state.SystemID.ValueString() {
		resp.Diagnostics.AddError("Error when updating with invalid input", "may not change resource `system_id`")
	}
	if resp.Diagnostics.HasError() {
		return
	}

	diags = updateRedfishNIC(ctx, service, &state, &plan)
	if resp.Diagnostics.Append(diags...); resp.Diagnostics.HasError() {
		return
	}
	tflog.Trace(ctx, "resource_RedfishNIC update: finished remote settings update")

	diags = readRedfishNIC(ctx, service, &plan)
	if resp.Diagnostics.Append(diags...); resp.Diagnostics.HasError() {
		return
	}
	tflog.Trace(ctx, "resource_RedfishNIC update: finished state update")

	// Save into State
	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	tflog.Trace(ctx, "resource_RedfishNIC update: finished")
}

// Delete deletes the resource and removes the Terraform state on success.
// nolint: unused,revive
func (r *RedfishNICResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	tflog.Trace(ctx, "resource_RedfishNIC delete: started")
	// Get State Data
	var state models.NICResource
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.State.RemoveResource(ctx)
	tflog.Trace(ctx, "resource_RedfishNIC delete: finished")
}

// ImportState import state for existing nic
func (*RedfishNICResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	type creds struct {
		Username                string `json:"username"`
		Password                string `json:"password"`
		Endpoint                string `json:"endpoint"`
		SslInsecure             bool   `json:"ssl_insecure"`
		SystemID                string `json:"system_id"`
		NetworkAdapterID        string `json:"network_adapter_id"`
		NetworkDeviceFunctionID string `json:"network_device_function_id"`
	}

	var c creds
	err := json.Unmarshal([]byte(req.ID), &c)
	if err != nil {
		resp.Diagnostics.AddError("Error while unmarshalling id", err.Error())
	}
	server := models.RedfishServer{
		User:        types.StringValue(c.Username),
		Password:    types.StringValue(c.Password),
		Endpoint:    types.StringValue(c.Endpoint),
		SslInsecure: types.BoolValue(c.SslInsecure),
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root(NICComponmentSchemaID), "importId")...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("redfish_server"), []models.RedfishServer{server})...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("system_id"), c.SystemID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("network_adapter_id"), c.NetworkAdapterID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("network_device_function_id"), c.NetworkDeviceFunctionID)...)
}

func updateRedfishNIC(ctx context.Context, service *gofish.Service, state, plan *models.NICResource) diag.Diagnostics {
	var diags diag.Diagnostics

	applyTime := plan.ApplyTime.ValueString()
	jobWait := true
	resetType := plan.ResetType.ValueString()
	resetTimeout := plan.ResetTimeout.ValueInt64()
	jobTimeout := plan.JobTimeout.ValueInt64()
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
		_, err := pOp.PowerOperation(resetType, resetTimeout, intervalNICJobCheckTime)
		if err != nil {
			diags.AddError(RedfishJobErrorMsg, err.Error())
			return diags
		}
	}

	var jobURL string
	if oemNetworkAttributesChanged(ctx, plan, state) {
		jobURL, diags = updateNicOemNetworkAttributes(ctx, service, plan)
	} else if networkAttributesChanged(ctx, plan, state) {
		jobURL, diags = updateNicNetworktributes(ctx, service, plan)
	} else {
		jobWait = false
		tflog.Trace(ctx, "Both `oem_network_attributes` and `network_attributes` are not changed. Skip Update for NIC.")
	}
	if diags.HasError() {
		return diags
	}

	var err error
	if jobWait && jobURL != "" {
		// jobURL could be JobService and TaskService
		if strings.Contains(jobURL, "JobService") {
			err = common.WaitForJobToFinish(service, jobURL, intervalNICJobCheckTime, jobTimeout)
		} else {
			err = common.WaitForTaskToFinish(service, jobURL, intervalNICJobCheckTime, jobTimeout)
		}
		if err != nil {
			diags.AddError(RedfishJobErrorMsg, err.Error())
			return diags
		}
	}
	time.Sleep(60 * time.Second)
	tflog.Trace(ctx, "Job has been completed")

	return diags
}

// nolint: gocyclo,revive
func updateNicNetworktributes(ctx context.Context, service *gofish.Service, plan *models.NICResource) (jobURL string, diags diag.Diagnostics) {
	tflog.Info(ctx, "updateNicNetworktributes: started")
	applyTime := plan.ApplyTime.ValueString()
	networkAttrsError := "there was an issue when creating/updating network attributes: "
	if applyTime == string(redfishcommon.ImmediateApplyTime) {
		diags.AddError(networkAttrsError+errMessageInvalidInput, "`Immediate` is not supported by `network_attributes`")
		return
	}

	objectAsOptions := basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true}
	var attrsState models.NetworkDeviceFunctionSettings
	if diags = plan.Networktributes.As(ctx, &attrsState, objectAsOptions); diags.HasError() {
		return
	}

	// get networkDeviceFunction by system id, adapter id and networkDeviceFunction id
	_, networketworkDeviceFunc, err := getNetworkDeviceFunction(service, plan.SystemID.ValueString(),
		plan.NetworkAdapterID.ValueString(), plan.NetworkDeviceFunctionID.ValueString())
	if err != nil {
		diags.AddError(fmt.Sprintf("%s: error when retrieving NetworkDeviceFunction", networkAttrsError), err.Error())
		return
	}
	// get OEM data
	dellDeviceFunction, _ := dell.NetworkDeviceFunction(networketworkDeviceFunc)
	if dellDeviceFunction.SettingsObject.ODataID == "" {
		diags.AddError(networkAttrsError, "error get NetworkAttributes SettingsObject from NetworkDeviceFunction Extension")
		return
	}

	var ethChangable, fcChangable, iscsiChangable bool

	if !attrsState.FibreChannel.IsNull() && !attrsState.FibreChannel.IsUnknown() {
		if networketworkDeviceFunc.FibreChannel.PermanentWWNN == "" && networketworkDeviceFunc.FibreChannel.PermanentWWPN == "" {
			diags.AddError(networkAttrsError+errMessageInvalidInput, "maynot configure `fibre_channel` if no FibreChannel NIC is present")
			return
		}
		fcChangable = true
	}
	if !attrsState.Ethernet.IsNull() && !attrsState.Ethernet.IsUnknown() {
		if networketworkDeviceFunc.Ethernet.PermanentMACAddress == "" {
			diags.AddError(networkAttrsError+errMessageInvalidInput, "maynot configure `ethernet` if no Ethernet NIC is present")
			return
		}
		ethChangable = true
	}
	if !attrsState.ISCSIBoot.IsNull() && !attrsState.ISCSIBoot.IsUnknown() {
		if networketworkDeviceFunc.ISCSIBoot.AuthenticationMethod == "" && networketworkDeviceFunc.ISCSIBoot.IPAddressType == "" {
			diags.AddError(networkAttrsError+errMessageInvalidInput, "maynot configure `iscsi_boot` if no ISCSIBoot is present")
			return
		}
		iscsiChangable = true
	}

	// Set the body to send
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
	if iscsiChangable {
		if patchBody["iSCSIBoot"], diags = getIscsiBootPatchBody(ctx, &attrsState); diags.HasError() {
			return "", diags
		}
	}

	if ethChangable {
		if patchBody["Ethernet"], diags = getEthernetPatchBody(ctx, &attrsState); diags.HasError() {
			return "", diags
		}
	}
	if fcChangable {
		if patchBody["FibreChannel"], diags = getFibreChannelPatchBody(ctx, &attrsState); diags.HasError() {
			return "", diags
		}
	}
	if !attrsState.NetDevFuncType.IsUnknown() && !attrsState.NetDevFuncType.IsNull() {
		patchBody["NetDevFuncType"] = attrsState.NetDevFuncType.ValueString()
	}

	resp, err := service.GetClient().Patch(dellDeviceFunction.SettingsObject.ODataID, patchBody)
	if err != nil {
		diags.AddError(fmt.Sprintf("%s: patch request to iDRAC failed", networkAttrsError), err.Error())
		return
	}
	defer resp.Body.Close() // #nosec G104

	// check if location is present in the response header
	location, err := resp.Location()
	if err != nil {
		diags.AddError(fmt.Sprintf("%s: get location failed after patch request to iDRAC ", networkAttrsError), err.Error())
		return
	}
	return location.EscapedPath(), diags
}

func updateNicOemNetworkAttributes(ctx context.Context, service *gofish.Service, plan *models.NICResource) (jobURL string, diags diag.Diagnostics) {
	tflog.Info(ctx, "updateNicOemNetworkAttributes: started")
	applyTime := plan.ApplyTime.ValueString()
	oemNetworkAttrsError := "there was an issue when creating/updating ome network attributes"

	objectAsOptions := basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true}
	var oemAttrsState models.OemNetworkAttributes
	if diags = plan.OemNetworkAttributes.As(ctx, &oemAttrsState, objectAsOptions); diags.HasError() {
		return
	}

	// get networkDeviceFunction by system id, adapter id and networkDeviceFunction id
	_, networketworkDeviceFunc, err := getNetworkDeviceFunction(service, plan.SystemID.ValueString(),
		plan.NetworkAdapterID.ValueString(), plan.NetworkDeviceFunctionID.ValueString())
	if err != nil {
		diags.AddError(fmt.Sprintf("%s: error when retrieving NetworkDeviceFunction", oemNetworkAttrsError), err.Error())
		return
	}

	// get OEM data
	dellDeviceFunction, _ := dell.NetworkDeviceFunction(networketworkDeviceFunc)
	if dellDeviceFunction.DellNetworkAttributes.ODataID == "" {
		diags.AddError(oemNetworkAttrsError, "error get DellNetworkAttributes ODataID from NetworkDeviceFunction Extension")
		return
	}
	oemNetworkAttrURL := dellDeviceFunction.DellNetworkAttributes.ODataID
	oemNetworkAttrSettingsURL := fmt.Sprintf("%v/Settings", oemNetworkAttrURL)
	oemNetworkAttrClearPendingURL := fmt.Sprintf("%v/Actions/DellManager.ClearPending", oemNetworkAttrSettingsURL)
	dellNetworkAttributes, err := dell.GetDellNetworkAttributes(service.GetClient(), oemNetworkAttrURL)
	if err != nil {
		diags.AddError(fmt.Sprintf("%s: error when retrieving DellNetworkAttributes", oemNetworkAttrsError), err.Error())
		return
	}

	// clear pending
	if oemAttrsState.ClearPending.ValueBool() {
		emptyPostBody := make(map[string]interface{})
		_, err := service.GetClient().Post(oemNetworkAttrClearPendingURL, emptyPostBody)
		if err != nil {
			errStr := err.Error()
			// if clear_pending failed due to no pending data to delete, just log it and continue.
			if !strings.Contains(errStr, "No pending data to delete") {
				diags.AddError(fmt.Sprintf("%s: post request for clear_pending to iDRAC failed", oemNetworkAttrsError), errStr)
				return
			}
			tflog.Warn(ctx, fmt.Sprintf("%s: oem network attributes clear_pending failed. Error: %s", oemNetworkAttrsError, errStr))
		}
	}

	// Get attributes
	attributesTf := make(map[string]string)
	if diags = oemAttrsState.Attributes.ElementsAs(ctx, &attributesTf, true); diags.HasError() {
		return
	}
	// get NetworkAttributeRegistry to check parameters before posting them to redfish
	networkAttributeRegistry, err := getNetworkAttributeRegistry(service, dellNetworkAttributes.ID)
	if err != nil {
		diags.AddError(fmt.Sprintf("%s: Could not get oem network attribute registry from iDRAC", oemNetworkAttrsError), err.Error())
		return
	}
	err = assertOemNetworkAttributes(attributesTf, networkAttributeRegistry)
	if err != nil {
		diags.AddError(fmt.Sprintf("%s: oem network attribute registry from iDRAC does not match input", oemNetworkAttrsError), err.Error())
		return
	}
	if len(attributesTf) == 0 {
		// if nothing to patch, just return
		return
	}
	// Set right attributes to patch (values from map are all string. It needs int and string)
	// re-use setManagerAttributesRightType to get right attributes for oem network attributes
	attributesToPatch, err := setManagerAttributesRightType(attributesTf, networkAttributeRegistry)
	if err != nil {
		diags.AddError(fmt.Sprintf("%s: Input oem network attributes could not be cast to the required type", oemNetworkAttrsError), err.Error())
		return
	}

	// Check that all attributes passed are compliant with the API
	err = checkManagerAttributes(networkAttributeRegistry, attributesToPatch)
	if err != nil {
		diags.AddError(fmt.Sprintf("%s: oem network attribute registry from iDRAC does not match input", oemNetworkAttrsError), err.Error())
		return
	}

	// Set the body to send
	patchBody := make(map[string]interface{})
	patchBody["Attributes"] = attributesToPatch
	patchBody[patchBodySettingsApplyTime] = map[string]interface{}{
		patchBodyApplyTime: plan.ApplyTime.ValueString(),
	}
	if strings.Contains(applyTime, "Maintenance") {
		patchBody[patchBodySettingsApplyTime] = map[string]interface{}{
			patchBodyApplyTime:                   plan.ApplyTime.ValueString(),
			"MaintenanceWindowStartTime":         plan.MaintenanceWindow.StartTime.ValueString(),
			"MaintenanceWindowDurationInSeconds": plan.MaintenanceWindow.Duration.ValueInt64(),
		}
	}

	resp, err := service.GetClient().Patch(oemNetworkAttrSettingsURL, patchBody)
	if err != nil {
		diags.AddError(fmt.Sprintf("%s: patch request to iDRAC failed", oemNetworkAttrsError), err.Error())
		return
	}
	defer resp.Body.Close() // #nosec G104

	// check if location is present in the response header
	// for all supported applytime, there should exist location and jobURL when setting oem network attributes successfully
	// for Immediate type, it will auto create one power reboot action, no need extra restart
	location, err := resp.Location()
	if err != nil {
		diags.AddError(fmt.Sprintf("%s: get location failed after patch request to iDRAC ", oemNetworkAttrsError), err.Error())
		return
	}
	return location.EscapedPath(), diags
}

func assertOemNetworkAttributes(rawAttributes map[string]string, managerAttributeRegistry *dell.ManagerAttributeRegistry) error {
	var err error
	// make map of name to ID of attributes
	attributes := make(map[string]string)
	for _, dellAttr := range managerAttributeRegistry.Attributes {
		attributes[dellAttr.AttributeName] = dellAttr.ID
	}

	// check if all input attributes are present in registry
	for k := range rawAttributes {
		_, ok := attributes[k]
		if !ok {
			err = errors.Join(err, fmt.Errorf("couldn't find oem network attribute %s", k))
			continue
		}
	}
	return err
}

func readRedfishNIC(ctx context.Context, service *gofish.Service, state *models.NICResource) diag.Diagnostics {
	var diags diag.Diagnostics

	// get networkDeviceFunction by system id, adapter id and networkDeviceFunction id
	system, networketworkDeviceFunc, err := getNetworkDeviceFunction(service, state.SystemID.ValueString(),
		state.NetworkAdapterID.ValueString(), state.NetworkDeviceFunctionID.ValueString())
	if err != nil {
		diags.AddError("Error when retrieving NetworkDeviceFunction", err.Error())
		return diags
	}

	// get OEM data
	dellDeviceFunction, _ := dell.NetworkDeviceFunction(networketworkDeviceFunc)
	if dellDeviceFunction.DellNetworkAttributes.ODataID == "" {
		diags.AddError("there was an issue when reading NIC", "error get DellNetworkAttributes ODataID from NetworkDeviceFunction Extension")
		return diags
	}
	dellNetworkAttributes, err := dell.GetDellNetworkAttributes(service.GetClient(), dellDeviceFunction.DellNetworkAttributes.ODataID)
	if err != nil {
		diags.AddError("Error when retrieving DellNetworkAttributes", err.Error())
		return diags
	}
	if diags = parseDellNetworkAttributesIntoState(ctx, dellNetworkAttributes, state); diags.HasError() {
		return diags
	}
	if diags = parseNetworkDeviceFunctionIntoState(ctx, dellDeviceFunction, state); diags.HasError() {
		return diags
	}

	state.ID = types.StringValue("redfish_network_adapter_resource")
	state.SystemID = types.StringValue(system.ID)
	return diags
}
