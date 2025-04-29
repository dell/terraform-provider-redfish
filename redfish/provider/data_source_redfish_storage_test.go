/*
Copyright (c) 2021-2025 Dell Inc., or its subsidiaries. All Rights Reserved.

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
	"terraform-provider-redfish/gofish/dell"
	"testing"

	"github.com/bytedance/mockey"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

const mockErrorMessage = "mock error"

func TestAccRedfishStorageDataSourcefetch(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRedfishDataSourceStorageConfig(creds),
			},
			{
				Config: testAccStorageDatasourceWithControllerID(creds, os.Getenv("TF_STORAGE_CONTROLLER_IDS")),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.redfish_storage.storage", "storage.0.storage_controller_id"),
				),
			},
			{
				Config: testAccStorageDatasourceWithControllerName(creds, os.Getenv("TF_STORAGE_CONTROLLER_NAMES")),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.redfish_storage.storage", "storage.0.name"),
				),
			},
			{
				Config: testAccStorageDatasourceWithBothSysIDandControllerName(creds, os.Getenv("TF_STORAGE_CONTROLLER_NAMES")),
			},
			{
				Config: testAccStorageDatasourceWithBothSysIDandControllerId(creds, os.Getenv("TF_STORAGE_CONTROLLER_IDS")),
			},
		},
	})
}

func TestAccRedfishStorageDataSourceReadError(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRedfishDataSourceStorageConfig(creds),
			},
			{
				PreConfig: func() {
					FunctionMocker = mockey.Mock(NewConfig).Return(nil, fmt.Errorf(mockErrorMessage)).Build()
				},
				Config:      testAccRedfishDataSourceStorageConfig(creds),
				ExpectError: regexp.MustCompile(`.*` + mockErrorMessage + `*.`),
			},
		},
	})
	if FunctionMocker != nil {
		FunctionMocker.Release()
	}
}

func TestAccRedfishStorageDataSourceGetResourceError(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRedfishDataSourceStorageConfig(creds),
			},
			{
				PreConfig: func() {
					if FunctionMocker != nil {
						FunctionMocker.Release()
					}
					FunctionMocker = mockey.Mock(getSystemResource).Return(nil, fmt.Errorf(mockErrorMessage)).Build()
				},
				Config:      testAccRedfishDataSourceStorageConfig(creds),
				ExpectError: regexp.MustCompile(`.*` + mockErrorMessage + `*.`),
			},
			{
				PreConfig: func() {
					if FunctionMocker != nil {
						FunctionMocker.Release()
					}
					FunctionMocker = mockey.Mock(dell.Storage).Return(nil, fmt.Errorf(mockErrorMessage)).Build()
				},
				Config:      testAccRedfishDataSourceStorageConfig(creds),
				ExpectError: regexp.MustCompile(`.*` + mockErrorMessage + `*.`),
			},
		},
	})
	if FunctionMocker != nil {
		FunctionMocker.Release()
	}
}

// controller_names = ["PERC H730P Mini"]
// controller_ids = ["AHCI.Embedded.2-1"]

func testAccRedfishDataSourceStorageConfig(testingInfo TestingServerCredentials) string {
	return fmt.Sprintf(`
	data "redfish_storage" "storage" {
		redfish_server {
		  user         = "%s"
		  password     = "%s"
		  endpoint     = "%s"
		  ssl_insecure = true
		}
	  }
		`,
		testingInfo.Username,
		testingInfo.Password,
		testingInfo.Endpoint,
	)
}

func testAccStorageDatasourceWithControllerID(testingInfo TestingServerCredentials, controllerID string) string {
	return fmt.Sprintf(`
	data "redfish_storage" "storage" {
		redfish_server {
		  user         = "%s"
		  password     = "%s"
		  endpoint     = "%s"
		  ssl_insecure = true
		}
		controller_ids = %s
	  }
		`,
		testingInfo.Username,
		testingInfo.Password,
		testingInfo.Endpoint,
		controllerID,
	)
}

func testAccStorageDatasourceWithControllerName(testingInfo TestingServerCredentials, controllerName string) string {
	return fmt.Sprintf(`
	data "redfish_storage" "storage" {
		redfish_server {
		  user         = "%s"
		  password     = "%s"
		  endpoint     = "%s"
		  ssl_insecure = true
		}
		controller_names = %s
	  }
		`,
		testingInfo.Username,
		testingInfo.Password,
		testingInfo.Endpoint,
		controllerName,
	)
}

func testAccStorageDatasourceWithBothSysIDandControllerName(testingInfo TestingServerCredentials, controllerName string) string {
	return fmt.Sprintf(`
	data "redfish_storage" "storage" {
		redfish_server {
		  user         = "%s"
		  password     = "%s"
		  endpoint     = "%s"
		  ssl_insecure = true
		}
		system_id = "System.Embedded.1"
		controller_names = %s
	  }
		`,
		testingInfo.Username,
		testingInfo.Password,
		testingInfo.Endpoint,
		controllerName,
	)
}

func testAccStorageDatasourceWithBothSysIDandControllerId(testingInfo TestingServerCredentials, controllerID string) string {
	return fmt.Sprintf(`
	data "redfish_storage" "storage" {
		redfish_server {
		  user         = "%s"
		  password     = "%s"
		  endpoint     = "%s"
		  ssl_insecure = true
		}
		system_id = "System.Embedded.1"
		controller_ids = %s
	  }
		`,
		testingInfo.Username,
		testingInfo.Password,
		testingInfo.Endpoint,
		controllerID,
	)
}
