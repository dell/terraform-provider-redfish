package provider

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

// Test to create redfish virtual media - Positive
func TestAccRedfishVirtualMedia_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRedfishResourceVirtualMediaConfig(
					creds,
					"virtual_media",
					"http://linuxlib.us.dell.com/pub/redhat/RHEL8/8.9/latest/BaseOS/x86_64/iso/RHEL-8.9.0-20231023.21-x86_64-boot.iso",
					true,
					"HTTP",
					"Stream"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("redfish_virtual_media.virtual_media", "image", "http://linuxlib.us.dell.com/pub/redhat/RHEL8/8.9/latest/BaseOS/x86_64/iso/RHEL-8.9.0-20231023.21-x86_64-boot.iso"),
					resource.TestCheckResourceAttr("redfish_virtual_media.virtual_media", "inserted", "true"),
				),
			},
		},
	})
}

// Test to create redfish virtual media with invalid image extension - Negative
func TestAccRedfishVirtualMedia_InvalidImage_Negative(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRedfishResourceVirtualMediaConfig(
					creds,
					"virtual_media",
					"http://linuxlib.us.dell.com/pub/redhat/RHEL8/8.9/latest/BaseOS/x86_64/iso/RHEL-8.9.0-20231023.21-x86_64-boot.is",
					true,
					"HTTP",
					"Stream"),
				ExpectError: regexp.MustCompile("Couldn't mount Virtual Media"),
			},
		},
	})
}

// Test to create redfish virtual media with invalid transfer method - Negative
func TestAccRedfishVirtualMedia_InvalidTransferMethod_Negative(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRedfishResourceVirtualMediaConfig(
					creds,
					"virtual_media",
					"http://linuxlib.us.dell.com/pub/redhat/RHEL8/8.9/latest/BaseOS/x86_64/iso/RHEL-8.9.0-20231023.21-x86_64-boot.iso",
					true,
					"HTTP",
					"Upload"),
				ExpectError: regexp.MustCompile("Couldn't mount Virtual Media"),
			},
		},
	})
}

// Test to create redfish virtual media with invalid transfer protocol type - Negative
func TestAccRedfishVirtualMedia_InvalidTransferProtocol_Negative(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRedfishResourceVirtualMediaConfig(
					creds,
					"virtual_media",
					"http://linuxlib.us.dell.com/pub/redhat/RHEL8/8.9/latest/BaseOS/x86_64/iso/RHEL-8.9.0-20231023.21-x86_64-boot.iso",
					true,
					"HTTPS",
					"Stream"),
				ExpectError: regexp.MustCompile("Couldn't mount Virtual Media"),
			},
		},
	})
}

// Test to create redfish virtual media when no file shares are available to mount - Negative
func TestAccRedfishVirtualMediaNoMediaNegative_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRedfishResourceVirtualMediaConfig(
					creds,
					"virtual_media1",
					"http://linuxlib.us.dell.com/pub/redhat/RHEL8/8.9/latest/BaseOS/x86_64/iso/RHEL-8.9.0-20231023.21-x86_64-boot.iso",
					true,
					"HTTP",
					"Stream") +
					testAccRedfishResourceVirtualMediaConfigDependency(
						creds,
						"virtual_media2",
						"http://linuxlib.us.dell.com/pub/RHEL8/8.9/latest/BaseOS/x86_64/iso/RHEL-8.9.0-20231023.21-x86_64-boot.iso",
						true,
						"HTTP",
						"Stream",
						"redfish_virtual_media.virtual_media1") +
					testAccRedfishResourceVirtualMediaConfigDependency(
						creds,
						"virtual_media3",
						"http://linuxlib.us.dell.com/pub/RHEL8/8.9/latest/BaseOS/x86_64/iso/RHEL-8.9.0-20231023.21-x86_64-boot.iso",
						true,
						"HTTP",
						"Stream",
						"redfish_virtual_media.virtual_media2"),
				ExpectError: regexp.MustCompile("There are no Virtual Medias to mount"),
			},
		},
	})
}

// Test to create redfish virtual media on iDRAC 5.x - Positive
func TestAccRedfishVirtualMediaServer2_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRedfishResourceVirtualMediaConfigServer5x(
					creds,
					"virtual_media",
					"http://linuxlib.us.dell.com/pub/redhat/RHEL8/8.8/BaseOS/x86_64/os/images/efiboot.img",
					true,
					"HTTP",
					"Stream"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("redfish_virtual_media.virtual_media", "image", "http://linuxlib.us.dell.com/pub/redhat/RHEL8/8.8/BaseOS/x86_64/os/images/efiboot.img"),
					resource.TestCheckResourceAttr("redfish_virtual_media.virtual_media", "inserted", "true"),
				),
			},
		},
	})
}

// Test to create redfish virtual media with invalid transfer protocol type on iDRAC 5.x - Negative
func TestAccRedfishVirtualMediaServer2_InvalidTransferProtocol_Negative(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRedfishResourceVirtualMediaConfigServer5x(
					creds,
					"virtual_media",
					"http://linuxlib.us.dell.com/pub/redhat/RHEL8/8.8/BaseOS/x86_64/os/images/efiboot.img",
					true,
					"HTTPS",
					"Stream"),
				ExpectError: regexp.MustCompile("Couldn't mount Virtual Media"),
			},
		},
	})
}

