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

func TestAccRedfishIDRACAttributesBasic(t *testing.T) {
	version := os.Getenv("TF_TESTING_REDFISH_VERSION")
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			if version == "17" {
				t.Skip("Skipping 17G test")
			}
			testAccPreCheck(t)
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				PreConfig: func() {
					FunctionMocker = mockey.Mock(isServerGenerationSeventeenAndAbove).Return(false, nil).Build()
				},
				Config: testAccRedfishResourceIDracAttributesConfig(
					creds, "ironman"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("redfish_dell_idrac_attributes.idrac", "attributes.Users.3.Enable", "Disabled"),
					resource.TestCheckResourceAttr("redfish_dell_idrac_attributes.idrac", "attributes.Time.1.Timezone", "CST6CDT"),
				),
			},
		},
	})
}

func TestAccRedfishIDRACAttributesBasic17G(t *testing.T) {
	version := os.Getenv("TF_TESTING_REDFISH_VERSION")
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			if version != "17" {
				t.Skip("Skipping 17G test")
			}
			testAccPreCheck(t)
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				PreConfig: func() {
					FunctionMocker = mockey.Mock(isServerGenerationSeventeenAndAbove).Return(true, nil).Build()
				},
				Config: testAccRedfishResourceIDracAttributesConfig(
					creds, "ironman"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("redfish_dell_idrac_attributes.idrac", "attributes.Users.3.Enable", "Disabled"),
					resource.TestCheckResourceAttr("redfish_dell_idrac_attributes.idrac", "attributes.Time.1.Timezone", "CST6CDT"),
				),
			},
		},
	})
}

func TestAccRedfishIDRACAttributesInvalidAttribute(t *testing.T) {
	version := os.Getenv("TF_TESTING_REDFISH_VERSION")
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			if version == "17" {
				t.Skip("Skipping 17G test")
			}
			testAccPreCheck(t)
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				PreConfig: func() {
					FunctionMocker = mockey.Mock(isServerGenerationSeventeenAndAbove).Return(false, nil).Build()
				},
				Config: testAccRedfishResourceIDracAttributesConfigInvalid(
					creds),
				ExpectError: regexp.MustCompile("there was an issue when creating/updating idrac attributes"),
			},
		},
	})
}

func TestAccRedfishIDRACAttributesInvalidAttribute17G(t *testing.T) {
	version := os.Getenv("TF_TESTING_REDFISH_VERSION")
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			if version != "17" {
				t.Skip("Skipping 17G test")
			}
			testAccPreCheck(t)
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				PreConfig: func() {
					FunctionMocker = mockey.Mock(isServerGenerationSeventeenAndAbove).Return(true, nil).Build()
				},
				Config: testAccRedfishResourceIDracAttributesConfigInvalid(
					creds),
				ExpectError: regexp.MustCompile("there was an issue when creating/updating idrac attributes"),
			},
		},
	})
}

func TestAccRedfishIDRACAttributeImport(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `resource "redfish_dell_idrac_attributes" "idrac" {
				}`,
				ResourceName:  "redfish_dell_idrac_attributes.idrac",
				ImportState:   true,
				ImportStateId: "{\"username\":\"" + creds.Username + "\",\"password\":\"" + creds.Password + "\",\"endpoint\":\"" + creds.Endpoint + "\",\"ssl_insecure\":true}",
				ExpectError:   nil,
			},
		},
	})
}

func TestAccRedfishIDRACAttributeImportByFilter(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `resource "redfish_dell_idrac_attributes" "idrac" {
				}`,
				ResourceName:  "redfish_dell_idrac_attributes.idrac",
				ImportState:   true,
				ImportStateId: "{\"attributes\":[\"Users.2.UserName\"],\"username\":\"" + creds.Username + "\",\"password\":\"" + creds.Password + "\",\"endpoint\":\"" + creds.Endpoint + "\",\"ssl_insecure\":true}",
				ExpectError:   nil,
			},
		},
	})
}

