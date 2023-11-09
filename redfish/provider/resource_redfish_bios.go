package provider

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"path"
	"strconv"
	"strings"
	"terraform-provider-redfish/common"
	"terraform-provider-redfish/redfish/models"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"

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
	p *redfishProvider
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
		MarkdownDescription: "This Terraform resource is used to manage user entity of the iDRAC Server." +
			"We can create, read, modify and delete an existing user using this resource.",
		Description: "This Terraform resource is used to manage user entity of the iDRAC Server." +
			"We can create, read, modify and delete an existing user using this resource.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The ID of the resource.",
				Description:         "The ID of the resource.",
				Computed:            true,
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
		},
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
func (r *BiosResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
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
	if err != nil {
		diags.AddError("Error running job %w", err.Error())
	}
	resp.Diagnostics.Append(diags...)

	tflog.Trace(ctx, "resource_Bios create: updating state finished, saving ...")
	// Save into State
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	tflog.Trace(ctx, "resource_Bios create: finish")
}

// Read refreshes the Terraform state with the latest data.
func (r *BiosResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	tflog.Trace(ctx, "resource_Bios read: started")
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
		diags.AddError("Error running job %w", err.Error())
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
	if err != nil {
		diags.AddError("Error running job %v", err.Error())
	}
	resp.Diagnostics.Append(diags...)

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

func (r *BiosResource) updateRedfishDellBiosAttributes(ctx context.Context, service *gofish.Service, plan *models.Bios,
) (*models.Bios, diag.Diagnostics) {
	var diags diag.Diagnostics
	state := plan

	// Lock the mutex to avoid race conditions with other resources
	redfishMutexKV.Lock(plan.RedfishServer[0].Endpoint.ValueString())
	defer redfishMutexKV.Unlock(plan.RedfishServer[0].Endpoint.ValueString())

	bios, err := getBiosResource(service)
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

	attrsPayload, diags := getBiosAttrsToPatch(ctx, plan, attributes)
	if err != nil {
		diags.AddError("error getting BIOS attributes to patch", err.Error())
		return nil, diags
	}

	resetTimeout := plan.ResetTimeout.ValueInt64()
	biosConfigJobTimeout := plan.JobTimeout.ValueInt64()
	resetType := plan.ResetType.ValueString()

	log.Printf("[DEBUG] resetTimeout is set to %d  and Bios Config Job timeout is set to %d", resetTimeout, biosConfigJobTimeout)

	var biosTaskURI string
	if len(attrsPayload) != 0 {
		biosTaskURI, err = patchBiosAttributes(plan, bios, attrsPayload)
		if err != nil {
			diags.AddError("error updating bios attributes", err.Error())
			return nil, diags
		}

		// reboot the server
		pOp := powerOperator{ctx, service}
		_, err := pOp.PowerOperation(resetType, resetTimeout, intervalBiosConfigJobCheckTime)
		if err != nil {
			// TODO: handle this scenario
			diags.AddError("there was an issue restarting the server", err.Error())
			return nil, diags
		}

		// wait for the bios config job to finish
		err = common.WaitForJobToFinish(service, biosTaskURI, intervalBiosConfigJobCheckTime, biosConfigJobTimeout)
		if err != nil {
			diags.AddError("error waiting for Bios config monitor task to be completed", err.Error())
			return nil, diags
		}
		time.Sleep(60 * time.Second)
	} else {
		log.Printf("[DEBUG] BIOS attributes are already set")
	}

	state.ID = types.StringValue(bios.ODataID)

	err = r.readRedfishDellBiosAttributes(service, state)
	if err != nil {
		diags.AddError("unable to fetch currrent bios values", err.Error())
		return nil, diags
	}

	log.Printf("[DEBUG] %s: Update finished successfully", state.ID)
	return state, nil
}

func (*BiosResource) readRedfishDellBiosAttributes(service *gofish.Service, d *models.Bios) error {
	bios, err := getBiosResource(service)
	if err != nil {
		return fmt.Errorf("error fetching BIOS resource: %w", err)
	}

	old := d.Attributes.Elements()

	attributes := make(map[string]string)
	err = copyBiosAttributes(bios, attributes)
	if err != nil {
		return fmt.Errorf("error fetching BIOS attributes: %w", err)
	}

	attributesTF := make(map[string]attr.Value)
	for key, value := range attributes {
		if _, ok := old[key]; ok {
			attributesTF[key] = types.StringValue(value)
		}
	}

	d.Attributes = types.MapValueMust(types.StringType, attributesTF)

	return nil
}

func getBiosResource(service *gofish.Service) (*redfish.Bios, error) {
	system, err := getSystemResource(service)
	if err != nil {
		log.Printf("[ERROR]: Failed to get system resource: %s", err)
		return nil, err
	}

	bios, err := system.Bios()
	if err != nil {
		log.Printf("[ERROR]: Failed to get Bios resource: %s", err)
		return nil, err
	}

	return bios, nil
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
			err := fmt.Errorf("BIOS attribute %s not found", key)
			diags.AddError("There was an issue while creating/updating bios attriutes", err.Error())
			return attrsToPatch, diags
		}
		// check if the original value is an integer
		// if yes, then we need to convert accordingly
		if intOldVal, err := strconv.Atoi(attributes[key]); err == nil {
			intNewVal, err := strconv.Atoi(newVal)
			if err != nil {
				diags.AddError("There was an issue while creating/updating bios attriutes", err.Error())
				return attrsToPatch, diags
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
	return attrsToPatch, nil
}

func patchBiosAttributes(d *models.Bios, bios *redfish.Bios, attributes map[string]interface{}) (biosTaskURI string, err error) {
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
		log.Printf("error fetching data: %s", err)
		return "", err
	}
	oDataURI.Path = path.Join(oDataURI.Path, "Settings")
	settingsObjectURI := oDataURI.String()

	resp, err := bios.GetClient().Patch(settingsObjectURI, payload)
	if err != nil {
		log.Printf("[DEBUG] error sending the patch request: %s", err)
		return "", err
	}

	// check if location is present in the response header
	if location, err := resp.Location(); err == nil {
		log.Printf("[DEBUG] BIOS configuration job uri: %s", location.String())
		taskURI := location.EscapedPath()
		return taskURI, nil
	}
	return "", nil
}

// import (
// 	"context"
// 	"errors"
// 	"fmt"
// 	"github.com/dell/terraform-provider-redfish/common"
// 	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
// 	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
// 	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
// 	"github.com/stmcginnis/gofish"
// 	redfishcommon "github.com/stmcginnis/gofish/common"
// 	"github.com/stmcginnis/gofish/redfish"
// 	"log"
// 	"net/url"
// 	"path"
// 	"strconv"
// 	"strings"
// 	"time"
// )

// const (
// 	defaultBiosConfigServerResetTimeout int = 120
// 	defaultBiosConfigJobTimeout         int = 1200
// 	intervalBiosConfigJobCheckTime      int = 10
// )

// func resourceRedfishBios() *schema.Resource {
// 	return &schema.Resource{
// 		CreateContext: resourceRedfishBiosUpdate,
// 		ReadContext:   resourceRedfishBiosRead,
// 		UpdateContext: resourceRedfishBiosUpdate,
// 		DeleteContext: resourceRedfishBiosDelete,
// 		Schema:        getResourceRedfishBiosSchema(),
// 		CustomizeDiff: resourceRedfishBiosCustomizeDiff,
// 	}
// }

// func getResourceRedfishBiosSchema() map[string]*schema.Schema {
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
// 		"attributes": {
// 			Type:        schema.TypeMap,
// 			Optional:    true,
// 			Computed:    true,
// 			Description: "Bios attributes",
// 			Elem: &schema.Schema{
// 				Type: schema.TypeString,
// 			},
// 		},
// 		"settings_apply_time": {
// 			Type:     schema.TypeString,
// 			Optional: true,
// 			Description: "The time when the BIOS settings can be applied. Applicable value is 'OnReset' only. " +
// 				"In upcoming releases other apply time values will be supported. Default is \"OnReset\".",
// 			ValidateFunc: validation.StringInSlice([]string{
// 				string(redfishcommon.OnResetApplyTime),
// 			}, false),
// 			Default: string(redfishcommon.OnResetApplyTime),
// 		},
// 		"reset_type": {
// 			Type:     schema.TypeString,
// 			Optional: true,
// 			Description: "Reset type to apply on the computer system after the BIOS settings are applied. " +
// 				"Applicable values are 'ForceRestart', " +
// 				"'GracefulRestart', and 'PowerCycle'." +
// 				"Default = \"GracefulRestart\". ",
// 			ValidateFunc: validation.StringInSlice([]string{
// 				string(redfish.ForceRestartResetType),
// 				string(redfish.GracefulRestartResetType),
// 				string(redfish.PowerCycleResetType),
// 			}, false),
// 			Default: string(redfish.GracefulRestartResetType),
// 		},
// 		"reset_timeout": {
// 			Type:        schema.TypeInt,
// 			Optional:    true,
// 			Description: "reset_timeout is the time in seconds that the provider waits for the server to be reset before timing out.",
// 			Default:     defaultBiosConfigServerResetTimeout,
// 		},
// 		"bios_job_timeout": {
// 			Type:        schema.TypeInt,
// 			Optional:    true,
// 			Description: "bios_job_timeout is the time in seconds that the provider waits for the bios update job to be completed
// before timing out.",
// 			Default:     defaultBiosConfigJobTimeout,
// 		},
// 	}
// }

// func resourceRedfishBiosCustomizeDiff(_ context.Context, diff *schema.ResourceDiff, v interface{}) error {
// 	if diff.Id() == "" {
// 		return nil
// 	}

// 	if diff.HasChange("attributes") {
// 		o, n := diff.GetChange("attributes")
// 		oldAttribs := o.(map[string]interface{})
// 		newAttribs := n.(map[string]interface{})

// 		sameAttribs := biosAttributesMatch(oldAttribs, newAttribs)

// 		if sameAttribs {
// 			log.Printf("[DEBUG] Bios attributes have not changed. clearing diff")
// 			if err := diff.Clear("attributes"); err != nil {
// 				return err
// 			}
// 		} else {
// 			// Update the attributes value pairs
// 			for k, v := range newAttribs {
// 				oldAttribs[k] = v
// 			}

// 			if err := diff.SetNew("attributes", oldAttribs); err != nil {
// 				return err
// 			}
// 		}
// 	}

// 	return nil
// }

// func resourceRedfishBiosRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
// 	service, err := NewConfig(m.(*schema.ResourceData), d)
// 	if err != nil {
// 		return diag.Errorf(err.Error())
// 	}
// 	return readRedfishBiosResource(service, d)
// }

// func resourceRedfishBiosUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
// 	service, err := NewConfig(m.(*schema.ResourceData), d)
// 	if err != nil {
// 		return diag.Errorf(err.Error())
// 	}
// 	return updateRedfishBiosResource(service, d)
// }

// func resourceRedfishBiosDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
// 	var diags diag.Diagnostics
// 	d.SetId("")
// 	return diags
// }

// func updateRedfishBiosResource(service *gofish.Service, d *schema.ResourceData) diag.Diagnostics {
// 	log.Printf("[DEBUG] Beginning update")
// 	var diags diag.Diagnostics

// 	// Lock the mutex to avoid race conditions with other resources
// 	redfishMutexKV.Lock(getRedfishServerEndpoint(d))
// 	defer redfishMutexKV.Unlock(getRedfishServerEndpoint(d))

// 	resetType := d.Get("reset_type")

// 	bios, err := getBiosResource(service)
// 	if err != nil {
// 		return diag.Errorf("error fetching bios resource: %s", err)
// 	}

// 	attributes := make(map[string]string)
// 	err = copyBiosAttributes(bios, attributes)
// 	if err != nil {
// 		return diag.Errorf("error fetching bios attributes: %s", err)
// 	}

// 	attrsPayload, err := getBiosAttrsToPatch(d, attributes)
// 	if err != nil {
// 		return diag.Errorf("error getting BIOS attributes to patch: %s", err)
// 	}

// 	resetTimeout := d.Get("reset_timeout")

// 	biosConfigJobTimeout := d.Get("bios_job_timeout")

// 	log.Printf("[DEBUG] resetTimeout is set to %d  and Bios Config Job timeout is set to %d", resetTimeout.(int), biosConfigJobTimeout.(int))

// 	var biosTaskURI string
// 	if len(attrsPayload) != 0 {
// 		biosTaskURI, err = patchBiosAttributes(d, bios, attrsPayload)
// 		if err != nil {
// 			return diag.Errorf("error updating bios attributes: %s", err)
// 		}

// 		// reboot the server
// 		_, diags := PowerOperation(resetType.(string), resetTimeout.(int), intervalBiosConfigJobCheckTime, service)
// 		if diags.HasError() {
// 			// TODO: handle this scenario
// 			return diag.Errorf("there was an issue restarting the server")
// 		}

// 		// wait for the bios config job to finish
// 		err = common.WaitForJobToFinish(service, biosTaskURI, intervalBiosConfigJobCheckTime, biosConfigJobTimeout.(int))
// 		if err != nil {
// 			return diag.Errorf("Error waiting for Bios config monitor task (%s) to be completed: %s", biosTaskURI, err)
// 		}
// 		time.Sleep(30 * time.Second)
// 	} else {
// 		log.Printf("[DEBUG] BIOS attributes are already set")
// 	}

// 	if err = d.Set("attributes", attributes); err != nil {
// 		return diag.Errorf("error setting bios attributes: %s", err)
// 	}

// 	diags = readRedfishBiosResource(service, d)
// 	// Set the ID to @odata.id
// 	d.SetId(bios.ODataID)

// 	log.Printf("[DEBUG] %s: Update finished successfully", d.Id())
// 	return diags
// }

// func readRedfishBiosResource(service *gofish.Service, d *schema.ResourceData) diag.Diagnostics {

// 	log.Printf("[DEBUG] %s: Beginning read", d.Id())
// 	var diags diag.Diagnostics

// 	bios, err := getBiosResource(service)
// 	if err != nil {
// 		return diag.Errorf("error fetching BIOS resource: %s", err)
// 	}

// 	attributes := make(map[string]string)
// 	err = copyBiosAttributes(bios, attributes)
// 	if err != nil {
// 		return diag.Errorf("error fetching BIOS attributes: %s", err)
// 	}

// 	if err := d.Set("attributes", attributes); err != nil {
// 		return diag.Errorf("error setting bios attributes: %s", err)
// 	}

// 	log.Printf("[DEBUG] %s: Read finished successfully", d.Id())

// 	return diags
// }

// func copyBiosAttributes(bios *redfish.Bios, attributes map[string]string) error {

// 	// TODO: BIOS Attributes' values might be any of several types.
// 	// terraform-sdk currently does not support a map with different
// 	// value types. So we will convert int and float values to string.
// 	// copy from the BIOS attributes to the new bios attributes map
// 	for key, value := range bios.Attributes {
// 		if attrVal, ok := value.(string); ok {
// 			attributes[key] = attrVal
// 		} else {
// 			attributes[key] = fmt.Sprintf("%v", value)
// 		}
// 	}

// 	return nil
// }

// func patchBiosAttributes(d *schema.ResourceData, bios *redfish.Bios, attributes map[string]interface{}) (biosTaskURI string, err error) {

// 	payload := make(map[string]interface{})
// 	payload["Attributes"] = attributes

// 	if settingsApplyTime, ok := d.GetOk("settings_apply_time"); ok {
// 		allowedValues := bios.AllowedAttributeUpdateApplyTimes()
// 		allowed := false
// 		for i := range allowedValues {
// 			if strings.TrimSpace(settingsApplyTime.(string)) == (string)(allowedValues[i]) {
// 				allowed = true
// 				break
// 			}
// 		}

// 		if !allowed {
// 			errTxt := fmt.Sprintf("\"%s\" is not allowed as settings apply time", settingsApplyTime)
// 			err := errors.New(errTxt)
// 			return "", err
// 		}

// 		payload["@Redfish.SettingsApplyTime"] = map[string]interface{}{
// 			"ApplyTime": settingsApplyTime.(string),
// 		}
// 	}

// 	oDataURI, err := url.Parse(bios.ODataID)
// 	oDataURI.Path = path.Join(oDataURI.Path, "Settings")
// 	settingsObjectURI := oDataURI.String()

// 	resp, err := bios.GetClient().Patch(settingsObjectURI, payload)
// 	if err != nil {
// 		log.Printf("[DEBUG] error sending the patch request: %s", err)
// 		return "", err
// 	}

// 	// check if location is present in the response header
// 	if location, err := resp.Location(); err == nil {
// 		log.Printf("[DEBUG] BIOS configuration job uri: %s", location.String())

// 		taskURI := location.EscapedPath()
// 		return taskURI, nil

// 	}

// 	return "", nil
// }

// func getBiosResource(service *gofish.Service) (*redfish.Bios, error) {

// 	system, err := getSystemResource(service)
// 	if err != nil {
// 		log.Printf("[ERROR]: Failed to get system resource: %s", err)
// 		return nil, err
// 	}

// 	bios, err := system.Bios()
// 	if err != nil {
// 		log.Printf("[ERROR]: Failed to get Bios resource: %s", err)
// 		return nil, err
// 	}

// 	return bios, nil
// }

// func getBiosAttrsToPatch(d *schema.ResourceData, attributes map[string]string) (map[string]interface{}, error) {

// 	attrs := make(map[string]interface{})
// 	attrsToPatch := make(map[string]interface{})

// 	if v, ok := d.GetOk("attributes"); ok {
// 		attrs = v.(map[string]interface{})
// 	}

// 	for key, newVal := range attrs {
// 		if oldVal, ok := attributes[key]; ok {
// 			// check if the original value is an integer
// 			// if yes, then we need to convert accordingly
// 			if intOldVal, err := strconv.Atoi(attributes[key]); err == nil {
// 				intNewVal, err := strconv.Atoi(newVal.(string))
// 				if err != nil {
// 					return attrsToPatch, err
// 				}

// 				// Add to patch list if attribute value has changed
// 				if intNewVal != intOldVal {
// 					attrsToPatch[key] = intNewVal
// 				}
// 			} else {
// 				if newVal != oldVal {
// 					attrsToPatch[key] = newVal
// 				}
// 			}

// 		} else {
// 			err := fmt.Errorf("BIOS attribute %s not found", key)
// 			return attrsToPatch, err
// 		}
// 	}

// 	return attrsToPatch, nil
// }

// func biosAttributesMatch(oldAttribs, newAttribs map[string]interface{}) bool {
// 	log.Printf("[DEBUG] Begin biosAttributesMatch")

// 	for key, newVal := range newAttribs {
// 		log.Printf("[DEBUG] attribute: %v, newVal: %v", key, newVal)
// 		if oldVal, ok := oldAttribs[key]; ok {
// 			log.Printf("[DEBUG] found attribute: %v, oldVal: %v", key, oldVal)
// 			// check if the original value is an integer
// 			// if yes, then we need to convert accordingly
// 			if intOldVal, err := strconv.Atoi(oldVal.(string)); err == nil {
// 				intNewVal, err := strconv.Atoi(newVal.(string))
// 				if err != nil {
// 					return false
// 				}

// 				if intNewVal != intOldVal {
// 					return false
// 				}
// 			} else {
// 				if newVal != oldVal {
// 					return false
// 				}
// 			}
// 		} else {
// 			// attribute not found in the current state
// 			log.Printf("[DEBUG] attribute %v not found in the current state", key)
// 			return false
// 		}
// 	}
// 	return true
// }
