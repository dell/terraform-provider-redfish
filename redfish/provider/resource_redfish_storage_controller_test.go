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

	"github.com/bytedance/mockey"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

var storageControllerParams testingStorageControllerInputs
var storageController17GParams testingStorageControllerInputs

type testingStorageControllerInputs struct {
	TestingServerCredentials
	SystemID     string
	StorageID    string
	ControllerID string
}

func init() {
	storageControllerParams = testingStorageControllerInputs{
		TestingServerCredentials: creds,
		SystemID:                 "System.Embedded.1",
		StorageID:                "RAID.Integrated.1-1",
		ControllerID:             "RAID.Integrated.1-1",
	}

	storageController17GParams = testingStorageControllerInputs{
		TestingServerCredentials: creds,
		SystemID:                 "System.Embedded.1",
		StorageID:                "RAID.SL.1-1",
		ControllerID:             "RAID.SL.1-1",
	}
}

func TestAccRedfishStorageControllerAttributesCreate(t *testing.T) {
	version := os.Getenv("TF_TESTING_REDFISH_VERSION")
	if version == "17" {
		t.Skip("Skipping Tests for 17G")
	}
	storageControllerResourceName := "redfish_storage_controller.test"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// create using basic config
			{
				Config: testAccRedfishResourceStorageControllerBasicConfig(storageControllerParams),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(storageControllerResourceName, "system_id", "System.Embedded.1"),
					resource.TestCheckResourceAttr(storageControllerResourceName, "storage_id", "RAID.Integrated.1-1"),
					resource.TestCheckResourceAttr(storageControllerResourceName, "controller_id", "RAID.Integrated.1-1")),
			},
		},
	})
}

func TestAccRedfishStorageControllerAttributes17GCreate(t *testing.T) {
	version := os.Getenv("TF_TESTING_REDFISH_VERSION")
	if version != "17" {
		t.Skip("Skipping Tests for below 17G")
	}
	storageControllerResourceName := "redfish_storage_controller.test"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// create using basic config
			{
				Config: testAccRedfishResourceStorageControllerBasicConfig(storageController17GParams),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(storageControllerResourceName, "system_id", "System.Embedded.1"),
					resource.TestCheckResourceAttr(storageControllerResourceName, "storage_id", "RAID.SL.1-1"),
					resource.TestCheckResourceAttr(storageControllerResourceName, "controller_id", "RAID.SL.1-1")),
			},
		},
	})
}

func TestAccRedfishStorageControllerAttributesUpdateNonSecurity(t *testing.T) {
	version := os.Getenv("TF_TESTING_REDFISH_VERSION")
	if version == "17" {
		t.Skip("Skipping Tests for 17G")
	}
	storageControllerResourceName := "redfish_storage_controller.test"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// create using basic config
			{
				Config: testAccRedfishResourceStorageControllerBasicConfig(storageControllerParams),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(storageControllerResourceName, "system_id", "System.Embedded.1"),
					resource.TestCheckResourceAttr(storageControllerResourceName, "storage_id", "RAID.Integrated.1-1"),
					resource.TestCheckResourceAttr(storageControllerResourceName, "controller_id", "RAID.Integrated.1-1"),
				),
			},
			// update storage_controller attributes with one set of values
			{
				Config: testAccRedfishResourceStorageControllerFirstAvailableChoiceSelectedConfig(storageControllerParams),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(storageControllerResourceName, "storage_controller.oem.dell.dell_storage_controller.check_consistency_mode", "Normal"),
					resource.TestCheckResourceAttr(storageControllerResourceName, "storage_controller.oem.dell.dell_storage_controller.copyback_mode", "On"),
					resource.TestCheckResourceAttr(storageControllerResourceName, "storage_controller.oem.dell.dell_storage_controller.load_balance_mode", "Automatic"),
					resource.TestCheckResourceAttr(storageControllerResourceName, "storage_controller.oem.dell.dell_storage_controller.enhanced_auto_import_foreign_configuration_mode", "Disabled"),
					resource.TestCheckResourceAttr(storageControllerResourceName, "storage_controller.oem.dell.dell_storage_controller.patrol_read_unconfigured_area_mode", "Disabled"),
					resource.TestCheckResourceAttr(storageControllerResourceName, "storage_controller.oem.dell.dell_storage_controller.patrol_read_mode", "Disabled"),
					resource.TestCheckResourceAttr(storageControllerResourceName, "storage_controller.oem.dell.dell_storage_controller.background_initialization_rate_percent", "32"),
					resource.TestCheckResourceAttr(storageControllerResourceName, "storage_controller.oem.dell.dell_storage_controller.reconstruct_rate_percent", "32"),
					resource.TestCheckResourceAttr(storageControllerResourceName, "storage_controller.controller_rates.consistency_check_rate_percent", "32"),
					resource.TestCheckResourceAttr(storageControllerResourceName, "storage_controller.controller_rates.rebuild_rate_percent", "32"),
				),
			},
			// update storage_controller attributes with another set of values
			{
				Config: testAccRedfishResourceStorageControllerSecondAvailableChoiceSelectedConfig(storageControllerParams),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(storageControllerResourceName, "storage_controller.oem.dell.dell_storage_controller.check_consistency_mode", "StopOnError"),
					resource.TestCheckResourceAttr(storageControllerResourceName, "storage_controller.oem.dell.dell_storage_controller.copyback_mode", "Off"),
					resource.TestCheckResourceAttr(storageControllerResourceName, "storage_controller.oem.dell.dell_storage_controller.load_balance_mode", "Disabled"),
					resource.TestCheckResourceAttr(storageControllerResourceName, "storage_controller.oem.dell.dell_storage_controller.enhanced_auto_import_foreign_configuration_mode", "Enabled"),
					resource.TestCheckResourceAttr(storageControllerResourceName, "storage_controller.oem.dell.dell_storage_controller.patrol_read_unconfigured_area_mode", "Enabled"),
					resource.TestCheckResourceAttr(storageControllerResourceName, "storage_controller.oem.dell.dell_storage_controller.patrol_read_mode", "Automatic"),
					resource.TestCheckResourceAttr(storageControllerResourceName, "storage_controller.oem.dell.dell_storage_controller.background_initialization_rate_percent", "30"),
					resource.TestCheckResourceAttr(storageControllerResourceName, "storage_controller.oem.dell.dell_storage_controller.reconstruct_rate_percent", "30"),
					resource.TestCheckResourceAttr(storageControllerResourceName, "storage_controller.controller_rates.consistency_check_rate_percent", "30"),
					resource.TestCheckResourceAttr(storageControllerResourceName, "storage_controller.controller_rates.rebuild_rate_percent", "30"),
				),
			},
		},
	})
}

