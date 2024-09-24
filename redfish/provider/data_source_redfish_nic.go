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
	"math/big"
	"strings"
	"terraform-provider-redfish/gofish/dell"
	"terraform-provider-redfish/redfish/models"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stmcginnis/gofish"
	"github.com/stmcginnis/gofish/common"
	"github.com/stmcginnis/gofish/redfish"
)

var (
	_ datasource.DataSource              = &NICDatasource{}
	_ datasource.DataSourceWithConfigure = &NICDatasource{}
)

// Constants for NIC Component Schema strings for name and description that appear multiple times.
const (
	NICComponmentSchemaID                     = "id"
	NICComponmentSchemaOdataID                = "odata_id"
	NICComponmentSchemaName                   = "name"
	NICComponmentSchemaDescription            = "description"
	NICComponmentSchemaStatus                 = "status"
	NICComponmentSchemaPartNumber             = "part_number"
	NICComponmentSchemaSerialNumber           = "serial_number"
	NICSchemaDescriptionForSerialNumber       = "A manufacturer-allocated number used to identify the Small Form Factor pluggable(SFP) Transceiver"
	NICSchemaDescriptionForDeprecatedNoteV440 = "Note: This property is deprecated and not supported " +
		"in iDRAC firmware version 4.40.00.00 or later versions"
	NICSchemaDescriptionForDeprecatedNoteV420 = "Note: This property will be deprecated in Poweredge systems " +
		"with model YX5X and iDRAC firmware version 4.20.20.20 or later"
)

// NewNICDatasource is new datasource for NIC.
func NewNICDatasource() datasource.DataSource {
	return &NICDatasource{}
}

// NICDatasource to construct datasource.
type NICDatasource struct {
	p       *redfishProvider
	ctx     context.Context
	service *gofish.Service
}

// Configure implements datasource.DataSourceWithConfigure.
func (g *NICDatasource) Configure(_ context.Context, req datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	g.p = req.ProviderData.(*redfishProvider)
}

// Metadata implements datasource.DataSource.
func (*NICDatasource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "network"
}

// Schema implements datasource.DataSource.
func (*NICDatasource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "This Terraform datasource is used to query existing network interface cards(NIC) configuration including " +
			"network adapters, network ports, network device functions and their OEM attributes." +
			" The information fetched from this block can be further used for resource block.",
		Description: "This Terraform datasource is used to query existing network interface cards(NIC) configuration including " +
			"network adapters, network ports, network device functions and their OEM attributes." +
			" The information fetched from this block can be further used for resource block.",
		Attributes: NICDatasourceSchema(),
		Blocks: map[string]schema.Block{
			"nic_filter": schema.SingleNestedBlock{
				MarkdownDescription: "NIC filter for systems, nework adapters, network ports and network device functions",
				Description:         "NIC filter for systems, nework adapters, network ports and network device functions",
				Attributes:          NICFilterSchema(),
			},
			"redfish_server": schema.ListNestedBlock{
				MarkdownDescription: redfishServerMD,
				Description:         redfishServerMD,
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
					listvalidator.IsRequired(),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: RedfishServerDatasourceSchema(),
				},
			},
		},
	}
}

