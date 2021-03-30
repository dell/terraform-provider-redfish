package redfish

import (
	"context"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stmcginnis/gofish"
)

func dataSourceRedfishStorage() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceRedfishStorageRead,
		Schema:      getDataSourceRedfishStorageSchema(),
	}
}

func getDataSourceRedfishStorageSchema() map[string]*schema.Schema {
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
		"storage": {
			Type:        schema.TypeList,
			Description: "List of storage and disks attached available on this instance",
			Computed:    true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"storage_controller_id": {
						Description: "Disks attached to the storage resource",
						Computed:    true,
						Type:        schema.TypeString,
					},
					"drives": {
						Type:        schema.TypeList,
						Description: "Disks attached to the storage resource",
						Computed:    true,
						Elem: &schema.Schema{
							Type: schema.TypeString,
						},
					},
				},
			},
		},
	}
}

func dataSourceRedfishStorageRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	service, err := NewConfig(m.(*schema.ResourceData), d)
	if err != nil {
		return diag.Errorf(err.Error())
	}
	return readRedfishStorageCollection(service, d)
}

func readRedfishStorageCollection(service *gofish.Service, d *schema.ResourceData) diag.Diagnostics {
	var diags diag.Diagnostics
	m := make([]map[string]interface{}, 0) //List where all storage controller will be held

	systems, err := service.Systems()
	if err != nil {
		return diag.Errorf("Error when retrieving systems: %s", err)
	}
	storage, err := systems[0].Storage()
	if err != nil {
		return diag.Errorf("Error when retrieving storage: %s", err)
	}

	var mToAdd map[string]interface{} //Map where each controller and its disks will be written
	for _, s := range storage {
		mToAdd = make(map[string]interface{}) //Create new mToAdd instace
		mToAdd["storage_controller_id"] = s.ID
		drives, err := s.Drives()
		if err != nil {
			return diag.Errorf("Error when retrieving drives: %s", err)
		}

		driveNames := make([]interface{}, 0)
		for _, d := range drives {
			driveNames = append(driveNames, d.Name)
		}
		mToAdd["drives"] = driveNames
		m = append(m, mToAdd) //Insert controller into list
	}

	d.Set("storage", m)
	d.SetId(strconv.FormatInt(time.Now().Unix(), 10))

	return diags
}
