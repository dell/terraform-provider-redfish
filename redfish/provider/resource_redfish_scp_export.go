/*
Copyright (c) 2024 Dell Inc., or its subsidiaries. All Rights Reserved.

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
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"
	"terraform-provider-redfish/common"
	"terraform-provider-redfish/gofish/dell"
	"terraform-provider-redfish/redfish/models"

	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/stmcginnis/gofish"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource = &ScpExportResource{}
)

const (
	// defaultJobTimeout represents the default timeout value for the job in seconds
	defaultJobTimeout int64 = 3600
	// intervalJobCheckTime is the interval time to check the job status in seconds
	intervalJobCheckTime int64 = 10
)

// NewScpExportResource is a helper function to simplify the provider implementation.
func NewScpExportResource() resource.Resource {
	return &ScpExportResource{}
}

// ScpExportResource is the resource implementation.
type ScpExportResource struct {
	p *redfishProvider
}

// Configure implements resource.ResourceWithConfigure
func (r *ScpExportResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	r.p = req.ProviderData.(*redfishProvider)
}

// Metadata returns the resource type name.
func (*ScpExportResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "idrac_server_configuration_profile_export"
}

// Schema defines the schema for the resource.
func (*ScpExportResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Resource for managing iDRAC Server Configuration Profile export on iDRAC Server.",
		Version:             1,
		Attributes:          RedfishScpExportSchema(),
		Blocks:              RedfishServerResourceBlockMap(),
	}
}

// RedfishScpExportSchema defines the schema for the resource.
func RedfishScpExportSchema() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"id": schema.StringAttribute{
			MarkdownDescription: "ID of the export SCP resource",
			Description:         "ID of the export SCP resource",
			Computed:            true,
		},
		"file_content": schema.StringAttribute{
			MarkdownDescription: "File Content",
			Description:         "File Content",
			Computed:            true,
		},
		"export_format": schema.StringAttribute{
			MarkdownDescription: "Specify the output file format.",
			Description:         "Specify the output file format.",
			Optional:            true,
			Computed:            true,
			Default:             stringdefault.StaticString("XML"),
			Validators: []validator.String{
				stringvalidator.LengthAtLeast(1),
				stringvalidator.OneOf([]string{
					string("JSON"),
					string("XML"),
				}...),
			},
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
		"export_use": schema.StringAttribute{
			MarkdownDescription: "Specify the type of Server Configuration Profile (SCP) to be exported.",
			Description:         "Specify the type of Server Configuration Profile (SCP) to be exported.",
			Optional:            true,
			Computed:            true,
			Default:             stringdefault.StaticString("Default"),
			Validators: []validator.String{
				stringvalidator.LengthAtLeast(1),
				stringvalidator.OneOf([]string{
					string("Default"),
					string("Clone"),
					string("Replace"),
				}...),
			},
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
		"include_in_export": schema.ListAttribute{
			MarkdownDescription: "Include In Export",
			Description:         "Include In Export",
			Optional:            true,
			ElementType:         types.StringType,
			Computed:            true,
			Default: listdefault.StaticValue(
				types.ListValueMust(
					types.StringType,
					[]attr.Value{
						types.StringValue("Default"),
					},
				),
			),
			Validators: []validator.List{
				listvalidator.SizeAtLeast(1),
				listvalidator.ValueStringsAre(stringvalidator.OneOf([]string{
					string("Default"),
					string("IncludeReadOnly"),
					string("IncludePasswordHashValues"),
					string("IncludeCustomTelemetry"),
				}...)),
			},
			PlanModifiers: []planmodifier.List{
				listplanmodifier.RequiresReplace(),
			},
		},
		"share_parameters": schema.SingleNestedAttribute{
			MarkdownDescription: "Share Parameters",
			Description:         "Share Parameters",
			Required:            true,
			Attributes:          ShareParametersExportSchema(),
			PlanModifiers: []planmodifier.Object{
				objectplanmodifier.RequiresReplace(),
			},
		},
	}
}

// ShareParametersExportSchema returns the schema for the share parameters
func ShareParametersExportSchema() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"filename": schema.StringAttribute{
			MarkdownDescription: "File Name - The name of the server configuration profile file to export.",
			Description:         "File Name - The name of the server configuration profile file to export.",
			Required:            true,
			Validators: []validator.String{
				stringvalidator.LengthAtLeast(1),
			},
		},
		"ip_address": schema.StringAttribute{
			MarkdownDescription: "IPAddress - The IP address of the target export server.",
			Description:         "IPAddress - The IP address of the target export server.",
			Optional:            true,
			Validators: []validator.String{
				stringvalidator.LengthAtLeast(1),
			},
		},
		"ignore_certificate_warning": schema.BoolAttribute{
			MarkdownDescription: "Ignore Certificate Warning",
			Description:         "Ignore Certificate Warning",
			Optional:            true,
			Computed:            true,
			Default:             booldefault.StaticBool(false),
		},
		"password": schema.StringAttribute{
			MarkdownDescription: "Password - The password for the share server user account. This password is required if the share " +
				"type is set to \"CIFS\". It is required only if the share type is set to \"CIFS\". It is not required if the share " +
				"type is set to \"NFS\".",
			Description: "Password - The password for the share server user account. This password is required if the share " +
				"type is set to \"CIFS\". It is required only if the share type is set to \"CIFS\". It is not required if the share " +
				"type is set to \"NFS\".",
			Optional:  true,
			Sensitive: true,
			Validators: []validator.String{
				stringvalidator.LengthAtLeast(1),
			},
		},
		"port_number": schema.Int64Attribute{
			MarkdownDescription: "Port Number - The port number used to communicate with the share server. The default value is 80. ",
			Description:         "Port Number - The port number used to communicate with the share server. The default value is 80. ",
			Optional:            true,
			Computed:            true,
			Default:             int64default.StaticInt64(defaultSCPPort),
		},
		"proxy_support": schema.BoolAttribute{
			MarkdownDescription: "Proxy Support - Specifies whether or not to use a proxy server for the export operation. " +
				"If `true`, export operation will use a proxy server for communication with the export server. If `false`, " +
				"export operation will not use a proxy server for communication with the export server. Default value is `false`.",
			Description: "Password - The password for the share server user account. This password is required if the share " +
				"type is set to \"CIFS\". It is required only if the share type is set to \"CIFS\". It is not required if the share " +
				"type is set to \"NFS\".",
			Optional: true,
			Computed: true,
			Default:  booldefault.StaticBool(false),
		},
		"proxy_password": schema.StringAttribute{
			MarkdownDescription: "The password for the proxy server. This is required if the proxy_support parameter is set to `true`. " +
				"It is used for authenticating the proxy server credentials.",
			Description: "The password for the proxy server. This is required if the proxy_support parameter is set to `true`. " +
				"It is used for authenticating the proxy server credentials.",
			Optional:  true,
			Sensitive: true,
			Validators: []validator.String{
				stringvalidator.LengthAtLeast(1),
			},
		},
		"proxy_port": schema.Int64Attribute{
			MarkdownDescription: `The port number used by the proxy server. 
			This parameter is optional. 
			If not provided, the default port number (80) is used for the communication with the proxy server.`,
			Description: `The port number used by the proxy server. 
			This parameter is optional. 
			If not provided, the default port number (80) is used for the communication with the proxy server.`,
			Optional: true,
			Computed: true,
			Default:  int64default.StaticInt64(defaultSCPPort),
		},
		"proxy_server": schema.StringAttribute{
			MarkdownDescription: `The IP address or hostname of the proxy server.
			 This is the server that acts as a bridge between the iDRAC and the Server Configuration Profile share server. 
			 It is used to communicate with the Server Configuration Profile share server 
			 in order to export the Server Configuration Profile. If the Server Configuration Profile share server
			  is not accessible from the iDRAC directly, then a proxy server must be used in order to establish the connection. 
			  This parameter is optional. 
			  If it is not provided, the Server Configuration Profile export operation
			   will attempt to connect to the Server Configuration Profile share server directly.`,
			Description: `The IP address or hostname of the proxy server.
			 This is the server that acts as a bridge between the iDRAC and the Server Configuration Profile share server.
			  It is used to communicate with the Server Configuration Profile share server 
			  in order to export the Server Configuration Profile. If the Server Configuration Profile share server
			   is not accessible from the iDRAC directly, then a proxy server must be used in order to establish the connection.
			    This parameter is optional.
				 If it is not provided, the Server Configuration Profile export operation 
				 will attempt to connect to the Server Configuration Profile share server directly.`,
			Optional: true,
			Validators: []validator.String{
				stringvalidator.LengthAtLeast(1),
			},
		},

		"proxy_type": schema.StringAttribute{
			MarkdownDescription: "The type of proxy server to be used. " +
				"The default is \"HTTP\"." +
				" If set to \"SOCKS4\", a SOCKS4 proxy server must be specified." +
				" If set to \"HTTP\", an HTTP proxy server must be specified. " +
				"If not specified, the Server Configuration Profile export operation will" +
				" attempt to connect to the Server Configuration Profile share server directly.",
			Description: "The type of proxy server to be used. " +
				"The default is \"HTTP\"." +
				" If set to \"SOCKS4\", a SOCKS4 proxy server must be specified." +
				" If set to \"HTTP\", an HTTP proxy server must be specified. " +
				"If not specified, the Server Configuration Profile export operation will" +
				" attempt to connect to the Server Configuration Profile share server directly.",
			Optional: true,
			Computed: true,
			Validators: []validator.String{
				stringvalidator.LengthAtLeast(1),
				stringvalidator.OneOf([]string{
					string("HTTP"),
					string("SOCKS4"),
				}...),
			},
			Default: stringdefault.StaticString("HTTP"),
		},
		"proxy_username": schema.StringAttribute{
			MarkdownDescription: "The username to be used when connecting to the proxy server.",
			Description:         "The username to be used when connecting to the proxy server.",
			Optional:            true,
			Validators: []validator.String{
				stringvalidator.LengthAtLeast(1),
			},
		},
		"share_name": schema.StringAttribute{
			MarkdownDescription: `Share Name - The name of the directory or share on the server 
			that contains the Server Configuration Profile file to export.`,
			Description: `Share Name - The name of the directory or share on the server 
			that contains the Server Configuration Profile file to export.`,
			Optional: true,
			Validators: []validator.String{
				stringvalidator.LengthAtLeast(1),
			},
		},
		"share_type": schema.StringAttribute{
			MarkdownDescription: "Share Type - The type of share being used to export the Server Configuration Profile file.",
			Description:         "Share Type - The type of share being used to export the Server Configuration Profile file.",
			Required:            true,
			Validators: []validator.String{
				stringvalidator.LengthAtLeast(1),
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
			MarkdownDescription: "Filter configuration by target",
			Description:         "Filter configuration by target",
			Optional:            true,
			ElementType:         types.StringType,
			Computed:            true,
			Default: listdefault.StaticValue(
				types.ListValueMust(
					types.StringType,
					[]attr.Value{
						types.StringValue("ALL"),
					},
				),
			),
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
			MarkdownDescription: `Username - The username to use when authenticating with the server
			 that contains the Server Configuration Profile file being exported.`,
			Description: `Username - The username to use when authenticating with the server
			 that contains the Server Configuration Profile file being exported.`,
			Optional: true,
			Validators: []validator.String{
				stringvalidator.LengthAtLeast(1),
			},
		},
	}
}

// ValidateConfig validates the resource config.
func (*ScpExportResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	// Get Plan Data
	var plan models.TFRedfishScpExport
	diags := req.Config.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	if plan.ShareParameters.IsUnknown() {
		return
	}
	var sp models.TFShareParameters
	plan.ShareParameters.As(ctx, &sp, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true})
	shareType := sp.ShareType.ValueString()
	if shareType == "NFS" {
		if sp.IPAddress.IsNull() || sp.ShareName.IsNull() {
			resp.Diagnostics.AddError(
				"Export NFS Error",
				"When configuring the share type as ‘NFS’, it is essential to provide both the IP address and the share name.")
			return
		}
	} else if shareType == "CIFS" {
		if sp.IPAddress.IsNull() || sp.ShareName.IsNull() || sp.Username.IsNull() || sp.Password.IsNull() {
			resp.Diagnostics.AddError(
				"Export CIFS Error",
				"When configuring the share type as CIFS, it is essential to provide the IP address, share name, username and password.")
			return
		}
	} else if shareType == "HTTP" || shareType == "HTTPS" {
		if sp.IPAddress.IsNull() {
			resp.Diagnostics.AddError(
				"Export HTTP/HTTPS IP Error",
				fmt.Sprintf(
					"When configuring the share type as %s, it is essential to provide the IP address.",
					shareType))
			return
		}
		if sp.ProxySupport.ValueBool() {
			if sp.ProxyServer.IsNull() {
				resp.Diagnostics.AddError(
					"Export Proxy Error",
					fmt.Sprintf(
						"When configuring the share type as %s and Proxy Support is enabled, it is essential to provide the proxy server.",
						shareType))
				return
			}
		}
	}
	if resp.Diagnostics.HasError() {
		return
	}
}

// Create creates the resource and sets the initial Terraform state.
func (r *ScpExportResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	tflog.Trace(ctx, "resource_ScpExport create : Started")
	// Get Plan Data
	var plan models.TFRedfishScpExport
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if plan.ShareParameters.IsUnknown() {
		return
	}
	var sp models.TFShareParameters
	plan.ShareParameters.As(ctx, &sp, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true})

	redfishMutexKV.Lock(plan.RedfishServer[0].Endpoint.ValueString())
	defer redfishMutexKV.Unlock(plan.RedfishServer[0].Endpoint.ValueString())

	service, err := NewConfig(r.p, &plan.RedfishServer)
	if err != nil {
		resp.Diagnostics.AddError("service error - config create", err.Error())
		return
	}

	content, err := scpExportExecutor(ctx, service, plan)
	if err != nil {
		resp.Diagnostics.AddError("executor error", err.Error())
		return
	}
	plan.FileContent = types.StringValue(content)

	tflog.Trace(ctx, "resource_ScpExport create: updating state finished, saving ...")
	// Save into State
	plan.ID = types.StringValue("scpExport")
	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	tflog.Trace(ctx, "resource_ScpExport create: finish")
}

// Update updates the resource and sets the updated Terraform state on success.
func (*ScpExportResource) Update(_ context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError(
		"Error updating Export Server Configuration Profile.",
		"An update plan of Export Server Configuration Profile should never be invoked. This resource is supposed to be replaced on update.",
	)
}

// Read refreshes the Terraform state with the latest data.
func (*ScpExportResource) Read(_ context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	resp.State = req.State
}

// Delete deletes the resource and removes the Terraform state on success.
func (*ScpExportResource) Delete(ctx context.Context, _ resource.DeleteRequest, resp *resource.DeleteResponse) {
	resp.State.RemoveResource(ctx)
}

// scpExportExecutor executes the SCP export process.
func scpExportExecutor(ctx context.Context, service *gofish.Service, plan models.TFRedfishScpExport) (string, error) {
	var sp models.TFShareParameters
	plan.ShareParameters.As(ctx, &sp, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true})
	managers, err := service.Managers()
	if err != nil {
		return "", err
	}
	dellManager, err := dell.Manager(managers[0])
	if err != nil {
		return "", err
	}
	exportURL := dellManager.Actions.ExportSystemConfigurationTarget
	resp, err := service.GetClient().Post(exportURL, constructExportPayload(ctx, plan, dellManager.FirmwareVersion))
	if err != nil {
		return "", err
	}
	if location, err := resp.Location(); err == nil {
		taskURI := location.EscapedPath()
		if sp.ShareType.ValueString() == "LOCAL" {
			fileContent, err := common.GetJobAttachment(service, taskURI, intervalJobCheckTime, defaultJobTimeout)
			if err != nil {
				return "", err
			}
			return base64.StdEncoding.EncodeToString(fileContent), nil
		}
		if err := common.WaitForDellJobToFinish(service, taskURI, intervalJobCheckTime, defaultJobTimeout); err != nil {
			return "", err
		}
	}
	return "SCP exported successfully", nil
}

// constructExportPayload is a function that constructs the SCP export payload
func constructExportPayload(ctx context.Context, plan models.TFRedfishScpExport, firmwareVersion string) models.SCPExport {
	var sp models.TFShareParameters
	plan.ShareParameters.As(ctx, &sp, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true})
	var target, includeInExport []string
	plan.IncludeInExport.ElementsAs(context.Background(), &includeInExport, true)
	sp.Target.ElementsAs(context.Background(), &target, true)
	scpExport := models.SCPExport{
		ExportFormat:    plan.ExportFormat.ValueString(),
		ExportUse:       plan.ExportUse.ValueString(),
		IncludeInExport: includeInExport,
		ShareParameters: models.ShareParameters{},
	}
	// Set ignoreCertificateWarning to "Disabled" if the plan's ignoreCertificateWarning value is true,
	// otherwise set it to "Enabled".
	ignoreCertificateWarning := "Disabled"
	if !sp.IgnoreCertificateWarning.ValueBool() {
		ignoreCertificateWarning = "Enabled"
	}

	portNumber := strconv.FormatInt(sp.PortNumber.ValueInt64(), defaultIntBase)
	scpExport.ShareParameters = models.ShareParameters{
		FileName:                 sp.FileName.ValueString(),
		IPAddress:                sp.IPAddress.ValueString(),
		IgnoreCertificateWarning: ignoreCertificateWarning,
		Password:                 sp.Password.ValueString(),
		PortNumber:               portNumber,
		ShareName:                sp.ShareName.ValueString(),
		ShareType:                sp.ShareType.ValueString(),
		Username:                 sp.Username.ValueString(),
	}

	if strings.HasPrefix(firmwareVersion, "5.") {
		scpExport.ShareParameters.Target = strings.Join(target, ",")
	} else {
		scpExport.ShareParameters.Target = target
	}

	// Set proxySupport to "Enabled" if the plan's proxySupport value is true,
	// otherwise set it to "Disabled".
	proxySupport := "Enabled"
	if !sp.ProxySupport.ValueBool() {
		proxySupport = "Disabled"
	}
	proxyPassword := sp.ProxyPassword.ValueString()
	proxyPort := strconv.FormatInt(sp.ProxyPort.ValueInt64(), defaultIntBase)
	proxyServer := sp.ProxyServer.ValueString()
	proxyType := sp.ProxyType.ValueString()
	proxyUserName := sp.ProxyUserName.ValueString()

	scpExport.ShareParameters.ProxyPassword = proxyPassword
	scpExport.ShareParameters.ProxyPort = proxyPort
	scpExport.ShareParameters.ProxyServer = proxyServer
	scpExport.ShareParameters.ProxySupport = proxySupport
	scpExport.ShareParameters.ProxyType = proxyType
	scpExport.ShareParameters.ProxyUserName = proxyUserName
	return scpExport
}
