package redfish

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/url"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/dell/terraform-provider-redfish/common"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/stmcginnis/gofish"
	redfishcommon "github.com/stmcginnis/gofish/common"
	"github.com/stmcginnis/gofish/redfish"
)

const (
	defaultBiosConfigServerResetTimeout int = 120
	defaultBiosConfigJobTimeout         int = 1200
	intervalBiosConfigJobCheckTime      int = 10
)

func resourceRedfishBios() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceRedfishBiosUpdate,
		ReadContext:   resourceRedfishBiosRead,
		UpdateContext: resourceRedfishBiosUpdate,
		DeleteContext: resourceRedfishBiosDelete,
		Schema:        getResourceRedfishBiosSchema(),
		CustomizeDiff: resourceRedfishBiosCustomizeDiff,
	}
}

func getResourceRedfishBiosSchema() map[string]*schema.Schema {
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
		"attributes": {
			Type:        schema.TypeMap,
			Optional:    true,
			Computed:    true,
			Description: "Bios attributes",
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
		"settings_apply_time": {
			Type:     schema.TypeString,
			Optional: true,
			Description: "The time when the BIOS settings can be applied. Applicable value is 'OnReset' only. " +
				"In upcoming releases other apply time values will be supported. Default is \"OnReset\".",
			ValidateFunc: validation.StringInSlice([]string{
				string(redfishcommon.OnResetApplyTime),
			}, false),
			Default: string(redfishcommon.OnResetApplyTime),
		},
		"reset_type": {
			Type:     schema.TypeString,
			Optional: true,
			Description: "Reset type to apply on the computer system after the BIOS settings are applied. " +
				"Applicable values are 'ForceRestart', " +
				"'GracefulRestart', and 'PowerCycle'." +
				"Default = \"GracefulRestart\". ",
			ValidateFunc: validation.StringInSlice([]string{
				string(redfish.ForceRestartResetType),
				string(redfish.GracefulRestartResetType),
				string(redfish.PowerCycleResetType),
			}, false),
			Default: string(redfish.GracefulRestartResetType),
		},
		"reset_timeout": {
			Type:        schema.TypeInt,
			Optional:    true,
			Description: "reset_timeout is the time in seconds that the provider waits for the server to be reset before timing out.",
			Default:     defaultBiosConfigServerResetTimeout,
		},
		"bios_job_timeout": {
			Type:        schema.TypeInt,
			Optional:    true,
			Description: "bios_job_timeout is the time in seconds that the provider waits for the bios update job to be completed before timing out.",
			Default:     defaultBiosConfigJobTimeout,
		},
	}
}

func resourceRedfishBiosCustomizeDiff(_ context.Context, diff *schema.ResourceDiff, v interface{}) error {
	if diff.Id() == "" {
		return nil
	}

	if diff.HasChange("attributes") {
		o, n := diff.GetChange("attributes")
		oldAttribs := o.(map[string]interface{})
		newAttribs := n.(map[string]interface{})

		sameAttribs := biosAttributesMatch(oldAttribs, newAttribs)

		if sameAttribs {
			log.Printf("[DEBUG] Bios attributes have not changed. clearing diff")
			if err := diff.Clear("attributes"); err != nil {
				return err
			}
		} else {
			// Update the attributes value pairs
			for k, v := range newAttribs {
				oldAttribs[k] = v
			}

			if err := diff.SetNew("attributes", oldAttribs); err != nil {
				return err
			}
		}
	}

	return nil
}

func resourceRedfishBiosRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	service, err := NewConfig(m.(*schema.ResourceData), d)
	if err != nil {
		return diag.Errorf(err.Error())
	}
	return readRedfishBiosResource(service, d)
}

func resourceRedfishBiosUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	service, err := NewConfig(m.(*schema.ResourceData), d)
	if err != nil {
		return diag.Errorf(err.Error())
	}
	return updateRedfishBiosResource(service, d)
}

func resourceRedfishBiosDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	d.SetId("")
	return diags
}

