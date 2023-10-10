package provider

// import (
// 	"context"
// 	"fmt"

// 	"github.com/dell/terraform-provider-redfish/gofish/dell"
// 	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
// 	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
// 	"github.com/stmcginnis/gofish"
// )

// func dataSourceRedfishDellIdracAttributes() *schema.Resource {
// 	return &schema.Resource{
// 		ReadContext: dataSourceRedfishDellIdracAttributesRead,
// 		Schema:      getDataSourceRedfishDellIdracAttributesSchema(),
// 	}
// }

// func getDataSourceRedfishDellIdracAttributesSchema() map[string]*schema.Schema {
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
// 		"attributes": {
// 			Type:        schema.TypeMap,
// 			Computed:    true,
// 			Description: "iDRAC attributes available to set",
// 			Elem: &schema.Schema{
// 				Type: schema.TypeString,
// 			},
// 		},
// 	}
// }

// func dataSourceRedfishDellIdracAttributesRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
// 	service, err := NewConfig(m.(*schema.ResourceData), d)
// 	if err != nil {
// 		return diag.Errorf(err.Error())
// 	}
// 	return readDatasourceRedfishDellIdracAttributes(service, d)
// }

// func readDatasourceRedfishDellIdracAttributes(service *gofish.Service, d *schema.ResourceData) diag.Diagnostics {
// 	var diags diag.Diagnostics

// 	// get managers (Dell servers have only the iDRAC)
// 	managers, err := service.Managers()
// 	if err != nil {
// 		return diag.Errorf("there was an issue when reading idrac attributes - %s", err)
// 	}

// 	// Get OEM
// 	dellManager, err := dell.DellManager(managers[0])
// 	if err != nil {
// 		return diag.Errorf("there was an issue when reading idrac attributes - %s", err)
// 	}

// 	// Get Dell attributes
// 	dellAttributes, err := dellManager.DellAttributes()
// 	if err != nil {
// 		return diag.Errorf("there was an issue when reading idrac attributes - %s", err)
// 	}
// 	idracAttributes, err := getIdracAttributes(dellAttributes)
// 	if err != nil {
// 		return diag.Errorf("there was an issue when reading idrac attributes - %s", err)
// 	}

// 	attributesToReturn := make(map[string]string)

// 	for k, v := range idracAttributes.Attributes {
// 		if v != nil {
// 			attributesToReturn[k] = fmt.Sprintf("%v", v)
// 		} else {
// 			attributesToReturn[k] = ""
// 		}
// 	}

// 	err = d.Set("attributes", attributesToReturn)
// 	if err != nil {
// 		return diag.Errorf("there was an issue when reading idrac attributes - %s", err)
// 	}

// 	d.SetId(idracAttributes.ODataID)

// 	return diags
// }
