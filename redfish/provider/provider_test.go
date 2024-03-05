/*
Copyright (c) 2023-2024 Dell Inc., or its subsidiaries. All Rights Reserved.

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

package provider

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/joho/godotenv"
)

var (
	testAccProtoV6ProviderFactories map[string]func() (tfprotov6.ProviderServer, error)
	creds                           TestingServerCredentials
	image64Boot                     string
	imageEfiBoot                    string
	drive                           string
)

// // TestingServerCredentials Struct used to store the credentials we pass for testing. This allows us to pass testing
// // credentials via environment variables instead of having them hard coded
type TestingServerCredentials struct {
	Username  string
	Password  string
	Endpoint  string
	Endpoint2 string
	Insecure  bool
}

func init() {
	err := godotenv.Load("redfish_test.env")
	if err != nil {
		fmt.Println(err.Error())
	}

	testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
		// newProvider is an example function that returns a tfsdk.Provider
		"redfish": providerserver.NewProtocol6WithError(New()),
	}

	creds = TestingServerCredentials{
		Username:  os.Getenv("TF_TESTING_USERNAME"),
		Password:  os.Getenv("TF_TESTING_PASSWORD"),
		Endpoint:  os.Getenv("TF_TESTING_ENDPOINT"),
		Endpoint2: os.Getenv("TF_TESTING_ENDPOINT2"),
		Insecure:  false,
	}

	// virtual media environment variable
	image64Boot = os.Getenv("TF_TESTING_VIRTUAL_MEDIA_IMAGE_PATH_64Boot")
	imageEfiBoot = os.Getenv("TF_TESTING_VIRTUAL_MEDIA_IMAGE_PATH_EfiBoot")
	// storage volume environment varibale
	drive = os.Getenv("TF_TESTING_STORAGE_VOLUME_DRIVE")
}

func testAccPreCheck(t *testing.T) {
	if v := os.Getenv("TF_TESTING_USERNAME"); v == "" {
		t.Fatal("TF_TESTING_USERNAME must be set for acceptance tests")
	}

	if v := os.Getenv("TF_TESTING_PASSWORD"); v == "" {
		t.Fatal("TF_TESTING_PASSWORD must be set for acceptance tests")
	}

	if v := os.Getenv("TF_TESTING_ENDPOINT"); v == "" {
		t.Fatal("TF_TESTING_ENDPOINT must be set for acceptance tests")
	}

	if v := os.Getenv("TF_TESTING_ENDPOINT2"); v == "" {
		t.Fatal("TF_TESTING_ENDPOINT2 must be set for acceptance tests")
	}
}

// getID returns the ID of the resource in import scenarios
func getID(d *terraform.State, name string) (string, error) {
	allRes := d.RootModule().Resources
	if res, ok := allRes[name]; ok {
		return res.Primary.ID, nil
	}
	return "", fmt.Errorf("resource %s not found", name)
}
