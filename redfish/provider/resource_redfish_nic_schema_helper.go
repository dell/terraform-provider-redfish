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

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/stmcginnis/gofish/redfish"
)

func convertTerraformValueToGoBasicValue(ctx context.Context, v attr.Value) (interface{}, error) {
	vTypeStr := v.Type(ctx).String()
	if strings.HasPrefix(vTypeStr, "basetypes.StringType") {
		if des, ok := v.(basetypes.StringValue); ok {
			return des.ValueString(), nil
		}
	}
	if strings.HasPrefix(vTypeStr, "basetypes.Int64Type") {
		if des, ok := v.(basetypes.Int64Value); ok {
			return int(des.ValueInt64()), nil
		}
	}
	if strings.HasPrefix(vTypeStr, "basetypes.BoolType") {
		if des, ok := v.(basetypes.BoolValue); ok {
			return des.ValueBool(), nil
		}
	}

	return nil, fmt.Errorf("unsupported type: %s", vTypeStr)
}

func getOemNetworkAttributesModelType() map[string]attr.Type {
	return map[string]attr.Type{
		fieldNameAttributes:            types.MapType{ElemType: types.StringType},
		fieldNameClearPending:          types.BoolType,
		"attribute_registry":           types.StringType,
		NICComponmentSchemaOdataID:     types.StringType,
		NICComponmentSchemaID:          types.StringType,
		NICComponmentSchemaName:        types.StringType,
		NICComponmentSchemaDescription: types.StringType,
	}
}

func getNetworkDevFuncSettingsModelType() map[string]attr.Type {
	return map[string]attr.Type{
		NICComponmentSchemaOdataID:          types.StringType,
		NICComponmentSchemaID:               types.StringType,
		NICComponmentSchemaName:             types.StringType,
		NICComponmentSchemaDescription:      types.StringType,
		"max_virtual_functions":             types.Int64Type,
		fieldNameNetDevFuncType:             types.StringType,
		"physical_port_assignment":          types.StringType,
		"net_dev_func_capabilities":         types.ListType{ElemType: types.StringType},
		"assignable_physical_ports":         types.ListType{ElemType: types.StringType},
		"assignable_physical_network_ports": types.ListType{ElemType: types.StringType},
		"status":                            types.ObjectType{AttrTypes: getStatusModelType()},
		fieldNameIscsiBoot:                  types.ObjectType{AttrTypes: getISCSIBootModelType()},
		fieldNameEthernet:                   types.ObjectType{AttrTypes: getEthernetModelType()},
		fieldNameFibreChannel:               types.ObjectType{AttrTypes: getFibreChannelModelType()},
	}
}

func getFibreChannelModelType() map[string]attr.Type {
	return map[string]attr.Type{
		fieldNameAllowFipVlanDiscovery: types.BoolType,
		fieldNameBootTargets:           types.ListType{ElemType: types.ObjectType{AttrTypes: getBootTargetModelType()}},
		"fcoe_active_vlan_id":          types.Int64Type,
		fieldNameFcoeLocalVlanID:       types.Int64Type,
		"permanent_wwnn":               types.StringType,
		"permanent_wwpn":               types.StringType,
		fieldNameWWNN:                  types.StringType,
		fieldNameWWNNSource:            types.StringType,
		fieldNameWWPN:                  types.StringType,
		"fibre_channel_id":             types.StringType,
	}
}

func getBootTargetModelType() map[string]attr.Type {
	return map[string]attr.Type{
		fieldNameBootPriority: types.Int64Type,
		fieldNameLunID:        types.StringType,
		fieldNameWWPN:         types.StringType,
	}
}

func getEthernetModelType() map[string]attr.Type {
	return map[string]attr.Type{
		fieldNameMACAddress:     types.StringType,
		fieldNameMTUSize:        types.Int64Type,
		"permanent_mac_address": types.StringType,
		fieldNameVLAN:           types.ObjectType{AttrTypes: newVLANModelType()},
	}
}

func newVLANModelType() map[string]attr.Type {
	return map[string]attr.Type{
		fieldNameVLANID:      types.Int64Type,
		fieldNameVLANEnabled: types.BoolType,
	}
}

func getStatusModelType() map[string]attr.Type {
	return map[string]attr.Type{
		"state":         types.StringType,
		fieldNameHealth: types.StringType,
		"health_rollup": types.StringType,
	}
}

