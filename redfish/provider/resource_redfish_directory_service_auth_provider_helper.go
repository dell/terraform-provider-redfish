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
	"strconv"
	"strings"
	"terraform-provider-redfish/redfish/models"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

const (
	// RSASecurID2FA is rsa Secure Id 2 factor authentication
	RSASecurID2FA = "RSASecurID2FA"
	// Disabled disable the service
	Disabled = "Disabled"
	// Enabled enable the service
	Enabled = "Enabled"

	lowerLimit = 15
	upperLimit = 300
)

// nolint: gocyclo,revive
func newActiveDirectoryChanged(ctx context.Context, plan, state *models.DirectoryServiceAuthProviderResource) bool {
	objectAsOptions := basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true}

	if plan.ActiveDirectoryResource.IsNull() || plan.ActiveDirectoryResource.IsUnknown() {
		return false
	}
	var activeDirectoryPlan models.ActiveDirectoryResource
	if !plan.ActiveDirectoryResource.IsNull() && !plan.ActiveDirectoryResource.IsUnknown() {
		if diags := plan.ActiveDirectoryResource.As(ctx, &activeDirectoryPlan, objectAsOptions); diags.HasError() {
			return false
		}
	}

	var activeDirectoryState models.ActiveDirectoryResource
	if !state.ActiveDirectoryResource.IsNull() && !state.ActiveDirectoryResource.IsUnknown() {
		if diags := state.ActiveDirectoryResource.As(ctx, &activeDirectoryState, objectAsOptions); diags.HasError() {
			return false
		}
	}

	var directoryPlan models.DirectoryResource
	if !activeDirectoryPlan.Directory.IsNull() && !activeDirectoryPlan.Directory.IsUnknown() {
		if diags := activeDirectoryPlan.Directory.As(ctx, &directoryPlan, objectAsOptions); diags.HasError() {
			return false
		}
	}

	var directoryState models.DirectoryResource
	if !activeDirectoryState.Directory.IsNull() && !activeDirectoryState.Directory.IsUnknown() {
		if diags := activeDirectoryState.Directory.As(ctx, &directoryState, objectAsOptions); diags.HasError() {
			return false
		}
	}

	if directoryPlan.RemoteRoleMapping.String() != "" && directoryPlan.RemoteRoleMapping.String() != directoryState.RemoteRoleMapping.String() {
		return true
	}

	if len(directoryPlan.ServiceAddresses) != len(directoryState.ServiceAddresses) {
		return true
	}

	if checkListMatching(directoryPlan.ServiceAddresses, directoryState.ServiceAddresses) {
		return true
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

	return newAttributesChanged(plan.ActiveDirectoryAttributes, state.ActiveDirectoryAttributes)
}

