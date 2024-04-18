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
	"encoding/json"
	"fmt"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"terraform-provider-redfish/redfish/models"

	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	tfpath "github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/stmcginnis/gofish"
	"github.com/stmcginnis/gofish/redfish"
)

const (
	minUserNameLength = 1
	maxUserNameLength = 16
	minPasswordLength = 4
	maxPasswordLength = 40
	maxUserID         = 16
	minUserID         = 2
	password          = "password"
	username          = "username"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource = &UserAccountResource{}
)

// NewUserAccountResource is a helper function to simplify the provider implementation.
func NewUserAccountResource() resource.Resource {
	return &UserAccountResource{}
}

// UserAccountResource is the resource implementation.
type UserAccountResource struct {
	p   *redfishProvider
	ctx context.Context
}

// Configure implements resource.ResourceWithConfigure
func (r *UserAccountResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	r.p = req.ProviderData.(*redfishProvider)
}

// Metadata returns the resource type name.
func (*UserAccountResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "user_account"
}

// Schema defines the schema for the resource.
func (*UserAccountResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "This Terraform resource is used to manage user entity of the iDRAC Server. We can create, read, " +
			"modify and delete an existing user using this resource.",
		Description: "This Terraform resource is used to manage user entity of the iDRAC Server. We can create, read, " +
			"modify and delete an existing user using this resource.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The ID of the resource. Cannot be updated.",
				Description:         "The ID of the resource. Cannot be updated.",
				Computed:            true,
			},
			"user_id": schema.StringAttribute{
				MarkdownDescription: "The ID of the user. Cannot be updated.",
				Description:         "The ID of the user. Cannot be updated.",
				Optional:            true,
				Computed:            true,
			},
			username: schema.StringAttribute{
				MarkdownDescription: "The name of the user",
				Description:         "The name of the user",
				Optional:            true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(minUserNameLength, maxUserNameLength),
				},
				DeprecationMessage: "Single user support is deprecated and will be removed in an upcoming release. Use 'users' block instead.",
			},
			password: schema.StringAttribute{
				MarkdownDescription: "Password of the user",
				Description:         "Password of the user",
				Optional:            true,
				Sensitive:           true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(minPasswordLength, maxPasswordLength),
					stringvalidator.AlsoRequires(tfpath.MatchRoot(username)),
				},
			},
			"role_id": schema.StringAttribute{
				MarkdownDescription: "Role of the user. Applicable values are 'Operator', 'Administrator', 'None', and 'ReadOnly'. ",
				Description:         "Role of the user. Applicable values are 'Operator', 'Administrator', 'None', and 'ReadOnly'. ",
				Optional:            true,
				Computed:            true,
				Validators: []validator.String{
					stringvalidator.OneOf([]string{
						"Operator",
						"Administrator",
						"ReadOnly",
						"None",
					}...),
				},
			},
			"enabled": schema.BoolAttribute{
				MarkdownDescription: "If the user is currently active or not.",
				Description:         "If the user is currently active or not.",
				Optional:            true,
				Computed:            true,
			},
			"users": schema.ListNestedAttribute{
				Description:         "To create/delete/modify multiple users.",
				MarkdownDescription: "To create/delete/modify multiple users.",
				Optional:            true,
				Validators: []validator.List{
					listvalidator.ExactlyOneOf(tfpath.Expressions{
						tfpath.MatchRoot(username),
					}...),
					listvalidator.ConflictsWith(tfpath.Expressions{
						tfpath.MatchRoot(username),
						tfpath.MatchRoot(password),
						tfpath.MatchRoot("role_id"),
						tfpath.MatchRoot("enabled"),
					}...),
				},
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"user_id": schema.StringAttribute{
							MarkdownDescription: "The ID of the users. Cannot be updated.",
							Description:         "The ID of the users. Cannot be updated.",
							Optional:            true,
							Computed:            true,
						},
						username: schema.StringAttribute{
							MarkdownDescription: "The name of the users",
							Description:         "The name of the users",
							Required:            true,
							Validators: []validator.String{
								stringvalidator.LengthBetween(minUserNameLength, maxUserNameLength),
							},
						},
						password: schema.StringAttribute{
							MarkdownDescription: "Password of the users",
							Description:         "Password of the users",
							Required:            true,
							Sensitive:           true,
							Validators: []validator.String{
								stringvalidator.LengthBetween(minPasswordLength, maxPasswordLength),
							},
						},
						"role": schema.StringAttribute{
							MarkdownDescription: "Role of the users. Applicable values are 'Operator', 'Administrator', 'None', and 'ReadOnly'. " +
								"Default is \"None\"",
							Description: "Role of the users. Applicable values are 'Operator', 'Administrator', 'None', and 'ReadOnly'. " +
								"Default is \"None\"",
							Optional: true,
							Computed: true,
							Default:  stringdefault.StaticString("None"),
							Validators: []validator.String{
								stringvalidator.OneOf([]string{
									"Operator",
									"Administrator",
									"ReadOnly",
									"None",
								}...),
							},
						},
						"enabled": schema.BoolAttribute{
							MarkdownDescription: "If the users is/are currently active or not.",
							Description:         "If the users is/are currently active or not.",
							Optional:            true,
							Computed:            true,
						},
					},
				},
			},
		},
		Blocks: RedfishServerResourceBlockMap(),
	}
}

