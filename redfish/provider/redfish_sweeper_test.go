package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/stmcginnis/gofish"
)

func TestMain(m *testing.M) {
	resource.TestMain(m)
}

func getSweeperClient(region string) (*gofish.Service, error) {
	endpoint := "https://" + creds.Endpoint
	clientConfig := gofish.ClientConfig{
		Endpoint:  endpoint,
		Username:  creds.Username,
		Password:  creds.Password,
		BasicAuth: true,
		Insecure:  true,
	}
	api, err := gofish.Connect(clientConfig)
	if err != nil {
		return nil, fmt.Errorf("Unable to create sweeper client %v", err)
	}

	return api.Service, nil
}
