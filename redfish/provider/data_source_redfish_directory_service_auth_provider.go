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
	"fmt"
	"strings"
	"terraform-provider-redfish/gofish/dell"
	"terraform-provider-redfish/redfish/models"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stmcginnis/gofish"
	"github.com/stmcginnis/gofish/common"
	"github.com/stmcginnis/gofish/redfish"
)

// Constants for ActiveDirectory and LDAP
const (
	ActiveDirectory = "ActiveDirectory"
	LDAP            = "LDAP"
	BLANK           = ""
)

var (
	_ datasource.DataSource              = &DirectoryServiceAuthProviderDatasource{}
	_ datasource.DataSourceWithConfigure = &DirectoryServiceAuthProviderDatasource{}
)

// NewDirectoryServiceAuthProviderDatasource is new datasource for directory Service auth provider
func NewDirectoryServiceAuthProviderDatasource() datasource.DataSource {
	return &DirectoryServiceAuthProviderDatasource{}
}

// DirectoryServiceAuthProviderDatasource to construct datasource
type DirectoryServiceAuthProviderDatasource struct {
	p       *redfishProvider
	ctx     context.Context
	service *gofish.Service
}

// Configure implements datasource.DataSourceWithConfigure
func (g *DirectoryServiceAuthProviderDatasource) Configure(_ context.Context, req datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	g.p = req.ProviderData.(*redfishProvider)
}

// Metadata implements datasource.DataSource
func (*DirectoryServiceAuthProviderDatasource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "directory_service_auth_provider"
}

// Schema implements datasource.DataSource
func (*DirectoryServiceAuthProviderDatasource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "This Terraform datasource is used to query existing Directory Service auth provider." +
			" The information fetched from this block can be further used for resource block.",
		Description: "This Terraform datasource is used to query existing Directory Service auth provider." +
			" The information fetched from this block can be further used for resource block.",
		Attributes: DirectoryServiceAuthProviderDatasourceSchema(),
		Blocks:     RedfishServerDatasourceBlockMap(),
	}
}

// DirectoryServiceAuthProviderDatasourceSchema to define the DirectoryServiceAuthProvider data-source schema
func DirectoryServiceAuthProviderDatasourceSchema() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"id": schema.StringAttribute{
			MarkdownDescription: "ID of the Directory Service Auth Provider data-source",
			Description:         "ID of the Directory Service Auth Provider data-source",
			Computed:            true,
		},
		"directory_service_auth_provider": schema.SingleNestedAttribute{
			MarkdownDescription: "Directory Service Auth Provider Attributes.",
			Description:         "Directory Service Auth Provider Attributes.",
			Attributes:          DirectoryServiceAuthProviderSchema(),
			Computed:            true,
		},
		"active_directory_attributes": schema.MapAttribute{
			MarkdownDescription: "ActiveDirectory.* attributes in Dell iDRAC attributes.",
			Description:         "ActiveDirectory.* attributes in Dell iDRAC attributes.",
			ElementType:         types.StringType,
			Computed:            true,
		},
		"ldap_attributes": schema.MapAttribute{
			MarkdownDescription: "LDAP.* attributes in Dell iDRAC attributes.",
			Description:         "LDAP.* attributes in Dell iDRAC attributes.",
			ElementType:         types.StringType,
			Computed:            true,
		},
	}
}

