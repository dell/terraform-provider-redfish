package redfish

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/stmcginnis/gofish"
	"github.com/stmcginnis/gofish/redfish"
)

func resourceRedfishVirtualMedia() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceRedfishVirtualMediaCreate,
		ReadContext:   resourceRedfishVirtualMediaRead,
		UpdateContext: resourceRedfishVirtualMediaUpdate,
		DeleteContext: resourceRedfishVirtualMediaDelete,
		Schema:        getResourceRedfishVirtualMediaSchema(),
	}
}

func getResourceRedfishVirtualMediaSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"redfish_server": {
			Type:        schema.TypeList,
			Required:    true,
			Description: "List of server BMCs and their respective user credentials",
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"user": {
						Type:        schema.TypeString,
						Optional:    true,
						Description: "User name for login",
					},
					"password": {
						Type:        schema.TypeString,
						Optional:    true,
						Description: "User password for login",
						Sensitive:   true,
					},
					"endpoint": {
						Type:        schema.TypeString,
						Required:    true,
						Description: "Server BMC IP address or hostname",
					},
					"ssl_insecure": {
						Type:        schema.TypeBool,
						Optional:    true,
						Description: "This field indicates whether the SSL/TLS certificate must be verified or not",
					},
				},
			},
		},
		"image": {
			Type:        schema.TypeString,
			Description: "The URI of the remote media to attach to the virtual media",
			Required:    true,
		},
		"inserted": {
			Type:        schema.TypeBool,
			Description: "The URI of the remote media to attach to the virtual media",
			Computed:    true,
		},
		"transfer_method": {
			Type:        schema.TypeString,
			Description: "Indicates how the data is transferred",
			Optional:    true,
			Computed:    true,
			ValidateFunc: validation.StringInSlice([]string{
				string(redfish.StreamTransferMethod),
				string(redfish.UploadTransferMethod),
			}, false),
		},
		"transfer_protocol_type": {
			Type:        schema.TypeString,
			Description: "The protocol used to transfer.",
			Optional:    true,
			Computed:    true,
			ValidateFunc: validation.StringInSlice([]string{
				string(redfish.CIFSTransferProtocolType),
				string(redfish.FTPTransferProtocolType),
				string(redfish.SFTPTransferProtocolType),
				string(redfish.HTTPTransferProtocolType),
				string(redfish.HTTPSTransferProtocolType),
				string(redfish.NFSTransferProtocolType),
				string(redfish.SCPTransferProtocolType),
				string(redfish.TFTPTransferProtocolType),
				string(redfish.OEMTransferProtocolType),
			}, false),
		},
		"write_protected": {
			Type:        schema.TypeBool,
			Description: "Indicates whether the remote device media prevents writing to that media.",
			Optional:    true,
			Computed:    true,
		},
	}
}

func resourceRedfishVirtualMediaCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	service, err := NewConfig(m.(*schema.ResourceData), d)
	if err != nil {
		return diag.Errorf(err.Error())
	}
	return createRedfishVirtualMedia(service, d)
}

func resourceRedfishVirtualMediaRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	service, err := NewConfig(m.(*schema.ResourceData), d)
	if err != nil {
		return diag.Errorf(err.Error())
	}
	return readRedfishVirtualMedia(service, d)
}

func resourceRedfishVirtualMediaUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	service, err := NewConfig(m.(*schema.ResourceData), d)
	if err != nil {
		return diag.Errorf(err.Error())
	}
	if diags := updateRedfishVirtualMedia(ctx, service, d, m); diags.HasError() {
		return diags
	}
	return resourceRedfishVirtualMediaRead(ctx, d, m)
}

func resourceRedfishVirtualMediaDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	service, err := NewConfig(m.(*schema.ResourceData), d)
	if err != nil {
		return diag.Errorf(err.Error())
	}
	return deleteRedfishVirtualMedia(service, d)
}

