# Information about Redfish Terraform Provider
This guide will explain different parts of the provider and will give an overview about how the provider is built, to onboard quicker developers that might be interested in this project.  

## 1. Provider's way of operation
When you think of Terraform, normally users tend to think that the way a provider connects with a cloud provider is via a single endpoint. Well actually that's the way it works. Cloud providers provider and endpoint and users point to that endpoint when configuring terraform.  
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

With the Redfish Terraform Provider, that operating model have been changed because of the way we interact with the infrastructure (Redfish endpoints).  
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

This comes with some challenges when it comes to develop a terraform provider, because Terraform hasn't been thought for that. There were two ways this provider could have been implemented:
  - Create different alias for each redfish endpoint in the .tf files.
  - **Create a provider that supports different redfish endpoints and apply the defined resources to all of them**.
The chosen approach was the second one, as .tf files are less verbose and practically the provider will be much useful. *This doesn't mean that users cannot use aliases for different kind of servers.*

## 2. Provider declaration
The provider schema is the following:
![](provider/images/provider_schema.png)  
As it can be seen, the entry *redfish_server* has been set to a *schema.TypeList*. This gives the provider the ability of accepting different *redfish_server* blocks in the .tf file.  
Inside each *redfish_server* block, users will define the *user*, *password*, *endpoint* and *ssl_insecure* values. This means that within the same provider, different servers can be provisioned.  
  
In terms of the .tf file, this could be an example:
![](provider/images/provider_declaration.png)  
As shown in the figure, the provider block *redfish* contains two *redfish_server* blocks inside it. This means that two servers will be managed using that provider instance. From now on, all resources specified for this provider will be applied to **all** redfish endpoints defined within it.

## 3. How this is implemented