func getISCSIBootModelType() map[string]attr.Type {
	return map[string]attr.Type{
		fieldNameAuthenticationMethod:       types.StringType,
		fieldNameChapSec:                    types.StringType,
		fieldNameChapUsername:               types.StringType,
		fieldNameIPAddressType:              types.StringType,
		fieldNameIPMaskDNSViaDHCP:           types.BoolType,
		fieldNameInitiatorDefaultGateway:    types.StringType,
		fieldNameInitiatorIPAddress:         types.StringType,
		fieldNameInitiatorName:              types.StringType,
		fieldNameInitiatorNetmask:           types.StringType,
		fieldNameMutualChapSec:              types.StringType,
		fieldNameMutualChapUsername:         types.StringType,
		fieldNamePrimaryDNS:                 types.StringType,
		fieldNamePrimaryLun:                 types.Int64Type,
		fieldNamePrimaryTargetIPAddress:     types.StringType,
		fieldNamePrimaryTargetName:          types.StringType,
		fieldNamePrimaryTargetTCPPort:       types.Int64Type,
		fieldNamePrimaryVLANEnable:          types.BoolType,
		fieldNamePrimaryVLANID:              types.Int64Type,
		fieldNameRouterAdvertisementEnabled: types.BoolType,
		fieldNameSecondaryDNS:               types.StringType,
		fieldNameSecondaryLun:               types.Int64Type,
		fieldNameSecondaryTargetIPAddress:   types.StringType,
		fieldNameSecondaryTargetName:        types.StringType,
		fieldNameSecondaryTargetTCPPort:     types.Int64Type,
		fieldNameSecondaryVLANEnable:        types.BoolType,
		fieldNameSecondaryVLANID:            types.Int64Type,
		fieldNameTargetInfoViaDHCP:          types.BoolType,
	}
}

func getISCSIBootObjectValue(ctx context.Context, dellDeviceFunction *dell.NetworkDeviceFunctionExtended, state *models.NICResource,
	objectAsOptions basetypes.ObjectAsOptions,
) (basetypes.ObjectValue, diag.Diagnostics) {
	emptyObj := types.ObjectNull(getISCSIBootModelType())
	var oldDevFuncSettings models.NetworkDeviceFunctionSettings
	var oldISCSIBoot models.ISCSIBoot
	if diags := state.Networktributes.As(ctx, &oldDevFuncSettings, objectAsOptions); diags.HasError() {
		return emptyObj, diags
	}
	if diags := oldDevFuncSettings.ISCSIBoot.As(ctx, &oldISCSIBoot, objectAsOptions); diags.HasError() {
		return emptyObj, diags
	}
	if (oldDevFuncSettings.ISCSIBoot.IsNull() || oldDevFuncSettings.ISCSIBoot.IsUnknown()) &&
		dellDeviceFunction.ISCSIBoot.AuthenticationMethod == "" && dellDeviceFunction.ISCSIBoot.IPAddressType == "" {
		return emptyObj, nil
	}

	iscsiBootItemMap := map[string]attr.Value{
		fieldNameAuthenticationMethod:       types.StringValue(string(dellDeviceFunction.ISCSIBoot.AuthenticationMethod)),
		fieldNameChapSec:                    types.StringValue(dellDeviceFunction.ISCSIBoot.CHAPSecret),
		fieldNameChapUsername:               types.StringValue(dellDeviceFunction.ISCSIBoot.CHAPUsername),
		fieldNameIPAddressType:              types.StringValue(string(dellDeviceFunction.ISCSIBoot.IPAddressType)),
		fieldNameIPMaskDNSViaDHCP:           types.BoolValue(dellDeviceFunction.ISCSIBoot.IPMaskDNSViaDHCP),
		fieldNameInitiatorDefaultGateway:    types.StringValue(dellDeviceFunction.ISCSIBoot.InitiatorDefaultGateway),
		fieldNameInitiatorIPAddress:         types.StringValue(dellDeviceFunction.ISCSIBoot.InitiatorIPAddress),
		fieldNameInitiatorName:              types.StringValue(dellDeviceFunction.ISCSIBoot.InitiatorName),
		fieldNameInitiatorNetmask:           types.StringValue(dellDeviceFunction.ISCSIBoot.InitiatorNetmask),
		fieldNameMutualChapSec:              types.StringValue(dellDeviceFunction.ISCSIBoot.MutualCHAPSecret),
		fieldNameMutualChapUsername:         types.StringValue(dellDeviceFunction.ISCSIBoot.MutualCHAPUsername),
		fieldNamePrimaryDNS:                 types.StringValue(dellDeviceFunction.ISCSIBoot.PrimaryDNS),
		fieldNamePrimaryLun:                 types.Int64Value(int64(dellDeviceFunction.ISCSIBoot.PrimaryLUN)),
		fieldNamePrimaryTargetIPAddress:     types.StringValue(dellDeviceFunction.ISCSIBoot.PrimaryTargetIPAddress),
		fieldNamePrimaryTargetName:          types.StringValue(dellDeviceFunction.ISCSIBoot.PrimaryTargetName),
		fieldNamePrimaryTargetTCPPort:       types.Int64Value(int64(dellDeviceFunction.ISCSIBoot.PrimaryTargetTCPPort)),
		fieldNamePrimaryVLANEnable:          types.BoolValue(dellDeviceFunction.ISCSIBoot.PrimaryVLANEnable),
		fieldNamePrimaryVLANID:              types.Int64Value(int64(dellDeviceFunction.ISCSIBoot.PrimaryVLANID)),
		fieldNameRouterAdvertisementEnabled: types.BoolValue(dellDeviceFunction.ISCSIBoot.RouterAdvertisementEnabled),
		fieldNameSecondaryDNS:               types.StringValue(dellDeviceFunction.ISCSIBoot.SecondaryDNS),
		fieldNameSecondaryLun:               types.Int64Value(int64(dellDeviceFunction.ISCSIBoot.SecondaryLUN)),
		fieldNameSecondaryTargetIPAddress:   types.StringValue(dellDeviceFunction.ISCSIBoot.SecondaryTargetIPAddress),
		fieldNameSecondaryTargetName:        types.StringValue(dellDeviceFunction.ISCSIBoot.SecondaryTargetName),
		fieldNameSecondaryTargetTCPPort:     types.Int64Value(int64(dellDeviceFunction.ISCSIBoot.SecondaryTargetTCPPort)),
		fieldNameSecondaryVLANEnable:        types.BoolValue(dellDeviceFunction.ISCSIBoot.SecondaryVLANEnable),
		fieldNameSecondaryVLANID:            types.Int64Value(int64(dellDeviceFunction.ISCSIBoot.SecondaryVLANID)),
		fieldNameTargetInfoViaDHCP:          types.BoolValue(dellDeviceFunction.ISCSIBoot.TargetInfoViaDHCP),
	}
	for key, value := range oldDevFuncSettings.ISCSIBoot.Attributes() {
		if !value.IsUnknown() && !value.IsNull() {
			iscsiBootItemMap[key] = value
		}
	}

	return types.ObjectValue(getISCSIBootModelType(), iscsiBootItemMap)
}

