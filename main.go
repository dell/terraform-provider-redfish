package main

import (
	"context"
	"flag"
	"log"
	"terraform-provider-redfish/redfish/provider"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
)

// Run "go generate" to format example terraform files and generate the docs for the registry/website

// If you do not have terraform installed, you can remove the formatting command, but its suggested to
// ensure the documentation is formatted properly.
//go:generate terraform fmt -recursive ./examples/

// Run the docs generation tool, check its repository for more information on how it works and how docs
// can be customized.
//go:generate go run github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs

func main() {
	var debug bool

	flag.BoolVar(&debug, "debug", false, "set to true to run the provider with support for debuggers like delve")
	flag.Parse()

	err := providerserver.Serve(context.Background(), provider.New, providerserver.ServeOpts{
		Address: "registry.terraform.io/dell/redfish",
		Debug:   debug,
	})
	if err != nil {
		log.Fatal(err.Error())
	}
}

// func main() {

// 	var debugMode bool

// 	// Set this flag to true if you want the provider to run in debug mode. Leaving it as is will cause it to run
// 	// normally.
// 	flag.BoolVar(&debugMode, "debug", false, "set to true to run the provider with support for debuggers like delve")
// 	flag.Parse()

// 	opts := &plugin.ServeOpts{
// 		Debug:        debugMode,
// 		ProviderAddr: "registry.terraform.io/dell/redfish",
// 		ProviderFunc: redfish.Provider,
// 	}

// 	if debugMode {
// 		err := plugin.Debug(context.Background(), "registry.terraform.io/dell/redfish", opts)
// 		if err != nil {
// 			log.Fatal(err.Error())
// 		}
// 		return
// 	}

// 	plugin.Serve(opts)

// }
