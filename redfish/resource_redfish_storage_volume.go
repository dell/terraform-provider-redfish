package redfish

import (
	"context"
	"fmt"
	"github.com/dell/terraform-provider-redfish/common"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stmcginnis/gofish"
	"github.com/stmcginnis/gofish/redfish"
	_ "log"
	"net/http"
)

func resourceRedfishStorageVolume() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceStorageVolumeCreate,
		ReadContext:   resourceStorageVolumeRead,
		UpdateContext: resourceStorageVolumeUpdate,
		DeleteContext: resourceStorageVolumeDelete,
		Schema:        getResourceStorageVolumeSchema(),
	}
}

func getResourceStorageVolumeSchema() map[string]*schema.Schema {
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
		"storage_controller_id": &schema.Schema{
			Type:        schema.TypeString,
			Required:    true,
			Description: "This value must be the storage controller ID the user want to manage. I.e: RAID.Integrated.1-1",
		},
		"volume_name": &schema.Schema{
			Type:        schema.TypeString,
			Required:    true,
			Description: "This value is the desired name for the volume to be given",
		},
		"volume_type": &schema.Schema{
			Type:        schema.TypeString,
			Required:    true,
			Description: "This value specifies the raid level the virtual disk is going to have. Possible values are: NonRedundant (RAID-0), Mirrored (RAID-1), StripedWithParity (RAID-5), SpannedMirrors (RAID-10) or SpannedStripesWithParity (RAID-50)",
		},
		"volume_disks": &schema.Schema{
			Type:        schema.TypeList,
			Required:    true,
			Description: "This list contains the physical disks names to create the volume within a disk controller",
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
		"settings_apply_time": &schema.Schema{
			Type:        schema.TypeString,
			Description: "Flag to make the operation either \"Immediate\" or \"OnReset\". By default value is \"Immediate\"",
			Optional:    true,
		},
		"job_id": &schema.Schema{
			Type:        schema.TypeString,
			Description: "This parameter will return the jobID from the job that will carry out the operation if \"settings_apply_time\" is different from \"Immediate\".",
			Computed:    true,
		},
	}
}

func resourceStorageVolumeCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	service, err := NewConfig(m.(*schema.ResourceData), d)
	if err != nil {
		return diag.Errorf(err.Error())
	}
	return createRedfishStorageVolume(service, d)
}

func resourceStorageVolumeRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	service, err := NewConfig(m.(*schema.ResourceData), d)
	if err != nil {
		return diag.Errorf(err.Error())
	}
	return readRedfishStorageVolume(service, d)
}

func resourceStorageVolumeUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	service, err := NewConfig(m.(*schema.ResourceData), d)
	if err != nil {
		return diag.Errorf(err.Error())
	}
	if diags := updateRedfishStorageVolume(ctx, service, d, m); diags.HasError() {
		return diags
	}
	return resourceStorageVolumeRead(ctx, d, m)
}

func resourceStorageVolumeDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	service, err := NewConfig(m.(*schema.ResourceData), d)
	if err != nil {
		return diag.Errorf(err.Error())
	}
	return deleteRedfishStorageVolume(service, d)
}

func createRedfishStorageVolume(service *gofish.Service, d *schema.ResourceData) diag.Diagnostics {
	var diags diag.Diagnostics
	//Get user config
	storageID := d.Get("storage_controller_id").(string)
	volumeType := d.Get("volume_type").(string)
	volumeName := d.Get("volume_name").(string)
	driveNamesRaw := d.Get("volume_disks").([]interface{})
	applyTime, ok := d.GetOk("settings_apply_time")
	if !ok {
		//If settingsApplyTime has not set, by default use Immediate
		applyTime = "Immediate"
	}
	//Convert from []interface{} to []string for using
	driveNames := make([]string, len(driveNamesRaw))
	for i, raw := range driveNamesRaw {
		driveNames[i] = raw.(string)
	}

	//Get storage
	storage, err := getStorageController(service, storageID)
	if err != nil {
		return diag.Errorf("Error when getting the storage struct: %s", err)
	}
	//Get drives
	drives, err := getDrives(storage, driveNames)
	if err != nil {
		return diag.Errorf("Error when getting the drives: %s", err)
	}
	jobID, err := createVolume(service, storage.ODataID, volumeType, volumeName, drives, applyTime.(string))
	if err != nil {
		return diag.Errorf("Error when creating the virtual disk on disk controller %s - %s", storageID, err)
	}

	//Need to figure out how to proceed with settingsApplyTime (Immediate or OnReset)
	switch applyTime.(string) {
	case "Immediate":
		err = common.WaitForJobToFinish(service, jobID, common.TimeBetweenAttempts, common.Timeout)
		if err != nil {
			return diag.Errorf("Error, job %s wasn't able to complete", jobID)
		}

		volumeID, err := getVolumeID(storage, volumeName)
		if err != nil {
			return diag.Errorf("Error. The volume ID with volume name %s on %s controller was not found", volumeName, storageID)
		}
		d.Set("job_id", "")
		d.SetId(volumeID)
		return diags
	case "OnReset":
		//TODO - Implement for not Immediate scenarios
	}
	return diag.Errorf("Error. The \"settingsApplyTime\" you chose doesn't exist")
}

