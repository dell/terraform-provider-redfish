package redfish

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccRedfishStorageVolumeCreate_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccRedfishResourceStorageVolumeMinConfig(
					creds,
					"RAID.Integrated.1-1",
					"TerraformVol1",
					"NonRedundant",
					"Solid State Disk 0:0:1",
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("redfish_storage_volume.volume", "storage_controller_id", "RAID.Integrated.1-1"),
					resource.TestCheckResourceAttr("redfish_storage_volume.volume", "volume_type", "NonRedundant"),
				),
			},
		},
	})
}

func TestAccRedfishStorageVolume_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccRedfishResourceStorageVolumeConfig(
					creds,
					"RAID.Integrated.1-1",
					"TerraformVol1",
					"NonRedundant",
					"Solid State Disk 0:0:1",
					"Immediate",
					"Off",
					"UnprotectedWriteBack",
					"PowerCycle",
					100,
					200,
					1073323223,
					131072,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("redfish_storage_volume.volume", "storage_controller_id", "RAID.Integrated.1-1"),
					resource.TestCheckResourceAttr("redfish_storage_volume.volume", "write_cache_policy", "UnprotectedWriteBack"),
				),
			},
		},
	})
}

func TestAccRedfishStorageVolume_InvaldiController(t *testing.T) {
	resource.Test(t, resource.TestCase{
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccRedfishResourceStorageVolumeConfig(
					creds,
					"Invalid-ID",
					"TerraformVol1",
					"NonRedundant",
					"Solid State Disk 0:0:1",
					"Immediate",
					"Off",
					"UnprotectedWriteBack",
					"PowerCycle",
					100,
					200,
					1073323223,
					131072,
				),
				ExpectError: regexp.MustCompile("Error when getting the storage"),
			},
		},
	})
}

func TestAccRedfishStorageVolume_InvaldiDrive(t *testing.T) {
	resource.Test(t, resource.TestCase{
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccRedfishResourceStorageVolumeConfig(
					creds,
					"RAID.Integrated.1-1",
					"TerraformVol1",
					"NonRedundant",
					"Invalid-Drive",
					"Immediate",
					"Off",
					"UnprotectedWriteBack",
					"PowerCycle",
					100,
					200,
					1073323223,
					131072,
				),
				ExpectError: regexp.MustCompile("Error when getting the drives"),
			},
		},
	})
}

func TestAccRedfishStorageVolume_InvaldiVolumeType(t *testing.T) {
	resource.Test(t, resource.TestCase{
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccRedfishResourceStorageVolumeConfig(
					creds,
					"RAID.Integrated.1-1",
					"TerraformVol1",
					"Mirrored",
					"Solid State Disk 0:0:1",
					"Immediate",
					"Off",
					"UnprotectedWriteBack",
					"PowerCycle",
					100,
					200,
					1073323223,
					131072,
				),
				ExpectError: regexp.MustCompile("Error when creating the virtual disk on disk controller"),
			},
		},
	})
}

func TestAccRedfishStorageVolumeUpdate_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccRedfishResourceStorageVolumeConfig(
					creds,
					"RAID.Integrated.1-1",
					"TerraformVol1",
					"NonRedundant",
					"Solid State Disk 0:0:1",
					"Immediate",
					"Off",
					"UnprotectedWriteBack",
					"PowerCycle",
					100,
					200,
					1073323223,
					131072,
				),

				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("redfish_storage_volume.volume", "volume_name", "TerraformVol1"),
					resource.TestCheckResourceAttr("redfish_storage_volume.volume", "read_cache_policy", "Off"),
				),
			},

			{
				Config: testAccRedfishResourceStorageVolumeConfig(
					creds,
					"RAID.Integrated.1-1",
					"TerraformVol1",
					"NonRedundant",
					"Solid State Disk 0:0:1",
					"Immediate",
					"AdaptiveReadAhead",
					"UnprotectedWriteBack",
					"PowerCycle",
					100,
					200,
					1073323223,
					131072,
				),

				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("redfish_storage_volume.volume", "volume_name", "TerraformVol1"),
					resource.TestCheckResourceAttr("redfish_storage_volume.volume", "read_cache_policy", "AdaptiveReadAhead"),
				),
			},
		},
	})
}

func testAccRedfishResourceStorageVolumeConfig(testingInfo TestingServerCredentials,
	storage_controller_id string,
	volume_name string,
	volume_type string,
	drives string,
	settings_apply_time string,
	read_cache_policy string,
	write_cache_policy string,
	reset_type string,
	reset_timeout int,
	volume_job_timeout int,
	capacity_bytes int,
	optimum_io_size_bytes int,

) string {
	return fmt.Sprintf(`
	resource "redfish_storage_volume" "volume" {
		redfish_server {
		  user         = "%s"
		  password     = "%s"
		  endpoint     = "https://%s"
		  ssl_insecure = true
		}
	  
		storage_controller_id = "%s"
		volume_name           = "%s"
		volume_type           = "%s"
		drives                = ["%s"]
		settings_apply_time   = "%s"
		read_cache_policy 	  = "%s"
		write_cache_policy 	  = "%s"
		reset_type 			  = "%s"
		reset_timeout 		  = %d
		volume_job_timeout 	  = %d
		capacity_bytes = %d
  		optimum_io_size_bytes = %d
	  }
	  `,
		testingInfo.Username,
		testingInfo.Password,
		testingInfo.Endpoint,
		storage_controller_id,
		volume_name,
		volume_type,
		drives,
		settings_apply_time,
		read_cache_policy,
		write_cache_policy,
		reset_type,
		reset_timeout,
		volume_job_timeout,
		capacity_bytes,
		optimum_io_size_bytes,
	)
}

func testAccRedfishResourceStorageVolumeMinConfig(testingInfo TestingServerCredentials,
	storage_controller_id string,
	volume_name string,
	volume_type string,
	drives string,
) string {
	return fmt.Sprintf(`
	resource "redfish_storage_volume" "volume" {
		redfish_server {
		  user         = "%s"
		  password     = "%s"
		  endpoint     = "https://%s"
		  ssl_insecure = true
		}
	  
		storage_controller_id = "%s"
		volume_name           = "%s"
		volume_type           = "%s"
		drives                = ["%s"]
	  }
	  `,
		testingInfo.Username,
		testingInfo.Password,
		testingInfo.Endpoint,
		storage_controller_id,
		volume_name,
		volume_type,
		drives,
	)
}
