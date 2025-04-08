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
	"encoding/json"
	"fmt"
	"io"
	"regexp"
	"terraform-provider-redfish/gofish/dell"
	"testing"

	"github.com/bytedance/mockey"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccRedfishSystemAttributesBasic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRedfishResourceSystemAttributesConfig(creds),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("redfish_dell_system_attributes.system", "attributes.SupportInfo.1.Outsourced", "Yes"),
					// If `PlatformCapability.1.PSPFCCapable` is Enabled then only will be able to modify `ServerPwr.1.PSPFCEnabled`
					// resource.TestCheckResourceAttr("redfish_dell_system_attributes.system", "attributes.ServerPwr.1.PSPFCEnabled", "Disabled"),
				),
			},
		},
	})
}

func TestAccRedfishSystemAttributesInvalidAttribute(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRedfishResourceSystemConfigInvalid(
					creds),
				ExpectError: regexp.MustCompile("there was an issue when creating/updating System attributes"),
			},
		},
	})
}

func TestAccRedfishSystemManagerInvalidAttribute(t *testing.T) {
	var funcMocker1 *mockey.Mocker
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRedfishResourceSystemAttributesConfig(
					creds),
			},
			{
				PreConfig: func() {
					FunctionMocker = mockey.Mock(assertSystemAttributes).Return(fmt.Errorf("mock error")).Build()
				},
				Config:      testAccRedfishResourceSystemManagerConfigInvalid(creds),
				ExpectError: regexp.MustCompile(`.*mock error*.`),
			},
			{
				PreConfig: func() {
					if FunctionMocker != nil {
						FunctionMocker.Release()
						funcMocker1 = mockey.Mock(assertSystemAttributes).Return(nil).Build()
						FunctionMocker = mockey.Mock(setManagerAttributesRightType).Return(nil, fmt.Errorf("mock error")).Build()
					}
				},
				Config:      testAccRedfishResourceSystemManagerAttributesTypeInvalid(creds),
				ExpectError: regexp.MustCompile(`.*mock error*.`),
			},
		},
	})
	if funcMocker1 != nil {
		funcMocker1.Release()
	}
	if FunctionMocker != nil {
		FunctionMocker.Release()
	}
}

func TestAccRedfishSystemAttributesUpdate(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRedfishResourceSystemAttributesConfig(creds),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("redfish_dell_system_attributes.system", "attributes.SupportInfo.1.Outsourced", "Yes"),
					// If `PlatformCapability.1.PSPFCCapable` is Enabled then only will be able to modify `ServerPwr.1.PSPFCEnabled`
					// resource.TestCheckResourceAttr("redfish_dell_system_attributes.system", "attributes.ServerPwr.1.PSPFCEnabled", "Disabled"),
				),
			},
			{
				Config: testAccRedfishResourceSystemAttributesUpdateConfig(creds),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("redfish_dell_system_attributes.system", "attributes.SupportInfo.1.Outsourced", "No"),
					// If `PlatformCapability.1.PSPFCCapable` is Enabled then only will be able to modify `ServerPwr.1.PSPFCEnabled`
					// resource.TestCheckResourceAttr("redfish_dell_system_attributes.system", "attributes.ServerPwr.1.PSPFCEnabled", "Enabled"),
				),
			},
		},
	})
}

func TestAccRedfishSystemAttributesCreateConfigErr(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				PreConfig: func() {
					FunctionMocker = mockey.Mock(NewConfig).Return(nil, fmt.Errorf("mock error")).Build()
				},
				Config:      testAccRedfishResourceSystemAttributesConfig(creds),
				ExpectError: regexp.MustCompile(`.*mock error*.`),
			},
		},
	})
	if FunctionMocker != nil {
		FunctionMocker.Release()
	}
}

func TestAccRedfishSystemAttributesReadConfigErr(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRedfishResourceSystemAttributesConfig(creds),
			},
			{
				PreConfig: func() {
					FunctionMocker = mockey.Mock(NewConfig).Return(nil, fmt.Errorf("mock error")).Build()
				},
				Config:      testAccRedfishResourceSystemAttributesConfig(creds),
				ExpectError: regexp.MustCompile(`.*mock error*.`),
			},
		},
	})
	if FunctionMocker != nil {
		FunctionMocker.Release()
	}
}

