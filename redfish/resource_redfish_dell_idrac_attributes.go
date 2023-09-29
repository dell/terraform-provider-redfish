package redfish

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/dell/terraform-provider-redfish/gofish/dell"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stmcginnis/gofish"
)

func resourceRedfishDellIdracAttributes() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceRedfishDellIdracAttributesCreate,
		ReadContext:   resourceRedfishDellIdracAttributesRead,
		UpdateContext: resourceRedfishDellIdracAttributesUpdate,
		DeleteContext: resourceRedfishDellIdracAttributesDelete,
		Schema:        getResourceRedfishDellIdracAttributesSchema(),
	}
}

func getResourceRedfishDellIdracAttributesSchema() map[string]*schema.Schema {
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
			Type:     schema.TypeMap,
			Required: true,
			Description: "iDRAC attributes. To check allowed attributes please either use the datasource for dell idrac attributes or query " +
				"/redfish/v1/Managers/iDRAC.Embedded.1/Oem/Dell/DellAttributes/iDRAC.Embedded.1. To get allowed values for those attributes, check " +
				"/redfish/v1/Registries/ManagerAttributeRegistry/ManagerAttributeRegistry.v1_0_0.json from a Redfish Instance",
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
	}
}

func resourceRedfishDellIdracAttributesCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	service, err := NewConfig(m.(*schema.ResourceData), d)
	if err != nil {
		return diag.Errorf(err.Error())
	}
	return updateRedfishDellIdracAttributes(ctx, service, d, m)
}

func resourceRedfishDellIdracAttributesRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	service, err := NewConfig(m.(*schema.ResourceData), d)
	if err != nil {
		return diag.Errorf(err.Error())
	}
	return readRedfishDellIdracAttributes(ctx, service, d, m)
}

func resourceRedfishDellIdracAttributesUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	service, err := NewConfig(m.(*schema.ResourceData), d)
	if err != nil {
		return diag.Errorf(err.Error())
	}
	if diags := updateRedfishDellIdracAttributes(ctx, service, d, m); diags.HasError() {
		return diags
	}
	return resourceRedfishDellIdracAttributesRead(ctx, d, m)
}

func resourceRedfishDellIdracAttributesDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	service, err := NewConfig(m.(*schema.ResourceData), d)
	if err != nil {
		return diag.Errorf(err.Error())
	}
	return deleteRedfishDellIdracAttributes(ctx, service, d, m)
}

func updateRedfishDellIdracAttributes(ctx context.Context, service *gofish.Service, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	// Get attributes
	attributesTf := d.Get("attributes").(map[string]interface{})

	// get managerAttributeRegistry to check parameters before posting them to redfish
	managerAttributeRegistry, err := getManagerAttributeRegistry(service)
	if err != nil {
		return diag.Errorf("there was an issue when creating/updating idrac attributes - %s", err)
	}

	// Set right attributes to patch (values from map are all string. It needs int and string)
	attributesToPatch, err := setManagerAttributesRightType(attributesTf, managerAttributeRegistry)
	if err != nil {
		return diag.Errorf("there was an issue when creating/updating idrac attributes - %s", err)
	}

	// Check that all attributes passed are compliant with the API
	err = checkManagerAttributes(managerAttributeRegistry, attributesToPatch)
	if err != nil {
		return diag.Errorf("there was an issue when creating/updating idrac attributes - %s", err)
	}

	// get managers (Dell servers have only the iDRAC)
	managers, err := service.Managers()
	if err != nil {
		return diag.Errorf("there was an issue when creating/updating idrac attributes - %s", err)
	}

	// Get OEM
	dellManager, err := dell.DellManager(managers[0])
	if err != nil {
		return diag.Errorf("there was an issue when creating/updating idrac attributes - %s", err)
	}

	// Get Dell attributes
	dellAttributes, err := dellManager.DellAttributes()
	if err != nil {
		return diag.Errorf("there was an issue when creating/updating idrac attributes - %s", err)
	}
	idracAttributes, err := getIdracAttributes(dellAttributes)
	if err != nil {
		return diag.Errorf("there was an issue when creating/updating idrac attributes - %s", err)
	}

	// Set the body to send
	patchBody := struct {
		ApplyTime  string `json:"@Redfish.OperationApplyTime"`
		Attributes map[string]interface{}
	}{
		ApplyTime:  "Immediate",
		Attributes: attributesToPatch,
	}

	response, err := service.GetClient().Patch(idracAttributes.ODataID, patchBody)
	if err != nil {
		return diag.Errorf("there was an issue when creating/updating idrac attributes - %s", err)
	}
	response.Body.Close() //#nosec G104 -- TBD

	d.SetId(idracAttributes.ODataID)
	readRedfishDellIdracAttributes(ctx, service, d, m)

	return diags
}

