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
	"fmt"
	"strconv"
	"strings"
	"terraform-provider-redfish/common"
	"terraform-provider-redfish/gofish/dell"
	"terraform-provider-redfish/redfish/models"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listdefault"
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
	_ resource.Resource = &ScpImportResource{}
)

const (
	defaultScpImportTimeout int64 = 1200
	minScpImportTimeout     int64 = 300
	maxScpImportTimeout     int64 = 3600
	defaultSCPPort          int64 = 80
	defaultIntBase          int   = 10
)

// NewScpImportResource is a helper function to simplify the provider implementation.
func NewScpImportResource() resource.Resource {
	return &ScpImportResource{}
}

// ScpImportResource is the resource implementation.
type ScpImportResource struct {
	p *redfishProvider
}

// Configure implements resource.ResourceWithConfigure
func (r *ScpImportResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	r.p = req.ProviderData.(*redfishProvider)
}

// Metadata returns the resource type name.
func (*ScpImportResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "idrac_server_configuration_profile_import"
}

// Schema defines the schema for the resource.
func (*ScpImportResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Resource for managing iDRAC Server Configuration Profile Import on iDRAC Server.",
		Version:             1,
		Attributes:          RedfishScpImportSchema(),
		Blocks:              RedfishServerResourceBlockMap(),
	}
}

