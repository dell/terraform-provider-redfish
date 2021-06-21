package redfish

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stmcginnis/gofish"
	"github.com/stmcginnis/gofish/redfish"
	//"log"
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
		"virtual_media_id": {
			Type:        schema.TypeString,
			Description: "ID from the virtual media to be used. I.E: RemovableDisk",
			Required:    true,
		},
		"image": {
			Type:        schema.TypeString,
			Description: "The URI of the remote media to attach to the virtual media",
			Required:    true,
		},
		"inserted": {
			Type:        schema.TypeBool,
			Description: "The URI of the remote media to attach to the virtual media",
			Optional:    true,
		},
		"username": {
			Type:        schema.TypeString,
			Description: "The username to access the image parameter-specific URI",
			Optional:    true,
		},
		"password": {
			Type:        schema.TypeString,
			Description: "The password to access the image parameter-specific URI",
			Optional:    true,
		},
		"transfer_method": {
			Type:        schema.TypeString,
			Description: "",
			Optional:    true,
		},
		"transfer_protocol_type": {
			Type:        schema.TypeString,
			Description: "",
			Optional:    true,
		},
		"write_protected": {
			Type:        schema.TypeBool,
			Description: "",
			Optional:    true,
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

	//Get terraform schema data
	virtualMediaID := d.Get("virtual_media_id").(string)
	image := d.Get("image").(string)

	var username string
	if v, ok := d.GetOk("username"); ok {
		username = v.(string)
	}
	var password string
	if v, ok := d.GetOk("password"); ok {
		password = v.(string)
	}
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
	} else {
		inserted = true //If inserted is not set, set it to true
	}
	var writeProtected bool
	if v, ok := d.GetOkExists("write_protected"); ok {
		writeProtected = v.(bool)
	} else {
		writeProtected = true //If write_protected is not set, set it to true
	}

	//Get OOB Manager card - managers[0] will be our oob card
	managers, err := service.Managers()
	if err != nil {
		return diag.Errorf("Couldn't retrieve managers from redfish API: %s", err)
	}

	virtualMediaCollection, err := managers[0].VirtualMedia()
	if err != nil {
		return diag.Errorf("Couldn't retrieve virtual media collection from redfish API: %s", err)
	}
	//Get specific virtual media
	virtualMedia, err := getVirtualMedia(virtualMediaID, virtualMediaCollection)
	if err != nil {
		return diag.Errorf("Virtual Media selected doesn't exist: %s", err)
	}

	virtualMediaConfig := redfish.VirtualMediaConfig{
		Image:                image,
		Inserted:             inserted,
		Password:             password,
		TransferMethod:       transferMethod,
		TransferProtocolType: transferProtocolType,
		UserName:             username,
		WriteProtected:       writeProtected,
	}

	err = virtualMedia.InsertMediaConfig(virtualMediaConfig)
	if err != nil {
		return diag.Errorf("Couldn't mount Virtual Media: %s", err)
	}

	d.SetId(virtualMedia.ODataID)
	return diags
}

func readRedfishVirtualMedia(service *gofish.Service, d *schema.ResourceData) diag.Diagnostics {
	var diags diag.Diagnostics

	virtualMedia, err := redfish.GetVirtualMedia(service.Client, d.Id())
	if err != nil {
		return diag.Errorf("Virtual Media doesn't exist: %s", err) //This error won't be triggered ever
	}

	if len(virtualMedia.Image) == 0 { //Nothing is mounted here
		d.SetId("")
		return diags
	}

	//Get terraform schema data
	image := d.Get("image").(string)

	var username string
	if v, ok := d.GetOk("username"); ok {
		username = v.(string)
	}
	var password string
	if v, ok := d.GetOk("password"); ok {
		password = v.(string)
	}
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
	} else {
		inserted = true //If inserted is not set, set it to true
	}
	var writeProtected bool
	if v, ok := d.GetOkExists("write_protected"); ok {
		writeProtected = v.(bool)
	} else {
		writeProtected = true //If write_protected is not set, set it to true
	}

	if virtualMedia.Image != image {
		d.Set("image", image)
	}
	if virtualMedia.UserName != username {
		d.Set("username", username)
	}
	if virtualMedia.Password != password {
		d.Set("password", password)
	}
	if string(virtualMedia.TransferMethod) != transferMethod {
		d.Set("transfer_method", transferMethod)
	}
	if string(virtualMedia.TransferProtocolType) != transferProtocolType {
		d.Set("transfer_protocol_type", transferProtocolType)
	}
	if virtualMedia.Inserted != inserted {
		d.Set("inserted", inserted)
	}
	if virtualMedia.WriteProtected != writeProtected {
		d.Set("write_protected", writeProtected)
	}

	return diags
}

func updateRedfishVirtualMedia(ctx context.Context, service *gofish.Service, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	//Hot update os not possible. Unmount and mount needs to be done to update
	virtualMedia, err := redfish.GetVirtualMedia(service.Client, d.Id())
	if err != nil {
		return diag.Errorf("Virtual Media doesn't exist: %s", err) //This error won't be triggered ever
	}

	err = virtualMedia.EjectMedia()
	if err != nil {
		return diag.Errorf("There was an error when ejecting media: %s", err)
	}

	//Get terraform schema data
	image := d.Get("image").(string)

	var username string
	if v, ok := d.GetOk("username"); ok {
		username = v.(string)
	}
	var password string
	if v, ok := d.GetOk("password"); ok {
		password = v.(string)
	}
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
		Password:             password,
		TransferMethod:       transferMethod,
		TransferProtocolType: transferProtocolType,
		UserName:             username,
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

	virtualMedia, err := redfish.GetVirtualMedia(service.Client, d.Id())
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
