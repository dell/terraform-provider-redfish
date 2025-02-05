/*
Copyright (c) 2025 Dell Inc., or its subsidiaries. All Rights Reserved.

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

package helper

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"terraform-provider-redfish/gofish/dell"
	"terraform-provider-redfish/redfish/models"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/stmcginnis/gofish"
	redfishcommon "github.com/stmcginnis/gofish/common"
	"github.com/stmcginnis/gofish/redfish"
)

// nolint: gocyclo,revive
const (
	defaultNICJobTimeout                int64 = 1200
	intervalNICJobCheckTime             int64 = 10
	defaultNICResetTimeout              int64 = 120
	fieldDescriptionNetDevFuncID              = "ID of the network device function"
	errMessageInvalidInput                    = "input params are not valid"
	noteMessageUpdateOneAttrsOnly             = "Please update one of network_attributes or oem_network_attributes at a time."
	noteMessageUpdateAttrsExclusive           = "Note: `oem_network_attributes` is mutually exclusive with `network_attributes`. "
	patchBodySettingsApplyTime                = "@Redfish.SettingsApplyTime"
	patchBodyApplyTime                        = "ApplyTime"
	fieldNameClearPending                     = "clear_pending"
	fieldNameAttributes                       = "attributes"
	fieldNameWWPN                             = "wwpn"
	fieldNameWWNN                             = "wwnn"
	fieldNameWWNNSource                       = "wwn_source"
	fieldNameBootPriority                     = "boot_priority"
	fieldNameLunID                            = "lun_id"
	fieldNameFibreChannel                     = "fibre_channel"
	fieldNameNetDevFuncType                   = "net_dev_func_type"
	fieldNameEthernet                         = "ethernet"
	fieldNameIscsiBoot                        = "iscsi_boot"
	fieldNameHealth                           = "health"
	fieldNameMACAddress                       = "mac_address"
	fieldNameMTUSize                          = "mtu_size"
	fieldNameVLAN                             = "vlan"
	fieldNameVLANID                           = "vlan_id"
	fieldNameVLANEnabled                      = "vlan_enabled"
	fieldNameAllowFipVlanDiscovery            = "allow_fip_vlan_discovery" // nolint: gosec
	fieldNameBootTargets                      = "boot_targets"
	fieldNameFcoeLocalVlanID                  = "fcoe_local_vlan_id"
	fieldNameAuthenticationMethod             = "authentication_method"
	fieldNameChapSec                          = "chap_secret"
	fieldNameChapUsername                     = "chap_username"
	fieldNameIPAddressType                    = "ip_address_type"
	fieldNameIPMaskDNSViaDHCP                 = "ip_mask_dns_via_dhcp"
	fieldNameInitiatorDefaultGateway          = "initiator_default_gateway"
	fieldNameInitiatorIPAddress               = "initiator_ip_address"
	fieldNameInitiatorName                    = "initiator_name"
	fieldNameInitiatorNetmask                 = "initiator_netmask"
	fieldNameMutualChapSec                    = "mutual_chap_secret"
	fieldNameMutualChapUsername               = "mutual_chap_username"
	fieldNamePrimaryDNS                       = "primary_dns"
	fieldNamePrimaryLun                       = "primary_lun"
	fieldNamePrimaryTargetIPAddress           = "primary_target_ip_address"
	fieldNamePrimaryTargetName                = "primary_target_name"
	fieldNamePrimaryTargetTCPPort             = "primary_target_tcp_port"
	fieldNamePrimaryVLANEnable                = "primary_vlan_enable"
	fieldNamePrimaryVLANID                    = "primary_vlan_id"
	fieldNameRouterAdvertisementEnabled       = "router_advertisement_enabled"
	fieldNameSecondaryDNS                     = "secondary_dns"
	fieldNameSecondaryLun                     = "secondary_lun"
	fieldNameSecondaryTargetIPAddress         = "secondary_target_ip_address"
	fieldNameSecondaryTargetName              = "secondary_target_name"
	fieldNameSecondaryTargetTCPPort           = "secondary_target_tcp_port"
	fieldNameSecondaryVLANEnable              = "secondary_vlan_enable"
	fieldNameSecondaryVLANID                  = "secondary_vlan_id"
	fieldNameTargetInfoViaDHCP                = "target_info_via_dhcp"
	// NICComponmentSchema attributes
	NICComponmentSchemaID               = "id"
	NICComponmentSchemaOdataID          = "odata_id"
	NICComponmentSchemaName             = "name"
	NICComponmentSchemaDescription      = "description"
	NICComponmentSchemaStatus           = "status"
	NICComponmentSchemaPartNumber       = "part_number"
	NICComponmentSchemaSerialNumber     = "serial_number"
	NICSchemaDescriptionForSerialNumber = "A manufacturer-allocated number used to identify the Small Form Factor pluggable(SFP) " +
		"Transceiver"
	NICSchemaDescriptionForDeprecatedNoteV440 = "Note: This property is deprecated and not supported " +
		"in iDRAC firmware version 4.40.00.00 or later versions"
	NICSchemaDescriptionForDeprecatedNoteV420 = "Note: This property will be deprecated in Poweredge systems " +
		"with model YX5X and iDRAC firmware version 4.20.20.20 or later"
		// RedfishJobErrorMsg specifies error details occured while tracking job details
	RedfishJobErrorMsg = "Error, job wasn't able to complete"
)

/* func updateRedfishNIC(ctx context.Context, service *gofish.Service, state, plan *models.NICResource) diag.Diagnostics {
	var diags diag.Diagnostics

	applyTime := plan.ApplyTime.ValueString()
	jobWait := true
	resetType := plan.ResetType.ValueString()
	resetTimeout := plan.ResetTimeout.ValueInt64()
	jobTimeout := plan.JobTimeout.ValueInt64()
	if applyTime == string(redfishcommon.AtMaintenanceWindowStartApplyTime) ||
		applyTime == string(redfishcommon.InMaintenanceWindowOnResetApplyTime) {
		if plan.MaintenanceWindow == nil || plan.MaintenanceWindow.StartTime.IsUnknown() {
			diags.AddError("Input param is not valid",
				"Please set `maintenance_window` when `apply_time` is `AtMaintenanceWindowStart` or `InMaintenanceWindowOnReset`")
			return diags
		}
		// when apply_time is AtMaintenanceWindowStart or InMaintenanceWindowOnReset, skip wait for job to finish
		jobWait = false
	}

	// Lock the mutex to avoid race conditions with othe                                           r resources
	redfishMutexKV.Lock(plan.RedfishServer[0].Endpoint.ValueString())
	defer redfishMutexKV.Unlock(plan.RedfishServer[0].Endpoint.ValueString())

	// OnReset case
	if applyTime == string(redfishcommon.OnResetApplyTime) {
		// Reboot the server
		pOp := powerOperator{ctx, service, plan.SystemID.ValueString()}
		_, err := pOp.PowerOperation(resetType, resetTimeout, intervalNICJobCheckTime)
		if err != nil {
			diags.AddError(RedfishJobErrorMsg, err.Error())
			return diags
		}
	}

	var jobURL string
	if oemNetworkAttributesChanged(ctx, plan, state) {
		jobURL, diags = updateNicOemNetworkAttributes(ctx, service, plan)
	} else if networkAttributesChanged(ctx, plan, state) {
		jobURL, diags = updateNicNetworktributes(ctx, service, plan)
	} else {
		jobWait = false
		tflog.Trace(ctx, "Both `oem_network_attributes` and `network_attributes` are not changed. Skip Update for NIC.")
	}
	if diags.HasError() {
		return diags
	}

	var err error
	if jobWait && jobURL != "" {
		// jobURL could be JobService and TaskService
		if strings.Contains(jobURL, "JobService") {
			err = common.WaitForJobToFinish(service, jobURL, intervalNICJobCheckTime, jobTimeout)
		} else {
			err = common.WaitForTaskToFinish(service, jobURL, intervalNICJobCheckTime, jobTimeout)
		}
		if err != nil {
			diags.AddError(RedfishJobErrorMsg, err.Error())
			return diags
		}
	}
	time.Sleep(60 * time.Second)
	tflog.Trace(ctx, "Job has been completed")

	return diags
} */

