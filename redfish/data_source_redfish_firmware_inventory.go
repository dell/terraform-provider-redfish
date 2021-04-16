package redfish

import (
	"context"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stmcginnis/gofish"
)

func dataSourceRedfishFirmwareInventory() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceRedfishFirmwareInventoryRead,
		Schema:      getDataSourceRedfishFirmwareInventorySchema(),
	}
}

func getDataSourceRedfishFirmwareInventorySchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"odata_id": {
			Type:        schema.TypeString,
			Description: "OData ID for the Firmware Inventory resource",
			Computed:    true,
		},
		"id": {
			Type:        schema.TypeString,
			Description: "Id",
			Computed:    true,
		},
		"inventory": {
			Type:        schema.TypeList,
			Description: "Firmware Inventory",
			Computed:    true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"EntityName": {
						Type:     schema.TypeString,
						Computed: true,
					},
					"EntityId": {
						Type:     schema.TypeString,
						Computed: true,
					},
					"Version": {
						Type:     schema.TypeString,
						Computed: true,
					},
				},
			},
		},
	}
}

func dataSourceRedfishFirmwareInventoryRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	service, err := NewConfig(m.(*schema.ResourceData), d)
	if err != nil {
		return diag.Errorf(err.Error())
	}
	return readRedfishFirmwareInventory(service, d)
}

func readRedfishFirmwareInventory(service *gofish.Service, d *schema.ResourceData) diag.Diagnostics {
	var diags diag.Diagnostics

	type InventoryItem struct {
		entityId   string
		entityName string
		version    string
	}

	updateService, err := service.UpdateService()
	if err != nil {
		return diag.Errorf("Error fetching UpdateService collection: %s", err)
	}

	var inv InventoryItem
	inventoryItemList := make([]InventoryItem, 10)
	fwInventories, err := updateService.FirmwareInventories()
	for _, fwInv := range fwInventories {

		if strings.HasPrefix(fwInv.Entity.ID, "Installed") {
			inv.entityId = fwInv.Entity.ID
			inv.entityName = fwInv.Entity.Name
			inv.version = fwInv.Version

			inventoryItemList = append(inventoryItemList, inv)
		}
	}
	inventory := make(map[string]interface{})
	inventory["inventory"] = inventoryItemList

	if err := d.Set("odata_id", updateService.ODataID); err != nil {
		return diag.Errorf("error setting UpdateService OData ID: %s", err)
	}

	if err := d.Set("id", updateService.ID); err != nil {
		return diag.Errorf("error setting UpdateService ID: %s", err)
	}

	if err := d.Set("attributes", inventory); err != nil {
		return diag.Errorf("error setting Firmware Inventory: %s", err)
	}

	return diags
}
