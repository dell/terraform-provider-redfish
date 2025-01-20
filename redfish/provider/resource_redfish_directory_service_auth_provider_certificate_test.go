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

func TestAccRedfishDirectoryServiceAuthProviderCertificateBasic(t *testing.T) {
	valid_ds_cert := os.Getenv("VALID_DS_CERT")
	valid_ds_cert_update := os.Getenv("VALID_DS_CERT_UPDATE")
	invalid_ds_cert := os.Getenv("INVALID_DS_CERT")
	terraformDSAuthProviderCertificateResourceName := "redfish_directory_service_auth_provider_certificate.ds_auth_certificate"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		ExternalProviders: map[string]resource.ExternalProvider{
			"local": {
				Source: "hashicorp/local",
			},
		},
		Steps: []resource.TestStep{
			{
				// create certificate import resource
				Config: testAccRedfishDirectoryServiceAuthProviderCertificateConfig(creds, valid_ds_cert, "PEM"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(terraformDSAuthProviderCertificateResourceName, "certificate_type", "PEM"),
				),
			},

			{
				// update certificate import resource
				Config: testAccRedfishDirectoryServiceAuthProviderCertificateConfig(creds, valid_ds_cert_update, "PEM"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(terraformDSAuthProviderCertificateResourceName, "certificate_type", "PEM"),
				),
			},
			{
				// update certificate import resource with invalid cert
				Config:      testAccRedfishDirectoryServiceAuthProviderCertificateConfig(creds, invalid_ds_cert, "PEM"),
				ExpectError: regexp.MustCompile("There was an error while creating/updating Certificate resource"),
			},
		},
	})
}

func testAccRedfishDirectoryServiceAuthProviderCertificateConfig(testingInfo TestingServerCredentials, certfile string, certificateType string) string {
	return fmt.Sprintf(`

	data "local_file" "ds_certificate" {
		filename = "%s"
	  }

	resource "redfish_directory_service_auth_provider_certificate" "ds_auth_certificate" {
		redfish_server {
		  user         = "%s"
		  password     = "%s"
		  endpoint     = "%s"
		  ssl_insecure = true
		}
		certificate_type = "%s"
		certificate_string = data.local_file.ds_certificate.content	
	}
	  `,
		certfile,
		testingInfo.Username,
		testingInfo.Password,
		testingInfo.Endpoint,
		certificateType,
	)
}
