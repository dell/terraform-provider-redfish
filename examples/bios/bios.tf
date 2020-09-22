resource "redfish_bios" "bios" {
  attributes = {
    "NumLock" = "On"
  }
  settings_apply_time = "OnReset"
}

data "redfish_bios" "bios" {
}

output "bios_attributes" {
  value = "${data.redfish_bios.bios.attributes}"
}
