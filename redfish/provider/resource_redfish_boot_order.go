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
	"net/http"
	"strings"
	"terraform-provider-redfish/common"
	"terraform-provider-redfish/gofish/dell"
	"terraform-provider-redfish/redfish/models"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	tfpath "github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/stmcginnis/gofish"
	"github.com/stmcginnis/gofish/redfish"
)

const (
	defaultBootOrderResetTimeout  int64 = 120
	defaultBootOrderJobTimeout    int64 = 1200
	intervalBootOrderJobCheckTime int64 = 10
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource = &BootOrderResource{}
)

// NewBootOrderResource is a helper function to simplify the provider implementation.
func NewBootOrderResource() resource.Resource {
	return &BootOrderResource{}
}

// BootOrderResource is the resource implementation.
type BootOrderResource struct {
	p   *redfishProvider
	ctx context.Context
}

// Schema implements resource.Resource.
func (*BootOrderResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "This Terraform resource is used to configure Boot Order and enable/disable Boot Options of the iDRAC Server." +
			" We can Read the existing configurations or modify them using this resource.",
		Description: "This Terraform resource is used to configure Boot Order and enable/disable Boot Options of the iDRAC Server." +
			" We can Read the existing configurations or modify them using this resource.",
		Attributes: BootOrderSchema(),
		Blocks:     RedfishServerResourceBlockMap(),
	}
}

// BootOrderSchema to define the Boot Order resource schema
func BootOrderSchema() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"id": schema.StringAttribute{
			MarkdownDescription: "ID of the Boot Order Resource",
			Description:         "ID of the Boot Order Resource",
			Computed:            true,
		},
		"boot_options": schema.ListNestedAttribute{
			Description:         "Options to enable or disable the boot device.",
			MarkdownDescription: "Options to enable or disable the boot device.",
			Optional:            true,
			Computed:            true,
			NestedObject: schema.NestedAttributeObject{
				Attributes: map[string]schema.Attribute{
					"boot_option_reference": schema.StringAttribute{
						Description:         "FQDD of the boot device.",
						MarkdownDescription: "FQDD of the boot device.",
						Optional:            true,
						Computed:            true,
					},
					"boot_option_enabled": schema.BoolAttribute{
						Description:         "Enable or disable the boot device.",
						MarkdownDescription: "Enable or disable the boot device.",
						Required:            true,
					},
				},
			},
		},
		"boot_order": schema.ListAttribute{
			MarkdownDescription: "sets the boot devices in the required boot order sequences.",
			Description:         "sets the boot devices in the required boot order sequences.",
			Computed:            true,
			Optional:            true,
			ElementType:         types.StringType,
			Validators: []validator.List{
				listvalidator.ConflictsWith(tfpath.Expressions{
					tfpath.MatchRoot("boot_options"),
				}...),
				listvalidator.AtLeastOneOf(tfpath.Expressions{
					tfpath.MatchRoot("boot_options"),
					tfpath.MatchRoot("boot_order"),
				}...),
			},
		},
		"reset_type": schema.StringAttribute{
			Required: true,
			Description: "Reset type allows to choose the type of restart to apply when firmware upgrade is scheduled." +
				" Possible values are: \"ForceRestart\", \"GracefulRestart\" or \"PowerCycle\"",
			MarkdownDescription: "Reset type allows to choose the type of restart to apply when firmware upgrade is scheduled." +
				" Possible values are: \"ForceRestart\", \"GracefulRestart\" or \"PowerCycle\"",
			Validators: []validator.String{
				stringvalidator.OneOf([]string{
					string(redfish.ForceRestartResetType),
					string(redfish.GracefulRestartResetType),
					string(redfish.PowerCycleResetType),
				}...),
			},
		},
		"reset_timeout": schema.Int64Attribute{
			Optional:            true,
			Computed:            true,
			Default:             int64default.StaticInt64(int64(defaultBootOrderResetTimeout)),
			Description:         "Time in seconds that the provider waits for the server to be reset before timing out.",
			MarkdownDescription: "Time in seconds that the provider waits for the server to be reset before timing out.",
		},
		"boot_order_job_timeout": schema.Int64Attribute{
			Optional:            true,
			Computed:            true,
			Default:             int64default.StaticInt64(int64(defaultBootOrderJobTimeout)),
			Description:         "Time in seconds that the provider waits for the simple update job to be completed before timing out.",
			MarkdownDescription: "Time in seconds that the provider waits for the BootSource override job to be completed before timing out.",
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
	}
}

