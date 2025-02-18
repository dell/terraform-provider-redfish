/*
Copyright (c) 2024 Dell Inc., or its subsidiaries. All Rights Reserved.

Licensed under the Mozilla Public License Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://mozilla.org/MPL/2.0/


Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

# kerberod file is not supported by 17G and this configuration works only for below 17G
data "local_file" "kerberos" {
  # this is the path to the kerberos keytab file that we want to upload.
  # this file must be base64 encoded format
  filename = "/root/directoryservice/new/terraform-provider-redfish/test-data/kerberos_file.txt"
}

# redfish_directory_service_auth_provider Terraform resource is used to configure Directory Service Auth Provider Active Directory and LDAP Service
# Available action: Create, Update (Active Directory, LDAP)
# Active Directory (Create, Update): remote_role_mapping, service_addresses, service_enabled,authentication, active_directory_attributes
# LDAP (Create, Update): remote_role_mapping, service_addresses, service_enabled,ldap_service, ldap_attributes
resource "redfish_directory_service_auth_provider" "ds_auth" {
  for_each = var.rack1

  redfish_server {
    # Alias name for server BMCs. The key in provider's `redfish_servers` map
    # `redfish_alias` is used to align with enhancements to password management.
    # When using redfish_alias, provider's `redfish_servers` is required.
    redfish_alias = each.key

    user         = each.value.user
    password     = each.value.password
    endpoint     = each.value.endpoint
    ssl_insecure = each.value.ssl_insecure
  }

  #Note: `active_directory` is mutually inclusive with `active_directory_attributes`.
  #Note: `ldap` is mutually inclusive with `ldap_attributes`.
  #Note: `active_directory` is mutually exclusive with `ldap`.
  #Note: `active_directory_attributes` is mutually exclusive with `ldap_attributes`. 
  active_directory = {
    directory = {
      # remote_role_mapping = [
      #     {
      #         local_role = "Administrator",
      #         remote_group = "idracgroup"
      #     }
      # ],
      # To Update service addresses for 17G please provide configuration in active_directory_attributes
      # This configuration will be working once the issue for 17G get resolved 
      # service_addresses = [
      #     "yulanadhost11.yulan.pie.lab.emc.com"
      #  ],
      service_enabled = true,
      # authentication configuration works for below 17G server, 17G server do not support kerberos
      authentication = {
        kerberos_key_tab_file = data.local_file.kerberos.content
      }
    }
  }

  active_directory_attributes = {
    "ActiveDirectory.1.AuthTimeout"          = "120",
    "ActiveDirectory.1.CertValidationEnable" = "Enabled",
    "ActiveDirectory.1.DCLookupEnable"       = "Enabled",

    # DomainController can be configured when DCLookupEnable is Disabled
    # To update service addresses for 17G please provide below configuration
    #"ActiveDirectory.1.DomainController1"= "yulanadhost1.yulan.pie.lab.emc.com",
    #"ActiveDirectory.1.DomainController2"= "yulanadhost2.yulan.pie.lab.emc.com",
    #"ActiveDirectory.1.DomainController3"= "yulanadhost3.yulan.pie.lab.emc.com",

    # RacName and RacDomain can be configured when Schema is Extended Schema which is supported by below 17G server
    "ActiveDirectory.1.RacDomain" = "test",
    "ActiveDirectory.1.RacName"   = "test",

    # SSOEnable configuration can be done for below 17G server and it's not supported by 17G server
    # if SSOEnable is Enabled make sure ActiveDirectory Service is enabled and valid kerberos_key_tab_file is provided
    "ActiveDirectory.1.SSOEnable" = "Disabled",

    # Schema can be Extended Schema or Standard Schema
    # Schema configuration can be done for below 17G, and this configuration is not supported by 17G 
    "ActiveDirectory.1.Schema" = "Extended Schema",
    "UserDomain.1.Name"        = "yulan.pie.lab.emc.com",

    # DCLookupByUserDomain must be configured when DCLookupEnable is enabled 
    "ActiveDirectory.1.DCLookupByUserDomain" : "Enabled",

    # DCLookupDomainName must be configured when DCLookupByUserDomain is Disabled and DCLookupEnable is Enabled
    #"ActiveDirectory.1.DCLookupDomainName"="test", 

    #"ActiveDirectory.1.GCLookupEnable" = "Disabled"

    # for 17G below configuration can be performed without schema configuration
    # at least any one from GlobalCatalog1,GlobalCatalog2,GlobalCatalog3 must be configured when Schema is Standard and GCLookupEnable is Disabled
    # "ActiveDirectory.1.GlobalCatalog1" = "yulanadhost11.yulan.pie.lab.emc.com",
    # "ActiveDirectory.1.GlobalCatalog2" = "yulanadhost11.yulan.pie.lab.emc.com",
    # "ActiveDirectory.1.GlobalCatalog3" = "yulanadhost11.yulan.pie.lab.emc.com", 

    # GCRootDomain can be configured when GCLookupEnable is Enabled  
    #"ActiveDirectory.1.GCRootDomain" = "test"     
  }



  #    ldap = {
  #		directory = {
  #			remote_role_mapping = [
  #				{
  #					local_role = "Administrator",
  #					remote_group = "cn = idracgroup,cn = users,dc = yulan,dc = pie,dc = lab,dc = emc,dc = com"
  #				}        
  #			],
  #   To Update LDAP service addresses for 17G please provide configuration in ldap_attributes
  #			service_addresses = [
  #				"yulanadhost12.yulan.pie.lab.emc.com"
  #			],
  #			service_enabled = false
  #		},
  #		ldap_service = {
  #			search_settings = {
  #			    base_distinguished_names = [
  #				  "dc = yulan,dc = pie,dc = lab,dc = emc,dc = com"
  #				],
  #				group_name_attribute = "name",
  #				user_name_attribute = "member"
  #			}
  #		}
  #	}
  #		
  #	ldap_attributes = {
  #	  "LDAP.1.GroupAttributeIsDN" = "Enabled"
  #	  "LDAP.1.Port" = "636",
  #	  "LDAP.1.BindDN" = "cn = adtester,cn = users,dc = yulan,dc = pie,dc = lab,dc = emc,dc = com",
  #	  "LDAP.1.BindPassword" = "",
  #	  "LDAP.1.SearchFilter" = "(objectclass = *)",
  #   To Update LDAP service addresses for 17G please provide below configuration
  #   "LDAP.1.Server": "yulanadhost12.yulan.pie.lab.emc.com",
  #	  
  #		  }

}