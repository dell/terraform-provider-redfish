package redfish

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

// Test to create redfish virtual media - Positive
func TestAccRedfishVirtualMedia_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccRedfishResourceVirtualMediaConfig(
					creds,
					"virtual_media",
					"http://linuxlib.us.dell.com/pub/redhat/RHEL8/8.8/BaseOS/x86_64/os/images/efiboot.img",
					true,
					"HTTP",
					"Stream"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("redfish_virtual_media.virtual_media", "image", "http://linuxlib.us.dell.com/pub/redhat/RHEL8/8.8/BaseOS/x86_64/os/images/efiboot.img"),
					resource.TestCheckResourceAttr("redfish_virtual_media.virtual_media", "inserted", "true"),
				),
			},
		},
	})
}

// Test to create virtual media with invalid image path - Negative
func TestAccRedfishVirtualMediaInvalid_basic(t *testing.T) {

	resource.Test(t, resource.TestCase{
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccRedfishResourceVirtualMediaConfig(
					creds,
					"virtual_media",
					"http://linuxlib.com/pub/redhat/RHEL8/8.8/BaseOS/x86_64/os/images/efiboot.img",
					true,
					"HTTP",
					"Stream"),
				ExpectError: regexp.MustCompile("Unable to locate the ISO or IMG image file or folder"),
			},
		},
	})
}

// Test to create redfish virtual media when no file shares are available to mount - Negative
func TestAccRedfishVirtualMediaNoMediaNegative_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccRedfishResourceVirtualMediaConfig(
					creds,
					"virtual_media1",
					"http://linuxlib.us.dell.com/pub/redhat/RHEL8/8.8/BaseOS/x86_64/iso/RHEL-8.8.0-20230411.3-x86_64-boot.iso",
					true,
					"HTTP",
					"Stream") +
					testAccRedfishResourceVirtualMediaConfig(
						creds,
						"virtual_media2",
						"http://linuxlib.us.dell.com/pub/redhat/RHEL8/8.8/BaseOS/x86_64/os/images/efiboot.img",
						true,
						"HTTP",
						"Stream") +
					testAccRedfishResourceVirtualMediaConfig(
						creds,
						"virtual_media3",
						"http://linuxlib.us.dell.com/pub/redhat/RHEL8/8.8/BaseOS/x86_64/os/images/efiboot.img",
						true,
						"HTTP",
						"Stream"),
				ExpectError: regexp.MustCompile("There are no Virtual Medias to mount"),
			},
		},
	})
}

// Test to create redfish virtual media on iDRAC 5.x - Positive
func TestAccRedfishVirtualMediaServer2_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccRedfishResourceVirtualMediaConfigServer5x(
					creds,
					"virtual_media",
					"http://linuxlib.us.dell.com/pub/redhat/RHEL8/8.8/BaseOS/x86_64/os/images/efiboot.img",
					true,
					"HTTP",
					"Stream"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("redfish_virtual_media.virtual_media", "image", "http://linuxlib.us.dell.com/pub/redhat/RHEL8/8.8/BaseOS/x86_64/os/images/efiboot.img"),
					resource.TestCheckResourceAttr("redfish_virtual_media.virtual_media", "inserted", "true"),
				),
			},
		},
	})
}

// Test to update redfish virtual media - Positive
func TestAccRedfishVirtualMediaUpdate_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccRedfishResourceVirtualMediaConfig(
					creds,
					"virtual_media",
					"http://linuxlib.us.dell.com/pub/redhat/RHEL8/8.8/BaseOS/x86_64/os/images/efiboot.img",
					true,
					"HTTP",
					"Stream"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("redfish_virtual_media.virtual_media", "image", "http://linuxlib.us.dell.com/pub/redhat/RHEL8/8.8/BaseOS/x86_64/os/images/efiboot.img"),
					resource.TestCheckResourceAttr("redfish_virtual_media.virtual_media", "inserted", "true"),
				),
			},
			{
				Config: testAccRedfishResourceVirtualMediaConfig(
					creds,
					"virtual_media",
					"http://linuxlib.us.dell.com/pub/redhat/RHEL8/8.8/BaseOS/x86_64/iso/RHEL-8.8.0-20230411.3-x86_64-boot.iso",
					true,
					"HTTP",
					"Stream"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("redfish_virtual_media.virtual_media", "image", "http://linuxlib.us.dell.com/pub/redhat/RHEL8/8.8/BaseOS/x86_64/iso/RHEL-8.8.0-20230411.3-x86_64-boot.iso"),
					resource.TestCheckResourceAttr("redfish_virtual_media.virtual_media", "inserted", "true"),
				),
			},
		},
	})
}

func testAccRedfishResourceVirtualMediaConfig(testingInfo TestingServerCredentials,
	resource_name string,
	image string,
	write_protected bool,
	transfer_protocol_type string,
	transfer_method string) string {
	return fmt.Sprintf(`
		
		resource "redfish_virtual_media" "%s" {
		
		  redfish_server {
			user = "%s"
			password = "%s"
			endpoint = "https://%s"
			ssl_insecure = true
		  }
		  image = "%s"
		  write_protected = %t
		  transfer_protocol_type = "%s"
		  transfer_method = "%s"
		}
		`,
		resource_name,
		testingInfo.Username,
		testingInfo.Password,
		testingInfo.Endpoint,
		image,
		write_protected,
		transfer_protocol_type,
		transfer_method,
	)
}

func testAccRedfishResourceVirtualMediaConfigServer5x(testingInfo TestingServerCredentials,
	resource_name string,
	image string,
	write_protected bool,
	transfer_protocol_type string,
	transfer_method string) string {
	return fmt.Sprintf(`
		
		resource "redfish_virtual_media" "%s" {
		
		  redfish_server {
			user = "%s"
			password = "%s"
			endpoint = "https://%s"
			ssl_insecure = true
		  }
		  image = "%s"
		  write_protected = %t
		  transfer_protocol_type = "%s"
		  transfer_method = "%s"
		}
		`,
		resource_name,
		testingInfo.Username,
		testingInfo.Password,
		testingInfo.Endpoint2,
		image,
		write_protected,
		transfer_protocol_type,
		transfer_method,
	)
}
