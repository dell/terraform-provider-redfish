package redfish

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/stmcginnis/gofish"
	"github.com/stmcginnis/gofish/redfish"
	"path/filepath"
	"strconv"
)

func resourceRedfishUserAccount() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceRedfishUserAccountCreate,
		ReadContext:   resourceRedfishUserAccountRead,
		UpdateContext: resourceRedfishUserAccountUpdate,
		DeleteContext: resourceRedfishUserAccountDelete,
		Schema:        getResourceRedfishUserAccountSchema(),
	}
}

func getResourceRedfishUserAccountSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"redfish_server": {
			Type:        schema.TypeList,
			Required:    true,
			Description: "This list contains the different redfish endpoints to manage (different servers)",
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"user": {
						Type:        schema.TypeString,
						Optional:    true,
						Description: "This field is the user to login against the redfish API",
					},
					"password": {
						Type:        schema.TypeString,
						Optional:    true,
						Description: "This field is the password related to the user given",
					},
					"endpoint": {
						Type:        schema.TypeString,
						Required:    true,
						Description: "This field is the endpoint where the redfish API is placed",
					},
					"ssl_insecure": {
						Type:        schema.TypeBool,
						Optional:    true,
						Description: "This field indicates if the SSL/TLS certificate must be verified",
					},
				},
			},
		},
		"user_id": {
			Type:     schema.TypeString,
			Optional: true,
			Computed: true,
		},
		"username": {
			Type:     schema.TypeString,
			Required: true,
		},
		"password": {
			Type:      schema.TypeString,
			Required:  true,
			Sensitive: true,
		},
		"enabled": {
			Type:     schema.TypeBool,
			Optional: true,
			Default:  false,
		},
		"role_id": {
			Type:     schema.TypeString,
			Optional: true,
			Default:  "None",
			Description: "Applicable values are 'Operator', 'Administrator', 'None', and 'ReadOnly'. " +
				"Default is \"None\".",
			ValidateFunc: validation.StringInSlice([]string{
				"Operator",
				"Administrator",
				"ReadOnly",
				"None",
			}, false),
		},
	}
}

func resourceRedfishUserAccountCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	service, err := NewConfig(m.(*schema.ResourceData), d)
	if err != nil {
		return diag.Errorf(err.Error())
	}
	return createRedfishUserAccount(service, d)
}

func resourceRedfishUserAccountRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	service, err := NewConfig(m.(*schema.ResourceData), d)
	if err != nil {
		return diag.Errorf(err.Error())
	}
	return readRedfishUserAccount(service, d)
}

func resourceRedfishUserAccountUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	service, err := NewConfig(m.(*schema.ResourceData), d)
	if err != nil {
		return diag.Errorf(err.Error())
	}
	if diags := updateRedfishUserAccount(ctx, service, d, m); diags.HasError() {
		return diags
	}
	return resourceRedfishUserAccountRead(ctx, d, m)
}

func resourceRedfishUserAccountDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	service, err := NewConfig(m.(*schema.ResourceData), d)
	if err != nil {
		return diag.Errorf(err.Error())
	}
	return deleteRedfishUserAccount(service, d)
}

func createRedfishUserAccount(service *gofish.Service, d *schema.ResourceData) diag.Diagnostics {
	var diags diag.Diagnostics

	// Lock the mutex to avoid race conditions with other resources
	redfishMutexKV.Lock(getRedfishServerEndpoint(d))
	defer redfishMutexKV.Unlock(getRedfishServerEndpoint(d))

	accountList, err := getAccountList(service)
	if err != nil {
		return diag.Errorf("Error when retrieving account list %v", err)
	}

	// check if username already exists
	err = checkUserNameExists(accountList, d.Get("username").(string))
	if err != nil {
		return diag.Errorf(err.Error())
	}

	// check if user id is valid or not
	userIdInt, err := strconv.Atoi(d.Get("user_id").(string))
	if !(userIdInt > 2 && userIdInt <= 16) {
		return diag.Errorf("User_id can vary between 3 to 16 only")
	}

	payload := make(map[string]interface{})
	for _, account := range accountList {
		if len(account.UserName) == 0 && account.ID != "1" { //ID 1 is reserved
			payload["UserName"] = d.Get("username").(string)
			payload["Password"] = d.Get("password").(string)
			payload["Enabled"] = d.Get("enabled").(bool)
			payload["RoleId"] = d.Get("role_id").(string)
			if len(d.Get("user_id").(string)) > 0 {
				// update the account.ODataID URL to new account ID
				account.ID = d.Get("user_id").(string)
				url, _ := filepath.Split(account.ODataID)
				account.ODataID = url + account.ID
			}
			//Ideally a go routine for each server should be done
			res, err := service.Client.Patch(account.ODataID, payload)
			if err != nil {
				return diag.Errorf("Error when contacting the redfish API %v", err) //This error might happen when a user was created outside terraform
			}
			if res.StatusCode != 200 {
				return diag.Errorf("There was an issue with the APIClient. HTTP error code %d", res.StatusCode)
			}
			//Set ID to terraform state file
			d.SetId(account.ID)
			diags = readRedfishUserAccount(service, d)
			return diags
		}
	}
	//No room for new users
	return diag.Errorf("There are no room for new users")
}

