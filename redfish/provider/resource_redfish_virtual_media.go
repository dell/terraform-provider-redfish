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

const (
	serviceError = "service error"
	mountError   = "Couldn't mount Virtual Media"
)

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
		"redfish_server": schema.SingleNestedAttribute{
			MarkdownDescription: "Redfish Server",
			Description:         "Redfish Server",
			Required:            true,
			Attributes:          RedfishServerSchema(),
		},
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
		Version:             1,
		Attributes:          VirtualMediaSchema(),
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
		resp.Diagnostics.AddError(mountError, "Unable to Process the request. Image extension should be .iso or .img")
		return
	}
	// Validate transfer method
	if plan.TransferMethod.ValueString() == "Upload" {
		resp.Diagnostics.AddError(mountError, "Unable to Process the request. TransferMethod upload is not supported.")
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
		resp.Diagnostics.AddError(serviceError, err.Error())
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
			// Get specific virtual media
			virtualMedia, err := getVirtualMedia(virtualMediaCollection[index].ID, virtualMediaCollection)
			if err != nil {
				resp.Diagnostics.AddError("Virtual Media selected doesn't exist: %s", err.Error())
				return
			}
			if !virtualMedia.Inserted {
				err = virtualMedia.InsertMediaConfig(virtualMediaConfig)
				if err != nil {
					resp.Diagnostics.AddError(mountError, err.Error())
					return
				}
				// Get virtual media details
				virtualMedia, err := redfish.GetVirtualMedia(service.GetClient(), virtualMedia.ODataID)
				if err != nil {
					resp.Diagnostics.AddError("Virtual Media selected doesn't exist: %s", err.Error())
					return
				}
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
		// Get specific virtual media
		virtualMedia, err := getVirtualMedia(virtualMediaID, virtualMediaCollection)
		if err != nil {
			resp.Diagnostics.AddError("Virtual Media selected doesn't exist: ", err.Error())
			return
		}
		if !virtualMedia.Inserted {
			err = virtualMedia.InsertMediaConfig(virtualMediaConfig)
			if err != nil {
				resp.Diagnostics.AddError("Couldn't mount Virtual Media ", err.Error())
				return
			}
			// Get virtual media details
			virtualMedia, err := redfish.GetVirtualMedia(service.GetClient(), virtualMedia.ODataID)
			if err != nil {
				resp.Diagnostics.AddError("Virtual Media selected doesn't exist: %s", err.Error())
				return
			}
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
	tflog.Trace(ctx, "resource_power update: finished")
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
		resp.Diagnostics.AddError(serviceError, err.Error())
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
	service, err := NewConfig(r.p, &state.RedfishServer)
	if err != nil {
		resp.Diagnostics.AddError(serviceError, err.Error())
		return
	}

	// Validate image extension
	image := plan.Image.ValueString()
	if !strings.HasSuffix(image, ".iso") && !strings.HasSuffix(image, ".img") {
		resp.Diagnostics.AddError(mountError, "Unable to Process the request. Image extension should be .iso or .img")
		return
	}

	// Validate transfer method
	if plan.TransferMethod.ValueString() == "Upload" {
		resp.Diagnostics.AddError(mountError, "Unable to Process the request. TransferMethod upload is not supported.")
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
		resp.Diagnostics.AddError(serviceError, err.Error())
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

// func getResourceRedfishVirtualMediaSchema() map[string]*schema.Schema {
// 	return map[string]*schema.Schema{
// 		"redfish_server": {
// 			Type:        schema.TypeList,
// 			Required:    true,
// 			Description: "List of server BMCs and their respective user credentials",
// 			Elem: &schema.Resource{
// 				Schema: map[string]*schema.Schema{
// 					"user": {
// 						Type:        schema.TypeString,
// 						Optional:    true,
// 						Description: "User name for login",
// 					},
// 					"password": {
// 						Type:        schema.TypeString,
// 						Optional:    true,
// 						Description: "User password for login",
// 						Sensitive:   true,
// 					},
// 					"endpoint": {
// 						Type:        schema.TypeString,
// 						Required:    true,
// 						Description: "Server BMC IP address or hostname",
// 					},
// 					"ssl_insecure": {
// 						Type:        schema.TypeBool,
// 						Optional:    true,
// 						Description: "This field indicates whether the SSL/TLS certificate must be verified or not",
// 					},
// 				},
// 			},
// 		},
// 		"image": {
// 			Type:        schema.TypeString,
// 			Description: "The URI of the remote media to attach to the virtual media",
// 			Required:    true,
// 		},
// 		"inserted": {
// 			Type:        schema.TypeBool,
// 			Description: "The URI of the remote media to attach to the virtual media",
// 			Computed:    true,
// 		},
// 		"transfer_method": {
// 			Type:        schema.TypeString,
// 			Description: "Indicates how the data is transferred",
// 			Optional:    true,
// 			Computed:    true,
// 			ValidateFunc: validation.StringInSlice([]string{
// 				string(redfish.StreamTransferMethod),
// 				string(redfish.UploadTransferMethod),
// 			}, false),
// 		},
// 		"transfer_protocol_type": {
// 			Type:        schema.TypeString,
// 			Description: "The protocol used to transfer.",
// 			Optional:    true,
// 			Computed:    true,
// 			ValidateFunc: validation.StringInSlice([]string{
// 				string(redfish.CIFSTransferProtocolType),
// 				string(redfish.FTPTransferProtocolType),
// 				string(redfish.SFTPTransferProtocolType),
// 				string(redfish.HTTPTransferProtocolType),
// 				string(redfish.HTTPSTransferProtocolType),
// 				string(redfish.NFSTransferProtocolType),
// 				string(redfish.SCPTransferProtocolType),
// 				string(redfish.TFTPTransferProtocolType),
// 				string(redfish.OEMTransferProtocolType),
// 			}, false),
// 		},
// 		"write_protected": {
// 			Type:        schema.TypeBool,
// 			Description: "Indicates whether the remote device media prevents writing to that media.",
// 			Optional:    true,
// 			Computed:    true,
// 		},
// 	}
// }

// func resourceRedfishVirtualMediaCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
// 	service, err := NewConfig(m.(*schema.ResourceData), d)
// 	if err != nil {
// 		return diag.Errorf(err.Error())
// 	}
// 	return createRedfishVirtualMedia(service, d)
// }

// func resourceRedfishVirtualMediaRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
// 	service, err := NewConfig(m.(*schema.ResourceData), d)
// 	if err != nil {
// 		return diag.Errorf(err.Error())
// 	}
// 	return readRedfishVirtualMedia(service, d)
// }

// func resourceRedfishVirtualMediaUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
// 	service, err := NewConfig(m.(*schema.ResourceData), d)
// 	if err != nil {
// 		return diag.Errorf(err.Error())
// 	}
// 	if diags := updateRedfishVirtualMedia(ctx, service, d, m); diags.HasError() {
// 		return diags
// 	}
// 	return resourceRedfishVirtualMediaRead(ctx, d, m)
// }

// func resourceRedfishVirtualMediaDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
// 	service, err := NewConfig(m.(*schema.ResourceData), d)
// 	if err != nil {
// 		return diag.Errorf(err.Error())
// 	}
// 	return deleteRedfishVirtualMedia(service, d)
// }

// func createRedfishVirtualMedia(service *gofish.Service, d *schema.ResourceData) diag.Diagnostics {
// 	var diags diag.Diagnostics

// 	// Lock the mutex to avoid race conditions with other resources
// 	redfishMutexKV.Lock(getRedfishServerEndpoint(d))
// 	defer redfishMutexKV.Unlock(getRedfishServerEndpoint(d))

// 	//Get terraform schema data
// 	image := d.Get("image").(string)
// 	if !strings.HasSuffix(image, ".iso") && !strings.HasSuffix(image, ".img") {
// 		return diag.Errorf("Unable to Process the request.")
// 	}

// 	var transferMethod string
// 	if v, ok := d.GetOk("transfer_method"); ok {
// 		transferMethod = v.(string)
// 	}
// 	if transferMethod == "Upload" {
// 		return diag.Errorf("Unable to Process the request.")
// 	}
// 	var transferProtocolType string
// 	if v, ok := d.GetOk("transfer_protocol_type"); ok {
// 		transferProtocolType = v.(string)
// 	}
// 	var inserted bool
// 	if v, ok := d.GetOkExists("inserted"); ok {
// 		inserted = v.(bool)
// 	} else {
// 		inserted = true //If inserted is not set, set it to true
// 	}
// 	var writeProtected bool
// 	if v, ok := d.GetOkExists("write_protected"); ok {
// 		writeProtected = v.(bool)
// 	} else {
// 		writeProtected = true //If write_protected is not set, set it to true
// 	}

// 	virtualMediaConfig := redfish.VirtualMediaConfig{
// 		Image:                image,
// 		Inserted:             inserted,
// 		TransferMethod:       transferMethod,
// 		TransferProtocolType: transferProtocolType,
// 		WriteProtected:       writeProtected,
// 	}

// 	//Get Systems details
// 	systems, err := service.Systems()
// 	if err != nil {
// 		return diag.Errorf("Error when retrieving systems: %s", err)
// 	}
// 	if len(systems) == 0 {
// 		return diag.Errorf("There is no system available")
// 	}

// 	virtualMediaCollection, err := systems[0].VirtualMedia()
// 	if err != nil {
// 		return diag.Errorf("Couldn't retrieve virtual media collection from redfish API: %s", err)
// 	}

// 	if len(virtualMediaCollection) != 0 {
// 		for index := range virtualMediaCollection {
// 			//Get specific virtual media
// 			virtualMedia, err := getVirtualMedia(virtualMediaCollection[index].ID, virtualMediaCollection)
// 			if err != nil {
// 				return diag.Errorf("Virtual Media selected doesn't exist: %s", err)
// 			}
// 			if !virtualMedia.Inserted {
// 				err = virtualMedia.InsertMediaConfig(virtualMediaConfig)
// 				if err != nil {
// 					return diag.Errorf("Couldn't mount Virtual Media: %s", err)
// 				}

// 				d.SetId(virtualMedia.ODataID)
// 				diags = readRedfishVirtualMedia(service, d)
// 				return diags
// 			}
// 		}
// 	} else {
// 		// This implementation is added to support iDRAC firmware version 5.x. As virtual media can only be accessed through Managers card on 5.x.
// 		//Get OOB Manager card - managers[0] will be our oob card
// 		managers, err := service.Managers()
// 		if err != nil {
// 			return diag.Errorf("Couldn't retrieve managers from redfish API: %s", err)
// 		}

// 		virtualMediaCollection, err := managers[0].VirtualMedia()
// 		if err != nil {
// 			return diag.Errorf("Couldn't retrieve virtual media collection from redfish API: %s", err)
// 		}

// 		var virtualMediaID string
// 		if strings.HasSuffix(image, ".iso") {
// 			virtualMediaID = "CD"
// 		} else {
// 			virtualMediaID = "RemovableDisk"
// 		}

// 		virtualMedia, err := getVirtualMedia(virtualMediaID, virtualMediaCollection)
// 		if err != nil {
// 			return diag.Errorf("Virtual Media selected doesn't exist: %s", err)
// 		}
// 		if !virtualMedia.Inserted {
// 			err = virtualMedia.InsertMediaConfig(virtualMediaConfig)
// 			if err != nil {
// 				return diag.Errorf("Couldn't mount Virtual Media: %s", err)
// 			}

// 			d.SetId(virtualMedia.ODataID)
// 			diags = readRedfishVirtualMedia(service, d)
// 			return diags
// 		}
// 	}

// 	return diag.Errorf("There are no Virtual Medias to mount")
// }

// func readRedfishVirtualMedia(service *gofish.Service, d *schema.ResourceData) diag.Diagnostics {
// 	var diags diag.Diagnostics

// 	virtualMedia, err := redfish.GetVirtualMedia(service.GetClient(), d.Id())
// 	if err != nil {
// 		return diag.Errorf("Virtual Media doesn't exist: %s", err) //This error won't be triggered ever
// 	}

// 	if len(virtualMedia.Image) == 0 { //Nothing is mounted here
// 		d.SetId("")
// 		return diags
// 	}

// 	//Get terraform schema data
// 	image := d.Get("image").(string)

// 	var transferMethod string
// 	if v, ok := d.GetOk("transfer_method"); ok {
// 		transferMethod = v.(string)
// 	}
// 	var transferProtocolType string
// 	if v, ok := d.GetOk("transfer_protocol_type"); ok {
// 		transferProtocolType = v.(string)
// 	}
// 	var inserted bool
// 	if v, ok := d.GetOkExists("inserted"); ok {
// 		inserted = v.(bool)
// 	}
// 	var writeProtected bool
// 	if v, ok := d.GetOkExists("write_protected"); ok {
// 		writeProtected = v.(bool)
// 	}

// 	if virtualMedia.Image != image {
// 		d.Set("image", virtualMedia.Image)
// 	}
// 	if string(virtualMedia.TransferMethod) != transferMethod {
// 		d.Set("transfer_method", virtualMedia.TransferMethod)
// 	}
// 	if string(virtualMedia.TransferProtocolType) != transferProtocolType {
// 		d.Set("transfer_protocol_type", virtualMedia.TransferProtocolType)
// 	}
// 	if virtualMedia.Inserted != inserted {
// 		d.Set("inserted", virtualMedia.Inserted)
// 	}
// 	if virtualMedia.WriteProtected != writeProtected {
// 		d.Set("write_protected", virtualMedia.WriteProtected)
// 	}

// 	return diags
// }

// func updateRedfishVirtualMedia(ctx context.Context, service *gofish.Service, d *schema.ResourceData, m interface{}) diag.Diagnostics {
// 	var diags diag.Diagnostics

// 	// Lock the mutex to avoid race conditions with other resources
// 	redfishMutexKV.Lock(getRedfishServerEndpoint(d))
// 	defer redfishMutexKV.Unlock(getRedfishServerEndpoint(d))

// 	//Hot update os not possible. Unmount and mount needs to be done to update
// 	virtualMedia, err := redfish.GetVirtualMedia(service.GetClient(), d.Id())
// 	if err != nil {
// 		return diag.Errorf("Virtual Media doesn't exist: %s", err) //This error won't be triggered ever
// 	}

// 	err = virtualMedia.EjectMedia()
// 	if err != nil {
// 		return diag.Errorf("There was an error when ejecting media: %s", err)
// 	}

// 	//Get terraform schema data
// 	image := d.Get("image").(string)
// 	if !strings.HasSuffix(image, ".iso") && !strings.HasSuffix(image, ".img") {
// 		return diag.Errorf("Unable to Process the request.")
// 	}

// 	var transferMethod string
// 	if v, ok := d.GetOk("transfer_method"); ok {
// 		transferMethod = v.(string)
// 	}
// 	if transferMethod == "Upload" {
// 		return diag.Errorf("Unable to Process the request.")
// 	}
// 	var transferProtocolType string
// 	if v, ok := d.GetOk("transfer_protocol_type"); ok {
// 		transferProtocolType = v.(string)
// 	}
// 	var inserted bool
// 	if v, ok := d.GetOkExists("inserted"); ok {
// 		inserted = v.(bool)
// 	} else {
// 		inserted = true //If inserted is not set, set it to true
// 	}
// 	var writeProtected bool
// 	if v, ok := d.GetOkExists("write_protected"); ok {
// 		writeProtected = v.(bool)
// 	} else {
// 		writeProtected = true //If write_protected is not set, set it to true
// 	}

// 	virtualMediaConfig := redfish.VirtualMediaConfig{
// 		Image:                image,
// 		Inserted:             inserted,
// 		TransferMethod:       transferMethod,
// 		TransferProtocolType: transferProtocolType,
// 		WriteProtected:       writeProtected,
// 	}

// 	err = virtualMedia.InsertMediaConfig(virtualMediaConfig)
// 	if err != nil {
// 		return diag.Errorf("Couldn't mount Virtual Media: %s", err)
// 	}

// 	return diags
// }

// func deleteRedfishVirtualMedia(service *gofish.Service, d *schema.ResourceData) diag.Diagnostics {
// 	var diags diag.Diagnostics

// 	// Lock the mutex to avoid race conditions with other resources
// 	redfishMutexKV.Lock(getRedfishServerEndpoint(d))
// 	defer redfishMutexKV.Unlock(getRedfishServerEndpoint(d))

// 	virtualMedia, err := redfish.GetVirtualMedia(service.GetClient(), d.Id())
// 	if err != nil {
// 		return diag.Errorf("Virtual Media doesn't exist: %s", err) //This error won't be triggered ever
// 	}

// 	err = virtualMedia.EjectMedia()
// 	if err != nil {
// 		return diag.Errorf("There was an error when ejecting media: %s", err)
// 	}

// 	return diags
// }

// func getVirtualMedia(virtualMediaID string, vms []*redfish.VirtualMedia) (*redfish.VirtualMedia, error) {
// 	for _, v := range vms {
// 		if v.ID == virtualMediaID {
// 			return v, nil
// 		}
// 	}
// 	return nil, fmt.Errorf("VirtualMedia with ID %s doesn't exist", virtualMediaID)
// }