func TestAccRedfishSystemAttributesImport(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `resource "redfish_dell_system_attributes" "system" {
				}`,
				ResourceName:  "redfish_dell_system_attributes.system",
				ImportState:   true,
				ImportStateId: "{\"username\":\"" + creds.Username + "\",\"password\":\"" + creds.Password + "\",\"endpoint\":\"" + creds.Endpoint + "\",\"ssl_insecure\":true}",
				ExpectError:   nil,
			},
		},
	})
}

func TestAccRedfishSystemAttributesImportCheck(t *testing.T) {
	var systemAttributeResourceName = "redfish_dell_system_attributes.system"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRedfishResourceSystemAttributesConfig(creds),
			},
			{
				Config:        testAccRedfishResourceSystemAttributesConfig(creds),
				ResourceName:  systemAttributeResourceName,
				ImportState:   true,
				ExpectError:   nil,
				ImportStateId: "{\"username\":\"" + creds.Username + "\",\"password\":\"" + creds.Password + "\",\"endpoint\":\"" + creds.Endpoint + "\",\"attributes\":[\"SupportInfo.1.Outsourced\"],\"ssl_insecure\":true}",
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("redfish_dell_system_attributes.system", "attributes.SupportInfo.1.Outsourced", "Yes"),
					// If `PlatformCapability.1.PSPFCCapable` is Enabled then only will be able to modify `ServerPwr.1.PSPFCEnabled`
					// resource.TestCheckResourceAttr("redfish_dell_system_attributes.system", "attributes.ServerPwr.1.PSPFCEnabled", "Disabled"),
				),
			},
		},
	})
}

func TestAccRedfishSystemAttributesPSPFCEnabled(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				PreConfig: func() {
					FunctionMocker = mockey.Mock(dell.Manager).Return(nil, fmt.Errorf("Could not get OEM from iDRAC manager")).Build()
				},
				Config:      testAccRedfishResourceSystemAttributesEnabledConfig(creds),
				ExpectError: regexp.MustCompile(`.*Could not get OEM from iDRAC manager*.`),
			},
			{
				PreConfig: func() {
					if FunctionMocker != nil {
						FunctionMocker.Release()
					}
					FunctionMocker = mockey.Mock(isServerGenerationSeventeenAndAbove).Return(nil, fmt.Errorf("Error retrieving the server generation")).Build()
				},
				Config:      testAccRedfishResourceSystemAttributesEnabledConfig(creds),
				ExpectError: regexp.MustCompile(`.*Error retrieving the server generation*.`),
			},
			{
				PreConfig: func() {
					if FunctionMocker != nil {
						FunctionMocker.Release()
					}
					FunctionMocker = mockey.Mock(io.ReadAll).Return(nil, fmt.Errorf("Failed to parse response body")).Build()
				},
				Config:      testAccRedfishResourceSystemAttributesEnabledConfig(creds),
				ExpectError: regexp.MustCompile(`.*Failed to parse response body*.`),
			},
			{
				PreConfig: func() {
					if FunctionMocker != nil {
						FunctionMocker.Release()
					}
					FunctionMocker = mockey.Mock(json.Unmarshal).Return(fmt.Errorf("Cannot convert response to string")).Build()
				},
				Config:      testAccRedfishResourceSystemAttributesEnabledConfig(creds),
				ExpectError: regexp.MustCompile(`.*Cannot convert response to string*.`),
			},
			// this scenario is working for 17G.
			// Commenting as of now smooth running in both AT and UT for both versions.
			// {
			// 	PreConfig: func() {
			// 		if FunctionMocker != nil {
			// 			FunctionMocker.Release()
			// 		}
			// 	},
			// 	Config:      testAccRedfishResourceSystemAttributesEnabledConfig(creds),
			// 	ExpectError: regexp.MustCompile(`.*As PSPFCCapable Attributes disabled, Unable to update the PSPFCEnabled Attribute.*.`),
			// },
			{
				PreConfig: func() {
					if FunctionMocker != nil {
						FunctionMocker.Release()
					}
					FunctionMocker = mockey.Mock(checkManagerAttributes).Return(fmt.Errorf("Manager attribute registry from iDRAC does not match input")).Build()
				},
				Config:      testAccRedfishResourceSystemAttributesConfig(creds),
				ExpectError: regexp.MustCompile(`.*Manager attribute registry from iDRAC does not match input*.`),
			},
			{
				PreConfig: func() {
					if FunctionMocker != nil {
						FunctionMocker.Release()
					}
					FunctionMocker = mockey.Mock(getSystemAttributes).Return(nil, fmt.Errorf("Could not get system attributes")).Build()
				},
				Config:      testAccRedfishResourceSystemAttributesConfig(creds),
				ExpectError: regexp.MustCompile(`.*Could not get system attributes*.`),
			},
		},
	})
	if FunctionMocker != nil {
		FunctionMocker.Release()
	}
}

