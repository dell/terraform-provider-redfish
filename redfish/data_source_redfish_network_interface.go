package redfish

import (
	"context"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stmcginnis/gofish"
	"github.com/stmcginnis/gofish/redfish"
)

func dataSourceRedfishNetworkInterface() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceRedfishNetworkInterfaceRead,
		Schema:      getDataSourceRedfishNetworkInterfaceSchema(),
	}
}

func getDataSourceRedfishNetworkInterfaceSchema() map[string]*schema.Schema {
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
		"interfaces": {
			Type:        schema.TypeList,
			Description: "",
			Computed:    true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"id": {
						Description: "",
						Computed:    true,
						Type:        schema.TypeString,
					},
					"ports": {
						Description: "",
						Computed:    true,
						Type:        schema.TypeList,
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"id": {
									Description: "",
									Computed:    true,
									Type:        schema.TypeString,
								},
								"addresses": {
									Description: "",
									Computed:    true,
									Type:        schema.TypeList,
									Elem: &schema.Schema{
										Type: schema.TypeString,
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func dataSourceRedfishNetworkInterfaceRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	service, err := NewConfig(m.(*schema.ResourceData), d)
	if err != nil {
		return diag.Errorf(err.Error())
	}
	return readRedfishNetworkInterface(service, d)
}

func readRedfishNetworkInterface(service *gofish.Service, d *schema.ResourceData) diag.Diagnostics {
	var diags diag.Diagnostics

	systems, err := service.Systems()
	if err != nil {
		return diag.Errorf("Error when retrieving systems: %s", err)
	}

	interfaces, err := systems[0].NetworkInterfaces()
	if err != nil {
		return diag.Errorf("Error when retrieving network interfaces: %s", err)
	}

	result := make([]map[string]interface{}, 0)
	for _, i := range interfaces {
		ports, err := getNetworkPorts(i)
		if err != nil {
			return diag.Errorf("Error when retrieving network ports: %s", err)
		}

		result = append(result, map[string]interface{}{
			"id":    i.ID,
			"ports": ports,
		})
	}

	d.Set("interfaces", result)
	d.SetId(strconv.FormatInt(time.Now().Unix(), 10))

	return diags
}

func getNetworkPorts(i *redfish.NetworkInterface) ([]map[string]interface{}, error) {

	ports, err := i.NetworkPorts()
	if err != nil {
		return nil, err
	}

	result := make([]map[string]interface{}, 0)
	for _, p := range ports {
		result = append(result, map[string]interface{}{
			"id":        p.ID,
			"addresses": p.AssociatedNetworkAddresses,
		})
	}

	return result, nil
}
