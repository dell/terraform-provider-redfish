/*
Copyright (c) 2020-2025 Dell Inc., or its subsidiaries. All Rights Reserved.

Licensed under the Mozilla Public License Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://mozilla.org/MPL/2.0/


Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package provider

import (
	"fmt"
	"os"
	"regexp"
	"testing"
	"time"

	"github.com/bytedance/mockey"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

// getVolumeImportConf returns the import configuration for the storage volume
func getVolumeImportConf(d *terraform.State, creds TestingServerCredentials) (string, error) {
	id, err := getID(d, "redfish_storage_volume.volume")
	if err != nil {
		return id, err
	}
	return fmt.Sprintf("{\"id\":\"%s\",\"username\":\"%s\",\"password\":\"%s\",\"endpoint\":\"%s\",\"ssl_insecure\":true}",
		id, creds.Username, creds.Password, creds.Endpoint), nil
}

func TestAccRedfishStorageVolume_InvalidController(t *testing.T) {
	version := os.Getenv("TF_TESTING_REDFISH_VERSION")
	if version == "17" {
		t.Skip("Skipping StorageVolume Tests for 17G")
	}
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRedfishResourceStorageVolumeConfig(
					creds,
					"Invalid-ID",
					"TerraformVol1",
					"RAID0",
					drive,
					"Immediate",
					"Off",
					"UnprotectedWriteBack",
					"PowerCycle",
					100,
					200,
					1073323223,
					131072,
				),
				ExpectError: regexp.MustCompile("error when getting the storage"),
			},
		},
	})
}

func TestAccRedfishStorageVolume_InvalidDrive(t *testing.T) {
	version := os.Getenv("TF_TESTING_REDFISH_VERSION")
	if version == "17" {
		t.Skip("Skipping StorageVolume Tests for 17G")
	}
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				PreConfig: func() {
					time.Sleep(180 * time.Second)
				},
				Config: testAccRedfishResourceStorageVolumeConfig(
					creds,
					"RAID.Integrated.1-1",
					"TerraformVol1",
					"RAID0",
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
	version := os.Getenv("TF_TESTING_REDFISH_VERSION")
	if version == "17" {
		t.Skip("Skipping StorageVolume Tests for 17G")
	}
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRedfishResourceStorageVolumeConfig(
					creds,
					"RAID.Integrated.1-1",
					"TerraformVol1",
					"RAID1",
					drive,
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
	version := os.Getenv("TF_TESTING_REDFISH_VERSION")
	if version == "17" {
		t.Skip("Skipping StorageVolume Tests for 17G")
	}
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRedfishResourceStorageVolumeConfig(
					creds,
					"RAID.Integrated.1-1",
					"TerraformVol1",
					"RAID0",
					drive,
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
				ExpectNonEmptyPlan: true,
			},
			{
				Config: testAccRedfishResourceStorageVolumeConfig(
					creds,
					"RAID.Integrated.1-1",
					"TerraformVol1",
					"RAID0",
					drive,
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
				ExpectNonEmptyPlan: true,
			},
			{
				Config: testAccRedfishResourceStorageVolumeConfig(
					creds,
					"RAID.Integrated.1-1",
					"TerraformVol1",
					"RAID0",
					drive,
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
					resource.TestCheckResourceAttr("redfish_storage_volume.volume", "read_cache_policy", "AdaptiveReadAhead"),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccRedfishStorageVolumeCreate_basic(t *testing.T) {
	version := os.Getenv("TF_TESTING_REDFISH_VERSION")
	if version == "17" {
		t.Skip("Skipping StorageVolume Tests for 17G")
	}
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRedfishResourceStorageVolumeMinConfig(
					creds,
					"RAID.Integrated.1-1",
					"TerraformVol1",
					"RAID0",
					drive,
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("redfish_storage_volume.volume", "storage_controller_id", "RAID.Integrated.1-1"),
				),
				// / TBD: non empty plan fix for
				ExpectNonEmptyPlan: true,
			},
			// test import
			{
				ResourceName: "redfish_storage_volume.volume",
				ImportState:  true,
				ImportStateIdFunc: func(d *terraform.State) (string, error) {
					return getVolumeImportConf(d, creds)
				},
				ExpectError: nil,
			},
			// test import -Negative
			{
				ResourceName:  "redfish_storage_volume.volume",
				ImportState:   true,
				ImportStateId: "{\"id\":\"invalid\",\"username\":\"" + creds.Username + "\",\"password\":\"" + creds.Password + "\",\"endpoint\":\"" + creds.Endpoint + "\",\"ssl_insecure\":true}",
				ExpectError:   regexp.MustCompile("There was an error with the API"),
			},
		},
	})
}

func TestAccRedfishStorageVolume_basic(t *testing.T) {
	version := os.Getenv("TF_TESTING_REDFISH_VERSION")
	if version == "17" {
		t.Skip("Skipping StorageVolume Tests for 17G")
	}
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRedfishResourceStorageVolumeConfig(
					creds,
					"RAID.Integrated.1-1",
					"TerraformVol1",
					"RAID0",
					drive,
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
				// / TBD: non empty plan fix
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccRedfishStorageVolume_OnReset(t *testing.T) {
	version := os.Getenv("TF_TESTING_REDFISH_VERSION")
	if version == "17" {
		t.Skip("Skipping StorageVolume Tests for 17G")
	}
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				PreConfig: func() {
					time.Sleep(120 * time.Second)
				},
				Config: testAccRedfishResourceStorageVolumeConfig(
					creds,
					"RAID.Integrated.1-1",
					"TerraformVol1",
					"RAID0",
					drive,
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
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccRedfishStorageVolumeMockNewConfigErr(t *testing.T) {
	version := os.Getenv("TF_TESTING_REDFISH_VERSION")
	if version == "17" {
		t.Skip("Skipping StorageVolume Tests for 17G")
	}
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				PreConfig: func() {
					FunctionMocker = mockey.Mock(NewConfig).Return(nil, fmt.Errorf("mock error")).Build()
				},
				Config: testAccRedfishResourceStorageVolumeConfig(
					creds,
					"RAID.Integrated.1-1",
					"TerraformVol1",
					"RAID0",
					drive,
					"Immediate",
					"AdaptiveReadAhead",
					"UnprotectedWriteBack",
					"PowerCycle",
					100,
					1200,
					1073323222,
					131072,
				),
				ExpectError: regexp.MustCompile(`.*mock error*.`),
			},
			{
				PreConfig: func() {
					if FunctionMocker != nil {
						FunctionMocker.Release()
					}
				},
				Config: testAccRedfishResourceStorageVolumeConfig(
					creds,
					"RAID.Integrated.1-1",
					"TerraformVol1",
					"RAID0",
					drive,
					"Immediate",
					"AdaptiveReadAhead",
					"UnprotectedWriteBack",
					"PowerCycle",
					100,
					1200,
					1073323222,
					131072,
				),
				ExpectNonEmptyPlan: true,
			},
			{
				PreConfig: func() {
					FunctionMocker = mockey.Mock(NewConfig).Return(nil, fmt.Errorf("mock error")).Build()
				},
				Config: testAccRedfishResourceStorageVolumeConfig(
					creds,
					"RAID.Integrated.1-1",
					"TerraformVol1",
					"RAID0",
					drive,
					"Immediate",
					"ReadAhead",
					"UnprotectedWriteBack",
					"PowerCycle",
					100,
					1200,
					1073323222,
					131072,
				),
				ExpectError: regexp.MustCompile(`.*mock error*.`),
			},
			{
				PreConfig: func() {
					if FunctionMocker != nil {
						FunctionMocker.Release()
					}
				},
				Config: testAccRedfishResourceStorageVolumeConfig(
					creds,
					"RAID.Integrated.1-1",
					"TerraformVol1",
					"RAID0",
					drive,
					"Immediate",
					"ReadAhead",
					"UnprotectedWriteBack",
					"PowerCycle",
					100,
					1200,
					1073323222,
					131072,
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccRedfishStorageVolume17GInvalidController(t *testing.T) {
	version := os.Getenv("TF_TESTING_REDFISH_VERSION")
	if version != "17" {
		t.Skip("Skipping StorageVolume Tests for below 17G")
	}
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRedfishResourceStorageVolume17GConfig(
					creds,
					"Invalid-ID",
					"TerraformVol1",
					"RAID0",
					drive,
					"Immediate",
					"Off",
					"WriteThrough",
					"PowerCycle",
					"Enabled",
					100,
					1200,
					107374182400,
					65536,
				),
				ExpectError: regexp.MustCompile("error when getting the storage"),
			},
		},
	})
}

func TestAccRedfishStorageVolume17GInvalidDrive(t *testing.T) {
	version := os.Getenv("TF_TESTING_REDFISH_VERSION")
	if version != "17" {
		t.Skip("Skipping StorageVolume Tests for below 17G")
	}
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				PreConfig: func() {
					time.Sleep(180 * time.Second)
				},
				Config: testAccRedfishResourceStorageVolume17GConfig(
					creds,
					"RAID.SL.1-1",
					"TerraformVol1",
					"RAID0",
					"Invalid-Drive",
					"Immediate",
					"Off",
					"WriteThrough",
					"PowerCycle",
					"Enabled",
					100,
					1200,
					107374182400,
					65536,
				),
				ExpectError: regexp.MustCompile("Error when getting the drives"),
			},
		},
	})
}

func TestAccRedfishStorageVolume17GInvalidVolumeType(t *testing.T) {
	version := os.Getenv("TF_TESTING_REDFISH_VERSION")
	if version != "17" {
		t.Skip("Skipping StorageVolume Tests for below 17G")
	}
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRedfishResourceStorageVolume17GConfig(
					creds,
					"RAID.SL.1-1",
					"TerraformVol1",
					"RAID1",
					drive,
					"Immediate",
					"Off",
					"WriteThrough",
					"PowerCycle",
					"Enabled",
					100,
					1200,
					107374182400,
					65536,
				),
				ExpectError: regexp.MustCompile("Error when creating the virtual disk on disk controller"),
			},
		},
	})
}

func TestAccRedfishStorageVolume17GCreate(t *testing.T) {
	version := os.Getenv("TF_TESTING_REDFISH_VERSION")
	if version != "17" {
		t.Skip("Skipping StorageVolume Tests for below 17G")
	}
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRedfishResourceStorageVolume17GConfig(
					creds,
					"RAID.SL.1-1",
					"TerraformVol1",
					"RAID0",
					drive,
					"Immediate",
					"Off",
					"WriteThrough",
					"PowerCycle",
					"Enabled",
					100,
					1200,
					107374182400,
					65536,
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("redfish_storage_volume.volume", "storage_controller_id", "RAID.SL.1-1"),
					resource.TestCheckResourceAttr("redfish_storage_volume.volume", "read_cache_policy", "Off"),
					resource.TestCheckResourceAttr("redfish_storage_volume.volume", "write_cache_policy", "WriteThrough"),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccRedfishStorageVolume17GUpdate(t *testing.T) {
	version := os.Getenv("TF_TESTING_REDFISH_VERSION")
	if version != "17" {
		t.Skip("Skipping StorageVolume Tests for below 17G")
	}
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRedfishResourceStorageVolume17GConfig(
					creds,
					"RAID.SL.1-1",
					"TerraformVol1",
					"RAID0",
					drive,
					"Immediate",
					"Off",
					"WriteThrough",
					"PowerCycle",
					"Enabled",
					100,
					1200,
					107374182400,
					65536,
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("redfish_storage_volume.volume", "storage_controller_id", "RAID.SL.1-1"),
					resource.TestCheckResourceAttr("redfish_storage_volume.volume", "disk_cache_policy", "Enabled"),
				),
				ExpectNonEmptyPlan: true,
			},
			{
				Config: testAccRedfishResourceStorageVolume17GConfig(
					creds,
					"RAID.SL.1-1",
					"TerraformVol1",
					"RAID0",
					drive,
					"Immediate",
					"Off",
					"WriteThrough",
					"PowerCycle",
					"Disabled",
					100,
					1200,
					107374182400,
					65536,
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("redfish_storage_volume.volume", "storage_controller_id", "RAID.SL.1-1"),
					resource.TestCheckResourceAttr("redfish_storage_volume.volume", "disk_cache_policy", "Disabled"),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccRedfishStorageVolume17GOnReset(t *testing.T) {
	version := os.Getenv("TF_TESTING_REDFISH_VERSION")
	if version != "17" {
		t.Skip("Skipping StorageVolume Tests for below 17G")
	}
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				PreConfig: func() {
					time.Sleep(120 * time.Second)
				},
				Config: testAccRedfishResourceStorageVolume17GConfig(
					creds,
					"RAID.SL.1-1",
					"TerraformVol1",
					"RAID0",
					drive,
					"OnReset",
					"Off",
					"WriteThrough",
					"PowerCycle",
					"Enabled",
					500,
					2000,
					107374182400,
					65536,
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("redfish_storage_volume.volume", "storage_controller_id", "RAID.SL.1-1"),
					resource.TestCheckResourceAttr("redfish_storage_volume.volume", "disk_cache_policy", "Enabled"),
				),
				ExpectNonEmptyPlan: true,
			},
			{
				PreConfig: func() {
					time.Sleep(120 * time.Second)
				},
				Config: testAccRedfishResourceStorageVolume17GConfig(
					creds,
					"RAID.SL.1-1",
					"TerraformVol1",
					"RAID0",
					drive,
					"OnReset",
					"Off",
					"WriteThrough",
					"PowerCycle",
					"Disabled",
					500,
					2000,
					107374182400,
					65536,
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("redfish_storage_volume.volume", "storage_controller_id", "RAID.SL.1-1"),
					resource.TestCheckResourceAttr("redfish_storage_volume.volume", "disk_cache_policy", "Disabled"),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccRedfishStorageVolume17GImport(t *testing.T) {
	version := os.Getenv("TF_TESTING_REDFISH_VERSION")
	if version != "17" {
		t.Skip("Skipping StorageVolume Tests for below 17G")
	}
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRedfishResourceStorageVolume17GConfig(
					creds,
					"RAID.SL.1-1",
					"TerraformVol1",
					"RAID0",
					drive,
					"Immediate",
					"Off",
					"WriteThrough",
					"PowerCycle",
					"Enabled",
					100,
					1200,
					107374182400,
					65536,
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("redfish_storage_volume.volume", "storage_controller_id", "RAID.SL.1-1"),
					resource.TestCheckResourceAttr("redfish_storage_volume.volume", "read_cache_policy", "Off"),
					resource.TestCheckResourceAttr("redfish_storage_volume.volume", "write_cache_policy", "WriteThrough"),
				),
				ExpectNonEmptyPlan: true,
			},
			// test import
			{
				ResourceName: "redfish_storage_volume.volume",
				ImportState:  true,
				ImportStateIdFunc: func(d *terraform.State) (string, error) {
					return getVolumeImportConf(d, creds)
				},
				ExpectError: nil,
			},
			// test import -Negative
			{
				ResourceName:  "redfish_storage_volume.volume",
				ImportState:   true,
				ImportStateId: "{\"id\":\"invalid\",\"username\":\"" + creds.Username + "\",\"password\":\"" + creds.Password + "\",\"endpoint\":\"" + creds.Endpoint + "\",\"ssl_insecure\":true}",
				ExpectError:   regexp.MustCompile("There was an error with the API"),
			},
		},
	})
}

func TestAccRedfishStorageVolume17GMockNewConfigErr(t *testing.T) {
	version := os.Getenv("TF_TESTING_REDFISH_VERSION")
	if version != "17" {
		t.Skip("Skipping StorageVolume Tests for below 17G")
	}
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				PreConfig: func() {
					FunctionMocker = mockey.Mock(NewConfig).Return(nil, fmt.Errorf("mock error")).Build()
				},
				Config: testAccRedfishResourceStorageVolume17GConfig(
					creds,
					"RAID.SL.1-1",
					"TerraformVol1",
					"RAID0",
					drive,
					"Immediate",
					"Off",
					"WriteThrough",
					"PowerCycle",
					"Enabled",
					100,
					1200,
					107374182400,
					65536,
				),
				ExpectError: regexp.MustCompile(`.*mock error*.`),
			},
			{
				PreConfig: func() {
					if FunctionMocker != nil {
						FunctionMocker.Release()
					}
				},
				Config: testAccRedfishResourceStorageVolume17GConfig(
					creds,
					"RAID.SL.1-1",
					"TerraformVol1",
					"RAID0",
					drive,
					"Immediate",
					"Off",
					"WriteThrough",
					"PowerCycle",
					"Enabled",
					100,
					1200,
					107374182400,
					65536,
				),
				ExpectNonEmptyPlan: true,
			},
			{
				PreConfig: func() {
					FunctionMocker = mockey.Mock(NewConfig).Return(nil, fmt.Errorf("mock error")).Build()
				},
				Config: testAccRedfishResourceStorageVolume17GConfig(
					creds,
					"RAID.SL.1-1",
					"TerraformVol1",
					"RAID0",
					drive,
					"Immediate",
					"Off",
					"WriteThrough",
					"PowerCycle",
					"Disabled",
					100,
					1200,
					107374182400,
					65536,
				),
				ExpectError: regexp.MustCompile(`.*mock error*.`),
			},
			{
				PreConfig: func() {
					if FunctionMocker != nil {
						FunctionMocker.Release()
					}
				},
				Config: testAccRedfishResourceStorageVolume17GConfig(
					creds,
					"RAID.SL.1-1",
					"TerraformVol1",
					"RAID0",
					drive,
					"Immediate",
					"Off",
					"WriteThrough",
					"PowerCycle",
					"Disabled",
					100,
					1200,
					107374182400,
					65536,
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

// Wrote this test to test the encrypted property.
// However since we do not have the proper equiptment in our lab and had to borrow will comment out until we do.
// This way the rest of the test can run without failure.
// TODO: Uncomment when we have proper equiptment

// func TestAccRedfishStorageVolume_Encrypted(t *testing.T) {
// 	resource.Test(t, resource.TestCase{
// 		PreCheck:                 func() { testAccPreCheck(t) },
// 		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
// 		Steps: []resource.TestStep{
// 			{
// 				Config: testAccRedfishResourceStorageVolumeEncryptedConfig(
// 					creds,
// 					"RAID.SL.3-1",
// 					"TerraformVol1",
// 					"RAID0",
// 					drive,
// 					"Immediate",
// 					"Off",
// 					"WriteThrough",
// 					"PowerCycle",
// 					500,
// 					2000,
// 					true,
// 				),
// 				Check: resource.ComposeAggregateTestCheckFunc(
// 					resource.TestCheckResourceAttr("redfish_storage_volume.volume", "storage_controller_id", "RAID.SL.3-1"),
// 					resource.TestCheckResourceAttr("redfish_storage_volume.volume", "write_cache_policy", "WriteThrough"),
// 					resource.TestCheckResourceAttr("redfish_storage_volume.volume", "encrypted", "true"),
// 				),
// 				ExpectNonEmptyPlan: true,
// 			},
// 		},
// 	})
// }

func testAccRedfishResourceStorageVolumeConfig(testingInfo TestingServerCredentials,
	storage_controller_id string,
	volume_name string,
	raid_type string,
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
		  endpoint     = "%s"
		  ssl_insecure = true
		}

		storage_controller_id = "%s"
		volume_name           = "%s"
		raid_type           = "%s"
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
		raid_type,
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

func testAccRedfishResourceStorageVolume17GConfig(testingInfo TestingServerCredentials,
	storage_controller_id string,
	volume_name string,
	raid_type string,
	drives string,
	settings_apply_time string,
	read_cache_policy string,
	write_cache_policy string,
	reset_type string,
	disk_cache_policy string,
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
		  endpoint     = "%s"
		  ssl_insecure = true
		}

		storage_controller_id = "%s"
		volume_name           = "%s"
		raid_type           = "%s"
		drives                = ["%s"]
		settings_apply_time   = "%s"
		read_cache_policy 	  = "%s"
		write_cache_policy 	  = "%s"
		reset_type 			  = "%s"
		disk_cache_policy 	  = "%s"
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
		raid_type,
		drives,
		settings_apply_time,
		read_cache_policy,
		write_cache_policy,
		reset_type,
		disk_cache_policy,
		reset_timeout,
		volume_job_timeout,
		capacity_bytes,
		optimum_io_size_bytes,
	)
}

func testAccRedfishResourceStorageVolumeEncryptedConfig(testingInfo TestingServerCredentials,
	storage_controller_id string,
	volume_name string,
	raid_type string,
	drives string,
	settings_apply_time string,
	read_cache_policy string,
	write_cache_policy string,
	reset_type string,
	reset_timeout int,
	volume_job_timeout int,
	encrypted bool,
) string {
	return fmt.Sprintf(`
	resource "redfish_storage_volume" "volume" {
		redfish_server {
		  user         = "%s"
		  password     = "%s"
		  endpoint     = "%s"
		  ssl_insecure = true
		}

		storage_controller_id = "%s"
		volume_name           = "%s"
		raid_type           = "%s"
		drives                = ["%s"]
		settings_apply_time   = "%s"
		read_cache_policy 	  = "%s"
		write_cache_policy 	  = "%s"
		reset_type 			  = "%s"
		reset_timeout 		  = %d
		volume_job_timeout 	  = %d
		encrypted = %t
	  }
	  `,
		testingInfo.Username,
		testingInfo.Password,
		testingInfo.Endpoint,
		storage_controller_id,
		volume_name,
		raid_type,
		drives,
		settings_apply_time,
		read_cache_policy,
		write_cache_policy,
		reset_type,
		reset_timeout,
		volume_job_timeout,
		encrypted,
	)
}

func testAccRedfishResourceStorageVolumeMinConfig(testingInfo TestingServerCredentials,
	storage_controller_id string,
	volume_name string,
	raid_type string,
	drives string,
) string {
	return fmt.Sprintf(`
	resource "redfish_storage_volume" "volume" {
		redfish_server {
		  user         = "%s"
		  password     = "%s"
		  endpoint     = "%s"
		  ssl_insecure = true
		}
	    system_id = "System.Embedded.1"
		storage_controller_id = "%s"
		volume_name           = "%s"
		raid_type           = "%s"
		drives                = ["%s"]
	  }
	  `,
		testingInfo.Username,
		testingInfo.Password,
		testingInfo.Endpoint,
		storage_controller_id,
		volume_name,
		raid_type,
		drives,
	)
}