// nolint: gocyclo,revive
func newLDAPChanged(ctx context.Context, plan, state *models.DirectoryServiceAuthProviderResource) bool {
	objectAsOptions := basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true}

	if plan.LDAPResource.IsNull() || plan.LDAPResource.IsUnknown() {
		return false
	}
	var ldapPlan models.LDAPResource
	if !plan.LDAPResource.IsNull() && !plan.LDAPResource.IsUnknown() {
		if diags := plan.LDAPResource.As(ctx, &ldapPlan, objectAsOptions); diags.HasError() {
			return false
		}
	}
	var ldapState models.LDAPResource
	if !state.LDAPResource.IsNull() && !state.LDAPResource.IsUnknown() {
		if diags := state.LDAPResource.As(ctx, &ldapState, objectAsOptions); diags.HasError() {
			return false
		}
	}
	var directoryPlan models.DirectoryResource
	if !ldapPlan.Directory.IsNull() && !ldapPlan.Directory.IsUnknown() {
		if diags := ldapPlan.Directory.As(ctx, &directoryPlan, objectAsOptions); diags.HasError() {
			return false
		}
	}

	var directoryState models.DirectoryResource
	if !ldapState.Directory.IsNull() && !ldapState.Directory.IsUnknown() {
		if diags := ldapState.Directory.As(ctx, &directoryState, objectAsOptions); diags.HasError() {
			return false
		}
	}

	if directoryPlan.RemoteRoleMapping.String() != "" && directoryPlan.RemoteRoleMapping.String() != directoryState.RemoteRoleMapping.String() {
		return true
	}

	if len(directoryPlan.ServiceAddresses) != len(directoryState.ServiceAddresses) {
		return true
	}

	if checkListMatching(directoryPlan.ServiceAddresses, directoryState.ServiceAddresses) {
		return true
	}

	var ldapServicePlan models.LDAPServiceResource
	if !ldapPlan.LDAPService.IsNull() && !ldapPlan.LDAPService.IsUnknown() {
		if diags := ldapPlan.LDAPService.As(ctx, &ldapServicePlan, objectAsOptions); diags.HasError() {
			return false
		}
	}

	var ldapServiceState models.LDAPServiceResource
	if !ldapState.LDAPService.IsNull() && !ldapState.LDAPService.IsUnknown() {
		if diags := ldapState.LDAPService.As(ctx, &ldapServiceState, objectAsOptions); diags.HasError() {
			return false
		}
	}

	var ldapSearchSettingsPlan models.SearchSettingsResource
	if !ldapServicePlan.SearchSettings.IsNull() && !ldapServicePlan.SearchSettings.IsUnknown() {
		if diags := ldapServicePlan.SearchSettings.As(ctx, &ldapSearchSettingsPlan, objectAsOptions); diags.HasError() {
			return false
		}
	}

	var ldapSearchSettingsState models.SearchSettingsResource
	if !ldapServiceState.SearchSettings.IsNull() && !ldapServiceState.SearchSettings.IsUnknown() {
		if diags := ldapServiceState.SearchSettings.As(ctx, &ldapSearchSettingsState, objectAsOptions); diags.HasError() {
			return false
		}
	}

	if ldapSearchSettingsPlan.GroupNameAttribute.ValueString() != "" &&
		ldapSearchSettingsPlan.GroupNameAttribute.ValueString() != ldapSearchSettingsState.GroupNameAttribute.ValueString() {
		return true
	}
	if ldapSearchSettingsPlan.UsernameAttribute.ValueString() != "" &&
		ldapSearchSettingsPlan.UsernameAttribute.ValueString() != ldapSearchSettingsState.UsernameAttribute.ValueString() {
		return true
	}

	if len(ldapSearchSettingsPlan.BaseDistinguishedNames) != len(ldapSearchSettingsState.BaseDistinguishedNames) {
		return true
	}

	if checkListMatching(ldapSearchSettingsPlan.BaseDistinguishedNames, ldapSearchSettingsState.BaseDistinguishedNames) {
		return true
	}

	return newAttributesChanged(plan.LDAPAttributes, state.LDAPAttributes)
}

func newAttributesChanged(attrsPlan, attrsState types.Map) bool {
	if attrsPlan.IsNull() || attrsPlan.IsUnknown() {
		return false
	}

	if attrsPlan.String() != "" && attrsPlan.String() == attrsState.String() {
		return false
	}
	return true
}

func checkListMatching(planList []types.String, stateList []types.String) bool {
	changed := false
	for _, planTarget := range planList {
		changed = true
		for _, stateTarget := range stateList {
			if planTarget.String() == stateTarget.String() {
				changed = false
				break
			}
		}
	}
	return changed
}

func checkAttributeskeyPresent(attributes types.Map, prefix string, suffix string) bool {
	for k := range attributes.Elements() {
		if strings.HasPrefix(k, prefix) && strings.HasSuffix(k, suffix) {
			return true
		}
	}
	return false
}