// Configure implements resource.ResourceWithConfigure
func (r *BootOrderResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	r.p = req.ProviderData.(*redfishProvider)
}

// Metadata returns the resource type name.
func (*BootOrderResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "boot_order"
}

// Create implements resource.Resource.
func (r *BootOrderResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	r.ctx = ctx
	tflog.Trace(ctx, "resource_Bios create : Started")
	var diags diag.Diagnostics

	// Get Plan Data
	var plan models.BootOrder
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

	diags = r.bootOperation(ctx, service, &plan)
	resp.Diagnostics.Append(diags...)
	if diags.HasError() {
		return
	}

	state, diags := r.updateServer(service, plan)
	if diags.HasError() {
		resp.Diagnostics.AddError("Update server failed", "")
		return
	}

	tflog.Trace(ctx, "resource_Bios create: updating state finished, saving ...")
	// Save into State
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	tflog.Trace(ctx, "resource_Bios create: finish")
}

// Update implements resource.Resource.
func (r *BootOrderResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	r.ctx = ctx
	tflog.Trace(ctx, "resource_boot_order update : Started")
	// Get Plan Data
	var plan models.BootOrder
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

	diags = r.bootOperation(ctx, service, &plan)
	resp.Diagnostics.Append(diags...)
	if diags.HasError() {
		return
	}

	state, diags := r.updateServer(service, plan)
	if diags.HasError() {
		resp.Diagnostics.AddError("Update server failed", "")
		return
	}

	tflog.Trace(ctx, "resource_Boot_order update: updating state finished, saving ...")
	// Save into State
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	tflog.Trace(ctx, "resource_Boot_order update: finish")
}

// Delete implements resource.Resource.
func (*BootOrderResource) Delete(ctx context.Context, _ resource.DeleteRequest, resp *resource.DeleteResponse) {
	resp.State.RemoveResource(ctx)
}

// Read implements resource.Resource.
func (r *BootOrderResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	tflog.Trace(ctx, "resource_simple_update read : Started")
	// Get Plan Data
	var newState, state models.BootOrder
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

	system, err := getSystemResource(service, state.SystemID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("[ERROR]: Failed to get updated system resource", err.Error())
		return
	}

	diags = r.readRedfishBootAttributes(system, &newState, &state)
	if diags.HasError() {
		return
	}

	// Save into State
	diags = resp.State.Set(ctx, &newState)
	resp.Diagnostics.Append(diags...)
	tflog.Trace(ctx, "resource_Bios read: finished")
}

// ImportState implements Import functionality for Boot Order Resource
func (*BootOrderResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	type creds struct {
		Username     string `json:"username"`
		Password     string `json:"password"`
		Endpoint     string `json:"endpoint"`
		SslInsecure  bool   `json:"ssl_insecure"`
		SystemID     string `json:"system_id"`
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

	redfishServer := tfpath.Root("redfish_server")
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, redfishServer, []models.RedfishServer{server})...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, tfpath.Root("system_id"), types.StringValue(c.SystemID))...)
}