// Read implements datasource.DataSource
func (g *DirectoryServiceAuthProviderDatasource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var plan models.DirectoryServiceAuthProviderDatasource
	diags := req.Config.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	api, err := NewConfig(g.p, &plan.RedfishServer)
	if err != nil {
		resp.Diagnostics.AddError("service error", err.Error())
		return
	}
	defer api.Logout()
	g.ctx = ctx
	g.service = api.Service
	state, diags := g.readDatasourceRedfishDSAuthProvider(plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (g *DirectoryServiceAuthProviderDatasource) readDatasourceRedfishDSAuthProvider(d models.DirectoryServiceAuthProviderDatasource) (
	models.DirectoryServiceAuthProviderDatasource, diag.Diagnostics,
) {
	var diags diag.Diagnostics

	accountService, err := g.service.AccountService()
	if err != nil {
		diags.AddError("Error fetching Account Service", err.Error())
		return d, diags
	}

	// write the current time as ID
	d.ID = types.StringValue(fmt.Sprintf("%d", time.Now().Unix()))

	auth := newDSAuthProviderState(accountService)
	d.DirectoryServiceAuthProvider = auth

	if diags = loadActiveDirectoryAttributesState(g.service, &d); diags.HasError() {
		return d, diags
	}

	if diags = loadLDAPAttributesState(g.service, &d); diags.HasError() {
		return d, diags
	}

	return d, diags
}

func loadActiveDirectoryAttributesState(service *gofish.Service, d *models.DirectoryServiceAuthProviderDatasource) diag.Diagnostics {
	var idracAttributesState models.DellIdracAttributes
	if diags := readDatasourceRedfishDellIdracAttributes(service, &idracAttributesState); diags.HasError() {
		return diags
	}

	// nolint: gocyclo, gocognit,revive
	activeDirectoryAttributes := []string{".SSOEnable", ".AuthTimeout", ".DCLookupEnable", ".Schema", ".GCLookupEnable", ".GlobalCatalog1", ".GlobalCatalog2", ".GlobalCatalog3", ".RacName", ".RacDomain"}

	attributesToReturn := make(map[string]attr.Value)
	for k, v := range idracAttributesState.Attributes.Elements() {
		if strings.HasPrefix(k, "ActiveDirectory.") {
			for _, input := range activeDirectoryAttributes {
				if strings.HasSuffix(k, input) {
					attributesToReturn[k] = v
				}
			}
		}
		// nolint: revive
		if (strings.HasPrefix(k, "UserDomain.") && strings.HasSuffix(k, ".Name")) || (strings.HasPrefix(k, "ADGroup.") && strings.HasSuffix(k, ".Name")) {
			attributesToReturn[k] = v
		}
	}

	d.ActiveDirectoryAttributes = types.MapValueMust(types.StringType, attributesToReturn)
	return nil
}

func loadLDAPAttributesState(service *gofish.Service, d *models.DirectoryServiceAuthProviderDatasource) diag.Diagnostics {
	var idracAttributesState models.DellIdracAttributes
	if diags := readDatasourceRedfishDellIdracAttributes(service, &idracAttributesState); diags.HasError() {
		return diags
	}

	// nolint: gocyclo, gocognit,revive
	ldapAttributes := []string{".GroupAttributeIsDN", ".Port", ".BindDN", ".BindPassword", ".SearchFilter"}
	attributesToReturn := make(map[string]attr.Value)
	for k, v := range idracAttributesState.Attributes.Elements() {
		if strings.HasPrefix(k, "LDAP.") {
			for _, input := range ldapAttributes {
				if strings.HasSuffix(k, input) {
					attributesToReturn[k] = v
				}
			}
		}
	}

	d.LDAPAttributes = types.MapValueMust(types.StringType, attributesToReturn)
	return nil
}

func newDSAuthProviderState(accountService *redfish.AccountService) *models.DirectoryServiceAuthProvider {
	return &models.DirectoryServiceAuthProvider{
		ODataID:                            types.StringValue(accountService.ODataID),
		ID:                                 types.StringValue(accountService.ID),
		Name:                               types.StringValue(accountService.Name),
		Description:                        types.StringValue(accountService.Description),
		AccountLockoutCounterResetAfter:    types.Int64Value(int64(accountService.AccountLockoutCounterResetAfter)),
		AccountLockoutDuration:             types.Int64Value(int64(accountService.AccountLockoutThreshold)),
		AccountLockoutThreshold:            types.Int64Value(int64(accountService.AccountLockoutThreshold)),
		Accounts:                           newAccountsState(accountService),
		ActiveDirectory:                    newActiveDirectoryState(accountService),
		AdditionalExternalAccountProviders: newAdditionalExternalAccountProvidersState(accountService),
		AuthFailureLoggingThreshold:        types.Int64Value(int64(accountService.AuthFailureLoggingThreshold)),
		LDAP:                               newLDAPState(accountService),
		LocalAccountAuth:                   newLocalAccountAuthState(accountService.LocalAccountAuth),
		MaxPasswordLength:                  types.Int64Value(int64(accountService.MaxPasswordLength)),
		MinPasswordLength:                  types.Int64Value(int64(accountService.MinPasswordLength)),
		PasswordExpirationDays:             types.Int64Value(int64(accountService.PasswordExpirationDays)),
		PrivilegeMap:                       newPrivilegeMapState(accountService),
		Roles:                              newRolesState(accountService),
		ServiceEnabled:                     types.BoolValue(accountService.ServiceEnabled),
		Status:                             newDSAuthProviderStatusState(accountService.Status),
		SupportedAccountTypes:              newSupportedAccountTypesState(accountService.SupportedAccountTypes),
		SupportedOEMAccountTypes:           newSupportedOEMAccountTypesState(accountService.SupportedOEMAccountTypes),
	}
}

func newAccountsState(input *redfish.AccountService) types.String {
	dellAccount, accError := dell.DirectoryServiceAuthProvider(input)

	if accError != nil {
		return types.StringValue(BLANK)
	}

	return types.StringValue(dellAccount.Accounts.ODataID)
}

func newAdditionalExternalAccountProvidersState(input *redfish.AccountService) types.String {
	dellAdditional, addErr := dell.DirectoryServiceAuthProvider(input)

	if addErr != nil {
		return types.StringValue(BLANK)
	}

	return types.StringValue(dellAdditional.AdditionalExternalAccountProviders.ODataID)
}

func newPrivilegeMapState(input *redfish.AccountService) types.String {
	dellPrivilegeMap, mapErr := dell.DirectoryServiceAuthProvider(input)

	if mapErr != nil {
		return types.StringValue(BLANK)
	}

	return types.StringValue(dellPrivilegeMap.PrivilegeMap.ODataID)
}

func newRolesState(input *redfish.AccountService) types.String {
	dellRole, roleErr := dell.DirectoryServiceAuthProvider(input)

	if roleErr != nil {
		return types.StringValue(BLANK)
	}

	return types.StringValue(dellRole.Roles.ODataID)
}

func newLocalAccountAuthState(input redfish.LocalAccountAuth) types.String {
	return types.StringValue(string(input))
}

func newDirectoryState(input *redfish.AccountService, directoryType string) *models.Directory {
	var inData *redfish.ExternalAccountProvider
	if ActiveDirectory == directoryType {
		inData = &input.ActiveDirectory
	}
	if LDAP == directoryType {
		inData = &input.LDAP
	}

	return &models.Directory{
		Certificates:        newCertificatesState(input, directoryType),
		AccountProviderType: newAccountProviderTypeState(inData.AccountProviderType),
		Authentication:      newAuthenticationState(&inData.Authentication),
		RemoteRoleMapping:   newRemoteRoleMappingState(inData.RemoteRoleMapping),
		ServiceAddresses:    newServiceAddressState(inData.ServiceAddresses),
		ServiceEnabled:      types.BoolValue(inData.ServiceEnabled),
	}
}

func newActiveDirectoryState(input *redfish.AccountService) *models.ActiveDirectory {
	inData := input.ActiveDirectory
	return &models.ActiveDirectory{
		Directory:      newDirectoryState(input, ActiveDirectory),
		KerberosKeytab: types.StringValue(inData.Authentication.KerberosKeytab),
	}
}

func newCertificatesState(input *redfish.AccountService, directoryType string) types.String {
	dellCertificate, certErr := dell.DirectoryServiceAuthProvider(input)

	if certErr != nil {
		return types.StringValue(BLANK)
	}

	if directoryType == ActiveDirectory {
		return types.StringValue(dellCertificate.ActiveDirectoryCertificate.ODataID)
	}

	if directoryType == LDAP {
		return types.StringValue(dellCertificate.LDAPCertificate.ODataID)
	}

	return types.StringValue(BLANK)
}

func newAccountProviderTypeState(inputs redfish.AccountProviderTypes) types.String {
	return types.StringValue(string(inputs))
}

func newAuthenticationState(input *redfish.Authentication) *models.Authentication {
	return &models.Authentication{
		AuthenticationType: newAuthenticationTypeState(input.AuthenticationType),
	}
}

func newRemoteRoleMappingState(input []redfish.RoleMapping) []models.RemoteRoleMapping {
	var output []models.RemoteRoleMapping
	for _, v := range input {
		output = append(output, models.RemoteRoleMapping{
			RemoteGroup: types.StringValue(v.RemoteGroup),
			LocalRole:   types.StringValue(v.LocalRole),
		})
	}
	return output
}

func newDSAuthProviderStatusState(input common.Status) models.Status {
	return models.Status{
		Health:       types.StringValue(string(input.Health)),
		HealthRollup: types.StringValue(string(input.HealthRollup)),
		State:        types.StringValue(string(input.State)),
	}
}

func newSupportedOEMAccountTypesState(input []string) []types.String {
	out := make([]types.String, 0)
	for _, input := range input {
		out = append(out, types.StringValue(string(input)))
	}
	return out
}

func newServiceAddressState(input []string) []types.String {
	out := make([]types.String, 0)
	for _, input := range input {
		out = append(out, types.StringValue(input))
	}
	return out
}

func newLDAPState(input *redfish.AccountService) *models.LDAP {
	return &models.LDAP{
		Directory:   newDirectoryState(input, LDAP),
		LDAPService: newLDAPServiceState(&input.LDAP.LDAPService),
	}
}

func newLDAPServiceState(input *redfish.LDAPService) *models.LDAPService {
	return &models.LDAPService{
		SearchSettings: newSearchSettingsState(&input.SearchSettings),
	}
}

func newSearchSettingsState(input *redfish.LDAPSearchSettings) *models.SearchSettings {
	return &models.SearchSettings{
		BaseDistinguishedNames: newBaseDistinguishedNamesState(input.BaseDistinguishedNames),
		GroupNameAttribute:     types.StringValue(input.GroupNameAttribute),
		UsernameAttribute:      types.StringValue(input.UsernameAttribute),
	}
}

func newBaseDistinguishedNamesState(inputs []string) []types.String {
	out := make([]types.String, 0)
	for _, input := range inputs {
		out = append(out, types.StringValue(string(input)))
	}
	return out
}

func newAuthenticationTypeState(inputs redfish.AuthenticationTypes) types.String {
	return types.StringValue(string(inputs))
}

func newSupportedAccountTypesState(inputs []redfish.AccountTypes) []types.String {
	out := make([]types.String, 0)
	for _, input := range inputs {
		out = append(out, types.StringValue(string(input)))
	}
	return out
}

// DirectoryServiceAuthProviderSchema is a function that returns the schema for Directory Service Auth Provider
func DirectoryServiceAuthProviderSchema() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"odata_id": schema.StringAttribute{
			MarkdownDescription: "OData ID for the Account Service instance",
			Description:         "OData ID for the Account Service instance",
			Computed:            true,
		},
		"id": schema.StringAttribute{
			MarkdownDescription: "ID of the Account Service",
			Description:         "ID of the Account Service",
			Computed:            true,
		},
		"name": schema.StringAttribute{
			MarkdownDescription: "Name of the Account Service.",
			Description:         "Name of the Account Service.",
			Computed:            true,
		},
		"description": schema.StringAttribute{
			MarkdownDescription: "Description of the Account Service",
			Description:         "Description Of the Account Service",
			Computed:            true,
		},
		"account_lockout_counter_reset_after": schema.Int64Attribute{
			MarkdownDescription: "Account Lockout Counter Reset After",
			Description:         "Account Lockout Counter Reset After",
			Computed:            true,
		},
		"account_lockout_duration": schema.Int64Attribute{
			MarkdownDescription: "Account Lockout Duration",
			Description:         "Account Lockout Duration",
			Computed:            true,
		},
		"account_lockout_threshold": schema.Int64Attribute{
			MarkdownDescription: "Account Lockout Threshold",
			Description:         "Account Lockout Threshold",
			Computed:            true,
		},
		"accounts": schema.StringAttribute{
			MarkdownDescription: "Accounts is a link to a Resource Collection of type ManagerAccountCollection.",
			Description:         "Accounts is a link to a Resource Collection of type ManagerAccountCollection.",
			Computed:            true,
		},
		"active_directory": schema.SingleNestedAttribute{
			MarkdownDescription: "Active Directory",
			Description:         "Active Directory",
			Attributes:          ActiveDirectorySchema(),
			Computed:            true,
		},
		"additional_external_account_providers": schema.StringAttribute{
			MarkdownDescription: "AdditionalExternalAccountProviders is the additional external account providers that this Account Service uses.",
			Description:         "AdditionalExternalAccountProviders is the additional external account providers that this Account Service uses.",
			Computed:            true,
		},
		"auth_failure_logging_threshold": schema.Int64Attribute{
			MarkdownDescription: "Auth Failure Logging Threshold",
			Description:         "Auth Failure Logging Threshold",
			Computed:            true,
		},
		"ldap": schema.SingleNestedAttribute{
			MarkdownDescription: "LDAP",
			Description:         "LDAP",
			Attributes:          LDAPSchema(),
			Computed:            true,
		},
		"local_account_auth": schema.StringAttribute{
			MarkdownDescription: "Local Account Auth",
			Description:         "Local Account Auth",
			Computed:            true,
		},
		"max_password_length": schema.Int64Attribute{
			MarkdownDescription: "Maximum Length of the Password",
			Description:         "Maximum Length of the Password",
			Computed:            true,
		},
		"min_password_length": schema.Int64Attribute{
			MarkdownDescription: "Minimum Length of the Password",
			Description:         "Minimum Length of the Password",
			Computed:            true,
		},
		"password_expiration_days": schema.Int64Attribute{
			MarkdownDescription: "Password Expiration Days",
			Description:         "Password Expiration Days",
			Computed:            true,
		},
		"privilege_map": schema.StringAttribute{
			MarkdownDescription: "Privilege Map",
			Description:         "Privilege Map",
			Computed:            true,
		},
		"roles": schema.StringAttribute{
			MarkdownDescription: "roles is a link to a Resource Collection of type RoleCollection.",
			Description:         "roles is a link to a Resource Collection of type RoleCollection.",
			Computed:            true,
		},
		"service_enabled": schema.BoolAttribute{
			MarkdownDescription: "ServiceEnabled indicate whether the Accountr Service is enabled.",
			Description:         "ServiceEnabled indicate whether the Accountr Service is enabled.",
			Computed:            true,
		},
		"status": schema.SingleNestedAttribute{
			MarkdownDescription: "Status is any status or health properties of the Resource.",
			Description:         "Status is any status or health properties of the Resource.",
			Computed:            true,
			Attributes:          StatusSchema(),
		},
		"supported_account_types": schema.ListAttribute{
			ElementType:         types.StringType,
			MarkdownDescription: "SupportedAccountTypes is an array of the account types supported by the service.",
			Description:         "SupportedAccountTypes is an array of the account types supported by the service.",
			Computed:            true,
		},
		"supported_oem_account_types": schema.ListAttribute{
			ElementType:         types.StringType,
			MarkdownDescription: "SupportedOEMAccountTypes is an array of the OEM account types supported by the service.",
			Description:         "SupportedOEMAccountTypes is an array of the OEM account types supported by the service.",
			Computed:            true,
		},
	}
}

