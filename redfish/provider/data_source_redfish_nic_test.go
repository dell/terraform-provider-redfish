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
	"os"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccRedfishNICDataSource_fetch(t *testing.T) {
	nicDatasourceName := "data.redfish_network.nic"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRedfishDataSourceNICConfig(creds),
			},
			{
				Config: testAccNICDatasourceWithSystemID(creds),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr(nicDatasourceName, "network_interfaces.0.odata_id", regexp.MustCompile(`.*System.Embedded.1*.`)),
				),
			},
			{
				Config:      testAccNICDatasourceWithInvalidSystemID(creds),
				ExpectError: regexp.MustCompile(`.*Error one or more of the filtered system ids are not valid*.`),
			},
			{
				Config: testAccNICDatasourceWithAdapterID(creds, os.Getenv("NETWORK_ADAPTER_ID_1")),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(nicDatasourceName, "network_interfaces.#"),
					resource.TestCheckResourceAttrSet(nicDatasourceName, "network_interfaces.0.odata_id"),
					resource.TestCheckResourceAttrSet(nicDatasourceName, "network_interfaces.0.network_adapter.odata_id"),
				),
			},
			{
				Config:      testAccNICDatasourceWithInvalidAdapterID(creds),
				ExpectError: regexp.MustCompile(`.*Error one or more of the filtered network adapter ids are not valid*.`),
			},
			{
				Config: testAccNICDatasourceWithConfiguredFilter(creds, os.Getenv("NETWORK_ADAPTER_ID_1"), os.Getenv("NETWORK_PORT_IDS_1"), os.Getenv("NETWORK_DEVICE_FUNCTION_IDS_1")),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(nicDatasourceName, "network_interfaces.#"),
					resource.TestCheckResourceAttrSet(nicDatasourceName, "network_interfaces.0.odata_id"),
					resource.TestCheckResourceAttrSet(nicDatasourceName, "network_interfaces.0.network_adapter.odata_id"),
					resource.TestCheckResourceAttrSet(nicDatasourceName, "network_interfaces.0.network_ports.#"),
					resource.TestCheckResourceAttrSet(nicDatasourceName, "network_interfaces.0.network_device_functions.#"),
				),
			},
			{
				Config:      testAccNICDatasourceWithInvalidPortID(creds, os.Getenv("NETWORK_ADAPTER_ID_1")),
				ExpectError: regexp.MustCompile(`.*Error one or more of the filtered network port ids are not valid*.`),
			},
			{
				Config: testAccNICDatasourceWithTwoAdapterIDs(creds, os.Getenv("NETWORK_ADAPTER_ID_1"), os.Getenv("NETWORK_PORT_IDS_1"), os.Getenv("NETWORK_DEVICE_FUNCTION_IDS_1"), os.Getenv("NETWORK_ADAPTER_ID_2"), os.Getenv("NETWORK_PORT_IDS_2"), os.Getenv("NETWORK_DEVICE_FUNCTION_IDS_2")),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(nicDatasourceName, "network_interfaces.#", "2"),
					resource.TestMatchResourceAttr(nicDatasourceName, "network_interfaces.0.odata_id", regexp.MustCompile(`.*System.Embedded.1*.`)),
				),
			},
		},
	})
}

func testAccRedfishDataSourceNICConfig(testingInfo TestingServerCredentials) string {
	return fmt.Sprintf(`
	data "redfish_network" "nic" {	  
		redfish_server {
		  user         = "%s"
		  password     = "%s"
		  endpoint     = "%s"
		  ssl_insecure = true
		}
	}`,
		testingInfo.Username,
		testingInfo.Password,
		testingInfo.Endpoint,
	)
}

