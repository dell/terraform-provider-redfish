package redfish

import (
	"context"
	// "fmt"
	// "github.com/dell/terraform-provider-redfish/common"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stmcginnis/gofish"
	"github.com/stmcginnis/gofish/redfish"
	// "log"
)

func resourceUserAccount() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceUserAccountCreate,
		ReadContext:   resourceUserAccountRead,
		UpdateContext: resourceUserAccountUpdate,
		DeleteContext: resourceUserAccountDelete,
		Schema: map[string]*schema.Schema{
			"redfish_server": {
				Type:        schema.TypeList,
				Required:    true,
				Description: "This list contains the different redfish endpoints to manage (different servers)",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"user": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "This field is the user to login against the redfish API",
						},
						"password": {
							Type:        schema.TypeString,
							Required:    true,
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
		},
	}
}

func resourceUserAccountCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	//Connect clients
	service, err := NewConfig(d)
	if err != nil {
		return diag.Errorf(err.Error())
	}

	accountList, err := getAccountList(service)
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
			res, err := service.Client.Patch(account.ODataID, payload)
			if err != nil {
				return diag.Errorf("Error when contacting the redfish API %v", err) //This error might happen when a user was created outside terraform
			}
			if res.StatusCode != 200 {
				return diag.Errorf("There was an issue with the APIClient. HTTP error code %d", res.StatusCode)
			}
			//Set ID to terraform state file
			d.SetId(account.ID)
			return diags
		}
	}
	//No room for new users
	return diag.Errorf("There are no room for new users")
}

func resourceUserAccountRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	//Connect clients
	service, err := NewConfig(d)
	if err != nil {
		return diag.Errorf(err.Error())
	}

	accountList, err := getAccountList(service)
	if err != nil {
		return diag.Errorf("Error when retrieving account list %v", err)
	}

	account, err := getAccount(accountList, d.Id())
	if err != nil {
		return diag.Errorf("Error when retrieving accounts %v", err)
	}
	if account == nil { //User doesn't exist. Needs to be recreated.
		d.Set("username", "")
		d.Set("enabled", "")
		d.Set("role_id", "")
		return diags
	}
	if d.Get("username") != account.UserName || d.Get("enabled") != account.Enabled || d.Get("role_id") != account.RoleID {
		// If something is different an update needs to be triggered
		d.Set("username", account.UserName)
		d.Set("enabled", account.Enabled)
		d.Set("role_id", account.RoleID)
		return diags
	}

	return diags
}

func resourceUserAccountUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	return resourceUserAccountRead(ctx, d, m)
}

func resourceUserAccountDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
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
