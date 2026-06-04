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
	"testing"
)

func TestControllerPCISlotString(t *testing.T) {
	tests := []struct {
		name     string
		pcislot  any
		expected string
	}{
		{
			name:     "string value",
			pcislot:  "1",
			expected: "1",
		},
		{
			name:     "numeric value",
			pcislot:  1,
			expected: "1",
		},
		{
			name:     "float value",
			pcislot:  1.0,
			expected: "1",
		},
		{
			name:     "nil value",
			pcislot:  nil,
			expected: "",
		},
		{
			name:     "empty string",
			pcislot:  "",
			expected: "",
		},
		{
			name:     "numeric zero",
			pcislot:  0,
			expected: "0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := Controller{PCISlot: tt.pcislot}
			result := c.PCISlotString()
			if result != tt.expected {
				t.Errorf("PCISlotString() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestDellStorageControllerPCISlotString(t *testing.T) {
	tests := []struct {
		name     string
		pcislot  any
		expected string
	}{
		{
			name:     "string value",
			pcislot:  "2",
			expected: "2",
		},
		{
			name:     "numeric value",
			pcislot:  2,
			expected: "2",
		},
		{
			name:     "float value",
			pcislot:  2.0,
			expected: "2",
		},
		{
			name:     "nil value",
			pcislot:  nil,
			expected: "",
		},
		{
			name:     "empty string",
			pcislot:  "",
			expected: "",
		},
		{
			name:     "numeric zero",
			pcislot:  0,
			expected: "0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := DellStorageController{PCISlot: tt.pcislot}
			result := d.PCISlotString()
			if result != tt.expected {
				t.Errorf("PCISlotString() = %v, want %v", result, tt.expected)
			}
		})
	}
}