// DirectorySchema is a function that returns the schema for Directory
func DirectorySchema() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"certificates": schema.StringAttribute{
			MarkdownDescription: "Certificates is a link to a resource collection of type CertificateCollection" +
				" that contains certificates the external account provider uses.",
			Description: "Certificates is a link to a resource collection of type CertificateCollection" +
				" that contains certificates the external account provider uses.",
			Computed: true,
		},
		"account_provider_type": schema.StringAttribute{
			MarkdownDescription: "AccountProviderType is the type of external account provider to which this service connects.",
			Description:         "AccountProviderType is the type of external account provider to which this service connects.",
			Computed:            true,
		},
		"authentication": schema.SingleNestedAttribute{
			MarkdownDescription: "Authentication information for the account provider.",
			Description:         "Authentication information for the account provider.",
			Attributes:          AuthenticationSchema(),
			Computed:            true,
		},
		"remote_role_mapping": schema.ListNestedAttribute{
			MarkdownDescription: "Mapping rules that are used to convert the account providers account information to the local Redfish role",
			Description:         "Mapping rules that are used to convert the account providers account information to the local Redfish role",
			NestedObject: schema.NestedAttributeObject{
				Attributes: RemoteRoleMappingSchema(),
			},
			Computed: true,
		},
		"service_addresses": schema.ListAttribute{
			MarkdownDescription: "ServiceAddresses of the account providers",
			Description:         "ServiceAddresses of the account providers",
			Computed:            true,
			ElementType:         types.StringType,
		},
		"service_enabled": schema.BoolAttribute{
			MarkdownDescription: "ServiceEnabled indicate whether this service is enabled.",
			Description:         "ServiceEnabled indicate whether this service is enabled.",
			Computed:            true,
		},
	}
}

