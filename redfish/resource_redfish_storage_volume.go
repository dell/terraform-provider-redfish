package redfish

import (
	"github.com/dell/terraform-provider-redfish/common"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	// "github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"context"
	"fmt"
	"github.com/stmcginnis/gofish"
	"github.com/stmcginnis/gofish/redfish"
	//	"strings"
)

const (
	// This constants are used to avoid hardcoding the terraform input variables
	storageControllerStr string = "storage_controller"
	volumeNameStr        string = "volume_name"
	raidLevelStr         string = "raid_level"
	volumeDisks          string = "volume_disks"
	settingsApplyTimeStr string = "settings_apply_time"
	biosConfigJobURIStr  string = "bios_config_job_uri"
)

func resourceRedfishStorageVolume() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceStorageVolumeCreate,
		ReadContext:   resourceStorageVolumeRead,
		UpdateContext: resourceStorageVolumeUpdate,
		DeleteContext: resourceStorageVolumeDelete,
		Schema: map[string]*schema.Schema{
			storageControllerStr: &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				Description: "This value must be the disk controller the user want to manage. I.e: RAID.Integrated.1-1",
			},
			volumeNameStr: &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				Description: "This value is the desired name for the volume to be given",
			},
			raidLevelStr: &schema.Schema{ //Call it volumeType
				Type:        schema.TypeString,
				Required:    true,
				Description: "This value specifies the raid level the virtual disk is going to have. Possible values are: NonRedundant (RAID-0), Mirrored (RAID-1), StripedWithParity (RAID-5), SpannedMirrors (RAID-10) or SpannedStripesWithParity (RAID-50)",
			},
			volumeDisks: &schema.Schema{
				Type:        schema.TypeList,
				Required:    true,
				Description: "This list contains the disks to create the volume within a disk controller",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			settingsApplyTimeStr: &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
			biosConfigJobURIStr: &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			/*ValidateFunc: xxxx,
				biosConfigJobURIStr: {
				Type:        schema.TypeString,
				Description: "Volume configuration job URI",
				Computed:    true,
			},*/
		},
	}
}

func resourceStorageVolumeCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := m.(*gofish.APIClient)
	//Get user config
	storageController := d.Get(storageControllerStr).(string)
	raidLevel := d.Get(raidLevelStr).(string)
	volumeName := d.Get(volumeNameStr).(string)
	applyTime := d.Get(settingsApplyTimeStr).(string)
	driveNamesRaw := d.Get(volumeDisks).([]interface{})

	//Convert from []interface{} to []string for using
	driveNames := make([]string, len(driveNamesRaw))
	for i, raw := range driveNamesRaw {
		driveNames[i] = raw.(string)
	}

	//Need to figure out how to proceed with settingsApplyTime (Immediate or OnReset)
	jobID, err := createVolume(conn, storageController, raidLevel, volumeName, driveNames, applyTime)
	if err != nil {
		return diag.Errorf("Error when creating the virtual disk on disk controller %s - %s", storageController, err)
	}
	if applyTime == "Immediate" {
		err = common.WaitForJobToFinish(conn, jobID, common.TimeBetweenAttempts, common.Timeout)
		if err != nil {
			return diag.Errorf("Error. Job %s wasn't able to complete", jobID)
		}
		// Get new volumeID
		volumeID, err := getVolumeID(conn, storageController, volumeName)
		if err != nil {
			return diag.Errorf("Error. The volume ID with volume name %s on %s controller was not found", volumeName, storageController)
		}
		d.Set(biosConfigJobURIStr, "")
		d.SetId(volumeID)
	} else {
		//TODO - Implement for not Immediate scenarios
		d.Set(biosConfigJobURIStr, jobID)
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
	//Get user config
	//If applyTime has been set to Immediate, the volumeID of the resource will be the ODataID of the volume just created.
	//If applyTime is OnReset, the volumeID will be the JobID
	volumeID := d.Id()
	applyTime := d.Get(settingsApplyTimeStr).(string)
	//DELETE VOLUME
	if applyTime == "Immediate" {
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
			storageController := d.Get(storageControllerStr).(string)
			volumeName := d.Get(volumeNameStr).(string)
			actualVolumeID, err := getVolumeID(conn, storageController, volumeName)
			if err != nil {
				return diag.Errorf("Issue when getting the actual volumeID: %s", err)
			}
			//MAYBE WE NEED TO SET A JOB INSTEAD OF DELETING IT RIGHTAWAY
			_, err = deleteVolume(conn, actualVolumeID)
			d.SetId("")
		} else {
			//Get rid of the Job that will create the volume
			//IMPORTART LIMITATION. TO DELETE A TASK IN DELL EMC REDFISH IMPLEMENTATION, NEEDS TO BE DONE THROUGH ITS MANAGER/redfish/v1/Managers/iDRAC.Embedded.1/Jobs/%s
			//KEEPING THIS CODE FOR THE FUTURE
			/*res, err := conn.Delete(volumeID)
			if err != nil {
				return diag.Errorf("Issue when deleting the job: %s", err)
			}
			defer res.Body.Close()
			if res.StatusCode != 200 {
				return diag.Errorf("Issue when deleting the job. The HTTP Delete response returned %v code", res.StatusCode)
			}*/
			err := common.DeleteDellJob(conn, task.ID)
			if err != nil {
				return diag.Errorf("Issue when deleting the task: %s", err)
			}
			d.SetId("")
		}
	}
	return diags
}

