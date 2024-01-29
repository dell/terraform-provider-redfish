package provider

import (
	"context"
	"fmt"
	"terraform-provider-redfish/gofish/dell"
	"terraform-provider-redfish/redfish/models"
	"time"

	"github.com/stmcginnis/gofish"
	"github.com/stmcginnis/gofish/redfish"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ datasource.DataSource              = &StorageDatasource{}
	_ datasource.DataSourceWithConfigure = &StorageDatasource{}
)

// NewStorageDatasource is new datasource for storage
func NewStorageDatasource() datasource.DataSource {
	return &StorageDatasource{}
}

// StorageDatasource to construct datasource
type StorageDatasource struct {
	p       *redfishProvider
	ctx     context.Context
	service *gofish.Service
}

// Configure implements datasource.DataSourceWithConfigure
func (g *StorageDatasource) Configure(_ context.Context, req datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	g.p = req.ProviderData.(*redfishProvider)
}

// Metadata implements datasource.DataSource
func (*StorageDatasource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "storage"
}

// Schema implements datasource.DataSource
func (*StorageDatasource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "This Terraform datasource is used to query existing storage volume details." +
			" The information fetched from this block can be further used for resource block.",
		Description: "This Terraform datasource is used to query existing storage volume details." +
			" The information fetched from this block can be further used for resource block.",
		Attributes: StorageDatasourceSchema(),
		Blocks:     RedfishServerDatasourceBlockMap(),
	}
}

// StorageDatasourceSchema to define the storage data-source schema
func StorageDatasourceSchema() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"id": schema.StringAttribute{
			MarkdownDescription: "ID of the storage data-source",
			Description:         "ID of the storage data-source",
			Computed:            true,
		},
		"storage": schema.ListNestedAttribute{
			MarkdownDescription: "List of storage controllers",
			Description:         "List of storage controllers",
			NestedObject: schema.NestedAttributeObject{
				Attributes: StorageSchema(),
			},
			Computed: true,
		},
	}
}

// // StorageControllerSchema to define the storage data-source schema
// func StorageControllerSchema() map[string]schema.Attribute {
// 	return map[string]schema.Attribute{
// 		"storage_controller_id": schema.StringAttribute{
// 			MarkdownDescription: "ID of the storage controller",
// 			Description:         "ID of the storage controller",
// 			Computed:            true,
// 		},
// 		"drives": schema.ListAttribute{
// 			MarkdownDescription: "List of drives on the storage controller",
// 			Description:         "List of drives on the storage controller",
// 			Computed:            true,
// 			ElementType:         types.StringType,
// 		},
// 		"storage_controllers": schema.MapAttribute{
// 			Computed:    true,
// 			ElementType: types.StringType,
// 		},
// 		"dell_data": schema.StringAttribute{
// 			MarkdownDescription: "ID of the storage controller",
// 			Description:         "ID of the storage controller",
// 			Computed:            true,
// 		},
// 	}
// }