func getEthernetObjectValue(ctx context.Context, dellDeviceFunction *dell.NetworkDeviceFunctionExtended, state *models.NICResource,
	objectAsOptions basetypes.ObjectAsOptions,
) (basetypes.ObjectValue, diag.Diagnostics) {
	var diags diag.Diagnostics
	emptyObj := types.ObjectNull(getEthernetModelType())

	var oldDevFuncSettings models.NetworkDeviceFunctionSettings
	var oldEthernet models.EthernetSettings
	if diags = state.Networktributes.As(ctx, &oldDevFuncSettings, objectAsOptions); diags.HasError() {
		return emptyObj, diags
	}
	if diags = oldDevFuncSettings.Ethernet.As(ctx, &oldEthernet, objectAsOptions); diags.HasError() {
		return emptyObj, diags
	}
	if (oldDevFuncSettings.Ethernet.IsNull() || oldDevFuncSettings.Ethernet.IsUnknown()) && dellDeviceFunction.Ethernet.PermanentMACAddress == "" {
		return emptyObj, diags
	}

	// build ethernet.vlan value into terraform object type
	var oldEthernetVlan models.VLAN
	if diags = oldEthernet.VLAN.As(ctx, &oldEthernetVlan, objectAsOptions); diags.HasError() {
		return emptyObj, diags
	}
	vlanItemMap := map[string]attr.Value{
		fieldNameVLANID:      types.Int64Value(int64(dellDeviceFunction.DellEthernet.VLAN.VLANID)),
		fieldNameVLANEnabled: types.BoolValue(dellDeviceFunction.DellEthernet.VLAN.VLANEnabled),
	}
	if !oldEthernetVlan.VLANID.IsUnknown() && !oldEthernetVlan.VLANID.IsNull() {
		vlanItemMap[fieldNameVLANID] = oldEthernetVlan.VLANID
	}
	if !oldEthernetVlan.VLANEnabled.IsUnknown() && !oldEthernetVlan.VLANEnabled.IsNull() {
		vlanItemMap[fieldNameVLANEnabled] = oldEthernetVlan.VLANEnabled
	}
	vlanItemObj, diags := types.ObjectValue(newVLANModelType(), vlanItemMap)
	if diags.HasError() {
		return emptyObj, diags
	}

	ethernetItemMap := map[string]attr.Value{
		fieldNameMACAddress:     types.StringValue(dellDeviceFunction.DellEthernet.MACAddress),
		fieldNameMTUSize:        types.Int64Value(int64(dellDeviceFunction.DellEthernet.MTUSize)),
		"permanent_mac_address": types.StringValue(dellDeviceFunction.DellEthernet.PermanentMACAddress),
		fieldNameVLAN:           vlanItemObj,
	}
	if !oldEthernet.MTUSize.IsUnknown() && !oldEthernet.MTUSize.IsNull() {
		ethernetItemMap[fieldNameMTUSize] = oldEthernet.MTUSize
	}
	if !oldEthernet.MACAddress.IsUnknown() && !oldEthernet.MACAddress.IsNull() {
		ethernetItemMap[fieldNameMACAddress] = oldEthernet.MACAddress
	}
	if oldEthernet.VLAN.IsUnknown() && oldEthernet.VLAN.IsNull() {
		ethernetItemMap[fieldNameVLAN] = types.ObjectNull(newVLANModelType())
	}

	return types.ObjectValue(getEthernetModelType(), ethernetItemMap)
}

