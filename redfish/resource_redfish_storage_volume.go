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
	jobID, err := createVolume(conn, storageController, raidLevel, volumeName, driveNames)
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
		d.SetId(volumeID)
	}
	//TODO - Implement for not Immediate scenarios
	//TODO - Implement for not Immediate scenarios
	//TODO - Implement for not Immediate scenarios
	return diags
}

func resourceStorageVolumeRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	/*conn := m.(*gofish.APIClient)
	storageController := d.Get(storageControllerStr).(string)
	volumeName := d.Get(volumeNameStr).(string)
	volumeID, err := getVolumeID(conn, storageController, volumeName)
	if err != nil {
		return diag.Errorf("Error when creating the virtual disk on disk controller %s - %s", storageController, err)
	}
	d.SetId(volumeID)*/
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
	}
	//TODO - Implement for not Immediate scenarios
	//TODO - Implement for not Immediate scenarios
	//TODO - Implement for not Immediate scenarios
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

func deleteAllVolumes(c *gofish.APIClient, diskControllerName string) (err error) {
	storage, err := getStorageController(c, diskControllerName)
	if err != nil {
		return err
	}
	// TEST CODE
	/*opeartionValues, err := storage.GetOperationApplyTimeValues()
	if err != nil {
		return err
	}
	fmt.Printf(string(len(opeartionValues)))*/
	// END TEST CODE
	volumes, err := storage.Volumes()
	if err != nil {
		return fmt.Errorf("Error when retreiving volumes from %v from the Redfish API", storage.Entity.Name)
	}
	for _, v := range volumes {
		res, err := c.Delete(v.Entity.ODataID)
		if err != nil {
			return fmt.Errorf("Error while deleting the volume %v", v.Entity.Name)
		}
		defer res.Body.Close()
		fmt.Printf(v.Entity.ODataID)
	}
	return nil
}

func deleteVolume(c *gofish.APIClient, volumeURI string) (jobID string, err error) {
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
func createVolume(c *gofish.APIClient, diskControllerName string, raidMode string, volumeName string, driveNames []string) (jobID string, err error) {
	//At the moment is creates a virtual disk using all disk from the disk controller
	//Get storage controller to get @odata.id from volumes
	storage, err := getStorageController(c, diskControllerName)
	if err != nil {
		return "", err
	}
	/*drives, err := getAllDrivesStorageController(c, diskControllerName)
	if err != nil {
		panic(err)
	}*/
	drives, err := getDrivesStorageController(c, diskControllerName, driveNames)
	if err != nil {
		return "", err
	}
	newVolume := make(map[string]interface{})
	newVolume["VolumeType"] = raidMode
	newVolume["Name"] = volumeName
	//newVolume["Drives"] = []interface{}{}
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
		}
	}
	return volumeID, nil
}