// Read implements datasource.DataSource
func (g *StorageDatasource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var plan models.StorageDatasource
	diags := req.Config.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	service, err := NewConfig(g.p, &plan.RedfishServer)
	if err != nil {
		resp.Diagnostics.AddError("service error", err.Error())
		return
	}
	g.ctx = ctx
	g.service = service
	state, diags := g.readDatasourceRedfishStorage(plan)
	resp.Diagnostics.Append(diags...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (g *StorageDatasource) readDatasourceRedfishStorage(d models.StorageDatasource) (models.StorageDatasource, diag.Diagnostics) {
	var diags diag.Diagnostics
	// write the current time as ID
	d.ID = types.StringValue(fmt.Sprintf("%d", time.Now().Unix()))

	systems, err := g.service.Systems()
	if err != nil {
		diags.AddError("Error fetching computer systems collection", err.Error())
		return d, diags
	}

	storage, err := systems[0].Storage()
	if err != nil {
		diags.AddError("Error fetching storage", err.Error())
		return d, diags
	}

	d.Storages = make([]models.Storage, 0)
	for _, s := range storage {
		dellStorage, _ := dell.Storage(s)
		terraformData := newStorage(*dellStorage)
		d.Storages = append(d.Storages, terraformData)
		// mToAdd := models.StorageControllerData{
		// 	ID: types.StringValue(s.ID),
		// }
		// drives, err := s.Drives()
		// if err != nil {
		// 	diags.AddError(fmt.Sprintf("Error when retrieving drives: %s", s.ID), err.Error())
		// 	continue
		// }

		// driveNames := make([]types.String, 0)
		// for _, d := range drives {
		// 	driveNames = append(driveNames, types.StringValue(d.Name))
		// }
		// mToAdd.Drives = driveNames
		// d.Storages = append(d.Storages, mToAdd)
	}

	return d, diags
}

func newStorage(input dell.StorageExtended) models.Storage {
	drives,_ := input.Drives()
	return models.Storage{
		OdataContext:                        types.StringValue(input.ODataContext),
		OdataID:                             types.StringValue(input.ODataID),
		OdataType:                           types.StringValue(input.ODataType),
		// Controllers:                         newControllers(input.),
		Description:                         types.StringValue(input.Description),
		Drives:                              newDrivesList(drives),
		DrivesOdataCount:                    types.Int64Value(int64(input.DrivesCount)),
		ID:                                  types.StringValue(input.ID),
		// Identifiers:                         newIdentifiersList(input.I),
		// IdentifiersOdataCount:               types.Int64Value(int64(input.IdentifiersOdataCount)),
		// Links:                               newLinks(input.Links),
		Name:                                types.StringValue(input.Name),
		Oem:                                 newOem(input.OemData),
		// Status:                              newStatus(input.Status),
		// StorageControllers:                  newStorageControllersList(input.StorageControllers),
		// StorageControllersRedfishDeprecated: types.StringValue(input.StorageControllersRedfishDeprecated),
		// StorageControllersOdataCount:        types.Int64Value(int64(input.StorageControllersOdataCount)),
		// Volumes:                             newVolumes(input.Volumes),
	}
}

// newDrivesList converts list of redfish.Drives to list of models.Drives
func newDrivesList(inputs []*redfish.Drive) []models.Drives {
	out := make([]models.Drives, 0)
	for _, input := range inputs {
		out = append(out, newDrives(input))
	}
	return out
}

// // newIdentifiersList converts list of redfish.Identifiers to list of models.Identifiers
// func newIdentifiersList(inputs []redfish.Ide) []models.Identifiers {
// 	out := make([]models.Identifiers, 0)
// 	for _, input := range inputs {
// 		out = append(out, newIdentifiers(input))
// 	}
// 	return out
// }

// newStorageControllersList converts list of redfish.StorageControllers to list of models.StorageControllers
func newStorageControllersList(inputs []redfish.StorageController) []models.StorageControllers {
	out := make([]models.StorageControllers, 0)
	for _, input := range inputs {
		out = append(out, newStorageControllers(input))
	}
	return out
}

// // newControllers converts redfish.Controllers to models.Controllers
// func newControllers(input redfish.Controllers) models.Controllers {
// 	return models.Controllers{
// 		OdataID: types.StringValue(input.Odata),
// 	}
// }

// newDrives converts redfish.Drives to models.Drives
func newDrives(input *redfish.Drive) models.Drives {
	return models.Drives{
		OdataID: types.StringValue(input.ODataType),
	}
}

// // newIdentifiers converts redfish.Identifiers to models.Identifiers
// func newIdentifiers(input redfish.Iden) models.Identifiers {
// 	return models.Identifiers{
// 		DurableName:       types.StringValue(input.DurableName),
// 		DurableNameFormat: types.StringValue(input.DurableNameFormat),
// 	}
// }

// // newEnclosures converts redfish.Enclosures to models.Enclosures
// func newEnclosures(input redfish.Enclosures) models.Enclosures {
// 	return models.Enclosures{
// 		OdataID: types.StringValue(input.OdataID),
// 	}
// }

// // newSimpleStorage converts redfish.SimpleStorage to models.SimpleStorage
// func newSimpleStorage(input redfish.SimpleStorage) models.SimpleStorage {
// 	return models.SimpleStorage{
// 		OdataID: types.StringValue(input.OdataID),
// 	}
// }

// newLinks converts redfish.Links to models.Links
// func newLinks(input redfish.Links) models.Links {
// 	return models.Links{
// 		Enclosures:           newEnclosuresList(input.Enclosures),
// 		EnclosuresOdataCount: types.Int64Value(int64(input.EnclosuresOdataCount)),
// 		SimpleStorage:        newSimpleStorage(input.SimpleStorage),
// 	}
// }

// // newEnclosuresList converts list of redfish.Enclosures to list of models.Enclosures
// func newEnclosuresList(inputs []redfish.Enclosures) []models.Enclosures {
// 	out := make([]models.Enclosures, 0)
// 	for _, input := range inputs {
// 		out = append(out, newEnclosures(input))
// 	}
// 	return out
// }

// newDellController converts redfish.DellController to models.DellController
func newDellController(input dell.DellController) models.DellController {
	return models.DellController{
		OdataContext:                     types.StringValue(input.OdataContext),
		OdataID:                          types.StringValue(input.OdataID),
		OdataType:                        types.StringValue(input.OdataType),
		AlarmState:                       types.StringValue(input.AlarmState),
		AutoConfigBehavior:               types.StringValue(input.AutoConfigBehavior),
		BootVirtualDiskFQDD:              types.StringValue(input.BootVirtualDiskFQDD),
		CacheSizeInMB:                    types.Int64Value(int64(input.CacheSizeInMB)),
		CachecadeCapability:              types.StringValue(input.CachecadeCapability),
		ConnectorCount:                   types.Int64Value(int64(input.ConnectorCount)),
		ControllerFirmwareVersion:        types.StringValue(input.ControllerFirmwareVersion),
		CurrentControllerMode:            types.StringValue(input.CurrentControllerMode),
		Description:                      types.StringValue(input.Description),
		Device:                           types.StringValue(input.Device),
		DeviceCardDataBusWidth:           types.StringValue(input.DeviceCardDataBusWidth),
		DeviceCardSlotLength:             types.StringValue(input.DeviceCardSlotLength),
		DeviceCardSlotType:               types.StringValue(input.DeviceCardSlotType),
		DriverVersion:                    types.StringValue(input.DriverVersion),
		EncryptionCapability:             types.StringValue(input.EncryptionCapability),
		EncryptionMode:                   types.StringValue(input.EncryptionMode),
		ID:                               types.StringValue(input.ID),
		LastSystemInventoryTime:          types.StringValue(input.LastSystemInventoryTime),
		LastUpdateTime:                   types.StringValue(input.LastUpdateTime),
		MaxAvailablePCILinkSpeed:         types.StringValue(input.MaxAvailablePCILinkSpeed),
		MaxPossiblePCILinkSpeed:          types.StringValue(input.MaxPossiblePCILinkSpeed),
		Name:                             types.StringValue(input.Name),
		PatrolReadState:                  types.StringValue(input.PatrolReadState),
		PersistentHotspare:               types.StringValue(input.PersistentHotspare),
		RealtimeCapability:               types.StringValue(input.RealtimeCapability),
		RollupStatus:                     types.StringValue(input.RollupStatus),
		SASAddress:                       types.StringValue(input.SASAddress),
		SecurityStatus:                   types.StringValue(input.SecurityStatus),
		SharedSlotAssignmentAllowed:      types.StringValue(input.SharedSlotAssignmentAllowed),
		SlicedVDCapability:               types.StringValue(input.SlicedVDCapability),
		SupportControllerBootMode:        types.StringValue(input.SupportControllerBootMode),
		SupportEnhancedAutoForeignImport: types.StringValue(input.SupportEnhancedAutoForeignImport),
		SupportRAID10UnevenSpans:         types.StringValue(input.SupportRAID10UnevenSpans),
		SupportsLKMtoSEKMTransition:      types.StringValue(input.SupportsLKMtoSEKMTransition),
		T10PICapability:                  types.StringValue(input.T10PICapability),
	}
}

// newDellControllerBattery converts redfish.DellControllerBattery to models.DellControllerBattery
func newDellControllerBattery(input dell.DellControllerBattery) models.DellControllerBattery {
	return models.DellControllerBattery{
		OdataContext:  types.StringValue(input.OdataContext),
		OdataID:       types.StringValue(input.OdataID),
		OdataType:     types.StringValue(input.OdataType),
		Description:   types.StringValue(input.Description),
		Fqdd:          types.StringValue(input.Fqdd),
		ID:            types.StringValue(input.ID),
		Name:          types.StringValue(input.Name),
		PrimaryStatus: types.StringValue(input.PrimaryStatus),
		RAIDState:     types.StringValue(input.RAIDState),
	}
}

// newDell converts redfish.Dell to models.Dell
func newDell(input dell.StorageOEM) models.Dell {
	return models.Dell{
		OdataType:             types.StringValue(input.OdataType),
		DellController:        newDellController(input.DellController),
		DellControllerBattery: newDellControllerBattery(input.DellControllerBattery),
	}
}

// newOem converts redfish.Oem to models.Oem
func newOem(input dell.StorageOEM) models.Oem {
	return models.Oem{
		Dell: newDell(input),
	}
}

// newStatus converts redfish.Status to models.Status
// func newStatus(input redfish.Status) models.Status {
// 	return models.Status{
// 		Health:       types.StringValue(input.Health),
// 		HealthRollup: types.StringValue(input.HealthRollup),
// 		State:        types.StringValue(input.State),
// 	}
// }

// newAssembly converts redfish.Assembly to models.Assembly
func newAssembly(input redfish.Assembly) models.Assembly {
	return models.Assembly{
		OdataID: types.StringValue(input.ODataID),
	}
}

// newCacheSummary converts redfish.CacheSummary to models.CacheSummary
func newCacheSummary(input redfish.CacheSummary) models.CacheSummary {
	return models.CacheSummary{
		TotalCacheSizeMiB: types.Int64Value(int64(input.TotalCacheSizeMiB)),
	}
}

// // newControllerRates converts redfish.ControllerRates to models.ControllerRates
// func newControllerRates(input redfish.Control) models.ControllerRates {
// 	return models.ControllerRates{
// 		ConsistencyCheckRatePercent: types.Int64Value(int64(input.ConsistencyCheckRatePercent)),
// 		RebuildRatePercent:          types.Int64Value(int64(input.RebuildRatePercent)),
// 	}
// }

// // newPCIeFunctions converts redfish.PCIeFunctions to models.PCIeFunctions
// func newPCIeFunctions(input redfish.PCIeFunction) models.PCIeFunctions {
// 	return models.PCIeFunctions{
// 		OdataID: types.StringValue(input.OdataID),
// 	}
// }

// // newLinks converts redfish.Links to models.Links
// func newLinks(input redfish.LinkStatus) models.Links {
// 	return models.Links{
// 		PCIeFunctions:           newPCIeFunctionsList(input.PCIeFunctions),
// 		PCIeFunctionsOdataCount: types.Int64Value(int64(input.PCIeFunctionsOdataCount)),
// 	}
// }

// // newPCIeFunctionsList converts list of redfish.PCIeFunctions to list of models.PCIeFunctions
// func newPCIeFunctionsList(inputs []redfish.PCIeFunction) []models.PCIeFunctions {
// 	out := make([]models.PCIeFunctions, 0)
// 	for _, input := range inputs {
// 		out = append(out, newPCIeFunctions(input))
// 	}
// 	return out
// }

// newStorageControllers converts redfish.StorageControllers to models.StorageControllers
func newStorageControllers(input redfish.StorageController) models.StorageControllers {
	assembly, _ := input.Assembly()
	return models.StorageControllers{
		OdataID:               types.StringValue(input.ODataID),
		Assembly:              newAssembly(*assembly),
		CacheSummary:          newCacheSummary(input.CacheSummary),
		// ControllerRates:       newControllerRates(input.Co),
		FirmwareVersion:       types.StringValue(input.FirmwareVersion),
		// Identifiers:           newIdentifiersList(input.Identifiers),
		// IdentifiersOdataCount: types.Int64Value(int64(input.IdentifiersOdataCount)),
		// Links:                 newLinks(input.Links),
		Manufacturer:          types.StringValue(input.Manufacturer),
		MemberID:              types.StringValue(input.MemberID),
		Model:                 types.StringValue(input.Model),
		Name:                  types.StringValue(input.Name),
		SpeedGbps:             types.Int64Value(int64(input.SpeedGbps)),
		// Status:                newStatus(input.Status),
		// SupportedControllerProtocols: newtypes.StringList(input.SupportedControllerProtocols),
		// SupportedControllerProtocolsOdataCount: types.Int64Value(int64(input.SupportedControllerProtocolsOdataCount)),
		// SupportedDeviceProtocols: newtypes.StringList(input.SupportedDeviceProtocols),
		// SupportedDeviceProtocolsOdataCount: types.Int64Value(int64(input.SupportedDeviceProtocolsOdataCount)),
		// SupportedRAIDTypes: newtypes.StringList(input.SupportedRAIDTypes),
		// SupportedRAIDTypesOdataCount: types.Int64Value(int64(input.SupportedRAIDTypesOdataCount)),
	}
}

// newVolumes converts redfish.Volumes to models.Volumes
func newVolumes(input redfish.Volume) models.Volumes {
	return models.Volumes{
		OdataID: types.StringValue(input.ODataID),
	}
}

// StorageSchema is a function that returns the schema for Storage
func StorageSchema() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"odata_context": schema.StringAttribute{
			MarkdownDescription: "",
			Description:         "",
			Computed:            true,
		},
		"odata_id": schema.StringAttribute{
			MarkdownDescription: "",
			Description:         "",
			Computed:            true,
		},
		"odata_type": schema.StringAttribute{
			MarkdownDescription: "",
			Description:         "",
			Computed:            true,
		},
		"controllers": schema.SingleNestedAttribute{
			MarkdownDescription: "",
			Description:         "",
			Computed:            true,
			Attributes:          ControllersSchema(),
		},
		"description": schema.StringAttribute{
			MarkdownDescription: "",
			Description:         "",
			Computed:            true,
		},
		"drives": schema.ListNestedAttribute{
			MarkdownDescription: "",
			Description:         "",
			Computed:            true,
			NestedObject:        schema.NestedAttributeObject{Attributes: DrivesSchema()},
		},
		"drives_odata_count": schema.Int64Attribute{
			MarkdownDescription: "",
			Description:         "",
			Computed:            true,
		},
		"id": schema.StringAttribute{
			MarkdownDescription: "",
			Description:         "",
			Computed:            true,
		},
		"identifiers": schema.ListNestedAttribute{
			MarkdownDescription: "",
			Description:         "",
			Computed:            true,
			NestedObject:        schema.NestedAttributeObject{Attributes: IdentifiersSchema()},
		},
		"identifiers_odata_count": schema.Int64Attribute{
			MarkdownDescription: "",
			Description:         "",
			Computed:            true,
		},
		"links": schema.SingleNestedAttribute{
			MarkdownDescription: "",
			Description:         "",
			Computed:            true,
			Attributes:          LinksStorageSchema(),
		},
		"name": schema.StringAttribute{
			MarkdownDescription: "",
			Description:         "",
			Computed:            true,
		},
		"oem": schema.SingleNestedAttribute{
			MarkdownDescription: "",
			Description:         "",
			Computed:            true,
			Attributes:          OemSchema(),
		},
		"status": schema.SingleNestedAttribute{
			MarkdownDescription: "",
			Description:         "",
			Computed:            true,
			Attributes:          StatusSchema(),
		},
		"storage_controllers": schema.ListNestedAttribute{
			MarkdownDescription: "",
			Description:         "",
			Computed:            true,
			NestedObject:        schema.NestedAttributeObject{Attributes: StorageControllersSchema()},
		},
		"storage_controllers_redfish_deprecated": schema.StringAttribute{
			MarkdownDescription: "",
			Description:         "",
			Computed:            true,
		},
		"storage_controllers_odata_count": schema.Int64Attribute{
			MarkdownDescription: "",
			Description:         "",
			Computed:            true,
		},
		"volumes": schema.SingleNestedAttribute{
			MarkdownDescription: "",
			Description:         "",
			Computed:            true,
			Attributes:          VolumesSchema(),
		},
	}
}

