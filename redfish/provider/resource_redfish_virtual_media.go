/*
Copyright (c) 2023-2024 Dell Inc., or its subsidiaries. All Rights Reserved.

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

package provider

import (
	"context"
	"encoding/json"
	"strings"
	"terraform-provider-redfish/redfish/helper"
	"terraform-provider-redfish/redfish/models"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/stmcginnis/gofish/redfish"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource = &virtualMediaResource{}
)

// NewVirtualMediaResource is a helper function to simplify the provider implementation.
func NewVirtualMediaResource() resource.Resource {
	return &virtualMediaResource{}
}

// virtualMediaResource is the resource implementation.
type virtualMediaResource struct {
	p *redfishProvider
}

// Configure implements resource.ResourceWithConfigure
func (r *virtualMediaResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	r.p = req.ProviderData.(*redfishProvider)
}

// Metadata returns the resource type name.
func (*virtualMediaResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "virtual_media"
}

// VirtualMediaSchema defines the schema for the resource.
func VirtualMediaSchema() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"id": schema.StringAttribute{
			MarkdownDescription: "ID of the virtual media resource",
			Description:         "ID of the virtual media resource",
			Computed:            true,
		},
		"image": schema.StringAttribute{
			Required:            true,
			Description:         "The URI of the remote media to attach to the virtual media",
			MarkdownDescription: "The URI of the remote media to attach to the virtual media",
		},
		"inserted": schema.BoolAttribute{
			Computed:            true,
			Description:         "Describes whether virtual media is attached or detached",
			MarkdownDescription: "Describes whether virtual media is attached or detached",
		},
		"transfer_method": schema.StringAttribute{
			Optional:            true,
			Computed:            true,
			Description:         "Indicates how the data is transferred",
			MarkdownDescription: "Indicates how the data is transferred",
			Validators: []validator.String{
				stringvalidator.OneOf([]string{
					"Stream",
					"Upload",
				}...),
			},
		},
		"transfer_protocol_type": schema.StringAttribute{
			Optional:            true,
			Computed:            true,
			Description:         "The protocol used to transfer.",
			MarkdownDescription: "The protocol used to transfer.",
			Validators: []validator.String{
				stringvalidator.OneOf([]string{
					"CIFS",
					"FTP",
					"SFTP",
					"HTTP",
					"HTTPS",
					"NFS",
					"SCP",
					"TFTP",
					"OEM",
				}...),
			},
		},
		"write_protected": schema.BoolAttribute{
			Optional:            true,
			Computed:            true,
			Description:         "Indicates whether the remote device media prevents writing to that media.",
			MarkdownDescription: "Indicates whether the remote device media prevents writing to that media.",
			Default:             booldefault.StaticBool(true),
		},
		"system_id": schema.StringAttribute{
			MarkdownDescription: "System ID of the system",
			Description:         "System ID of the system",
			Computed:            true,
			Optional:            true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplaceIfConfigured(),
			},
		},
	}
}

// Schema defines the schema for the resource.
func (*virtualMediaResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "This Terraform resource is used to configure virtual media on the iDRAC Server." +
			" We can Read, Attach, Detach the virtual media or Modify the attached image using this resource.",
		Description: "This Terraform resource is used to configure virtual media on the iDRAC Server." +
			" We can Read, Attach, Detach the virtual media or Modify the attached image using this resource.",

		Attributes: VirtualMediaSchema(),
		Blocks:     RedfishServerResourceBlockMap(),
	}
}

// Create creates the resource and sets the initial Terraform state.
func (r *virtualMediaResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	tflog.Trace(ctx, "resource_virtual_media create : Started")
	// Get Plan Data
	var plan models.VirtualMedia
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	// Validate image extension
	image := plan.Image.ValueString()
	if !strings.HasSuffix(image, ".iso") && !strings.HasSuffix(image, ".img") {
		resp.Diagnostics.AddError(RedfishVirtualMediaMountError, "Unable to Process the request. Image extension should be .iso or .img")
		return
	}
	// Validate transfer method
	if plan.TransferMethod.ValueString() == "Upload" {
		resp.Diagnostics.AddError(RedfishVirtualMediaMountError, "Unable to Process the request. TransferMethod upload is not supported.")
		return
	}
	virtualMediaConfig := redfish.VirtualMediaConfig{
		Image:                image,
		Inserted:             plan.Inserted.ValueBool(),
		TransferMethod:       redfish.TransferMethod(plan.TransferMethod.ValueString()),
		TransferProtocolType: redfish.TransferProtocolType(plan.TransferProtocolType.ValueString()),
		WriteProtected:       plan.WriteProtected.ValueBool(),
	}
	api, err := NewConfig(r.p, &plan.RedfishServer)
	if err != nil {
		resp.Diagnostics.AddError(ServiceErrorMsg, err.Error())
		return
	}
	service := api.Service
	// Get Systems details
	system, err := getSystemResource(service, plan.SystemID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error when retrieving systems", err.Error())
		return
	}
	env, d := helper.GetVMEnv(service, system)

	resp.Diagnostics = append(resp.Diagnostics, d...)
	if resp.Diagnostics.HasError() {
		return
	}
	defer api.Logout()
	service, virtualMediaCollection := service, env.Collection
	if !env.Manager {
		// This implementation is added to support iDRAC firmware version 6.x/7.x.
		plan.SystemID = types.StringValue(env.System.ID)
		for index := range virtualMediaCollection {
			virtualMedia, err := helper.InsertMedia(virtualMediaCollection[index].ID, virtualMediaCollection, virtualMediaConfig, service)
			if err != nil {
				resp.Diagnostics.AddError("Error while inserting virtual media", err.Error())
				return
			}
			if virtualMedia != nil {
				// Save into State
				result := helper.UpdateVirtualMediaState(virtualMedia, plan)
				diags = resp.State.Set(ctx, &result)
				resp.Diagnostics.Append(diags...)
				tflog.Trace(ctx, "resource_virtual_media create: finished")
				return
			}
		}
	} else {
		// This implementation is added to support iDRAC firmware version 5.x. As virtual media can only be accessed through Managers card on 5.x.
		// Get OOB Manager card - managers[0] will be our oob card
		var virtualMediaID string
		plan.SystemID = types.StringValue("")
		if strings.HasSuffix(plan.Image.ValueString(), ".iso") {
			virtualMediaID = "CD"
		} else {
			virtualMediaID = "RemovableDisk"
		}

		virtualMedia, err := helper.InsertMedia(virtualMediaID, virtualMediaCollection, virtualMediaConfig, service)
		if err != nil {
			resp.Diagnostics.AddError("Error while inserting virtual media", err.Error())
			return
		}
		if virtualMedia != nil {
			// Save into State
			result := helper.UpdateVirtualMediaState(virtualMedia, plan)
			diags = resp.State.Set(ctx, &result)
			resp.Diagnostics.Append(diags...)
			tflog.Trace(ctx, "resource_virtual_media create: finished")
			return
		}
	}
	// if no virtual media is available
	resp.Diagnostics.AddError("Error: There are no Virtual Medias to mount", "Please detach media and try again")
	resp.Diagnostics.Append(diags...)
	tflog.Trace(ctx, "resource_virtual_media create: finished")
}

// Read refreshes the Terraform state with the latest data.
func (r *virtualMediaResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	tflog.Trace(ctx, "resource_virtual_media read: started")
	// Get State
	var state models.VirtualMedia
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get service
	api, err := NewConfig(r.p, &state.RedfishServer)
	if err != nil {
		resp.Diagnostics.AddError(ServiceErrorMsg, err.Error())
		return
	}
	service := api.Service
	defer api.Logout()

	// Get virtual media details
	virtualMedia, err := redfish.GetVirtualMedia(service.GetClient(), state.VirtualMediaID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Virtual Media doesn't exist: ", err.Error())
		return
	}

	if len(virtualMedia.Image) == 0 { // Nothing is mounted here
		return
	}

	// Save into State
	result := helper.UpdateVirtualMediaState(virtualMedia, state)
	diags = resp.State.Set(ctx, &result)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Trace(ctx, "resource_virtual_media read: finished")
}

// VMediaImportConfig is the JSON configuration for importing a virtual media
type VMediaImportConfig struct {
	helper.ServerConf
	SystemID     string `json:"system_id"`
	ID           string `json:"id"`
	RedfishAlias string `json:"redfish_alias"`
}

// ImportState is the RPC called to import state for existing Virtual Media
func (r *virtualMediaResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	var c VMediaImportConfig
	err := json.Unmarshal([]byte(req.ID), &c)
	if err != nil {
		resp.Diagnostics.AddError("Error while unmarshalling import configuration", err.Error())
		return
	}

	server := models.RedfishServer{
		User:         types.StringValue(c.Username),
		Password:     types.StringValue(c.Password),
		Endpoint:     types.StringValue(c.Endpoint),
		SslInsecure:  types.BoolValue(c.SslInsecure),
		RedfishAlias: types.StringValue(c.RedfishAlias),
	}

	creds := []models.RedfishServer{server}

	api, err := NewConfig(r.p, &creds)
	if err != nil {
		resp.Diagnostics.AddError(ServiceErrorMsg, err.Error())
		return
	}
	service := api.Service
	// Get Systems details
	system, err := getSystemResource(service, c.SystemID)
	if err != nil {
		resp.Diagnostics.AddError("Error when retrieving systems", err.Error())
		return
	}
	env, d := helper.GetVMEnv(service, system)
	resp.Diagnostics = append(resp.Diagnostics, d...)
	if resp.Diagnostics.HasError() {
		return
	}
	defer api.Logout()

	// get virtual media with given ID
	var media *redfish.VirtualMedia
	for _, vm := range env.Collection {
		if vm.ODataID == c.ID {
			media = vm
			break
		}
	}
	if media == nil {
		resp.Diagnostics.AddError("Virtual Media with ID "+c.ID+" doesn't exist.", "")
		return
	}

	// check if virtual media is mounted
	if len(media.Image) == 0 { // Nothing is mounted here
		resp.Diagnostics.AddError("Virtual Media with ID "+c.ID+" is not mounted.", "")
		return
	}

	// Save into State
	result := helper.UpdateVirtualMediaState(media, models.VirtualMedia{
		RedfishServer: creds,
	})
	diags := resp.State.Set(ctx, &result)
	resp.Diagnostics.Append(diags...)
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *virtualMediaResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Lock the mutex to avoid race conditions with other resources
	// redfishMutexKV.Lock(getRedfishServerEndpoint(d))
	// defer redfishMutexKV.Unlock(getRedfishServerEndpoint(d))

	tflog.Trace(ctx, "resource_virtual_media update: started")
	// Get state Data
	var plan, state models.VirtualMedia
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get plan Data
	diags = req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get service
	api, err := NewConfig(r.p, &plan.RedfishServer)
	if err != nil {
		resp.Diagnostics.AddError(ServiceErrorMsg, err.Error())
		return
	}
	service := api.Service
	defer api.Logout()

	// Validate image extension
	image := plan.Image.ValueString()
	if !strings.HasSuffix(image, ".iso") && !strings.HasSuffix(image, ".img") {
		resp.Diagnostics.AddError(RedfishVirtualMediaMountError, "Unable to Process the request. Image extension should be .iso or .img")
		return
	}

	// Validate transfer method
	if plan.TransferMethod.ValueString() == "Upload" {
		resp.Diagnostics.AddError(RedfishVirtualMediaMountError, "Unable to Process the request. TransferMethod upload is not supported.")
		return
	}

	virtualMediaConfig := redfish.VirtualMediaConfig{
		Image:                plan.Image.ValueString(),
		Inserted:             plan.Inserted.ValueBool(),
		TransferMethod:       redfish.TransferMethod(plan.TransferMethod.ValueString()),
		TransferProtocolType: redfish.TransferProtocolType(plan.TransferProtocolType.ValueString()),
		WriteProtected:       plan.WriteProtected.ValueBool(),
	}

	virtualMediaConfigState := redfish.VirtualMediaConfig{
		Image:                state.Image.ValueString(),
		Inserted:             state.Inserted.ValueBool(),
		TransferMethod:       redfish.TransferMethod(state.TransferMethod.ValueString()),
		TransferProtocolType: redfish.TransferProtocolType(state.TransferProtocolType.ValueString()),
		WriteProtected:       state.WriteProtected.ValueBool(),
	}

	// Hot update is not possible. Unmount and mount needs to be done to update
	virtualMedia, err := redfish.GetVirtualMedia(service.GetClient(), state.VirtualMediaID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Virtual Media doesn't exist: %s", err.Error()) // This error won't be triggered ever
		return
	}

	err = virtualMedia.EjectMedia()
	if err != nil {
		resp.Diagnostics.AddError("There was an error when ejecting media: ", err.Error())
		return
	}

	err = virtualMedia.InsertMediaConfig(virtualMediaConfig)
	if err != nil {
		resp.Diagnostics.AddError("Couldn't mount Virtual Media ", err.Error())
		// if insert media fails, again performing insert with original(state) config
		err = virtualMedia.InsertMediaConfig(virtualMediaConfigState)
		if err != nil {
			resp.Diagnostics.AddError("Couldn't mount Virtual Media ", err.Error())
			return
		}
		return
	}

	// Get virtual media details
	virtualMedia, err = redfish.GetVirtualMedia(service.GetClient(), state.VirtualMediaID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Virtual Media doesn't exist: %s", err.Error()) // This error won't be triggered ever
		return
	}

	// Save into State
	result := helper.UpdateVirtualMediaState(virtualMedia, state)
	diags = resp.State.Set(ctx, &result)
	resp.Diagnostics.Append(diags...)
	tflog.Trace(ctx, "resource_virtual_media update: finished")
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *virtualMediaResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	tflog.Trace(ctx, "resource_virtual_media delete: started")
	// Get State Data
	var state models.VirtualMedia
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get service
	api, err := NewConfig(r.p, &state.RedfishServer)
	if err != nil {
		resp.Diagnostics.AddError(ServiceErrorMsg, err.Error())
		return
	}
	service := api.Service
	defer api.Logout()

	// Get virtual media details
	virtualMedia, err := redfish.GetVirtualMedia(service.GetClient(), state.VirtualMediaID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Virtual Media doesn't exist: ", err.Error()) // This error won't be triggered ever
		return
	}

	// Eject virtual media
	err = virtualMedia.EjectMedia()
	if err != nil {
		resp.Diagnostics.AddError("There was an error when ejecting media: ", err.Error())
		return
	}

	// Save into State
	result := helper.UpdateVirtualMediaState(virtualMedia, state)
	diags = resp.State.Set(ctx, &result)
	resp.Diagnostics.Append(diags...)
	tflog.Trace(ctx, "resource_virtual_media delete: finished")
}
