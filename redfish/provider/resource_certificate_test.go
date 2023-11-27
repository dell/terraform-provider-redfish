package provider

import (
	"fmt"
	"os"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

// test redfish bios settings
func TestAccRedfishCertificate_basic(t *testing.T) {
	var valid_cert = os.Getenv("VALID_CERT")
	var invalid_cert = os.Getenv("INVALID_CERT")
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
			endpoint = "https://%s"
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
