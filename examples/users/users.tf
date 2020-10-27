provider "redfish" {
  redfish_server { 
        endpoint = "https://localhost:5000"
        user = "root"
        password = "calvin"
        ssl_insecure = true
  }
  /*redfish_endpoints { 
        redfish_endpoint = "https://192.168.1.2:5000"
        user = "root"
        password = "calvin"
        ssl_insecure = true
  }*/
}

resource "redfish_user_account" "users" {
    username = "mike"
    password = "test1234"
    enabled = true
}