// OemNetworkAttributesChanged is helper function
func OemNetworkAttributesChanged(_ context.Context, plan, state *models.NICResource) bool {
	if plan.OemNetworkAttributes.IsUnknown() || plan.OemNetworkAttributes.IsNull() {
		return false
	}
	return oemClearPendingChanged(plan, state) || oemNetworkAttributesAttrsChanged(plan, state)
}

// UpdateNicNetworktributes is a helper function
// nolint: gocyclo,revive
func UpdateNicNetworktributes(ctx context.Context, service *gofish.Service, system *redfish.ComputerSystem, plan *models.NICResource) (jobURL string, diags diag.Diagnostics) {
	tflog.Info(ctx, "updateNicNetworktributes: started")
	applyTime := plan.ApplyTime.ValueString()
	networkAttrsError := "there was an issue when creating/updating network attributes: "
	if applyTime == string(redfishcommon.ImmediateApplyTime) {
		diags.AddError(networkAttrsError+errMessageInvalidInput, "`Immediate` is not supported by `network_attributes`")
		return
	}

	objectAsOptions := basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true}
	var attrsState models.NetworkDeviceFunctionSettings
	if diags = plan.Networktributes.As(ctx, &attrsState, objectAsOptions); diags.HasError() {
		return
	}

	// get networkDeviceFunction by system id, adapter id and networkDeviceFunction id
	_, networketworkDeviceFunc, err := getNetworkDeviceFunction(system,
		plan.NetworkAdapterID.ValueString(), plan.NetworkDeviceFunctionID.ValueString())
	if err != nil {
		diags.AddError(fmt.Sprintf("%s: error when retrieving NetworkDeviceFunction", networkAttrsError), err.Error())
		return
	}
	// get OEM data
	dellDeviceFunction, _ := dell.NetworkDeviceFunction(networketworkDeviceFunc)
	if dellDeviceFunction.SettingsObject.ODataID == "" {
		diags.AddError(networkAttrsError, "error get NetworkAttributes SettingsObject from NetworkDeviceFunction Extension")
		return
	}

	var ethChangable, fcChangable, iscsiChangable bool

	if !attrsState.FibreChannel.IsNull() && !attrsState.FibreChannel.IsUnknown() {
		if networketworkDeviceFunc.FibreChannel.PermanentWWNN == "" && networketworkDeviceFunc.FibreChannel.PermanentWWPN == "" {
			diags.AddError(networkAttrsError+errMessageInvalidInput, "maynot configure `fibre_channel` if no FibreChannel NIC is present")
			return
		}
		fcChangable = true
	}
	if !attrsState.Ethernet.IsNull() && !attrsState.Ethernet.IsUnknown() {
		if networketworkDeviceFunc.Ethernet.PermanentMACAddress == "" {
			diags.AddError(networkAttrsError+errMessageInvalidInput, "maynot configure `ethernet` if no Ethernet NIC is present")
			return
		}
		ethChangable = true
	}
	if !attrsState.ISCSIBoot.IsNull() && !attrsState.ISCSIBoot.IsUnknown() {
		if networketworkDeviceFunc.ISCSIBoot.AuthenticationMethod == "" && networketworkDeviceFunc.ISCSIBoot.IPAddressType == "" {
			diags.AddError(networkAttrsError+errMessageInvalidInput, "maynot configure `iscsi_boot` if no ISCSIBoot is present")
			return
		}
		iscsiChangable = true
	}

	// Set the body to send
	patchBody := make(map[string]interface{})
	patchBody[patchBodySettingsApplyTime] = map[string]interface{}{
		patchBodyApplyTime: plan.ApplyTime.ValueString(),
	}
	if strings.Contains(plan.ApplyTime.ValueString(), "Maintenance") {
		patchBody[patchBodySettingsApplyTime] = map[string]interface{}{
			patchBodyApplyTime:                   plan.ApplyTime.ValueString(),
			"MaintenanceWindowStartTime":         plan.MaintenanceWindow.StartTime.ValueString(),
			"MaintenanceWindowDurationInSeconds": plan.MaintenanceWindow.Duration.ValueInt64(),
		}
	}
	if iscsiChangable {
		if patchBody["iSCSIBoot"], diags = getIscsiBootPatchBody(ctx, &attrsState); diags.HasError() {
			return "", diags
		}
	}

	if ethChangable {
		if patchBody["Ethernet"], diags = getEthernetPatchBody(ctx, &attrsState); diags.HasError() {
			return "", diags
		}
	}
	if fcChangable {
		if patchBody["FibreChannel"], diags = getFibreChannelPatchBody(ctx, &attrsState); diags.HasError() {
			return "", diags
		}
	}
	if !attrsState.NetDevFuncType.IsUnknown() && !attrsState.NetDevFuncType.IsNull() {
		patchBody["NetDevFuncType"] = attrsState.NetDevFuncType.ValueString()
	}

	resp, err := service.GetClient().Patch(dellDeviceFunction.SettingsObject.ODataID, patchBody)
	if err != nil {
		diags.AddError(fmt.Sprintf("%s: patch request to iDRAC failed", networkAttrsError), err.Error())
		return
	}
	defer resp.Body.Close() // #nosec G104

	// check if location is present in the response header
	location, err := resp.Location()
	if err != nil {
		diags.AddError(fmt.Sprintf("%s: get location failed after patch request to iDRAC ", networkAttrsError), err.Error())
		return
	}
	return location.EscapedPath(), diags
}

