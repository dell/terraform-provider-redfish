provider "redfish" {
  redfish_endpoint = "https://localhost:5000"
  user = "root"
  password = "calvin"
  ssl_insecure = true
}

resource "redfish_storage_volume" "volume" {
    storage_controller = "RAID.Integrated.1-1"
    volume_name = "MyVol"
    raid_level = "Mirrored"
    settings_apply_time = "Immediate"
}

