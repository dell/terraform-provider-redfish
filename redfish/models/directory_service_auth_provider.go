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

package models

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// DirectoryServiceAuthProviderDatasource to construct terraform schema for the auth provider resource.
type DirectoryServiceAuthProviderDatasource struct {
	ID                           types.String                  `tfsdk:"id"`
	RedfishServer                []RedfishServer               `tfsdk:"redfish_server"`
	DirectoryServiceAuthProvider *DirectoryServiceAuthProvider `tfsdk:"directory_service_auth_provider"`
	ActiveDirectoryAttributes    types.Map                     `tfsdk:"active_directory_attributes"`
	LDAPAttributes               types.Map                     `tfsdk:"ldap_attributes"`
}

// DirectoryServiceAuthProvider is the tfsdk model of DirectoryServiceAuthProvider
type DirectoryServiceAuthProvider struct {
	ODataID                            types.String     `tfsdk:"odata_id"`
	ID                                 types.String     `tfsdk:"id"`
	Name                               types.String     `tfsdk:"name"`
	Description                        types.String     `tfsdk:"description"`
	AccountLockoutCounterResetAfter    types.Int64      `tfsdk:"account_lockout_counter_reset_after"`
	AccountLockoutDuration             types.Int64      `tfsdk:"account_lockout_duration"`
	AccountLockoutThreshold            types.Int64      `tfsdk:"account_lockout_threshold"`
	ActiveDirectory                    *ActiveDirectory `tfsdk:"active_directory"`
	AdditionalExternalAccountProviders types.String     `tfsdk:"additional_external_account_providers"`
	AuthFailureLoggingThreshold        types.Int64      `tfsdk:"auth_failure_logging_threshold"`
	LDAP                               *LDAP            `tfsdk:"ldap"`
	Accounts                           types.String     `tfsdk:"accounts"`
	LocalAccountAuth                   types.String     `tfsdk:"local_account_auth"`
	MaxPasswordLength                  types.Int64      `tfsdk:"max_password_length"`
	MinPasswordLength                  types.Int64      `tfsdk:"min_password_length"`
	PasswordExpirationDays             types.Int64      `tfsdk:"password_expiration_days"`
	PrivilegeMap                       types.String     `tfsdk:"privilege_map"`
	Roles                              types.String     `tfsdk:"roles"`
	ServiceEnabled                     types.Bool       `tfsdk:"service_enabled"`
	Status                             Status           `tfsdk:"status"`
	SupportedAccountTypes              []types.String   `tfsdk:"supported_account_types"`
	SupportedOEMAccountTypes           []types.String   `tfsdk:"supported_oem_account_types"`
}

// Directory is the tfsdk model of Directory
type Directory struct {
	Certificates        types.String        `tfsdk:"certificates"`
	AccountProviderType types.String        `tfsdk:"account_provider_type"`
	Authentication      *Authentication     `tfsdk:"authentication"`
	RemoteRoleMapping   []RemoteRoleMapping `tfsdk:"remote_role_mapping"`
	ServiceAddresses    []types.String      `tfsdk:"service_addresses"`
	ServiceEnabled      types.Bool          `tfsdk:"service_enabled"`
}

// ActiveDirectory is the tfsdk model of ActiveDirectory
type ActiveDirectory struct {
	Directory *Directory `tfsdk:"directory"`
}

// LDAP is the tfsdk model of LDAP
type LDAP struct {
	Directory   *Directory   `tfsdk:"directory"`
	LDAPService *LDAPService `tfsdk:"ldap_service"`
}

// Authentication is the tfsdk model of Authentication
type Authentication struct {
	AuthenticationType types.String `tfsdk:"authentication_type"`
}

// RemoteRoleMapping is the tfsdk model of RemoteRoleMapping
type RemoteRoleMapping struct {
	RemoteGroup types.String `tfsdk:"remote_group"`
	LocalRole   types.String `tfsdk:"local_role"`
}

// LDAPService is the tfsdk model of LDAPService
type LDAPService struct {
	SearchSettings *SearchSettings `tfsdk:"search_settings"`
}

// SearchSettings is the tfsdk model of SearchSettings
type SearchSettings struct {
	BaseDistinguishedNames []types.String `tfsdk:"base_distinguished_names"`
	UsernameAttribute      types.String   `tfsdk:"user_name_attribute"`
	GroupNameAttribute     types.String   `tfsdk:"group_name_attribute"`
}
