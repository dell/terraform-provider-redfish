package redfish

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func Provider() *schema.Provider {
	provider := &schema.Provider{
		Schema: map[string]*schema.Schema{
			"redfish_server": {
				Type:        schema.TypeList,
				Required:    true,
				Description: "This list contains the different redfish endpoints to manage (different servers)",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"user": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "This field is the user to login against the redfish API",
						},
						"password": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "This field is the password related to the user given",
						},
						"endpoint": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "This field is the endpoint where the redfish API is placed",
						},
						"ssl_insecure": {
							Type:        schema.TypeBool,
							Optional:    true,
							Description: "This field indicates if the SSL/TLS certificate must be verified",
						},
					},
				},
			},
		},

		ResourcesMap: map[string]*schema.Resource{
			"redfish_user_account": resourceUserAccount(),
			//	"redfish_bios":           resourceRedfishBios(),
			"redfish_storage_volume": resourceRedfishStorageVolume(),
		},

		DataSourcesMap: map[string]*schema.Resource{
			//	"redfish_bios": dataSourceRedfishBios(),
		},
	}

	provider.ConfigureFunc = func(d *schema.ResourceData) (interface{}, error) {
		terraformVersion := provider.TerraformVersion
		if terraformVersion == "" {
			// Terraform 0.12 introduced this field to the protocol
			// We can therefore assume that if it's missing it's 0.10 or 0.11
			terraformVersion = "0.11+compatible"
		}
		return providerConfigure(d, terraformVersion)
	}

	return provider
}

func providerConfigure(d *schema.ResourceData, terraformVersion string) (interface{}, error) {
	/*Redfish token issued by iDRAC needs to be revoked when the provider is done.
	At the moment, the terraform SDK (Provider.StopFunc) is not implemented. To follow up, please refer to this pull request:
	https://github.com/hashicorp/terraform-plugin-sdk/pull/377
	*/
	c, err := NewConfig(d)
	if err != nil {
		return nil, err
	}
	return c, nil
}
