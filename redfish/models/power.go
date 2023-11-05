package models

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Power to construct terraform schema for power resource.
type Power struct {
	PowerId            types.String  `tfsdk:"id"`
	RedfishServer      []RedfishServer `tfsdk:"redfish_server"`
	DesiredPowerAction types.String  `tfsdk:"desired_power_action"`
	MaximumWaitTime    types.Int64   `tfsdk:"maximum_wait_time"`
	CheckInterval      types.Int64   `tfsdk:"check_interval"`
	PowerState         types.String  `tfsdk:"power_state"`
}
