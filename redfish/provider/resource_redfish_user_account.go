package provider

import (
	"context"
	"fmt"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"terraform-provider-redfish/redfish/models"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
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
	p *redfishProvider
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
			"username": schema.StringAttribute{
				MarkdownDescription: "The name of the user",
				Description:         "The name of the user",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(minUserNameLength, maxUserNameLength),
				},
			},
			"password": schema.StringAttribute{
				MarkdownDescription: "Password of the user",
				Description:         "Password of the user",
				Required:            true,
				Sensitive:           true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(minPasswordLength, maxPasswordLength),
				},
			},
			"role_id": schema.StringAttribute{
				MarkdownDescription: "Role of the user. Applicable values are 'Operator', 'Administrator', 'None', and 'ReadOnly'. " +
					"Default is \"None\"",
				Description: "Role of the user. Applicable values are 'Operator', 'Administrator', 'None', and 'ReadOnly'. " +
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
				MarkdownDescription: "If the user is currently active or not.",
				Description:         "If the user is currently active or not.",
				Optional:            true,
			},
		},
		Blocks: RedfishServerResourceBlockMap(),
	}
}

// Create creates the resource and sets the initial Terraform state.
func (r *UserAccountResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	tflog.Trace(ctx, "resource_UserAccount create : Started")

	// Get Plan Data
	var plan models.UserAccount
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	service, err := NewConfig(r.p, &plan.RedfishServer)
	if err != nil {
		resp.Diagnostics.AddError(ServiceErrorMsg, err.Error())
		return
	}

	tflog.Trace(ctx, "resource_UserAccount create: updating state finished, saving ...")
	password := plan.Password.ValueString()
	userName := plan.Username.ValueString()
	userID := plan.UserID.ValueString()

	// validate Password
	err = validatePassword(password)
	if err != nil {
		resp.Diagnostics.AddError("Password validation failed", err.Error())
		return
	}

	accountList, err := getAccountList(service)
	if err != nil {
		resp.Diagnostics.AddError("Error when retrieving account list", err.Error())
		return
	}

	// check if username already exists
	err = checkUserNameExists(accountList, userName)
	if err != nil {
		resp.Diagnostics.AddError("Cannot check exsting user", err.Error())
		return
	}

	// check if user id already exists
	err = checkUserIDExists(accountList, userID)
	if err != nil {
		resp.Diagnostics.AddError("User ID already exists", err.Error())
		return
	}

	// check if user id is valid or not
	if len(userID) > 0 {
		userIdInt, err := strconv.Atoi(userID)
		if !(userIdInt > minUserID && userIdInt <= maxUserID) {
			resp.Diagnostics.AddError("User_id can vary between 3 to 16 only", "Please update user ID")
			return
		}
		if err != nil {
			resp.Diagnostics.AddError("Invalid user ID", "Cannot convert user ID to int")
			return
		}
	}

	payload := make(map[string]interface{})
	for _, account := range accountList {
		if len(account.UserName) == 0 && account.ID != "1" { // ID 1 is reserved
			payload["UserName"] = userName
			payload["Password"] = password
			payload["Enabled"] = plan.Enabled.ValueBool()
			payload["RoleId"] = plan.RoleID.ValueString()
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
				resp.Diagnostics.AddError(RedfishAPIErrorMsg, err.Error()) // This error might happen when a user was created outside terraform
				return
			}
			break
		} else if account.ID == "17" {
			// No room for new users
			resp.Diagnostics.AddError("There is no room for new users", "Please remove an existing user to proceed")
			return
		}
	}

	_, account, err := getUserAccountFromID(service, userID)
	if err != nil {
		resp.Diagnostics.AddError(RedfishFetchErrorMsg, err.Error())
		return
	}

	result := models.UserAccount{}
	r.updateServer(&plan, &result, account, operationCreate)

	// Save into State
	diags = resp.State.Set(ctx, result)
	resp.Diagnostics.Append(diags...)
	tflog.Trace(ctx, "resource_UserAccount create: finish")
}

