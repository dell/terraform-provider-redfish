package redfish

import (
	"os"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// TestingServerCredentials Struct used to store the credentials we pass for testing. This allows us to pass testing
// credentials via environment variables instead of having them hard coded
type TestingServerCredentials struct {
	Username  string
	Password  string
	Endpoint  string
	Endpoint2 string
	Insecure  bool
}

var testAccProviders map[string]*schema.Provider
var testAccProvider *schema.Provider
var creds TestingServerCredentials

func init() {

	testAccProvider = Provider()
	testAccProviders = map[string]*schema.Provider{
		"redfish": testAccProvider,
	}

	creds = TestingServerCredentials{
		Username:  os.Getenv("TF_TESTING_USERNAME"),
		Password:  os.Getenv("TF_TESTING_PASSWORD"),
		Endpoint:  os.Getenv("TF_TESTING_ENDPOINT"),
		Endpoint2: os.Getenv("TF_TESTING_ENDPOINT2"),
		Insecure:  false,
	}

}
