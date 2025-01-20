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
	"io"
	"terraform-provider-redfish/gofish/dell"
	"terraform-provider-redfish/redfish/models"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/stmcginnis/gofish"
)

const (
	createCertAPI     = "/redfish/v1/AccountService/ActiveDirectory/Certificates"
	replaceCertAPI    = "/redfish/v1/CertificateService/Actions/CertificateService.ReplaceCertificate"
	pem               = "PEM"
	certificateType   = "CertificateType"
	certificateString = "CertificateString"
)

var _ resource.Resource = &RedfishDirectoryServiceAuthProviderCertificateResource{}

// NewRedfishDirectoryServiceAuthProviderCertificateResource is new Resource for directory Service auth provider certificate
func NewRedfishDirectoryServiceAuthProviderCertificateResource() resource.Resource {
	return &RedfishDirectoryServiceAuthProviderCertificateResource{}
}

// RedfishDirectoryServiceAuthProviderCertificateResource to construct resource
type RedfishDirectoryServiceAuthProviderCertificateResource struct {
	p   *redfishProvider
	ctx context.Context
}

// Configure implements resource.ResourceWithConfigure
// nolint: revive
func (r *RedfishDirectoryServiceAuthProviderCertificateResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	r.p = req.ProviderData.(*redfishProvider)
}

// Metadata returns the resource type name.
// nolint: revive
func (*RedfishDirectoryServiceAuthProviderCertificateResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "directory_service_auth_provider_certificate"
}

// Schema defines the schema for the resource.
func (*RedfishDirectoryServiceAuthProviderCertificateResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "This Terraform resource is used to configure Directory Service Auth Provider certificate",
		Description:         "This Terraform resource is used to configure Directory Service Auth Provider certificate",

		Attributes: DirectoryServiceAuthProviderCertificateResourceSchema(),
		Blocks:     RedfishServerResourceBlockMap(),
	}
}

// DirectoryServiceAuthProviderCertificateResourceSchema is a function that returns the schema for Directory Service Auth Provider Certificate
func DirectoryServiceAuthProviderCertificateResourceSchema() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"id": schema.StringAttribute{
			MarkdownDescription: "ID of the Directory Service Auth Provider Certificate resource",
			Description:         "ID of the Directory Service Auth Provider Certificate resource",
			Computed:            true,
		},
		"certificate_type": schema.StringAttribute{
			MarkdownDescription: "certificate Type",
			Description:         "certificate Type",
			Required:            true,
			Validators: []validator.String{
				stringvalidator.LengthAtLeast(1),
				stringvalidator.OneOf([]string{
					string(pem),
				}...),
			},
		},
		"certificate_string": schema.StringAttribute{
			MarkdownDescription: "Encrypted Certificate",
			Description:         "Encrypted Certificate",
			Required:            true,
		},
	}
}

// Create creates the resource and sets the initial Terraform state.
// nolint: revive
func (r *RedfishDirectoryServiceAuthProviderCertificateResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	r.ctx = ctx
	tflog.Trace(ctx, "resource_directory_service_auth_provider_certificate create : Started")
	// Get Plan Data
	var plan models.DirectoryServiceAuthProviderCertificateResource
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	api, err := NewConfig(r.p, &plan.RedfishServer)
	if err != nil {
		resp.Diagnostics.AddError("service error", err.Error())
		return
	}
	service := api.Service
	defer api.Logout()
	certURI, count, diags := getCertificateDetailsURI(service)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if count == 0 && plan.CertificateType.ValueString() == pem {
		diags = createRedfishDirectoryServiceAuthCertificate(service, &plan)
	} else {
		diags = updateRedfishDirectoryServiceAuthCertificate(service, certURI, &plan)
	}
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	plan.ID = types.StringValue("redfish_directory_service_auth_provider_certificate_" + plan.CertificateType.ValueString())
	tflog.Trace(ctx, "resource_directory_service_auth_provider_certificate create: finished state update")

	// Save into State
	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	tflog.Trace(ctx, "resource_directory_service_auth_provider_certificate create: finish")
}

