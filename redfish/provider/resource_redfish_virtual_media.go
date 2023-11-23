package provider

import (
	"context"
	"fmt"
	"log"
	"strings"
	"terraform-provider-redfish/redfish/models"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/stmcginnis/gofish"
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
	}
}

// Schema defines the schema for the resource.
func (*virtualMediaResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Resource for managing virtual media.",

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
		TransferMethod:       plan.TransferMethod.ValueString(),
		TransferProtocolType: plan.TransferProtocolType.ValueString(),
		WriteProtected:       plan.WriteProtected.ValueBool(),
	}
	// Get service
	service, err := NewConfig(r.p, &plan.RedfishServer)
	if err != nil {
		resp.Diagnostics.AddError(ServiceErrorMsg, err.Error())
		return
	}
	// Get Systems details
	systems, err := service.Systems()
	if err != nil {
		resp.Diagnostics.AddError("Error when retrieving systems", err.Error())
		return
	}
	if len(systems) == 0 {
		resp.Diagnostics.AddError("There is no system available", err.Error())
		return
	}
	// Get virtual media collection
	virtualMediaCollection, err := systems[0].VirtualMedia()
	if err != nil {
		resp.Diagnostics.AddError("Couldn't retrieve virtual media collection from redfish API", err.Error())
		return
	}
	if len(virtualMediaCollection) != 0 {
		for index := range virtualMediaCollection {
			virtualMedia, err := insertMedia(virtualMediaCollection[index].ID, virtualMediaCollection, virtualMediaConfig, service)
			if err != nil {
				resp.Diagnostics.AddError("Error while inserting virtual media", err.Error())
				return
			}
			if virtualMedia != nil {
				// Save into State
				result := models.VirtualMedia{}
				r.updateVirtualMediaState(&result, *virtualMedia, &plan)
				diags = resp.State.Set(ctx, &result)
				resp.Diagnostics.Append(diags...)
				tflog.Trace(ctx, "resource_virtual_media create: finished")
				return
			}
		}
	} else {
		// This implementation is added to support iDRAC firmware version 5.x. As virtual media can only be accessed through Managers card on 5.x.
		// Get OOB Manager card - managers[0] will be our oob card
		managers, err := service.Managers()
		if err != nil {
			resp.Diagnostics.AddError("Couldn't retrieve managers from redfish API: ", err.Error())
			return
		}
		// Get virtual media collection
		virtualMediaCollection, err := managers[0].VirtualMedia()
		if err != nil {
			resp.Diagnostics.AddError("Couldn't retrieve virtual media collection from redfish API: ", err.Error())
			return
		}
		var virtualMediaID string
		if strings.HasSuffix(plan.Image.ValueString(), ".iso") {
			virtualMediaID = "CD"
		} else {
			virtualMediaID = "RemovableDisk"
		}

		virtualMedia, err := insertMedia(virtualMediaID, virtualMediaCollection, virtualMediaConfig, service)
		if err != nil {
			resp.Diagnostics.AddError("Error while inserting virtual media", err.Error())
			return
		}
		if virtualMedia != nil {
			// Save into State
			result := models.VirtualMedia{}
			r.updateVirtualMediaState(&result, *virtualMedia, &plan)
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
	service, err := NewConfig(r.p, &state.RedfishServer)
	if err != nil {
		resp.Diagnostics.AddError(ServiceErrorMsg, err.Error())
		return
	}

	// Get virtual media details
	virtualMedia, err := redfish.GetVirtualMedia(service.GetClient(), state.VirtualMediaID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Virtual Media doesn't exist: ", err.Error()) // This error won't be triggered ever
		return
	}

	if len(virtualMedia.Image) == 0 { // Nothing is mounted here
		return
	}

	// Save into State
	result := models.VirtualMedia{}
	r.updateVirtualMediaState(&result, *virtualMedia, &state)
	diags = resp.State.Set(ctx, &result)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Trace(ctx, "resource_virtual_media read: finished")
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
	service, err := NewConfig(r.p, &plan.RedfishServer)
	if err != nil {
		resp.Diagnostics.AddError(ServiceErrorMsg, err.Error())
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
		Image:                plan.Image.ValueString(),
		Inserted:             plan.Inserted.ValueBool(),
		TransferMethod:       plan.TransferMethod.ValueString(),
		TransferProtocolType: plan.TransferProtocolType.ValueString(),
		WriteProtected:       plan.WriteProtected.ValueBool(),
	}

	virtualMediaConfigState := redfish.VirtualMediaConfig{
		Image:                state.Image.ValueString(),
		Inserted:             state.Inserted.ValueBool(),
		TransferMethod:       state.TransferMethod.ValueString(),
		TransferProtocolType: state.TransferProtocolType.ValueString(),
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
	result := models.VirtualMedia{}
	r.updateVirtualMediaState(&result, *virtualMedia, &plan)
	log.Printf("result: %v\n", result)
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
	service, err := NewConfig(r.p, &state.RedfishServer)
	if err != nil {
		resp.Diagnostics.AddError(ServiceErrorMsg, err.Error())
		return
	}

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
	result := models.VirtualMedia{}
	r.updateVirtualMediaState(&result, *virtualMedia, &state)
	diags = resp.State.Set(ctx, &result)
	resp.Diagnostics.Append(diags...)
	tflog.Trace(ctx, "resource_virtual_media delete: finished")
}

func getVirtualMedia(virtualMediaID string, vms []*redfish.VirtualMedia) (*redfish.VirtualMedia, error) {
	for _, v := range vms {
		if v.ID == virtualMediaID {
			return v, nil
		}
	}
	return nil, fmt.Errorf("VirtualMedia with ID %s doesn't exist", virtualMediaID)
}

func insertMedia(id string, collection []*redfish.VirtualMedia, config redfish.VirtualMediaConfig, s *gofish.Service) (*redfish.VirtualMedia, error) {
	virtualMedia, err := getVirtualMedia(id, collection)
	if err != nil {
		return nil, fmt.Errorf("Virtual Media selected doesn't exist: %v", err.Error())
	}
	if !virtualMedia.Inserted {
		err = virtualMedia.InsertMediaConfig(config)
		if err != nil {
			return nil, fmt.Errorf("Couldn't mount Virtual Media: %v", err.Error())
		}
		virtualMedia, err := redfish.GetVirtualMedia(s.GetClient(), virtualMedia.ODataID)
		if err != nil {
			return nil, fmt.Errorf("Virtual Media selected doesn't exist: %s", err.Error())
		}
		return virtualMedia, nil
	}
	return nil, err
}

// updateVirtualMediaState - Update virtual media details from response to state
func (virtualMediaResource) updateVirtualMediaState(state *models.VirtualMedia, response redfish.VirtualMedia, plan *models.VirtualMedia) {
	state.VirtualMediaID = types.StringValue(response.ODataID)
	state.Image = types.StringValue(response.Image)
	state.Inserted = types.BoolValue(response.Inserted)
	state.TransferMethod = types.StringValue(string(response.TransferMethod))
	state.TransferProtocolType = types.StringValue(string(response.TransferProtocolType))
	state.WriteProtected = types.BoolValue(response.WriteProtected)
	state.RedfishServer = plan.RedfishServer
}
