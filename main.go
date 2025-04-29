/*
Copyright (c) 2020-2025 Dell Inc., or its subsidiaries. All Rights Reserved.

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
//nolint:all
//go:generate go run github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs@v0.19.4 generate --rendered-website-dir docs --provider-name terraform-provider-redfish

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
