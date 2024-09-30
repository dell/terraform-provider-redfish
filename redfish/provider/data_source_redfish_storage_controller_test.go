/*
Copyright (c) 2024 Dell Inc., or its subsidiaries. All Rights Reserved.

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
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccRedfishStorageControllerDataSource_fetch(t *testing.T) {
	storageControllerDatasourceName := "data.redfish_storage_controller.test"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRedfishDataSourceStorageControllerConfig(creds),
			},
			{
				Config: testAccStorageControllerDatasourceWithEmptySystemFilter(creds),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(storageControllerDatasourceName, "storage_controllers.#", "3"),
				),
			},
			{
				Config: testAccStorageControllerDatasourceWithSystemID(creds),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(storageControllerDatasourceName, "storage_controllers.#", "3"),
					resource.TestMatchResourceAttr(storageControllerDatasourceName, "storage_controllers.0.odata_id", regexp.MustCompile(`.*System.Embedded.1*.`)),
					resource.TestMatchResourceAttr(storageControllerDatasourceName, "storage_controllers.1.odata_id", regexp.MustCompile(`.*System.Embedded.1*.`)),
					resource.TestMatchResourceAttr(storageControllerDatasourceName, "storage_controllers.2.odata_id", regexp.MustCompile(`.*System.Embedded.1*.`)),
				),
			},
			{
				Config: testAccStorageControllerDatasourceWithStorageID(creds),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(storageControllerDatasourceName, "storage_controllers.#", "1"),
					resource.TestMatchResourceAttr(storageControllerDatasourceName, "storage_controllers.0.odata_id", regexp.MustCompile(`.*System.Embedded.1*.`)),
					resource.TestMatchResourceAttr(storageControllerDatasourceName, "storage_controllers.0.odata_id", regexp.MustCompile(`.*RAID.Integrated.1-1*.`)),
				),
			},
			{
				Config: testAccStorageControllerDatasourceWithControllerID(creds),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(storageControllerDatasourceName, "storage_controllers.#", "1"),
					resource.TestMatchResourceAttr(storageControllerDatasourceName, "storage_controllers.0.odata_id", regexp.MustCompile(`.*System.Embedded.1*.`)),
					resource.TestMatchResourceAttr(storageControllerDatasourceName, "storage_controllers.0.odata_id", regexp.MustCompile(`.*RAID.Integrated.1-1*.`)),
					resource.TestMatchResourceAttr(storageControllerDatasourceName, "storage_controllers.0.odata_id", regexp.MustCompile(`.*RAID.Integrated.1-1*.`)),
				),
			},
			{
				Config: testAccStorageControllerDatasourceWithMultipleStorageIDs(creds),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(storageControllerDatasourceName, "storage_controllers.#", "2"),
					resource.TestMatchResourceAttr(storageControllerDatasourceName, "storage_controllers.0.odata_id", regexp.MustCompile(`.*System.Embedded.1*.`)),
					resource.TestMatchResourceAttr(storageControllerDatasourceName, "storage_controllers.0.odata_id", regexp.MustCompile(`.*AHCI.Embedded.1-1*.`)),
					resource.TestMatchResourceAttr(storageControllerDatasourceName, "storage_controllers.0.odata_id", regexp.MustCompile(`.*AHCI.Embedded.1-1*.`)),
					resource.TestMatchResourceAttr(storageControllerDatasourceName, "storage_controllers.1.odata_id", regexp.MustCompile(`.*System.Embedded.1*.`)),
					resource.TestMatchResourceAttr(storageControllerDatasourceName, "storage_controllers.1.odata_id", regexp.MustCompile(`.*RAID.Integrated.1-1*.`)),
					resource.TestMatchResourceAttr(storageControllerDatasourceName, "storage_controllers.1.odata_id", regexp.MustCompile(`.*RAID.Integrated.1-1*.`)),
				),
			},
			{
				Config:      testAccStorageControllerDatasourceWithInvalidSystemID(creds),
				ExpectError: regexp.MustCompile(`.*Error one or more of the filtered system ids are not valid*.`),
			},
			{
				Config:      testAccStorageControllerDatasourceWithInvalidStorageID(creds),
				ExpectError: regexp.MustCompile(`.*Error one or more of the filtered storage ids are not valid*.`),
			},
			{
				Config:      testAccStorageControllerDatasourceWithInvalidControllerID(creds),
				ExpectError: regexp.MustCompile(`.*Error one or more of the filtered storage controller ids are not valid*.`),
			},
		},
	})
}

func testAccRedfishDataSourceStorageControllerConfig(testingInfo TestingServerCredentials) string {
	return fmt.Sprintf(`
	data "redfish_storage_controller" "test" {	  
		redfish_server {
		  user         = "%s"
		  password     = "%s"
		  endpoint     = "https://%s"
		  ssl_insecure = true
		}
	}`,
		testingInfo.Username,
		testingInfo.Password,
		testingInfo.Endpoint,
	)
}

func testAccStorageControllerDatasourceWithEmptySystemFilter(testingInfo TestingServerCredentials) string {
	return fmt.Sprintf(`
	data "redfish_storage_controller" "test" {	  
		redfish_server {
		  user         = "%s"
		  password     = "%s"
		  endpoint     = "https://%s"
		  ssl_insecure = true
		}
		storage_controller_filter {
			systems = []
		}
	}`,
		testingInfo.Username,
		testingInfo.Password,
		testingInfo.Endpoint,
	)
}

func testAccStorageControllerDatasourceWithSystemID(testingInfo TestingServerCredentials) string {
	return fmt.Sprintf(`
	data "redfish_storage_controller" "test" {	  
		redfish_server {
		  user         = "%s"
		  password     = "%s"
		  endpoint     = "https://%s"
		  ssl_insecure = true
		}
		storage_controller_filter {
			systems = [
				{
				system_id = "System.Embedded.1"
				}
			]
		}
	}`,
		testingInfo.Username,
		testingInfo.Password,
		testingInfo.Endpoint,
	)
}

func testAccStorageControllerDatasourceWithStorageID(testingInfo TestingServerCredentials) string {
	return fmt.Sprintf(`
	data "redfish_storage_controller" "test" {	  
		redfish_server {
		  user         = "%s"
		  password     = "%s"
		  endpoint     = "https://%s"
		  ssl_insecure = true
		}
		storage_controller_filter {
			systems = [
				{
				system_id = "System.Embedded.1"
				storages = [
					{
					storage_id = "RAID.Integrated.1-1"
					}
				]
				}
			]
		}
	}`,
		testingInfo.Username,
		testingInfo.Password,
		testingInfo.Endpoint,
	)
}

func testAccStorageControllerDatasourceWithControllerID(testingInfo TestingServerCredentials) string {
	return fmt.Sprintf(`
	data "redfish_storage_controller" "test" {	  
		redfish_server {
		  user         = "%s"
		  password     = "%s"
		  endpoint     = "https://%s"
		  ssl_insecure = true
		}
		storage_controller_filter {
			systems = [
				{
				system_id = "System.Embedded.1"
				storages = [
					{
					storage_id = "RAID.Integrated.1-1"
					controller_ids = ["RAID.Integrated.1-1"]
					}
				]
				}
			]
		}
	}`,
		testingInfo.Username,
		testingInfo.Password,
		testingInfo.Endpoint,
	)
}

func testAccStorageControllerDatasourceWithMultipleStorageIDs(testingInfo TestingServerCredentials) string {
	return fmt.Sprintf(`
	data "redfish_storage_controller" "test" {	  
		redfish_server {
		  user         = "%s"
		  password     = "%s"
		  endpoint     = "https://%s"
		  ssl_insecure = true
		}
		storage_controller_filter {
			systems = [
				{
				system_id = "System.Embedded.1"
				storages = [
					{
					storage_id = "RAID.Integrated.1-1"
					controller_ids = ["RAID.Integrated.1-1"]
					},
					{
					storage_id = "AHCI.Embedded.1-1"
					controller_ids = ["AHCI.Embedded.1-1"]
					}
				]
				}
			]
		}
	}`,
		testingInfo.Username,
		testingInfo.Password,
		testingInfo.Endpoint,
	)
}

func testAccStorageControllerDatasourceWithInvalidSystemID(testingInfo TestingServerCredentials) string {
	return fmt.Sprintf(`
	data "redfish_storage_controller" "test" {	  
		redfish_server {
		  user         = "%s"
		  password     = "%s"
		  endpoint     = "https://%s"
		  ssl_insecure = true
		}
		storage_controller_filter {
			systems = [
				{
				system_id = "InvalidSystemID"
				storages = [
					{
					storage_id = "RAID.Integrated.1-1"
					controller_ids = ["RAID.Integrated.1-1"]
					}
				]
				}
			]
		}
	}`,
		testingInfo.Username,
		testingInfo.Password,
		testingInfo.Endpoint,
	)
}

func testAccStorageControllerDatasourceWithInvalidStorageID(testingInfo TestingServerCredentials) string {
	return fmt.Sprintf(`
	data "redfish_storage_controller" "test" {	  
		redfish_server {
		  user         = "%s"
		  password     = "%s"
		  endpoint     = "https://%s"
		  ssl_insecure = true
		}
		storage_controller_filter {
			systems = [
				{
				system_id = "System.Embedded.1"
				storages = [
					{
					storage_id = "InvalidStorageID"
					controller_ids = ["RAID.Integrated.1-1"]
					}
				]
				}
			]
		}
	}`,
		testingInfo.Username,
		testingInfo.Password,
		testingInfo.Endpoint,
	)
}

func testAccStorageControllerDatasourceWithInvalidControllerID(testingInfo TestingServerCredentials) string {
	return fmt.Sprintf(`
	data "redfish_storage_controller" "test" {	  
		redfish_server {
		  user         = "%s"
		  password     = "%s"
		  endpoint     = "https://%s"
		  ssl_insecure = true
		}
		storage_controller_filter {
			systems = [
				{
				system_id = "System.Embedded.1"
				storages = [
					{
					storage_id = "RAID.Integrated.1-1"
					controller_ids = ["InvalidControllerID"]
					}
				]
				}
			]
		}
	}`,
		testingInfo.Username,
		testingInfo.Password,
		testingInfo.Endpoint,
	)
}