// Read refreshes the Terraform state with the latest data.
func (r *UserAccountResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	tflog.Trace(ctx, "resource_UserAccount read: started")

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

	_, account, err := getUserAccountFromID(service, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(RedfishFetchErrorMsg, err.Error())
	}

	if account == nil { // User doesn't exist. Needs to be recreated.
		resp.Diagnostics.AddError("Error when retrieving accounts", "User does not eists, needs to be recreated")
		return
	}

	r.updateServer(nil, &state, account, operationRead)

	tflog.Trace(ctx, "resource_UserAccount read: finished reading state")
	// Save into State
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	tflog.Trace(ctx, "resource_UserAccount read: finished")
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *UserAccountResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Get state Data
	tflog.Trace(ctx, "resource_UserAccount update: started")

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

	service, err := NewConfig(r.p, &plan.RedfishServer)
	if err != nil {
		resp.Diagnostics.AddError(ServiceErrorMsg, err.Error())
		return
	}

	// validate Password
	err = validatePassword(plan.Password.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Password validation failed", err.Error())
		return
	}

	accountList, account, err := getUserAccountFromID(service, state.ID.ValueString())
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
	_, account, err = getUserAccountFromID(service, account.ID)
	if err != nil {
		resp.Diagnostics.AddError(RedfishFetchErrorMsg, err.Error())
	}
	r.updateServer(&plan, &state, account, operationUpdate)

	tflog.Trace(ctx, "resource_UserAccount update: finished state update")
	// Save into State
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	tflog.Trace(ctx, "resource_UserAccount update: finished")
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *UserAccountResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	tflog.Trace(ctx, "resource_UserAccount delete: started")
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

	_, account, err := getUserAccountFromID(service, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(RedfishFetchErrorMsg, err.Error())
	}

	// First set Role ID as "" and Enabled as false
	payload := make(map[string]interface{})
	payload["Enable"] = "false"
	payload["RoleId"] = "None"
	_, err = service.GetClient().Patch(account.ODataID, payload)
	if err != nil {
		resp.Diagnostics.AddError(RedfishAPIErrorMsg, err.Error())
		return
	}

	// second PATCH call to remove username.
	payload = make(map[string]interface{})
	payload["UserName"] = ""
	_, err = service.GetClient().Patch(account.ODataID, payload)
	if err != nil {
		resp.Diagnostics.AddError(RedfishAPIErrorMsg, err.Error())
		return
	}

	tflog.Trace(ctx, "resource_UserAccount delete: finished")
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
	}
}

// import (
// 	"context"
// 	"fmt"
// 	"path/filepath"
// 	"regexp"
// 	"strconv"
// 	"strings"

// 	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
// 	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
// 	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
// 	"github.com/stmcginnis/gofish"
// 	"github.com/stmcginnis/gofish/redfish"
// )

// func resourceRedfishUserAccount() *schema.Resource {
// 	return &schema.Resource{
// 		CreateContext: resourceRedfishUserAccountCreate,
// 		ReadContext:   resourceRedfishUserAccountRead,
// 		UpdateContext: resourceRedfishUserAccountUpdate,
// 		DeleteContext: resourceRedfishUserAccountDelete,
// 		Schema:        getResourceRedfishUserAccountSchema(),
// 	}
// }

// func getResourceRedfishUserAccountSchema() map[string]*schema.Schema {
// 	return map[string]*schema.Schema{
// 		"redfish_server": {
// 			Type:        schema.TypeList,
// 			Required:    true,
// 			Description: "This list contains the different redfish endpoints to manage (different servers)",
// 			Elem: &schema.Resource{
// 				Schema: map[string]*schema.Schema{
// 					"user": {
// 						Type:        schema.TypeString,
// 						Optional:    true,
// 						Description: "This field is the user to login against the redfish API",
// 					},
// 					"password": {
// 						Type:        schema.TypeString,
// 						Optional:    true,
// 						Description: "This field is the password related to the user given",
// 					},
// 					"endpoint": {
// 						Type:        schema.TypeString,
// 						Required:    true,
// 						Description: "This field is the endpoint where the redfish API is placed",
// 					},
// 					"ssl_insecure": {
// 						Type:        schema.TypeBool,
// 						Optional:    true,
// 						Description: "This field indicates if the SSL/TLS certificate must be verified",
// 					},
// 				},
// 			},
// 		},
// 		"user_id": {
// 			Type:        schema.TypeString,
// 			Optional:    true,
// 			Computed:    true,
// 			Description: "The ID of the user. Cannot be updated.",
// 		},
// 		"username": {
// 			Type:         schema.TypeString,
// 			Required:     true,
// 			ValidateFunc: validation.StringLenBetween(1, 16),
// 			Description:  "The name of the user.",
// 		},
// 		"password": {
// 			Type:         schema.TypeString,
// 			Required:     true,
// 			Sensitive:    true,
// 			ValidateFunc: validation.StringLenBetween(4, 40),
// 			Description:  "Password of the user.",
// 		},
// 		"enabled": {
// 			Type:        schema.TypeBool,
// 			Optional:    true,
// 			Default:     false,
// 			Description: "If the user is currently active or not.",
// 		},
// 		"role_id": {
// 			Type:     schema.TypeString,
// 			Optional: true,
// 			Default:  "None",
// 			Description: "Applicable values are 'Operator', 'Administrator', 'None', and 'ReadOnly'. " +
// 				"Default is \"None\".",
// 			ValidateFunc: validation.StringInSlice([]string{
// 				"Operator",
// 				"Administrator",
// 				"ReadOnly",
// 				"None",
// 			}, false),
// 		},
// 	}
// }

// func resourceRedfishUserAccountCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
// 	service, err := NewConfig(m.(*schema.ResourceData), d)
// 	if err != nil {
// 		return diag.Errorf(err.Error())
// 	}
// 	return createRedfishUserAccount(service, d)
// }

// func resourceRedfishUserAccountRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
// 	service, err := NewConfig(m.(*schema.ResourceData), d)
// 	if err != nil {
// 		return diag.Errorf(err.Error())
// 	}
// 	return readRedfishUserAccount(service, d)
// }

// func resourceRedfishUserAccountUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
// 	service, err := NewConfig(m.(*schema.ResourceData), d)
// 	if err != nil {
// 		return diag.Errorf(err.Error())
// 	}
// 	if diags := updateRedfishUserAccount(ctx, service, d, m); diags.HasError() {
// 		return diags
// 	}
// 	return resourceRedfishUserAccountRead(ctx, d, m)
// }

// func resourceRedfishUserAccountDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
// 	service, err := NewConfig(m.(*schema.ResourceData), d)
// 	if err != nil {
// 		return diag.Errorf(err.Error())
// 	}
// 	return deleteRedfishUserAccount(service, d)
// }

// func createRedfishUserAccount(service *gofish.Service, d *schema.ResourceData) diag.Diagnostics {
// 	var diags diag.Diagnostics

// 	// Lock the mutex to avoid race conditions with other resources
// 	redfishMutexKV.Lock(getRedfishServerEndpoint(d))
// 	defer redfishMutexKV.Unlock(getRedfishServerEndpoint(d))

// 	// validate Password
// 	err := validatePassword(d.Get("password").(string))
// 	if err != nil {
// 		return diag.Errorf(err.Error())
// 	}

// 	accountList, err := getAccountList(service)
// 	if err != nil {
// 		return diag.Errorf("Error when retrieving account list %v", err)
// 	}

// 	// check if username already exists
// 	err = checkUserNameExists(accountList, d.Get("username").(string))
// 	if err != nil {
// 		return diag.Errorf(err.Error())
// 	}

// 	// check if user id already exists
// 	err = checkUserIDExists(accountList, d.Get("user_id").(string))
// 	if err != nil {
// 		return diag.Errorf(err.Error())
// 	}

// 	// check if user id is valid or not
// 	userIdInt, err := strconv.Atoi(d.Get("user_id").(string))
// 	if len(d.Get("user_id").(string)) > 0 && !(userIdInt > 2 && userIdInt <= 16) {
// 		return diag.Errorf("User_id can vary between 3 to 16 only")
// 	}

// 	payload := make(map[string]interface{})
// 	for _, account := range accountList {
// 		if len(account.UserName) == 0 && account.ID != "1" { //ID 1 is reserved
// 			payload["UserName"] = d.Get("username").(string)
// 			payload["Password"] = d.Get("password").(string)
// 			payload["Enabled"] = d.Get("enabled").(bool)
// 			payload["RoleId"] = d.Get("role_id").(string)
// 			if len(d.Get("user_id").(string)) > 0 {
// 				// update the account.ODataID URL to new account ID
// 				account.ID = d.Get("user_id").(string)
// 				url, _ := filepath.Split(account.ODataID)
// 				account.ODataID = url + account.ID
// 			}
// 			//Ideally a go routine for each server should be done
// 			res, err := service.GetClient().Patch(account.ODataID, payload)
// 			if err != nil {
//     		This error might happen when a user was created outside terraform
// 				return diag.Errorf("Error when contacting the redfish API %v", err)
// 			}
// 			if res.StatusCode != 200 {
// 				return diag.Errorf("There was an issue with the APIClient. HTTP error code %d", res.StatusCode)
// 			}
// 			//Set ID to terraform state file
// 			d.SetId(account.ID)
// 			diags = readRedfishUserAccount(service, d)
// 			return diags
// 		}
// 	}
// 	//No room for new users
// 	return diag.Errorf("There are no room for new users")
// }

// func readRedfishUserAccount(service *gofish.Service, d *schema.ResourceData) diag.Diagnostics {
// 	var diags diag.Diagnostics