func readRedfishUserAccount(service *gofish.Service, d *schema.ResourceData) diag.Diagnostics {
	var diags diag.Diagnostics

	accountList, err := getAccountList(service)
	if err != nil {
		return diag.Errorf("Error when retrieving account list %v", err)
	}

	account, err := getAccount(accountList, d.Id())
	if err != nil {
		return diag.Errorf("Error when retrieving accounts %v", err)
	}
	if account == nil { //User doesn't exist. Needs to be recreated.
		d.SetId("")
		return diags
	}

	d.Set("username", account.UserName)
	d.Set("enabled", account.Enabled)
	d.Set("role_id", account.RoleID)

	return diags
}

func updateRedfishUserAccount(ctx context.Context, service *gofish.Service, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	var userUpdated bool

	// Lock the mutex to avoid race conditions with other resources
	redfishMutexKV.Lock(getRedfishServerEndpoint(d))
	defer redfishMutexKV.Unlock(getRedfishServerEndpoint(d))

	accountList, err := getAccountList(service)
	if err != nil {
		return diag.Errorf("Error when retrieving account list %v", err)
	}

	account, err := getAccount(accountList, d.Id())
	if err != nil {
		return diag.Errorf("Error when retrieving accounts %v", err)
	}

	if d.Get("user_id") != account.ID {
		return diag.Errorf("user_id cannot be updated")
	}

	// check if the username already exists
	if d.Get("username") != account.UserName {
		err = checkUserNameExists(accountList, d.Get("username").(string))
		if err != nil {
			return diag.Errorf(err.Error())
		}
		userUpdated = true
	}
	if userUpdated || d.Get("enabled") != account.Enabled || d.Get("role_id") != account.RoleID || d.Get("password") != account.Password {
		payload := make(map[string]interface{})
		payload["UserName"] = d.Get("username")
		payload["Password"] = d.Get("password")
		payload["Enabled"] = d.Get("enabled")
		payload["RoleId"] = d.Get("role_id")
		res, err := service.Client.Patch(account.ODataID, payload)
		if err != nil {
			return diag.Errorf("Error when contacting the redfish API %v", err)
		}
		if res.StatusCode != 200 {
			return diag.Errorf("There was an issue with the server. HTTP error code %d", res.StatusCode)
		}
	}
	return diags
}

func deleteRedfishUserAccount(service *gofish.Service, d *schema.ResourceData) diag.Diagnostics {
	var diags diag.Diagnostics

	// Lock the mutex to avoid race conditions with other resources
	redfishMutexKV.Lock(getRedfishServerEndpoint(d))
	defer redfishMutexKV.Unlock(getRedfishServerEndpoint(d))

	accountList, err := getAccountList(service)
	if err != nil {
		return diag.Errorf("Error when retrieving account list %v", err)
	}

	account, err := getAccount(accountList, d.Id())
	if err != nil {
		return diag.Errorf("Error when retrieving accounts %v", err)
	}
	// First set Role ID as "" and Enabled as false
	payload := make(map[string]interface{})
	payload["Enable"] = "false"
	payload["RoleId"] = "None"
	res, err := service.Client.Patch(account.ODataID, payload)
	if err != nil {
		return diag.Errorf("Error when contacting the redfish API %v", err)
	}
	if res.StatusCode != 200 {
		return diag.Errorf("There was an issue with the server. HTTP error code %d", res.StatusCode)
	}

	// second PATCH call to remove username.
	payload = make(map[string]interface{})
	payload["UserName"] = ""
	res, err = service.Client.Patch(account.ODataID, payload)
	if err != nil {
		return diag.Errorf("Error when contacting the redfish API %v", err)
	}
	if res.StatusCode != 200 {
		return diag.Errorf("There was an issue with the server. HTTP error code %d", res.StatusCode)
	}

	d.SetId("")
	return diags
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
	return nil, nil //This will be returned if there are no errors but the user does not exist
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
