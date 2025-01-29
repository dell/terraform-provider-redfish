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
	"terraform-provider-redfish/redfish/helper"
	"terraform-provider-redfish/redfish/models"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/stmcginnis/gofish"
	"github.com/stmcginnis/gofish/redfish"
)

const (
	ldapService            = "ldap_service"
	baseDistinguishedNames = "base_distinguished_names"
	authentication         = "authentication"
	kerberosKeyTabFile     = "kerberos_key_tab_file"
	localRole              = "local_role"
	remoteGroup            = "remote_group"
	directory              = "directory"
	remoteRoleMapping      = "remote_role_mapping"
	serviceAddresses       = "service_addresses"
	serviceEnabled         = "service_enabled"
	searchSettings         = "search_settings"
)

func getActiveDirectoryModelType() map[string]attr.Type {
	return map[string]attr.Type{
		directory:      types.ObjectType{AttrTypes: getDirectoryModelType()},
		authentication: types.ObjectType{AttrTypes: getAuthentcationModelType()},
	}
}

func getLDAPModelType() map[string]attr.Type {
	return map[string]attr.Type{
		directory:   types.ObjectType{AttrTypes: getDirectoryModelType()},
		ldapService: types.ObjectType{AttrTypes: getLDAPServiceModelType()},
	}
}

func getAuthentcationModelType() map[string]attr.Type {
	return map[string]attr.Type{
		kerberosKeyTabFile: types.StringType,
	}
}

func getDirectoryModelType() map[string]attr.Type {
	return map[string]attr.Type{
		remoteRoleMapping: types.ListType{ElemType: types.ObjectType{AttrTypes: getRemoteRoleMappingModelType()}},
		serviceAddresses:  types.ListType{ElemType: types.StringType},
		serviceEnabled:    types.BoolType,
	}
}

func getLDAPServiceModelType() map[string]attr.Type {
	return map[string]attr.Type{
		searchSettings: types.ObjectType{AttrTypes: getSearchSettingsModelType()},
	}
}

func getSearchSettingsModelType() map[string]attr.Type {
	return map[string]attr.Type{
		baseDistinguishedNames: types.ListType{ElemType: types.StringType},
		"user_name_attribute":  types.StringType,
		"group_name_attribute": types.StringType,
	}
}

func getRemoteRoleMappingModelType() map[string]attr.Type {
	return map[string]attr.Type{
		remoteGroup: types.StringType,
		localRole:   types.StringType,
	}
}

