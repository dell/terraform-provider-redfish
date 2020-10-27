package redfish

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stmcginnis/gofish"
	"log"
)

// ClientConfig is a struct created to hold the redfish endpoint as well as the API for keeping track of the subresources created
type ClientConfig struct {
	Endpoint string
	API      *gofish.APIClient
}

// NewConfig function creates the needed gofish structs to query the redfish API
func NewConfig(d *schema.ResourceData) ([]*ClientConfig, error) {
	//Slice where all API clients will be returned
	var clients []*ClientConfig
	serverConfigs := d.Get("redfish_server").([]interface{})
	for _, v := range serverConfigs {
		clientConfig := gofish.ClientConfig{
			Endpoint:  v.(map[string]interface{})["endpoint"].(string),
			Username:  v.(map[string]interface{})["user"].(string),
			Password:  v.(map[string]interface{})["password"].(string),
			BasicAuth: true,
			Insecure:  v.(map[string]interface{})["ssl_insecure"].(bool),
		}
		api, err := gofish.Connect(clientConfig)
		if err != nil {
			return nil, fmt.Errorf("Error connecting to redfish API: %v", err)
		}
		log.Printf("Connection with the redfish endpoint %v was sucessful\n", v.(map[string]interface{})["endpoint"].(string))
		clients = append(clients, &ClientConfig{Endpoint: v.(map[string]interface{})["endpoint"].(string), API: api})
	}
	return clients, nil
}