func TestAccRedfishStorageControllerAttributes17GUpdateNonSecurity(t *testing.T) {
	version := os.Getenv("TF_TESTING_REDFISH_VERSION")
	if version != "17" {
		t.Skip("Skipping Tests for below 17G")
	}
	storageControllerResourceName := "redfish_storage_controller.test"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// create using basic config
			{
				Config: testAccRedfishResourceStorageControllerBasicConfig(storageController17GParams),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(storageControllerResourceName, "system_id", "System.Embedded.1"),
					resource.TestCheckResourceAttr(storageControllerResourceName, "storage_id", "RAID.SL.1-1"),
					resource.TestCheckResourceAttr(storageControllerResourceName, "controller_id", "RAID.SL.1-1"),
				),
			},
			// update storage_controller attributes with one set of values
			{
				Config: testAccRedfishResourceStorageController17GFirstAvailableChoiceSelectedConfig(storageController17GParams),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(storageControllerResourceName, "storage_controller.oem.dell.dell_storage_controller.background_initialization_rate_percent", "32"),
					resource.TestCheckResourceAttr(storageControllerResourceName, "storage_controller.controller_rates.consistency_check_rate_percent", "32"),
				),
			},
			// update storage_controller attributes with another set of values
			{
				Config: testAccRedfishResourceStorageController17GSecondAvailableChoiceSelectedConfig(storageController17GParams),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(storageControllerResourceName, "storage_controller.oem.dell.dell_storage_controller.background_initialization_rate_percent", "30"),
					resource.TestCheckResourceAttr(storageControllerResourceName, "storage_controller.controller_rates.consistency_check_rate_percent", "30"),
				),
			},
		},
	})
}

func TestAccRedfishStorageControllerAttributesUpdateSecurity(t *testing.T) {
	version := os.Getenv("TF_TESTING_REDFISH_VERSION")
	if version == "17" {
		t.Skip("Skipping Tests for 17G")
	}
	storageControllerResourceName := "redfish_storage_controller.test"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// create using basic config
			{
				Config: testAccRedfishResourceStorageControllerBasicConfig(storageControllerParams),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(storageControllerResourceName, "system_id", "System.Embedded.1"),
					resource.TestCheckResourceAttr(storageControllerResourceName, "storage_id", "RAID.Integrated.1-1"),
					resource.TestCheckResourceAttr(storageControllerResourceName, "controller_id", "RAID.Integrated.1-1"),
				),
			},
			// update security attributes with SetControllerKey action
			{
				Config: testAccRedfishResourceStorageControllerSecuritySetControllerKeyConfig(storageControllerParams),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(storageControllerResourceName, "security.action", "SetControllerKey"),
					resource.TestCheckResourceAttr(storageControllerResourceName, "security.key_id", "testkey1"),
					resource.TestCheckResourceAttr(storageControllerResourceName, "security.key", "Test123##"),
				),
			},
			// update security attributes with ReKey action
			{
				Config: testAccRedfishResourceStorageControllerSecurityReKeyConfig(storageControllerParams),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(storageControllerResourceName, "security.action", "ReKey"),
					resource.TestCheckResourceAttr(storageControllerResourceName, "security.key_id", "testkey2"),
					resource.TestCheckResourceAttr(storageControllerResourceName, "security.key", "Test123###"),
					resource.TestCheckResourceAttr(storageControllerResourceName, "security.old_key", "Test123##"),
					resource.TestCheckResourceAttr(storageControllerResourceName, "security.mode", "LKM"),
				),
			},
			// update security attributes with RemoveControllerKey action
			{
				Config: testAccRedfishResourceStorageControllerSecurityRemoveControllerKeyConfig(storageControllerParams),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(storageControllerResourceName, "security.action", "RemoveControllerKey"),
					resource.TestCheckResourceAttr(storageControllerResourceName, "security.key_id", ""),
				),
			},
		},
	})
}

func TestAccRedfishStorageControllerAttributes17GUpdateSecurity(t *testing.T) {
	version := os.Getenv("TF_TESTING_REDFISH_VERSION")
	if version != "17" {
		t.Skip("Skipping Tests for below 17G")
	}
	storageControllerResourceName := "redfish_storage_controller.test"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// create using basic config
			{
				Config: testAccRedfishResourceStorageControllerBasicConfig(storageController17GParams),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(storageControllerResourceName, "system_id", "System.Embedded.1"),
					resource.TestCheckResourceAttr(storageControllerResourceName, "storage_id", "RAID.SL.1-1"),
					resource.TestCheckResourceAttr(storageControllerResourceName, "controller_id", "RAID.SL.1-1"),
				),
			},
			// update security attributes with EnableSecurity action
			{
				Config: testAccRedfishResourceStorageController17GEnableSecurityConfig(storageController17GParams),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(storageControllerResourceName, "security.action", "EnableSecurity"),
					resource.TestCheckResourceAttr(storageControllerResourceName, "security.mode", "Enabled"),
				),
			},
			// update security attributes with DisableSecurity action
			{
				Config: testAccRedfishResourceStorageController17GDisableSecurityConfig(storageController17GParams),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(storageControllerResourceName, "security.action", "DisableSecurity"),
					resource.TestCheckResourceAttr(storageControllerResourceName, "security.mode", "Disabled"),
				),
			},
		},
	})
}

