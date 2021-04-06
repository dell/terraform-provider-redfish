terraform {
    required_providers {
        redfish = {
            source = "dell/redfish"
        }
    }
}

resource "redfish_storage_volume" "volume" {
    for_each = var.rack1

    redfish_server {
        user = each.value.user
        password = each.value.password
        endpoint = each.value.endpoint
        ssl_insecure = each.value.ssl_insecure
    } 
    
    storage_controller_id = "RAID.Integrated.1-1"
    volume_name = "TerraformVol"
    volume_type = "Mirrored"
    drives = ["Solid State Disk 0:1:0", "Solid State Disk 0:1:1"]
    settings_apply_time = "Immediate"
    // settings_apply_time = "OnReset"
}

