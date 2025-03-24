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
	"errors"
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
	_ resource.Resource = &dellSystemAttributesResource{}
)

// NewDellSystemAttributesResource is a helper function to simplify the provider implementation.
func NewDellSystemAttributesResource() resource.Resource {
	return &dellSystemAttributesResource{}
}

// dellSystemAttributesResource is the resource implementation.
type dellSystemAttributesResource struct {
	p *redfishProvider
}

// Configure implements resource.ResourceWithConfigure
func (r *dellSystemAttributesResource) Configure(ctx context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	r.p = req.ProviderData.(*redfishProvider)
	tflog.Trace(ctx, "resource_DellSystemAttributes configured ")
}

// Metadata returns the resource type name.
func (*dellSystemAttributesResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "dell_system_attributes"
}

// Schema defines the schema for the resource.
func (*dellSystemAttributesResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "This Terraform resource is used to configure System attributes of the iDRAC Server." +
			" We can Read the existing configurations or modify them using this resource.",
		Description: "This Terraform resource is used to configure System attributes of the iDRAC Server." +
			" We can Read the existing configurations or modify them using this resource.",

		Attributes: DellSystemAttributesSchema(),
		Blocks:     RedfishServerResourceBlockMap(),
	}
}

// DellSystemAttributesSchema to define the system attribute schema
func DellSystemAttributesSchema() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"id": schema.StringAttribute{
			MarkdownDescription: "ID of the System attributes resource",
			Description:         "ID of the System attributes resource",
			Computed:            true,
		},
		"attributes": schema.MapAttribute{
			MarkdownDescription: "System attributes. " +
				"To check allowed attributes please either use the datasource for dell System attributes or query " +
				"/redfish/v1/Managers/iDRAC.Embedded.1/Oem/Dell/DellAttributes/System.Embedded.1 " +
				"To get allowed values for those attributes, check " +
				"/redfish/v1/Registries/ManagerAttributeRegistry/ManagerAttributeRegistry.v1_0_0.json from a Redfish Instance",
			Description: "System attributes. " +
				"To check allowed attributes please either use the datasource for dell System attributes or query " +
				"/redfish/v1/Managers/iDRAC.Embedded.1/Oem/Dell/DellAttributes/System.Embedded.1 " +
				"To get allowed values for those attributes, check " +
				"/redfish/v1/Registries/ManagerAttributeRegistry/ManagerAttributeRegistry.v1_0_0.json from a Redfish Instance",
			ElementType: types.StringType,
			Required:    true,
		},
	}
}

