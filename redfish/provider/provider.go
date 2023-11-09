package provider

import (

	// "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"context"
	"terraform-provider-redfish/mutexkv"
	"terraform-provider-redfish/redfish/models"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
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
	Username string
	Password string
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

	p.Username = config.Username.ValueString()
	p.Password = config.Password.ValueString()

	resp.ResourceData = p
	resp.DataSourceData = p

	tflog.Trace(ctx, p.Username+" "+p.Password)
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
	}
}
