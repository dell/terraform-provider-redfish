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

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// test redfish bios settings
func TestAccRedfishCertificate_basic(t *testing.T) {
	valid_cert := os.Getenv("VALID_CERT")
	invalid_cert := os.Getenv("INVALID_CERT")
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
				Config: testAccRedfishResourceCustomCertificate(
					creds, valid_cert),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("redfish_certificate.cert", "certificate_type", "CustomCertificate"),
				),
			},
			{
				Config: testAccRedfishResourceCustomCertificate(
					creds, invalid_cert),
				ExpectError: regexp.MustCompile("Couldn't upload certificate from redfish API"),
			},
		},
	})
}

func testAccRedfishResourceCustomCertificate(testingInfo TestingServerCredentials, certfile string) string {
	return fmt.Sprintf(`
		data "local_file" "cert" {
			filename = "%s"
	  	}
		resource "redfish_certificate" "cert"  {
		  redfish_server {
			user = "%s"
			password = "%s"
			endpoint = "%s"
			ssl_insecure = true
		  }

		  certificate_type = "CustomCertificate"
		  passphrase = "12345"
		  ssl_certificate_content = data.local_file.cert.content
		}
		`,
		certfile,
		testingInfo.Username,
		testingInfo.Password,
		testingInfo.Endpoint,
	)
}
