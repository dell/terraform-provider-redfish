package redfish

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/stmcginnis/gofish"
)

func resourceUser() *schema.Resource {
	return &schema.Resource{
		Create: resourceUserCreate,
		Read:   resourceUserRead,
		Update: resourceUserUpdate,
		Delete: resourceUserDelete,

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
		},
	}
}

func resourceUserCreate(d *schema.ResourceData, m interface{}) error {
	c := m.(*gofish.APIClient)
	//Retreive the service root
	service := c.Service

	accountService, err := service.AccountService()
	if err != nil {
		return err
	}
	//Get list of accounts
	accounts, err := accountService.Accounts()
	if err != nil {
		return err
	}
	payload := make(map[string]string)
	for _, account := range accounts {
		if len(account.UserName) == 0 && account.ID != "1" { //ID 1 is reserved
			payload["UserName"] = d.Get("username").(string)
			payload["Password"] = d.Get("password").(string)
			//We should include more params
			res, err := c.Patch(account.ODataID, payload)
			if err != nil {
				return err
			}
			if res.StatusCode == 200 {
				d.SetId(account.ID)
				return resourceUserRead(d, m)
			} else {
				return fmt.Errorf("There was an issue with the APIClient. HTTP error code %d", res.StatusCode)
			}
		}
	}
	//No room for new users
	return fmt.Errorf("There are no room for new users")
}

func resourceUserRead(d *schema.ResourceData, m interface{}) error {
	return nil
}

func resourceUserUpdate(d *schema.ResourceData, m interface{}) error {
	return resourceUserRead(d, m)
}

func resourceUserDelete(d *schema.ResourceData, m interface{}) error {
	c := m.(*gofish.APIClient)
	//Retreive the service root
	service := c.Service

	accountService, err := service.AccountService()
	if err != nil {
		return err
	}
	//Get list of accounts
	accounts, err := accountService.Accounts()
	if err != nil {
		return err
	}
	payload := make(map[string]string)
	for _, account := range accounts {
		if account.ID == d.Id() {
			payload["UserName"] = ""
			//payload["Password"] = ""
			//We should include more params
			res, err := c.Patch(account.ODataID, payload)
			if err != nil {
				return err
			}
			if res.StatusCode == 200 {
				d.SetId("")
				return nil
			}
			return fmt.Errorf("There was an issue with the APIClient. HTTP error code %d", res.StatusCode)
		}
	}
	return fmt.Errorf("No user to remove")
}
