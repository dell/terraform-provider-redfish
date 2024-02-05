package provider

import (
	"context"
	"fmt"
	"terraform-provider-redfish/gofish/dell"
	"terraform-provider-redfish/redfish/models"
	"time"

	"github.com/stmcginnis/gofish"
	"github.com/stmcginnis/gofish/common"
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
		"controller_ids": schema.ListAttribute{
			MarkdownDescription: "ID of the storage controller ID",
			Description:         "ID of the storage controller ID",
			Optional:            true,
			ElementType:         types.StringType,
		},
		"controller_names": schema.ListAttribute{
			MarkdownDescription: "ID of the storage controller name",
			Description:         "ID of the storage controller name",
			Optional:            true,
			ElementType:         types.StringType,
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
	controllerIDs := make([]string, 0)
	controllerNames := make([]string, 0)
	d.ControllerIDs.ElementsAs(g.ctx, &controllerIDs, false)
	d.ControllerNames.ElementsAs(g.ctx, &controllerNames, false)
	controllers := append(controllerIDs, controllerNames...)
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
		if len(controllers) > 0 {
			if !contains(controllers, s.Name, s.ID) {
				continue
			}
		}
		dellStorage, _ := dell.Storage(s)
		terraformData := newStorage(*dellStorage)
		drives, err := s.Drives()
		if err != nil {
			diags.AddError(fmt.Sprintf("Error when retrieving drives: %s", s.ID), err.Error())
			continue
		}

		driveNames := make([]types.String, 0)
		for _, d := range drives {
			driveNames = append(driveNames, types.StringValue(d.Name))
		}
		terraformData.Drives = driveNames
		d.Storages = append(d.Storages, terraformData)
	}

	return d, diags
}

func contains(s []string, str1 string, str2 string) bool {
	for _, v := range s {
		if v == str1 || v == str2 {
			return true
		}
	}

	return false
}

func newStorage(extendedStorage dell.StorageExtended) models.Storage {
	input := extendedStorage.Storage
	return models.Storage{
		OdataContext:                 types.StringValue(input.ODataContext),
		OdataID:                      types.StringValue(input.ODataID),
		OdataType:                    types.StringValue(input.ODataType),
		Description:                  types.StringValue(input.Description),
		DrivesOdataCount:             types.Int64Value(int64(input.DrivesCount)),
		ID:                           types.StringValue(input.ID),
		Name:                         types.StringValue(input.Name),
		Oem:                          newOem(extendedStorage.OemData),
		Status:                       newStatus(input.Status),
		StorageControllers:           newStorageControllersList(input.StorageControllers),
		StorageControllersOdataCount: types.Int64Value(int64(input.StorageControllersCount)),
	}
}

// newStorageControllersList converts list of redfish.StorageControllers to list of models.StorageControllers
func newStorageControllersList(inputs []redfish.StorageController) []models.StorageControllers {
	out := make([]models.StorageControllers, 0)
	for _, input := range inputs {
		out = append(out, newStorageControllers(input))
	}
	return out
}

// newStorageControllers converts redfish.StorageControllers to models.StorageControllers
func newStorageControllers(input redfish.StorageController) models.StorageControllers {
	return models.StorageControllers{
		OdataID:                      types.StringValue(input.ODataID),
		CacheSummary:                 newCacheSummary(input.CacheSummary),
		FirmwareVersion:              types.StringValue(input.FirmwareVersion),
		Manufacturer:                 types.StringValue(input.Manufacturer),
		MemberID:                     types.StringValue(input.MemberID),
		Model:                        types.StringValue(input.Model),
		Name:                         types.StringValue(input.Name),
		SpeedGbps:                    types.Int64Value(int64(input.SpeedGbps)),
		Status:                       newStatus(input.Status),
		SupportedControllerProtocols: newProtocols(input.SupportedControllerProtocols),
		SupportedDeviceProtocols:     newProtocols(input.SupportedDeviceProtocols),
		SupportedRAIDTypes:           newRAIDTypes(input.SupportedRAIDTypes),
	}
}

func newProtocols(inputs []common.Protocol) []types.String {
	out := make([]types.String, 0)
	for _, input := range inputs {
		out = append(out, types.StringValue(string(input)))
	}
	return out
}

func newRAIDTypes(inputs []redfish.RAIDType) []types.String {
	out := make([]types.String, 0)
	for _, input := range inputs {
		out = append(out, types.StringValue(string(input)))
	}
	return out
}

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
func newStatus(input common.Status) models.Status {
	return models.Status{
		Health:       types.StringValue(string(input.Health)),
		HealthRollup: types.StringValue(string(input.HealthRollup)),
		State:        types.StringValue(string(input.State)),
	}
}