// nolint: gocyclo,revive
func getFibreChannelObjectValue(ctx context.Context, dellDeviceFunction *dell.NetworkDeviceFunctionExtended, state *models.NICResource,
	objectAsOptions basetypes.ObjectAsOptions,
) (basetypes.ObjectValue, diag.Diagnostics) {
	var diags diag.Diagnostics
	emptyObj := types.ObjectNull(getFibreChannelModelType())

	var oldDevFuncSettings models.NetworkDeviceFunctionSettings
	var oldFibreChannel models.FibreChannelSettings

	if diags = state.Networktributes.As(ctx, &oldDevFuncSettings, objectAsOptions); diags.HasError() {
		return emptyObj, diags
	}
	if diags = oldDevFuncSettings.FibreChannel.As(ctx, &oldFibreChannel, objectAsOptions); diags.HasError() {
		return emptyObj, diags
	}
	if (oldDevFuncSettings.FibreChannel.IsNull() && oldDevFuncSettings.FibreChannel.IsUnknown()) &&
		dellDeviceFunction.FibreChannel.PermanentWWNN == "" && dellDeviceFunction.FibreChannel.PermanentWWPN == "" {
		return emptyObj, diags
	}

	fibreChannelItemMap := map[string]attr.Value{
		fieldNameAllowFipVlanDiscovery: types.BoolValue(dellDeviceFunction.FibreChannel.AllowFIPVLANDiscovery),
		"fcoe_active_vlan_id":          types.Int64Value(int64(dellDeviceFunction.FibreChannel.FCoEActiveVLANID)),
		fieldNameFcoeLocalVlanID:       types.Int64Value(int64(dellDeviceFunction.FibreChannel.FCoELocalVLANID)),
		"permanent_wwnn":               types.StringValue(dellDeviceFunction.FibreChannel.PermanentWWNN),
		"permanent_wwpn":               types.StringValue(dellDeviceFunction.FibreChannel.PermanentWWPN),
		fieldNameWWNN:                  types.StringValue(dellDeviceFunction.FibreChannel.WWNN),
		fieldNameWWNNSource:            types.StringValue(string(dellDeviceFunction.FibreChannel.WWNSource)),
		fieldNameWWPN:                  types.StringValue(dellDeviceFunction.FibreChannel.WWPN),
		"fibre_channel_id":             types.StringValue(dellDeviceFunction.FibreChannel.FibreChannelID),
	}

	if !oldFibreChannel.WWNN.IsUnknown() && !oldFibreChannel.WWNN.IsNull() {
		fibreChannelItemMap[fieldNameWWNN] = oldFibreChannel.WWNN
	}
	if !oldFibreChannel.WWPN.IsUnknown() && !oldFibreChannel.WWPN.IsNull() {
		fibreChannelItemMap[fieldNameWWPN] = oldFibreChannel.WWPN
	}
	if !oldFibreChannel.FCoELocalVLANId.IsUnknown() && !oldFibreChannel.FCoELocalVLANId.IsNull() {
		fibreChannelItemMap[fieldNameFcoeLocalVlanID] = oldFibreChannel.FCoELocalVLANId
	}
	if !oldFibreChannel.WWNSource.IsUnknown() && !oldFibreChannel.WWNSource.IsNull() {
		fibreChannelItemMap[fieldNameWWNNSource] = oldFibreChannel.WWNSource
	}
	if !oldFibreChannel.AllowFIPVLANDiscovery.IsUnknown() && !oldFibreChannel.AllowFIPVLANDiscovery.IsNull() {
		fibreChannelItemMap[fieldNameAllowFipVlanDiscovery] = oldFibreChannel.AllowFIPVLANDiscovery
	}
	if !oldFibreChannel.BootTargets.IsUnknown() && !oldFibreChannel.BootTargets.IsNull() {
		bootTargetItemObjList := make([]attr.Value, 0)
		var oldFcBootTargets []models.BootTarget
		if diags = oldFibreChannel.BootTargets.ElementsAs(ctx, &oldFcBootTargets, false); diags.HasError() {
			return emptyObj, diags
		}
		for _, oldElement := range oldFcBootTargets {
			var newActiveBootTarget redfish.BootTargets
			for _, newElement := range dellDeviceFunction.FibreChannel.BootTargets {
				if newElement.LUNID == oldElement.LUNID.ValueString() || newElement.WWPN == oldElement.WWPN.ValueString() {
					newActiveBootTarget = newElement
					break
				}
			}
			bootTargetItemMap := map[string]attr.Value{
				fieldNameBootPriority: types.Int64Value(int64(newActiveBootTarget.BootPriority)),
				fieldNameLunID:        types.StringValue(newActiveBootTarget.LUNID),
				fieldNameWWPN:         types.StringValue(newActiveBootTarget.WWPN),
			}
			if !oldElement.LUNID.IsUnknown() && !oldElement.LUNID.IsNull() {
				bootTargetItemMap[fieldNameLunID] = oldElement.LUNID
			}
			if !oldElement.BootPriority.IsUnknown() && !oldElement.BootPriority.IsNull() {
				bootTargetItemMap[fieldNameBootPriority] = oldElement.BootPriority
			}
			if !oldElement.WWPN.IsUnknown() && !oldElement.WWPN.IsNull() {
				bootTargetItemMap[fieldNameWWPN] = oldElement.WWPN
			}
			bootTargetItemObj, diags := types.ObjectValue(getBootTargetModelType(), bootTargetItemMap)
			if diags.HasError() {
				return emptyObj, diags
			}
			bootTargetItemObjList = append(bootTargetItemObjList, bootTargetItemObj)
		}
		fibreChannelItemMap[fieldNameBootTargets], diags = types.ListValue(types.ObjectType{AttrTypes: getBootTargetModelType()}, bootTargetItemObjList)
		if diags.HasError() {
			return emptyObj, diags
		}
	} else {
		bootTargetItemObjList := make([]attr.Value, 0)
		for _, t := range dellDeviceFunction.FibreChannel.BootTargets {
			bootTargetItemMap := map[string]attr.Value{
				fieldNameBootPriority: types.Int64Value(int64(t.BootPriority)),
				fieldNameLunID:        types.StringValue(t.LUNID),
				fieldNameWWPN:         types.StringValue(t.WWPN),
			}

			bootTargetItemObj, diags := types.ObjectValue(getBootTargetModelType(), bootTargetItemMap)
			if diags.HasError() {
				return emptyObj, diags
			}
			bootTargetItemObjList = append(bootTargetItemObjList, bootTargetItemObj)
		}
		fibreChannelItemMap[fieldNameBootTargets], diags = types.ListValue(types.ObjectType{AttrTypes: getBootTargetModelType()}, bootTargetItemObjList)
		if diags.HasError() {
			return emptyObj, diags
		}
	}
	return types.ObjectValue(getFibreChannelModelType(), fibreChannelItemMap)
}

