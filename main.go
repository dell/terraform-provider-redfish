package main

import (
	"github.com/dell/terraform-provider-redfish/redfish"
	"github.com/hashicorp/terraform-plugin-sdk/v2/plugin"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: redfish.Provider})
}