func createRedfishVirtualMedia(service *gofish.Service, d *schema.ResourceData) diag.Diagnostics {
	var diags diag.Diagnostics

	// Lock the mutex to avoid race conditions with other resources
	redfishMutexKV.Lock(getRedfishServerEndpoint(d))
	defer redfishMutexKV.Unlock(getRedfishServerEndpoint(d))

	//Get terraform schema data
	image := d.Get("image").(string)
	if !strings.HasSuffix(image, ".iso") && !strings.HasSuffix(image, ".img") {
		return diag.Errorf("Unable to Process the request because the value entered for the parameter Image is not supported by the implementation. Please provide an image with extension iso or img.")
	}

	var transferMethod string
	if v, ok := d.GetOk("transfer_method"); ok {
		transferMethod = v.(string)
	}
	if transferMethod == "Upload" {
		return diag.Errorf("Unable to Process the request because the value entered for the parameter TransferMethod is not supported by the implementation.")
	}
	var transferProtocolType string
	if v, ok := d.GetOk("transfer_protocol_type"); ok {
		transferProtocolType = v.(string)
	}
	var inserted bool
	if v, ok := d.GetOkExists("inserted"); ok {
		inserted = v.(bool)
	} else {
		inserted = true //If inserted is not set, set it to true
	}
	var writeProtected bool
	if v, ok := d.GetOkExists("write_protected"); ok {
		writeProtected = v.(bool)
	} else {
		writeProtected = true //If write_protected is not set, set it to true
	}

	virtualMediaConfig := redfish.VirtualMediaConfig{
		Image:                image,
		Inserted:             inserted,
		TransferMethod:       transferMethod,
		TransferProtocolType: transferProtocolType,
		WriteProtected:       writeProtected,
	}

	//Get Systems details
	systems, err := service.Systems()
	if err != nil {
		return diag.Errorf("Error when retrieving systems: %s", err)
	}
	if len(systems) == 0 {
		return diag.Errorf("There is no system available")
	}

	virtualMediaCollection, err := systems[0].VirtualMedia()
	if err != nil {
		return diag.Errorf("Couldn't retrieve virtual media collection from redfish API: %s", err)
	}

	if len(virtualMediaCollection) != 0 {
		for index := range virtualMediaCollection {
			//Get specific virtual media
			virtualMedia, err := getVirtualMedia(virtualMediaCollection[index].ID, virtualMediaCollection)
			if err != nil {
				return diag.Errorf("Virtual Media selected doesn't exist: %s", err)
			}
			if !virtualMedia.Inserted {
				err = virtualMedia.InsertMediaConfig(virtualMediaConfig)
				if err != nil {
					return diag.Errorf("Couldn't mount Virtual Media: %s", err)
				}

				d.SetId(virtualMedia.ODataID)
				diags = readRedfishVirtualMedia(service, d)
				return diags
			}
		}
	} else {
		// This implementation is added to support iDRAC firmware version 5.x. As virtual media can only be accessed through Managers card on 5.x.
		//Get OOB Manager card - managers[0] will be our oob card
		managers, err := service.Managers()
		if err != nil {
			return diag.Errorf("Couldn't retrieve managers from redfish API: %s", err)
		}

		virtualMediaCollection, err := managers[0].VirtualMedia()
		if err != nil {
			return diag.Errorf("Couldn't retrieve virtual media collection from redfish API: %s", err)
		}

		var virtualMediaID string
		if strings.HasSuffix(image, ".iso") {
			virtualMediaID = "CD"
		} else {
			virtualMediaID = "RemovableDisk"
		}

		virtualMedia, err := getVirtualMedia(virtualMediaID, virtualMediaCollection)
		if err != nil {
			return diag.Errorf("Virtual Media selected doesn't exist: %s", err)
		}
		if !virtualMedia.Inserted {
			err = virtualMedia.InsertMediaConfig(virtualMediaConfig)
			if err != nil {
				return diag.Errorf("Couldn't mount Virtual Media: %s", err)
			}

			d.SetId(virtualMedia.ODataID)
			diags = readRedfishVirtualMedia(service, d)
			return diags
		}
	}

	return diag.Errorf("There are no Virtual Medias to mount")
}

