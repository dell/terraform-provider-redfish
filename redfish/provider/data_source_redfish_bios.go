package provider

import (
	"context"
	"fmt"
	"terraform-provider-redfish/redfish/models"

	"github.com/stmcginnis/gofish"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ datasource.DataSource              = &BiosDatasource{}
	_ datasource.DataSourceWithConfigure = &BiosDatasource{}
)

// NewBiosDatasource is new datasource for idrac attributes
func NewBiosDatasource() datasource.DataSource {
	return &BiosDatasource{}
}

// BiosDatasource to construct datasource
type BiosDatasource struct {
	p       *redfishProvider
	ctx     context.Context
	service *gofish.Service
}

// Configure implements datasource.DataSourceWithConfigure
func (g *BiosDatasource) Configure(_ context.Context, req datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	g.p = req.ProviderData.(*redfishProvider)
}

// Metadata implements datasource.DataSource
func (*BiosDatasource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "dell_idrac_attributes"
}

// Schema implements datasource.DataSource
func (*BiosDatasource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Data source to provide redfish infiziya",
		Attributes:          BiosDatasourceSchema(),
	}
}

// BiosDatasourceSchema to define the idrac attribute schema
func BiosDatasourceSchema() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"id": schema.StringAttribute{
			MarkdownDescription: "ID of the BIOS data-source",
			Description:         "ID of the BIOS data-source",
			Computed:            true,
		},
		"odata_id": schema.StringAttribute{
			MarkdownDescription: "OData ID of the BIOS data-source",
			Description:         "OData ID of the BIOS data-source",
			Computed:            true,
		},
		"redfish_server": schema.SingleNestedAttribute{
			MarkdownDescription: "Redfish Server",
			Description:         "Redfish Server",
			Required:            true,
			Attributes:          RedfishServerDatasourceSchema(),
		},
		"attributes": schema.MapAttribute{
			MarkdownDescription: "BIOS attributes.",
			Description:         "BIOS attributes.",
			ElementType:         types.StringType,
			Computed:            true,
		},
	}
}

// Read implements datasource.DataSource
func (g *BiosDatasource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var plan models.BiosDatasource
	diags := req.Config.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if plan.ID.IsUnknown() {
		plan.ID = types.StringValue("placeholder")
	}
	service, err := NewConfig(g.p, &plan.RedfishServer)
	if err != nil {
		resp.Diagnostics.AddError("service error", err.Error())
		return
	}
	g.ctx = ctx
	g.service = service
	state, diags := g.readDatasourceRedfishBios(plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (g *BiosDatasource) readDatasourceRedfishBios(d models.BiosDatasource) (models.BiosDatasource, diag.Diagnostics) {
	var diags diag.Diagnostics

	systems, err := g.service.Systems()
	if err != nil {
		diags.AddError("Error fetching computer systems collection", err.Error())
		return d, diags
	}

	bios, err := systems[0].Bios()
	if err != nil {
		diags.AddError("Error fetching bios", err.Error())
		return d, diags
	}

	// TODO: BIOS Attributes' values might be any of several types.
	// terraform-sdk currently does not support a map with different
	// value types. So we will convert int and float values to string
	attributes := make(map[string]attr.Value)

	// copy from the BIOS attributes to the new bios attributes map
	for key, value := range bios.Attributes {
		if attr_val, ok := value.(string); ok {
			attributes[key] = types.StringValue(attr_val)
		} else {
			attributes[key] = types.StringValue(fmt.Sprintf("%v", value))
		}
	}

	d.OdataID = types.StringValue(bios.ODataID)
	d.ID = types.StringValue(bios.ID)
	d.Attributes, diags = types.MapValue(types.StringType, attributes)

	return d, diags
}

// func dataSourceRedfishBios() *schema.Resource {
// 	return &schema.Resource{
// 		ReadContext: dataSourceRedfishBiosRead,
// 		Schema:      getDataSourceRedfishBiosSchema(),
// 	}
// }

// func getDataSourceRedfishBiosSchema() map[string]*schema.Schema {
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
// 		"odata_id": {
// 			Type:        schema.TypeString,
// 			Description: "OData ID for the Bios resource",
// 			Computed:    true,
// 		},
// 		"attributes": {
// 			Type:        schema.TypeMap,
// 			Description: "Bios attributes",
// 			Elem: &schema.Schema{
// 				Type:     schema.TypeString,
// 				Computed: true,
// 			},
// 			Computed: true,
// 		},
// 		"id": {
// 			Type:        schema.TypeString,
// 			Description: "Id",
// 			Computed:    true,
// 		},
// 	}
// }

// func dataSourceRedfishBiosRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
// 	service, err := NewConfig(m.(*schema.ResourceData), d)
// 	if err != nil {
// 		return diag.Errorf(err.Error())
// 	}
// 	return readRedfishBios(service, d)
// }

// func readRedfishBios(service *gofish.Service, d *schema.ResourceData) diag.Diagnostics {
// 	var diags diag.Diagnostics

// 	systems, err := service.Systems()
// 	if err != nil {
// 		return diag.Errorf("error fetching computer systems collection: %s", err)

// 	}

// 	bios, err := systems[0].Bios()
// 	if err != nil {
// 		return diag.Errorf("error fetching bios: %s", err)
// 	}

// 	// TODO: BIOS Attributes' values might be any of several types.
// 	// terraform-sdk currently does not support a map with different
// 	// value types. So we will convert int and float values to string
// 	attributes := make(map[string]string)

// 	// copy from the BIOS attributes to the new bios attributes map
// 	for key, value := range bios.Attributes {
// 		if attr_val, ok := value.(string); ok {
// 			attributes[key] = attr_val
// 		} else {
// 			attributes[key] = fmt.Sprintf("%v", value)
// 		}
// 	}

// 	if err := d.Set("odata_id", bios.ODataID); err != nil {
// 		return diag.Errorf("error setting bios OData ID: %s", err)
// 	}

// 	if err := d.Set("id", bios.ID); err != nil {
// 		return diag.Errorf("error setting bios ID: %s", err)
// 	}

// 	if err := d.Set("attributes", attributes); err != nil {
// 		return diag.Errorf("error setting bios attributes: %s", err)
// 	}

// 	// Set the ID to the redfish endpoint + bios @odata.id
// 	serverConfig := d.Get("redfish_server").([]interface{})
// 	endpoint := serverConfig[0].(map[string]interface{})["endpoint"].(string)
// 	biosResourceId := endpoint + bios.ODataID
// 	d.SetId(biosResourceId)

// 	return diags
// }