// setManagerAttributesRightType gets a map[string]interface{} from terraform, where all keys are strings,
// and returns a map[string]interface{} where values are either string or ints, and can be used for PATCH
func setManagerAttributesRightType(rawAttributes map[string]string, registry *dell.ManagerAttributeRegistry) (map[string]interface{}, error) {
	patchMap := make(map[string]interface{})

	for k, v := range rawAttributes {
		attrType, err := registry.GetAttributeType(k)
		if err != nil {
			return nil, err
		}
		switch attrType {
		case "int":
			t, err := strconv.Atoi(v)
			if err != nil {
				return nil, fmt.Errorf("property %s must be an integer", k)
			}
			patchMap[k] = t
		case "string":
			patchMap[k] = v
		}
	}

	return patchMap, nil
}

func checkManagerAttributes(attrRegistry *dell.ManagerAttributeRegistry, attributes map[string]interface{}) error {
	var errStr string // Here will be collected all attribute errors to show to users

	for k, v := range attributes {
		err := attrRegistry.CheckAttribute(k, v)
		if err != nil {
			errStr += fmt.Sprintf("%s - %s\n", k, err.Error())
		}
	}
	if len(errStr) > 0 {
		return fmt.Errorf("%s", errStr)
	}

	return nil
}

// nolint: gocyclo,revive
// UpdateNicOemNetworkAttributes is a helper function
func UpdateNicOemNetworkAttributes(ctx context.Context, service *gofish.Service, system *redfish.ComputerSystem, plan *models.NICResource) (jobURL string, diags diag.Diagnostics) {
	tflog.Info(ctx, "updateNicOemNetworkAttributes: started")
	applyTime := plan.ApplyTime.ValueString()
	oemNetworkAttrsError := "there was an issue when creating/updating ome network attributes"

	objectAsOptions := basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true}
	var oemAttrsState models.OemNetworkAttributes
	if diags = plan.OemNetworkAttributes.As(ctx, &oemAttrsState, objectAsOptions); diags.HasError() {
		return
	}

	// get networkDeviceFunction by system id, adapter id and networkDeviceFunction id
	_, networketworkDeviceFunc, err := getNetworkDeviceFunction(system,
		plan.NetworkAdapterID.ValueString(), plan.NetworkDeviceFunctionID.ValueString())
	if err != nil {
		diags.AddError(fmt.Sprintf("%s: error when retrieving NetworkDeviceFunction", oemNetworkAttrsError), err.Error())
		return
	}

	// get OEM data
	dellDeviceFunction, _ := dell.NetworkDeviceFunction(networketworkDeviceFunc)
	if dellDeviceFunction.DellNetworkAttributes.ODataID == "" {
		diags.AddError(oemNetworkAttrsError, "error get DellNetworkAttributes ODataID from NetworkDeviceFunction Extension")
		return
	}
	oemNetworkAttrURL := dellDeviceFunction.DellNetworkAttributes.ODataID
	oemNetworkAttrSettingsURL := fmt.Sprintf("%v/Settings", oemNetworkAttrURL)
	oemNetworkAttrClearPendingURL := fmt.Sprintf("%v/Actions/DellManager.ClearPending", oemNetworkAttrSettingsURL)
	dellNetworkAttributes, err := dell.GetDellNetworkAttributes(service.GetClient(), oemNetworkAttrURL)
	if err != nil {
		diags.AddError(fmt.Sprintf("%s: error when retrieving DellNetworkAttributes", oemNetworkAttrsError), err.Error())
		return
	}

	// clear pending
	if oemAttrsState.ClearPending.ValueBool() {
		emptyPostBody := make(map[string]interface{})
		_, err := service.GetClient().Post(oemNetworkAttrClearPendingURL, emptyPostBody)
		if err != nil {
			errStr := err.Error()
			// if clear_pending failed due to no pending data to delete, just log it and continue.
			if !strings.Contains(errStr, "No pending data to delete") {
				diags.AddError(fmt.Sprintf("%s: post request for clear_pending to iDRAC failed", oemNetworkAttrsError), errStr)
				return
			}
			tflog.Warn(ctx, fmt.Sprintf("%s: oem network attributes clear_pending failed. Error: %s", oemNetworkAttrsError, errStr))
		}
	}

	// Get attributes
	attributesTf := make(map[string]string)
	if diags = oemAttrsState.Attributes.ElementsAs(ctx, &attributesTf, true); diags.HasError() {
		return
	}
	// get NetworkAttributeRegistry to check parameters before posting them to redfish
	networkAttributeRegistry, err := getNetworkAttributeRegistry(service, dellNetworkAttributes.ID)
	if err != nil {
		diags.AddError(fmt.Sprintf("%s: Could not get oem network attribute registry from iDRAC", oemNetworkAttrsError), err.Error())
		return
	}
	err = assertOemNetworkAttributes(attributesTf, networkAttributeRegistry)
	if err != nil {
		diags.AddError(fmt.Sprintf("%s: oem network attribute registry from iDRAC does not match input", oemNetworkAttrsError), err.Error())
		return
	}
	if len(attributesTf) == 0 {
		// if nothing to patch, just return
		return
	}
	// Set right attributes to patch (values from map are all string. It needs int and string)
	// re-use setManagerAttributesRightType to get right attributes for oem network attributes
	attributesToPatch, err := setManagerAttributesRightType(attributesTf, networkAttributeRegistry)
	if err != nil {
		diags.AddError(fmt.Sprintf("%s: Input oem network attributes could not be cast to the required type", oemNetworkAttrsError), err.Error())
		return
	}

	// Check that all attributes passed are compliant with the API
	err = checkManagerAttributes(networkAttributeRegistry, attributesToPatch)
	if err != nil {
		diags.AddError(fmt.Sprintf("%s: oem network attribute registry from iDRAC does not match input", oemNetworkAttrsError), err.Error())
		return
	}

	// Set the body to send
	patchBody := make(map[string]interface{})
	patchBody["Attributes"] = attributesToPatch
	patchBody[patchBodySettingsApplyTime] = map[string]interface{}{
		patchBodyApplyTime: plan.ApplyTime.ValueString(),
	}
	if strings.Contains(applyTime, "Maintenance") {
		patchBody[patchBodySettingsApplyTime] = map[string]interface{}{
			patchBodyApplyTime:                   plan.ApplyTime.ValueString(),
			"MaintenanceWindowStartTime":         plan.MaintenanceWindow.StartTime.ValueString(),
			"MaintenanceWindowDurationInSeconds": plan.MaintenanceWindow.Duration.ValueInt64(),
		}
	}

	resp, err := service.GetClient().Patch(oemNetworkAttrSettingsURL, patchBody)
	if err != nil {
		diags.AddError(fmt.Sprintf("%s: patch request to iDRAC failed", oemNetworkAttrsError), err.Error())
		return
	}
	defer resp.Body.Close() // #nosec G104

	// check if location is present in the response header
	// for all supported applytime, there should exist location and jobURL when setting oem network attributes successfully
	// for Immediate type, it will auto create one power reboot action, no need extra restart
	location, err := resp.Location()
	if err != nil {
		diags.AddError(fmt.Sprintf("%s: get location failed after patch request to iDRAC ", oemNetworkAttrsError), err.Error())
		return
	}
	return location.EscapedPath(), diags
}

