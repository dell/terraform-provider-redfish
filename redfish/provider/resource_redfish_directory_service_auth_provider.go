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
	"fmt"
	"io"
	"terraform-provider-redfish/redfish/models"

	"github.com/hashicorp/terraform-plugin-framework-validators/mapvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/objectvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	tfpath "github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/stmcginnis/gofish"
	"github.com/stmcginnis/gofish/redfish"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource = &RedfishDirectoryServiceAuthProviderResource{}
)

// NewDirectoryServiceAuthProviderResource is a helper function to simplify the provider implementation.
func NewRedfishDirectoryServiceAuthProviderResource() resource.Resource {
	return &RedfishDirectoryServiceAuthProviderResource{}
}

// DirectoryServiceAuthProviderResource is the resource implementation.
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
		MarkdownDescription: "This Terraform resource is used to configure Directory Service Active Directory and LDAP Configuration" +
			" We can Read the existing configurations or modify them using this resource.",
		Description: "This Terraform resource is used to configure Directory Service Active Directory and LDAP Configuration" +
			" We can Read the existing configurations or modify them using this resource.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "ID of the Directory Service Auth Provider resource",
				Description:         "ID of the Directory Service Auth Provider resource",
				Computed:            true,
			},
			"active_directory": schema.SingleNestedAttribute{
				MarkdownDescription: "Active Directory",
				Description:         "Active Directory",
				Attributes:          ActiveDirectoryResourceSchema(),
				Computed:            true,
				Optional:            true,
				Validators: []validator.Object{
					//objectvalidator.ConflictsWith(path.MatchRelative().AtParent().AtName("ldap")),
					/*objectvalidator.ConflictsWith(
						path.MatchRelative().AtParent().AtName("ldap"),
						path.MatchRelative().AtParent().AtName("ldap_attributes"),
					),*/
					objectvalidator.AtLeastOneOf(
						path.MatchRoot("ldap"),
						path.MatchRoot("ldap_attributes"),
					),
				},
			},
			"ldap": schema.SingleNestedAttribute{
				MarkdownDescription: "LDAP",
				Description:         "LDAP",
				Attributes:          LDAPResourceSchema(),
				Computed:            true,
				Optional:            true,
				Validators: []validator.Object{
					/*objectvalidator.ConflictsWith(
						path.MatchRelative().AtParent().AtName("active_directory"),
						path.MatchRelative().AtParent().AtName("active_directory_attributes"),
					),*/
					objectvalidator.AtLeastOneOf(
						path.MatchRoot("active_directory"),
						path.MatchRoot("active_directory_attributes"),
					),
				},
			},
			"active_directory_attributes": schema.MapAttribute{
				MarkdownDescription: "ActiveDirectory.* attributes in Dell iDRAC attributes.",
				Description:         "ActiveDirectory.* attributes in Dell iDRAC attributes.",
				ElementType:         types.StringType,
				Computed:            true,
				Optional:            true,
				Validators: []validator.Map{
					/*mapvalidator.ConflictsWith(
						path.MatchRelative().AtParent().AtName("ldap"),
						path.MatchRelative().AtParent().AtName("ldap_attributes"),
					),*/
					mapvalidator.AtLeastOneOf(
						path.MatchRoot("ldap_attributes"),
						path.MatchRoot("ldap"),
					),
				},
			},
			"ldap_attributes": schema.MapAttribute{
				MarkdownDescription: "LDAP.* attributes in Dell iDRAC attributes.",
				Description:         "LDAP.* attributes in Dell iDRAC attributes.",
				ElementType:         types.StringType,
				Computed:            true,
				Optional:            true,
				Validators: []validator.Map{
					/*mapvalidator.ConflictsWith(
						path.MatchRelative().AtParent().AtName("active_directory"),
						path.MatchRelative().AtParent().AtName("active_directory_attributes"),
					),*/
					mapvalidator.AtLeastOneOf(
						path.MatchRoot("active_directory_attributes"),
						path.MatchRoot("active_directory"),
					),
				},
			},
		},
		Blocks: RedfishServerResourceBlockMap(),
	}
}

func AuthenticationResourceSchema() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"kerberos_key_tab_file": schema.StringAttribute{
			MarkdownDescription: "KerberosKeytab is a Base64-encoded version of the Kerberos keytab for this Service",
			Description:         "KerberosKeytab is a Base64-encoded version of the Kerberos keytab for this Service",
			Computed:            true,
			Optional:            true,
		},
	}
}

// RemoteRoleMappingSchema is a function that returns the schema for Boot Options
func RemoteRoleMappingResourceSchema() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"remote_group": schema.StringAttribute{
			MarkdownDescription: "Name of the remote group.",
			Description:         "Name of the remote group.",
			Computed:            true,
			Optional:            true,
		},
		"local_role": schema.StringAttribute{
			MarkdownDescription: "Role Assigned to the Group.",
			Description:         "Role Assigned to the Group.",
			Computed:            true,
			Optional:            true,
			Validators: []validator.String{
				stringvalidator.LengthAtLeast(1),
				stringvalidator.OneOf([]string{
					string("Administrator"),
					string("Operator"),
					string("ReadOnly"),
					string("None"),
				}...),
			},
		},
	}
}

func DirectoryResourceSchema() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"remote_role_mapping": schema.ListNestedAttribute{
			MarkdownDescription: "Mapping rules that are used to convert the account providers account information to the local Redfish role",
			Description:         "Mapping rules that are used to convert the account providers account information to the local Redfish role",
			NestedObject: schema.NestedAttributeObject{
				Attributes: RemoteRoleMappingResourceSchema(),
			},
			Computed: true,
			Optional: true,
		},
		"service_addresses": schema.ListAttribute{
			MarkdownDescription: "ServiceAddresses of the account providers",
			Description:         "ServiceAddresses of the account providers",
			Computed:            true,
			Optional:            true,
			ElementType:         types.StringType,
		},
		"service_enabled": schema.BoolAttribute{
			MarkdownDescription: "ServiceEnabled indicate whether this service is enabled.",
			Description:         "ServiceEnabled indicate whether this service is enabled.",
			Computed:            true,
			Optional:            true,
		},
	}
}

// ActiveDirectorySchema is a function that returns the schema for Boot Options
func ActiveDirectoryResourceSchema() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"directory": schema.SingleNestedAttribute{
			MarkdownDescription: "Directory for Active Directory .",
			Description:         "Directory for Active Directory",
			Attributes:          DirectoryResourceSchema(),
			Computed:            true,
			Optional:            true,
		},
		"authentication": schema.SingleNestedAttribute{
			MarkdownDescription: "Authentication information for the account provider.",
			Description:         "Authentication information for the account provider.",
			Attributes:          AuthenticationResourceSchema(),
			Computed:            true,
			Optional:            true,
		},
	}
}