func testAccNICDatasourceWithSystemID(testingInfo TestingServerCredentials) string {
	return fmt.Sprintf(`
	data "redfish_network" "nic" {	  
		redfish_server {
		  user         = "%s"
		  password     = "%s"
		  endpoint     = "%s"
		  ssl_insecure = true
		}
		nic_filter {
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

func testAccNICDatasourceWithInvalidSystemID(testingInfo TestingServerCredentials) string {
	return fmt.Sprintf(`
	data "redfish_network" "nic" {	  
		redfish_server {
		  user         = "%s"
		  password     = "%s"
		  endpoint     = "%s"
		  ssl_insecure = true
		}
		nic_filter {
		  systems = [
		  {
			system_id = "InvalidSystemID"
		  }
		  ]
		}
	}`,
		testingInfo.Username,
		testingInfo.Password,
		testingInfo.Endpoint,
	)
}

func testAccNICDatasourceWithAdapterID(testingInfo TestingServerCredentials, network_adapter_id string) string {
	return fmt.Sprintf(`
	data "redfish_network" "nic" {	
		redfish_server {
		  user         = "%s"
		  password     = "%s"
		  endpoint     = "%s"
		  ssl_insecure = true
		}
		nic_filter {
		  systems = [
		  {
			system_id = "System.Embedded.1"
			network_adapters = [
			{
			  network_adapter_id = "%s"
			}
			]
		  }
		  ]
		}
	}`,
		testingInfo.Username,
		testingInfo.Password,
		testingInfo.Endpoint,
		network_adapter_id,
	)
}

func testAccNICDatasourceWithInvalidAdapterID(testingInfo TestingServerCredentials) string {
	return fmt.Sprintf(`
	data "redfish_network" "nic" {	
		redfish_server {
		  user         = "%s"
		  password     = "%s"
		  endpoint     = "%s"
		  ssl_insecure = true
		}
		nic_filter {
		  systems = [
		  {
			system_id = "System.Embedded.1"
			network_adapters = [
			{
			  network_adapter_id = "InvalidAdapterID"
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

func testAccNICDatasourceWithConfiguredFilter(testingInfo TestingServerCredentials, network_adapter_id string, network_port_ids string, network_device_function_ids string) string {
	return fmt.Sprintf(`
	data "redfish_network" "nic" {	
		redfish_server {
		  user         = "%s"
		  password     = "%s"
		  endpoint     = "%s"
		  ssl_insecure = true
		}

		nic_filter {
		  systems = [
		  {
			system_id = "System.Embedded.1"
			network_adapters = [
			{
			  network_adapter_id = "%s"
			  network_port_ids = %s
			  network_device_function_ids = %s
			}
			]
		  }
		  ]
		}
	}`,
		testingInfo.Username,
		testingInfo.Password,
		testingInfo.Endpoint,
		network_adapter_id,
		network_port_ids,
		network_device_function_ids,
	)
}

func testAccNICDatasourceWithInvalidPortID(testingInfo TestingServerCredentials, network_adapter_id string) string {
	return fmt.Sprintf(`
	data "redfish_network" "nic" {	
		redfish_server {
		  user         = "%s"
		  password     = "%s"
		  endpoint     = "%s"
		  ssl_insecure = true
		}

		nic_filter {
		  systems = [
		  {
			system_id = "System.Embedded.1"
			network_adapters = [
			{
			  network_adapter_id = "%s"
			  network_port_ids = ["InvalidPortID"]
			}
			]
		  }
		  ]
		}
	}`,
		testingInfo.Username,
		testingInfo.Password,
		testingInfo.Endpoint,
		network_adapter_id,
	)
}

func testAccNICDatasourceWithTwoAdapterIDs(testingInfo TestingServerCredentials, network_adapter_id_1 string, network_port_ids_1 string, network_device_function_ids_1 string, network_adapter_id_2 string, network_port_ids_2 string, network_device_function_ids_2 string) string {
	return fmt.Sprintf(`
	data "redfish_network" "nic" {	
		redfish_server {
		  user         = "%s"
		  password     = "%s"
		  endpoint     = "%s"
		  ssl_insecure = true
		}

		nic_filter {
		  systems = [
		  {
			system_id = "System.Embedded.1"
			network_adapters = [
			{
			  network_adapter_id = "%s"
			  network_port_id = %s
			  network_device_function_ids = %s			  
			},
			{
			  network_adapter_id = "%s"
			  network_port_ids = %s
			  network_device_function_ids = %s			  
			}
			]
		  }
		  ]
		}
	}`,
		testingInfo.Username,
		testingInfo.Password,
		testingInfo.Endpoint,
		network_adapter_id_1,
		network_port_ids_1,
		network_device_function_ids_1,
		network_adapter_id_2,
		network_port_ids_2,
		network_device_function_ids_2,
	)
}
