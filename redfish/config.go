package redfish

import (
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/stmcginnis/gofish"
)

// NewConfig function creates the needed gofish structs to query the redfish API
func NewConfig(d *schema.ResourceData) (*gofish.APIClient, error) {
	//Check if the ssl config param has been set
	var sslMode bool
	if v, ok := d.GetOk("ssl_insecure"); ok {
		sslMode = v.(bool)
	}
	clientConfig := gofish.ClientConfig{
		Endpoint:  d.Get("redfish_endpoint").(string),
		Username:  d.Get("user").(string),
		Password:  d.Get("password").(string),
		BasicAuth: true,
		Insecure:  sslMode,
	}
	return gofish.Connect(clientConfig)
}
