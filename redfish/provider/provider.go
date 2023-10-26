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
func (p *redfishProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "redfish_"
}

// Schema implements provider.Provider.
func (p *redfishProvider) Schema(ctx context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
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

func (p *redfishProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewPowerResource,
		NewVirtualMediaResource,
	}
}

func (p *redfishProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{}
}

// func Provider() *schema.Provider {
// 	provider := &schema.Provider{
// 		Schema: map[string]*schema.Schema{
// 			"user": {
// 				Type:        schema.TypeString,
// 				Optional:    true,
// 				Description: "Default value. This field is the user to login against the redfish API",
// 			},
// 			"password": {
// 				Type:        schema.TypeString,
// 				Optional:    true,
// 				Description: "Default value. This field is the password related to the user given",
// 			},
// 		},

// 		ResourcesMap: map[string]*schema.Resource{
// 			"redfish_user_account":          resourceRedfishUserAccount(),
// 			"redfish_bios":                  resourceRedfishBios(),
// 			"redfish_storage_volume":        resourceRedfishStorageVolume(),
// 			"redfish_virtual_media":         resourceRedfishVirtualMedia(),
// 			"redfish_power":                 resourceRedFishPower(),
// 			"redfish_simple_update":         resourceRedfishSimpleUpdate(),
// 			"redfish_dell_idrac_attributes": resourceRedfishDellIdracAttributes(),
// 		},

// 		DataSourcesMap: map[string]*schema.Resource{
// 			"redfish_bios":                  dataSourceRedfishBios(),
// 			"redfish_virtual_media":         dataSourceRedfishVirtualMedia(),
// 			"redfish_storage":               dataSourceRedfishStorage(),
// 			"redfish_firmware_inventory":    dataSourceRedfishFirmwareInventory(),
// 			"redfish_dell_idrac_attributes": dataSourceRedfishDellIdracAttributes(),
// 			"redfish_system_boot":           dataSourceRedfishSystemBoot(),
// 		},
// 	}

// 	provider.ConfigureFunc = func(d *schema.ResourceData) (interface{}, error) {
// 		terraformVersion := provider.TerraformVersion
// 		if terraformVersion == "" {
// 			// Terraform 0.12 introduced this field to the protocol
// 			// We can therefore assume that if it's missing it's 0.10 or 0.11
// 			terraformVersion = "0.11+compatible"
// 		}
// 		return providerConfigure(d, terraformVersion)
// 	}

// 	return provider
// }

// func providerConfigure(d *schema.ResourceData, terraformVersion string) (interface{}, error) {
// 	/*Redfish token issued by iDRAC needs to be revoked when the provider is done.
// 	At the moment, the terraform SDK (Provider.StopFunc) is not implemented. To follow up, please refer to this pull request:
// 	https://github.com/hashicorp/terraform-plugin-sdk/pull/377
// 	*/

// 	return d, nil
// }
