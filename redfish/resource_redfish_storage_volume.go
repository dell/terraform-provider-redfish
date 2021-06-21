package redfish

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/dell/terraform-provider-redfish/common"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/stmcginnis/gofish"
	redfishcommon "github.com/stmcginnis/gofish/common"
	"github.com/stmcginnis/gofish/redfish"
)

const (
	defaultStorageVolumeResetTimeout  int = 120
	defaultStorageVolumeJobTimeout    int = 1200
	intervalStorageVolumeJobCheckTime int = 10
)

func resourceRedfishStorageVolume() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceRedfishStorageVolumeCreate,
		ReadContext:   resourceRedfishStorageVolumeRead,
		UpdateContext: resourceRedfishStorageVolumeUpdate,
		DeleteContext: resourceRedfishStorageVolumeDelete,
		Schema:        getResourceRedfishStorageVolumeSchema(),
	}
}

func getResourceRedfishStorageVolumeSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"redfish_server": {
			Type:        schema.TypeList,
			Required:    true,
			Description: "This list contains the different redfish endpoints to manage (different servers)",
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"user": {
						Type:        schema.TypeString,
						Optional:    true,
						Description: "This field is the user to login against the redfish API",
					},
					"password": {
						Type:        schema.TypeString,
						Optional:    true,
						Description: "This field is the password related to the user given",
					},
					"endpoint": {
						Type:        schema.TypeString,
						Required:    true,
						Description: "This field is the endpoint where the redfish API is placed",
					},
					"ssl_insecure": {
						Type:        schema.TypeBool,
						Optional:    true,
						Description: "This field indicates if the SSL/TLS certificate must be verified",
					},
				},
			},
		},
		"storage_controller_id": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "This value must be the storage controller ID the user want to manage. I.e: RAID.Integrated.1-1",
		},
		"volume_name": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "This value is the desired name for the volume to be given",
		},
		"volume_type": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "This value specifies the raid level the virtual disk is going to have. Possible values are: NonRedundant (RAID-0), Mirrored (RAID-1), StripedWithParity (RAID-5), SpannedMirrors (RAID-10) or SpannedStripesWithParity (RAID-50)",
			ValidateFunc: validation.StringInSlice([]string{
				// string(redfish.RawDeviceVolumeType),
				string(redfish.NonRedundantVolumeType),
				string(redfish.MirroredVolumeType),
				string(redfish.StripedWithParityVolumeType),
				string(redfish.SpannedMirrorsVolumeType),
				string(redfish.SpannedStripesWithParityVolumeType),
			}, false),
		},
		"drives": {
			Type:        schema.TypeList,
			Required:    true,
			Description: "This list contains the physical disks names to create the volume within a disk controller",
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
		"settings_apply_time": {
			Type:        schema.TypeString,
			Description: "Flag to make the operation either \"Immediate\" or \"OnReset\". By default value is \"Immediate\"",
			Optional:    true,
			ValidateFunc: validation.StringInSlice([]string{
				string(redfishcommon.ImmediateApplyTime),
				string(redfishcommon.OnResetApplyTime)}, false),
		},
		"reset_type": {
			Type:     schema.TypeString,
			Optional: true,
			Description: "Reset type allows to choose the type of restart to apply when settings_apply_time is set to \"OnReset\"" +
				"Possible values are: \"ForceRestart\", \"GracefulRestart\" or \"PowerCycle\". If not set, \"ForceRestart\" is the default.",
			ValidateFunc: validation.StringInSlice([]string{
				string(redfish.ForceRestartResetType),
				string(redfish.GracefulRestartResetType),
				string(redfish.PowerCycleResetType),
			}, false),
		},
		"reset_timeout": {
			Type:     schema.TypeInt,
			Optional: true,
			Description: "reset_timeout is the time in seconds that the provider waits for the server to be reset" +
				"(if settings_apply_time is set to \"OnReset\") before timing out. Default is 120s.",
		},
		"volume_job_timeout": {
			Type:     schema.TypeInt,
			Optional: true,
			Description: "volume_job_timeout is the time in seconds that the provider waits for the volume job to be completed before timing out." +
				"Default is 1200s",
		},
	}
}

func resourceRedfishStorageVolumeCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	service, err := NewConfig(m.(*schema.ResourceData), d)
	if err != nil {
		return diag.Errorf(err.Error())
	}
	return createRedfishStorageVolume(service, d)
}

func resourceRedfishStorageVolumeRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	service, err := NewConfig(m.(*schema.ResourceData), d)
	if err != nil {
		return diag.Errorf(err.Error())
	}
	return readRedfishStorageVolume(service, d)
}

func resourceRedfishStorageVolumeUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	service, err := NewConfig(m.(*schema.ResourceData), d)
	if err != nil {
		return diag.Errorf(err.Error())
	}
	if diags := updateRedfishStorageVolume(ctx, service, d, m); diags.HasError() {
		return diags
	}
	return resourceRedfishStorageVolumeRead(ctx, d, m)
}

func resourceRedfishStorageVolumeDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	service, err := NewConfig(m.(*schema.ResourceData), d)
	if err != nil {
		return diag.Errorf(err.Error())
	}
	return deleteRedfishStorageVolume(service, d)
}

func createRedfishStorageVolume(service *gofish.Service, d *schema.ResourceData) diag.Diagnostics {
	var diags diag.Diagnostics

	// Lock the mutex to avoid race conditions with other resources
	redfishMutexKV.Lock(getRedfishServerEndpoint(d))
	defer redfishMutexKV.Unlock(getRedfishServerEndpoint(d))

	// Get user config
	storageID := d.Get("storage_controller_id").(string)
	volumeType := d.Get("volume_type").(string)
	volumeName := d.Get("volume_name").(string)
	driveNamesRaw := d.Get("drives").([]interface{})
	applyTime, ok := d.GetOk("settings_apply_time")
	if !ok {
		// If settingsApplyTime has not set, by default use Immediate
		applyTime = "Immediate"
	}

	// Convert from []interface{} to []string for using
	driveNames := make([]string, len(driveNamesRaw))
	for i, raw := range driveNamesRaw {
		driveNames[i] = raw.(string)
	}

	volumeJobTimeout, ok := d.GetOk("volume_job_timeout")
	if !ok {
		volumeJobTimeout = defaultStorageVolumeJobTimeout
	}

	// Get storage
	systems, err := service.Systems()
	if err != nil {
		return diag.Errorf("Error when retreiving the Systems from the Redfish API")
	}

	storageControllers, err := systems[0].Storage()
	if err != nil {
		return diag.Errorf("Error when retreiving the Storage from %v from the Redfish API", systems[0].Name)
	}

	storage, err := getStorageController(storageControllers, storageID)
	if err != nil {
		return diag.Errorf("Error when getting the storage struct: %s", err)
	}

	// Check if settings_apply_time is doable on this controller
	operationApplyTimes, err := storage.GetOperationApplyTimeValues()
	if err != nil {
		return diag.Errorf("couldn't retrieve operationApplyTimes from %s controller", storage.Name)
	}
	if !checkOperationApplyTimes(applyTime.(string), operationApplyTimes) {
		return diag.Errorf("Storage controller %s does not support settings_apply_time: %s", storageID, applyTime)
	}

	//Get drives
	allStorageDrives, err := storage.Drives()
	if err != nil {
		return diag.Errorf("Error when getting the drives attached to controller - %s", err)
	}
	drives, err := getDrives(allStorageDrives, driveNames)
	if err != nil {
		return diag.Errorf("Error when getting the drives: %s", err)
	}

	// Create volume job
	jobID, err := createVolume(service, storage.ODataID, volumeType, volumeName, drives, applyTime.(string))
	if err != nil {
		return diag.Errorf("Error when creating the virtual disk on disk controller %s - %s", storageID, err)
	}

	// Immediate or OnReset scenarios
	switch applyTime.(string) {
	case string(redfishcommon.OnResetApplyTime): // OnReset case
		// Get reset_timeout and reset_type from schema
		resetType, ok := d.GetOk("reset_type")
		if !ok {
			resetType = string(redfish.ForceRestartResetType)
		}
		resetTimeout, ok := d.GetOk("reset_timeout")
		if !ok {
			resetTimeout = defaultStorageVolumeResetTimeout
		}

		// Reboot the server
		_, diags := PowerOperation(resetType.(string), resetTimeout.(int), intervalSimpleUpdateJobCheckTime, service)
		if diags.HasError() {
			// Handle this scenario - TBD
			return diag.Errorf("there was an issue when restarting the server")
		}

	}

	// Wait for the job to finish
	err = common.WaitForJobToFinish(service, jobID, intervalStorageVolumeJobCheckTime, volumeJobTimeout.(int))
	if err != nil {
		return diag.Errorf("Error, job %s wasn't able to complete: %s", jobID, err)
	}

	//Get storage volumes
	volumes, err := storage.Volumes()
	if err != nil {
		return diag.Errorf("there was an issue when retrieving volumes - %s", err)
	}
	volumeID, err := getVolumeID(volumes, volumeName)
	if err != nil {
		return diag.Errorf("Error. The volume ID with volume name %s on %s controller was not found", volumeName, storageID)
	}

	d.SetId(volumeID)
	return diags

}

