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

package models

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// DirectoryServiceAuthProviderCertificateResource to construct terraform schema for the auth provider certificate resource.
type DirectoryServiceAuthProviderCertificateResource struct {
	ID            types.String    `tfsdk:"id"`
	RedfishServer []RedfishServer `tfsdk:"redfish_server"`

	CertificateType   types.String `tfsdk:"certificate_type"`
	CertificateString types.String `tfsdk:"certificate_string"`
}

// DirectoryServiceCertificateResource is the tfsdk model of DirectoryServiceCertificate
type DirectoryServiceCertificateResource struct {
	Name                  types.String   `tfsdk:"name"`
	Description           types.String   `tfsdk:"description"`
	ValidNotAfter         types.String   `tfsdk:"valid_not_after"`
	Subject               Subject        `tfsdk:"subject"`
	Issuer                Subject        `tfsdk:"issuer"`
	ValidNotBefore        types.String   `tfsdk:"valid_not_before"`
	SerialNumber          types.String   `tfsdk:"serial_number"`
	CertificateUsageTypes []types.String `tfsdk:"certificate_usage_types"`
}

// DirectoryServiceAuthProviderCertificateDatasource to construct terraform schema for the auth provider certificate resource.
type DirectoryServiceAuthProviderCertificateDatasource struct {
	ID                                      types.String                             `tfsdk:"id"`
	RedfishServer                           []RedfishServer                          `tfsdk:"redfish_server"`
	DirectoryServiceAuthProviderCertificate *DirectoryServiceAuthProviderCertificate `tfsdk:"directory_service_auth_provider_certificate"`
	CertificateFilter                       *CertificateFilter                       `tfsdk:"certificate_filter"`
}

// DirectoryServiceAuthProviderCertificate is the tfsdk model of DirectoryServiceAuthProviderCertificate
type DirectoryServiceAuthProviderCertificate struct {
	DirectoryServiceCertificate *DirectoryServiceCertificate `tfsdk:"directory_service_certificate"`
}

// DirectoryServiceCertificate is the tfsdk model of DirectoryServiceCertificate
type DirectoryServiceCertificate struct {
	ODataId               types.String   `tfsdk:"odata_id"`
	Name                  types.String   `tfsdk:"name"`
	Description           types.String   `tfsdk:"description"`
	ValidNotAfter         types.String   `tfsdk:"valid_not_after"`
	Subject               Subject        `tfsdk:"subject"`
	Issuer                Subject        `tfsdk:"issuer"`
	ValidNotBefore        types.String   `tfsdk:"valid_not_before"`
	SerialNumber          types.String   `tfsdk:"serial_number"`
	CertificateUsageTypes []types.String `tfsdk:"certificate_usage_types"`
}

// Subject is the tfsdk model of Subject
type Subject struct {
	CommonName         types.String `tfsdk:"common_name"`
	Organization       types.String `tfsdk:"organization"`
	City               types.String `tfsdk:"city"`
	Country            types.String `tfsdk:"country"`
	Email              types.String `tfsdk:"email"`
	OrganizationalUnit types.String `tfsdk:"organizational_unit"`
	State              types.String `tfsdk:"state"`
}

// CertificateFilter is the tfsdk model of CertificateFilter
type CertificateFilter struct {
	CertificateProviderType types.String `tfsdk:"certificate_provider_type"`
	CertificateId           types.String `tfsdk:"certificate_id"`
}

// CertificateCollection is the json model of CertificateCollection
type CertificateCollection struct {
	OdataContext string               `json:"@odata.context"`
	OdataID      string               `json:"@odata.id"`
	OdataType    string               `json:"@odata.type"`
	Name         string               `json:"Name"`
	Description  string               `json:"Description"`
	Members      []CertificateMembers `json:"Members"`
	MembersCount int                  `json:"Members@odata.count"`
}

// CertificateMembers is the json model of CertificateMembers
type CertificateMembers struct {
	OdataID string `json:"@odata.id"`
}

// Certificate is the json model of Certificate
type Certificate struct {
	OdataContext          string             `json:"@odata.context"`
	ODataID               string             `json:"@odata.id"`
	Name                  string             `json:"Name"`
	Description           string             `json:"Description"`
	ValidNotAfter         string             `json:"ValidNotAfter"`
	Subject               CertificateSubject `json:"Subject"`
	Issuer                CertificateSubject `json:"Issuer"`
	ValidNotBefore        string             `json:"ValidNotBefore"`
	SerialNumber          string             `json:"SerialNumber"`
	CertificateUsageTypes []string           `json:"CertificateUsageTypes"`
}

// CertificateSubject is the json model of CertificateSubject
type CertificateSubject struct {
	CommonName         string `json:"CommonName"`
	Organization       string `json:"Organization"`
	City               string `json:"City"`
	Country            string `json:"Country"`
	Email              string `json:"Email"`
	OrganizationalUnit string `json:"OrganizationalUnit"`
	State              string `json:"State"`
}
