terraform {
    required_providers {
        redfish = {
            version = "~> 0.2.0"
            source = "dell.com/dell/redfish"
        }
    }
}

provider "redfish" {}

resource "redfish_virtual_media" "vm" {
    for_each = var.rack1

    redfish_server {
        user = each.value.user
        password = each.value.password
        endpoint = each.value.endpoint
        ssl_insecure = each.value.ssl_insecure
    }

    virtual_media_id = "CD"
    image = "http://web.svd-miguel02.dell-atc.lan/centos/7.6.1810/image.iso"
}
