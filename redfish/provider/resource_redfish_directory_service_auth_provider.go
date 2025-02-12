/*
Copyright (c) 2024 Dell Inc., or its subsidiaries. All Rights Reserved.

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
	"io"
	"terraform-provider-redfish/redfish/models"

	tfpath "github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/stmcginnis/gofish"
	"github.com/stmcginnis/gofish/redfish"
)

const (
	noteMessageUpdateOneServiceOnly = "Please update one of active_directory or ldap at a time."
	noteADMessageInclusive          = "Note: `active_directory` is mutually inclusive with `active_directory_attributes`."
	noteLDAPMessageInclusive        = "Note: `ldap` is mutually inclusive with `ldap_attributes`."
	noteMessageExclusive            = "Note: `active_directory` is mutually exclusive with `ldap`."
	noteAttributesMessageExclusive  = "Note: `active_directory_attributes` is mutually exclusive with `ldap_attributes`."
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource = &RedfishDirectoryServiceAuthProviderResource{}
)

// NewRedfishDirectoryServiceAuthProviderResource is a helper function to simplify the provider implementation.
func NewRedfishDirectoryServiceAuthProviderResource() resource.Resource {
	return &RedfishDirectoryServiceAuthProviderResource{}
}

// RedfishDirectoryServiceAuthProviderResource is the resource implementation.
type RedfishDirectoryServiceAuthProviderResource struct {
	p   *redfishProvider
	ctx context.Context
}

// Configure implements resource.ResourceWithConfigure
func (r *RedfishDirectoryServiceAuthProviderResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	r.p = req.ProviderData.(*redfishProvider)
}

// Metadata returns the resource type name.
func (*RedfishDirectoryServiceAuthProviderResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "directory_service_auth_provider"
}

// Schema defines the schema for the resource.
func (*RedfishDirectoryServiceAuthProviderResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "This Terraform resource is used to configure Directory Service Auth Provider Active Directory and LDAP Service" +
			" We can Read the existing configurations or modify them using this resource.",
		Description: "This Terraform resource is used to configure Directory Service Auth Provider Active Directory and LDAP Service" +
			" We can Read the existing configurations or modify them using this resource.",

		Attributes: DirectoryServiceAuthProviderResourceSchema(),
		Blocks:     RedfishServerResourceBlockMap(),
	}
}

// Create creates the resource and sets the initial Terraform state.
func (r *RedfishDirectoryServiceAuthProviderResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	r.ctx = ctx
	tflog.Trace(ctx, "resource_directory_service_auth_provider create : Started")
	// Get Plan Data
	var plan, emptyState models.DirectoryServiceAuthProviderResource
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

	activeServiceChanged := newActiveDirectoryChanged(ctx, &plan, &emptyState)
	ldapServiceChanged := newLDAPChanged(ctx, &plan, &emptyState)

	if activeServiceChanged && ldapServiceChanged {
		resp.Diagnostics.AddError("Error when creating both of `ActiveDirectory` and `LDAP`",
			noteMessageUpdateOneServiceOnly)
		return
	}
	diags = r.updateRedfishDirectoryServiceAuth(ctx, service, &plan, &emptyState)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Trace(ctx, "resource_directory_service_auth_provider create: updating state finished, saving ...")
	diags = r.readRedfishDirectoryServiceAuthProvider(ctx, service, &plan)
	if resp.Diagnostics.Append(diags...); resp.Diagnostics.HasError() {
		return
	}
	tflog.Trace(ctx, "resource_directory_service_auth_provider create: finished state update")

	// Save into State
	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	tflog.Trace(ctx, "resource_directory_service_auth_provider create: finish")
}

// Read refreshes the Terraform state with the latest data.
func (r *RedfishDirectoryServiceAuthProviderResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	tflog.Trace(ctx, "resource_directory_service_auth_provider read: started")
	r.ctx = ctx
	var state models.DirectoryServiceAuthProviderResource
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

	diags = r.readRedfishDirectoryServiceAuthProvider(ctx, service, &state)
	if diags.HasError() {
		diags.AddError("Error running job", "error in reading the directory service")
	}

	var idracAttribute models.DellIdracAttributes
	diags = readRedfishDellIdracAttributes(ctx, service, &idracAttribute)
	resp.Diagnostics.Append(diags...)

	tflog.Trace(ctx, "resource_directory_service_auth_provider read: finished reading state")
	// Save into State
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	tflog.Trace(ctx, "resource_directory_service_auth_provider read: finished")
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *RedfishDirectoryServiceAuthProviderResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	r.ctx = ctx
	var state, plan models.DirectoryServiceAuthProviderResource
	// Get state Data
	tflog.Trace(ctx, "resource_directory_service_auth_provider update: started")
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
	api, err := NewConfig(r.p, &plan.RedfishServer)
	if err != nil {
		resp.Diagnostics.AddError("service error", err.Error())
		return
	}
	service := api.Service
	defer api.Logout()

	activeServiceChanged := newActiveDirectoryChanged(ctx, &plan, &state)
	ldapServiceChanged := newLDAPChanged(ctx, &plan, &state)
	if activeServiceChanged && ldapServiceChanged {
		resp.Diagnostics.AddError("Error when updating both of `ActiveDirectory` and `LDAP`",
			"Please update one of active_directory or ldap at a time.")
		return
	}
	diags = r.updateRedfishDirectoryServiceAuth(ctx, service, &plan, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Trace(ctx, "resource_directory_service_auth_provider update: finished state update")
	// Save into State
	diags = r.readRedfishDirectoryServiceAuthProvider(ctx, service, &plan)
	if diags.HasError() {
		diags.AddError("Error running job", "error in reading the directory service")
	}
	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	tflog.Trace(ctx, "resource_directory_service_auth_provider update: finished")
}

// Delete deletes the resource and removes the Terraform state on success.
func (*RedfishDirectoryServiceAuthProviderResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	tflog.Trace(ctx, "resource_directory_service_auth_provider delete: started")
	// Get State Data
	var state models.DirectoryServiceAuthProviderResource
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.State.RemoveResource(ctx)
	tflog.Trace(ctx, "resource_directory_service_auth_provider delete: finished")
}

// ImportState import state for existing resource
// nolint:revive
func (*RedfishDirectoryServiceAuthProviderResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	type creds struct {
		Username     string `json:"username"`
		Password     string `json:"password"`
		Endpoint     string `json:"endpoint"`
		SslInsecure  bool   `json:"ssl_insecure"`
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
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, tfpath.Root("id"), "importId")...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, redfishServer, []models.RedfishServer{server})...)
}

// nolint: gofumpt
func (*RedfishDirectoryServiceAuthProviderResource) updateRedfishDirectoryServiceAuth(ctx context.Context, service *gofish.Service, plan,
	state *models.DirectoryServiceAuthProviderResource) diag.Diagnostics {
	var diags diag.Diagnostics
	// Lock the mutex to avoid race conditions with other resources
	redfishMutexKV.Lock(plan.RedfishServer[0].Endpoint.ValueString())
	defer redfishMutexKV.Unlock(plan.RedfishServer[0].Endpoint.ValueString())

	activeServiceChanged := newActiveDirectoryChanged(ctx, plan, state)
	ldapServiceChanged := newLDAPChanged(ctx, plan, state)

	// make a call to get the device is 17G or below
	isGenerationSeventeenAndAbove, err := isServerGenerationSeventeenAndAbove(service)
	if err != nil {
		diags.AddError("Error retrieving the server generation", err.Error())
		return diags
	}
	// get the account service resource and ODATA_ID will be used to make a patch call
	accountService, err := service.AccountService()
	if err != nil {
		diags.AddError("error fetching accountservice resource", err.Error())
		return diags
	}

	accountServiceURI := accountService.ODataID

	if activeServiceChanged {
		if diags = updateActiveDirectory(ctx, accountServiceURI, service, plan, isGenerationSeventeenAndAbove); diags.HasError() {
			return diags
		}
	} else if ldapServiceChanged {
		if diags = updateLDAP(ctx, accountServiceURI, service, plan, isGenerationSeventeenAndAbove); diags.HasError() {
			return diags
		}
	}
	return diags
}

func getAccountServiceDetails(service *gofish.Service) (*redfish.AccountService, error) {
	accountService, err := service.AccountService()
	if err != nil {
		return nil, err
	}

	return accountService, nil
}

// nolint: revive
func (*RedfishDirectoryServiceAuthProviderResource) readRedfishDirectoryServiceAuthProvider(ctx context.Context, service *gofish.Service, state *models.DirectoryServiceAuthProviderResource) (diags diag.Diagnostics) {
	// var diags diag.Diagnostics
	// call function to check the generation of device
	isGenerationSeventeenAndAbove, err := isServerGenerationSeventeenAndAbove(service)
	if err != nil {
		diags.AddError("Error retrieving the server generation", err.Error())
		return diags
	}

	accountService, err := getAccountServiceDetails(service)
	if err != nil {
		diags.AddError("Error fetching Account Service", err.Error())
		return diags
	}

	if diags = parseActiveDirectoryIntoState(ctx, accountService, service, state, isGenerationSeventeenAndAbove); diags.HasError() {
		diags.AddError("ActiveDir state null", "ActiveDirectory state null")
		return diags
	}
	if diags = parseLDAPIntoState(ctx, accountService, service, state, isGenerationSeventeenAndAbove); diags.HasError() {
		diags.AddError("oldLDAPState state null", "oldLDAPState state null")
		return diags
	}
	state.ID = types.StringValue("redfish_directory_service_auth_provider")
	return diags
}

// nolint: revive
func updateActiveDirectory(ctx context.Context, serviceURI string, service *gofish.Service, plan *models.DirectoryServiceAuthProviderResource, isSeventeenGen bool) (diags diag.Diagnostics) {
	// var diags diag.Diagnostic

	// Check for all valid scenario
	if authTimeOutCheck, diags := isValidAuthTime(ActiveDirectory, ".AuthTimeout", plan); diags.HasError() || !authTimeOutCheck {
		return diags
	}

	if ssoCheck, diags := isSSOEnabledWithValidFile(ctx, ActiveDirectory, "SSOEnable", plan, isSeventeenGen); diags.HasError() || !ssoCheck {
		return diags
	}

	dcLookupDomainCheck, diags := isValidDCLookupDomainConfig(ctx, ActiveDirectory, "DCLookupEnable", plan, isSeventeenGen)
	if diags.HasError() || !dcLookupDomainCheck {
		return diags
	}
	if schemacheck, diags := isValidSchemaSelection(ctx, ActiveDirectory, "Schema", plan, isSeventeenGen); diags.HasError() || !schemacheck {
		return diags
	}

	patchBody := make(map[string]interface{})
	if patchBody[ActiveDirectory], diags = getActiveDirectoryPatchBody(ctx, plan, isSeventeenGen); diags.HasError() {
		return diags
	}

	// make a patch call to update the account service activeDirectory configuration
	response, err := service.GetClient().Patch(serviceURI, patchBody)
	if err != nil {
		diags.AddError("There was an error while creating/ updating Active Directory",
			"There was an error while creating/ updating Active Directory "+err.Error())
		return diags
	}
	if response != nil {
		body, err := io.ReadAll(response.Body)
		if err != nil {
			diags.AddError("error reading response body", "error "+string(body))
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
			diags.AddError("Error updating AccountService Details", string(errorMsg))
			return diags
		}
	}

	defer response.Body.Close()
	var idracAttributesPlan models.DellIdracAttributes
	idracAttributesPlan.Attributes = plan.ActiveDirectoryAttributes
	idracAttributesPlan.ID = plan.ID
	idracAttributesPlan.RedfishServer = plan.RedfishServer
	diags = updateRedfishDellIdracAttributes(ctx, service, &idracAttributesPlan)
	if diags.HasError() {
		return diags
	}
	plan.ActiveDirectoryAttributes = idracAttributesPlan.Attributes
	return diags
}

// nolint: revive
func updateLDAP(ctx context.Context, serviceURI string, service *gofish.Service, plan *models.DirectoryServiceAuthProviderResource, isSeventeenGen bool) (diags diag.Diagnostics) {
	patchBody := make(map[string]interface{})

	// check Server address is configured or not
	if isValid, diags := isValidLDAPConfig(ctx, plan, isSeventeenGen); diags.HasError() || !isValid {
		return diags
	}
	if patchBody["LDAP"], diags = getLDAPPatchBody(ctx, plan, isSeventeenGen); diags.HasError() {
		return diags
	}

	// make a patch call to update the account service activeDirectory configuration
	response, err := service.GetClient().Patch(serviceURI, patchBody)
	if err != nil {
		diags.AddError("There was an error while creating/ updating LDAP", "There was an error while creating/ updating LDAP")
		return diags
	}

	if response != nil {
		body, err := io.ReadAll(response.Body)
		if err != nil {
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
			diags.AddError("Error updating AccountService Details", string(errorMsg))
			return diags
		}
	}
	defer response.Body.Close()
	var idracAttributesPlan models.DellIdracAttributes
	idracAttributesPlan.Attributes = plan.LDAPAttributes
	idracAttributesPlan.ID = plan.ID
	idracAttributesPlan.RedfishServer = plan.RedfishServer
	diags = updateRedfishDellIdracAttributes(ctx, service, &idracAttributesPlan)
	// resp.Diagnostics.Append(diags...)
	if diags.HasError() {
		diags.AddError("Idrac update fail", "Idrac update fail due to invalid LDAP configuration")
		return diags
	}

	plan.LDAPAttributes = idracAttributesPlan.Attributes
	return diags
}
