package provider

import (
	"context"
	"errors"
	"fmt"
	"log"
	"terraform-provider-redfish/redfish/models"
	"time"

	datasourceSchema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	resourceSchema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/stmcginnis/gofish"
	"github.com/stmcginnis/gofish/redfish"
)

// RedfishServerSchema to construct schema of redfish server
func RedfishServerSchema() map[string]resourceSchema.Attribute {
	return map[string]resourceSchema.Attribute{
		"user": resourceSchema.StringAttribute{
			Optional:    true,
			Description: "User name for login",
		},
		"password": resourceSchema.StringAttribute{
			Optional:    true,
			Description: "User password for login",
			Sensitive:   true,
		},
		"endpoint": resourceSchema.StringAttribute{
			Required:    true,
			Description: "Server BMC IP address or hostname",
		},
		"validate_cert": resourceSchema.BoolAttribute{
			Optional:    true,
			Description: "This field indicates whether the SSL/TLS certificate must be verified or not",
		},
	}
}

// RedfishServerDatasourceSchema to construct schema of redfish server
func RedfishServerDatasourceSchema() map[string]datasourceSchema.Attribute {
	return map[string]datasourceSchema.Attribute{
		"user": datasourceSchema.StringAttribute{
			Optional:    true,
			Description: "User name for login",
		},
		"password": datasourceSchema.StringAttribute{
			Optional:    true,
			Description: "User password for login",
			Sensitive:   true,
		},
		"endpoint": datasourceSchema.StringAttribute{
			Required:    true,
			Description: "Server BMC IP address or hostname",
		},
		"validate_cert": datasourceSchema.BoolAttribute{
			Optional:    true,
			Description: "This field indicates whether the SSL/TLS certificate must be verified or not",
		},
	}
}

// Based on an instance of Service from the gofish library, retrieve a concrete system on which we can take action
func getSystemResource(service *gofish.Service) (*redfish.ComputerSystem, error) {
	systems, err := service.Systems()
	if err != nil {
		return nil, err
	}
	if len(systems) == 0 {
		return nil, errors.New("no computer systems found")
	}

	return systems[0], err
}

// NewConfig function creates the needed gofish structs to query the redfish API
// See https://github.com/stmcginnis/gofish for details. This function returns a Service struct which can then be
// used to make any required API calls.
func NewConfig(pconfig *redfishProvider, rserver *models.RedfishServer) (*gofish.Service, error) {
	var redfishClientUser, redfishClientPass string

	if len(rserver.User.ValueString()) > 0 {
		redfishClientUser = rserver.User.ValueString()
	} else if len(pconfig.Username) > 0 {
		redfishClientUser = pconfig.Username
	} else {
		return nil, fmt.Errorf("error. Either provide username at provider level or resource level. Please check your configuration")
	}

	if len(rserver.Password.ValueString()) > 0 {
		redfishClientPass = rserver.Password.ValueString()
	} else if len(pconfig.Password) > 0 {
		redfishClientPass = pconfig.Password
	} else {
		return nil, fmt.Errorf("error. Either provide password at provider level or resource level. Please check your configuration")
	}

	if len(redfishClientUser) == 0 || len(redfishClientPass) == 0 {
		return nil, fmt.Errorf("error. Either Redfish client username or password has not been set. Please check your configuration")
	}

	clientConfig := gofish.ClientConfig{
		Endpoint:  rserver.Endpoint.ValueString(),
		Username:  redfishClientUser,
		Password:  redfishClientPass,
		BasicAuth: true,
		Insecure:  !rserver.ValidateCert.ValueBool(),
	}
	api, err := gofish.Connect(clientConfig)
	if err != nil {
		return nil, fmt.Errorf("error connecting to redfish API: %v", err)
	}
	log.Printf("Connection with the redfish endpoint %v was sucessful\n", rserver.Endpoint.ValueString())
	return api.Service, nil
}

type powerOperator struct {
	ctx     context.Context
	service *gofish.Service
}

// PowerOperation Executes a power operation against the target server. It takes four arguments. The first is the reset
// type. See the struct "ResetType" at https://github.com/stmcginnis/gofish/blob/main/redfish/computersystem.go for all
// possible options. The second is maximumWaitTime which is the maximum amount of time to wait for the server to reach
// the expected power state before considering it a failure. The third is checkInterval which is how often to check the
// server's power state for updates. The last is a pointer to a gofish.Service object with which the function can
// interact with the server. It will return a tuple consisting of the server's power state at time of return and
// diagnostics
func (p powerOperator) PowerOperation(resetType string, maximumWaitTime int64, checkInterval int64) (redfish.PowerState, error) {
	const powerON redfish.PowerState = "On"
	const powerOFF redfish.PowerState = "Off"
	system, err := getSystemResource(p.service)
	if err != nil {
		tflog.Error(p.ctx, fmt.Sprintf("Failed to identify system: %s", err))
		return "", fmt.Errorf("failed to identify system: %w", err)
	}

	var targetPowerState redfish.PowerState

	if resetType == "ForceOff" || resetType == "GracefulShutdown" {
		if system.PowerState == powerOFF {
			tflog.Trace(p.ctx, "Server already powered off. No action required.")
			return redfish.OffPowerState, nil
		}
		targetPowerState = powerOFF
	}

	if resetType == "On" || resetType == "ForceOn" {
		if system.PowerState == powerON {
			tflog.Trace(p.ctx, "Server already powered on. No action required")
			return redfish.OnPowerState, nil
		}
		targetPowerState = powerON
	}

	if resetType == "ForceRestart" || resetType == "GracefulRestart" || resetType == "PowerCycle" || resetType == "Nmi" {
		// If someone asks for a reset while the server is off, change the reset type to on instead
		if system.PowerState == powerOFF {
			resetType = "On"
		} else {
			targetPowerState = powerON
		}
	}

	if resetType == "PushPowerButton" {
		// In case of Push Power button toggle the current state
		if system.PowerState == powerOFF {
			targetPowerState = powerON
		} else {
			targetPowerState = powerOFF
		}
	}

	// Run the power operation against the target server
	tflog.Trace(p.ctx, fmt.Sprintf("Performing system.Reset(%s)", resetType))
	if err = system.Reset(redfish.ResetType(resetType)); err != nil {
		tflog.Warn(p.ctx, fmt.Sprintf("system.Reset returned an error: %s", err))
		return system.PowerState, err
	}

	// Wait for the server to be in the correct power state
	var totalTime int64
	for totalTime < maximumWaitTime {
		time.Sleep(time.Duration(checkInterval) * time.Second)
		totalTime += checkInterval
		tflog.Trace(p.ctx, fmt.Sprintf("Total time is %d seconds. Checking power state now.", totalTime))

		system, err := getSystemResource(p.service)
		if err != nil {
			tflog.Error(p.ctx, fmt.Sprintf("Failed to identify system: %s", err))
			return targetPowerState, err
		}
		if system.PowerState == targetPowerState {
			tflog.Debug(p.ctx, "system.Reset successful")
			return system.PowerState, nil
		}
	}

	// If we've reached here it means the system never reached the appropriate target state
	// We will instead set the power state to whatever the current state is and return
	// TODO : Change to warning when updated to plugin framework
	tflog.Warn(p.ctx, "The system failed to update the server's power status within the maximum wait time specified!")
	return system.PowerState, nil
}