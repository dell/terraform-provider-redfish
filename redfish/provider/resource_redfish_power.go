package provider

import (
	"context"
	"terraform-provider-redfish/redfish/models"
	"time"

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

// NewpowerResource is a helper function to simplify the provider implementation.
func NewPowerResource() resource.Resource {
	return &powerResource{}
}

// powerResource is the resource implementation.
type powerResource struct {
	p *redfishProvider
}

// Configure implements resource.ResourceWithConfigure
func (r *powerResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	r.p = req.ProviderData.(*redfishProvider)
}

// Metadata returns the resource type name.
func (r *powerResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "power"
}

func PowerSchema() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"id": schema.StringAttribute{
			MarkdownDescription: "ID of the power resource",
			Description:         "ID of the power resource",
			Computed:            true,
		},

		"redfish_server": schema.SingleNestedAttribute{
			MarkdownDescription: "Redfish Server",
			Description:         "Redfish Server",
			Required:            true,
			Attributes:          RedfishServerSchema(),
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
			Default:  int64default.StaticInt64(120),
		},

		"check_interval": schema.Int64Attribute{
			MarkdownDescription: "The frequency with which to check the server's power state in seconds",
			Description:         "The frequency with which to check the server's power state in seconds",
			Optional:            true,
			Computed:            true,
			Default:             int64default.StaticInt64(10),
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
func (r *powerResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Resource for managing power.",
		Version:             1,
		Attributes:          PowerSchema(),
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
	redfishMutexKV.Lock(plan.RedfishServer.Endpoint.ValueString())
	defer redfishMutexKV.Unlock(plan.RedfishServer.Endpoint.ValueString())

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
func (r *powerResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
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
func (r *powerResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
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

// import (
// 	"context"
// 	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
// 	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
// 	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
// 	"github.com/stmcginnis/gofish/redfish"
// 	"log"
// 	"time"
// )

// func resourceRedFishPower() *schema.Resource {
// 	return &schema.Resource{
// 		CreateContext: resourceRedfishPowerUpdate,
// 		ReadContext:   resourceRedfishPowerRead,
// 		UpdateContext: resourceRedfishPowerUpdate,
// 		DeleteContext: resourceRedfishPowerDelete,
// 		Schema:        getResourceRedfishPowerSchema(),
// 		CustomizeDiff: CheckPowerDiff(),
// 	}
// }

// const (
// 	defaultMaximumPowerConfigServerTimeout int = 120
// 	intervalPowerConfigJobCheckTime        int = 10
// )

// // Custom function for calculating power state. Given some desired_power_action, we know what the expected power
// // state should be after the action is applied. Instead of marking the value of power_state as unknown during
// // PlanResourceChange, we can calculate exactly what the end power_state should be.
// func CheckPowerDiff(funcs ...schema.CustomizeDiffFunc) schema.CustomizeDiffFunc {
// 	return func(ctx context.Context, d *schema.ResourceDiff, meta interface{}) error {

// 		resetType, ok := d.GetOk("desired_power_action")

// 		if !ok || resetType == nil {
// 			log.Printf("[ERROR]: There was a problem getting the desired_power_action")
// 			return nil
// 		}

// 		if resetType == "ForceOff" || resetType == "GracefulShutdown" {
// 			d.SetNew("power_state", "Off")
// 		} else if resetType == "ForceOn" || resetType == "On" {
// 			d.SetNew("power_state", "On")
// 		} else if resetType == "ForceRestart" || resetType == "PowerCycle" || resetType == "Nmi" {
// 			d.SetNew("power_state", "Reset_On")
// 		}
// 		// Note - if they select PushPowerButton then this function does nothing because we don't know what the value
// 		// will be. We just let Terraform set everything to unknown as per normal

// 		return nil
// 	}
// }

// func getResourceRedfishPowerSchema() map[string]*schema.Schema {
// 	return map[string]*schema.Schema{
// 		"redfish_server": {
// 			Type:        schema.TypeList,
// 			Required:    true,
// 			Description: "List of server BMCs and their respective user credentials",
// 			Elem: &schema.Resource{
// 				Schema: map[string]*schema.Schema{
// 					"user": {
// 						Type:        schema.TypeString,
// 						Optional:    true,
// 						Description: "User name for login",
// 					},
// 					"password": {
// 						Type:        schema.TypeString,
// 						Optional:    true,
// 						Description: "User password for login",
// 						Sensitive:   true,
// 					},
// 					"endpoint": {
// 						Type:        schema.TypeString,
// 						Required:    true,
// 						Description: "Server BMC IP address or hostname",
// 					},
// 					"ssl_insecure": {
// 						Type:        schema.TypeBool,
// 						Optional:    true,
// 						Description: "This field indicates whether the SSL/TLS certificate must be verified or not",
// 					},
// 				},
// 			},
// 		},
// 		"desired_power_action": {
// 			Type:     schema.TypeString,
// 			Required: true,
// 			Description: "Desired power setting. Applicable values are 'On','ForceOn','ForceOff','ForceRestart'," +
// 				"'GracefulRestart','GracefulShutdown','PowerCycle', 'PushPowerButton', 'Nmi'",
// 			ValidateFunc: validation.StringInSlice([]string{
// 				string(redfish.OnResetType),
// 				string(redfish.ForceOnResetType),
// 				string(redfish.ForceOffResetType),
// 				string(redfish.ForceRestartResetType),
// 				string(redfish.GracefulRestartResetType),
// 				string(redfish.GracefulShutdownResetType),
// 				string(redfish.PushPowerButtonResetType),
// 				string(redfish.PowerCycleResetType),
// 				string(redfish.NmiResetType),
// 			}, false),
// 		},
// 		"maximum_wait_time": {
// 			Type:     schema.TypeInt,
// 			Optional: true,
// 			Description: "The maximum amount of time to wait for the server to enter the correct power state before" +
// 				"giving up in seconds",
// 			Default: defaultMaximumPowerConfigServerTimeout,
// 		},
// 		"check_interval": {
// 			Type:        schema.TypeInt,
// 			Optional:    true,
// 			Description: "The frequency with which to check the server's power state in seconds",
// 			Default:     intervalPowerConfigJobCheckTime,
// 		},
// 		"power_state": {
// 			Type:     schema.TypeString,
// 			Optional: true,
// 			Description: "Desired power setting. Applicable values 'On','ForceOn','ForceOff','ForceRestart'," +
// 				"'GracefulRestart','GracefulShutdown','PowerCycle', 'PushPowerButton', 'Nmi'.",
// 		},
// 	}
// }

// func resourceRedfishPowerRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

// 	log.Printf("[DEBUG] %s: Beginning read", d.Id())
// 	var diags diag.Diagnostics

// 	service, err := NewConfig(m.(*schema.ResourceData), d)
// 	if err != nil {
// 		return diag.Errorf(err.Error())
// 	}

// 	system, err := getSystemResource(service)
// 	if err != nil {
// 		log.Printf("[ERROR]: Failed to identify system: %s", err)
// 		return diag.Errorf(err.Error())
// 	}

// 	if err := d.Set("power_state", system.PowerState); err != nil {
// 		return diag.Errorf("[ERROR]: Could not retrieve system power state. %s", err)
// 	}

// 	resetType, ok := d.GetOk("desired_power_action")

// 	if !ok || resetType == nil {
// 		log.Printf("[ERROR]: ")
// 		return diags
// 	}

// 	if err := d.Set("desired_power_action", resetType); err != nil {
// 		return diag.Errorf("[ERROR]: Could not retrieve system power state. %s", err)
// 	}

// 	return diags
// }

// func resourceRedfishPowerUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

// 	var diags diag.Diagnostics

// 	// Lock the mutex to avoid race conditions with other resources
// 	redfishMutexKV.Lock(getRedfishServerEndpoint(d))
// 	defer redfishMutexKV.Unlock(getRedfishServerEndpoint(d))

// 	resetType, ok := d.GetOk("desired_power_action")

// 	if !ok || resetType == nil {
// 		log.Printf("[ERROR]: There was a problem getting the desired_power_action")
// 		return diags
// 	}

// 	// Takes the m interface and feeds it the user input data d. You can then reference it with X.GetOk("user")
// 	service, err := NewConfig(m.(*schema.ResourceData), d)

// 	if err != nil {
// 		return diag.Errorf(err.Error())
// 	}

// 	system, err := getSystemResource(service)
// 	if err != nil {
// 		log.Printf("[ERROR]: Failed to identify system: %s", err)
// 		return diag.Errorf(err.Error())
// 	}

// 	d.SetId(system.SerialNumber + "_power")

// 	maxTimeout := d.Get("maximum_wait_time")

// 	checkInterval := d.Get("check_interval")

// 	powerState, diags := PowerOperation(resetType.(string), maxTimeout.(int), checkInterval.(int), service)

// 	// time to allow changes to get reflected
// 	time.Sleep(10 * time.Second)

// 	if (resetType == "ForceRestart" || resetType == "GracefulRestart" || resetType == "PowerCycle" || resetType == "Nmi") && powerState == "On" {
// 		powerState = "Reset_On"
// 	}

// 	d.Set("power_state", powerState)

// 	return diags
// }

// func resourceRedfishPowerDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

// 	var diags diag.Diagnostics

// 	d.SetId("")

// 	return diags
// }