func assertOemNetworkAttributes(rawAttributes map[string]string, managerAttributeRegistry *dell.ManagerAttributeRegistry) error {
	var err error
	// make map of name to ID of attributes
	attributes := make(map[string]string)
	for _, dellAttr := range managerAttributeRegistry.Attributes {
		attributes[dellAttr.AttributeName] = dellAttr.ID
	}

	// check if all input attributes are present in registry
	for k := range rawAttributes {
		_, ok := attributes[k]
		if !ok {
			err = errors.Join(err, fmt.Errorf("couldn't find oem network attribute %s", k))
			continue
		}
	}
	return err
}

// ReadRedfishNIC function read the redfish NIC
func ReadRedfishNIC(ctx context.Context, service *gofish.Service, system *redfish.ComputerSystem, state *models.NICResource) diag.Diagnostics {
	var diags diag.Diagnostics

	// get networkDeviceFunction by system id, adapter id and networkDeviceFunction id
	system, networketworkDeviceFunc, err := getNetworkDeviceFunction(system,
		state.NetworkAdapterID.ValueString(), state.NetworkDeviceFunctionID.ValueString())
	if err != nil {
		diags.AddError("Error when retrieving NetworkDeviceFunction", err.Error())
		return diags
	}

	// get OEM data
	dellDeviceFunction, _ := dell.NetworkDeviceFunction(networketworkDeviceFunc)
	if dellDeviceFunction.DellNetworkAttributes.ODataID == "" {
		diags.AddError("there was an issue when reading NIC", "error get DellNetworkAttributes ODataID from NetworkDeviceFunction Extension")
		return diags
	}
	dellNetworkAttributes, err := dell.GetDellNetworkAttributes(service.GetClient(), dellDeviceFunction.DellNetworkAttributes.ODataID)
	if err != nil {
		diags.AddError("Error when retrieving DellNetworkAttributes", err.Error())
		return diags
	}
	if diags = parseDellNetworkAttributesIntoState(ctx, dellNetworkAttributes, state); diags.HasError() {
		return diags
	}
	if diags = parseNetworkDeviceFunctionIntoState(ctx, dellDeviceFunction, state); diags.HasError() {
		return diags
	}

	state.ID = types.StringValue("redfish_network_adapter_resource")
	state.SystemID = types.StringValue(system.ID)
	return diags
}

