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

const (
	// HTTP represents the constant value for HTTP protocol
	HTTP = "HTTP"
	// HTTPS represents the constant value for HTTPS protocol
	HTTPS = "HTTPS"

	// defaultJobTimeout represents the default timeout value for the job in seconds
	defaultJobTimeout int64 = 1200
	// intervalJobCheckTime is the interval time to check the job status in seconds
	intervalJobCheckTime int64 = 10
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
		"id": schema.StringAttribute{
			MarkdownDescription: "ID of the Export SCP data-source",
			Description:         "ID of the Export SCP data-source",
			Computed:            true,
		},
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
			Optional:            true,
		},
		"file_content": schema.StringAttribute{
			MarkdownDescription: "File Content",
			Description:         "File Content",
			Computed:            true,
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
			Validators: []validator.String{
				stringvalidator.OneOf([]string{
					string("Enabled"),
					string("Disabled"),
				}...),
			},
		},
		"proxy_type": schema.StringAttribute{
			MarkdownDescription: "proxy_type",
			Description:         "proxy_type",
			Optional:            true,
			Validators: []validator.String{
				stringvalidator.OneOf([]string{
					string(HTTP),
					string("SOCKS4"),
				}...),
			},
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
					string(HTTP),
					string(HTTPS),
					string("LOCAL"),
				}...),
			},
		},
		"target": schema.ListAttribute{
			MarkdownDescription: "Filter configuration by target",
			Description:         "Filter configuration by target",
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

// setDefaultValues to set default values for scp export attributes
func setDefaultValues(plan *models.RedfishScpExport) {
	if plan.FileName.IsNull() {
		plan.FileName = types.StringValue("export_scp")
	}
	if plan.ExportFormat.IsNull() {
		plan.ExportFormat = types.StringValue("XML")
	}
	if plan.ExportUse.IsNull() {
		plan.ExportUse = types.StringValue("Default")
	}
	if plan.IncludeInExport.IsNull() {
		plan.IncludeInExport = types.ListValueMust(types.StringType, []attr.Value{types.StringValue("Default")})
	}
	if plan.Target.IsNull() {
		plan.Target = types.ListValueMust(types.StringType, []attr.Value{types.StringValue("ALL")})
	}
	if plan.ProxySupport.IsNull() {
		plan.ProxySupport = types.StringValue("Disabled")
	}
	if plan.ShareType.ValueString() == HTTP || plan.ShareType.ValueString() == HTTPS {
		if plan.PortNumber.IsNull() {
			if plan.ShareType.ValueString() == HTTP {
				plan.PortNumber = types.StringValue("80")
			} else {
				plan.PortNumber = types.StringValue("443")
			}
		}
		if plan.ProxySupport.ValueString() == "Enabled" {
			if plan.ProxyPort.IsNull() {
				plan.ProxyPort = types.StringValue("80")
			}
			if plan.ProxyType.IsNull() {
				plan.ProxyType = types.StringValue(HTTP)
			}
		}
	}
}

func validateShareType(plan *models.RedfishScpExport, resp *datasource.ReadResponse) {
	shareType := plan.ShareType.ValueString()
	validationError := "The validation encountered an error."
	if shareType == "NFS" {
		if plan.IPAddress.IsNull() || plan.ShareName.IsNull() {
			resp.Diagnostics.AddError(
				validationError,
				"When configuring the share type as ‘NFS’, it is essential to provide both the IP address and the share name.")
			return
		}
	} else if shareType == "CIFS" {
		if plan.IPAddress.IsNull() || plan.ShareName.IsNull() || plan.Username.IsNull() || plan.Password.IsNull() {
			resp.Diagnostics.AddError(
				validationError,
				"When configuring the share type as CIFS, it is essential to provide the IP address, share name, username and password.")
			return
		}
	} else if shareType == HTTP || shareType == HTTPS {
		if plan.IPAddress.IsNull() {
			resp.Diagnostics.AddError(
				validationError,
				fmt.Sprintf(
					"When configuring the share type as %s, it is essential to provide the IP address.",
					shareType))
			return
		}
		if plan.ProxySupport.ValueString() == "Enabled" {
			if plan.ProxyServer.IsNull() {
				resp.Diagnostics.AddError(
					validationError,
					fmt.Sprintf(
						"When configuring the share type as %s and Proxy Support is enabled, it is essential to provide the proxy server.",
						shareType))
				return
			}
		}
	}
}

// Read implements datasource.DataSource
func (g *SCPExportDatasource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var plan models.RedfishScpExport
	diags := req.Config.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	// set default value
	setDefaultValues(&plan)
	if plan.ShareType.IsNull() {
		plan.ShareType = types.StringValue("LOCAL")
	} else {
		validateShareType(&plan, resp)
		if resp.Diagnostics.HasError() {
			return
		}
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	service, err := NewConfig(g.p, &plan.RedfishServer)
	if err != nil {
		resp.Diagnostics.AddError("service error", err.Error())
		return
	}
	managers, err := service.Managers()
	if err != nil {
		resp.Diagnostics.AddError("error while retrieving managers", err.Error())
		return
	}

	// Get OEM
	dellManager, err := dell.Manager(managers[0])
	if err != nil {
		resp.Diagnostics.AddError("error while retrieving dell manager", err.Error())
		return
	}
	exportURL := dellManager.Actions.ExportSystemConfigurationTarget
	var includeInExport, target []string
	resp.Diagnostics.Append(plan.IncludeInExport.ElementsAs(ctx, &includeInExport, true)...)
	resp.Diagnostics.Append(plan.Target.ElementsAs(ctx, &target, true)...)
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
			Target:                   target,
			Username:                 plan.Username.ValueString(),
			Workgroup:                plan.Workgroup.ValueString(),
		},
	}
	response, err := service.GetClient().Post(exportURL, payload)
	if err != nil {
		resp.Diagnostics.AddError("error during export", err.Error())
		return
	}
	if location, err := response.Location(); err == nil {
		// tflog.Trace(r.ctx, "[DEBUG] BIOS configuration job uri: "+location.String())
		taskURI := location.EscapedPath()
		if plan.ShareType.ValueString() == "LOCAL" {
			fileContent, err := common.GetJobAttachment(service, taskURI, intervalJobCheckTime, defaultJobTimeout)
			if err != nil {
				resp.Diagnostics.AddError(
					"error waiting for SCP Export monitor task to be completed",
					err.Error())
			}
			plan.FileContent = types.StringValue(string(fileContent))
		} else {
			err = common.WaitForTaskToFinish(service, taskURI, intervalJobCheckTime, defaultJobTimeout)
			if err != nil {
				resp.Diagnostics.AddError(
					"error waiting for SCP Export monitor task to be completed",
					err.Error())
			}
		}
	}
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}