func TestAccRedfishIDRACAttributesCreateConfigErr(t *testing.T) {

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				PreConfig: func() {
					FunctionMocker = mockey.Mock(NewConfig).Return(nil, fmt.Errorf("mock error")).Build()
				},
				Config:      testAccRedfishResourceIDracAttributesConfig(creds, "avengers"),
				ExpectError: regexp.MustCompile(`.*mock error*.`),
			},
		},
	})
	if FunctionMocker != nil {
		FunctionMocker.Release()
	}
}

func TestAccRedfishIDRACAttributesGetManagerAttributeRegistryErr(t *testing.T) {

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				PreConfig: func() {
					FunctionMocker = mockey.Mock(getManagerAttributeRegistry).Return(nil, fmt.Errorf("mock error")).Build()
				},
				Config:      testAccRedfishResourceIDracAttributesConfig(creds, "avengers"),
				ExpectError: regexp.MustCompile(`.*mock error*.`),
			},
		},
	})
	if FunctionMocker != nil {
		FunctionMocker.Release()
	}
}

func TestAccRedfishIDRACAttributesSetManagerAttributesRightTypeErr(t *testing.T) {
	var funcMocker1 *mockey.Mocker
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				PreConfig: func() {
					FunctionMocker = mockey.Mock(isServerGenerationSeventeenAndAbove).Return(true, nil).Build()
					funcMocker1 = mockey.Mock(setManagerAttributesRightType).Return(nil, fmt.Errorf("mock error")).Build()
				},
				Config:      testAccRedfishResourceIDracAttributesConfig(creds, "avengers"),
				ExpectError: regexp.MustCompile(`.*mock error*.`),
			},
		},
	})
	if FunctionMocker != nil {
		FunctionMocker.Release()
	}
	if funcMocker1 != nil {
		funcMocker1.Release()
	}
}

func TestAccRedfishIDRACAttributesGetIdracAttributesErr(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				PreConfig: func() {
					FunctionMocker = mockey.Mock(getIdracAttributes).Return(nil, fmt.Errorf("mock error")).Build()
				},
				Config:      testAccRedfishResourceIDracAttributesConfig(creds, "avengers"),
				ExpectError: regexp.MustCompile(`.*mock error*.`),
			},
		},
	})
	if FunctionMocker != nil {
		FunctionMocker.Release()
	}
}

func TestAccRedfishIDRACAttributes17Error(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				PreConfig: func() {
					FunctionMocker = mockey.Mock(isServerGenerationSeventeenAndAbove).Return(true, nil).Build()
				},
				Config:      testAccRedfishResourceIDrac17GAttributesError(creds, "avengers"),
				ExpectError: regexp.MustCompile("Need to use Role attribute for getting and setting the privileges 'Users.x.Role'"),
			},
		},
	})
	if FunctionMocker != nil {
		FunctionMocker.Release()
	}
}

func TestAccRedfishIDRACAttributesBelow17GConfig(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				PreConfig: func() {
					FunctionMocker = mockey.Mock(isServerGenerationSeventeenAndAbove).Return(false, nil).Build()
				},
				Config:      testAccRedfishResourceIDracBelow17GConfigError(creds, "avengers"),
				ExpectError: regexp.MustCompile("Need to use Privilege attribute for getting and setting the privileges 'Users.x.Privilege'"),
			},
		},
	})
	if FunctionMocker != nil {
		FunctionMocker.Release()
	}
}

func TestAccRedfishIDRACAttributes17GServerError(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				PreConfig: func() {
					FunctionMocker = mockey.Mock(isServerGenerationSeventeenAndAbove).Return(nil, fmt.Errorf("mock error")).Build()
				},
				Config:      testAccRedfishResourceIDracBelow17GConfigError(creds, "avengers"),
				ExpectError: regexp.MustCompile("Error retrieving the server generation"),
			},
		},
	})
	if FunctionMocker != nil {
		FunctionMocker.Release()
	}
}

func TestAccRedfishIDRACAttributes17GParam(t *testing.T) {
	version := os.Getenv("TF_TESTING_REDFISH_VERSION")
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			if version != "17" {
				t.Skip("Skipping 17G test")
			}
			testAccPreCheck(t)
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				PreConfig: func() {
					FunctionMocker = mockey.Mock(isServerGenerationSeventeenAndAbove).Return(true, nil).Build()
				},
				Config: testAccRedfishResourceIDracBelow17GConfigError(creds, "avengers"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("redfish_dell_idrac_attributes.idrac", "attributes.Users.3.Role", "ReadOnly"),
				),
			},
		},
	})
	if FunctionMocker != nil {
		FunctionMocker.Release()
	}
}

