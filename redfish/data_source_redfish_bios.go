package redfish

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceRedfishBios() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceRedfishBiosRead,
		Schema: getDataSourceRedfishBiosSchema(),
	}
}

func getDataSourceRedfishBiosSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"redfish_server": {
			Type: schema.TypeList,
			Required: true,
			Description: "List of server BMCs and their respective user credentials",
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"user": {
						Type: schema.TypeString,
						Optional: true,
						Description: "User name for login",
					},
					"password": {
						Type: schema.TypeString,
						Optional: true,
						Description: "User password for login",
						Sensitive: true,
					},
					"endpoint": {
						Type: schema.TypeString,
						Required: true,
						Description: "Server BMC IP address or hostname",
					},
					"ssl_insecure": {
						Type: schema.TypeBool,
						Optional: true,
						Description: "This field indicates whether the SSL/TLS certificate must be verified or not",
					},
				},
			},
		},
		"odata_id": {
			Type: schema.TypeString,
			Description: "OData ID for the Bios resource",
			Computed: true,
		},
		"attributes": {
			Type: schema.TypeMap,
			Description: "Bios attributes",
			Elem: &schema.Schema{
				Type: schema.TypeString,
				Computed: true,
			},
			Computed: true,
		},
		"id": {
			Type: schema.TypeString,
			Description: "Id",
			Computed: true,
		},
	}
}

func dataSourceRedfishBiosRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	service, err := NewConfig(m.(*schema.ResourceData), d)
	if err != nil {
		return diag.Errorf(err.Error())
	}

	var diags diag.Diagnostics

	systems, err := service.Systems()
	if err != nil {
		return diag.Errorf("error fetching computer systems collection: %s", err)
	}

	bios, err := systems[0].Bios()
	if err != nil {
		return diag.Errorf("error fetching bios: %s", err)
	}

	// BIOS attributes values might be any of several types.
	// terraform-sdk currently does not support a map with different
	// value types. So we will convert int and float values to string
	// See https://stackoverflow.com/questions/66991765/terraform-sdk-custom-provider-how-to-accept-json-input-in-data-source
	attributes := make(map[string]string)

	// copy from the BIOS attributes to the new bios attributes map
	for key, value := range bios.Attributes {
		if attr_val, ok := value.(string); ok {
			attributes[key] = attr_val
		} else {
			attributes[key] = fmt.Sprintf("%v", value)
		}
	}

	if err := d.Set("odata_id", bios.ODataID); err != nil {
		return diag.Errorf("error setting bios OData ID: %s", err)
	}

	if err := d.Set("id", bios.ID); err != nil {
                return diag.Errorf("error setting bios ID: %s", err)
        }

	if err := d.Set("attributes", attributes); err != nil {
                return diag.Errorf("error setting bios attributes: %s", err)
        }

	// Set the ID to the redfish endpoint + bios @odata.id
	serverConfig := d.Get("redfish_server").([]interface{})
	endpoint := serverConfig[0].(map[string]interface{})["endpoint"].(string)
	biosResourceId := endpoint + bios.ODataID
	d.SetId(biosResourceId)

	return diags
}

