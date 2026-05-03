/*
Copyright (c) 2025 Dell Inc., or its subsidiaries. All Rights Reserved.

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

package helper

import (
	"testing"

	"terraform-provider-redfish/redfish/models"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stmcginnis/gofish/redfish"
)

// planWithValues returns a VirtualMedia plan populated with the given values,
// simulating what Terraform passes in from the user's configuration.
func planWithValues(image, transferMethod, transferProtocol string, writeProtected bool, systemID string) models.VirtualMedia {
	return models.VirtualMedia{
		Image:                types.StringValue(image),
		TransferMethod:       types.StringValue(transferMethod),
		TransferProtocolType: types.StringValue(transferProtocol),
		WriteProtected:       types.BoolValue(writeProtected),
		SystemID:             types.StringValue(systemID),
	}
}

// TestUpdateVirtualMediaState_NormalServer verifies that when the Redfish API
// returns fully-populated fields (iDRAC 6.x / 7.x behaviour), the response
// values are written into state and the plan values are not used as fallback.
func TestUpdateVirtualMediaState_NormalServer(t *testing.T) {
	plan := planWithValues(
		"https://example.com/old.iso",
		"Stream",
		"HTTPS",
		true,
		"System.Embedded.1",
	)

	response := &redfish.VirtualMedia{}
	response.ODataID = "/redfish/v1/Systems/System.Embedded.1/VirtualMedia/CD"
	response.Image = "https://example.com/bootstrap/server01.iso"
	response.Inserted = true
	response.TransferMethod = redfish.TransferMethod("Stream")
	response.TransferProtocolType = redfish.TransferProtocolType("HTTPS")
	response.WriteProtected = true

	result := UpdateVirtualMediaState(response, plan)

	if result.Image.ValueString() != response.Image {
		t.Errorf("Image: got %q, want %q", result.Image.ValueString(), response.Image)
	}
	if result.TransferMethod.ValueString() != string(response.TransferMethod) {
		t.Errorf("TransferMethod: got %q, want %q", result.TransferMethod.ValueString(), string(response.TransferMethod))
	}
	if result.TransferProtocolType.ValueString() != string(response.TransferProtocolType) {
		t.Errorf("TransferProtocolType: got %q, want %q", result.TransferProtocolType.ValueString(), string(response.TransferProtocolType))
	}
	if result.WriteProtected.ValueBool() != response.WriteProtected {
		t.Errorf("WriteProtected: got %v, want %v", result.WriteProtected.ValueBool(), response.WriteProtected)
	}
	if result.VirtualMediaID.ValueString() != response.ODataID {
		t.Errorf("VirtualMediaID: got %q, want %q", result.VirtualMediaID.ValueString(), response.ODataID)
	}
	if result.SystemID.ValueString() != plan.SystemID.ValueString() {
		t.Errorf("SystemID: got %q, want %q", result.SystemID.ValueString(), plan.SystemID.ValueString())
	}
}

// TestUpdateVirtualMediaState_R670EmptyResponse simulates R670 behaviour where
// the server returns empty strings for Image/TransferMethod/TransferProtocolType
// and false for WriteProtected after a successful InsertMedia call.
// The fix must fall back to plan values so Terraform does not see
// "inconsistent result after apply".
func TestUpdateVirtualMediaState_R670EmptyResponse(t *testing.T) {
	plan := planWithValues(
		"https://example.com/bootstrap/R670-server.iso",
		"Stream",
		"HTTPS",
		true,
		"System.Embedded.1",
	)

	// Simulate R670: Inserted is true (media was mounted) but all other fields
	// come back empty/zero from the GET response.
	response := &redfish.VirtualMedia{}
	response.ODataID = "/redfish/v1/Systems/System.Embedded.1/VirtualMedia/CD"
	response.Image = ""
	response.Inserted = true
	response.TransferMethod = redfish.TransferMethod("")
	response.TransferProtocolType = redfish.TransferProtocolType("")
	response.WriteProtected = false

	result := UpdateVirtualMediaState(response, plan)

	if result.Image.ValueString() != plan.Image.ValueString() {
		t.Errorf("Image: got %q, want plan value %q", result.Image.ValueString(), plan.Image.ValueString())
	}
	if result.TransferMethod.ValueString() != plan.TransferMethod.ValueString() {
		t.Errorf("TransferMethod: got %q, want plan value %q", result.TransferMethod.ValueString(), plan.TransferMethod.ValueString())
	}
	if result.TransferProtocolType.ValueString() != plan.TransferProtocolType.ValueString() {
		t.Errorf("TransferProtocolType: got %q, want plan value %q", result.TransferProtocolType.ValueString(), plan.TransferProtocolType.ValueString())
	}
	if result.WriteProtected.ValueBool() != plan.WriteProtected.ValueBool() {
		t.Errorf("WriteProtected: got %v, want plan value %v", result.WriteProtected.ValueBool(), plan.WriteProtected.ValueBool())
	}
	if result.VirtualMediaID.ValueString() != response.ODataID {
		t.Errorf("VirtualMediaID: got %q, want %q", result.VirtualMediaID.ValueString(), response.ODataID)
	}
	if result.Inserted.ValueBool() != response.Inserted {
		t.Errorf("Inserted: got %v, want %v", result.Inserted.ValueBool(), response.Inserted)
	}
}

// TestUpdateVirtualMediaState_R670WriteProtectedFalse verifies that when the
// user explicitly configures write_protected=false on R670 (which also returns
// Image=""), the plan value (false) is preserved — not accidentally overwritten.
func TestUpdateVirtualMediaState_R670WriteProtectedFalse(t *testing.T) {
	plan := planWithValues(
		"https://example.com/bootstrap/server.iso",
		"Stream",
		"HTTPS",
		false, // user explicitly set false
		"",
	)

	response := &redfish.VirtualMedia{}
	response.ODataID = "/redfish/v1/Systems/System.Embedded.1/VirtualMedia/CD"
	response.Image = ""
	response.Inserted = true
	response.TransferMethod = redfish.TransferMethod("")
	response.TransferProtocolType = redfish.TransferProtocolType("")
	response.WriteProtected = false

	result := UpdateVirtualMediaState(response, plan)

	if result.WriteProtected.ValueBool() != false {
		t.Errorf("WriteProtected: got %v, want false (plan value)", result.WriteProtected.ValueBool())
	}
}

// TestUpdateVirtualMediaState_PartialResponse verifies that when only some fields
// are returned by the server, per-field fallback logic works correctly.
// Image is populated but TransferMethod/TransferProtocolType are empty.
func TestUpdateVirtualMediaState_PartialResponse(t *testing.T) {
	plan := planWithValues(
		"https://example.com/bootstrap/server.iso",
		"Stream",
		"HTTPS",
		true,
		"",
	)

	response := &redfish.VirtualMedia{}
	response.ODataID = "/redfish/v1/Systems/System.Embedded.1/VirtualMedia/CD"
	response.Image = "https://example.com/bootstrap/server.iso"
	response.Inserted = true
	response.TransferMethod = redfish.TransferMethod("")    // empty from server
	response.TransferProtocolType = redfish.TransferProtocolType("") // empty from server
	response.WriteProtected = true

	result := UpdateVirtualMediaState(response, plan)

	// Image came from response
	if result.Image.ValueString() != response.Image {
		t.Errorf("Image: got %q, want response value %q", result.Image.ValueString(), response.Image)
	}
	// TransferMethod fell back to plan
	if result.TransferMethod.ValueString() != plan.TransferMethod.ValueString() {
		t.Errorf("TransferMethod: got %q, want plan value %q", result.TransferMethod.ValueString(), plan.TransferMethod.ValueString())
	}
	// TransferProtocolType fell back to plan
	if result.TransferProtocolType.ValueString() != plan.TransferProtocolType.ValueString() {
		t.Errorf("TransferProtocolType: got %q, want plan value %q", result.TransferProtocolType.ValueString(), plan.TransferProtocolType.ValueString())
	}
	// WriteProtected: Image was non-empty so response value used
	if result.WriteProtected.ValueBool() != response.WriteProtected {
		t.Errorf("WriteProtected: got %v, want response value %v", result.WriteProtected.ValueBool(), response.WriteProtected)
	}
}
