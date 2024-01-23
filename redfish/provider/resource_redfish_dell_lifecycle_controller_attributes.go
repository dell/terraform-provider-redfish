package provider

import (
	"context"
	"encoding/json"
	"fmt"
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
	_ resource.Resource = &dellLCAttributesResource{}
)

// NewDellLCAttributesResource is a helper function to simplify the provider implementation.
func NewDellLCAttributesResource() resource.Resource {
	return &dellLCAttributesResource{}
}

// DellLCAttributesResource is the resource implementation.
type dellLCAttributesResource struct {
	p *redfishProvider
}

// Configure implements resource.ResourceWithConfigure
func (r *dellLCAttributesResource) Configure(ctx context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	r.p = req.ProviderData.(*redfishProvider)
	tflog.Trace(ctx, "resource_DellLCAttributes configured ")
}

// Metadata returns the resource type name.
func (*dellLCAttributesResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "dell_lc_attributes"
}

// Schema defines the schema for the resource.
func (*dellLCAttributesResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "This Terraform resource is used to configure Lifecycle Controller attributes of the iDRAC Server." +
			" We can Read the existing configurations or modify them using this resource.",
		Description: "This Terraform resource is used to configure Lifecycle Controller attributes of the iDRAC Server." +
			" We can Read the existing configurations or modify them using this resource.",

		Attributes: DellLCAttributesSchema(),
		Blocks:     RedfishServerResourceBlockMap(),
	}
}

// DellLCAttributesSchema to define the lifecycle controller attribute schema
func DellLCAttributesSchema() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"id": schema.StringAttribute{
			MarkdownDescription: "ID of the LC attributes resource",
			Description:         "ID of the LC attributes resource",
			Computed:            true,
		},
		"attributes": schema.MapAttribute{
			MarkdownDescription: "Lifecycle Controller attributes. " +
				"To check allowed attributes please either use the datasource for dell LC attributes or query " +
				"/redfish/v1/Managers/iDRAC.Embedded.1/Oem/Dell/DellAttributes/LifecycleController.Embedded.1 " +
				"To get allowed values for those attributes, check " +
				"/redfish/v1/Registries/ManagerAttributeRegistry/ManagerAttributeRegistry.v1_0_0.json from a Redfish Instance",
			Description: "iDRAC attributes. " +
				"To check allowed attributes please either use the datasource for dell LC attributes or query " +
				"/redfish/v1/Managers/iDRAC.Embedded.1/Oem/Dell/DellAttributes/LifecycleController.Embedded.1 " +
				"To get allowed values for those attributes, check " +
				"/redfish/v1/Registries/ManagerAttributeRegistry/ManagerAttributeRegistry.v1_0_0.json from a Redfish Instance",
			ElementType: types.StringType,
			Required:    true,
		},
	}
}