// Create creates the resource and sets the initial Terraform state.
func (r *dellSystemAttributesResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	tflog.Trace(ctx, "resource_DellSystemAttributes create : Started")
	// Get Plan Data
	var plan models.DellSystemAttributes
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

	diags = updateRedfishDellSystemAttributes(ctx, service, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Trace(ctx, "resource_DellSystemAttributes create: updating state finished, saving ...")
	// Save into State
	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	tflog.Trace(ctx, "resource_DellSystemAttributes create: finish")
}

// Read refreshes the Terraform state with the latest data.
func (r *dellSystemAttributesResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	tflog.Trace(ctx, "resource_DellSystemAttributes read: started")
	var state models.DellSystemAttributes
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

	diags = readRedfishDellSystemAttributes(ctx, service, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Trace(ctx, "resource_DellSystemAttributes read: finished reading state")
	// Save into State
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	tflog.Trace(ctx, "resource_DellSystemAttributes read: finished")
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *dellSystemAttributesResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Get state Data
	tflog.Trace(ctx, "resource_DellSystemAttributes update: started")
	var plan models.DellSystemAttributes

	// Get plan Data
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

	diags = updateRedfishDellSystemAttributes(ctx, service, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Trace(ctx, "resource_DellSystemAttributes update: finished state update")
	// Save into State
	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	tflog.Trace(ctx, "resource_DellSystemAttributes update: finished")
}

// Delete deletes the resource and removes the Terraform state on success.
func (*dellSystemAttributesResource) Delete(ctx context.Context, _ resource.DeleteRequest, resp *resource.DeleteResponse) {
	tflog.Trace(ctx, "resource_DellSystemAttributes delete: started")

	resp.State.RemoveResource(ctx)
	tflog.Trace(ctx, "resource_DellSystemAttributes delete: finished")
}

// ImportState import state for existing DellSystemAttributes
func (r *dellSystemAttributesResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	type creds struct {
		Username     string   `json:"username"`
		Password     string   `json:"password"`
		Endpoint     string   `json:"endpoint"`
		SslInsecure  bool     `json:"ssl_insecure"`
		Attributes   []string `json:"attributes"`
		RedfishAlias string   `json:"redfish_alias"`
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

	api, d := r.getEnv(&srv)
	resp.Diagnostics = append(resp.Diagnostics, d...)
	if resp.Diagnostics.HasError() {
		return
	}
	service := api.Service
	defer api.Logout()

	managerAttributeRegistry, err := getManagerAttributeRegistry(service)
	if err != nil {
		resp.Diagnostics.AddError("Error while getting manager attributes registry", err.Error())
		return
	}

	var systemAttr []string
	for _, managerAttr := range managerAttributeRegistry.Attributes {
		if strings.HasPrefix(managerAttr.ID, "System") {
			systemAttr = append(systemAttr, managerAttr.AttributeName)
		}
	}
	for readAttr := range readAttributes {
		if !slices.Contains(systemAttr, readAttr) {
			resp.Diagnostics.AddError("Invalid System attributes provided", "")
			return
		}
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, attributes, types.MapValueMust(types.StringType, readAttributes))...)
}

func (r *dellSystemAttributesResource) getEnv(rserver *[]models.RedfishServer) (*gofish.APIClient, diag.Diagnostics) {
	var d diag.Diagnostics
	// Get service
	api, err := NewConfig(r.p, rserver)
	if err != nil {
		d.AddError(ServiceErrorMsg, err.Error())
		return nil, d
	}
	return api, nil
}

func updateRedfishDellSystemAttributes(ctx context.Context, service *gofish.Service, d *models.DellSystemAttributes) diag.Diagnostics {
	tflog.Info(ctx, "updateRedfishDellSystemAttributes: started")
	var diags diag.Diagnostics
	idracError := "there was an issue when creating/updating System attributes"
	d.ID = types.StringValue("placeholder")
	// Get attributes
	attributesTf := make(map[string]string)
	diags.Append(d.Attributes.ElementsAs(ctx, &attributesTf, true)...)
	// get managerAttributeRegistry to check parameters before posting them to redfish
	managerAttributeRegistry, err := getManagerAttributeRegistry(service)
	if err != nil {
		diags.AddError(fmt.Sprintf("%s: Could not get manager attribute registry from iDRAC", idracError), err.Error())
		return diags
	}
	// get managers (Dell servers have only the iDRAC)
	managers, err := service.Managers()
	if err != nil {
		diags.AddError(fmt.Sprintf("%s: Could not get manager from iDRAC", idracError), err.Error())
		return diags
	}

	// Get OEM
	dellManager, err := dell.Manager(managers[0])
	if err != nil {
		diags.AddError(fmt.Sprintf("%s: Could not get OEM from iDRAC manager", idracError), err.Error())
		return diags
	}
	// Suppressed API to check PSPFCCapable status
	const suppressedAPI = "/Oem/Dell/DellAttributes/iDRAC.Embedded.1/Suppressed"
	suppressedURI := dellManager.Manager.ODataID + suppressedAPI
	supResp, err := service.GetClient().Get(suppressedURI)
	if err != nil {
		diags.AddError(fmt.Sprintf("%s: Could not get PlatformCapability.1.PSPFCCapable from iDRAC manager", idracError), err.Error())
		return diags
	}

	// Parasing the response of suppressed API
	readResponse, err := io.ReadAll(supResp.Body)
	if err != nil {
		diags.AddError("Failed to parse response body", err.Error())
		return diags
	}
	// ReadResponse is the response body of the suppressed API
	// Here we are unmarshaling the response body into a map[string]interface{}
	// to check the PSPFCCapable status.
	// If the status is "Enabled" then we can proceed with the modification of
	// PSPFCEnabled attribute, otherwise it is disabled.
	var decodedAttrData map[string]interface{}
	err = json.Unmarshal(readResponse, &decodedAttrData)
	if err != nil {
		diags.AddError("Cannot convert response to string", err.Error())
		return diags
	}
	capableAttr, ok := decodedAttrData["Attributes"].(map[string]interface{})["PlatformCapability.1.PSPFCCapable"]
	if !ok {
		diags.AddError("Failed to read decoded data", "Failed to read decoded data")
		return diags
	}
	// Check the status of PSPFCCapable, if enabled then procceed with modification of PSPFCEnabled attribute,
	// Otherwise throw an error because PSPFCCapable is disabled.
	if value, ok := attributesTf["ServerPwr.1.PSPFCEnabled"]; ok {
		if value == "Enabled" && capableAttr.(string) == "Disabled" {
			const attributeErr = "As PSPFCCapable Attributes disabled, Unable to update the PSPFCEnabled Attribute."
			diags.AddError(attributeErr, attributeErr)
			return diags
		}
	}

	err = assertSystemAttributes(attributesTf, managerAttributeRegistry)
	if err != nil {
		diags.AddError(fmt.Sprintf("%s: System attribute registry from iDRAC does not match input", idracError), err.Error())
		return diags
	}
	// Set right attributes to patch (values from map are all string. It needs int and string)
	attributesToPatch, err := setManagerAttributesRightType(attributesTf, managerAttributeRegistry)
	if err != nil {
		diags.AddError(fmt.Sprintf("%s: Input system attributes could not be cast to the required type", idracError), err.Error())
		return diags
	}

	// Check that all attributes passed are compliant with the API
	err = checkManagerAttributes(managerAttributeRegistry, attributesToPatch)
	if err != nil {
		diags.AddError(fmt.Sprintf("%s: Manager attribute registry from iDRAC does not match input", idracError), err.Error())
		return diags
	}

	// Get Dell attributes
	dellAttributes, err := dellManager.DellAttributes()
	if err != nil {
		diags.AddError(fmt.Sprintf("%s: Could not get dell manager attributes", idracError), err.Error())
		return diags
	}
	systemAttributes, err := getSystemAttributes(dellAttributes)
	if err != nil {
		diags.AddError(fmt.Sprintf("%s: Could not get system attributes", idracError), err.Error())
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

	response, err := service.GetClient().Patch(systemAttributes.ODataID, patchBody)
	if err != nil {
		diags.AddError(fmt.Sprintf("%s: patch request to iDRAC failed", idracError), err.Error())
		return diags
	}
	response.Body.Close() // #nosec G104
	d.ID = types.StringValue(systemAttributes.ODataID)
	diags = readRedfishDellSystemAttributes(ctx, service, d)
	return diags
}

func readRedfishDellSystemAttributes(ctx context.Context, service *gofish.Service, d *models.DellSystemAttributes) diag.Diagnostics {
	tflog.Info(ctx, "readRedfishDellSystemAttributes: started")
	var diags diag.Diagnostics
	idracError := "there was an issue when reading System attributes"
	// get managers (Dell servers have only the iDRAC)
	managers, err := service.Managers()
	if err != nil {
		diags.AddError(fmt.Sprintf("%s: Could not get manager from iDRAC", idracError), err.Error())
		return diags
	}

	// Get OEM
	dellManager, err := dell.Manager(managers[0])
	if err != nil {
		diags.AddError(fmt.Sprintf("%s: Could not get OEM from iDRAC manager", idracError), err.Error())
		return diags
	}

	// Get Dell attributes
	dellAttributes, err := dellManager.DellAttributes()
	if err != nil {
		diags.AddError(fmt.Sprintf("%s: Could not get dell manager attributes", idracError), err.Error())
		return diags
	}
	systemAttributes, err := getSystemAttributes(dellAttributes)
	if err != nil {
		diags.AddError(fmt.Sprintf("%s: Could not get system attributes", idracError), err.Error())
		return diags
	}

	// Get config attributes
	old := d.Attributes.Elements()
	readAttributes := make(map[string]attr.Value)

	if !d.Attributes.IsNull() {
		for k, v := range old {
			// Check if attribute from config exists in System attributes
			attrValue := systemAttributes.Attributes[k]
			// This is done to avoid triggering an update when reading Password values,
			// that are shown as null (nil to Go)
			if attrValue != nil {
				attributeValue(attrValue, readAttributes, k)
			} else {
				readAttributes[k] = v.(types.String)
			}
		}
	} else {
		for k, attrValue := range systemAttributes.Attributes {
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

func getSystemAttributes(attributes []*dell.Attributes) (*dell.Attributes, error) {
	for _, a := range attributes {
		if strings.Contains(a.ID, "System") {
			return a, nil
		}
	}
	return nil, fmt.Errorf("couldn't find SystemAttributes")
}

func assertSystemAttributes(rawAttributes map[string]string, managerAttributeRegistry *dell.ManagerAttributeRegistry) error {
	var err error
	// make map of name to ID of attributes
	attributes := make(map[string]string)
	for _, dellAttr := range managerAttributeRegistry.Attributes {
		attributes[dellAttr.AttributeName] = dellAttr.ID
	}

	// check if all input attributes are present in registry
	// if present, make sure that its ID starts with System, ie. it is a System attribute
	for k := range rawAttributes {
		attrID, ok := attributes[k]
		if !ok {
			err = errors.Join(err, fmt.Errorf("couldn't find manager attribute %s", k))
			continue
		}
		// check if attribute is a system attribute
		if !strings.HasPrefix(attrID, "System") {
			err = errors.Join(err, fmt.Errorf("attribute %s is not a system attribute, its ID is %s", k, attrID))
		}
	}
	return err
}
