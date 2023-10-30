package provider

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"terraform-provider-redfish/gofish/dell"
	"terraform-provider-redfish/redfish/models"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/stmcginnis/gofish"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource = &dellIdracAttributesResource{}
)

// NewDellIdracAttributesResource is a helper function to simplify the provider implementation.
func NewDellIdracAttributesResource() resource.Resource {
	return &dellIdracAttributesResource{}
}

// dellIdracAttributesResource is the resource implementation.
type dellIdracAttributesResource struct {
	p *redfishProvider
}

// Configure implements resource.ResourceWithConfigure
func (r *dellIdracAttributesResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		resp.Diagnostics.AddError("error", "provider data is empty")
		return
	}
	r.p = req.ProviderData.(*redfishProvider)
	tflog.Trace(ctx, "resource_DellIdracAttributes configured ")
}

// Metadata returns the resource type name.
func (*dellIdracAttributesResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "dell_idrac_attributes"
}

// Schema defines the schema for the resource.
func (*dellIdracAttributesResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Resource for managing DellIdracAttributes on OpenManage Enterprise.",
		Version:             1,
		Attributes:          DellIdracAttributesSchema(),
	}
}

// DellIdracAttributesSchema to define the idrac attribute schema
func DellIdracAttributesSchema() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"id": schema.StringAttribute{
			MarkdownDescription: "ID of the iDRAC attributes resource",
			Description:         "ID of the iDRAC attributes resource",
			Computed:            true,
		},
		"redfish_server": schema.SingleNestedAttribute{
			MarkdownDescription: "Redfish Server",
			Description:         "Redfish Server",
			Required:            true,
			Attributes:          RedfishServerSchema(),
		},
		"attributes": schema.MapAttribute{
			MarkdownDescription: "iDRAC attributes. " +
				"To check allowed attributes please either use the datasource for dell idrac attributes or query " +
				"/redfish/v1/Managers/iDRAC.Embedded.1/Oem/Dell/DellAttributes/iDRAC.Embedded.1. " +
				"To get allowed values for those attributes, check " +
				"/redfish/v1/Registries/ManagerAttributeRegistry/ManagerAttributeRegistry.v1_0_0.json from a Redfish Instance",
			Description: "iDRAC attributes. " +
				"To check allowed attributes please either use the datasource for dell idrac attributes or query " +
				"/redfish/v1/Managers/iDRAC.Embedded.1/Oem/Dell/DellAttributes/iDRAC.Embedded.1. " +
				"To get allowed values for those attributes, check " +
				"/redfish/v1/Registries/ManagerAttributeRegistry/ManagerAttributeRegistry.v1_0_0.json from a Redfish Instance",
			ElementType: types.StringType,
			Required:    true,
		},
	}
}

// Create creates the resource and sets the initial Terraform state.
func (r *dellIdracAttributesResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	tflog.Trace(ctx, "resource_DellIdracAttributes create : Started")
	// Get Plan Data
	var plan models.DellIdracAttributes
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

	diags = updateRedfishDellIdracAttributes(ctx, service, &plan)
	resp.Diagnostics.Append(diags...)

	tflog.Trace(ctx, "resource_DellIdracAttributes create: updating state finished, saving ...")
	// Save into State
	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	tflog.Trace(ctx, "resource_DellIdracAttributes create: finish")
}

// Read refreshes the Terraform state with the latest data.
func (r *dellIdracAttributesResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	tflog.Trace(ctx, "resource_DellIdracAttributes read: started")
	var state models.DellIdracAttributes
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

	diags = readRedfishDellIdracAttributes(ctx, service, &state)
	resp.Diagnostics.Append(diags...)
	tflog.Trace(ctx, "resource_DellIdracAttributes read: finished reading state")
	// Save into State
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	tflog.Trace(ctx, "resource_DellIdracAttributes read: finished")
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *dellIdracAttributesResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Get state Data
	tflog.Trace(ctx, "resource_DellIdracAttributes update: started")
	var plan models.DellIdracAttributes

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

	diags = updateRedfishDellIdracAttributes(ctx, service, &plan)
	resp.Diagnostics.Append(diags...)

	tflog.Trace(ctx, "resource_DellIdracAttributes update: finished state update")
	// Save into State
	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	tflog.Trace(ctx, "resource_DellIdracAttributes update: finished")
}

// Delete deletes the resource and removes the Terraform state on success.
func (*dellIdracAttributesResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	tflog.Trace(ctx, "resource_DellIdracAttributes delete: started")
	// Get State Data
	var state models.DellIdracAttributes
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.State.RemoveResource(ctx)
	tflog.Trace(ctx, "resource_DellIdracAttributes delete: finished")
}

