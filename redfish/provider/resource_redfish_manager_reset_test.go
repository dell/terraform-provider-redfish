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
	"testing"

	"github.com/bytedance/mockey"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stmcginnis/gofish"
)

// Test to create manager reset resource with invalid reset type- Negative
func TestAccRedfishManagerReset_Invalid_ResetType_Negative(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccRedfishResourceManagerResetConfig(creds, "iDRAC.Embedded.1", "On"),
				ExpectError: regexp.MustCompile("Invalid Attribute Value Match"),
			},
		},
	})
}

// Test to create manager reset resource with invalid manager id- Negative
func TestAccRedfishManagerReset_Invalid_ManagerID_Negative(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccRedfishResourceManagerResetConfig(creds, "iDRAC.Embedded.0", "GracefulRestart"),
				ExpectError: regexp.MustCompile("invalid Manager ID provided"),
			},
		},
	})
}

// Test to update manager reset resource with invalid maanger id- Negative
func TestAccRedfishManagerReset_Update_Negative(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRedfishResourceManagerResetConfig(creds, "iDRAC.Embedded.1", "GracefulRestart"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("redfish_manager_reset.manager_reset", "id", "iDRAC.Embedded.1"),
				),
			},
			{
				Config:      testAccRedfishResourceManagerResetConfig(creds, "iDRAC.Embedded", "GracefulRestart"),
				ExpectError: regexp.MustCompile("invalid Manager ID provided"),
			},
		},
	})
}

// Test to perform manager reset create with Mock err
func TestAccRedfishManagerReset_CreateMockErr(t *testing.T) {
	var funcMocker *mockey.Mocker
	service := &gofish.Service{}
	api := &gofish.APIClient{
		Service: service,
	}
	//manager := &redfish.Manager{}
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				PreConfig: func() {
					FunctionMocker = mockey.Mock(NewConfig).Return(nil, fmt.Errorf("mock error")).Build()
				},
				Config:      testAccRedfishResourceManagerResetConfig(creds, "iDRAC.Embedded.1", "GracefulRestart"),
				ExpectError: regexp.MustCompile(`.*mock error*.`),
			},
			{
				PreConfig: func() {
					if FunctionMocker != nil {
						FunctionMocker.Release()
					}
					funcMocker = mockey.Mock(NewConfig).Return(api, nil).Build()
					FunctionMocker = mockey.Mock(getManager).Return(nil, fmt.Errorf("mock error")).Build()
				},
				Config:      testAccRedfishResourceManagerResetConfig(creds, "iDRAC.Embedded.1", "GracefulRestart"),
				ExpectError: regexp.MustCompile(`.*mock error*.`),
			},
			/*{
				PreConfig: func() {
					if funcMocker != nil {
						funcMocker.Release()
					}
					if FunctionMocker != nil {
						FunctionMocker.Release()
					}
					funcMocker = mockey.Mock(NewConfig).Return(api, nil).Build()
					funcMocker1 = mockey.Mock(getManager).Return(manager, nil).Build()
					FunctionMocker = mockey.Mock(manager.Reset).Return(fmt.Errorf("mock error")).Build()
				},
				Config:      testAccRedfishResourceManagerResetConfig(creds, "iDRAC.Embedded.1", "GracefulRestart"),
				ExpectError: regexp.MustCompile(`.*mock error*.`),
			},
			{
				PreConfig: func() {
					if funcMocker != nil {
						funcMocker.Release()
					}
					if funcMocker1 != nil {
						funcMocker1.Release()
					}
					if FunctionMocker != nil {
						FunctionMocker.Release()
					}
					funcMocker = mockey.Mock(NewConfig).Return(api, nil).Build()
					funcMocker1 = mockey.Mock(getManager).Return(manager, nil).Build()
					funcMocker2 = mockey.Mock(redfish.ResetType).Return("GracefulRestart").Build()
					funcMocker2 = mockey.Mock(manager.Reset).Return(nil).Build()
					FunctionMocker = mockey.Mock(NewConfig).Return(nil, fmt.Errorf("mock error")).Build()
				},
				Config:      testAccRedfishResourceManagerResetConfig(creds, "iDRAC.Embedded.1", "GracefulRestart"),
				ExpectError: regexp.MustCompile(`.*mock error*.`),
			},
			{
				PreConfig: func() {
					if funcMocker != nil {
						funcMocker.Release()
					}
					if funcMocker1 != nil {
						funcMocker1.Release()
					}
					if funcMocker2 != nil {
						funcMocker2.Release()
					}
					if FunctionMocker != nil {
						FunctionMocker.Release()
					}
					funcMocker = mockey.Mock(NewConfig).Return(api, nil).Build()
					funcMocker1 = mockey.Mock(getManager).Return(manager, nil).Build()
					funcMocker2 = mockey.Mock(manager.Reset).Return(nil).Build()
					funcMocker3 = mockey.Mock(NewConfig).Return(api, nil).Build()
					FunctionMocker = mockey.Mock(getActiveAliasRedfishServer).Return(fmt.Errorf("mock error")).Build()
				},
				Config:      testAccRedfishResourceManagerResetConfig(creds, "iDRAC.Embedded.1", "GracefulRestart"),
				ExpectError: regexp.MustCompile(`.*mock error*.`),
			},*/
		},
	})

	if funcMocker != nil {
		funcMocker.Release()
	}
	if FunctionMocker != nil {
		FunctionMocker.Release()
	}
}

