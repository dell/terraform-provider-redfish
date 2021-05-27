package dell

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/stmcginnis/gofish/common"
)

// Attributes handles the Dell attribute values that may be any of several
// types and adds some basic helper methods to make accessing values easier.
type Attributes map[string]interface{}

// String gets the string representation of the attribute value.
func (a Attributes) String(name string) string {
	if val, ok := a[name]; ok {
		return fmt.Sprintf("%v", val)
	}

	return ""
}

// Float64 gets the value as a float64 or 0 if that is not possible.
func (a Attributes) Float64(name string) float64 {
	if val, ok := a[name]; ok {
		return val.(float64)
	}

	return 0
}

// Int gets the value as an integer or 0 if that is not possible.
func (a Attributes) Int(name string) int {
	// Integer values may be interpeted as float64, so get it as that first,
	// then coerce down to int.
	floatVal := int(a.Float64(name))
	return (floatVal)
}

// Bool gets the value as a boolean or returns false.
func (a Attributes) Bool(name string) bool {
	maybeBool := a.String(name)
	maybeBool = strings.ToLower(maybeBool)
	return (maybeBool == "true" ||
		maybeBool == "1" ||
		maybeBool == "enabled")
}

type DellAttributes struct {
	common.Entity

	// ODataContext is the odata context.
	ODataContext string `json:"@odata.context"`
	// ODataType is the odata type.
	ODataType string `json:"@odata.type"`
	// Description provides a description of this resource.
	Description string
	// This property shall contain the list of Dell attributes and their values
	// as determined by the manufacturer or provider.
	Attributes Attributes
	// settingsTarget is the URL to send settings update requests to.
	settingsObject common.Link
	// settingsApplyTimes is a set of allowed settings update apply times. If none
	// are specified, then the system does not provide that information.
	settingsApplyTimes []common.ApplyTime
	// rawData holds the original serialized JSON so we can compare updates.
	rawData []byte
}

func (d *DellAttributes) UnmarshalJSON(data []byte) error {
	type temp DellAttributes

	var t struct {
		temp
		Settings common.Settings `json:"@Redfish.Settings"`
	}

	err := json.Unmarshal(data, &t)
	if err != nil {
		return err
	}

	*d = DellAttributes(t.temp)
	d.settingsObject = t.Settings.SettingsObject
	d.settingsApplyTimes = t.settingsApplyTimes
	d.rawData = data

	return nil
}

// GetDellAttributes return a DellAttributes pointer given a client and a uri to query
func GetDellAttributes(c common.Client, uri string) (*DellAttributes, error) {
	resp, err := c.Get(uri)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var dellAttributes DellAttributes
	err = json.NewDecoder(resp.Body).Decode(&dellAttributes)
	if err != nil {
		return nil, err
	}

	dellAttributes.SetClient(c)
	return &dellAttributes, nil
}

// ListReferenceDellAttributes return an slice of DellAttributes pointers given a client and common.Links
func ListReferenceDellAttributes(c common.Client, links common.Links) ([]*DellAttributes, error) {
	var result []*DellAttributes

	if len(links) == 0 {
		return result, nil
	}

	for _, attLink := range links {
		attr, err := GetDellAttributes(c, string(attLink))
		if err != nil {
			return nil, err
		}
		result = append(result, attr)
	}

	return result, nil
}