// ImportState import state for existing DellIdracAttributes
func (*dellIdracAttributesResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func updateRedfishDellIdracAttributes(ctx context.Context, service *gofish.Service, d *models.DellIdracAttributes) diag.Diagnostics {
	var diags diag.Diagnostics
	idracError := "there was an issue when creating/updating idrac attributes"
	d.ID = types.StringValue("placeholder")
	// Get attributes
	attributesTf := make(map[string]string)
	diags.Append(d.Attributes.ElementsAs(ctx, &attributesTf, true)...)
	// get managerAttributeRegistry to check parameters before posting them to redfish
	managerAttributeRegistry, err := getManagerAttributeRegistry(service)
	if err != nil {
		diags.AddError(idracError, err.Error())
		return diags
	}
	// Set right attributes to patch (values from map are all string. It needs int and string)
	attributesToPatch, err := setManagerAttributesRightType(attributesTf, managerAttributeRegistry)
	if err != nil {
		diags.AddError(idracError, err.Error())
		return diags
	}

	// Check that all attributes passed are compliant with the API
	err = checkManagerAttributes(managerAttributeRegistry, attributesToPatch)
	if err != nil {
		diags.AddError(idracError, err.Error())
		return diags
	}

	// get managers (Dell servers have only the iDRAC)
	managers, err := service.Managers()
	if err != nil {
		diags.AddError(idracError, err.Error())
		return diags
	}

	// Get OEM
	dellManager, err := dell.DellManager(managers[0])
	if err != nil {
		diags.AddError(idracError, err.Error())
		return diags
	}

	// Get Dell attributes
	dellAttributes, err := dellManager.DellAttributes()
	if err != nil {
		diags.AddError(idracError, err.Error())
		return diags
	}
	idracAttributes, err := getIdracAttributes(dellAttributes)
	if err != nil {
		diags.AddError(idracError, err.Error())
		return diags
	}

	// Set the body to send
	patchBody := struct {
		ApplyTime  string `json:"@Redfish.OperationApplyTime"`
		Attributes map[string]interface{}
	}{
		ApplyTime:  "Immediate",
		Attributes: attributesToPatch,
	}

	response, err := service.GetClient().Patch(idracAttributes.ODataID, patchBody)
	if err != nil {
		diags.AddError(idracError, err.Error())
		return diags
	}
	response.Body.Close()
	d.ID = types.StringValue(idracAttributes.ODataID)
	diags = readRedfishDellIdracAttributes(ctx, service, d)
	return diags
}

func readRedfishDellIdracAttributes(_ context.Context, service *gofish.Service, d *models.DellIdracAttributes) diag.Diagnostics {
	var diags diag.Diagnostics
	idracError := "there was an issue when reading idrac attributes"
	// get managers (Dell servers have only the iDRAC)
	managers, err := service.Managers()
	if err != nil {
		diags.AddError(idracError, err.Error())
		return diags
	}

	// Get OEM
	dellManager, err := dell.DellManager(managers[0])
	if err != nil {
		diags.AddError(idracError, err.Error())
		return diags
	}

	// Get Dell attributes
	dellAttributes, err := dellManager.DellAttributes()
	if err != nil {
		diags.AddError(idracError, err.Error())
		return diags
	}
	idracAttributes, err := getIdracAttributes(dellAttributes)
	if err != nil {
		diags.AddError(idracError, err.Error())
		return diags
	}

	// Get config attributes
	old := d.Attributes.Elements()
	readAttributes := make(map[string]attr.Value)

	for k, v := range old {
		// Check if attribute from config exists in idrac attributes
		attrValue := idracAttributes.Attributes[k] 
		// This is done to avoid triggering an update when reading Password values, 
		// that are shown as null (nil to Go)
		if attrValue != nil {                      
			readAttributes[k] = v
		} else {
			readAttributes[k] = v.(types.String)
		}
	}
	d.Attributes = types.MapValueMust(types.StringType, readAttributes)
	return diags
}

func getManagerAttributeRegistry(service *gofish.Service) (*dell.ManagerAttributeRegistry, error) {
	registries, err := service.Registries()
	if err != nil {
		return nil, err
	}

	for _, r := range registries {
		if r.ID == "ManagerAttributeRegistry" {
			// Get ManagerAttributesRegistry
			managerAttrRegistry, err := dell.GetDellManagerAttributeRegistry(service.GetClient(), r.Location[0].URI)
			if err != nil {
				return nil, err
			}
			return managerAttrRegistry, nil
		}
	}

	return nil, fmt.Errorf("error. Couldn't retrieve ManagerAttributeRegistry")
}

func getIdracAttributes(attributes []*dell.DellAttributes) (*dell.DellAttributes, error) {
	for _, a := range attributes {
		if strings.Contains(a.ID, "iDRAC") {
			return a, nil
		}
	}
	return nil, fmt.Errorf("couldn't find iDRACAttributes")
}

func checkManagerAttributes(attrRegistry *dell.ManagerAttributeRegistry, attributes map[string]interface{}) error {
	var errors string // Here will be collected all attribute errors to show to users

	for k, v := range attributes {
		err := attrRegistry.CheckAttribute(k, v)
		if err != nil {
			errors += fmt.Sprintf("%s - %s\n", k, err.Error())
		}
	}
	if len(errors) > 0 {
		return fmt.Errorf(errors)
	}

	return nil
}

// setManagerAttributesRightType gets a map[string]interface{} from terraform, where all keys are strings,
// and returns a map[string]interface{} where values are either string or ints, and can be used for PATCH
func setManagerAttributesRightType(rawAttributes map[string]string, registry *dell.ManagerAttributeRegistry) (map[string]interface{}, error) {
	patchMap := make(map[string]interface{})

	for k, v := range rawAttributes {
		attrType, err := registry.GetAttributeType(k)
		if err != nil {
			return nil, err
		}
		switch attrType {
		case "int":
			t, err := strconv.Atoi(v)
			if err != nil {
				return nil, fmt.Errorf("property %s must be an integer", k)
			}
			patchMap[k] = t
		case "string":
			patchMap[k] = v
		}
	}

	return patchMap, nil
}