func getStatusObjectValue(_ context.Context, dellDeviceFunction *dell.NetworkDeviceFunctionExtended) (basetypes.ObjectValue, diag.Diagnostics) {
	statusItemMap := map[string]attr.Value{
		"state":         types.StringValue(string(dellDeviceFunction.Status.State)),
		fieldNameHealth: types.StringValue(string(dellDeviceFunction.Status.Health)),
		"health_rollup": types.StringValue(string(dellDeviceFunction.Status.HealthRollup)),
	}
	return types.ObjectValue(getStatusModelType(), statusItemMap)
}

func getStringListValue(_ context.Context, stringList []types.String) (basetypes.ListValue, diag.Diagnostics) {
	listValue := make([]attr.Value, 0)
	for _, v := range stringList {
		listValue = append(listValue, v)
	}
	return types.ListValue(types.StringType, listValue)
}

func parseDellNetworkAttributesIntoState(ctx context.Context, attrs *dell.NetworkAttributes, state *models.NICResource) diag.Diagnostics {
	var diags diag.Diagnostics
	// Get config attributes
	readAttributes := make(map[string]attr.Value)
	objectAsOptions := basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true}

	var oldAttrsObj models.OemNetworkAttributes
	if diags = state.OemNetworkAttributes.As(ctx, &oldAttrsObj, objectAsOptions); diags.HasError() {
		return diags
	}
	mapObj := oldAttrsObj.Attributes
	if oldAttrsObj.Attributes.IsUnknown() || oldAttrsObj.Attributes.IsNull() {
		for k, attrValue := range attrs.Attributes {
			if attrValue != nil {
				attributeValue(attrValue, readAttributes, k)
			} else {
				readAttributes[k] = types.StringValue("")
			}
		}
		mapObj = types.MapValueMust(types.StringType, readAttributes)
	}

	oemNetworkAttrsValueMap := map[string]attr.Value{
		fieldNameAttributes:            mapObj,
		fieldNameClearPending:          types.BoolNull(),
		"attribute_registry":           types.StringValue(attrs.AttributeRegistry),
		NICComponmentSchemaOdataID:     types.StringValue(attrs.ODataID),
		NICComponmentSchemaID:          types.StringValue(attrs.ID),
		NICComponmentSchemaName:        types.StringValue(attrs.Name),
		NICComponmentSchemaDescription: types.StringValue(attrs.Description),
	}
	if !oldAttrsObj.ClearPending.IsUnknown() {
		oemNetworkAttrsValueMap[fieldNameClearPending] = oldAttrsObj.ClearPending
	}
	oemNetworkAttrsObj, diags := types.ObjectValue(getOemNetworkAttributesModelType(), oemNetworkAttrsValueMap)
	if diags.HasError() {
		return diags
	}
	state.OemNetworkAttributes = oemNetworkAttrsObj
	return diags
}

