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
	"github.com/hashicorp/terraform-plugin-framework-validators/mapvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/objectvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const (
	comma = " , "
)

// DirectoryServiceAuthProviderResourceSchema defines the schema for the resource.
func DirectoryServiceAuthProviderResourceSchema() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"id": schema.StringAttribute{
			MarkdownDescription: "ID of the Directory Service Auth Provider resource",
			Description:         "ID of the Directory Service Auth Provider resource",
			Computed:            true,
		},
		"active_directory": schema.SingleNestedAttribute{
			MarkdownDescription: "Active Directory" + noteADMessageInclusive + comma + noteMessageExclusive,
			Description:         "Active Directory" + noteADMessageInclusive + comma + noteMessageExclusive,
			Attributes:          ActiveDirectoryResourceSchema(),
			Computed:            true,
			Optional:            true,
			Validators: []validator.Object{
				objectvalidator.AtLeastOneOf(
					path.MatchRoot("ldap")),
				objectvalidator.AlsoRequires(
					path.MatchRoot("active_directory_attributes")),
			},
		},
		"ldap": schema.SingleNestedAttribute{
			MarkdownDescription: "LDAP" + noteLDAPMessageInclusive + comma + noteMessageExclusive,
			Description:         "LDAP" + noteLDAPMessageInclusive + comma + noteMessageExclusive,
			Attributes:          LDAPResourceSchema(),
			Computed:            true,
			Optional:            true,
			Validators: []validator.Object{
				objectvalidator.AtLeastOneOf(
					path.MatchRoot("active_directory")),
				objectvalidator.AlsoRequires(
					path.MatchRoot("ldap_attributes")),
			},
		},
		"active_directory_attributes": schema.MapAttribute{
			MarkdownDescription: "ActiveDirectory.* attributes in Dell iDRAC attributes." + noteADMessageInclusive + comma +
				noteAttributesMessageExclusive,
			Description: "ActiveDirectory.* attributes in Dell iDRAC attributes." + noteADMessageInclusive + comma +
				noteAttributesMessageExclusive,
			ElementType: types.StringType,
			Computed:    true,
			Optional:    true,
			Validators: []validator.Map{
				mapvalidator.AlsoRequires(
					path.MatchRoot("active_directory")),
				mapvalidator.AtLeastOneOf(
					path.MatchRoot("ldap_attributes")),
			},
		},
		"ldap_attributes": schema.MapAttribute{
			MarkdownDescription: "LDAP.* attributes in Dell iDRAC attributes." + noteLDAPMessageInclusive + comma + noteAttributesMessageExclusive,
			Description:         "LDAP.* attributes in Dell iDRAC attributes." + noteLDAPMessageInclusive + comma + noteAttributesMessageExclusive,
			ElementType:         types.StringType,
			Computed:            true,
			Optional:            true,
			Validators: []validator.Map{
				mapvalidator.AlsoRequires(
					path.MatchRoot("ldap")),
				mapvalidator.AtLeastOneOf(
					path.MatchRoot("active_directory_attributes")),
			},
		},
	}
}

// ActiveDirectoryResourceSchema is a function that returns the schema for Active Directory Resource
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

// AuthenticationResourceSchema is a function that returns the schema for Authentication Resource
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

// RemoteRoleMappingResourceSchema is a function that returns the schema for Remote Role Mapping
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

// DirectoryResourceSchema is a function that returns the schema for Directory Resource
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

// LDAPResourceSchema is a function that returns the schema for LDAPResource
func LDAPResourceSchema() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"ldap_service": schema.SingleNestedAttribute{
			MarkdownDescription: "LDAPService is any additional mapping information needed to parse a generic LDAP service.",
			Description:         "LDAPService is any additional mapping information needed to parse a generic LDAP service.",
			Attributes:          LDAPServiceResourceSchema(),
			Computed:            true,
			Optional:            true,
		},
		"directory": schema.SingleNestedAttribute{
			MarkdownDescription: "Directory for LDAP.",
			Description:         "Directory for LDAP",
			Attributes:          DirectoryResourceSchema(),
			Computed:            true,
			Optional:            true,
		},
	}
}

// LDAPServiceResourceSchema is a function that returns the schema for LDAPService Resource
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

// SearchSettingsResourceSchema is a function that returns the schema for SearchSettings Resource
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
