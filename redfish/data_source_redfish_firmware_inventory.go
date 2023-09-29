package redfish

import (
	"context"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stmcginnis/gofish"
	"github.com/stmcginnis/gofish/redfish"
)

type InventoryItem struct {
	entityID   string
	entityName string
	version    string
}

func dataSourceRedfishFirmwareInventory() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceRedfishFirmwareInventoryRead,
		Schema:      getDataSourceRedfishFirmwareInventorySchema(),
	}
}

func getDataSourceRedfishFirmwareInventorySchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"redfish_server": {
			Type:        schema.TypeList,
			Required:    true,
			Description: "List of server BMCs and their respective user credentials",
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"user": {
						Type:        schema.TypeString,
						Optional:    true,
						Description: "User name for login",
					},
					"password": {
						Type:        schema.TypeString,
						Optional:    true,
						Description: "User password for login",
						Sensitive:   true,
					},
					"endpoint": {
						Type:        schema.TypeString,
						Required:    true,
						Description: "Server BMC IP address or hostname",
					},
					"ssl_insecure": {
						Type:        schema.TypeBool,
						Optional:    true,
						Description: "This field indicates whether the SSL/TLS certificate must be verified or not",
					},
				},
			},
		},
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
					"entity_name": {
						Type:     schema.TypeString,
						Computed: true,
					},
					"entity_id": {
						Type:     schema.TypeString,
						Computed: true,
					},
					"version": {
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

func flattenInventoryItems(inventoryItems *[]InventoryItem) []interface{} {
	if inventoryItems != nil {
		inv := make([]interface{}, len(*inventoryItems))
		for i, invItem := range *inventoryItems {
			inv[i] = map[string]any{
				"entity_name": invItem.entityID,
				"entity_id":   invItem.entityName,
				"version":     invItem.version,
			}
		}
		return inv
	}
	return make([]interface{}, 0)
}

func getInventoryItems(fwInventories []*redfish.SoftwareInventory) []InventoryItem {
	var inv InventoryItem

	inventoryItemList := make([]InventoryItem, 0)

	for _, fwInv := range fwInventories {
		if strings.HasPrefix(fwInv.Entity.ID, "Installed") {
			inv.entityID = fwInv.Entity.ID
			inv.entityName = fwInv.Entity.Name
			inv.version = fwInv.Version

			inventoryItemList = append(inventoryItemList, inv)
		}
	}
	return inventoryItemList
}

func readRedfishFirmwareInventory(service *gofish.Service, d *schema.ResourceData) diag.Diagnostics {
	var diags diag.Diagnostics

	updateService, err := service.UpdateService()
	if err != nil {
		return diag.Errorf("Error fetching UpdateService collection: %s", err)
	}

	fwInventories, err := updateService.FirmwareInventories()
	if err != nil {
		return diag.Errorf("Error fetching Firmware Inventory: %s", err)
	}

	// Filter inventory which are prefixed as "Installed"
	inventoryItems := getInventoryItems(fwInventories)

	// Flatten array of InventoryItem to array of key-value pair objects
	inventoryList := flattenInventoryItems(&inventoryItems)

	if err := d.Set("odata_id", updateService.ODataID); err != nil {
		return diag.Errorf("error setting UpdateService OData ID: %s", err)
	}

	if err := d.Set("id", updateService.ID); err != nil {
		return diag.Errorf("error setting UpdateService ID: %s", err)
	}

	if err := d.Set("inventory", inventoryList); err != nil {
		return diag.Errorf("error setting Firmware Inventory: %s", err)
	}

	serverConfig := d.Get("redfish_server").([]interface{})
	endpoint := serverConfig[0].(map[string]interface{})["endpoint"].(string)
	fwResourceID := endpoint + updateService.ODataID
	d.SetId(fwResourceID)

	return diags
}
