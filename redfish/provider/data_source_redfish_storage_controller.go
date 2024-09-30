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

	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
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
	_ datasource.DataSource              = &StorageControllerDatasource{}
	_ datasource.DataSourceWithConfigure = &StorageControllerDatasource{}
)

// NewStorageControllerDatasource is new datasource for StorageControllerDatasource.
func NewStorageControllerDatasource() datasource.DataSource {
	return &StorageControllerDatasource{}
}

// StorageControllerDatasource to construct datasource.
type StorageControllerDatasource struct {
	p       *redfishProvider
	ctx     context.Context
	service *gofish.Service
}

// Configure implements datasource.DataSourceWithConfigure.
func (g *StorageControllerDatasource) Configure(_ context.Context, req datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	g.p = req.ProviderData.(*redfishProvider)
}

// Metadata implements datasource.DataSource.
func (*StorageControllerDatasource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "storage_controller"
}

// Schema implements datasource.DataSource.
func (*StorageControllerDatasource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "This Terraform datasource is used to query existing storage controller configuration." +
			" The information fetched from this block can be further used for resource block.",
		Description: "This Terraform datasource is used to query existing storage controller configuration." +
			" The information fetched from this block can be further used for resource block.",
		Attributes: StorageControllerDatasourceSchema(),
		Blocks: map[string]schema.Block{
			"storage_controller_filter": schema.SingleNestedBlock{
				MarkdownDescription: "Storage Controller filter for systems, storages and controllers",
				Description:         "Storage Controller filter for systems, storages and controllers",
				Attributes:          StorageControllerFilterSchema(),
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
func (g *StorageControllerDatasource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var plan models.StorageControllerDatasource
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
	state, diags := g.readDatasourceRedfishStorageController(plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// readDatasourceRedfishStorageController populates the storage controllers in the datasource model.
// nolint: gocyclo, gocognit, revive
func (g *StorageControllerDatasource) readDatasourceRedfishStorageController(d models.StorageControllerDatasource) (models.StorageControllerDatasource, diag.Diagnostics) {
	var diags diag.Diagnostics

	stringJoinSplit := " ,"

	// write the current time as ID
	d.ID = types.StringValue(fmt.Sprintf("%d", time.Now().Unix()))

	var systemFilter []models.SystemsFilter
	if d.StorageControllerFilter != nil {
		// get the filter for systems
		systemFilter = d.StorageControllerFilter.Systems
	}

	// get all the systems
	systemList, err := g.service.Systems()
	if err != nil {
		diags.AddError("Error fetching computer systems collection", err.Error())
		return d, diags
	}
	var validSystemIDs []string

	// iterating through all the systems
	for _, system := range systemList {
		foundSystem := false
		var storageFilter []models.StoragesFilter

		for _, filteredSystemItem := range systemFilter {
			// if the system under consideration is part of the filter
			if filteredSystemItem.SystemID.ValueString() == system.ID {
				foundSystem = true
				// get the filter for storages
				storageFilter = filteredSystemItem.Storages
				break
			}
		}

		// if the filter exists but the system under consideration is not part of the filter
		if len(systemFilter) > 0 && !foundSystem {
			continue
		}

		// reached here means
		// either the filter itself doesn't exist in which case we need to consider all systems
		// or the filter exists and the system is part of it
		validSystemID := system.ID
		validSystemIDs = append(validSystemIDs, validSystemID)

		// get all storages
		storageList, err := system.Storage()
		if err != nil {
			diags.AddError("Error fetching storages collection", err.Error())
			return d, diags
		}
		var validStorageIDs []string

		// iterating through all the storages
		for _, storage := range storageList {
			foundStorage := false
			var controllerFilter []string

			for _, filteredStorageItem := range storageFilter {
				// if the storage under consideration is part of the filter
				if filteredStorageItem.StorageID.ValueString() == storage.ID {
					foundStorage = true
					// get the filter for storage controllers
					for _, controllerID := range filteredStorageItem.ControllerIDs {
						controllerFilter = append(controllerFilter, controllerID.ValueString())
					}
					break
				}
			}

			// if the filter exists but the storage under consideration is not part of the filter
			if len(storageFilter) > 0 && !foundStorage {
				continue
			}

			// reached here means
			// either the filter itself doesn't exist in which case we need to consider all storages
			// or the filter exists and the storage is part of it
			validStorageID := storage.ID
			validStorageIDs = append(validStorageIDs, validStorageID)

			// get all storage controllers
			storageControllerList, err := storage.Controllers()
			if err != nil {
				diags.AddError("Error fetching storage controllers collection", err.Error())
				return d, diags
			}
			var validStorageControllerIDs []string

			// iterating through all the storage controllers
			for _, storageController := range storageControllerList {
				foundController := false

				for _, controllerID := range controllerFilter {
					// if the storage controller under consideration is part of the filter
					if controllerID == storageController.ID {
						foundController = true
						break
					}
				}

				// if the filter exists but the storage controller under consideration is not part of the filter
				if len(controllerFilter) > 0 && !foundController {
					continue
				}

				// reached here means
				// either the filter itself doesn't exist in which case we need to consider all storage controllers
				// or the filter exists and the storage controller is part of it
				validStorageControllerID := storageController.ID
				validStorageControllerIDs = append(validStorageControllerIDs, validStorageControllerID)

				d.StorageControllers = append(d.StorageControllers, newStorageController(storageController))
			}
			// check for an invalid storage controller id for a storage id for a system id in the filter
			if len(controllerFilter) != 0 && len(validStorageControllerIDs) != len(controllerFilter) {
				diags.AddError(
					fmt.Sprintf("Error one or more of the filtered storage controller ids are not valid for the system id %s and storage id %s", validSystemID, validStorageID),
					fmt.Sprintf("Valid storage controller ids are [%v]", strings.Join(validStorageControllerIDs, stringJoinSplit)),
				)
				return d, diags
			}

		}
		// check for an invalid storage id for a system id in the filter
		if len(storageFilter) != 0 && len(validStorageIDs) != len(storageFilter) {
			diags.AddError(
				fmt.Sprintf("Error one or more of the filtered storage ids are not valid for the system id %s", validSystemID),
				fmt.Sprintf("Valid storage ids are [%v]", strings.Join(validStorageIDs, stringJoinSplit)),
			)
			return d, diags
		}

	}
	// check for an invalid system id in the filter
	if len(systemFilter) != 0 && len(validSystemIDs) != len(systemFilter) {
		diags.AddError(
			"Error one or more of the filtered system ids are not valid.",
			fmt.Sprintf("Valid system ids are [%v]", strings.Join(validSystemIDs, stringJoinSplit)),
		)
		return d, diags
	}

	return d, diags
}

// newStorageController converts redfish.StorageController to models.StorageController
func newStorageController(storageController *redfish.StorageController) models.StorageController {
	return models.StorageController{
		ODataID:                      types.StringValue(storageController.ODataID),
		ID:                           types.StringValue(storageController.ID),
		Description:                  newStorageControllerDescription(storageController),
		Name:                         types.StringValue(storageController.Name),
		Assembly:                     newStorageControllerAssembly(storageController),
		CacheSummary:                 newStorageControllerCacheSummary(storageController.CacheSummary),
		ControllerRates:              newStorageControllerControllerRates(storageController.ControllerRates),
		FirmwareVersion:              types.StringValue(storageController.FirmwareVersion),
		Identifiers:                  newStorageControllerIdentifiers(storageController.Identifiers),
		Links:                        newStorageControllerLinks(storageController),
		Manufacturer:                 types.StringValue(storageController.Manufacturer),
		Model:                        types.StringValue(storageController.Model),
		Oem:                          newStorageControllerOEM(storageController),
		SpeedGbps:                    types.Float64Value(float64(storageController.SpeedGbps)),
		Status:                       newStorageControllerStatus(storageController.Status),
		SupportedControllerProtocols: newStorageControllerSupportedControllerProtocols(storageController.SupportedControllerProtocols),
		SupportedDeviceProtocols:     newStorageControllerSupportedDeviceProtocols(storageController.SupportedDeviceProtocols),
		SupportedRAIDTypes:           newStorageControllerSupportedRAIDTypes(storageController.SupportedRAIDTypes),
	}
}

// newStorageControllerAssembly given redfish.StorageController populates models.Assembly
func newStorageControllerAssembly(storageController *redfish.StorageController) models.Assembly {
	storageControllerExtended, _ := dell.StorageController(storageController)
	assembly := storageControllerExtended.Assembly
	return models.Assembly{
		ODataID: types.StringValue(assembly.ODataID),
	}
}

// newStorageControllerLinks given redfish.StorageController populates models.Links
func newStorageControllerLinks(storageController *redfish.StorageController) models.Links {
	storageControllerExtended, _ := dell.StorageController(storageController)
	pcieFunctions := storageControllerExtended.PCIeFunctions
	return models.Links{
		PCIeFunctions: newStorageControllerLinksPCIeFunctions(pcieFunctions),
	}
}

// newStorageControllerLinksPCIeFunctions given []string populates []models.PCIeFunction
func newStorageControllerLinksPCIeFunctions(pcieFunctions []string) []models.PCIeFunction {
	var values []models.PCIeFunction
	for _, val := range pcieFunctions {
		var value models.PCIeFunction
		value.ODataID = types.StringValue(val)
		values = append(values, value)
	}
	return values
}

// newStorageControllerDescription given redfish.StorageController populates description
func newStorageControllerDescription(storageController *redfish.StorageController) types.String {
	storageControllerExtended, _ := dell.StorageController(storageController)
	description := storageControllerExtended.Description
	return types.StringValue(description)
}

// newStorageControllerOEM given redfish.StorageController populates models.StorageControllerOEM
func newStorageControllerOEM(storageController *redfish.StorageController) models.StorageControllerOEM {
	storageControllerExtended, _ := dell.StorageController(storageController)
	return models.StorageControllerOEM{
		Dell: newStorageControllerOEMDell(storageControllerExtended.Oem.Dell),
	}
}

// newStorageControllerOEMDell given dell.StorageControllerOEMDell populates models.StorageControllerOEMDell
func newStorageControllerOEMDell(input dell.StorageControllerOEMDell) models.StorageControllerOEMDell {
	return models.StorageControllerOEMDell{
		DellStorageController: newDellStorageController(input.DellStorageController),
	}
}

// newDellStorageController given dell.DellStorageController populates models.DellStorageController
func newDellStorageController(input dell.DellStorageController) models.DellStorageController {
	return models.DellStorageController{
		AlarmState:                          types.StringValue(input.AlarmState),
		AutoConfigBehavior:                  types.StringValue(input.AutoConfigBehavior),
		BackgroundInitializationRatePercent: types.Int64Value(input.BackgroundInitializationRatePercent),
		BatteryLearnMode:                    types.StringValue(input.BatteryLearnMode),
		BootVirtualDiskFQDD:                 types.StringValue(input.BootVirtualDiskFQDD),
		CacheSizeInMB:                       types.Int64Value(input.CacheSizeInMB),
		CachecadeCapability:                 types.StringValue(input.CachecadeCapability),
		CheckConsistencyMode:                types.StringValue(input.CheckConsistencyMode),
		ConnectorCount:                      types.Int64Value(input.ConnectorCount),
		ControllerBootMode:                  types.StringValue(input.ControllerBootMode),
		ControllerFirmwareVersion:           types.StringValue(input.ControllerFirmwareVersion),
		ControllerMode:                      types.StringValue(input.ControllerMode),
		CopybackMode:                        types.StringValue(input.CopybackMode),
		CurrentControllerMode:               types.StringValue(input.CurrentControllerMode),
		Device:                              types.StringValue(input.Device),
		DeviceCardDataBusWidth:              types.StringValue(input.DeviceCardDataBusWidth),
		DeviceCardSlotLength:                types.StringValue(input.DeviceCardSlotLength),
		DeviceCardSlotType:                  types.StringValue(input.DeviceCardSlotType),
		DriverVersion:                       types.StringValue(input.DriverVersion),
		EncryptionCapability:                types.StringValue(input.EncryptionCapability),
		EncryptionMode:                      types.StringValue(input.EncryptionMode),
		EnhancedAutoImportForeignConfigurationMode: types.StringValue(input.EnhancedAutoImportForeignConfigurationMode),
		KeyID:                            types.StringValue(input.KeyID),
		LastSystemInventoryTime:          types.StringValue(input.LastSystemInventoryTime),
		LastUpdateTime:                   types.StringValue(input.LastUpdateTime),
		LoadBalanceMode:                  types.StringValue(input.LoadBalanceMode),
		MaxAvailablePCILinkSpeed:         types.StringValue(input.MaxAvailablePCILinkSpeed),
		MaxDrivesInSpanCount:             types.Int64Value(input.MaxDrivesInSpanCount),
		MaxPossiblePCILinkSpeed:          types.StringValue(input.MaxPossiblePCILinkSpeed),
		MaxSpansInVolumeCount:            types.Int64Value(input.MaxSpansInVolumeCount),
		MaxSupportedVolumesCount:         types.Int64Value(input.MaxSupportedVolumesCount),
		PCISlot:                          types.StringValue(input.PCISlot),
		PatrolReadIterationsCount:        types.Int64Value(input.PatrolReadIterationsCount),
		PatrolReadMode:                   types.StringValue(input.PatrolReadMode),
		PatrolReadRatePercent:            types.Int64Value(input.PatrolReadRatePercent),
		PatrolReadState:                  types.StringValue(input.PatrolReadState),
		PatrolReadUnconfiguredAreaMode:   types.StringValue(input.PatrolReadUnconfiguredAreaMode),
		PersistentHotspare:               types.StringValue(input.PersistentHotspare),
		PersistentHotspareMode:           types.StringValue(input.PersistentHotspareMode),
		RAIDMode:                         types.StringValue(input.RAIDMode),
		RealtimeCapability:               types.StringValue(input.RealtimeCapability),
		ReconstructRatePercent:           types.Int64Value(input.ReconstructRatePercent),
		RollupStatus:                     types.StringValue(input.RollupStatus),
		SASAddress:                       types.StringValue(input.SASAddress),
		SecurityStatus:                   types.StringValue(input.SecurityStatus),
		SharedSlotAssignmentAllowed:      types.StringValue(input.SharedSlotAssignmentAllowed),
		SlicedVDCapability:               types.StringValue(input.SlicedVDCapability),
		SpindownIdleTimeSeconds:          types.Int64Value(input.SpindownIdleTimeSeconds),
		SupportControllerBootMode:        types.StringValue(input.SupportControllerBootMode),
		SupportEnhancedAutoForeignImport: types.StringValue(input.SupportEnhancedAutoForeignImport),
		SupportRAID10UnevenSpans:         types.StringValue(input.SupportRAID10UnevenSpans),
		SupportedInitializationTypes:     newTypesStringList(input.SupportedInitializationTypes),
		SupportsLKMtoSEKMTransition:      types.StringValue(input.SupportsLKMtoSEKMTransition),
		T10PICapability:                  types.StringValue(input.T10PICapability),
	}
}

// newStorageControllerCacheSummary converts redfish.CacheSummary to models.CacheSummary
func newStorageControllerCacheSummary(cacheSummary redfish.CacheSummary) models.CacheSummary {
	return models.CacheSummary{
		TotalCacheSizeMiB: types.Int64Value(int64(cacheSummary.TotalCacheSizeMiB)),
	}
}

// newStorageControllerControllerRates converts redfish.Rates to models.ControllerRates
func newStorageControllerControllerRates(controllerRates redfish.Rates) models.ControllerRates {
	return models.ControllerRates{
		ConsistencyCheckRatePercent: types.Int64Value(int64(controllerRates.ConsistencyCheckRatePercent)),
		RebuildRatePercent:          types.Int64Value(int64(controllerRates.RebuildRatePercent)),
	}
}

// newStorageControllerIdentifiers converts []common.Identifier to models.Identifier
func newStorageControllerIdentifiers(identifiers []common.Identifier) []models.Identifier {
	var values []models.Identifier
	for _, identifier := range identifiers {
		var val models.Identifier
		val.DurableName = types.StringValue(identifier.DurableName)
		val.DurableNameFormat = types.StringValue(string(identifier.DurableNameFormat))
		values = append(values, val)
	}
	return values
}

// newStorageControllerStatus converts common.Status to models.Status
func newStorageControllerStatus(status common.Status) models.Status {
	return models.Status{
		Health:       types.StringValue(string(status.Health)),
		HealthRollup: types.StringValue(string(status.HealthRollup)),
		State:        types.StringValue(string(status.State)),
	}
}

// newStorageControllerSupportedControllerProtocols converts []common.Protocol to []types.String
func newStorageControllerSupportedControllerProtocols(supportedControllerProtocols []common.Protocol) []types.String {
	var values []string
	for _, val := range supportedControllerProtocols {
		values = append(values, string(val))
	}
	return newTypesStringList(values)
}

// newStorageControllerSupportedDeviceProtocols converts []common.Protocol to []types.String
func newStorageControllerSupportedDeviceProtocols(supportedDeviceProtocols []common.Protocol) []types.String {
	var values []string
	for _, val := range supportedDeviceProtocols {
		values = append(values, string(val))
	}
	return newTypesStringList(values)
}

// newStorageControllerSupportedRAIDTypes converts []redfish.RAIDType to []types.String
func newStorageControllerSupportedRAIDTypes(supportedRAIDTypes []redfish.RAIDType) []types.String {
	var values []string
	for _, val := range supportedRAIDTypes {
		values = append(values, string(val))
	}
	return newTypesStringList(values)
}

// StorageControllerFilterSchema to construct schema of storage controller filter.
func StorageControllerFilterSchema() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"systems": schema.ListNestedAttribute{
			Optional:    true,
			Description: "Filter for systems, storages and storage controllers",
			Validators: []validator.List{
				listvalidator.UniqueValues(),
			},
			NestedObject: schema.NestedAttributeObject{
				Attributes: map[string]schema.Attribute{
					"system_id": schema.StringAttribute{
						Required:    true,
						Description: "Filter for systems",
					},
					"storages": schema.ListNestedAttribute{
						Optional:    true,
						Description: "Filter for storages and storage controllers",
						Validators: []validator.List{
							listvalidator.UniqueValues(),
						},
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"storage_id": schema.StringAttribute{
									Required:    true,
									Description: "Filter for storages",
								},
								"controller_ids": schema.SetAttribute{
									Optional:    true,
									ElementType: types.StringType,
									Description: "Filter for storage controllers",
								},
							},
						},
					},
				},
			},
		},
	}
}

// StorageControllerDatasourceSchema to define the Storage Controller data-source schema.
func StorageControllerDatasourceSchema() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"id": schema.StringAttribute{
			MarkdownDescription: "ID of the storage controller data-source.",
			Description:         "ID of the storage controller data-source.",
			Computed:            true,
		},
		"storage_controllers": schema.ListNestedAttribute{
			MarkdownDescription: "List of storage controllers fetched.",
			Description:         "List of storage controllers fetched.",
			NestedObject: schema.NestedAttributeObject{
				Attributes: StorageControllerSchema(),
			},
			Computed: true,
		},
	}
}