// Read implements datasource.DataSource.
func (g *NICDatasource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var plan models.NICDatasource
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
	state, diags := g.readDatasourceRedfishNIC(plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// nolint: gocyclo, gocognit,revive
func (g *NICDatasource) readDatasourceRedfishNIC(d models.NICDatasource) (models.NICDatasource, diag.Diagnostics) {
	var diags diag.Diagnostics

	// write the current time as ID
	d.ID = types.StringValue(fmt.Sprintf("%d", time.Now().Unix()))

	// re-use readDatasourceRedfishDellIdracAttributes and select NIC.* attributes
	if diags = loadNICAttributesState(g.service, &d); diags.HasError() {
		return d, diags
	}

	stringJoinSplit := " ,"
	systems, err := g.service.Systems()
	if err != nil {
		diags.AddError("Error fetching computer systems collection", err.Error())
		return d, diags
	}

	var validSystems []string
	var systemFilters []models.SystemFilter
	if d.NICFilter != nil {
		systemFilters = d.NICFilter.Systems
	}
	for _, system := range systems {
		var foundSystem bool
		var adapterFilters []models.NetworkAdapterFilter
		for _, filter := range systemFilters {
			if filter.SystemID.ValueString() == system.ID {
				foundSystem = true
				adapterFilters = filter.NetworkAdapters
				break
			}
		}
		if len(systemFilters) > 0 && !foundSystem {
			continue
		}
		validSystemID := system.ID
		validSystems = append(validSystems, validSystemID)

		var validAdapters []string
		networkInterfaces, err := system.NetworkInterfaces()
		if err != nil {
			diags.AddError("Error fetching NetworkInterfaces collection", err.Error())
			return d, diags
		}
		for _, networkInterface := range networkInterfaces {
			found := false
			var adapterFilter models.NetworkAdapterFilter
			for _, filter := range adapterFilters {
				if filter.NetworkAdapterID.ValueString() == networkInterface.ID {
					found = true
					adapterFilter = filter
					break
				}
			}
			if len(adapterFilters) > 0 && !found {
				continue
			}
			validAdapterID := networkInterface.ID
			validAdapters = append(validAdapters, validAdapterID)

			var validNetworkPorts, validNetworkDeviceFunctions []string
			adapter, err := networkInterface.NetworkAdapter()
			if err != nil {
				diags.AddError("Error fetching NetworkAdapter: %s", err.Error())
				return d, diags
			}
			ports, err := adapter.NetworkPorts()
			if err != nil {
				diags.AddError("Error fetching NetworkPorts collection", err.Error())
				return d, diags
			}
			filteredPorts := make([]*redfish.NetworkPort, 0)
			for _, filter := range adapterFilter.NetworkPortIDs {
				for _, port := range ports {
					if filter.ValueString() == port.ID {
						validNetworkPorts = append(validNetworkPorts, port.ID)
						filteredPorts = append(filteredPorts, port)
					}
				}
			}
			// check network ports filter diff
			if len(adapterFilter.NetworkPortIDs) != 0 && len(validNetworkPorts) != len(adapterFilter.NetworkPortIDs) {
				diags.AddError(
					fmt.Sprintf("Error one or more of the filtered network port ids are not valid for system:%s, adapter:%s",
						validSystemID, validAdapterID),
					fmt.Sprintf("Valid network port ids are [%v]", strings.Join(validNetworkPorts, stringJoinSplit)),
				)
				return d, diags
			}

			if len(filteredPorts) > 0 {
				ports = filteredPorts
			}
			deviceFunctions, err := adapter.NetworkDeviceFunctions()
			if err != nil {
				diags.AddError("Error fetching NetworkDeviceFunctions collection", err.Error())
				return d, diags
			}
			filteredDeviceFunctions := make([]*redfish.NetworkDeviceFunction, 0)
			for _, filter := range adapterFilter.NetworkDeviceFunctionIDs {
				for _, devFunc := range deviceFunctions {
					if filter.ValueString() == devFunc.ID {
						validNetworkDeviceFunctions = append(validNetworkDeviceFunctions, devFunc.ID)
						filteredDeviceFunctions = append(filteredDeviceFunctions, devFunc)
					}
				}
			}
			// check network device functions filter diff
			if len(adapterFilter.NetworkDeviceFunctionIDs) != 0 && len(validNetworkDeviceFunctions) != len(adapterFilter.NetworkDeviceFunctionIDs) {
				diags.AddError(
					fmt.Sprintf("Error one or more of the filtered network device function ids are not valid for system:%s, adapter:%s",
						validSystemID, validAdapterID),
					fmt.Sprintf("Valid network device function ids are [%v]", strings.Join(validNetworkDeviceFunctions, stringJoinSplit)),
				)
				return d, diags
			}

			if len(filteredDeviceFunctions) > 0 {
				deviceFunctions = filteredDeviceFunctions
			}
			d.NICs = append(d.NICs, newNetworkInterfaceState(networkInterface, adapter, ports, deviceFunctions))
		}
		// check adapters filter diff
		if len(adapterFilters) != 0 && len(validAdapters) != len(adapterFilters) {
			diags.AddError(
				fmt.Sprintf("Error one or more of the filtered network adapter ids are not valid for system:%s", validSystemID),
				fmt.Sprintf("Valid network adapter ids are [%v]", strings.Join(validAdapters, stringJoinSplit)),
			)
			return d, diags
		}
	}
	// check systems filter diff
	if len(systemFilters) != 0 && len(validSystems) != len(systemFilters) {
		diags.AddError(
			"Error one or more of the filtered system ids are not valid.",
			fmt.Sprintf("Valid system ids are [%v]", strings.Join(validSystems, stringJoinSplit)),
		)
		return d, diags
	}
	return d, diags
}

func newNetworkInterfaceState(networkInterface *redfish.NetworkInterface, adapter *redfish.NetworkAdapter,
	ports []*redfish.NetworkPort, deviceFunctions []*redfish.NetworkDeviceFunction,
) models.NetworkInterface {
	return models.NetworkInterface{
		ODataID:                types.StringValue(networkInterface.ODataID),
		Description:            types.StringValue(networkInterface.Description),
		ID:                     types.StringValue(networkInterface.ID),
		Name:                   types.StringValue(networkInterface.Name),
		Status:                 newNetworkStatus(networkInterface.Status),
		NetworkAdapter:         newNetworkAdapter(adapter),
		NetworkPorts:           newNetworkPorts(ports),
		NetworkDeviceFunctions: newNetworkDeviceFunctions(deviceFunctions),
	}
}

// newNetworkStatus converts redfish.Status to models.Status
func newNetworkStatus(input common.Status) models.Status {
	return models.Status{
		Health:       types.StringValue(string(input.Health)),
		HealthRollup: types.StringValue(string(input.HealthRollup)),
		State:        types.StringValue(string(input.State)),
	}
}

func newNetworkAdapter(adapter *redfish.NetworkAdapter) models.NetworkAdapter {
	return models.NetworkAdapter{
		ODataID:      types.StringValue(adapter.ODataID),
		ID:           types.StringValue(adapter.ID),
		Name:         types.StringValue(adapter.Name),
		Description:  types.StringValue(adapter.Description),
		Manufacturer: types.StringValue(adapter.Manufacturer),
		Model:        types.StringValue(adapter.Model),
		PartNumber:   types.StringValue(adapter.PartNumber),
		SerialNumber: types.StringValue(adapter.SerialNumber),
		Status:       newNetworkStatus(adapter.Status),
		Controllers:  newNetworkCollector(adapter.Controllers),
	}
}

func newNetworkCollector(collectors []redfish.Controllers) []models.NetworkCollector {
	networkCollectors := make([]models.NetworkCollector, 0)
	for _, collector := range collectors {
		networkCollectors = append(networkCollectors, models.NetworkCollector{
			FirmwarePackageVersion: types.StringValue(collector.FirmwarePackageVersion),
			ControllerCapabilities: newControllerCapabilities(collector.ControllerCapabilities),
		})
	}
	return networkCollectors
}

func newControllerCapabilities(controllerCapabilities redfish.ControllerCapabilities) models.ControllerCapabilities {
	return models.ControllerCapabilities{
		DataCenterBridging:    newDataCenterBridging(controllerCapabilities),
		NPAR:                  newNPAR(controllerCapabilities),
		NPIV:                  newNPIV(controllerCapabilities),
		VirtualizationOffload: newVirtualizationOffload(controllerCapabilities),
	}
}

func newDataCenterBridging(controllerCapabilities redfish.ControllerCapabilities) models.DataCenterBridging {
	return models.DataCenterBridging{
		Capable: types.BoolValue(controllerCapabilities.DataCenterBridging.Capable),
	}
}

func newNPAR(controllerCapabilities redfish.ControllerCapabilities) models.NPAR {
	return models.NPAR{
		NparCapable: types.BoolValue(controllerCapabilities.NPAR.NparCapable),
		NparEnabled: types.BoolValue(controllerCapabilities.NPAR.NparEnabled),
	}
}

func newNPIV(controllerCapabilities redfish.ControllerCapabilities) models.NPIV {
	return models.NPIV{
		MaxDeviceLogins: types.Int64Value(int64(controllerCapabilities.NPIV.MaxDeviceLogins)),
		MaxPortLogins:   types.Int64Value(int64(controllerCapabilities.NPIV.MaxPortLogins)),
	}
}

func newVirtualizationOffload(controllerCapabilities redfish.ControllerCapabilities) models.VirtualizationOffload {
	return models.VirtualizationOffload{
		SRIOV:           newSRIOV(controllerCapabilities),
		VirtualFunction: newVirtualFunction(controllerCapabilities),
	}
}

func newSRIOV(controllerCapabilities redfish.ControllerCapabilities) models.SRIOV {
	return models.SRIOV{
		SRIOVVEPACapable: types.BoolValue(controllerCapabilities.VirtualizationOffload.SRIOV.SRIOVVEPACapable),
	}
}

func newVirtualFunction(controllerCapabilities redfish.ControllerCapabilities) models.VirtualFunction {
	return models.VirtualFunction{
		DeviceMaxCount:         types.Int64Value(int64(controllerCapabilities.VirtualizationOffload.VirtualFunction.DeviceMaxCount)),
		MinAssignmentGroupSize: types.Int64Value(int64(controllerCapabilities.VirtualizationOffload.VirtualFunction.MinAssignmentGroupSize)),
		NetworkPortMaxCount:    types.Int64Value(int64(controllerCapabilities.VirtualizationOffload.VirtualFunction.NetworkPortMaxCount)),
	}
}

func newNetworkPorts(ports []*redfish.NetworkPort) []models.NetworkPort {
	networkPorts := make([]models.NetworkPort, 0)
	for _, port := range ports {
		// Get Dell NetworkPort extension
		dellNetworkPort, _ := dell.NetworkPort(port)

		networkPorts = append(networkPorts, models.NetworkPort{
			ODataID:                       types.StringValue(dellNetworkPort.ODataID),
			ID:                            types.StringValue(dellNetworkPort.ID),
			Name:                          types.StringValue(dellNetworkPort.Name),
			Status:                        newNetworkStatus(dellNetworkPort.Status),
			Description:                   types.StringValue(dellNetworkPort.Description),
			ActiveLinkTechnology:          types.StringValue(string(dellNetworkPort.ActiveLinkTechnology)),
			CurrentLinkSpeedMbps:          types.Int64Value(int64(dellNetworkPort.CurrentLinkSpeedMbps)),
			EEEEnabled:                    types.BoolValue(dellNetworkPort.EEEEnabled),
			FlowControlConfiguration:      types.StringValue(string(dellNetworkPort.FlowControlConfiguration)),
			FlowControlStatus:             types.StringValue(string(dellNetworkPort.FlowControlStatus)),
			LinkStatus:                    types.StringValue(string(dellNetworkPort.LinkStatus)),
			PhysicalPortNumber:            types.StringValue(dellNetworkPort.PhysicalPortNumber),
			VendorID:                      types.StringValue(dellNetworkPort.VendorID),
			WakeOnLANEnabled:              types.BoolValue(dellNetworkPort.WakeOnLANEnabled),
			AssociatedNetworkAddresses:    newTypesStringList(dellNetworkPort.AssociatedNetworkAddresses),
			SupportedEthernetCapabilities: newSupportedEthernetCapabilities(dellNetworkPort.SupportedEthernetCapabilities),
			NetDevFuncMaxBWAlloc:          newNetDevFuncMaxBWAllocs(dellNetworkPort.NetDevFuncMaxBWAlloc),
			NetDevFuncMinBWAlloc:          newNetDevFuncMinBWAllocs(dellNetworkPort.NetDevFuncMinBWAlloc),
			SupportedLinkCapabilities:     newSupportedLinkCapabilities(dellNetworkPort.SupportedLinkCapabilitiesExtended),
			OemData:                       newNetworkPortOEM(dellNetworkPort.OemData),
		})
	}
	return networkPorts
}

func newNetDevFuncMaxBWAllocs(inputs []redfish.NetDevFuncMaxBWAlloc) []models.NetDevFuncMaxBWAlloc {
	out := make([]models.NetDevFuncMaxBWAlloc, 0)
	for _, input := range inputs {
		out = append(out, models.NetDevFuncMaxBWAlloc{
			MaxBWAllocPercent:     types.Int64Value(int64(input.MaxBWAllocPercent)),
			NetworkDeviceFunction: types.StringValue(input.NetworkDeviceFunction.ODataID),
		})
	}
	return out
}

func newNetDevFuncMinBWAllocs(inputs []redfish.NetDevFuncMinBWAlloc) []models.NetDevFuncMinBWAlloc {
	out := make([]models.NetDevFuncMinBWAlloc, 0)
	for _, input := range inputs {
		out = append(out, models.NetDevFuncMinBWAlloc{
			MinBWAllocPercent:     types.Int64Value(int64(input.MinBWAllocPercent)),
			NetworkDeviceFunction: types.StringValue(input.NetworkDeviceFunction.ODataID),
		})
	}
	return out
}

func newNetworkPortOEM(input dell.NetworkPortOEM) models.NetworkPortOEM {
	return models.NetworkPortOEM{
		DellNetworkTransceiver: newDellNetworkTransceiver(input.Dell.DellNetworkTransceiver),
	}
}

func newDellNetworkTransceiver(input dell.NetworkTransceiver) *models.DellNetworkTransceiver {
	if input.ID == "" {
		return nil
	}
	return &models.DellNetworkTransceiver{
		ODataID:           types.StringValue(input.ODataID),
		ID:                types.StringValue(input.ID),
		DeviceDescription: types.StringValue(input.DeviceDescription),
		Name:              types.StringValue(input.Name),
		FQDD:              types.StringValue(input.FQDD),
		IdentifierType:    types.StringValue(input.IdentifierType),
		InterfaceType:     types.StringValue(input.InterfaceType),
		PartNumber:        types.StringValue(input.PartNumber),
		Revision:          types.StringValue(input.Revision),
		SerialNumber:      types.StringValue(input.SerialNumber),
		VendorName:        types.StringValue(input.VendorName),
	}
}

func newSupportedLinkCapabilities(inputs []dell.SupportedLinkCapabilityExtended) []models.SupportedLinkCapability {
	out := make([]models.SupportedLinkCapability, 0)
	for _, input := range inputs {
		out = append(out, models.SupportedLinkCapability{
			AutoSpeedNegotiation:  types.BoolValue(input.AutoSpeedNegotiation),
			LinkNetworkTechnology: types.StringValue(string(input.LinkNetworkTechnology)),
			LinkSpeedMbps:         types.Int64Value(int64(input.LinkSpeedMbps)),
		})
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

func newTypesStringList(input []string) []types.String {
	out := make([]types.String, 0)
	for _, i := range input {
		out = append(out, types.StringValue(i))
	}
	return out
}

func newSupportedEthernetCapabilities(inputs []redfish.SupportedEthernetCapabilities) []types.String {
	out := make([]types.String, 0)
	for _, input := range inputs {
		out = append(out, types.StringValue(string(input)))
	}
	return out
}

func newNetworkDeviceFunctions(deviceFunctions []*redfish.NetworkDeviceFunction) []models.NetworkDeviceFunction {
	networkDeviceFunctions := make([]models.NetworkDeviceFunction, 0)
	for _, deviceFunction := range deviceFunctions {
		// Get Dell NetworkDeviceFunction extension
		dellDeviceFunction, _ := dell.NetworkDeviceFunction(deviceFunction)

		networkDeviceFunctions = append(networkDeviceFunctions, models.NetworkDeviceFunction{
			ODataID:                        types.StringValue(dellDeviceFunction.ODataID),
			ID:                             types.StringValue(dellDeviceFunction.ID),
			Name:                           types.StringValue(dellDeviceFunction.Name),
			Status:                         newNetworkStatus(dellDeviceFunction.Status),
			Description:                    types.StringValue(dellDeviceFunction.Description),
			MaxVirtualFunctions:            types.Int64Value(int64(dellDeviceFunction.MaxVirtualFunctions)),
			NetDevFuncType:                 types.StringValue(string(dellDeviceFunction.NetDevFuncType)),
			NetDevFuncCapabilities:         newNetDevFuncCapabilities(dellDeviceFunction.NetDevFuncCapabilities),
			Ethernet:                       newEthernet(dellDeviceFunction.DellEthernet),
			FibreChannel:                   newFibreChannel(dellDeviceFunction.FibreChannel),
			ISCSIBoot:                      newISCSIBoot(dellDeviceFunction.ISCSIBoot),
			PhysicalPortAssignment:         types.StringValue(dellDeviceFunction.DellPhysicalPortAssignment.ODataID),
			AssignablePhysicalPorts:        newEntityStringList(dellDeviceFunction.DellAssignablePhysicalPorts),
			AssignablePhysicalNetworkPorts: newEntityStringList(dellDeviceFunction.DellAssignablePhysicalNetworkPorts),
			OemData:                        newNetworkDeviceFunctionOEM(dellDeviceFunction.OemData),
		})
	}
	return networkDeviceFunctions
}

func newNetworkDeviceFunctionOEM(input dell.NetworkDeviceFunctionOEM) models.NetworkDeviceFunctionOEM {
	return models.NetworkDeviceFunctionOEM{
		DellNIC:             newDellNIC(input.Dell.DellNIC),
		DellNICPortMetrics:  newDellNICPortMetrics(input.Dell.DellNICPortMetrics),
		DellNICCapabilities: newDellNICCapabilities(input.Dell.DellNICCapabilities),
		DellFC:              newDellFC(input.Dell.DellFC),
		DellFCPortMetrics:   newDellFCPortMetrics(input.Dell.DellFCPortMetrics),
		DellFCCapabilities:  newDellFCCapabilities(input.Dell.DellFCCapabilities),
	}
}

func newDellFCPortMetrics(input dell.FCPortMetrics) *models.DellFCPortMetrics {
	if input.ID == "" {
		return nil
	}
	return &models.DellFCPortMetrics{
		ODataID:             types.StringValue(input.ODataID),
		ID:                  types.StringValue(input.ID),
		Name:                types.StringValue(input.Name),
		FCInvalidCRCs:       types.Int64Value(int64(input.FCInvalidCRCs)),
		FCLinkFailures:      types.Int64Value(int64(input.FCLinkFailures)),
		FCLossOfSignals:     types.Int64Value(int64(input.FCLossOfSignals)),
		FCRxKBCount:         types.Int64Value(int64(input.FCRxKBCount)),
		FCRxSequences:       types.Int64Value(int64(input.FCRxSequences)),
		FCRxTotalFrames:     types.Int64Value(int64(input.FCRxTotalFrames)),
		FCTxKBCount:         types.Int64Value(int64(input.FCTxKBCount)),
		FCTxSequences:       types.Int64Value(int64(input.FCTxSequences)),
		FCTxTotalFrames:     types.Int64Value(int64(input.FCTxTotalFrames)),
		OSDriverState:       types.StringValue(string(input.OSDriverState)),
		PortStatus:          types.StringValue(string(input.PortStatus)),
		RXInputPowerStatus:  types.StringValue(string(input.RXInputPowerStatus)),
		RXInputPowermW:      types.NumberValue(big.NewFloat(input.RXInputPowermW)),
		TXBiasCurrentStatus: types.StringValue(string(input.TXBiasCurrentStatus)),
		TXBiasCurrentmW:     types.NumberValue(big.NewFloat(input.TXBiasCurrentmW)),
		TXOutputPowerStatus: types.StringValue(string(input.TXOutputPowerStatus)),
		TXOutputPowermW:     types.NumberValue(big.NewFloat(input.TXOutputPowermW)),
		TemperatureStatus:   types.StringValue(string(input.TemperatureStatus)),
		TemperatureCelsius:  types.NumberValue(big.NewFloat(input.TemperatureCelsius)),
		VoltageStatus:       types.StringValue(string(input.VoltageStatus)),
		VoltageValueVolts:   types.NumberValue(big.NewFloat(input.VoltageValueVolts)),
	}
}

func newDellFC(input dell.FC) *models.DellFC {
	if input.ID == "" {
		return nil
	}
	return &models.DellFC{
		ODataID:                 types.StringValue(input.ODataID),
		ID:                      types.StringValue(input.ID),
		Name:                    types.StringValue(input.Name),
		Bus:                     types.Int64Value(int64(input.Bus)),
		CableLengthMetres:       types.Int64Value(int64(input.CableLengthMetres)),
		ChipType:                types.StringValue(input.ChipType),
		Device:                  types.Int64Value(int64(input.Device)),
		DeviceDescription:       types.StringValue(input.DeviceDescription),
		DeviceName:              types.StringValue(input.DeviceName),
		EFIVersion:              types.StringValue(input.EFIVersion),
		FCTapeEnable:            types.StringValue(input.FCTapeEnable),
		FCOSDriverVersion:       types.StringValue(input.FCOSDriverVersion),
		FCoEOSDriverVersion:     types.StringValue(input.FCoEOSDriverVersion),
		FabricLoginRetryCount:   types.Int64Value(int64(input.FabricLoginRetryCount)),
		FabricLoginTimeout:      types.Int64Value(int64(input.FabricLoginTimeout)),
		FamilyVersion:           types.StringValue(input.FamilyVersion),
		FramePayloadSize:        types.StringValue(input.FramePayloadSize),
		Function:                types.Int64Value(int64(input.Function)),
		HardZoneAddress:         types.Int64Value(int64(input.HardZoneAddress)),
		HardZoneEnable:          types.StringValue(input.HardZoneEnable),
		IdentifierType:          types.StringValue(input.IdentifierType),
		ISCSIOSDriverVersion:    types.StringValue(input.ISCSIOSDriverVersion),
		LanDriverVersion:        types.StringValue(input.LanDriverVersion),
		LinkDownTimeout:         types.Int64Value(int64(input.LinkDownTimeout)),
		LoopResetDelay:          types.Int64Value(int64(input.LoopResetDelay)),
		PartNumber:              types.StringValue(input.PartNumber),
		PortDownRetryCount:      types.Int64Value(int64(input.PortDownRetryCount)),
		PortDownTimeout:         types.Int64Value(int64(input.PortDownTimeout)),
		PortLoginRetryCount:     types.Int64Value(int64(input.PortLoginRetryCount)),
		PortLoginTimeout:        types.Int64Value(int64(input.PortLoginTimeout)),
		ProductName:             types.StringValue(input.ProductName),
		RDMAOSDriverVersion:     types.StringValue(input.RDMAOSDriverVersion),
		Revision:                types.StringValue(input.Revision),
		SecondFCTargetLUN:       types.Int64Value(int64(input.SecondFCTargetLUN)),
		SecondFCTargetWWPN:      types.StringValue(input.SecondFCTargetWWPN),
		SerialNumber:            types.StringValue(input.SerialNumber),
		TransceiverPartNumber:   types.StringValue(input.TransceiverPartNumber),
		TransceiverSerialNumber: types.StringValue(input.TransceiverSerialNumber),
		TransceiverVendorName:   types.StringValue(input.TransceiverVendorName),
		VendorName:              types.StringValue(input.VendorName),
	}
}

func newDellFCCapabilities(input dell.FCCapabilities) *models.DellFCCapabilities {
	if input.ID == "" {
		return nil
	}
	return &models.DellFCCapabilities{
		ODataID:                        types.StringValue(input.ODataID),
		ID:                             types.StringValue(input.ID),
		Name:                           types.StringValue(input.Name),
		FCMaxNumberExchanges:           types.Int64Value(int64(input.FCMaxNumberExchanges)),
		FCMaxNumberOutStandingCommands: types.Int64Value(int64(input.FCMaxNumberOutStandingCommands)),
		FeatureLicensingSupport:        types.StringValue(input.FeatureLicensingSupport),
		FlexAddressingSupport:          types.StringValue(input.FlexAddressingSupport),
		OnChipThermalSensor:            types.StringValue(input.OnChipThermalSensor),
		PersistencePolicySupport:       types.StringValue(input.PersistencePolicySupport),
		UEFISupport:                    types.StringValue(input.UEFISupport),
	}
}

func newDellNICCapabilities(input dell.NICCapabilities) *models.DellNICCapabilities {
	if input.ID == "" {
		return nil
	}
	return &models.DellNICCapabilities{
		ODataID:                          types.StringValue(input.ODataID),
		ID:                               types.StringValue(input.ID),
		Name:                             types.StringValue(input.Name),
		BPESupport:                       types.StringValue(input.BPESupport),
		CongestionNotification:           types.StringValue(input.CongestionNotification),
		DCBExchangeProtocol:              types.StringValue(input.DCBExchangeProtocol),
		ETS:                              types.StringValue(input.ETS),
		EVBModesSupport:                  types.StringValue(input.EVBModesSupport),
		FCoEBootSupport:                  types.StringValue(input.FCoEBootSupport),
		FCoEMaxIOsPerSession:             types.Int64Value(int64(input.FCoEMaxIOsPerSession)),
		FCoEMaxNPIVPerPort:               types.Int64Value(int64(input.FCoEMaxNPIVPerPort)),
		FCoEMaxNumberExchanges:           types.Int64Value(int64(input.FCoEMaxNumberExchanges)),
		FCoEMaxNumberLogins:              types.Int64Value(int64(input.FCoEMaxNumberLogins)),
		FCoEMaxNumberOfFCTargets:         types.Int64Value(int64(input.FCoEMaxNumberOfFCTargets)),
		FCoEMaxNumberOutStandingCommands: types.Int64Value(int64(input.FCoEMaxNumberOutStandingCommands)),
		FCoEOffloadSupport:               types.StringValue(input.FCoEOffloadSupport),
		FeatureLicensingSupport:          types.StringValue(input.FeatureLicensingSupport),
		FlexAddressingSupport:            types.StringValue(input.FlexAddressingSupport),
		IPSecOffloadSupport:              types.StringValue(input.IPSecOffloadSupport),
		MACSecSupport:                    types.StringValue(input.MACSecSupport),
		NWManagementPassThrough:          types.StringValue(input.NWManagementPassThrough),
		NicPartitioningSupport:           types.StringValue(input.NicPartitioningSupport),
		OSBMCManagementPassThrough:       types.StringValue(input.OSBMCManagementPassThrough),
		OnChipThermalSensor:              types.StringValue(input.OnChipThermalSensor),
		OpenFlowSupport:                  types.StringValue(input.OpenFlowSupport),
		PXEBootSupport:                   types.StringValue(input.PXEBootSupport),
		PartitionWOLSupport:              types.StringValue(input.PartitionWOLSupport),
		PersistencePolicySupport:         types.StringValue(input.PersistencePolicySupport),
		PriorityFlowControl:              types.StringValue(input.PriorityFlowControl),
		RDMASupport:                      types.StringValue(input.RDMASupport),
		RemotePHY:                        types.StringValue(input.RemotePHY),
		TCPChimneySupport:                types.StringValue(input.TCPChimneySupport),
		TCPOffloadEngineSupport:          types.StringValue(input.TCPOffloadEngineSupport),
		VEB:                              types.StringValue(input.VEB),
		VEBVEPAMultiChannel:              types.StringValue(input.VEBVEPAMultiChannel),
		VEBVEPASingleChannel:             types.StringValue(input.VEBVEPASingleChannel),
		VirtualLinkControl:               types.StringValue(input.VirtualLinkControl),
		ISCSIBootSupport:                 types.StringValue(input.ISCSIBootSupport),
		ISCSIOffloadSupport:              types.StringValue(input.ISCSIOffloadSupport),
		UEFISupport:                      types.StringValue(input.UEFISupport),
	}
}

func newDellNICPortMetrics(input dell.NICPortMetrics) *models.DellNICPortMetrics {
	if input.ID == "" {
		return nil
	}
	return &models.DellNICPortMetrics{
		ODataID:                      types.StringValue(input.ODataID),
		ID:                           types.StringValue(input.ID),
		Name:                         types.StringValue(input.Name),
		DiscardedPkts:                types.Int64Value(int64(input.DiscardedPkts)),
		FQDD:                         types.StringValue(input.FQDD),
		OSDriverState:                types.StringValue(input.OSDriverState),
		RxBytes:                      types.Int64Value(int64(input.RxBytes)),
		RxBroadcast:                  types.Int64Value(int64(input.RxBroadcast)),
		RxErrorPktAlignmentErrors:    types.Int64Value(int64(input.RxErrorPktAlignmentErrors)),
		RxErrorPktFCSErrors:          types.Int64Value(int64(input.RxErrorPktFCSErrors)),
		RxJabberPkt:                  types.Int64Value(int64(input.RxJabberPkt)),
		RxMutlicastPackets:           types.Int64Value(int64(input.RxMutlicastPackets)),
		RxPauseXOFFFrames:            types.Int64Value(int64(input.RxPauseXOFFFrames)),
		RxPauseXONFrames:             types.Int64Value(int64(input.RxPauseXONFrames)),
		RxRuntPkt:                    types.Int64Value(int64(input.RxRuntPkt)),
		RxUnicastPackets:             types.Int64Value(int64(input.RxUnicastPackets)),
		TxBroadcast:                  types.Int64Value(int64(input.TxBroadcast)),
		TxBytes:                      types.Int64Value(int64(input.TxBytes)),
		TxErrorPktExcessiveCollision: types.Int64Value(int64(input.TxErrorPktExcessiveCollision)),
		TxErrorPktLateCollision:      types.Int64Value(int64(input.TxErrorPktLateCollision)),
		TxErrorPktMultipleCollision:  types.Int64Value(int64(input.TxErrorPktMultipleCollision)),
		TxErrorPktSingleCollision:    types.Int64Value(int64(input.TxErrorPktSingleCollision)),
		TxMutlicastPackets:           types.Int64Value(int64(input.TxMutlicastPackets)),
		TxPauseXOFFFrames:            types.Int64Value(int64(input.TxPauseXOFFFrames)),
		TxPauseXONFrames:             types.Int64Value(int64(input.TxPauseXONFrames)),
		StartStatisticTime:           types.StringValue(input.StartStatisticTime),
		StatisticTime:                types.StringValue(input.StatisticTime),
		FCCRCErrorCount:              types.Int64Value(int64(input.FCCRCErrorCount)),
		FCOELinkFailures:             types.Int64Value(int64(input.FCOELinkFailures)),
		FCOEPktRxCount:               types.Int64Value(int64(input.FCOEPktRxCount)),
		FCOEPktTxCount:               types.Int64Value(int64(input.FCOEPktTxCount)),
		FCOERxPktDroppedCount:        types.Int64Value(int64(input.FCOERxPktDroppedCount)),
		LanFCSRxErrors:               types.Int64Value(int64(input.LanFCSRxErrors)),
		LanUnicastPktRXCount:         types.Int64Value(int64(input.LanUnicastPktRXCount)),
		LanUnicastPktTXCount:         types.Int64Value(int64(input.LanUnicastPktTXCount)),
		RDMARxTotalBytes:             types.Int64Value(int64(input.RDMARxTotalBytes)),
		RDMARxTotalPackets:           types.Int64Value(int64(input.RDMARxTotalPackets)),
		RDMATotalProtectionErrors:    types.Int64Value(int64(input.RDMATotalProtectionErrors)),
		RDMATotalProtocolErrors:      types.Int64Value(int64(input.RDMATotalProtocolErrors)),
		RDMATxTotalBytes:             types.Int64Value(int64(input.RDMATxTotalBytes)),
		RDMATxTotalPackets:           types.Int64Value(int64(input.RDMATxTotalPackets)),
		RDMATxTotalReadReqPkts:       types.Int64Value(int64(input.RDMATxTotalReadReqPkts)),
		RDMATxTotalSendPkts:          types.Int64Value(int64(input.RDMATxTotalSendPkts)),
		RDMATxTotalWritePkts:         types.Int64Value(int64(input.RDMATxTotalWritePkts)),
		TxUnicastPackets:             types.Int64Value(int64(input.TxUnicastPackets)),
		PartitionLinkStatus:          types.StringValue(input.PartitionLinkStatus),
		PartitionOSDriverState:       types.StringValue(input.PartitionOSDriverState),
		RXInputPowerStatus:           types.StringValue(input.RXInputPowerStatus),
		RxFalseCarrierDetection:      types.Int64Value(int64(input.RxFalseCarrierDetection)),
		TXBiasCurrentStatus:          types.StringValue(input.TXBiasCurrentStatus),
		TXOutputPowerStatus:          types.StringValue(input.TXOutputPowerStatus),
		TemperatureStatus:            types.StringValue(input.TemperatureStatus),
		VoltageStatus:                types.StringValue(input.VoltageStatus),
		RXInputPowermW:               types.NumberValue(big.NewFloat(input.RXInputPowermW)),
		TXBiasCurrentmA:              types.NumberValue(big.NewFloat(input.TXBiasCurrentmA)),
		TXOutputPowermW:              types.NumberValue(big.NewFloat(input.TXOutputPowermW)),
		TemperatureCelsius:           types.NumberValue(big.NewFloat(input.TemperatureCelsius)),
		VoltageValueVolts:            types.NumberValue(big.NewFloat(input.VoltageValueVolts)),
	}
}

func newDellNIC(input dell.NIC) *models.DellNIC {
	if input.ID == "" {
		return nil
	}
	return &models.DellNIC{
		ODataID:                  types.StringValue(input.ODataID),
		ID:                       types.StringValue(input.ID),
		DeviceDescription:        types.StringValue(input.DeviceDescription),
		Name:                     types.StringValue(input.Name),
		FQDD:                     types.StringValue(input.FQDD),
		IdentifierType:           types.StringValue(input.IdentifierType),
		PartNumber:               types.StringValue(input.PartNumber),
		Revision:                 types.StringValue(input.Revision),
		SerialNumber:             types.StringValue(input.SerialNumber),
		VendorName:               types.StringValue(input.VendorName),
		BusNumber:                types.Int64Value(int64(input.BusNumber)),
		ControllerBIOSVersion:    types.StringValue(input.ControllerBIOSVersion),
		DataBusWidth:             types.StringValue(input.DataBusWidth),
		EFIVersion:               types.StringValue(input.EFIVersion),
		FCoEOffloadMode:          types.StringValue(input.FCoEOffloadMode),
		FCOSDriverVersion:        types.StringValue(input.FCOSDriverVersion),
		FamilyVersion:            types.StringValue(input.FamilyVersion),
		InstanceID:               types.StringValue(input.InstanceID),
		LastSystemInventoryTime:  types.StringValue(input.LastSystemInventoryTime),
		LinkDuplex:               types.StringValue(input.LinkDuplex),
		LastUpdateTime:           types.StringValue(input.LastUpdateTime),
		MediaType:                types.StringValue(input.MediaType),
		NICMode:                  types.StringValue(input.NICMode),
		PCIDeviceID:              types.StringValue(input.PCIDeviceID),
		PCIVendorID:              types.StringValue(input.PCIVendorID),
		PCISubDeviceID:           types.StringValue(input.PCISubDeviceID),
		PCISubVendorID:           types.StringValue(input.PCISubVendorID),
		ProductName:              types.StringValue(input.ProductName),
		Protocol:                 types.StringValue(input.Protocol),
		SNAPIState:               types.StringValue(input.SNAPIState),
		SNAPISupport:             types.StringValue(input.SNAPISupport),
		SlotLength:               types.StringValue(input.SlotLength),
		SlotType:                 types.StringValue(input.SlotType),
		VPISupport:               types.StringValue(input.VPISupport),
		ISCSIOffloadMode:         types.StringValue(input.ISCSIOffloadMode),
		TransceiverVendorName:    types.StringValue(input.TransceiverVendorName),
		CableLengthMetres:        types.Int64Value(int64(input.CableLengthMetres)),
		PermanentFCOEMACAddress:  types.StringValue(input.PermanentFCOEMACAddress),
		PermanentiSCSIMACAddress: types.StringValue(input.PermanentiSCSIMACAddress),
		TransceiverPartNumber:    types.StringValue(input.TransceiverPartNumber),
		TransceiverSerialNumber:  types.StringValue(input.TransceiverSerialNumber),
	}
}

func newEthernet(input dell.Ethernet) *models.Ethernet {
	if input.MACAddress == "" && input.PermanentMACAddress == "" {
		return nil
	}
	return &models.Ethernet{
		MACAddress:          types.StringValue(input.MACAddress),
		MTUSize:             types.Int64Value(int64(input.MTUSize)),
		PermanentMACAddress: types.StringValue(input.PermanentMACAddress),
		VLAN:                newVLAN(input.VLAN),
	}
}

func newVLAN(input dell.VLAN) models.VLAN {
	if !input.VLANEnabled {
		return models.VLAN{VLANEnabled: types.BoolValue(input.VLANEnabled)}
	}
	return models.VLAN{
		VLANEnabled: types.BoolValue(input.VLANEnabled),
		VLANID:      types.Int64Value(int64(input.VLANID)),
	}
}

func newFibreChannel(input redfish.FibreChannel) *models.FibreChannel {
	if input.PermanentWWNN == "" && input.PermanentWWPN == "" && input.WWNN == "" && input.WWPN == "" {
		return nil
	}
	return &models.FibreChannel{
		FibreChannelId:        types.StringValue(input.FibreChannelID),
		AllowFIPVLANDiscovery: types.BoolValue(input.AllowFIPVLANDiscovery),
		BootTargets:           newBootTargets(input.BootTargets),
		FCoEActiveVLANId:      types.Int64Value(int64(input.FCoEActiveVLANID)),
		FCoELocalVLANId:       types.Int64Value(int64(input.FCoELocalVLANID)),
		PermanentWWNN:         types.StringValue(input.PermanentWWNN),
		PermanentWWPN:         types.StringValue(input.PermanentWWPN),
		WWNN:                  types.StringValue(input.WWNN),
		WWNSource:             types.StringValue(string(input.WWNSource)),
		WWPN:                  types.StringValue(input.WWPN),
	}
}

func newBootTargets(inputs []redfish.BootTargets) []models.BootTarget {
	out := make([]models.BootTarget, 0)
	for _, input := range inputs {
		out = append(out, models.BootTarget{
			BootPriority: types.Int64Value(int64(input.BootPriority)),
			LUNID:        types.StringValue(input.LUNID),
			WWPN:         types.StringValue(input.WWPN),
		})
	}
	return out
}

func newISCSIBoot(input redfish.ISCSIBoot) *models.ISCSIBoot {
	if input.InitiatorName == "" && input.InitiatorDefaultGateway == "" && input.PrimaryDNS == "" && input.IPAddressType == "" {
		return nil
	}
	return &models.ISCSIBoot{
		AuthenticationMethod:       types.StringValue(string(input.AuthenticationMethod)),
		CHAPSecret:                 types.StringValue(input.CHAPSecret),
		CHAPUsername:               types.StringValue(input.CHAPUsername),
		IPAddressType:              types.StringValue(string(input.IPAddressType)),
		IPMaskDNSViaDHCP:           types.BoolValue(input.IPMaskDNSViaDHCP),
		InitiatorDefaultGateway:    types.StringValue(input.InitiatorDefaultGateway),
		InitiatorIPAddress:         types.StringValue(input.InitiatorIPAddress),
		InitiatorName:              types.StringValue(input.InitiatorName),
		InitiatorNetmask:           types.StringValue(input.InitiatorNetmask),
		MutualCHAPSecret:           types.StringValue(input.MutualCHAPSecret),
		MutualCHAPUsername:         types.StringValue(input.MutualCHAPUsername),
		PrimaryDNS:                 types.StringValue(input.PrimaryDNS),
		PrimaryLUN:                 types.Int64Value(int64(input.PrimaryLUN)),
		PrimaryTargetIPAddress:     types.StringValue(input.PrimaryTargetIPAddress),
		PrimaryTargetName:          types.StringValue(input.PrimaryTargetName),
		PrimaryTargetTCPPort:       types.Int64Value(int64(input.PrimaryTargetTCPPort)),
		PrimaryVLANEnable:          types.BoolValue(input.PrimaryVLANEnable),
		PrimaryVLANId:              types.Int64Value(int64(input.PrimaryVLANID)),
		RouterAdvertisementEnabled: types.BoolValue(input.RouterAdvertisementEnabled),
		SecondaryDNS:               types.StringValue(input.SecondaryDNS),
		SecondaryLUN:               types.Int64Value(int64(input.SecondaryLUN)),
		SecondaryTargetIPAddress:   types.StringValue(input.SecondaryTargetIPAddress),
		SecondaryTargetName:        types.StringValue(input.SecondaryTargetName),
		SecondaryTargetTCPPort:     types.Int64Value(int64(input.SecondaryTargetTCPPort)),
		SecondaryVLANEnable:        types.BoolValue(input.SecondaryVLANEnable),
		SecondaryVLANId:            types.Int64Value(int64(input.SecondaryVLANID)),
		TargetInfoViaDHCP:          types.BoolValue(input.TargetInfoViaDHCP),
	}
}

func newNetDevFuncCapabilities(inputs []redfish.NetworkDeviceTechnology) []types.String {
	out := make([]types.String, 0)
	for _, input := range inputs {
		out = append(out, types.StringValue(string(input)))
	}
	return out
}

func loadNICAttributesState(service *gofish.Service, d *models.NICDatasource) diag.Diagnostics {
	var idracAttributesState models.DellIdracAttributes
	if diags := readDatasourceRedfishDellIdracAttributes(service, &idracAttributesState); diags.HasError() {
		return diags
	}

	attributesToReturn := make(map[string]attr.Value)
	for k, v := range idracAttributesState.Attributes.Elements() {
		if strings.HasPrefix(k, "NIC.") {
			attributesToReturn[k] = v
		}
	}

	d.NICAttributes = types.MapValueMust(types.StringType, attributesToReturn)
	return nil
}

// NICFilterSchema to construct schema of nic filter.
func NICFilterSchema() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"systems": schema.ListNestedAttribute{
			Optional:    true,
			Description: "Filter for systems, nework adapters, network ports and network device functions",
			Validators: []validator.List{
				listvalidator.UniqueValues(),
			},
			NestedObject: schema.NestedAttributeObject{
				Attributes: map[string]schema.Attribute{
					"system_id": schema.StringAttribute{
						Required:    true,
						Description: "Filter for systems",
					},
					"network_adapters": schema.ListNestedAttribute{
						Optional:    true,
						Description: "Filter for nework adapters, network ports and network device functions",
						Validators: []validator.List{
							listvalidator.UniqueValues(),
						},
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"network_adapter_id": schema.StringAttribute{
									Required:    true,
									Description: "Filter for network adapters",
								},
								"network_port_ids": schema.SetAttribute{
									Optional:    true,
									ElementType: types.StringType,
									Description: "Filter for network ports",
								},
								"network_device_function_ids": schema.SetAttribute{
									Optional:    true,
									ElementType: types.StringType,
									Description: "Filter for network device functions",
								},
							},
						},
					},
				},
			},
		},
	}
}