// Test to perform manager reset
func TestAccRedfishManagerReset_Create(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRedfishResourceManagerResetConfig(creds, "iDRAC.Embedded.1", "GracefulRestart"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("redfish_manager_reset.manager_reset", "id", "iDRAC.Embedded.1"),
				),
			},
		},
	})
}

// Test to perform manager reset with Mock err
func TestAccRedfishManagerReset_ReadMockErr(t *testing.T) {
	var funcMocker *mockey.Mocker
	service := &gofish.Service{}
	api := &gofish.APIClient{
		Service: service,
	}
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRedfishResourceManagerResetConfig(creds, "iDRAC.Embedded.1", "GracefulRestart"),
			},
			{
				PreConfig: func() {
					FunctionMocker = mockey.Mock(NewConfig).Return(nil, fmt.Errorf("mock error")).Build()
				},
				Config:      testAccRedfishResourceManagerResetConfig(creds, "iDRAC.Embedded.1", "GracefulRestart"),
				ExpectError: regexp.MustCompile(`.*mock error*.`),
			},
			{
				PreConfig: func() {
					if FunctionMocker != nil {
						FunctionMocker.Release()
					}
					funcMocker = mockey.Mock(NewConfig).Return(api, nil).Build()
					FunctionMocker = mockey.Mock(getManager).Return(nil, fmt.Errorf("mock error")).Build()
				},
				Config:      testAccRedfishResourceManagerResetConfig(creds, "iDRAC.Embedded.1", "GracefulRestart"),
				ExpectError: regexp.MustCompile(`.*mock error*.`),
			},
		},
	})

	if funcMocker != nil {
		funcMocker.Release()
	}
	if FunctionMocker != nil {
		FunctionMocker.Release()
	}
}

func testAccRedfishResourceManagerResetConfig(testingInfo TestingServerCredentials,
	managerID string,
	resetType string,
) string {
	return fmt.Sprintf(`
		
	resource "redfish_manager_reset" "manager_reset" {
		redfish_server {
		  user         = "%s"
		  password     = "%s"
		  endpoint     = "%s"
		  ssl_insecure = true
		}
	  
		id = "%s"
		reset_type = "%s"
	}
		`,
		testingInfo.Username,
		testingInfo.Password,
		testingInfo.Endpoint,
		managerID,
		resetType,
	)
}