func TestAccRedfishStorageControllerAttributesError(t *testing.T) {
	version := os.Getenv("TF_TESTING_REDFISH_VERSION")
	if version == "17" {
		t.Skip("Skipping Tests for 17G")
	}
	storageControllerResourceName := "redfish_storage_controller.test"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// create using basic config
			{
				Config: testAccRedfishResourceStorageControllerBasicConfig(storageControllerParams),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(storageControllerResourceName, "system_id", "System.Embedded.1"),
					resource.TestCheckResourceAttr(storageControllerResourceName, "storage_id", "RAID.Integrated.1-1"),
					resource.TestCheckResourceAttr(storageControllerResourceName, "controller_id", "RAID.Integrated.1-1"),
				),
			},
			// error scenario when updating using a different system id
			{
				Config:      testAccRedfishResourceStorageControllerDifferentSystemIDConfig(storageControllerParams),
				ExpectError: regexp.MustCompile("Error when updating with invalid input"),
			},
			// error scenario when updating using a different storage id
			{
				Config:      testAccRedfishResourceStorageControllerDifferentStorageIDConfig(storageControllerParams),
				ExpectError: regexp.MustCompile("Error when updating with invalid input"),
			},
			// error scenario when updating using a different controller id
			{
				Config:      testAccRedfishResourceStorageControllerDifferentControllerIDConfig(storageControllerParams),
				ExpectError: regexp.MustCompile("Error when updating with invalid input"),
			},
			// error scenario when updating controller mode and some other storage controller attribute
			{
				Config:      testAccRedfishResourceStorageControllerControllerModeAndOtherAttributeUpdateConfig(storageControllerParams),
				ExpectError: regexp.MustCompile("While updating `controller_mode`, no other property should be changed."),
			},
			// error scenario when updating controller mode and security attribute
			{
				Config:      testAccRedfishResourceStorageControllerControllerModeAndSecurityUpdateConfig(storageControllerParams),
				ExpectError: regexp.MustCompile("While updating `controller_mode`, no other property should be changed."),
			},
			// error scenario when updating controller mode to HBA and having enhanced_auto_import_foreign_configuration_mode as Enabled
			{
				Config: testAccRedfishResourceStorageControllerControllerModeAndEnhancedAutoImportForeignConfigurationModeConfig(storageControllerParams),
				ExpectError: regexp.MustCompile("Either with `controller_mode` attribute set to `RAID`, set `enhanced_auto_import_foreign_configuration_mode` attribute to `Disabled` first " +
					"or now that the `controller_mode` attribute is set to `HBA`, ensure `enhanced_auto_import_foreign_configuration_mode` attribute is commented."),
			},
			// error scenario when updating controller mode without an on reset type of apply time
			{
				Config:      testAccRedfishResourceStorageControllerControllerModeWithoutOnResetApplyTimeConfig(storageControllerParams),
				ExpectError: regexp.MustCompile("While updating `controller_mode`, the `apply_time` should be `OnReset`."),
			},
			// error scenario when updating security and some other storage controller attribute
			{
				Config:      testAccRedfishResourceStorageControllerSecurityAndOtherAttributeUpdateConfig(storageControllerParams),
				ExpectError: regexp.MustCompile("Attributes of both `security` and `storage_controller` were changed."),
			},
			// error scenario when updating security with maintenance type of apply time
			{
				Config:      testAccRedfishResourceStorageControllerSecurityWithMaintenanceTypeApplyTimeConfig(storageControllerParams),
				ExpectError: regexp.MustCompile("While updating `security` attributes, the `apply_time` should be `Immediate` or `OnReset`."),
			},
			// error scenario when updating security without specifying the action
			{
				Config:      testAccRedfishResourceStorageControllerSecurityWithoutActionConfig(storageControllerParams),
				ExpectError: regexp.MustCompile("Security updates will not be applied since the `action` is not specified."),
			},
			// error scenario when updating security with an incorrect config for SetControllerKey action
			{
				Config:      testAccRedfishResourceStorageControllerSecuritySetControllerKeyIncorrectConfig(storageControllerParams),
				ExpectError: regexp.MustCompile("With `action` set to `SetControllerKey`, the `key` needs to be set."),
			},
			// error scenario when updating security with an incorrect config for ReKey action
			{
				Config:      testAccRedfishResourceStorageControllerSecurityReKeyIncorrectConfig(storageControllerParams),
				ExpectError: regexp.MustCompile("With `action` set to `ReKey`, the `old_key` needs to be set."),
			},
			// error scenario when updating security with an incorrect config for RemoveControllerKey action
			{
				Config:      testAccRedfishResourceStorageControllerSecurityRemoveControllerKeyIncorrectConfig(storageControllerParams),
				ExpectError: regexp.MustCompile("With `action` set to `RemoveControllerKey`, the `key_id` needs to be commented."),
			},
			// error scenario when performing ReKey when key is not present.
			{
				Config:      testAccRedfishResourceStorageControllerSecurityReKeyConfig(storageControllerParams),
				ExpectError: regexp.MustCompile("Post request to IDRAC failed"),
			},
			// error scenario when performing RemoveControllerKey when key is not present.
			{
				Config:      testAccRedfishResourceStorageControllerSecurityRemoveControllerKeyConfig(storageControllerParams),
				ExpectError: regexp.MustCompile("Post request to IDRAC failed"),
			},
			// setting the key using SetControllerKey
			{
				Config: testAccRedfishResourceStorageControllerSecuritySetControllerKeyConfig(storageControllerParams),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(storageControllerResourceName, "security.action", "SetControllerKey"),
					resource.TestCheckResourceAttr(storageControllerResourceName, "security.key_id", "testkey1"),
					resource.TestCheckResourceAttr(storageControllerResourceName, "security.key", "Test123##"),
				),
			},
			// update security attributes with ReKey action
			{
				Config: testAccRedfishResourceStorageControllerSecurityReKeyConfig(storageControllerParams),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(storageControllerResourceName, "security.action", "ReKey"),
					resource.TestCheckResourceAttr(storageControllerResourceName, "security.key_id", "testkey2"),
					resource.TestCheckResourceAttr(storageControllerResourceName, "security.key", "Test123###"),
					resource.TestCheckResourceAttr(storageControllerResourceName, "security.old_key", "Test123##"),
					resource.TestCheckResourceAttr(storageControllerResourceName, "security.mode", "LKM"),
				),
			},
			// error scenario when performing SetControllerKey when key is present.
			{
				Config:      testAccRedfishResourceStorageControllerSecuritySetControllerKeyConfig(storageControllerParams),
				ExpectError: regexp.MustCompile("Post request to IDRAC failed"),
			},
			// removing the key using RemoveControllerKey
			{
				Config: testAccRedfishResourceStorageControllerSecurityRemoveControllerKeyConfig(storageControllerParams),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(storageControllerResourceName, "security.action", "RemoveControllerKey"),
					resource.TestCheckResourceAttr(storageControllerResourceName, "security.key_id", ""),
				),
			},
		},
	})
}

func TestAccRedfishStorageControllerAttributes17GError(t *testing.T) {
	version := os.Getenv("TF_TESTING_REDFISH_VERSION")
	if version != "17" {
		t.Skip("Skipping Tests for below 17G")
	}
	storageControllerResourceName := "redfish_storage_controller.test"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// create using basic config
			{
				Config: testAccRedfishResourceStorageControllerBasicConfig(storageController17GParams),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(storageControllerResourceName, "system_id", "System.Embedded.1"),
					resource.TestCheckResourceAttr(storageControllerResourceName, "storage_id", "RAID.SL.1-1"),
					resource.TestCheckResourceAttr(storageControllerResourceName, "controller_id", "RAID.SL.1-1"),
				),
			},
			// error scenario when updating using a different system id
			{
				Config:      testAccRedfishResourceStorageControllerDifferentSystemIDConfig(storageController17GParams),
				ExpectError: regexp.MustCompile("Error when updating with invalid input"),
			},
			// error scenario when updating using a different storage id
			{
				Config:      testAccRedfishResourceStorageControllerDifferentStorageIDConfig(storageController17GParams),
				ExpectError: regexp.MustCompile("Error when updating with invalid input"),
			},
			// error scenario when updating using a different controller id
			{
				Config:      testAccRedfishResourceStorageControllerDifferentControllerIDConfig(storageController17GParams),
				ExpectError: regexp.MustCompile("Error when updating with invalid input"),
			},
			// error scenario when updating security and some other storage controller attribute
			{
				Config:      testAccRedfishResourceStorageController17GSecurityAndOtherAttributeUpdateConfig(storageController17GParams),
				ExpectError: regexp.MustCompile("Attributes of both `security` and `storage_controller` were changed."),
			},
			// error scenario when updating security with maintenance type of apply time
			{
				Config:      testAccRedfishResourceStorageController17GSecurityWithMaintenanceTypeApplyTimeConfig(storageController17GParams),
				ExpectError: regexp.MustCompile("In 17G and above, the `apply_time` values `AtMaintenanceWindowStart` and `InMaintenanceWindowOnReset` are not valid"),
			},
			// error scenario when updating storage controller with maintenance type of apply time
			{
				Config:      testAccRedfishResourceStorageController17GStorageControllerWithMaintenanceTypeApplyTimeConfig(storageController17GParams),
				ExpectError: regexp.MustCompile("In 17G and above, the `apply_time` values `AtMaintenanceWindowStart` and `InMaintenanceWindowOnReset` are not valid"),
			},
			// error scenario when performing DisableSecurity action when security is already disabled.
			{
				Config:      testAccRedfishResourceStorageController17GDisableSecurityConfig(storageController17GParams),
				ExpectError: regexp.MustCompile("Post request to IDRAC failed"),
			},
		},
	})
}

