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
	"net/http"
	"terraform-provider-redfish/common"
	"terraform-provider-redfish/redfish/models"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/stmcginnis/gofish"
	"github.com/stmcginnis/gofish/redfish"
)

const (
	defaultBootSourceOverrideResetTimeout  int   = 120
	defaultBootSourceOverrideJobTimeout    int   = 1200
	intervalBootSourceOverrideJobCheckTime int64 = 10
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource = &BootSourceOverrideResource{}
)

// NewBootSourceOverrideResource is a helper function to simplify the provider implementation.
func NewBootSourceOverrideResource() resource.Resource {
	return &BootSourceOverrideResource{}
}

// BootSourceOverrideResource is the resource implementation.
type BootSourceOverrideResource struct {
	p   *redfishProvider
	ctx context.Context
}

// Schema implements resource.Resource.
func (*BootSourceOverrideResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "This Terraform resource is used to configure Boot sources of the iDRAC Server.",
		Description:         "This Terraform resource is used to configure Boot sources of the iDRAC Server.",
		Attributes:          BootSourceOverrideSchema(),
		Blocks: map[string]schema.Block{
			"redfish_server": schema.ListNestedBlock{
				MarkdownDescription: "List of server BMCs and their respective user credentials",
				Description:         "List of server BMCs and their respective user credentials",
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
					listvalidator.IsRequired(),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: RedfishServerSchema(),
				},
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

// BootSourceOverrideSchema to define the Boot Source Override resource schema
func BootSourceOverrideSchema() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"id": schema.StringAttribute{
			MarkdownDescription: "ID of the Boot Source Override Resource",
			Description:         "ID of the Boot Source Override Resource",
			Computed:            true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"boot_source_override_mode": schema.StringAttribute{
			MarkdownDescription: "The BIOS boot mode to be used when boot source is booted from.",
			Description:         "The BIOS boot mode to be used when boot source is booted from.",
			Optional:            true,
			Validators: []validator.String{
				stringvalidator.OneOf([]string{
					string(redfish.LegacyBootSourceOverrideMode),
					string(redfish.UEFIBootSourceOverrideMode),
				}...),
			},
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplaceIfConfigured(),
			},
		},
		"boot_source_override_enabled": schema.StringAttribute{
			MarkdownDescription: "The state of the Boot Source Override feature.",
			Description:         "The state of the Boot Source Override feature.",
			Optional:            true,
			Validators: []validator.String{
				stringvalidator.OneOf([]string{
					string(redfish.ContinuousBootSourceOverrideEnabled),
					string(redfish.DisabledBootSourceOverrideEnabled),
					string(redfish.OnceBootSourceOverrideEnabled),
				}...),
			},
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplaceIfConfigured(),
			},
		},
		"boot_source_override_target": schema.StringAttribute{
			MarkdownDescription: "The boot source override target device to use during the next boot instead of the normal boot device.",
			Description:         "The boot source override target device to use during the next boot instead of the normal boot device.",
			Optional:            true,
			Validators: []validator.String{
				stringvalidator.OneOf([]string{
					string(redfish.PxeBootSourceOverrideTarget),
					string(redfish.BiosSetupBootSourceOverrideTarget),
					string(redfish.NoneBootSourceOverrideTarget),
					string(redfish.FloppyBootSourceOverrideTarget),
					string(redfish.CdBootSourceOverrideTarget),
					string(redfish.UsbBootSourceOverrideTarget),
					string(redfish.HddBootSourceOverrideTarget),
					string(redfish.UtilitiesBootSourceOverrideTarget),
					string(redfish.DiagsBootSourceOverrideTarget),
					string(redfish.UefiShellBootSourceOverrideTarget),
					string(redfish.UefiTargetBootSourceOverrideTarget),
					string(redfish.SDCardBootSourceOverrideTarget),
					string(redfish.UefiHTTPBootSourceOverrideTarget),
					string(redfish.RemoteDriveBootSourceOverrideTarget),
					string(redfish.UefiBootNextBootSourceOverrideTarget),
				}...),
			},
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplaceIfConfigured(),
			},
		},
		"uefi_target_boot_source_override": schema.StringAttribute{
			MarkdownDescription: "The UEFI device path of the device from which to boot when boot_source_override_target is UefiTarget",
			Description:         "The UEFI device path of the device from which to boot when boot_source_override_target is UefiTarget",
			Optional:            true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplaceIfConfigured(),
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
			Default:             int64default.StaticInt64(int64(defaultBootSourceOverrideResetTimeout)),
			Description:         "Time in seconds that the provider waits for the server to be reset before timing out.",
			MarkdownDescription: "Time in seconds that the provider waits for the server to be reset before timing out.",
		},
		"boot_source_job_timeout": schema.Int64Attribute{
			Optional:            true,
			Computed:            true,
			Default:             int64default.StaticInt64(int64(defaultBootSourceOverrideJobTimeout)),
			Description:         "Time in seconds that the provider waits for the BootSource override job to be completed before timing out.",
			MarkdownDescription: "Time in seconds that the provider waits for the BootSource override job to be completed before timing out.",
		},
	}
}

