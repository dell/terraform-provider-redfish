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

func TestAccRedfishDirectoryServiceAuthProviderBasic(t *testing.T) {
	version := os.Getenv("TF_TESTING_REDFISH_VERSION")
	if version == "17" {
		t.Skip("Skipping DirectoryService for 17G")
	}
	terraformDSAuthProviderResourceName := "redfish_directory_service_auth_provider.ds_auth"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				// error create with both `ActiveDirectory` and `LDAP`
				PreConfig: func() {
					FunctionMocker = mockey.Mock(isServerGenerationSeventeenAndAbove).Return(false, nil).Build()
				},
				Config:      testAccRedfishDirectoryServiceAuthProviderErrorConfig(creds),
				ExpectError: regexp.MustCompile("Error when creating both of `ActiveDirectory` and `LDAP`"),
			},

			{
				// create with `ActiveDirectory`
				Config: testAccRedfishDirectoryServiceAuthProviderADConfig(creds),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(terraformDSAuthProviderResourceName, "active_directory.directory.service_enabled", "true"),
				),
			},

			{
				// update with `LDAP`
				Config: testAccRedfishDirectoryServiceAuthProviderLDAPConfig(creds),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(terraformDSAuthProviderResourceName, "ldap.directory.service_enabled", "false"),
				),
			},
			{
				// error update with both `ActiveDirectory` and `LDAP`
				Config:      testAccRedfishDirectoryServiceAuthProviderErrorConfig(creds),
				ExpectError: regexp.MustCompile("Error when updating both of `ActiveDirectory` and `LDAP`"),
			},
			{
				// update with `ActiveDirectory`
				Config: testAccRedfishDirectoryServiceAuthProviderAD_UpdateConfig(creds),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(terraformDSAuthProviderResourceName, "active_directory.directory.service_enabled", "false"),
					resource.TestCheckResourceAttr(terraformDSAuthProviderResourceName, "active_directory_attributes.ActiveDirectory.1.AuthTimeout", "130"),
				),
			},
			{
				// update with `ActiveDirectory` and standard Schema
				Config: testAccRedfishDirectoryServiceAuthProviderADWithStandardSchema_UpdateConfig(creds),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(terraformDSAuthProviderResourceName, "active_directory.directory.service_enabled", "true"),
				),
			},
		},
	})
	if FunctionMocker != nil {
		FunctionMocker.Release()
	}
}

func TestAccRedfishDirectoryServiceAuthProviderBasic_17GConfig(t *testing.T) {
	version := os.Getenv("TF_TESTING_REDFISH_VERSION")
	if version != "17" {
		t.Skip("Skipping DirectoryService 17G tests for below 17G")
	}
	terraformDSAuthProviderResourceName := "redfish_directory_service_auth_provider.ds_auth"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				// error create with both `ActiveDirectory` and `LDAP`
				PreConfig: func() {
					FunctionMocker = mockey.Mock(isServerGenerationSeventeenAndAbove).Return(true, nil).Build()
				},
				Config:      testAccRedfishDirectoryServiceAuthProviderError17GConfig(creds),
				ExpectError: regexp.MustCompile("Error when creating both of `ActiveDirectory` and `LDAP`"),
			},
			{
				// create with `ActiveDirectory`
				Config: testAccRedfishDirectoryServiceAuthProvider17GConfig(creds),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(terraformDSAuthProviderResourceName, "active_directory.directory.service_enabled", "false"),
				),
			},

			{
				// update with `LDAP`
				Config: testAccRedfishDirectoryServiceAuthProviderLDAP17GConfig(creds),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(terraformDSAuthProviderResourceName, "ldap.directory.service_enabled", "false"),
				),
			},
			{
				// error update with both `ActiveDirectory` and `LDAP`
				Config:      testAccRedfishDirectoryServiceAuthProviderError17GConfig(creds),
				ExpectError: regexp.MustCompile("Error when updating both of `ActiveDirectory` and `LDAP`"),
			},
			{
				// update with `ActiveDirectory`
				Config: testAccRedfishDirectoryServiceAuthProviderAD_Update17GConfig(creds),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(terraformDSAuthProviderResourceName, "active_directory.directory.service_enabled", "true"),
					resource.TestCheckResourceAttr(terraformDSAuthProviderResourceName, "active_directory_attributes.ActiveDirectory.1.AuthTimeout", "130"),
				),
			},
		},
	})
	if FunctionMocker != nil {
		FunctionMocker.Release()
	}
}

func TestAccRedfishDirectoryServiceAuthProviderInvalidCase(t *testing.T) {
	version := os.Getenv("TF_TESTING_REDFISH_VERSION")
	if version == "17" {
		t.Skip("Skipping DirectoryService test for 17G")
	}
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				// error for Active Directory Service
				PreConfig: func() {
					FunctionMocker = mockey.Mock(isServerGenerationSeventeenAndAbove).Return(false, nil).Build()
				},
				Config:      testAccRedfishDirectoryServiceAuthProviderADWithStandardSchema_ServiceErrorConfig(creds),
				ExpectError: regexp.MustCompile("Error updating AccountService Details"),
			},
			{
				// error for LDAP Service
				Config:      testAccRedfishDirectoryServiceAuthProviderLDAP_ServiceErrorConfig(creds),
				ExpectError: regexp.MustCompile("Error updating AccountService Details"),
			},
			{
				// error for empty AuthTimeout in ActiveDirectory
				Config:      testAccRedfishDirectoryServiceAuthProviderEmptyAuth(creds),
				ExpectError: regexp.MustCompile("Invalid AuthTimeout, Please provide all the required configuration"),
			},
			{
				// error for Invalid AuthTimeout in ActiveDirectory
				Config:      testAccRedfishDirectoryServiceAuthProviderInvalidAuthTimeoutString(creds),
				ExpectError: regexp.MustCompile("Invalid AuthTimeout"),
			},

			{
				// error for Invalid AuthTimeout not in (15,300) in ActiveDirectory
				Config:      testAccRedfishDirectoryServiceAuthProviderInvalidAuth(creds),
				ExpectError: regexp.MustCompile("Invalid AuthTimeout, AuthTimeout must be between 15 and 300"),
			},
			{
				// error ActiveDirectoryService Disabled and SSOEnable Enabled
				Config:      testAccRedfishDirectoryServiceAuthProviderADDisableSSOEnable(creds),
				ExpectError: regexp.MustCompile("Please provide valid Configuration for SSO"),
			},

			{
				// error ActiveDirectoryService Enabled and SSOEnable Enabled and no Kerberos key tab
				Config:      testAccRedfishDirectoryServiceAuthProviderADEnSSOEnNoKb(creds),
				ExpectError: regexp.MustCompile("Please provide valid kerberos key tab file when SSO is enabled"),
			},

			{
				// error DCLookupEnable Enabled and service address as empty
				Config:      testAccRedfishDirectoryServiceAuthProviderDClookUpEnServiceAddEmpty(creds),
				ExpectError: regexp.MustCompile("ServiceAddresses is not Configured for Disabled DCLookUp"),
			},

			{
				// error DCLookupEnable Disabled and DCLookupByUserDomain config
				Config:      testAccRedfishDirectoryServiceAuthProviderDCLookupByUserDomainConfig(creds),
				ExpectError: regexp.MustCompile("DCLookupByUserDomain can not be Configured for Disabled DCLookUp"),
			},

			{
				// error DCLookupEnable Disabled and DCLookupDomainName config
				Config:      testAccRedfishDirectoryServiceAuthProviderDCLookupDomainNameConfig(creds),
				ExpectError: regexp.MustCompile("DCLookupDomainName can not be configured for Disabled DCLookUp"),
			},

			{
				// error DCLookupEnable Enabled and ServiceAddress non empty
				Config:      testAccRedfishDirectoryServiceAuthProviderDDCLookupEnableNoServiceAddConfig(creds),
				ExpectError: regexp.MustCompile("Service address can not be configured for Enabled DCLookUp"),
			},

			{
				// error DCLookupEnable Enabled and DCLookupByUserDomain does not exist
				Config:      testAccRedfishDirectoryServiceAuthProviderDCLookupByUserDomainEmptyConfig(creds),
				ExpectError: regexp.MustCompile("DCLookupByUserDomain must be configured for Enabled DCLookUp"),
			},

			{
				// error DCLookupEnable Enabled, DCLookupByUserDomain Disabled and without DCLookupDomainName
				Config:      testAccRedfishDirectoryServiceAuthProviderWithoutDCLookupDomainNameConfig(creds),
				ExpectError: regexp.MustCompile("DCLookupDomainName must be configured for Disabled DCLookupByUserDomain"),
			},
			{
				// error DCLookupEnable Enabled, DCLookupByUserDomain Disabled and DCLookupDomainName Empty
				Config:      testAccRedfishDirectoryServiceAuthProviderDCLookupDomainNameEmptyConfig(creds),
				ExpectError: regexp.MustCompile("DCLookupDomainName must be configured for Disabled DCLookupByUserDomain"),
			},

			{
				// error DCLookupEnable Enabled DCLookupByUserDomain Enabled and DCLookupDomainName non Empty
				Config:      testAccRedfishDirectoryServiceAuthProviderDCLookupEnableDCLookupDomainNameConfig(creds),
				ExpectError: regexp.MustCompile("DCLookupDomainName can not be configured for Enabled DCLookupByUserDomain"),
			},
			{
				// error DCLookupEnable Invalid
				Config:      testAccRedfishDirectoryServiceAuthProviderDCLookupEnableInvalidConfig(creds),
				ExpectError: regexp.MustCompile("Invalid configuration for DCLookUp"),
			},
			{
				// error service address as emoty for LDAP
				Config:      testAccRedfishDirectoryServiceAuthProviderLDAPServiceAddEmptyConfig(creds),
				ExpectError: regexp.MustCompile("ServiceAddresses is not be Configured for LDAP"),
			},
		},
	})
	if FunctionMocker != nil {
		FunctionMocker.Release()
	}
}

func TestAccRedfishDirectoryServiceAuthProvider17GInvalidCase(t *testing.T) {
	version := os.Getenv("TF_TESTING_REDFISH_VERSION")
	if version != "17" {
		t.Skip("Skipping DirectoryService for below 17G")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{

			{
				// error for empty AuthTimeout in ActiveDirectory
				PreConfig: func() {
					FunctionMocker = mockey.Mock(isServerGenerationSeventeenAndAbove).Return(true, nil).Build()
				},
				Config:      testAccRedfishDirectoryServiceAuthProviderEmptyAuth17GConfig(creds),
				ExpectError: regexp.MustCompile("Invalid AuthTimeout, Please provide all the required configuration"),
			},
			{
				// error for Invalid AuthTimeout in ActiveDirectory
				Config:      testAccRedfishDirectoryServiceAuthProviderInvalidAuthTimeoutString17GConfig(creds),
				ExpectError: regexp.MustCompile("Invalid AuthTimeout"),
			},

			{
				// error for Invalid AuthTimeout not in (15,300) in ActiveDirectory
				Config:      testAccRedfishDirectoryServiceAuthProviderInvalidAuth17GConfig(creds),
				ExpectError: regexp.MustCompile("Invalid AuthTimeout, AuthTimeout must be between 15 and 300"),
			},
			{
				// error DCLookupEnable Disabled and service address as non emoty for 17G
				Config:      testAccRedfishDirectoryServiceAuthProviderDClookUpDsServiceAddNonEmpty17GConfig(creds),
				ExpectError: regexp.MustCompile("ServiceAddresses can not be Configured for 17 Gen"),
			},

			{
				// error DCLookupEnable Disabled and no domaincontroller configured
				Config:      testAccRedfishDirectoryServiceAuthProviderDClookUpDsServiceAddEmpty17GConfig(creds),
				ExpectError: regexp.MustCompile("DomainController server address is not Configured for Disabled DCLookUp"),
			},

			{
				// error DCLookupEnable Disabled and DCLookupByUserDomain config
				Config:      testAccRedfishDirectoryServiceAuthProviderDCLookupByUserDomainConfig17GConfig(creds),
				ExpectError: regexp.MustCompile("DCLookupByUserDomain can not be Configured for Disabled DCLookUp"),
			},

			{
				// error DCLookupEnable Disabled and DCLookupDomainName config
				Config:      testAccRedfishDirectoryServiceAuthProviderDCLookupDomainName17GConfig(creds),
				ExpectError: regexp.MustCompile("DCLookupDomainName can not be configured for Disabled DCLookUp"),
			},

			{
				// error DCLookupEnable Enabled and DomainController non empty
				Config:      testAccRedfishDirectoryServiceAuthProviderDDCLookupEnableNoServiceAdd17GConfig(creds),
				ExpectError: regexp.MustCompile("DomainController server address can not be configured for Enabled DCLookUp"),
			},

			{
				// error DCLookupEnable Enabled and DCLookupByUserDomain does not exist
				Config:      testAccRedfishDirectoryServiceAuthProviderDCLookupByUserDomainEmpty17GConfig(creds),
				ExpectError: regexp.MustCompile("DCLookupByUserDomain must be configured for Enabled DCLookUp"),
			},

			{
				// error DCLookupEnable Enabled, DCLookupByUserDomain Disabled and without DCLookupDomainName
				Config:      testAccRedfishDirectoryServiceAuthProviderWithoutDCLookupDomainName17GConfig(creds),
				ExpectError: regexp.MustCompile("DCLookupDomainName must be configured for Disabled DCLookupByUserDomain"),
			},
			{
				// error DCLookupEnable Enabled, DCLookupByUserDomain Disabled and DCLookupDomainName Empty
				Config:      testAccRedfishDirectoryServiceAuthProviderDCLookupDomainNameEmpty17GConfig(creds),
				ExpectError: regexp.MustCompile("DCLookupDomainName must be configured for Disabled DCLookupByUserDomain"),
			},

			{
				// error DCLookupEnable Enabled DCLookupByUserDomain Enabled and DCLookupDomainName non Empty
				Config:      testAccRedfishDirectoryServiceAuthProviderDCLookupEnableDCLookupDomainName17GConfig(creds),
				ExpectError: regexp.MustCompile("DCLookupDomainName can not be configured for Enabled DCLookupByUserDomain"),
			},
			{
				// error DCLookupEnable Invalid
				Config:      testAccRedfishDirectoryServiceAuthProviderDCLookupEnableInvalid17GConfig(creds),
				ExpectError: regexp.MustCompile("Invalid configuration for DCLookUp"),
			},

			{
				// error service address as non emoty for 17G with LDAP
				Config:      testAccRedfishDirectoryServiceAuthProviderLDAPServiceAddNonEmpty17GConfig(creds),
				ExpectError: regexp.MustCompile("ServiceAddresses can not be Configured for 17G LDAP"),
			},

			{
				// error no Server Address configured for 17G with LDAP
				Config:      testAccRedfishDirectoryServiceAuthProviderLDAPServerAddressEmpty17GConfig(creds),
				ExpectError: regexp.MustCompile("server address is not Configured for 17G LDAP"),
			},
		},
	})
}

