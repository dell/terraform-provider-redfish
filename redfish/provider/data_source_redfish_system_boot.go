package provider

// import (
// 	"context"

// 	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
// 	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
// 	"github.com/stmcginnis/gofish"
// 	"github.com/stmcginnis/gofish/redfish"
// )

// func dataSourceRedfishSystemBoot() *schema.Resource {
// 	return &schema.Resource{
// 		ReadContext: dataSourceRedfishSystemBootRead,
// 		Schema:      getDataSourceRedfishSystemBootSchema(),
// 	}
// }

// func getDataSourceRedfishSystemBootSchema() map[string]*schema.Schema {
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
// 		"resource_id": {
// 			Type:        schema.TypeString,
// 			Optional:    true,
// 			Description: "Resource ID of the computer system resource. If not provided, then the first system resource is used from the computer system collection",
// 		},
// 		"boot_order": {
// 			Type:        schema.TypeList,
// 			Computed:    true,
// 			Description: "An array of BootOptionReference strings that represent the persistent boot order for this computer system",
// 			Elem: &schema.Schema{
// 				Type: schema.TypeString,
// 			},
// 		},
// 		"boot_source_override_enabled": {
// 			Type:        schema.TypeString,
// 			Computed:    true,
// 			Description: "The state of the boot source override feature",
// 		},
// 		"boot_source_override_mode": {
// 			Type:        schema.TypeString,
// 			Computed:    true,
// 			Description: "The BIOS boot mode to use when the system boots from the BootSourceOverrideTarget boot source",
// 		},
// 		"boot_source_override_target": {
// 			Type:        schema.TypeString,
// 			Computed:    true,
// 			Description: "The current boot source to use at the next boot instead of the normal boot device, if BootSourceOverrideEnabled is true",
// 		},
// 		"uefi_target_boot_source_override": {
// 			Type:        schema.TypeString,
// 			Computed:    true,
// 			Description: "The UEFI device path of the device from which to boot when BootSourceOverrideTarget is UefiTarget",
// 		},
// 	}
// }

// func dataSourceRedfishSystemBootRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
// 	service, err := NewConfig(m.(*schema.ResourceData), d)
// 	if err != nil {
// 		return diag.Errorf(err.Error())
// 	}

// 	return readRedfishSystemBoot(service, d)
// }

// func readRedfishSystemBoot(service *gofish.Service, d *schema.ResourceData) diag.Diagnostics {
// 	var diags diag.Diagnostics

// 	systems, err := service.Systems()
// 	if err != nil {
// 		return diag.Errorf("Error when retrieving systems: %s", err)
// 	}

// 	// get the boot resource
// 	var computerSystem *redfish.ComputerSystem
// 	var boot redfish.Boot
// 	if systemResourceId, ok := d.GetOk("resource_id"); ok {
// 		for key := range systems {
// 			if systems[key].ID == systemResourceId {
// 				computerSystem = systems[key]
// 				boot = systems[key].Boot
// 				break
// 			}
// 		}

// 		if computerSystem == nil {
// 			return diag.Errorf("Could not find a ComputerSystem resource with resource ID = %s", systemResourceId)
// 		}
// 	} else {
// 		// use the first system resource in the collection if resource
// 		// ID is not provided
// 		computerSystem = systems[0]
// 		boot = systems[0].Boot
// 	}

// 	if err := d.Set("boot_order", boot.BootOrder); err != nil {
// 		return diag.Errorf("error setting BootOrder: %s", err)
// 	}

// 	if err := d.Set("boot_source_override_enabled", boot.BootSourceOverrideEnabled); err != nil {
// 		return diag.Errorf("error setting BootSourceOverrideEnabled: %s", err)
// 	}

// 	if err := d.Set("boot_source_override_mode", boot.BootSourceOverrideMode); err != nil {
// 		return diag.Errorf("error setting BootSourceOverrideMode: %s", err)
// 	}

// 	if err := d.Set("boot_source_override_target", boot.BootSourceOverrideTarget); err != nil {
// 		return diag.Errorf("error setting BootSourceOverrideTarget: %s", err)
// 	}

// 	if err := d.Set("uefi_target_boot_source_override", boot.UefiTargetBootSourceOverride); err != nil {
// 		return diag.Errorf("error setting UefiTargetBootSourceOverride: %s", err)
// 	}

// 	// set computer system ODataID as the resource ID
// 	d.SetId(computerSystem.ODataID)
// 	return diags
// }
