package redfish

import (
	"context"
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
				Type:     schema.TypeMap,
				Optional: true,
				Computed: true,
			},
		},
	}
}

func resourceUserAccountCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	userIDs := make(map[string]string)

	c := m.([]*ClientConfig)
	for _, v := range c {
		client := v.API.(*gofish.APIClient)
		accountList, err := getAccountList(client.Service)
		if err != nil {
			return diag.Errorf("Error when retrieving account list %v", err)
		}
		payload := make(map[string]interface{})
		for _, account := range accountList {
			if len(account.UserName) == 0 && account.ID != "1" { //ID 1 is reserved
				payload["UserName"] = d.Get("username").(string)
				payload["Password"] = d.Get("password").(string)
				payload["Enabled"] = d.Get("enabled").(bool)
				payload["RoleId"] = d.Get("role_id").(string)
				//Ideally a go routine for each server should be done
				res, err := v.API.Patch(account.ODataID, payload)
				if err != nil {
					//If something fails, we have to keep track of what's done so far
					d.SetId("Users")
					d.Set("users_id", userIDs)
					return diag.Errorf("Error when contacting the redfish API %v", err) //This error might happen when a user was created outside terraform
				}
				if res.StatusCode != 200 {
					d.SetId("Users")
					d.Set("users_id", userIDs)
					return diag.Errorf("There was an issue with the APIClient. HTTP error code %d", res.StatusCode)
				}
				userIDs[v.Endpoint] = account.ID
				break //Finish the loop, don't want another user created
			}
		}
		//No room for new users
		//return fmt.Errorf("There are no room for new users")
	}
	d.SetId("Users")
	d.Set("users_id", userIDs)
	return diags

}

func resourceUserAccountRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	users := d.Get("users_id").(map[string]interface{})
	// readUsers := make(map[string]string)
	c := m.([]*ClientConfig)
	for _, v := range c {
		log.Printf("[ReadContext] Checking client with endpoint %s", v.Endpoint)
		client := v.API.(*gofish.APIClient)
		accountList, err := getAccountList(client.Service)
		if err != nil {
			return diag.Errorf("Error when retrieving account list %v", err)
		}
		//users[v.Endpoint] is nil if user is not in the map
		// var account *redfish.ManagerAccount = nil
		if _, ok := users[v.Endpoint]; ok { //We only care about resources created by Terraform
			account, err := getAccount(accountList, users[v.Endpoint].(string))
			if err != nil {
				return diag.Errorf("Error when retrieving accounts %v", err)
			}
			if account == nil {
				//If account is nil means that does not exist and we need to create it
				d.Set("username", "")
				d.Set("enabled", "")
				d.Set("role_id", "")
				return diags
			}
			if d.Get("username") != account.UserName || d.Get("enabled") != account.Enabled || d.Get("role_id") != account.RoleID {
				// If something is different, even just one, we need to trigger an update and return
				d.Set("username", account.UserName)
				d.Set("enabled", account.Enabled)
				d.Set("role_id", account.RoleID)
				return diags
			}
		}
	}
	return diags
}

func resourceUserAccountUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.([]*ClientConfig)
	users := d.Get("users_id").(map[string]interface{})
	for _, v := range c {
		client := v.API.(*gofish.APIClient)
		accountList, err := getAccountList(client.Service)
		if err != nil {
			return diag.Errorf("Error when retrieving account list %v", err)
		}
		if _, ok := users[v.Endpoint]; ok { //We only care about resources created by Terraform
			account, err := getAccount(accountList, users[v.Endpoint].(string))
			if err != nil {
				return diag.Errorf("Error when retrieving accounts %v", err)
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
						res, err := v.API.Patch(account.ODataID, payload)
						if err != nil {
							d.Set("users_id", users) //I'll get rid of this when it's done in a go routine
							return diag.Errorf("Error when contacting the redfish API %v", err)
						}
						if res.StatusCode != 200 {
							d.Set("users_id", users) //I'll get rid of this when it's done in a go routine
							return diag.Errorf("There was an issue with the APIClient. HTTP error code %d", res.StatusCode)
						}
						users[v.Endpoint] = account.ID
						break //Finish the loop, don't want another user created
					}
				}
			} else {
				if d.Get("username") != account.UserName || d.Get("enabled") != account.Enabled || d.Get("role_id") != account.RoleID {
					payload := make(map[string]interface{})
					payload["UserName"] = d.Get("username")
					payload["Password"] = d.Get("password")
					payload["Enabled"] = d.Get("enabled")
					payload["RoleId"] = d.Get("role_id")
					res, err := v.API.Patch(account.ODataID, payload) //null!!! Myabe nil scenario should be taken apart from the conditional above
					if err != nil {
						d.Set("users_id", users) //I'll get rid of this when it's done in a go routine
						return diag.Errorf("Error when contacting the redfish API %v", err)
					}
					if res.StatusCode != 200 {
						d.Set("users_id", users) //I'll get rid of this when it's done in a go routine
						return diag.Errorf("There was an issue with the APIClient. HTTP error code %d", res.StatusCode)
					}
				}
			}
		}
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
	//Get subresources
	users := d.Get("users_id").(map[string]interface{})
	c := m.([]*ClientConfig)
	for _, v := range c {
		client := v.API.(*gofish.APIClient)
		accountList, err := getAccountList(client.Service)
		if err != nil {
			return diag.Errorf("Error when retrieving account list %v", err)
		}
		//users[v.Endpoint] is nil if user is not in the map
		// var account *redfish.ManagerAccount = nil
		if _, ok := users[v.Endpoint]; ok {
			account, err := getAccount(accountList, users[v.Endpoint].(string))
			if err != nil {
				return diag.Errorf("Error when retrieving accounts %v", err)
			}
			if account == nil {
				//return diag.Errorf("The user account does not exist")
				delete(users, users[v.Endpoint].(string))
			}
			payload := make(map[string]interface{})
			payload["UserName"] = ""
			res, err := v.API.Patch(account.ODataID, payload)
			if err != nil {
				return diag.Errorf("Error when contacting the redfish API %v", err)
			}
			if res.StatusCode != 200 {
				return diag.Errorf("There was an issue with the APIClient. HTTP error code %d", res.StatusCode)
			}
		}
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