// newCacheSummary converts redfish.CacheSummary to models.CacheSummary
func newCacheSummary(input redfish.CacheSummary) models.CacheSummary {
	return models.CacheSummary{
		TotalCacheSizeMiB: types.Int64Value(int64(input.TotalCacheSizeMiB)),
	}
}

// StorageSchema is a function that returns the schema for Storage
func StorageSchema() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"odata_context": schema.StringAttribute{
			MarkdownDescription: "odata context of storage",
			Description:         "odata context of storage",
			Computed:            true,
		},
		"odata_id": schema.StringAttribute{
			MarkdownDescription: "odata id of storage",
			Description:         "odata id of storage",
			Computed:            true,
		},
		"odata_type": schema.StringAttribute{
			MarkdownDescription: "odata type of storage",
			Description:         "odata type of storage",
			Computed:            true,
		},
		"description": schema.StringAttribute{
			MarkdownDescription: "description of the storage",
			Description:         "description of the storage",
			Computed:            true,
		},
		"drives": schema.ListAttribute{
			MarkdownDescription: "drives on the storage",
			Description:         "drives on the storage",
			Computed:            true,
			ElementType:         types.StringType,
		},
		"drives_odata_count": schema.Int64Attribute{
			MarkdownDescription: "drives count",
			Description:         "drive count",
			Computed:            true,
		},
		"storage_controller_id": schema.StringAttribute{
			MarkdownDescription: "storage controller id",
			Description:         "storage controller id",
			Computed:            true,
		},
		"name": schema.StringAttribute{
			MarkdownDescription: "name of the storage",
			Description:         "name of the storage",
			Computed:            true,
		},
		"oem": schema.SingleNestedAttribute{
			MarkdownDescription: "oem attributes of storage controller",
			Description:         "oem attributes of storage controller",
			Computed:            true,
			Attributes:          OemSchema(),
		},
		"status": schema.SingleNestedAttribute{
			MarkdownDescription: "status of the storage",
			Description:         "status of the storage",
			Computed:            true,
			Attributes:          StatusSchema(),
		},
		"storage_controllers": schema.ListNestedAttribute{
			MarkdownDescription: "storage controllers list",
			Description:         "storage contollers list",
			Computed:            true,
			NestedObject:        schema.NestedAttributeObject{Attributes: StorageControllersSchema()},
		},
		"storage_controllers_odata_count": schema.Int64Attribute{
			MarkdownDescription: "storage controller count",
			Description:         "storage controller count",
			Computed:            true,
		},
	}
}