func getkAttributeskeyValue(attributes types.Map, prefix string, suffix string) string {
	for k, v := range attributes.Elements() {
		if strings.HasPrefix(k, prefix) && strings.HasSuffix(k, suffix) {
			value := v.String()
			value = value[1 : len(value)-1]
			return value
		}
	}
	return ""
}

func isValid2FactorAuth(attributes types.Map) (bool, diag.Diagnostics) {
	var diags diag.Diagnostics
	// attributes := attrsState.ActiveDirectoryAttributes
	checkey2FA := checkAttributeskeyPresent(attributes, RSASecurID2FA, "RSASecurIDAccessKey")
	checkID2FA := checkAttributeskeyPresent(attributes, RSASecurID2FA, "RSASecurIDClientID")
	checkServer2FA := checkAttributeskeyPresent(attributes, RSASecurID2FA, "RSASecurIDAuthenticationServer")

	if checkey2FA || checkID2FA || checkServer2FA {
		checkey2FAValue := getkAttributeskeyValue(attributes, RSASecurID2FA, "RSASecurIDAccessKey")
		checID2FAValue := getkAttributeskeyValue(attributes, RSASecurID2FA, "RSASecurIDClientID")
		checkServer2FAValue := getkAttributeskeyValue(attributes, RSASecurID2FA, "RSASecurIDAuthenticationServer")

		if checkey2FAValue == "" || checID2FAValue == "" || checkServer2FAValue == "" {
			diags.AddError("Missing RSASecurID2FA required params", "Please provide all the required configuration for 2 factor autentication")
			return false, diags
		}
	}

	return true, diags
}

func isValidAuthTime(prefix string, suffix string, attrsState *models.DirectoryServiceAuthProviderResource) (bool, diag.Diagnostics) {
	var diags diag.Diagnostics
	attributes := attrsState.ActiveDirectoryAttributes
	check := checkAttributeskeyPresent(attributes, prefix, suffix)

	if !check {
		diags.AddError("Invalid AuthTimeout, Please provide all the required configuration", "Please provide all the required configuration")
		return false, diags
	}
	value := getkAttributeskeyValue(attributes, prefix, suffix)

	intValue, err := strconv.Atoi(value)
	if err != nil {
		diags.AddError("Invalid AuthTimeout", "Invalid AuthTimeout")
		return false, diags
	}

	if intValue < lowerLimit || intValue > upperLimit {
		diags.AddError("Invalid AuthTimeout, AuthTimeout must be between 15 and 300", "AuthTimeout must be between 15 and 300")
		return false, diags
	}

	return true, nil
}

// nolint: revive
func isSSOEnabledWithValidFile(ctx context.Context, prefix string, suffix string, attrsState *models.DirectoryServiceAuthProviderResource) (bool, diag.Diagnostics) {
	var diags diag.Diagnostics

	attributes := attrsState.ActiveDirectoryAttributes
	check := checkAttributeskeyPresent(attributes, prefix, suffix)

	if check {
		value := getkAttributeskeyValue(attributes, prefix, suffix)
		if value == Enabled {
			objectAsOptions := basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true}
			var activeDirectoryPlan models.ActiveDirectoryResource
			if !attrsState.ActiveDirectoryResource.IsNull() && !attrsState.ActiveDirectoryResource.IsUnknown() {
				if diags := attrsState.ActiveDirectoryResource.As(ctx, &activeDirectoryPlan, objectAsOptions); diags.HasError() {
					return false, diags
				}
			}

			var directoryPlan models.DirectoryResource
			if !activeDirectoryPlan.Directory.IsNull() && !activeDirectoryPlan.Directory.IsUnknown() {
				if diags := activeDirectoryPlan.Directory.As(ctx, &directoryPlan, objectAsOptions); diags.HasError() {
					return false, diags
				}
			}

			if !directoryPlan.ServiceEnabled.ValueBool() {
				diags.AddError("Please provide valid Configuration for SSO", "SSO can't be enabled when Active Directory service is Disabled")
				return false, diags
			}
			var authenticationPlan models.AuthenticationResource
			if !activeDirectoryPlan.Authentication.IsNull() && !activeDirectoryPlan.Authentication.IsUnknown() {
				if diags := activeDirectoryPlan.Authentication.As(ctx, &authenticationPlan, objectAsOptions); diags.HasError() {
					return false, diags
				}
			}

			if activeDirectoryPlan.Authentication.IsNull() || activeDirectoryPlan.Authentication.IsUnknown() ||
				authenticationPlan.KerberosKeytab.IsNull() || authenticationPlan.KerberosKeytab.IsUnknown() {
				diags.AddError("Please provide valid kerberos key tab file when SSO is enabled",
					"Please provide valid kerberos key tab file when SSO is enabled")
				return false, diags
			}

			return true, diags
		}

		if value == Disabled {
			diags.AddWarning("Disabled ", "inside Disabled")
			return true, diags
		}
	}

	return true, diags
}

