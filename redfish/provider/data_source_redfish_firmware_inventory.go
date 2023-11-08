package provider

import (
	"context"
	"fmt"
	"strings"
	"terraform-provider-redfish/redfish/models"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stmcginnis/gofish"
	"github.com/stmcginnis/gofish/redfish"
)

var (
	_ datasource.DataSource              = &FirmwareInventoryDatasource{}
	_ datasource.DataSourceWithConfigure = &FirmwareInventoryDatasource{}
)

// NewFirmwareInventoryDatasource is new datasource for FirmwareInventory
func NewFirmwareInventoryDatasource() datasource.DataSource {
	return &FirmwareInventoryDatasource{}
}

// FirmwareInventoryDatasource to construct datasource
type FirmwareInventoryDatasource struct {
	p *redfishProvider
}

// Configure implements datasource.DataSourceWithConfigure
func (g *FirmwareInventoryDatasource) Configure(_ context.Context, req datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	g.p = req.ProviderData.(*redfishProvider)
}

// Metadata implements datasource.DataSource
func (*FirmwareInventoryDatasource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "firmware_inventory"
}

// Schema implements datasource.DataSource
func (*FirmwareInventoryDatasource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Data source to fetch Firmware Inventory details via RedFish.",
		Attributes:          FirmwareInventoryDatasourceSchema(),
		Blocks:              RedfishServerDatasourceBlockMap(),
	}
}

// FirmwareInventoryDatasourceSchema to define the Firmware Inventory data-source schema
func FirmwareInventoryDatasourceSchema() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"id": schema.StringAttribute{
			MarkdownDescription: "ID of the Firmware Inventory data-source",
			Description:         "ID of the Firmware Inventory data-source",
			Computed:            true,
		},
		"odata_id": schema.StringAttribute{
			MarkdownDescription: "OData ID of the Firmware Inventory data-source",
			Description:         "OData ID of the Firmware Inventory data-source",
			Computed:            true,
		},
		"inventory": schema.ListNestedAttribute{
			MarkdownDescription: "Firmware Inventory.",
			Description:         "Firmware Inventory.",
			Computed:            true,
			NestedObject: schema.NestedAttributeObject{
				Attributes: map[string]schema.Attribute{
					"entity_name": schema.StringAttribute{
						Computed: true,
						Description: "entity name of the firmware inventory",
						MarkdownDescription: "entity name of the firmware inventory",
					},
					"entity_id": schema.StringAttribute{
						Computed: true,
						Description: "entity id of the firmware inventory",
						MarkdownDescription: "entity id of the firmware inventory",
					},
					"version": schema.StringAttribute{
						Computed: true,
						Description: "firmware inventory version",
						MarkdownDescription: "firmware inventory version",
					},
				},
			},
		},
	}
}

// Read implements datasource.DataSource
func (g *FirmwareInventoryDatasource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var plan models.FirmwareInventory
	diags := req.Config.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	service, err := NewConfig(g.p, &plan.RedfishServer)
	if err != nil {
		resp.Diagnostics.AddError("service error", err.Error())
		return
	}
	state, err := readRedfishFirmwareInventory(service)
	if err != nil {
		diags.AddError("failed to fetch firmware inventory details", err.Error())
	}
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func getInventoryItems(fwInventories []*redfish.SoftwareInventory) []models.Inventory {
	var inv models.Inventory

	inventoryItemList := make([]models.Inventory, 0)

	for _, fwInv := range fwInventories {
		if strings.HasPrefix(fwInv.Entity.ID, "Installed") {
			inv.EntityId = types.StringValue(fwInv.Entity.ID)
			inv.EntityName = types.StringValue(fwInv.Entity.Name)
			inv.Version = types.StringValue(fwInv.Version)
			inventoryItemList = append(inventoryItemList, inv)
		}
	}
	return inventoryItemList
}

func readRedfishFirmwareInventory(service *gofish.Service) (*models.FirmwareInventory, error) {
	updateService, err := service.UpdateService()
	if err != nil {
		return nil, fmt.Errorf("error fetching UpdateService collection: %s", err.Error())
	}

	fwInventories, err := updateService.FirmwareInventories()
	if err != nil {
		return nil, fmt.Errorf("error fetching Firmware Inventory: %s", err)
	}

	// Filter inventory which are prefixed as "Installed"
	inventoryItems := getInventoryItems(fwInventories)

	firmwareState := models.FirmwareInventory{
		OdataID:   types.StringValue(updateService.ODataID),
		ID:        types.StringValue(updateService.ID),
		Inventory: inventoryItems,
	}
	return &firmwareState, nil
}
