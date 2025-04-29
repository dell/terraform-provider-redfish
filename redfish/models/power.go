/*
Copyright (c) 2023-2025 Dell Inc., or its subsidiaries. All Rights Reserved.

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

package models

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Power to construct terraform schema for power resource.
type Power struct {
	PowerId            types.String    `tfsdk:"id"`
	RedfishServer      []RedfishServer `tfsdk:"redfish_server"`
	DesiredPowerAction types.String    `tfsdk:"desired_power_action"`
	MaximumWaitTime    types.Int64     `tfsdk:"maximum_wait_time"`
	CheckInterval      types.Int64     `tfsdk:"check_interval"`
	PowerState         types.String    `tfsdk:"power_state"`
	SystemID           types.String    `tfsdk:"system_id"`
}