func updateRedfishBiosResource(service *gofish.Service, d *schema.ResourceData) diag.Diagnostics {
	log.Printf("[DEBUG] Beginning update")
	var diags diag.Diagnostics

	// Lock the mutex to avoid race conditions with other resources
	redfishMutexKV.Lock(getRedfishServerEndpoint(d))
	defer redfishMutexKV.Unlock(getRedfishServerEndpoint(d))

	resetType := d.Get("reset_type")

	bios, err := getBiosResource(service)
	if err != nil {
		return diag.Errorf("error fetching bios resource: %s", err)
	}

	attributes := make(map[string]string)
	err = copyBiosAttributes(bios, attributes)
	if err != nil {
		return diag.Errorf("error fetching bios attributes: %s", err)
	}

	attrsPayload, err := getBiosAttrsToPatch(d, attributes)
	if err != nil {
		return diag.Errorf("error getting BIOS attributes to patch: %s", err)
	}

	resetTimeout := d.Get("reset_timeout")

	biosConfigJobTimeout := d.Get("bios_job_timeout")

	log.Printf("[DEBUG] resetTimeout is set to %d  and Bios Config Job timeout is set to %d", resetTimeout.(int), biosConfigJobTimeout.(int))

	var biosTaskURI string
	if len(attrsPayload) != 0 {
		biosTaskURI, err = patchBiosAttributes(d, bios, attrsPayload)
		if err != nil {
			return diag.Errorf("error updating bios attributes: %s", err)
		}

		// reboot the server
		_, diags := PowerOperation(resetType.(string), resetTimeout.(int), intervalBiosConfigJobCheckTime, service)
		if diags.HasError() {
			// TODO: handle this scenario
			return diag.Errorf("there was an issue restarting the server")
		}

		// wait for the bios config job to finish
		err = common.WaitForJobToFinish(service, biosTaskURI, intervalBiosConfigJobCheckTime, biosConfigJobTimeout.(int))
		if err != nil {
			return diag.Errorf("Error waiting for Bios config monitor task (%s) to be completed: %s", biosTaskURI, err)
		}
		time.Sleep(30 * time.Second)
	} else {
		log.Printf("[DEBUG] BIOS attributes are already set")
	}

	if err = d.Set("attributes", attributes); err != nil {
		return diag.Errorf("error setting bios attributes: %s", err)
	}

	diags = readRedfishBiosResource(service, d)
	// Set the ID to @odata.id
	d.SetId(bios.ODataID)

	log.Printf("[DEBUG] %s: Update finished successfully", d.Id())
	return diags
}

func readRedfishBiosResource(service *gofish.Service, d *schema.ResourceData) diag.Diagnostics {

	log.Printf("[DEBUG] %s: Beginning read", d.Id())
	var diags diag.Diagnostics

	bios, err := getBiosResource(service)
	if err != nil {
		return diag.Errorf("error fetching BIOS resource: %s", err)
	}

	attributes := make(map[string]string)
	err = copyBiosAttributes(bios, attributes)
	if err != nil {
		return diag.Errorf("error fetching BIOS attributes: %s", err)
	}

	if err := d.Set("attributes", attributes); err != nil {
		return diag.Errorf("error setting bios attributes: %s", err)
	}

	log.Printf("[DEBUG] %s: Read finished successfully", d.Id())

	return diags
}

func copyBiosAttributes(bios *redfish.Bios, attributes map[string]string) error {

	// TODO: BIOS Attributes' values might be any of several types.
	// terraform-sdk currently does not support a map with different
	// value types. So we will convert int and float values to string.
	// copy from the BIOS attributes to the new bios attributes map
	for key, value := range bios.Attributes {
		if attrVal, ok := value.(string); ok {
			attributes[key] = attrVal
		} else {
			attributes[key] = fmt.Sprintf("%v", value)
		}
	}

	return nil
}

