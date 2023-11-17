package models

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// RedfishManagerReset to construct terraform schema for manager reset resource.
type RedfishManagerReset struct {
	Id            types.String    `tfsdk:"id"`
	ManagerID     types.String    `tfsdk:"manager_id"`
	ResetType     types.String    `tfsdk:"reset_type"`
	RedfishServer []RedfishServer `tfsdk:"redfish_server"`
}
