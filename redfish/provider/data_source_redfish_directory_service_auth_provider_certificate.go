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

package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"terraform-provider-redfish/gofish/dell"
	"terraform-provider-redfish/redfish/models"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stmcginnis/gofish"
)

const (
	// StatusCodeSuccess will denote http.response success code
	StatusCodeSuccess int = 200
)

var (
	_ datasource.DataSource              = &DirectoryServiceAuthProviderCertificateDatasource{}
	_ datasource.DataSourceWithConfigure = &DirectoryServiceAuthProviderCertificateDatasource{}
)

// NewDirectoryServiceAuthProviderCertificateDatasource is new datasource for directory Service auth provider
func NewDirectoryServiceAuthProviderCertificateDatasource() datasource.DataSource {
	return &DirectoryServiceAuthProviderCertificateDatasource{}
}

// DirectoryServiceAuthProviderCertificateDatasource to construct datasource
type DirectoryServiceAuthProviderCertificateDatasource struct {
	p       *redfishProvider
	ctx     context.Context
	service *gofish.Service
}

// Configure implements datasource.DataSourceWithConfigure
// nolint: revive
func (g *DirectoryServiceAuthProviderCertificateDatasource) Configure(_ context.Context, req datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	g.p = req.ProviderData.(*redfishProvider)
}

// Metadata implements datasource.DataSource
// nolint: revive
func (*DirectoryServiceAuthProviderCertificateDatasource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "directory_service_auth_provider_certificate"
}

// Schema implements datasource.DataSource
func (*DirectoryServiceAuthProviderCertificateDatasource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "This Terraform datasource is used to query existing Directory Service auth provider Certificate." +
			" The information fetched from this block can be further used for resource block.",
		Description: "This Terraform datasource is used to query existing Directory Service auth provider Certificate." +
			" The information fetched from this block can be further used for resource block.",
		Attributes: DirectoryServiceAuthProviderCertificateDatasourceSchema(),
		Blocks: map[string]schema.Block{
			"certificate_filter": schema.SingleNestedBlock{
				MarkdownDescription: "Certificate filter for Directory Service Auth Provider",
				Description:         "Certificate filter for Directory Service Auth Provider",
				Attributes:          CertificateFilterSchema(),
			},
			"redfish_server": schema.ListNestedBlock{
				MarkdownDescription: redfishServerMD,
				Description:         redfishServerMD,
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
					listvalidator.IsRequired(),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: RedfishServerDatasourceSchema(),
				},
			},
		},
	}
}

// DirectoryServiceAuthProviderCertificateDatasourceSchema to define the DirectoryServiceAuthProvider certificate data-source schema
func DirectoryServiceAuthProviderCertificateDatasourceSchema() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"id": schema.StringAttribute{
			MarkdownDescription: "ID of the Directory Service Auth Provider Certificate data-source",
			Description:         "ID of the Directory Service Auth Provider Certificate data-source",
			Computed:            true,
		},
		"directory_service_auth_provider_certificate": schema.SingleNestedAttribute{
			MarkdownDescription: "Directory Service Auth Provider Certificate Details.",
			Description:         "Directory Service Auth Provider Certificate Details.",
			Attributes:          DirectoryServiceAuthProviderCertificateSchema(),
			Computed:            true,
		},
	}
}

// DirectoryServiceAuthProviderCertificateSchema to define the DirectoryServiceAuthProvider Certificate schema
func DirectoryServiceAuthProviderCertificateSchema() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"directory_service_certificate": schema.SingleNestedAttribute{
			MarkdownDescription: "Directory Service Certificate Details.",
			Description:         "Directory Service Certificate Details.",
			Attributes:          DirectoryServiceCertificateSchema(),
			Computed:            true,
		},
		"security_certificate": schema.MapAttribute{
			MarkdownDescription: "SecurityCertificate attributes in Dell iDRAC attributes. ",
			Description:         "SecurityCertificate attributes in Dell iDRAC attributes.",
			ElementType:         types.StringType,
			Computed:            true,
		},
	}
}

// DirectoryServiceCertificateSchema is a function that returns the schema for Directory Service Auth Provider Certificate
func DirectoryServiceCertificateSchema() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"odata_id": schema.StringAttribute{
			MarkdownDescription: "OData ID for the Certificate",
			Description:         "OData ID for the Certificate",
			Computed:            true,
		},
		"name": schema.StringAttribute{
			MarkdownDescription: "Name of the Certificate",
			Description:         "Name of the Certificate",
			Computed:            true,
		},
		"description": schema.StringAttribute{
			MarkdownDescription: "Description of the Certificate",
			Description:         "Description Of the Certificate",
			Computed:            true,
		},
		"valid_not_after": schema.StringAttribute{
			MarkdownDescription: "The date when the certificate is no longer valid",
			Description:         "The date when the certificate is no longer valid",
			Computed:            true,
		},
		"subject": schema.SingleNestedAttribute{
			MarkdownDescription: "The subject of the certificate",
			Description:         "The subject of the certificate",
			Attributes:          SubjectSchema(),
			Computed:            true,
		},
		"issuer": schema.SingleNestedAttribute{
			MarkdownDescription: "The issuer of the certificate",
			Description:         "The issuer of the certificate",
			Attributes:          SubjectSchema(),
			Computed:            true,
		},
		"valid_not_before": schema.StringAttribute{
			MarkdownDescription: "The date when the certificate becomes valid",
			Description:         "The date when the certificate becomes valid",
			Computed:            true,
		},
		"serial_number": schema.StringAttribute{
			MarkdownDescription: "The serial number of the certificate",
			Description:         "The serial number of the certificate",
			Computed:            true,
		},
		"certificate_usage_types": schema.ListAttribute{
			ElementType:         types.StringType,
			MarkdownDescription: "The types or purposes for this certificate",
			Description:         "The types or purposes for this certificate",
			Computed:            true,
		},
	}
}