// RedfishScpImportSchema defines the schema for the resource.
func RedfishScpImportSchema() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"id": schema.StringAttribute{
			MarkdownDescription: "ID of the Import SCP resource",
			Description:         "ID of the Import SCP resource",
			Computed:            true,
		},
		"host_power_state": schema.StringAttribute{
			MarkdownDescription: `Host Power State. This attribute allows you to specify the power state of the host when the
				iDRAC is performing the import operation. Accepted values are: "On" or "Off". If this attribute is not specified
				or is set to "On", the host is powered on before the import operation. If it is set to "Off", the host is powered
				off before the import operation. Note that the host will be powered back on after the import is completed.`,
			Description: `Host Power State. This attribute allows you to specify the power state of the
			host when the iDRAC is performing the import operation. Accepted values are: "On" or "Off".
			If this attribute is not specified or is set to "On", the host is powered on before the
			import operation. If it is set to "Off", the host is powered off before the import operation.
			Note that the host will be powered back on after the import is completed.`,
			Optional: true,
			Computed: true,
			Default:  stringdefault.StaticString("On"),
			Validators: []validator.String{
				stringvalidator.LengthAtLeast(1),
				stringvalidator.OneOf([]string{
					string("On"),
					string("Off"),
				}...),
			},
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
		"import_buffer": schema.StringAttribute{
			MarkdownDescription: "Buffer content to perform Import." +
				"This is only required for localstore and is not applicable for CIFS/NFS style Import. " +
				"If the import buffer is empty, then it will perform the import from the source path specified in share parameters.",
			Description: "Buffer content to perform Import." +
				"This is only required for localstore and is not applicable for CIFS/NFS style Import. If the import buffer is empty," +
				"then it will perform the import from the source path specified in share parameters.",
			Optional: true,
			Validators: []validator.String{
				stringvalidator.LengthAtLeast(1),
			},
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
		"shutdown_type": schema.StringAttribute{
			MarkdownDescription: "Shutdown Type. This attribute specifies the type of shutdown that should be performed " +
				"before importing the server configuration profile. Accepted values are: \"Graceful\" (default), " +
				"\"Forced\", or \"NoReboot\". If set to \"Graceful\", the server will be gracefully shut down before the " +
				"import. If set to \"Forced\", the server will be forcefully shut down before the import. If set to " +
				"\"NoReboot\", the server will not be restarted after the import. Note that if the server is powered off " +
				"before the import operation, it will not be powered back on after the import is completed. If the server " +
				"is powered on before the import operation, it will be powered off during the import process if this " +
				"attribute is set to \"Forced\" or \"NoReboot\", and will be powered back on after the import is completed " +
				"if this attribute is set to \"Graceful\" or \"NoReboot\".",
			Description: "Shutdown Type. This attribute specifies the type of shutdown that should be performed " +
				"before importing the server configuration profile. Accepted values are: \"Graceful\" (default), " +
				"\"Forced\", or \"NoReboot\". If set to \"Graceful\", the server will be gracefully shut down before the " +
				"import. If set to \"Forced\", the server will be forcefully shut down before the import. If set to " +
				"\"NoReboot\", the server will not be restarted after the import. Note that if the server is powered off " +
				"before the import operation, it will not be powered back on after the import is completed. If the server " +
				"is powered on before the import operation, it will be powered off during the import process if this " +
				"attribute is set to \"Forced\" or \"NoReboot\", and will be powered back on after the import is completed " +
				"if this attribute is set to \"Graceful\" or \"NoReboot\".",
			Optional: true,
			Computed: true,
			Default:  stringdefault.StaticString("Graceful"),
			Validators: []validator.String{
				stringvalidator.LengthAtLeast(1),
				stringvalidator.OneOf([]string{
					string("Graceful"),
					string("Forced"),
					string("NoReboot"),
				}...),
			},
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
		"time_to_wait": schema.Int64Attribute{
			MarkdownDescription: `Time To Wait (in seconds) - specifies the time to wait for the server configuration profile
				to be imported. This is useful for ensuring that the server is powered off before the import operation, and for waiting
				for the import to complete before powering the server back on. The default value is 1200 seconds (or 20 minutes), but can
				be set to a lower value of 300 seconds (or 5 minutes) upto a max value of 3600 seconds (or 60 minutes) if desired. Note
				that this attribute only applies if the server is powered on before the import operation, or if the server is powered off
				before the import operation and the shutdown type is set to "Graceful" or "NoReboot". The minimum value is 300 seconds, and
				the maximum value is 3600 seconds (or 1 hour). If the server is powered off before the import operation and the shutdown
				type is set to "Forced" or "NoReboot", the import operation will occur immediately and the server will not be powered
				back on after the import is completed.`,
			Description: `Time To Wait (in seconds) - specifies the time to wait for the server configuration profile
				to be imported. This is useful for ensuring that the server is powered off before the import operation, and for waiting
				for the import to complete before powering the server back on. The default value is 1200 seconds (or 20 minutes), but can
				be set to a lower value of 300 seconds (or 5 minutes) upto a max value of 3600 seconds (or 60 minutes) if desired. Note
				that this attribute only applies if the server is powered on before the import operation, or if the server is powered off
				before the import operation and the shutdown type is set to "Graceful" or "NoReboot". The minimum value is 300 seconds, and
				the maximum value is 3600 seconds (or 1 hour). If the server is powered off before the import operation and the shutdown
				type is set to "Forced" or "NoReboot", the import operation will occur immediately and the server will not be powered
				back on after the import is completed.`,
			Optional: true,
			Computed: true,
			Default:  int64default.StaticInt64(defaultScpImportTimeout),
			Validators: []validator.Int64{
				int64validator.Between(minScpImportTimeout, maxScpImportTimeout),
			},
			PlanModifiers: []planmodifier.Int64{
				int64planmodifier.RequiresReplace(),
			},
		},
		"share_parameters": schema.SingleNestedAttribute{
			MarkdownDescription: "Share Parameters",
			Description:         "Share Parameters",
			Required:            true,
			Attributes:          ShareParametersSchema(),
			PlanModifiers: []planmodifier.Object{
				objectplanmodifier.RequiresReplace(),
			},
		},
	}
}

// ShareParametersSchema returns the schema for the share parameters
func ShareParametersSchema() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"filename": schema.StringAttribute{
			MarkdownDescription: "File Name - The name of the server configuration profile file to import. This is the name of the file " +
				"that was previously exported using the Server Configuration Profile Export operation. This file is typically " +
				"in the xml/json format",
			Description: "File Name - The name of the server configuration profile file to import. This is the name of the file " +
				"that was previously exported using the Server Configuration Profile Export operation. This file is typically " +
				"in the xml/json format",
			Required: true,
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
			MarkdownDescription: "Proxy Support - Specifies whether or not to use a proxy server for the import operation. " +
				"If `true`, import operation will use a proxy server for communication with the export server. If `false`, " +
				"import operation will not use a proxy server for communication with the export server. Default value is `false`.",
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
			 in order to import the Server Configuration Profile. If the Server Configuration Profile share server
			  is not accessible from the iDRAC directly, then a proxy server must be used in order to establish the connection. 
			  This parameter is optional. 
			  If it is not provided, the Server Configuration Profile import operation
			   will attempt to connect to the Server Configuration Profile share server directly.`,
			Description: `The IP address or hostname of the proxy server.
			 This is the server that acts as a bridge between the iDRAC and the Server Configuration Profile share server.
			  It is used to communicate with the Server Configuration Profile share server 
			  in order to import the Server Configuration Profile. If the Server Configuration Profile share server
			   is not accessible from the iDRAC directly, then a proxy server must be used in order to establish the connection.
			    This parameter is optional.
				 If it is not provided, the Server Configuration Profile import operation 
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
				"If not specified, the Server Configuration Profile import operation will" +
				" attempt to connect to the Server Configuration Profile share server directly.",
			Description: "The type of proxy server to be used. " +
				"The default is \"HTTP\"." +
				" If set to \"SOCKS4\", a SOCKS4 proxy server must be specified." +
				" If set to \"HTTP\", an HTTP proxy server must be specified. " +
				"If not specified, the Server Configuration Profile import operation will" +
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
			that contains the Server Configuration Profile file to import.`,
			Description: `Share Name - The name of the directory or share on the server 
			that contains the Server Configuration Profile file to import.`,
			Optional: true,
			Validators: []validator.String{
				stringvalidator.LengthAtLeast(1),
			},
		},
		"share_type": schema.StringAttribute{
			MarkdownDescription: "Share Type - The type of share being used to import the Server Configuration Profile file.",
			Description:         "Share Type - The type of share being used to import the Server Configuration Profile file.",
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
			 that contains the Server Configuration Profile file being imported.`,
			Description: `Username - The username to use when authenticating with the server
			 that contains the Server Configuration Profile file being imported.`,
			Optional: true,
			Validators: []validator.String{
				stringvalidator.LengthAtLeast(1),
			},
		},
	}
}

