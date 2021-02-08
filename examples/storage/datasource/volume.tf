provider "redfish" {}

data "redfish_storage_volume" "volume" {
    for_each = var.rack1

    redfish_server {
        user = each.value.user
        password = each.value.password
        endpoint = each.value.endpoint
        ssl_insecure = each.value.ssl_insecure
    }
}

#output "storage_volume" {
#    value = data.redfish_storage_volume.volume
#}