func readRedfishStorageVolume(service *gofish.Service, d *schema.ResourceData) diag.Diagnostics {
	var diags diag.Diagnostics

	//Check if the volume exists
	_, err := redfish.GetVolume(service.Client, d.Id())
	if err != nil {
		e, ok := err.(*redfishcommon.Error)
		if !ok {
			return diag.Errorf("There was an error with the API: %s", err)
		}
		if e.HTTPReturnedStatusCode == http.StatusNotFound {
			log.Printf("Volume %s doesn't exist", d.Id())
			d.SetId("")
			return diags
		}
		return diag.Errorf("Status code %d - %s", e.HTTPReturnedStatusCode, e.Error())
	}

	/*
		- If it has jobID, if finished, get the volumeID
		Also never EVER trigger an update regarding disk properties for safety reasons
	*/

	return diags
}

func updateRedfishStorageVolume(ctx context.Context, service *gofish.Service, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	/*
		Since we are dealing with storage, better not to try to update anything
	*/
	return diags
}

func deleteRedfishStorageVolume(service *gofish.Service, d *schema.ResourceData) diag.Diagnostics {
	var diags diag.Diagnostics

	// Lock the mutex to avoid race conditions with other resources
	redfishMutexKV.Lock(getRedfishServerEndpoint(d))
	defer redfishMutexKV.Unlock(getRedfishServerEndpoint(d))

	// Get vars from schema
	applyTime, ok := d.GetOk("settings_apply_time")
	if !ok {
		// If settingsApplyTime has not set, by default use Immediate
		applyTime = "Immediate"
	}
	volumeJobTimeout, ok := d.GetOk("volume_job_timeout")
	if !ok {
		volumeJobTimeout = defaultStorageVolumeJobTimeout
	}

	jobID, err := deleteVolume(service, d.Id())
	if err != nil {
		return diag.Errorf("Error. There was an error when deleting volume %s - %s", d.Id(), err)
	}

	switch applyTime.(string) {
	case string(redfishcommon.OnResetApplyTime): // OnReset case
		// Get reset_timeout and reset_type from schema
		resetType, ok := d.GetOk("reset_type")
		if !ok {
			resetType = string(redfish.ForceRestartResetType)
		}
		resetTimeout, ok := d.GetOk("reset_timeout")
		if !ok {
			resetTimeout = defaultStorageVolumeResetTimeout
		}

		// Reboot the server
		_, diags := PowerOperation(resetType.(string), resetTimeout.(int), intervalSimpleUpdateJobCheckTime, service)
		if diags.HasError() {
			// Handle this scenario - TBD
			return diag.Errorf("there was an issue when restarting the server")
		}
	}

	//WAIT FOR VOLUME TO DELETE
	err = common.WaitForJobToFinish(service, jobID, intervalStorageVolumeJobCheckTime, volumeJobTimeout.(int))
	if err != nil {
		return diag.Errorf("Error, timeout reached when waiting for job %s to finish. %s", jobID, err)
	}

	return diags
}