func TestAccRedfishDirectoryServiceAuthProviderInvalidSchema_Config(t *testing.T) {
	version := os.Getenv("TF_TESTING_REDFISH_VERSION")
	if version == "17" {
		t.Skip("Skipping DirectoryService for 17G")
	}
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				// error Extended Schema without RacName and RacDomain
				PreConfig: func() {
					FunctionMocker = mockey.Mock(isServerGenerationSeventeenAndAbove).Return(false, nil).Build()
				},
				Config:      testAccRedfishDirectoryServiceAuthProviderExtendedNoRacConfig(creds),
				ExpectError: regexp.MustCompile("Please provide valid config for RacName and RacDomain"),
			},
			{
				// error Extended Schema with empty RacName and RacDomain
				Config:      testAccRedfishDirectoryServiceAuthProviderExtendedEmptyRacConfig(creds),
				ExpectError: regexp.MustCompile("RacName and RacDomain can not be null for Extended Schema"),
			},
			{
				// error Extended Schema with GCLookup config
				Config:      testAccRedfishDirectoryServiceAuthProviderExtendedGCLookUpConfig(creds),
				ExpectError: regexp.MustCompile("GCLookupEnable, GlobalCatalog1, GlobalCatalog2, GlobalCatalog3 can not be configured for Extended Schema"),
			},

			{
				// error Extended Schema with remote role mapping config
				Config:      testAccRedfishDirectoryServiceAuthProviderExtendedRemoteRoleConfig(creds),
				ExpectError: regexp.MustCompile("RemoteRoleMapping can not be configured for Extended Schema"),
			},
			{
				// error Extended Schema with groupdomain config
				Config:      testAccRedfishDirectoryServiceAuthProviderExtendedADGroupDomainConfig(creds),
				ExpectError: regexp.MustCompile("Domain can not be configured for Extended Schema"),
			},
			{
				// error Standard Schema with RacName and RacDomain
				Config:      testAccRedfishDirectoryServiceAuthProviderStandardSchemaAndRacConfig(creds),
				ExpectError: regexp.MustCompile("RacName and RacDomain can not be configured for Standard Schema"),
			},
			{
				// error Standard Schema without GCLookup config
				Config:      testAccRedfishDirectoryServiceAuthProviderStandardSchemaNoGCLookUpConfig(creds),
				ExpectError: regexp.MustCompile("GCLookupEnable must be configured for Standard Schema"),
			},
			{
				// error Standard Schema, GCLookup Enabled, no GCRootDomain
				Config:      testAccRedfishDirectoryServiceAuthProviderStandardSchemaNoGCRootConfig(creds),
				ExpectError: regexp.MustCompile("GCRootDomain must be configured for Enabled GCLookupEnable"),
			},
			{
				// error Standard Schema GCLookup Enabled, with global catalog config
				Config:      testAccRedfishDirectoryServiceAuthProviderStandardSchemaGlobalCatalogConfig(creds),
				ExpectError: regexp.MustCompile("GlobalCatalog can not be configured for Enabled GCLookupEnable"),
			},
			{
				// error Standard Schema GCLookup Disabled with no Globalcatalog config
				Config:      testAccRedfishDirectoryServiceAuthProviderStandardSchemaNoGlobalCatalogConfig(creds),
				ExpectError: regexp.MustCompile("Invalid GlobalCatalog configuration for Standard Schema"),
			},
			{
				// error Standard Schema GCLookup Disabled with GcRootDomain config
				Config:      testAccRedfishDirectoryServiceAuthProviderStandardSchemaGCRootConfig(creds),
				ExpectError: regexp.MustCompile("GCRootDomain can not be configured for Disabled GCLookupEnable"),
			},
			{
				// error Standard Schema and Invalid GCLookup config
				Config:      testAccRedfishDirectoryServiceAuthProviderStandardSchemaInvalidGCLookUpConfig(creds),
				ExpectError: regexp.MustCompile("Invalid configuration for Standard Schema"),
			},
		},
	})
	if FunctionMocker != nil {
		FunctionMocker.Release()
	}
}

func TestAccRedfishDirectoryServiceAuthProviderInvalidSchema_17GConfig(t *testing.T) {
	version := os.Getenv("TF_TESTING_REDFISH_VERSION")
	if version != "17" {
		t.Skip("Skipping DirectoryService for 17G")
	}
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				// error Standard Schema with RacName and RacDomain
				PreConfig: func() {
					FunctionMocker = mockey.Mock(isServerGenerationSeventeenAndAbove).Return(true, nil).Build()
				},
				Config:      testAccRedfishDirectoryServiceAuthProviderStandardSchemaAndRac17GConfig(creds),
				ExpectError: regexp.MustCompile("RacName and RacDomain can not be configured for 17G which does not support extended schema"),
			},
			{
				// error Standard Schema without GCLookup config
				Config:      testAccRedfishDirectoryServiceAuthProviderStandardSchemaNoGCLookUp17GConfig(creds),
				ExpectError: regexp.MustCompile("GCLookupEnable must be configured"),
			},
			{
				// error Standard Schema, GCLookup Enabled, no GCRootDomain
				Config:      testAccRedfishDirectoryServiceAuthProviderStandardSchemaNoGCRoot17GConfig(creds),
				ExpectError: regexp.MustCompile("GCRootDomain must be configured for Enabled GCLookupEnable"),
			},
			{
				// error Standard Schema GCLookup Enabled, with global catalog config
				Config:      testAccRedfishDirectoryServiceAuthProviderStandardSchemaGlobalCatalog17GConfig(creds),
				ExpectError: regexp.MustCompile("GlobalCatalog can not be configured for Enabled GCLookupEnable"),
			},
			{
				// error Standard Schema GCLookup Disabled with no Globalcatalog config
				Config:      testAccRedfishDirectoryServiceAuthProviderStandardSchemaNoGlobalCatalog17GConfig(creds),
				ExpectError: regexp.MustCompile("Invalid GlobalCatalog configuration for Standard Schema"),
			},
			{
				// error Standard Schema GCLookup Disabled with GcRootDomain config
				Config:      testAccRedfishDirectoryServiceAuthProviderStandardSchemaGCRoot17GConfig(creds),
				ExpectError: regexp.MustCompile("GCRootDomain can not be configured for Disabled GCLookupEnable"),
			},
			{
				// error Standard Schema and Invalid GCLookup config
				Config:      testAccRedfishDirectoryServiceAuthProviderStandardSchemaInvalidGCLookUp17GConfig(creds),
				ExpectError: regexp.MustCompile("Invalid configuration for Standard Schema"),
			},
		},
	})
	if FunctionMocker != nil {
		FunctionMocker.Release()
	}
}