func TestAccRedfishStorageControllerAttributesImport(t *testing.T) {
	version := os.Getenv("TF_TESTING_REDFISH_VERSION")
	if version == "17" {
		t.Skip("Skipping Tests for 17G")
	}
	storageControllerResourceName := "redfish_storage_controller.test"
	importReqID := fmt.Sprintf("{\"system_id\":\"%s\",\"storage_id\":\"%s\",\"controller_id\":\"%s\",\"username\":\"%s\",\"password\":\"%s\",\"endpoint\":\"%s\",\"ssl_insecure\":true}",
		storageControllerParams.SystemID, storageControllerParams.StorageID, storageControllerParams.ControllerID, storageControllerParams.Username, storageControllerParams.Password, storageControllerParams.Endpoint)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `resource "redfish_storage_controller" "test" {
				}`,
				ResourceName:  storageControllerResourceName,
				ImportState:   true,
				ImportStateId: importReqID,
				ExpectError:   nil,
			},
		},
	})
}

func TestAccRedfishStorageControllerAttributes17GImport(t *testing.T) {
	version := os.Getenv("TF_TESTING_REDFISH_VERSION")
	if version != "17" {
		t.Skip("Skipping Tests for below 17G")
	}
	storageControllerResourceName := "redfish_storage_controller.test"
	importReqID := fmt.Sprintf("{\"system_id\":\"%s\",\"storage_id\":\"%s\",\"controller_id\":\"%s\",\"username\":\"%s\",\"password\":\"%s\",\"endpoint\":\"%s\",\"ssl_insecure\":true}",
		storageController17GParams.SystemID, storageController17GParams.StorageID, storageController17GParams.ControllerID, storageController17GParams.Username, storageController17GParams.Password, storageController17GParams.Endpoint)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `resource "redfish_storage_controller" "test" {
				}`,
				ResourceName:  storageControllerResourceName,
				ImportState:   true,
				ImportStateId: importReqID,
				ExpectError:   nil,
			},
		},
	})
}

func TestAccRedfishStorageControllerAttributesMock(t *testing.T) {
	version := os.Getenv("TF_TESTING_REDFISH_VERSION")
	if version == "17" {
		t.Skip("Skipping Tests for 17G")
	}
	storageControllerResourceName := "redfish_storage_controller.test"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// mock NewConfig error when creating
			{
				PreConfig: func() {
					FunctionMocker = mockey.Mock(NewConfig).Return(nil, fmt.Errorf("mock error")).Build()
				},
				Config:      testAccRedfishResourceStorageControllerBasicConfig(storageControllerParams),
				ExpectError: regexp.MustCompile(`.*mock error*.`),
			},
			// creating
			{
				PreConfig: func() {
					if FunctionMocker != nil {
						FunctionMocker.Release()
					}
				},
				Config: testAccRedfishResourceStorageControllerBasicConfig(storageControllerParams),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(storageControllerResourceName, "system_id", "System.Embedded.1"),
					resource.TestCheckResourceAttr(storageControllerResourceName, "storage_id", "RAID.Integrated.1-1"),
					resource.TestCheckResourceAttr(storageControllerResourceName, "controller_id", "RAID.Integrated.1-1")),
			},
			// mock NewConfig error when updating
			{
				PreConfig: func() {
					FunctionMocker = mockey.Mock(NewConfig).Return(nil, fmt.Errorf("mock error")).Build()
				},
				Config:      testAccRedfishResourceStorageControllerFirstAvailableChoiceSelectedConfig(storageControllerParams),
				ExpectError: regexp.MustCompile(`.*mock error*.`),
			},
			// updating
			{
				PreConfig: func() {
					if FunctionMocker != nil {
						FunctionMocker.Release()
					}
				},
				Config: testAccRedfishResourceStorageControllerFirstAvailableChoiceSelectedConfig(storageControllerParams),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(storageControllerResourceName, "storage_controller.oem.dell.dell_storage_controller.check_consistency_mode", "Normal"),
					resource.TestCheckResourceAttr(storageControllerResourceName, "storage_controller.oem.dell.dell_storage_controller.copyback_mode", "On"),
					resource.TestCheckResourceAttr(storageControllerResourceName, "storage_controller.oem.dell.dell_storage_controller.load_balance_mode", "Automatic"),
					resource.TestCheckResourceAttr(storageControllerResourceName, "storage_controller.oem.dell.dell_storage_controller.enhanced_auto_import_foreign_configuration_mode", "Disabled"),
					resource.TestCheckResourceAttr(storageControllerResourceName, "storage_controller.oem.dell.dell_storage_controller.patrol_read_unconfigured_area_mode", "Disabled"),
					resource.TestCheckResourceAttr(storageControllerResourceName, "storage_controller.oem.dell.dell_storage_controller.patrol_read_mode", "Disabled"),
					resource.TestCheckResourceAttr(storageControllerResourceName, "storage_controller.oem.dell.dell_storage_controller.background_initialization_rate_percent", "32"),
					resource.TestCheckResourceAttr(storageControllerResourceName, "storage_controller.oem.dell.dell_storage_controller.reconstruct_rate_percent", "32"),
					resource.TestCheckResourceAttr(storageControllerResourceName, "storage_controller.controller_rates.consistency_check_rate_percent", "32"),
					resource.TestCheckResourceAttr(storageControllerResourceName, "storage_controller.controller_rates.rebuild_rate_percent", "32"),
				),
			},
			// updating
			{
				Config: testAccRedfishResourceStorageControllerSecondAvailableChoiceSelectedConfig(storageControllerParams),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(storageControllerResourceName, "storage_controller.oem.dell.dell_storage_controller.check_consistency_mode", "StopOnError"),
					resource.TestCheckResourceAttr(storageControllerResourceName, "storage_controller.oem.dell.dell_storage_controller.copyback_mode", "Off"),
					resource.TestCheckResourceAttr(storageControllerResourceName, "storage_controller.oem.dell.dell_storage_controller.load_balance_mode", "Disabled"),
					resource.TestCheckResourceAttr(storageControllerResourceName, "storage_controller.oem.dell.dell_storage_controller.enhanced_auto_import_foreign_configuration_mode", "Enabled"),
					resource.TestCheckResourceAttr(storageControllerResourceName, "storage_controller.oem.dell.dell_storage_controller.patrol_read_unconfigured_area_mode", "Enabled"),
					resource.TestCheckResourceAttr(storageControllerResourceName, "storage_controller.oem.dell.dell_storage_controller.patrol_read_mode", "Automatic"),
					resource.TestCheckResourceAttr(storageControllerResourceName, "storage_controller.oem.dell.dell_storage_controller.background_initialization_rate_percent", "30"),
					resource.TestCheckResourceAttr(storageControllerResourceName, "storage_controller.oem.dell.dell_storage_controller.reconstruct_rate_percent", "30"),
					resource.TestCheckResourceAttr(storageControllerResourceName, "storage_controller.controller_rates.consistency_check_rate_percent", "30"),
					resource.TestCheckResourceAttr(storageControllerResourceName, "storage_controller.controller_rates.rebuild_rate_percent", "30"),
				),
			},
		},
	})
}

