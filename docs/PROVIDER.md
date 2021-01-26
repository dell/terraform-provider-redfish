# Information about Redfish Terraform Provider
This guide will explain different parts of the provider and will give an overview about how the provider is built, to onboard quicker developers that might be interested in this project.  

## 1. Provider's way of operation
When you think of Terraform, normally operators tend to think that the way a provider connects with a cloud provider is via a single endpoint. Well, actually that's the way it works. Cloud providers provide and endpoint and operators point to that endpoint when configuring terraform.  
~~~
  +-----------------+
  | Cloud provider  |
  +-------+---------+
          ^
          |
          |
+---------+----------+
| Terraform provider |
+--------------------+
~~~

With the **Redfish Terraform Provider**, that operating model has been changed because of the way the provider interacts with the infrastructure (Redfish endpoints).
In a regular scenario (for instance a datacenter), operators don't just have one endpoint, but a bunch of them. Each redfish endpoint corresponds to each physical server.  
~~~
+------------------+     +------------------+      ...N        +------------------+
| PowerEdge Server |     | PowerEdge Server |  +-----------+   | PowerEdge Server |
| with Redfish API |     | with Redfish API |                  | with Redfish API |
+-------+----------+     +--------+---------+                  +---------+--------+
        ^                         ^                   ^                  ^
        |                         |                   |                  |
        +-------------------------+----------+--------+------------------+
                                             |
                                             |
                                  +----------+---------+
                                  | Terraform Provider |
                                  +--------------------+

~~~

## How this is overcomed
Normally the provider is initialized in the provider block, giving it your cloud credentials to deal with the infra. Something like this:
~~~	
provider "aws" {
	region     = "eu-west-1"
	access_key = "myaccesskey"
	secret_key = "mysecretkey"
}
~~~
When that is done, then operators would start writing the resources wanted to be deployed in there.  

  
With this **terraform redfish provider** a different approach had to be followed since there are multiple endpoints. What has been done (and kudos to Kyriakos Oikonomakos from Hashicorp to propose this) was to initialize the client at resource level. This allow operators to manage different servers in one shot. Take a look into this example:  
    
users.tf
~~~
provider "redfish" {}

resource "redfish_user_account" "rr" {
    for_each = var.rack1

    redfish_server {
        user = each.value.user
        password = each.value.password
        endpoint = each.value.endpoint
        ssl_insecure = each.value.ssl_insecure
    }      

    username = "mike"
    password = "test1234"
    enabled = true
}
~~~
  

terraform.tfvars
~~~
rack1 = {
    "my-server-1" = {
        user = "root"
        password = "calvin"
        endpoint = "https://my-server-1.myawesomecompany.org"
        ssl_insecure = true
    },
    "my-server-2" = {
        user = "root"
        password = "calvin"
        endpoint = "https://my-server-2.myawesomecompany.org"
        ssl_insecure = true
    },
}
~~~
  
By following this, operators will be creating two users in two different servers, using this provider and the Redfish API.  
*Remember, in every CRUD operation, the client must be initialized.*

## Overwriding client credentials
There might be scenarios where operators have the same credentials for all machines to be managed. In that case they don't need to write over and over again the *user* and *password* for all servers. There credentials can be written at provider block level. 
~~~
provider "redfish" {
    user = "root"
    password = "calvin"
}
~~~
And then, when defining the infrastructure, just add the *endpoint* and *ssl_insecure* values:

~~~
rack1 = {
    "my-server-1" = {
        endpoint = "https://my-server-1.myawesomecompany.org"
        ssl_insecure = true
    },
    "my-server-2" = {
        endpoint = "https://my-server-2.myawesomecompany.org"
        ssl_insecure = true
    },
}
~~~

Also, the rule for this is to use the most specific client values, so in case the client credentials are placed at both, provider block and resource level, **the ones defined at resource level** will be used. 
  