// Create creates the resource and sets the initial Terraform state.
func (r *UserAccountResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	tflog.Trace(ctx, "resource_user_account create : Started")

	r.ctx = ctx
	// Get Plan Data
	var plan models.UserAccount
	var userID string
	var userIDs []string
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Lock the mutex to avoid race conditions with other resources
	redfishMutexKV.Lock(plan.RedfishServer[0].Endpoint.ValueString())
	defer redfishMutexKV.Unlock(plan.RedfishServer[0].Endpoint.ValueString())

	service, err := NewConfig(r.p, &plan.RedfishServer)
	if err != nil {
		resp.Diagnostics.AddError(ServiceErrorMsg, err.Error())
		return
	}

	if plan.UserDetails.IsUnknown() || plan.UserDetails.IsNull() {
		userID, diags = r.createUser(plan, service, models.UserDetails{})
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}
		_, account, err := GetUserAccountFromID(service, userID)
		if err != nil {
			resp.Diagnostics.AddError(RedfishFetchErrorMsg, err.Error())
			return
		}

		result := models.UserAccount{}
		r.updateServer(&plan, &result, account, operationCreate)

		tflog.Trace(ctx, "resource_user_account create: updating state finished, saving ...")

		// Save into State
		diags = resp.State.Set(ctx, result)
		resp.Diagnostics.Append(diags...)
		return
	}

	userList, diags := r.getUserDetailsList(&plan)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	for _, user := range userList {
		userID, diags = r.createUser(plan, service, user)
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}
		userIDs = append(userIDs, userID)
	}

	tflog.Trace(ctx, "resource_user_account create: updating state finished, saving ...")
	result := models.UserAccount{}
	diags = r.updateServerMultipleUser(&plan, &result, service, userIDs)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	// Save into State
	diags = resp.State.Set(ctx, result)
	resp.Diagnostics.Append(diags...)
	tflog.Trace(ctx, "resource_user_account create: finish")
}