func TestAccRedfishStorageControllerAttributes17GMock(t *testing.T) {
	version := os.Getenv("TF_TESTING_REDFISH_VERSION")
	if version != "17" {
		t.Skip("Skipping Tests for below 17G")
	}
	storageControllerResourceName := "redfish_storage_controller.test"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// mock NewConfig error when creating
			{
				PreConfig: func() {
					FunctionMocker = mockey.Mock(NewConfig).Return(nil, fmt.Errorf("mock error")).Build()
				},
				Config:      testAccRedfishResourceStorageControllerBasicConfig(storageController17GParams),
				ExpectError: regexp.MustCompile(`.*mock error*.`),
			},
			// creating
			{
				PreConfig: func() {
					if FunctionMocker != nil {
						FunctionMocker.Release()
					}
				},
				Config: testAccRedfishResourceStorageControllerBasicConfig(storageController17GParams),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(storageControllerResourceName, "system_id", "System.Embedded.1"),
					resource.TestCheckResourceAttr(storageControllerResourceName, "storage_id", "RAID.SL.1-1"),
					resource.TestCheckResourceAttr(storageControllerResourceName, "controller_id", "RAID.SL.1-1")),
			},
			// mock NewConfig error when updating
			{
				PreConfig: func() {
					FunctionMocker = mockey.Mock(NewConfig).Return(nil, fmt.Errorf("mock error")).Build()
				},
				Config:      testAccRedfishResourceStorageController17GFirstAvailableChoiceSelectedConfig(storageController17GParams),
				ExpectError: regexp.MustCompile(`.*mock error*.`),
			},
			// updating
			{
				PreConfig: func() {
					if FunctionMocker != nil {
						FunctionMocker.Release()
					}
				},
				Config: testAccRedfishResourceStorageController17GFirstAvailableChoiceSelectedConfig(storageController17GParams),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(storageControllerResourceName, "storage_controller.oem.dell.dell_storage_controller.background_initialization_rate_percent", "32"),
					resource.TestCheckResourceAttr(storageControllerResourceName, "storage_controller.controller_rates.consistency_check_rate_percent", "32"),
				),
			},
			// updating
			{
				Config: testAccRedfishResourceStorageController17GSecondAvailableChoiceSelectedConfig(storageController17GParams),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(storageControllerResourceName, "storage_controller.oem.dell.dell_storage_controller.background_initialization_rate_percent", "30"),
					resource.TestCheckResourceAttr(storageControllerResourceName, "storage_controller.controller_rates.consistency_check_rate_percent", "30"),
				),
			},
		},
	})
}

func testAccRedfishResourceStorageControllerBasicConfig(testingInfo testingStorageControllerInputs) string {
	return fmt.Sprintf(`
	resource "redfish_storage_controller" "test" {
		redfish_server {
			user         = "%s"
			password     = "%s"
			endpoint     = "%s"
			ssl_insecure = true
		}
		system_id = "%s"
		storage_id = "%s"
		controller_id = "%s"
		apply_time = "Immediate"
		job_timeout = 1200
	}
		`,
		testingInfo.Username,
		testingInfo.Password,
		testingInfo.Endpoint,
		testingInfo.SystemID,
		testingInfo.StorageID,
		testingInfo.ControllerID,
	)
}

func testAccRedfishResourceStorageControllerFirstAvailableChoiceSelectedConfig(testingInfo testingStorageControllerInputs) string {
	return fmt.Sprintf(`
	resource "redfish_storage_controller" "test" {
		redfish_server {
			user         = "%s"
			password     = "%s"
			endpoint     = "%s"
			ssl_insecure = true
		}
		system_id = "%s"
		storage_id = "%s"
		controller_id = "%s"
		apply_time = "Immediate"
		job_timeout = 1200
		storage_controller = {
			oem = {
				dell = {
					dell_storage_controller = {
						check_consistency_mode = "Normal"
						copyback_mode = "On"
						load_balance_mode = "Automatic"
						enhanced_auto_import_foreign_configuration_mode = "Disabled"
						patrol_read_unconfigured_area_mode = "Disabled"
						patrol_read_mode = "Disabled"
						background_initialization_rate_percent = 32
						reconstruct_rate_percent = 32
					}
				}
			}

			controller_rates = {
				consistency_check_rate_percent = 32
				rebuild_rate_percent = 32
			}
		}
	}
		`,
		testingInfo.Username,
		testingInfo.Password,
		testingInfo.Endpoint,
		testingInfo.SystemID,
		testingInfo.StorageID,
		testingInfo.ControllerID,
	)
}

func testAccRedfishResourceStorageController17GFirstAvailableChoiceSelectedConfig(testingInfo testingStorageControllerInputs) string {
	return fmt.Sprintf(`
	resource "redfish_storage_controller" "test" {
		redfish_server {
			user         = "%s"
			password     = "%s"
			endpoint     = "%s"
			ssl_insecure = true
		}
		system_id = "%s"
		storage_id = "%s"
		controller_id = "%s"
		apply_time = "Immediate"
		job_timeout = 1200
		storage_controller = {
			oem = {
				dell = {
					dell_storage_controller = {
						background_initialization_rate_percent = 32
					}
				}
			}

			controller_rates = {
				consistency_check_rate_percent = 32
			}
		}
	}
		`,
		testingInfo.Username,
		testingInfo.Password,
		testingInfo.Endpoint,
		testingInfo.SystemID,
		testingInfo.StorageID,
		testingInfo.ControllerID,
	)
}

