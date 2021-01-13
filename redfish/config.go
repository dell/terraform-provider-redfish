package redfish

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stmcginnis/gofish"
	"log"
)

// NewConfig function creates the needed gofish structs to query the redfish API
func NewConfig(d *schema.ResourceData) (*gofish.Service, error) {
	//Slice where all API clients will be returned
	serverConfig := d.Get("redfish_server").([]interface{}) //It must be just one element
	clientConfig := gofish.ClientConfig{
		Endpoint:  serverConfig[0].(map[string]interface{})["endpoint"].(string),
		Username:  serverConfig[0].(map[string]interface{})["user"].(string),
		Password:  serverConfig[0].(map[string]interface{})["password"].(string),
		BasicAuth: true,
		Insecure:  serverConfig[0].(map[string]interface{})["ssl_insecure"].(bool),
	}
	api, err := gofish.Connect(clientConfig)
	if err != nil {
		return nil, fmt.Errorf("Error connecting to redfish API: %v", err)
	}
	log.Printf("Connection with the redfish endpoint %v was sucessful\n", serverConfig[0].(map[string]interface{})["endpoint"].(string))
	return api.Service, nil
}
