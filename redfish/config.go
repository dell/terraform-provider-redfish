package redfish

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stmcginnis/gofish"
	"log"
)

// NewConfig function creates the needed gofish structs to query the redfish API
/*

 */
func NewConfig(provider *schema.ResourceData, resource *schema.ResourceData) (*gofish.Service, error) {
	//Get redfish connection details from resource block
	var providerUser, providerPassword string

	if v, ok := provider.GetOk("user"); ok {
		providerUser = v.(string)
	}
	if v, ok := provider.GetOk("password"); ok {
		providerPassword = v.(string)
	}

	resourceServerConfig := resource.Get("redfish_server").([]interface{}) //It must be just one element

	//Overwrite parameters (just user and password for client connection)
	//Get redfish username at resource level over provider level
	var redfishClientUser, redfishClientPass string
	if len(resourceServerConfig[0].(map[string]interface{})["user"].(string)) > 0 {
		redfishClientUser = resourceServerConfig[0].(map[string]interface{})["user"].(string)
		log.Println("Using redfish user from resource")
	} else {
		redfishClientUser = providerUser
		log.Println("Using redfish user from provider")
	}
	//Get redfish password at resource level over provider level
	if len(resourceServerConfig[0].(map[string]interface{})["password"].(string)) > 0 {
		redfishClientPass = resourceServerConfig[0].(map[string]interface{})["password"].(string)
		log.Println("Using redfish password from resource")
	} else {
		redfishClientPass = providerPassword
		log.Println("Using redfish password from provider")
	}
	//If for some reason none user or pass has been set at provider/resource level, trow an error
	if len(redfishClientUser) == 0 || len(redfishClientPass) == 0 {
		return nil, fmt.Errorf("Error. Either Redfish client username or password has not been set. Please check your configuration")
	}

	clientConfig := gofish.ClientConfig{
		Endpoint:  resourceServerConfig[0].(map[string]interface{})["endpoint"].(string),
		Username:  redfishClientUser,
		Password:  redfishClientPass,
		BasicAuth: true,
		Insecure:  resourceServerConfig[0].(map[string]interface{})["ssl_insecure"].(bool),
	}
	api, err := gofish.Connect(clientConfig)
	if err != nil {
		return nil, fmt.Errorf("Error connecting to redfish API: %v", err)
	}
	log.Printf("Connection with the redfish endpoint %v was sucessful\n", resourceServerConfig[0].(map[string]interface{})["endpoint"].(string))
	return api.Service, nil
}