func testAccRedfishResourceStorageControllerSecondAvailableChoiceSelectedConfig(testingInfo testingStorageControllerInputs) string {
	return fmt.Sprintf(`
	resource "redfish_storage_controller" "test" {
		redfish_server {
			user         = "%s"
			password     = "%s"
			endpoint     = "%s"
			ssl_insecure = true
		}
		system_id = "%s"
		storage_id = "%s"
		controller_id = "%s"
		apply_time = "Immediate"
		job_timeout = 1200
		storage_controller = {
			oem = {
				dell = {
					dell_storage_controller = {
						check_consistency_mode = "StopOnError"
						copyback_mode = "Off"
						load_balance_mode = "Disabled"
						enhanced_auto_import_foreign_configuration_mode = "Enabled"
						patrol_read_unconfigured_area_mode = "Enabled"
						patrol_read_mode = "Automatic"
						background_initialization_rate_percent = 30
						reconstruct_rate_percent = 30
					}
				}
			}

			controller_rates = {
				consistency_check_rate_percent = 30
				rebuild_rate_percent = 30
			}
		}
	}
		`,
		testingInfo.Username,
		testingInfo.Password,
		testingInfo.Endpoint,
		testingInfo.SystemID,
		testingInfo.StorageID,
		testingInfo.ControllerID,
	)
}

func testAccRedfishResourceStorageController17GSecondAvailableChoiceSelectedConfig(testingInfo testingStorageControllerInputs) string {
	return fmt.Sprintf(`
	resource "redfish_storage_controller" "test" {
		redfish_server {
			user         = "%s"
			password     = "%s"
			endpoint     = "%s"
			ssl_insecure = true
		}
		system_id = "%s"
		storage_id = "%s"
		controller_id = "%s"
		apply_time = "Immediate"
		job_timeout = 1200
		storage_controller = {
			oem = {
				dell = {
					dell_storage_controller = {
						background_initialization_rate_percent = 30
					}
				}
			}

			controller_rates = {
				consistency_check_rate_percent = 30
			}
		}
	}
		`,
		testingInfo.Username,
		testingInfo.Password,
		testingInfo.Endpoint,
		testingInfo.SystemID,
		testingInfo.StorageID,
		testingInfo.ControllerID,
	)
}

func testAccRedfishResourceStorageControllerSecuritySetControllerKeyConfig(testingInfo testingStorageControllerInputs) string {
	return fmt.Sprintf(`
	resource "redfish_storage_controller" "test" {
		redfish_server {
			user         = "%s"
			password     = "%s"
			endpoint     = "%s"
			ssl_insecure = true
		}
		system_id = "%s"
		storage_id = "%s"
		controller_id = "%s"
		apply_time = "Immediate"
		job_timeout = 1200
		security = {
			action = "SetControllerKey"
			key_id = "testkey1"
			key = "Test123##"
		}
	}
		`,
		testingInfo.Username,
		testingInfo.Password,
		testingInfo.Endpoint,
		testingInfo.SystemID,
		testingInfo.StorageID,
		testingInfo.ControllerID,
	)
}

func testAccRedfishResourceStorageControllerSecurityReKeyConfig(testingInfo testingStorageControllerInputs) string {
	return fmt.Sprintf(`
	resource "redfish_storage_controller" "test" {
		redfish_server {
			user         = "%s"
			password     = "%s"
			endpoint     = "%s"
			ssl_insecure = true
		}
		system_id = "%s"
		storage_id = "%s"
		controller_id = "%s"
		apply_time = "Immediate"
		job_timeout = 1200
		security = {
			action = "ReKey"
			key_id = "testkey2"
			key = "Test123###"
			old_key = "Test123##"
			mode = "LKM"
		}
	}
		`,
		testingInfo.Username,
		testingInfo.Password,
		testingInfo.Endpoint,
		testingInfo.SystemID,
		testingInfo.StorageID,
		testingInfo.ControllerID,
	)
}

func testAccRedfishResourceStorageControllerSecurityRemoveControllerKeyConfig(testingInfo testingStorageControllerInputs) string {
	return fmt.Sprintf(`
	resource "redfish_storage_controller" "test" {
		redfish_server {
			user         = "%s"
			password     = "%s"
			endpoint     = "%s"
			ssl_insecure = true
		}
		system_id = "%s"
		storage_id = "%s"
		controller_id = "%s"
		apply_time = "Immediate"
		job_timeout = 1200
		security = {
			action = "RemoveControllerKey"
		}
	}
		`,
		testingInfo.Username,
		testingInfo.Password,
		testingInfo.Endpoint,
		testingInfo.SystemID,
		testingInfo.StorageID,
		testingInfo.ControllerID,
	)
}

func testAccRedfishResourceStorageController17GEnableSecurityConfig(testingInfo testingStorageControllerInputs) string {
	return fmt.Sprintf(`
	resource "redfish_storage_controller" "test" {
		redfish_server {
			user         = "%s"
			password     = "%s"
			endpoint     = "%s"
			ssl_insecure = true
		}
		system_id = "%s"
		storage_id = "%s"
		controller_id = "%s"
		apply_time = "Immediate"
		job_timeout = 1200
		security = {
			action = "EnableSecurity"
		}
	}
		`,
		testingInfo.Username,
		testingInfo.Password,
		testingInfo.Endpoint,
		testingInfo.SystemID,
		testingInfo.StorageID,
		testingInfo.ControllerID,
	)
}

func testAccRedfishResourceStorageController17GDisableSecurityConfig(testingInfo testingStorageControllerInputs) string {
	return fmt.Sprintf(`
	resource "redfish_storage_controller" "test" {
		redfish_server {
			user         = "%s"
			password     = "%s"
			endpoint     = "%s"
			ssl_insecure = true
		}
		system_id = "%s"
		storage_id = "%s"
		controller_id = "%s"
		apply_time = "Immediate"
		job_timeout = 1200
		security = {
			action = "DisableSecurity"
		}
	}
		`,
		testingInfo.Username,
		testingInfo.Password,
		testingInfo.Endpoint,
		testingInfo.SystemID,
		testingInfo.StorageID,
		testingInfo.ControllerID,
	)
}

func testAccRedfishResourceStorageControllerDifferentSystemIDConfig(testingInfo testingStorageControllerInputs) string {
	return fmt.Sprintf(`
	resource "redfish_storage_controller" "test" {
		redfish_server {
			user         = "%s"
			password     = "%s"
			endpoint     = "%s"
			ssl_insecure = true
		}
		system_id = "%s"
		storage_id = "%s"
		controller_id = "%s"
		apply_time = "Immediate"
		job_timeout = 1200
	}
		`,
		testingInfo.Username,
		testingInfo.Password,
		testingInfo.Endpoint,
		"Different_SystemID",
		testingInfo.StorageID,
		testingInfo.ControllerID,
	)
}

func testAccRedfishResourceStorageControllerDifferentStorageIDConfig(testingInfo testingStorageControllerInputs) string {
	return fmt.Sprintf(`
	resource "redfish_storage_controller" "test" {
		redfish_server {
			user         = "%s"
			password     = "%s"
			endpoint     = "%s"
			ssl_insecure = true
		}
		system_id = "%s"
		storage_id = "%s"
		controller_id = "%s"
		apply_time = "Immediate"
		job_timeout = 1200
	}
		`,
		testingInfo.Username,
		testingInfo.Password,
		testingInfo.Endpoint,
		testingInfo.SystemID,
		"Different_StorageID",
		testingInfo.ControllerID,
	)
}

