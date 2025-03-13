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
	"time"

	"github.com/bytedance/mockey"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// test redfish Boot Order
func TestAccRedfishBootSourceOverride_basic(t *testing.T) {
	version := os.Getenv("TF_TESTING_REDFISH_VERSION")
	if version == "17" {
		t.Skip("Skipping Boot Source Override Test for 17G")
	}
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				PreConfig: func() {
					FunctionMocker = mockey.Mock(isServerGenerationSeventeenAndAbove).Return(false, nil).Build()
				},
				Config: testAccRedfishResourceBootSourceLegacyconfig(creds),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("redfish_boot_source_override.boot", "boot_source_override_mode", "Legacy"),
				),
			},
		},
	})
	if FunctionMocker != nil {
		FunctionMocker.Release()
	}
}

func TestAccRedfishBootSourceOverride_updated(t *testing.T) {
	version := os.Getenv("TF_TESTING_REDFISH_VERSION")
	if version == "17" {
		t.Skip("Skipping Boot Source Override Test for 17G")
	}
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				PreConfig: func() {
					FunctionMocker = mockey.Mock(isServerGenerationSeventeenAndAbove).Return(false, nil).Build()
				},
				Config: testAccRedfishResourceBootSourceResetType(creds),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("redfish_boot_source_override.boot", "boot_source_override_mode", "UEFI"),
				),
			},
			{
				PreConfig: func() {
					time.Sleep(120 * time.Second)
				},
				Config: testAccRedfishResourceBootSourceUEFIconfig(creds),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("redfish_boot_source_override.boot", "boot_source_override_mode", "UEFI"),
				),
			},
		},
	})
	if FunctionMocker != nil {
		FunctionMocker.Release()
	}
}

// test redfish Boot Source Override for 17G
func TestAccRedfishBootSourceOverride_17Gbasic(t *testing.T) {
	version := os.Getenv("TF_TESTING_REDFISH_VERSION")
	if version != "17" {
		t.Skip("Skipping Boot Source Override Test for below 17G")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				PreConfig: func() {
					FunctionMocker = mockey.Mock(isServerGenerationSeventeenAndAbove).Return(true, nil).Build()
				},
				Config:      testAccRedfishResourceBootSourceLegacyconfig(creds),
				ExpectError: regexp.MustCompile("BootSourceOverrideMode is not supported by 17G server"),
			},
			{
				Config: testAccRedfishResourceBootSource17Gconfig(creds),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("redfish_boot_source_override.boot", "boot_source_override_target", "UefiTarget"),
				),
			},
		},
	})
	if FunctionMocker != nil {
		FunctionMocker.Release()
	}
}

func TestAccRedfishBootSourceOverride_17Gupdated(t *testing.T) {
	version := os.Getenv("TF_TESTING_REDFISH_VERSION")
	if version != "17" {
		t.Skip("Skipping Boot Source Override Test for below 17G")
	}
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				PreConfig: func() {
					FunctionMocker = mockey.Mock(isServerGenerationSeventeenAndAbove).Return(true, nil).Build()
				},
				Config: testAccRedfishResourceBootSource17GUpdateconfig(creds),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("redfish_boot_source_override.boot", "boot_source_override_target", "UefiTarget"),
				),
			},
			{
				PreConfig: func() {
					time.Sleep(120 * time.Second)
				},
				Config: testAccRedfishResourceBootSource17Gconfig(creds),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("redfish_boot_source_override.boot", "boot_source_override_target", "UefiTarget"),
				),
			},
		},
	})
	if FunctionMocker != nil {
		FunctionMocker.Release()
	}
}