// Test to update redfish virtual media with invalid transfer protocol type on iDRAC 5.x - Negative
func TestAccRedfishVirtualMediaServer2Update_InvalidTransferProtocol_Negative(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRedfishResourceVirtualMediaConfigServer5x(
					creds,
					"virtual_media",
					"http://linuxlib.us.dell.com/pub/redhat/RHEL8/8.8/BaseOS/x86_64/os/images/efiboot.img",
					true,
					"HTTP",
					"Stream"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("redfish_virtual_media.virtual_media", "image", "http://linuxlib.us.dell.com/pub/redhat/RHEL8/8.8/BaseOS/x86_64/os/images/efiboot.img"),
					resource.TestCheckResourceAttr("redfish_virtual_media.virtual_media", "inserted", "true"),
				),
			},
			{
				Config: testAccRedfishResourceVirtualMediaConfigServer5x(
					creds,
					"virtual_media",
					"http://linuxlib.us.dell.com/pub/redhat/RHEL8/8.8/BaseOS/x86_64/os/images/efiboot.img",
					true,
					"HTTPS",
					"Stream"),
				ExpectError: regexp.MustCompile("Couldn't mount Virtual Media"),
			},
		},
	})
}

// Test to update redfish virtual media - Negative
func TestAccRedfishVirtualMediaUpdate_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRedfishResourceVirtualMediaConfig(
					creds,
					"virtual_media",
					"http://linuxlib.us.dell.com/pub/redhat/RHEL8/8.9/latest/BaseOS/x86_64/iso/RHEL-8.9.0-20231023.21-x86_64-boot.iso",
					true,
					"HTTP",
					"Stream"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("redfish_virtual_media.virtual_media", "image", "http://linuxlib.us.dell.com/pub/redhat/RHEL8/8.9/latest/BaseOS/x86_64/iso/RHEL-8.9.0-20231023.21-x86_64-boot.iso"),
					resource.TestCheckResourceAttr("redfish_virtual_media.virtual_media", "inserted", "true"),
				),
			},
			{
				Config: testAccRedfishResourceVirtualMediaConfig(
					creds,
					"virtual_media",
					"http://linuxlib.us.dell.com/pub/redhat/RHEL8/8.9/latest/BaseOS/x86_64/iso/RHEL-8.9.0-20231023.21-x86_64-boot.iso",
					false,
					"HTTP",
					"Stream"),
				ExpectError: regexp.MustCompile("Provider produced inconsistent result after apply"),
			},
		},
	})
}

// Test to update redfish virtual media with invalid image - Negative
func TestAccRedfishVirtualMediaUpdate_InvalidImage_Negative(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRedfishResourceVirtualMediaConfig(
					creds,
					"virtual_media",
					"http://linuxlib.us.dell.com/pub/redhat/RHEL8/8.9/latest/BaseOS/x86_64/iso/RHEL-8.9.0-20231023.21-x86_64-boot.iso",
					true,
					"HTTP",
					"Stream"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("redfish_virtual_media.virtual_media", "image", "http://linuxlib.us.dell.com/pub/redhat/RHEL8/8.9/latest/BaseOS/x86_64/iso/RHEL-8.9.0-20231023.21-x86_64-boot.iso"),
					resource.TestCheckResourceAttr("redfish_virtual_media.virtual_media", "inserted", "true"),
				),
			},
			{
				Config: testAccRedfishResourceVirtualMediaConfig(
					creds,
					"virtual_media",
					"http://linuxlib.us.dell.com/pub/redhat/RHEL8/8.8/BaseOS/x86_64/iso/RHEL-8.8.0-20230411.3-x86_64-boot.is",
					true,
					"HTTP",
					"Stream"),
				ExpectError: regexp.MustCompile("Couldn't mount Virtual Media"),
			},
		},
	})
}

// Test to update redfish virtual media with invalid transfer method - Negative
func TestAccRedfishVirtualMediaUpdate_InvalidTransferMethod_Negative(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRedfishResourceVirtualMediaConfig(
					creds,
					"virtual_media",
					"http://linuxlib.us.dell.com/pub/redhat/RHEL8/8.9/latest/BaseOS/x86_64/iso/RHEL-8.9.0-20231023.21-x86_64-boot.iso",
					true,
					"HTTP",
					"Stream"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("redfish_virtual_media.virtual_media", "image", "http://linuxlib.us.dell.com/pub/redhat/RHEL8/8.9/latest/BaseOS/x86_64/iso/RHEL-8.9.0-20231023.21-x86_64-boot.iso"),
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
					"Upload"),
				ExpectError: regexp.MustCompile("Couldn't mount Virtual Media"),
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

func testAccRedfishResourceVirtualMediaConfigDependency(testingInfo TestingServerCredentials,
	resource_name string,
	image string,
	write_protected bool,
	transfer_protocol_type string,
	transfer_method string,
	depends_on string) string {
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

		  depends_on = [
			"%s"
		  ]
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
		depends_on,
	)
}