func TestAccRedfishDirectoryServiceAuthProviderImport(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				PreConfig: func() {
					FunctionMocker = mockey.Mock(isServerGenerationSeventeenAndAbove).Return(false, nil).Build()
				},
				Config: `resource "redfish_directory_service_auth_provider" "ds_auth" {
					}`,
				ResourceName:  "redfish_directory_service_auth_provider.ds_auth",
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

func testAccRedfishDirectoryServiceAuthProviderErrorConfig(testingInfo TestingServerCredentials) string {
	return fmt.Sprintf(`
	resource "redfish_directory_service_auth_provider" "ds_auth" {
		redfish_server {
		  user         = "%s"
		  password     = "%s"
		  endpoint     = "%s"
		  ssl_insecure = true
		}
	  
		active_directory = {
			directory = {
				service_enabled = false
				
			}
		}
		
		active_directory_attributes = {
			"ActiveDirectory.1.AuthTimeout"= "110",
			"ActiveDirectory.1.CertValidationEnable"= "Enabled",
			"ActiveDirectory.1.DCLookupEnable"= "Enabled",
			"ActiveDirectory.1.RacDomain"= "test",
			"ActiveDirectory.1.RacName"= "test",
			"ActiveDirectory.1.SSOEnable"= "Disabled",
			"ActiveDirectory.1.Schema"= "Extended Schema",
			"UserDomain.1.Name"= "yulan1.pie.lab.emc.com",
			"ActiveDirectory.1.DCLookupByUserDomain"="Enabled",          
		}
		
		ldap = {
			directory = {
				remote_role_mapping = [
					{
						local_role = "Administrator",
						remote_group = "cn = idracgroup,cn = users,dc = yulan,dc = pie,dc = lab,dc = emc,dc = com"
					}        
				],
				service_addresses = [
					"yulanadhost12.yulan.pie.lab.emc.com"
				],
				service_enabled = true
			},
			ldap_service = {
				search_settings = {
					base_distinguished_names = [
						  "dc = yulan11,dc = pie,dc = lab,dc = emc,dc = com"
					],
					group_name_attribute = "name",
					user_name_attribute = "member"
				}
			}
		}
		
		 ldap_attributes = {
		  "LDAP.1.GroupAttributeIsDN" = "Enabled"
		  "LDAP.1.Port" = "636",
		  "LDAP.1.BindDN" = "cn = adtester,cn = users,dc = yulan,dc = pie,dc = lab,dc = emc,dc = com",
		  "LDAP.1.BindPassword" = "",
		  "LDAP.1.SearchFilter" = "(objectclass = *)"
		  }
		}
	  `,
		testingInfo.Username,
		testingInfo.Password,
		testingInfo.Endpoint,
	)
}

func testAccRedfishDirectoryServiceAuthProviderError17GConfig(testingInfo TestingServerCredentials) string {
	return fmt.Sprintf(`
	resource "redfish_directory_service_auth_provider" "ds_auth" {
		redfish_server {
		  user         = "%s"
		  password     = "%s"
		  endpoint     = "%s"
		  ssl_insecure = true
		}
	  
		active_directory = {
			directory = {
				service_enabled = false
				
			}
		}
		
		active_directory_attributes = {
			"ActiveDirectory.1.AuthTimeout"= "110",
			"ActiveDirectory.1.CertValidationEnable"= "Enabled",
			"ActiveDirectory.1.DCLookupEnable"= "Enabled",
			"UserDomain.1.Name"= "yulan1.pie.lab.emc.com",
			"ActiveDirectory.1.DCLookupByUserDomain"="Enabled",          
		}
		
		ldap = {
			directory = {
				remote_role_mapping = [
					{
						local_role = "Administrator",
						remote_group = "cn = idracgroup,cn = users,dc = yulan,dc = pie,dc = lab,dc = emc,dc = com"
					}        
				],
				service_addresses = [
					"yulanadhost12.yulan.pie.lab.emc.com"
				],
				service_enabled = true
			},
			ldap_service = {
				search_settings = {
					base_distinguished_names = [
						  "dc = yulan11,dc = pie,dc = lab,dc = emc,dc = com"
					],
					group_name_attribute = "name",
					user_name_attribute = "member"
				}
			}
		}
		
		 ldap_attributes = {
		  "LDAP.1.GroupAttributeIsDN" = "Enabled"
		  "LDAP.1.Port" = "636",
		  "LDAP.1.BindDN" = "cn = adtester,cn = users,dc = yulan,dc = pie,dc = lab,dc = emc,dc = com",
		  "LDAP.1.BindPassword" = "",
		  "LDAP.1.SearchFilter" = "(objectclass = *)"
		  }
		}
	  `,
		testingInfo.Username,
		testingInfo.Password,
		testingInfo.Endpoint,
	)
}

func testAccRedfishDirectoryServiceAuthProvider17GConfig(testingInfo TestingServerCredentials) string {
	return fmt.Sprintf(`
	resource "redfish_directory_service_auth_provider" "ds_auth" {
		redfish_server {
		  user         = "%s"
		  password     = "%s"
		  endpoint     = "%s"
		  ssl_insecure = true
		}

		active_directory = {
			directory = {
				service_enabled = false
				remote_role_mapping = [
           {
              local_role = "Administrator",
               remote_group = "xxxx"
           }
       ],
			}
		}

		active_directory_attributes = {
			"ActiveDirectory.1.AuthTimeout"= "110",
			"ActiveDirectory.1.CertValidationEnable"= "Enabled",
			"ActiveDirectory.1.DCLookupEnable"= "Enabled",
			"UserDomain.1.Name"        = "yulan.pie.lab.emc.com",
			"ActiveDirectory.1.DCLookupByUserDomain" : "Enabled",
			"ActiveDirectory.1.GCLookupEnable" = "Disabled",
			"ActiveDirectory.1.GlobalCatalog1" = "yulanadhost.yulan.pie.lab.emc.com",
		}
	}
	  `,
		testingInfo.Username,
		testingInfo.Password,
		testingInfo.Endpoint,
	)
}

func testAccRedfishDirectoryServiceAuthProviderADConfig(testingInfo TestingServerCredentials) string {
	return fmt.Sprintf(`
	resource "redfish_directory_service_auth_provider" "ds_auth" {
		redfish_server {
		  user         = "%s"
		  password     = "%s"
		  endpoint     = "%s"
		  ssl_insecure = true
		}
		active_directory = {
			directory = {
				service_enabled = true
				
			}
		}
		
		active_directory_attributes = {
			"ActiveDirectory.1.AuthTimeout"= "120",
			"ActiveDirectory.1.CertValidationEnable"= "Enabled",
			"ActiveDirectory.1.DCLookupEnable"= "Enabled",
			"ActiveDirectory.1.RacDomain"= "test",
			"ActiveDirectory.1.RacName"= "test",
			"ActiveDirectory.1.SSOEnable"= "Disabled",
			"ActiveDirectory.1.Schema"= "Extended Schema",
			"UserDomain.1.Name"= "yulan.pie.lab.emc.com",
			"ActiveDirectory.1.DCLookupByUserDomain"="Enabled",          
		}
	}
	  `,
		testingInfo.Username,
		testingInfo.Password,
		testingInfo.Endpoint,
	)
}

func testAccRedfishDirectoryServiceAuthProviderLDAPConfig(testingInfo TestingServerCredentials) string {
	return fmt.Sprintf(`
	resource "redfish_directory_service_auth_provider" "ds_auth" {
		redfish_server {
		  user         = "%s"
		  password     = "%s"
		  endpoint     = "%s"
		  ssl_insecure = true
		}
		ldap = {
			directory = {
				remote_role_mapping = [
					{
						local_role = "Administrator",
						remote_group = "cn = idracgroup,cn = users,dc = yulan,dc = pie,dc = lab,dc = emc,dc = com"
					}        
				],
				service_addresses = [
					"yulanadhost12.yulan.pie.lab.emc.com"
				],
				service_enabled = false
			},
			ldap_service = {
				search_settings = {
					base_distinguished_names = [
						  "dc = yulan,dc = pie,dc = lab,dc = emc,dc = com"
					],
					group_name_attribute = "name",
					user_name_attribute = "member"
				}
			}
		}
		
		 ldap_attributes = {
		  "LDAP.1.GroupAttributeIsDN" = "Enabled"
		  "LDAP.1.Port" = "636",
		  "LDAP.1.BindDN" = "cn = adtester,cn = users,dc = yulan,dc = pie,dc = lab,dc = emc,dc = com",
		  "LDAP.1.BindPassword" = "",
		  "LDAP.1.SearchFilter" = "(objectclass = *)"
		  }
		}
	  `,
		testingInfo.Username,
		testingInfo.Password,
		testingInfo.Endpoint,
	)
}

func testAccRedfishDirectoryServiceAuthProviderLDAP17GConfig(testingInfo TestingServerCredentials) string {
	return fmt.Sprintf(`
	resource "redfish_directory_service_auth_provider" "ds_auth" {
		redfish_server {
		  user         = "%s"
		  password     = "%s"
		  endpoint     = "%s"
		  ssl_insecure = true
		}
		ldap = {
			directory = {
				remote_role_mapping = [
					{
						local_role = "Administrator",
						remote_group = "cn = idracgroup,cn = users,dc = yulan,dc = pie,dc = lab,dc = emc,dc = com"
					}        
				],
				#service_addresses = [
				#	"yulanadhost.yulan.pie.lab.emc.com"
				#],
				service_enabled = false
			},
			ldap_service = {
				search_settings = {
					base_distinguished_names = [
						  "dc = yulan,dc = pie,dc = lab,dc = emc,dc = com"
					],
					group_name_attribute = "name",
					user_name_attribute = "member"
				}
			}
		}
		
		 ldap_attributes = {
		  "LDAP.1.GroupAttributeIsDN" = "Enabled"
		  "LDAP.1.Port" = "636",
		  "LDAP.1.BindDN" = "cn = adtester,cn = users,dc = yulan,dc = pie,dc = lab,dc = emc,dc = com",
		  "LDAP.1.BindPassword" = "",
		  "LDAP.1.SearchFilter" = "(objectclass = *)",
		  "LDAP.1.Server" = "yulanadhost.yulan.pie.lab.emc.com"
		  }
		}
	  `,
		testingInfo.Username,
		testingInfo.Password,
		testingInfo.Endpoint,
	)
}

func testAccRedfishDirectoryServiceAuthProviderLDAP_ServiceErrorConfig(testingInfo TestingServerCredentials) string {
	return fmt.Sprintf(`
	resource "redfish_directory_service_auth_provider" "ds_auth" {
		redfish_server {
		  user         = "%s"
		  password     = "%s"
		  endpoint     = "%s"
		  ssl_insecure = true
		}
		ldap = {
			directory = {
				remote_role_mapping = [
					{
						local_role = "Administrator",
						remote_group = "cn = idracgroup,cn = users,dc = yulan,dc = pie,dc = lab,dc = emc,dc = com"
					}        
				],
				service_addresses = [
					"yulanadhost12.yulan.pie.lab.emc.com",
					"yulanadhost12.yulan.pie.lab.emc.com"
				],
				service_enabled = false
			},
			ldap_service = {
				search_settings = {
					base_distinguished_names = [
						  "dc = yulan,dc = pie,dc = lab,dc = emc,dc = com"
					],
					group_name_attribute = "name",
					user_name_attribute = "member"
				}
			}
		}
		
		 ldap_attributes = {
		  "LDAP.1.GroupAttributeIsDN" = "Enabled"
		  "LDAP.1.Port" = "636",
		  "LDAP.1.BindDN" = "cn = adtester,cn = users,dc = yulan,dc = pie,dc = lab,dc = emc,dc = com",
		  "LDAP.1.BindPassword" = "",
		  "LDAP.1.SearchFilter" = "(objectclass = *)"
		  }
		}
	  `,
		testingInfo.Username,
		testingInfo.Password,
		testingInfo.Endpoint,
	)
}

func testAccRedfishDirectoryServiceAuthProviderAD_UpdateConfig(testingInfo TestingServerCredentials) string {
	return fmt.Sprintf(`
	resource "redfish_directory_service_auth_provider" "ds_auth" {
		redfish_server {
		  user         = "%s"
		  password     = "%s"
		  endpoint     = "%s"
		  ssl_insecure = true
		}
	  
		active_directory = {
			directory = {
				service_enabled = false,
				authentication = {
					kerberos_key_tab_file = ""
				}
			}
		}
		
		active_directory_attributes = {
			"ActiveDirectory.1.AuthTimeout"= "130",
			"ActiveDirectory.1.CertValidationEnable"= "Enabled",
			"ActiveDirectory.1.DCLookupEnable"= "Enabled",
			"ActiveDirectory.1.RacDomain"= "test",
			"ActiveDirectory.1.RacName"= "test",
			"ActiveDirectory.1.SSOEnable"= "Disabled",
			"ActiveDirectory.1.Schema"= "Extended Schema",
			"UserDomain.1.Name"= "yulan.pie.lab.emc.com",
			"ActiveDirectory.1.DCLookupByUserDomain"="Enabled",          
		}
	}
	  `,
		testingInfo.Username,
		testingInfo.Password,
		testingInfo.Endpoint,
	)
}

func testAccRedfishDirectoryServiceAuthProviderAD_Update17GConfig(testingInfo TestingServerCredentials) string {
	return fmt.Sprintf(`
	resource "redfish_directory_service_auth_provider" "ds_auth" {
		redfish_server {
		  user         = "%s"
		  password     = "%s"
		  endpoint     = "%s"
		  ssl_insecure = true
		}
	  
		active_directory = {
			directory = {
				remote_role_mapping = [
					{
						local_role = "Administrator",
						remote_group = "xxxx"
					},
					{
						local_role = "Operator",
						remote_group = "abcd"
					}
				],
				#service_addresses = [
				#	"yulanadhost1.yulan.pie.lab.emc.com",
				#	"yulanadhost.yulan.pie.lab.emc.com",
				#	"yulanadhost2.yulan.pie.lab.emc.com"
				#],
				service_enabled = true
			}
		}
		
		active_directory_attributes = {
			"ActiveDirectory.1.AuthTimeout"= "130",
			"ActiveDirectory.1.CertValidationEnable"= "Enabled",
			"ActiveDirectory.1.DCLookupEnable"= "Disabled",
			"UserDomain.1.Name"= "yulan.pie.lab.emc.com",
			"UserDomain.2.Name"= "yulan2.pie.lab.emc.com",
			"UserDomain.3.Name"= "yulan3.pie.lab.emc.com",
			"ActiveDirectory.1.GCLookupEnable" = "Disabled",
			"ActiveDirectory.1.GlobalCatalog1" = "yulanadhost21.yulan.pie.lab.emc.com", 
			"ActiveDirectory.1.DomainController1": "yulanadhost1.yulan.pie.lab.emc.com",        
		}
	}
	  `,
		testingInfo.Username,
		testingInfo.Password,
		testingInfo.Endpoint,
	)
}

func testAccRedfishDirectoryServiceAuthProviderADWithStandardSchema_UpdateConfig(testingInfo TestingServerCredentials) string {
	return fmt.Sprintf(`
	resource "redfish_directory_service_auth_provider" "ds_auth" {
		redfish_server {
		  user         = "%s"
		  password     = "%s"
		  endpoint     = "%s"
		  ssl_insecure = true
		}
	  
		active_directory = {
			directory = {
				remote_role_mapping = [
					{
						local_role = "Administrator",
						remote_group = "xxxx"
					},
					{
						local_role = "Operator",
						remote_group = "abcd"
					}
				],
				service_addresses = [
					"yulanadhost1.yulan.pie.lab.emc.com",
					"yulanadhost.yulan.pie.lab.emc.com",
					"yulanadhost2.yulan.pie.lab.emc.com"
				 ],
				service_enabled = true
			}
		}
		
		active_directory_attributes = {
			"ActiveDirectory.1.AuthTimeout"= "130",
			"ActiveDirectory.1.CertValidationEnable"= "Enabled",
			"ActiveDirectory.1.DCLookupEnable"= "Disabled",
			"ActiveDirectory.1.SSOEnable"= "Disabled",
			"ActiveDirectory.1.Schema"= "Standard Schema",
			"UserDomain.1.Name"= "yulan.pie.lab.emc.com",
			"UserDomain.2.Name"= "yulan2.pie.lab.emc.com",
			"UserDomain.3.Name"= "yulan3.pie.lab.emc.com",
			#"ActiveDirectory.1.DCLookupByUserDomain"="Enabled", 
			"ActiveDirectory.1.GCLookupEnable" = "Disabled",
			"ActiveDirectory.1.GlobalCatalog1" = "yulanadhost21.yulan.pie.lab.emc.com",         
		}
	}
	  `,
		testingInfo.Username,
		testingInfo.Password,
		testingInfo.Endpoint,
	)
}
func testAccRedfishDirectoryServiceAuthProviderADWithStandardSchema_ServiceErrorConfig(testingInfo TestingServerCredentials) string {
	return fmt.Sprintf(`
	resource "redfish_directory_service_auth_provider" "ds_auth" {
		redfish_server {
		  user         = "%s"
		  password     = "%s"
		  endpoint     = "%s"
		  ssl_insecure = true
		}
	  
		active_directory = {
			directory = {
				remote_role_mapping = [
					{
						local_role = "Administrator",
						remote_group = "xxxx"
					},
					{
						local_role = "Operator",
						remote_group = "abcd"
					}
				],
				service_addresses = [
					"yulanadhost1.yulan.pie.lab.emc.com",
					"yulanadhost.yulan.pie.lab.emc.com",
					"yulanadhost2.yulan.pie.lab.emc.com",
					"yulanadhost2.yulan.pie.lab.emc.com"
				 ],
				service_enabled = true
			}
		}
		
		active_directory_attributes = {
			"ActiveDirectory.1.AuthTimeout"= "130",
			"ActiveDirectory.1.CertValidationEnable"= "Enabled",
			"ActiveDirectory.1.DCLookupEnable"= "Disabled",
			"ActiveDirectory.1.SSOEnable"= "Disabled",
			"ActiveDirectory.1.Schema"= "Standard Schema",
			"UserDomain.1.Name"= "yulan.pie.lab.emc.com",
			"UserDomain.2.Name"= "yulan2.pie.lab.emc.com",
			"UserDomain.3.Name"= "yulan3.pie.lab.emc.com",
			#"ActiveDirectory.1.DCLookupByUserDomain"="Enabled", 
			"ActiveDirectory.1.GCLookupEnable" = "Disabled",
			"ActiveDirectory.1.GlobalCatalog1" = "yulanadhost21.yulan.pie.lab.emc.com",         
		}
	}
	  `,
		testingInfo.Username,
		testingInfo.Password,
		testingInfo.Endpoint,
	)
}

func testAccRedfishDirectoryServiceAuthProviderEmptyAuth(testingInfo TestingServerCredentials) string {
	return fmt.Sprintf(`
	resource "redfish_directory_service_auth_provider" "ds_auth" {
		redfish_server {
		  user         = "%s"
		  password     = "%s"
		  endpoint     = "%s"
		  ssl_insecure = true
		}
	  
		active_directory = {
			directory = {
				
				service_addresses = [
					"yulanadhost11.yulan.pie.lab.emc.com"
				 ],
				service_enabled = true
			}
		}
		
		active_directory_attributes = {
			"ActiveDirectory.1.CertValidationEnable"= "Enabled",
			"ActiveDirectory.1.DCLookupEnable"= "Disabled",
			"ActiveDirectory.1.RacDomain"= "test",
			"ActiveDirectory.1.RacName"= "test",
			"ActiveDirectory.1.SSOEnable"= "Disabled",
			"ActiveDirectory.1.Schema"= "Extended Schema",
			"UserDomain.1.Name"= "yulan.pie.lab.emc.com"           
		}
	}
	  `,
		testingInfo.Username,
		testingInfo.Password,
		testingInfo.Endpoint,
	)
}

func testAccRedfishDirectoryServiceAuthProviderEmptyAuth17GConfig(testingInfo TestingServerCredentials) string {
	return fmt.Sprintf(`
	resource "redfish_directory_service_auth_provider" "ds_auth" {
		redfish_server {
		  user         = "%s"
		  password     = "%s"
		  endpoint     = "%s"
		  ssl_insecure = true
		}
	  
		active_directory = {
			directory = {
				
				#service_addresses = [
				#	"yulanadhost11.yulan.pie.lab.emc.com"
				# ],
				service_enabled = true
			}
		}
		
		active_directory_attributes = {
			"ActiveDirectory.1.CertValidationEnable"= "Enabled",
			"ActiveDirectory.1.DCLookupEnable"= "Disabled",
			"UserDomain.1.Name"= "yulan.pie.lab.emc.com"           
		}
	}
	  `,
		testingInfo.Username,
		testingInfo.Password,
		testingInfo.Endpoint,
	)
}

func testAccRedfishDirectoryServiceAuthProviderInvalidAuthTimeoutString(testingInfo TestingServerCredentials) string {
	return fmt.Sprintf(`
	resource "redfish_directory_service_auth_provider" "ds_auth" {
		redfish_server {
		  user         = "%s"
		  password     = "%s"
		  endpoint     = "%s"
		  ssl_insecure = true
		}
	  
		active_directory = {
			directory = {
				
				service_addresses = [
					"yulanadhost11.yulan.pie.lab.emc.com"
				 ],
				service_enabled = true
			}
		}
		
		active_directory_attributes = {
			"ActiveDirectory.1.AuthTimeout"= "Invalid",
			"ActiveDirectory.1.CertValidationEnable"= "Enabled",
			"ActiveDirectory.1.DCLookupEnable"= "Disabled",
			"ActiveDirectory.1.RacDomain"= "test",
			"ActiveDirectory.1.RacName"= "test",
			"ActiveDirectory.1.SSOEnable"= "Disabled",
			"ActiveDirectory.1.Schema"= "Extended Schema",
			#"ADGroup.1.Domain" = "yulan.pie.lab.emc.com",
			"UserDomain.1.Name"= "yulan.pie.lab.emc.com"           
		}
	}
	  `,
		testingInfo.Username,
		testingInfo.Password,
		testingInfo.Endpoint,
	)
}

func testAccRedfishDirectoryServiceAuthProviderInvalidAuthTimeoutString17GConfig(testingInfo TestingServerCredentials) string {
	return fmt.Sprintf(`
	resource "redfish_directory_service_auth_provider" "ds_auth" {
		redfish_server {
		  user         = "%s"
		  password     = "%s"
		  endpoint     = "%s"
		  ssl_insecure = true
		}
	  
		active_directory = {
			directory = {
				
				#service_addresses = [
				#	"yulanadhost11.yulan.pie.lab.emc.com"
				# ],
				service_enabled = true
			}
		}
		
		active_directory_attributes = {
			"ActiveDirectory.1.AuthTimeout"= "Invalid",
			"ActiveDirectory.1.CertValidationEnable"= "Enabled",
			"ActiveDirectory.1.DCLookupEnable"= "Disabled",
			#"ADGroup.1.Domain" = "yulan.pie.lab.emc.com",
			"UserDomain.1.Name"= "yulan.pie.lab.emc.com"           
		}
	}
	  `,
		testingInfo.Username,
		testingInfo.Password,
		testingInfo.Endpoint,
	)
}

func testAccRedfishDirectoryServiceAuthProviderInvalidAuth(testingInfo TestingServerCredentials) string {
	return fmt.Sprintf(`
	resource "redfish_directory_service_auth_provider" "ds_auth" {
		redfish_server {
		  user         = "%s"
		  password     = "%s"
		  endpoint     = "%s"
		  ssl_insecure = true
		}
	  
		active_directory = {
			directory = {
				
				service_addresses = [
					"yulanadhost11.yulan.pie.lab.emc.com"
				 ],
				service_enabled = true
			}
		}
		
		active_directory_attributes = {
			"ActiveDirectory.1.AuthTimeout"= "12",
			"ActiveDirectory.1.CertValidationEnable"= "Enabled",
			"ActiveDirectory.1.DCLookupEnable"= "Disabled",
			"ActiveDirectory.1.RacDomain"= "test",
			"ActiveDirectory.1.RacName"= "test",
			"ActiveDirectory.1.SSOEnable"= "Disabled",
			"ActiveDirectory.1.Schema"= "Extended Schema",
			#"ADGroup.1.Domain" = "yulan.pie.lab.emc.com",
			"UserDomain.1.Name"= "yulan.pie.lab.emc.com"           
		}
	}
	  `,
		testingInfo.Username,
		testingInfo.Password,
		testingInfo.Endpoint,
	)
}

func testAccRedfishDirectoryServiceAuthProviderInvalidAuth17GConfig(testingInfo TestingServerCredentials) string {
	return fmt.Sprintf(`
	resource "redfish_directory_service_auth_provider" "ds_auth" {
		redfish_server {
		  user         = "%s"
		  password     = "%s"
		  endpoint     = "%s"
		  ssl_insecure = true
		}
	  
		active_directory = {
			directory = {
				
				#service_addresses = [
				#	"yulanadhost11.yulan.pie.lab.emc.com"
				# ],
				service_enabled = true
			}
		}
		
		active_directory_attributes = {
			"ActiveDirectory.1.AuthTimeout"= "12",
			"ActiveDirectory.1.CertValidationEnable"= "Enabled",
			"ActiveDirectory.1.DCLookupEnable"= "Disabled",
			#"ADGroup.1.Domain" = "yulan.pie.lab.emc.com",
			"UserDomain.1.Name"= "yulan.pie.lab.emc.com"           
		}
	}
	  `,
		testingInfo.Username,
		testingInfo.Password,
		testingInfo.Endpoint,
	)
}

func testAccRedfishDirectoryServiceAuthProviderADDisableSSOEnable(testingInfo TestingServerCredentials) string {
	return fmt.Sprintf(`
	resource "redfish_directory_service_auth_provider" "ds_auth" {
		redfish_server {
		  user         = "%s"
		  password     = "%s"
		  endpoint     = "%s"
		  ssl_insecure = true
		}
	  
		active_directory = {
			directory = {
				
				service_addresses = [
					"yulanadhost11.yulan.pie.lab.emc.com"
				 ],
				service_enabled = false
			}
		}
		
		active_directory_attributes = {
			"ActiveDirectory.1.AuthTimeout"= "120",
			"ActiveDirectory.1.CertValidationEnable"= "Enabled",
			"ActiveDirectory.1.DCLookupEnable"= "Disabled",
			"ActiveDirectory.1.RacDomain"= "test",
			"ActiveDirectory.1.RacName"= "test",
			"ActiveDirectory.1.SSOEnable"= "Enabled",
			"ActiveDirectory.1.Schema"= "Extended Schema",
			#"ADGroup.1.Domain" = "yulan.pie.lab.emc.com",
			"UserDomain.1.Name"= "yulan.pie.lab.emc.com"           
		}
	}
	  `,
		testingInfo.Username,
		testingInfo.Password,
		testingInfo.Endpoint,
	)
}

func testAccRedfishDirectoryServiceAuthProviderADEnSSOEnNoKb(testingInfo TestingServerCredentials) string {
	return fmt.Sprintf(`
	resource "redfish_directory_service_auth_provider" "ds_auth" {
		redfish_server {
		  user         = "%s"
		  password     = "%s"
		  endpoint     = "%s"
		  ssl_insecure = true
		}
	  
		active_directory = {
			directory = {
				
				service_addresses = [
					"yulanadhost11.yulan.pie.lab.emc.com"
				 ],
				service_enabled = true
			}
		}
		
		active_directory_attributes = {
			"ActiveDirectory.1.AuthTimeout"= "120",
			"ActiveDirectory.1.CertValidationEnable"= "Enabled",
			"ActiveDirectory.1.DCLookupEnable"= "Disabled",
			"ActiveDirectory.1.RacDomain"= "test",
			"ActiveDirectory.1.RacName"= "test",
			"ActiveDirectory.1.SSOEnable"= "Enabled",
			"ActiveDirectory.1.Schema"= "Extended Schema",
			"UserDomain.1.Name"= "yulan.pie.lab.emc.com"           
		}
	}
	  `,
		testingInfo.Username,
		testingInfo.Password,
		testingInfo.Endpoint,
	)
}

func testAccRedfishDirectoryServiceAuthProviderDClookUpEnServiceAddEmpty(testingInfo TestingServerCredentials) string {
	return fmt.Sprintf(`
	resource "redfish_directory_service_auth_provider" "ds_auth" {
		redfish_server {
		  user         = "%s"
		  password     = "%s"
		  endpoint     = "%s"
		  ssl_insecure = true
		}
	  
		active_directory = {
			directory = {
				service_enabled = true,
			}
		}
		
		active_directory_attributes = {
			"ActiveDirectory.1.AuthTimeout"= "120",
			"ActiveDirectory.1.CertValidationEnable"= "Enabled",
			"ActiveDirectory.1.DCLookupEnable"= "Disabled",
			"ActiveDirectory.1.RacDomain"= "test",
			"ActiveDirectory.1.RacName"= "test",
			"ActiveDirectory.1.SSOEnable"= "Disabled",
			"ActiveDirectory.1.Schema"= "Extended Schema",
			"UserDomain.1.Name"= "yulan.pie.lab.emc.com"           
		}
	}
	  `,
		testingInfo.Username,
		testingInfo.Password,
		testingInfo.Endpoint,
	)
}

func testAccRedfishDirectoryServiceAuthProviderDClookUpDsServiceAddNonEmpty17GConfig(testingInfo TestingServerCredentials) string {
	return fmt.Sprintf(`
	resource "redfish_directory_service_auth_provider" "ds_auth" {
		redfish_server {
		  user         = "%s"
		  password     = "%s"
		  endpoint     = "%s"
		  ssl_insecure = true
		}
	  
		active_directory = {
			directory = {
				service_enabled = false,
				service_addresses = [
				"yulanadhost12.yulan.pie.lab.emc.com"
			]
			}
		}
		
		active_directory_attributes = {
			"ActiveDirectory.1.AuthTimeout"= "130",
			"ActiveDirectory.1.CertValidationEnable"= "Enabled",
			"ActiveDirectory.1.DCLookupEnable"= "Disabled",
			"UserDomain.1.Name"= "yulan.pie.lab.emc.com"           
		}
	}
	  `,
		testingInfo.Username,
		testingInfo.Password,
		testingInfo.Endpoint,
	)
}

func testAccRedfishDirectoryServiceAuthProviderDClookUpDsServiceAddEmpty17GConfig(testingInfo TestingServerCredentials) string {
	return fmt.Sprintf(`
	resource "redfish_directory_service_auth_provider" "ds_auth" {
		redfish_server {
		  user         = "%s"
		  password     = "%s"
		  endpoint     = "%s"
		  ssl_insecure = true
		}
	  
		active_directory = {
			directory = {
				service_enabled = true,
			}
		}
		
		active_directory_attributes = {
			"ActiveDirectory.1.AuthTimeout"= "120",
			"ActiveDirectory.1.CertValidationEnable"= "Enabled",
			"ActiveDirectory.1.DCLookupEnable"= "Disabled",
			"UserDomain.1.Name"= "yulan.pie.lab.emc.com"           
		}
	}
	  `,
		testingInfo.Username,
		testingInfo.Password,
		testingInfo.Endpoint,
	)
}

func testAccRedfishDirectoryServiceAuthProviderLDAPServiceAddNonEmpty17GConfig(testingInfo TestingServerCredentials) string {
	return fmt.Sprintf(`
	resource "redfish_directory_service_auth_provider" "ds_auth" {
		redfish_server {
		  user         = "%s"
		  password     = "%s"
		  endpoint     = "%s"
		  ssl_insecure = true
		}
		ldap = {
			directory = {
				remote_role_mapping = [
					{
						local_role = "Administrator",
						remote_group = "cn = idracgroup,cn = users,dc = yulan,dc = pie,dc = lab,dc = emc,dc = com"
					}        
				],
				service_addresses = [
					"yulanadhost12.yulan.pie.lab.emc.com"
				],
				service_enabled = false
			},
			ldap_service = {
				search_settings = {
					base_distinguished_names = [
						  "dc = yulan,dc = pie,dc = lab,dc = emc,dc = com"
					],
					group_name_attribute = "name",
					user_name_attribute = "member"
				}
			}
		}
		
		 ldap_attributes = {
		  "LDAP.1.GroupAttributeIsDN" = "Enabled"
		  "LDAP.1.Port" = "636",
		  "LDAP.1.BindDN" = "cn = adtester,cn = users,dc = yulan,dc = pie,dc = lab,dc = emc,dc = com",
		  "LDAP.1.BindPassword" = "",
		  "LDAP.1.SearchFilter" = "(objectclass = *)"
		  }
		}
	  `,
		testingInfo.Username,
		testingInfo.Password,
		testingInfo.Endpoint,
	)
}

func testAccRedfishDirectoryServiceAuthProviderLDAPServiceAddEmptyConfig(testingInfo TestingServerCredentials) string {
	return fmt.Sprintf(`
	resource "redfish_directory_service_auth_provider" "ds_auth" {
		redfish_server {
		  user         = "%s"
		  password     = "%s"
		  endpoint     = "%s"
		  ssl_insecure = true
		}
		ldap = {
			directory = {
				remote_role_mapping = [
					{
						local_role = "Administrator",
						remote_group = "cn = idracgroup,cn = users,dc = yulan,dc = pie,dc = lab,dc = emc,dc = com"
					}        
				],
				#service_addresses = [
				#	"yulanadhost12.yulan.pie.lab.emc.com"
				#],
				service_enabled = false
			},
			ldap_service = {
				search_settings = {
					base_distinguished_names = [
						  "dc = yulan,dc = pie,dc = lab,dc = emc,dc = com"
					],
					group_name_attribute = "name",
					user_name_attribute = "member"
				}
			}
		}
		
		 ldap_attributes = {
		  "LDAP.1.GroupAttributeIsDN" = "Enabled"
		  "LDAP.1.Port" = "636",
		  "LDAP.1.BindDN" = "cn = adtester,cn = users,dc = yulan,dc = pie,dc = lab,dc = emc,dc = com",
		  "LDAP.1.BindPassword" = "",
		  "LDAP.1.SearchFilter" = "(objectclass = *)"
		  }
		}
	  `,
		testingInfo.Username,
		testingInfo.Password,
		testingInfo.Endpoint,
	)
}

func testAccRedfishDirectoryServiceAuthProviderLDAPServerAddressEmpty17GConfig(testingInfo TestingServerCredentials) string {
	return fmt.Sprintf(`
	resource "redfish_directory_service_auth_provider" "ds_auth" {
		redfish_server {
		  user         = "%s"
		  password     = "%s"
		  endpoint     = "%s"
		  ssl_insecure = true
		}
		ldap = {
			directory = {
				remote_role_mapping = [
					{
						local_role = "Administrator",
						remote_group = "cn = idracgroup,cn = users,dc = yulan,dc = pie,dc = lab,dc = emc,dc = com"
					}        
				],
				service_enabled = false
			},
			ldap_service = {
				search_settings = {
					base_distinguished_names = [
						  "dc = yulan,dc = pie,dc = lab,dc = emc,dc = com"
					],
					group_name_attribute = "name",
					user_name_attribute = "member"
				}
			}
		}
		
		 ldap_attributes = {
		  "LDAP.1.GroupAttributeIsDN" = "Enabled"
		  "LDAP.1.Port" = "636",
		  "LDAP.1.BindDN" = "cn = adtester,cn = users,dc = yulan,dc = pie,dc = lab,dc = emc,dc = com",
		  "LDAP.1.BindPassword" = "",
		  "LDAP.1.SearchFilter" = "(objectclass = *)"
		  }
		}
	  `,
		testingInfo.Username,
		testingInfo.Password,
		testingInfo.Endpoint,
	)
}

func testAccRedfishDirectoryServiceAuthProviderDCLookupByUserDomainConfig(testingInfo TestingServerCredentials) string {
	return fmt.Sprintf(`
	resource "redfish_directory_service_auth_provider" "ds_auth" {
		redfish_server {
		  user         = "%s"
		  password     = "%s"
		  endpoint     = "%s"
		  ssl_insecure = true
		}
	  
		active_directory = {
			directory = {
				service_addresses = [
					"yulanadhost12.yulan.pie.lab.emc.com"
				],
				service_enabled = true,
				authentication = {
					kerberos_key_tab_file = ""
				}
			}
		}
		
		active_directory_attributes = {
			"ActiveDirectory.1.AuthTimeout"= "120",
			"ActiveDirectory.1.CertValidationEnable"= "Enabled",
			"ActiveDirectory.1.DCLookupEnable"= "Disabled",
			"ActiveDirectory.1.RacDomain"= "test",
			"ActiveDirectory.1.RacName"= "test",
			"ActiveDirectory.1.SSOEnable"= "Disabled",
			"ActiveDirectory.1.Schema"= "Extended Schema",
			"UserDomain.1.Name"= "yulan.pie.lab.emc.com",
			"ActiveDirectory.1.DCLookupByUserDomain"="Disabled",           
		}
	}
	  `,
		testingInfo.Username,
		testingInfo.Password,
		testingInfo.Endpoint,
	)
}

func testAccRedfishDirectoryServiceAuthProviderDCLookupByUserDomainConfig17GConfig(testingInfo TestingServerCredentials) string {
	return fmt.Sprintf(`
	resource "redfish_directory_service_auth_provider" "ds_auth" {
		redfish_server {
		  user         = "%s"
		  password     = "%s"
		  endpoint     = "%s"
		  ssl_insecure = true
		}
	  
		active_directory = {
			directory = {
				#service_addresses = [
				#	"yulanadhost12.yulan.pie.lab.emc.com"
				#],
				service_enabled = true,
			}
		}
		
		active_directory_attributes = {
			"ActiveDirectory.1.AuthTimeout"= "120",
			"ActiveDirectory.1.CertValidationEnable"= "Enabled",
			"ActiveDirectory.1.DCLookupEnable"= "Disabled",
			"ActiveDirectory.1.DomainController1": "yulanadhost1.yulan.pie.lab.emc.com",
			"UserDomain.1.Name"= "yulan.pie.lab.emc.com",
			"ActiveDirectory.1.DCLookupByUserDomain"="Disabled",           
		}
	}
	  `,
		testingInfo.Username,
		testingInfo.Password,
		testingInfo.Endpoint,
	)
}
func testAccRedfishDirectoryServiceAuthProviderDCLookupDomainNameConfig(testingInfo TestingServerCredentials) string {
	return fmt.Sprintf(`
	resource "redfish_directory_service_auth_provider" "ds_auth" {
		redfish_server {
		  user         = "%s"
		  password     = "%s"
		  endpoint     = "%s"
		  ssl_insecure = true
		}
	  
		active_directory = {
			directory = {
				service_addresses = [
					"yulanadhost12.yulan.pie.lab.emc.com"
				],
				service_enabled = true,
				authentication = {
					kerberos_key_tab_file = ""
				}
			}
		}
		
		active_directory_attributes = {
			"ActiveDirectory.1.AuthTimeout"= "120",
			"ActiveDirectory.1.CertValidationEnable"= "Enabled",
			"ActiveDirectory.1.DCLookupEnable"= "Disabled",
			"ActiveDirectory.1.RacDomain"= "test",
			"ActiveDirectory.1.RacName"= "test",
			"ActiveDirectory.1.SSOEnable"= "Disabled",
			"ActiveDirectory.1.Schema"= "Extended Schema",
			"UserDomain.1.Name"= "yulan.pie.lab.emc.com",
			#"ActiveDirectory.1.DCLookupByUserDomain"="Disabled", 
			"ActiveDirectory.1.DCLookupDomainName"="test",          
		}
	}
	  `,
		testingInfo.Username,
		testingInfo.Password,
		testingInfo.Endpoint,
	)
}

func testAccRedfishDirectoryServiceAuthProviderDCLookupDomainName17GConfig(testingInfo TestingServerCredentials) string {
	return fmt.Sprintf(`
	resource "redfish_directory_service_auth_provider" "ds_auth" {
		redfish_server {
		  user         = "%s"
		  password     = "%s"
		  endpoint     = "%s"
		  ssl_insecure = true
		}
	  
		active_directory = {
			directory = {
				#service_addresses = [
				#	"yulanadhost12.yulan.pie.lab.emc.com"
				#],
				service_enabled = true,
			}
		}
		
		active_directory_attributes = {
			"ActiveDirectory.1.AuthTimeout"= "120",
			"ActiveDirectory.1.CertValidationEnable"= "Enabled",
			"ActiveDirectory.1.DCLookupEnable"= "Disabled",
			"ActiveDirectory.1.DomainController1": "yulanadhost1.yulan.pie.lab.emc.com",
			"UserDomain.1.Name"= "yulan.pie.lab.emc.com",
			"ActiveDirectory.1.DCLookupDomainName"="test",          
		}
	}
	  `,
		testingInfo.Username,
		testingInfo.Password,
		testingInfo.Endpoint,
	)
}

func testAccRedfishDirectoryServiceAuthProviderDDCLookupEnableNoServiceAddConfig(testingInfo TestingServerCredentials) string {
	return fmt.Sprintf(`
	resource "redfish_directory_service_auth_provider" "ds_auth" {
		redfish_server {
		  user         = "%s"
		  password     = "%s"
		  endpoint     = "%s"
		  ssl_insecure = true
		}
	  
		active_directory = {
			directory = {
				service_addresses = [
            "yulanadhost12.yulan.pie.lab.emc.com"
        ],
				service_enabled = true,
				authentication = {
					kerberos_key_tab_file = ""
				}
			}
		}
		
		active_directory_attributes = {
			"ActiveDirectory.1.AuthTimeout"= "120",
			"ActiveDirectory.1.CertValidationEnable"= "Enabled",
			"ActiveDirectory.1.DCLookupEnable"= "Enabled",
			"ActiveDirectory.1.RacDomain"= "test",
			"ActiveDirectory.1.RacName"= "test",
			"ActiveDirectory.1.SSOEnable"= "Disabled",
			"ActiveDirectory.1.Schema"= "Extended Schema",
			"UserDomain.1.Name"= "yulan.pie.lab.emc.com",
			"ActiveDirectory.1.DCLookupByUserDomain"="Disabled", 
			"ActiveDirectory.1.DCLookupDomainName"="test",          
		}
	}
	  `,
		testingInfo.Username,
		testingInfo.Password,
		testingInfo.Endpoint,
	)
}

func testAccRedfishDirectoryServiceAuthProviderDDCLookupEnableNoServiceAdd17GConfig(testingInfo TestingServerCredentials) string {
	return fmt.Sprintf(`
	resource "redfish_directory_service_auth_provider" "ds_auth" {
		redfish_server {
		  user         = "%s"
		  password     = "%s"
		  endpoint     = "%s"
		  ssl_insecure = true
		}
	  
		active_directory = {
			directory = {
				#service_addresses = [
                #  "yulanadhost12.yulan.pie.lab.emc.com"
                #],
				service_enabled = true,
			}
		}
		
		active_directory_attributes = {
			"ActiveDirectory.1.AuthTimeout"= "120",
			"ActiveDirectory.1.CertValidationEnable"= "Enabled",
			"ActiveDirectory.1.DCLookupEnable"= "Enabled",
			"ActiveDirectory.1.DomainController1": "yulanadhost1.yulan.pie.lab.emc.com",
			"UserDomain.1.Name"= "yulan.pie.lab.emc.com",
			"ActiveDirectory.1.DCLookupByUserDomain"="Disabled", 
			"ActiveDirectory.1.DCLookupDomainName"="test",          
		}
	}
	  `,
		testingInfo.Username,
		testingInfo.Password,
		testingInfo.Endpoint,
	)
}

func testAccRedfishDirectoryServiceAuthProviderDCLookupByUserDomainEmptyConfig(testingInfo TestingServerCredentials) string {
	return fmt.Sprintf(`
	resource "redfish_directory_service_auth_provider" "ds_auth" {
		redfish_server {
		  user         = "%s"
		  password     = "%s"
		  endpoint     = "%s"
		  ssl_insecure = true
		}
	  
		active_directory = {
			directory = {
				service_enabled = true,
				authentication = {
					kerberos_key_tab_file = ""
				}
			}
		}
		
		active_directory_attributes = {
			"ActiveDirectory.1.AuthTimeout"= "120",
			"ActiveDirectory.1.CertValidationEnable"= "Enabled",
			"ActiveDirectory.1.DCLookupEnable"= "Enabled",
			"ActiveDirectory.1.RacDomain"= "test",
			"ActiveDirectory.1.RacName"= "test",
			"ActiveDirectory.1.SSOEnable"= "Disabled",
			"ActiveDirectory.1.Schema"= "Extended Schema",
			"UserDomain.1.Name"= "yulan.pie.lab.emc.com", 
			"ActiveDirectory.1.DCLookupDomainName"="test",          
		}
	}
	  `,
		testingInfo.Username,
		testingInfo.Password,
		testingInfo.Endpoint,
	)
}

func testAccRedfishDirectoryServiceAuthProviderDCLookupByUserDomainEmpty17GConfig(testingInfo TestingServerCredentials) string {
	return fmt.Sprintf(`
	resource "redfish_directory_service_auth_provider" "ds_auth" {
		redfish_server {
		  user         = "%s"
		  password     = "%s"
		  endpoint     = "%s"
		  ssl_insecure = true
		}
	  
		active_directory = {
			directory = {
				service_enabled = true,
			}
		}
		
		active_directory_attributes = {
			"ActiveDirectory.1.AuthTimeout"= "120",
			"ActiveDirectory.1.CertValidationEnable"= "Enabled",
			"ActiveDirectory.1.DCLookupEnable"= "Enabled",
			"UserDomain.1.Name"= "yulan.pie.lab.emc.com", 
			"ActiveDirectory.1.DCLookupDomainName"="test",          
		}
	}
	  `,
		testingInfo.Username,
		testingInfo.Password,
		testingInfo.Endpoint,
	)
}

func testAccRedfishDirectoryServiceAuthProviderWithoutDCLookupDomainNameConfig(testingInfo TestingServerCredentials) string {
	return fmt.Sprintf(`
	resource "redfish_directory_service_auth_provider" "ds_auth" {
		redfish_server {
		  user         = "%s"
		  password     = "%s"
		  endpoint     = "%s"
		  ssl_insecure = true
		}
	  
		active_directory = {
			directory = {
				service_enabled = true,
				authentication = {
					kerberos_key_tab_file = ""
				}
			}
		}
		
		active_directory_attributes = {
			"ActiveDirectory.1.AuthTimeout"= "120",
			"ActiveDirectory.1.CertValidationEnable"= "Enabled",
			"ActiveDirectory.1.DCLookupEnable"= "Enabled",
			"ActiveDirectory.1.RacDomain"= "test",
			"ActiveDirectory.1.RacName"= "test",
			"ActiveDirectory.1.SSOEnable"= "Disabled",
			"ActiveDirectory.1.Schema"= "Extended Schema",
			"UserDomain.1.Name"= "yulan.pie.lab.emc.com",
			"ActiveDirectory.1.DCLookupByUserDomain"="Disabled", 
			        
		}
	}
	  `,
		testingInfo.Username,
		testingInfo.Password,
		testingInfo.Endpoint,
	)
}

func testAccRedfishDirectoryServiceAuthProviderWithoutDCLookupDomainName17GConfig(testingInfo TestingServerCredentials) string {
	return fmt.Sprintf(`
	resource "redfish_directory_service_auth_provider" "ds_auth" {
		redfish_server {
		  user         = "%s"
		  password     = "%s"
		  endpoint     = "%s"
		  ssl_insecure = true
		}
	  
		active_directory = {
			directory = {
				service_enabled = true,
			}
		}
		
		active_directory_attributes = {
			"ActiveDirectory.1.AuthTimeout"= "120",
			"ActiveDirectory.1.CertValidationEnable"= "Enabled",
			"ActiveDirectory.1.DCLookupEnable"= "Enabled",
			"UserDomain.1.Name"= "yulan.pie.lab.emc.com",
			"ActiveDirectory.1.DCLookupByUserDomain"="Disabled", 
			        
		}
	}
	  `,
		testingInfo.Username,
		testingInfo.Password,
		testingInfo.Endpoint,
	)
}

func testAccRedfishDirectoryServiceAuthProviderDCLookupDomainNameEmptyConfig(testingInfo TestingServerCredentials) string {
	return fmt.Sprintf(`
	resource "redfish_directory_service_auth_provider" "ds_auth" {
		redfish_server {
		  user         = "%s"
		  password     = "%s"
		  endpoint     = "%s"
		  ssl_insecure = true
		}
	  
		active_directory = {
			directory = {
				service_enabled = true,
				authentication = {
					kerberos_key_tab_file = ""
				}
			}
		}
		
		active_directory_attributes = {
			"ActiveDirectory.1.AuthTimeout"= "120",
			"ActiveDirectory.1.CertValidationEnable"= "Enabled",
			"ActiveDirectory.1.DCLookupEnable"= "Enabled",
			"ActiveDirectory.1.RacDomain"= "test",
			"ActiveDirectory.1.RacName"= "test",
			"ActiveDirectory.1.SSOEnable"= "Disabled",
			"ActiveDirectory.1.Schema"= "Extended Schema",
			"UserDomain.1.Name"= "yulan.pie.lab.emc.com",
			"ActiveDirectory.1.DCLookupByUserDomain"="Disabled", 
			"ActiveDirectory.1.DCLookupDomainName"="",        
		}
	}
	  `,
		testingInfo.Username,
		testingInfo.Password,
		testingInfo.Endpoint,
	)
}

func testAccRedfishDirectoryServiceAuthProviderDCLookupDomainNameEmpty17GConfig(testingInfo TestingServerCredentials) string {
	return fmt.Sprintf(`
	resource "redfish_directory_service_auth_provider" "ds_auth" {
		redfish_server {
		  user         = "%s"
		  password     = "%s"
		  endpoint     = "%s"
		  ssl_insecure = true
		}
	  
		active_directory = {
			directory = {
				service_enabled = true,
			}
		}
		
		active_directory_attributes = {
			"ActiveDirectory.1.AuthTimeout"= "120",
			"ActiveDirectory.1.CertValidationEnable"= "Enabled",
			"ActiveDirectory.1.DCLookupEnable"= "Enabled",
			"UserDomain.1.Name"= "yulan.pie.lab.emc.com",
			"ActiveDirectory.1.DCLookupByUserDomain"="Disabled", 
			"ActiveDirectory.1.DCLookupDomainName"="",        
		}
	}
	  `,
		testingInfo.Username,
		testingInfo.Password,
		testingInfo.Endpoint,
	)
}

func testAccRedfishDirectoryServiceAuthProviderDCLookupEnableDCLookupDomainNameConfig(testingInfo TestingServerCredentials) string {
	return fmt.Sprintf(`
	resource "redfish_directory_service_auth_provider" "ds_auth" {
		redfish_server {
		  user         = "%s"
		  password     = "%s"
		  endpoint     = "%s"
		  ssl_insecure = true
		}
	  
		active_directory = {
			directory = {
				service_enabled = true,
				authentication = {
					kerberos_key_tab_file = ""
				}
			}
		}
		
		active_directory_attributes = {
			"ActiveDirectory.1.AuthTimeout"= "120",
			"ActiveDirectory.1.CertValidationEnable"= "Enabled",
			"ActiveDirectory.1.DCLookupEnable"= "Enabled",
			"ActiveDirectory.1.RacDomain"= "test",
			"ActiveDirectory.1.RacName"= "test",
			"ActiveDirectory.1.SSOEnable"= "Disabled",
			"ActiveDirectory.1.Schema"= "Extended Schema",
			"UserDomain.1.Name"= "yulan.pie.lab.emc.com",
			"ActiveDirectory.1.DCLookupByUserDomain"="Enabled", 
			"ActiveDirectory.1.DCLookupDomainName"="test",          
		}
	}
	  `,
		testingInfo.Username,
		testingInfo.Password,
		testingInfo.Endpoint,
	)
}

func testAccRedfishDirectoryServiceAuthProviderDCLookupEnableDCLookupDomainName17GConfig(testingInfo TestingServerCredentials) string {
	return fmt.Sprintf(`
	resource "redfish_directory_service_auth_provider" "ds_auth" {
		redfish_server {
		  user         = "%s"
		  password     = "%s"
		  endpoint     = "%s"
		  ssl_insecure = true
		}
	  
		active_directory = {
			directory = {
				service_enabled = true,
			}
		}
		
		active_directory_attributes = {
			"ActiveDirectory.1.AuthTimeout"= "120",
			"ActiveDirectory.1.CertValidationEnable"= "Enabled",
			"ActiveDirectory.1.DCLookupEnable"= "Enabled",
			"UserDomain.1.Name"= "yulan.pie.lab.emc.com",
			"ActiveDirectory.1.DCLookupByUserDomain"="Enabled", 
			"ActiveDirectory.1.DCLookupDomainName"="test",          
		}
	}
	  `,
		testingInfo.Username,
		testingInfo.Password,
		testingInfo.Endpoint,
	)
}

func testAccRedfishDirectoryServiceAuthProviderDCLookupEnableInvalidConfig(testingInfo TestingServerCredentials) string {
	return fmt.Sprintf(`
	resource "redfish_directory_service_auth_provider" "ds_auth" {
		redfish_server {
		  user         = "%s"
		  password     = "%s"
		  endpoint     = "%s"
		  ssl_insecure = true
		}
	  
		active_directory = {
			directory = {
				service_enabled = true,
				authentication = {
					kerberos_key_tab_file = ""
				}
			}
		}
		
		active_directory_attributes = {
			"ActiveDirectory.1.AuthTimeout"= "120",
			"ActiveDirectory.1.CertValidationEnable"= "Enabled",
			"ActiveDirectory.1.DCLookupEnable"= "Invalid",
			"ActiveDirectory.1.RacDomain"= "test",
			"ActiveDirectory.1.RacName"= "test",
			"ActiveDirectory.1.SSOEnable"= "Disabled",
			"ActiveDirectory.1.Schema"= "Extended Schema",
			"UserDomain.1.Name"= "yulan.pie.lab.emc.com",
			"ActiveDirectory.1.DCLookupByUserDomain"="Enabled", 
			"ActiveDirectory.1.DCLookupDomainName"="test",          
		}
	}
	  `,
		testingInfo.Username,
		testingInfo.Password,
		testingInfo.Endpoint,
	)
}
func testAccRedfishDirectoryServiceAuthProviderDCLookupEnableInvalid17GConfig(testingInfo TestingServerCredentials) string {
	return fmt.Sprintf(`
	resource "redfish_directory_service_auth_provider" "ds_auth" {
		redfish_server {
		  user         = "%s"
		  password     = "%s"
		  endpoint     = "%s"
		  ssl_insecure = true
		}
	  
		active_directory = {
			directory = {
				service_enabled = true,
			}
		}
		
		active_directory_attributes = {
			"ActiveDirectory.1.AuthTimeout"= "120",
			"ActiveDirectory.1.CertValidationEnable"= "Enabled",
			"ActiveDirectory.1.DCLookupEnable"= "Invalid",
			"UserDomain.1.Name"= "yulan.pie.lab.emc.com",
			"ActiveDirectory.1.DCLookupByUserDomain"="Enabled", 
			"ActiveDirectory.1.DCLookupDomainName"="test",          
		}
	}
	  `,
		testingInfo.Username,
		testingInfo.Password,
		testingInfo.Endpoint,
	)
}

func testAccRedfishDirectoryServiceAuthProviderExtendedNoRacConfig(testingInfo TestingServerCredentials) string {
	return fmt.Sprintf(`
	resource "redfish_directory_service_auth_provider" "ds_auth" {
		redfish_server {
		  user         = "%s"
		  password     = "%s"
		  endpoint     = "%s"
		  ssl_insecure = true
		}
	  
		active_directory = {
			directory = {
				service_enabled = true,
				authentication = {
					kerberos_key_tab_file = ""
				}
			}
		}
		
		active_directory_attributes = {
			"ActiveDirectory.1.AuthTimeout"= "120",
			"ActiveDirectory.1.CertValidationEnable"= "Enabled",
			"ActiveDirectory.1.DCLookupEnable"= "Enabled",
			"ActiveDirectory.1.SSOEnable"= "Disabled",
			"ActiveDirectory.1.Schema"= "Extended Schema",
			"UserDomain.1.Name"= "yulan.pie.lab.emc.com",
			"ActiveDirectory.1.DCLookupByUserDomain"="Enabled",          
		}
	}
	  `,
		testingInfo.Username,
		testingInfo.Password,
		testingInfo.Endpoint,
	)
}

func testAccRedfishDirectoryServiceAuthProviderExtendedEmptyRacConfig(testingInfo TestingServerCredentials) string {
	return fmt.Sprintf(`
	resource "redfish_directory_service_auth_provider" "ds_auth" {
		redfish_server {
		  user         = "%s"
		  password     = "%s"
		  endpoint     = "%s"
		  ssl_insecure = true
		}
	  
		active_directory = {
			directory = {
				service_enabled = true,
				authentication = {
					kerberos_key_tab_file = ""
				}
			}
		}
		
		active_directory_attributes = {
			"ActiveDirectory.1.AuthTimeout"= "120",
			"ActiveDirectory.1.CertValidationEnable"= "Enabled",
			"ActiveDirectory.1.DCLookupEnable"= "Enabled",
			"ActiveDirectory.1.SSOEnable"= "Disabled",
			"ActiveDirectory.1.Schema"= "Extended Schema",
			"ActiveDirectory.1.RacDomain"= "",
			"ActiveDirectory.1.RacName"= "",
			"UserDomain.1.Name"= "yulan.pie.lab.emc.com",
			"ActiveDirectory.1.DCLookupByUserDomain"="Enabled",          
		}
	}
	  `,
		testingInfo.Username,
		testingInfo.Password,
		testingInfo.Endpoint,
	)
}

func testAccRedfishDirectoryServiceAuthProviderExtendedGCLookUpConfig(testingInfo TestingServerCredentials) string {
	return fmt.Sprintf(`
	resource "redfish_directory_service_auth_provider" "ds_auth" {
		redfish_server {
		  user         = "%s"
		  password     = "%s"
		  endpoint     = "%s"
		  ssl_insecure = true
		}
	  
		active_directory = {
			directory = {
				service_enabled = true,
				authentication = {
					kerberos_key_tab_file = ""
				}
			}
		}
		
		active_directory_attributes = {
			"ActiveDirectory.1.AuthTimeout"= "120",
			"ActiveDirectory.1.CertValidationEnable"= "Enabled",
			"ActiveDirectory.1.DCLookupEnable"= "Enabled",
			"ActiveDirectory.1.SSOEnable"= "Disabled",
			"ActiveDirectory.1.Schema"= "Extended Schema",
			"ActiveDirectory.1.RacDomain"= "test",
			"ActiveDirectory.1.RacName"= "test",
			"UserDomain.1.Name"= "yulan.pie.lab.emc.com",
			"ActiveDirectory.1.DCLookupByUserDomain"="Enabled", 
			"ActiveDirectory.1.GCLookupEnable" = "Disabled",         
		}
	}
	  `,
		testingInfo.Username,
		testingInfo.Password,
		testingInfo.Endpoint,
	)
}

func testAccRedfishDirectoryServiceAuthProviderExtendedRemoteRoleConfig(testingInfo TestingServerCredentials) string {
	return fmt.Sprintf(`
	resource "redfish_directory_service_auth_provider" "ds_auth" {
		redfish_server {
		  user         = "%s"
		  password     = "%s"
		  endpoint     = "%s"
		  ssl_insecure = true
		}
	  
		active_directory = {
			directory = {
				remote_role_mapping = [
					{
						local_role = "Administrator",
						remote_group = "xxxx"
					}
				],
				service_enabled = true,
				authentication = {
					kerberos_key_tab_file = ""
				}
			}
		}
		
		active_directory_attributes = {
			"ActiveDirectory.1.AuthTimeout"= "120",
			"ActiveDirectory.1.CertValidationEnable"= "Enabled",
			"ActiveDirectory.1.DCLookupEnable"= "Enabled",
			"ActiveDirectory.1.SSOEnable"= "Disabled",
			"ActiveDirectory.1.Schema"= "Extended Schema",
			"ActiveDirectory.1.RacDomain"= "test",
			"ActiveDirectory.1.RacName"= "test",
			"UserDomain.1.Name"= "yulan.pie.lab.emc.com",
			"ActiveDirectory.1.DCLookupByUserDomain"="Enabled", 
			       
		}
	}
	  `,
		testingInfo.Username,
		testingInfo.Password,
		testingInfo.Endpoint,
	)
}

func testAccRedfishDirectoryServiceAuthProviderExtendedADGroupDomainConfig(testingInfo TestingServerCredentials) string {
	return fmt.Sprintf(`
	resource "redfish_directory_service_auth_provider" "ds_auth" {
		redfish_server {
		  user         = "%s"
		  password     = "%s"
		  endpoint     = "%s"
		  ssl_insecure = true
		}
	  
		active_directory = {
			directory = {
				service_enabled = true,
				authentication = {
					kerberos_key_tab_file = ""
				}
			}
		}
		
		active_directory_attributes = {
			"ActiveDirectory.1.AuthTimeout"= "120",
			"ActiveDirectory.1.CertValidationEnable"= "Enabled",
			"ActiveDirectory.1.DCLookupEnable"= "Enabled",
			"ActiveDirectory.1.SSOEnable"= "Disabled",
			"ActiveDirectory.1.Schema"= "Extended Schema",
			"ActiveDirectory.1.RacDomain"= "test",
			"ActiveDirectory.1.RacName"= "test",
			"UserDomain.1.Name"= "yulan.pie.lab.emc.com",
			"ActiveDirectory.1.DCLookupByUserDomain"="Enabled", 
			"ADGroup.1.Domain" = "yulan.pie.lab.emc.com",
			       
		}
	}
	  `,
		testingInfo.Username,
		testingInfo.Password,
		testingInfo.Endpoint,
	)
}

func testAccRedfishDirectoryServiceAuthProviderStandardSchemaAndRacConfig(testingInfo TestingServerCredentials) string {
	return fmt.Sprintf(`
	resource "redfish_directory_service_auth_provider" "ds_auth" {
		redfish_server {
		  user         = "%s"
		  password     = "%s"
		  endpoint     = "%s"
		  ssl_insecure = true
		}
	  
		active_directory = {
			directory = {
				service_enabled = true,
				authentication = {
					kerberos_key_tab_file = ""
				}
			}
		}
		
		active_directory_attributes = {
			"ActiveDirectory.1.AuthTimeout"= "120",
			"ActiveDirectory.1.CertValidationEnable"= "Enabled",
			"ActiveDirectory.1.DCLookupEnable"= "Enabled",
			"ActiveDirectory.1.SSOEnable"= "Disabled",
			"ActiveDirectory.1.Schema"= "Standard Schema",
			"ActiveDirectory.1.RacDomain"= "test",
			"ActiveDirectory.1.RacName"= "test",
			"UserDomain.1.Name"= "yulan.pie.lab.emc.com",
			"ActiveDirectory.1.DCLookupByUserDomain"="Enabled",          
		}
	}
	  `,
		testingInfo.Username,
		testingInfo.Password,
		testingInfo.Endpoint,
	)
}

func testAccRedfishDirectoryServiceAuthProviderStandardSchemaAndRac17GConfig(testingInfo TestingServerCredentials) string {
	return fmt.Sprintf(`
	resource "redfish_directory_service_auth_provider" "ds_auth" {
		redfish_server {
		  user         = "%s"
		  password     = "%s"
		  endpoint     = "%s"
		  ssl_insecure = true
		}
	  
		active_directory = {
			directory = {
				service_enabled = true,
			}
		}
		
		active_directory_attributes = {
			"ActiveDirectory.1.AuthTimeout"= "120",
			"ActiveDirectory.1.CertValidationEnable"= "Enabled",
			"ActiveDirectory.1.DCLookupEnable"= "Enabled",
			"ActiveDirectory.1.RacDomain"= "test",
			"ActiveDirectory.1.RacName"= "test",
			"UserDomain.1.Name"= "yulan.pie.lab.emc.com",
			"ActiveDirectory.1.DCLookupByUserDomain"="Enabled",          
		}
	}
	  `,
		testingInfo.Username,
		testingInfo.Password,
		testingInfo.Endpoint,
	)
}

func testAccRedfishDirectoryServiceAuthProviderStandardSchemaNoGCLookUpConfig(testingInfo TestingServerCredentials) string {
	return fmt.Sprintf(`
	resource "redfish_directory_service_auth_provider" "ds_auth" {
		redfish_server {
		  user         = "%s"
		  password     = "%s"
		  endpoint     = "%s"
		  ssl_insecure = true
		}
	  
		active_directory = {
			directory = {
				service_enabled = true,
				authentication = {
					kerberos_key_tab_file = ""
				}
			}
		}
		
		active_directory_attributes = {
			"ActiveDirectory.1.AuthTimeout"= "120",
			"ActiveDirectory.1.CertValidationEnable"= "Enabled",
			"ActiveDirectory.1.DCLookupEnable"= "Enabled",
			"ActiveDirectory.1.SSOEnable"= "Disabled",
			"ActiveDirectory.1.Schema"= "Standard Schema",
			"UserDomain.1.Name"= "yulan.pie.lab.emc.com",
			"ActiveDirectory.1.DCLookupByUserDomain"="Enabled",          
		}
	}
	  `,
		testingInfo.Username,
		testingInfo.Password,
		testingInfo.Endpoint,
	)
}

func testAccRedfishDirectoryServiceAuthProviderStandardSchemaNoGCRootConfig(testingInfo TestingServerCredentials) string {
	return fmt.Sprintf(`
	resource "redfish_directory_service_auth_provider" "ds_auth" {
		redfish_server {
		  user         = "%s"
		  password     = "%s"
		  endpoint     = "%s"
		  ssl_insecure = true
		}
	  
		active_directory = {
			directory = {
				service_enabled = true,
				authentication = {
					kerberos_key_tab_file = ""
				}
			}
		}
		
		active_directory_attributes = {
			"ActiveDirectory.1.AuthTimeout"= "120",
			"ActiveDirectory.1.CertValidationEnable"= "Enabled",
			"ActiveDirectory.1.DCLookupEnable"= "Enabled",
			"ActiveDirectory.1.SSOEnable"= "Disabled",
			"ActiveDirectory.1.Schema"= "Standard Schema",
			"UserDomain.1.Name"= "yulan.pie.lab.emc.com",
			"ActiveDirectory.1.DCLookupByUserDomain"="Enabled", 
			"ActiveDirectory.1.GCLookupEnable" = "Enabled",         
		}
	}
	  `,
		testingInfo.Username,
		testingInfo.Password,
		testingInfo.Endpoint,
	)
}

func testAccRedfishDirectoryServiceAuthProviderStandardSchemaGlobalCatalogConfig(testingInfo TestingServerCredentials) string {
	return fmt.Sprintf(`
	resource "redfish_directory_service_auth_provider" "ds_auth" {
		redfish_server {
		  user         = "%s"
		  password     = "%s"
		  endpoint     = "%s"
		  ssl_insecure = true
		}
	  
		active_directory = {
			directory = {
				service_enabled = true,
				authentication = {
					kerberos_key_tab_file = ""
				}
			}
		}
		
		active_directory_attributes = {
			"ActiveDirectory.1.AuthTimeout"= "120",
			"ActiveDirectory.1.CertValidationEnable"= "Enabled",
			"ActiveDirectory.1.DCLookupEnable"= "Enabled",
			"ActiveDirectory.1.SSOEnable"= "Disabled",
			"ActiveDirectory.1.Schema"= "Standard Schema",
			"UserDomain.1.Name"= "yulan.pie.lab.emc.com",
			"ActiveDirectory.1.DCLookupByUserDomain"="Enabled", 
			"ActiveDirectory.1.GCLookupEnable" = "Enabled",
			"ActiveDirectory.1.GCRootDomain" = "test",
			"ActiveDirectory.1.GlobalCatalog1" = "yulanadhost11.yulan.pie.lab.emc.com",         
		}
	}
	  `,
		testingInfo.Username,
		testingInfo.Password,
		testingInfo.Endpoint,
	)
}

func testAccRedfishDirectoryServiceAuthProviderStandardSchemaNoGlobalCatalogConfig(testingInfo TestingServerCredentials) string {
	return fmt.Sprintf(`
	resource "redfish_directory_service_auth_provider" "ds_auth" {
		redfish_server {
		  user         = "%s"
		  password     = "%s"
		  endpoint     = "%s"
		  ssl_insecure = true
		}
	  
		active_directory = {
			directory = {
				service_enabled = true,
				authentication = {
					kerberos_key_tab_file = ""
				}
			}
		}
		
		active_directory_attributes = {
			"ActiveDirectory.1.AuthTimeout"= "120",
			"ActiveDirectory.1.CertValidationEnable"= "Enabled",
			"ActiveDirectory.1.DCLookupEnable"= "Enabled",
			"ActiveDirectory.1.SSOEnable"= "Disabled",
			"ActiveDirectory.1.Schema"= "Standard Schema",
			"UserDomain.1.Name"= "yulan.pie.lab.emc.com",
			"ActiveDirectory.1.DCLookupByUserDomain"="Enabled", 
			"ActiveDirectory.1.GCLookupEnable" = "Disabled",
			"ActiveDirectory.1.GlobalCatalog1" = "",         
		}
	}
	  `,
		testingInfo.Username,
		testingInfo.Password,
		testingInfo.Endpoint,
	)
}

func testAccRedfishDirectoryServiceAuthProviderStandardSchemaGCRootConfig(testingInfo TestingServerCredentials) string {
	return fmt.Sprintf(`
	resource "redfish_directory_service_auth_provider" "ds_auth" {
		redfish_server {
		  user         = "%s"
		  password     = "%s"
		  endpoint     = "%s"
		  ssl_insecure = true
		}
	  
		active_directory = {
			directory = {
				service_enabled = true,
				authentication = {
					kerberos_key_tab_file = ""
				}
			}
		}
		
		active_directory_attributes = {
			"ActiveDirectory.1.AuthTimeout"= "120",
			"ActiveDirectory.1.CertValidationEnable"= "Enabled",
			"ActiveDirectory.1.DCLookupEnable"= "Enabled",
			"ActiveDirectory.1.SSOEnable"= "Disabled",
			"ActiveDirectory.1.Schema"= "Standard Schema",
			"UserDomain.1.Name"= "yulan.pie.lab.emc.com",
			"ActiveDirectory.1.DCLookupByUserDomain"="Enabled", 
			"ActiveDirectory.1.GCLookupEnable" = "Disabled",
			"ActiveDirectory.1.GCRootDomain" = "test",
			"ActiveDirectory.1.GlobalCatalog1" = "yulanadhost11.yulan.pie.lab.emc.com",         
		}
	}
	  `,
		testingInfo.Username,
		testingInfo.Password,
		testingInfo.Endpoint,
	)
}

func testAccRedfishDirectoryServiceAuthProviderStandardSchemaInvalidGCLookUpConfig(testingInfo TestingServerCredentials) string {
	return fmt.Sprintf(`
	resource "redfish_directory_service_auth_provider" "ds_auth" {
		redfish_server {
		  user         = "%s"
		  password     = "%s"
		  endpoint     = "%s"
		  ssl_insecure = true
		}
	  
		active_directory = {
			directory = {
				service_enabled = true,
				authentication = {
					kerberos_key_tab_file = ""
				}
			}
		}
		
		active_directory_attributes = {
			"ActiveDirectory.1.AuthTimeout"= "120",
			"ActiveDirectory.1.CertValidationEnable"= "Enabled",
			"ActiveDirectory.1.DCLookupEnable"= "Enabled",
			"ActiveDirectory.1.SSOEnable"= "Disabled",
			"ActiveDirectory.1.Schema"= "Standard Schema",
			"UserDomain.1.Name"= "yulan.pie.lab.emc.com",
			"ActiveDirectory.1.DCLookupByUserDomain"="Enabled", 
			"ActiveDirectory.1.GCLookupEnable" = "Invalid",
			"ActiveDirectory.1.GCRootDomain" = "test",
			"ActiveDirectory.1.GlobalCatalog1" = "yulanadhost11.yulan.pie.lab.emc.com",         
		}
	}
	  `,
		testingInfo.Username,
		testingInfo.Password,
		testingInfo.Endpoint,
	)
}

func testAccRedfishDirectoryServiceAuthProviderStandardSchemaNoGCLookUp17GConfig(testingInfo TestingServerCredentials) string {
	return fmt.Sprintf(`
	resource "redfish_directory_service_auth_provider" "ds_auth" {
		redfish_server {
		  user         = "%s"
		  password     = "%s"
		  endpoint     = "%s"
		  ssl_insecure = true
		}
	  
		active_directory = {
			directory = {
				service_enabled = true,
				
			}
		}
		
		active_directory_attributes = {
			"ActiveDirectory.1.AuthTimeout"= "120",
			"ActiveDirectory.1.CertValidationEnable"= "Enabled",
			"ActiveDirectory.1.DCLookupEnable"= "Enabled",
			"UserDomain.1.Name"= "yulan.pie.lab.emc.com",
			"ActiveDirectory.1.DCLookupByUserDomain"="Enabled",          
		}
	}
	  `,
		testingInfo.Username,
		testingInfo.Password,
		testingInfo.Endpoint,
	)
}

func testAccRedfishDirectoryServiceAuthProviderStandardSchemaNoGCRoot17GConfig(testingInfo TestingServerCredentials) string {
	return fmt.Sprintf(`
	resource "redfish_directory_service_auth_provider" "ds_auth" {
		redfish_server {
		  user         = "%s"
		  password     = "%s"
		  endpoint     = "%s"
		  ssl_insecure = true
		}
	  
		active_directory = {
			directory = {
				service_enabled = true,
			}
		}
		
		active_directory_attributes = {
			"ActiveDirectory.1.AuthTimeout"= "120",
			"ActiveDirectory.1.CertValidationEnable"= "Enabled",
			"ActiveDirectory.1.DCLookupEnable"= "Enabled",
			"UserDomain.1.Name"= "yulan.pie.lab.emc.com",
			"ActiveDirectory.1.DCLookupByUserDomain"="Enabled", 
			"ActiveDirectory.1.GCLookupEnable" = "Enabled",         
		}
	}
	  `,
		testingInfo.Username,
		testingInfo.Password,
		testingInfo.Endpoint,
	)
}

func testAccRedfishDirectoryServiceAuthProviderStandardSchemaGlobalCatalog17GConfig(testingInfo TestingServerCredentials) string {
	return fmt.Sprintf(`
	resource "redfish_directory_service_auth_provider" "ds_auth" {
		redfish_server {
		  user         = "%s"
		  password     = "%s"
		  endpoint     = "%s"
		  ssl_insecure = true
		}
	  
		active_directory = {
			directory = {
				service_enabled = true,
			}
		}
		
		active_directory_attributes = {
			"ActiveDirectory.1.AuthTimeout"= "120",
			"ActiveDirectory.1.CertValidationEnable"= "Enabled",
			"ActiveDirectory.1.DCLookupEnable"= "Enabled",
			"UserDomain.1.Name"= "yulan.pie.lab.emc.com",
			"ActiveDirectory.1.DCLookupByUserDomain"="Enabled", 
			"ActiveDirectory.1.GCLookupEnable" = "Enabled",
			"ActiveDirectory.1.GCRootDomain" = "test",
			"ActiveDirectory.1.GlobalCatalog1" = "yulanadhost11.yulan.pie.lab.emc.com",         
		}
	}
	  `,
		testingInfo.Username,
		testingInfo.Password,
		testingInfo.Endpoint,
	)
}

func testAccRedfishDirectoryServiceAuthProviderStandardSchemaNoGlobalCatalog17GConfig(testingInfo TestingServerCredentials) string {
	return fmt.Sprintf(`
	resource "redfish_directory_service_auth_provider" "ds_auth" {
		redfish_server {
		  user         = "%s"
		  password     = "%s"
		  endpoint     = "%s"
		  ssl_insecure = true
		}
	  
		active_directory = {
			directory = {
				service_enabled = true,
			}
		}
		
		active_directory_attributes = {
			"ActiveDirectory.1.AuthTimeout"= "120",
			"ActiveDirectory.1.CertValidationEnable"= "Enabled",
			"ActiveDirectory.1.DCLookupEnable"= "Enabled",
			"UserDomain.1.Name"= "yulan.pie.lab.emc.com",
			"ActiveDirectory.1.DCLookupByUserDomain"="Enabled", 
			"ActiveDirectory.1.GCLookupEnable" = "Disabled",
			"ActiveDirectory.1.GlobalCatalog1" = "",         
		}
	}
	  `,
		testingInfo.Username,
		testingInfo.Password,
		testingInfo.Endpoint,
	)
}

func testAccRedfishDirectoryServiceAuthProviderStandardSchemaGCRoot17GConfig(testingInfo TestingServerCredentials) string {
	return fmt.Sprintf(`
	resource "redfish_directory_service_auth_provider" "ds_auth" {
		redfish_server {
		  user         = "%s"
		  password     = "%s"
		  endpoint     = "%s"
		  ssl_insecure = true
		}
	  
		active_directory = {
			directory = {
				service_enabled = true,
			}
		}
		
		active_directory_attributes = {
			"ActiveDirectory.1.AuthTimeout"= "120",
			"ActiveDirectory.1.CertValidationEnable"= "Enabled",
			"ActiveDirectory.1.DCLookupEnable"= "Enabled",
			"UserDomain.1.Name"= "yulan.pie.lab.emc.com",
			"ActiveDirectory.1.DCLookupByUserDomain"="Enabled", 
			"ActiveDirectory.1.GCLookupEnable" = "Disabled",
			"ActiveDirectory.1.GCRootDomain" = "test",
			"ActiveDirectory.1.GlobalCatalog1" = "yulanadhost11.yulan.pie.lab.emc.com",         
		}
	}
	  `,
		testingInfo.Username,
		testingInfo.Password,
		testingInfo.Endpoint,
	)
}

func testAccRedfishDirectoryServiceAuthProviderStandardSchemaInvalidGCLookUp17GConfig(testingInfo TestingServerCredentials) string {
	return fmt.Sprintf(`
	resource "redfish_directory_service_auth_provider" "ds_auth" {
		redfish_server {
		  user         = "%s"
		  password     = "%s"
		  endpoint     = "%s"
		  ssl_insecure = true
		}
	  
		active_directory = {
			directory = {
				service_enabled = true,
			}
		}
		
		active_directory_attributes = {
			"ActiveDirectory.1.AuthTimeout"= "120",
			"ActiveDirectory.1.CertValidationEnable"= "Enabled",
			"ActiveDirectory.1.DCLookupEnable"= "Enabled",

			"UserDomain.1.Name"= "yulan.pie.lab.emc.com",
			"ActiveDirectory.1.DCLookupByUserDomain"="Enabled", 
			"ActiveDirectory.1.GCLookupEnable" = "Invalid",
			"ActiveDirectory.1.GCRootDomain" = "test",
			"ActiveDirectory.1.GlobalCatalog1" = "yulanadhost11.yulan.pie.lab.emc.com",         
		}
	}
	  `,
		testingInfo.Username,
		testingInfo.Password,
		testingInfo.Endpoint,
	)
}
