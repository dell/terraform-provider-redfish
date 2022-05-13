# Information about Redfish Terraform Provider
This guide will explain different parts of the provider and will give an overview about how the provider is built, to onboard quicker developers that might be interested in this project.  

## 1. Provider's way of operation
When you think of Terraform, normally operators tend to think that the way a provider connects with a cloud provider is via a single endpoint. Well, actually that's the way it works. Cloud providers provide an endpoint and operators point to that endpoint when configuring terraform.  
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

## How we overcome this

Normally the provider is initialized in the provider block, giving it your cloud credentials to deal with the infrastructure. Something like this:
~~~	
provider "aws" {
	region     = "eu-west-1"
	access_key = "myaccesskey"
	secret_key = "mysecretkey"
}
~~~
When that is done, then operators would start writing the resources they want to deploy in those regions.  

  
With this **terraform redfish provider** a different approach had to be followed since there are multiple endpoints. What has been done (and kudos to Kyriakos Oikonomakos from Hashicorp for proposing this) was to initialize the client at the resource level. This allows operators to manage different servers from one central point. Take a look into this example:  
    
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
  
By doing this, operators create two users on two different servers using this provider and the Redfish API.  
*Remember, in every CRUD operation, the client must be initialized.*

## Overwriting client credentials
There might be scenarios where operators have the same credentials for all machines they want to manage. In that case they don't need to repeatedly write the *user* and *password* for all servers. They can write their credentials at the provider block level.
~~~
provider "redfish" {
    user = "root"
    password = "calvin"
}
~~~

After the user specifies their credentials, they will next need to define the infrastructure. Instead of defining credentials for each endpoint they need only provide the *endpoint* and *ssl_insecure* values:

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

Terraform will always use the most specific client values. In the case client credentials are defined at both the provider block and resource level, **the credentials defined at the resource level** will be used. 
  