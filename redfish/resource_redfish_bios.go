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

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/stmcginnis/gofish"
	"github.com/stmcginnis/gofish/common"
	"github.com/stmcginnis/gofish/redfish"
)

func resourceRedfishBios() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceRedfishBiosUpdate,
		ReadContext:   resourceRedfishBiosRead,
		UpdateContext: resourceRedfishBiosUpdate,
		DeleteContext: resourceRedfishBiosDelete,
		Schema:        getResourceRedfishBiosSchema(),
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
			Description: "Bios attributes",
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
		"settings_apply_time": {
			Type:     schema.TypeString,
			Optional: true,
			Description: "The time when the BIOS settings can be applied. Applicable values are " +
				"'OnReset', 'Immediate', 'AtMaintenanceWindowStart' and 'InMaintenanceWindowStart'. " +
				"Default is \"\" which will not create a BIOS configuration job.",
			ValidateFunc: validation.StringInSlice([]string{
				string(common.ImmediateApplyTime),
				string(common.OnResetApplyTime),
				string(common.AtMaintenanceWindowStartApplyTime),
				string(common.InMaintenanceWindowOnResetApplyTime),
			}, false),
		},
		"action_after_apply": {
			Type:     schema.TypeString,
			Optional: true,
			Description: "Action to perform on the target after the BIOS settings are applied. " +
				"Default=nil : no action after apply" +
				"Applicable values are nil, 'On','ForceOn','ForceOff','ForceRestart','GracefulRestart'," +
				"'GracefulShutdown','PushPowerButton','PowerCycle','Nmi'.",
			ValidateFunc: validation.StringInSlice([]string{
				string(redfish.OnResetType),
				string(redfish.ForceOnResetType),
				string(redfish.ForceOffResetType),
				string(redfish.ForceRestartResetType),
				string(redfish.GracefulRestartResetType),
				string(redfish.GracefulShutdownResetType),
				string(redfish.PushPowerButtonResetType),
				string(redfish.PowerCycleResetType),
			}, false),
		},
		"task_monitor_uri": {
			Type:        schema.TypeString,
			Description: "URI of the BIOS configuration task monitor",
			Computed:    true,
		},
	}
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

	// check if there is already a bios config job in progress
	// if yes, then check the current status of the job. If it
	// has not completed yet, then don't perform another operation
	pending := false
	var taskURI string
	if v, ok := d.GetOk("task_monitor_uri"); ok {
		log.Printf("[DEBUG] %s: Bios config task monitor uri is \"%s\"", d.Id(), v.(string))
		taskURI, _ = v.(string)
		if len(taskURI) > 0 {
			task, _ := redfish.GetTask(service.Client, taskURI)
			if task != nil {
				if task.TaskState != redfish.CompletedTaskState {
					log.Printf("[DEBUG] %s: BIOS config task state = %s", d.Id(), task.TaskState)
					pending = true
				}
			} else {
				// Task does not exist or there was an error
				if err := d.Set("task_monitor_uri", ""); err != nil {
					return diag.Errorf("error updating task_monitor_uri: %s", err)
				}
			}
		}
	}

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

	if len(attrsPayload) != 0 {
		if !pending {
			err = patchBiosAttributes(d, bios, attrsPayload)
			if err != nil {
				return diag.Errorf("error updating bios attributes: %s", err)
			}
		} else {
			log.Printf("[DEBUG] Not updating the attributes as a previous BIOS job is already scheduled")
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Warning,
				Summary:  "Unable to update bios attributes",
				Detail:   "Unable to update bios attributes as a previous BIOS job is already scheduled.",
			})
		}
	} else {
		log.Printf("[DEBUG] BIOS attributes are already set")
	}
	if err := d.Set("attributes", attributes); err != nil {
		return diag.Errorf("error setting bios attributes: %s", err)
	}
	// Set the ID to @odata.id
	d.SetId(bios.ODataID)
	actionAfterApply, ok := d.GetOk("action_after_apply")

	if ok && actionAfterApply != nil {
		resetSystem(service, d, (redfish.ResetType)(actionAfterApply.(string)))

		if v, ok := d.GetOk("task_monitor_uri"); ok {
			log.Printf("[DEBUG] %s: Bios config task monitor uri is \"%s\"", d.Id(), v.(string))
			taskURI, _ = v.(string)
			if len(taskURI) > 0 {
				createStateConf := &resource.StateChangeConf{
					Pending: []string{
						string(redfish.NewTaskState),
						string(redfish.StartingTaskState),
						string(redfish.RunningTaskState),
						string(redfish.PendingTaskState),
						string(redfish.StoppingTaskState),
						string(redfish.CancellingTaskState),
					},
					Target: []string{
						string(redfish.CompletedTaskState),
					},
					Refresh: func() (interface{}, string, error) {
						resp, err := redfish.GetTask(service.Client, taskURI)
						if err != nil {
							return 0, "", err
						}
						return resp, string(resp.TaskState), nil
					},
					Timeout:    d.Timeout(schema.TimeoutCreate),
					Delay:      10 * time.Second,
					MinTimeout: 5 * time.Second,
					ContinuousTargetOccurence: 5,
				}
				_, err = createStateConf.WaitForState()
				if err != nil {
					return diag.Errorf("Error waiting for Bios config monitor task (%s) to be completed: %s", d.Id(), err)
				}
			}
		}
	}


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

func patchBiosAttributes(d *schema.ResourceData, bios *redfish.Bios, attributes map[string]interface{}) error {

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
			return err
		}

		payload["@Redfish.SettingsApplyTime"] = map[string]interface{}{
			"ApplyTime": settingsApplyTime.(string),
		}
	}

	oDataURI, err := url.Parse(bios.ODataID)
	oDataURI.Path = path.Join(oDataURI.Path, "Settings")
	settingsObjectURI := oDataURI.String()

	resp, err := bios.Client.Patch(settingsObjectURI, payload)
	if err != nil {
		log.Printf("[DEBUG] error sending the patch request: %s", err)
		return err
	}

	// check if location is present in the response header
	if location, err := resp.Location(); err == nil {
		log.Printf("[DEBUG] BIOS configuration job uri: %s", location.String())

		taskURI := location.EscapedPath()

		if err = d.Set("task_monitor_uri", taskURI); err != nil {
			log.Printf("[DEBUG] error setting the task uri: %s", err)
			return err
		}
	}

	return nil
}

func resetSystem(service *gofish.Service, d *schema.ResourceData, resetType redfish.ResetType) error {

	system, err := getSystemResource(service)
	if err != nil {
		log.Printf("[ERROR]: Failed to identify system: %s", err)
		return err
	}

	if system.PowerState == redfish.OffPowerState {
		log.Printf("[WARN]: will not perform reset because system is Off.  Bios changes will be applied at next boot.")
		return nil
	}

	log.Printf("[TRACE]: Performing system.Reset(%s)", resetType)
	if err = system.Reset(resetType); err != nil {
		log.Printf("[WARN]: system.Reset returned an error: %s", err)
		return err
	}

	log.Printf("[TRACE]: system.Reset successful")
	return err
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