// LDAPResourceSchema is a function that returns the schema for Boot Options
func LDAPResourceSchema() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"directory": schema.SingleNestedAttribute{
			MarkdownDescription: "Directory for LDAP.",
			Description:         "Directory for LDAP",
			Attributes:          DirectoryResourceSchema(),
			Computed:            true,
			Optional:            true,
		},
		"ldap_service": schema.SingleNestedAttribute{
			MarkdownDescription: "LDAPService is any additional mapping information needed to parse a generic LDAP service.",
			Description:         "LDAPService is any additional mapping information needed to parse a generic LDAP service.",
			Attributes:          LDAPServiceResourceSchema(),
			Computed:            true,
			Optional:            true,
		},
	}
}

// LDAPServiceSchema is a function that returns the schema for Boot Options
func LDAPServiceResourceSchema() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"search_settings": schema.SingleNestedAttribute{
			MarkdownDescription: "SearchSettings is the required settings to search an external LDAP service.",
			Description:         "SearchSettings is the required settings to search an external LDAP service.",
			Attributes:          SearchSettingsResourceSchema(),
			Computed:            true,
			Optional:            true,
		},
	}
}

// SearchSettingsResourceSchema is a function that returns the schema for Boot Options
func SearchSettingsResourceSchema() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"base_distinguished_names": schema.ListAttribute{
			MarkdownDescription: "BaseDistinguishedNames is an array of base distinguished names to use to search an external LDAP service.",
			Description:         "BaseDistinguishedNames is an array of base distinguished names to use to search an external LDAP service.",
			Computed:            true,
			Optional:            true,
			ElementType:         types.StringType,
		},

		"user_name_attribute": schema.StringAttribute{
			MarkdownDescription: "UsernameAttribute is the attribute name that contains the LDAP user name.",
			Description:         "UsernameAttribute is the attribute name that contains the LDAP user name.",
			Computed:            true,
			Optional:            true,
		},

		"group_name_attribute": schema.StringAttribute{
			MarkdownDescription: "GroupNameAttribute is the attribute name that contains the LDAP group name.",
			Description:         "GroupNameAttribute is the attribute name that contains the LDAP group name.",
			Computed:            true,
			Optional:            true,
		},
	}
}

