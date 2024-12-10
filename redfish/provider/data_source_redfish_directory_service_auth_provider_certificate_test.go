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

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccRedfishDirectoryServiceAuthProviderCertificate_fetch(t *testing.T) {
	dSAPCertificateDatasourceName := "data.redfish_directory_service_auth_provider_certificate.test"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccRedfishDirectoryServiceAuthProviderCertificatewithoutFilterConfig(creds),
				ExpectError: regexp.MustCompile(`Missing Configuration for Required Attribute`),
			},
			{
				Config:      testAccDirectoryServiceAuthProviderCertificateWithEmptyCertificateProviderFilter(creds),
				ExpectError: regexp.MustCompile(`Invalid CertificateProviderType`),
			},

			{
				Config:      testAccDirectoryServiceAuthProviderCertificateWithInvalidCertificateProviderFilter(creds),
				ExpectError: regexp.MustCompile(`Invalid CertificateProviderType`),
			},
			{
				Config: testAccDirectoryServiceAuthProviderCertificateWithFilter(creds, os.Getenv("TF_TESTING_DS_AP_CERTIFICATE_ID")),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(dSAPCertificateDatasourceName, "directory_service_auth_provider_certificate.directory_service_certificate.certificate_usage_types.0", "User"),
				),
			},
			{
				Config: testAccDirectoryServiceAuthProviderCertificateWithoutCertificateId(creds),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(dSAPCertificateDatasourceName, "directory_service_auth_provider_certificate.directory_service_certificate.certificate_usage_types.0", "User"),
				),
			},
			{
				Config:      testAccDirectoryServiceAuthProviderCertificateWithEmptyCertificateId(creds),
				ExpectError: regexp.MustCompile(`CertificateId can't be empty value`),
			},
			{
				Config:      testAccDirectoryServiceAuthProviderCertificateWithInvalidCertificateId(creds),
				ExpectError: regexp.MustCompile(`Error fetching Certificate`),
			},
		},
	})
}

func testAccRedfishDirectoryServiceAuthProviderCertificatewithoutFilterConfig(testingInfo TestingServerCredentials) string {
	return fmt.Sprintf(`
	data "redfish_directory_service_auth_provider_certificate" "test" {	  
		redfish_server {
		  user         = "%s"
		  password     = "%s"
		  endpoint     = "%s"
		  ssl_insecure = true
		}
	}`,
		testingInfo.Username,
		testingInfo.Password,
		testingInfo.Endpoint,
	)
}

func testAccDirectoryServiceAuthProviderCertificateWithEmptyCertificateProviderFilter(testingInfo TestingServerCredentials) string {
	return fmt.Sprintf(`
	data "redfish_directory_service_auth_provider_certificate" "test" {	  
		redfish_server {
		  user         = "%s"
		  password     = "%s"
		  endpoint     = "%s"
		  ssl_insecure = true
		}
		certificate_filter {
			certificate_provider_type = ""
		  }
	}`,
		testingInfo.Username,
		testingInfo.Password,
		testingInfo.Endpoint,
	)
}

func testAccDirectoryServiceAuthProviderCertificateWithInvalidCertificateProviderFilter(testingInfo TestingServerCredentials) string {
	return fmt.Sprintf(`
	data "redfish_directory_service_auth_provider_certificate" "test" {	  
		redfish_server {
		  user         = "%s"
		  password     = "%s"
		  endpoint     = "%s"
		  ssl_insecure = true
		}
		certificate_filter {
			certificate_provider_type = "INVALID"
		  }
	}`,
		testingInfo.Username,
		testingInfo.Password,
		testingInfo.Endpoint,
	)
}

func testAccDirectoryServiceAuthProviderCertificateWithFilter(testingInfo TestingServerCredentials, certificateId string) string {
	return fmt.Sprintf(`
	data "redfish_directory_service_auth_provider_certificate" "test" {	  
		redfish_server {
		  user         = "%s"
		  password     = "%s"
		  endpoint     = "%s"
		  ssl_insecure = true
		}
		certificate_filter {
			certificate_provider_type = "LDAP"
			certificate_id ="%s"
		}
	}`,
		testingInfo.Username,
		testingInfo.Password,
		testingInfo.Endpoint,
		certificateId,
	)
}

func testAccDirectoryServiceAuthProviderCertificateWithoutCertificateId(testingInfo TestingServerCredentials) string {
	return fmt.Sprintf(`
	data "redfish_directory_service_auth_provider_certificate" "test" {	  
		redfish_server {
		  user         = "%s"
		  password     = "%s"
		  endpoint     = "%s"
		  ssl_insecure = true
		}
		certificate_filter {
			certificate_provider_type = "LDAP"
		}
	}`,
		testingInfo.Username,
		testingInfo.Password,
		testingInfo.Endpoint,
	)
}

func testAccDirectoryServiceAuthProviderCertificateWithEmptyCertificateId(testingInfo TestingServerCredentials) string {
	return fmt.Sprintf(`
	data "redfish_directory_service_auth_provider_certificate" "test" {	  
		redfish_server {
		  user         = "%s"
		  password     = "%s"
		  endpoint     = "%s"
		  ssl_insecure = true
		}
		certificate_filter {
			certificate_provider_type = "LDAP"
			certificate_id =""
		}
	}`,
		testingInfo.Username,
		testingInfo.Password,
		testingInfo.Endpoint,
	)
}

func testAccDirectoryServiceAuthProviderCertificateWithInvalidCertificateId(testingInfo TestingServerCredentials) string {
	return fmt.Sprintf(`
	data "redfish_directory_service_auth_provider_certificate" "test" {	  
		redfish_server {
		  user         = "%s"
		  password     = "%s"
		  endpoint     = "%s"
		  ssl_insecure = true
		}
		certificate_filter {
			certificate_provider_type = "LDAP"
			certificate_id ="INVALID"
		}
	}`,
		testingInfo.Username,
		testingInfo.Password,
		testingInfo.Endpoint,
	)
}
