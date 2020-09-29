package redfish

import (
	"context"
	"errors"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/stmcginnis/gofish"
	"github.com/stmcginnis/gofish/common"
	"github.com/stmcginnis/gofish/redfish"
	"log"
	"strconv"
	"strings"
)

func resourceRedfishBios() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceRedfishBiosUpdate,
		ReadContext:   resourceRedfishBiosRead,
		UpdateContext: resourceRedfishBiosUpdate,
		DeleteContext: resourceRedfishBiosDelete,
		Schema: map[string]*schema.Schema{
			"attributes": {
				Type:        schema.TypeMap,
				Optional:    true,
				Description: "Bios attributes",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},

			"settings_apply_time": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The time when the BIOS settings can be applied. Applicable values are 'OnReset', 'Immediate', 'AtMaintenanceWindowStart' and 'InMaintenanceWindowStart'.",
				ValidateFunc: validation.StringInSlice([]string{
					string(common.ImmediateApplyTime),
					string(common.OnResetApplyTime),
					string(common.AtMaintenanceWindowStartApplyTime),
					string(common.InMaintenanceWindowOnResetApplyTime),
				}, false),
			},

			"bios_config_job_uri": {
				Type:        schema.TypeString,
				Description: "BIOS configuration job uri",
				Computed:    true,
			},
		},
	}
}

func resourceRedfishBiosUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

	log.Printf("[DEBUG] Beginning update")
	var diags diag.Diagnostics

	conn := m.(*gofish.APIClient)

	// check if there is already a bios config job in progress
	// if yes, then check the current status of the job. If it
	// has not completed yet, then don't perform another operation
	pending := false
	if v, ok := d.GetOk("bios_config_job_uri"); ok {
		log.Printf("[DEBUG] %s: Bios config job uri is \"%s\"", d.Id(), v.(string))
		taskUri, _ := v.(string)
		if len(taskUri) > 0 {
			task, _ := redfish.GetTask(conn, taskUri)
			if task != nil {
				if task.TaskState != redfish.CompletedTaskState {
					log.Println("[DEBUG] %s: BIOS config task state = %s", d.Id(), task.TaskState)
					pending = true
				}
			} else {
				// Task does not exist or there was an error
				if err := d.Set("bios_config_job_uri", ""); err != nil {
					return diag.Errorf("error updating bios_config_job_uri: %s", err)
				}
			}
		}
	}

	bios, err := getBios(conn)
	if err != nil {
		return diag.Errorf("error fetching bios resource: %s", err)
	}

	attributes := make(map[string]string)
	err = copyBiosAttributes(bios, attributes)
	if err != nil {
		return diag.Errorf("error fetching bios attributes: %s", err)
	}

	attrsToPatch := make(map[string]interface{})
	if v, ok := d.GetOk("attributes"); ok {
		attrsToPatch = v.(map[string]interface{})
	}

	attrsPayload := make(map[string]interface{})

	for key, val := range attrsToPatch {
		if oldVal, ok := attributes[key]; ok {
			// check if the original value is an integer
			// if yes, then we need to convert accordingly
			if intOldVal, err := strconv.Atoi(attributes[key]); err == nil {
				intVal, err := strconv.Atoi(val.(string))
				if err != nil {
					return diag.Errorf("Failed typecast to int for bios attribute: %s", key)
				}

				// Add to payload if attribute value has changed
				if intVal != intOldVal {
					attrsPayload[key] = intVal
				}
			} else {
				if val != oldVal {
					attrsPayload[key] = val
				}
			}

		} else {
			return diag.Errorf("BIOS attribute %s not found", key)
		}
	}

	if len(attrsPayload) != 0 {
		if !pending {
			err = updateBiosAttributes(d, bios, attrsPayload)
			if err != nil {
				return diag.Errorf("error updating bios attributes: %s", err)
			}
		} else {
			log.Printf("[DEBUG] Not updating the attributes as a previous BIOS job is pending")
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Warning,
				Summary: "Unable to update bios attributes",
				Detail: "Unable to update bios attributes as a previous BIOS job is pending",
			})
		}
	} else {
		log.Printf("[DEBUG] BIOS attributes are already set")
	}

	if err := d.Set("attributes", attributes); err != nil {
		return diag.Errorf("error setting bios attributes: %s", err)
	}

	// Set the ID to the @odata.id
	d.SetId(bios.ODataID)

	log.Printf("[DEBUG] %s: Update finished successfully", d.Id())
	return diags
}

func resourceRedfishBiosRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

	log.Printf("[DEBUG] %s: Beginning read", d.Id())
	var diags diag.Diagnostics

	conn := m.(*gofish.APIClient)

	bios, err := getBios(conn)
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

	// Set the ID to the @odata.id
	d.SetId(bios.ODataID)

	log.Printf("[DEBUG] %s: Read finished successfully", d.Id())

	return diags
}

func resourceRedfishBiosDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

	var diags diag.Diagnostics

	d.SetId("")

	return diags
}

func getBios(conn *gofish.APIClient) (*redfish.Bios, error) {

	service := conn.Service
	systems, err := service.Systems()
	if err != nil {
		return nil, err
	}

	bios, err := systems[0].Bios()
	if err != nil {
		return nil, err
	}

	return bios, nil
}

func copyBiosAttributes(bios *redfish.Bios, attributes map[string]string) error {

	// TODO: BIOS Attributes' values might be any of several types.
	// terraform-sdk currently does not support a map with different
	// value types. So we will convert int and float values to string.
	// copy from the BIOS attributes to the new bios attributes map
	for key, value := range bios.Attributes {
		if attr_val, ok := value.(string); ok {
			attributes[key] = attr_val
		} else {
			attributes[key] = fmt.Sprintf("%v", value)
		}
	}

	return nil
}

func updateBiosAttributes(d *schema.ResourceData, bios *redfish.Bios, attributes map[string]interface{}) error {

	payload := make(map[string]interface{})
	payload["Attributes"] = attributes

	if settingsApplyTime, ok := d.GetOk("settings_apply_time"); ok {
		allowedValues := bios.AllowedAttributeUpdateApplyTimes()
		allowed := false
		for i := range allowedValues {
			if strings.TrimSpace(settingsApplyTime.(string)) == (string)(allowedValues[i]) {
				allowed = true
			}
		}

		if !allowed {
			errTxt := fmt.Sprintf("\"%s\" is not allowed as settings apply time", settingsApplyTime)
			err := errors.New(errTxt)
			return err
		}

		payload["@Redfish.SettingsApplyTime"] = map[string]interface{}{
			"ApplyTime": settingsApplyTime.(string),
		}
	}

	settingsObjectURI := bios.ODataID + "/Settings"

	resp, err := bios.Client.Patch(settingsObjectURI, payload)
	if err != nil {
		log.Println("[DEBUG] error sending the patch request: %s", err)
		return err
	}

	// check if location is present in the response header
	if location, err := resp.Location(); err == nil {
		log.Printf("[DEBUG] BIOS configuration job uri:", location.String())

		taskUri := location.EscapedPath()

		if err = d.Set("bios_config_job_uri", taskUri); err != nil {
			log.Printf("[DEBUG] error setting the task uri: %s", err)
			return err
		}
	}

	return nil
}