func TestAccRedfishIDRACAttributesBelow17GParam(t *testing.T) {
	version := os.Getenv("TF_TESTING_REDFISH_VERSION")
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			if version == "17" {
				t.Skip("Skipping 17G test")
			}
			testAccPreCheck(t)
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				PreConfig: func() {
					FunctionMocker = mockey.Mock(isServerGenerationSeventeenAndAbove).Return(false, nil).Build()
				},
				Config: testAccRedfishResourceIDrac17GAttributesError(creds, "avengers"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("redfish_dell_idrac_attributes.idrac", "attributes.Users.3.Privilege", "511"),
				),
			},
		},
	})
	if FunctionMocker != nil {
		FunctionMocker.Release()
	}
}

func testAccRedfishResourceIDracAttributesConfig(testingInfo TestingServerCredentials, username string) string {
	return fmt.Sprintf(`
	resource "redfish_dell_idrac_attributes" "idrac" {
		redfish_server {
		  user         = "%s"
		  password     = "%s"
		  endpoint     = "%s"
		  ssl_insecure = true
		}
	  
		attributes = {
		  "Users.3.Enable"    		  = "Disabled"
		  "Users.3.UserName"  		  = "%s"
		  "Users.3.Password"  		  = "test1234"
		  "Time.1.Timezone"   		  = "CST6CDT"
		  "SysLog.1.PowerLogInterval" = 5
		}
	  }
	  `,
		testingInfo.Username,
		testingInfo.Password,
		testingInfo.Endpoint,
		username,
	)
}

func testAccRedfishResourceIDracAttributesConfigInvalid(testingInfo TestingServerCredentials) string {
	return fmt.Sprintf(`
	resource "redfish_dell_idrac_attributes" "idrac" {
		redfish_server {
		  user         = "%s"
		  password     = "%s"
		  endpoint     = "%s"
		  ssl_insecure = true
		}
	  
		attributes = {
		  "Users.3.Enable"            = "Disabled"
		  "Users.3.UserName"          = "mike"
		  "Users.3.Password"          = "test1234"
		  "Time.1.Timezone"			  = "CST6CDT",
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

func testAccRedfishResourceIDrac17GAttributesError(testingInfo TestingServerCredentials, username string) string {
	return fmt.Sprintf(`
	resource "redfish_dell_idrac_attributes" "idrac" {
		redfish_server {
		  user         = "%s"
		  password     = "%s"
		  endpoint     = "%s"
		  ssl_insecure = true
		}
	  
		attributes = {
		  "Users.3.Enable"    		  = "Disabled"
		  "Users.3.UserName"  		  = "%s"
		  "Users.3.Password"  		  = "test1234"
		  "Users.3.Privilege" 		  	= 511
		  "Time.1.Timezone"   		  = "CST6CDT"
		  "SysLog.1.PowerLogInterval" = 5
		}
	  }
	  `,
		testingInfo.Username,
		testingInfo.Password,
		testingInfo.Endpoint,
		username,
	)
}

func testAccRedfishResourceIDracBelow17GConfigError(testingInfo TestingServerCredentials, username string) string {
	return fmt.Sprintf(`
	resource "redfish_dell_idrac_attributes" "idrac" {
		redfish_server {
		  user         = "%s"
		  password     = "%s"
		  endpoint     = "%s"
		  ssl_insecure = true
		}
	  
		attributes = {
		  "Users.3.Enable"    		  = "Disabled"
		  "Users.3.UserName"  		  = "%s"
		  "Users.3.Password"  		  = "test1234"
		  "Users.3.Role" 		  	= "ReadOnly"
		  "Time.1.Timezone"   		  = "CST6CDT"
		  "SysLog.1.PowerLogInterval" = 5
		}
	  }
	  `,
		testingInfo.Username,
		testingInfo.Password,
		testingInfo.Endpoint,
		username,
	)
}