// nolint: gocyclo, revive
func isValidSchemaSelection(ctx context.Context, prefix string, suffix string, attrsState *models.DirectoryServiceAuthProviderResource) (bool, diag.Diagnostics) {
	var diags diag.Diagnostics
	attributes := attrsState.ActiveDirectoryAttributes
	check := checkAttributeskeyPresent(attributes, prefix, suffix)
	if check {
		value := getkAttributeskeyValue(attributes, prefix, suffix)
		iDRAcName := checkAttributeskeyPresent(attributes, ActiveDirectory, "RacName")
		iDRAcDomain := checkAttributeskeyPresent(attributes, ActiveDirectory, "RacDomain")
		groupDomain := checkAttributeskeyPresent(attributes, "ADGroup", "Domain")
		gcLookupEnable := checkAttributeskeyPresent(attributes, ActiveDirectory, "GCLookupEnable")
		gc1 := checkAttributeskeyPresent(attributes, ActiveDirectory, "GlobalCatalog1")
		gc2 := checkAttributeskeyPresent(attributes, ActiveDirectory, "GlobalCatalog2")
		gc3 := checkAttributeskeyPresent(attributes, ActiveDirectory, "GlobalCatalog3")
		objectAsOptions := basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true}
		var activeDirectoryPlan models.ActiveDirectoryResource
		if !attrsState.ActiveDirectoryResource.IsNull() && !attrsState.ActiveDirectoryResource.IsUnknown() {
			if diags := attrsState.ActiveDirectoryResource.As(ctx, &activeDirectoryPlan, objectAsOptions); diags.HasError() {
				return false, diags
			}
		}

		var directoryPlan models.DirectoryResource
		if !activeDirectoryPlan.Directory.IsNull() && !activeDirectoryPlan.Directory.IsUnknown() {
			if diags := activeDirectoryPlan.Directory.As(ctx, &directoryPlan, objectAsOptions); diags.HasError() {
				return false, diags
			}
		}

		var remoteRoleMapping []models.RemoteRoleMapping
		if diags := directoryPlan.RemoteRoleMapping.ElementsAs(ctx, &remoteRoleMapping, true); diags.HasError() {
			return false, diags
		}

		if value == "Extended Schema" {
			if !iDRAcName || !iDRAcDomain {
				diags.AddError("Please provide valid config for RacName and RacDomain",
					"RacName and RacDomain must be configured for Extended Schema")
				return false, diags
			}
			iDRAcNameValue := getkAttributeskeyValue(attributes, ActiveDirectory, "RacName")
			iDRAcDomainValue := getkAttributeskeyValue(attributes, ActiveDirectory, "RacDomain")
			if iDRAcNameValue == "" || iDRAcDomainValue == "" {
				diags.AddError("RacName and RacDomain can not be null for Extended Schema",
					"Please provide valid configuration for RacName and RacDomain")
				return false, diags
			}
			if gcLookupEnable || gc1 || gc2 || gc3 {
				diags.AddError("GCLookupEnable, GlobalCatalog1, GlobalCatalog2, GlobalCatalog3 can not be configured for Extended Schema",
					"GCLookupEnable, GlobalCatalog1, GlobalCatalog2, GlobalCatalog3 can not be configured for Extended Schema")
				return false, diags
			}

			if !activeDirectoryPlan.Directory.IsNull() && !activeDirectoryPlan.Directory.IsUnknown() {
				if len(remoteRoleMapping) != 0 {
					diags.AddError("RemoteRoleMapping can not be configured for Extended Schema",
						"RemoteRoleMapping can not be configured for Extended Schema")
					return false, diags

				}
			}

			if groupDomain {
				diags.AddError("Domain can not be configured for Extended Schema", "Domain can not be configured for Extended Schema")
				return false, diags
			}
		}
		if value == "Standard Schema" {
			if iDRAcName || iDRAcDomain {
				diags.AddError("RacName and RacDomain can not be configured for Standard Schema",
					"RacName and RacDomain can not be configured for Standard Schema")
				return false, diags
			}
			if !gcLookupEnable {
				diags.AddError("GCLookupEnable must be configured for Standard Schema", "GCLookupEnable must be configured for Standard Schema")
				return false, diags
			}
			gcLookupEnableValue := getkAttributeskeyValue(attributes, ActiveDirectory, "GCLookupEnable")

			if gcLookupEnableValue == Enabled {
				gcRootDomain := checkAttributeskeyPresent(attributes, ActiveDirectory, "GCRootDomain")
				if !gcRootDomain || getkAttributeskeyValue(attributes, ActiveDirectory, "GCRootDomain") == "" {
					diags.AddError("GCRootDomain must be configured for Enabled GCLookupEnable",
						"GCRootDomain must be configured for Enabled GCLookupEnable")
					return false, diags
				}

				if gc1 || gc2 || gc3 {
					diags.AddError("GlobalCatalog can not be configured for Enabled GCLookupEnable",
						" GlobalCatalog can not be configured for Enabled GCLookupEnable")
					return false, diags
				}
			} else if gcLookupEnableValue == Disabled {
				gc1Value := getkAttributeskeyValue(attributes, ActiveDirectory, "GlobalCatalog1")
				gc2Value := getkAttributeskeyValue(attributes, ActiveDirectory, "GlobalCatalog2")
				gc3Value := getkAttributeskeyValue(attributes, ActiveDirectory, "GlobalCatalog3")

				if gc1Value == "" && gc2Value == "" && gc3Value == "" {
					diags.AddError("Invalid GlobalCatalog configuration for Standard Schema",
						"Atleast any one from GlobalCatalog1, GlobalCatalog2, GlobalCatalog3 must be configured for Disabled GCLookupEnable")
					return false, diags
				}

				gcRootDomain := checkAttributeskeyPresent(attributes, ActiveDirectory, "GCRootDomain")
				if gcRootDomain {
					diags.AddError("GCRootDomain can not be configured for Disabled GCLookupEnable",
						"GCRootDomain can not be configured for Disabled GCLookupEnable")
					return false, diags
				}
			} else {
				diags.AddError("Invalid configuration for Standard Schema", "Please provide valid configuration for Standard Schema")
				return false, diags
			}
		}
	}
	return true, diags
}