func (r *BootOrderResource) bootOperation(ctx context.Context, service *gofish.Service, plan *models.BootOrder) diag.Diagnostics {
	// Lock the mutex to avoid race conditions with other resources
	redfishMutexKV.Lock(plan.RedfishServer[0].Endpoint.ValueString())
	defer redfishMutexKV.Unlock(plan.RedfishServer[0].Endpoint.ValueString())

	resp, diags := r.updateRedfishDellBootAttributes(service, plan)
	if diags.HasError() {
		return diags
	}
	diags.Append(r.restartServer(ctx, service, resp, plan)...)
	return diags
}

func (r *BootOrderResource) updateRedfishDellBootAttributes(service *gofish.Service, d *models.BootOrder) (*http.Response, diag.Diagnostics) {
	var resp *http.Response
	var diags diag.Diagnostics
	var err error

	if len(d.BootOptions.Elements()) > 0 {
		resp, diags = r.updateBootOptions(service, d)
		return resp, diags
	}
	resp, err = r.setBootOrder(service, d)
	if err != nil {
		diags.AddError("Boot Operation Failed", err.Error())
	}
	return resp, diags
}

func (r *BootOrderResource) readRedfishBootAttributes(system *redfish.ComputerSystem, d,
	plan *models.BootOrder,
) diag.Diagnostics {
	var diags diag.Diagnostics

	boot := system.Boot
	bootOrder := []attr.Value{}
	for _, bootOptionReference := range boot.BootOrder {
		bootOrder = append(bootOrder, types.StringValue(string(bootOptionReference)))
	}

	d.BootOrder, diags = types.ListValue(types.StringType, bootOrder)
	if diags.HasError() {
		return diags
	}

	d.ID = types.StringValue(system.ODataID)
	d.RedfishServer = plan.RedfishServer
	if plan.JobTimeout.ValueInt64() > 0 {
		d.JobTimeout = plan.JobTimeout
	} else {
		d.JobTimeout = types.Int64Value(defaultBootOrderJobTimeout)
	}
	if plan.ResetTimeout.ValueInt64() > 0 {
		d.ResetTimeout = plan.ResetTimeout
	} else {
		d.ResetTimeout = types.Int64Value(defaultBootOrderResetTimeout)
	}
	d.ResetType = plan.ResetType
	stateval, diags := r.getUpdatedBootOptions(system, plan)
	d.BootOptions = stateval
	d.SystemID = types.StringValue(system.ID)
	return diags
}

func (r *BootOrderResource) getUpdatedBootOptions(system *redfish.ComputerSystem, plan *models.BootOrder) (types.List, diag.Diagnostics) {
	var diags diag.Diagnostics

	responseBootOptions, err := system.BootOptions()
	if err != nil {
		diags.AddError("Cannot read Boot Options", err.Error())
	}

	bootOptionsTypes := map[string]attr.Type{
		"boot_option_reference": types.StringType,
		"boot_option_enabled":   types.BoolType,
	}

	bootOptionsEleType := types.ObjectType{
		AttrTypes: bootOptionsTypes,
	}

	objectBootOptions := []attr.Value{}

	if len(plan.BootOptions.Elements()) > 0 || len(plan.BootOrder.Elements()) > 0 {
		// if there are no elements in plan then CRUD returns empty list
		if len(plan.BootOptions.Elements()) == 0 {
			return types.ListNull(bootOptionsEleType), diags
		}

		// to avoid state and plan mismatch error fetch only boot options that are a part of plan
		newPlan, diags := r.getBootOptionsList(plan)
		if diags.HasError() {
			return types.ListNull(bootOptionsEleType), diags
		}
		for _, rbp := range responseBootOptions {
			toBeAdded := false
			for _, planObject := range newPlan {
				if rbp.BootOptionReference == planObject.BootOptionReference.ValueString() {
					toBeAdded = true
					break
				}
			}
			if !toBeAdded {
				continue
			}
			objVal, diags := getUpdatedValues(rbp, bootOptionsTypes)
			objectBootOptions = append(objectBootOptions, objVal)
			if diags.HasError() {
				return types.ListNull(bootOptionsEleType), diags
			}
		}
	} else { //  fetch all boot options in case of import
		for _, rbp := range responseBootOptions {
			objVal, diags := getUpdatedValues(rbp, bootOptionsTypes)
			objectBootOptions = append(objectBootOptions, objVal)
			if diags.HasError() {
				return types.ListNull(bootOptionsEleType), diags
			}
		}
	}
	diags.Append(diags...)
	if diags.HasError() {
		return types.ListNull(bootOptionsEleType), diags
	}

	stateVal, diags := types.ListValue(bootOptionsEleType, objectBootOptions)
	return stateVal, diags
}

