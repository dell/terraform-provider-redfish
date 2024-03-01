/*
Copyright (c) 2021-2024 Dell Inc., or its subsidiaries. All Rights Reserved.

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
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccRedfishStorageDataSource_fetch(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRedfishDataSourceStorageConfig(creds),
			},
			{
				Config: testAccStorageDatasourceWithControllerID(creds),
			},
			{
				Config: testAccStorageDatasourceWithControllerName(creds),
			},
			{
				Config: testAccStorageDatasourceWithBothConfig(creds),
			},
		},
	})
}

// controller_names = ["PERC H730P Mini"]
// controller_ids = ["AHCI.Embedded.2-1"]

func testAccRedfishDataSourceStorageConfig(testingInfo TestingServerCredentials) string {
	return fmt.Sprintf(`
	data "redfish_storage" "storage" {	  
		redfish_server {
		  user         = "%s"
		  password     = "%s"
		  endpoint     = "https://%s"
		  ssl_insecure = true
		}
	  }
		`,
		testingInfo.Username,
		testingInfo.Password,
		testingInfo.Endpoint,
	)
}

func testAccStorageDatasourceWithControllerID(testingInfo TestingServerCredentials) string {
	return fmt.Sprintf(`
	data "redfish_storage" "storage" {	  
		redfish_server {
		  user         = "%s"
		  password     = "%s"
		  endpoint     = "https://%s"
		  ssl_insecure = true
		}
		controller_ids = ["AHCI.Embedded.2-1"]
	  }
		`,
		testingInfo.Username,
		testingInfo.Password,
		testingInfo.Endpoint,
	)
}

func testAccStorageDatasourceWithControllerName(testingInfo TestingServerCredentials) string {
	return fmt.Sprintf(`
	data "redfish_storage" "storage" {	
		redfish_server {
		  user         = "%s"
		  password     = "%s"
		  endpoint     = "https://%s"
		  ssl_insecure = true
		}
		controller_names = ["PERC H730P Mini"]
	  }
		`,
		testingInfo.Username,
		testingInfo.Password,
		testingInfo.Endpoint,
	)
}

func testAccStorageDatasourceWithBothConfig(testingInfo TestingServerCredentials) string {
	return fmt.Sprintf(`
	data "redfish_storage" "storage" {	
		redfish_server {
		  user         = "%s"
		  password     = "%s"
		  endpoint     = "https://%s"
		  ssl_insecure = true
		}
		controller_ids = ["AHCI.Embedded.2-1"]
		controller_names = ["PERC H730P Mini"]
	  }
		`,
		testingInfo.Username,
		testingInfo.Password,
		testingInfo.Endpoint,
	)
}