// StorageControllerSchema is a function that returns the schema for Storage Controller.
func StorageControllerSchema() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"odata_id": schema.StringAttribute{
			Description: "The unique identifier for a resource.",
			Computed:    true,
		},
		"id": schema.StringAttribute{
			Description: "The unique identifier for this resource within the collection of similar resources.",
			Computed:    true,
		},
		"description": schema.StringAttribute{
			Description: "The description of this resource. Used for commonality in the schema definitions.",
			Computed:    true,
		},
		"name": schema.StringAttribute{
			Description: "The name of the resource or array member.",
			Computed:    true,
		},
		"assembly": schema.SingleNestedAttribute{
			MarkdownDescription: "A reference to a resource.",
			Description:         "A reference to a resource.",
			Computed:            true,
			Attributes:          AssemblySchema(),
		},
		"cache_summary": schema.SingleNestedAttribute{
			MarkdownDescription: "This type describes the cache memory of the storage controller in general detail.",
			Description:         "This type describes the cache memory of the storage controller in general detail.",
			Computed:            true,
			Attributes:          CacheSummarySchema(),
		},
		"controller_rates": schema.SingleNestedAttribute{
			MarkdownDescription: "This type describes the various controller rates used for processes such as volume rebuild or consistency checks.",
			Description:         "This type describes the various controller rates used for processes such as volume rebuild or consistency checks.",
			Computed:            true,
			Attributes:          ControllerRatesSchema(),
		},
		"firmware_version": schema.StringAttribute{
			MarkdownDescription: "The firmware version of this storage controller.",
			Description:         "The firmware version of this storage controller.",
			Computed:            true,
		},
		"identifiers": schema.ListNestedAttribute{
			MarkdownDescription: "Any additional identifiers for a resource.",
			Description:         "Any additional identifiers for a resource.",
			Computed:            true,
			NestedObject: schema.NestedAttributeObject{
				Attributes: IdentifiersSchema(),
			},
		},
		"links": schema.SingleNestedAttribute{
			MarkdownDescription: "The links to other resources that are related to this resource.",
			Description:         "The links to other resources that are related to this resource.",
			Computed:            true,
			Attributes:          LinksSchema(),
		},
		"manufacturer": schema.StringAttribute{
			MarkdownDescription: "The manufacturer of this storage controller.",
			Description:         "The manufacturer of this storage controller.",
			Computed:            true,
		},
		"model": schema.StringAttribute{
			MarkdownDescription: "The model number for the storage controller.",
			Description:         "The model number for the storage controller.",
			Computed:            true,
		},
		"oem": schema.SingleNestedAttribute{
			MarkdownDescription: "The OEM extension to the StorageController resource.",
			Description:         "The OEM extension to the StorageController resource.",
			Computed:            true,
			Attributes:          StorageControllerOEMSchema(),
		},
		"speed_gbps": schema.Float64Attribute{
			MarkdownDescription: "The maximum speed of the storage controller's device interface.",
			Description:         "The maximum speed of the storage controller's device interface.",
			Computed:            true,
		},
		"status": schema.SingleNestedAttribute{
			MarkdownDescription: "The status and health of a resource and its children.",
			Description:         "The status and health of a resource and its children.",
			Computed:            true,
			Attributes:          StatusSchema(),
		},
		"supported_controller_protocols": schema.ListAttribute{
			MarkdownDescription: "The supported set of protocols for communicating to this storage controller.",
			Description:         "The supported set of protocols for communicating to this storage controller.",
			Computed:            true,
			ElementType:         types.StringType,
		},
		"supported_device_protocols": schema.ListAttribute{
			MarkdownDescription: "The protocols that the storage controller can use to communicate with attached devices.",
			Description:         "The protocols that the storage controller can use to communicate with attached devices.",
			Computed:            true,
			ElementType:         types.StringType,
		},
		"supported_raid_types": schema.ListAttribute{
			MarkdownDescription: "The set of RAID types supported by the storage controller.",
			Description:         "The set of RAID types supported by the storage controller.",
			Computed:            true,
			ElementType:         types.StringType,
		},
	}
}