// Read refreshes the Terraform state with the latest data.
func (r *UserAccountResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	tflog.Trace(ctx, "resource_user_account read: started")

	var state models.UserAccount
	var userIDs []string

	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	service, err := NewConfig(r.p, &state.RedfishServer)
	if err != nil {
		resp.Diagnostics.AddError(ServiceErrorMsg, err.Error())
		return
	}

	if !(state.UserDetails.IsUnknown() || state.UserDetails.IsNull()) {
		userList := make([]models.UserDetails, 0)
		diags.Append(state.UserDetails.ElementsAs(ctx, &userList, false)...)
		for _, user := range userList {
			userIDs = append(userIDs, user.UserID.ValueString())
		}
		r.updateServerMultipleUser(nil, &state, service, userIDs)
	} else {
		_, account, err := GetUserAccountFromID(service, state.ID.ValueString())
		if err != nil {
			resp.Diagnostics.AddError(RedfishFetchErrorMsg, err.Error())
		}
		if account == nil { // User doesn't exist. Needs to be recreated.
			resp.Diagnostics.AddError("Error when retrieving accounts", "User does not exists, needs to be recreated")
			return
		}
		r.updateServer(nil, &state, account, operationRead)
	}

	tflog.Trace(ctx, "resource_user_account read: finished reading state")
	// Save into State
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	tflog.Trace(ctx, "resource_user_account read: finished")
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *UserAccountResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Get state Data
	tflog.Trace(ctx, "resource_user_account update: started")

	var state, plan models.UserAccount

	// Get current state
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

	// Lock the mutex to avoid race conditions with other resources
	redfishMutexKV.Lock(plan.RedfishServer[0].Endpoint.ValueString())
	defer redfishMutexKV.Unlock(plan.RedfishServer[0].Endpoint.ValueString())

	service, err := NewConfig(r.p, &plan.RedfishServer)
	if err != nil {
		resp.Diagnostics.AddError(ServiceErrorMsg, err.Error())
		return
	}

	// validate Password
	err = validatePassword(plan.Password.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(RedfishPasswordErrorMsg, err.Error())
		return
	}

	accountList, account, err := GetUserAccountFromID(service, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(RedfishFetchErrorMsg, err.Error())
	}

	if plan.UserID.ValueString() != "" && plan.UserID.ValueString() != account.ID {
		resp.Diagnostics.AddError("user_id cannot be updated", "")
		return
	}

	// check if the username already exists
	if plan.Username.ValueString() != account.UserName {
		err = checkUserNameExists(accountList, plan.Username.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("Cannot check exsting user", err.Error())
			return
		}
	}

	payload := make(map[string]interface{})
	payload["UserName"] = plan.Username.ValueString()
	payload["Password"] = plan.Password.ValueString()
	payload["Enabled"] = plan.Enabled.ValueBool()
	payload["RoleId"] = plan.RoleID.ValueString()
	_, err = service.GetClient().Patch(account.ODataID, payload)
	if err != nil {
		resp.Diagnostics.AddError(RedfishAPIErrorMsg, err.Error())
		return
	}

	// get user which is updated
	_, account, err = GetUserAccountFromID(service, account.ID)
	if err != nil {
		resp.Diagnostics.AddError(RedfishFetchErrorMsg, err.Error())
	}
	if account == nil { // User doesn't exist. Needs to be recreated.
		resp.Diagnostics.AddError("Error when retrieving accounts", "User does not exists, needs to be recreated")
		return
	}
	r.updateServer(&plan, &state, account, operationUpdate)

	tflog.Trace(ctx, "resource_user_account update: finished state update")
	// Save into State
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	tflog.Trace(ctx, "resource_user_account update: finished")
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *UserAccountResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	tflog.Trace(ctx, "resource_user_account delete: started")
	// Get State Data
	var state models.UserAccount
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	service, err := NewConfig(r.p, &state.RedfishServer)
	if err != nil {
		resp.Diagnostics.AddError(ServiceErrorMsg, err.Error())
		return
	}

	if !state.UserDetails.IsUnknown() && !state.UserDetails.IsNull() {
		userList := make([]models.UserDetails, 0)
		diags.Append(state.UserDetails.ElementsAs(ctx, &userList, false)...)
		for _, user := range userList {
			r.deleteUser(service, user.UserID.ValueString())
		}
	} else {
		r.deleteUser(service, state.ID.ValueString())
	}
	tflog.Trace(ctx, "resource_user_account delete: finished")
}

// ImportState import state for existing user
func (*UserAccountResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	type creds struct {
		Username    string `json:"username"`
		Password    string `json:"password"`
		Endpoint    string `json:"endpoint"`
		SslInsecure bool   `json:"ssl_insecure"`
		Id          string `json:"id"`
	}

	var c creds
	err := json.Unmarshal([]byte(req.ID), &c)
	if err != nil {
		resp.Diagnostics.AddError("Error while unmarshalling id", err.Error())
	}

	server := models.RedfishServer{
		User:        types.StringValue(c.Username),
		Password:    types.StringValue(c.Password),
		Endpoint:    types.StringValue(c.Endpoint),
		SslInsecure: types.BoolValue(c.SslInsecure),
	}

	idAttrPath := tfpath.Root("id")
	redfishServer := tfpath.Root("redfish_server")
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, idAttrPath, c.Id)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, redfishServer, []models.RedfishServer{server})...)
}

func (UserAccountResource) deleteUser(service *gofish.Service, userID string) (string, diag.Diagnostics) {
	var diags diag.Diagnostics
	_, account, err := GetUserAccountFromID(service, userID)
	if err != nil {
		diags.AddError(RedfishPasswordErrorMsg, err.Error())
		return "", diags
	}

	// First set Role ID as "" and Enabled as false
	payload := make(map[string]interface{})
	payload["Enable"] = "false"
	payload["RoleId"] = "None"
	_, err = service.GetClient().Patch(account.ODataID, payload)
	if err != nil {
		diags.AddError(RedfishPasswordErrorMsg, err.Error())
		return "", diags
	}

	// second PATCH call to remove username.
	payload = make(map[string]interface{})
	payload["UserName"] = ""
	_, err = service.GetClient().Patch(account.ODataID, payload)
	if err != nil {
		diags.AddError(RedfishPasswordErrorMsg, err.Error())
		return "", diags
	}
	return "", nil
}