// ValidateConfig validates the resource config.
func (*ScpImportResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var plan models.RedfishScpImport
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
				"Import NFS Error",
				"When configuring the share type as ‘NFS’, it is essential to provide both the IP address and the share name.")
			return
		}
	} else if shareType == "CIFS" {
		if sp.IPAddress.IsNull() || sp.ShareName.IsNull() || sp.Username.IsNull() || sp.Password.IsNull() {
			resp.Diagnostics.AddError(
				"Import CIFS Error",
				"When configuring the share type as CIFS, it is essential to provide the IP address, share name, username and password.")
			return
		}
	} else if shareType == "HTTP" || shareType == "HTTPS" {
		if sp.IPAddress.IsNull() {
			resp.Diagnostics.AddError(
				"Import HTTP/HTTPS IP Error",
				fmt.Sprintf(
					"When configuring the share type as %s, it is essential to provide the IP address.",
					shareType))
			return
		}
		if sp.ProxySupport.ValueBool() {
			if sp.ProxyServer.IsNull() {
				resp.Diagnostics.AddError(
					"Import Proxy Error",
					fmt.Sprintf(
						"When configuring the share type as %s and Proxy Support is enabled, it is essential to provide the proxy server.",
						shareType))
				return
			}
		}
	}
}

// Create creates the resource and sets the initial Terraform state.
func (r *ScpImportResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	tflog.Trace(ctx, "resource_ScpImport create : Started")
	// Get Plan Data
	var plan models.RedfishScpImport
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
		resp.Diagnostics.AddError("service error", err.Error())
		return
	}

	log, err := scpImportExecutor(ctx, service, plan)
	if err != nil {
		resp.Diagnostics.AddError(log, err.Error())
		return
	}

	tflog.Trace(ctx, "resource_ScpImport create: updating state finished, saving ...")
	// Save into State
	plan.ID = types.StringValue("scpimport")
	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	tflog.Trace(ctx, "resource_ScpImport create: finish")
}

