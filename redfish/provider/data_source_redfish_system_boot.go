package provider

import (
	"context"
	"terraform-provider-redfish/redfish/models"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stmcginnis/gofish"
	"github.com/stmcginnis/gofish/redfish"
)

var (
	_ datasource.DataSource              = &SystemBootDatasource{}
	_ datasource.DataSourceWithConfigure = &SystemBootDatasource{}
)

// NewSystemBootDatasource is new datasource for group devices
func NewSystemBootDatasource() datasource.DataSource {
	return &SystemBootDatasource{}
}

// SystemBootDatasource to construct datasource
type SystemBootDatasource struct {
	p *redfishProvider
}

// Configure implements datasource.DataSourceWithConfigure
func (g *SystemBootDatasource) Configure(_ context.Context, req datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	g.p = req.ProviderData.(*redfishProvider)
}

// Metadata implements datasource.DataSource
func (*SystemBootDatasource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "system_boot"
}

// Schema implements datasource.DataSource
func (*SystemBootDatasource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Data source to fetch Firmware Inventory details via RedFish.",
		Attributes:          SystemBootDatasourceSchema(),
		Blocks:              RedfishServerDatasourceBlockMap(),
	}
}

// SystemBootDatasourceSchema to define the system boot datasource schema
func SystemBootDatasourceSchema() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"id": schema.StringAttribute{
			MarkdownDescription: "Resource ID of the computer system. If not provided, the first system resource is used",
			Description:         "Resource ID of the computer system. If not provided, the first system resource is used",
			Optional:            true,
		},
		"boot_order": schema.ListAttribute{
			MarkdownDescription: "An array of BootOptionReference strings that represent the persistent boot order for this computer system",
			Description:         "An array of BootOptionReference strings that represent the persistent boot order for this computer system",
			Computed:            true,
			ElementType:         types.StringType,
		},
		"boot_source_override_enabled": schema.StringAttribute{
			MarkdownDescription: "The state of the boot source override feature",
			Description:         "The state of the boot source override feature",
			Computed:            true,
		},
		"boot_source_override_mode": schema.StringAttribute{
			MarkdownDescription: "The BIOS boot mode to use when the system boots from the BootSourceOverrideTarget boot source",
			Description:         "The BIOS boot mode to use when the system boots from the BootSourceOverrideTarget boot source",
			Computed:            true,
		},
		"boot_source_override_target": schema.StringAttribute{
			MarkdownDescription: "Current boot source to use at next boot instead of the normal boot device, if BootSourceOverrideEnabled is true",
			Description:         "Current boot source to use at next boot instead of the normal boot device, if BootSourceOverrideEnabled is true",
			Computed:            true,
		},
		"uefi_target_boot_source_override": schema.StringAttribute{
			MarkdownDescription: "The UEFI device path of the device from which to boot when BootSourceOverrideTarget is UefiTarget",
			Description:         "The UEFI device path of the device from which to boot when BootSourceOverrideTarget is UefiTarget",
			Computed:            true,
		},
	}
}

// Read implements datasource.DataSource
func (g *SystemBootDatasource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var plan models.SystemBootDataSource
	diags := req.Config.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	service, err := NewConfig(g.p, &plan.RedfishServer)
	if err != nil {
		resp.Diagnostics.AddError("service error", err.Error())
		return
	}
	state, diags := readRedfishSystemBoot(service, &plan)
	if err != nil {
		diags.AddError("failed to fetch firmware inventory details", err.Error())
	}
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func readRedfishSystemBoot(service *gofish.Service, d *models.SystemBootDataSource) (*models.SystemBootDataSource, diag.Diagnostics) {
	var diags diag.Diagnostics

	systems, err := service.Systems()
	if err != nil {
		diags.AddError("Error when retrieving systems", err.Error())
		return nil, diags
	}

	// get the boot resource
	var computerSystem *redfish.ComputerSystem
	var boot redfish.Boot
	if d.ResourceID.ValueString() != "" {
		for key := range systems {
			if systems[key].ID == d.ResourceID.ValueString() {
				computerSystem = systems[key]
				boot = systems[key].Boot
				break
			}
		}

		if computerSystem == nil {
			diags.AddError("Could not find a ComputerSystem", "")
			return nil, diags
		}
	} else {
		// use the first system resource in the collection if resource
		// ID is not provided
		computerSystem = systems[0]
		boot = systems[0].Boot
	}

	bootOrder := []attr.Value{}
	for _, bootOptionReference := range boot.BootOrder {
		bootOrder = append(bootOrder, types.StringValue(string(bootOptionReference)))
	}

	d.BootOrder, diags = types.ListValue(types.StringType, bootOrder)
	d.BootSourceOverrideEnabled = types.StringValue(string(boot.BootSourceOverrideEnabled))
	d.BootSourceOverrideMode = types.StringValue(string(boot.BootSourceOverrideMode))
	d.BootSourceOverrideTarget = types.StringValue(string(boot.BootSourceOverrideTarget))
	d.UefiTargetBootSourceOverride = types.StringValue(string(boot.UefiTargetBootSourceOverride))
	d.ResourceID = types.StringValue(computerSystem.ODataID)

	return d, diags
}

// import (
// 	"context"

// 	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
// 	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
// 	"github.com/stmcginnis/gofish"
// 	"github.com/stmcginnis/gofish/redfish"
// )

// func dataSourceRedfishSystemBoot() *schema.Resource {
// 	return &schema.Resource{
// 		ReadContext: dataSourceRedfishSystemBootRead,
// 		Schema:      getDataSourceRedfishSystemBootSchema(),
// 	}
// }

// func getDataSourceRedfishSystemBootSchema() map[string]*schema.Schema {
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
// 		"resource_id": {
// 			Type:        schema.TypeString,
// 			Optional:    true,
// 			Description: "Resource ID of the computer system resource.",
// 		},
// 		"boot_order": {
// 			Type:        schema.TypeList,
// 			Computed:    true,
// 			Description: "An array of BootOptionReference strings that represent the persistent boot order for this computer system",
// 			Elem: &schema.Schema{
// 				Type: schema.TypeString,
// 			},
// 		},
// 		"boot_source_override_enabled": {
// 			Type:        schema.TypeString,
// 			Computed:    true,
// 			Description: "The state of the boot source override feature",
// 		},
// 		"boot_source_override_mode": {
// 			Type:        schema.TypeString,
// 			Computed:    true,
// 			Description: "The BIOS boot mode to use when the system boots from the BootSourceOverrideTarget boot source",
// 		},
// 		"boot_source_override_target": {
// 			Type:        schema.TypeString,
// 			Computed:    true,
// 			Description: "The current boot source to use at the next boot instead of the normal boot device, if BootSourceOverrideEnabled is true",
// 		},
// 		"uefi_target_boot_source_override": {
// 			Type:        schema.TypeString,
// 			Computed:    true,
// 			Description: "The UEFI device path of the device from which to boot when BootSourceOverrideTarget is UefiTarget",
// 		},
// 	}
// }

// func dataSourceRedfishSystemBootRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
// 	service, err := NewConfig(m.(*schema.ResourceData), d)
// 	if err != nil {
// 		return diag.Errorf(err.Error())
// 	}

// 	return readRedfishSystemBoot(service, d)
// }
