provider "redfish" {}

data "redfish_virtual_media" "vm" {
    for_each = var.rack1

    redfish_server {
        user = each.value.user
        password = each.value.password
        endpoint = each.value.endpoint
        ssl_insecure = each.value.ssl_insecure
    }      
}

#output "virtual_media" {
#    value = data.redfish_virtual_media.vm
#}