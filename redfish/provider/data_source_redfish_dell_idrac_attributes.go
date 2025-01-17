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
	"terraform-provider-redfish/redfish/helper"
	"terraform-provider-redfish/redfish/models"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ datasource.DataSource              = &DellIdracAttributesDatasource{}
	_ datasource.DataSourceWithConfigure = &DellIdracAttributesDatasource{}
)

// NewDellIdracAttributesDatasource is new datasource for idrac attributes
func NewDellIdracAttributesDatasource() datasource.DataSource {
	return &DellIdracAttributesDatasource{}
}

// DellIdracAttributesDatasource to construct datasource
type DellIdracAttributesDatasource struct {
	p *redfishProvider
}

// Configure implements datasource.DataSourceWithConfigure
func (g *DellIdracAttributesDatasource) Configure(_ context.Context, req datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	g.p = req.ProviderData.(*redfishProvider)
}

// Metadata implements datasource.DataSource
func (*DellIdracAttributesDatasource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "dell_idrac_attributes"
}

// Schema implements datasource.DataSource
func (*DellIdracAttributesDatasource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "This Terraform datasource is used to query existing iDRAC configuration." +
			" The information fetched from this block can be further used for resource block.",
		Description: "This Terraform datasource is used to query existing iDRAC configuration." +
			" The information fetched from this block can be further used for resource block.",
		Attributes: DellIdracAttributesSchemaDatasource(),
		Blocks:     RedfishServerDatasourceBlockMap(),
	}
}

// DellIdracAttributesSchemaDatasource to define the idrac attribute schema
func DellIdracAttributesSchemaDatasource() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"id": schema.StringAttribute{
			MarkdownDescription: "ID of the iDRAC attributes resource",
			Description:         "ID of the iDRAC attributes resource",
			Computed:            true,
		},
		"attributes": schema.MapAttribute{
			MarkdownDescription: "iDRAC attributes. " +
				"To check allowed attributes please either use the datasource for dell idrac attributes or query " +
				"/redfish/v1/Managers/iDRAC.Embedded.1/Oem/Dell/DellAttributes/iDRAC.Embedded.1. " +
				"To get allowed values for those attributes, check " +
				"/redfish/v1/Registries/ManagerAttributeRegistry/ManagerAttributeRegistry.v1_0_0.json from a Redfish Instance",
			Description: "iDRAC attributes. " +
				"To check allowed attributes please either use the datasource for dell idrac attributes or query " +
				"/redfish/v1/Managers/iDRAC.Embedded.1/Oem/Dell/DellAttributes/iDRAC.Embedded.1. " +
				"To get allowed values for those attributes, check " +
				"/redfish/v1/Registries/ManagerAttributeRegistry/ManagerAttributeRegistry.v1_0_0.json from a Redfish Instance",
			ElementType: types.StringType,
			Computed:    true,
		},
	}
}

// Read implements datasource.DataSource
func (g *DellIdracAttributesDatasource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state models.DellIdracAttributes
	diags := req.Config.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if state.ID.IsUnknown() {
		state.ID = types.StringValue("placeholder")
	}
	api, err := NewConfig(g.p, &state.RedfishServer)
	if err != nil {
		resp.Diagnostics.AddError("service error", err.Error())
		return
	}
	service := api.Service
	defer api.Logout()
	diags = helper.ReadDatasourceRedfishDellIdracAttributes(service, &state)
	resp.Diagnostics.Append(diags...)
	if diags.HasError() {
		return
	}

	resp.Diagnostics.Append(diags...)
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}