func (UserAccountResource) updateServer(plan, state *models.UserAccount, account *redfish.ManagerAccount, operation operation) {
	state.ID = types.StringValue(account.ID)
	state.Username = types.StringValue(account.UserName)
	state.Enabled = types.BoolValue(account.Enabled)
	state.RoleID = types.StringValue(account.RoleID)
	state.UserID = types.StringValue(account.ID)
	if operation != operationRead {
		state.Password = plan.Password
		state.RedfishServer = plan.RedfishServer
		state.UserDetails = plan.UserDetails
	}
}

func (r UserAccountResource) updateServerMultipleUser(plan, state *models.UserAccount, service *gofish.Service,
	userIDs []string,
) diag.Diagnostics {
	var diags diag.Diagnostics
	var userDetails []models.UserDetails
	var userData models.UserDetails
	var userList []models.UserDetails
	var role string

	if plan != nil {
		state.RedfishServer = plan.RedfishServer
		userList, diags = r.getUserDetailsList(plan)
		if plan.UserDetails.IsUnknown() || plan.UserDetails.IsNull() {
			role = "role_id"
		} else {
			role = "role"
		}
	} else {
		userList, diags = r.getUserDetailsList(state)
	}
	if diags.HasError() {
		diags.AddError("Error in processing user details", "Error when retrieving user account details as list")
		return diags
	}

	// get userdetails from server
	for i, userID := range userIDs {
		_, account, err := GetUserAccountFromID(service, userID)
		if err != nil {
			diags.AddError("Error in fecthing user details", err.Error())
			return diags
		}
		if plan != nil {
			userData = models.UserDetails{
				Password: userList[i].Password,
				RoleID:   types.StringValue(account.RoleID),
				UserID:   types.StringValue(account.ID),
				Username: types.StringValue(account.UserName),
				Enabled:  types.BoolValue(account.Enabled),
			}
		} else {
			userData = models.UserDetails{
				RoleID:   types.StringValue(account.RoleID),
				UserID:   types.StringValue(account.ID),
				Username: types.StringValue(account.UserName),
				Enabled:  types.BoolValue(account.Enabled),
				Password: userList[i].Password,
			}
		}
		userDetails = append(userDetails, userData)
	}

	userOptionsTypes := map[string]attr.Type{
		role:      types.StringType,
		"user_id": types.StringType,
		username:  types.StringType,
		"enabled": types.BoolType,
		password:  types.StringType,
	}

	userOptionsEleType := types.ObjectType{
		AttrTypes: userOptionsTypes,
	}

	userDetailsValue, diags := types.ListValueFrom(context.Background(), userOptionsEleType, userDetails)
	if diags.HasError() {
		diags.AddError("Error in processing user details", "Error in updating user details to state file")
		return diags
	}

	state.UserDetails = userDetailsValue
	state.ID = types.StringValue("user")
	return nil
}

func getAccountList(c *gofish.Service) ([]*redfish.ManagerAccount, error) {
	accountService, err := c.AccountService()
	if err != nil {
		return nil, err
	}
	accounts, err := accountService.Accounts()
	if err != nil {
		return nil, err
	}
	return accounts, nil
}

func getAccount(accountList []*redfish.ManagerAccount, id string) (*redfish.ManagerAccount, error) {
	for _, account := range accountList {
		if account.ID == id && len(account.UserName) > 0 {
			return account, nil
		}
	}
	return nil, nil // This will be returned if there are no errors but the user does not exist
}

// To check if given username is equal to any existing username
func checkUserNameExists(accountList []*redfish.ManagerAccount, username string) error {
	for _, account := range accountList {
		if username == account.UserName {
			return fmt.Errorf("user %v already exists against ID %v. Please enter a different user name", username, account.ID)
		}
	}
	return nil
}

// To check if given ID already exists
func checkUserIDExists(accountList []*redfish.ManagerAccount, userID string) error {
	for _, account := range accountList {
		if len(userID) > 0 && userID == account.ID && len(account.UserName) != 0 {
			return fmt.Errorf("user ID %v already exists. Please enter a valid user ID", userID)
		}
	}
	return nil
}