// ControllersSchema is a function that returns the schema for Controllers
func ControllersSchema() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"odata_id": schema.StringAttribute{
			MarkdownDescription: "",
			Description:         "",
			Computed:            true,
		},
	}
}

// DrivesSchema is a function that returns the schema for Drives
func DrivesSchema() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"odata_id": schema.StringAttribute{
			MarkdownDescription: "",
			Description:         "",
			Computed:            true,
		},
	}
}

// IdentifiersSchema is a function that returns the schema for Identifiers
func IdentifiersSchema() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"durable_name": schema.StringAttribute{
			MarkdownDescription: "",
			Description:         "",
			Computed:            true,
		},
		"durable_name_format": schema.StringAttribute{
			MarkdownDescription: "",
			Description:         "",
			Computed:            true,
		},
	}
}

// EnclosuresSchema is a function that returns the schema for Enclosures
func EnclosuresSchema() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"odata_id": schema.StringAttribute{
			MarkdownDescription: "",
			Description:         "",
			Computed:            true,
		},
	}
}

// SimpleStorageSchema is a function that returns the schema for SimpleStorage
func SimpleStorageSchema() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"odata_id": schema.StringAttribute{
			MarkdownDescription: "",
			Description:         "",
			Computed:            true,
		},
	}
}

// LinksSchema is a function that returns the schema for Links
func LinksStorageSchema() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"enclosures": schema.ListNestedAttribute{
			MarkdownDescription: "",
			Description:         "",
			Computed:            true,
			NestedObject:        schema.NestedAttributeObject{Attributes: EnclosuresSchema()},
		},
		"enclosures_odata_count": schema.Int64Attribute{
			MarkdownDescription: "",
			Description:         "",
			Computed:            true,
		},
		"simple_storage": schema.SingleNestedAttribute{
			MarkdownDescription: "",
			Description:         "",
			Computed:            true,
			Attributes:          SimpleStorageSchema(),
		},
	}
}

