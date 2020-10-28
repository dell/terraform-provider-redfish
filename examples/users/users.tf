provider "redfish" {
  redfish_server { 
        endpoint = "https://localhost:5007"
        user = "root"
        password = "calvin"
        ssl_insecure = true
  }
  redfish_server { 
        endpoint = "https://localhost:5008"
        user = "root"
        password = "calvin"
        ssl_insecure = true
  }
}

resource "redfish_user_account" "users" {
    username = "mike"
    password = "test1234"
    enabled = true
}