// nolint: gocyclo, revive
func isValidDCLookupDomainConfig(ctx context.Context, prefix string, suffix string, attrsState *models.DirectoryServiceAuthProviderResource) (bool, diag.Diagnostics) {
	var diags diag.Diagnostics
	attributes := attrsState.ActiveDirectoryAttributes
	check := checkAttributeskeyPresent(attributes, prefix, suffix)
	if check {
		objectAsOptions := basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true}
		var activeDirectoryPlan models.ActiveDirectoryResource
		if !attrsState.ActiveDirectoryResource.IsNull() && !attrsState.ActiveDirectoryResource.IsUnknown() {
			if diags := attrsState.ActiveDirectoryResource.As(ctx, &activeDirectoryPlan, objectAsOptions); diags.HasError() {
				return false, diags
			}
		}

		var directoryPlan models.DirectoryResource
		if !activeDirectoryPlan.Directory.IsNull() && !activeDirectoryPlan.Directory.IsUnknown() {
			if diags := activeDirectoryPlan.Directory.As(ctx, &directoryPlan, objectAsOptions); diags.HasError() {
				return false, diags
			}
		}
		dcLookupEnableValue := getkAttributeskeyValue(attributes, ActiveDirectory, "DCLookupEnable")
		serviceAddress := directoryPlan.ServiceAddresses

		serviceAddressList := make([]interface{}, 0)
		for _, target := range serviceAddress {
			serviceAddressList = append(serviceAddressList, target.ValueString())
		}
		userDomain := checkAttributeskeyPresent(attributes, ActiveDirectory, "DCLookupByUserDomain")
		specifyDomain := checkAttributeskeyPresent(attributes, ActiveDirectory, "DCLookupDomainName")
		if dcLookupEnableValue == Disabled {
			// diags.AddError("Invalid configuration for DCLookUp", "Service address must be configured for DCLookUp"+strconv.Itoa(len(serviceAddressList)))
			if len(serviceAddressList) == 0 {
				diags.AddError("ServiceAddresses is not Configured for Disabled DCLookUp",
					"Atleast one Service address must be configured for Disabled DCLookUp")
				return false, diags
			}
			if userDomain {
				diags.AddError("DCLookupByUserDomain can not be Configured for Disabled DCLookUp",
					"DCLookupByUserDomain can not be Configured for Disabled DCLookUp")
				return false, diags
			}
			if specifyDomain {
				diags.AddError("DCLookupDomainName can not be configured for Disabled DCLookUp",
					"DCLookupDomainName can not be configured for Disabled DCLookUp")
				return false, diags
			}
		} else if dcLookupEnableValue == Enabled {
			if len(serviceAddressList) != 0 {
				diags.AddError("Service address can not be configured for Enabled DCLookUp", "Service address can not be configured for Enabled DCLookUp")
				return false, diags
			}
			if !userDomain {
				diags.AddError("DCLookupByUserDomain must be configured for Enabled DCLookUp",
					"DCLookupByUserDomain must be configured for Enabled DCLookUp")
				return false, diags
			}
			userDomainValue := getkAttributeskeyValue(attributes, ActiveDirectory, "DCLookupByUserDomain")
			if userDomainValue == Disabled {
				if !specifyDomain {
					diags.AddError("DCLookupDomainName must be configured for Disabled DCLookupByUserDomain",
						"DCLookupDomainName must be configured for Disabled DCLookupByUserDomain")
					return false, diags
				}
				specifyDomainValue := getkAttributeskeyValue(attributes, ActiveDirectory, "DCLookupDomainName")
				if specifyDomainValue == "" {
					diags.AddError("DCLookupDomainName must be configured for Disabled DCLookupByUserDomain",
						"DCLookupDomainName must be configured for Disabled DCLookupByUserDomain")
					return false, diags
				}
			} else if userDomainValue == Enabled {
				specifyDomain := checkAttributeskeyPresent(attributes, ActiveDirectory, "DCLookupDomainName")
				if specifyDomain {
					diags.AddError("DCLookupDomainName can not be configured for Enabled DCLookupByUserDomain",
						"DCLookupDomainName can not be configured for Enabled DCLookupByUserDomain")
					return false, diags
				}
			}
			return true, diags

		} else {
			diags.AddError("Invalid configuration for DCLookUp", "Please provide valid configuration for DCLookUp")
			return false, diags
		}
	}
	return true, diags
}
