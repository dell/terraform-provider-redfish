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
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

const testAccVMedResName = "redfish_virtual_media.virtual_media"

// getVMedImportConf returns the import configuration for the virtual media
func getVMedImportConf(d *terraform.State, creds TestingServerCredentials) (string, error) {
	id, err := getID(d, testAccVMedResName)
	if err != nil {
		return id, err
	}
	return fmt.Sprintf("{\"id\":\"%s\",\"username\":\"%s\",\"password\":\"%s\",\"endpoint\":\"%s\",\"ssl_insecure\":true}",
		id, creds.Username, creds.Password, creds.Endpoint), nil
}

// Test to create redfish virtual media - Positive
func TestAccRedfishVirtualMedia_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRedfishResourceVirtualMediaConfig(
					creds,
					"virtual_media",
					image64Boot,
					true,
					"HTTP",
					"Stream"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(testAccVMedResName, "image", image64Boot),
					resource.TestCheckResourceAttr(testAccVMedResName, "inserted", "true"),
				),
			},
			// check that import is creating correct state
			{
				ResourceName:      testAccVMedResName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: func(d *terraform.State) (string, error) {
					return getVMedImportConf(d, creds)
				},
				ImportStateVerifyIgnore: []string{"redfish_server.0.redfish_alias"},
			},
			// check that wrong import ID gives error
			{
				ResourceName: testAccVMedResName,
				ImportState:  true,
				ImportStateId: fmt.Sprintf("{\"id\":\"invalid\",\"username\":\"%s\",\"password\":\"%s\",\"endpoint\":\"%s\",\"ssl_insecure\":true}",
					creds.Username, creds.Password, creds.Endpoint),
				ExpectError: regexp.MustCompile("Virtual Media with ID invalid doesn't exist"),
			},
		},
	})
}

// Test to create redfish virtual media with invalid image extension - Negative
func TestAccRedfishVirtualMedia_InvalidImage_Negative(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRedfishResourceVirtualMediaConfig(
					creds,
					"virtual_media",
					"http://linuxlib.us.dell.com/pub/redhat/RHEL8/8.9/latest/BaseOS/x86_64/iso/RHEL-8.9.0-20231023.21-x86_64-boot.is",
					true,
					"HTTP",
					"Stream"),
				ExpectError: regexp.MustCompile("[C|c]ouldn't mount Virtual Media"),
			},
		},
	})
}

// Test to create redfish virtual media with invalid transfer method - Negative
func TestAccRedfishVirtualMedia_InvalidTransferMethod_Negative(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRedfishResourceVirtualMediaConfig(
					creds,
					"virtual_media",
					image64Boot,
					true,
					"HTTP",
					"Upload"),
				ExpectError: regexp.MustCompile("[C|c]ouldn't mount Virtual Media"),
			},
		},
	})
}

// Test to create redfish virtual media with invalid transfer protocol type - Negative
func TestAccRedfishVirtualMedia_InvalidTransferProtocol_Negative(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRedfishResourceVirtualMediaConfig(
					creds,
					"virtual_media",
					image64Boot,
					true,
					"HTTPS",
					"Stream"),
				ExpectError: regexp.MustCompile("[C|c]ouldn't mount Virtual Media"),
			},
		},
	})
}

// Test to create redfish virtual media when no file shares are available to mount - Negative
func TestAccRedfishVirtualMediaNoMediaNegative_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRedfishResourceVirtualMediaConfig(
					creds,
					"virtual_media1",
					image64Boot,
					true,
					"HTTP",
					"Stream") +
					testAccRedfishResourceVirtualMediaConfigDependency(
						creds,
						"virtual_media2",
						"http://linuxlib.us.dell.com/pub/Distros/RedHat/RHEL8/8.9/RHEL-8.9.0-20231030.60-x86_64-dvd1.iso",
						true,
						"HTTP",
						"Stream",
						"redfish_virtual_media.virtual_media1") +
					testAccRedfishResourceVirtualMediaConfigDependency(
						creds,
						"virtual_media3",
						"http://linuxlib.us.dell.com/pub/Distros/RedHat/RHEL8/8.9/RHEL-8.9.0-20231030.60-x86_64-dvd1.iso",
						true,
						"HTTP",
						"Stream",
						"redfish_virtual_media.virtual_media2"),
				ExpectError: regexp.MustCompile("There are no Virtual Medias to mount"),
			},
		},
	})
}