// nolint: revive
func getRemoteRoleMappingObjectValue(ctx context.Context, service []redfish.RoleMapping, state *models.DirectoryResource) ([]attr.Value, diag.Diagnostics) {
	var diags diag.Diagnostics
	remoteRoleMapList := make([]attr.Value, 0)
	var oldRemoteRolMap []models.RemoteRoleMapping
	if !state.RemoteRoleMapping.IsNull() && !state.RemoteRoleMapping.IsUnknown() {
		if diags := state.RemoteRoleMapping.ElementsAs(ctx, &oldRemoteRolMap, true); diags.HasError() {
			diags.AddError("oldRemoteRolMap nil ", "oldRemoteRolMap nil")
			return remoteRoleMapList, diags
		}
	}

	for _, oldElement := range oldRemoteRolMap {
		for _, serviceElement := range service {
			if serviceElement.LocalRole == oldElement.LocalRole.ValueString() && serviceElement.RemoteGroup == oldElement.RemoteGroup.ValueString() {
				remoteRoleItemMap := map[string]attr.Value{
					remoteGroup: types.StringValue(serviceElement.RemoteGroup),
					localRole:   types.StringValue(serviceElement.LocalRole),
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
	if len(remoteRoleMapList) == 0 {
		for _, serviceElement := range service {
			remoteRoleItemMap := map[string]attr.Value{
				remoteGroup: types.StringValue(serviceElement.RemoteGroup),
				localRole:   types.StringValue(serviceElement.LocalRole),
			}
			remoteRoleItemObj, diags := types.ObjectValue(getRemoteRoleMappingModelType(), remoteRoleItemMap)
			if diags.HasError() {
				return remoteRoleMapList, diags
			}
			remoteRoleMapList = append(remoteRoleMapList, remoteRoleItemObj)
		}
	}

	return remoteRoleMapList, diags
}

// nolint: revive
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
	serviceAddress := directoryPlan.ServiceAddresses
	activeDIrectoryRoleMappingList := service.ActiveDirectory.RemoteRoleMapping
	remoteRoleMappingObj, diags := getRemoteRoleMappingObjectValue(ctx, activeDIrectoryRoleMappingList, &directoryPlan)
	if diags.HasError() {
		return emptyObj, diags
	}

	remoteRoleList, diags := types.ListValue(types.ObjectType{AttrTypes: getRemoteRoleMappingModelType()}, remoteRoleMappingObj)

	if diags.HasError() {
		return emptyObj, diags
	}
	serviceAddressList, diags := getConfigDataList(service.ActiveDirectory.ServiceAddresses, serviceAddress)

	if diags.HasError() {
		return emptyObj, diags
	}

	directoryMap := map[string]attr.Value{
		remoteRoleMapping: remoteRoleList,
		serviceAddresses:  serviceAddressList,
		serviceEnabled:    types.BoolValue(service.ActiveDirectory.ServiceEnabled),
	}
	return types.ObjectValue(getDirectoryModelType(), directoryMap)
}

func getConfigDataList(input []string, stateServiceAddress []types.String) (basetypes.ListValue, diag.Diagnostics) {
	out := make([]attr.Value, 0)
	if len(stateServiceAddress) != 0 {
		for _, stateInput := range stateServiceAddress {
			for _, i := range input {
				if stateInput.ValueString() == i {
					out = append(out, types.StringValue(i))
					break
				}
			}
		}
	}

	if len(out) == 0 {
		for _, target := range input {
			out = append(out, types.StringValue(target))
		}
	}

	return types.ListValue(types.StringType, out)
}

// nolint: revive
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

	serviceAddress := directoryPlan.ServiceAddresses
	ldapRemoteRoleList := service.LDAP.RemoteRoleMapping

	remoteRoleMappingObj, diags := getRemoteRoleMappingObjectValue(ctx, ldapRemoteRoleList, &directoryPlan)
	if diags.HasError() {
		return emptyObj, diags
	}

	remoteRoleList, diags := types.ListValue(types.ObjectType{AttrTypes: getRemoteRoleMappingModelType()}, remoteRoleMappingObj)

	if diags.HasError() {
		return emptyObj, diags
	}

	serviceAddressList, diags := getConfigDataList(service.LDAP.ServiceAddresses, serviceAddress)
	if diags.HasError() {
		return emptyObj, diags
	}

	directoryMap := map[string]attr.Value{
		remoteRoleMapping: remoteRoleList,
		serviceAddresses:  serviceAddressList,
		serviceEnabled:    types.BoolValue(service.LDAP.ServiceEnabled),
	}

	return types.ObjectValue(getDirectoryModelType(), directoryMap)
}

// nolint: revive
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
	}

	ldapServiceMap := map[string]attr.Value{
		searchSettings: ldapSearchSetting,
	}

	return types.ObjectValue(getLDAPServiceModelType(), ldapServiceMap)
}

// nolint: revive
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

	baseDistinguished := oldSearchSettings.BaseDistinguishedNames
	baseDistinguishedList, diags := getConfigDataList(service.LDAP.LDAPService.SearchSettings.BaseDistinguishedNames, baseDistinguished)
	if diags.HasError() {
		return emptyObj, diags
	}

	searchSettingsMap := map[string]attr.Value{
		baseDistinguishedNames: baseDistinguishedList,
		"user_name_attribute":  types.StringValue(service.LDAP.LDAPService.SearchSettings.UsernameAttribute),
		"group_name_attribute": types.StringValue(service.LDAP.LDAPService.SearchSettings.GroupNameAttribute),
	}

	return types.ObjectValue(getSearchSettingsModelType(), searchSettingsMap)
}

// nolint: revive
func getAuthentcationObjectValue(ctx context.Context, service *redfish.AccountService, directoryPlan models.ActiveDirectoryResource, objectAsOptions basetypes.ObjectAsOptions) (basetypes.ObjectValue, diag.Diagnostics) {
	var authenticationPlan models.AuthenticationResource
	var authentication map[string]attr.Value
	emptyObj := types.ObjectNull(getAuthentcationModelType())
	if !directoryPlan.Authentication.IsNull() && !directoryPlan.Authentication.IsUnknown() {
		diags := directoryPlan.Authentication.As(ctx, &authenticationPlan, objectAsOptions)
		if diags.HasError() {
			return emptyObj, diags
		}
		authentication = map[string]attr.Value{
			kerberosKeyTabFile: types.StringValue(authenticationPlan.KerberosKeytab.ValueString()),
		}
	}

	if directoryPlan.Authentication.IsNull() || directoryPlan.Authentication.IsUnknown() {
		authentication = map[string]attr.Value{
			kerberosKeyTabFile: types.StringValue(service.ActiveDirectory.Authentication.KerberosKeytab),
		}
	}
	return types.ObjectValue(getAuthentcationModelType(), authentication)
}

