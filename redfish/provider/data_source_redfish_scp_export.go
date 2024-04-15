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
	"strings"
	"terraform-provider-redfish/common"
	"terraform-provider-redfish/gofish/dell"
	"terraform-provider-redfish/redfish/models"

	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ datasource.DataSource              = &SCPExportDatasource{}
	_ datasource.DataSourceWithConfigure = &SCPExportDatasource{}
)

// NewSCPExportDatasource is new datasource for storage
func NewSCPExportDatasource() datasource.DataSource {
	return &SCPExportDatasource{}
}

// SCPExportDatasource to construct datasource
type SCPExportDatasource struct {
	p *redfishProvider
	// ctx     context.Context
	// service *gofish.Service
}

// Configure implements datasource.DataSourceWithConfigure
func (g *SCPExportDatasource) Configure(_ context.Context, req datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	g.p = req.ProviderData.(*redfishProvider)
}

// Metadata implements datasource.DataSource
func (*SCPExportDatasource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "scp_export"
}

// Schema implements datasource.DataSource
func (*SCPExportDatasource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "This Terraform datasource is used to query existing storage details from iDRAC." +
			" The information fetched from this block can be further used for resource block.",
		Description: "This Terraform datasource is used to query existing storage details from iDRAC." +
			" The information fetched from this block can be further used for resource block.",
		Attributes: RedfishScpExportSchema(),
		Blocks:     RedfishServerDatasourceBlockMap(),
	}
}

// RedfishScpExportSchema is a function that returns the schema for RedfishScpExport
func RedfishScpExportSchema() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"export_format": schema.StringAttribute{
			MarkdownDescription: "Specify the output file format.",
			Description:         "Specify the output file format.",
			Optional:            true,
			Validators: []validator.String{
				stringvalidator.OneOf([]string{
					string("JSON"),
					string("XML"),
				}...),
			},
		},
		"export_use": schema.StringAttribute{
			MarkdownDescription: "Specify the type of Server Configuration Profile (SCP) to be exported.",
			Description:         "Specify the type of Server Configuration Profile (SCP) to be exported.",
			Optional:            true,
			Validators: []validator.String{
				stringvalidator.OneOf([]string{
					string("Default"),
					string("Clone"),
					string("Replace"),
				}...),
			},
		},
		"include_in_export": schema.ListAttribute{
			MarkdownDescription: "Include In Export",
			Description:         "Include In Export",
			Optional:            true,
			ElementType:         types.StringType,
			Validators: []validator.List{
				listvalidator.SizeAtLeast(1),
				listvalidator.ValueStringsAre(stringvalidator.OneOf([]string{
					string("Default"),
					string("IncludeReadOnly"),
					string("IncludePasswordHashValues"),
					string("IncludeCustomTelemetry"),
				}...)),
			},
		},
		"filename": schema.StringAttribute{
			MarkdownDescription: "File Name",
			Description:         "File Name",
			Required:            true,
		},
		"ip_address": schema.StringAttribute{
			MarkdownDescription: "IPAddress",
			Description:         "IPAddress",
			Optional:            true,
		},
		"ignore_certificate_warning": schema.StringAttribute{
			MarkdownDescription: "Ignore Certificate Warning",
			Description:         "Ignore Certificate Warning",
			Optional:            true,
			Validators: []validator.String{
				stringvalidator.OneOf([]string{
					string("Enabled"),
					string("Disabled"),
				}...),
			},
		},
		"password": schema.StringAttribute{
			MarkdownDescription: "Password",
			Description:         "Password",
			Optional:            true,
		},
		"port_number": schema.StringAttribute{
			MarkdownDescription: "port_number",
			Description:         "port_number",
			Optional:            true,
		},
		"proxy_password": schema.StringAttribute{
			MarkdownDescription: "proxy_password",
			Description:         "proxy_password",
			Optional:            true,
		},
		"proxy_port": schema.StringAttribute{
			MarkdownDescription: "proxy_port",
			Description:         "proxy_port",
			Optional:            true,
		},
		"proxy_server": schema.StringAttribute{
			MarkdownDescription: "proxy_server",
			Description:         "proxy_server",
			Optional:            true,
		},
		"proxy_support": schema.StringAttribute{
			MarkdownDescription: "proxy_support",
			Description:         "proxy_support",
			Optional:            true,
		},
		"proxy_type": schema.StringAttribute{
			MarkdownDescription: "proxy_type",
			Description:         "proxy_type",
			Optional:            true,
		},
		"proxy_username": schema.StringAttribute{
			MarkdownDescription: "proxy_username",
			Description:         "proxy_username",
			Optional:            true,
		},
		"share_name": schema.StringAttribute{
			MarkdownDescription: "Share Name",
			Description:         "Share Name",
			Optional:            true,
		},
		"share_type": schema.StringAttribute{
			MarkdownDescription: "Share Type",
			Description:         "Share Type",
			Optional:            true,
			Validators: []validator.String{
				stringvalidator.OneOf([]string{
					string("NFS"),
					string("CIFS"),
					string("HTTP"),
					string("HTTPS"),
					string("LOCAL"),
				}...),
			},
		},
		"target": schema.ListAttribute{
			MarkdownDescription: "Include In Export",
			Description:         "Include In Export",
			Optional:            true,
			ElementType:         types.StringType,
			Validators: []validator.List{
				listvalidator.SizeAtLeast(1),
				listvalidator.ValueStringsAre(stringvalidator.OneOf([]string{
					string("ALL"),
					string("IDRAC"),
					string("BIOS"),
					string("NIC"),
					string("RAID"),
					string("FC"),
					string("InfiniBand"),
					string("SupportAssist"),
					string("EventFilters"),
					string("System"),
					string("LifecycleController"),
					string("AHCI"),
					string("PCIeSSD"),
				}...)),
			},
		},
		"username": schema.StringAttribute{
			MarkdownDescription: "User Name",
			Description:         "User Name",
			Optional:            true,
		},
		"workgroup": schema.StringAttribute{
			MarkdownDescription: "workgroup",
			Description:         "workgroup",
			Optional:            true,
		},
	}
}