// DellControllerSchema is a function that returns the schema for DellController
func DellControllerSchema() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"odata_context": schema.StringAttribute{
			MarkdownDescription: "odata context",
			Description:         "odata context",
			Computed:            true,
		},
		"odata_id": schema.StringAttribute{
			MarkdownDescription: "odata id",
			Description:         "odata id",
			Computed:            true,
		},
		"odata_type": schema.StringAttribute{
			MarkdownDescription: "odata type",
			Description:         "odata type",
			Computed:            true,
		},
		"alarm_state": schema.StringAttribute{
			MarkdownDescription: "alarm state",
			Description:         "alarm state",
			Computed:            true,
		},
		"auto_config_behavior": schema.StringAttribute{
			MarkdownDescription: "auto config behavior",
			Description:         "auto config behavior",
			Computed:            true,
		},
		"boot_virtual_disk_fqdd": schema.StringAttribute{
			MarkdownDescription: "boot virtual disk fqdd",
			Description:         "boot virtual disk fqdd",
			Computed:            true,
		},
		"cache_size_in_mb": schema.Int64Attribute{
			MarkdownDescription: "cache size in mb",
			Description:         "cache size in mb",
			Computed:            true,
		},
		"cachecade_capability": schema.StringAttribute{
			MarkdownDescription: "cachecade capability",
			Description:         "cachecade capability",
			Computed:            true,
		},
		"connector_count": schema.Int64Attribute{
			MarkdownDescription: "connector count",
			Description:         "connector count",
			Computed:            true,
		},
		"controller_firmware_version": schema.StringAttribute{
			MarkdownDescription: "controller firmware version",
			Description:         "controller firmware version",
			Computed:            true,
		},
		"current_controller_mode": schema.StringAttribute{
			MarkdownDescription: "current controller mode",
			Description:         "current controller mode",
			Computed:            true,
		},
		"description": schema.StringAttribute{
			MarkdownDescription: "description",
			Description:         "description",
			Computed:            true,
		},
		"device": schema.StringAttribute{
			MarkdownDescription: "device",
			Description:         "device",
			Computed:            true,
		},
		"device_card_data_bus_width": schema.StringAttribute{
			MarkdownDescription: "device card data bus width",
			Description:         "device card data bus width",
			Computed:            true,
		},
		"device_card_slot_length": schema.StringAttribute{
			MarkdownDescription: "device card slot length",
			Description:         "device card slot length",
			Computed:            true,
		},
		"device_card_slot_type": schema.StringAttribute{
			MarkdownDescription: "device card slot type",
			Description:         "device card slot type",
			Computed:            true,
		},
		"driver_version": schema.StringAttribute{
			MarkdownDescription: "driver version",
			Description:         "driver version",
			Computed:            true,
		},
		"encryption_capability": schema.StringAttribute{
			MarkdownDescription: "encryption capability",
			Description:         "encryption capability",
			Computed:            true,
		},
		"encryption_mode": schema.StringAttribute{
			MarkdownDescription: "encryption mode",
			Description:         "encryption mode",
			Computed:            true,
		},
		"id": schema.StringAttribute{
			MarkdownDescription: "id",
			Description:         "id",
			Computed:            true,
		},
		"key_id": schema.StringAttribute{
			MarkdownDescription: "key id",
			Description:         "key id",
			Computed:            true,
		},
		"last_system_inventory_time": schema.StringAttribute{
			MarkdownDescription: "last system inventory time",
			Description:         "last system inventory time",
			Computed:            true,
		},
		"last_update_time": schema.StringAttribute{
			MarkdownDescription: "last update time",
			Description:         "last update time",
			Computed:            true,
		},
		"max_available_pci_link_speed": schema.StringAttribute{
			MarkdownDescription: "max available pci link speed",
			Description:         "max available pci link speed",
			Computed:            true,
		},
		"max_possible_pci_link_speed": schema.StringAttribute{
			MarkdownDescription: "max possible pci link speed",
			Description:         "max possible pci link speed",
			Computed:            true,
		},
		"name": schema.StringAttribute{
			MarkdownDescription: "name",
			Description:         "name",
			Computed:            true,
		},
		"pci_slot": schema.StringAttribute{
			MarkdownDescription: "pci slot",
			Description:         "pci slot",
			Computed:            true,
		},
		"patrol_read_state": schema.StringAttribute{
			MarkdownDescription: "patrol read state",
			Description:         "patrol read state",
			Computed:            true,
		},
		"persistent_hotspare": schema.StringAttribute{
			MarkdownDescription: "persistent hotspare",
			Description:         "persistent hotspare",
			Computed:            true,
		},
		"realtime_capability": schema.StringAttribute{
			MarkdownDescription: "realtime capability",
			Description:         "realtime capability",
			Computed:            true,
		},
		"rollup_status": schema.StringAttribute{
			MarkdownDescription: "rollup status",
			Description:         "rollup status",
			Computed:            true,
		},
		"sas_address": schema.StringAttribute{
			MarkdownDescription: "sas address",
			Description:         "sas address",
			Computed:            true,
		},
		"security_status": schema.StringAttribute{
			MarkdownDescription: "security status",
			Description:         "security status",
			Computed:            true,
		},
		"shared_slot_assignment_allowed": schema.StringAttribute{
			MarkdownDescription: "shared slot assignment allowed",
			Description:         "shared slot assignment allowed",
			Computed:            true,
		},
		"sliced_vd_capability": schema.StringAttribute{
			MarkdownDescription: "sliced vd capability",
			Description:         "sliced vd capability",
			Computed:            true,
		},
		"support_controller_boot_mode": schema.StringAttribute{
			MarkdownDescription: "support controller boot mode",
			Description:         "support controller boot mode",
			Computed:            true,
		},
		"support_enhanced_auto_foreign_import": schema.StringAttribute{
			MarkdownDescription: "support enhanced auto foreign import",
			Description:         "support enhanced auto foreign import",
			Computed:            true,
		},
		"support_raid_10_uneven_spans": schema.StringAttribute{
			MarkdownDescription: "support raid 10 uneven spans",
			Description:         "support raid 10 uneven spans",
			Computed:            true,
		},
		"supports_lk_mto_sekm_transition": schema.StringAttribute{
			MarkdownDescription: "supports lk mto sekm transition",
			Description:         "supports lk mto sekm transition",
			Computed:            true,
		},
		"t_10_pi_capability": schema.StringAttribute{
			MarkdownDescription: "t 10 pi capability",
			Description:         "t 10 pi capability",
			Computed:            true,
		},
	}
}

