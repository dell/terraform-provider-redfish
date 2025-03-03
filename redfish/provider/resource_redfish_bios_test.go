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
	"regexp"
	"testing"

	"github.com/bytedance/mockey"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// test redfish bios settings
func TestAccRedfishBios_basic(t *testing.T) {
	version := os.Getenv("TF_TESTING_REDFISH_VERSION")
	if version == "17" {
		t.Skip("Skipping Bios Tests for 17G")
	}
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				PreConfig: func() {
					FunctionMocker = mockey.Mock(isServerGenerationSeventeenAndAbove).Return(false, nil).Build()
				},
				Config: testAccRedfishResourceBiosConfigOn(
					creds),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("redfish_bios.bios", "attributes.NumLock", "On"),
				),
			},
			{
				Config: testAccRedfishResourceBiosConfigOff(
					creds),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("redfish_bios.bios", "attributes.NumLock", "Off"),
				),
			},
		},
	})

	if FunctionMocker != nil {
		FunctionMocker.Release()
	}
}

func TestAccRedfishBios_17Gbasic(t *testing.T) {
	version := os.Getenv("TF_TESTING_REDFISH_VERSION")
	if version != "17" {
		t.Skip("Skipping Bios Tests for below 17G")
	}
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				PreConfig: func() {
					FunctionMocker = mockey.Mock(isServerGenerationSeventeenAndAbove).Return(true, nil).Build()
				},
				Config: testAccRedfishResourceBiosConfigOn(
					creds),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("redfish_bios.bios", "attributes.NumLock", "On"),
				),
			},
			{
				Config: testAccRedfishResourceBiosConfigOff17G(
					creds),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("redfish_bios.bios", "attributes.NumLock", "Off"),
				),
			},
			{
				Config: testAccRedfishResourceBiosConfigOff(
					creds),
				ExpectError: regexp.MustCompile("AcPwrRcvryUserDelay Configuration is not supported by 17G device"),
			},
		},
	})
	if FunctionMocker != nil {
		FunctionMocker.Release()
	}
}

func TestAccRedfishBios_InvalidSettings(t *testing.T) {
	version := os.Getenv("TF_TESTING_REDFISH_VERSION")
	if version == "17" {
		t.Skip("Skipping Bios Tests for 17G")
	}
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				PreConfig: func() {
					FunctionMocker = mockey.Mock(isServerGenerationSeventeenAndAbove).Return(false, nil).Build()
				},
				Config: testAccRedfishResourceBiosConfigInvalidSettingsApplyTime(
					creds),
				ExpectError: regexp.MustCompile("Attribute settings_apply_time value must be one of"),
			},
		},
	})
}

func TestAccRedfishBios_InvalidAttributes(t *testing.T) {
	version := os.Getenv("TF_TESTING_REDFISH_VERSION")
	if version == "17" {
		t.Skip("Skipping Bios Tests for 17G")
	}
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				PreConfig: func() {
					FunctionMocker = mockey.Mock(isServerGenerationSeventeenAndAbove).Return(false, nil).Build()
				},
				Config: testAccRedfishResourceBiosConfigInvalidAttributes(
					creds),
				ExpectError: regexp.MustCompile("Attribute settings_apply_time value must be one of"),
			},
		},
	})
	if FunctionMocker != nil {
		FunctionMocker.Release()
	}
}

func TestAccRedfishBios_Mock(t *testing.T) {
	var funcMocker1 *mockey.Mocker
	version := os.Getenv("TF_TESTING_REDFISH_VERSION")
	if version == "17" {
		t.Skip("Skipping Bios Tests for 17G")
	}
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// mock NewConfig error when creating
			{
				PreConfig: func() {
					FunctionMocker = mockey.Mock(NewConfig).Return(nil, fmt.Errorf("mock error")).Build()

				},
				Config:      testAccRedfishResourceBiosConfigOn(creds),
				ExpectError: regexp.MustCompile(`.*mock error*.`),
			},
			// creating
			{
				PreConfig: func() {
					if FunctionMocker != nil {
						FunctionMocker.Release()
					}
					funcMocker1 = mockey.Mock(isServerGenerationSeventeenAndAbove).Return(false, nil).Build()

				},
				Config: testAccRedfishResourceBiosConfigOn(creds),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("redfish_bios.bios", "attributes.NumLock", "On"),
				),
			},
			// mock NewConfig error when updating
			{
				PreConfig: func() {
					FunctionMocker = mockey.Mock(NewConfig).Return(nil, fmt.Errorf("mock error")).Build()
				},
				Config:      testAccRedfishResourceBiosConfigOff(creds),
				ExpectError: regexp.MustCompile(`.*mock error*.`),
			},
			// updating
			{
				PreConfig: func() {
					if FunctionMocker != nil {
						FunctionMocker.Release()
					}
				},
				Config: testAccRedfishResourceBiosConfigOff(
					creds),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("redfish_bios.bios", "attributes.NumLock", "Off"),
				),
			},
		},
	})

	if funcMocker1 != nil {
		funcMocker1.Release()
	}
}