// If `PlatformCapability.1.PSPFCCapable` is Enabled then only will be able to modify `ServerPwr.1.PSPFCEnabled`
// Hence commenting the `ServerPwr.1.PSPFCEnabled` from the configurations
func testAccRedfishResourceSystemAttributesConfig(testingInfo TestingServerCredentials) string {
	return fmt.Sprintf(`
	resource "redfish_dell_system_attributes" "system" {
		redfish_server {
		  user         = "%s"
		  password     = "%s"
		  endpoint     = "%s"
		  ssl_insecure = true
		}

		attributes = {
			"SupportInfo.1.Outsourced" = "Yes"
		}
	  }
	  `,
		testingInfo.Username,
		testingInfo.Password,
		testingInfo.Endpoint,
	)
}

func testAccRedfishResourceSystemAttributesUpdateConfig(testingInfo TestingServerCredentials) string {
	return fmt.Sprintf(`
	resource "redfish_dell_system_attributes" "system" {
		redfish_server {
		  user         = "%s"
		  password     = "%s"
		  endpoint     = "%s"
		  ssl_insecure = true
		}

		attributes = {
			"SupportInfo.1.Outsourced" = "No"
		}
	  }
	  `,
		testingInfo.Username,
		testingInfo.Password,
		testingInfo.Endpoint,
	)
}

func testAccRedfishResourceSystemConfigInvalid(testingInfo TestingServerCredentials) string {
	return fmt.Sprintf(`
	resource "redfish_dell_system_attributes" "system" {
		redfish_server {
		  user         = "%s"
		  password     = "%s"
		  endpoint     = "%s"
		  ssl_insecure = true
		}

		attributes = {
			"SupportInfo.1.Outsourced" = "Yes",
		  	"SysLog.1.PowerLogInterval" = 5,
		  	"InvalidAttribute" 		  = "invalid",
		}
	  }
	  `,
		testingInfo.Username,
		testingInfo.Password,
		testingInfo.Endpoint,
	)
}

func testAccRedfishResourceSystemManagerConfigInvalid(testingInfo TestingServerCredentials) string {
	return fmt.Sprintf(`
	resource "redfish_dell_system_attributes" "system" {
		redfish_server {
		  user         = "%s"
		  password     = "%s"
		  endpoint     = "%s"
		  ssl_insecure = true
		}

		attributes = {
			"SysLog.1.PowerLogInterval" = 5,
			"InvalidAttribute" 		  = "invalid",
		}
	  }
	  `,
		testingInfo.Username,
		testingInfo.Password,
		testingInfo.Endpoint,
	)
}

func testAccRedfishResourceSystemManagerAttributesTypeInvalid(testingInfo TestingServerCredentials) string {
	return fmt.Sprintf(`
	resource "redfish_dell_system_attributes" "system" {
		redfish_server {
		  user         = "%s"
		  password     = "%s"
		  endpoint     = "%s"
		  ssl_insecure = true
		}

		attributes = {
			"invalid" = 9,
		}
	  }
	  `,
		testingInfo.Username,
		testingInfo.Password,
		testingInfo.Endpoint,
	)
}

func testAccRedfishResourceSystemAttributesEnabledConfig(testingInfo TestingServerCredentials) string {
	return fmt.Sprintf(`
	resource "redfish_dell_system_attributes" "system" {
		redfish_server {
		  user         = "%s"
		  password     = "%s"
		  endpoint     = "%s"
		  ssl_insecure = true
		}

		attributes = {
			"SupportInfo.1.Outsourced" = "Yes",
			"ServerPwr.1.PSPFCEnabled" = "Enabled"
		}
	  }
	  `,
		testingInfo.Username,
		testingInfo.Password,
		testingInfo.Endpoint,
	)
}