// nolint: gocyclo, revive
func parseActiveDirectoryIntoState(ctx context.Context, acctService *redfish.AccountService, service *gofish.Service, state *models.DirectoryServiceAuthProviderResource) (diags diag.Diagnostics) {
	objectAsOptions := basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true}
	var oldActiveDirectory models.ActiveDirectoryResource
	if !state.ActiveDirectoryResource.IsNull() && !state.ActiveDirectoryResource.IsUnknown() {
		diags = state.ActiveDirectoryResource.As(ctx, &oldActiveDirectory, objectAsOptions)
		if diags.HasError() {
			return diags
		}

	}
	directoryObj, diags := getActiveDirectoryObjectValue(ctx, acctService, state, objectAsOptions)
	if diags.HasError() {
		return diags
	}

	authenticationObj, diags := getAuthentcationObjectValue(ctx, acctService, oldActiveDirectory, objectAsOptions)
	if diags.HasError() {
		return diags
	}
	var idracAttributesPlan models.DellIdracAttributes
	if !state.ActiveDirectoryAttributes.IsNull() && !state.ActiveDirectoryAttributes.IsUnknown() {
		idracAttributesPlan.Attributes = state.ActiveDirectoryAttributes
	}
	if diags := readRedfishDellIdracAttributes(ctx, service, &idracAttributesPlan); diags.HasError() {
		return diags
	}
	var activeDirAttributes types.Map
	if !state.ActiveDirectoryAttributes.IsNull() && !state.ActiveDirectoryAttributes.IsUnknown() {
		activeDirAttributes = state.ActiveDirectoryAttributes
	}

	if state.ActiveDirectoryAttributes.IsNull() || state.ActiveDirectoryAttributes.IsUnknown() {
		// nolint: gocyclo, gocognit,revive
		activeDirectoryAttributes := []string{".CertValidationEnable", ".SSOEnable", ".AuthTimeout", ".DCLookupEnable", ".DCLookupByUserDomain", ".DCLookupDomainName", ".Schema", ".GCLookupEnable", ".GCRootDomain", ".GlobalCatalog1", ".GlobalCatalog2", ".GlobalCatalog3", ".RacName", ".RacDomain"}

		attributesToReturn := make(map[string]attr.Value)
		for k, v := range idracAttributesPlan.Attributes.Elements() {
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

		activeDirAttributes = types.MapValueMust(types.StringType, attributesToReturn)
	}
	activeDirectoryMap := map[string]attr.Value{
		directory:      directoryObj,
		authentication: authenticationObj,
	}
	state.ActiveDirectoryAttributes = activeDirAttributes
	state.ActiveDirectoryResource, diags = types.ObjectValue(getActiveDirectoryModelType(), activeDirectoryMap)

	return diags
}

