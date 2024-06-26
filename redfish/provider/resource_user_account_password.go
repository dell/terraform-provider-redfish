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
	"fmt"
	"terraform-provider-redfish/redfish/models"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stmcginnis/gofish/redfish"

	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource = &UserAccountPasswordResource{}
)

// NewUserAccountPasswordResource is a helper function to simplify the provider implementation.
func NewUserAccountPasswordResource() resource.Resource {
	return &UserAccountPasswordResource{}
}

// UserAccountPasswordResource is the resource implementation.
type UserAccountPasswordResource struct {
	p *redfishProvider
}

// Configure implements resource.ResourceWithConfigure
func (r *UserAccountPasswordResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	r.p = req.ProviderData.(*redfishProvider)
}

// Metadata returns the resource type name.
func (*UserAccountPasswordResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "user_account_password"
}

// Schema defines the schema for the resource.
func (*UserAccountPasswordResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "This Terraform resource is used to update password for a user of the iDRAC Server.",
		Description:         "This Terraform resource is used to update password for a user of the iDRAC Server.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The ID of the resource.",
				Description:         "The ID of the resource.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"endpoint": schema.StringAttribute{
				MarkdownDescription: "The endpoint of the iDRAC.",
				Description:         "The endpoint of the iDRAC.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplaceIfConfigured(),
				},
			},
			"username": schema.StringAttribute{
				MarkdownDescription: "The name of the user",
				Description:         "The name of the user",
				Optional:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplaceIfConfigured(),
				},
			},
			"old_password": schema.StringAttribute{
				MarkdownDescription: "Old/current password of the user to be updated",
				Description:         "Old/current password of the user to be updated",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplaceIfConfigured(),
				},
			},
			"new_password": schema.StringAttribute{
				MarkdownDescription: "New Password of the user for login",
				Description:         "New Password of the user for login",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplaceIfConfigured(),
				},
			},
			"ssl_insecure": schema.BoolAttribute{
				MarkdownDescription: "This field indicates whether the SSL/TLS certificate must be verified or not",
				Description:         "This field indicates whether the SSL/TLS certificate must be verified or not",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplaceIfConfigured(),
				},
			},
		},
	}
}

// Create creates the resource and sets the initial Terraform state.
func (r *UserAccountPasswordResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	tflog.Trace(ctx, "resource_user_account_password create : Started")
	var state, plan models.UserAccountPassword

	// Get Plan Data
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// create a redfsh server object from plan
	redfishServer := []models.RedfishServer{
		{
			Endpoint:    plan.Endpoint,
			User:        plan.Username,
			Password:    plan.OldPassword,
			SslInsecure: plan.SslInsecure,
		},
	}

	// Lock the mutex to avoid race conditions with other resources
	redfishMutexKV.Lock(redfishServer[0].Endpoint.ValueString())
	defer redfishMutexKV.Unlock(redfishServer[0].Endpoint.ValueString())

	api, err := NewConfig(r.p, &redfishServer)
	if err != nil {
		resp.Diagnostics.AddError(ServiceErrorMsg, err.Error())
		return
	}
	service := api.Service
	defer api.Logout()

	// Fetch user account for which password needs to be updated
	accountList, err := GetAccountList(service)
	if err != nil {
		resp.Diagnostics.AddError("unable to access user data, please check access credentials", err.Error())
		return
	}
	userAccount, err := fetchAccountFromUserName(accountList, plan.Username.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("failed to fetch user account", err.Error())
		return
	}

	// run patch request with new password
	payload := make(map[string]interface{})
	payload["UserName"] = plan.Username.ValueString()
	payload["Password"] = plan.NewPassword.ValueString()

	_, err = service.GetClient().Patch(userAccount.ODataID, payload)
	if err != nil {
		resp.Diagnostics.AddError("password update failed", err.Error())
		return
	}

	// update password to new password and check if login is successful
	redfishServer[0].Password = types.StringValue(plan.NewPassword.ValueString())

	api, err = NewConfig(r.p, &redfishServer)
	if err != nil {
		resp.Diagnostics.AddError("login failed using new password", err.Error())
		return
	}
	service = api.Service
	defer api.Logout()

	systems, err := service.Systems()
	if len(systems) == 0 || err != nil {
		resp.Diagnostics.AddError("login failed using new password", err.Error())
		return
	}

	tflog.Trace(ctx, "resource_user_account_Password create: updating state finished, saving ...")
	state = plan
	state.ID = types.StringValue(userAccount.ODataID)
	// Save into State
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	tflog.Trace(ctx, "resource_user_account_Password create: finish")
}

// Read refreshes the Terraform state with the latest data.
func (*UserAccountPasswordResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	tflog.Trace(ctx, "resource_user_accountPassword read: started")
	// Get Plan Data
	var state models.UserAccountPassword
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	tflog.Trace(ctx, "resource_user_account_Password Read: finish")
}

// Update updates the resource and sets the updated Terraform state on success.
func (*UserAccountPasswordResource) Update(_ context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError(
		"Error updating User Password Resource.",
		"This resource is supposed to be replaced on update.",
	)
}

// Delete deletes the resource and removes the Terraform state on success.
func (*UserAccountPasswordResource) Delete(ctx context.Context, _ resource.DeleteRequest, resp *resource.DeleteResponse) {
	resp.State.RemoveResource(ctx)
}

func fetchAccountFromUserName(accountList []*redfish.ManagerAccount, username string) (*redfish.ManagerAccount, error) {
	for _, account := range accountList {
		if username == account.UserName {
			return account, nil
		}
	}
	return nil, fmt.Errorf("account not found")
}
