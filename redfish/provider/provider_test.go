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

var testAccProtoV6ProviderFactories map[string]func() (tfprotov6.ProviderServer, error)
var creds TestingServerCredentials
var image64Boot string 
var imageEfiBoot string 
var drive string

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
