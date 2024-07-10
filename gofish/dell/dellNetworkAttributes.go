/*
Copyright (c) 2024 Dell Inc., or its subsidiaries. All Rights Reserved.

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

	"github.com/stmcginnis/gofish/common"
)

// NetworkAttributes is used to represent Dell network attributes
type NetworkAttributes struct {
	common.Entity

	// ODataContext is the odata context.
	ODataContext string `json:"@odata.context"`
	// ODataType is the odata type.
	ODataType string `json:"@odata.type"`
	// Description provides a description of this resource.
	Description string
	// AttributeRegistry for this network device function
	AttributeRegistry string
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

// UnmarshalJSON unmarshals NetworkAttributes JSON object from the raw JSON
func (d *NetworkAttributes) UnmarshalJSON(data []byte) error {
	type temp NetworkAttributes

	var t struct {
		temp
		Settings common.Settings `json:"@Redfish.Settings"`
	}

	err := json.Unmarshal(data, &t)
	if err != nil {
		return err
	}

	*d = NetworkAttributes(t.temp)
	d.settingsObject = t.Settings.SettingsObject
	d.settingsApplyTimes = t.settingsApplyTimes
	d.rawData = data

	return nil
}

// GetDellNetworkAttributes return a DellNetworkAttributes pointer given a client and a uri to query
func GetDellNetworkAttributes(c common.Client, uri string) (*NetworkAttributes, error) {
	resp, err := c.Get(uri)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var dellNetworkAttributes NetworkAttributes
	err = json.NewDecoder(resp.Body).Decode(&dellNetworkAttributes)
	if err != nil {
		return nil, err
	}

	dellNetworkAttributes.SetClient(c)
	return &dellNetworkAttributes, nil
}