// Test to create redfish virtual media on iDRAC 5.x - Positive
func TestAccRedfishVirtualMediaServer2_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRedfishResourceVirtualMediaConfigServer5x(
					creds,
					"virtual_media",
					imageEfiBoot,
					true,
					"HTTP",
					"Stream"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(testAccVMedResName, "image", imageEfiBoot),
					resource.TestCheckResourceAttr(testAccVMedResName, "inserted", "true"),
				),
			},
		},
	})
}

// Test to create redfish virtual media with invalid transfer protocol type on iDRAC 5.x - Negative
func TestAccRedfishVirtualMediaServer2_InvalidTransferProtocol_Negative(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRedfishResourceVirtualMediaConfigServer5x(
					creds,
					"virtual_media",
					imageEfiBoot,
					true,
					"HTTPS",
					"Stream"),
				ExpectError: regexp.MustCompile("[C|c]ouldn't mount Virtual Media"),
			},
		},
	})
}

// Test to update redfish virtual media with invalid transfer protocol type on iDRAC 5.x - Negative
func TestAccRedfishVirtualMediaServer2Update_InvalidTransferProtocol_Negative(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRedfishResourceVirtualMediaConfigServer5x(
					creds,
					"virtual_media",
					imageEfiBoot,
					true,
					"HTTP",
					"Stream"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(testAccVMedResName, "image", imageEfiBoot),
					resource.TestCheckResourceAttr(testAccVMedResName, "inserted", "true"),
				),
			},
			{
				Config: testAccRedfishResourceVirtualMediaConfigServer5x(
					creds,
					"virtual_media",
					imageEfiBoot,
					true,
					"HTTPS",
					"Stream"),
				ExpectError: regexp.MustCompile("[C|c]ouldn't mount Virtual Media"),
			},
		},
	})
}

// Test to update redfish virtual media - Negative
func TestAccRedfishVirtualMediaUpdate_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRedfishResourceVirtualMediaConfig(
					creds,
					"virtual_media",
					image64Boot,
					true,
					"HTTP",
					"Stream"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(testAccVMedResName, "image", image64Boot),
					resource.TestCheckResourceAttr(testAccVMedResName, "inserted", "true"),
				),
			},
			// check that import is creating correct state
			{
				ResourceName:      testAccVMedResName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: func(d *terraform.State) (string, error) {
					return getVMedImportConf(d, creds)
				},
				ImportStateVerifyIgnore: []string{"redfish_server.0.redfish_alias"},
			},
			{
				Config: testAccRedfishResourceVirtualMediaConfig(
					creds,
					"virtual_media",
					image64Boot,
					false,
					"HTTP",
					"Stream"),
				ExpectError: regexp.MustCompile("Provider produced inconsistent result after apply"),
			},
			// check that import is creating correct state
			{
				ResourceName:      testAccVMedResName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: func(d *terraform.State) (string, error) {
					return getVMedImportConf(d, creds)
				},
				ImportStateVerifyIgnore: []string{"redfish_server.0.redfish_alias"},
			},
		},
	})
}