// Update updates the resource and sets the updated Terraform state on success.
func (*ScpImportResource) Update(_ context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError(
		"Error updating Import Server Configuration Profile.",
		"An update plan of Import Server Configuration Profile should never be invoked. This resource is supposed to be replaced on update.",
	)
}

// Read refreshes the Terraform state with the latest data.
func (*ScpImportResource) Read(_ context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	resp.State = req.State
}

// Delete deletes the resource and removes the Terraform state on success.
func (*ScpImportResource) Delete(ctx context.Context, _ resource.DeleteRequest, resp *resource.DeleteResponse) {
	resp.State.RemoveResource(ctx)
}

// scpImportExecutor is a function that imports a server configuration profile (SCP) into a Redfish service.
//
// Parameters:
// - service: a pointer to a gofish.Service object representing the Redfish service.
// - scpImportPayload: a models.SCPImport object containing the payload for the SCP import.
// - jobCheckIntervalTime: an int64 representing the interval time for checking the job status.
// - jobDefaultTimeout: an int64 representing the default timeout for the job.
//
// Returns:
// - string: a message indicating the result of the SCP import.
// - error: an error object if there was an error during the import process.
func scpImportExecutor(ctx context.Context, service *gofish.Service, plan models.RedfishScpImport) (string, error) {
	managers, err := service.Managers()
	if err != nil {
		return "error while retrieving managers", err
	}
	dellManager, err := dell.Manager(managers[0])
	if err != nil {
		return "error while retrieving dell manager", err
	}
	importURL := dellManager.Actions.ImportSystemConfigurationTarget
	response, err := service.GetClient().Post(importURL, constructPayload(ctx, plan, dellManager.FirmwareVersion))
	if err != nil {
		return "error during import", err
	}

	if location, err := response.Location(); err == nil {
		taskURI := location.EscapedPath()
		err = common.WaitForDellJobToFinish(service, taskURI, intervalJobCheckTime, defaultJobTimeout)
		if err != nil {
			return "error waiting for SCP Export monitor task to be completed", err
		}
	}
	return "The server configuration profile was successfully imported", nil
}

// constructPayload constructs a SCPImport payload from a RedfishScpImport plan.
//
// It takes a RedfishScpImport plan as input and returns a SCPImport payload.
// The function extracts the necessary values from the plan and constructs the
// SCPImport payload by populating its fields. The function also constructs the
// ShareParameters field of the SCPImport payload by extracting values from the
// plan.
//
// Parameters:
// - plan: The RedfishScpImport plan from which to construct the SCPImport payload.
//
// Return:
// - models.SCPImport: The constructed SCPImport payload.
func constructPayload(ctx context.Context, plan models.RedfishScpImport, firmwareVersion string) models.SCPImport {
	var sp models.TFShareParameters
	plan.ShareParameters.As(ctx, &sp, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true})

	var target []string
	sp.Target.ElementsAs(context.Background(), &target, true)
	scpImport := models.SCPImport{
		HostPowerState: plan.HostPowerState.ValueString(),
		ImportBuffer:   plan.ImportBuffer.ValueString(),
		ShutdownType:   plan.ShutdownType.ValueString(),
		TimeToWait:     plan.TimeToWait.ValueInt64(),
	}

	// Set ignoreCertificateWarning to "Disabled" if the plan's ignoreCertificateWarning value is true,
	// otherwise set it to "Enabled".
	ignoreCertificateWarning := "Disabled"
	if !sp.IgnoreCertificateWarning.ValueBool() {
		ignoreCertificateWarning = "Enabled"
	}

	portNumber := strconv.FormatInt(sp.PortNumber.ValueInt64(), defaultIntBase)
	scpImport.ShareParameters = models.ShareParameters{
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
		scpImport.ShareParameters.Target = strings.Join(target, ", ")
	} else {
		scpImport.ShareParameters.Target = target
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

	scpImport.ShareParameters.ProxyPassword = proxyPassword
	scpImport.ShareParameters.ProxyPort = proxyPort
	scpImport.ShareParameters.ProxyServer = proxyServer
	scpImport.ShareParameters.ProxySupport = proxySupport
	scpImport.ShareParameters.ProxyType = proxyType
	scpImport.ShareParameters.ProxyUserName = proxyUserName

	return scpImport
}
