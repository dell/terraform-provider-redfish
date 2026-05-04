/*
Copyright (c) 2025 Dell Inc., or its subsidiaries. All Rights Reserved.

Licensed under the Mozilla Public License Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://mozilla.org/MPL/2.0/


Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package helper

import (
	"fmt"
	"terraform-provider-redfish/redfish/models"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stmcginnis/gofish"
	"github.com/stmcginnis/gofish/redfish"
)

// GetVMEnv is a helper function to get Virtual Media Environment
func GetVMEnv(service *gofish.Service, system *redfish.ComputerSystem) (
	VirtualMediaEnvironment, diag.Diagnostics,
) {
	var d diag.Diagnostics
	var env VirtualMediaEnvironment

	env.Service = service
	env.System = system
	// Get virtual media collection from system
	virtualMediaCollection, err := system.VirtualMedia()
	if err != nil {
		d.AddError("Couldn't retrieve virtual media collection from redfish API", err.Error())
		return env, d
	}
	if len(virtualMediaCollection) != 0 {
		// This happens in iDRAC 6.x and later
		env.Collection = virtualMediaCollection
		env.Manager = false
		return env, d
	}
	// This implementation is added to support iDRAC firmware version 5.x. As virtual media can only be accessed through Managers card on 5.x.
	// Get OOB Manager card - managers[0] will be our oob card
	env.Manager = true
	managers, err := service.Managers()
	if err != nil {
		d.AddError("Couldn't retrieve managers from redfish API: ", err.Error())
		return env, d
	}
	// Get virtual media collection from manager
	virtualMediaCollection, err = managers[0].VirtualMedia()
	if err != nil {
		d.AddError("Couldn't retrieve virtual media collection from redfish API: ", err.Error())
		return env, d
	}
	env.Collection = virtualMediaCollection
	return env, d
}

// GetVirtualMedia is a helper function to get the Virtual Media
func GetVirtualMedia(virtualMediaID string, vms []*redfish.VirtualMedia) (*redfish.VirtualMedia, error) {
	for _, v := range vms {
		if v.ID == virtualMediaID {
			return v, nil
		}
	}
	return nil, fmt.Errorf("virtual media with ID %s doesn't exist", virtualMediaID)
}

// InsertMedia is a helper function to indert a media
func InsertMedia(id string, collection []*redfish.VirtualMedia, config redfish.VirtualMediaConfig, s *gofish.Service) (*redfish.VirtualMedia, error) {
	virtualMedia, err := GetVirtualMedia(id, collection)
	if err != nil {
		return nil, fmt.Errorf("virtual media selected doesn't exist: %w", err)
	}
	virtualMedia.SetETag("")

	if !virtualMedia.Inserted {
		err = virtualMedia.InsertMediaConfig(config)
		if err != nil {
			return nil, fmt.Errorf("couldn't mount Virtual Media: %w", err)
		}
		virtualMedia, err := redfish.GetVirtualMedia(s.GetClient(), virtualMedia.ODataID)
		if err != nil {
			return nil, fmt.Errorf("virtual media selected doesn't exist: %w", err)
		}
		return virtualMedia, nil
	}
	return nil, err
}

// UpdateVirtualMediaState - Copy virtual media details from response to state object.
// When the server (e.g. R670) does not reflect back user-supplied fields in the GET
// response (returning empty strings / zero bools), the plan/state values are used as
// a fallback so Terraform does not see an "inconsistent result after apply" error.
func UpdateVirtualMediaState(response *redfish.VirtualMedia, plan models.VirtualMedia) models.VirtualMedia {
	image := plan.Image
	if response.Image != "" {
		image = types.StringValue(response.Image)
	}

	transferMethod := plan.TransferMethod
	if string(response.TransferMethod) != "" {
		transferMethod = types.StringValue(string(response.TransferMethod))
	}

	transferProtocolType := plan.TransferProtocolType
	if string(response.TransferProtocolType) != "" {
		transferProtocolType = types.StringValue(string(response.TransferProtocolType))
	}

	// WriteProtected has no sentinel empty value (bool), so use Image presence as an
	// indicator that the server populated all fields. When Image is empty the server
	// did not reflect the fields back and we keep the plan/state value.
	writeProtected := plan.WriteProtected
	if response.Image != "" {
		writeProtected = types.BoolValue(response.WriteProtected)
	}

	return models.VirtualMedia{
		VirtualMediaID:       types.StringValue(response.ODataID),
		Image:                image,
		Inserted:             types.BoolValue(response.Inserted),
		TransferMethod:       transferMethod,
		TransferProtocolType: transferProtocolType,
		WriteProtected:       writeProtected,
		SystemID:             plan.SystemID,
		RedfishServer:        plan.RedfishServer,
	}
}

// GetNejectVirtualMedia - is a helper function to get virtual media and eject the media
func GetNejectVirtualMedia(service *gofish.Service, uri string) (*redfish.VirtualMedia, error) {
	virtualMedia, err := redfish.GetVirtualMedia(service.GetClient(), uri)
	if err != nil {
		return nil, fmt.Errorf("virtual Media doesn't exist:  %w", err) // This error won't be triggered ever
	}
	virtualMedia.SetETag("")

	// Eject virtual media
	err = virtualMedia.EjectMedia()
	if err != nil {
		return nil, fmt.Errorf("there was an error when ejecting media: %w", err)
	}

	return virtualMedia, nil
}

// VirtualMediaEnvironment is schema for virtual media environment
type VirtualMediaEnvironment struct {
	Manager    bool
	Collection []*redfish.VirtualMedia
	Service    *gofish.Service
	System     *redfish.ComputerSystem
}

// VMediaImportConfig is the JSON configuration for importing a virtual media
type VMediaImportConfig struct {
	ServerConf
	SystemID     string `json:"system_id"`
	ID           string `json:"id"`
	RedfishAlias string `json:"redfish_alias"`
}

// ServerConf represents the common credentials in import config
type ServerConf struct {
	Username    string `json:"username"`
	Password    string `json:"password"`
	Endpoint    string `json:"endpoint"`
	SslInsecure bool   `json:"ssl_insecure"`
}