// NICDatasourceSchema to define the NIC data-source schema.
func NICDatasourceSchema() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		NICComponmentSchemaID: schema.StringAttribute{
			MarkdownDescription: "ID of the network interface cards data-source",
			Description:         "ID of the network interface cards data-source",
			Computed:            true,
		},
		"nic_attributes": schema.MapAttribute{
			MarkdownDescription: "nic.* attributes in Dell iDRAC attributes.",
			Description:         "nic.* attributes in Dell iDRAC attributes.",
			ElementType:         types.StringType,
			Computed:            true,
		},
		"network_interfaces": schema.ListNestedAttribute{
			MarkdownDescription: "List of network interface cards(NIC) fetched.",
			Description:         "List of network interface cards(NIC) fetched.",
			NestedObject: schema.NestedAttributeObject{
				Attributes: NICSchema(),
			},
			Computed: true,
		},
	}
}

// NICSchema is a function that returns the schema for NIC.
func NICSchema() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		NICComponmentSchemaOdataID: schema.StringAttribute{
			Computed:    true,
			Description: "OData ID for the NIC instance",
		},
		NICComponmentSchemaID: schema.StringAttribute{
			Computed:    true,
			Description: "ID of the NIC data-source",
		},
		NICComponmentSchemaName: schema.StringAttribute{
			Computed:    true,
			Description: "Name of the NIC data-source",
		},
		NICComponmentSchemaDescription: schema.StringAttribute{
			Computed:    true,
			Description: "Description of the NIC data-source",
		},
		NICComponmentSchemaStatus: schema.SingleNestedAttribute{
			MarkdownDescription: "The status and health of a resource and its children",
			Description:         "The status and health of a resource and its children.",
			Computed:            true,
			Attributes:          NetworkStatusSchema(),
		},
		"network_adapter": schema.SingleNestedAttribute{
			MarkdownDescription: "Network adapter fetched",
			Description:         "Network adapter fetched",
			Computed:            true,
			Attributes:          NetworkAdapterDataSourceSchema(),
		},
		"network_ports": schema.ListNestedAttribute{
			MarkdownDescription: "List of network ports fetched",
			Description:         "List of network ports fetched",
			Computed:            true,
			NestedObject: schema.NestedAttributeObject{
				Attributes: NetworkPortDataSourceSchema(),
			},
		},
		"network_device_functions": schema.ListNestedAttribute{
			MarkdownDescription: "List of network device functions fetched",
			Description:         "List of network device functions fetched",
			Computed:            true,
			NestedObject: schema.NestedAttributeObject{
				Attributes: NetworkDeviceFunctionDataSourceSchema(),
			},
		},
	}
}

// NetworkStatusSchema is a function that returns the schema for Status
func NetworkStatusSchema() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"health": schema.StringAttribute{
			MarkdownDescription: "health",
			Description:         "health",
			Computed:            true,
		},
		"health_rollup": schema.StringAttribute{
			MarkdownDescription: "health rollup",
			Description:         "health rollup",
			Computed:            true,
		},
		"state": schema.StringAttribute{
			MarkdownDescription: "state of the storage controller",
			Description:         "state of the storage controller",
			Computed:            true,
		},
	}
}

