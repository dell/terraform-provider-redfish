provider "redfish" {
  redfish_server { 
        endpoint = "192.168.1.1"
        user = "root"
        password = "calvin"
        ssl_insecure = true
  }
  redfish_server { 
        endpoint = "192.168.1.2"
        user = "root"
        password = "calvin"
        ssl_insecure = true
  }
}

resource "redfish_storage_volume" "volume" {
    storage_controller_id = "RAID.Integrated.1-1"
    volume_name = "TerraformVol"
    volume_type = "Mirrored"
    volume_disks = ["Solid State Disk 0:1:0", "Solid State Disk 0:1:1"]
    settings_apply_time = "Immediate"
    // settings_apply_time = "OnReset"
}

