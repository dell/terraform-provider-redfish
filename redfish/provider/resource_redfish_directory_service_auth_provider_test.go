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
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccRedfishDirectoryServiceAuthProviderBasic(t *testing.T) {
	terraformDSAuthProviderResourceName := "redfish_directory_service_auth_provider.ds_auth"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				// error create with both `ActiveDirectory` and `LDAP`
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
				// update with `ActiveDirectory`
				Config: testAccRedfishDirectoryServiceAuthProviderAD_UpdateConfig(creds),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(terraformDSAuthProviderResourceName, "active_directory.directory.service_enabled", "false"),
					resource.TestCheckResourceAttr(terraformDSAuthProviderResourceName, "active_directory_attributes.ActiveDirectory.1.AuthTimeout", "130"),
				),
			},
		},
	})
}

func TestAccRedfishDirectoryServiceAuthProviderInvalidCase(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				// error for empty AuthTimeout in ActiveDirectory
				Config:      testAccRedfishDirectoryServiceAuthProviderEmptyAuth(creds),
				ExpectError: regexp.MustCompile("Invalid AuthTimeout, Please provide all the required configuration"),
			},

			{
				// error for Invalid AuthTimeout in ActiveDirectory
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
				// error DCLookupEnable Enabled, DCLookupByUserDomain Disabled and DCLookupDomainName Empty
				Config:      testAccRedfishDirectoryServiceAuthProviderDCLookupDomainNameEmptyConfig(creds),
				ExpectError: regexp.MustCompile("DCLookupDomainName must be configured for Disabled DCLookupByUserDomain"),
			},

			{
				// error DCLookupEnable Enabled DCLookupByUserDomain Enabled and DCLookupDomainName non Empty
				Config:      testAccRedfishDirectoryServiceAuthProviderDCLookupEnableDCLookupDomainNameConfig(creds),
				ExpectError: regexp.MustCompile("DCLookupDomainName can not be configured for Enabled DCLookupByUserDomain"),
			},
		},
	})
}

func TestAccRedfishDirectoryServiceAuthProviderInvalidSchema_Config(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				// error Extended Schema without RacName and RacDomain
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
		},
	})
}

func TestAccRedfishDirectoryServiceAuthProviderImport(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `resource "redfish_directory_service_auth_provider" "ds_auth" {
					}`,
				ResourceName:  "redfish_directory_service_auth_provider.ds_auth",
				ImportState:   true,
				ImportStateId: "{\"username\":\"" + creds.Username + "\",\"password\":\"" + creds.Password + "\",\"endpoint\":\"" + creds.Endpoint + "\",\"ssl_insecure\":true}",
				ExpectError:   nil,
			},
		},
	})
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
				remote_role_mapping = [
					{
						local_role = "Administrator",
						remote_group = "xxxx"
					}
				],
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
			"ActiveDirectory.1.DCLookupByUserDomain":"Enabled",          
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
			"ActiveDirectory.1.DCLookupByUserDomain":"Enabled",          
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
			"ActiveDirectory.1.DCLookupByUserDomain":"Disabled",           
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
			#"ActiveDirectory.1.DCLookupByUserDomain":"Disabled", 
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
			"ActiveDirectory.1.DCLookupByUserDomain":"Disabled", 
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
			"ActiveDirectory.1.DCLookupByUserDomain":"Disabled", 
			        
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
			"ActiveDirectory.1.DCLookupByUserDomain":"Enabled", 
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
			"ActiveDirectory.1.DCLookupByUserDomain":"Enabled",          
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
			"ActiveDirectory.1.DCLookupByUserDomain":"Enabled",          
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
			"ActiveDirectory.1.DCLookupByUserDomain":"Enabled", 
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
			"ActiveDirectory.1.DCLookupByUserDomain":"Enabled", 
			       
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
			"ActiveDirectory.1.DCLookupByUserDomain":"Enabled", 
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
			"ActiveDirectory.1.DCLookupByUserDomain":"Enabled",          
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
			"ActiveDirectory.1.DCLookupByUserDomain":"Enabled",          
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
			"ActiveDirectory.1.DCLookupByUserDomain":"Enabled", 
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
			"ActiveDirectory.1.DCLookupByUserDomain":"Enabled", 
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
			"ActiveDirectory.1.DCLookupByUserDomain":"Enabled", 
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
			"ActiveDirectory.1.DCLookupByUserDomain":"Enabled", 
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
