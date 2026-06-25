package provider

import (
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// storageVolumeDataSourceModel describes the data source data model.
type storageVolumeDataSourceModel struct {
	RedfishServer types.Object `tfsdk:"redfish_server"`
	Filter        types.Object `tfsdk:"filter"`
	
	// Output attributes
	ID            types.String `tfsdk:"id"`
	Name          types.String `tfsdk:"name"`
	CapacityBytes types.Int64  `tfsdk:"capacity_bytes"`
	RAIDType      types.String `tfsdk:"raid_type"`
	Status        types.String `tfsdk:"status"`
	Encrypted     types.Bool   `tfsdk:"encrypted"`
	ControllerID  types.String `tfsdk:"controller_id"`
	CreationTime  types.String `tfsdk:"creation_time"`
	PhysicalDisks types.List   `tfsdk:"physical_disks"`
	OptimumIOSize types.Int64  `tfsdk:"optimum_io_size"`
	BlockSize     types.Int64  `tfsdk:"block_size"`
}

// filterModel describes the filter block.
type filterModel struct {
	Name types.String `tfsdk:"name"`
	ID   types.String `tfsdk:"id"`
}

// storageVolumeDataSourceSchema returns the schema for the data source.
func storageVolumeDataSourceSchema() schema.Schema {
	return schema.Schema{
		MarkdownDescription: "Retrieves information about existing storage volumes from PowerEdge servers via iDRAC Redfish API.",
		Description:         "Retrieves information about existing storage volumes from PowerEdge servers via iDRAC Redfish API.",

		Attributes: map[string]schema.Attribute{
			"redfish_server": schema.SingleNestedAttribute{
				MarkdownDescription: "iDRAC connection configuration",
				Description:         "iDRAC connection configuration",
				Required:            true,
				Attributes: map[string]schema.Attribute{
					"endpoint": schema.StringAttribute{
						MarkdownDescription: "iDRAC endpoint URL (e.g., https://192.168.1.100)",
						Description:         "iDRAC endpoint URL",
						Required:            true,
					},
					"username": schema.StringAttribute{
						MarkdownDescription: "iDRAC username",
						Description:         "iDRAC username",
						Required:            true,
					},
					"password": schema.StringAttribute{
						MarkdownDescription: "iDRAC password",
						Description:         "iDRAC password",
						Required:            true,
						Sensitive:           true,
					},
					"ssl_insecure": schema.BoolAttribute{
						MarkdownDescription: "Skip SSL certificate verification (default: false)",
						Description:         "Skip SSL certificate verification",
						Optional:            true,
					},
				},
			},
			"filter": schema.SingleNestedAttribute{
				MarkdownDescription: "Volume filter criteria (at least one required)",
				Description:         "Volume filter criteria",
				Required:            true,
				Attributes: map[string]schema.Attribute{
					"name": schema.StringAttribute{
						MarkdownDescription: "Volume name to query",
						Description:         "Volume name to query",
						Optional:            true,
					},
					"id": schema.StringAttribute{
						MarkdownDescription: "Volume ID to query",
						Description:         "Volume ID to query",
						Optional:            true,
					},
				},
			},
			"id": schema.StringAttribute{
				MarkdownDescription: "Volume ID",
				Description:         "Volume ID",
				Computed:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Volume name",
				Description:         "Volume name",
				Computed:            true,
			},
			"capacity_bytes": schema.Int64Attribute{
				MarkdownDescription: "Volume capacity in bytes",
				Description:         "Volume capacity in bytes",
				Computed:            true,
			},
			"raid_type": schema.StringAttribute{
				MarkdownDescription: "RAID type (e.g., RAID0, RAID1, RAID5, RAID6, RAID10, RAID50, RAID60)",
				Description:         "RAID type",
				Computed:            true,
			},
			"status": schema.StringAttribute{
				MarkdownDescription: "Volume status (e.g., Available, InUse, Degraded, Failed)",
				Description:         "Volume status",
				Computed:            true,
			},
			"encrypted": schema.BoolAttribute{
				MarkdownDescription: "Whether volume is encrypted",
				Description:         "Whether volume is encrypted",
				Computed:            true,
			},
			"controller_id": schema.StringAttribute{
				MarkdownDescription: "Storage controller ID",
				Description:         "Storage controller ID",
				Computed:            true,
			},
			"creation_time": schema.StringAttribute{
				MarkdownDescription: "Volume creation timestamp (ISO 8601 format)",
				Description:         "Volume creation timestamp",
				Computed:            true,
			},
			"physical_disks": schema.ListAttribute{
				MarkdownDescription: "Physical disk IDs that comprise the volume",
				Description:         "Physical disk IDs",
				Computed:            true,
				ElementType:         types.StringType,
			},
			"optimum_io_size": schema.Int64Attribute{
				MarkdownDescription: "Optimum I/O size in bytes",
				Description:         "Optimum I/O size in bytes",
				Computed:            true,
			},
			"block_size": schema.Int64Attribute{
				MarkdownDescription: "Block size in bytes",
				Description:         "Block size in bytes",
				Computed:            true,
			},
		},
	}
}
