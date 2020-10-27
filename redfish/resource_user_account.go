package redfish

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stmcginnis/gofish"
	"github.com/stmcginnis/gofish/redfish"
)

func resourceUserAccount() *schema.Resource {
	return &schema.Resource{
		Create: resourceUserAccountCreate,
		Read:   resourceUserAccountRead,
		Update: resourceUserAccountUpdate,
		Delete: resourceUserAccountDelete,

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
			},
			"role_id": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
			"users_id": &schema.Schema{
				Type:     schema.TypeMap,
				Computed: true,
			},
		},
	}
}

func resourceUserAccountCreate(d *schema.ResourceData, m interface{}) error {
	userIDs := make(map[string]string)

	c := m.([]*ClientConfig)
	for _, v := range c {
		accountList, err := getAccountList(v.API)
		if err != nil {
			return err
		}
		payload := make(map[string]interface{})
		for _, account := range accountList {
			if len(account.UserName) == 0 && account.ID != "1" { //ID 1 is reserved
				payload["UserName"] = d.Get("username").(string)
				payload["Password"] = d.Get("password").(string)
				if value, set := d.GetOk("enabled"); set {
					payload["Enabled"] = value
				} else {
					payload["Enabled"] = false
				}
				if value, set := d.GetOk("role_id"); set {
					payload["RoleId"] = value
				} else {
					payload["RoleId"] = "None"
				}
				//Ideally a go routine for each server should be done
				res, err := v.API.Patch(account.ODataID, payload)
				if err != nil {
					return err
				}
				if res.StatusCode != 200 {
					return fmt.Errorf("There was an issue with the APIClient. HTTP error code %d", res.StatusCode)
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
	return resourceUserAccountRead(d, m)

}

func resourceUserAccountRead(d *schema.ResourceData, m interface{}) error {
	/*	userIDs := d.Get("users_id").(map[string]string)
		c := m.([]*gofish.APIClient)
		for _, v := range c {
			account, err := getAccount(v, userIDs[v.Service.ODataID])
			if err != nil {
				return err
			}
			if account == nil {
				d.SetId("")
				return nil
			}
			d.Set("username", account.UserName)
			d.Set("enabled", account.Enabled)
			d.Set("role_id", account.RoleID)
		}*/
	return nil
}

func resourceUserAccountUpdate(d *schema.ResourceData, m interface{}) error {
	c := m.([]*ClientConfig)
	for _, v := range c {
		account, err := getAccount(v.API, d.Get("users_id").(map[string]interface{})[v.Endpoint].(string))
		if err != nil {
			return err
		}
		payload := make(map[string]interface{})
		payload["UserName"] = d.Get("username")
		payload["Password"] = d.Get("password")
		payload["Enabled"] = d.Get("enabled")
		payload["RoleId"] = d.Get("role_id")
		res, err := v.API.Patch(account.ODataID, payload)
		if err != nil {
			return err
		}
		if res.StatusCode != 200 {
			return fmt.Errorf("There was an issue with the APIClient. HTTP error code %d", res.StatusCode)
		}
	}
	return resourceUserAccountRead(d, m)
}

func resourceUserAccountDelete(d *schema.ResourceData, m interface{}) error {
	c := m.([]*ClientConfig)
	for _, v := range c {
		account, err := getAccount(v.API, d.Get("users_id").(map[string]interface{})[v.Endpoint].(string))
		if err != nil {
			return err
		}
		if account == nil {
			return fmt.Errorf("The user account does not exist")
		}
		payload := make(map[string]interface{})
		payload["UserName"] = ""
		res, err := v.API.Patch(account.ODataID, payload)
		if err != nil {
			return err
		}
		if res.StatusCode != 200 {
			return fmt.Errorf("There was an issue with the APIClient. HTTP error code %d", res.StatusCode)
		}
	}
	d.SetId("")
	return nil

}

func getAccountList(c *gofish.APIClient) ([]*redfish.ManagerAccount, error) {
	service := c.Service
	accountService, err := service.AccountService()
	if err != nil {
		return nil, err
	}
	accounts, err := accountService.Accounts()
	if err != nil {
		return nil, err
	}
	return accounts, nil
}

func getAccount(c *gofish.APIClient, id string) (*redfish.ManagerAccount, error) {
	accountList, err := getAccountList(c)
	if err != nil {
		return nil, err
	}
	for _, account := range accountList {
		if account.ID == id && len(account.UserName) > 0 {
			return account, nil
		}
	}
	return nil, nil //This will be returned if there was no errors but the user does not exist
}