// nolint:gocyclo, revive
func parseLDAPIntoState(ctx context.Context, acctService *redfish.AccountService, service *gofish.Service, state *models.DirectoryServiceAuthProviderResource) (diags diag.Diagnostics) {
	objectAsOptions := basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true}
	var oldLDAP models.LDAPResource
	if !state.LDAPResource.IsNull() && !state.LDAPResource.IsUnknown() {
		if diags := state.LDAPResource.As(ctx, &oldLDAP, objectAsOptions); diags.HasError() {
			return diags
		}
	}

	directoryObj, diags := getLDAPDirectoryObjectValue(ctx, acctService, state, objectAsOptions)
	if diags.HasError() {
		return diags
	}

	ldapServiceObj, diags := getLDAPServiceObjectValue(ctx, acctService, state, objectAsOptions)

	if diags.HasError() {
		return diags
	}
	var idracAttributesPlan models.DellIdracAttributes
	if !state.LDAPAttributes.IsNull() && !state.LDAPAttributes.IsUnknown() {
		idracAttributesPlan.Attributes = state.LDAPAttributes
	}

	if diags = readRedfishDellIdracAttributes(ctx, service, &idracAttributesPlan); diags.HasError() {
		return diags
	}

	var ldapDirAttributes types.Map
	if !state.LDAPAttributes.IsNull() && !state.LDAPAttributes.IsUnknown() {
		ldapDirAttributes = state.LDAPAttributes
	}

	if state.LDAPAttributes.IsNull() || state.LDAPAttributes.IsUnknown() {
		// nolint: gocyclo, gocognit,revive
		ldapAttributes := []string{".CertValidationEnable", ".GroupAttributeIsDN", ".Port", ".BindDN", ".BindPassword", ".SearchFilter"}
		attributesToReturn := make(map[string]attr.Value)
		for k, v := range idracAttributesPlan.Attributes.Elements() {
			if strings.HasPrefix(k, "LDAP.") {
				for _, input := range ldapAttributes {
					if strings.HasSuffix(k, input) {
						attributesToReturn[k] = v
					}
				}
			}
		}
		ldapDirAttributes = types.MapValueMust(types.StringType, attributesToReturn)
	}
	ldapMap := map[string]attr.Value{
		directory:   directoryObj,
		ldapService: ldapServiceObj,
	}
	state.LDAPResource, diags = types.ObjectValue(getLDAPModelType(), ldapMap)
	state.LDAPAttributes = ldapDirAttributes
	return diags
}

