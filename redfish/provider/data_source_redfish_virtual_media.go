package provider

import (
	"context"
	"log"
	"strconv"
	"terraform-provider-redfish/redfish/models"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stmcginnis/gofish"
)

var (
	_ datasource.DataSource              = &DellVirtualMediaDatasource{}
	_ datasource.DataSourceWithConfigure = &DellVirtualMediaDatasource{}
)

// NewDellVirtualMediaDatasource is new datasource for group devices
func NewDellVirtualMediaDatasource() datasource.DataSource {
	return &DellVirtualMediaDatasource{}
}

// DellVirtualMediaDatasource is struct for virtual media datasource
type DellVirtualMediaDatasource struct {
	p *redfishProvider
}

// Configure implements datasource.DataSourceWithConfigure
func (g *DellVirtualMediaDatasource) Configure(_ context.Context, req datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	g.p = req.ProviderData.(*redfishProvider)
}

// Metadata implements datasource.DataSource
func (*DellVirtualMediaDatasource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "virtual_media"
}

// Schema implements datasource.DataSource
func (*DellVirtualMediaDatasource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "datasource for virtual media.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "ID of the virtual media datasource",
				Description:         "ID of the virtual media datasource",
				Computed:            true,
			},
			"virtual_media": schema.ListNestedAttribute{
				MarkdownDescription: "List of virtual media available on this instance",
				Description:         "List of virtual media available on this instance",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"odata_id": schema.StringAttribute{
							Computed:    true,
							Description: "OData ID for the Virtual Media resource",
						},
						"id": schema.StringAttribute{
							Computed:    true,
							Description: "Id of the virtual media resource",
						},
					},
				},
			},
		},
		Blocks: RedfishServerDatasourceBlockMap(),
	}
}

// Read implements datasource.DataSource
func (g *DellVirtualMediaDatasource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state models.VirtualMediaDataSource
	diags := req.Config.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if state.ID.IsUnknown() {
		state.ID = types.StringValue("placeholder")
	}
	service, err := NewConfig(g.p, &state.RedfishServer)
	if err != nil {
		resp.Diagnostics.AddError("service error", err.Error())
		return
	}
	diags = readRedfishDellVirtualMediaCollection(service, &state)
	resp.Diagnostics.Append(diags...)
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

func readRedfishDellVirtualMediaCollection(service *gofish.Service, d *models.VirtualMediaDataSource) diag.Diagnostics {
	var diags diag.Diagnostics
	const intBase = 10
	// Get manager.Since this provider is thought to work with individual servers, should be only one.
	manager, err := service.Managers()
	if err != nil {
		diags.AddError("Error retrieving the managers:", err.Error())
		return diags
	}

	// Get virtual media
	dellvirtualMedia, err := manager[0].VirtualMedia()
	if err != nil {
		diags.AddError("Error retrieving the virtual media instances", err.Error())
		return diags
	}

	vms := make([]models.VirtualMediaData, 0)
	for _, v := range dellvirtualMedia {
		var vmToAdd models.VirtualMediaData
		log.Printf("Adding %s - %s", v.ODataID, v.ID)
		vmToAdd.OdataId = types.StringValue(v.ODataID)
		vmToAdd.Id = types.StringValue(v.ID)
		vms = append(vms, vmToAdd)
	}
	d.VirtualMediaData = vms
	d.ID = types.StringValue(strconv.FormatInt(time.Now().Unix(), intBase))
	return diags
}
