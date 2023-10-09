package provider

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/joho/godotenv"
)

var testAccProtoV6ProviderFactories map[string]func() (tfprotov6.ProviderServer, error)
var creds TestingServerCredentials

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
		panic(err)
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
		t.Fatal("TF_TESTING_ENDPOINT must be set for acceptance tests")
	}
}

func skipTest() bool {
	return os.Getenv("TF_ACC") == "" || os.Getenv("ACC_DETAIL") == ""
}

// // TestingServerCredentials Struct used to store the credentials we pass for testing. This allows us to pass testing
// // credentials via environment variables instead of having them hard coded
// type TestingServerCredentials struct {
// 	Username  string
// 	Password  string
// 	Endpoint  string
// 	Endpoint2 string
// 	Insecure  bool
// }

// var testAccProviders map[string]*schema.Provider
// var testAccProvider *schema.Provider
// var creds TestingServerCredentials

// func init() {

// 	testAccProvider = Provider()
// 	testAccProviders = map[string]*schema.Provider{
// 		"redfish": testAccProvider,
// 	}

// 	creds = TestingServerCredentials{
// 		Username:  os.Getenv("TF_TESTING_USERNAME"),
// 		Password:  os.Getenv("TF_TESTING_PASSWORD"),
// 		Endpoint:  os.Getenv("TF_TESTING_ENDPOINT"),
// 		Endpoint2: os.Getenv("TF_TESTING_ENDPOINT2"),
// 		Insecure:  false,
// 	}

// }
