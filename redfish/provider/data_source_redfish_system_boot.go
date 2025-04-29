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
	"terraform-provider-redfish/redfish/models"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
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
		MarkdownDescription: "Data source to fetch System Boot details via RedFish." +
			" The information fetched from this block can be further used for resource block.",
		Description: "Data source to fetch System Boot details via RedFish." +
			" The information fetched from this block can be further used for resource block.",
		Attributes: SystemBootDatasourceSchema(),
		Blocks:     RedfishServerDatasourceBlockMap(),
	}
}

// SystemBootDatasourceSchema to define the system boot datasource schema
func SystemBootDatasourceSchema() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"id": schema.StringAttribute{
			MarkdownDescription: "Resource ID of the computer system used.",
			Description:         "Resource ID of the computer system used.",
			Computed:            true,
		},
		"system_id": schema.StringAttribute{
			MarkdownDescription: "System ID of the computer system. If not provided, the first system resource is used",
			Description:         "System ID of the computer system. If not provided, the first system resource is used",
			Optional:            true,
			Computed:            true,
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
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
	api, err := NewConfig(g.p, &plan.RedfishServer)
	if err != nil {
		resp.Diagnostics.AddError("service error", err.Error())
		return
	}
	service := api.Service
	defer api.Logout()
	state, diags := readRedfishSystemBoot(service, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func readRedfishSystemBoot(service *gofish.Service, d models.SystemBootDataSource) (models.SystemBootDataSource, diag.Diagnostics) {
	var diags diag.Diagnostics
	// get the boot resource
	var computerSystem *redfish.ComputerSystem
	var boot redfish.Boot

	system, err := getSystemResource(service, d.SystemID.ValueString())
	if err != nil {
		diags.AddError("Error fetching computer system", err.Error())
		return d, diags
	}
	computerSystem = system

	if computerSystem != nil {
		boot = computerSystem.Boot
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
	d.SystemID = types.StringValue(computerSystem.ID)
	d.ID = types.StringValue(computerSystem.ODataID)

	return d, diags
}
