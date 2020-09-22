package redfish

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/stmcginnis/gofish"
	"github.com/stmcginnis/gofish/common"
	"github.com/stmcginnis/gofish/redfish"
)

func resourceRedfishStorageVolume() *schema.Resource {
	return &schema.Resource{
		CreateContext: xxxUpdate,
		ReadContext:   xxxRead,
		UpdateContext: xxxUpdate,
		DeleteContext: xxxDelete,
		Schema: map[string]*schema.Schema{
			"storage_controller": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "This value must be the disk controller the user want to manage. I.e: RAID.Integrated.1-1",
			},
			"raid_level": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "This value specifies the raid level the virtual disk is going to have. Possible values are: NonRedundant (RAID-0), Mirrored (RAID-1), StripedWithParity (RAID-5), SpannedMirrors (RAID-10) or SpannedStripesWithParity (RAID-50)",
			},
			"settings_apply_time": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: xxxx,
			},
			"bios_config_job_uri": {
				Type:        schema.TypeString,
				Description: "Volume configuration job uri",
				Computed:    true,
			},
		},
	}
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
	// opeartionValues, err := storage.GetOperationApplyTimeValues()
	// if err != nil {
	// 	return err
	// }
	// fmt.Printf(string(len(opeartionValues)))
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

func getAllDrivesStorageController(c *gofish.APIClient, diskControllerName string) ([]*redfish.Drive, error) {
	storage, err := getStorageController(c, diskControllerName)
	if err != nil {
		return nil, err
	}
	drives, err := storage.Drives()
	if err != nil {
		return nil, err
	}
	return drives, nil
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
*/
func createVolume(c *gofish.APIClient, diskControllerName string, raidMode string, volumeName string) (err error) {
	//At the moment is creates a virtual disk using all disk from the disk controller
	//Get storage controller to get @odata.id from volumes
	storage, err := getStorageController(c, diskControllerName)
	if err != nil {
		return err
	}
	drives, err := getAllDrivesStorageController(c, diskControllerName)
	if err != nil {
		panic(err)
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
		return err
	}
	if res.StatusCode != 202 {
		return fmt.Errorf("The query was unsucessfull")
	}
	return nil
}