func getStorageController(storageControllers []*redfish.Storage, diskControllerID string) (*redfish.Storage, error) {
	for _, storage := range storageControllers {
		if storage.Entity.ID == diskControllerID {
			return storage, nil
		}
	}
	return nil, fmt.Errorf("error. Didn't find the storage controller %v", diskControllerID)
}

func deleteVolume(service *gofish.Service, volumeURI string) (jobID string, err error) {
	//TODO - Check if we can delete immediately or if we need to schedule a job
	res, err := service.Client.Delete(volumeURI)
	if err != nil {
		return "", fmt.Errorf("error while deleting the volume %s", volumeURI)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusAccepted {
		return "", fmt.Errorf("the operation was not successful. Return code was different from 202 ACCEPTED")
	}
	jobID = res.Header.Get("Location")
	if len(jobID) == 0 {
		return "", fmt.Errorf("there was some error when retreiving the jobID")
	}
	return jobID, nil
}

func getDrives(drives []*redfish.Drive, driveNames []string) ([]*redfish.Drive, error) {
	drivesToReturn := []*redfish.Drive{}
	for _, v := range drives {
		for _, w := range driveNames {
			if v.Name == w {
				drivesToReturn = append(drivesToReturn, v)
			}
		}
	}
	if len(driveNames) != len(drivesToReturn) {
		return nil, fmt.Errorf("any of the drives you inserted doesn't exist")
	}
	return drivesToReturn, nil
}

/*
createVolume creates a virtualdisk on a disk controller by using the redfish API
Parameters:
	c -> client API
	service -> Service struct from gofish
	storageLink -> ODataID of the storage object (i.e. /redfish/v1/.../RAID.Integrated.1-1)
	volumeType -> raid mode to apply to that set of disks
		Modes:
			- RAID-0 -> "NonRedundant"
			- RAID-1 -> "Mirrored"
			- RAID-5 -> "StripedWithParity"
			- RAID-10 -> "SpannedMirrors"
			- RAID-50 -> "SpannedStripesWithParity"
	volumeName -> Name for the volume
	driveNames -> Drives to use for the raid configuration
*/
func createVolume(service *gofish.Service,
	storageLink string,
	volumeType string,
	volumeName string,
	drives []*redfish.Drive,
	applyTime string) (jobID string, err error) {

	newVolume := make(map[string]interface{})
	newVolume["VolumeType"] = volumeType
	newVolume["Name"] = volumeName
	newVolume["@Redfish.OperationApplyTime"] = applyTime
	var listDrives []map[string]string
	for _, drive := range drives {
		storageDrive := make(map[string]string)
		storageDrive["@odata.id"] = drive.Entity.ODataID
		listDrives = append(listDrives, storageDrive)
	}
	newVolume["Drives"] = listDrives
	volumesURL := fmt.Sprintf("%v/Volumes", storageLink)
	res, err := service.Client.Post(volumesURL, newVolume)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusAccepted {
		return "", fmt.Errorf("the query was unsucessfull")
	}
	jobID = res.Header.Get("Location")
	if len(jobID) == 0 {
		return "", fmt.Errorf("there was some error when retreiving the jobID")
	}
	return jobID, nil
}

func getVolumeID(volumes []*redfish.Volume, volumeName string) (volumeLink string, err error) {
	for _, v := range volumes {
		if v.Name == volumeName {
			volumeLink = v.ODataID
			return volumeLink, nil
		}
	}
	return "", fmt.Errorf("couldn't find a volume with the provided name")
}

func checkOperationApplyTimes(optionToCheck string, storageOperationApplyTimes []redfishcommon.OperationApplyTime) (result bool) {
	for _, v := range storageOperationApplyTimes {
		if optionToCheck == string(v) {
			return true
		}
	}
	return false
}