// NetworkAdapterDataSourceSchema is a function that returns the schema for network adapter.
func NetworkAdapterDataSourceSchema() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		NICComponmentSchemaOdataID: schema.StringAttribute{
			Computed:    true,
			Description: "OData ID for the network adapter",
		},
		NICComponmentSchemaID: schema.StringAttribute{
			Computed:    true,
			Description: "ID of the network adapter",
		},
		NICComponmentSchemaName: schema.StringAttribute{
			Computed:    true,
			Description: "Name of the network adapter",
		},
		NICComponmentSchemaDescription: schema.StringAttribute{
			Computed:    true,
			Description: "Description of the network adapter",
		},
		NICComponmentSchemaStatus: schema.SingleNestedAttribute{
			MarkdownDescription: "The status and health of a resource and its children",
			Description:         "The status and health of a resource and its children.",
			Computed:            true,
			Attributes:          NetworkStatusSchema(),
		},
		"manufacturer": schema.StringAttribute{
			Computed:    true,
			Description: "The manufacturer or OEM of this network adapter",
		},
		"model": schema.StringAttribute{
			Computed:    true,
			Description: "The model string for this network adapter",
		},
		NICComponmentSchemaPartNumber: schema.StringAttribute{
			Computed:    true,
			Description: "Part number for this network adapter",
		},
		NICComponmentSchemaSerialNumber: schema.StringAttribute{
			Computed:    true,
			Description: "The serial number for this network adapter",
		},
		"controllers": schema.ListNestedAttribute{
			Description:         "A network controller ASIC that makes up part of a network adapter",
			MarkdownDescription: "A network controller ASIC that makes up part of a network adapter",
			Computed:            true,
			NestedObject: schema.NestedAttributeObject{
				Attributes: NetworkControllerSchema(),
			},
		},
	}
}

// NetworkControllerSchema is a function that returns the schema for network controller.
func NetworkControllerSchema() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"firmware_package_version": schema.StringAttribute{
			MarkdownDescription: "The version of the user-facing firmware package",
			Description:         "The version of the user-facing firmware package",
			Computed:            true,
		},
		"controller_capabilities": schema.SingleNestedAttribute{
			MarkdownDescription: "The capabilities of this controller",
			Description:         "The capabilities of this controller",
			Computed:            true,
			Attributes:          ControllerCapabilitiesSchema(),
		},
	}
}

// ControllerCapabilitiesSchema is a function that returns the schema for controller capabilities.
func ControllerCapabilitiesSchema() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"data_center_bridging": schema.SingleNestedAttribute{
			MarkdownDescription: "Data center bridging (DCB) for capabilities of a controller",
			Description:         "Data center bridging (DCB) for capabilities of a controller",
			Computed:            true,
			Attributes:          DataCenterBridgingSchema(),
		},
		"npar": schema.SingleNestedAttribute{
			MarkdownDescription: "NIC Partitioning capability, status, and configuration for a controller",
			Description:         "NIC Partitioning capability, status, and configuration for a controller",
			Computed:            true,
			Attributes:          NparSchema(),
		},
		"npiv": schema.SingleNestedAttribute{
			MarkdownDescription: "N_Port ID Virtualization (NPIV) capabilities for a controller",
			Description:         "N_Port ID Virtualization (NPIV) capabilities for a controller",
			Computed:            true,
			Attributes:          NpivSchema(),
		},
		"virtualization_offload": schema.SingleNestedAttribute{
			MarkdownDescription: "A Virtualization offload capability of a controller",
			Description:         "A Virtualization offload capability of a controller",
			Computed:            true,
			Attributes:          VirtualizationOffloadSchema(),
		},
	}
}

// VirtualizationOffloadSchema is a function that returns the schema for virtualization offload.
func VirtualizationOffloadSchema() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"sriov": schema.SingleNestedAttribute{
			MarkdownDescription: "Single-root input/output virtualization (SR-IOV) capabilities",
			Description:         "Single-root input/output virtualization (SR-IOV) capabilities",
			Computed:            true,
			Attributes:          SriovSchema(),
		},
		"virtual_function": schema.SingleNestedAttribute{
			MarkdownDescription: "A virtual function of a controller",
			Description:         "A virtual function of a controller",
			Computed:            true,
			Attributes:          VirtualFunctionSchema(),
		},
	}
}

// VirtualFunctionSchema is a function that returns the schema for virtual function.
func VirtualFunctionSchema() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"device_max_count": schema.Int64Attribute{
			Description:         "The maximum number of virtual functions supported by this controller",
			MarkdownDescription: "The maximum number of virtual functions supported by this controller",
			Computed:            true,
		},
		"min_assignment_group_size": schema.Int64Attribute{
			Description: "The minimum number of virtual functions that can be allocated or " +
				"moved between physical functions for this controller",
			MarkdownDescription: "The minimum number of virtual functions that can be allocated or " +
				"moved between physical functions for this controller",
			Computed: true,
		},
		"network_port_max_count": schema.Int64Attribute{
			Description:         "The maximum number of virtual functions supported per network port for this controller",
			MarkdownDescription: "The maximum number of virtual functions supported per network port for this controller",
			Computed:            true,
		},
	}
}

// SriovSchema is a function that returns the schema for sriov.
func SriovSchema() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"sriov_vepa_capable": schema.BoolAttribute{
			Description: "An indication of whether this controller supports single root input/output virtualization (SR-IOV)" +
				"in Virtual Ethernet Port Aggregator (VEPA) mode",
			MarkdownDescription: "An indication of whether this controller supports single root input/output virtualization (SR-IOV)" +
				"in Virtual Ethernet Port Aggregator (VEPA) mode",
			Computed: true,
		},
	}
}

// NpivSchema is a function that returns the schema for NPIV.
func NpivSchema() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"max_device_logins": schema.Int64Attribute{
			Description: "The maximum number of N_Port ID Virtualization (NPIV) logins allowed simultaneously " +
				"from all ports on this controller",
			MarkdownDescription: "The maximum number of N_Port ID Virtualization (NPIV) logins allowed simultaneously " +
				"from all ports on this controller",
			Computed: true,
		},
		"max_port_logins": schema.Int64Attribute{
			Description:         "The maximum number of N_Port ID Virtualization (NPIV) logins allowed per physical port on this controller",
			MarkdownDescription: "The maximum number of N_Port ID Virtualization (NPIV) logins allowed per physical port on this controller",
			Computed:            true,
		},
	}
}

// NparSchema is a function that returns the schema for NPAR.
func NparSchema() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"npar_capable": schema.BoolAttribute{
			Description:         "An indication of whether the controller supports NIC function partitioning",
			MarkdownDescription: "An indication of whether the controller supports NIC function partitioning",
			Computed:            true,
		},
		"npar_enabled": schema.BoolAttribute{
			Description:         "An indication of whether NIC function partitioning is active on this controller",
			MarkdownDescription: "An indication of whether NIC function partitioning is active on this controller.",
			Computed:            true,
		},
	}
}

// DataCenterBridgingSchema is a function that returns the schema for data center bridging.
func DataCenterBridgingSchema() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"capable": schema.BoolAttribute{
			Description:         "An indication of whether this controller is capable of data center bridging (DCB)",
			MarkdownDescription: "An indication of whether this controller is capable of data center bridging (DCB)",
			Computed:            true,
		},
	}
}

// NetworkPortDataSourceSchema is a function that returns the schema for network port.
func NetworkPortDataSourceSchema() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		NICComponmentSchemaOdataID: schema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "OData ID for the network port",
			Description:         "OData ID for the network port",
		},
		NICComponmentSchemaID: schema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "ID of the network port",
			Description:         "ID of the network port",
		},
		NICComponmentSchemaName: schema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "name of the network port",
			Description:         "name of the network port",
		},
		NICComponmentSchemaDescription: schema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "description of the network port",
			Description:         "description of the network port",
		},
		NICComponmentSchemaStatus: schema.SingleNestedAttribute{
			MarkdownDescription: "status of the network port",
			Description:         "status of the network port",
			Computed:            true,
			Attributes:          NetworkStatusSchema(),
		},
		"active_link_technology": schema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "Network port active link technology",
			Description:         "Network port active link technology",
		},
		"flow_control_configuration": schema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "The locally configured 802.3x flow control setting for this network port",
			Description:         "The locally configured 802.3x flow control setting for this network port.",
		},
		"flow_control_status": schema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "The 802.3x flow control behavior negotiated with the link partner for this network port (Ethernet-only)",
			Description:         "The 802.3x flow control behavior negotiated with the link partner for this network port (Ethernet-only)",
		},
		"vendor_id": schema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "The vendor Identification for this port",
			Description:         "The vendor Identification for this port",
		},
		"link_status": schema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "The status of the link between this port and its link partner",
			Description:         "The status of the link between this port and its link partner",
		},
		"associated_network_addresses": schema.ListAttribute{
			Description: "An array of configured MAC or WWN network addresses that are associated with this network port, " +
				"including the programmed address of the lowest numbered network device function, the configured but not active address, " +
				"if applicable, the address for hardware port teaming, or other network addresses",
			MarkdownDescription: "An array of configured MAC or WWN network addresses that are associated with this network port, " +
				"including the programmed address of the lowest numbered network device function, the configured but not active address, " +
				"if applicable, the address for hardware port teaming, or other network addresses",
			ElementType: types.StringType,
			Computed:    true,
		},
		"supported_ethernet_capabilities": schema.ListAttribute{
			Description:         "The set of Ethernet capabilities that this port supports",
			MarkdownDescription: "The set of Ethernet capabilities that this port supports.",
			ElementType:         types.StringType,
			Computed:            true,
		},
		"current_link_speed_mbps": schema.Int64Attribute{
			Computed:            true,
			MarkdownDescription: "Network port current link speed",
			Description:         "Network port current link speed.",
		},
		"physical_port_number": schema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "The physical port number label for this port",
			Description:         "The physical port number label for this port",
		},
		"eee_enabled": schema.BoolAttribute{
			Computed:            true,
			MarkdownDescription: "An indication of whether IEEE 802.3az Energy-Efficient Ethernet (EEE) is enabled for this network port",
			Description:         "An indication of whether IEEE 802.3az Energy-Efficient Ethernet (EEE) is enabled for this network port",
		},
		"wake_on_lan_enabled": schema.BoolAttribute{
			Computed:            true,
			MarkdownDescription: "An indication of whether Wake on LAN (WoL) is enabled for this network port",
			Description:         "An indication of whether Wake on LAN (WoL) is enabled for this network port",
		},
		"net_dev_func_max_bw_alloc": schema.ListNestedAttribute{
			MarkdownDescription: "A maximum bandwidth allocation percentage for a network device functions associated a port",
			Description:         "A maximum bandwidth allocation percentage for a network device functions associated a port",
			Computed:            true,
			NestedObject: schema.NestedAttributeObject{
				Attributes: NetDevFuncMaxBWAllocSchema(),
			},
		},
		"net_dev_func_min_bw_alloc": schema.ListNestedAttribute{
			MarkdownDescription: "A minimum bandwidth allocation percentage for a network device functions associated a port",
			Description:         "A minimum bandwidth allocation percentage for a network device functions associated a port",
			Computed:            true,
			NestedObject: schema.NestedAttributeObject{
				Attributes: NetDevFuncMinBWAllocSchema(),
			},
		},
		"supported_link_capabilities": schema.ListNestedAttribute{
			MarkdownDescription: "The link capabilities of an associated port",
			Description:         "The link capabilities of an associated port",
			Computed:            true,
			NestedObject: schema.NestedAttributeObject{
				Attributes: SupportedLinkCapabilitySchema(),
			},
		},
		"oem": schema.SingleNestedAttribute{
			MarkdownDescription: "The OEM extension for this network port",
			Description:         "The OEM extension for this network port",
			Computed:            true,
			Attributes:          NetworkPortOemSchema(),
		},
	}
}

// NetDevFuncMaxBWAllocSchema is a function that returns the schema for net dev func max bw alloc.
func NetDevFuncMaxBWAllocSchema() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"max_bw_alloc_percent": schema.Int64Attribute{
			MarkdownDescription: "The maximum bandwidth allocation percentage allocated to the corresponding network device function instance",
			Description:         "The maximum bandwidth allocation percentage allocated to the corresponding network device function instance",
			Computed:            true,
		},
		"network_device_function": schema.StringAttribute{
			MarkdownDescription: "List of network device functions for NetDevFuncMaxBWAlloc associated with this port",
			Description:         "List of network device functions for NetDevFuncMaxBWAlloc associated with this port",
			Computed:            true,
		},
	}
}

// NetDevFuncMinBWAllocSchema is a function that returns the schema for net dev func min bw alloc.
func NetDevFuncMinBWAllocSchema() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"min_bw_alloc_percent": schema.Int64Attribute{
			MarkdownDescription: "The minimum bandwidth allocation percentage allocated to the corresponding network device function instance",
			Description:         "The minimum bandwidth allocation percentage allocated to the corresponding network device function instance",
			Computed:            true,
		},
		"network_device_function": schema.StringAttribute{
			MarkdownDescription: "List of network device functions for NetDevFuncMinBWAlloc associated with this port",
			Description:         "List of network device functions for NetDevFuncMinBWAlloc associated with this port",
			Computed:            true,
		},
	}
}

// SupportedLinkCapabilitySchema is a function that returns the schema for supported link capability.
func SupportedLinkCapabilitySchema() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"auto_speed_negotiation": schema.BoolAttribute{
			MarkdownDescription: "An indication of whether the port is capable of autonegotiating speed",
			Description:         "An indication of whether the port is capable of autonegotiating speed",
			Computed:            true,
		},
		"link_network_technology": schema.StringAttribute{
			MarkdownDescription: "The link network technology capabilities of this port",
			Description:         "The link network technology capabilities of this port",
			Computed:            true,
		},
		"link_speed_mbps": schema.Int64Attribute{
			MarkdownDescription: "The speed of the link in Mbit/s when this link network technology is active",
			Description:         "The speed of the link in Mbit/s when this link network technology is active",
			Computed:            true,
		},
	}
}

// NetworkPortOemSchema is a function that returns the schema for network port oem.
func NetworkPortOemSchema() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"dell_network_transceiver": schema.SingleNestedAttribute{
			MarkdownDescription: "Dell Network Transceiver",
			Description:         "Dell Network Transceiver",
			Computed:            true,
			Attributes:          DellNetworkTransceiverSchema(),
		},
	}
}

// DellNetworkTransceiverSchema is a function that returns the schema for dell network transceiver.
func DellNetworkTransceiverSchema() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		NICComponmentSchemaOdataID: schema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "The unique identifier for a resource",
			Description:         "The unique identifier for a resource.",
		},
		NICComponmentSchemaID: schema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "The unique identifier for this resource within the collection of similar resources",
			Description:         "The unique identifier for this resource within the collection of similar resources",
		},
		NICComponmentSchemaName: schema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "The name of the resource or array member",
			Description:         "The name of the resource or array member",
		},
		"device_description": schema.StringAttribute{
			Computed: true,
			MarkdownDescription: "A string that contains the friendly Fully Qualified Device Description (FQDD), which is a property that " +
				"describes the device and its location",
			Description: "A string that contains the friendly Fully Qualified Device Description (FQDD), which is a property that " +
				"describes the device and its location",
		},
		"fqdd": schema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "A string that contains the Fully Qualified Device Description (FQDD) for the DellNetworkTransceiver",
			Description:         "A string that contains the Fully Qualified Device Description (FQDD) for the DellNetworkTransceiver",
		},
		"identifier_type": schema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "This property represents the type of Small Form Factor pluggable(SFP) Transceiver for the DellNetworkTransceiver",
			Description:         "This property represents the type of Small Form Factor pluggable(SFP) Transceiver for the DellNetworkTransceiver",
		},
		"interface_type": schema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "This property represents the interface type of Small Form Factor pluggable(SFP) Transceiver",
			Description:         "This property represents the interface type of Small Form Factor pluggable(SFP) Transceiver",
		},
		NICComponmentSchemaPartNumber: schema.StringAttribute{
			Computed: true,
			MarkdownDescription: "The part number assigned by the organization that is responsible for producing or SFP" +
				"(manufacturing the Small Form Factor pluggable) Transceivers",
			Description: "The part number assigned by the organization that is responsible for producing or SFP" +
				"(manufacturing the Small Form Factor pluggable) Transceivers",
		},
		"revision": schema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "This property represents the revision number of the Small Form Factor pluggable(SFP) Transceiver",
			Description:         "This property represents the revision number of the Small Form Factor pluggable(SFP) Transceiver",
		},
		NICComponmentSchemaSerialNumber: schema.StringAttribute{
			Computed:            true,
			MarkdownDescription: NICSchemaDescriptionForSerialNumber,
			Description:         NICSchemaDescriptionForSerialNumber,
		},
		"vendor_name": schema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "This property represents the vendor name of Small Form Factor pluggable(SFP) Transceiver for the object.",
			Description:         "This property represents the vendor name of Small Form Factor pluggable(SFP) Transceiver for the object.",
		},
	}
}