// DellControllerBatterySchema is a function that returns the schema for DellControllerBattery
func DellControllerBatterySchema() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"odata_context": schema.StringAttribute{
			MarkdownDescription: "odata context",
			Description:         "odata context",
			Computed:            true,
		},
		"odata_id": schema.StringAttribute{
			MarkdownDescription: "odata id",
			Description:         "odata id",
			Computed:            true,
		},
		"odata_type": schema.StringAttribute{
			MarkdownDescription: "odata type",
			Description:         "odata type",
			Computed:            true,
		},
		"description": schema.StringAttribute{
			MarkdownDescription: "description",
			Description:         "description",
			Computed:            true,
		},
		"fqdd": schema.StringAttribute{
			MarkdownDescription: "fqdd",
			Description:         "fqdd",
			Computed:            true,
		},
		"id": schema.StringAttribute{
			MarkdownDescription: "id",
			Description:         "id",
			Computed:            true,
		},
		"name": schema.StringAttribute{
			MarkdownDescription: "name",
			Description:         "name",
			Computed:            true,
		},
		"primary_status": schema.StringAttribute{
			MarkdownDescription: "primary_status",
			Description:         "primary_status",
			Computed:            true,
		},
		"raid_state": schema.StringAttribute{
			MarkdownDescription: "raid state",
			Description:         "raid state",
			Computed:            true,
		},
	}
}

// DellSchema is a function that returns the schema for Dell
func DellSchema() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"odata_type": schema.StringAttribute{
			MarkdownDescription: "odata type",
			Description:         "odata type",
			Computed:            true,
		},
		"dell_controller": schema.SingleNestedAttribute{
			MarkdownDescription: "dell controller",
			Description:         "dell controller",
			Computed:            true,
			Attributes:          DellControllerSchema(),
		},
		"dell_controller_battery": schema.SingleNestedAttribute{
			MarkdownDescription: "dell controller battery",
			Description:         "dell controller battery",
			Computed:            true,
			Attributes:          DellControllerBatterySchema(),
		},
	}
}

// OemSchema is a function that returns the schema for Oem
func OemSchema() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"dell": schema.SingleNestedAttribute{
			MarkdownDescription: "dell attributes",
			Description:         "dell attributes",
			Computed:            true,
			Attributes:          DellSchema(),
		},
	}
}

// StatusSchema is a function that returns the schema for Status
func StatusSchema() map[string]schema.Attribute {
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
			MarkdownDescription: "state",
			Description:         "state",
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

// StorageControllersSchema is a function that returns the schema for StorageControllers
func StorageControllersSchema() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"odata_id": schema.StringAttribute{
			MarkdownDescription: "odata id",
			Description:         "odata id",
			Computed:            true,
		},
		"cache_summary": schema.SingleNestedAttribute{
			MarkdownDescription: "cache summary",
			Description:         "cache summary",
			Computed:            true,
			Attributes:          CacheSummarySchema(),
		},
		"firmware_version": schema.StringAttribute{
			MarkdownDescription: "firmware version",
			Description:         "firmware version",
			Computed:            true,
		},
		"manufacturer": schema.StringAttribute{
			MarkdownDescription: "manufacturer",
			Description:         "manufacturer",
			Computed:            true,
		},
		"member_id": schema.StringAttribute{
			MarkdownDescription: "member id",
			Description:         "member id",
			Computed:            true,
		},
		"model": schema.StringAttribute{
			MarkdownDescription: "model",
			Description:         "model",
			Computed:            true,
		},
		"name": schema.StringAttribute{
			MarkdownDescription: "name",
			Description:         "name",
			Computed:            true,
		},
		"speed_gbps": schema.Int64Attribute{
			MarkdownDescription: "speed gbps",
			Description:         "speed gbps",
			Computed:            true,
		},
		"status": schema.SingleNestedAttribute{
			MarkdownDescription: "status",
			Description:         "status",
			Computed:            true,
			Attributes:          StatusSchema(),
		},
		"supported_controller_protocols": schema.ListAttribute{
			MarkdownDescription: "supported controller protocols",
			Description:         "supported controller protocols",
			Computed:            true,
			ElementType:         types.StringType,
		},
		"supported_device_protocols": schema.ListAttribute{
			MarkdownDescription: "supported device protocols",
			Description:         "supported device protocols",
			Computed:            true,
			ElementType:         types.StringType,
		},
		"supported_raid_types": schema.ListAttribute{
			MarkdownDescription: "supported raid types",
			Description:         "supported raid types",
			Computed:            true,
			ElementType:         types.StringType,
		},
	}
}
