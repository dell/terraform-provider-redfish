package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccRedfishStorageDataSource_fetch(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRedfishDataSourceStorageConfig(creds),
			},
			{
				Config: testAccStorageDatasourceWithControllerID(creds),
			},
			{
				Config: testAccStorageDatasourceWithControllerName(creds),
			},
			{
				Config: testAccStorageDatasourceWithBothConfig(creds),
			},
		},
	})
}

// controller_names = ["PERC H730P Mini"]
// controller_ids = ["AHCI.Embedded.2-1"]

func testAccRedfishDataSourceStorageConfig(testingInfo TestingServerCredentials) string {
	return fmt.Sprintf(`
	data "redfish_storage" "storage" {	  
		redfish_server {
		  user         = "%s"
		  password     = "%s"
		  endpoint     = "https://%s"
		  ssl_insecure = true
		}
	  }
		`,
		testingInfo.Username,
		testingInfo.Password,
		testingInfo.Endpoint,
	)
}

func testAccStorageDatasourceWithControllerID(testingInfo TestingServerCredentials) string {
	return fmt.Sprintf(`
	data "redfish_storage" "storage" {	  
		redfish_server {
		  user         = "%s"
		  password     = "%s"
		  endpoint     = "https://%s"
		  ssl_insecure = true
		}
		controller_ids = ["AHCI.Embedded.2-1"]
	  }
		`,
		testingInfo.Username,
		testingInfo.Password,
		testingInfo.Endpoint,
	)
}

func testAccStorageDatasourceWithControllerName(testingInfo TestingServerCredentials) string {
	return fmt.Sprintf(`
	data "redfish_storage" "storage" {	
		redfish_server {
		  user         = "%s"
		  password     = "%s"
		  endpoint     = "https://%s"
		  ssl_insecure = true
		}
		controller_names = ["PERC H730P Mini"]
	  }
		`,
		testingInfo.Username,
		testingInfo.Password,
		testingInfo.Endpoint,
	)
}

func testAccStorageDatasourceWithBothConfig(testingInfo TestingServerCredentials) string {
	return fmt.Sprintf(`
	data "redfish_storage" "storage" {	
		redfish_server {
		  user         = "%s"
		  password     = "%s"
		  endpoint     = "https://%s"
		  ssl_insecure = true
		}
		controller_ids = ["AHCI.Embedded.2-1"]
		controller_names = ["PERC H730P Mini"]
	  }
		`,
		testingInfo.Username,
		testingInfo.Password,
		testingInfo.Endpoint,
	)
}
