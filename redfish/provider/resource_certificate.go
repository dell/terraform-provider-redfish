/*
Copyright (c) 2023-2024 Dell Inc., or its subsidiaries. All Rights Reserved.

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
	"fmt"
	"io"
	"net/http"
	"strings"
	"terraform-provider-redfish/redfish/models"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

const (
	createSSLCertAPI = "/Oem/Dell/DelliDRACCardService/Actions/DelliDRACCardService.ImportSSLCertificate"
	resetSSLCertAPI  = "/Oem/Dell/DelliDRACCardService/Actions/DelliDRACCardService.SSLResetCfg"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource = &certificateResource{}
)

// NewCertificateResource is a helper function to simplify the provider implementation.
func NewCertificateResource() resource.Resource {
	return &certificateResource{}
}

// certificateResource is the resource implementation.
type certificateResource struct {
	p *redfishProvider
}

// Configure implements resource.ResourceWithConfigure
func (r *certificateResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	r.p = req.ProviderData.(*redfishProvider)
}

// Metadata returns the resource type name.
func (*certificateResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "certificate"
}

// Schema defines the schema for the resource.
func (*certificateResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Resource for import the ssl certificate to iDRAC, on the basis of input parameter Type." +
			" After importing the certificate, the iDRAC will automatically restart.",
		Description: "Resource for import the ssl certificate to iDRAC, on the basis of input parameter Type." +
			" After importing the certificate, the iDRAC will automatically restart.",
		Version:    1,
		Attributes: RedfishSSLCertificateSchema(),
		Blocks:     RedfishServerResourceBlockMap(),
	}
}

// RedfishSSLCertificateSchema is a function that returns the schema for RedfishSSLCertificate
func RedfishSSLCertificateSchema() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"id": schema.StringAttribute{
			MarkdownDescription: "ID",
			Description:         "ID",
			Computed:            true,
		},
		"certificate_type": schema.StringAttribute{
			MarkdownDescription: "Type of the certificate to be imported.",
			Description:         "Type of the certificate to be imported.",
			Required:            true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
			Validators: []validator.String{
				stringvalidator.OneOf(
					"CustomCertificate",
					"Server",
				),
			},
		},
		"passphrase": schema.StringAttribute{
			MarkdownDescription: "A passphrase for certificate file. Note: This is optional parameter for CSC certificate," +
				" and not required for Server and CA certificates.",
			Description: "A passphrase for certificate file. Note: This is optional parameter for CSC certificate," +
				" and not required for Server and CA certificates.",
			Optional: true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
		"ssl_certificate_content": schema.StringAttribute{
			MarkdownDescription: `SSLCertificate File require content of certificate 
				supported certificate type: 
				"CustomCertificate" - The certificate must be converted pkcs#12 format to encoded in Base64` +
				` and entire Base64 Content is required. The passphrase that was used to convert the certificate` +
				` to pkcs#12 format must also be provided in "passphrase" attribute. ` +
				`"Server" - Certificate Content is required.` +
				` Note - The certificate should be signed with hashing algorithm equivalent to sha256.`,
			Description: `SSLCertificate File require content of certificate 
				supported certificate type: 
				"CustomCertificate" - The certificate must be converted pkcs#12 format to encoded in Base64` +
				` and entire Base64 Content is required. The passphrase that was used to convert the certificate` +
				` to pkcs#12 format must also be provided in "passphrase" attribute. ` +
				`"Server" - Certificate Content is required.` +
				` Note - The certificate should be signed with hashing algorithm equivalent to sha256.`,
			Required: true,
			Validators: []validator.String{
				stringvalidator.LengthAtLeast(1),
			},
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
	}
}

// Create creates the resource and sets the initial Terraform state.
func (r *certificateResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	tflog.Trace(ctx, "resource_certificate create : Started")

	// Get Plan Data
	var plan models.RedfishSSLCertificate
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	redfishMutexKV.Lock(plan.RedfishServer[0].Endpoint.ValueString())
	defer redfishMutexKV.Unlock(plan.RedfishServer[0].Endpoint.ValueString())

	payload := models.SSLCertificate{
		CertificateType:    plan.CertificateType.ValueString(),
		Passphrase:         plan.Passphrase.ValueString(),
		SSLCertificateFile: plan.SSLCertificateFile.ValueString(),
	}

	params := CertUtilsParam{
		ctx:     ctx,
		pconfig: r.p,
		rserver: &plan.RedfishServer,
		api:     createSSLCertAPI,
		payload: payload,
	}

	ok, summary, details := certutils(params)
	if !ok {
		resp.Diagnostics.AddError(summary, details)
		return
	}

	tflog.Debug(ctx, "resource_certificate create: updating state finished, saving ...")
	// Save into State
	plan.ID = types.StringValue("placeholder")
	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	tflog.Trace(ctx, "resource_certificate create: finish")
}

// Read refreshes the Terraform state with the latest data.
func (*certificateResource) Read(_ context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// read refresh changes nothing
	resp.State = req.State
}

// Update updates the resource and sets the updated Terraform state on success.
func (*certificateResource) Update(_ context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Update should never happen
	resp.Diagnostics.AddError(
		"Error updating Certificate.",
		"An update plan of Certificate should never be invoked. This resource is supposed to be replaced on update.",
	)
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *certificateResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	tflog.Trace(ctx, "resource_certificate delete: started")
	// Get State Data
	var state models.RedfishSSLCertificate
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	redfishMutexKV.Lock(state.RedfishServer[0].Endpoint.ValueString())
	defer redfishMutexKV.Unlock(state.RedfishServer[0].Endpoint.ValueString())

	payload := strings.NewReader(`{}`)

	params := CertUtilsParam{
		ctx:     ctx,
		pconfig: r.p,
		rserver: &state.RedfishServer,
		api:     resetSSLCertAPI,
		payload: payload,
	}

	ok, summary, details := certutils(params)
	if !ok {
		resp.Diagnostics.AddError(summary, details)
		return
	}

	resp.State.RemoveResource(ctx)
	tflog.Trace(ctx, "resource_certificate delete: finished")
}

// CertUtilsParam to get parameters for certutils function
type CertUtilsParam struct {
	ctx     context.Context
	pconfig *redfishProvider
	rserver *[]models.RedfishServer
	api     string
	payload interface{}
}

func certutils(params CertUtilsParam) (ok bool, summary string, details string) {
	// Get service
	api, err := NewConfig(params.pconfig, params.rserver)
	if err != nil {
		return false, ServiceErrorMsg, err.Error()
	}
	service := api.Service
	defer api.Logout()
	managers, err := service.Managers()
	if err != nil {
		return false, "Couldn't retrieve managers from redfish API: ", err.Error()
	}
	uri := managers[0].ODataID + params.api
	res, err1 := service.GetClient().Post(uri, params.payload)
	if err1 != nil {
		return false, "Couldn't upload certificate from redfish API: ", err1.Error()
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		body, err := io.ReadAll(res.Body)
		if err != nil {
			return false, "Couldn't upload certificate from redfish API: ", err.Error()
		}
		return false, "Couldn't upload certificate from redfish API: ", string(body)
	}

	// Check iDRAC status
	checker := ServerStatusChecker{
		Service:  service,
		Endpoint: (*params.rserver)[0].Endpoint.ValueString(),
		Interval: defaultCheckInterval,
		Timeout:  defaultCheckTimeout,
	}
	err = checker.Check(params.ctx, "")
	if err != nil {
		return false, "Error while rebooting iDRAC. Operation may take longer duration to complete", err.Error()
	}

	return true, fmt.Sprintf("%v api", params.api), "successful execution"
}