func testAccRedfishResourceStorageControllerDifferentControllerIDConfig(testingInfo testingStorageControllerInputs) string {
	return fmt.Sprintf(`
	resource "redfish_storage_controller" "test" {
		redfish_server {
			user         = "%s"
			password     = "%s"
			endpoint     = "%s"
			ssl_insecure = true
		}
		system_id = "%s"
		storage_id = "%s"
		controller_id = "%s"
		apply_time = "Immediate"
		job_timeout = 1200
	}
		`,
		testingInfo.Username,
		testingInfo.Password,
		testingInfo.Endpoint,
		testingInfo.SystemID,
		testingInfo.StorageID,
		"Different_ControllerID",
	)
}

func testAccRedfishResourceStorageControllerControllerModeAndOtherAttributeUpdateConfig(testingInfo testingStorageControllerInputs) string {
	return fmt.Sprintf(`
	resource "redfish_storage_controller" "test" {
		redfish_server {
			user         = "%s"
			password     = "%s"
			endpoint     = "%s"
			ssl_insecure = true
		}
		system_id = "%s"
		storage_id = "%s"
		controller_id = "%s"
		apply_time = "OnReset"
		reset_type = "ForceRestart"
		reset_timeout = 120
		job_timeout = 1200
		storage_controller = {
			oem = {
				dell = {
					dell_storage_controller = {
						controller_mode = "HBA"
						check_consistency_mode = "Normal"
						copyback_mode = "On"
						load_balance_mode = "Automatic"
						enhanced_auto_import_foreign_configuration_mode = "Disabled"
						patrol_read_unconfigured_area_mode = "Disabled"
						patrol_read_mode = "Disabled"
						background_initialization_rate_percent = 32
						reconstruct_rate_percent = 32
					}
				}
			}

			controller_rates = {
				consistency_check_rate_percent = 32
				rebuild_rate_percent = 32
			}
		}
	}
		`,
		testingInfo.Username,
		testingInfo.Password,
		testingInfo.Endpoint,
		testingInfo.SystemID,
		testingInfo.StorageID,
		testingInfo.ControllerID,
	)
}

func testAccRedfishResourceStorageControllerControllerModeAndSecurityUpdateConfig(testingInfo testingStorageControllerInputs) string {
	return fmt.Sprintf(`
	resource "redfish_storage_controller" "test" {
		redfish_server {
			user         = "%s"
			password     = "%s"
			endpoint     = "%s"
			ssl_insecure = true
		}
		system_id = "%s"
		storage_id = "%s"
		controller_id = "%s"
		apply_time = "OnReset"
		reset_type = "ForceRestart"
		reset_timeout = 120
		job_timeout = 1200
		storage_controller = {
			oem = {
				dell = {
					dell_storage_controller = {
						controller_mode = "HBA"
					}
				}
			}
		}
		security = {
			action = "SetControllerKey"
			key_id = "testkey1"
			key = "Test123##"
		}
	}
		`,
		testingInfo.Username,
		testingInfo.Password,
		testingInfo.Endpoint,
		testingInfo.SystemID,
		testingInfo.StorageID,
		testingInfo.ControllerID,
	)
}

func testAccRedfishResourceStorageControllerControllerModeAndEnhancedAutoImportForeignConfigurationModeConfig(testingInfo testingStorageControllerInputs) string {
	return fmt.Sprintf(`
	resource "redfish_storage_controller" "test" {
		redfish_server {
			user         = "%s"
			password     = "%s"
			endpoint     = "%s"
			ssl_insecure = true
		}
		system_id = "%s"
		storage_id = "%s"
		controller_id = "%s"
		apply_time = "OnReset"
		reset_type = "ForceRestart"
		reset_timeout = 120
		job_timeout = 1200
		storage_controller = {
			oem = {
				dell = {
					dell_storage_controller = {
						controller_mode = "HBA"
						enhanced_auto_import_foreign_configuration_mode = "Enabled"
					}
				}
			}
		}
	}
		`,
		testingInfo.Username,
		testingInfo.Password,
		testingInfo.Endpoint,
		testingInfo.SystemID,
		testingInfo.StorageID,
		testingInfo.ControllerID,
	)
}

func testAccRedfishResourceStorageControllerControllerModeWithoutOnResetApplyTimeConfig(testingInfo testingStorageControllerInputs) string {
	return fmt.Sprintf(`
	resource "redfish_storage_controller" "test" {
		redfish_server {
			user         = "%s"
			password     = "%s"
			endpoint     = "%s"
			ssl_insecure = true
		}
		system_id = "%s"
		storage_id = "%s"
		controller_id = "%s"
		apply_time = "Immediate"
		job_timeout = 1200
		storage_controller = {
			oem = {
				dell = {
					dell_storage_controller = {
						controller_mode = "HBA"
					}
				}
			}
		}
	}
		`,
		testingInfo.Username,
		testingInfo.Password,
		testingInfo.Endpoint,
		testingInfo.SystemID,
		testingInfo.StorageID,
		testingInfo.ControllerID,
	)
}

func testAccRedfishResourceStorageControllerSecurityAndOtherAttributeUpdateConfig(testingInfo testingStorageControllerInputs) string {
	return fmt.Sprintf(`
	resource "redfish_storage_controller" "test" {
		redfish_server {
			user         = "%s"
			password     = "%s"
			endpoint     = "%s"
			ssl_insecure = true
		}
		system_id = "%s"
		storage_id = "%s"
		controller_id = "%s"
		apply_time = "Immediate"
		job_timeout = 1200
		storage_controller = {
			oem = {
				dell = {
					dell_storage_controller = {
						check_consistency_mode = "Normal"
						copyback_mode = "On"
						load_balance_mode = "Automatic"
						enhanced_auto_import_foreign_configuration_mode = "Disabled"
						patrol_read_unconfigured_area_mode = "Disabled"
						patrol_read_mode = "Disabled"
						background_initialization_rate_percent = 32
						reconstruct_rate_percent = 32
					}
				}
			}
			controller_rates = {
				consistency_check_rate_percent = 32
				rebuild_rate_percent = 32
			}
		}
		security = {
			action = "SetControllerKey"
			key_id = "testkey1"
			key = "Test123##"
		}
	}
		`,
		testingInfo.Username,
		testingInfo.Password,
		testingInfo.Endpoint,
		testingInfo.SystemID,
		testingInfo.StorageID,
		testingInfo.ControllerID,
	)
}