// Create creates the resource and sets the initial Terraform state.
func (r *RedfishDirectoryServiceAuthProviderResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	r.ctx = ctx
	tflog.Trace(ctx, "resource_directory_service_auth_provider create : Started")
	// Get Plan Data
	var plan models.DirectoryServiceAuthProviderResource
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

	state, diags := r.updateRedfishDirectoryServiceAuth(ctx, service, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Trace(ctx, "resource_directory_service_auth_provider create: updating state finished, saving ...")
	// Save into State
	diags = resp.State.Set(ctx, &state)
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

	// Get state Data
	tflog.Trace(ctx, "resource_directory_service_auth_provider update: started")
	var plan models.DirectoryServiceAuthProviderResource

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

	state, diags := r.updateRedfishDirectoryServiceAuth(ctx, service, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Trace(ctx, "resource_directory_service_auth_provider update: finished state update")
	// Save into State
	diags = resp.State.Set(ctx, &state)
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
func (*RedfishDirectoryServiceAuthProviderResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
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

func (r *RedfishDirectoryServiceAuthProviderResource) updateRedfishDirectoryServiceAuth(ctx context.Context, service *gofish.Service, plan *models.DirectoryServiceAuthProviderResource,
) (*models.DirectoryServiceAuthProviderResource, diag.Diagnostics) {
	var diags diag.Diagnostics
	state := plan

	//directoryService := plan.DirectoryServiceAuthProvider

	// Lock the mutex to avoid race conditions with other resources
	redfishMutexKV.Lock(plan.RedfishServer[0].Endpoint.ValueString())
	defer redfishMutexKV.Unlock(plan.RedfishServer[0].Endpoint.ValueString())

	activeServiceChanged := newActiveDirectoryChanged(ctx, plan)
	ldapServiceChanged := newLDAPChanged(ctx, plan)

	if activeServiceChanged && ldapServiceChanged {
		diags.AddError("Error when creating both of `ActiveDirectory` and `LDAP`",
			"Please update one of active_directory or ldap at a time.")
		return nil, diags
	}

	if !activeServiceChanged && !ldapServiceChanged {
		diags.AddError("nothing to create for `ActiveDirectory` and `LDAP`",
			"nothing to create for `ActiveDirectory` and `LDAP`")
		return nil, diags
	}

	// get the account service resource and ODATA_ID will be used to make a patch call
	accountService, err := service.AccountService()
	if err != nil {
		diags.AddError("error fetching accountservice resource", err.Error())
		return nil, diags
	}

	// Set the body to send
	patchBody := make(map[string]interface{})

	if activeServiceChanged {
		if patchBody["ActiveDirectory"], diags = getActiveDirectoryPatchBody(ctx, plan); diags.HasError() {
			return nil, diags
		}
	} else if ldapServiceChanged {
		if patchBody["LDAP"], diags = getLDAPPatchBody(ctx, plan); diags.HasError() {
			return nil, diags
		}
	}

	// make a patch call to update the account service activeDirectory or LDAP configuration
	response, err := service.GetClient().Patch(accountService.ODataID, patchBody)

	if response != nil {
		body, err := io.ReadAll(response.Body)
		if err != nil {
			diags.AddError("error reading response body", "error coming ********")
			//diags.AddError("error reading response body", )
			return nil, diags
		}

		if err == nil {
			diags.AddError("patch call to update account", string(body))
		}

		readResponse := make(map[string]json.RawMessage)
		err = json.Unmarshal(body, &readResponse)
		if err != nil {
			diags.AddError("Error unmarshalling response body", err.Error())
			return nil, diags
		}

		// check for extended error message in response
		errorMsg, ok := readResponse["error"]
		if ok {
			diags.AddError("Error updating AccountService Details", string(errorMsg))
			return nil, diags
		}
	}

	if err != nil {
		diags.AddError("There was an error while creating/ updating Active Directory", err.Error())
		return nil, diags
	}
	response.Body.Close() // #nosec G104
	//d.ID = types.StringValue(lcAttributes.ODataID)
	//diags = readRedfishDellLCAttributes(ctx, service, d)
	//return diags

	isActiveDirectoryAttributesChanged := newActiveDirectorAttributesChanged(ctx, plan)
	isLDAPAttributesChanged := newLDAPAttributesChanged(ctx, plan)
	var idracAttributesPlan models.DellIdracAttributes

	if isActiveDirectoryAttributesChanged {
		idracAttributesPlan.Attributes = plan.ActiveDirectoryAttributes
	}

	if isLDAPAttributesChanged {
		idracAttributesPlan.Attributes = plan.LDAPAttributes
	}
	//idracAttributesPlan.Attributes = plan.ActiveDirectoryAttributes
	idracAttributesPlan.ID = plan.ID
	idracAttributesPlan.RedfishServer = plan.RedfishServer
	diags = updateRedfishDellIdracAttributes(ctx, service, &idracAttributesPlan)
	//resp.Diagnostics.Append(diags...)
	if diags.HasError() {
		return nil, diags
	}

	if isActiveDirectoryAttributesChanged {
		state.ActiveDirectoryAttributes = idracAttributesPlan.Attributes
	}

	if isLDAPAttributesChanged {
		state.LDAPAttributes = idracAttributesPlan.Attributes
	}

	//state.ActiveDirectoryAttributes = idracAttributesPlan.Attributes
	diags = r.readRedfishDirectoryServiceAuthProvider(ctx, service, plan)
	if diags.HasError() {
		diags.AddError("unable to fetch currrent ActiveDirectory values", "unable to fetch currrent ActiveDirectory values")
		return nil, diags
	}

	tflog.Debug(ctx, state.ID.ValueString()+": Update finished successfully")
	return state, nil
}

/*
func getAttributesForDirectory(attrState basetypes.ObjectType) basetypes.ObjectType {

}*/

func (*RedfishDirectoryServiceAuthProviderResource) readRedfishDirectoryServiceAuthProvider(ctx context.Context, service *gofish.Service, state *models.DirectoryServiceAuthProviderResource) diag.Diagnostics {
	var diags diag.Diagnostics
	accountService, err := service.AccountService()
	if err != nil {
		diags.AddError("Error fetching Account Service", err.Error())
		return diags
	}

	// write the current time as ID
	// Get configured data into state file
	oldActiveDirState := state.ActiveDirectoryResource
	oldLDAPState := state.LDAPResource

	var idracAttributesPlan models.DellIdracAttributes

	//emptyActiveDirObj := types.ObjectNull(getActiveDirectoryModelType())
	//emptyLDAPObj := types.ObjectNull(getActiveDirectoryModelType())
	if !oldActiveDirState.IsNull() && !oldActiveDirState.IsUnknown() {
		if diags = parseActiveDirectoryIntoState(ctx, accountService, state); diags.HasError() {
			diags.AddError("ActiveDir state null", "ActiveDirectory state null")
			return diags
		}
		idracAttributesPlan.Attributes = state.ActiveDirectoryAttributes
		//state.LDAPResource = emptyLDAPObj
	}

	if !oldLDAPState.IsNull() && !oldLDAPState.IsUnknown() {
		if diags = parseLDAPIntoState(ctx, accountService, state); diags.HasError() {
			diags.AddError("oldLDAPState state null", "oldLDAPState state null")
			return diags
		}
		//idracAttributesPlan.Attributes = state.LDAPAttributes
		//state.ActiveDirectoryResource = emptyActiveDirObj
	}

	state.ID = types.StringValue("redfish_directory_service_auth_provider")

	//var idracAttribute models.DellIdracAttributes
	diags = readRedfishDellIdracAttributes(ctx, service, &idracAttributesPlan)

	/*activeDirectoryAttributesToReturn := make(map[string]attr.Value)
	ldapAttributesToReturn := make(map[string]attr.Value)
	for k, v := range idracAttributesPlan.Attributes.Elements() {
		if strings.HasPrefix(k, "ActiveDirectory.") || strings.HasPrefix(k, "UserDomain.") || strings.HasPrefix(k, "ADGroup.") {
			activeDirectoryAttributesToReturn[k] = v
		}

		if strings.HasPrefix(k, "ActiveDirectory.") {
			ldapAttributesToReturn[k] = v
		}
	}*/
	//var nilMap types.Map
	if !oldActiveDirState.IsNull() && !oldActiveDirState.IsUnknown() {
		state.ActiveDirectoryAttributes = idracAttributesPlan.Attributes
		//state.LDAPAttributes = nilMap
	}
	if !oldLDAPState.IsNull() && !oldLDAPState.IsUnknown() {
		state.LDAPAttributes = idracAttributesPlan.Attributes
		//state.ActiveDirectoryAttributes = nilMap
	}
	return diags
}

/*
func readRedfishDellIdracAttributesDS(_ context.Context, service *gofish.Service, d *models.DirectoryServiceAuthProviderResource) diag.Diagnostics {
	var diags diag.Diagnostics
	idracError := "there was an issue when reading idrac attributes"
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
	idracAttributes, err := getIdracAttributes(dellAttributes)
	if err != nil {
		diags.AddError(idracError, err.Error())
		return diags
	}

	// Get config attributes
	old := d.ActiveDirectoryAttributes.Elements()
	readAttributes := make(map[string]attr.Value)

	if !d.ActiveDirectoryAttributes.IsNull() {
		for k, v := range old {
			// Check if attribute from config exists in idrac attributes
			attrValue := idracAttributes.Attributes[k]
			// This is done to avoid triggering an update when reading Password values,
			// that are shown as null (nil to Go)
			if attrValue != nil {
				attributeValue(attrValue, readAttributes, k)
			} else {
				readAttributes[k] = v.(types.String)
			}
		}
	} else {
		for k, attrValue := range idracAttributes.Attributes {
			if attrValue != nil {
				attributeValue(attrValue, readAttributes, k)
			} else {
				readAttributes[k] = types.StringValue("")
			}
		}
	}
	d.ActiveDirectoryAttributes = types.MapValueMust(types.StringType, readAttributes)
	return diags
}
*/

func getRemoteRoleMappingValue(ctx context.Context, service *redfish.AccountService, state *models.DirectoryServiceAuthProviderResource, objectAsOptions basetypes.ObjectAsOptions) ([]attr.Value, diag.Diagnostics) {

	var diags diag.Diagnostics
	remoteRoleMapList := make([]attr.Value, 0)
	//objectAsOptions = basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true}
	var oldActiveDR models.ActiveDirectoryResource
	var oldDirectory models.DirectoryResource
	var oldRemoteRolMap []models.RemoteRoleMapping
	if !state.ActiveDirectoryResource.IsNull() && !state.ActiveDirectoryResource.IsUnknown() {
		if diags := state.ActiveDirectoryResource.As(ctx, &oldActiveDR, objectAsOptions); diags.HasError() {
			diags.AddError("oldActiveDR nil ", "oldActiveDR nil")
			return remoteRoleMapList, diags
		}
	}

	if !oldActiveDR.Directory.IsNull() && !oldActiveDR.Directory.IsUnknown() {
		if diags := oldActiveDR.Directory.As(ctx, &oldDirectory, objectAsOptions); diags.HasError() {
			diags.AddError("oldDirectory nil ", "oldDirectory nil")
			return remoteRoleMapList, diags
		}
	}

	if !oldDirectory.RemoteRoleMapping.IsNull() && !oldDirectory.RemoteRoleMapping.IsUnknown() {
		if diags := oldDirectory.RemoteRoleMapping.ElementsAs(ctx, &oldRemoteRolMap, true); diags.HasError() {
			diags.AddError("oldRemoteRolMap nil ", "oldRemoteRolMap nil")
			return remoteRoleMapList, diags
		}
	}

	for _, oldElement := range oldRemoteRolMap {
		for _, serviceElement := range service.ActiveDirectory.RemoteRoleMapping {
			//var newRemoteRoleMap redfish.RoleMapping
			if serviceElement.LocalRole == oldElement.LocalRole.ValueString() && serviceElement.RemoteGroup == oldElement.RemoteGroup.ValueString() {
				remoteRoleItemMap := map[string]attr.Value{
					"remote_group": types.StringValue(serviceElement.RemoteGroup),
					"local_role":   types.StringValue(serviceElement.LocalRole),
				}
				remoteRoleItemObj, diags := types.ObjectValue(getRemoteRoleMappingModelType(), remoteRoleItemMap)
				if diags.HasError() {
					return remoteRoleMapList, diags
				}
				remoteRoleMapList = append(remoteRoleMapList, remoteRoleItemObj)
				break
			}

		}
	}
	return remoteRoleMapList, diags
}

func getActiveDirectoryObjectValue(ctx context.Context, service *redfish.AccountService, state *models.DirectoryServiceAuthProviderResource, objectAsOptions basetypes.ObjectAsOptions) (basetypes.ObjectValue, diag.Diagnostics) {

	emptyObj := types.ObjectNull(getDirectoryModelType())
	var oldActiveDirRes models.ActiveDirectoryResource
	if !state.ActiveDirectoryResource.IsNull() && !state.ActiveDirectoryResource.IsUnknown() {
		if diags := state.ActiveDirectoryResource.As(ctx, &oldActiveDirRes, objectAsOptions); diags.HasError() {
			diags.AddError("oldActiveDirRes nill ", "oldActiveDirRes nil")
			return emptyObj, diags
		}
	}

	var directoryPlan models.DirectoryResource
	if !oldActiveDirRes.Directory.IsNull() && !oldActiveDirRes.Directory.IsUnknown() {
		if diags := oldActiveDirRes.Directory.As(ctx, &directoryPlan, objectAsOptions); diags.HasError() {
			diags.AddError("directoryPlan nill ", "directoryPlan nil")
			return emptyObj, diags
		}
	}

	var serviceAddress []string
	if !directoryPlan.ServiceAddresses.IsNull() && !directoryPlan.ServiceAddresses.IsUnknown() {
		if diags := directoryPlan.ServiceAddresses.ElementsAs(ctx, &serviceAddress, true); diags.HasError() {
			diags.AddError("serviceAddress nill ", "serviceAddress nil")
			return emptyObj, diags
		}
	}

	/*var oldLDAPRes models.LDAPResource
	if diags := state.LDAPResource.As(ctx, &oldLDAPRes, objectAsOptions); diags.HasError() {
		return emptyObj, diags
	}*/
	remoteRoleMappingObj, diags := getRemoteRoleMappingValue(ctx, service, state, objectAsOptions)
	if diags.HasError() {
		return emptyObj, diags
	}

	remoteRoleList, diags := types.ListValue(types.ObjectType{AttrTypes: getRemoteRoleMappingModelType()}, remoteRoleMappingObj)

	if diags.HasError() {
		return emptyObj, diags
	}

	//serviceAddressListValue := newTypesStringList(service.ActiveDirectory.ServiceAddresses)

	//serviceAddressList, diags := getStringListValue(ctx, serviceAddressListValue)
	serviceAddressList, diags := getConfigDataList(service.ActiveDirectory.ServiceAddresses, serviceAddress)

	if diags.HasError() {
		return emptyObj, diags
	}

	directoryMap := map[string]attr.Value{
		"remote_role_mapping": remoteRoleList,
		"service_addresses":   serviceAddressList,
		"service_enabled":     types.BoolValue(service.ActiveDirectory.ServiceEnabled),
	}

	return types.ObjectValue(getDirectoryModelType(), directoryMap)
}

func getConfigDataList(input []string, stateServiceAddress []string) (basetypes.ListValue, diag.Diagnostics) {
	out := make([]attr.Value, 0)

	for _, stateInput := range stateServiceAddress {
		for _, i := range input {
			if stateInput == i {
				out = append(out, types.StringValue(i))
				break
			}
		}
	}
	return types.ListValue(types.StringType, out)
}

// func getStringListValue(_ context.Context, stringList []types.String) (basetypes.ListValue, diag.Diagnostics) {
// 	listValue := make([]attr.Value, 0)
// 	for _, v := range stringList {
// 		listValue = append(listValue, v)
// 	}
// 	return types.ListValue(types.StringType, listValue)
// }

func getLDAPDirectoryObjectValue(ctx context.Context, service *redfish.AccountService, state *models.DirectoryServiceAuthProviderResource, objectAsOptions basetypes.ObjectAsOptions) (basetypes.ObjectValue, diag.Diagnostics) {

	emptyObj := types.ObjectNull(getDirectoryModelType())

	var oldLDAPRes models.LDAPResource
	if !state.LDAPResource.IsNull() && !state.LDAPResource.IsUnknown() {
		if diags := state.LDAPResource.As(ctx, &oldLDAPRes, objectAsOptions); diags.HasError() {
			return emptyObj, diags
		}
	}

	var directoryPlan models.DirectoryResource
	if !oldLDAPRes.Directory.IsNull() && !oldLDAPRes.Directory.IsUnknown() {
		if diags := oldLDAPRes.Directory.As(ctx, &directoryPlan, objectAsOptions); diags.HasError() {
			return emptyObj, diags
		}
	}

	var serviceAddress []string
	if !directoryPlan.ServiceAddresses.IsNull() && !directoryPlan.ServiceAddresses.IsUnknown() {
		if diags := directoryPlan.ServiceAddresses.ElementsAs(ctx, &serviceAddress, true); diags.HasError() {
			diags.AddError("serviceAddress nill ", "serviceAddress nil")
			return emptyObj, diags
		}
	}
	remoteRoleMappingObj, diags := getRemoteRoleMappingValue(ctx, service, state, objectAsOptions)
	if diags.HasError() {
		return emptyObj, diags
	}

	remoteRoleList, diags := types.ListValue(types.ObjectType{AttrTypes: getRemoteRoleMappingModelType()}, remoteRoleMappingObj)

	if diags.HasError() {
		return emptyObj, diags
	}

	serviceAddressList, diags := getConfigDataList(service.LDAP.ServiceAddresses, serviceAddress)

	//serviceAddressListValue := newTypesStringList(service.ActiveDirectory.ServiceAddresses)

	//serviceAddressList, diags := getStringListValue(ctx, serviceAddressListValue)

	if diags.HasError() {
		return emptyObj, diags
	}

	directoryMap := map[string]attr.Value{
		"remote_role_mapping": remoteRoleList,
		"service_addresses":   serviceAddressList,
		"service_enabled":     types.BoolValue(service.LDAP.ServiceEnabled),
	}

	return types.ObjectValue(getDirectoryModelType(), directoryMap)
}

func getLDAPServiceObjectValue(ctx context.Context, service *redfish.AccountService, state *models.DirectoryServiceAuthProviderResource, objectAsOptions basetypes.ObjectAsOptions) (basetypes.ObjectValue, diag.Diagnostics) {

	emptyObj := types.ObjectNull(getLDAPServiceModelType())

	var oldLDAPRes models.LDAPResource
	if !state.LDAPResource.IsNull() && !state.LDAPResource.IsUnknown() {
		if diags := state.LDAPResource.As(ctx, &oldLDAPRes, objectAsOptions); diags.HasError() {
			return emptyObj, diags
		}
	}

	var oldLDAPService models.LDAPService
	if !oldLDAPRes.LDAPService.IsNull() && !oldLDAPRes.LDAPService.IsUnknown() {
		if diags := oldLDAPRes.LDAPService.As(ctx, &oldLDAPService, objectAsOptions); diags.HasError() {
			return emptyObj, diags
		}
	}

	ldapSearchSetting, diags := getLDAPSearchSettingsObjectValue(ctx, service, state, objectAsOptions)

	if diags.HasError() {
		return emptyObj, diags
		//return ,diags
	}

	ldapServiceMap := map[string]attr.Value{
		"search_settings": ldapSearchSetting,
	}

	return types.ObjectValue(getLDAPServiceModelType(), ldapServiceMap)
}

func getLDAPSearchSettingsObjectValue(ctx context.Context, service *redfish.AccountService, state *models.DirectoryServiceAuthProviderResource, objectAsOptions basetypes.ObjectAsOptions) (basetypes.ObjectValue, diag.Diagnostics) {
	emptyObj := types.ObjectNull(getSearchSettingsModelType())
	var oldLDAPRes models.LDAPResource
	if !state.LDAPResource.IsNull() && !state.LDAPResource.IsUnknown() {
		if diags := state.LDAPResource.As(ctx, &oldLDAPRes, objectAsOptions); diags.HasError() {
			return emptyObj, diags
		}
	}
	var oldLDAPService models.LDAPServiceResource
	if !oldLDAPRes.LDAPService.IsNull() && !oldLDAPRes.LDAPService.IsUnknown() {
		if diags := oldLDAPRes.LDAPService.As(ctx, &oldLDAPService, objectAsOptions); diags.HasError() {
			return emptyObj, diags
		}
	}
	var oldSearchSettings models.SearchSettingsResource
	if !oldLDAPService.SearchSettings.IsNull() && !oldLDAPService.SearchSettings.IsUnknown() {
		if diags := oldLDAPService.SearchSettings.As(ctx, &oldSearchSettings, objectAsOptions); diags.HasError() {
			return emptyObj, diags
		}
	}

	var baseDistinguished []string
	if !oldSearchSettings.BaseDistinguishedNames.IsNull() && !oldSearchSettings.BaseDistinguishedNames.IsUnknown() {
		if diags := oldSearchSettings.BaseDistinguishedNames.ElementsAs(ctx, &baseDistinguished, true); diags.HasError() {
			diags.AddError("serviceAddress nill ", "serviceAddress nil")
			return emptyObj, diags
		}
	}

	baseDistinguishedList, diags := getConfigDataList(service.LDAP.LDAPService.SearchSettings.BaseDistinguishedNames, baseDistinguished)

	//baseDistinguishedListValue := newTypesStringList(service.LDAP.LDAPService.SearchSettings.BaseDistinguishedNames)

	//baseDistinguishedList, diags := getStringListValue(ctx, baseDistinguishedListValue)

	if diags.HasError() {
		return emptyObj, diags
	}

	searchSettingsMap := map[string]attr.Value{
		"base_distinguished_names": baseDistinguishedList,
		"user_name_attribute":      types.StringValue(service.LDAP.LDAPService.SearchSettings.UsernameAttribute),
		"group_name_attribute":     types.StringValue(service.LDAP.LDAPService.SearchSettings.GroupNameAttribute),
	}

	return types.ObjectValue(getSearchSettingsModelType(), searchSettingsMap)

}

func parseActiveDirectoryIntoState(ctx context.Context, service *redfish.AccountService, state *models.DirectoryServiceAuthProviderResource) diag.Diagnostics {
	//objectAsOptions := basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true}

	objectAsOptions := basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true}
	var oldActiveDirectory models.ActiveDirectoryResource
	if !state.ActiveDirectoryResource.IsNull() && !state.ActiveDirectoryResource.IsUnknown() {
		if diags := state.ActiveDirectoryResource.As(ctx, &oldActiveDirectory, objectAsOptions); diags.HasError() {
			diags.AddError("state.ActiveDirectoryResource.IsNull", "state.ActiveDirectoryResource.IsNull")
			return diags
		}
	}

	directoryObj, diags := getActiveDirectoryObjectValue(ctx, service, state, objectAsOptions)
	if diags.HasError() {
		return diags
	}

	authenticationObj, diags := getAuthentcationObjectValue(ctx, service, state, objectAsOptions)
	if diags.HasError() {
		return diags
	}

	activeDirectoryMap := map[string]attr.Value{
		"directory":      directoryObj,
		"authentication": authenticationObj,
	}
	state.ActiveDirectoryResource, diags = types.ObjectValue(getActiveDirectoryModelType(), activeDirectoryMap)
	return diags
}

func parseLDAPIntoState(ctx context.Context, service *redfish.AccountService, state *models.DirectoryServiceAuthProviderResource) diag.Diagnostics {
	//objectAsOptions := basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true}

	objectAsOptions := basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true}
	var oldLDAP models.LDAPResource
	if !state.LDAPResource.IsNull() && !state.LDAPResource.IsUnknown() {
		if diags := state.LDAPResource.As(ctx, &oldLDAP, objectAsOptions); diags.HasError() {
			return diags
		}
	}

	directoryObj, diags := getLDAPDirectoryObjectValue(ctx, service, state, objectAsOptions)
	if diags.HasError() {
		return diags
	}

	ldapServiceObj, diags := getLDAPServiceObjectValue(ctx, service, state, objectAsOptions)

	if diags.HasError() {
		return diags
	}
	ldapMap := map[string]attr.Value{
		"directory":    directoryObj,
		"ldap_service": ldapServiceObj,
	}
	state.ActiveDirectoryResource, diags = types.ObjectValue(getLDAPModelType(), ldapMap)
	return diags
}

func getActiveDirectoryModelType() map[string]attr.Type {
	return map[string]attr.Type{
		"directory":      types.ObjectType{AttrTypes: getDirectoryModelType()},
		"authentication": types.ObjectType{AttrTypes: getAuthentcationModelType()},
	}
}

func getLDAPModelType() map[string]attr.Type {
	return map[string]attr.Type{
		"directory":    types.ObjectType{AttrTypes: getDirectoryModelType()},
		"ldap_service": types.ObjectType{AttrTypes: getLDAPServiceModelType()},
	}
}

func getAuthentcationModelType() map[string]attr.Type {
	return map[string]attr.Type{
		"kerberos_key_tab_file": types.StringType,
	}
}

func getDirectoryModelType() map[string]attr.Type {
	return map[string]attr.Type{
		"remote_role_mapping": types.ListType{ElemType: types.ObjectType{AttrTypes: getRemoteRoleMappingModelType()}},
		"service_addresses":   types.ListType{ElemType: types.StringType},
		"service_enabled":     types.BoolType,
	}
}

func getLDAPServiceModelType() map[string]attr.Type {
	return map[string]attr.Type{
		"search_settings": types.ObjectType{AttrTypes: getSearchSettingsModelType()},
	}
}

func getSearchSettingsModelType() map[string]attr.Type {
	return map[string]attr.Type{
		"base_distinguished_names": types.ListType{ElemType: types.StringType},
		"user_name_attribute":      types.StringType,
		"group_name_attribute":     types.StringType,
	}
}

func getRemoteRoleMappingModelType() map[string]attr.Type {
	return map[string]attr.Type{
		"remote_group": types.StringType,
		"local_role":   types.StringType,
	}
}

func getAuthentcationObjectValue(ctx context.Context, service *redfish.AccountService, state *models.DirectoryServiceAuthProviderResource, objectAsOptions basetypes.ObjectAsOptions) (basetypes.ObjectValue, diag.Diagnostics) {
	authentication := map[string]attr.Value{
		"kerberos_key_tab_file": types.StringValue(service.ActiveDirectory.Authentication.KerberosKeytab),
	}

	return types.ObjectValue(getAuthentcationModelType(), authentication)
}

/*
	func newDSAuthProviderResourceState(accountService *redfish.AccountService) (basetypes.ObjectValue, diag.Diagnostics) {
		return map[string]attr.Value{
			"directory":      getDirectoryObjectValue,
			"authentication": getAuthentcationObjectValue,
		}
	}
*/
/*func newDirectoryResourceState(input *redfish.AccountService, directoryType string) *models.Directory {
	var inData *redfish.ExternalAccountProvider
	if ActiveDirectory == directoryType {
		inData = &input.ActiveDirectory
	}
	if LDAP == directoryType {
		inData = &input.LDAP
	}

	return &models.Directory{
		RemoteRoleMapping: newRemoteRoleMappingState(inData.RemoteRoleMapping),
		ServiceAddresses:  newServiceAddressState(inData.ServiceAddresses),
		ServiceEnabled:    types.BoolValue(inData.ServiceEnabled),
	}
}*/

func newActiveDirectoryChanged(ctx context.Context, attrsState *models.DirectoryServiceAuthProviderResource) bool {
	objectAsOptions := basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true}

	var activeDirectoryPlan models.ActiveDirectoryResource
	if !attrsState.ActiveDirectoryResource.IsNull() && !attrsState.ActiveDirectoryResource.IsUnknown() {
		if diags := attrsState.ActiveDirectoryResource.As(ctx, &activeDirectoryPlan, objectAsOptions); diags.HasError() {
			return false
		}
	}

	var directoryPlan models.DirectoryResource
	if !activeDirectoryPlan.Directory.IsNull() && !activeDirectoryPlan.Directory.IsUnknown() {
		if diags := activeDirectoryPlan.Directory.As(ctx, &directoryPlan, objectAsOptions); diags.HasError() {
			return false
		}
		if (!directoryPlan.ServiceEnabled.IsNull() && !directoryPlan.ServiceEnabled.IsUnknown()) ||
			(!directoryPlan.ServiceAddresses.IsNull() && !directoryPlan.ServiceAddresses.IsUnknown()) {
			return true
		}
	}

	if !directoryPlan.RemoteRoleMapping.IsNull() && !directoryPlan.RemoteRoleMapping.IsUnknown() {
		var remoteRoleMapping []models.RemoteRoleMapping

		if diags := directoryPlan.RemoteRoleMapping.ElementsAs(ctx, &remoteRoleMapping, true); diags.HasError() {
			return false
		}
		if len(remoteRoleMapping) != 0 {
			return true
		}
	}

	var authenticationPlan models.AuthenticationResource
	if !activeDirectoryPlan.Authentication.IsNull() && !activeDirectoryPlan.Authentication.IsUnknown() {
		if diags := activeDirectoryPlan.Authentication.As(ctx, &authenticationPlan, objectAsOptions); diags.HasError() {
			return false
		}
		if !authenticationPlan.KerberosKeytab.IsNull() && !authenticationPlan.KerberosKeytab.IsUnknown() {
			return true
		}
	}

	return newActiveDirectorAttributesChanged(ctx, attrsState)
}

func newLDAPChanged(ctx context.Context, attrsState *models.DirectoryServiceAuthProviderResource) bool {
	objectAsOptions := basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true}

	var ldapPlan models.LDAPResource
	if !attrsState.LDAPResource.IsNull() && !attrsState.LDAPResource.IsUnknown() {
		if diags := attrsState.LDAPResource.As(ctx, &ldapPlan, objectAsOptions); diags.HasError() {
			return false
		}
	}

	var directoryPlan models.DirectoryResource
	if !ldapPlan.Directory.IsNull() && !ldapPlan.Directory.IsUnknown() {
		if diags := ldapPlan.Directory.As(ctx, &directoryPlan, objectAsOptions); diags.HasError() {
			return false
		}
		if (!directoryPlan.ServiceEnabled.IsNull() && !directoryPlan.ServiceEnabled.IsUnknown()) ||
			(!directoryPlan.ServiceAddresses.IsNull() && !directoryPlan.ServiceAddresses.IsUnknown()) {
			return true
		}
	}

	if !directoryPlan.RemoteRoleMapping.IsNull() && !directoryPlan.RemoteRoleMapping.IsUnknown() {
		var remoteRoleMapping []models.RemoteRoleMapping

		if diags := directoryPlan.RemoteRoleMapping.ElementsAs(ctx, &remoteRoleMapping, true); diags.HasError() {
			return false
		}
		if len(remoteRoleMapping) != 0 {
			return true
		}
	}

	var ldapServicePlan models.LDAPServiceResource
	if !ldapPlan.LDAPService.IsNull() && !ldapPlan.LDAPService.IsUnknown() {

		if diags := ldapPlan.LDAPService.As(ctx, &ldapServicePlan, objectAsOptions); diags.HasError() {
			return false
		}
	}

	var ldapSearchSettingsPlan models.SearchSettingsResource
	if !ldapServicePlan.SearchSettings.IsNull() && !ldapServicePlan.SearchSettings.IsUnknown() {
		if diags := ldapServicePlan.SearchSettings.As(ctx, &ldapSearchSettingsPlan, objectAsOptions); diags.HasError() {
			return false
		}

		if (!ldapSearchSettingsPlan.GroupNameAttribute.IsNull() && !ldapSearchSettingsPlan.GroupNameAttribute.IsUnknown()) ||
			(!ldapSearchSettingsPlan.UsernameAttribute.IsNull() && !ldapSearchSettingsPlan.UsernameAttribute.IsUnknown()) {
			return true
		}
	}

	if !ldapSearchSettingsPlan.BaseDistinguishedNames.IsNull() && !ldapSearchSettingsPlan.BaseDistinguishedNames.IsUnknown() {
		var baseDistinguishedList []string

		if diags := ldapSearchSettingsPlan.BaseDistinguishedNames.ElementsAs(ctx, &baseDistinguishedList, true); diags.HasError() {
			return false
		}
		if len(baseDistinguishedList) != 0 {
			return true
		}
	}

	return newLDAPAttributesChanged(ctx, attrsState)
}

