package dell

import (
	// "encoding/json"
	// "strings"

	"testing"

	"github.com/stmcginnis/gofish"
)

func TestGetSystem(t *testing.T) {
	clientConfig := gofish.ClientConfig{
		Endpoint:  "https://<ip>",
		Username:  "<username>",
		Password:  "<password>",
		BasicAuth: true,
		Insecure:  true,
	}
	api, err := gofish.Connect(clientConfig)
	if err != nil {
		t.Fatalf("error connecting to redfish API: %s", err.Error())
	}
	t.Logf("Connected to redfish API: %s", api.Service.ODataID)

	serv, err1 := GetService(api)
	if err1 != nil {
		t.Fatalf("error getting service root: %s", err1.Error())
	}
	t.Logf("Got service root with system collection ID: %s", serv.Params.Systems)

	system, err2 := serv.GetSystem()
	if err2 != nil {
		t.Fatalf("error getting system: %s", err2.Error())
	}
	t.Logf("Got system with storage collection ID: %s", system.Params.Storage)

	storages, err3 := system.GetStorages()
	if err3 != nil {
		t.Fatalf("error getting system: %s", err3.Error())
	}
	t.Logf("Got storages with storage collection ID: %s", system.Params.Storage)
	for i, s := range storages {
		t.Logf("%d ==> Got storage with ID: %s and controller collection ID: %s", i, s.parent.ODataID, s.Params.Controllers)
	}
}