// Create creates the resource and sets the initial Terraform state.
func (r *dellLCAttributesResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	tflog.Trace(ctx, "resource_DellLCAttributes create : Started")
	// Get Plan Data
	var plan models.DellLCAttributes
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

	diags = updateRedfishDellLCAttributes(ctx, service, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Trace(ctx, "resource_DellLCAttributes create: updating state finished, saving ...")
	// Save into State
	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	tflog.Trace(ctx, "resource_DellLCAttributes create: finish")
}

// Read refreshes the Terraform state with the latest data.
func (r *dellLCAttributesResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	tflog.Trace(ctx, "resource_DellLCAttributes read: started")
	var state models.DellLCAttributes
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

	diags = readRedfishDellLCAttributes(ctx, service, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Trace(ctx, "resource_DellLCAttributes read: finished reading state")
	// Save into State
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	tflog.Trace(ctx, "resource_DellLCAttributes read: finished")
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *dellLCAttributesResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Get state Data
	tflog.Trace(ctx, "resource_DellLCAttributes update: started")
	var plan models.DellLCAttributes

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

	diags = updateRedfishDellLCAttributes(ctx, service, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Trace(ctx, "resource_DellLCAttributes update: finished state update")
	// Save into State
	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	tflog.Trace(ctx, "resource_DellLCAttributes update: finished")
}

// Delete deletes the resource and removes the Terraform state on success.
func (*dellLCAttributesResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	tflog.Trace(ctx, "resource_DellLCAttributes delete: started")

	resp.State.RemoveResource(ctx)
	tflog.Trace(ctx, "resource_DellLCAttributes delete: finished")
}

// ImportState import state for existing DellLCAttributes
func (*dellLCAttributesResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	type creds struct {
		Username    string   `json:"username"`
		Password    string   `json:"password"`
		Endpoint    string   `json:"endpoint"`
		SslInsecure bool     `json:"ssl_insecure"`
		Attributes  []string `json:"attributes"`
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

	idAttrPath := path.Root("id")
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, idAttrPath, "importId")...)

	redfishServer := path.Root("redfish_server")
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, redfishServer, []models.RedfishServer{server})...)

	attributes := path.Root("attributes")
	if c.Attributes == nil {
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, attributes, types.MapNull(types.StringType))...)
		return
	}
	readAttributes := make(map[string]attr.Value)
	for _, k := range c.Attributes {
		readAttributes[k] = types.StringValue("")
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, attributes, types.MapValueMust(types.StringType, readAttributes))...)
}

func updateRedfishDellLCAttributes(ctx context.Context, service *gofish.Service, d *models.DellLCAttributes) diag.Diagnostics {
	var diags diag.Diagnostics
	idracError := "there was an issue when creating/updating LC attributes"
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
	dellManager, err := dell.Manager(managers[0])
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
	lcAttributes, err := getLCAttributes(dellAttributes)
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

	response, err := service.GetClient().Patch(lcAttributes.ODataID, patchBody)
	if err != nil {
		diags.AddError(idracError, err.Error())
		return diags
	}
	response.Body.Close() // #nosec G104
	d.ID = types.StringValue(lcAttributes.ODataID)
	diags = readRedfishDellLCAttributes(ctx, service, d)
	return diags
}

func readRedfishDellLCAttributes(_ context.Context, service *gofish.Service, d *models.DellLCAttributes) diag.Diagnostics {
	var diags diag.Diagnostics
	idracError := "there was an issue when reading LC attributes"
	// get managers (Dell servers have only the iDRAC)
	managers, err := service.Managers()
	if err != nil {
		diags.AddError(idracError, err.Error())
		return diags
	}

	// Get OEM
	dellManager, err := dell.Manager(managers[0])
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
	lcAttributes, err := getLCAttributes(dellAttributes)
	if err != nil {
		diags.AddError(idracError, err.Error())
		return diags
	}

	// Get config attributes
	old := d.Attributes.Elements()
	readAttributes := make(map[string]attr.Value)

	if !d.Attributes.IsNull() {
		for k, v := range old {
			// Check if attribute from config exists in LC attributes
			attrValue := lcAttributes.Attributes[k]
			// This is done to avoid triggering an update when reading Password values,
			// that are shown as null (nil to Go)
			if attrValue != nil {
				attributeValue(attrValue, readAttributes, k)
			} else {
				readAttributes[k] = v.(types.String)
			}
		}
	} else {
		for k, attrValue := range lcAttributes.Attributes {
			if attrValue != nil {
				attributeValue(attrValue, readAttributes, k)
			} else {
				readAttributes[k] = types.StringValue("")
			}
		}
	}
	d.Attributes = types.MapValueMust(types.StringType, readAttributes)
	return diags
}

func getLCAttributes(attributes []*dell.Attributes) (*dell.Attributes, error) {
	for _, a := range attributes {
		if strings.Contains(a.ID, "LCAttributes") {
			return a, nil
		}
	}
	return nil, fmt.Errorf("couldn't find LCAttributes")
}