func getUpdatedValues(responseBootOption *redfish.BootOption, bootOptionsTypes map[string]attr.Type) (attr.Value, diag.Diagnostics) {
	var diags diag.Diagnostics

	obj := map[string]attr.Value{
		"boot_option_reference": types.StringValue(responseBootOption.BootOptionReference),
		"boot_option_enabled":   types.BoolValue(responseBootOption.BootOptionEnabled),
	}
	objVal, diags := types.ObjectValue(bootOptionsTypes, obj)

	return objVal, diags
}

func (r *BootOrderResource) updateBootOptions(service *gofish.Service, d *models.BootOrder) (*http.Response, diag.Diagnostics) {
	var url string
	var diags diag.Diagnostics

	system, err := getSystemResource(service, d.SystemID.ValueString())
	if err != nil {
		diags.AddError("[ERROR]: Failed to get system resource", err.Error())
		return nil, diags
	}

	type Payload struct {
		BootOptionEnabled bool `json:"BootOptionEnabled"`
	}

	var payload Payload
	var resp *http.Response

	bootOptions, err := system.BootOptions()
	if err != nil {
		diags.AddError("unable to fetch boot Options", err.Error())
		return nil, diags
	}

	if len(bootOptions) == 0 {
		diags.AddError("unable to fetch boot Options Boot Options are not specified", "")
		return nil, diags
	}
	url = bootOptions[0].Entity.ODataID
	lastIndx := strings.LastIndex(url, "/")
	url = url[:lastIndx]

	serverBootOptions, diags := r.getBootOptionsList(d)
	if diags.HasError() {
		return nil, diags
	}
	for _, ele := range serverBootOptions {
		payload.BootOptionEnabled = ele.BootOptionEnabled.ValueBool()
		finalURL := fmt.Sprintf("%s/%s", url, ele.BootOptionReference.ValueString())
		resp, err = service.GetClient().Patch(finalURL, payload)
		if err != nil {
			diags.AddError("Unable to update boot option data", err.Error())
			return nil, diags
		}
	}
	return resp, nil
}

func (*BootOrderResource) setBootOrder(service *gofish.Service, d *models.BootOrder) (*http.Response, error) {
	var resp *http.Response
	var uri string
	system, err := getSystemResource(service, d.SystemID.ValueString())
	if err != nil {
		return nil, fmt.Errorf("[ERROR]: Failed to get system resource %w", err)
	}

	boot := system.Boot
	// get existing boot order
	existingBootOrder := boot.BootOrder
	newBootOrder := d.BootOrder.Elements()

	// compare two boot orders
	if len(newBootOrder) > 0 {
		for _, d := range newBootOrder {
			flag := false
			for _, val := range existingBootOrder {
				if strings.Trim(d.String(), "\"") == val {
					flag = true
				}
			}
			if !flag {
				return nil, fmt.Errorf("new boot order and old boot order must be equal")
			}
		}
		// check if all boot devices are present
		if len(newBootOrder) != len(existingBootOrder) {
			return nil, fmt.Errorf("unable to complete the operation because all boot devices are required for this operation")
		}
	}

	type Boot struct {
		BootOrder []string
	}
	type Payload struct {
		Boot Boot `json:"Boot"`
	}
	var payload Payload
	if len(newBootOrder) > 0 {
		for _, d := range newBootOrder {
			payload.Boot.BootOrder = append(
				payload.Boot.BootOrder, strings.Trim(d.String(), "\""),
			)
		}
	}
	isGenerationSeventeenAndAbove, err := isServerGenerationSeventeenAndAbove(service)
	if err != nil {
		return nil, fmt.Errorf("Error retrieving the server generation %w", err)
	}
	// for 17G use system settings api for PATCH call
	if isGenerationSeventeenAndAbove {
		res, err := dell.ComputerSystems(system)
		uri = res.Settings.OdataID
		if err != nil {
			return nil, fmt.Errorf("Error retrieving the systems settings URI %w", err)
		}
	} else {
		// Below 17G will have System API for PATCH call
		uri = system.ODataID
	}

	resp, err = service.GetClient().Patch(uri, payload)
	if err != nil {
		return resp, fmt.Errorf("cannot update boot order %w", err)
	}
	return resp, nil
}