func getActiveDirectoryPatchBody(ctx context.Context, attrsState *models.DirectoryServiceAuthProviderResource) (map[string]interface{}, diag.Diagnostics) {
	supportedActiveDirectory := map[string]string{
		"service_enabled":     "ServiceEnabled",
		"service_addresses":   "ServiceAddresses",
		"remote_role_mapping": "RemoteRoleMapping",
		"authentication":      "Authentication",
	}
	supportedRemoteRoleMappingParams := map[string]string{
		"remote_group": "RemoteGroup",
		"local_role":   "LocalRole",
	}

	supportedAuthentication := map[string]string{
		"kerberos_key_tab_file": "KerberosKeytab",
	}

	objectAsOptions := basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true}

	var activeDirectoryPlan models.ActiveDirectoryResource
	if !attrsState.ActiveDirectoryResource.IsNull() && !attrsState.ActiveDirectoryResource.IsUnknown() {
		if diags := attrsState.ActiveDirectoryResource.As(ctx, &activeDirectoryPlan, objectAsOptions); diags.HasError() {
			return nil, diags
		}
	}

	var directoryPlan models.DirectoryResource
	if !activeDirectoryPlan.Directory.IsNull() && !activeDirectoryPlan.Directory.IsUnknown() {
		if diags := activeDirectoryPlan.Directory.As(ctx, &directoryPlan, objectAsOptions); diags.HasError() {
			return nil, diags
		}
	}

	//  var remoteRoleMappingPlan []models.RemoteRoleMapping
	//  if diags := directoryPlan.RemoteRoleMapping.As(ctx, &remoteRoleMappingPlan, objectAsOptions); diags.HasError() {
	//  	return nil, diags
	//  }

	var authenticationPlan models.AuthenticationResource
	if !activeDirectoryPlan.Authentication.IsNull() && !activeDirectoryPlan.Authentication.IsUnknown() {
		if diags := activeDirectoryPlan.Authentication.As(ctx, &authenticationPlan, objectAsOptions); diags.HasError() {
			return nil, diags
		}
	}

	patchBody := make(map[string]interface{})
	if !activeDirectoryPlan.Directory.IsNull() && !activeDirectoryPlan.Directory.IsUnknown() {
		for key, value := range activeDirectoryPlan.Directory.Attributes() {
			if !value.IsUnknown() && !value.IsNull() {
				goValue, err := convertTerraformValueToGoBasicValue(ctx, value)
				if err != nil {
					tflog.Trace(ctx, fmt.Sprintf("Failed to convert Ethernet value to go value: %s", err.Error()))
					continue
				}
				if fieldName, ok := supportedActiveDirectory[key]; ok {
					patchBody[fieldName] = goValue
				}
			}
		}
	}

	// get list of remote role mapping
	if !directoryPlan.RemoteRoleMapping.IsNull() && !directoryPlan.RemoteRoleMapping.IsUnknown() {
		var remoteRoleMapping []models.RemoteRoleMapping

		if diags := directoryPlan.RemoteRoleMapping.ElementsAs(ctx, &remoteRoleMapping, true); diags.HasError() {
			return nil, diags
		}

		remoteRoleMappingList := make([]interface{}, 0)
		for _, target := range remoteRoleMapping {
			remoteRoleMappingBody := make(map[string]interface{})
			if !target.LocalRole.IsNull() && !target.LocalRole.IsUnknown() {
				remoteRoleMappingBody[supportedRemoteRoleMappingParams["local_role"]] = target.LocalRole.ValueString()
			}
			if !target.RemoteGroup.IsNull() && !target.RemoteGroup.IsUnknown() {
				remoteRoleMappingBody[supportedRemoteRoleMappingParams["remote_group"]] = target.RemoteGroup.ValueString()
			}
			if len(remoteRoleMappingBody) > 0 {
				remoteRoleMappingList = append(remoteRoleMappingList, remoteRoleMappingBody)
			}
		}

		patchBody[supportedActiveDirectory["remote_role_mapping"]] = remoteRoleMappingList
	}

	if !directoryPlan.ServiceAddresses.IsNull() && !directoryPlan.ServiceAddresses.IsUnknown() {
		var serviceAddress []string

		if diags := directoryPlan.ServiceAddresses.ElementsAs(ctx, &serviceAddress, true); diags.HasError() {
			return nil, diags
		}

		serviceAddressList := make([]interface{}, 0)
		for _, target := range serviceAddress {
			serviceAddressList = append(serviceAddressList, target)
		}

		patchBody[supportedActiveDirectory["service_addresses"]] = serviceAddressList
	}

	// get directory patch body
	if !activeDirectoryPlan.Authentication.IsNull() && !activeDirectoryPlan.Authentication.IsUnknown() {
		authenticationPatchBody := make(map[string]interface{})
		for key, value := range activeDirectoryPlan.Authentication.Attributes() {
			if !value.IsUnknown() && !value.IsNull() {
				goValue, err := convertTerraformValueToGoBasicValue(ctx, value)
				if err != nil {
					tflog.Trace(ctx, fmt.Sprintf("Failed to convert VLAN value to go value: %s", err.Error()))
					continue
				}
				if fieldName, ok := supportedAuthentication[key]; ok {
					authenticationPatchBody[fieldName] = goValue
				}
			}
		}
		patchBody[supportedActiveDirectory["authentication"]] = authenticationPatchBody
	}

	return patchBody, nil
}

