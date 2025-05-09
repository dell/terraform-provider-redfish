/*
Copyright (c) 2023-2025 Dell Inc., or its subsidiaries. All Rights Reserved.

Licensed under the Mozilla Public License Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://mozilla.org/MPL/2.0/


Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package provider

import (
	"context"
	"fmt"
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
	_ datasource.DataSource              = &BiosDatasource{}
	_ datasource.DataSourceWithConfigure = &BiosDatasource{}
)

// NewBiosDatasource is new datasource for bios
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
	resp.TypeName = req.ProviderTypeName + "bios"
}

// Schema implements datasource.DataSource
func (*BiosDatasource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "This Terraform datasource is used to query existing Bios configuration." +
			" The information fetched from this block can be further used for resource block.",
		Description: "This Terraform datasource is used to query existing Bios configuration." +
			" The information fetched from this block can be further used for resource block.",
		Attributes: BiosDatasourceSchema(),
		Blocks:     RedfishServerDatasourceBlockMap(),
	}
}

// BiosDatasourceSchema to define the bios data-source schema
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
		"attributes": schema.MapAttribute{
			MarkdownDescription: "BIOS attributes.",
			Description:         "BIOS attributes.",
			ElementType:         types.StringType,
			Computed:            true,
		},
		"boot_options": schema.ListNestedAttribute{
			MarkdownDescription: "List of BIOS boot options.",
			Description:         "List of BIOS boot options.",
			NestedObject: schema.NestedAttributeObject{
				Attributes: BootOptionsSchema(),
			},
			Computed: true,
		},
		"system_id": schema.StringAttribute{
			MarkdownDescription: "System ID of the system",
			Description:         "System ID of the system",
			Computed:            true,
			Optional:            true,
		},
	}
}

// Read implements datasource.DataSource
func (g *BiosDatasource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var plan models.BiosDatasource
	diags := req.Config.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	api, err := NewConfig(g.p, &plan.RedfishServer)
	if err != nil {
		resp.Diagnostics.AddError("service error", err.Error())
		return
	}
	defer api.Logout()
	g.ctx = ctx
	g.service = api.Service
	state, diags := g.readDatasourceRedfishBios(plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (g *BiosDatasource) readDatasourceRedfishBios(d models.BiosDatasource) (models.BiosDatasource, diag.Diagnostics) {
	var diags diag.Diagnostics

	system, err := getSystemResource(g.service, d.SystemID.ValueString())
	if err != nil {
		diags.AddError("Error fetching computer systems collection", err.Error())
		return d, diags
	}

	if system == nil {
		diags.AddError("System not found", "System with ID "+d.SystemID.ValueString()+" not found")
		return d, diags
	}

	bios, err := system.Bios()
	if err != nil {
		diags.AddError("Error fetching bios", err.Error())
		return d, diags
	}

	if bios == nil {
		diags.AddError("BIOS not found", "BIOS not found for system with ID "+d.SystemID.ValueString())
		return d, diags
	}

	bootOptions, err := system.BootOptions()
	if err != nil {
		diags.AddError("Error fetching boot options", err.Error())
		return d, diags
	}

	// TODO: BIOS Attributes' values might be any of several types.
	// terraform-sdk currently does not support a map with different
	// value types. So we will convert int and float values to string
	attributes := make(map[string]attr.Value)

	// copy from the BIOS attributes to the new bios attributes map
	if bios.Attributes != nil {
		for key, value := range bios.Attributes {
			if attrVal, ok := value.(string); ok {
				attributes[key] = types.StringValue(attrVal)
			} else {
				attributes[key] = types.StringValue(fmt.Sprintf("%v", value))
			}
		}
	}

	d.OdataID = types.StringValue(bios.ODataID)
	d.ID = types.StringValue(bios.ID)
	d.Attributes, diags = types.MapValue(types.StringType, attributes)

	bootOptionsList := make([]models.BiosBootOptions, 0)
	for _, bootOption := range bootOptions {
		bootOptionsList = append(bootOptionsList, newBootOption(bootOption))
	}
	d.BootOptions = bootOptionsList

	return d, diags
}

// newBootOption converts client.BiosBootOptions to models.BiosBootOptions
func newBootOption(input *redfish.BootOption) models.BiosBootOptions {
	return models.BiosBootOptions{
		BootOptionEnabled:   types.BoolValue(input.BootOptionEnabled),
		BootOptionReference: types.StringValue(input.BootOptionReference),
		DisplayName:         types.StringValue(input.DisplayName),
		ID:                  types.StringValue(input.ID),
		Name:                types.StringValue(input.Name),
		UefiDevicePath:      types.StringValue(input.UefiDevicePath),
	}
}

// BootOptionsSchema is a function that returns the schema for Boot Options
func BootOptionsSchema() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"boot_option_enabled": schema.BoolAttribute{
			MarkdownDescription: "Enable or disable the boot device.",
			Description:         "Whether the boot device is enabled or not.",
			Computed:            true,
		},
		"boot_option_reference": schema.StringAttribute{
			MarkdownDescription: "FQDD of the boot device.",
			Description:         "FQDD of the boot device.",
			Computed:            true,
		},
		"display_name": schema.StringAttribute{
			MarkdownDescription: "Display name of the boot option",
			Description:         "Display name of the boot option",
			Computed:            true,
		},
		"id": schema.StringAttribute{
			MarkdownDescription: "ID of the boot option",
			Description:         "ID of the boot option",
			Computed:            true,
		},
		"name": schema.StringAttribute{
			MarkdownDescription: "Name of the boot option",
			Description:         "Name of the boot option",
			Computed:            true,
		},
		"uefi_device_path": schema.StringAttribute{
			MarkdownDescription: "Device path of the boot option",
			Description:         "Device path of the boot option",
			Computed:            true,
		},
	}
}
