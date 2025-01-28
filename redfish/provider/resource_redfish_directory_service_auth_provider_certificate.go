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

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
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
	certURI, count, diags := helper.GetCertificateDetailsURI(service)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	// Lock the mutex to avoid race conditions with other resources
	redfishMutexKV.Lock(plan.RedfishServer[0].Endpoint.ValueString())
	defer redfishMutexKV.Unlock(plan.RedfishServer[0].Endpoint.ValueString())
	if count == 0 && plan.CertificateType.ValueString() == pem {
		diags = helper.CreateRedfishDirectoryServiceAuthCertificate(service, &plan)
	} else {
		diags = helper.UpdateRedfishDirectoryServiceAuthCertificate(service, certURI, &plan)
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
	certURI, _, diags := helper.GetCertificateDetailsURI(service)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	redfishMutexKV.Lock(plan.RedfishServer[0].Endpoint.ValueString())
	defer redfishMutexKV.Unlock(plan.RedfishServer[0].Endpoint.ValueString())
	diags = helper.UpdateRedfishDirectoryServiceAuthCertificate(service, certURI, &plan)
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
