package provider

import (
	"context"
	"terraform-provider-redfish/redfish/models"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/stmcginnis/gofish/redfish"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource = &powerResource{}
)

// NewPowerResource is a helper function to simplify the provider implementation.
func NewPowerResource() resource.Resource {
	return &powerResource{}
}

// powerResource is the resource implementation.
type powerResource struct {
	p *redfishProvider
}

// Configure implements resource.ResourceWithConfigure
func (r *powerResource) Configure(ctx context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	r.p = req.ProviderData.(*redfishProvider)
	tflog.Trace(ctx, "resource_power configured")
}

// Metadata returns the resource type name.
func (*powerResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "power"
}

// PowerSchema to design the schema for power resource.
func PowerSchema() map[string]schema.Attribute {
	const waitTime = 120
	const checkInterval = 10
	return map[string]schema.Attribute{
		"id": schema.StringAttribute{
			MarkdownDescription: "ID of the power resource",
			Description:         "ID of the power resource",
			Computed:            true,
		},
		"desired_power_action": schema.StringAttribute{
			MarkdownDescription: "Desired power setting. Applicable values are 'On','ForceOn','ForceOff','ForceRestart'," +
				"'GracefulRestart','GracefulShutdown','PowerCycle', 'PushPowerButton', 'Nmi'",
			Description: "Desired power setting. Applicable values are 'On','ForceOn','ForceOff','ForceRestart'," +
				"'GracefulRestart','GracefulShutdown','PowerCycle', 'PushPowerButton', 'Nmi'",
			Required: true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplaceIfConfigured(),
			},
			Validators: []validator.String{
				stringvalidator.OneOf(
					string(redfish.OnResetType),
					string(redfish.ForceOnResetType),
					string(redfish.ForceOffResetType),
					string(redfish.ForceRestartResetType),
					string(redfish.GracefulRestartResetType),
					string(redfish.GracefulShutdownResetType),
					string(redfish.PushPowerButtonResetType),
					string(redfish.PowerCycleResetType),
					string(redfish.NmiResetType),
				),
			},
		},

		"maximum_wait_time": schema.Int64Attribute{
			MarkdownDescription: "The maximum amount of time to wait for the server to enter the correct power state before" +
				"giving up in seconds",
			Description: "The maximum amount of time to wait for the server to enter the correct power state before" +
				"giving up in seconds",
			Optional: true,
			Computed: true,
			Default:  int64default.StaticInt64(waitTime),
		},

		"check_interval": schema.Int64Attribute{
			MarkdownDescription: "The frequency with which to check the server's power state in seconds",
			Description:         "The frequency with which to check the server's power state in seconds",
			Optional:            true,
			Computed:            true,
			Default:             int64default.StaticInt64(checkInterval),
		},

		"power_state": schema.StringAttribute{
			MarkdownDescription: "Desired power setting. Applicable values 'On','ForceOn','ForceOff','ForceRestart'," +
				"'GracefulRestart','GracefulShutdown','PowerCycle', 'PushPowerButton', 'Nmi'.",
			Description: "Desired power setting. Applicable values 'On','ForceOn','ForceOff','ForceRestart'," +
				"'GracefulRestart','GracefulShutdown','PowerCycle', 'PushPowerButton', 'Nmi'.",
			Computed: true,
		},
	}
}

// Schema defines the schema for the resource.
func (*powerResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Resource for managing power.",
		Version:             1,
		Attributes:          PowerSchema(),
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
			},
		},
	}
}

// Create creates the resource and sets the initial Terraform state.
func (r *powerResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	tflog.Trace(ctx, "resource_power create : Started")
	// Get Plan Data
	var plan models.Power
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	// 	// Lock the mutex to avoid race conditions with other resources
	redfishMutexKV.Lock(plan.RedfishServer[0].Endpoint.ValueString())
	defer redfishMutexKV.Unlock(plan.RedfishServer[0].Endpoint.ValueString())

	service, err := NewConfig(r.p, &plan.RedfishServer)
	if err != nil {
		resp.Diagnostics.AddError("service error", err.Error())
		return
	}
	system, err := getSystemResource(service)
	if err != nil {
		resp.Diagnostics.AddError("system error", err.Error())
		return
	}

	plan.PowerId = types.StringValue(system.SerialNumber + "_power")

	resetType := plan.DesiredPowerAction.ValueString()
	pOp := powerOperator{ctx, service}
	powerState, pErr := pOp.PowerOperation(resetType, plan.MaximumWaitTime.ValueInt64(), plan.CheckInterval.ValueInt64())
	if pErr != nil {
		return
	}
	// time to allow changes to get reflected
	time.Sleep(10 * time.Second)

	if (resetType == "ForceRestart" || resetType == "GracefulRestart" || resetType == "PowerCycle" || resetType == "Nmi") && powerState == "On" {
		powerState = "Reset_On"
	}

	plan.PowerState = types.StringValue(string(powerState))

	tflog.Trace(ctx, "resource_power create: updating state finished, saving ...")
	// Save into State
	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	tflog.Trace(ctx, "resource_power create: finish")
}

// Read refreshes the Terraform state with the latest data.
func (r *powerResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	tflog.Trace(ctx, "resource_power read: started")
	var state models.Power
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

	system, err := getSystemResource(service)
	if err != nil {
		resp.Diagnostics.AddError("system error", err.Error())
		return
	}

	state.PowerState = types.StringValue(string(system.PowerState))

	tflog.Trace(ctx, "resource_power read: finished reading state")
	// Save into State
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	tflog.Trace(ctx, "resource_power read: finished")
}

// Update updates the resource and sets the updated Terraform state on success.
func (*powerResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Get state Data
	tflog.Trace(ctx, "resource_power update: started")
	var state, plan models.Power
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

	state.MaximumWaitTime = plan.MaximumWaitTime
	state.CheckInterval = plan.CheckInterval
	state.RedfishServer = plan.RedfishServer

	tflog.Trace(ctx, "resource_power update: finished state update")
	// Save into State
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	tflog.Trace(ctx, "resource_power update: finished")
}

// Delete deletes the resource and removes the Terraform state on success.
func (*powerResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	tflog.Trace(ctx, "resource_power delete: started")
	// Get State Data
	var state models.Power
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.State.RemoveResource(ctx)
	tflog.Trace(ctx, "resource_power delete: finished")
}