func parseNetworkDeviceFunctionIntoState(ctx context.Context, dellDeviceFunction *dell.NetworkDeviceFunctionExtended,
	state *models.NICResource,
) diag.Diagnostics {
	objectAsOptions := basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true}
	var oldDevFuncSettings models.NetworkDeviceFunctionSettings
	if diags := state.Networktributes.As(ctx, &oldDevFuncSettings, objectAsOptions); diags.HasError() {
		return diags
	}

	// build ISCSIBoot value into terraform object type
	iscsiBootItemObj, diags := getISCSIBootObjectValue(ctx, dellDeviceFunction, state, objectAsOptions)
	if diags.HasError() {
		return diags
	}
	// build status value into terraform object type
	statusItemObj, diags := getStatusObjectValue(ctx, dellDeviceFunction)
	if diags.HasError() {
		return diags
	}
	// build FibreChannel value into terraform object type
	fibreChannelItemObj, diags := getFibreChannelObjectValue(ctx, dellDeviceFunction, state, objectAsOptions)
	if diags.HasError() {
		return diags
	}
	// build ethernet value into terraform object type
	ethernetItemObj, diags := getEthernetObjectValue(ctx, dellDeviceFunction, state, objectAsOptions)
	if diags.HasError() {
		return diags
	}

	// parse string list to terraform list type
	capabilitiesList, diags := getStringListValue(ctx, newNetDevFuncCapabilities(dellDeviceFunction.NetDevFuncCapabilities))
	if diags.HasError() {
		return diags
	}
	physicalPortsList, diags := getStringListValue(ctx, newEntityStringList(dellDeviceFunction.DellAssignablePhysicalPorts))
	if diags.HasError() {
		return diags
	}
	physicalNetworkPortsList, diags := getStringListValue(ctx, newEntityStringList(dellDeviceFunction.DellAssignablePhysicalNetworkPorts))
	if diags.HasError() {
		return diags
	}

	netDevFuncItemMap := map[string]attr.Value{
		NICComponmentSchemaOdataID:          types.StringValue(dellDeviceFunction.ODataID),
		NICComponmentSchemaID:               types.StringValue(dellDeviceFunction.ID),
		NICComponmentSchemaName:             types.StringValue(dellDeviceFunction.Name),
		NICComponmentSchemaDescription:      types.StringValue(dellDeviceFunction.Description),
		"max_virtual_functions":             types.Int64Value(int64(dellDeviceFunction.MaxVirtualFunctions)),
		fieldNameNetDevFuncType:             types.StringValue(string(dellDeviceFunction.NetDevFuncType)),
		"physical_port_assignment":          types.StringValue(dellDeviceFunction.DellPhysicalPortAssignment.ODataID),
		"net_dev_func_capabilities":         capabilitiesList,
		"assignable_physical_ports":         physicalPortsList,
		"assignable_physical_network_ports": physicalNetworkPortsList,
		"status":                            statusItemObj,
		fieldNameIscsiBoot:                  iscsiBootItemObj,
		fieldNameEthernet:                   ethernetItemObj,
		fieldNameFibreChannel:               fibreChannelItemObj,
	}
	if !oldDevFuncSettings.NetDevFuncType.IsUnknown() && !oldDevFuncSettings.NetDevFuncType.IsNull() {
		netDevFuncItemMap[fieldNameNetDevFuncType] = oldDevFuncSettings.NetDevFuncType
	}

	state.Networktributes, diags = types.ObjectValue(getNetworkDevFuncSettingsModelType(), netDevFuncItemMap)

	return diags
}

