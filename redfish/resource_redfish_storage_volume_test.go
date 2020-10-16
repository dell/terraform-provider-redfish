package redfish_test

import (
	"fmt"
	"github.com/stmcginnis/gofish"
	"github.com/stmcginnis/gofish/common"
	"net/http"
	"testing"
)

var testClient common.TestClient

func TestGetStorageController(t *testing.T) {
	/*
		Possible cases:
			- The controller exists
			- The controller does not exist
			- Errors with the Client
	*/
	//Building the mocked client as well as the predefined responses
	var testClient = common.TestClient{}
	responseBuilder := &responseBuilder{}
	testClient.CustomReturnForActions = make(map[string][]interface{})

	//gofish.ServiceRoot will only make a Get call
	rootResponse := responseBuilder.Status("200 OK").StatusCode(200).Body(rootRedfishJSON).Build()
	testClient.CustomReturnForActions[http.MethodGet] = append(testClient.CustomReturnForActions[http.MethodGet], &rootResponse)
	service, err := gofish.ServiceRoot(&testClient)
	if err != nil {
		t.Errorf("Something went wrong with the mocked client: %s", err)
	}

	//service.Systems() will make 1 + N GET calls (where N is the number of systems, normally just one).
	//	- First one to get the system collection
	//	- The N following correspond to the number of systems embedded (Normally just one)
	systemsResponse := responseBuilder.Status("200 OK").StatusCode(200).Body(systemsRedfishJSON).Build()
	testClient.CustomReturnForActions[http.MethodGet] = append(testClient.CustomReturnForActions[http.MethodGet], &systemsResponse)
	embeddedSystemResponse := responseBuilder.Status("200 OK").StatusCode(200).Body(systemEmbeddedRedfishJSON).Build()
	testClient.CustomReturnForActions[http.MethodGet] = append(testClient.CustomReturnForActions[http.MethodGet], &embeddedSystemResponse)
	embedded, err := service.Systems()
	if err != nil {
		t.Errorf("Something went wrong with the mocked client: %s", err)
	}

	//embedded[0].Storage() will make 1 + N calls, (where N is the number of storage controllers)
	//	- First one to get the storage collection
	//	- The N following correspond to the number of storage controllers (The example collection has 3 controllers)
	storageResponse := responseBuilder.Status("200 OK").StatusCode(200).Body(storageRedfishJSON).Build()
	testClient.CustomReturnForActions[http.MethodGet] = append(testClient.CustomReturnForActions[http.MethodGet], &storageResponse)
	firstStorageResponse := responseBuilder.Status("200 OK").StatusCode(200).Body(storage1RedfishJSON).Build()
	testClient.CustomReturnForActions[http.MethodGet] = append(testClient.CustomReturnForActions[http.MethodGet], &firstStorageResponse)
	secondStorageResponse := responseBuilder.Status("200 OK").StatusCode(200).Body(storage2RedfishJSON).Build()
	testClient.CustomReturnForActions[http.MethodGet] = append(testClient.CustomReturnForActions[http.MethodGet], &secondStorageResponse)
	thirdStorageResponse := responseBuilder.Status("200 OK").StatusCode(200).Body(storage3RedfishJSON).Build()
	testClient.CustomReturnForActions[http.MethodGet] = append(testClient.CustomReturnForActions[http.MethodGet], &thirdStorageResponse)

	storage, err := embedded[0].Storage()
	if err != nil {
		t.Errorf("Something went wrong with the mocked client: %s", err)
	}
	fmt.Println(storage[0].Description)
}
