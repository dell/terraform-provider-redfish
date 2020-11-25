package redfish

import (
	"context"
	"fmt"
	"github.com/dell/terraform-provider-redfish/common"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stmcginnis/gofish"
	"github.com/stmcginnis/gofish/redfish"
	"log"
)

func resourceUserAccount() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceUserAccountCreate,
		ReadContext:   resourceUserAccountRead,
		UpdateContext: resourceUserAccountUpdate,
		DeleteContext: resourceUserAccountDelete,
		Schema: map[string]*schema.Schema{
			"username": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"password": &schema.Schema{
				Type:      schema.TypeString,
				Required:  true,
				Sensitive: true,
			},
			"enabled": &schema.Schema{
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"role_id": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Default:  "None",
			},
			"users_id": &schema.Schema{
				Type: schema.TypeMap,
				//Optional: true,
				Computed: true,
			},
		},
	}
}

func resourceUserAccountCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	userResult := make(chan common.ResourceResult, len(m.([]*ClientConfig)))
	c := m.([]*ClientConfig)
	for _, v := range c {
		go func(v *ClientConfig, userResult chan common.ResourceResult) {
			//client := v.API.(*gofish.APIClient)
			accountList, err := getAccountList(v.Service)
			if err != nil {
				userResult <- common.ResourceResult{Endpoint: v.Endpoint, ID: "", Error: true, ErrorMsg: fmt.Sprintf("[%v] Error when retrieving account list %v", v.Endpoint, err)}
				return
			}
			payload := make(map[string]interface{})
			for _, account := range accountList {
				if len(account.UserName) == 0 && account.ID != "1" { //ID 1 is reserved
					payload["UserName"] = d.Get("username").(string)
					payload["Password"] = d.Get("password").(string)
					payload["Enabled"] = d.Get("enabled").(bool)
					payload["RoleId"] = d.Get("role_id").(string)
					res, err := v.Service.Client.Patch(account.ODataID, payload)
					if err != nil {
						userResult <- common.ResourceResult{Endpoint: v.Endpoint, ID: "", Error: true, ErrorMsg: fmt.Sprintf("[%v] Error when contacting the redfish API %v", v.Endpoint, err)}
						return
					}
					if res.StatusCode != 200 {
						userResult <- common.ResourceResult{Endpoint: v.Endpoint, ID: "", Error: true, ErrorMsg: fmt.Sprintf("[%v] There was an issue with the APIClient. HTTP error code %d", v.Endpoint, res.StatusCode)}
						return
					}
					userResult <- common.ResourceResult{Endpoint: v.Endpoint, ID: account.ID, Error: false, ErrorMsg: ""}
					return //Finish the loop, don't want another user created
				}
			}
			//No room for new users
			userResult <- common.ResourceResult{Endpoint: v.Endpoint, ID: "", Error: true, ErrorMsg: fmt.Sprintf("[%v] No room for new users", v.Endpoint)}
		}(v, userResult)
	}
	userIDs := make(map[string]string)
	var errorMsg string
	for i := 0; i < len(m.([]*ClientConfig)); i++ {
		result := <-userResult
		if result.Error {
			errorMsg += result.ErrorMsg
		}
		userIDs[result.Endpoint] = result.ID
	}
	close(userResult)
	d.SetId("Users")
	d.Set("users_id", userIDs)
	if len(errorMsg) > 0 {
		return diag.Errorf(errorMsg)
	}
	return diags

}

func resourceUserAccountRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	userChanged := make(chan common.ResourceChanged, len(m.([]*ClientConfig)))
	users := d.Get("users_id").(map[string]interface{})
	c := m.([]*ClientConfig)
	for _, v := range c {
		go func(v *ClientConfig, userChanged chan common.ResourceChanged, d *schema.ResourceData) {
			log.Printf("[ReadContext] Checking client with endpoint %s", v.Endpoint)
			//client := v.API.(*gofish.APIClient)
			accountList, err := getAccountList(v.Service)
			if err != nil {
				userChanged <- common.ResourceChanged{Error: true, ErrorMessage: fmt.Sprintf("[%v] Error when retrieving account list %v", v.Endpoint, err)}
				return
			}
			account, err := getAccount(accountList, users[v.Endpoint].(string))
			if err != nil {
				userChanged <- common.ResourceChanged{Error: true, ErrorMessage: fmt.Sprintf("[%v] Error when retrieving accounts %v", v.Endpoint, err)}
				return
			}
			if account == nil || d.Get("username") != account.UserName || d.Get("enabled") != account.Enabled || d.Get("role_id") != account.RoleID {
				// If something is different, even just one, we need to trigger an update and return
				log.Printf("[ReadContext] Need to update users on client %s", v.Endpoint)
				userChanged <- common.ResourceChanged{HasChanged: true}
				return
			}
			userChanged <- common.ResourceChanged{HasChanged: false}
			log.Printf("[ReadContext] Nothing to update regarding users on client %s", v.Endpoint)
		}(v, userChanged, d)
	}
	var errorMsg string
	for i := 0; i < len(m.([]*ClientConfig)); i++ {
		changed := <-userChanged
		if changed.Error {
			errorMsg += changed.ErrorMessage
		} else {
			if changed.HasChanged { //If needs update
				d.Set("username", "")
				d.Set("enabled", "")
				d.Set("role_id", "")
				break
			}
		}
	}
	close(userChanged)
	if len(errorMsg) > 0 {
		return diag.Errorf(errorMsg)
	}
	return diags
}

func resourceUserAccountUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.([]*ClientConfig)
	users := d.Get("users_id").(map[string]interface{})
	userResult := make(chan common.ResourceResult, len(m.([]*ClientConfig)))
	for _, v := range c {
		go func(v *ClientConfig, userResult chan common.ResourceResult) {
			//client := v.API.(*gofish.APIClient)
			accountList, err := getAccountList(v.Service)
			if err != nil {
				userResult <- common.ResourceResult{Endpoint: v.Endpoint, Error: true, ErrorMsg: fmt.Sprintf("[%v] Error when retrieving account list %v", v.Endpoint, err)}
				return
			}
			account, err := getAccount(accountList, users[v.Endpoint].(string))
			if err != nil {
				userResult <- common.ResourceResult{Endpoint: v.Endpoint, Error: true, ErrorMsg: fmt.Sprintf("[%v] Error when retrieving accounts %v", v.Endpoint, err)}
				return
			}
			//If account does not exist or if params are not right, perform POST
			if account == nil {
				//Create a new one as we do in create
				payload := make(map[string]interface{})
				for _, account := range accountList {
					if len(account.UserName) == 0 && account.ID != "1" { //ID 1 is reserved
						payload["UserName"] = d.Get("username").(string)
						payload["Password"] = d.Get("password").(string)
						payload["Enabled"] = d.Get("enabled").(bool)
						payload["RoleId"] = d.Get("role_id").(string)
						res, err := v.Service.Client.Patch(account.ODataID, payload)
						if err != nil {
							userResult <- common.ResourceResult{Endpoint: v.Endpoint, Error: true, ErrorMsg: fmt.Sprintf("[%v] Error when contacting the redfish API %v", v.Endpoint, err)}
							return
						}
						if res.StatusCode != 200 {
							userResult <- common.ResourceResult{Endpoint: v.Endpoint, Error: true, ErrorMsg: fmt.Sprintf("[%v] There was an issue with the APIClient. HTTP error code %d", v.Endpoint, res.StatusCode)}
							return
						}
						userResult <- common.ResourceResult{Endpoint: v.Endpoint, ID: account.ID}
						return
					}
				}
				// No more room for users
				userResult <- common.ResourceResult{Endpoint: v.Endpoint, Error: true, ErrorMsg: fmt.Sprintf("[%v] No room for new users", v.Endpoint)}
				return
			}
			if d.Get("username") != account.UserName || d.Get("enabled") != account.Enabled || d.Get("role_id") != account.RoleID {
				payload := make(map[string]interface{})
				payload["UserName"] = d.Get("username")
				payload["Password"] = d.Get("password")
				payload["Enabled"] = d.Get("enabled")
				payload["RoleId"] = d.Get("role_id")
				res, err := v.Service.Client.Patch(account.ODataID, payload)
				if err != nil {
					userResult <- common.ResourceResult{Endpoint: v.Endpoint, Error: true, ErrorMsg: fmt.Sprintf("[%v] Error when contacting the redfish API %v", v.Endpoint, err)}
					return
				}
				if res.StatusCode != 200 {
					userResult <- common.ResourceResult{Endpoint: v.Endpoint, Error: true, ErrorMsg: fmt.Sprintf("[%v] There was an issue with the APIClient. HTTP error code %d", v.Endpoint, res.StatusCode)}
					return
				}
			}
			userResult <- common.ResourceResult{Endpoint: v.Endpoint, ID: account.ID}
			return
		}(v, userResult)
	}
	userIDs := make(map[string]string)
	var errorMsg string
	for i := 0; i < len(m.([]*ClientConfig)); i++ {
		result := <-userResult
		if result.Error {
			errorMsg += result.ErrorMsg
		}
		userIDs[result.Endpoint] = result.ID
	}
	close(userResult)
	if len(errorMsg) > 0 {
		return diag.Errorf(errorMsg)
	}
	d.Set("username", d.Get("username"))
	d.Set("password", d.Get("password"))
	d.Set("enabled", d.Get("enabled"))
	d.Set("role_id", d.Get("role_id"))
	d.Set("users_id", users)
	return resourceUserAccountRead(ctx, d, m)
}

func resourceUserAccountDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	userResult := make(chan common.ResourceResult, len(m.([]*ClientConfig)))
	//Get subresources
	users := d.Get("users_id").(map[string]interface{})
	c := m.([]*ClientConfig)
	for _, v := range c {
		go func(v *ClientConfig, userResult chan common.ResourceResult) {
			//client := v.API.(*gofish.APIClient)
			accountList, err := getAccountList(v.Service)
			if err != nil {
				userResult <- common.ResourceResult{Endpoint: v.Endpoint, Error: true, ErrorMsg: fmt.Sprintf("[%v] Error when retrieving account list %v", v.Endpoint, err)}
				return
			}
			account, err := getAccount(accountList, users[v.Endpoint].(string))
			if err != nil {
				userResult <- common.ResourceResult{Endpoint: v.Endpoint, Error: true, ErrorMsg: fmt.Sprintf("[%v] Error when retrieving accounts %v", v.Endpoint, err)}
				return
			}
			if account == nil {
				userResult <- common.ResourceResult{Endpoint: v.Endpoint, ID: ""}
				return
			}
			payload := make(map[string]interface{})
			payload["UserName"] = ""
			res, err := v.Service.Client.Patch(account.ODataID, payload)
			if err != nil {
				userResult <- common.ResourceResult{Endpoint: v.Endpoint, Error: true, ErrorMsg: fmt.Sprintf("[%v] Error when contacting the redfish API %v", v.Endpoint, err)}
				return
			}
			if res.StatusCode != 200 {
				userResult <- common.ResourceResult{Endpoint: v.Endpoint, Error: true, ErrorMsg: fmt.Sprintf("[%v] There was an issue with the APIClient. HTTP error code %d", v.Endpoint, res.StatusCode)}
				return
			}
			userResult <- common.ResourceResult{Endpoint: v.Endpoint, ID: ""}
			return
		}(v, userResult)

	}
	var errorMsg string
	for i := 0; i < len(m.([]*ClientConfig)); i++ {
		result := <-userResult
		if result.Error {
			errorMsg += result.ErrorMsg
		} else {
			delete(users, result.Endpoint)
		}
	}
	close(userResult)
	d.Set("users_id", users)
	if len(errorMsg) > 0 {
		return diag.Errorf(errorMsg)
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
	return nil, nil //This will be returned if there was no errors but the user does not exist
}
