package provider

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccRedfishStorageVolume_InvalidController(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRedfishResourceStorageVolumeConfig(
					creds,
					"Invalid-ID",
					"TerraformVol1",
					"NonRedundant",
					"Physical Disk 0:1:0",
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

func TestAccRedfishStorageVolume_InvalidDrive(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
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

func TestAccRedfishStorageVolume_InvalidVolumeType(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRedfishResourceStorageVolumeConfig(
					creds,
					"RAID.Integrated.1-1",
					"TerraformVol1",
					"Mirrored",
					"Physical Disk 0:1:0",
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

func TestAccRedfishStorageVolumeCreate_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRedfishResourceStorageVolumeMinConfig(
					creds,
					"RAID.Integrated.1-1",
					"TerraformVol1",
					"NonRedundant",
					"Physical Disk 0:1:1",
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("redfish_storage_volume.volume", "storage_controller_id", "RAID.Integrated.1-1"),
					resource.TestCheckResourceAttr("redfish_storage_volume.volume", "volume_type", "NonRedundant"),
				),
			},
		},
	})
}

func TestAccRedfishStorageVolume_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRedfishResourceStorageVolumeConfig(
					creds,
					"RAID.Integrated.1-1",
					"TerraformVol1",
					"NonRedundant",
					"Physical Disk 0:1:1",
					"Immediate",
					"AdaptiveReadAhead",
					"UnprotectedWriteBack",
					"PowerCycle",
					100,
					1200,
					1073323222,
					131072,
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("redfish_storage_volume.volume", "storage_controller_id", "RAID.Integrated.1-1"),
					resource.TestCheckResourceAttr("redfish_storage_volume.volume", "write_cache_policy", "UnprotectedWriteBack"),
				),
			},
		},
	})
}

func TestAccRedfishStorageVolumeUpdate_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRedfishResourceStorageVolumeConfig(
					creds,
					"RAID.Integrated.1-1",
					"TerraformVol1",
					"NonRedundant",
					"Physical Disk 0:1:1",
					"Immediate",
					"AdaptiveReadAhead",
					"UnprotectedWriteBack",
					"PowerCycle",
					100,
					1200,
					1073323222,
					131072,
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("redfish_storage_volume.volume", "storage_controller_id", "RAID.Integrated.1-1"),
					resource.TestCheckResourceAttr("redfish_storage_volume.volume", "read_cache_policy", "AdaptiveReadAhead"),
				),
			},

			{
				Config: testAccRedfishResourceStorageVolumeConfig(
					creds,
					"RAID.Integrated.1-1",
					"TerraformVol1",
					"NonRedundant",
					"Physical Disk 0:1:1",
					"Immediate",
					"ReadAhead",
					"UnprotectedWriteBack",
					"PowerCycle",
					100,
					1200,
					1073323222,
					131072,
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("redfish_storage_volume.volume", "storage_controller_id", "RAID.Integrated.1-1"),
					resource.TestCheckResourceAttr("redfish_storage_volume.volume", "read_cache_policy", "ReadAhead"),
				),
			},
		},
	})
}

func TestAccRedfishStorageVolume_OnReset(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRedfishResourceStorageVolumeConfig(
					creds,
					"RAID.Integrated.1-1",
					"TerraformVol1",
					"NonRedundant",
					"Physical Disk 0:1:1",
					"OnReset",
					"AdaptiveReadAhead",
					"UnprotectedWriteBack",
					"PowerCycle",
					500,
					2000,
					1073323222,
					131072,
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("redfish_storage_volume.volume", "storage_controller_id", "RAID.Integrated.1-1"),
					resource.TestCheckResourceAttr("redfish_storage_volume.volume", "write_cache_policy", "UnprotectedWriteBack"),
				),
			},
		},
	})
}

// Test to import volume - positive
func TestAccRedfishStorageVolumeImport_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:        testAccRedfishResourceStorageVolumeImportConfig(),
				ResourceName:  "redfish_storage_volume.volume",
				ImportState:   true,
				ImportStateId: "{\"id\":\"/redfish/v1/Systems/System.Embedded.1/Storage/RAID.Integrated.1-1/Volumes/Disk.Virtual.1:RAID.Integrated.1-1\",\"username\":\"" + creds.Username + "\",\"password\":\"" + creds.Password + "\",\"endpoint\":\"" + creds.Endpoint + "\",\"ssl_insecure\":true}",
				ExpectError:   nil,
			},
		},
	})
}

// Test to import volume - negative
func TestAccRedfishStorageVolumeImport_invalid(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:        testAccRedfishResourceStorageVolumeImportConfig(),
				ResourceName:  "redfish_storage_volume.volume",
				ImportState:   true,
				ImportStateId: "{\"id\":\"invalid\",\"username\":\"" + creds.Username + "\",\"password\":\"" + creds.Password + "\",\"endpoint\":\"" + creds.Endpoint + "\",\"ssl_insecure\":true}",
				ExpectError:   regexp.MustCompile("There was an error with the API"),
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

func testAccRedfishResourceStorageVolumeImportConfig() string {
	return fmt.Sprintf(`
	resource "redfish_storage_volume" "volume" {
	}`)
}
