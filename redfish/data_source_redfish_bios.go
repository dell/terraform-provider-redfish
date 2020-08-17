package redfish

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stmcginnis/gofish"
)

func dataSourceRedfishBios() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceRedfishBiosRead,
		Schema: map[string]*schema.Schema{
			"odata_id": {
				Type: schema.TypeString,
				Description: "ODataID",
				Computed: true,
			},
			"attributes": {
				Type: schema.TypeMap,
				Description: "Bios attributes",
				Elem: &schema.Schema{
					Type: schema.TypeString,
					Computed: true,
				},
				Computed:    true,
			},
			"id": {
				Type: schema.TypeString,
				Description: "Id",
				Computed: true,
			},
		},
	}
}

func dataSourceRedfishBiosRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*gofish.APIClient)

	service := conn.Service
	systems, err := service.Systems()
	if err != nil {
		return diag.Errorf("error fetching computer systems collection: %s", err)

	}

	bios, err := systems[0].Bios()
	if err != nil {
		return diag.Errorf("error fetching bios: %s", err)
	}

	// TODO: BIOS Attributes' values might be any of several types.
	// terraform-sdk currently does not support a map with different
	// value types. So we will convert int and float values to string
	bios_attributes_map := make(map[string]string)

	bios_id := bios.ID
	bios_odata_id := bios.ODataID

	// copy from the BIOS attributes to the new bios attributes map
	for key, value := range bios.Attributes {
		if attr_val, ok := value.(string); ok {
			bios_attributes_map[key] = attr_val
		} else {
			bios_attributes_map[key] = fmt.Sprintf("%v", value)
		}
	}

	if err := d.Set("odata_id", bios.ODataID); err != nil {
		return diag.Errorf("error setting bios OData ID: %s", err)
	}

	if err := d.Set("id", bios.ID); err != nil {
                return diag.Errorf("error setting bios ID: %s", err)
        }

	if err := d.Set("attributes", bios_attributes_map); err != nil {
                return diag.Errorf("error setting bios attributes: %s", err)
        }

	// Set the ID to the @odata.id
	d.SetId(bios.ODataID)

	return nil
}
