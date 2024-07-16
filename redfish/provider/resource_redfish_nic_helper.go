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

	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/stmcginnis/gofish"
	"github.com/stmcginnis/gofish/redfish"
)

func networkAttributesChanged(ctx context.Context, plan, state *models.NICResource) bool {
	if plan.Networktributes.IsUnknown() || plan.Networktributes.IsNull() {
		return false
	}

	objectAsOptions := basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true}
	var attrsPlan models.NetworkDeviceFunctionSettings
	var attrsState models.NetworkDeviceFunctionSettings
	if diags := plan.Networktributes.As(ctx, &attrsPlan, objectAsOptions); diags.HasError() {
		tflog.Debug(ctx, "resource_redfish_nic: networkAttributesChanged: plan.Networktributes.As: error")
		return false
	}
	if diags := state.Networktributes.As(ctx, &attrsState, objectAsOptions); diags.HasError() {
		tflog.Debug(ctx, "resource_redfish_nic: networkAttributesChanged: state.Networktributes.As: error")
		return false
	}

	// check NetDevFuncType change
	if attrsPlan.NetDevFuncType.ValueString() != "" && attrsPlan.NetDevFuncType.ValueString() != attrsState.NetDevFuncType.ValueString() {
		return true
	}
	return networkEthernetChanged(ctx, &attrsPlan, &attrsState) || networkFibreChannelChanged(ctx, &attrsPlan, &attrsState) ||
		networkISCSIBootChanged(ctx, &attrsPlan, &attrsState)
}

func networkEthernetChanged(ctx context.Context, plan, state *models.NetworkDeviceFunctionSettings) bool {
	if plan.Ethernet.IsNull() || plan.Ethernet.IsUnknown() {
		return false
	}
	objectAsOptions := basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true}
	var attrsPlan models.EthernetSettings
	var attrsState models.EthernetSettings
	if diags := plan.Ethernet.As(ctx, &attrsPlan, objectAsOptions); diags.HasError() {
		tflog.Debug(ctx, "resource_redfish_nic: networkEthernetChanged: plan.Ethernet.As: error")
		return false
	}
	if diags := state.Ethernet.As(ctx, &attrsState, objectAsOptions); diags.HasError() {
		tflog.Debug(ctx, "resource_redfish_nic: networkEthernetChanged: state.Ethernet.As: error")
		return false
	}

	if !attrsPlan.MTUSize.IsNull() && !attrsPlan.MTUSize.IsUnknown() &&
		attrsPlan.MTUSize.ValueInt64() != attrsState.MTUSize.ValueInt64() {
		return true
	}
	if attrsPlan.MACAddress.ValueString() != "" && attrsPlan.MACAddress.ValueString() != attrsState.MACAddress.ValueString() {
		return true
	}

	return networkVLanChanged(ctx, &attrsPlan, &attrsState)
}

func networkVLanChanged(_ context.Context, plan, state *models.EthernetSettings) bool {
	if plan.VLAN.IsNull() || plan.VLAN.IsUnknown() {
		return false
	}
	stateAttrs := state.VLAN.Attributes()
	for key, planValue := range plan.VLAN.Attributes() {
		if planValue.IsNull() || planValue.IsUnknown() {
			continue
		}
		if stateValue, ok := stateAttrs[key]; !ok {
			return true
		} else if !planValue.Equal(stateValue) {
			return true
		}
	}
	return false
}

func networkISCSIBootChanged(_ context.Context, plan, state *models.NetworkDeviceFunctionSettings) bool {
	if plan.ISCSIBoot.IsUnknown() || plan.ISCSIBoot.IsNull() {
		return false
	}

	stateAttrs := state.ISCSIBoot.Attributes()
	for key, planValue := range plan.ISCSIBoot.Attributes() {
		if planValue.IsNull() || planValue.IsUnknown() {
			continue
		}
		if stateValue, ok := stateAttrs[key]; !ok {
			return true
		} else if !planValue.Equal(stateValue) {
			return true
		}
	}
	return false
}

func networkFibreChannelChanged(ctx context.Context, plan, state *models.NetworkDeviceFunctionSettings) bool {
	if plan.FibreChannel.IsNull() || plan.FibreChannel.IsUnknown() {
		return false
	}
	objectAsOptions := basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true}
	var attrsPlan models.FibreChannelSettings
	var attrsState models.FibreChannelSettings
	if diags := plan.FibreChannel.As(ctx, &attrsPlan, objectAsOptions); diags.HasError() {
		tflog.Debug(ctx, "resource_redfish_nic: networkFibreChannelChanged: plan.FibreChannel.As: error")
		return false
	}
	if diags := state.FibreChannel.As(ctx, &attrsState, objectAsOptions); diags.HasError() {
		tflog.Debug(ctx, "resource_redfish_nic: networkFibreChannelChanged: state.FibreChannel.As: error")
		return false
	}

	if !attrsPlan.AllowFIPVLANDiscovery.IsNull() && !attrsPlan.AllowFIPVLANDiscovery.IsUnknown() &&
		attrsPlan.AllowFIPVLANDiscovery.ValueBool() != attrsState.AllowFIPVLANDiscovery.ValueBool() {
		return true
	}
	if !attrsPlan.FCoELocalVLANId.IsNull() && !attrsPlan.FCoELocalVLANId.IsUnknown() &&
		attrsPlan.FCoELocalVLANId.ValueInt64() != attrsState.FCoELocalVLANId.ValueInt64() {
		return true
	}
	if attrsPlan.WWNN.ValueString() != "" && attrsPlan.WWNN.ValueString() != attrsState.WWNN.ValueString() {
		return true
	}
	if attrsPlan.WWPN.ValueString() != "" && attrsPlan.WWPN.ValueString() != attrsState.WWPN.ValueString() {
		return true
	}
	if attrsPlan.WWNSource.ValueString() != "" && attrsPlan.WWNSource.ValueString() != attrsState.WWNSource.ValueString() {
		return true
	}

	return networkBootTargetsChanged(ctx, &attrsPlan, &attrsState)
}