func readRedfishStorageVolume(service *gofish.Service, d *schema.ResourceData) diag.Diagnostics {
	var diags diag.Diagnostics
	/*
		Here we gotta check:
			- If the volume exists
			- If it has jobID, if finished, get the volumeID

		Also never EVER trigger an update regarding disk properties for safety reasons
	*/
	return diags
}

func updateRedfishStorageVolume(ctx context.Context, service *gofish.Service, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	/*
		Since we are dealing with storage, betten not to try to update anything
	*/
	return diags
}

func deleteRedfishStorageVolume(service *gofish.Service, d *schema.ResourceData) diag.Diagnostics {
	var diags diag.Diagnostics
	switch l := len(d.Get("job_id").(string)); l {
	case 0: //This case means job_id has not been set, meaning a volume is in place.
		jobID, err := deleteVolume(service, d.Id())
		if err != nil {
			return diag.Errorf("Error. There was an error when deleting volume %s - %s", d.Id(), err)
		}
		//WAIT FOR VOLUME TO DELETE
		err = common.WaitForJobToFinish(service, jobID, common.TimeBetweenAttempts, common.Timeout)
		if err != nil {
			return diag.Errorf("Error, timeout reached when waiting for job %s to finish. %s", jobID, err)
		}
	default: //This case means job_id has been set.
		//Looks like now it's possible to use HTTP DELETE against the taskID on iDRAC 4.40.00.00
	}
	return diags
}

func getStorageController(service *gofish.Service, diskControllerID string) (*redfish.Storage, error) {
	systems, err := service.Systems()
	if err != nil {
		return nil, fmt.Errorf("Error when retreiving the Systems from the Redfish API")
	}
	sg, err := systems[0].Storage()
	if err != nil {
		return nil, fmt.Errorf("Error when retreiving the Storage from %v from the Redfish API", systems[0].Name)
	}
	for _, storage := range sg {
		if storage.Entity.ID == diskControllerID {
			return storage, nil
		}
	}
	return nil, fmt.Errorf("Error. Didn't find the storage controller %v", diskControllerID)
}

func deleteVolume(service *gofish.Service, volumeURI string) (jobID string, err error) {
	//TODO - Check if we can delete immediately or if we need to schedule a job
	res, err := service.Client.Delete(volumeURI)
	if err != nil {
		return "", fmt.Errorf("Error while deleting the volume %s", volumeURI)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusAccepted {
		return "", fmt.Errorf("The operation was not successful. Return code was different from 202 ACCEPTED")
	}
	jobID = res.Header.Get("Location")
	if len(jobID) == 0 {
		return "", fmt.Errorf("There was some error when retreiving the jobID")
	}
	return jobID, nil
}

func getDrives(storage *redfish.Storage, driveNames []string) ([]*redfish.Drive, error) {
	drivesToReturn := []*redfish.Drive{}
	drives, err := storage.Drives()
	if err != nil {
		return nil, err
	}
	for _, v := range drives {
		for _, w := range driveNames {
			if v.Name == w {
				drivesToReturn = append(drivesToReturn, v)
			}
		}
	}
	if len(driveNames) != len(drivesToReturn) {
		return nil, fmt.Errorf("Any of the drives you inserted doesn't exist")
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
		return "", fmt.Errorf("The query was unsucessfull")
	}
	jobID = res.Header.Get("Location")
	if len(jobID) == 0 {
		return "", fmt.Errorf("There was some error when retreiving the jobID")
	}
	return jobID, nil
}

func getVolumeID(storage *redfish.Storage, volumeName string) (volumeLink string, err error) {
	if err != nil {
		return "", err
	}
	//Get storage volumes
	volumes, err := storage.Volumes()
	if err != nil {
		return "", err
	}
	for _, v := range volumes {
		if v.Name == volumeName {
			volumeLink = v.ODataID
			return volumeLink, nil
		}
	}
	return "", fmt.Errorf("Couldn't find a volume with the provided name")
}
