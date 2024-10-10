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
				Config: testAccNICDatasourceWithAdapterID(creds),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(nicDatasourceName, "network_interfaces.#", "1"),
					resource.TestMatchResourceAttr(nicDatasourceName, "network_interfaces.0.odata_id", regexp.MustCompile(`.*System.Embedded.1*.`)),
					resource.TestMatchResourceAttr(nicDatasourceName, "network_interfaces.0.network_adapter.odata_id", regexp.MustCompile(`.*NIC.Integrated.1*.`)),
				),
			},
			{
				Config:      testAccNICDatasourceWithInvalidAdapterID(creds),
				ExpectError: regexp.MustCompile(`.*Error one or more of the filtered network adapter ids are not valid*.`),
			},
			{
				Config: testAccNICDatasourceWithConfiguredFilter(creds),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(nicDatasourceName, "network_interfaces.#", "1"),
					resource.TestMatchResourceAttr(nicDatasourceName, "network_interfaces.0.odata_id", regexp.MustCompile(`.*System.Embedded.1*.`)),
					resource.TestMatchResourceAttr(nicDatasourceName, "network_interfaces.0.network_adapter.odata_id", regexp.MustCompile(`.*NIC.Integrated.1*.`)),
					resource.TestCheckResourceAttr(nicDatasourceName, "network_interfaces.0.network_ports.#", "2"),
					resource.TestCheckResourceAttr(nicDatasourceName, "network_interfaces.0.network_device_functions.#", "2"),
				),
			},
			{
				Config:      testAccNICDatasourceWithInvalidPortID(creds),
				ExpectError: regexp.MustCompile(`.*Error one or more of the filtered network port ids are not valid*.`),
			},
			{
				Config: testAccNICDatasourceWithTwoAdapterIDs(creds),
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
		  endpoint     = "https://%s"
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
		  endpoint     = "https://%s"
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
		  endpoint     = "https://%s"
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

func testAccNICDatasourceWithAdapterID(testingInfo TestingServerCredentials) string {
	return fmt.Sprintf(`
	data "redfish_network" "nic" {	
		redfish_server {
		  user         = "%s"
		  password     = "%s"
		  endpoint     = "https://%s"
		  ssl_insecure = true
		}
		nic_filter {
		  systems = [
		  {
			system_id = "System.Embedded.1"
			network_adapters = [
			{
			  network_adapter_id = "NIC.Integrated.1"
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

func testAccNICDatasourceWithInvalidAdapterID(testingInfo TestingServerCredentials) string {
	return fmt.Sprintf(`
	data "redfish_network" "nic" {	
		redfish_server {
		  user         = "%s"
		  password     = "%s"
		  endpoint     = "https://%s"
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

func testAccNICDatasourceWithConfiguredFilter(testingInfo TestingServerCredentials) string {
	return fmt.Sprintf(`
	data "redfish_network" "nic" {	
		redfish_server {
		  user         = "%s"
		  password     = "%s"
		  endpoint     = "https://%s"
		  ssl_insecure = true
		}

		nic_filter {
		  systems = [
		  {
			system_id = "System.Embedded.1"
			network_adapters = [
			{
			  network_adapter_id = "NIC.Integrated.1"
			  network_port_ids = ["NIC.Integrated.1-1", "NIC.Integrated.1-2"]
			  network_device_function_ids = ["NIC.Integrated.1-3-1", "NIC.Integrated.1-2-1"]				  
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

func testAccNICDatasourceWithInvalidPortID(testingInfo TestingServerCredentials) string {
	return fmt.Sprintf(`
	data "redfish_network" "nic" {	
		redfish_server {
		  user         = "%s"
		  password     = "%s"
		  endpoint     = "https://%s"
		  ssl_insecure = true
		}

		nic_filter {
		  systems = [
		  {
			system_id = "System.Embedded.1"
			network_adapters = [
			{
			  network_adapter_id = "NIC.Integrated.1"
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
	)
}

func testAccNICDatasourceWithTwoAdapterIDs(testingInfo TestingServerCredentials) string {
	return fmt.Sprintf(`
	data "redfish_network" "nic" {	
		redfish_server {
		  user         = "%s"
		  password     = "%s"
		  endpoint     = "https://%s"
		  ssl_insecure = true
		}

		nic_filter {
		  systems = [
		  {
			system_id = "System.Embedded.1"
			network_adapters = [
			{
			  network_adapter_id = "NIC.Integrated.1"
			  network_port_ids = ["NIC.Integrated.1-1", "NIC.Integrated.1-2"]
			  network_device_function_ids = ["NIC.Integrated.1-3-1", "NIC.Integrated.1-2-1"]				  
			},
			{
			  network_adapter_id = "FC.Slot.1"
			  network_port_ids = ["FC.Slot.1-2"]
			  network_device_function_ids = ["FC.Slot.1-2"]				  
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