func readRedfishDellIdracAttributes(ctx context.Context, service *gofish.Service, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	// get managers (Dell servers have only the iDRAC)
	managers, err := service.Managers()
	if err != nil {
		return diag.Errorf("there was an issue when reading idrac attributes - %s", err)
	}

	// Get OEM
	dellManager, err := dell.DellManager(managers[0])
	if err != nil {
		return diag.Errorf("there was an issue when reading idrac attributes - %s", err)
	}

	// Get Dell attributes
	dellAttributes, err := dellManager.DellAttributes()
	if err != nil {
		return diag.Errorf("there was an issue when reading idrac attributes - %s", err)
	}
	idracAttributes, err := getIdracAttributes(dellAttributes)
	if err != nil {
		return diag.Errorf("there was an issue when reading idrac attributes - %s", err)
	}

	// Get config attributes
	old := d.Get("attributes")
	oldAttr := old.(map[string]interface{})
	readAttributes := make(map[string]string)

	for k, v := range oldAttr {
		attrValue := idracAttributes.Attributes[k] // Check if attribute from config exists in idrac attributes
		if attrValue != nil {                      // This is done to avoid triggering an update when reading Password values, that are shown as null (nil to Go)
			readAttributes[k] = fmt.Sprintf("%v", attrValue)
		} else {
			readAttributes[k] = v.(string)
		}
	}

	err = d.Set("attributes", readAttributes)
	if err != nil {
		return diag.Errorf("there was an issue when setting read attributes - %s", err)
	}

	return diags
}

func deleteRedfishDellIdracAttributes(ctx context.Context, service *gofish.Service, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	d.SetId("")

	return diags
}

func getManagerAttributeRegistry(service *gofish.Service) (*dell.ManagerAttributeRegistry, error) {
	registries, err := service.Registries()
	if err != nil {
		return nil, err
	}

	for _, r := range registries {
		if r.ID == "ManagerAttributeRegistry" {
			// Get ManagerAttributesRegistry
			managerAttrRegistry, err := dell.GetDellManagerAttributeRegistry(service.GetClient(), r.Location[0].URI)
			if err != nil {
				return nil, err
			}
			return managerAttrRegistry, nil
		}
	}

	return nil, fmt.Errorf("error. Couldn't retrieve ManagerAttributeRegistry")
}

func getIdracAttributes(attributes []*dell.DellAttributes) (*dell.DellAttributes, error) {
	for _, a := range attributes {
		if strings.Contains(a.ID, "iDRAC") {
			return a, nil
		}
	}
	return nil, fmt.Errorf("couldn't find iDRACAttributes")
}

func checkManagerAttributes(attrRegistry *dell.ManagerAttributeRegistry, attributes map[string]interface{}) error {
	var errors string // Here will be collected all attribute errors to show to users

	for k, v := range attributes {
		err := attrRegistry.CheckAttribute(k, v)
		if err != nil {
			errors += fmt.Sprintf("%s - %s\n", k, err.Error())
		}
	}
	if len(errors) > 0 {
		return fmt.Errorf(errors)
	}

	return nil
}

// setManagerAttributesRightType gets a map[string]interface{} from terraform, where all keys are strings,
// and returns a map[string]interface{} where values are either string or ints, and can be used for PATCH
func setManagerAttributesRightType(rawAttributes map[string]interface{}, registry *dell.ManagerAttributeRegistry) (map[string]interface{}, error) {
	patchMap := make(map[string]interface{})

	for k, v := range rawAttributes {
		attrType, err := registry.GetAttributeType(k)
		if err != nil {
			return nil, err
		}

		switch attrType {
		case "int":
			t, err := strconv.Atoi(v.(string))
			if err != nil {
				return nil, fmt.Errorf("property %s must be an integer", k)
			}
			patchMap[k] = t
		case "string":
			patchMap[k] = v
		}
	}

	return patchMap, nil
}