func patchBiosAttributes(d *schema.ResourceData, bios *redfish.Bios, attributes map[string]interface{}) (biosTaskURI string, err error) {

	payload := make(map[string]interface{})
	payload["Attributes"] = attributes

	if settingsApplyTime, ok := d.GetOk("settings_apply_time"); ok {
		allowedValues := bios.AllowedAttributeUpdateApplyTimes()
		allowed := false
		for i := range allowedValues {
			if strings.TrimSpace(settingsApplyTime.(string)) == (string)(allowedValues[i]) {
				allowed = true
				break
			}
		}

		if !allowed {
			errTxt := fmt.Sprintf("\"%s\" is not allowed as settings apply time", settingsApplyTime)
			err := errors.New(errTxt)
			return "", err
		}

		payload["@Redfish.SettingsApplyTime"] = map[string]interface{}{
			"ApplyTime": settingsApplyTime.(string),
		}
	}

	oDataURI, err := url.Parse(bios.ODataID)
	if err != nil {
		log.Printf("[DEBUG] error parsing odata id: %s", err)
		return "", err
	}

	oDataURI.Path = path.Join(oDataURI.Path, "Settings")
	settingsObjectURI := oDataURI.String()

	resp, err := bios.GetClient().Patch(settingsObjectURI, payload)
	if err != nil {
		log.Printf("[DEBUG] error sending the patch request: %s", err)
		return "", err
	}

	// check if location is present in the response header
	if location, err := resp.Location(); err == nil {
		log.Printf("[DEBUG] BIOS configuration job uri: %s", location.String())

		taskURI := location.EscapedPath()
		return taskURI, nil

	}

	return "", nil
}

func getBiosResource(service *gofish.Service) (*redfish.Bios, error) {

	system, err := getSystemResource(service)
	if err != nil {
		log.Printf("[ERROR]: Failed to get system resource: %s", err)
		return nil, err
	}

	bios, err := system.Bios()
	if err != nil {
		log.Printf("[ERROR]: Failed to get Bios resource: %s", err)
		return nil, err
	}

	return bios, nil
}

func getBiosAttrsToPatch(d *schema.ResourceData, attributes map[string]string) (map[string]interface{}, error) {

	attrs := make(map[string]interface{})
	attrsToPatch := make(map[string]interface{})

	if v, ok := d.GetOk("attributes"); ok {
		attrs = v.(map[string]interface{})
	}

	for key, newVal := range attrs {
		if oldVal, ok := attributes[key]; ok {
			// check if the original value is an integer
			// if yes, then we need to convert accordingly
			if intOldVal, err := strconv.Atoi(attributes[key]); err == nil {
				intNewVal, err := strconv.Atoi(newVal.(string))
				if err != nil {
					return attrsToPatch, err
				}

				// Add to patch list if attribute value has changed
				if intNewVal != intOldVal {
					attrsToPatch[key] = intNewVal
				}
			} else {
				if newVal != oldVal {
					attrsToPatch[key] = newVal
				}
			}

		} else {
			err := fmt.Errorf("BIOS attribute %s not found", key)
			return attrsToPatch, err
		}
	}

	return attrsToPatch, nil
}

func biosAttributesMatch(oldAttribs, newAttribs map[string]interface{}) bool {
	log.Printf("[DEBUG] Begin biosAttributesMatch")

	for key, newVal := range newAttribs {
		log.Printf("[DEBUG] attribute: %v, newVal: %v", key, newVal)
		if oldVal, ok := oldAttribs[key]; ok {
			log.Printf("[DEBUG] found attribute: %v, oldVal: %v", key, oldVal)
			// check if the original value is an integer
			// if yes, then we need to convert accordingly
			if intOldVal, err := strconv.Atoi(oldVal.(string)); err == nil {
				intNewVal, err := strconv.Atoi(newVal.(string))
				if err != nil {
					return false
				}

				if intNewVal != intOldVal {
					return false
				}
			} else {
				if newVal != oldVal {
					return false
				}
			}
		} else {
			// attribute not found in the current state
			log.Printf("[DEBUG] attribute %v not found in the current state", key)
			return false
		}
	}
	return true
}
