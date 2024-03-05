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
	"terraform-provider-redfish/redfish/models"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/stmcginnis/gofish/redfish"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource = &managerResetResource{}
)

const (
	defaultCheckInterval int = 5
	defaultCheckTimeout  int = 300
)

// NewManagerResetResource is a helper function to simplify the provider implementation.
func NewManagerResetResource() resource.Resource {
	return &managerResetResource{}
}

// managerResetResource is the resource implementation.
type managerResetResource struct {
	p *redfishProvider
}

// Configure implements resource.ResourceWithConfigure
func (r *managerResetResource) Configure(ctx context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	r.p = req.ProviderData.(*redfishProvider)
	tflog.Trace(ctx, "resource_manager_reset configured")
}

// Metadata returns the resource type name.
func (*managerResetResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "manager_reset"
}

// ManagerResetSchema to design the schema for manager reset resource.
func ManagerResetSchema() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"id": schema.StringAttribute{
			MarkdownDescription: "The value of the Id property of the Manager resource",
			Description:         "The value of the Id property of the Manager resource",
			Required:            true,
			Validators: []validator.String{
				stringvalidator.LengthAtLeast(1),
			},
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
		"reset_type": schema.StringAttribute{
			MarkdownDescription: "The type of the reset operation to be performed. Accepted value: GracefulRestart",
			Description:         "The type of the reset operation to be performed. Accepted value: GracefulRestart",
			Required:            true,
			Validators: []validator.String{
				stringvalidator.OneOf(
					string(redfish.GracefulRestartResetType),
				),
			},
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
	}
}

// Schema defines the schema for the resource.
func (*managerResetResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "This resource is used to reset the manager.",
		Description:         "This resource is used to reset the manager.",
		Attributes:          ManagerResetSchema(),
		Blocks:              RedfishServerResourceBlockMap(),
	}
}

// Create creates the resource and sets the initial Terraform state.
func (r *managerResetResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	tflog.Trace(ctx, "resource_manager_reset create : Started")
	// Get Plan Data
	var plan models.RedfishManagerReset
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Lock the mutex to avoid race conditions with other resources
	redfishMutexKV.Lock(plan.RedfishServer[0].Endpoint.ValueString())
	defer redfishMutexKV.Unlock(plan.RedfishServer[0].Endpoint.ValueString())

	resetType := plan.ResetType.ValueString()
	managerID := plan.Id.ValueString()

	// Get manager
	manager, err := getManager(r, plan, managerID)
	if err != nil {
		resp.Diagnostics.AddError("Error while retrieving manager from redfish API", err.Error())
		return
	}

	// Perform manager reset
	err = manager.Reset(redfish.ResetType(resetType))
	if err != nil {
		resp.Diagnostics.AddError("Error resetting manager", err.Error())
		return
	}

	service, err := NewConfig(r.p, &plan.RedfishServer)
	if err != nil {
		resp.Diagnostics.AddError("Error while getting service", err.Error())
		return
	}

	// Check iDRAC status
	checker := ServerStatusChecker{
		Service:  service,
		Endpoint: (plan.RedfishServer)[0].Endpoint.ValueString(),
		Interval: defaultCheckInterval,
		Timeout:  defaultCheckTimeout,
	}
	err = checker.Check(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Error while rebooting iDRAC. Operation may take longer duration to complete", err.Error())
		return
	}

	tflog.Trace(ctx, "resource_manager_reset create: updating state finished, saving ...")
	// Save into State
	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	tflog.Trace(ctx, "resource_manager_reset create: finish")
}

// Read refreshes the Terraform state with the latest data.
func (r *managerResetResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	tflog.Trace(ctx, "resource_manager_reset read: started")
	var state models.RedfishManagerReset
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	managerID := state.Id.ValueString()

	// Get manager
	manager, err := getManager(r, state, managerID)
	if err != nil {
		resp.Diagnostics.AddError("Error while retrieving manager from redfish API", err.Error())
		return
	}

	state.Id = types.StringValue(manager.ID)

	// Save into State
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	tflog.Trace(ctx, "resource_manager_reset read: finished")
}

// Update updates the resource and sets the updated Terraform state on success.
func (*managerResetResource) Update(_ context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Update should never happen, it will destroy and create in case of update
	resp.Diagnostics.AddError(
		"Error updating Manager reset.",
		"An update plan of Manager Reset should never be invoked. This resource is supposed to be replaced on update.",
	)
}

// Delete deletes the resource and removes the Terraform state on success.
func (*managerResetResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	tflog.Trace(ctx, "resource_manager_reset delete: started")
	// Get State Data
	var state models.RedfishManagerReset
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.State.RemoveResource(ctx)
	tflog.Trace(ctx, "resource_manager_reset delete: finished")
}

func getManagerFromCollection(managers []*redfish.Manager, managerID string) (*redfish.Manager, error) {
	for _, manager := range managers {
		if manager.ID == managerID {
			return manager, nil
		}
	}
	return nil, fmt.Errorf("invalid Manager ID provided")
}

func getManager(r *managerResetResource, d models.RedfishManagerReset, managerID string) (*redfish.Manager, error) {
	service, err := NewConfig(r.p, &d.RedfishServer)
	if err != nil {
		return nil, err
	}

	managers, err := service.Managers()
	if err != nil {
		return nil, err
	}

	manager, err := getManagerFromCollection(managers, managerID)
	if err != nil {
		return nil, err
	}
	return manager, nil
}