// Read implements datasource.DataSource
func (g *SCPExportDatasource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var plan models.RedfishScpExport
	diags := req.Config.Get(ctx, &plan)
	// set default value
	if plan.ExportFormat.IsNull() {
		plan.ExportFormat = types.StringValue("XML")
	}
	if plan.ExportUse.IsNull() {
		plan.ExportUse = types.StringValue("Default")
	}
	if plan.IncludeInExport.IsNull() {
		plan.IncludeInExport = types.ListValueMust(types.StringType,[]attr.Value{types.StringValue("Default")})
	}
	if plan.Target.IsNull() {
		plan.Target = types.ListValueMust(types.StringType,[]attr.Value{types.StringValue("ALL")})
	}
	if plan.ShareType.IsNull(){
		plan.ShareType = types.StringValue("LOCAL")
	} else {
		if plan.ShareType.ValueString() == "NFS" {
			if plan.IPAddress.IsNull() || plan.ShareName.IsNull(){
				resp.Diagnostics.AddError("The validation encountered an error.", "When configuring the share type as ‘NFS’, it is essential to provide both the IP address and the share name.")
				return 
			}
		} else if plan.ShareType.ValueString() == "CIFS" {
			if plan.IPAddress.IsNull() || plan.ShareName.IsNull() || plan.Username.IsNull() || plan.Password.IsNull() {
				resp.Diagnostics.AddError("The validation encountered an error.", "When configuring the share type as CIFS, it is essential to provide the IP address, share name, username and password.")
				return
			}
		}
	}
	resp.Diagnostics.Append(diags...)
	service, err := NewConfig(g.p, &plan.RedfishServer)
	if err != nil {
		resp.Diagnostics.AddError("service error", err.Error())
		return
	}
	managers, err := service.Managers()
	if err != nil {
		diags.AddError("__add__", err.Error())
		return
	}

	// Get OEM
	dellManager, err := dell.Manager(managers[0])
	if err != nil {
		diags.AddError("__add__", err.Error())
		return
	}
	exportUrl := dellManager.Actions.ExportSystemConfigurationTarget
	// plan.ExportFormat = types.StringValue(exportUrl)
	// plan.FileName = types.StringValue(string(dellManager.OemActions))
	var includeInExport, Target []string
	diags.Append(plan.IncludeInExport.ElementsAs(ctx, &includeInExport, true)...)
	diags.Append(plan.Target.ElementsAs(ctx, &Target, true)...)
	file := plan.FileName.ValueString() + "." + strings.ToLower(plan.ExportFormat.ValueString())
	payload := models.SCPExport{
		ExportFormat:    plan.ExportFormat.ValueString(),
		ExportUse:       plan.ExportUse.ValueString(),
		IncludeInExport: includeInExport,
		ShareParameters: models.ShareParameters{
			FileName:                 file,
			IPAddress:                plan.IPAddress.ValueString(),
			IgnoreCertificateWarning: plan.IgnoreCertificateWarning.ValueString(),
			Password:                 plan.Password.ValueString(),
			PortNumber:               plan.PortNumber.ValueString(),
			ProxyPassword:            plan.ProxyPassword.ValueString(),
			ProxyPort:                plan.ProxyPort.ValueString(),
			ProxyServer:              plan.ProxyServer.ValueString(),
			ProxySupport:             plan.ProxySupport.ValueString(),
			ProxyType:                plan.ProxyType.ValueString(),
			ProxyUserName:            plan.ProxyUserName.ValueString(),
			ShareName:                plan.ShareName.ValueString(),
			ShareType:                plan.ShareType.ValueString(),
			Target:                   Target,
			Username:                 plan.Username.ValueString(),
			Workgroup:                plan.Workgroup.ValueString(),
		},
	}
	response, err := service.GetClient().Post(exportUrl, payload)
	if err != nil {
		diags.AddWarning(err.Error(), "")
		resp.Diagnostics.Append(diags...)
		return
	}
	if location, err := response.Location(); err == nil {
		// tflog.Trace(r.ctx, "[DEBUG] BIOS configuration job uri: "+location.String())
		taskURI := location.EscapedPath()
		if plan.ShareType.ValueString() == "LOCAL" {
			err = common.GetJobAttachment(service, taskURI, file, 30, 200)
			if err != nil {
				diags.AddError("error waiting for SCP Export monitor task to be completed", err.Error())
				return
			}
		}
		err = common.WaitForTaskToFinish(service, taskURI, 10, 200)
		if err != nil {
			diags.AddError("error waiting for SCP Export monitor task to be completed", err.Error())
			return
		}
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}
