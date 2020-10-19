package redfish

import (
	"fmt"
	"github.com/stmcginnis/gofish"
	"github.com/stmcginnis/gofish/common"
	"net/http"
	"testing"
)

/*
Calls path
service---------systems---------embeddedSystem---------storage(collection)-----o---drives
																			    \--volumes (might have volumes or not)
*/

/*
setStorageMockedClient's mission is to return an storage struct that's needed for the tests
	and the mocked responses to the GET requests, for drives or volumes
	Params:
		- Collection: struct to return with next GET calls ready (i.e. embeddedSystem, drives or volumes)
		- options: string that can be used to set special cases:
			- if storageSubCollection is set to volumes:
				- "empty": means volumes empty collection
				- "included": means at lest there is one volume
	Returns:
		- interface{}: struct wanted by the user (i.e. gofish.Service, redfish.Storage)
		- error: errors when executing the function
*/
func setStorageMockedClient(collection string, options string) (interface{}, error) {
	testClient := &common.TestClient{}
	responseBuilder := &responseBuilder{}
	testClient.CustomReturnForActions = make(map[string][]interface{})

	rootResponse := responseBuilder.Status("200 OK").StatusCode(200).Body(rootRedfishJSON).Build()
	testClient.CustomReturnForActions[http.MethodGet] = append(testClient.CustomReturnForActions[http.MethodGet], &rootResponse)
	service, err := gofish.ServiceRoot(testClient)
	if err != nil {
		return nil, fmt.Errorf("Something went wrong with the mocked client: %s", err)
	}

	//service.Systems() will make 1 + N GET calls (where N is the number of systems, normally just one).
	//	- First one to get the system collection
	//	- The N following correspond to the number of systems embedded (Normally just one)
	systemsResponse := responseBuilder.Status("200 OK").StatusCode(200).Body(systemsRedfishJSON).Build()
	testClient.CustomReturnForActions[http.MethodGet] = append(testClient.CustomReturnForActions[http.MethodGet], &systemsResponse)
	embeddedSystemResponse := responseBuilder.Status("200 OK").StatusCode(200).Body(systemEmbeddedRedfishJSON).Build()
	testClient.CustomReturnForActions[http.MethodGet] = append(testClient.CustomReturnForActions[http.MethodGet], &embeddedSystemResponse)

	/*embedded, err := service.Systems()
	if err != nil {
		return nil, fmt.Errorf("Something went wrong with the mocked client: %s", err)
	}*/

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

	if collection == "service" {
		return service, nil
	}

	embedded, err := service.Systems()
	if err != nil {
		return nil, fmt.Errorf("Something went wrong with the mocked client: %s", err)
	}

	storage, err := embedded[0].Storage()
	if err != nil {
		return nil, fmt.Errorf("Something went wrong with the mocked client: %s", err)
	}

	switch collection {
	case "drives":
		firstDiskResponse := responseBuilder.Status("200 OK").StatusCode(200).Body(drive1RedfishJSON).Build()
		testClient.CustomReturnForActions[http.MethodGet] = append(testClient.CustomReturnForActions[http.MethodGet], &firstDiskResponse)
		secondDiskResponse := responseBuilder.Status("200 OK").StatusCode(200).Body(drive2RedfishJSON).Build()
		testClient.CustomReturnForActions[http.MethodGet] = append(testClient.CustomReturnForActions[http.MethodGet], &secondDiskResponse)
		return storage[0], nil
	case "volumes":
		switch options {
		case "empty":
			volumesResponse := responseBuilder.Status("200 OK").StatusCode(200).Body(noVolumesRedfishJSON).Build()
			testClient.CustomReturnForActions[http.MethodGet] = append(testClient.CustomReturnForActions[http.MethodGet], &volumesResponse)
			return storage[0], nil
		case "included":
			volumesResponse := responseBuilder.Status("200 OK").StatusCode(200).Body(volumesRedfishJSON).Build()
			testClient.CustomReturnForActions[http.MethodGet] = append(testClient.CustomReturnForActions[http.MethodGet], &volumesResponse)
			volumeResponse := responseBuilder.Status("200 OK").StatusCode(200).Body(volumeRedfishJSON).Build()
			testClient.CustomReturnForActions[http.MethodGet] = append(testClient.CustomReturnForActions[http.MethodGet], &volumeResponse)
			return storage[0], nil
		}

	}
	return nil, fmt.Errorf("No matches for building the test client")
}

func TestGetStorageController(t *testing.T) {
	/*
		Possible cases:
			- The controller exists
			- The controller does not exist
			- Errors with the Client (not sure if we can check that with the mocked client)
	*/
	cases := []struct {
		noTest     int
		storageID  string
		shouldPass bool
	}{
		{1, "RAID.Integrated.1-1", true},
		{2, "RAID.Integrated.1-2", false},
	}
	for _, v := range cases {
		service, err := setStorageMockedClient("service", "")
		if err != nil {
			t.Errorf("There was an error with the mocked client")
		}
		_, err = getStorageController(service.(*gofish.Service), v.storageID)
		if v.shouldPass {
			if err != nil {
				t.Errorf("Test number %v failed %v", v.noTest, err)
			}
		}

	}
}

func TestDeleteVolume(t *testing.T) {
	/*
		Possible cases:
			- The controller exists
			- The controller does not exist
			- Errors with the Client (not sure if we can check that with the mocked client)
	*/
}