// Test to import bios - positive
func TestAccRedfishBios_Import(t *testing.T) {
	version := os.Getenv("TF_TESTING_REDFISH_VERSION")
	if version == "17" {
		t.Skip("Skipping Bios Tests for 17G")
	}
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				PreConfig: func() {
					FunctionMocker = mockey.Mock(isServerGenerationSeventeenAndAbove).Return(false, nil).Build()
				},
				Config: testAccRedfishResourceBiosConfigOn(
					creds),
				ResourceName:  "redfish_bios.bios",
				ImportState:   true,
				ImportStateId: "{\"username\":\"" + creds.Username + "\",\"password\":\"" + creds.Password + "\",\"endpoint\":\"" + creds.Endpoint + "\",\"ssl_insecure\":true}",
				ExpectError:   nil,
			},
		},
	})
	if FunctionMocker != nil {
		FunctionMocker.Release()
	}
}

func TestAccRedfishBios_ImportSystemID(t *testing.T) {
	version := os.Getenv("TF_TESTING_REDFISH_VERSION")
	if version == "17" {
		t.Skip("Skipping Bios Tests for 17G")
	}
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				PreConfig: func() {
					FunctionMocker = mockey.Mock(isServerGenerationSeventeenAndAbove).Return(false, nil).Build()
				},
				Config: testAccRedfishResourceBiosConfigOn(
					creds),
				ResourceName:  "redfish_bios.bios",
				ImportState:   true,
				ImportStateId: "{\"username\":\"" + creds.Username + "\",\"password\":\"" + creds.Password + "\",\"endpoint\":\"" + creds.Endpoint + "\",\"ssl_insecure\":true,\"system_id\":\"System.Embedded.1\"}",
				ExpectError:   nil,
			},
		},
	})
	if FunctionMocker != nil {
		FunctionMocker.Release()
	}
}

func TestAccRedfishBios_17GInvalidSettings(t *testing.T) {
	version := os.Getenv("TF_TESTING_REDFISH_VERSION")
	if version != "17" {
		t.Skip("Skipping Bios Tests for below 17G")
	}
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				PreConfig: func() {
					FunctionMocker = mockey.Mock(isServerGenerationSeventeenAndAbove).Return(true, nil).Build()
				},
				Config: testAccRedfishResourceBiosConfigInvalidSettingsApplyTime(
					creds),
				ExpectError: regexp.MustCompile("Attribute settings_apply_time value must be one of"),
			},
		},
	})
	if FunctionMocker != nil {
		FunctionMocker.Release()
	}
}

func TestAccRedfishBios_17GInvalidAttributes(t *testing.T) {
	version := os.Getenv("TF_TESTING_REDFISH_VERSION")
	if version != "17" {
		t.Skip("Skipping Bios Tests for below 17G")
	}
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				PreConfig: func() {
					FunctionMocker = mockey.Mock(isServerGenerationSeventeenAndAbove).Return(true, nil).Build()
				},
				Config: testAccRedfishResourceBiosConfigInvalidAttributes(
					creds),
				ExpectError: regexp.MustCompile("Attribute settings_apply_time value must be one of"),
			},
		},
	})
	if FunctionMocker != nil {
		FunctionMocker.Release()
	}
}