// Test to update redfish virtual media with invalid image - Negative
func TestAccRedfishVirtualMediaUpdate_InvalidImage_Negative(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRedfishResourceVirtualMediaConfig(
					creds,
					"virtual_media",
					image64Boot,
					true,
					"HTTP",
					"Stream"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(testAccVMedResName, "image", image64Boot),
					resource.TestCheckResourceAttr(testAccVMedResName, "inserted", "true"),
				),
			},
			{
				Config: testAccRedfishResourceVirtualMediaConfig(
					creds,
					"virtual_media",
					"http://linuxlib.us.dell.com/pub/redhat/RHEL8/8.8/BaseOS/x86_64/iso/RHEL-8.8.0-20230411.3-x86_64-boot.is",
					true,
					"HTTP",
					"Stream"),
				ExpectError: regexp.MustCompile("[C|c]ouldn't mount Virtual Media"),
			},
		},
	})
}

// Test to update redfish virtual media with invalid transfer method - Negative
func TestAccRedfishVirtualMediaUpdate_InvalidTransferMethod_Negative(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRedfishResourceVirtualMediaConfig(
					creds,
					"virtual_media",
					image64Boot,
					true,
					"HTTP",
					"Stream"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(testAccVMedResName, "image", image64Boot),
					resource.TestCheckResourceAttr(testAccVMedResName, "inserted", "true"),
				),
			},
			{
				Config: testAccRedfishResourceVirtualMediaConfig(
					creds,
					"virtual_media",
					"http://linuxlib.us.dell.com/pub/redhat/RHEL8/8.8/BaseOS/x86_64/iso/RHEL-8.8.0-20230411.3-x86_64-boot.iso",
					true,
					"HTTP",
					"Upload"),
				ExpectError: regexp.MustCompile("[C|c]ouldn't mount Virtual Media"),
			},
		},
	})
}

func testAccRedfishResourceVirtualMediaConfig(testingInfo TestingServerCredentials,
	resource_name string,
	image string,
	write_protected bool,
	transfer_protocol_type string,
	transfer_method string,
) string {
	return fmt.Sprintf(`
		
		resource "redfish_virtual_media" "%s" {
		
		  redfish_server {
			user = "%s"
			password = "%s"
			endpoint = "%s"
			ssl_insecure = true
		  }
		  image = "%s"
		  write_protected = %t
		  transfer_protocol_type = "%s"
		  transfer_method = "%s"
		}
		`,
		resource_name,
		testingInfo.Username,
		testingInfo.Password,
		testingInfo.Endpoint,
		image,
		write_protected,
		transfer_protocol_type,
		transfer_method,
	)
}

func testAccRedfishResourceVirtualMediaConfigServer5x(testingInfo TestingServerCredentials,
	resource_name string,
	image string,
	write_protected bool,
	transfer_protocol_type string,
	transfer_method string,
) string {
	return fmt.Sprintf(`
		
		resource "redfish_virtual_media" "%s" {
		
		  redfish_server {
			user = "%s"
			password = "%s"
			endpoint = "%s"
			ssl_insecure = true
		  }
		  image = "%s"
		  write_protected = %t
		  transfer_protocol_type = "%s"
		  transfer_method = "%s"
		}
		`,
		resource_name,
		testingInfo.Username,
		testingInfo.Password,
		testingInfo.Endpoint2,
		image,
		write_protected,
		transfer_protocol_type,
		transfer_method,
	)
}

func testAccRedfishResourceVirtualMediaConfigDependency(testingInfo TestingServerCredentials,
	resource_name string,
	image string,
	write_protected bool,
	transfer_protocol_type string,
	transfer_method string,
	depends_on string,
) string {
	return fmt.Sprintf(`
		resource "redfish_virtual_media" "%s" {

		  redfish_server {
			user = "%s"
			password = "%s"
			endpoint = "%s"
			ssl_insecure = true
		  }
		  image = "%s"
		  write_protected = %t
		  transfer_protocol_type = "%s"
		  transfer_method = "%s"

		  depends_on = [
			"%s"
		  ]
		}
		`,
		resource_name,
		testingInfo.Username,
		testingInfo.Password,
		testingInfo.Endpoint,
		image,
		write_protected,
		transfer_protocol_type,
		transfer_method,
		depends_on,
	)
}