// 	accountList, err := getAccountList(service)
// 	if err != nil {
// 		return diag.Errorf("Error when retrieving account list %v", err)
// 	}

// 	account, err := getAccount(accountList, d.Id())
// 	if err != nil {
// 		return diag.Errorf("Error when retrieving accounts %v", err)
// 	}
// 	if account == nil { //User doesn't exist. Needs to be recreated.
// 		d.SetId("")
// 		return diags
// 	}

// 	d.Set("username", account.UserName)
// 	d.Set("enabled", account.Enabled)
// 	d.Set("role_id", account.RoleID)
// 	d.Set("user_id", account.ID)

// 	return diags
// }

// func updateRedfishUserAccount(ctx context.Context, service *gofish.Service, d *schema.ResourceData, m interface{}) diag.Diagnostics {
// 	var diags diag.Diagnostics
// 	var userUpdated bool

// 	// Lock the mutex to avoid race conditions with other resources
// 	redfishMutexKV.Lock(getRedfishServerEndpoint(d))
// 	defer redfishMutexKV.Unlock(getRedfishServerEndpoint(d))

// 	// validate Password
// 	err := validatePassword(d.Get("password").(string))
// 	if err != nil {
// 		return diag.Errorf(err.Error())
// 	}

// 	accountList, err := getAccountList(service)
// 	if err != nil {
// 		return diag.Errorf("Error when retrieving account list %v", err)
// 	}

// 	account, err := getAccount(accountList, d.Id())
// 	if err != nil {
// 		return diag.Errorf("Error when retrieving accounts %v", err)
// 	}

// 	if d.Get("user_id").(string) != account.ID {
// 		return diag.Errorf("user_id cannot be updated")
// 	}

// 	// check if the username already exists
// 	if d.Get("username") != account.UserName {
// 		err = checkUserNameExists(accountList, d.Get("username").(string))
// 		if err != nil {
// 			return diag.Errorf(err.Error())
// 		}
// 		userUpdated = true
// 	}
// 	if userUpdated || d.Get("enabled") != account.Enabled || d.Get("role_id") != account.RoleID || d.Get("password") != account.Password {
// 		payload := make(map[string]interface{})
// 		payload["UserName"] = d.Get("username")
// 		payload["Password"] = d.Get("password")
// 		payload["Enabled"] = d.Get("enabled")
// 		payload["RoleId"] = d.Get("role_id")
// 		res, err := service.GetClient().Patch(account.ODataID, payload)
// 		if err != nil {
// 			return diag.Errorf("Error when contacting the redfish API %v", err)
// 		}
// 		if res.StatusCode != 200 {
// 			return diag.Errorf("There was an issue with the server. HTTP error code %d", res.StatusCode)
// 		}
// 	}
// 	return diags
// }

// func deleteRedfishUserAccount(service *gofish.Service, d *schema.ResourceData) diag.Diagnostics {
// 	var diags diag.Diagnostics

// 	// Lock the mutex to avoid race conditions with other resources
// 	redfishMutexKV.Lock(getRedfishServerEndpoint(d))
// 	defer redfishMutexKV.Unlock(getRedfishServerEndpoint(d))

// 	accountList, err := getAccountList(service)
// 	if err != nil {
// 		return diag.Errorf("Error when retrieving account list %v", err)
// 	}

// 	account, err := getAccount(accountList, d.Id())
// 	if err != nil {
// 		return diag.Errorf("Error when retrieving accounts %v", err)
// 	}
// 	// First set Role ID as "" and Enabled as false
// 	payload := make(map[string]interface{})
// 	payload["Enable"] = "false"
// 	payload["RoleId"] = "None"
// 	res, err := service.GetClient().Patch(account.ODataID, payload)
// 	if err != nil {
// 		return diag.Errorf("Error when contacting the redfish API %v", err)
// 	}
// 	if res.StatusCode != 200 {
// 		return diag.Errorf("There was an issue with the server. HTTP error code %d", res.StatusCode)
// 	}

// 	// second PATCH call to remove username.
// 	payload = make(map[string]interface{})
// 	payload["UserName"] = ""
// 	res, err = service.GetClient().Patch(account.ODataID, payload)
// 	if err != nil {
// 		return diag.Errorf("Error when contacting the redfish API %v", err)
// 	}
// 	if res.StatusCode != 200 {
// 		return diag.Errorf("There was an issue with the server. HTTP error code %d", res.StatusCode)
// 	}

// 	d.SetId("")
// 	return diags
// }

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

func getUserAccountFromID(service *gofish.Service, userID string) ([]*redfish.ManagerAccount, *redfish.ManagerAccount, error) {
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