// ConvertTerraformValueToGoBasicValue function convert terraform to go type
func ConvertTerraformValueToGoBasicValue(ctx context.Context, v attr.Value) (interface{}, error) {
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

func attributeValue(attrValue interface{}, readAttributes map[string]attr.Value, k string) {
	if _, ok := attrValue.(float64); ok {
		readAttributes[k] = types.StringValue(fmt.Sprintf("%.0f", attrValue))
	} else {
		readAttributes[k] = types.StringValue(attrValue.(string))
	}
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

func newNetDevFuncCapabilities(inputs []redfish.NetworkDeviceTechnology) []types.String {
	out := make([]types.String, 0)
	for _, input := range inputs {
		out = append(out, types.StringValue(string(input)))
	}
	return out
}

func newEntityStringList(input []dell.Entity) []types.String {
	out := make([]types.String, 0)
	for _, i := range input {
		out = append(out, types.StringValue(i.ODataID))
	}
	return out
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
			goValue, err := ConvertTerraformValueToGoBasicValue(ctx, value)
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
			goValue, err := ConvertTerraformValueToGoBasicValue(ctx, value)
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
			goValue, err := ConvertTerraformValueToGoBasicValue(ctx, value)
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
				goValue, err := ConvertTerraformValueToGoBasicValue(ctx, value)
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

// NetworkAttributesChanged is a bool function to return true on changed
func NetworkAttributesChanged(ctx context.Context, plan, state *models.NICResource) bool {
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

func getNetworkDeviceFunction(system *redfish.ComputerSystem, networkAdapterID, networkDeviceFuncID string) (*redfish.ComputerSystem,
	*redfish.NetworkDeviceFunction, error,
) {
	// get system by id, if system id is empty, use the first one.
	/* system, err := getSystemResource(service, systemID)
	if err != nil {
		return nil, nil, err
	} */

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
