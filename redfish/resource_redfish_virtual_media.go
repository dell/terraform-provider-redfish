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
		CreateContext: resourceVirtualMediaCreate,
		ReadContext:   resourceVirtualMediaRead,
		UpdateContext: resourceVirtualMediaUpdate,
		DeleteContext: resourceVirtualMediaDelete,
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
			Required:    true,
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

func resourceVirtualMediaCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	service, err := NewConfig(m.(*schema.ResourceData), d)
	if err != nil {
		return diag.Errorf(err.Error())
	}
	return createRedfishVirtualMedia(service, d)
}

func resourceVirtualMediaRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	service, err := NewConfig(m.(*schema.ResourceData), d)
	if err != nil {
		return diag.Errorf(err.Error())
	}
	return readRedfishVirtualMedia(service, d)
}

func resourceVirtualMediaUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	service, err := NewConfig(m.(*schema.ResourceData), d)
	if err != nil {
		return diag.Errorf(err.Error())
	}
	if diags := updateRedfishStorageVolume(ctx, service, d, m); diags.HasError() {
		return diags
	}
	return resourceVirtualMediaRead(ctx, d, m)
}

func resourceVirtualMediaDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
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
		return diag.Errorf("%s", err)
	}

	return diags
}

func readRedfishVirtualMedia(service *gofish.Service, d *schema.ResourceData) diag.Diagnostics {
	var diags diag.Diagnostics
	return diags
}

func updateRedfishVirtualMedia(ctx context.Context, service *gofish.Service, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	return diags
}

func deleteRedfishVirtualMedia(service *gofish.Service, d *schema.ResourceData) diag.Diagnostics {
	var diags diag.Diagnostics
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
