package models

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// UserAccount struct
type UserAccount struct {
	ID            types.String    `tfsdk:"id"`
	Enabled       types.Bool      `tfsdk:"enabled"`
	Password      types.String    `tfsdk:"password"`
	RedfishServer []RedfishServer `tfsdk:"redfish_server"`
	RoleID        types.String    `tfsdk:"role_id"`
	UserID        types.String    `tfsdk:"user_id"`
	Username      types.String    `tfsdk:"username"`
}