func networkBootTargetsChanged(ctx context.Context, plan, state *models.FibreChannelSettings) bool {
	if plan.BootTargets.IsUnknown() {
		return false
	}
	if state.BootTargets.IsUnknown() {
		return true
	}
	var bootTargetsPlan []models.BootTarget
	var bootTargetsState []models.BootTarget
	if diags := plan.BootTargets.ElementsAs(ctx, &bootTargetsPlan, true); diags.HasError() {
		tflog.Debug(ctx, "resource_redfish_nic: networkBootTargetsChanged: plan.BootTargets.ElementsAs: error")
		return false
	}
	if diags := state.BootTargets.ElementsAs(ctx, &bootTargetsState, true); diags.HasError() {
		tflog.Debug(ctx, "resource_redfish_nic: networkBootTargetsChanged: state.BootTargets.ElementsAs: error")
		return false
	}
	if len(bootTargetsPlan) != len(bootTargetsState) {
		return true
	}
	for i, planTarget := range bootTargetsPlan {
		stateTarget := bootTargetsState[i]
		if !planTarget.BootPriority.IsUnknown() && !planTarget.BootPriority.IsNull() &&
			planTarget.BootPriority.ValueInt64() != stateTarget.BootPriority.ValueInt64() {
			return true
		}
		if planTarget.LUNID.ValueString() != "" && planTarget.LUNID.ValueString() != stateTarget.LUNID.ValueString() {
			return true
		}
		if planTarget.WWPN.ValueString() != "" && planTarget.WWPN.ValueString() != stateTarget.WWPN.ValueString() {
			return true
		}
	}
	return false
}

func oemNetworkAttributesChanged(_ context.Context, plan, state *models.NICResource) bool {
	if plan.OemNetworkAttributes.IsUnknown() || plan.OemNetworkAttributes.IsNull() {
		return false
	}
	return oemClearPendingChanged(plan, state) || oemNetworkAttributesAttrsChanged(plan, state)
}

func oemNetworkAttributesAttrsChanged(plan, state *models.NICResource) bool {
	var planAttrStr, stateAttrStr string
	planAttr, ok := plan.OemNetworkAttributes.Attributes()[fieldNameAttributes]
	if !ok {
		return ok
	}
	planAttrStr = planAttr.String()

	if stateAttr, ok := state.OemNetworkAttributes.Attributes()[fieldNameAttributes]; ok {
		stateAttrStr = stateAttr.String()
	}
	return planAttrStr != stateAttrStr
}

func oemClearPendingChanged(plan, state *models.NICResource) bool {
	planClearPending, stateClearPending := true, true
	if planAttr, ok := plan.OemNetworkAttributes.Attributes()[fieldNameClearPending]; !ok || planAttr.String() == "false" {
		return false
	}
	if stateAttr, ok := state.OemNetworkAttributes.Attributes()[fieldNameClearPending]; !ok || stateAttr.String() == "false" {
		stateClearPending = false
	}
	return planClearPending != stateClearPending
}

func getNetworkAttributeRegistry(service *gofish.Service, deviceID string) (*dell.ManagerAttributeRegistry, error) {
	registries, err := service.Registries()
	if err != nil {
		return nil, err
	}

	for _, r := range registries {
		if strings.HasSuffix(r.ID, deviceID) && strings.HasPrefix(r.ID, "NetworkAttribute") {
			// Get NetworkAttributesRegistry
			// re-use ManagerAttributeRegistry for NetworkAttributesRegistry
			networkAttrRegistry, err := dell.GetDellManagerAttributeRegistry(service.GetClient(), r.Location[0].URI)
			if err != nil {
				return nil, err
			}
			return networkAttrRegistry, nil
		}
	}

	return nil, fmt.Errorf("error. Couldn't retrieve NetworkAttributesRegistry")
}

func getNetworkDeviceFunction(service *gofish.Service, systemID, networkAdapterID, networkDeviceFuncID string) (*redfish.ComputerSystem,
	*redfish.NetworkDeviceFunction, error,
) {
	// get system by id, if system id is empty, use the first one.
	system, err := getSystemResource(service, systemID)
	if err != nil {
		return nil, nil, err
	}

	// get network adapter by id
	adapter, err := getNetworkAdapter(system, networkAdapterID)
	if err != nil {
		return system, nil, err
	}

	networkDeviceFuncs, err := adapter.NetworkDeviceFunctions()
	if err != nil {
		return system, nil, err
	}

	// get network device function by id
	for _, f := range networkDeviceFuncs {
		if f.ID == networkDeviceFuncID {
			return system, f, nil
		}
	}
	return system, nil, fmt.Errorf("couldn't find network device function: %s", networkDeviceFuncID)
}

func getNetworkAdapter(system *redfish.ComputerSystem, adapterID string) (*redfish.NetworkAdapter, error) {
	networkInterfaceList, err := system.NetworkInterfaces()
	if err != nil {
		return nil, err
	}
	for _, n := range networkInterfaceList {
		if n.ID == adapterID {
			adapter, err := n.NetworkAdapter()
			if err != nil {
				return nil, err
			}
			return adapter, nil
		}
	}
	return nil, fmt.Errorf("couldn't find network adapter: %s", adapterID)
}