// Configure implements resource.ResourceWithConfigure
func (r *BootSourceOverrideResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	r.p = req.ProviderData.(*redfishProvider)
}

// Metadata returns the resource type name.
func (*BootSourceOverrideResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "boot_source_override"
}

// Create implements cration of boot override resource.
func (r *BootSourceOverrideResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	r.ctx = ctx
	tflog.Trace(ctx, "resource_Bios create : Started")
	var diags diag.Diagnostics

	// Get Plan Data
	var plan models.BootSourceOverride
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

	plan.ID = types.StringValue("boot_sources")

	tflog.Trace(ctx, "resource_Bios create: updating state finished, saving ...")
	// Save into State
	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	tflog.Trace(ctx, "resource_Bios create: finish")
}

// Update implements resource.Resource.
func (*BootSourceOverrideResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	tflog.Trace(ctx, "resource_Boot_source update: updating state finished, saving ...")
	// Get Plan Data
	var plan, state models.BootSourceOverride
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	state.JobTimeout = plan.JobTimeout
	state.ResetTimeout = plan.ResetTimeout
	state.ResetType = plan.ResetType

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	tflog.Trace(ctx, "resource_Boot_source update: finish")
}

// Delete implements resource.Resource.
func (*BootSourceOverrideResource) Delete(ctx context.Context, _ resource.DeleteRequest, resp *resource.DeleteResponse) {
	resp.State.RemoveResource(ctx)
}

// Read implements resource.Resource.
func (*BootSourceOverrideResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	tflog.Trace(ctx, "resource_boot_source read : Started")
	// Get Plan Data
	var state models.BootSourceOverride
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	tflog.Trace(ctx, "resource_Boot_source Read: finish")
}

func (r *BootSourceOverrideResource) bootOperation(ctx context.Context, service *gofish.Service, plan *models.BootSourceOverride) diag.Diagnostics {
	// Lock the mutex to avoid race conditions with other resources
	redfishMutexKV.Lock(plan.RedfishServer[0].Endpoint.ValueString())
	defer redfishMutexKV.Unlock(plan.RedfishServer[0].Endpoint.ValueString())

	var resp *http.Response
	var diags diag.Diagnostics

	system, err := getSystemResource(service)
	if err != nil {
		diags.AddError("[ERROR]: Failed to get system resource", err.Error())
		return diags
	}

	type Boot struct {
		BootSourceOverrideMode    redfish.BootSourceOverrideMode
		BootSourceOverrideEnabled redfish.BootSourceOverrideEnabled
		BootSourceOverrideTarget  redfish.BootSourceOverrideTarget
	}
	type Payload struct {
		Boot Boot `json:"Boot"`
	}

	var payload Payload
	payload.Boot.BootSourceOverrideMode = redfish.BootSourceOverrideMode(plan.BootSourceOverrideMode.ValueString())
	payload.Boot.BootSourceOverrideEnabled = redfish.BootSourceOverrideEnabled(plan.BootSourceOverrideEnabled.ValueString())
	payload.Boot.BootSourceOverrideTarget = redfish.BootSourceOverrideTarget(plan.BootSourceOverrideTarget.ValueString())

	resp, err = service.GetClient().Patch(system.ODataID, payload)
	if err != nil {
		diags.AddError("Cannot update boot override details ", err.Error())
		return diags
	}

	diags.Append(r.restartServer(ctx, service, resp, plan)...)
	return diags
}

func (*BootSourceOverrideResource) restartServer(ctx context.Context, service *gofish.Service,
	resp *http.Response, plan *models.BootSourceOverride,
) diag.Diagnostics {
	// Power Operation parameters
	var diags diag.Diagnostics
	resetType := plan.ResetType.ValueString()
	resetTimeout := plan.ResetTimeout.ValueInt64()
	bootSourceOverrideJobTimeout := plan.JobTimeout.ValueInt64()

	// reboot the server
	pOp := powerOperator{ctx, service}
	_, err := pOp.PowerOperation(resetType, resetTimeout, intervalBootSourceOverrideJobCheckTime)
	if err != nil {
		diags.AddError("there was an issue restarting the server ", err.Error())
		return diags
	}

	jobID := resp.Header.Get("Location")
	// wait for the bios config job to finish
	err = common.WaitForJobToFinish(service, jobID, intervalBootSourceOverrideJobCheckTime, bootSourceOverrideJobTimeout)
	if err != nil {
		diags.AddError("error waiting for Bios config monitor task to be completed", err.Error())
		return diags
	}
	time.Sleep(60 * time.Second)
	return nil
}