// Read refreshes the Terraform state with the latest data.
func (r *RedfishDirectoryServiceAuthProviderCertificateResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	tflog.Trace(ctx, "resource_directory_service_auth_provider_certificate read: started")
	r.ctx = ctx
	var state models.DirectoryServiceAuthProviderCertificateResource
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	state.ID = types.StringValue("redfish_directory_service_auth_provider_certificate_" + state.CertificateType.ValueString())
	tflog.Trace(ctx, "resource_directory_service_auth_provider_certificate read: finished reading state")
	// Save into State
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	tflog.Trace(ctx, "resource_directory_service_auth_provider_certificate read: finished")
}

// Update updates the resource and sets the updated Terraform state on success.
// nolint: revive
func (r *RedfishDirectoryServiceAuthProviderCertificateResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	r.ctx = ctx
	var state, plan models.DirectoryServiceAuthProviderCertificateResource
	// Get state Data
	tflog.Trace(ctx, "resource_directory_service_auth_provider_certificate update: started")
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get plan Data
	diags = req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	api, err := NewConfig(r.p, &plan.RedfishServer)
	if err != nil {
		resp.Diagnostics.AddError("service error", err.Error())
		return
	}
	service := api.Service
	defer api.Logout()
	certURI, _, diags := getCertificateDetailsURI(service)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	diags = updateRedfishDirectoryServiceAuthCertificate(service, certURI, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	plan.ID = types.StringValue("redfish_directory_service_auth_provider_certificate_" + plan.CertificateType.ValueString())
	tflog.Trace(ctx, "resource_directory_service_auth_provider_certificate update: finished state update")
	// Save into State
	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	tflog.Trace(ctx, "resource_directory_service_auth_provider_certificate update: finished")
}

// Delete deletes the resource and removes the Terraform state on success.
// nolint: revive
func (*RedfishDirectoryServiceAuthProviderCertificateResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	tflog.Trace(ctx, "resource_directory_service_auth_provider_certificate delete: started")
	// Get State Data
	var state models.DirectoryServiceAuthProviderCertificateResource
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.State.RemoveResource(ctx)
	tflog.Trace(ctx, "resource_directory_service_auth_provider_certificate delete: finished")
}

// nolint: gofumpt
func updateRedfishDirectoryServiceAuthCertificate(service *gofish.Service, certURI string,
	plan *models.DirectoryServiceAuthProviderCertificateResource) diag.Diagnostics {
	var diags diag.Diagnostics
	// Lock the mutex to avoid race conditions with other resources
	redfishMutexKV.Lock(plan.RedfishServer[0].Endpoint.ValueString())
	defer redfishMutexKV.Unlock(plan.RedfishServer[0].Endpoint.ValueString())

	if plan.CertificateType.ValueString() == pem {
		if diags = updateCertificate(certURI, service, plan); diags.HasError() {
			return diags
		}
	}
	return diags
}

// nolint: gofumpt
func createRedfishDirectoryServiceAuthCertificate(service *gofish.Service,
	plan *models.DirectoryServiceAuthProviderCertificateResource) diag.Diagnostics {
	var diags diag.Diagnostics
	// Lock the mutex to avoid race conditions with other resources
	redfishMutexKV.Lock(plan.RedfishServer[0].Endpoint.ValueString())
	defer redfishMutexKV.Unlock(plan.RedfishServer[0].Endpoint.ValueString())
	if diags = createCertificate(service, plan); diags.HasError() {
		return diags
	}
	return diags
}

func getCertificateDetailsURI(service *gofish.Service) (string, int, diag.Diagnostics) {
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
		diags.AddError("Unable to fetch Certificate URI", "Unable to fetch Certificate URI")
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
