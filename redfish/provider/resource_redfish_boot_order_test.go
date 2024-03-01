/*
Copyright (c) 2023-2024 Dell Inc., or its subsidiaries. All Rights Reserved.

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
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

// test redfish Boot Order
func TestAccRedfishBootOrder_basic(t *testing.T) {

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRedfishResourceBootOrder(creds, `["Boot0003","Boot0004","Boot0005"]`),
			},
			{
				ResourceName:  "redfish_boot_order.boot",
				ImportState:   true,
				ImportStateId: "{\"username\":\"" + creds.Username + "\",\"password\":\"" + creds.Password + "\",\"endpoint\":\"https://" + creds.Endpoint + "\",\"ssl_insecure\":true}",
				ExpectError:   nil,
				// ImportStateVerify: true, // state is not verified since there are multiple boot options and import fetches all while using CRUD you can change specific boot options or none
			},
		},
	})
}

func TestAccRedfishBootOrderOptions_basic(t *testing.T) {

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRedfishResourceBootOptions(creds, os.Getenv("TF_TESTING_BOOT_OPTION_REFERENCE"), true),
			},
			{
				ResourceName:  "redfish_boot_order.boot",
				ImportState:   true,
				ImportStateId: "{\"username\":\"" + creds.Username + "\",\"password\":\"" + creds.Password + "\",\"endpoint\":\"https://" + creds.Endpoint + "\",\"ssl_insecure\":true}",
				ExpectError:   nil,
			},
			{
				Config: testAccRedfishResourceBootOptions(creds, os.Getenv("TF_TESTING_BOOT_OPTION_REFERENCE"), false),
			},
		},
	})
}

func testAccRedfishResourceBootOrder(testingInfo TestingServerCredentials, bootOrder string) string {
	return fmt.Sprintf(`

	resource "redfish_boot_order" "boot" {
		redfish_server {
			user = "%s"
			password = "%s"
			endpoint = "https://%s"
			ssl_insecure = true
		}
	   
		reset_type="ForceRestart"
		boot_order=%s
	}	  
	`,
		testingInfo.Username,
		testingInfo.Password,
		testingInfo.Endpoint,
		bootOrder,
	)
}

func testAccRedfishResourceBootOptions(testingInfo TestingServerCredentials, bootOptionReference string, bootOptionEnabled bool) string {
	return fmt.Sprintf(`

	resource "redfish_boot_order" "boot" {
		redfish_server {
			user = "%s"
			password = "%s"
			endpoint = "https://%s"
			ssl_insecure = true
		}
	    reset_timeout=400
		boot_order_job_timeout=4000
		reset_type="ForceRestart"   
		boot_options = [{boot_option_reference="%s", boot_option_enabled=%t}]
	}	  
	`,
		testingInfo.Username,
		testingInfo.Password,
		testingInfo.Endpoint,
		bootOptionReference,
		bootOptionEnabled,
	)
}
