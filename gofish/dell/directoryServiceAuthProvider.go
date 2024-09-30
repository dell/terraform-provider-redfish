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

package dell

import (
	"encoding/json"

	"github.com/stmcginnis/gofish/redfish"
)

// DirectoryServiceAuthProviderExtended contains gofish AccountService data, as well as Accounts,
// AdditionalExternalAccountProviders,Roles,PrivilegeMap,ActiveDirectoryCertificate,LDAPCertificate links
type DirectoryServiceAuthProviderExtended struct {
	*redfish.AccountService
	Accounts                           *redfish.ManagerAccount
	AdditionalExternalAccountProviders *redfish.ExternalAccountProvider
	Roles                              *redfish.Role
	PrivilegeMap                       *redfish.PrivilegeRegistry
	ActiveDirectoryCertificate         *redfish.Certificate
	LDAPCertificate                    *redfish.Certificate
}

// DirectoryServiceAuthProvider returns a Dell.DirectoryServiceAuthProvider pointer given a redfish.AccountService pointer from Gofish
// This is the wrapper that extracts and parses AccountService data, as well as Accounts,AdditionalExternalAccountProviders,Roles,
// PrivilegeMap,ActiveDirectoryCertificate,LDAPCertificate links.
func DirectoryServiceAuthProvider(accountService *redfish.AccountService) (*DirectoryServiceAuthProviderExtended, error) {
	dellAccount := &DirectoryServiceAuthProviderExtended{
		AccountService:                     accountService,
		Accounts:                           &redfish.ManagerAccount{},
		AdditionalExternalAccountProviders: &redfish.ExternalAccountProvider{},
		Roles:                              &redfish.Role{},
		PrivilegeMap:                       &redfish.PrivilegeRegistry{},
		ActiveDirectoryCertificate:         &redfish.Certificate{},
		LDAPCertificate:                    &redfish.Certificate{},
	}
	rawDataBytes, err := GetRawDataBytes(accountService)
	if err != nil {
		return dellAccount, err
	}

	if accountRawData, found := GetNodeFromRawDataBytes(rawDataBytes, "Accounts"); found == nil {
		var accountData *redfish.ManagerAccount
		if err = json.Unmarshal(accountRawData, &accountData); err == nil {
			dellAccount.Accounts = accountData
		}
	}

	if additionalRawData, found := GetNodeFromRawDataBytes(rawDataBytes, "AdditionalExternalAccountProviders"); found == nil {
		var additionalData *redfish.ExternalAccountProvider
		if err = json.Unmarshal(additionalRawData, &additionalData); err == nil {
			dellAccount.AdditionalExternalAccountProviders = additionalData
		}
	}
	if rolesRawData, found := GetNodeFromRawDataBytes(rawDataBytes, "Roles"); found == nil {
		var rolesData *redfish.Role
		if err = json.Unmarshal(rolesRawData, &rolesData); err == nil {
			dellAccount.Roles = rolesData
		}
	}
	if privilegeMapRawData, found := GetNodeFromRawDataBytes(rawDataBytes, "PrivilegeMap"); found == nil {
		var privilegeMapData *redfish.PrivilegeRegistry
		if err = json.Unmarshal(privilegeMapRawData, &privilegeMapData); err == nil {
			dellAccount.PrivilegeMap = privilegeMapData
		}
	}
	if activeDirectoryRawData, found := GetNodeFromRawDataBytes(rawDataBytes, "ActiveDirectory"); found == nil {
		if activeDirectoryCertificateRawData, found := GetNodeFromRawDataBytes(activeDirectoryRawData, "Certificates"); found == nil {
			var activeDirectoryData *redfish.Certificate
			if err = json.Unmarshal(activeDirectoryCertificateRawData, &activeDirectoryData); err == nil {
				dellAccount.ActiveDirectoryCertificate = activeDirectoryData
			}
		}
	}

	if ldapDirectoryRawData, found := GetNodeFromRawDataBytes(rawDataBytes, "LDAP"); found == nil {
		if ldapCertificateRawData, found := GetNodeFromRawDataBytes(ldapDirectoryRawData, "Certificates"); found == nil {
			var ldapData *redfish.Certificate
			if err = json.Unmarshal(ldapCertificateRawData, &ldapData); err == nil {
				dellAccount.LDAPCertificate = ldapData
			}
		}
	}
	return dellAccount, nil
}