func testAccRedfishResourceStorageController17GSecurityAndOtherAttributeUpdateConfig(testingInfo testingStorageControllerInputs) string {
	return fmt.Sprintf(`
	resource "redfish_storage_controller" "test" {
		redfish_server {
			user         = "%s"
			password     = "%s"
			endpoint     = "%s"
			ssl_insecure = true
		}
		system_id = "%s"
		storage_id = "%s"
		controller_id = "%s"
		apply_time = "Immediate"
		job_timeout = 1200
		storage_controller = {
			oem = {
				dell = {
					dell_storage_controller = {
						background_initialization_rate_percent = 32
					}
				}
			}
			controller_rates = {
				consistency_check_rate_percent = 32
			}
		}
		security = {
			action = "EnableSecurity"
		}
	}
		`,
		testingInfo.Username,
		testingInfo.Password,
		testingInfo.Endpoint,
		testingInfo.SystemID,
		testingInfo.StorageID,
		testingInfo.ControllerID,
	)
}

func testAccRedfishResourceStorageControllerSecurityWithMaintenanceTypeApplyTimeConfig(testingInfo testingStorageControllerInputs) string {
	return fmt.Sprintf(`
	resource "redfish_storage_controller" "test" {
		redfish_server {
			user         = "%s"
			password     = "%s"
			endpoint     = "%s"
			ssl_insecure = true
		}
		system_id = "%s"
		storage_id = "%s"
		controller_id = "%s"
		apply_time = "AtMaintenanceWindowStart"
		job_timeout = 1200
		maintenance_window = {
			start_time = "2024-10-15T22:45:00-05:00"
			duration = 600
		}
		security = {
			action = "SetControllerKey"
			key_id = "testkey1"
			key = "Test123##"
		}
	}
		`,
		testingInfo.Username,
		testingInfo.Password,
		testingInfo.Endpoint,
		testingInfo.SystemID,
		testingInfo.StorageID,
		testingInfo.ControllerID,
	)
}

func testAccRedfishResourceStorageController17GSecurityWithMaintenanceTypeApplyTimeConfig(testingInfo testingStorageControllerInputs) string {
	return fmt.Sprintf(`
	resource "redfish_storage_controller" "test" {
		redfish_server {
			user         = "%s"
			password     = "%s"
			endpoint     = "%s"
			ssl_insecure = true
		}
		system_id = "%s"
		storage_id = "%s"
		controller_id = "%s"
		apply_time = "AtMaintenanceWindowStart"
		job_timeout = 1200
		maintenance_window = {
			start_time = "2024-10-15T22:45:00-05:00"
			duration = 600
		}
		security = {
			action = "EnableSecurity"
		}
	}
		`,
		testingInfo.Username,
		testingInfo.Password,
		testingInfo.Endpoint,
		testingInfo.SystemID,
		testingInfo.StorageID,
		testingInfo.ControllerID,
	)
}

func testAccRedfishResourceStorageController17GStorageControllerWithMaintenanceTypeApplyTimeConfig(testingInfo testingStorageControllerInputs) string {
	return fmt.Sprintf(`
	resource "redfish_storage_controller" "test" {
		redfish_server {
			user         = "%s"
			password     = "%s"
			endpoint     = "%s"
			ssl_insecure = true
		}
		system_id = "%s"
		storage_id = "%s"
		controller_id = "%s"
		apply_time = "AtMaintenanceWindowStart"
		job_timeout = 1200
		maintenance_window = {
			start_time = "2024-10-15T22:45:00-05:00"
			duration = 600
		}
		storage_controller = {
			oem = {
				dell = {
					dell_storage_controller = {
						background_initialization_rate_percent = 32
					}
				}
			}
			controller_rates = {
				consistency_check_rate_percent = 32
			}
		}
	}
		`,
		testingInfo.Username,
		testingInfo.Password,
		testingInfo.Endpoint,
		testingInfo.SystemID,
		testingInfo.StorageID,
		testingInfo.ControllerID,
	)
}

func testAccRedfishResourceStorageControllerSecurityWithoutActionConfig(testingInfo testingStorageControllerInputs) string {
	return fmt.Sprintf(`
	resource "redfish_storage_controller" "test" {
		redfish_server {
			user         = "%s"
			password     = "%s"
			endpoint     = "%s"
			ssl_insecure = true
		}
		system_id = "%s"
		storage_id = "%s"
		controller_id = "%s"
		apply_time = "Immediate"
		job_timeout = 1200
		security = {
			key_id = "testkey1"
			key = "Test123##"
		}
	}
		`,
		testingInfo.Username,
		testingInfo.Password,
		testingInfo.Endpoint,
		testingInfo.SystemID,
		testingInfo.StorageID,
		testingInfo.ControllerID,
	)
}

func testAccRedfishResourceStorageControllerSecuritySetControllerKeyIncorrectConfig(testingInfo testingStorageControllerInputs) string {
	return fmt.Sprintf(`
	resource "redfish_storage_controller" "test" {
		redfish_server {
			user         = "%s"
			password     = "%s"
			endpoint     = "%s"
			ssl_insecure = true
		}
		system_id = "%s"
		storage_id = "%s"
		controller_id = "%s"
		apply_time = "Immediate"
		job_timeout = 1200
		security = {
			action = "SetControllerKey"
			key_id = "testkey1"
		}
	}
		`,
		testingInfo.Username,
		testingInfo.Password,
		testingInfo.Endpoint,
		testingInfo.SystemID,
		testingInfo.StorageID,
		testingInfo.ControllerID,
	)
}

func testAccRedfishResourceStorageControllerSecurityReKeyIncorrectConfig(testingInfo testingStorageControllerInputs) string {
	return fmt.Sprintf(`
	resource "redfish_storage_controller" "test" {
		redfish_server {
			user         = "%s"
			password     = "%s"
			endpoint     = "%s"
			ssl_insecure = true
		}
		system_id = "%s"
		storage_id = "%s"
		controller_id = "%s"
		apply_time = "Immediate"
		job_timeout = 1200
		security = {
			action = "ReKey"
			key_id = "testkey2"
			key = "Test123###"
			mode = "LKM"
		}
	}
		`,
		testingInfo.Username,
		testingInfo.Password,
		testingInfo.Endpoint,
		testingInfo.SystemID,
		testingInfo.StorageID,
		testingInfo.ControllerID,
	)
}

func testAccRedfishResourceStorageControllerSecurityRemoveControllerKeyIncorrectConfig(testingInfo testingStorageControllerInputs) string {
	return fmt.Sprintf(`
	resource "redfish_storage_controller" "test" {
		redfish_server {
			user         = "%s"
			password     = "%s"
			endpoint     = "%s"
			ssl_insecure = true
		}
		system_id = "%s"
		storage_id = "%s"
		controller_id = "%s"
		apply_time = "Immediate"
		job_timeout = 1200
		security = {
			action = "RemoveControllerKey"
			key_id = "testkey1"
		}
	}
		`,
		testingInfo.Username,
		testingInfo.Password,
		testingInfo.Endpoint,
		testingInfo.SystemID,
		testingInfo.StorageID,
		testingInfo.ControllerID,
	)
}