// DellControllerSchema is a function that returns the schema for DellController
func DellControllerSchema() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"odata_context": schema.StringAttribute{
			MarkdownDescription: "",
			Description:         "",
			Computed:            true,
		},
		"odata_id": schema.StringAttribute{
			MarkdownDescription: "",
			Description:         "",
			Computed:            true,
		},
		"odata_type": schema.StringAttribute{
			MarkdownDescription: "",
			Description:         "",
			Computed:            true,
		},
		"alarm_state": schema.StringAttribute{
			MarkdownDescription: "",
			Description:         "",
			Computed:            true,
		},
		"auto_config_behavior": schema.StringAttribute{
			MarkdownDescription: "",
			Description:         "",
			Computed:            true,
		},
		"boot_virtual_disk_fqdd": schema.StringAttribute{
			MarkdownDescription: "",
			Description:         "",
			Computed:            true,
		},
		"cache_size_in_mb": schema.Int64Attribute{
			MarkdownDescription: "",
			Description:         "",
			Computed:            true,
		},
		"cachecade_capability": schema.StringAttribute{
			MarkdownDescription: "",
			Description:         "",
			Computed:            true,
		},
		"connector_count": schema.Int64Attribute{
			MarkdownDescription: "",
			Description:         "",
			Computed:            true,
		},
		"controller_firmware_version": schema.StringAttribute{
			MarkdownDescription: "",
			Description:         "",
			Computed:            true,
		},
		"current_controller_mode": schema.StringAttribute{
			MarkdownDescription: "",
			Description:         "",
			Computed:            true,
		},
		"description": schema.StringAttribute{
			MarkdownDescription: "",
			Description:         "",
			Computed:            true,
		},
		"device": schema.StringAttribute{
			MarkdownDescription: "",
			Description:         "",
			Computed:            true,
		},
		"device_card_data_bus_width": schema.StringAttribute{
			MarkdownDescription: "",
			Description:         "",
			Computed:            true,
		},
		"device_card_slot_length": schema.StringAttribute{
			MarkdownDescription: "",
			Description:         "",
			Computed:            true,
		},
		"device_card_slot_type": schema.StringAttribute{
			MarkdownDescription: "",
			Description:         "",
			Computed:            true,
		},
		"driver_version": schema.StringAttribute{
			MarkdownDescription: "",
			Description:         "",
			Computed:            true,
		},
		"encryption_capability": schema.StringAttribute{
			MarkdownDescription: "",
			Description:         "",
			Computed:            true,
		},
		"encryption_mode": schema.StringAttribute{
			MarkdownDescription: "",
			Description:         "",
			Computed:            true,
		},
		"id": schema.StringAttribute{
			MarkdownDescription: "",
			Description:         "",
			Computed:            true,
		},
		"key_id": schema.StringAttribute{
			MarkdownDescription: "",
			Description:         "",
			Computed:            true,
		},
		"last_system_inventory_time": schema.StringAttribute{
			MarkdownDescription: "",
			Description:         "",
			Computed:            true,
		},
		"last_update_time": schema.StringAttribute{
			MarkdownDescription: "",
			Description:         "",
			Computed:            true,
		},
		"max_available_pci_link_speed": schema.StringAttribute{
			MarkdownDescription: "",
			Description:         "",
			Computed:            true,
		},
		"max_possible_pci_link_speed": schema.StringAttribute{
			MarkdownDescription: "",
			Description:         "",
			Computed:            true,
		},
		"name": schema.StringAttribute{
			MarkdownDescription: "",
			Description:         "",
			Computed:            true,
		},
		"pci_slot": schema.StringAttribute{
			MarkdownDescription: "",
			Description:         "",
			Computed:            true,
		},
		"patrol_read_state": schema.StringAttribute{
			MarkdownDescription: "",
			Description:         "",
			Computed:            true,
		},
		"persistent_hotspare": schema.StringAttribute{
			MarkdownDescription: "",
			Description:         "",
			Computed:            true,
		},
		"realtime_capability": schema.StringAttribute{
			MarkdownDescription: "",
			Description:         "",
			Computed:            true,
		},
		"rollup_status": schema.StringAttribute{
			MarkdownDescription: "",
			Description:         "",
			Computed:            true,
		},
		"sas_address": schema.StringAttribute{
			MarkdownDescription: "",
			Description:         "",
			Computed:            true,
		},
		"security_status": schema.StringAttribute{
			MarkdownDescription: "",
			Description:         "",
			Computed:            true,
		},
		"shared_slot_assignment_allowed": schema.StringAttribute{
			MarkdownDescription: "",
			Description:         "",
			Computed:            true,
		},
		"sliced_vd_capability": schema.StringAttribute{
			MarkdownDescription: "",
			Description:         "",
			Computed:            true,
		},
		"support_controller_boot_mode": schema.StringAttribute{
			MarkdownDescription: "",
			Description:         "",
			Computed:            true,
		},
		"support_enhanced_auto_foreign_import": schema.StringAttribute{
			MarkdownDescription: "",
			Description:         "",
			Computed:            true,
		},
		"support_raid_10_uneven_spans": schema.StringAttribute{
			MarkdownDescription: "",
			Description:         "",
			Computed:            true,
		},
		"supports_lk_mto_sekm_transition": schema.StringAttribute{
			MarkdownDescription: "",
			Description:         "",
			Computed:            true,
		},
		"t_10_pi_capability": schema.StringAttribute{
			MarkdownDescription: "",
			Description:         "",
			Computed:            true,
		},
	}
}