// AssemblySchema is a function that returns the schema for Assembly.
func AssemblySchema() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"odata_id": schema.StringAttribute{
			MarkdownDescription: "The link to the assembly associated with this storage controller.",
			Description:         "The link to the assembly associated with this storage controller.",
			Computed:            true,
		},
	}
}

// ControllerRatesSchema is a function that returns the schema for Controller Rates.
func ControllerRatesSchema() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"consistency_check_rate_percent": schema.Int64Attribute{
			MarkdownDescription: "This property describes the controller rate for consistency check",
			Description:         "This property describes the controller rate for consistency check",
			Computed:            true,
		},
		"rebuild_rate_percent": schema.Int64Attribute{
			MarkdownDescription: "This property describes the controller rate for volume rebuild",
			Description:         "This property describes the controller rate for volume rebuild",
			Computed:            true,
		},
	}
}

// IdentifiersSchema is a function that returns the schema for Identifiers.
func IdentifiersSchema() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"durable_name": schema.StringAttribute{
			MarkdownDescription: "This property describes the durable name for the storage controller.",
			Description:         "This property describes the durable name for the storage controller.",
			Computed:            true,
		},
		"durable_name_format": schema.StringAttribute{
			MarkdownDescription: "This property describes the durable name format for the storage controller.",
			Description:         "This property describes the durable name format for the storage controller.",
			Computed:            true,
		},
	}
}