func (r *BootOrderResource) updateServer(service *gofish.Service, plan models.BootOrder) (*models.BootOrder, diag.Diagnostics) {
	var diags diag.Diagnostics
	// Fetch Updated details
	system, err := getSystemResource(service, plan.SystemID.ValueString())
	if err != nil {
		diags.AddError("[ERROR]: Failed to get updated system resource", err.Error())
		return nil, diags
	}

	state := models.BootOrder{}
	diags = r.readRedfishBootAttributes(system, &state, &plan)
	if err != nil {
		diags.AddError("State update failed", err.Error())
		return nil, diags
	}
	return &state, diags
}

func (*BootOrderResource) restartServer(ctx context.Context, service *gofish.Service, resp *http.Response, plan *models.BootOrder) diag.Diagnostics {
	// Power Operation parameters
	var diags diag.Diagnostics
	resetType := plan.ResetType.ValueString()
	resetTimeout := plan.ResetTimeout.ValueInt64()
	bootOrderJobTimeout := plan.JobTimeout.ValueInt64()

	jobID := resp.Header.Get("Location")
	if jobID == "" {
		diags.AddWarning("this configuration is already set ", "Update the configuration and run again")
		return diags
	}
	// Below 17G device returns location as /redfish/v1/TaskService/Tasks/JOB_ID for same GET call return status as 200 with all the job status.
	// where as 17G device returns location as /redfish/v1/TaskService/TaskMonitors/JOB_ID for same GET call return no content hence
	// we are replacing TaskMonitors to Tasks.
	jobID = strings.Replace(jobID, "TaskMonitors", "Tasks", 1)

	// reboot the server
	pOp := powerOperator{ctx, service, plan.SystemID.ValueString()}
	_, err := pOp.PowerOperation(resetType, resetTimeout, intervalBootOrderJobCheckTime)
	if err != nil {
		diags.AddError("there was an issue restarting the server ", err.Error())
		return diags
	}
	if jobID != "" {
		// wait for the bios config job to finish
		if strings.Contains(jobID, "Job") {
			err = common.WaitForJobToFinish(service, jobID, intervalBootOrderJobCheckTime, bootOrderJobTimeout)
		} else {
			err = common.WaitForTaskToFinish(service, jobID, intervalBootOrderJobCheckTime, bootOrderJobTimeout)
		}
		if err != nil {
			diags.AddError("error waiting for Bios config monitor task to be completed", err.Error())
			return diags
		}
	}
	time.Sleep(180 * time.Second)
	return nil
}

// getBootOptionsList converts list of Boot Options from tf model to go type
func (r *BootOrderResource) getBootOptionsList(d *models.BootOrder) ([]models.BootOptions, diag.Diagnostics) {
	var diags diag.Diagnostics
	bootList := make([]models.BootOptions, 0)
	diags.Append(d.BootOptions.ElementsAs(r.ctx, &bootList, false)...)
	return bootList, diags
}