// Test to import bios - positive
func TestAccRedfishBios_17GImport(t *testing.T) {
	version := os.Getenv("TF_TESTING_REDFISH_VERSION")
	if version != "17" {
		t.Skip("Skipping Bios Tests for below 17G")
	}
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				PreConfig: func() {
					FunctionMocker = mockey.Mock(isServerGenerationSeventeenAndAbove).Return(true, nil).Build()
				},
				Config: testAccRedfishResourceBiosConfigOn(
					creds),
				ResourceName:  "redfish_bios.bios",
				ImportState:   true,
				ImportStateId: "{\"username\":\"" + creds.Username + "\",\"password\":\"" + creds.Password + "\",\"endpoint\":\"" + creds.Endpoint + "\",\"ssl_insecure\":true}",
				ExpectError:   nil,
			},
		},
	})
	if FunctionMocker != nil {
		FunctionMocker.Release()
	}
}

func TestAccRedfishBios_17GImportSystemID(t *testing.T) {
	version := os.Getenv("TF_TESTING_REDFISH_VERSION")
	if version != "17" {
		t.Skip("Skipping Bios Tests for below 17G")
	}
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				PreConfig: func() {
					FunctionMocker = mockey.Mock(isServerGenerationSeventeenAndAbove).Return(true, nil).Build()
				},
				Config: testAccRedfishResourceBiosConfigOn(
					creds),
				ResourceName:  "redfish_bios.bios",
				ImportState:   true,
				ImportStateId: "{\"username\":\"" + creds.Username + "\",\"password\":\"" + creds.Password + "\",\"endpoint\":\"" + creds.Endpoint + "\",\"ssl_insecure\":true,\"system_id\":\"System.Embedded.1\"}",
				ExpectError:   nil,
			},
		},
	})
	if FunctionMocker != nil {
		FunctionMocker.Release()
	}
}

func testAccRedfishResourceBiosConfigOn(testingInfo TestingServerCredentials) string {
	return fmt.Sprintf(`

		resource "redfish_bios" "bios"  {
		
		  redfish_server {
			user = "%s"
			password = "%s"
			endpoint = "%s"
			ssl_insecure = true
		  }

		  attributes = {
			"NumLock" = "On"
		  }
		  reset_type = "ForceRestart"
		//   system_id = "System.Embedded.1"
		}
		`,
		testingInfo.Username,
		testingInfo.Password,
		testingInfo.Endpoint,
	)
}

func testAccRedfishResourceBiosConfigOff(testingInfo TestingServerCredentials) string {
	return fmt.Sprintf(`

		resource "redfish_bios" "bios"  {
		
		  redfish_server {
			user = "%s"
			password = "%s"
			endpoint = "%s"
			ssl_insecure = true
		  }

		  attributes = {
			"NumLock" = "Off"
			"AcPwrRcvryUserDelay" = 70
		  }
		  reset_type = "ForceRestart"
   		  bios_job_timeout = 1200
		  reset_timeout = 120
		}
		`,
		testingInfo.Username,
		testingInfo.Password,
		testingInfo.Endpoint,
	)
}

func testAccRedfishResourceBiosConfigOff17G(testingInfo TestingServerCredentials) string {
	return fmt.Sprintf(`

		resource "redfish_bios" "bios"  {
		
		  redfish_server {
			user = "%s"
			password = "%s"
			endpoint = "%s"
			ssl_insecure = true
		  }

		  attributes = {
			"NumLock" = "Off"
			#"AcPwrRcvryUserDelay" = 70
		  }
		  reset_type = "ForceRestart"
   		  bios_job_timeout = 1200
		  reset_timeout = 120
		}
		`,
		testingInfo.Username,
		testingInfo.Password,
		testingInfo.Endpoint,
	)
}

func testAccRedfishResourceBiosConfigInvalidSettingsApplyTime(testingInfo TestingServerCredentials) string {
	return fmt.Sprintf(`

		resource "redfish_bios" "bios"  {
		
		  redfish_server {
			user = "%s"
			password = "%s"
			endpoint = "%s"
			ssl_insecure = true
		  }

		  attributes = {
		  }
		  settings_apply_time = "random"
		}
		`,
		testingInfo.Username,
		testingInfo.Password,
		testingInfo.Endpoint,
	)
}

func testAccRedfishResourceBiosConfigInvalidAttributes(testingInfo TestingServerCredentials) string {
	return fmt.Sprintf(`

		resource "redfish_bios" "bios"  {
		
		  redfish_server {
			user = "%s"
			password = "%s"
			endpoint = "%s"
			ssl_insecure = true
		  }

		  attributes = {
		  }
		  settings_apply_time = "ForceRestart"
		}
		`,
		testingInfo.Username,
		testingInfo.Password,
		testingInfo.Endpoint,
	)
}