// DellControllerBatterySchema is a function that returns the schema for DellControllerBattery
func DellControllerBatterySchema() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"odata_context": schema.StringAttribute{
			MarkdownDescription: "",
			Description:         "",
			Computed:            true,
		},
		"odata_id": schema.StringAttribute{
			MarkdownDescription: "",
			Description:         "",
			Computed:            true,
		},
		"odata_type": schema.StringAttribute{
			MarkdownDescription: "",
			Description:         "",
			Computed:            true,
		},
		"description": schema.StringAttribute{
			MarkdownDescription: "",
			Description:         "",
			Computed:            true,
		},
		"fqdd": schema.StringAttribute{
			MarkdownDescription: "",
			Description:         "",
			Computed:            true,
		},
		"id": schema.StringAttribute{
			MarkdownDescription: "",
			Description:         "",
			Computed:            true,
		},
		"name": schema.StringAttribute{
			MarkdownDescription: "",
			Description:         "",
			Computed:            true,
		},
		"primary_status": schema.StringAttribute{
			MarkdownDescription: "",
			Description:         "",
			Computed:            true,
		},
		"raid_state": schema.StringAttribute{
			MarkdownDescription: "",
			Description:         "",
			Computed:            true,
		},
	}
}

// DellSchema is a function that returns the schema for Dell
func DellSchema() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"odata_type": schema.StringAttribute{
			MarkdownDescription: "",
			Description:         "",
			Computed:            true,
		},
		"dell_controller": schema.SingleNestedAttribute{
			MarkdownDescription: "",
			Description:         "",
			Computed:            true,
			Attributes:          DellControllerSchema(),
		},
		"dell_controller_battery": schema.SingleNestedAttribute{
			MarkdownDescription: "",
			Description:         "",
			Computed:            true,
			Attributes:          DellControllerBatterySchema(),
		},
	}
}