// LinksSchema is a function that returns the schema for Links.
func LinksSchema() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"pcie_functions": schema.ListNestedAttribute{
			MarkdownDescription: "PCIeFunctions",
			Description:         "PCIeFunctions",
			Computed:            true,
			NestedObject: schema.NestedAttributeObject{
				Attributes: PCIeFunctionsSchema(),
			},
		},
	}
}

// PCIeFunctionsSchema is a function that returns the schema for PCIeFunctions.
func PCIeFunctionsSchema() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"odata_id": schema.StringAttribute{
			MarkdownDescription: "The link to the PCIeFunctions",
			Description:         "The link to the PCIeFunctions",
			Computed:            true,
		},
	}
}

// StorageControllerOEMSchema is a function that returns the schema for StorageControllerOEM.
func StorageControllerOEMSchema() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"dell": schema.SingleNestedAttribute{
			MarkdownDescription: "Dell",
			Description:         "Dell",
			Computed:            true,
			Attributes:          StorageControllerOEMDellSchema(),
		},
	}
}

// StorageControllerOEMDellSchema is a function that returns the schema for StorageControllerOEMDell.
func StorageControllerOEMDellSchema() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"dell_storage_controller": schema.SingleNestedAttribute{
			MarkdownDescription: "Dell Storage Controller",
			Description:         "Dell Storage Controller",
			Computed:            true,
			Attributes:          DellStorageControllerSchema(),
		},
	}
}