func getStorageController(c *gofish.APIClient, diskControllerName string) (*redfish.Storage, error) {
	service := c.Service
	systems, err := service.Systems()
	if err != nil {
		return nil, fmt.Errorf("Error when retreiving the Systems from the Redfish API")
	}
	sg, err := systems[0].Storage()
	if err != nil {
		return nil, fmt.Errorf("Error when retreiving the Storage from %v from the Redfish API", systems[0].Name)
	}
	for _, storage := range sg {
		if storage.Entity.ID == diskControllerName {
			return storage, nil
		}
	}
	return nil, fmt.Errorf("Error. Didn't find the storage controller %v", diskControllerName)
}

func deleteVolume(c *gofish.APIClient, volumeURI string) (jobID string, err error) {
	//TODO - Check if we can delete immediately or if we need to schedule a job
	res, err := c.Delete(volumeURI)
	if err != nil {
		return "", fmt.Errorf("Error while deleting the volume %s", volumeURI)
	}
	defer res.Body.Close()
	jobID = res.Header.Get("Location")
	if len(jobID) == 0 {
		return "", fmt.Errorf("There was some error when retreiving the jobID")
	}
	return jobID, nil
}

func getDrivesStorageController(c *gofish.APIClient, diskControllerName string, driveNames []string) ([]*redfish.Drive, error) {
	var drivesToReturn = []*redfish.Drive{}
	for _, v := range driveNames {
		drive, err := getDrive(c, diskControllerName, v)
		if err != nil {
			return nil, err
		}
		drivesToReturn = append(drivesToReturn, drive)
	}
	return drivesToReturn, nil
}

func getDrive(c *gofish.APIClient, diskControllerName string, driveName string) (*redfish.Drive, error) {
	storage, err := getStorageController(c, diskControllerName)
	if err != nil {
		return nil, err
	}
	drives, err := storage.Drives()
	if err != nil {
		return nil, err
	}
	for _, v := range drives {
		if v.Name == driveName {
			return v, nil
		}
	}
	return nil, fmt.Errorf("The drive %s you were trying to find does not exist", driveName)
}

/*
createVolume creates a virtualdisk on a disk controller by using the redfish API
Parameters:
	c -> gofish API
	diskControllerName -> ID of the disk controller to manage (i.e. RAID.Integrated.1-1)
	raidMode -> raid mode to apply to that set of disks
		Modes:
			- RAID-0 -> "NonRedundant"
			- RAID-1 -> "Mirrored"
			- RAID-5 -> "StripedWithParity"
			- RAID-10 -> "SpannedMirrors"
			- RAID-50 -> "SpannedStripesWithParity"
	volumeName -> Name for the volume
	driveNames -> Drives to use for the raid configuration
*/
func createVolume(c *gofish.APIClient,
	diskControllerName string,
	raidMode string,
	volumeName string,
	driveNames []string,
	applyTime string) (jobID string, err error) {
	//At the moment is creates a virtual disk using all disk from the disk controller
	//Get storage controller to get @odata.id from volumes
	storage, err := getStorageController(c, diskControllerName)
	if err != nil {
		return "", err
	}
	drives, err := getDrivesStorageController(c, diskControllerName, driveNames)
	if err != nil {
		return "", err
	}
	newVolume := make(map[string]interface{})
	newVolume["VolumeType"] = raidMode
	newVolume["Name"] = volumeName
	newVolume["@Redfish.OperationApplyTime"] = applyTime
	var listDrives []map[string]string
	for _, drive := range drives {
		storageDrive := make(map[string]string)
		storageDrive["@odata.id"] = drive.Entity.ODataID
		listDrives = append(listDrives, storageDrive)
	}
	newVolume["Drives"] = listDrives
	volumesURL := fmt.Sprintf("%v/Volumes", storage.ODataID)
	res, err := c.Post(volumesURL, newVolume)
	if err != nil {
		return "", err
	}
	if res.StatusCode != 202 {
		return "", fmt.Errorf("The query was unsucessfull")
	}
	jobID = res.Header.Get("Location")
	if len(jobID) == 0 {
		return "", fmt.Errorf("There was some error when retreiving the jobID")
	}
	return jobID, nil
}

func getVolumeID(c *gofish.APIClient, diskControllerName string, volumeName string) (volumeID string, err error) {
	storage, err := getStorageController(c, diskControllerName)
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
			volumeID = v.ODataID
			return volumeID, nil
		}
	}
	return "", nil
}
