package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDataSourceRedfishStorageVolume_ByName(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceRedfishStorageVolumeConfig_ByName(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						"data.redfish_storage_volume.test",
						"name",
						"TestVolume-1",
					),
					resource.TestCheckResourceAttrSet(
						"data.redfish_storage_volume.test",
						"id",
					),
					resource.TestCheckResourceAttrSet(
						"data.redfish_storage_volume.test",
						"capacity_bytes",
					),
					resource.TestCheckResourceAttr(
						"data.redfish_storage_volume.test",
						"raid_type",
						"RAID1",
					),
					resource.TestCheckResourceAttr(
						"data.redfish_storage_volume.test",
						"encrypted",
						"true",
					),
				),
			},
		},
	})
}

func testAccDataSourceRedfishStorageVolumeConfig_ByName() string {
	return `
data "redfish_storage_volume" "test" {
  redfish_server {
    endpoint     = "` + testAccRedfishEndpoint + `"
    username     = "` + testAccRedfishUsername + `"
    password     = "` + testAccRedfishPassword + `"
    ssl_insecure = true
  }

  filter {
    name = "TestVolume-1"
  }
}

output "volume_id" {
  value = data.redfish_storage_volume.test.id
}
`
}

func TestAccDataSourceRedfishStorageVolume_ByID(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceRedfishStorageVolumeConfig_ByID(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						"data.redfish_storage_volume.test",
						"id",
						"Disk.Virtual.0:RAID.Integrated.1-1",
					),
					resource.TestCheckResourceAttrSet(
						"data.redfish_storage_volume.test",
						"name",
					),
					resource.TestCheckResourceAttrSet(
						"data.redfish_storage_volume.test",
						"capacity_bytes",
					),
				),
			},
		},
	})
}

func testAccDataSourceRedfishStorageVolumeConfig_ByID() string {
	return `
data "redfish_storage_volume" "test" {
  redfish_server {
    endpoint     = "` + testAccRedfishEndpoint + `"
    username     = "` + testAccRedfishUsername + `"
    password     = "` + testAccRedfishPassword + `"
    ssl_insecure = true
  }

  filter {
    id = "Disk.Virtual.0:RAID.Integrated.1-1"
  }
}
`
}

func TestAccDataSourceRedfishStorageVolume_AllAttributes(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceRedfishStorageVolumeConfig_AllAttributes(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.redfish_storage_volume.test", "id"),
					resource.TestCheckResourceAttrSet("data.redfish_storage_volume.test", "name"),
					resource.TestCheckResourceAttrSet("data.redfish_storage_volume.test", "capacity_bytes"),
					resource.TestCheckResourceAttrSet("data.redfish_storage_volume.test", "raid_type"),
					resource.TestCheckResourceAttrSet("data.redfish_storage_volume.test", "status"),
					resource.TestCheckResourceAttrSet("data.redfish_storage_volume.test", "encrypted"),
					resource.TestCheckResourceAttrSet("data.redfish_storage_volume.test", "controller_id"),
					resource.TestCheckResourceAttrSet("data.redfish_storage_volume.test", "physical_disks.#"),
				),
			},
		},
	})
}

func testAccDataSourceRedfishStorageVolumeConfig_AllAttributes() string {
	return `
data "redfish_storage_volume" "test" {
  redfish_server {
    endpoint     = "` + testAccRedfishEndpoint + `"
    username     = "` + testAccRedfishUsername + `"
    password     = "` + testAccRedfishPassword + `"
    ssl_insecure = true
  }

  filter {
    name = "TestVolume-2"
  }
}
`
}

// Unit tests (without TF_ACC)

func TestStorageVolumeDataSource_Schema(t *testing.T) {
	ds := NewStorageVolumeDataSource()
	schema := storageVolumeDataSourceSchema()

	// Verify required attributes
	if schema.Attributes["redfish_server"].IsRequired() != true {
		t.Error("redfish_server should be required")
	}

	if schema.Attributes["filter"].IsRequired() != true {
		t.Error("filter should be required")
	}

	// Verify computed attributes
	if schema.Attributes["id"].IsComputed() != true {
		t.Error("id should be computed")
	}

	if schema.Attributes["name"].IsComputed() != true {
		t.Error("name should be computed")
	}

	// Verify sensitive attribute
	redfishServerAttr := schema.Attributes["redfish_server"].(schema.SingleNestedAttribute)
	passwordAttr := redfishServerAttr.Attributes["password"].(schema.StringAttribute)
	if passwordAttr.IsSensitive() != true {
		t.Error("password should be sensitive")
	}
}