// To validate password
func validatePassword(password string) error {
	hasLowerCase := regexp.MustCompile(`[a-z]`).MatchString(password)
	hasUpperCase := regexp.MustCompile(`[A-Z]`).MatchString(password)
	hasNumber := regexp.MustCompile(`\d`).MatchString(password)
	hasSpecialChar := strings.ContainsAny(password, "'-!\"#$%&()*,./:;?@[\\]^_`{|}~+<=>")
	if !hasLowerCase || !hasUpperCase || !hasNumber || !hasSpecialChar {
		return fmt.Errorf("validation failed. The password must include one uppercase and one lower case letter, one number and a special character")
	}
	return nil
}

// GetUserAccountFromID fetches specific user details for the given userID
func GetUserAccountFromID(service *gofish.Service, userID string) ([]*redfish.ManagerAccount, *redfish.ManagerAccount, error) {
	accountList, err := getAccountList(service)
	if err != nil {
		return nil, nil, fmt.Errorf("error when retrieving account list %v", err.Error())
	}

	// get user which is created
	account, err := getAccount(accountList, userID)
	if err != nil {
		return nil, nil, fmt.Errorf("error when retrieving accounts %v", err.Error())
	}
	return accountList, account, nil
}

func (*UserAccountResource) createUser(plan models.UserAccount, service *gofish.Service, user models.UserDetails) (string, diag.Diagnostics) {
	var diags diag.Diagnostics
	var password, userName, userID, roleID string
	var enabled bool
	if plan.UserDetails.IsNull() || plan.UserDetails.IsUnknown() {
		password = plan.Password.ValueString()
		userName = plan.Username.ValueString()
		userID = plan.UserID.ValueString()
		roleID = plan.RoleID.ValueString()
		enabled = plan.Enabled.ValueBool()
	} else {
		password = user.Password.ValueString()
		userName = user.Username.ValueString()
		userID = user.UserID.ValueString()
		enabled = user.Enabled.ValueBool()
		roleID = user.RoleID.ValueString()
	}

	// validate Password
	err := validatePassword(password)
	if err != nil {
		diags.AddError(RedfishPasswordErrorMsg, err.Error())
		return "", diags
	}

	accountList, err := getAccountList(service)
	if err != nil {
		diags.AddError("Error when retrieving account list ", err.Error())
		return "", diags
	}

	// check if username already exists
	err = checkUserNameExists(accountList, userName)
	if err != nil {
		diags.AddError("Cannot check exsting user ", err.Error())
		return "", diags
	}

	// check if user id already exists
	err = checkUserIDExists(accountList, userID)
	if err != nil {
		diags.AddError("User ID already exists %v", err.Error())
		return "", diags
	}

	// check if user id is valid or not
	if len(userID) > 0 {
		userIdInt, err := strconv.Atoi(userID)
		if !(userIdInt > minUserID && userIdInt <= maxUserID) {
			diags.AddError("User_id can vary between 3 to 16 only", "Please update user ID")
			return "", diags
		}
		if err != nil {
			diags.AddError("Invalid user ID", "Cannot convert user ID to int")
			return "", diags
		}
	}

	payload := make(map[string]interface{})
	for _, account := range accountList {
		if len(account.UserName) == 0 && account.ID != "1" { // ID 1 is reserved
			payload["UserName"] = userName
			payload["Password"] = password
			payload["Enabled"] = enabled
			payload["RoleId"] = roleID
			if len(userID) > 0 {
				// update the account.ODataID URL to new account ID
				account.ID = userID
				url, _ := filepath.Split(account.ODataID)
				account.ODataID = url + account.ID
			} else {
				userID = account.ID
			}
			// Ideally a go routine for each server should be done
			_, err = service.GetClient().Patch(account.ODataID, payload)
			if err != nil {
				diags.AddError(RedfishAPIErrorMsg, err.Error()) // This error might happen when a user was created outside terraform
				return "", diags
			}
			break
		} else if account.ID == "17" {
			// No room for new users
			diags.AddError("There is no room for new users", "Please remove an existing user to proceed")
			return "", diags
		}
	}
	return userID, nil
}

func (r *UserAccountResource) getUserDetailsList(d *models.UserAccount) ([]models.UserDetails, diag.Diagnostics) {
	var diags diag.Diagnostics
	userList := make([]models.UserDetails, 0)
	diags.Append(d.UserDetails.ElementsAs(r.ctx, &userList, false)...)
	return userList, diags
}
