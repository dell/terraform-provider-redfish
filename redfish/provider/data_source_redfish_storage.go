package provider

import (
	"context"
	"fmt"
	"terraform-provider-redfish/redfish/models"
	"time"

	"github.com/stmcginnis/gofish"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ datasource.DataSource              = &StorageDatasource{}
	_ datasource.DataSourceWithConfigure = &StorageDatasource{}
)

// NewStorageDatasource is new datasource for storage
func NewStorageDatasource() datasource.DataSource {
	return &StorageDatasource{}
}

// StorageDatasource to construct datasource
type StorageDatasource struct {
	p       *redfishProvider
	ctx     context.Context
	service *gofish.Service
}

// Configure implements datasource.DataSourceWithConfigure
func (g *StorageDatasource) Configure(_ context.Context, req datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	g.p = req.ProviderData.(*redfishProvider)
}

// Metadata implements datasource.DataSource
func (*StorageDatasource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "storage"
}

// Schema implements datasource.DataSource
func (*StorageDatasource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "This Terraform datasource is used to query existing storage volume details." +
			" The information fetched from this block can be further used for resource block.",
		Description: "This Terraform datasource is used to query existing storage volume details." +
			" The information fetched from this block can be further used for resource block.",
		Attributes: StorageDatasourceSchema(),
		Blocks:     RedfishServerDatasourceBlockMap(),
	}
}

// StorageDatasourceSchema to define the storage data-source schema
func StorageDatasourceSchema() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"id": schema.StringAttribute{
			MarkdownDescription: "ID of the storage data-source",
			Description:         "ID of the storage data-source",
			Computed:            true,
		},
		"storage": schema.ListNestedAttribute{
			MarkdownDescription: "List of storage controllers",
			Description:         "List of storage controllers",
			NestedObject: schema.NestedAttributeObject{
				Attributes: StorageControllerSchema(),
			},
			Computed: true,
		},
	}
}

// StorageControllerSchema to define the storage data-source schema
func StorageControllerSchema() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"storage_controller_id": schema.StringAttribute{
			MarkdownDescription: "ID of the storage controller",
			Description:         "ID of the storage controller",
			Computed:            true,
		},
		"drives": schema.ListAttribute{
			MarkdownDescription: "List of drives on the storage controller",
			Description:         "List of drives on the storage controller",
			Computed:            true,
			ElementType:         types.StringType,
		},
	}
}

// Read implements datasource.DataSource
func (g *StorageDatasource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var plan models.StorageDatasource
	diags := req.Config.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	service, err := NewConfig(g.p, &plan.RedfishServer)
	if err != nil {
		resp.Diagnostics.AddError("service error", err.Error())
		return
	}
	g.ctx = ctx
	g.service = service
	state, diags := g.readDatasourceRedfishStorage(plan)
	resp.Diagnostics.Append(diags...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (g *StorageDatasource) readDatasourceRedfishStorage(d models.StorageDatasource) (models.StorageDatasource, diag.Diagnostics) {
	var diags diag.Diagnostics

	// write the current time as ID
	d.ID = types.StringValue(fmt.Sprintf("%d", time.Now().Unix()))

	systems, err := g.service.Systems()
	if err != nil {
		diags.AddError("Error fetching computer systems collection", err.Error())
		return d, diags
	}

	storage, err := systems[0].Storage()
	if err != nil {
		diags.AddError("Error fetching storage", err.Error())
		return d, diags
	}

	d.Storages = make([]models.StorageControllerData, 0)
	for _, s := range storage {
		mToAdd := models.StorageControllerData{
			ID: types.StringValue(s.ID),
		}
		drives, err := s.Drives()
		if err != nil {
			diags.AddError(fmt.Sprintf("Error when retrieving drives: %s", s.ID), err.Error())
			continue
		}

		driveNames := make([]types.String, 0)
		for _, d := range drives {
			driveNames = append(driveNames, types.StringValue(d.Name))
		}
		mToAdd.Drives = driveNames
		d.Storages = append(d.Storages, mToAdd)
	}

	return d, diags
}