// NetworkDeviceFunctionDataSourceSchema is a function that returns the schema for network device function data source.
func NetworkDeviceFunctionDataSourceSchema() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		NICComponmentSchemaOdataID: schema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "OData ID for the network device function",
			Description:         "OData ID for the network device function",
		},
		NICComponmentSchemaID: schema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "ID of the network device function",
			Description:         "ID of the network device function",
		},
		NICComponmentSchemaName: schema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "name of the network device function",
			Description:         "name of the network device function",
		},
		NICComponmentSchemaDescription: schema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "description of the network device function",
			Description:         "description of the network device function",
		},
		NICComponmentSchemaStatus: schema.SingleNestedAttribute{
			MarkdownDescription: "status of the network device function",
			Description:         "status of the network device function",
			Computed:            true,
			Attributes:          NetworkStatusSchema(),
		},
		"ethernet": schema.SingleNestedAttribute{
			MarkdownDescription: "This type describes Ethernet capabilities, status, and configuration for a network device function",
			Description:         "This type describes Ethernet capabilities, status, and configuration for a network device function",
			Computed:            true,
			Attributes:          EthernetSchema(),
		},
		"fibre_channel": schema.SingleNestedAttribute{
			MarkdownDescription: "This type describes Fibre Channel capabilities, status, and configuration for a network device function",
			Description:         "This type describes Fibre Channel capabilities, status, and configuration for a network device function",
			Computed:            true,
			Attributes:          FibreChannelSchema(),
		},
		"iscsi_boot": schema.SingleNestedAttribute{
			MarkdownDescription: "The iSCSI boot capabilities, status, and configuration for a network device function",
			Description:         "The iSCSI boot capabilities, status, and configuration for a network device function",
			Computed:            true,
			Attributes:          ISCSIBootSchema(),
		},
		"max_virtual_functions": schema.Int64Attribute{
			Computed:            true,
			MarkdownDescription: "The number of virtual functions that are available for this network device function",
			Description:         "The number of virtual functions that are available for this network device function",
		},
		"net_dev_func_capabilities": schema.ListAttribute{
			ElementType:         types.StringType,
			Computed:            true,
			MarkdownDescription: "An array of capabilities for this network device function",
			Description:         "An array of capabilities for this network device function",
		},
		"net_dev_func_type": schema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "The configured capability of this network device function",
			Description:         "The configured capability of this network device function",
		},
		"physical_port_assignment": schema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "A reference to a physical port assignment to this function",
			Description:         "A reference to a physical port assignment to this function",
		},
		"assignable_physical_ports": schema.ListAttribute{
			ElementType:         types.StringType,
			Computed:            true,
			MarkdownDescription: "A reference to assignable physical ports to this function",
			Description:         "A reference to assignable physical ports to this function",
		},
		"assignable_physical_network_ports": schema.ListAttribute{
			ElementType:         types.StringType,
			Computed:            true,
			MarkdownDescription: "A reference to assignable physical network ports to this function",
			Description:         "A reference to assignable physical network ports to this function",
		},
		"oem": schema.SingleNestedAttribute{
			MarkdownDescription: "The OEM extension for this network network function",
			Description:         "The OEM extension for this network network function",
			Computed:            true,
			Attributes:          NetworkDeviceFunctionOemSchema(),
		},
	}
}

// NetworkDeviceFunctionOemSchema is a function that returns the schema for network device function oem.
func NetworkDeviceFunctionOemSchema() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"dell_nic": schema.SingleNestedAttribute{
			MarkdownDescription: "The OEM extension of Dell NIC for this network device function",
			Description:         "The OEM extension of Dell NIC for this network device function",
			Computed:            true,
			Attributes:          DellNICSchema(),
		},
		"dell_nic_port_metrics": schema.SingleNestedAttribute{
			MarkdownDescription: "The OEM extension of Dell NIC port metrics for this network device function",
			Description:         "The OEM extension of Dell NIC port metrics for this network device function",
			Computed:            true,
			Attributes:          DellNICPortMetricsSchema(),
		},
		"dell_nic_capabilities": schema.SingleNestedAttribute{
			MarkdownDescription: "The OEM extension of Dell NIC capabilities for this network device function",
			Description:         "The OEM extension of Dell NIC capabilities for this network device function",
			Computed:            true,
			Attributes:          DellNICCapabilitiesSchema(),
		},
		"dell_fc": schema.SingleNestedAttribute{
			MarkdownDescription: "The OEM extension of Dell FC for this network device function",
			Description:         "The OEM extension of Dell FC for this network device function",
			Computed:            true,
			Attributes:          DellFCSchema(),
		},
		"dell_fc_port_metrics": schema.SingleNestedAttribute{
			MarkdownDescription: "The OEM extension of Dell FC port metrics for this network device function",
			Description:         "The OEM extension of Dell FC port metrics for this network device function",
			Computed:            true,
			Attributes:          DellFCPortMetricsSchema(),
		},
		"dell_fc_port_capabilities": schema.SingleNestedAttribute{
			MarkdownDescription: "The OEM extension of Dell FC capabilities for this network device function",
			Description:         "The OEM extension of Dell FC capabilities for this network device function",
			Computed:            true,
			Attributes:          DellFCCapabilitiesSchema(),
		},
	}
}

// DellFCCapabilitiesSchema is a function that returns the schema for dell fc capabilities.
func DellFCCapabilitiesSchema() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		NICComponmentSchemaOdataID: schema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "OData ID of DellFCCapabilities for the network device function",
			Description:         "OData ID of DellFCCapabilities for the network device function",
		},
		NICComponmentSchemaID: schema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "ID of the DellFCCapabilities ",
			Description:         "ID of the DellFCCapabilities",
		},
		NICComponmentSchemaName: schema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "Name of the DellFCCapabilities",
			Description:         "Name of the DellFCCapabilities",
		},
		"fc_max_number_exchanges": schema.Int64Attribute{
			Computed:            true,
			MarkdownDescription: "This property represents the maximum number of exchanges",
			Description:         "This property represents the maximum number of exchanges",
		},
		"fc_max_number_out_standing_commands": schema.Int64Attribute{
			Computed:            true,
			MarkdownDescription: "This property represents the maximum number of outstanding commands across all connections",
			Description:         "This property represents the maximum number of outstanding commands across all connections",
		},
		"feature_licensing_support": schema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "The property provides details of the FC's feature licensing support",
			Description:         "The property provides details of the FC's feature licensing support",
		},
		"flex_addressing_support": schema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "The property provides detail of the FC's port's flex addressing support",
			Description:         "The property provides detail of the FC's port's flex addressing support",
		},
		"on_chip_thermal_sensor": schema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "The property provides details of the FC's on-chip thermal sensor support",
			Description:         "The property provides details of the FC's on-chip thermal sensor support",
		},
		"persistence_policy_support": schema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "This property specifies if the card supports persistence policy",
			Description:         "This property specifies if the card supports persistence policy",
		},
		"uefi_support": schema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "The property provides details of the FC's port's UEFI support",
			Description:         "The property provides details of the FC's port's UEFI support",
		},
	}
}

// DellNICCapabilitiesSchema is a function that returns the schema for dell nic capabilities.
func DellNICCapabilitiesSchema() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		NICComponmentSchemaOdataID: schema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "OData ID of DellNICCapabilities for the network device function",
			Description:         "OData ID of DellNICCapabilities for the network device function",
		},
		NICComponmentSchemaID: schema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "ID of DellNICCapabilities",
			Description:         "ID of DellNICCapabilities",
		},
		NICComponmentSchemaName: schema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "Name of DellNICCapabilities",
			Description:         "Name of DellNICCapabilities",
		},
		"bpe_support": schema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "This property represents Bridge Port Extension (BPE) for the ports of the NIC",
			Description:         "This property represents Bridge Port Extension (BPE) for the ports of the NIC",
		},
		"congestion_notification": schema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "This property represents congestion notification support for a NIC port",
			Description:         "This property represents congestion notification support for a NIC port",
		},
		"dcb_exchange_protocol": schema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "This property represents DCB Exchange protocol support for a NIC port",
			Description:         "This property represents DCB Exchange protocol support for a NIC port",
		},
		"ets": schema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "This property represents Enhanced Transmission Selection support for a NIC port",
			Description:         "This property represents Enhanced Transmission Selection support for a NIC port",
		},
		"evb_modes_support": schema.StringAttribute{
			Computed: true,
			MarkdownDescription: "This property represents EVB Edge Virtual Bridging) mode support for the ports of the NIC. " +
				"Possible values are 0 Unknown, 2 Supported, 3 Not Supported",
			Description: "This property represents EVB Edge Virtual Bridging) mode support for the ports of the NIC. " +
				"Possible values are 0 Unknown, 2 Supported, 3 Not Supported",
		},
		"fcoe_boot_support": schema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "The property shall represent FCoE boot support for a NIC port",
			Description:         "The property shall represent FCoE boot support for a NIC port",
		},
		"fcoe_max_ios_per_session": schema.Int64Attribute{
			Computed:            true,
			MarkdownDescription: "This property represents the maximum number of I/Os per connection supported by the NIC",
			Description:         "This property represents the maximum number of I/Os per connection supported by the NIC",
		},
		"fcoe_max_npiv_per_port": schema.Int64Attribute{
			Computed:            true,
			MarkdownDescription: "This property represents the maximum number of NPIV per port supported by the DellNICCapabilities",
			Description:         "This property represents the maximum number of NPIV per port supported by the DellNICCapabilities",
		},
		"fcoe_max_number_exchanges": schema.Int64Attribute{
			Computed:            true,
			MarkdownDescription: "This property represents the maximum number of exchanges for the NIC",
			Description:         "This property represents the maximum number of exchanges for the NIC",
		},
		"fcoe_max_number_logins": schema.Int64Attribute{
			Computed:            true,
			MarkdownDescription: "This property represents the maximum logins per port for the NIC",
			Description:         "This property represents the maximum logins per port for the NIC",
		},
		"fcoe_max_number_of_fc_targets": schema.Int64Attribute{
			Computed:            true,
			MarkdownDescription: "This property represents the maximum number of FCoE targets supported by the NIC",
			Description:         "This property represents the maximum number of FCoE targets supported by the NIC",
		},
		"fcoe_max_number_outstanding_commands": schema.Int64Attribute{
			Computed:            true,
			MarkdownDescription: "This property represents the maximum number of outstanding commands supported across all connections for the NIC",
			Description:         "This property represents the maximum number of outstanding commands supported across all connections for the NIC",
		},
		"fcoe_offload_support": schema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "The property shall represent FCoE offload support for the NIC",
			Description:         "The property shall represent FCoE offload support for the NIC",
		},
		"feature_licensing_support": schema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "This property represents feature licensing support for the NIC",
			Description:         "This property represents feature licensing support for the NIC",
		},
		"flex_addressing_support": schema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "The property shall represent flex adddressing support for a NIC port",
			Description:         "The property shall represent flex adddressing support for a NIC port",
		},
		"ipsec_offload_support": schema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "This property represents IPSec offload support for a NIC port",
			Description:         "This property represents IPSec offload support for a NIC port",
		},
		"mac_sec_support": schema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "This property represents secure MAC support for a NIC port",
			Description:         "This property represents secure MAC support for a NIC port",
		},
		"nw_management_pass_through": schema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "This property represents network management passthrough support for a NIC port",
			Description:         "This property represents network management passthrough support for a NIC port",
		},
		"nic_partitioning_support": schema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "This property represents partitioning support for the NIC",
			Description:         "This property represents partitioning support for the NIC",
		},
		"os_bmc_management_pass_through": schema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "This property represents OS-inband to BMC-out-of-band management passthrough support for a NIC port",
			Description:         "This property represents OS-inband to BMC-out-of-band management passthrough support for a NIC port",
		},
		"on_chip_thermal_sensor": schema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "This property represents on-chip thermal sensor support for the NIC",
			Description:         "This property represents on-chip thermal sensor support for the NIC",
		},
		"open_flow_support": schema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "This property represents open-flow support for a NIC port",
			Description:         "This property represents open-flow support for a NIC port",
		},
		"pxe_boot_support": schema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "The property shall represent PXE boot support for a NIC port",
			Description:         "The property shall represent PXE boot support for a NIC port",
		},
		"partition_wol_support": schema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "This property represents Wake-On-LAN support for a NIC partition",
			Description:         "This property represents Wake-On-LAN support for a NIC partition",
		},
		"persistence_policy_support": schema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "This property specifies whether the card supports persistence policy",
			Description:         "This property specifies whether the card supports persistence policy",
		},
		"priority_flow_control": schema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "This property represents priority flow-control support for a NIC port",
			Description:         "This property represents priority flow-control support for a NIC port",
		},
		"rdma_support": schema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "This property represents RDMA support for a NIC port",
			Description:         "This property represents RDMA support for a NIC port",
		},
		"remote_phy": schema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "This property represents remote PHY support for a NIC port",
			Description:         "This property represents remote PHY support for a NIC port",
		},
		"tcp_chimney_support": schema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "This property represents TCP Chimney support for a NIC port",
			Description:         "This property represents TCP Chimney support for a NIC port",
		},
		"tcp_offload_engine_support": schema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "This property represents the support of TCP Offload Engine for a NIC port",
			Description:         "This property represents the support of TCP Offload Engine for a NIC port",
		},
		"veb": schema.StringAttribute{
			Computed: true,
			MarkdownDescription: "This property provides details about the VEB (Virtual Ethernet Bridging) -" +
				" single channel support for the ports of the NIC",
			Description: "This property provides details about the VEB (Virtual Ethernet Bridging) -" +
				" single channel support for the ports of the NIC",
		},
		"veb_vepa_multi_channel": schema.StringAttribute{
			Computed: true,
			MarkdownDescription: "This property provides details about the Virtual Ethernet Bridging and Virtual Ethernet Port Aggregator" +
				" (VEB-VEPA) multichannel support for the ports of the NIC",
			Description: "This property provides details about the Virtual Ethernet Bridging and Virtual Ethernet Port Aggregator" +
				" (VEB-VEPA) multichannel support for the ports of the NIC",
		},
		"veb_vepa_single_channel": schema.StringAttribute{
			Computed: true,
			MarkdownDescription: "This property provides details about the VEB-VEPA" +
				" (Virtual Ethernet Bridging and Virtual Ethernet Port Aggregator) - single channel support for the ports of the NIC",
			Description: "This property provides details about the VEB-VEPA" +
				" (Virtual Ethernet Bridging and Virtual Ethernet Port Aggregator) - single channel support for the ports of the NIC",
		},
		"virtual_link_control": schema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "This property represents virtual link-control support for a NIC partition",
			Description:         "This property represents virtual link-control support for a NIC partition",
		},
		"iscsi_boot_support": schema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "The property shall represent iSCSI boot support for a NIC port",
			Description:         "The property shall represent iSCSI boot support for a NIC port",
		},
		"iscsi_offload_support": schema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "The property shall represent iSCSI offload support for a NIC port",
			Description:         "The property shall represent iSCSI offload support for a NIC port",
		},
		"uefi_support": schema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "This property represents UEFI support for a NIC port",
			Description:         "This property represents UEFI support for a NIC port",
		},
	}
}