// ActiveDirectorySchema is a function that returns the schema for Active Directory
func ActiveDirectorySchema() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"directory": schema.SingleNestedAttribute{
			MarkdownDescription: "Directory for Active Directory .",
			Description:         "Directory for Active Directory",
			Attributes:          DirectorySchema(),
			Computed:            true,
		},
		"kerberos_key_tab_file": schema.StringAttribute{
			MarkdownDescription: "KerberosKeytab is a Base64-encoded version of the Kerberos keytab for this Service",
			Description:         "KerberosKeytab is a Base64-encoded version of the Kerberos keytab for this Service",
			Computed:            true,
		},
	}
}

// AuthenticationSchema is a function that returns the schema for Authentication
func AuthenticationSchema() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"authentication_type": schema.StringAttribute{
			MarkdownDescription: "AuthenticationType is used to connect to the account provider",
			Description:         "AuthenticationType is used to connect to the account provider",
			Computed:            true,
		},
	}
}

// RemoteRoleMappingSchema is a function that returns the schema for RemoteRoleMapping
func RemoteRoleMappingSchema() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"remote_group": schema.StringAttribute{
			MarkdownDescription: "Name of the remote group.",
			Description:         "Name of the remote group.",
			Computed:            true,
		},
		"local_role": schema.StringAttribute{
			MarkdownDescription: "Role Assigned to the Group.",
			Description:         "Role Assigned to the Group.",
			Computed:            true,
		},
	}
}

