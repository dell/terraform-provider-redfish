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

	// "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"context"
	"terraform-provider-redfish/mutexkv"
	"terraform-provider-redfish/redfish/models"

	"github.com/hashicorp/terraform-plugin-framework-validators/mapvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// This is a global MutexKV for use within this plugin
var redfishMutexKV = mutexkv.NewMutexKV()

// Ensure the implementation satisfies the provider.Provider interface.
var _ provider.Provider = &redfishProvider{}

// New - returns new provider struct definition.
func New() provider.Provider {
	return &redfishProvider{}
}

type redfishProvider struct {
	models.ProviderConfig
}

// Metadata - provider metadata AKA name.
func (*redfishProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "redfish_"
}

// Schema implements provider.Provider.
func (*redfishProvider) Schema(ctx context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Terraform Provider Redfish",
		Attributes: map[string]schema.Attribute{
			"user": schema.StringAttribute{
				MarkdownDescription: "This field is the user to login against the redfish API",
				Description:         "This field is the user to login against the redfish API",
				Optional:            true,
			},
			"password": schema.StringAttribute{
				MarkdownDescription: "This field is the password related to the user given",
				Description:         "This field is the password related to the user given",
				Optional:            true,
				Sensitive:           true,
			},
			"redfish_servers": schema.MapNestedAttribute{
				MarkdownDescription: "Map of server BMCs with their alias keys and respective user credentials. " +
					"This is required when resource/datasource's `redfish_alias` is not null",
				Description: "Map of server BMCs with their alias keys and respective user credentials. " +
					"This is required when resource/datasource's `redfish_alias` is not null",
				Optional: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"user": schema.StringAttribute{
							Optional:    true,
							Description: "User name for login",
						},
						"password": schema.StringAttribute{
							Optional:    true,
							Description: "User password for login",
							Sensitive:   true,
						},
						"endpoint": schema.StringAttribute{
							Required:    true,
							Description: "Server BMC IP address or hostname",
						},
						"ssl_insecure": schema.BoolAttribute{
							Optional:    true,
							Description: "This field indicates whether the SSL/TLS certificate must be verified or not",
						},
					},
				},
				Validators: []validator.Map{
					mapvalidator.KeysAre(stringvalidator.LengthAtLeast(1)),
				},
			},
		},
	}
	tflog.Trace(ctx, "resource schema created")
}

// Configure - provider pre-initiate calle function.
func (p *redfishProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	// If the upstream provider SDK or HTTP client requires configuration, such
	// as authentication or logging, this is a great opportunity to do so.
	tflog.Trace(ctx, "Started configuring the provider")
	config := models.ProviderConfig{}
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)

	if config.Username.IsUnknown() {
		// Cannot connect to client with an unknown value
		resp.Diagnostics.AddWarning(
			"Unable to create client",
			"Cannot use unknown value as username",
		)
		return
	}

	if config.Password.IsUnknown() {
		// Cannot connect to client with an unknown value
		resp.Diagnostics.AddWarning(
			"Unable to create client",
			"Cannot use unknown value as password",
		)
		return
	}

	p.Username = config.Username
	p.Password = config.Password
	p.Servers = config.Servers

	resp.ResourceData = p
	resp.DataSourceData = p

	tflog.Trace(ctx, config.Username.ValueString()+" "+config.Password.ValueString())
	tflog.Trace(ctx, "Finished configuring the provider")
}

// Resources function to add new resource
func (*redfishProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewPowerResource,
		NewVirtualMediaResource,
		NewUserAccountResource,
		NewSimpleUpdateResource,
		NewDellIdracAttributesResource,
		NewRedfishStorageVolumeResource,
		NewBiosResource,
		NewManagerResetResource,
		NewBootOrderResource,
		NewBootSourceOverrideResource,
		NewCertificateResource,
		NewDellLCAttributesResource,
		NewDellSystemAttributesResource,
		NewIdracFirmwareUpdateResource,
		NewUserAccountPasswordResource,
		NewScpImportResource,
		NewScpExportResource,
		NewRedfishNICResource,
	}
}

// DataSources function to add new data-source
func (*redfishProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewBiosDatasource,
		NewDellIdracAttributesDatasource,
		NewStorageDatasource,
		NewDellVirtualMediaDatasource,
		NewSystemBootDatasource,
		NewFirmwareInventoryDatasource,
		NewNICDatasource,
	}
}
