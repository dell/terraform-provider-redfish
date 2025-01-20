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
	"terraform-provider-redfish/redfish/helper"
	"terraform-provider-redfish/redfish/models"

	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
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
	state, diags := helper.ReadDatasourceRedfishDSAuthProviderCertificate(g.service, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