// DellStorageControllerSchema is a function that returns the schema for DellStorageController
func DellStorageControllerSchema() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"alarm_state": schema.StringAttribute{
			MarkdownDescription: "Alarm State",
			Description:         "Alarm State",
			Computed:            true,
		},
		"auto_config_behavior": schema.StringAttribute{
			MarkdownDescription: "Auto Config Behavior",
			Description:         "Auto Config Behavior",
			Computed:            true,
		},
		"background_initialization_rate_percent": schema.Int64Attribute{
			MarkdownDescription: "Background Initialization Rate Percent",
			Description:         "Background Initialization Rate Percent",
			Computed:            true,
		},
		"battery_learn_mode": schema.StringAttribute{
			MarkdownDescription: "Battery Learn Mode",
			Description:         "Battery Learn Mode",
			Computed:            true,
		},
		"boot_virtual_disk_fqdd": schema.StringAttribute{
			MarkdownDescription: "Boot Virtual Disk FQDD",
			Description:         "Boot Virtual Disk FQDD",
			Computed:            true,
		},
		"cache_size_in_mb": schema.Int64Attribute{
			MarkdownDescription: "Cache Size In MB",
			Description:         "Cache Size In MB",
			Computed:            true,
		},
		"cachecade_capability": schema.StringAttribute{
			MarkdownDescription: "Cachecade Capability",
			Description:         "Cachecade Capability",
			Computed:            true,
		},
		"check_consistency_mode": schema.StringAttribute{
			MarkdownDescription: "Check Consistency Mode",
			Description:         "Check Consistency Mode",
			Computed:            true,
		},
		"connector_count": schema.Int64Attribute{
			MarkdownDescription: "Connector Count",
			Description:         "Connector Count",
			Computed:            true,
		},
		"controller_boot_mode": schema.StringAttribute{
			MarkdownDescription: "Controller Boot Mode",
			Description:         "Controller Boot Mode",
			Computed:            true,
		},
		"controller_firmware_version": schema.StringAttribute{
			MarkdownDescription: "Controller Firmware Version",
			Description:         "Controller Firmware Version",
			Computed:            true,
		},
		"controller_mode": schema.StringAttribute{
			MarkdownDescription: "Controller Mode",
			Description:         "Controller Mode",
			Computed:            true,
		},
		"copyback_mode": schema.StringAttribute{
			MarkdownDescription: "Copyback Mode",
			Description:         "Copyback Mode",
			Computed:            true,
		},
		"current_controller_mode": schema.StringAttribute{
			MarkdownDescription: "Current Controller Mode",
			Description:         "Current Controller Mode",
			Computed:            true,
		},
		"device": schema.StringAttribute{
			MarkdownDescription: "Device",
			Description:         "Device",
			Computed:            true,
		},
		"device_card_data_bus_width": schema.StringAttribute{
			MarkdownDescription: "Device Card Data Bus Width",
			Description:         "Device Card Data Bus Width",
			Computed:            true,
		},
		"device_card_slot_length": schema.StringAttribute{
			MarkdownDescription: "Device Card Slot Length",
			Description:         "Device Card Slot Length",
			Computed:            true,
		},
		"device_card_slot_type": schema.StringAttribute{
			MarkdownDescription: "Device Card Slot Type",
			Description:         "Device Card Slot Type",
			Computed:            true,
		},
		"driver_version": schema.StringAttribute{
			MarkdownDescription: "Driver Version",
			Description:         "Driver Version",
			Computed:            true,
		},
		"encryption_capability": schema.StringAttribute{
			MarkdownDescription: "Encryption Capability",
			Description:         "Encryption Capability",
			Computed:            true,
		},
		"encryption_mode": schema.StringAttribute{
			MarkdownDescription: "Encryption Mode",
			Description:         "Encryption Mode",
			Computed:            true,
		},
		"enhanced_auto_import_foreign_configuration_mode": schema.StringAttribute{
			MarkdownDescription: "Enhanced Auto Import Foreign Configuration Mode",
			Description:         "Enhanced Auto Import Foreign Configuration Mode",
			Computed:            true,
		},
		"key_id": schema.StringAttribute{
			MarkdownDescription: "Key ID",
			Description:         "Key ID",
			Computed:            true,
		},
		"last_system_inventory_time": schema.StringAttribute{
			MarkdownDescription: "Last System Inventory Time",
			Description:         "Last System Inventory Time",
			Computed:            true,
		},
		"last_update_time": schema.StringAttribute{
			MarkdownDescription: "Last Update Time",
			Description:         "Last Update Time",
			Computed:            true,
		},
		"load_balance_mode": schema.StringAttribute{
			MarkdownDescription: "Load Balance Mode",
			Description:         "Load Balance Mode",
			Computed:            true,
		},
		"max_available_pci_link_speed": schema.StringAttribute{
			MarkdownDescription: "Max Available PCI Link Speed",
			Description:         "Max Available PCI Link Speed",
			Computed:            true,
		},
		"max_drives_in_span_count": schema.Int64Attribute{
			MarkdownDescription: "Max Drives In Span Count",
			Description:         "Max Drives In Span Count",
			Computed:            true,
		},
		"max_possible_pci_link_speed": schema.StringAttribute{
			MarkdownDescription: "Max Possible PCI Link Speed",
			Description:         "Max Possible PCI Link Speed",
			Computed:            true,
		},
		"max_spans_in_volume_count": schema.Int64Attribute{
			MarkdownDescription: "Max Spans In Volume Count",
			Description:         "Max Spans In Volume Count",
			Computed:            true,
		},
		"max_supported_volumes_count": schema.Int64Attribute{
			MarkdownDescription: "Max Supported Volumes Count",
			Description:         "Max Supported Volumes Count",
			Computed:            true,
		},
		"pci_slot": schema.StringAttribute{
			MarkdownDescription: "PCI Slot",
			Description:         "PCI Slot",
			Computed:            true,
		},
		"patrol_read_iterations_count": schema.Int64Attribute{
			MarkdownDescription: "Patrol Read Iterations Count",
			Description:         "Patrol Read Iterations Count",
			Computed:            true,
		},
		"patrol_read_mode": schema.StringAttribute{
			MarkdownDescription: "Patrol Read Mode",
			Description:         "Patrol Read Mode",
			Computed:            true,
		},
		"patrol_read_rate_percent": schema.Int64Attribute{
			MarkdownDescription: "Patrol Read Rate Percent",
			Description:         "Patrol Read Rate Percent",
			Computed:            true,
		},
		"patrol_read_state": schema.StringAttribute{
			MarkdownDescription: "Patrol Read State",
			Description:         "Patrol Read State",
			Computed:            true,
		},
		"patrol_read_unconfigured_area_mode": schema.StringAttribute{
			MarkdownDescription: "Patrol Read Unconfigured Area Mode",
			Description:         "Patrol Read Unconfigured Area Mode",
			Computed:            true,
		},
		"persistent_hotspare": schema.StringAttribute{
			MarkdownDescription: "Persistent Hotspare",
			Description:         "Persistent Hotspare",
			Computed:            true,
		},
		"persistent_hotspare_mode": schema.StringAttribute{
			MarkdownDescription: "Persistent Hotspare Mode",
			Description:         "Persistent Hotspare Mode",
			Computed:            true,
		},
		"raid_mode": schema.StringAttribute{
			MarkdownDescription: "RAID Mode",
			Description:         "RAID Mode",
			Computed:            true,
		},
		"real_time_capability": schema.StringAttribute{
			MarkdownDescription: "Realtime Capability",
			Description:         "Realtime Capability",
			Computed:            true,
		},
		"reconstruct_rate_percent": schema.Int64Attribute{
			MarkdownDescription: "Reconstruct Rate Percent",
			Description:         "Reconstruct Rate Percent",
			Computed:            true,
		},
		"rollup_status": schema.StringAttribute{
			MarkdownDescription: "Rollup Status",
			Description:         "Rollup Status",
			Computed:            true,
		},
		"sas_address": schema.StringAttribute{
			MarkdownDescription: "SAS Address",
			Description:         "SAS Address",
			Computed:            true,
		},
		"security_status": schema.StringAttribute{
			MarkdownDescription: "Security Status",
			Description:         "Security Status",
			Computed:            true,
		},
		"shared_slot_assignment_allowed": schema.StringAttribute{
			MarkdownDescription: "Shared Slot Assignment Allowed",
			Description:         "Shared Slot Assignment Allowed",
			Computed:            true,
		},
		"sliced_vd_capability": schema.StringAttribute{
			MarkdownDescription: "Sliced VD Capability",
			Description:         "Sliced VD Capability",
			Computed:            true,
		},
		"spindown_idle_time_seconds": schema.Int64Attribute{
			MarkdownDescription: "Spindown Idle Time Seconds",
			Description:         "Spindown Idle Time Seconds",
			Computed:            true,
		},
		"support_controller_boot_mode": schema.StringAttribute{
			MarkdownDescription: "Support Controller Boot Mode",
			Description:         "Support Controller Boot Mode",
			Computed:            true,
		},
		"support_enhanced_auto_foreign_import": schema.StringAttribute{
			MarkdownDescription: "Support Enhanced Auto Foreign Import",
			Description:         "Support Enhanced Auto Foreign Import",
			Computed:            true,
		},
		"support_raid10_uneven_spans": schema.StringAttribute{
			MarkdownDescription: "Support RAID10 Uneven Spans",
			Description:         "Support RAID10 Uneven Spans",
			Computed:            true,
		},
		"supported_initialization_types": schema.ListAttribute{
			MarkdownDescription: "Supported Initialization Types",
			Description:         "Supported Initialization Types",
			Computed:            true,
			ElementType:         types.StringType,
		},
		"supports_lkm_to_sekm_transition": schema.StringAttribute{
			MarkdownDescription: "Supports LKM to SEKM Transition",
			Description:         "Supports LKM to SEKM Transition",
			Computed:            true,
		},
		"t10_pi_capability": schema.StringAttribute{
			MarkdownDescription: "T10 PI Capability",
			Description:         "T10 PI Capability",
			Computed:            true,
		},
	}
}