// DellFCPortMetricsSchema is a function that returns the schema for dell fc port metrics.
func DellFCPortMetricsSchema() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		NICComponmentSchemaOdataID: schema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "OData ID of DellFCPortMetrics for the network device function",
			Description:         "OData ID of DellFCPortMetrics for the network device function",
		},
		NICComponmentSchemaID: schema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "ID of the DellFCPortMetrics",
			Description:         "ID of the DellFCPortMetrics",
		},
		NICComponmentSchemaName: schema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "Name of the DellFCPortMetrics",
			Description:         "Name of the DellFCPortMetrics",
		},
		"fc_invalid_crcs": schema.Int64Attribute{
			Computed:            true,
			MarkdownDescription: "This property represents invalid CRCs",
			Description:         "This property represents invalid CRCs",
		},
		"fc_link_failures": schema.Int64Attribute{
			Computed:            true,
			MarkdownDescription: "This property represents link failures",
			Description:         "This property represents link failures",
		},
		"fc_loss_of_signals": schema.Int64Attribute{
			Computed:            true,
			MarkdownDescription: "This property represents loss of signals",
			Description:         "This property represents loss of signals",
		},
		"fc_rx_kb_count": schema.Int64Attribute{
			Computed:            true,
			MarkdownDescription: "This property represents the KB count received",
			Description:         "This property represents the KB count received",
		},
		"fc_rx_sequences": schema.Int64Attribute{
			Computed:            true,
			MarkdownDescription: "This property represents the FC sequences received",
			Description:         "This property represents the FC sequences received",
		},
		"fc_rx_total_frames": schema.Int64Attribute{
			Computed:            true,
			MarkdownDescription: "This property represents the total FC frames received",
			Description:         "This property represents the total FC frames received.",
		},
		"fc_tx_kb_count": schema.Int64Attribute{
			Computed:            true,
			MarkdownDescription: "This property represents the KB count transmitted",
			Description:         "This property represents the KB count transmitted",
		},
		"fc_tx_sequences": schema.Int64Attribute{
			Computed:            true,
			MarkdownDescription: "This property represents the FC sequences transmitted",
			Description:         "This property represents the FC sequences transmitted",
		},
		"fc_tx_total_frames": schema.Int64Attribute{
			Computed:            true,
			MarkdownDescription: "This property represents the total FC frames transmitted",
			Description:         "This property represents the total FC frames transmitted",
		},
		"os_driver_state": schema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "This property indicates the OS driver states for the DellFCPortMetrics",
			Description:         "This property indicates the OS driver states for the DellFCPortMetrics",
		},
		"port_status": schema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "This property represents port status for the DellFCPortMetrics",
			Description:         "This property represents port status for the DellFCPortMetrics",
		},
		"rx_input_power_status": schema.StringAttribute{
			Computed: true,
			MarkdownDescription: "Indicates the status of Rx Input Power value limits for the DellFCPortMetrics. " +
				NICSchemaDescriptionForDeprecatedNoteV440,
			Description: "Indicates the status of Rx Input Power value limits for the DellFCPortMetrics. " +
				NICSchemaDescriptionForDeprecatedNoteV440,
		},
		"rx_input_power_mw": schema.NumberAttribute{
			Computed: true,
			MarkdownDescription: "Indicates the RX input power value of Small Form Factor pluggable(SFP) Transceiver for the DellFCPortMetrics. " +
				NICSchemaDescriptionForDeprecatedNoteV440,
			Description: "Indicates the RX input power value of Small Form Factor pluggable(SFP) Transceiver for the DellFCPortMetrics. " +
				NICSchemaDescriptionForDeprecatedNoteV440,
		},
		"tx_bias_current_status": schema.StringAttribute{
			Computed: true,
			MarkdownDescription: "Indicates the status of Tx Bias Current value limits for the DellFCPortMetrics. " +
				NICSchemaDescriptionForDeprecatedNoteV440,
			Description: "Indicates the status of Tx Bias Current value limits for the DellFCPortMetrics. " +
				NICSchemaDescriptionForDeprecatedNoteV440,
		},
		"tx_bias_current_mw": schema.NumberAttribute{
			Computed: true,
			MarkdownDescription: "Indicates the TX Bias current value of Small Form Factor pluggable(SFP) Transceiver for the DellFCPortMetrics. " +
				NICSchemaDescriptionForDeprecatedNoteV440,
			Description: "Indicates the TX Bias current value of Small Form Factor pluggable(SFP) Transceiver for the DellFCPortMetrics. " +
				NICSchemaDescriptionForDeprecatedNoteV440,
		},
		"tx_output_power_status": schema.StringAttribute{
			Computed: true,
			MarkdownDescription: "Indicates the status of Tx Output Power value limits for the DellFCPortMetrics. " +
				NICSchemaDescriptionForDeprecatedNoteV440,
			Description: "Indicates the status of Tx Output Power value limits for the DellFCPortMetrics. " +
				NICSchemaDescriptionForDeprecatedNoteV440,
		},
		"tx_output_power_mw": schema.NumberAttribute{
			Computed: true,
			MarkdownDescription: "Indicates the TX output power value of Small Form Factor pluggable(SFP) Transceiver for the DellFCPortMetrics. " +
				NICSchemaDescriptionForDeprecatedNoteV440,
			Description: "Indicates the TX output power value of Small Form Factor pluggable(SFP) Transceiver for the DellFCPortMetrics. " +
				NICSchemaDescriptionForDeprecatedNoteV440,
		},
		"temperature_status": schema.StringAttribute{
			Computed: true,
			MarkdownDescription: "Indicates the status of Temperature value limits for the DellFCPortMetrics. " +
				NICSchemaDescriptionForDeprecatedNoteV440,
			Description: "Indicates the status of Temperature value limits for the DellFCPortMetrics. " +
				NICSchemaDescriptionForDeprecatedNoteV440,
		},
		"temperature_celsius": schema.NumberAttribute{
			Computed: true,
			MarkdownDescription: "Indicates the temperature value of Small Form Factor pluggable(SFP) Transceiver for the DellFCPortMetrics. " +
				NICSchemaDescriptionForDeprecatedNoteV440,
			Description: "Indicates the temperature value of Small Form Factor pluggable(SFP) Transceiver for the DellFCPortMetrics. " +
				NICSchemaDescriptionForDeprecatedNoteV440,
		},
		"voltage_status": schema.StringAttribute{
			Computed: true,
			MarkdownDescription: "Indicates the status of voltage value limits for the DellFCPortMetrics. " +
				NICSchemaDescriptionForDeprecatedNoteV440,
			Description: "Indicates the status of voltage value limits for the DellFCPortMetrics. " +
				NICSchemaDescriptionForDeprecatedNoteV440,
		},
		"voltage_value_volts": schema.NumberAttribute{
			Computed: true,
			MarkdownDescription: "Indicates the voltage value of Small Form Factor pluggable(SFP) Transceiver for the DellFCPortMetrics. " +
				NICSchemaDescriptionForDeprecatedNoteV440,
			Description: "Indicates the voltage value of Small Form Factor pluggable(SFP) Transceiver for the DellFCPortMetrics. " +
				NICSchemaDescriptionForDeprecatedNoteV440,
		},
	}
}

// DellNICPortMetricsSchema is a function that returns the schema for dell nic port metrics.
func DellNICPortMetricsSchema() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		NICComponmentSchemaOdataID: schema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "OData ID of DellNICPortMetrics for the network device function",
			Description:         "OData ID of DellNICPortMetrics for the network device function",
		},
		NICComponmentSchemaID: schema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "ID of DellNICPortMetrics",
			Description:         "ID of DellNICPortMetrics",
		},
		NICComponmentSchemaName: schema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "Name of DellNICPortMetrics",
			Description:         "Name of DellNICPortMetrics",
		},
		"discarded_pkts": schema.Int64Attribute{
			Computed:            true,
			MarkdownDescription: "Indicates the total number of discarded packets",
			Description:         "Indicates the total number of discarded packets",
		},
		"fqdd": schema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "A string that contains the Fully Qualified Device Description (FQDD) for the DellNICPortMetrics",
			Description:         "A string that contains the Fully Qualified Device Description (FQDD) for the DellNICPortMetrics",
		},
		"os_driver_state": schema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "Indicates operating system driver states",
			Description:         "Indicates operating system driver states",
		},
		"rx_bytes": schema.Int64Attribute{
			Computed: true,
			MarkdownDescription: "Indicates the total number of bytes received, including host and remote management pass through traffic. " +
				"Remote management passthrough received traffic is applicable to LOMs only",
			Description: "Indicates the total number of bytes received, including host and remote management pass through traffic. " +
				"Remote management passthrough received traffic is applicable to LOMs only",
		},
		"rx_broadcast": schema.Int64Attribute{
			Computed:            true,
			MarkdownDescription: "Indicates the total number of good broadcast packets received",
			Description:         "Indicates the total number of good broadcast packets received",
		},
		"rx_error_pkt_alignment_errors": schema.Int64Attribute{
			Computed:            true,
			MarkdownDescription: "Indicates the total number of packets received with alignment errors",
			Description:         "Indicates the total number of packets received with alignment errors",
		},
		"rx_error_pkt_fcs_errors": schema.Int64Attribute{
			Computed:            true,
			MarkdownDescription: "Indicates the total number of packets received with FCS errors",
			Description:         "Indicates the total number of packets received with FCS errors",
		},
		"rx_jabber_pkt": schema.Int64Attribute{
			Computed:            true,
			MarkdownDescription: "Indicates the total number of frames that are too long",
			Description:         "Indicates the total number of frames that are too long",
		},
		"rx_mutlicast_packets": schema.Int64Attribute{
			Computed:            true,
			MarkdownDescription: "Indicates the total number of good multicast packets received",
			Description:         "Indicates the total number of good multicast packets received",
		},
		"rx_pause_xoff_frames": schema.Int64Attribute{
			Computed:            true,
			MarkdownDescription: "Indicates the flow control frames from the network to pause transmission",
			Description:         "Indicates the flow control frames from the network to pause transmission",
		},
		"rx_pause_xon_frames": schema.Int64Attribute{
			Computed:            true,
			MarkdownDescription: "Indicates the flow control frames from the network to resume transmission",
			Description:         "Indicates the flow control frames from the network to resume transmission",
		},
		"rx_runt_pkt": schema.Int64Attribute{
			Computed:            true,
			MarkdownDescription: "Indicates the total number of frames that are too short (< 64 bytes)",
			Description:         "Indicates the total number of frames that are too short (< 64 bytes)",
		},
		"rx_unicast_packets": schema.Int64Attribute{
			Computed:            true,
			MarkdownDescription: "Indicates the total number of good unicast packets received",
			Description:         "Indicates the total number of good unicast packets received",
		},
		"tx_bytes": schema.Int64Attribute{
			Computed: true,
			MarkdownDescription: "Indicates the total number of bytes transmitted, including host and remote management passthrough traffic. " +
				"Remote management passthrough transmitted traffic is applicable to LOMs only",
			Description: "Indicates the total number of bytes transmitted, including host and remote management passthrough traffic. " +
				"Remote management passthrough transmitted traffic is applicable to LOMs only",
		},
		"tx_broadcast": schema.Int64Attribute{
			Computed:            true,
			MarkdownDescription: "Indicates the total number of good broadcast packets transmitted",
			Description:         "Indicates the total number of good broadcast packets transmitted",
		},
		"tx_error_pkt_excessive_collision": schema.Int64Attribute{
			Computed:            true,
			MarkdownDescription: "Indicates the number of times a single transmitted packet encountered more than 15 collisions",
			Description:         "Indicates the number of times a single transmitted packet encountered more than 15 collisions",
		},
		"tx_error_pkt_late_collision": schema.Int64Attribute{
			Computed:            true,
			MarkdownDescription: "Indicates the number of collisions that occurred after one slot time (defined by IEEE 802.3)",
			Description:         "Indicates the number of collisions that occurred after one slot time (defined by IEEE 802.3)",
		},
		"tx_error_pkt_multiple_collision": schema.Int64Attribute{
			Computed:            true,
			MarkdownDescription: "Indicates the number of times that a transmitted packet encountered 2-15 collisions",
			Description:         "Indicates the number of times that a transmitted packet encountered 2-15 collisions",
		},
		"tx_error_pkt_single_collision": schema.Int64Attribute{
			Computed:            true,
			MarkdownDescription: "Indicates the number of times that a successfully transmitted packet encountered a single collision",
			Description:         "Indicates the number of times that a successfully transmitted packet encountered a single collision",
		},
		"tx_mutlicast_packets": schema.Int64Attribute{
			Computed:            true,
			MarkdownDescription: "Indicates the total number of good multicast packets transmitted",
			Description:         "Indicates the total number of good multicast packets transmitted",
		},
		"tx_pause_xoff_frames": schema.Int64Attribute{
			Computed:            true,
			MarkdownDescription: "Indicates the number of XOFF packets transmitted to the network",
			Description:         "Indicates the number of XOFF packets transmitted to the network",
		},
		"tx_pause_xon_frames": schema.Int64Attribute{
			Computed:            true,
			MarkdownDescription: "Indicates the number of XON packets transmitted to the network",
			Description:         "Indicates the number of XON packets transmitted to the network",
		},
		"start_statistic_time": schema.StringAttribute{
			Computed: true,
			MarkdownDescription: "Indicates the measurement time for the first NIC statistics. " +
				"The property is used with the StatisticTime property to calculate the duration over which the NIC statistics are gathered",
			Description: "Indicates the measurement time for the first NIC statistics. " +
				"The property is used with the StatisticTime property to calculate the duration over which the NIC statistics are gathered",
		},
		"statistic_time": schema.StringAttribute{
			Computed: true,
			MarkdownDescription: "Indicates the most recent measurement time for NIC statistics. " +
				"The property is used with the StatisticStartTime property to calculate the duration over which the NIC statistics are gathered",
			Description: "Indicates the most recent measurement time for NIC statistics. " +
				"The property is used with the StatisticStartTime property to calculate the duration over which the NIC statistics are gathered",
		},
		"fc_crc_error_count": schema.Int64Attribute{
			Computed:            true,
			MarkdownDescription: "Indicates the number of FC frames with CRC errors",
			Description:         "Indicates the number of FC frames with CRC errors",
		},
		"fcoe_link_failures": schema.Int64Attribute{
			Computed:            true,
			MarkdownDescription: "Indicates the number of FCoE/FIP login failures",
			Description:         "Indicates the number of FCoE/FIP login failures",
		},
		"fcoe_pkt_rx_count": schema.Int64Attribute{
			Computed:            true,
			MarkdownDescription: "Indicates the number of good (FCS valid) packets received with the active FCoE MAC address of the partition",
			Description:         "Indicates the number of good (FCS valid) packets received with the active FCoE MAC address of the partition",
		},
		"fcoe_pkt_tx_count": schema.Int64Attribute{
			Computed:            true,
			MarkdownDescription: "Indicates the number of good (FCS valid) packets transmitted that passed L2 filtering by a specific MAC address",
			Description:         "Indicates the number of good (FCS valid) packets transmitted that passed L2 filtering by a specific MAC address",
		},
		"fcoe_rx_pkt_dropped_count": schema.Int64Attribute{
			Computed:            true,
			MarkdownDescription: "Indicates the number of receive packets with FCS errors",
			Description:         "Indicates the number of receive packets with FCS errors",
		},
		"lan_fcs_rx_errors": schema.Int64Attribute{
			Computed:            true,
			MarkdownDescription: "Indicates the Lan FCS receive Errors",
			Description:         "Indicates the Lan FCS receive Errors",
		},
		"lan_unicast_pkt_rx_count": schema.Int64Attribute{
			Computed:            true,
			MarkdownDescription: "Indicates the total number of Lan Unicast Packets Received",
			Description:         "Indicates the total number of Lan Unicast Packets Received",
		},
		"lan_unicast_pkt_tx_count": schema.Int64Attribute{
			Computed:            true,
			MarkdownDescription: "Indicates the total number of Lan Unicast Packets Transmitted",
			Description:         "Indicates the total number of Lan Unicast Packets Transmitted",
		},
		"rdma_rx_total_bytes": schema.Int64Attribute{
			Computed:            true,
			MarkdownDescription: "Indicates the total number of RDMA bytes received",
			Description:         "Indicates the total number of RDMA bytes received",
		},
		"rdma_rx_total_packets": schema.Int64Attribute{
			Computed:            true,
			MarkdownDescription: "Indicates the total number of RDMA packets received",
			Description:         "Indicates the total number of RDMA packets received",
		},
		"rdma_total_protection_errors": schema.Int64Attribute{
			Computed:            true,
			MarkdownDescription: "Indicates the total number of RDMA Protection errors",
			Description:         "Indicates the total number of RDMA Protection errors",
		},
		"rdma_total_protocol_errors": schema.Int64Attribute{
			Computed:            true,
			MarkdownDescription: "Indicates the total number of RDMA Protocol errors",
			Description:         "Indicates the total number of RDMA Protocol errors",
		},
		"rdma_tx_total_bytes": schema.Int64Attribute{
			Computed:            true,
			MarkdownDescription: "Indicates the total number of RDMA bytes transmitted",
			Description:         "Indicates the total number of RDMA bytes transmitted",
		},
		"rdma_tx_total_packets": schema.Int64Attribute{
			Computed:            true,
			MarkdownDescription: "Indicates the total number of RDMA packets transmitted",
			Description:         "Indicates the total number of RDMA packets transmitted",
		},
		"rdma_tx_total_read_req_pkts": schema.Int64Attribute{
			Computed:            true,
			MarkdownDescription: "Indicates the total number of RDMA ReadRequest packets transmitted",
			Description:         "Indicates the total number of RDMA ReadRequest packets transmitted",
		},
		"rdma_tx_total_send_pkts": schema.Int64Attribute{
			Computed:            true,
			MarkdownDescription: "Indicates the total number of RDMA Send packets transmitted",
			Description:         "Indicates the total number of RDMA Send packets transmitted",
		},
		"rdma_tx_total_write_pkts": schema.Int64Attribute{
			Computed:            true,
			MarkdownDescription: "Indicates the total number of RDMA Write packets transmitted",
			Description:         "Indicates the total number of RDMA Write packets transmitted",
		},
		"tx_unicast_packets": schema.Int64Attribute{
			Computed:            true,
			MarkdownDescription: "Indicates the total number of good unicast packets transmitted for the DellFCPortMetrics",
			Description:         "Indicates the total number of good unicast packets transmitted for the DellFCPortMetrics",
		},
		"partition_link_status": schema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "Indicates whether the partition link is up or down for the DellFCPortMetrics",
			Description:         "Indicates whether the partition link is up or down for the DellFCPortMetrics",
		},
		"partition_os_driver_state": schema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "Indicates operating system driver states of the partitions for the DellFCPortMetrics",
			Description:         "Indicates operating system driver states of the partitions for the DellFCPortMetrics",
		},
		"rx_input_power_status": schema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "Indicates the status of Rx Input Power value limits for the DellFCPortMetrics",
			Description:         "Indicates the status of Rx Input Power value limits for the DellFCPortMetrics",
		},
		"rx_false_carrier_detection": schema.Int64Attribute{
			Computed:            true,
			MarkdownDescription: "Indicates the total number of false carrier errors received from PHY",
			Description:         "Indicates the total number of false carrier errors received from PHY",
		},
		"tx_bias_current_status": schema.StringAttribute{
			Computed: true,
			MarkdownDescription: "Indicates the status of Tx Bias Current value limits for the DellNICPortMetrics. " +
				NICSchemaDescriptionForDeprecatedNoteV440,
			Description: "Indicates the status of Tx Bias Current value limits for the DellNICPortMetrics. " +
				NICSchemaDescriptionForDeprecatedNoteV440,
		},
		"tx_output_power_status": schema.StringAttribute{
			Computed: true,
			MarkdownDescription: "Indicates the status of Tx Output Power value limits for the DellNICPortMetrics.. " +
				NICSchemaDescriptionForDeprecatedNoteV440,
			Description: "Indicates the status of Tx Output Power value limits for the DellNICPortMetrics.. " +
				NICSchemaDescriptionForDeprecatedNoteV440,
		},
		"temperature_status": schema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "Indicates the status of Temperature value limits for the DellNICPortMetrics.",
			Description:         "Indicates the status of Temperature value limits for the DellNICPortMetrics.",
		},
		"voltage_status": schema.StringAttribute{
			Computed: true,
			MarkdownDescription: "Indicates the status of voltage value limits for the DellNICPortMetrics. " +
				NICSchemaDescriptionForDeprecatedNoteV440,
			Description: "Indicates the status of voltage value limits for the DellNICPortMetrics. " +
				NICSchemaDescriptionForDeprecatedNoteV440,
		},
		"rx_input_power_mw": schema.NumberAttribute{
			Computed: true,
			MarkdownDescription: "Indicates the RX input power value of Small Form Factor pluggable(SFP) Transceiver for the DellNICPortMetrics. " +
				NICSchemaDescriptionForDeprecatedNoteV440,
			Description: "Indicates the RX input power value of Small Form Factor pluggable(SFP) Transceiver for the DellNICPortMetrics. " +
				NICSchemaDescriptionForDeprecatedNoteV440,
		},
		"tx_bias_current_ma": schema.NumberAttribute{
			Computed: true,
			MarkdownDescription: "Indicates the TX Bias current value of Small Form Factor pluggable(SFP) Transceiver for the DellNICPortMetrics. " +
				NICSchemaDescriptionForDeprecatedNoteV440,
			Description: "Indicates the TX Bias current value of Small Form Factor pluggable(SFP) Transceiver for the DellNICPortMetrics. " +
				NICSchemaDescriptionForDeprecatedNoteV440,
		},
		"tx_output_power_mw": schema.NumberAttribute{
			Computed: true,
			MarkdownDescription: "Indicates the TX output power value of Small Form Factor pluggable(SFP) Transceiver for the DellNICPortMetrics. " +
				NICSchemaDescriptionForDeprecatedNoteV440,
			Description: "Indicates the TX output power value of Small Form Factor pluggable(SFP) Transceiver for the DellNICPortMetrics. " +
				NICSchemaDescriptionForDeprecatedNoteV440,
		},
		"temperature_celsius": schema.NumberAttribute{
			Computed: true,
			MarkdownDescription: "Indicates the temperature value of Small Form Factor pluggable(SFP) Transceiver for the DellNICPortMetrics. " +
				NICSchemaDescriptionForDeprecatedNoteV440,
			Description: "Indicates the temperature value of Small Form Factor pluggable(SFP) Transceiver for the DellNICPortMetrics. " +
				NICSchemaDescriptionForDeprecatedNoteV440,
		},
		"voltage_value_volts": schema.NumberAttribute{
			Computed: true,
			MarkdownDescription: "Indicates the voltage value of Small Form Factor pluggable(SFP) Transceiver for the DellNICPortMetrics. " +
				NICSchemaDescriptionForDeprecatedNoteV440,
			Description: "Indicates the voltage value of Small Form Factor pluggable(SFP) Transceiver for the DellNICPortMetrics. " +
				NICSchemaDescriptionForDeprecatedNoteV440,
		},
	}
}

