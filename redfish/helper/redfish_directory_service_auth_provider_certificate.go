/*
Copyright (c) 2025 Dell Inc., or its subsidiaries. All Rights Reserved.

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

package helper

import (
	"encoding/json"
	"fmt"
	"io"
	"terraform-provider-redfish/gofish/dell"
	"terraform-provider-redfish/redfish/models"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stmcginnis/gofish"
)

// Constants for ActiveDirectory and LDAP
const (
	ActiveDirectory = "ActiveDirectory"
	LDAP            = "LDAP"
	BLANK           = ""
	// StatusCodeSuccess will denote http.response success code
	StatusCodeSuccess int = 200

	createCertAPI     = "/redfish/v1/AccountService/ActiveDirectory/Certificates"
	replaceCertAPI    = "/redfish/v1/CertificateService/Actions/CertificateService.ReplaceCertificate"
	pem               = "PEM"
	certificateType   = "CertificateType"
	certificateString = "CertificateString"
	certURIErr        = "Unable to fetch Certificate URI"
)

// nolint: revive
// ReadDatasourceRedfishDSAuthProviderCertificate is a helper function to read Certificate resource for DSAP
func ReadDatasourceRedfishDSAuthProviderCertificate(service *gofish.Service, d models.DirectoryServiceAuthProviderCertificateDatasource) (
	models.DirectoryServiceAuthProviderCertificateDatasource, diag.Diagnostics,
) {
	var diags diag.Diagnostics

	accountService, err := service.AccountService()
	if err != nil {
		diags.AddError("Error fetching Account Service", err.Error())
		return d, diags
	}

	// write the current time as ID
	d.ID = types.StringValue(fmt.Sprintf("%d", time.Now().Unix()))

	dellCertificate, certErr := dell.DirectoryServiceAuthProvider(accountService)

	if certErr != nil {
		diags.AddError(certURIErr, certURIErr)
	}

	var certificateURI string

	if d.CertificateFilter.CertificateProviderType.IsNull() || d.CertificateFilter.CertificateProviderType.IsUnknown() {
		diags.AddError("Invalid CertificateProviderType", "Please provide valid value for CertificateProviderType")
		return d, diags
	}

	if d.CertificateFilter.CertificateProviderType.ValueString() != ActiveDirectory && d.CertificateFilter.CertificateProviderType.ValueString() != LDAP {
		diags.AddError("Invalid CertificateProviderType", "Please provide valid value for CertificateProviderType")
		return d, diags
	}

	if d.CertificateFilter.CertificateProviderType.ValueString() == ActiveDirectory {
		certificateURI = dellCertificate.ActiveDirectoryCertificate.ODataID
	}
	if d.CertificateFilter.CertificateProviderType.ValueString() == LDAP {
		certificateURI = dellCertificate.LDAPCertificate.ODataID
	}
	var certificateDetailsURI string
	if d.CertificateFilter.CertificateId.IsNull() || d.CertificateFilter.CertificateId.IsUnknown() {
		response, err := service.GetClient().Get(certificateURI)
		if err != nil {
			diags.AddError("Error fetching Certificate collections", err.Error())
			return d, diags
		}

		if response.StatusCode != StatusCodeSuccess {
			return d, diags
		}
		body, err := io.ReadAll(response.Body)
		var certificateCollections models.CertificateCollection
		if err != nil {
			return d, diags
		}

		err = json.Unmarshal(body, &certificateCollections)
		if err != nil {
			diags.AddError("Error parsing Certificate Collection", err.Error())
			return d, diags
		}

		if certificateCollections.MembersCount != 0 {
			certificateDetailsURI = certificateCollections.Members[len(certificateCollections.Members)-1].OdataID
		} else {
			diags.AddError("Certificate Details are not Available", "Certificate Details are not Available")
			return d, diags
		}

	}

	if !d.CertificateFilter.CertificateId.IsNull() && !d.CertificateFilter.CertificateId.IsUnknown() && d.CertificateFilter.CertificateId.ValueString() == "" {
		diags.AddError("CertificateId can't be empty value", "CertificateId can't be empty value")
		return d, diags
	}

	if !d.CertificateFilter.CertificateId.IsNull() && !d.CertificateFilter.CertificateId.IsUnknown() {
		certificateDetailsURI = certificateURI + "/" + d.CertificateFilter.CertificateId.ValueString()
	}

	certResponse, err := service.GetClient().Get(certificateDetailsURI)
	// nolint: gofumpt
	if err != nil {
		diags.AddError("Error fetching Certificate", err.Error())
		return d, diags
	}

	if certResponse.StatusCode != StatusCodeSuccess {
		return d, diags
	}
	certBody, err := io.ReadAll(certResponse.Body)
	var certificate models.Certificate
	if err != nil {
		return d, diags
	}

	err = json.Unmarshal(certBody, &certificate)
	if err != nil {
		diags.AddError("Error parsing Certificate", err.Error())
		return d, diags
	}
	directoryServiceCertificate := newDSAuthProviderCertificateState(certificate)
	var directoryServiceAuthProviderCertificate models.DirectoryServiceAuthProviderCertificate
	directoryServiceAuthProviderCertificate.DirectoryServiceCertificate = directoryServiceCertificate
	d.DirectoryServiceAuthProviderCertificate = &directoryServiceAuthProviderCertificate
	if d.DirectoryServiceAuthProviderCertificate == nil {
		diags.AddError("DirectoryServiceAuthProviderCertificate null ", "DirectoryServiceAuthProviderCertificate null")
		return d, diags
	}
	return d, diags
}

func newDSAuthProviderCertificateState(certificateData models.Certificate) *models.DirectoryServiceCertificate {
	return &models.DirectoryServiceCertificate{
		ODataId:               types.StringValue(certificateData.ODataID),
		Name:                  types.StringValue(certificateData.Name),
		Description:           types.StringValue(certificateData.Description),
		ValidNotAfter:         types.StringValue(certificateData.ValidNotAfter),
		Subject:               newSubjectAndIssuerState(&certificateData.Subject),
		Issuer:                newSubjectAndIssuerState(&certificateData.Issuer),
		ValidNotBefore:        types.StringValue(certificateData.ValidNotBefore),
		SerialNumber:          types.StringValue(certificateData.SerialNumber),
		CertificateUsageTypes: newCertificateUsageTypeState(certificateData.CertificateUsageTypes),
	}
}

func newCertificateUsageTypeState(input []string) []types.String {
	out := make([]types.String, 0)
	for _, input := range input {
		out = append(out, types.StringValue(string(input)))
	}
	return out
}

func newSubjectAndIssuerState(input *models.CertificateSubject) models.Subject {
	return models.Subject{
		CommonName:         types.StringValue(input.CommonName),
		Organization:       types.StringValue(input.Organization),
		City:               types.StringValue(input.City),
		Country:            types.StringValue(input.Country),
		Email:              types.StringValue(input.Email),
		OrganizationalUnit: types.StringValue(input.OrganizationalUnit),
		State:              types.StringValue(input.State),
	}
}

// UpdateRedfishDirectoryServiceAuthCertificate is a helper function to update Certificate resource for DSAP
// nolint: gofumpt
func UpdateRedfishDirectoryServiceAuthCertificate(service *gofish.Service, certURI string,
	plan *models.DirectoryServiceAuthProviderCertificateResource) diag.Diagnostics {
	var diags diag.Diagnostics
	if plan.CertificateType.ValueString() == pem {
		if diags = updateCertificate(certURI, service, plan); diags.HasError() {
			return diags
		}
	}
	return diags
}

// CreateRedfishDirectoryServiceAuthCertificate is a helper function to create Certificate resource for DSAP
// nolint: gofumpt
func CreateRedfishDirectoryServiceAuthCertificate(service *gofish.Service,
	plan *models.DirectoryServiceAuthProviderCertificateResource) diag.Diagnostics {
	var diags diag.Diagnostics
	if diags = createCertificate(service, plan); diags.HasError() {
		return diags
	}
	return diags
}

// GetCertificateDetailsURI is a helper function to get Certificate URI
func GetCertificateDetailsURI(service *gofish.Service) (string, int, diag.Diagnostics) {
	var certificateDetailsURI string
	var diags diag.Diagnostics
	var certificateCollections models.CertificateCollection
	// get the account service resource and ODATA_ID will be used to make a patch call
	accountService, err := service.AccountService()
	if err != nil {
		diags.AddError("error fetching accountservice resource", err.Error())
		return "", 0, diags
	}

	dellCertificate, certErr := dell.DirectoryServiceAuthProvider(accountService)
	if certErr != nil {
		diags.AddError(certURIErr, certURIErr)
		return "", 0, diags
	}
	certificateURI := dellCertificate.ActiveDirectoryCertificate.ODataID
	response, err := service.GetClient().Get(certificateURI)
	if err != nil {
		diags.AddError("Error fetching Certificate collections", err.Error())
		return "", 0, diags
	}

	if response.StatusCode != StatusCodeSuccess {
		diags.AddError("Error", "Invalid")
		return "", 0, diags
	}
	body, err := io.ReadAll(response.Body)
	if err != nil {
		diags.AddError("Error while reading the response ", err.Error())
		return "", 0, diags
	}

	err = json.Unmarshal(body, &certificateCollections)
	if err != nil {
		diags.AddError("Error parsing Certificate Collection", err.Error())
		return "", 0, diags
	}

	if certificateCollections.MembersCount != 0 {
		certificateDetailsURI = certificateCollections.Members[len(certificateCollections.Members)-1].OdataID
		return certificateDetailsURI, certificateCollections.MembersCount, nil
	}
	return certificateDetailsURI, certificateCollections.MembersCount, nil
}

// nolint: revive
func updateCertificate(certURI string, service *gofish.Service, plan *models.DirectoryServiceAuthProviderCertificateResource) (diags diag.Diagnostics) {
	patchBody := make(map[string]interface{})
	patchBody[certificateType] = plan.CertificateType.ValueString()
	patchBody[certificateString] = plan.CertificateString.ValueString()
	patchBody["CertificateUri"] = map[string]interface{}{
		"@odata.id": certURI,
	}

	if diags = postCall(replaceCertAPI, patchBody, service); diags.HasError() {
		return diags
	}
	return diags
}

func createCertificate(service *gofish.Service, plan *models.DirectoryServiceAuthProviderCertificateResource) (diags diag.Diagnostics) {
	patchBody := make(map[string]interface{})
	patchBody[certificateType] = plan.CertificateType.ValueString()
	patchBody[certificateString] = plan.CertificateString.ValueString()
	if diags = postCall(createCertAPI, patchBody, service); diags.HasError() {
		return diags
	}
	return nil
}

func postCall(uri string, patchBody map[string]interface{}, service *gofish.Service) (diags diag.Diagnostics) {
	response, err := service.GetClient().Post(uri, patchBody)
	if err != nil {
		diags.AddError("There was an error while creating/updating Certificate resource",
			err.Error())
		return diags
	}
	if response != nil {
		body, err := io.ReadAll(response.Body)
		if err != nil {
			diags.AddError("Error reading response", "error "+string(body))
			return diags
		}
		readResponse := make(map[string]json.RawMessage)
		err = json.Unmarshal(body, &readResponse)
		if err != nil {
			diags.AddError("Error unmarshalling response", err.Error())
			return diags
		}
		// check for extended error message in response
		errorMsg, ok := readResponse["error"]
		if ok {
			diags.AddError("Error creating/updating Certificate resource ", string(errorMsg))
			return diags
		}
	}
	return diags
}