func getLDAPPatchBody(ctx context.Context, attrsState *models.DirectoryServiceAuthProviderResource) (map[string]interface{}, diag.Diagnostics) {
	supportedLDAP := map[string]string{
		"service_enabled":     "ServiceEnabled",
		"service_addresses":   "ServiceAddresses",
		"remote_role_mapping": "RemoteRoleMapping",
		"ldap_service":        "LDAPService",
	}
	supportedRemoteRoleMappingParams := map[string]string{
		"remote_group": "RemoteGroup",
		"local_role":   "LocalRole",
	}

	supportedLDAPService := map[string]string{
		"search_settings": "SearchSettings",
	}

	supportedSearchSetting := map[string]string{
		"base_distinguished_names": "BaseDistinguishedNames",
		"user_name_attribute":      "UsernameAttribute",
		"group_name_attribute":     "GroupNameAttribute",
	}

	objectAsOptions := basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true}

	var ldapPlan models.LDAPResource
	if !attrsState.LDAPResource.IsNull() && !attrsState.LDAPResource.IsUnknown() {
		if diags := attrsState.LDAPResource.As(ctx, &ldapPlan, objectAsOptions); diags.HasError() {
			return nil, diags
		}
	}

	var ldapServicePlan models.LDAPServiceResource
	if !ldapPlan.LDAPService.IsNull() && !ldapPlan.LDAPService.IsUnknown() {
		if diags := ldapPlan.LDAPService.As(ctx, &ldapServicePlan, objectAsOptions); diags.HasError() {
			return nil, diags
		}
	}

	var ldapSearchSettingsPlan models.SearchSettingsResource
	if !ldapServicePlan.SearchSettings.IsNull() && !ldapServicePlan.SearchSettings.IsUnknown() {
		if diags := ldapServicePlan.SearchSettings.As(ctx, &ldapSearchSettingsPlan, objectAsOptions); diags.HasError() {
			return nil, diags
		}
	}

	var directoryPlan models.DirectoryResource
	if !ldapPlan.Directory.IsNull() && !ldapPlan.Directory.IsUnknown() {
		if diags := ldapPlan.Directory.As(ctx, &directoryPlan, objectAsOptions); diags.HasError() {
			return nil, diags
		}
	}

	// var remoteRoleMappingPlan models.RemoteRoleMapping
	// if diags := directoryPlan.RemoteRoleMapping.As(ctx, &remoteRoleMappingPlan, objectAsOptions); diags.HasError() {
	// 	return nil, diags
	// }

	patchBody := make(map[string]interface{})
	if !ldapPlan.Directory.IsNull() && !ldapPlan.Directory.IsUnknown() {
		for key, value := range ldapPlan.Directory.Attributes() {
			if !value.IsUnknown() && !value.IsNull() {
				goValue, err := convertTerraformValueToGoBasicValue(ctx, value)
				if err != nil {
					tflog.Trace(ctx, fmt.Sprintf("Failed to convert Ethernet value to go value: %s", err.Error()))
					continue
				}
				if fieldName, ok := supportedLDAP[key]; ok {
					patchBody[fieldName] = goValue
				}
			}
		}
	}

	if !ldapPlan.LDAPService.IsNull() && !ldapPlan.LDAPService.IsUnknown() {
		ldapServicepatchBody := make(map[string]interface{})
		for key1, value1 := range ldapPlan.LDAPService.Attributes() {
			if !value1.IsUnknown() && !value1.IsNull() {
				if !ldapServicePlan.SearchSettings.IsNull() && !ldapServicePlan.SearchSettings.IsUnknown() {
					ldapSearchSettingPatchBody := make(map[string]interface{})
					for key, value := range ldapServicePlan.SearchSettings.Attributes() {
						if !value.IsUnknown() && !value.IsNull() {
							goValue, err := convertTerraformValueToGoBasicValue(ctx, value)
							if err != nil {
								// check basedistingushed value
								var baseDistinguishedList []string
								if !ldapSearchSettingsPlan.BaseDistinguishedNames.IsNull() && !ldapSearchSettingsPlan.BaseDistinguishedNames.IsUnknown() {
									if diags := ldapSearchSettingsPlan.BaseDistinguishedNames.ElementsAs(ctx, &baseDistinguishedList, true); diags.HasError() {
										return nil, diags
									}
								}
								if !ldapSearchSettingsPlan.BaseDistinguishedNames.IsNull() && !ldapSearchSettingsPlan.BaseDistinguishedNames.IsUnknown() {
									baseDistinguishedListValue := make([]interface{}, 0)
									for _, target := range baseDistinguishedList {
										baseDistinguishedListValue = append(baseDistinguishedListValue, target)
									}
									if fieldName, ok := supportedSearchSetting[key]; ok {
										ldapSearchSettingPatchBody[fieldName] = baseDistinguishedListValue
									}
								}
								//tflog.Trace(ctx, fmt.Sprintf("Failed to convert Ethernet value to go value: %s", err.Error()))
								continue
							}
							if fieldName, ok := supportedSearchSetting[key]; ok {
								ldapSearchSettingPatchBody[fieldName] = goValue
							}
						}
					}
					//ldapServicepatchBody[supportedLDAPService["search_settings"]] = ldapSearchSettingpatchBody
					if fieldName, ok := supportedLDAPService[key1]; ok {
						ldapServicepatchBody[fieldName] = ldapSearchSettingPatchBody
					}
				}

			}
		}
		patchBody[supportedLDAP["ldap_service"]] = ldapServicepatchBody
	}

	/*if !ldapServicePlan.SearchSettings.IsNull() && !ldapServicePlan.SearchSettings.IsUnknown() {
		ldapSearchSettingpatchBody := make(map[string]interface{})
		for key, value := range ldapServicePlan.SearchSettings.Attributes() {
			if !value.IsUnknown() && !value.IsNull() {
				goValue, err := convertTerraformValueToGoBasicValue(ctx, value)
				if err != nil {
					tflog.Trace(ctx, fmt.Sprintf("Failed to convert Ethernet value to go value: %s", err.Error()))
					continue
				}
				if fieldName, ok := supportedSearchSetting[key]; ok {
					ldapSearchSettingpatchBody[fieldName] = goValue
				}
			}
		}

		ldapServicepatchBody[supportedLDAPService["search_settings"]] = ldapSearchSettingpatchBody
		patchBody[supportedLDAP["ldap_service"]] = ldapServicepatchBody
	}*/

	// get list of remote role mapping
	if !directoryPlan.RemoteRoleMapping.IsNull() && !directoryPlan.RemoteRoleMapping.IsUnknown() {
		var remoteRoleMapping []models.RemoteRoleMapping

		if diags := directoryPlan.RemoteRoleMapping.ElementsAs(ctx, &remoteRoleMapping, true); diags.HasError() {
			return nil, diags
		}

		remoteRoleMappingList := make([]interface{}, 0)
		for _, target := range remoteRoleMapping {
			remoteRoleMappingBody := make(map[string]interface{})
			if !target.LocalRole.IsNull() && !target.LocalRole.IsUnknown() {
				remoteRoleMappingBody[supportedRemoteRoleMappingParams["local_role"]] = target.LocalRole.ValueString()
			}
			if !target.RemoteGroup.IsNull() && !target.RemoteGroup.IsUnknown() {
				remoteRoleMappingBody[supportedRemoteRoleMappingParams["remote_group"]] = target.RemoteGroup.ValueString()
			}
			if len(remoteRoleMappingBody) > 0 {
				remoteRoleMappingList = append(remoteRoleMappingList, remoteRoleMappingBody)
			}
		}

		patchBody[supportedLDAP["remote_role_mapping"]] = remoteRoleMappingList
	}

	if !directoryPlan.ServiceAddresses.IsNull() && !directoryPlan.ServiceAddresses.IsUnknown() {
		var serviceAddress []string

		if diags := directoryPlan.ServiceAddresses.ElementsAs(ctx, &serviceAddress, true); diags.HasError() {
			return nil, diags
		}

		serviceAddressList := make([]interface{}, 0)
		for _, target := range serviceAddress {
			serviceAddressList = append(serviceAddressList, target)
		}

		patchBody[supportedLDAP["service_addresses"]] = serviceAddressList
	}

	return patchBody, nil
}

func newActiveDirectorAttributesChanged(ctx context.Context, plan *models.DirectoryServiceAuthProviderResource) bool {
	if !plan.ActiveDirectoryAttributes.IsUnknown() && !plan.ActiveDirectoryAttributes.IsNull() {

		var actiMap map[string]string
		if diags := plan.ActiveDirectoryAttributes.ElementsAs(ctx, &actiMap, true); diags.HasError() {
			return false
		}

		if len(actiMap) == 0 {
			return false
		}

		return true
	}
	return false
}

func newLDAPAttributesChanged(ctx context.Context, plan *models.DirectoryServiceAuthProviderResource) bool {
	if !plan.LDAPAttributes.IsUnknown() && plan.LDAPAttributes.IsNull() {

		var ldapMap map[string]string
		if diags := plan.LDAPAttributes.ElementsAs(ctx, &ldapMap, true); diags.HasError() {
			return false
		}

		if len(ldapMap) == 0 {
			return false
		}

		return true
	}
	return false
}