// nolint: gocyclo, revive
func getActiveDirectoryPatchBody(ctx context.Context, attrsState *models.DirectoryServiceAuthProviderResource) (map[string]interface{}, diag.Diagnostics) {
	// var diags diag.Diagnostics
	supportedActiveDirectory := map[string]string{
		serviceEnabled:    "ServiceEnabled",
		serviceAddresses:  "ServiceAddresses",
		remoteRoleMapping: "RemoteRoleMapping",
		authentication:    "Authentication",
	}
	supportedRemoteRoleMappingParams := map[string]string{
		remoteGroup: "RemoteGroup",
		localRole:   "LocalRole",
	}

	supportedAuthentication := map[string]string{
		kerberosKeyTabFile: "KerberosKeytab",
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
				goValue, err := helper.ConvertTerraformValueToGoBasicValue(ctx, value)
				if err != nil {
					tflog.Trace(ctx, fmt.Sprintf("Failed to convert AD directory value to go value: %s", err.Error()))
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
		var remoteRoleMappingModel []models.RemoteRoleMapping

		if diags := directoryPlan.RemoteRoleMapping.ElementsAs(ctx, &remoteRoleMappingModel, true); diags.HasError() {
			return nil, diags
		}

		remoteRoleMappingList := make([]interface{}, 0)
		for _, target := range remoteRoleMappingModel {
			remoteRoleMappingBody := make(map[string]interface{})
			if !target.LocalRole.IsNull() && !target.LocalRole.IsUnknown() {
				remoteRoleMappingBody[supportedRemoteRoleMappingParams[localRole]] = target.LocalRole.ValueString()
			}
			if !target.RemoteGroup.IsNull() && !target.RemoteGroup.IsUnknown() {
				remoteRoleMappingBody[supportedRemoteRoleMappingParams[remoteGroup]] = target.RemoteGroup.ValueString()
			}
			if len(remoteRoleMappingBody) > 0 {
				remoteRoleMappingList = append(remoteRoleMappingList, remoteRoleMappingBody)
			}
		}

		patchBody[supportedActiveDirectory[remoteRoleMapping]] = remoteRoleMappingList
	}

	serviceAddress := directoryPlan.ServiceAddresses
	serviceAddressList := make([]interface{}, 0)
	for _, target := range serviceAddress {
		serviceAddressList = append(serviceAddressList, target.ValueString())
	}
	if len(serviceAddressList) != 0 {
		patchBody[supportedActiveDirectory[serviceAddresses]] = serviceAddressList
	}

	// get directory patch body
	if !activeDirectoryPlan.Authentication.IsNull() && !activeDirectoryPlan.Authentication.IsUnknown() {
		authenticationPatchBody := make(map[string]interface{})
		for key, value := range activeDirectoryPlan.Authentication.Attributes() {
			if !value.IsUnknown() && !value.IsNull() {
				goValue, err := helper.ConvertTerraformValueToGoBasicValue(ctx, value)
				if err != nil {
					tflog.Trace(ctx, fmt.Sprintf("Failed to convert AD authentication value to go value: %s", err.Error()))
					continue
				}
				if fieldName, ok := supportedAuthentication[key]; ok {
					authenticationPatchBody[fieldName] = goValue
				}
			}
		}
		patchBody[supportedActiveDirectory[authentication]] = authenticationPatchBody
	}

	return patchBody, nil
}

// nolint: gocyclo, revive
func getLDAPPatchBody(ctx context.Context, attrsState *models.DirectoryServiceAuthProviderResource) (map[string]interface{}, diag.Diagnostics) {
	// var diags diag.Diagnostics
	supportedLDAP := map[string]string{
		serviceEnabled:    "ServiceEnabled",
		serviceAddresses:  "ServiceAddresses",
		remoteRoleMapping: "RemoteRoleMapping",
		ldapService:       "LDAPService",
	}
	supportedRemoteRoleMappingParams := map[string]string{
		remoteGroup: "RemoteGroup",
		localRole:   "LocalRole",
	}

	supportedLDAPService := map[string]string{
		searchSettings: "SearchSettings",
	}

	supportedSearchSetting := map[string]string{
		baseDistinguishedNames: "BaseDistinguishedNames",
		"user_name_attribute":  "UsernameAttribute",
		"group_name_attribute": "GroupNameAttribute",
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
	patchBody := make(map[string]interface{})
	if !ldapPlan.Directory.IsNull() && !ldapPlan.Directory.IsUnknown() {
		for key, value := range ldapPlan.Directory.Attributes() {
			if !value.IsUnknown() && !value.IsNull() {
				goValue, err := helper.ConvertTerraformValueToGoBasicValue(ctx, value)
				if err != nil {
					tflog.Trace(ctx, fmt.Sprintf("Failed to convert LDAP Directory value to go value: %s", err.Error()))
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
							goValue, err := helper.ConvertTerraformValueToGoBasicValue(ctx, value)
							if err != nil {
								tflog.Trace(ctx, fmt.Sprintf("Failed to convert LDAP SearchSettings value to go value: %s", err.Error()))
								continue
							}
							if fieldName, ok := supportedSearchSetting[key]; ok {
								ldapSearchSettingPatchBody[fieldName] = goValue
							}
						}
					}
					baseDiss := ldapSearchSettingsPlan.BaseDistinguishedNames
					baseDissList := make([]interface{}, 0)
					for _, target := range baseDiss {
						baseDissList = append(baseDissList, target.ValueString())
					}
					if len(baseDissList) != 0 {
						ldapSearchSettingPatchBody[supportedSearchSetting[baseDistinguishedNames]] = baseDissList
					}

					if fieldName, ok := supportedLDAPService[key1]; ok {
						ldapServicepatchBody[fieldName] = ldapSearchSettingPatchBody
					}
				}
			}
		}
		patchBody[supportedLDAP[ldapService]] = ldapServicepatchBody
	}

	// get list of remote role mapping
	if !directoryPlan.RemoteRoleMapping.IsNull() && !directoryPlan.RemoteRoleMapping.IsUnknown() {
		var remoteRoleMappingModel []models.RemoteRoleMapping

		if diags := directoryPlan.RemoteRoleMapping.ElementsAs(ctx, &remoteRoleMappingModel, true); diags.HasError() {
			return nil, diags
		}

		remoteRoleMappingList := make([]interface{}, 0)
		for _, target := range remoteRoleMappingModel {
			remoteRoleMappingBody := make(map[string]interface{})
			if !target.LocalRole.IsNull() && !target.LocalRole.IsUnknown() {
				remoteRoleMappingBody[supportedRemoteRoleMappingParams[localRole]] = target.LocalRole.ValueString()
			}
			if !target.RemoteGroup.IsNull() && !target.RemoteGroup.IsUnknown() {
				remoteRoleMappingBody[supportedRemoteRoleMappingParams[remoteGroup]] = target.RemoteGroup.ValueString()
			}
			if len(remoteRoleMappingBody) > 0 {
				remoteRoleMappingList = append(remoteRoleMappingList, remoteRoleMappingBody)
			}
		}

		patchBody[supportedLDAP[remoteRoleMapping]] = remoteRoleMappingList
	}

	serviceAddress := directoryPlan.ServiceAddresses
	serviceAddressList := make([]interface{}, 0)
	for _, target := range serviceAddress {
		serviceAddressList = append(serviceAddressList, target.ValueString())
	}
	if len(serviceAddressList) != 0 {
		patchBody[supportedLDAP[serviceAddresses]] = serviceAddressList
	}
	return patchBody, nil
}
