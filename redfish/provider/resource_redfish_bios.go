/*
Copyright (c) 2023-2024 Dell Inc., or its subsidiaries. All Rights Reserved.

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
	"fmt"
	"net/url"
	"path"
	"strconv"
	"strings"
	"terraform-provider-redfish/common"
	"terraform-provider-redfish/redfish/models"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	tfpath "github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/mapplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/stmcginnis/gofish"
	redfishcommon "github.com/stmcginnis/gofish/common"
	"github.com/stmcginnis/gofish/redfish"
)

const (
	defaultBiosConfigServerResetTimeout = 120
	defaultBiosConfigJobTimeout         = 1200
	intervalBiosConfigJobCheckTime      = 10
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource = &BiosResource{}
)

// NewBiosResource is a helper function to simplify the provider implementation.
func NewBiosResource() resource.Resource {
	return &BiosResource{}
}

// BiosResource is the resource implementation.
type BiosResource struct {
	p   *redfishProvider
	ctx context.Context
}

// Configure implements resource.ResourceWithConfigure
func (r *BiosResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	r.p = req.ProviderData.(*redfishProvider)
}

// Metadata returns the resource type name.
func (*BiosResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "bios"
}

// Schema defines the schema for the resource.
func (*BiosResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "This Terraform resource is used to configure Bios attributes of the iDRAC Server." +
			" We can Read the existing configurations or modify them using this resource.",
		Description: "This Terraform resource is used to configure Bios attributes of the iDRAC Server." +
			" We can Read the existing configurations or modify them using this resource.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The ID of the resource.",
				Description:         "The ID of the resource.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"attributes": schema.MapAttribute{
				MarkdownDescription: "The Bios attribute map.",
				Description:         "The Bios attribute map.",
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.Map{
					mapplanmodifier.UseStateForUnknown(),
				},
			},
			"settings_apply_time": schema.StringAttribute{
				Optional: true,
				Description: "The time when the BIOS settings can be applied. Applicable value is 'OnReset' only. " +
					"In upcoming releases other apply time values will be supported. Default is \"OnReset\".",
				Validators: []validator.String{
					stringvalidator.OneOf([]string{
						string(redfishcommon.OnResetApplyTime),
					}...),
				},
				Default:  stringdefault.StaticString(string(redfishcommon.OnResetApplyTime)),
				Computed: true,
			},
			"reset_type": schema.StringAttribute{
				Optional: true,
				Description: "Reset type to apply on the computer system after the BIOS settings are applied. " +
					"Applicable values are 'ForceRestart', " +
					"'GracefulRestart', and 'PowerCycle'." +
					"Default = \"GracefulRestart\". ",
				Validators: []validator.String{
					stringvalidator.OneOf([]string{
						string(redfish.ForceRestartResetType),
						string(redfish.GracefulRestartResetType),
						string(redfish.PowerCycleResetType),
					}...),
				},
				Computed: true,
				Default:  stringdefault.StaticString(string(redfish.GracefulRestartResetType)),
			},
			"reset_timeout": schema.Int64Attribute{
				Optional:    true,
				Description: "reset_timeout is the time in seconds that the provider waits for the server to be reset before timing out.",
				Default:     int64default.StaticInt64(int64(defaultBiosConfigServerResetTimeout)),
				Computed:    true,
			},
			"bios_job_timeout": schema.Int64Attribute{
				Optional: true,
				Description: "bios_job_timeout is the time in seconds that the provider waits for the bios update job to be" +
					"completed before timing out.",
				Default:  int64default.StaticInt64(int64(defaultBiosConfigJobTimeout)),
				Computed: true,
			},
			"system_id": schema.StringAttribute{
				MarkdownDescription: "System ID of the system",
				Description:         "System ID of the system",
				Computed:            true,
				Optional:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplaceIfConfigured(),
				},
			},
		},
		Blocks: RedfishServerResourceBlockMap(),
	}
}

// Create creates the resource and sets the initial Terraform state.
func (r *BiosResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	r.ctx = ctx
	tflog.Trace(ctx, "resource_Bios create : Started")
	// Get Plan Data
	var plan models.Bios
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	service, err := NewConfig(r.p, &plan.RedfishServer)
	if err != nil {
		resp.Diagnostics.AddError("service error", err.Error())
		return
	}

	state, diags := r.updateRedfishDellBiosAttributes(ctx, service, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Trace(ctx, "resource_Bios create: updating state finished, saving ...")
	// Save into State
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	tflog.Trace(ctx, "resource_Bios create: finish")
}

// Read refreshes the Terraform state with the latest data.
func (r *BiosResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	tflog.Trace(ctx, "resource_Bios read: started")
	r.ctx = ctx
	var state models.Bios
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	service, err := NewConfig(r.p, &state.RedfishServer)
	if err != nil {
		resp.Diagnostics.AddError("service error", err.Error())
		return
	}

	err = r.readRedfishDellBiosAttributes(service, &state)
	if err != nil {
		diags.AddError("Error running job", err.Error())
	}
	resp.Diagnostics.Append(diags...)

	tflog.Trace(ctx, "resource_Bios read: finished reading state")
	// Save into State
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	tflog.Trace(ctx, "resource_Bios read: finished")
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *BiosResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	r.ctx = ctx

	// Get state Data
	tflog.Trace(ctx, "resource_Bios update: started")
	var plan models.Bios

	// Get plan Data
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	service, err := NewConfig(r.p, &plan.RedfishServer)
	if err != nil {
		resp.Diagnostics.AddError("service error", err.Error())
		return
	}

	state, diags := r.updateRedfishDellBiosAttributes(ctx, service, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Trace(ctx, "resource_Bios update: finished state update")
	// Save into State
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	tflog.Trace(ctx, "resource_Bios update: finished")
}

// Delete deletes the resource and removes the Terraform state on success.
func (*BiosResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	tflog.Trace(ctx, "resource_Bios delete: started")
	// Get State Data
	var state models.Bios
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.State.RemoveResource(ctx)
	tflog.Trace(ctx, "resource_Bios delete: finished")
}

// ImportState import state for existing resource
func (*BiosResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	type creds struct {
		Username    string `json:"username"`
		Password    string `json:"password"`
		Endpoint    string `json:"endpoint"`
		SslInsecure bool   `json:"ssl_insecure"`
		SystemID    string `json:"system_id"`
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

	redfishServer := tfpath.Root("redfish_server")
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, redfishServer, []models.RedfishServer{server})...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, tfpath.Root("system_id"), types.StringValue(c.SystemID))...)
}

func (r *BiosResource) updateRedfishDellBiosAttributes(ctx context.Context, service *gofish.Service, plan *models.Bios,
) (*models.Bios, diag.Diagnostics) {
	var diags diag.Diagnostics
	state := plan

	// Lock the mutex to avoid race conditions with other resources
	redfishMutexKV.Lock(plan.RedfishServer[0].Endpoint.ValueString())
	defer redfishMutexKV.Unlock(plan.RedfishServer[0].Endpoint.ValueString())

	system, err := getSystemResource(service, state.SystemID.ValueString())
	if err != nil {
		diags.AddError("error fetching system resource", err.Error())
		return nil, diags
	}

	state.SystemID = types.StringValue(system.ID)

	bios, err := system.Bios()
	if err != nil {
		diags.AddError("error fetching bios resource", err.Error())
		return nil, diags
	}

	attributes := make(map[string]string)
	err = copyBiosAttributes(bios, attributes)
	if err != nil {
		diags.AddError("error fetching bios resource", err.Error())
		return nil, diags
	}

	attrsPayload, diagsAttr := getBiosAttrsToPatch(ctx, plan, attributes)
	diags.Append(diagsAttr...)
	if diags.HasError() {
		return nil, diags
	}

	resetTimeout := plan.ResetTimeout.ValueInt64()
	biosConfigJobTimeout := plan.JobTimeout.ValueInt64()
	resetType := plan.ResetType.ValueString()

	tflog.Debug(ctx, fmt.Sprintf("resetTimeout is set to %d and Bios Config Job timeout is set to %d", resetTimeout, biosConfigJobTimeout))

	var biosTaskURI string
	if len(attrsPayload) != 0 {
		tflog.Info(ctx, "Submitting patch request for bios attributes")
		biosTaskURI, err = r.patchBiosAttributes(plan, bios, attrsPayload)
		if err != nil {
			diags.AddError("error updating bios attributes", err.Error())
			return nil, diags
		}

		tflog.Info(ctx, "Submitting patch request for bios attributes completed successfully")
		tflog.Info(ctx, "rebooting the server")
		// reboot the server
		pOp := powerOperator{ctx, service}
		_, err := pOp.PowerOperation(plan.SystemID.ValueString(), resetType, resetTimeout, intervalBiosConfigJobCheckTime)
		if err != nil {
			// TODO: handle this scenario
			diags.AddError("there was an issue restarting the server", err.Error())
			return nil, diags
		}

		tflog.Info(ctx, "rebooting the server completed successfully")
		tflog.Info(ctx, "Waiting for the bios config job to finish")
		// wait for the bios config job to finish
		err = common.WaitForTaskToFinish(service, biosTaskURI, intervalBiosConfigJobCheckTime, biosConfigJobTimeout)
		if err != nil {
			diags.AddError("error waiting for Bios config monitor task to be completed", err.Error())
			return nil, diags
		}
		tflog.Info(ctx, "Bios config job has completed successfully")
		time.Sleep(60 * time.Second)
	} else {
		tflog.Info(ctx, "BIOS attributes are already set")
	}

	state.ID = types.StringValue(bios.ODataID)

	err = r.readRedfishDellBiosAttributes(service, state)
	if err != nil {
		diags.AddError("unable to fetch currrent bios values", err.Error())
		return nil, diags
	}

	tflog.Debug(ctx, state.ID.ValueString()+": Update finished successfully")
	return state, nil
}

func (*BiosResource) readRedfishDellBiosAttributes(service *gofish.Service, d *models.Bios) error {
	system, err := getSystemResource(service, d.SystemID.ValueString())
	if err != nil {
		return fmt.Errorf("error fetching BIOS resource: %w", err)
	}

	d.SystemID = types.StringValue(system.ID)

	bios, err := system.Bios()
	if err != nil {
		return fmt.Errorf("error fetching BIOS resource: %w", err)
	}

	attributes := make(map[string]string)
	err = copyBiosAttributes(bios, attributes)
	if err != nil {
		return fmt.Errorf("error fetching BIOS attributes: %w", err)
	}

	attributesTF := make(map[string]attr.Value)
	if !d.Attributes.IsNull() {
		old := d.Attributes.Elements()
		for key, value := range attributes {
			if _, ok := old[key]; ok {
				attributesTF[key] = types.StringValue(value)
			}
		}
	} else {
		// only in case of import
		for key, value := range attributes {
			attributesTF[key] = types.StringValue(value)
		}
	}

	d.Attributes = types.MapValueMust(types.StringType, attributesTF)
	d.ID = types.StringValue(bios.ID)
	return nil
}

func copyBiosAttributes(bios *redfish.Bios, attributes map[string]string) error {
	// TODO: BIOS Attributes' values might be any of several types.
	// terraform-sdk currently does not support a map with different
	// value types. So we will convert int and float values to string.
	// copy from the BIOS attributes to the new bios attributes map
	// for key, value := range bios.Attributes {
	for key, value := range bios.Attributes {
		if attrVal, ok := value.(string); ok {
			attributes[key] = attrVal
		} else {
			attributes[key] = fmt.Sprintf("%v", value)
		}
	}
	return nil
}

func getBiosAttrsToPatch(ctx context.Context, d *models.Bios, attributes map[string]string) (map[string]interface{}, diag.Diagnostics) {
	var diags diag.Diagnostics
	attrs := make(map[string]string)
	attrsToPatch := make(map[string]interface{})
	diags.Append(d.Attributes.ElementsAs(ctx, &attrs, true)...)

	for key, newVal := range attrs {
		oldVal, ok := attributes[key]
		if !ok {
			diags.AddError("There was an issue while creating/updating bios attriutes", fmt.Sprintf("BIOS attribute %s not found", key))
			continue
		}
		// check if the original value is an integer
		// if yes, then we need to convert accordingly
		if intOldVal, err := strconv.Atoi(attributes[key]); err == nil {
			intNewVal, err := strconv.Atoi(newVal)
			if err != nil {
				err = fmt.Errorf("BIOS attribute %s is expected to be an integer: %w", key, err)
				diags.AddError("There was an issue while creating/updating bios attriutes", err.Error())
				continue
			}

			// Add to patch list if attribute value has changed
			if intNewVal != intOldVal {
				attrsToPatch[key] = intNewVal
			}
		} else {
			if newVal != oldVal {
				attrsToPatch[key] = newVal
			}
		}
	}
	return attrsToPatch, diags
}

func (r *BiosResource) patchBiosAttributes(d *models.Bios, bios *redfish.Bios, attributes map[string]interface{}) (biosTaskURI string, err error) {
	payload := make(map[string]interface{})
	payload["Attributes"] = attributes

	settingsApplyTime := d.SettingsApplyTime.ValueString()

	allowedValues := bios.AllowedAttributeUpdateApplyTimes()
	allowed := false
	for i := range allowedValues {
		if strings.TrimSpace(settingsApplyTime) == (string)(allowedValues[i]) {
			allowed = true
			break
		}
	}

	if !allowed {
		err := fmt.Errorf("\"%s\" is not allowed as settings apply time", settingsApplyTime)
		return "", err
	}

	payload["@Redfish.SettingsApplyTime"] = map[string]interface{}{
		"ApplyTime": settingsApplyTime,
	}

	oDataURI, err := url.Parse(bios.ODataID)
	if err != nil {
		tflog.Trace(r.ctx, "error fetching data: "+err.Error())
		return "", err
	}
	oDataURI.Path = path.Join(oDataURI.Path, "Settings")
	settingsObjectURI := oDataURI.String()

	resp, err := bios.GetClient().Patch(settingsObjectURI, payload)
	if err != nil {
		tflog.Trace(r.ctx, "[DEBUG] error sending the patch request:"+err.Error())
		return "", err
	}

	// check if location is present in the response header
	if location, err := resp.Location(); err == nil {
		tflog.Trace(r.ctx, "[DEBUG] BIOS configuration job uri: "+location.String())
		taskURI := location.EscapedPath()
		return taskURI, nil
	}
	return "", nil
}
