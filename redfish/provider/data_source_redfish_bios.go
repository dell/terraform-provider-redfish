/*
Copyright (c) 2023-2024 Dell Inc., or its subsidiaries. All Rights Reserved.

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
	}
}

// Read implements datasource.DataSource
func (g *BiosDatasource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var plan models.BiosDatasource
	diags := req.Config.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
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
		if attrVal, ok := value.(string); ok {
			attributes[key] = types.StringValue(attrVal)
		} else {
			attributes[key] = types.StringValue(fmt.Sprintf("%v", value))
		}
	}

	d.OdataID = types.StringValue(bios.ODataID)
	d.ID = types.StringValue(bios.ID)
	d.Attributes, diags = types.MapValue(types.StringType, attributes)

	return d, diags
}
