/*
Copyright (c) 2023-2024 Dell Inc., or its subsidiaries. All Rights Reserved.

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
	resp.TypeName = req.ProviderTypeName + "nic"
}

// Schema implements datasource.DataSource.
func (*NICDatasource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "This Terraform datasource is used to query existing network interface cards(NIC) configuration including network adapters, network ports, network device functions and their OEM attributes." +
			" The information fetched from this block can be further used for resource block.",
		Description: "This Terraform datasource is used to query existing network interface cards(NIC) configuration including network adapters, network ports, network device functions and their OEM attributes." +
			" The information fetched from this block can be further used for resource block.",
		Attributes: NICDatasourceSchema(),
		Blocks: map[string]schema.Block{
			"nic_filter": schema.ListNestedBlock{
				MarkdownDescription: "NIC filter for resources, nework adapters, network ports and network device functions",
				Description:         "NIC filter for resources, nework adapters, network ports and network device functions",
				Validators: []validator.List{
					listvalidator.UniqueValues(),
					listvalidator.SizeAtLeast(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: NICFilterSchema(),
				},
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
	service, err := NewConfig(g.p, &plan.RedfishServer)
	if err != nil {
		resp.Diagnostics.AddError("service error", err.Error())
		return
	}
	g.ctx = ctx
	g.service = service
	state, diags := g.readDatasourceRedfishNIC(plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	// notFound := setFilterDiff(controllers, foundControllers)
	// for _, cont := range notFound {
	// 	diags.AddError("Could not find Controller "+cont, "")  // add warning
	// }
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (g *NICDatasource) readDatasourceRedfishNIC(d models.NICDatasource) (models.NICDatasource, diag.Diagnostics) {
	var diags diag.Diagnostics

	// write the current time as ID
	d.ID = types.StringValue(fmt.Sprintf("%d", time.Now().Unix()))

	// re-use readDatasourceRedfishDellIdracAttributes and select NIC.* attributes
	if diag := loadNICAttributesState(g.service, &d); diag.HasError() {
		return d, diag
	}

	systems, err := g.service.Systems()
	if err != nil {
		diags.AddError("Error fetching computer systems collection", err.Error())
		return d, diags
	}

	for _, system := range systems {
		var found bool
		var adapterFilters []models.NetworkAdapterFilter
		for _, filter := range d.NICFilter {
			if filter.SystemID.ValueString() == system.ID {
				found = true
				adapterFilters = filter.NetworkAdapters
				break
			}
		}
		if d.NICFilter != nil && !found {
			continue
		}

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
						filteredPorts = append(filteredPorts, port)
					}
				}
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
						filteredDeviceFunctions = append(filteredDeviceFunctions, devFunc)
					}
				}
			}
			d.NICs = append(d.NICs, newNetworkInterfaceState(networkInterface, adapter, filteredPorts, filteredDeviceFunctions))
		}
	}
	return d, diags
}

func newNetworkInterfaceState(networkInterface *redfish.NetworkInterface, adapter *redfish.NetworkAdapter, ports []*redfish.NetworkPort, deviceFunctions []*redfish.NetworkDeviceFunction) models.NetworkInterface {
	return models.NetworkInterface{
		ODataID:                types.StringValue(networkInterface.ODataID),
		Description:            types.StringValue(networkInterface.Description),
		ID:                     types.StringValue(networkInterface.ID),
		Name:                   types.StringValue(networkInterface.Name),
		Status:                 newNetworkStatus(networkInterface.Status),
		NetworkAdapter:         newNetworkAdapter(adapter),
		NetworkPorts:           newNetworkPorts(ports),                     //todo TTHE
		NetworkDeviceFunctions: newNetworkDeviceFunctions(deviceFunctions), //todo TTHE
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
	// todo TTHE
	return nil
}

func newNetworkDeviceFunctions(deviceFunctions []*redfish.NetworkDeviceFunction) []models.NetworkDeviceFunction {
	//todo TTHE
	return nil
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
					},
					"network_device_function_ids": schema.SetAttribute{
						Optional:    true,
						ElementType: types.StringType,
					},
				},
			},
		},
	}
}

// NICDatasourceSchema to define the NIC data-source schema.
func NICDatasourceSchema() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"id": schema.StringAttribute{
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
		"odata_id": schema.StringAttribute{
			Computed:    true,
			Description: "OData ID for the NIC instance",
		},
		"id": schema.StringAttribute{
			Computed:    true,
			Description: "ID of the NIC data-source",
		},
		"name": schema.StringAttribute{
			Computed:    true,
			Description: "Name of the NIC data-source",
		},
		"description": schema.StringAttribute{
			Computed:    true,
			Description: "Description of the NIC data-source",
		},
		"status": schema.SingleNestedAttribute{
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
		//TODO tthe
		"network_ports": schema.ListNestedAttribute{
			MarkdownDescription: "List of network ports fetched",
			Description:         "List of network ports fetched",
			Computed:            true,
			NestedObject: schema.NestedAttributeObject{
				Attributes: NetworkPortDataSourceSchema(),
			},
		},
		//todo TTHE
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
		"odata_id": schema.StringAttribute{
			Computed:    true,
			Description: "OData ID for the network adapter",
		},
		"id": schema.StringAttribute{
			Computed:    true,
			Description: "ID of the network adapter",
		},
		"name": schema.StringAttribute{
			Computed:    true,
			Description: "Name of the network adapter",
		},
		"description": schema.StringAttribute{
			Computed:    true,
			Description: "Description of the network adapter",
		},
		"status": schema.SingleNestedAttribute{
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
		"part_number": schema.StringAttribute{
			Computed:    true,
			Description: "Part number for this network adapter",
		},
		"serial_number": schema.StringAttribute{
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
			Description:         "The minimum number of virtual functions that can be allocated or moved between physical functions for this controller",
			MarkdownDescription: "The minimum number of virtual functions that can be allocated or moved between physical functions for this controller",
			Computed:            true,
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
			Description:         "An indication of whether this controller supports single root input/output virtualization (SR-IOV) in Virtual Ethernet Port Aggregator (VEPA) mode",
			MarkdownDescription: "An indication of whether this controller supports single root input/output virtualization (SR-IOV) in Virtual Ethernet Port Aggregator (VEPA) mode",
			Computed:            true,
		},
	}
}

// NpivSchema is a function that returns the schema for NPIV.
func NpivSchema() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"max_device_logins": schema.Int64Attribute{
			Description:         "The maximum number of N_Port ID Virtualization (NPIV) logins allowed simultaneously from all ports on this controller",
			MarkdownDescription: "The maximum number of N_Port ID Virtualization (NPIV) logins allowed simultaneously from all ports on this controller",
			Computed:            true,
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
		// "odata_id": schema.StringAttribute{
		// 	Computed:    true,
		// MarkdownDescription: "OData ID for the NIC data-source",
		// 	Description: "OData ID for the NIC data-source",
		// },
		// "id": schema.StringAttribute{
		// 	Computed:    true,
		// 	MarkdownDescription: "ID of the NIC data-source",
		// 	Description: "ID of the NIC data-source",
		// },
		// "name": schema.StringAttribute{
		// 	Computed:    true,
		// 	MarkdownDescription: "name of the NIC data-source",
		// 	Description: "name of the NIC data-source",
		// },
		// "description": schema.StringAttribute{
		// 	Computed:    true,
		// 	MarkdownDescription: "description of the NIC data-source",
		// 	Description: "description of the NIC data-source",
		// },
		// "status": schema.SingleNestedAttribute{
		// 	MarkdownDescription: "status of the NIC",
		// 	Description:         "status of the NIC",
		// 	Computed:            true,
		// 	Attributes:          NetworkStatusSchema(),
		// },
		// "active_link_technology": schema.StringAttribute{
		// 	Computed:    true,
		// 	MarkdownDescription: "description of the NIC data-source",
		// 	Description: "description of the NIC data-source",
		// },
		// "flow_control_configuration": schema.StringAttribute{
		// 	Computed:    true,
		// 	MarkdownDescription: "description of the NIC data-source",
		// 	Description: "description of the NIC data-source",
		// },
		// "flow_control_status": schema.StringAttribute{
		// 	Computed:    true,
		// 	MarkdownDescription: "description of the NIC data-source",
		// 	Description: "description of the NIC data-source",
		// },
		// "vendor_id": schema.StringAttribute{
		// 	Computed:    true,
		// 	MarkdownDescription: "description of the NIC data-source",
		// 	Description: "description of the NIC data-source",
		// },
		// "link_status": schema.StringAttribute{
		// 	Computed:    true,
		// 	MarkdownDescription: "description of the NIC data-source",
		// 	Description: "description of the NIC data-source",
		// },
		// "associated_network_addresses": schema.ListAttribute{
		// 	Description:         "List of Domain Name Server IP addresses.",
		// 	MarkdownDescription: "List of Domain Name Server IP addresses.",
		// 	ElementType:         types.StringType,
		// 	Computed:            true,
		// },
		// "supported_ethernet_capabilities": schema.ListAttribute{
		// 	Description:         "List of Domain Name Server IP addresses.",
		// 	MarkdownDescription: "List of Domain Name Server IP addresses.",
		// 	ElementType:         types.StringType,
		// 	Computed:            true,
		// },
		// "current_link_speed_mbps": schema.Int64Attribute{
		// 	Computed:    true,
		// 	MarkdownDescription: "description of the NIC data-source",
		// 	Description: "description of the NIC data-source",
		// },
		// "physical_port_number": schema.Int64Attribute{
		// 	Computed:    true,
		// 	MarkdownDescription: "description of the NIC data-source",
		// 	Description: "description of the NIC data-source",
		// },
		// "eee_enabled": schema.BoolAttribute{
		// 	Computed:    true,
		// 	MarkdownDescription: "description of the NIC data-source",
		// 	Description: "description of the NIC data-source",
		// },
		// "wake_on_lan_enabled": schema.BoolAttribute{
		// 	Computed:    true,
		// 	MarkdownDescription: "description of the NIC data-source",
		// 	Description: "description of the NIC data-source",
		// },

		////////
		// "net_dev_func_max_bw_alloc": schema.ListNestedAttribute{
		// 	MarkdownDescription: "status of the NIC",
		// 	Description:         "status of the NIC",
		// 	Computed:            true,
		// 	NestedObject:        NetworkStatusSchema(), // todo TTHE
		// },
		// "net_dev_func_min_bw_alloc": schema.ListNestedAttribute{
		// 	MarkdownDescription: "status of the NIC",
		// 	Description:         "status of the NIC",
		// 	Computed:            true,
		// 	Attributes:          NetworkStatusSchema(), // todo TTHE
		// },
		// "supported_link_capabilities": schema.ListNestedAttribute{
		// 	MarkdownDescription: "status of the NIC",
		// 	Description:         "status of the NIC",
		// 	Computed:            true,
		// 	Attributes:          NetworkStatusSchema(), // todo TTHE

		// },
	}
}

func NetworkDeviceFunctionDataSourceSchema() map[string]schema.Attribute {
	// todo TTHE
	return map[string]schema.Attribute{}
}
