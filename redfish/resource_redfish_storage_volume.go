package redfish

import (
	"context"
	"fmt"
	"github.com/dell/terraform-provider-redfish/common"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stmcginnis/gofish"
	redfishcommon "github.com/stmcginnis/gofish/common"
	"github.com/stmcginnis/gofish/redfish"
	"net/http"
)

const (
	// This constants are used to avoid hardcoding the terraform input variables
	storageControllerID string = "storage_controller_id"
	volumeName          string = "volume_name"
	volumeType          string = "volume_type"
	volumeDisks         string = "volume_disks"
	settingsApplyTime   string = "settings_apply_time"
	biosConfigJobURI    string = "bios_config_job_uri"
)

func resourceRedfishStorageVolume() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceStorageVolumeCreate,
		ReadContext:   resourceStorageVolumeRead,
		UpdateContext: resourceStorageVolumeUpdate,
		DeleteContext: resourceStorageVolumeDelete,
		Schema: map[string]*schema.Schema{
			storageControllerID: &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				Description: "This value must be the storage controller ID the user want to manage. I.e: RAID.Integrated.1-1",
			},
			volumeName: &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				Description: "This value is the desired name for the volume to be given",
			},
			volumeType: &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				Description: "This value specifies the raid level the virtual disk is going to have. Possible values are: NonRedundant (RAID-0), Mirrored (RAID-1), StripedWithParity (RAID-5), SpannedMirrors (RAID-10) or SpannedStripesWithParity (RAID-50)",
			},
			volumeDisks: &schema.Schema{
				Type:        schema.TypeList,
				Required:    true,
				Description: "This list contains the physical disks names to create the volume within a disk controller",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			settingsApplyTime: &schema.Schema{
				Type:        schema.TypeString,
				Description: "Flag to make the operation either \"Immediate\" or \"OnReset\". By default value is \"Immediate\"",
				Optional:    true,
			},
			biosConfigJobURI: &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			/*TODO
			Implement validate function with redfish.GetOperationApplyTimeValues()*/
		},
	}
}

func resourceStorageVolumeCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := m.(*gofish.APIClient)
	service := conn.Service
	//Get user config
	storageID := d.Get(storageControllerID).(string)
	volumeType := d.Get(volumeType).(string)
	volumeName := d.Get(volumeName).(string)
	driveNamesRaw := d.Get(volumeDisks).([]interface{})
	applyTime, ok := d.GetOk(settingsApplyTime)
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
		return diag.Errorf("Issue when getting the storage struct: %s", err)
	}
	//Get drives
	drives, err := getDrives(storage, driveNames)
	if err != nil {
		return diag.Errorf("Issue when getting the drives: %s", err)
	}
	//Need to figure out how to proceed with settingsApplyTime (Immediate or OnReset)
	//func createVolume(service *gofish.Service, storageLink string, volumeType string, volumeName string, drives []*redfish.Drive, applyTime string) (jobID string, err error) {
	//Get redfish.drives

	jobID, err := createVolume(conn, storage.ODataID, volumeType, volumeName, drives, applyTime.(string))
	if err != nil {
		return diag.Errorf("Error when creating the virtual disk on disk controller %s - %s", storageID, err)
	}
	if applyTime.(string) == "Immediate" {
		err = common.WaitForJobToFinish(conn, jobID, common.TimeBetweenAttempts, common.Timeout)
		if err != nil {
			return diag.Errorf("Error. Job %s wasn't able to complete", jobID)
		}
		// Get new volumeID
		//getVolumeID(storage *redfish.Storage, volumeName string) (volumeLink string, err error)
		storage, err := getStorageController(service, storageID)
		if err != nil {
			return diag.Errorf("Issue when getting the storage struct: %s", err)
		}
		volumeID, err := getVolumeID(storage, volumeName)
		if err != nil {
			return diag.Errorf("Error. The volume ID with volume name %s on %s controller was not found", volumeName, storageID)
		}
		d.Set(biosConfigJobURI, "")
		d.SetId(volumeID)
	} else {
		//TODO - Implement for not Immediate scenarios
		d.Set(biosConfigJobURI, jobID)
		d.SetId(jobID)
	}

	//resourceStorageVolumeRead(ctx, d, m)
	return diags
}

func resourceStorageVolumeRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	return diags
}

func resourceStorageVolumeUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	return resourceStorageVolumeRead(ctx, d, m)
}

func resourceStorageVolumeDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics
	conn := m.(*gofish.APIClient)
	service := conn.Service
	//Get user config
	//If applyTime has been set to Immediate, the volumeID of the resource will be the ODataID of the volume just created.
	//If applyTime is OnReset, the volumeID will be the JobID
	volumeID := d.Id()
	applyTime, ok := d.GetOk(settingsApplyTime)
	if !ok {
		//If settingsApplyTime has not set, by default use Immediate
		applyTime = "Immediate"
	}
	//DELETE VOLUME
	if applyTime.(string) == "Immediate" {
		jobID, err := deleteVolume(conn, volumeID)
		if err != nil {
			return diag.Errorf("Error. There was an error when deleting volume %s", volumeID)
		}
		//WAIT FOR VOLUME TO DELETE
		err = common.WaitForJobToFinish(conn, jobID, common.TimeBetweenAttempts, common.Timeout)
		if err != nil {
			panic(err)
		}
	} else {
		//Check if the job has been completed or not. If not, kill the job. If so, kill the volume
		task, err := redfish.GetTask(conn, volumeID)
		if err != nil {
			return diag.Errorf("Issue when retrieving the tasks: %s", err)
		}
		if task.TaskState == redfish.CompletedTaskState {
			//Get the actual volumeID for destroying it
			storageID := d.Get(storageControllerID).(string)
			volumeName := d.Get(volumeName).(string)
			//getStorageController
			storage, err := getStorageController(service, storageID)
			if err != nil {
				return diag.Errorf("Issue when getting the storage struct: %s", err)
			}
			//getVolumeID(storage *redfish.Storage, volumeName string) (volumeLink string, err error)
			actualVolumeID, err := getVolumeID(storage, volumeName)
			if err != nil {
				return diag.Errorf("Issue when getting the actual volumeID: %s", err)
			}
			//MAYBE WE NEED TO SET A JOB INSTEAD OF DELETING IT RIGHTAWAY
			_, err = deleteVolume(conn, actualVolumeID)
			d.SetId("")
		} else {
			//Get rid of the Job that will create the volume
			//IMPORTART LIMITATION. TO DELETE A TASK IN DELL EMC REDFISH IMPLEMENTATION, NEEDS TO BE DONE THROUGH ITS MANAGER/redfish/v1/Managers/iDRAC.Embedded.1/Jobs/%s
			err := common.DeleteDellJob(conn, task.ID)
			if err != nil {
				return diag.Errorf("Issue when deleting the task: %s", err)
			}
			d.SetId("")
		}
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

func deleteVolume(c redfishcommon.Client, volumeURI string) (jobID string, err error) {
	//TODO - Check if we can delete immediately or if we need to schedule a job
	res, err := c.Delete(volumeURI)
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
	/*storage, err := getStorageController(service, diskControllerName)
	if err != nil {
		return nil, err
	}*/
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
func createVolume(client redfishcommon.Client,
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
	res, err := client.Post(volumesURL, newVolume)
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
	//storage, err := getStorageController(service, diskControllerName)
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