// SubjectSchema is a function that returns the schema for Subject or issuer
func SubjectSchema() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"common_name": schema.StringAttribute{
			MarkdownDescription: "The common name of the entity",
			Description:         "The common name of the entity",
			Computed:            true,
		},
		"organization": schema.StringAttribute{
			MarkdownDescription: "The name of the organization of the entity",
			Description:         "The name of the organization of the entity",
			Computed:            true,
		},
		"city": schema.StringAttribute{
			MarkdownDescription: "The city or locality of the organization of the entity",
			Description:         "The city or locality of the organization of the entity",
			Computed:            true,
		},
		"country": schema.StringAttribute{
			MarkdownDescription: "The country of the organization of the entity",
			Description:         "The country of the organization of the entity",
			Computed:            true,
		},
		"email": schema.StringAttribute{
			MarkdownDescription: "The email address of the contact within the organization of the entity",
			Description:         "The email address of the contact within the organization of the entity",
			Computed:            true,
		},
		"organizational_unit": schema.StringAttribute{
			MarkdownDescription: "The name of the unit or division of the organization of the entity",
			Description:         "The name of the unit or division of the organization of the entity",
			Computed:            true,
		},
		"state": schema.StringAttribute{
			MarkdownDescription: "The state, province, or region of the organization of the entity",
			Description:         "The state, province, or region of the organization of the entity",
			Computed:            true,
		},
	}
}

// CertificateFilterSchema to construct schema of certificate filter.
func CertificateFilterSchema() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"certificate_provider_type": schema.StringAttribute{
			Required:    true,
			Description: "Filter for CertificateProviderType",
		},
		"certificate_id": schema.StringAttribute{
			Optional:    true,
			Description: "CertificateId",
		},
	}
}

// Read implements datasource.DataSource
func (g *DirectoryServiceAuthProviderCertificateDatasource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var plan models.DirectoryServiceAuthProviderCertificateDatasource
	diags := req.Config.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	api, err := NewConfig(g.p, &plan.RedfishServer)
	if err != nil {
		resp.Diagnostics.AddError("service error", err.Error())
		return
	}
	defer api.Logout()
	g.ctx = ctx
	g.service = api.Service
	state, diags := g.readDatasourceRedfishDSAuthProviderCertificate(plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// nolint: revive
func (g *DirectoryServiceAuthProviderCertificateDatasource) readDatasourceRedfishDSAuthProviderCertificate(d models.DirectoryServiceAuthProviderCertificateDatasource) (
	models.DirectoryServiceAuthProviderCertificateDatasource, diag.Diagnostics,
) {
	var diags diag.Diagnostics

	accountService, err := g.service.AccountService()
	if err != nil {
		diags.AddError("Error fetching Account Service", err.Error())
		return d, diags
	}

	// write the current time as ID
	d.ID = types.StringValue(fmt.Sprintf("%d", time.Now().Unix()))

	dellCertificate, certErr := dell.DirectoryServiceAuthProvider(accountService)

	if certErr != nil {
		diags.AddError("Unable to fetch Certificate URI", "Unable to fetch Certificate URI")
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
		response, err := g.service.GetClient().Get(certificateURI)
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

	certResponse, err := g.service.GetClient().Get(certificateDetailsURI)
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

	if diags = newDSAuthProviderSecurityCertificate(g.service, d.DirectoryServiceAuthProviderCertificate); diags.HasError() {
		return d, diags
	}
	return d, diags
}

func newDSAuthProviderSecurityCertificate(service *gofish.Service, d *models.DirectoryServiceAuthProviderCertificate) diag.Diagnostics {
	var idracAttributesState models.DellIdracAttributes
	var diags diag.Diagnostics
	if diags := readDatasourceRedfishDellIdracAttributes(service, &idracAttributesState); diags.HasError() {
		return diags
	}

	attributesToReturn := make(map[string]attr.Value)
	middleindex, diags := getnumberForIdrac(&idracAttributesState)
	if middleindex != "" && !diags.HasError() {
		prefixValue := "SecurityCertificate." + middleindex
		for k, v := range idracAttributesState.Attributes.Elements() {
			if strings.HasPrefix(k, prefixValue) {
				attributesToReturn[k] = v
			}
		}
	}
	securityValue := types.MapValueMust(types.StringType, attributesToReturn)
	d.SecurityCertificate = securityValue
	return nil
}

func getnumberForIdrac(idracAttributesState *models.DellIdracAttributes) (string, diag.Diagnostics) {
	var diags diag.Diagnostics
	for k, v := range idracAttributesState.Attributes.Elements() {
		if strings.HasPrefix(k, "SecurityCertificate.") && strings.HasSuffix(k, ".CertificateType") {
			if v.String() == "\"RSA_CA\"" {
				arr := strings.Split(k, ".")
				return arr[1], nil
			}
		}
	}
	return "", diags
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
