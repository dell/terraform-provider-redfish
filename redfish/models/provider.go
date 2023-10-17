package models

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// providerConfig can be used to store data from the Terraform configuration.
type ProviderConfig struct {
	Username types.String `tfsdk:"user"`
	Password types.String `tfsdk:"password"`
}

type RedfishServer struct {
	User         types.String `tfsdk:"user"`
	Password     types.String `tfsdk:"password"`
	Endpoint     types.String `tfsdk:"endpoint"`
	ValidateCert types.Bool   `tfsdk:"validate_cert"`
}
