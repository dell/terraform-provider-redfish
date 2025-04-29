/*
Copyright (c) 2021-2025 Dell Inc., or its subsidiaries. All Rights Reserved.

Licensed under the Mozilla Public License Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://mozilla.org/MPL/2.0/


Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package dell

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/stmcginnis/gofish/common"
)

// AttributesMap handles the Dell attribute values that may be any of several
// types and adds some basic helper methods to make accessing values easier.
type AttributesMap map[string]interface{}

// String gets the string representation of the attribute value.
func (a AttributesMap) String(name string) string {
	if val, ok := a[name]; ok {
		return fmt.Sprintf("%v", val)
	}

	return ""
}

// Float64 gets the value as a float64 or 0 if that is not possible.
func (a AttributesMap) Float64(name string) float64 {
	if val, ok := a[name]; ok {
		return val.(float64)
	}

	return 0
}

// Int gets the value as an integer or 0 if that is not possible.
func (a AttributesMap) Int(name string) int {
	// Integer values may be interpeted as float64, so get it as that first,
	// then coerce down to int.
	floatVal := int(a.Float64(name))
	return (floatVal)
}

// Bool gets the value as a boolean or returns false.
func (a AttributesMap) Bool(name string) bool {
	maybeBool := a.String(name)
	maybeBool = strings.ToLower(maybeBool)
	return (maybeBool == "true" ||
		maybeBool == "1" ||
		maybeBool == "enabled")
}

// Attributes is used to represent Dell attributes
type Attributes struct {
	common.Entity

	// ODataContext is the odata context.
	ODataContext string `json:"@odata.context"`
	// ODataType is the odata type.
	ODataType string `json:"@odata.type"`
	// Description provides a description of this resource.
	Description string
	// This property shall contain the list of Dell attributes and their values
	// as determined by the manufacturer or provider.
	Attributes AttributesMap
	// settingsTarget is the URL to send settings update requests to.
	settingsObject common.Link
	// settingsApplyTimes is a set of allowed settings update apply times. If none
	// are specified, then the system does not provide that information.
	settingsApplyTimes []common.ApplyTime
	// rawData holds the original serialized JSON so we can compare updates.
	rawData []byte
}

// UnmarshalJSON unmarshals Dell Attributes JSON object from the raw JSON
func (d *Attributes) UnmarshalJSON(data []byte) error {
	type temp Attributes

	var t struct {
		temp
		Settings common.Settings `json:"@Redfish.Settings"`
	}

	err := json.Unmarshal(data, &t)
	if err != nil {
		return err
	}

	*d = Attributes(t.temp)
	d.settingsObject = t.Settings.SettingsObject
	d.settingsApplyTimes = t.settingsApplyTimes
	d.rawData = data

	return nil
}

// GetDellAttributes return a DellAttributes pointer given a client and a uri to query
func GetDellAttributes(c common.Client, uri string) (*Attributes, error) {
	resp, err := c.Get(uri)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var dellAttributes Attributes
	err = json.NewDecoder(resp.Body).Decode(&dellAttributes)
	if err != nil {
		return nil, err
	}

	dellAttributes.SetClient(c)
	return &dellAttributes, nil
}

// ListReferenceDellAttributes return an slice of DellAttributes pointers given a client and common.Links
func ListReferenceDellAttributes(c common.Client, links common.Links) ([]*Attributes, error) {
	var result []*Attributes

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