// Run below negative scenario for 17G and below 17G devices
func TestAccRedfishBootSourceOverride_MockErr(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				PreConfig: func() {
					FunctionMocker = mockey.Mock(NewConfig).Return(nil, fmt.Errorf("mock error")).Build()
				},
				Config:      testAccRedfishResourceBootSource17GUpdateconfig(creds),
				ExpectError: regexp.MustCompile(`.*mock error*.`),
			},
			{
				PreConfig: func() {
					if FunctionMocker != nil {
						FunctionMocker.Release()
					}
					FunctionMocker = mockey.Mock(getSystemResource).Return(nil, fmt.Errorf("mock error")).Build()
				},
				Config:      testAccRedfishResourceBootSource17GUpdateconfig(creds),
				ExpectError: regexp.MustCompile(`.*mock error*.`),
			},
			{
				PreConfig: func() {
					if FunctionMocker != nil {
						FunctionMocker.Release()
					}
					FunctionMocker = mockey.Mock(isServerGenerationSeventeenAndAbove).Return(nil, fmt.Errorf("mock error")).Build()
				},
				Config:      testAccRedfishResourceBootSource17GUpdateconfig(creds),
				ExpectError: regexp.MustCompile(`.*mock error*.`),
			},
		},
	})
	if FunctionMocker != nil {
		FunctionMocker.Release()
	}
}

func testAccRedfishResourceBootSourceLegacyconfig(testingInfo TestingServerCredentials) string {
	return fmt.Sprintf(`

	resource "redfish_boot_source_override" "boot" {
		redfish_server {
			user = "%s"
			password = "%s"
			endpoint = "%s"
			ssl_insecure = true
		}
	    system_id = "System.Embedded.1"
		boot_source_override_enabled = "Once"
		boot_source_override_target = "Pxe"
		boot_source_override_mode = "Legacy"
		reset_type    = "GracefulRestart"
	}	  
	`,
		testingInfo.Username,
		testingInfo.Password,
		testingInfo.Endpoint,
	)
}

func testAccRedfishResourceBootSource17Gconfig(testingInfo TestingServerCredentials) string {
	return fmt.Sprintf(`

	resource "redfish_boot_source_override" "boot" {
		redfish_server {
			user = "%s"
			password = "%s"
			endpoint = "%s"
			ssl_insecure = true
		}
	    system_id = "System.Embedded.1"
		boot_source_override_enabled = "Once"
		boot_source_override_target = "UefiTarget"
		reset_type    = "GracefulRestart"
	}	  
	`,
		testingInfo.Username,
		testingInfo.Password,
		testingInfo.Endpoint,
	)
}

func testAccRedfishResourceBootSourceUEFIconfig(testingInfo TestingServerCredentials) string {
	return fmt.Sprintf(`

	resource "redfish_boot_source_override" "boot" {
	  
		redfish_server {
			user = "%s"
			password = "%s"
			endpoint = "%s"
			ssl_insecure = true
		}
	   
		boot_source_override_enabled = "Once"
		boot_source_override_target = "UefiTarget"
		boot_source_override_mode = "UEFI"
		reset_type    = "GracefulRestart"
	}	  
	`,
		testingInfo.Username,
		testingInfo.Password,
		testingInfo.Endpoint,
	)
}

func testAccRedfishResourceBootSource17GUpdateconfig(testingInfo TestingServerCredentials) string {
	return fmt.Sprintf(`

	resource "redfish_boot_source_override" "boot" {
	  
		redfish_server {
			user = "%s"
			password = "%s"
			endpoint = "%s"
			ssl_insecure = true
		}
	   
		boot_source_override_enabled = "Once"
		boot_source_override_target = "UefiTarget"
		reset_type    = "ForceRestart"
	}	  
	`,
		testingInfo.Username,
		testingInfo.Password,
		testingInfo.Endpoint,
	)
}

func testAccRedfishResourceBootSourceResetType(testingInfo TestingServerCredentials) string {
	return fmt.Sprintf(`

	resource "redfish_boot_source_override" "boot" {
	  
		redfish_server {
			user = "%s"
			password = "%s"
			endpoint = "%s"
			ssl_insecure = true
		}
	   
		boot_source_override_enabled = "Once"
		boot_source_override_target = "UefiTarget"
		boot_source_override_mode = "UEFI"
		reset_type    = "ForceRestart"
	}	  
	`,
		testingInfo.Username,
		testingInfo.Password,
		testingInfo.Endpoint,
	)
}