// DellFCSchema is a function that returns the schema for dell fc.
func DellFCSchema() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		NICComponmentSchemaOdataID: schema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "OData ID of DellFC for the network device function",
			Description:         "OData ID of DellFC for the network device function",
		},
		NICComponmentSchemaID: schema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "ID of DellFC",
			Description:         "ID of DellFC",
		},
		NICComponmentSchemaName: schema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "Name of DellFC",
			Description:         "Name of DellFC",
		},
		"device_description": schema.StringAttribute{
			Computed: true,
			MarkdownDescription: "A string that contains the friendly Fully Qualified Device Description -" +
				" a property that describes the device and its location",
			Description: "A string that contains the friendly Fully Qualified Device Description -" +
				" a property that describes the device and its location",
		},
		"bus": schema.Int64Attribute{
			Computed:            true,
			MarkdownDescription: "This property represents the bus number of the PCI device",
			Description:         "This property represents the bus number of the PCI device",
		},
		"cable_length_metres": schema.Int64Attribute{
			Computed: true,
			MarkdownDescription: "This property represents the cable length of Small Form Factor pluggable(SFP) Transceiver for the DellFC. " +
				NICSchemaDescriptionForDeprecatedNoteV420,
			Description: "This property represents the cable length of Small Form Factor pluggable(SFP) Transceiver for the DellFC. " +
				NICSchemaDescriptionForDeprecatedNoteV420,
		},
		"chip_type": schema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "This property represents the chip type",
			Description:         "This property represents the chip type",
		},
		"device": schema.Int64Attribute{
			Computed:            true,
			MarkdownDescription: "This property represents the device number of the PCI device",
			Description:         "This property represents the device number of the PCI device",
		},
		"device_name": schema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "This property represents FC HBA device name",
			Description:         "This property represents FC HBA device name",
		},
		"function": schema.Int64Attribute{
			Computed:            true,
			MarkdownDescription: "This property represents the function number of the PCI device",
			Description:         "This property represents the function number of the PCI device",
		},
		"identifier_type": schema.StringAttribute{
			Computed: true,
			MarkdownDescription: "This property represents the type of Small Form Factor pluggable(SFP) Transceiver for the DellFC. " +
				NICSchemaDescriptionForDeprecatedNoteV440,
			Description: "This property represents the type of Small Form Factor pluggable(SFP) Transceiver for the DellFC. " +
				NICSchemaDescriptionForDeprecatedNoteV440,
		},
		"efi_version": schema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "This property represents the EFI version on the device",
			Description:         "This property represents the EFI version on the device",
		},
		"fc_tape_enable": schema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "This property represents the FC Tape state",
			Description:         "This property represents the FC Tape state",
		},
		"fc_os_driver_version": schema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "This property represents the FCOS OS Driver version for the DellFC",
			Description:         "This property represents the FCOS OS Driver version for the DellFC",
		},
		"fcoe_os_driver_version": schema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "This property represents the FCOE OS Driver version",
			Description:         "This property represents the FCOE OS Driver version",
		},
		"fabric_login_retry_count": schema.Int64Attribute{
			Computed:            true,
			MarkdownDescription: "This property represents the Fabric Login Retry Count",
			Description:         "This property represents the Fabric Login Retry Count",
		},
		"fabric_login_timeout": schema.Int64Attribute{
			Computed:            true,
			MarkdownDescription: "This property represents the Fabric Login Timeout in milliseconds",
			Description:         "This property represents the Fabric Login Timeout in milliseconds",
		},
		"family_version": schema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "This property represents the firmware version",
			Description:         "This property represents the firmware version",
		},
		"frame_payload_size": schema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "This property represents the frame payload size",
			Description:         "This property represents the frame payload size",
		},
		"hard_zone_address": schema.Int64Attribute{
			Computed:            true,
			MarkdownDescription: "This property represents the Hard Zone Address",
			Description:         "This property represents the Hard Zone Address",
		},
		"hard_zone_enable": schema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "This property represents the Hard Zone state",
			Description:         "This property represents the Hard Zone state",
		},
		"iscsi_os_driver_version": schema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "This property represents the ISCSI OS Driver version",
			Description:         "This property represents the ISCSI OS Driver version",
		},
		"lan_driver_version": schema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "This property represents the LAN Driver version",
			Description:         "This property represents the LAN Driver version",
		},
		"link_down_timeout": schema.Int64Attribute{
			Computed:            true,
			MarkdownDescription: "This property represents the Link Down Timeout in milliseconds",
			Description:         "This property represents the Link Down Timeout in milliseconds",
		},
		"loop_reset_delay": schema.Int64Attribute{
			Computed:            true,
			MarkdownDescription: "This property represents the Loop Reset Delay in seconds",
			Description:         "This property represents the Loop Reset Delay in seconds",
		},
		NICComponmentSchemaPartNumber: schema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "The part number assigned by the organization that is responsible for producing or manufacturing the FC device",
			Description:         "The part number assigned by the organization that is responsible for producing or manufacturing the FC device",
		},
		"port_down_retry_count": schema.Int64Attribute{
			Computed:            true,
			MarkdownDescription: "This property represents the Port Down Retry Count",
			Description:         "This property represents the Port Down Retry Count",
		},
		"port_down_timeout": schema.Int64Attribute{
			Computed:            true,
			MarkdownDescription: "This property represents the Port Down Timeout in milliseconds",
			Description:         "This property represents the Port Down Timeout in milliseconds",
		},
		"port_login_retry_count": schema.Int64Attribute{
			Computed:            true,
			MarkdownDescription: "This property represents the Port Login Retry Count",
			Description:         "This property represents the Port Login Retry Count",
		},
		"port_login_timeout": schema.Int64Attribute{
			Computed:            true,
			MarkdownDescription: "This property represents the Port Login Timeout in milliseconds",
			Description:         "This property represents the Port Login Timeout in milliseconds",
		},
		"product_name": schema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "This property represents the Product Name",
			Description:         "This property represents the Product Name",
		},
		"rdma_os_driver_version": schema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "This property represents the RDMA OS Driver version",
			Description:         "This property represents the RDMA OS Driver version",
		},
		"revision": schema.StringAttribute{
			Computed: true,
			MarkdownDescription: "This property represents the revision number of the Small Form Factor pluggable(SFP) Transceiver." +
				NICSchemaDescriptionForDeprecatedNoteV440,
			Description: "This property represents the revision number of the Small Form Factor pluggable(SFP) Transceiver." +
				NICSchemaDescriptionForDeprecatedNoteV440,
		},
		"second_fc_target_lun": schema.Int64Attribute{
			Computed:            true,
			MarkdownDescription: "This property represents the Second FC Target LUN",
			Description:         "This property represents the Second FC Target LUN",
		},
		"second_fc_target_wwpn": schema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "This property represents the Second FC Target World-Wide Port Name",
			Description:         "This property represents the Second FC Target World-Wide Port Name",
		},
		NICComponmentSchemaSerialNumber: schema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "A manufacturer-allocated number used to identify the FC device",
			Description:         "A manufacturer-allocated number used to identify the FC device",
		},
		"transceiver_part_number": schema.StringAttribute{
			Computed: true,
			MarkdownDescription: "The part number assigned by the organization that is responsible for producing or " +
				"manufacturing the Small Form Factor pluggable(SFP) Transceivers. " +
				NICSchemaDescriptionForDeprecatedNoteV440,
			Description: "The part number assigned by the organization that is responsible for producing or " +
				"manufacturing the Small Form Factor pluggable(SFP) Transceivers. " +
				NICSchemaDescriptionForDeprecatedNoteV440,
		},
		"transceiver_serial_number": schema.StringAttribute{
			Computed:            true,
			MarkdownDescription: NICSchemaDescriptionForSerialNumber + NICSchemaDescriptionForDeprecatedNoteV440,
			Description:         NICSchemaDescriptionForSerialNumber + NICSchemaDescriptionForDeprecatedNoteV440,
		},
		"transceiver_vendor_name": schema.StringAttribute{
			Computed: true,
			MarkdownDescription: "This property represents the vendor name of Small Form Factor pluggable(SFP) Transceiver for the DellFC. " +
				NICSchemaDescriptionForDeprecatedNoteV440,
			Description: "This property represents the vendor name of Small Form Factor pluggable(SFP) Transceiver for the DellFC. " +
				NICSchemaDescriptionForDeprecatedNoteV440,
		},
		"vendor_name": schema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "This property represents the Vendor Name",
			Description:         "This property represents the Vendor Name",
		},
	}
}