// LDAPSchema is a function that returns the schema for LDAP
func LDAPSchema() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"directory": schema.SingleNestedAttribute{
			MarkdownDescription: "Directory for LDAP.",
			Description:         "Directory for LDAP",
			Attributes:          DirectorySchema(),
			Computed:            true,
		},
		"ldap_service": schema.SingleNestedAttribute{
			MarkdownDescription: "LDAPService is any additional mapping information needed to parse a generic LDAP service.",
			Description:         "LDAPService is any additional mapping information needed to parse a generic LDAP service.",
			Attributes:          LDAPServiceSchema(),
			Computed:            true,
		},
	}
}

// LDAPServiceSchema is a function that returns the schema for LDAPService
func LDAPServiceSchema() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"search_settings": schema.SingleNestedAttribute{
			MarkdownDescription: "SearchSettings is the required settings to search an external LDAP service.",
			Description:         "SearchSettings is the required settings to search an external LDAP service.",
			Attributes:          SearchSettingsSchema(),
			Computed:            true,
		},
	}
}

// SearchSettingsSchema is a function that returns the schema for SearchSettings
func SearchSettingsSchema() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"base_distinguished_names": schema.ListAttribute{
			MarkdownDescription: "BaseDistinguishedNames is an array of base distinguished names to use to search an external LDAP service.",
			Description:         "BaseDistinguishedNames is an array of base distinguished names to use to search an external LDAP service.",
			Computed:            true,
			ElementType:         types.StringType,
		},

		"user_name_attribute": schema.StringAttribute{
			MarkdownDescription: "UsernameAttribute is the attribute name that contains the LDAP user name.",
			Description:         "UsernameAttribute is the attribute name that contains the LDAP user name.",
			Computed:            true,
		},

		"group_name_attribute": schema.StringAttribute{
			MarkdownDescription: "GroupNameAttribute is the attribute name that contains the LDAP group name.",
			Description:         "GroupNameAttribute is the attribute name that contains the LDAP group name.",
			Computed:            true,
		},
	}
}