func getFibreChannelPatchBody(ctx context.Context, attrsState *models.NetworkDeviceFunctionSettings) (map[string]interface{}, diag.Diagnostics) {
	supportedFcParams := map[string]string{
		fieldNameAllowFipVlanDiscovery: "AllowFIPVLANDiscovery",
		fieldNameFcoeLocalVlanID:       "FCoELocalVLANId",
		fieldNameWWNNSource:            "WWNSource",
		fieldNameWWPN:                  "WWPN",
		fieldNameWWNN:                  "WWNN",
		fieldNameBootTargets:           "BootTargets",
	}
	supportedBootTargetParams := map[string]string{
		fieldNameBootPriority: "BootPriority",
		fieldNameWWPN:         "WWPN",
		fieldNameLunID:        "LUNID",
	}
	objectAsOptions := basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true}

	var fcPlan models.FibreChannelSettings
	if diags := attrsState.FibreChannel.As(ctx, &fcPlan, objectAsOptions); diags.HasError() {
		return nil, diags
	}
	patchBody := make(map[string]interface{})
	for key, value := range attrsState.FibreChannel.Attributes() {
		if !value.IsUnknown() && !value.IsNull() {
			goValue, err := convertTerraformValueToGoBasicValue(ctx, value)
			if err != nil {
				tflog.Trace(ctx, fmt.Sprintf("Failed to convert FibreChannel value to go value: %s", err.Error()))
				continue
			}
			if fieldName, ok := supportedFcParams[key]; ok {
				patchBody[fieldName] = goValue
			}
		}
	}

	// get BootTargets patch body
	if !fcPlan.BootTargets.IsUnknown() {
		var bootTargetsPlan []models.BootTarget
		if diags := fcPlan.BootTargets.ElementsAs(ctx, &bootTargetsPlan, true); diags.HasError() {
			return nil, diags
		}

		bootTargetList := make([]interface{}, 0)
		for _, target := range bootTargetsPlan {
			bootTargetPatchBody := make(map[string]interface{})
			if !target.BootPriority.IsNull() && !target.BootPriority.IsUnknown() {
				bootTargetPatchBody[supportedBootTargetParams[fieldNameBootPriority]] = int(target.BootPriority.ValueInt64())
			}
			if !target.LUNID.IsNull() && !target.LUNID.IsUnknown() {
				bootTargetPatchBody[supportedBootTargetParams[fieldNameLunID]] = target.LUNID.ValueString()
			}
			if !target.WWPN.IsNull() && !target.WWPN.IsUnknown() {
				bootTargetPatchBody[supportedBootTargetParams[fieldNameWWPN]] = target.WWPN.ValueString()
			}
			if len(bootTargetPatchBody) > 0 {
				bootTargetList = append(bootTargetList, bootTargetPatchBody)
			}
		}

		patchBody[supportedFcParams[fieldNameBootTargets]] = bootTargetList
	}

	return patchBody, nil
}

