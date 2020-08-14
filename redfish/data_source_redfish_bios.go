package redfish

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/stmcginnis/gofish"
)

func dataSourceRedfishBios() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceRedfishBiosRead,

		Schema: map[string]*schema.Schema{
			"attributes": {
				Type:        schema.TypeMap,
				Description: "Bios attributes",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Computed:    true,
			},
		},
	}
}

func dataSourceRedfishBiosRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*gofish.APIClient)

	service := conn.Service
	systems, err := service.Systems()
	if err != nil {
		return fmt.Errorf("error fetching computer systems collection: %s", err)

	}

	bios, err := systems[0].Bios()
	if err != nil {
		return fmt.Errorf("error fetching bios: %s", err)
	}

	// TODO: BIOS Attributes' values might be any of several types.
	// terraform-sdk currently does not support a map with different
	// value types. So we will convert int and float values to string
	bios_attributes_map := make(map[string]string)

	// copy from the BIOS attributes to the new bios attributes map
	for key, value := range bios.Attributes {
		if attr_val, ok := value.(string); ok {
			bios_attributes_map[key] = attr_val
		} else {
			bios_attributes_map[key] = fmt.Sprintf("%v", value)
		}
	}

	d.SetId("attributes")
	if err := d.Set("attributes", bios_attributes_map); err != nil {
		return fmt.Errorf("error setting bios attributes: %s", err)
	}

	return nil
}
