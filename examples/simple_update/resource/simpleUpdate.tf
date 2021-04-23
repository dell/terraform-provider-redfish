terraform {
    required_providers {
        redfish = {
            source = "dell/redfish"
        }
    }
}

resource "redfish_simple_update" "update" {
    for_each = var.rack1

    redfish_server {
        user = each.value.user
        password = each.value.password
        endpoint = each.value.endpoint
        ssl_insecure = each.value.ssl_insecure
    } 
    
    transfer_protocol = "HTTP"
    target_firmware_image = "/home/mikeletux/Downloads/BIOS_FXC54_WN64_1.15.0.EXE"
    reset_type = "ForceRestart"
    // reset_timeout = 120 // If not set, by default will be 120s
    // simple_update_job_timeout = 1200 // If not set, by default will be 1200s
}
