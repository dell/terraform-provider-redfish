package redfish

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/dell/terraform-provider-redfish/common"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/stmcginnis/gofish"
	redfishcommon "github.com/stmcginnis/gofish/common"
	"github.com/stmcginnis/gofish/redfish"
)

const (
	defaultSimpleUpdateResetTimeout  int = 120
	defaultSimpleUpdateJobTimeout    int = 1200
	intervalSimpleUpdateJobCheckTime int = 10
)

func resourceRedfishSimpleUpdate() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceRedfishSimpleUpdateCreate,
		ReadContext:   resourceRedfishSimpleUpdateRead,
		UpdateContext: resourceRedfishSimpleUpdateUpdate,
		DeleteContext: resourceRedfishSimpleUpdateDelete,
		Schema:        getResourceRedfishSimpleUpdateSchema(),
	}
}

func getResourceRedfishSimpleUpdateSchema() map[string]*schema.Schema {
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
		"transfer_protocol": {
			Type:     schema.TypeString,
			Required: true,
			Description: "The network protocol that the Update Service uses to retrieve the software image file located at the URI provided " +
				"in ImageURI, if the URI does not contain a scheme." +
				"Accepted values: CIFS, FTP, SFTP, HTTP, HTTPS, NSF, SCP, TFTP, OEM, NFS",
		},
		/* For the time being, target_firmware_image will be the local path for our firmware packages.
		   It is intended to work along HTTP transfer protocol
		   In the future it could be used for targetting FTP/CIFS/NFS images
		   TBD - Think about a custom diff function that grabs only the file name and not the path, to avoid unneeded update triggers
		*/
		"target_firmware_image": {
			Type:     schema.TypeString,
			Required: true,
			Description: "Target firmware image used for firmware update on the redfish instance. " +
				"Make sure you place your firmware packages in the same folder as the module and set it as follows: \"${path.module}/BIOS_FXC54_WN64_1.15.0.EXE\"",
			// DiffSuppressFunc will allow moving fw packages through the filesystem without triggering an update if so.
			// At the moment it uses filename to see if they're the same. We need to strengthen that by somehow using hashing
			DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
				if filepath.Base(old) == filepath.Base(new) {
					return true
				}
				return false
			},
		},
		"reset_type": {
			Type:     schema.TypeString,
			Required: true,
			Description: "Reset type allows to choose the type of restart to apply when firmware upgrade is scheduled." +
				"Possible values are: \"ForceRestart\", \"GracefulRestart\" or \"PowerCycle\"",
			ValidateFunc: validation.StringInSlice([]string{
				string(redfish.ForceRestartResetType),
				string(redfish.GracefulRestartResetType),
				string(redfish.PowerCycleResetType),
			}, false),
		},
		"reset_timeout": {
			Type:        schema.TypeInt,
			Optional:    true,
			Description: "reset_timeout is the time in seconds that the provider waits for the server to be reset before timing out.",
		},
		"simple_update_job_timeout": {
			Type:        schema.TypeInt,
			Optional:    true,
			Description: "simple_update_job_timeout is the time in seconds that the provider waits for the simple update job to be completed before timing out.",
		},
		"software_id": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "Software ID from the firmware package uploaded",
		},
		"version": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "Software version from the firmware package uploaded",
		},
	}
}

func resourceRedfishSimpleUpdateCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	service, err := NewConfig(m.(*schema.ResourceData), d)
	if err != nil {
		return diag.Errorf(err.Error())
	}
	return updateRedfishSimpleUpdate(ctx, service, d, m)
}

func resourceRedfishSimpleUpdateRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	service, err := NewConfig(m.(*schema.ResourceData), d)
	if err != nil {
		return diag.Errorf(err.Error())
	}
	return readRedfishSimpleUpdate(service, d)
}

func resourceRedfishSimpleUpdateUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	service, err := NewConfig(m.(*schema.ResourceData), d)
	if err != nil {
		return diag.Errorf(err.Error())
	}
	if diags := updateRedfishSimpleUpdate(ctx, service, d, m); diags.HasError() {
		return diags
	}
	return resourceRedfishSimpleUpdateRead(ctx, d, m)
}

func resourceRedfishSimpleUpdateDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	service, err := NewConfig(m.(*schema.ResourceData), d)
	if err != nil {
		return diag.Errorf(err.Error())
	}
	return deleteRedfishSimpleUpdate(service, d)
}

func readRedfishSimpleUpdate(service *gofish.Service, d *schema.ResourceData) diag.Diagnostics {
	var diags diag.Diagnostics

	// Try to get software inventory
	_, err := redfish.GetSoftwareInventory(service.Client, d.Id())
	if err != nil {
		_, ok := err.(*redfishcommon.Error)
		if !ok {
			return diag.Errorf("there was an issue with the API")
		}
		// the firmware package previously applied has changed, trigger update
		d.SetId("")
	}

	return diags
}

func updateRedfishSimpleUpdate(ctx context.Context, service *gofish.Service, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	// Lock the mutex to avoid race conditions with other resources
	redfishMutexKV.Lock(getRedfishServerEndpoint(d))
	defer redfishMutexKV.Unlock(getRedfishServerEndpoint(d))

	transferProtocol := d.Get("transfer_protocol").(string)
	targetFirmwareImage := d.Get("target_firmware_image").(string)
	resetType := d.Get("reset_type").(string)
	resetTimeout, ok := d.GetOk("reset_timeout")
	if !ok {
		resetTimeout = defaultSimpleUpdateResetTimeout
	}
	simpleUpdateJobTimeout, ok := d.GetOk("simple_update_job_timeout")
	if !ok {
		simpleUpdateJobTimeout = defaultSimpleUpdateJobTimeout
	}
	log.Printf("[DEBUG] resetTimeout is set to %d and simpleUpdateJobTimeout to %d", resetTimeout.(int), simpleUpdateJobTimeout.(int))

	// Check if chosen reset type is supported before doing anything else
	systems, err := service.Systems()
	if err != nil {
		return diag.Errorf("Couldn't retrieve allowed reset types from systems - %s", err)
	}
	if ok := checkResetType(resetType, systems[0].SupportedResetTypes); !ok {
		return diag.Errorf("reset type %s is not available in this redfish implementation", resetType)
	}

	// Get update service from root
	updateService, err := service.UpdateService()
	if err != nil {
		return diag.Errorf("error while retrieving UpdateService - %s", err)
	}

	//Check if the transfer protocol is available in the redfish instance
	err = checkTransferProtocol(transferProtocol, updateService)
	if err != nil {
		var availableTransferProtocols string
		for _, v := range updateService.TransferProtocol {
			availableTransferProtocols += fmt.Sprintf("%s ", v)
		}
		return diag.Errorf("%s. Supported transfer protocols in this implementation: %s", err, availableTransferProtocols) // !!!! append list of supported transfer protocols
	}

	switch transferProtocol {
	case "HTTP":
		// Get ETag from FW inventory
		response, err := service.Client.Get(updateService.FirmwareInventory)
		if err != nil {
			diag.Errorf("error while retrieving Etag from FirmwareInventory")
		}
		response.Body.Close()
		etag := response.Header.Get("ETag")

		// Set custom headers
		customHeaders := map[string]string{
			"if-match": etag,
		}

		// Open file to upload
		file, err := openFile(targetFirmwareImage)
		if err != nil {
			return diag.Errorf("couldn't open FW file to upload - %s", err)
		}
		defer file.Close()

		// Set payload
		payload := map[string]io.Reader{
			"file": file,
		}

		// Upload FW Package to FW inventory
		response, err = service.Client.PostMultipartWithHeaders(updateService.HTTPPushURI, payload, customHeaders)
		if err != nil {
			return diag.Errorf("there was an issue when uploading FW package to redfish - %s", err)
		}
		response.Body.Close()
		packageLocation := response.Header.Get("Location")

		// Get package information ( SoftwareID - Version )
		packageInformation, err := redfish.GetSoftwareInventory(service.Client, packageLocation)
		if err != nil {
			return diag.Errorf("there was an issue when retrieving uploaded package information - %s", err)
		}

		// Set payload for POST call that'll trigger the update job scheduling
		triggerUpdatePayload := struct {
			ImageURI string
		}{
			ImageURI: packageLocation,
		}
		// Do the POST call agains Simple.Update service
		response, err = service.Client.Post(updateService.UpdateServiceTarget, triggerUpdatePayload)
		if err != nil {
			// Delete uploaded package - TBD
			return diag.Errorf("there was an issue when scheduling the update job - %s", err)
		}
		response.Body.Close()
		// Get jobid
		jobID := response.Header.Get("Location")

		// Reboot the server
		_, diags := PowerOperation(resetType, resetTimeout.(int), intervalSimpleUpdateJobCheckTime, service)
		if diags.HasError() {
			// Delete uploaded package - TBD
			return diag.Errorf("there was an issue when restarting the server")
		}

		// Check JID
		err = common.WaitForJobToFinish(service, jobID, intervalSimpleUpdateJobCheckTime, simpleUpdateJobTimeout.(int))
		if err != nil {
			// Delete uploaded package - TBD
			return diag.Errorf("there was an issue when waiting for the job to complete - %s", err)
		}

		// Get updated FW inventory
		fwInventory, err := updateService.FirmwareInventories()
		if err != nil {
			// TBD - HOW TO HANDLE WHEN FAILS BUT FIRMWARE WAS INSTALLED?
			return diag.Errorf("error when getting firmware inventory - %s", err)
		}

		// Get fw ID
		fwPackage, err := getFWfromInventory(fwInventory, packageInformation.SoftwareID, packageInformation.Version)
		if err != nil {
			// TBD - HOW TO HANDLE WHEN FAILS BUT FIRMWARE WAS INSTALLED?
			return diag.Errorf("error when retrieving fw package from fw inventory - %s", err)
		}
		d.Set("software_id", fwPackage.SoftwareID)
		d.Set("version", fwPackage.Version)
		d.SetId(fwPackage.ODataID)

	default:
		return diag.Errorf("Transfer protocol not available in this implementation")

	}

	return diags
}

