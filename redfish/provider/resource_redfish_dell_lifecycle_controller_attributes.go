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
	"io"
	"slices"
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
			Description: "Lifecycle Controller attributes. " +
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
func (*dellLCAttributesResource) Delete(ctx context.Context, _ resource.DeleteRequest, resp *resource.DeleteResponse) {
	tflog.Trace(ctx, "resource_DellLCAttributes delete: started")

	resp.State.RemoveResource(ctx)
	tflog.Trace(ctx, "resource_DellLCAttributes delete: finished")
}

// ImportState import state for existing DellLCAttributes
func (r *dellLCAttributesResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
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
	srv := []models.RedfishServer{server}

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

	service, d := r.getLCEnv(&srv)
	resp.Diagnostics = append(resp.Diagnostics, d...)
	if resp.Diagnostics.HasError() {
		return
	}

	managerAttributeRegistry, err := getManagerAttributeRegistry(service)
	if err != nil {
		resp.Diagnostics.AddError("Error while getting manager attributes registry", err.Error())
		return
	}

	var lcAttr []string
	for _, attr := range managerAttributeRegistry.Attributes {
		if strings.HasPrefix(attr.ID, "LifecycleController") {
			lcAttr = append(lcAttr, attr.AttributeName)
		}
	}
	for attr := range readAttributes {
		if !slices.Contains(lcAttr, attr) {
			resp.Diagnostics.AddError("Invalid LC attributes provided", "")
			return
		}
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, attributes, types.MapValueMust(types.StringType, readAttributes))...)
}

func (r *dellLCAttributesResource) getLCEnv(rserver *[]models.RedfishServer) (*gofish.Service, diag.Diagnostics) {
	var d diag.Diagnostics
	// Get service
	service, err := NewConfig(r.p, rserver)
	if err != nil {
		d.AddError(ServiceErrorMsg, err.Error())
		return nil, d
	}
	return service, nil
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

	if response != nil {
		body, err := io.ReadAll(response.Body)
		if err != nil {
			diags.AddError("error reading response body", err.Error())
			return diags
		}

		readResponse := make(map[string]json.RawMessage)
		err = json.Unmarshal(body, &readResponse)
		if err != nil {
			diags.AddError("Error unmarshalling response body", err.Error())
			return diags
		}

		// check for extended error message in response
		errorMsg, ok := readResponse["error"]
		if ok {
			diags.AddError("Error updating lifecycle controller attributes", string(errorMsg))
			return diags
		}
	}

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
		if strings.Contains(a.ID, "LCAttributes") || strings.Contains(a.ID, "LifecycleController.Embedded.1") {
			return a, nil
		}
	}
	return nil, fmt.Errorf("couldn't find LCAttributes")
}