// OemSchema is a function that returns the schema for Oem
func OemSchema() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"dell": schema.SingleNestedAttribute{
			MarkdownDescription: "",
			Description:         "",
			Computed:            true,
			Attributes:          DellSchema(),
		},
	}
}

// StatusSchema is a function that returns the schema for Status
func StatusSchema() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"health": schema.StringAttribute{
			MarkdownDescription: "",
			Description:         "",
			Computed:            true,
		},
		"health_rollup": schema.StringAttribute{
			MarkdownDescription: "",
			Description:         "",
			Computed:            true,
		},
		"state": schema.StringAttribute{
			MarkdownDescription: "",
			Description:         "",
			Computed:            true,
		},
	}
}

// AssemblySchema is a function that returns the schema for Assembly
func AssemblySchema() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"odata_id": schema.StringAttribute{
			MarkdownDescription: "",
			Description:         "",
			Computed:            true,
		},
	}
}

// CacheSummarySchema is a function that returns the schema for CacheSummary
func CacheSummarySchema() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"total_cache_size_mi_b": schema.Int64Attribute{
			MarkdownDescription: "",
			Description:         "",
			Computed:            true,
		},
	}
}

// ControllerRatesSchema is a function that returns the schema for ControllerRates
func ControllerRatesSchema() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"consistency_check_rate_percent": schema.Int64Attribute{
			MarkdownDescription: "",
			Description:         "",
			Computed:            true,
		},
		"rebuild_rate_percent": schema.Int64Attribute{
			MarkdownDescription: "",
			Description:         "",
			Computed:            true,
		},
	}
}

