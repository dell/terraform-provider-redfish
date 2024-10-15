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
      #         local_role = "None",
      #         remote_group = "idracgroup"
      #     }
      # ],
      # service_addresses = [
      #     "yulanadhost11.yulan.pie.lab.emc.com"
      #  ],
      service_enabled = true,
      authentication = {
        kerberos_key_tab_file = data.local_file.kerberos.content
      }
    }
  }

  active_directory_attributes = {
    "ActiveDirectory.1.AuthTimeout"          = "120",
    "ActiveDirectory.1.CertValidationEnable" = "Enabled",
    "ActiveDirectory.1.DCLookupEnable"       = "Enabled",

    # RacName and RacDomain can be configured when Schema is Extended Schema
    "ActiveDirectory.1.RacDomain" = "test",
    "ActiveDirectory.1.RacName"   = "test",

    # if SSOEnable is Enabled make sure ActiveDirectory Service is enabled and valid kerberos_key_tab_file is provided
    "ActiveDirectory.1.SSOEnable" = "Disabled",

    # Schema can be Extended Schema or Standard Schema
    "ActiveDirectory.1.Schema" = "Extended Schema",
    "UserDomain.1.Name"        = "yulan.pie.lab.emc.com",

    # DCLookupByUserDomain must be configured when DCLookupEnable is enabled 
    "ActiveDirectory.1.DCLookupByUserDomain" : "Enabled",

    # DCLookupDomainName must be configured when DCLookupByUserDomain is Disabled and DCLookupEnable is Enabled
    #"ActiveDirectory.1.DCLookupDomainName"="test", 

    #"ActiveDirectory.1.GCLookupEnable" = "Disabled"

    # at least any one from GlobalCatalog1,GlobalCatalog2,GlobalCatalog3 must be configured when Schema is Standard and GCLookupEnable is Disabled
    # "ActiveDirectory.1.GlobalCatalog1" = "yulanadhost11.yulan.pie.lab.emc.com",
    # "ActiveDirectory.1.GlobalCatalog2" = "yulanadhost11.yulan.pie.lab.emc.com",
    # "ActiveDirectory.1.GlobalCatalog3" = "yulanadhost11.yulan.pie.lab.emc.com", 

    # GCRootDomain can be configured when GCLookupEnable is Enabled  
    #"ActiveDirectory.1.GCRootDomain" = "test"  

    # RSA Secure configuration required Datacenter license 
    #"LDAP.1.RSASecurID2FALDAP":"Enabled",
    #"RSASecurID2FA.1.RSASecurIDAccessKey": "&#9679;&#9679;1", 
    #"RSASecurID2FA.1.RSASecurIDClientID": "&#9679;&#9679;1", 
    #"RSASecurID2FA.1.RSASecurIDAuthenticationServer": "",    
  }



  #    ldap = {
  #		directory = {
  #			remote_role_mapping = [
  #				{
  #					local_role = "Administrator",
  #					remote_group = "cn = idracgroup,cn = users,dc = yulan,dc = pie,dc = lab,dc = emc,dc = com"
  #				}        
  #			],
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
  #	  
  #    #"LDAP.1.RSASecurID2FALDAP":"Enabled",
  #    #"RSASecurID2FA.1.RSASecurIDAccessKey": "&#9679;&#9679;1", 
  #    #"RSASecurID2FA.1.RSASecurIDClientID": "&#9679;&#9679;1", 
  #    #"RSASecurID2FA.1.RSASecurIDAuthenticationServer": "",
  #		  }

}