func readRedfishVirtualMedia(service *gofish.Service, d *schema.ResourceData) diag.Diagnostics {
	var diags diag.Diagnostics

	virtualMedia, err := redfish.GetVirtualMedia(service.GetClient(), d.Id())
	if err != nil {
		return diag.Errorf("Virtual Media doesn't exist: %s", err) //This error won't be triggered ever
	}

	if len(virtualMedia.Image) == 0 { //Nothing is mounted here
		d.SetId("")
		return diags
	}

	//Get terraform schema data
	image := d.Get("image").(string)

	var transferMethod string
	if v, ok := d.GetOk("transfer_method"); ok {
		transferMethod = v.(string)
	}
	var transferProtocolType string
	if v, ok := d.GetOk("transfer_protocol_type"); ok {
		transferProtocolType = v.(string)
	}
	var inserted bool
	if v, ok := d.GetOkExists("inserted"); ok {
		inserted = v.(bool)
	}
	var writeProtected bool
	if v, ok := d.GetOkExists("write_protected"); ok {
		writeProtected = v.(bool)
	}

	if virtualMedia.Image != image {
		err = d.Set("image", virtualMedia.Image)
	}
	if string(virtualMedia.TransferMethod) != transferMethod {
		err = errors.Join(err, d.Set("transfer_method", virtualMedia.TransferMethod))
	}
	if string(virtualMedia.TransferProtocolType) != transferProtocolType {
		err = errors.Join(err, d.Set("transfer_protocol_type", virtualMedia.TransferProtocolType))
	}
	if virtualMedia.Inserted != inserted {
		err = errors.Join(err, d.Set("inserted", virtualMedia.Inserted))
	}
	if virtualMedia.WriteProtected != writeProtected {
		err = errors.Join(err, d.Set("write_protected", virtualMedia.WriteProtected))
	}

	return append(diags, diag.FromErr(err)...)
}

func updateRedfishVirtualMedia(ctx context.Context, service *gofish.Service, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	// Lock the mutex to avoid race conditions with other resources
	redfishMutexKV.Lock(getRedfishServerEndpoint(d))
	defer redfishMutexKV.Unlock(getRedfishServerEndpoint(d))

	//Hot update os not possible. Unmount and mount needs to be done to update
	virtualMedia, err := redfish.GetVirtualMedia(service.GetClient(), d.Id())
	if err != nil {
		return diag.Errorf("Virtual Media doesn't exist: %s", err) //This error won't be triggered ever
	}

	err = virtualMedia.EjectMedia()
	if err != nil {
		return diag.Errorf("There was an error when ejecting media: %s", err)
	}

	//Get terraform schema data
	image := d.Get("image").(string)
	if !strings.HasSuffix(image, ".iso") && !strings.HasSuffix(image, ".img") {
		return diag.Errorf("Unable to Process the request because the value entered for the parameter Image is not supported by the implementation. Please provide an image with extension iso or img.")
	}

	var transferMethod string
	if v, ok := d.GetOk("transfer_method"); ok {
		transferMethod = v.(string)
	}
	if transferMethod == "Upload" {
		return diag.Errorf("Unable to Process the request because the value entered for the parameter TransferMethod is not supported by the implementation.")
	}
	var transferProtocolType string
	if v, ok := d.GetOk("transfer_protocol_type"); ok {
		transferProtocolType = v.(string)
	}
	var inserted bool
	if v, ok := d.GetOkExists("inserted"); ok {
		inserted = v.(bool)
	} else {
		inserted = true //If inserted is not set, set it to true
	}
	var writeProtected bool
	if v, ok := d.GetOkExists("write_protected"); ok {
		writeProtected = v.(bool)
	} else {
		writeProtected = true //If write_protected is not set, set it to true
	}

	virtualMediaConfig := redfish.VirtualMediaConfig{
		Image:                image,
		Inserted:             inserted,
		TransferMethod:       transferMethod,
		TransferProtocolType: transferProtocolType,
		WriteProtected:       writeProtected,
	}

	err = virtualMedia.InsertMediaConfig(virtualMediaConfig)
	if err != nil {
		return diag.Errorf("Couldn't mount Virtual Media: %s", err)
	}

	return diags
}

func deleteRedfishVirtualMedia(service *gofish.Service, d *schema.ResourceData) diag.Diagnostics {
	var diags diag.Diagnostics

	// Lock the mutex to avoid race conditions with other resources
	redfishMutexKV.Lock(getRedfishServerEndpoint(d))
	defer redfishMutexKV.Unlock(getRedfishServerEndpoint(d))

	virtualMedia, err := redfish.GetVirtualMedia(service.GetClient(), d.Id())
	if err != nil {
		return diag.Errorf("Virtual Media doesn't exist: %s", err) //This error won't be triggered ever
	}

	err = virtualMedia.EjectMedia()
	if err != nil {
		return diag.Errorf("There was an error when ejecting media: %s", err)
	}

	return diags
}

func getVirtualMedia(virtualMediaID string, vms []*redfish.VirtualMedia) (*redfish.VirtualMedia, error) {
	for _, v := range vms {
		if v.ID == virtualMediaID {
			return v, nil
		}
	}
	return nil, fmt.Errorf("VirtualMedia with ID %s doesn't exist", virtualMediaID)
}