func deleteRedfishSimpleUpdate(service *gofish.Service, d *schema.ResourceData) diag.Diagnostics {
	var diags diag.Diagnostics

	d.SetId("")

	return diags
}

// checkResetType check if the resetType passed is within the allowableValues slice
func checkResetType(resetType string, allowableValues []redfish.ResetType) bool {
	for _, v := range allowableValues {
		if resetType == string(v) {
			return true
		}
	}
	return false
}

// openFile is a simple function that opens a file
func openFile(filePath string) (*os.File, error) {
	if f, err := os.Open(filePath); err != nil {
		return nil, fmt.Errorf("error when opening %s file - %s", filePath, err)
	} else {
		return f, nil
	}
}

// checkTransferProtocol checks if the chosen transfer protocol is available in the redfish instance
func checkTransferProtocol(transferProtocol string, updateService *redfish.UpdateService) error {
	for _, v := range updateService.TransferProtocol {
		if transferProtocol == v {
			return nil
		}
	}
	return fmt.Errorf("this transfer protocol is not available in this redfish instance")
}

// getFWfromInventory get the right SoftwareInventory struct if exists
func getFWfromInventory(softwareInventories []*redfish.SoftwareInventory, softwareID, version string) (*redfish.SoftwareInventory, error) {
	for _, v := range softwareInventories {
		if v.SoftwareID == softwareID && v.Version == version {
			return v, nil
		}
	}
	return nil, fmt.Errorf("couldn't find FW on Firmware inventory")
}