// PCIeFunctionsSchema is a function that returns the schema for PCIeFunctions
func PCIeFunctionsSchema() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"odata_id": schema.StringAttribute{
			MarkdownDescription: "",
			Description:         "",
			Computed:            true,
		},
	}
}

// LinksSchema is a function that returns the schema for Links
func LinksSchema() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"pc_ie_functions": schema.ListNestedAttribute{
			MarkdownDescription: "",
			Description:         "",
			Computed:            true,
			NestedObject:        schema.NestedAttributeObject{Attributes: PCIeFunctionsSchema()},
		},
		"pc_ie_functions_odata_count": schema.Int64Attribute{
			MarkdownDescription: "",
			Description:         "",
			Computed:            true,
		},
	}
}

// StorageControllersSchema is a function that returns the schema for StorageControllers
func StorageControllersSchema() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"odata_id": schema.StringAttribute{
			MarkdownDescription: "",
			Description:         "",
			Computed:            true,
		},
		"assembly": schema.SingleNestedAttribute{
			MarkdownDescription: "",
			Description:         "",
			Computed:            true,
			Attributes:          AssemblySchema(),
		},
		"cache_summary": schema.SingleNestedAttribute{
			MarkdownDescription: "",
			Description:         "",
			Computed:            true,
			Attributes:          CacheSummarySchema(),
		},
		"controller_rates": schema.SingleNestedAttribute{
			MarkdownDescription: "",
			Description:         "",
			Computed:            true,
			Attributes:          ControllerRatesSchema(),
		},
		"firmware_version": schema.StringAttribute{
			MarkdownDescription: "",
			Description:         "",
			Computed:            true,
		},
		"identifiers": schema.ListNestedAttribute{
			MarkdownDescription: "",
			Description:         "",
			Computed:            true,
			NestedObject:        schema.NestedAttributeObject{Attributes: IdentifiersSchema()},
		},
		"identifiers_odata_count": schema.Int64Attribute{
			MarkdownDescription: "",
			Description:         "",
			Computed:            true,
		},
		"links": schema.SingleNestedAttribute{
			MarkdownDescription: "",
			Description:         "",
			Computed:            true,
			Attributes:          LinksSchema(),
		},
		"manufacturer": schema.StringAttribute{
			MarkdownDescription: "",
			Description:         "",
			Computed:            true,
		},
		"member_id": schema.StringAttribute{
			MarkdownDescription: "",
			Description:         "",
			Computed:            true,
		},
		"model": schema.StringAttribute{
			MarkdownDescription: "",
			Description:         "",
			Computed:            true,
		},
		"name": schema.StringAttribute{
			MarkdownDescription: "",
			Description:         "",
			Computed:            true,
		},
		"speed_gbps": schema.Int64Attribute{
			MarkdownDescription: "",
			Description:         "",
			Computed:            true,
		},
		"status": schema.SingleNestedAttribute{
			MarkdownDescription: "",
			Description:         "",
			Computed:            true,
			Attributes:          StatusSchema(),
		},
		"supported_controller_protocols": schema.ListAttribute{
			MarkdownDescription: "",
			Description:         "",
			Computed:            true,
			ElementType:         types.StringType,
		},
		"supported_controller_protocols_odata_count": schema.Int64Attribute{
			MarkdownDescription: "",
			Description:         "",
			Computed:            true,
		},
		"supported_device_protocols": schema.ListAttribute{
			MarkdownDescription: "",
			Description:         "",
			Computed:            true,
			ElementType:         types.StringType,
		},
		"supported_device_protocols_odata_count": schema.Int64Attribute{
			MarkdownDescription: "",
			Description:         "",
			Computed:            true,
		},
		"supported_raid_types": schema.ListAttribute{
			MarkdownDescription: "",
			Description:         "",
			Computed:            true,
			ElementType:         types.StringType,
		},
		"supported_raid_types_odata_count": schema.Int64Attribute{
			MarkdownDescription: "",
			Description:         "",
			Computed:            true,
		},
	}
}

// VolumesSchema is a function that returns the schema for Volumes
func VolumesSchema() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"odata_id": schema.StringAttribute{
			MarkdownDescription: "",
			Description:         "",
			Computed:            true,
		},
	}
}
