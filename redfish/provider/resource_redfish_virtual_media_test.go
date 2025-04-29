/*
Copyright (c) 2023-2025 Dell Inc., or its subsidiaries. All Rights Reserved.

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
	"terraform-provider-redfish/redfish/helper"
	"testing"

	"github.com/bytedance/mockey"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/stmcginnis/gofish"
	"github.com/stmcginnis/gofish/redfish"
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
					virtualMediaTransferProtocolTypeValid,
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

// Test to read redfish virtual media on iDRAC 5.x - Positive with mocky
func TestAccRedfishVirtualMediaServer2_ReadMockErr(t *testing.T) {
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
					virtualMediaTransferProtocolTypeValid,
					"Stream"),
			},
			{
				PreConfig: func() {
					FunctionMocker = mockey.Mock(NewConfig).Return(nil, fmt.Errorf("mock error")).Build()
				},
				Config: testAccRedfishResourceVirtualMediaConfig(
					creds,
					"virtual_media",
					image64Boot,
					true,
					virtualMediaTransferProtocolTypeValid,
					"Stream",
				),
				ExpectError: regexp.MustCompile(`.*mock error*.`),
			},
			{
				PreConfig: func() {
					FunctionMocker.Release()
				},
				Config: testAccRedfishResourceVirtualMediaConfig(
					creds,
					"virtual_media",
					image64Boot,
					true,
					virtualMediaTransferProtocolTypeValid,
					"Stream",
				),
			},
		},
	})
	if FunctionMocker != nil {
		FunctionMocker.Release()
	}
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
					image64BootInvalid,
					true,
					virtualMediaTransferProtocolTypeValid,
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
					virtualMediaTransferProtocolTypeValid,
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
					virtualMediaTransferProtocolTypeInvalid,
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
					virtualMediaTransferProtocolTypeValid,
					"Stream") +
					testAccRedfishResourceVirtualMediaConfigDependency(
						creds,
						"virtual_media2",
						image64Dvd1,
						true,
						virtualMediaTransferProtocolTypeValid,
						"Stream",
						"redfish_virtual_media.virtual_media1") +
					testAccRedfishResourceVirtualMediaConfigDependency(
						creds,
						"virtual_media3",
						image64Dvd1,
						true,
						virtualMediaTransferProtocolTypeValid,
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
					virtualMediaTransferProtocolTypeValid,
					"Stream"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(testAccVMedResName, "image", imageEfiBoot),
					resource.TestCheckResourceAttr(testAccVMedResName, "inserted", "true"),
				),
			},
		},
	})
}

// Test to create redfish virtual media on iDRAC 5.x - Negative with mocky
func TestAccRedfishVirtualMediaServer2_CreateMockErr(t *testing.T) {
	var funcMocker1, funcMocker2, funcMocker3 *mockey.Mocker
	service := &gofish.Service{}
	api := &gofish.APIClient{
		Service: service,
	}
	system := &redfish.ComputerSystem{}
	redfishCollection := make([]*redfish.VirtualMedia, 0)
	env := helper.VirtualMediaEnvironment{
		Manager:    true,
		Collection: redfishCollection,
	}
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				PreConfig: func() {
					FunctionMocker = mockey.Mock(NewConfig).Return(nil, fmt.Errorf("mock error")).Build()
				},
				Config: testAccRedfishResourceVirtualMediaConfigServer5x(
					creds,
					"virtual_media",
					imageEfiBoot,
					true,
					virtualMediaTransferProtocolTypeValid,
					"Stream",
				),
				ExpectError: regexp.MustCompile(`.*mock error*.`),
			},
			{
				PreConfig: func() {
					if FunctionMocker != nil {
						FunctionMocker.Release()
					}
					funcMocker1 = mockey.Mock(NewConfig).Return(api, nil).Build()
					FunctionMocker = mockey.Mock(getSystemResource).Return(nil, fmt.Errorf("mock error")).Build()
				},
				Config: testAccRedfishResourceVirtualMediaConfigServer5x(
					creds,
					"virtual_media",
					imageEfiBoot,
					true,
					virtualMediaTransferProtocolTypeValid,
					"Stream",
				),
				ExpectError: regexp.MustCompile(`.*mock error*.`),
			},
			{
				PreConfig: func() {
					if FunctionMocker != nil {
						FunctionMocker.Release()
					}
					if funcMocker1 != nil {
						funcMocker1.Release()
					}
					funcMocker1 = mockey.Mock(NewConfig).Return(api, nil).Build()
					funcMocker2 = mockey.Mock(getSystemResource).Return(system, nil).Build()
					funcMocker3 = mockey.Mock(helper.GetVMEnv).Return(env, nil).Build()
					FunctionMocker = mockey.Mock(helper.InsertMedia).Return(nil, fmt.Errorf("mock error")).Build()
				},
				Config: testAccRedfishResourceVirtualMediaConfigServer5x(
					creds,
					"virtual_media",
					imageEfiBoot,
					true,
					virtualMediaTransferProtocolTypeValid,
					"Stream",
				),
				ExpectError: regexp.MustCompile(`.*mock error*.`),
			},
		},
	})
	if funcMocker1 != nil {
		funcMocker1.Release()
	}
	if funcMocker2 != nil {
		funcMocker2.Release()
	}
	if funcMocker3 != nil {
		funcMocker3.Release()
	}
	if FunctionMocker != nil {
		FunctionMocker.Release()
	}
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
					virtualMediaTransferProtocolTypeInvalid,
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
					virtualMediaTransferProtocolTypeValid,
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
					virtualMediaTransferProtocolTypeInvalid,
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
					virtualMediaTransferProtocolTypeValid,
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
					virtualMediaTransferProtocolTypeValid,
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

// Test to update redfish virtual media - MockErr
func TestAccRedfishVirtualMediaUpdate_MockErr(t *testing.T) {
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
					virtualMediaTransferProtocolTypeValid,
					"Stream"),
			},
			{
				PreConfig: func() {
					FunctionMocker = mockey.Mock(NewConfig).Return(nil, fmt.Errorf("mock error")).Build()
				},
				Config: testAccRedfishResourceVirtualMediaConfig(
					creds,
					"virtual_media",
					image64Boot,
					false,
					virtualMediaTransferProtocolTypeValid,
					"Stream",
				),
				ExpectError: regexp.MustCompile(`.*mock error*.`),
			},
			{
				PreConfig: func() {
					FunctionMocker.Release()
				},
				Config: testAccRedfishResourceVirtualMediaConfig(
					creds,
					"virtual_media",
					image64Boot,
					true,
					virtualMediaTransferProtocolTypeValid,
					"Stream",
				),
			},
			{
				PreConfig: func() {
					FunctionMocker = mockey.Mock(helper.GetNejectVirtualMedia).Return(nil, fmt.Errorf("mock error")).Build()
				},
				Config: testAccRedfishResourceVirtualMediaConfig(
					creds,
					"virtual_media",
					image64Boot,
					false,
					virtualMediaTransferProtocolTypeValid,
					"Stream",
				),
				ExpectError: regexp.MustCompile(`.*mock error*.`),
			},
			{
				PreConfig: func() {
					FunctionMocker.Release()
				},
				Config: testAccRedfishResourceVirtualMediaConfig(
					creds,
					"virtual_media",
					image64Boot,
					true,
					virtualMediaTransferProtocolTypeValid,
					"Stream",
				),
			},
		},
	})

	if FunctionMocker != nil {
		FunctionMocker.Release()
	}
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
					virtualMediaTransferProtocolTypeValid,
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
					image64BootInvalid,
					true,
					virtualMediaTransferProtocolTypeValid,
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
					virtualMediaTransferProtocolTypeValid,
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
					image64Dvd1,
					true,
					virtualMediaTransferProtocolTypeValid,
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

/*func testAccRedfishResourceVirtualMediaMockData() string {
	return fmt.Sprintf(`
   "@odata.context": "/redfish/v1/$metadata#VirtualMediaCollection.VirtualMediaCollection",
   "@odata.id": "/redfish/v1/Systems/System.Embedded.1/VirtualMedia",
   "@odata.type": "#VirtualMediaCollection.VirtualMediaCollection",
   "Description": "Collection of Virtual Media",
   "Members":[],
   "Members@odata.count": ,
   "Name": "VirtualMedia Collection"
		`,
	)
}*/
