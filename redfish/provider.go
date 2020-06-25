package redfish

import (
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func Provider() *schema.Provider {
	return &schema.Provider{
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
			"redfish_endpoint": {
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

		ResourcesMap: map[string]*schema.Resource{
			"redfish_server": resourceServer(), //Dummy resource
			"redfish_user":   resourceUser(),
		},

		ConfigureFunc: providerConfigure,
		//StopFunc: NEEDS TO BE IMPLEMENTED to revoke the redfish token
	}
}

func providerConfigure(d *schema.ResourceData) (interface{}, error) {
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
