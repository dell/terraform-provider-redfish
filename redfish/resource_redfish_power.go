package redfish

import (
	"context"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceRedFishPower() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceRedfishPowerUpdate,
		ReadContext:   resourceRedfishPowerRead,
		UpdateContext: resourceRedfishPowerUpdate,
		DeleteContext: resourceRedfishPowerDelete,
		Schema:        getResourceRedfishPowerSchema(),
		CustomizeDiff: CheckPowerDiff(),
	}
}

// Custom function for calculating power state. Given some desired_power_action, we know what the expected power
// state should be after the action is applied. Instead of marking the value of power_state as unknown during
// PlanResourceChange, we can calculate exactly what the end power_state should be.
func CheckPowerDiff(funcs ...schema.CustomizeDiffFunc) schema.CustomizeDiffFunc {
	return func(ctx context.Context, d *schema.ResourceDiff, meta interface{}) error {

		resetType, ok := d.GetOk("desired_power_action")

		if !ok || resetType == nil {
			log.Printf("[ERROR]: There was a problem getting the desired_power_action")
			return nil
		}

		if resetType == "ForceOff" || resetType == "GracefulShutdown" {
			d.SetNew("power_state", "Off")
		} else if resetType == "ForceOn" || resetType == "On" {
			d.SetNew("power_state", "On")
		} else if resetType == "ForceRestart" || resetType == "PowerCycle" {
			d.SetNew("power_state", "Reset_On")
		}
		// Note - if they select PushPowerButton then this function does nothing because we don't know what the value
		// will be. We just let Terraform set everything to unknown as per normal

		return nil
	}
}

func getResourceRedfishPowerSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"redfish_server": {
			Type:        schema.TypeList,
			Required:    true,
			Description: "List of server BMCs and their respective user credentials",
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"user": {
						Type:        schema.TypeString,
						Optional:    true,
						Description: "User name for login",
					},
					"password": {
						Type:        schema.TypeString,
						Optional:    true,
						Description: "User password for login",
						Sensitive:   true,
					},
					"endpoint": {
						Type:        schema.TypeString,
						Required:    true,
						Description: "Server BMC IP address or hostname",
					},
					"ssl_insecure": {
						Type:        schema.TypeBool,
						Optional:    true,
						Description: "This field indicates whether the SSL/TLS certificate must be verified or not",
					},
				},
			},
		},
		"desired_power_action": {
			Type:     schema.TypeString,
			Required: true,
			Description: "Desired power setting. Applicable values 'On','ForceOn','ForceOff','ForceRestart'," +
				"'GracefulRestart','GracefulShutdown','PowerCycle'",
		},
		"maximum_wait_time": {
			Type:     schema.TypeInt,
			Required: true,
			Description: "The maximum amount of time to wait for the server to enter the correct power state before" +
				"giving up in seconds",
		},
		"check_interval": {
			Type:        schema.TypeInt,
			Required:    true,
			Description: "The frequency with which to check the server's power state in seconds",
		},
		"power_state": {
			Type:     schema.TypeString,
			Computed: true,
			Description: "Desired power setting. Applicable values 'On','ForceOn','ForceOff','ForceRestart'," +
				"'GracefulRestart','GracefulShutdown','PowerCycle'.",
		},
	}
}

func resourceRedfishPowerRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

	log.Printf("[DEBUG] %s: Beginning read", d.Id())
	var diags diag.Diagnostics

	service, err := NewConfig(m.(*schema.ResourceData), d)
	if err != nil {
		return diag.Errorf(err.Error())
	}

	system, err := getSystemResource(service)
	if err != nil {
		log.Printf("[ERROR]: Failed to identify system: %s", err)
		return diag.Errorf(err.Error())
	}

	if err := d.Set("power_state", system.PowerState); err != nil {
		return diag.Errorf("[ERROR]: Could not retrieve system power state. %s", err)
	}

	resetType, ok := d.GetOk("desired_power_action")

	if !ok || resetType == nil {
		log.Printf("[ERROR]: ")
		return diags
	}

	if err := d.Set("desired_power_action", resetType); err != nil {
		return diag.Errorf("[ERROR]: Could not retrieve system power state. %s", err)
	}

	return diags
}

func resourceRedfishPowerUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

	var diags diag.Diagnostics

	// Lock the mutex to avoid race conditions with other resources
	redfishMutexKV.Lock(getRedfishServerEndpoint(d))
	defer redfishMutexKV.Unlock(getRedfishServerEndpoint(d))

	resetType, ok := d.GetOk("desired_power_action")

	if !ok || resetType == nil {
		log.Printf("[ERROR]: There was a problem getting the desired_power_action")
		return diags
	}

	// Takes the m interface and feeds it the user input data d. You can then reference it with X.GetOk("user")
	service, err := NewConfig(m.(*schema.ResourceData), d)

	if err != nil {
		return diag.Errorf(err.Error())
	}

	system, err := getSystemResource(service)
	if err != nil {
		log.Printf("[ERROR]: Failed to identify system: %s", err)
		return diag.Errorf(err.Error())
	}

	d.SetId(system.SerialNumber + "_power")

	powerState, diags := PowerOperation(resetType.(string), d.Get("maximum_wait_time").(int), d.Get("check_interval").(int), service)

	if (resetType == "ForceRestart" || resetType == "GracefulRestart" || resetType == "PowerCycle") && powerState == "On" {
		powerState = "Reset_On"
	}

	d.Set("power_state", powerState)

	return diags
}

func resourceRedfishPowerDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

	var diags diag.Diagnostics

	d.SetId("")

	return diags
}
