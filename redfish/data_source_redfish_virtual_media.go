package redfish

// import (
// 	"context"
// 	"log"
// 	"strconv"
// 	"time"

// 	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
// 	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
// 	"github.com/stmcginnis/gofish"
// )

// func dataSourceRedfishVirtualMedia() *schema.Resource {
// 	return &schema.Resource{
// 		ReadContext: dataSourceRedfishVirtualMediaRead,
// 		Schema:      getDataSourceRedfishVirtualMediaSchema(),
// 	}
// }

// func getDataSourceRedfishVirtualMediaSchema() map[string]*schema.Schema {
// 	return map[string]*schema.Schema{
// 		"redfish_server": {
// 			Type:        schema.TypeList,
// 			Required:    true,
// 			Description: "List of server BMCs and their respective user credentials",
// 			Elem: &schema.Resource{
// 				Schema: map[string]*schema.Schema{
// 					"user": {
// 						Type:        schema.TypeString,
// 						Optional:    true,
// 						Description: "User name for login",
// 					},
// 					"password": {
// 						Type:        schema.TypeString,
// 						Optional:    true,
// 						Description: "User password for login",
// 						Sensitive:   true,
// 					},
// 					"endpoint": {
// 						Type:        schema.TypeString,
// 						Required:    true,
// 						Description: "Server BMC IP address or hostname",
// 					},
// 					"ssl_insecure": {
// 						Type:        schema.TypeBool,
// 						Optional:    true,
// 						Description: "This field indicates whether the SSL/TLS certificate must be verified or not",
// 					},
// 				},
// 			},
// 		},
// 		"virtual_media": {
// 			Type:        schema.TypeList,
// 			Description: "List of virtual media available on this instance",
// 			Computed:    true,
// 			Elem: &schema.Resource{
// 				Schema: map[string]*schema.Schema{
// 					"odata_id": {
// 						Type:        schema.TypeString,
// 						Description: "OData ID for the Virtual Media resource",
// 						Computed:    true,
// 					},
// 					"id": {
// 						Type:        schema.TypeString,
// 						Description: "Id of the virtual media resource",
// 						Computed:    true,
// 					},
// 				},
// 			},
// 		},
// 	}
// }

// func dataSourceRedfishVirtualMediaRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
// 	service, err := NewConfig(m.(*schema.ResourceData), d)
// 	if err != nil {
// 		return diag.Errorf(err.Error())
// 	}
// 	return readRedfishVirtualMediaCollection(service, d)
// }

// func readRedfishVirtualMediaCollection(service *gofish.Service, d *schema.ResourceData) diag.Diagnostics {
// 	var diags diag.Diagnostics

// 	//Get manager.Since this provider is thought to work with individual servers, should be only one.
// 	manager, err := service.Managers()
// 	if err != nil {
// 		return diag.Errorf("Error retrieving the managers: %s", err)
// 	}

// 	//Get virtual media
// 	virtualMedia, err := manager[0].VirtualMedia()
// 	if err != nil {
// 		return diag.Errorf("Error retrieving the virtual media instances: %s", err)
// 	}

// 	vms := make([]map[string]interface{}, 0)
// 	for _, v := range virtualMedia {
// 		vmToAdd := make(map[string]interface{})
// 		log.Printf("Adding %s - %s", v.ODataID, v.ID)
// 		vmToAdd["odata_id"] = v.ODataID
// 		vmToAdd["id"] = v.ID
// 		vms = append(vms, vmToAdd)
// 	}
// 	d.Set("virtual_media", vms)
// 	d.SetId(strconv.FormatInt(time.Now().Unix(), 10))

// 	return diags
// }