// DellNICSchema is a function that returns the schema for dell nic.
func DellNICSchema() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		NICComponmentSchemaOdataID: schema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "OData ID of DellNIC for the network device function",
			Description:         "OData ID of DellNIC for the network device function",
		},
		NICComponmentSchemaID: schema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "ID of DellNIC",
			Description:         "ID of DellNIC",
		},
		NICComponmentSchemaName: schema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "name of DellNIC",
			Description:         "name of DellNIC",
		},
		"device_description": schema.StringAttribute{
			Computed: true,
			MarkdownDescription: "A string that contains the friendly Fully Qualified Device Description (FQDD), " +
				"which is a property that describes the device and its location",
			Description: "A string that contains the friendly Fully Qualified Device Description (FQDD), " +
				"which is a property that describes the device and its location",
		},
		"bus_number": schema.Int64Attribute{
			Computed:            true,
			MarkdownDescription: "The bus number where this PCI device resides",
			Description:         "The bus number where this PCI device resides",
		},
		"controller_bios_version": schema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "This property represents the firmware version of Controller BIOS",
			Description:         "This property represents the firmware version of Controller BIOS",
		},
		"data_bus_width": schema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "This property represents the data-bus width of the NIC PCI device",
			Description:         "This property represents the data-bus width of the NIC PCI device",
		},
		"efi_version": schema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "This property represents the firmware version of EFI",
			Description:         "This property represents the firmware version of EFI",
		},
		"fcoe_offload_mode": schema.StringAttribute{
			Computed: true,
			MarkdownDescription: "This property indicates if Fibre Channel over Ethernet (FCoE) personality is enabled or disabled" +
				" on current partition in a Converged Network Adaptor device",
			Description: "This property indicates if Fibre Channel over Ethernet (FCoE) personality is enabled or disabled" +
				" on current partition in a Converged Network Adaptor device",
		},
		"fc_os_driver_version": schema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "This property represents the FCOS OS Driver version for the DellNIC",
			Description:         "This property represents the FCOS OS Driver version for the DellNIC",
		},
		"fqdd": schema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "A string that contains the Fully Qualified Device Description (FQDD) for the DellNIC",
			Description:         "A string that contains the Fully Qualified Device Description (FQDD) for the DellNIC",
		},
		"family_version": schema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "Represents family version of firmware",
			Description:         "Represents family version of firmware",
		},
		"instance_id": schema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "A unique identifier for the instance",
			Description:         "A unique identifier for the instance",
		},
		"last_system_inventory_time": schema.StringAttribute{
			Computed: true,
			MarkdownDescription: "This property represents the time when System Inventory Collection On Reboot (CSIOR) was last performed or " +
				"the object was last updated on iDRAC. The value is represented in the format yyyymmddHHMMSS",
			Description: "This property represents the time when System Inventory Collection On Reboot (CSIOR) was last performed or " +
				"the object was last updated on iDRAC. The value is represented in the format yyyymmddHHMMSS",
		},
		"link_duplex": schema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "This property indicates whether the Link is full-duplex or half-duplex",
			Description:         "This property indicates whether the Link is full-duplex or half-duplex",
		},
		"last_update_time": schema.StringAttribute{
			Computed: true,
			MarkdownDescription: "This property represents the time when the data was last updated. The value is represented " +
				"in the format yyyymmddHHMMSS",
			Description: "This property represents the time when the data was last updated. The value is represented " +
				"in the format yyyymmddHHMMSS",
		},
		"media_type": schema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "The property shall represent the drive media type",
			Description:         "The property shall represent the drive media type",
		},
		"nic_mode": schema.StringAttribute{
			Computed: true,
			MarkdownDescription: "Represents if network interface card personality is enabled or disabled on current partition " +
				"in a Converged Network Adaptor device",
			Description: "Represents if network interface card personality is enabled or disabled on current partition " +
				"in a Converged Network Adaptor device",
		},
		"pci_device_id": schema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "This property contains a value assigned by the device manufacturer used to identify the type of device",
			Description:         "This property contains a value assigned by the device manufacturer used to identify the type of device",
		},
		"pci_vendor_id": schema.StringAttribute{
			Computed: true,
			MarkdownDescription: "This property represents the register that contains a value assigned by the PCI SIG used to " +
				"identify the manufacturer of the device",
			Description: "This property represents the register that contains a value assigned by the PCI SIG used to " +
				"identify the manufacturer of the device",
		},
		NICComponmentSchemaPartNumber: schema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "The part number assigned by the organization that is responsible for producing or manufacturing the NIC device",
			Description:         "The part number assigned by the organization that is responsible for producing or manufacturing the NIC device",
		},
		"pci_sub_device_id": schema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "Represents PCI sub device ID",
			Description:         "Represents PCI sub device ID",
		},
		"pci_sub_vendor_id": schema.StringAttribute{
			Computed: true,
			MarkdownDescription: "This property represents the subsystem vendor ID. ID information is reported from " +
				"a PCIDevice through protocol-specific requests",
			Description: "This property represents the subsystem vendor ID. ID information is reported from " +
				"a PCIDevice through protocol-specific requests",
		},
		"product_name": schema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "A string containing the product name",
			Description:         "A string containing the product name",
		},
		"protocol": schema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "Supported Protocol Types",
			Description:         "Supported Protocol Types",
		},
		"snapi_state": schema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "This property represents the SNAPI state",
			Description:         "This property represents the SNAPI state",
		},
		NICComponmentSchemaSerialNumber: schema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "A manufacturer-allocated number used to identify the NIC device",
			Description:         "A manufacturer-allocated number used to identify the NIC device",
		},
		"snapi_support": schema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "This property represents the SNAPI support",
			Description:         "This property represents the SNAPI support",
		},
		"slot_length": schema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "This property represents the represents the slot length of the NIC PCI device",
			Description:         "This property represents the represents the slot length of the NIC PCI device",
		},
		"slot_type": schema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "This property indicates the slot type of the NIC PCI device",
			Description:         "This property indicates the slot type of the NIC PCI device",
		},
		"vpi_support": schema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "This property represents the VPI support",
			Description:         "This property represents the VPI support",
		},
		"vendor_name": schema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "This property represents the vendor name",
			Description:         "This property represents the vendor name",
		},
		"iscsi_offload_mode": schema.StringAttribute{
			Computed: true,
			MarkdownDescription: "This property indicates if Internet Small Computer System Interface (iSCSI) personality is enabled or " +
				"disabled on current partition in a Converged Network Adaptor device",
			Description: "This property indicates if Internet Small Computer System Interface (iSCSI) personality is enabled or " +
				"disabled on current partition in a Converged Network Adaptor device",
		},
		"identifier_type": schema.StringAttribute{
			Computed: true,
			MarkdownDescription: "This property represents the type of Small Form Factor pluggable(SFP) Transceiver for the DellNIC. " +
				NICSchemaDescriptionForDeprecatedNoteV440,
			Description: "This property represents the type of Small Form Factor pluggable(SFP) Transceiver for the DellNIC. " +
				NICSchemaDescriptionForDeprecatedNoteV440,
		},
		"transceiver_vendor_name": schema.StringAttribute{
			Computed: true,
			MarkdownDescription: "This property represents the vendor name of Small Form Factor pluggable(SFP) Transceiver for the DellNIC." +
				NICSchemaDescriptionForDeprecatedNoteV440,
			Description: "This property represents the vendor name of Small Form Factor pluggable(SFP) Transceiver for the DellNIC." +
				NICSchemaDescriptionForDeprecatedNoteV440,
		},
		"cable_length_metres": schema.Int64Attribute{
			Computed: true,
			MarkdownDescription: "This property represents the cable length of Small Form Factor pluggable(SFP) Transceiver for the DellNIC. " +
				NICSchemaDescriptionForDeprecatedNoteV420,
			Description: "This property represents the cable length of Small Form Factor pluggable(SFP) Transceiver for the DellNIC. " +
				NICSchemaDescriptionForDeprecatedNoteV420,
		},
		"permanent_fcoe_emac_address": schema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "PermanentFCOEMACAddress defines the network address that is hardcoded into a port for FCoE",
			Description:         "PermanentFCOEMACAddress defines the network address that is hardcoded into a port for FCoE",
		},
		"permanent_iscsi_emac_address": schema.StringAttribute{
			Computed: true,
			MarkdownDescription: "PermanentAddress defines the network address that is hardcoded into a port for iSCSI. " +
				"This 'hardcoded' address can be changed using a firmware upgrade or a software configuration. " +
				"When this change is made, the field should be updated at the same time. " +
				"PermanentAddress should be left blank if no 'hardcoded' address exists for the NetworkAdapter.",
			Description: "PermanentAddress defines the network address that is hardcoded into a port for iSCSI. " +
				"This 'hardcoded' address can be changed using a firmware upgrade or a software configuration. " +
				"When this change is made, the field should be updated at the same time. " +
				"PermanentAddress should be left blank if no 'hardcoded' address exists for the NetworkAdapter.",
		},
		"revision": schema.StringAttribute{
			Computed: true,
			MarkdownDescription: "This property represents the revision number of the Small Form Factor pluggable(SFP) Transceiver. " +
				NICSchemaDescriptionForDeprecatedNoteV440,
			Description: "This property represents the revision number of the Small Form Factor pluggable(SFP) Transceiver. " +
				NICSchemaDescriptionForDeprecatedNoteV440,
		},
		"transceiver_part_number": schema.StringAttribute{
			Computed: true,
			MarkdownDescription: "The part number assigned by the organization that is responsible for producing or SFP Transceivers" +
				"(manufacturing the Small Form Factor pluggable). " +
				NICSchemaDescriptionForDeprecatedNoteV440,
			Description: "The part number assigned by the organization that is responsible for producing or SFP Transceivers" +
				"(manufacturing the Small Form Factor pluggable). " +
				NICSchemaDescriptionForDeprecatedNoteV440,
		},
		"transceiver_serial_number": schema.StringAttribute{
			Computed:            true,
			MarkdownDescription: NICSchemaDescriptionForSerialNumber + NICSchemaDescriptionForDeprecatedNoteV440,
			Description:         NICSchemaDescriptionForSerialNumber + NICSchemaDescriptionForDeprecatedNoteV440,
		},
	}
}

// EthernetSchema is a function that returns the schema for ethernet.
func EthernetSchema() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"mac_address": schema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "The currently configured MAC address",
			Description:         "The currently configured MAC address",
		},
		"mtu_size": schema.Int64Attribute{
			Computed:            true,
			MarkdownDescription: "The maximum transmission unit (MTU) configured for this network device function",
			Description:         "The maximum transmission unit (MTU) configured for this network device function",
		},
		"permanent_mac_address": schema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "The permanent MAC address assigned to this function",
			Description:         "The permanent MAC address assigned to this function",
		},
		"vlan": schema.SingleNestedAttribute{
			Computed:            true,
			MarkdownDescription: "The attributes of a VLAN",
			Description:         "The attributes of a VLAN",
			Attributes:          VLANSchema(),
		},
	}
}

// VLANSchema is a function that returns the schema for vlan.
func VLANSchema() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"vlan_id": schema.Int64Attribute{
			Computed:            true,
			MarkdownDescription: "The vlan id of the network device function",
			Description:         "The vlan id of the network device function",
		},
		"vlan_enabled": schema.BoolAttribute{
			Computed:            true,
			MarkdownDescription: "An indication of whether the VLAN is enabled",
			Description:         "An indication of whether the VLAN is enabled",
		},
	}
}

// FibreChannelSchema is a function that returns the schema for fibre channel.
func FibreChannelSchema() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"allow_fip_vlan_discovery": schema.BoolAttribute{
			Computed:            true,
			MarkdownDescription: "An indication of whether the FCoE Initialization Protocol (FIP) populates the FCoE VLAN ID",
			Description:         "An indication of whether the FCoE Initialization Protocol (FIP) populates the FCoE VLAN ID",
		},
		"boot_targets": schema.ListNestedAttribute{
			Description:         "A Fibre Channel boot target configured for a network device function",
			MarkdownDescription: "A Fibre Channel boot target configured for a network device function",
			Computed:            true,
			NestedObject: schema.NestedAttributeObject{
				Attributes: BootTargetSchema(),
			},
		},
		"fcoe_active_vlan_id": schema.Int64Attribute{
			Computed:            true,
			MarkdownDescription: "The active FCoE VLAN ID",
			Description:         "The active FCoE VLAN ID",
		},
		"fcoe_local_vlan_id": schema.Int64Attribute{
			Computed:            true,
			MarkdownDescription: "The locally configured FCoE VLAN ID",
			Description:         "The locally configured FCoE VLAN ID",
		},
		"permanent_wwnn": schema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "The permanent World Wide Node Name (WWNN) address assigned to this function",
			Description:         "The permanent World Wide Node Name (WWNN) address assigned to this function",
		},
		"permanent_wwpn": schema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "The permanent World Wide Port Name (WWPN) address assigned to this function",
			Description:         "The permanent World Wide Port Name (WWPN) address assigned to this function",
		},
		"wwnn": schema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "The currently configured World Wide Node Name (WWNN) address of this function",
			Description:         "The currently configured World Wide Node Name (WWNN) address of this function",
		},
		"wwn_source": schema.StringAttribute{
			Computed: true,
			MarkdownDescription: "The configuration source of the World Wide Names (WWN) for this World Wide Node Name (WWNN) and " +
				"World Wide Port Name (WWPN) connection",
			Description: "The configuration source of the World Wide Names (WWN) for this World Wide Node Name (WWNN) and " +
				"World Wide Port Name (WWPN) connection",
		},
		"wwpn": schema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "The currently configured World Wide Port Name (WWPN) address of this function",
			Description:         "The currently configured World Wide Port Name (WWPN) address of this function",
		},
		"fibre_channel_id": schema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "The Fibre Channel ID that the switch assigns for this interface",
			Description:         "The Fibre Channel ID that the switch assigns for this interface",
		},
	}
}

// BootTargetSchema is a function that returns the schema for boot target.
func BootTargetSchema() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"boot_priority": schema.Int64Attribute{
			Computed:            true,
			MarkdownDescription: "The relative priority for this entry in the boot targets array",
			Description:         "The relative priority for this entry in the boot targets array",
		},
		"lun_id": schema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "The logical unit number (LUN) ID from which to boot on the device to which the corresponding WWPN refers",
			Description:         "The logical unit number (LUN) ID from which to boot on the device to which the corresponding WWPN refers",
		},
		"wwpn": schema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "The World Wide Port Name (WWPN) from which to boot",
			Description:         "The World Wide Port Name (WWPN) from which to boot",
		},
	}
}

// ISCSIBootSchema is a function that returns the schema for iscsi boot.
func ISCSIBootSchema() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"authentication_method": schema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "The iSCSI boot authentication method for this network device function",
			Description:         "The iSCSI boot authentication method for this network device function",
		},
		"chap_secret": schema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "The shared secret for CHAP authentication",
			Description:         "The shared secret for CHAP authentication",
			Sensitive:           true,
		},
		"chap_username": schema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "The user name for CHAP authentication",
			Description:         "The user name for CHAP authentication",
		},
		"ip_address_type": schema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "The type of IP address being populated in the iSCSIBoot IP address fields",
			Description:         "The type of IP address being populated in the iSCSIBoot IP address fields",
		},
		"ip_mask_dns_via_dhcp": schema.BoolAttribute{
			Computed: true,
			MarkdownDescription: "An indication of whether the iSCSI boot initiator uses DHCP to obtain the initiator name, IP address, " +
				"and netmask",
			Description: "An indication of whether the iSCSI boot initiator uses DHCP to obtain the initiator name, IP address, " +
				"and netmask",
		},
		"initiator_default_gateway": schema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "The IPv6 or IPv4 iSCSI boot default gateway",
			Description:         "The IPv6 or IPv4 iSCSI boot default gateway",
		},
		"initiator_ip_address": schema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "The IPv6 or IPv4 address of the iSCSI initiator",
			Description:         "The IPv6 or IPv4 address of the iSCSI initiator",
		},
		"initiator_name": schema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "The iSCSI initiator name",
			Description:         "The iSCSI initiator name",
		},
		"initiator_netmask": schema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "The IPv6 or IPv4 netmask of the iSCSI boot initiator",
			Description:         "The IPv6 or IPv4 netmask of the iSCSI boot initiator",
		},
		"mutual_chap_secret": schema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "The CHAP secret for two-way CHAP authentication",
			Description:         "The CHAP secret for two-way CHAP authentication",
			Sensitive:           true,
		},
		"mutual_chap_username": schema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "The CHAP user name for two-way CHAP authentication",
			Description:         "The CHAP user name for two-way CHAP authentication",
		},
		"primary_dns": schema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "The IPv6 or IPv4 address of the primary DNS server for the iSCSI boot initiator",
			Description:         "The IPv6 or IPv4 address of the primary DNS server for the iSCSI boot initiator",
		},
		"primary_lun": schema.Int64Attribute{
			Computed:            true,
			MarkdownDescription: "The logical unit number (LUN) for the primary iSCSI boot target",
			Description:         "The logical unit number (LUN) for the primary iSCSI boot target",
		},
		"primary_target_ip_address": schema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "The IPv4 or IPv6 address for the primary iSCSI boot target",
			Description:         "The IPv4 or IPv6 address for the primary iSCSI boot target",
		},
		"primary_target_name": schema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "The name of the iSCSI primary boot target",
			Description:         "The name of the iSCSI primary boot target",
		},
		"primary_target_tcp_port": schema.Int64Attribute{
			Computed:            true,
			MarkdownDescription: "The TCP port for the primary iSCSI boot target",
			Description:         "The TCP port for the primary iSCSI boot target",
		},
		"primary_vlan_enable": schema.BoolAttribute{
			Computed:            true,
			MarkdownDescription: "An indication of whether the primary VLAN is enabled",
			Description:         "An indication of whether the primary VLAN is enabled",
		},
		"primary_vlan_id": schema.Int64Attribute{
			Computed:            true,
			MarkdownDescription: "The 802.1q VLAN ID to use for iSCSI boot from the primary target",
			Description:         "The 802.1q VLAN ID to use for iSCSI boot from the primary target",
		},
		"router_advertisement_enabled": schema.BoolAttribute{
			Computed:            true,
			MarkdownDescription: "An indication of whether IPv6 router advertisement is enabled for the iSCSI boot target",
			Description:         "An indication of whether IPv6 router advertisement is enabled for the iSCSI boot target",
		},
		"secondary_dns": schema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "The IPv6 or IPv4 address of the secondary DNS server for the iSCSI boot initiator",
			Description:         "The IPv6 or IPv4 address of the secondary DNS server for the iSCSI boot initiator",
		},
		"secondary_lun": schema.Int64Attribute{
			Computed:            true,
			MarkdownDescription: "The logical unit number (LUN) for the secondary iSCSI boot target",
			Description:         "The logical unit number (LUN) for the secondary iSCSI boot target",
		},
		"secondary_target_ip_address": schema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "The IPv4 or IPv6 address for the secondary iSCSI boot target",
			Description:         "The IPv4 or IPv6 address for the secondary iSCSI boot target",
		},
		"secondary_target_name": schema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "The name of the iSCSI secondary boot target",
			Description:         "The name of the iSCSI secondary boot target",
		},
		"secondary_target_tcp_port": schema.Int64Attribute{
			Computed:            true,
			MarkdownDescription: "The TCP port for the secondary iSCSI boot target",
			Description:         "The TCP port for the secondary iSCSI boot target",
		},
		"secondary_vlan_enable": schema.BoolAttribute{
			Computed:            true,
			MarkdownDescription: "An indication of whether the secondary VLAN is enabled",
			Description:         "An indication of whether the secondary VLAN is enabled",
		},
		"secondary_vlan_id": schema.Int64Attribute{
			Computed:            true,
			MarkdownDescription: "The 802.1q VLAN ID to use for iSCSI boot from the secondary target",
			Description:         "The 802.1q VLAN ID to use for iSCSI boot from the secondary target",
		},
		"target_info_via_dhcp": schema.BoolAttribute{
			Computed:            true,
			MarkdownDescription: "An indication of whether the iSCSI boot target name, LUN, IP address, and netmask should be obtained from DHCP",
			Description:         "An indication of whether the iSCSI boot target name, LUN, IP address, and netmask should be obtained from DHCP",
		},
	}
}
