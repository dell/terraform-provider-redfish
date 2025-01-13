package helper

import (
	"fmt"
	"strings"
	"terraform-provider-redfish/gofish/dell"
	"terraform-provider-redfish/redfish/models"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stmcginnis/gofish"
)

// ReadDatasourceRedfishDellIdracAttributes reads Dell iDRAC attributes from a Redfish service.
//
// Parameters:
// - service: The Redfish service to read the attributes from.
// - d: The DellIdracAttributes object to store the read attributes.
//
// Returns:
// - diag.Diagnostics: A diagnostics object containing any errors encountered during the read operation.
func ReadDatasourceRedfishDellIdracAttributes(service *gofish.Service, d *models.DellIdracAttributes) diag.Diagnostics {
	var diags diag.Diagnostics
	idracError := "there was an issue when reading idrac attributes"
	// get managers (Dell servers have only the iDRAC)
	managers, err := service.Managers()
	if err != nil {
		diags.AddError(idracError, err.Error())
		return diags
	}

	// Get OEM
	dellManager, err := dell.Manager(managers[0])
	if err != nil {
		diags.AddError(idracError, err.Error())
		return diags
	}

	// Get Dell attributes
	dellAttributes, err := dellManager.DellAttributes()
	if err != nil {
		diags.AddError(idracError, err.Error())
		return diags
	}
	idracAttributes, err := GetIdracAttributes(dellAttributes)
	if err != nil {
		diags.AddError(idracError, err.Error())
		return diags
	}

	attributesToReturn := make(map[string]attr.Value)

	for k, v := range idracAttributes.Attributes {
		if v != nil {
			attributesToReturn[k] = types.StringValue(fmt.Sprint(v))
		} else {
			attributesToReturn[k] = types.StringValue("")
		}
	}

	d.Attributes = types.MapValueMust(types.StringType, attributesToReturn)
	if err != nil {
		diags.AddError(idracError, err.Error())
		return diags
	}

	d.ID = types.StringValue(idracAttributes.ODataID)
	return diags
}

// GetIdracAttributes retrieves the iDRAC attributes from the given list of attributes.
//
// Parameters:
// - attributes: The list of attributes to search through.
//
// Returns:
// - *dell.Attributes: The iDRAC attributes if found, or nil if not found.
// - error: An error if the iDRAC attributes could not be found.
func GetIdracAttributes(attributes []*dell.Attributes) (*dell.Attributes, error) {
	for _, a := range attributes {
		if strings.Contains(a.ID, "iDRAC") {
			return a, nil
		}
	}
	return nil, fmt.Errorf("couldn't find iDRACAttributes")
}