func getIscsiBootPatchBody(ctx context.Context, attrsState *models.NetworkDeviceFunctionSettings) (map[string]interface{}, diag.Diagnostics) {
	supportedParams := map[string]string{
		fieldNameAuthenticationMethod:       "AuthenticationMethod",
		fieldNameChapSec:                    "CHAPSecret",
		fieldNameChapUsername:               "CHAPUsername",
		fieldNameIPAddressType:              "IPAddressType",
		fieldNameIPMaskDNSViaDHCP:           "IPMaskDNSViaDHCP",
		fieldNameInitiatorDefaultGateway:    "InitiatorDefaultGateway",
		fieldNameInitiatorIPAddress:         "InitiatorIPAddress",
		fieldNameInitiatorName:              "InitiatorName",
		fieldNameInitiatorNetmask:           "InitiatorNetmask",
		fieldNameMutualChapSec:              "MutualCHAPSecret",
		fieldNameMutualChapUsername:         "MutualCHAPUsername",
		fieldNamePrimaryDNS:                 "PrimaryDNS",
		fieldNamePrimaryLun:                 "PrimaryLUN",
		fieldNamePrimaryTargetIPAddress:     "PrimaryTargetIPAddress",
		fieldNamePrimaryTargetName:          "PrimaryTargetName",
		fieldNamePrimaryTargetTCPPort:       "PrimaryTargetTCPPort",
		fieldNamePrimaryVLANEnable:          "PrimaryVLANEnable",
		fieldNamePrimaryVLANID:              "PrimaryVLANId",
		fieldNameRouterAdvertisementEnabled: "RouterAdvertisementEnabled",
		fieldNameSecondaryDNS:               "SecondaryDNS",
		fieldNameSecondaryLun:               "SecondaryLUN",
		fieldNameSecondaryTargetIPAddress:   "SecondaryTargetIPAddress",
		fieldNameSecondaryTargetName:        "SecondaryTargetName",
		fieldNameSecondaryTargetTCPPort:     "SecondaryTargetTCPPort",
		fieldNameSecondaryVLANEnable:        "SecondaryVLANEnable",
		fieldNameSecondaryVLANID:            "SecondaryVLANId",
		fieldNameTargetInfoViaDHCP:          "TargetInfoViaDHCP",
	}
	objectAsOptions := basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true}
	var iscsiPlan models.ISCSIBoot

	if diags := attrsState.ISCSIBoot.As(ctx, &iscsiPlan, objectAsOptions); diags.HasError() {
		return nil, diags
	}
	patchBody := make(map[string]interface{})
	for key, value := range attrsState.ISCSIBoot.Attributes() {
		if !value.IsUnknown() && !value.IsNull() {
			goValue, err := convertTerraformValueToGoBasicValue(ctx, value)
			if err != nil {
				tflog.Trace(ctx, fmt.Sprintf("Failed to convert ISCSIBoot value to go value: %s", err.Error()))
				continue
			}
			if fieldName, ok := supportedParams[key]; ok {
				patchBody[fieldName] = goValue
			}
		}
	}
	return patchBody, nil
}

func getEthernetPatchBody(ctx context.Context, attrsState *models.NetworkDeviceFunctionSettings) (map[string]interface{}, diag.Diagnostics) {
	supportedEthParams := map[string]string{
		fieldNameMACAddress: "MACAddress",
		fieldNameMTUSize:    "MTUSize",
		fieldNameVLAN:       "VLAN",
	}
	supportedVlanParams := map[string]string{
		fieldNameVLANID:      "VLANEnable",
		fieldNameVLANEnabled: "VLANId",
	}
	objectAsOptions := basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true}

	var ethPlan models.EthernetSettings
	if diags := attrsState.Ethernet.As(ctx, &ethPlan, objectAsOptions); diags.HasError() {
		return nil, diags
	}
	patchBody := make(map[string]interface{})
	for key, value := range attrsState.Ethernet.Attributes() {
		if !value.IsUnknown() && !value.IsNull() {
			goValue, err := convertTerraformValueToGoBasicValue(ctx, value)
			if err != nil {
				tflog.Trace(ctx, fmt.Sprintf("Failed to convert Ethernet value to go value: %s", err.Error()))
				continue
			}
			if fieldName, ok := supportedEthParams[key]; ok {
				patchBody[fieldName] = goValue
			}
		}
	}

	// get vlan patch body
	if !ethPlan.VLAN.IsNull() && !ethPlan.VLAN.IsUnknown() {
		vlanPatchBody := make(map[string]interface{})
		for key, value := range ethPlan.VLAN.Attributes() {
			if !value.IsUnknown() && !value.IsNull() {
				goValue, err := convertTerraformValueToGoBasicValue(ctx, value)
				if err != nil {
					tflog.Trace(ctx, fmt.Sprintf("Failed to convert VLAN value to go value: %s", err.Error()))
					continue
				}
				if fieldName, ok := supportedVlanParams[key]; ok {
					vlanPatchBody[fieldName] = goValue
				}
			}
		}
		patchBody[supportedEthParams[fieldNameVLAN]] = vlanPatchBody
	}

	return patchBody, nil
}
