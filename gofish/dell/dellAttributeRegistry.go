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
	"reflect"

	"github.com/stmcginnis/gofish/common"
)

// AttributeEnumValue struct represents attributes to be used with enums
type AttributeEnumValue struct {
	ValueDisplayName string
	ValueName        string
}

// ManagerAttribute represent attributes to create Dell OEM Manager
type ManagerAttribute struct {
	AttributeName string
	DefaultValue  interface{} // This might be string or int.
	DisplayName   string
	DisplayOrder  int
	HelpText      string
	Hidden        bool
	ID            string
	MaxLength     int // To be used with strings/passwords
	MinLength     int // To be used with strings/passwords
	UpperBound    int // To be used with integers
	LowerBound    int // To be used with integers
	MenuPath      string
	Readonly      bool
	WriteOnly     bool
	Regex         string
	Type          string
	Value         []AttributeEnumValue // To be used with Enums
}

// SupportedSystem struct represents details of the supported systems
type SupportedSystem struct {
	FirmwareVersion string
	ProductName     string
	SystemId        string
}

// ManagerAttributeRegistry contains attriutes for manager attribute registry
type ManagerAttributeRegistry struct {
	*common.Resource
	Language     string
	OwningEntity string
	Attributes   []ManagerAttribute
	// Dependencies
	// Menus
	RegistryPrefix   string
	RegistryVersion  string
	SupportedSystems []SupportedSystem
}

// GetDellManagerAttributeRegistry is function to fetch manager attribute registry
func GetDellManagerAttributeRegistry(c common.Client, uri string) (*ManagerAttributeRegistry, error) {
	resp, err := c.Get(uri)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var managerAttributeRegistry ManagerAttributeRegistry
	err = json.NewDecoder(resp.Body).Decode(&managerAttributeRegistry)
	if err != nil {
		return nil, err
	}

	managerAttributeRegistry.SetClient(c)
	return &managerAttributeRegistry, nil
}

// UnmarshalJSON unmarshals Manager Attribute object from raw JSON
func (m *ManagerAttributeRegistry) UnmarshalJSON(data []byte) error {
	type temp ManagerAttributeRegistry
	type RegistryEntries struct {
		Attributes []ManagerAttribute
	}
	var t struct {
		temp
		RegistryEntries RegistryEntries
	}

	err := json.Unmarshal(data, &t)
	if err != nil {
		return err
	}

	*m = ManagerAttributeRegistry(t.temp)
	m.Attributes = t.RegistryEntries.Attributes

	return nil
}

// GetAttributeType returns an string that says if the attribute is "string" for Enumeration, Password and String or "int" if Integer
// error is set if not attribute was found
func (m *ManagerAttributeRegistry) GetAttributeType(attributeName string) (string, error) {
	attr, err := m.getAttribute(attributeName)
	if err != nil {
		return "", err
	}

	switch attr.Type {
	case "Integer":
		return "int", nil
	case "Enumeration", "Password", "String":
		return "string", nil
	}

	return "", fmt.Errorf("type out from Integer, Enumeration, Password or String")
}

// CheckAttribute checks if the given attribute name and value are compliant with the attribute registry.
func (m *ManagerAttributeRegistry) CheckAttribute(attributeName string, value interface{}) error {
	attr, err := m.getAttribute(attributeName)
	if err != nil {
		return err
	}

	// First check value is compliant
	switch v := reflect.ValueOf(value); v.Kind() {
	case reflect.String:
		if attr.Type != "String" && attr.Type != "Password" && attr.Type != "Enumeration" {
			return fmt.Errorf("value passed is string but attribute checked is neither String or Password type")
		}

		// Check if readonly
		if attr.Readonly {
			return fmt.Errorf("property %s cannot be written as it is read only", attributeName)
		}
		switch attr.Type {
		case "String", "Password":
			// Check string bounds
			if len(v.String()) < attr.MinLength || len(v.String()) > attr.MaxLength {
				return fmt.Errorf("value to check is not compliant. Attribute length %d, Min length %d, max length %d",
					len(v.String()), attr.MinLength, attr.MaxLength)
			}
		case "Enumeration":
			err := checkValueDisplayNameArray(v.String(), attr.Value)
			if err != nil {
				return err
			}
		}

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if attr.Type != "Integer" {
			return fmt.Errorf("value passed is integer but attribute checked is not integer")
		}
		// Check if readonly
		if attr.Readonly {
			return fmt.Errorf("property %s cannot be written as it is read only", attributeName)
		}
		// Check integer bounds
		if v.Int() < int64(attr.LowerBound) || v.Int() > int64(attr.UpperBound) {
			return fmt.Errorf("value to check is not compliant. Value is %d, lower bound is %d, upper bound is %d",
				v.Int(), attr.LowerBound, attr.UpperBound)
		}

	default:
		return fmt.Errorf("only integers or strings are allowed for attributes")
	}
	return nil
}

func (m *ManagerAttributeRegistry) getAttribute(attributeName string) (*ManagerAttribute, error) {
	for _, v := range m.Attributes {
		if v.AttributeName == attributeName {
			return &v, nil
		}
	}
	return nil, fmt.Errorf("attribute %s was not found", attributeName)
}

func checkValueDisplayNameArray(value string, possibleValues []AttributeEnumValue) error {
	for _, v := range possibleValues {
		if v.ValueDisplayName == value {
			return nil
		}
	}

	var helpErrMsg string
	for i, v := range possibleValues {
		helpErrMsg += v.ValueDisplayName
		if i < len(possibleValues)-1 {
			helpErrMsg += ", "
		}
	}
	return fmt.Errorf("enumeration value given is not permitted. Allowed values: %s", helpErrMsg)
}
