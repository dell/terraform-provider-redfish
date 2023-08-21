package redfish

import (
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stmcginnis/gofish"
	"github.com/stmcginnis/gofish/redfish"
)

// Based on an instance of Service from the gofish library, retrieve a concrete system on which we can take action
func getSystemResource(service *gofish.Service) (*redfish.ComputerSystem, error) {

	systems, err := service.Systems()

	if err != nil {
		return nil, err
	}
	if len(systems) == 0 {
		return nil, errors.New("No computer systems found")
	}

	return systems[0], err
}

// NewConfig function creates the needed gofish structs to query the redfish API
// See https://github.com/stmcginnis/gofish for details. This function returns a Service struct which can then be
// used to make any required API calls.
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

// PowerOperation Executes a power operation against the target server. It takes four arguments. The first is the reset
// type. See the struct "ResetType" at https://github.com/stmcginnis/gofish/blob/main/redfish/computersystem.go for all
// possible options. The second is maximumWaitTime which is the maximum amount of time to wait for the server to reach
// the expected power state before considering it a failure. The third is checkInterval which is how often to check the
// server's power state for updates. The last is a pointer to a gofish.Service object with which the function can
// interact with the server. It will return a tuple consisting of the server's power state at time of return and
// diagnostics
func PowerOperation(resetType string, maximumWaitTime int, checkInterval int, service *gofish.Service) (redfish.PowerState, diag.Diagnostics) {

	var diags diag.Diagnostics

	system, err := getSystemResource(service)
	if err != nil {
		log.Printf("[ERROR]: Failed to identify system: %s", err)
		return "", diag.Errorf(err.Error())
	}

	var targetPowerState redfish.PowerState

	if resetType == "ForceOff" || resetType == "GracefulShutdown" {
		if system.PowerState == "Off" {
			log.Printf("[TRACE]: Server already powered off. No action required.")
			return redfish.OffPowerState, diags
		} else {
			targetPowerState = "Off"
		}
	}

	if resetType == "On" || resetType == "ForceOn" {
		if system.PowerState == "On" {
			log.Printf("[TRACE]: Server already powered on. No action required.")
			return redfish.OnPowerState, diags
		} else {
			targetPowerState = "On"
		}
	}

	if resetType == "ForceRestart" || resetType == "GracefulRestart" || resetType == "PowerCycle" || resetType == "Nmi" {
		// If someone asks for a reset while the server is off, change the reset type to on instead
		if system.PowerState == "Off" {
			resetType = "On"
		}
		targetPowerState = "On"
	}

	if resetType == "PushPowerButton" {
		// In case of Push Power button toggle the current state
		if system.PowerState == "Off" {
			targetPowerState = "On"
		} else {
			targetPowerState = "Off"
		}
	}

	// Run the power operation against the target server
	log.Printf("[TRACE]: Performing system.Reset(%s)", resetType)
	if err = system.Reset(redfish.ResetType(resetType)); err != nil {
		log.Printf("[WARN]: system.Reset returned an error: %s", err)
		return system.PowerState, diag.Errorf(err.Error())
	}

	// Wait for the server to be in the correct power state
	totalTime := 0
	for totalTime < maximumWaitTime {

		time.Sleep(time.Duration(checkInterval) * time.Second)
		totalTime += checkInterval
		log.Printf("[TRACE]: Total time is %d seconds. Checking power state now.", totalTime)

		system, err := getSystemResource(service)
		if err != nil {
			log.Printf("[ERROR]: Failed to identify system: %s", err)
			return system.PowerState, diag.Errorf(err.Error())
		}

		if system.PowerState == targetPowerState {
			log.Printf("[TRACE]: system.Reset successful")
			return system.PowerState, diags
		}

	}

	// If we've reached here it means the system never reached the appropriate target state
	// We will instead set the power state to whatever the current state is and return
	log.Printf("[ERROR]: The system failed to correctly update the system power!")
	return system.PowerState, diags

}

// getRedfishServerEndpoint returns the endpoint from an schema. This might be useful
// when using MutexKV, since we need a way to differentiate mutex operations
// across servers
func getRedfishServerEndpoint(resource *schema.ResourceData) string {
	resourceServerConfig := resource.Get("redfish_server").([]interface{})
	return resourceServerConfig[0].(map[string]interface{})["endpoint"].(string)
}
