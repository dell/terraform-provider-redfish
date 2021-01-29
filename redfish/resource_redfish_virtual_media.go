package redfish

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stmcginnis/gofish"
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
			Type:        schema.TypeList,
			Description: "ID from the virtual media to be used. I.E: RemovableDisk",
			Required:    true,
		},
		"image": {
			Type:        schema.TypeList,
			Description: "The URI of the remote media to attach to the virtual media",
			Required:    true,
		},
		"username": {
			Type:        schema.TypeList,
			Description: "The username to access the image parameter-specific URI",
			Optional:    true,
		},
		"password": {
			Type:        schema.TypeList,
			Description: "The password to access the image parameter-specific URI",
			Optional:    true,
		},
		"transfer_method": {
			Type:        schema.TypeList,
			Description: "",
			Optional:    true,
		},
		"transfer_protocol_type": {
			Type:        schema.TypeList,
			Description: "",
			Optional:    true,
		},
		"write_protected": {
			Type:        schema.TypeList,
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
