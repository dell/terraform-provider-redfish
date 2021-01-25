package redfish

import (
	_ "context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stmcginnis/gofish"
	"github.com/stmcginnis/gofish/common"
	_ "github.com/stmcginnis/gofish/redfish"
	"net/http"
	"testing"
)

func TestCreateRedfishStorageVolume(t *testing.T) {
	var testClient *common.TestClient
	var responseBuilder *responseBuilder

	//Set mocked configmap
	configMap := map[string]interface{}{
		"storage_controller_id": "RAID.Integrated.1-1",
		"volume_name":           "MyVol",
		"volume_type":           "Mirrored",
		"volume_disks":          []interface{}{"Physical Disk 0:1:0", "Physical Disk 0:1:1"},
		"settings_apply_time":   "Immediate",
	}

	//First test - Create a volume with volume_type included in controller OperationApplyTime and POST succeed
	d := schema.TestResourceDataRaw(t, getResourceStorageVolumeSchema(), configMap)
	testClient = &common.TestClient{}
	responseBuilder = NewResponseBuilder()

	//Get mocked client
	service, err := setStorageMockedClient(testClient, responseBuilder)
	if err != nil {
		t.Errorf("Error when creating the mocked client: %v", err)
	}

	//Get operationApplyTimes
	volumesEmptyCollection := responseBuilder.Status("200 OK").StatusCode(200).Body(noVolumesRedfishJSON).Build()
	testClient.CustomReturnForActions[http.MethodGet] = append(testClient.CustomReturnForActions[http.MethodGet], &volumesEmptyCollection)

	firstDiskResponse := responseBuilder.Status("200 OK").StatusCode(200).Body(drive1RedfishJSON).Build()
	testClient.CustomReturnForActions[http.MethodGet] = append(testClient.CustomReturnForActions[http.MethodGet], &firstDiskResponse)

	secondDiskResponse := responseBuilder.Status("200 OK").StatusCode(200).Body(drive2RedfishJSON).Build()
	testClient.CustomReturnForActions[http.MethodGet] = append(testClient.CustomReturnForActions[http.MethodGet], &secondDiskResponse)

	//POST Response (201 return code for accepted)
	headers := map[string]string{"Location": "/redfish/v1/TaskService/Tasks/JID_1234567890"}
	postCall := responseBuilder.Status("202 ACCEPTED").StatusCode(202).Body("").Headers(headers).Build()
	testClient.CustomReturnForActions[http.MethodPost] = append(testClient.CustomReturnForActions[http.MethodPatch], &postCall)

	//Set mocked response for GetTask()
	sucessfulTask := responseBuilder.Status("200 OK").StatusCode(200).Body(successfulTask).Build()
	testClient.CustomReturnForActions[http.MethodGet] = append(testClient.CustomReturnForActions[http.MethodGet], &sucessfulTask)

	volumesCollection := responseBuilder.Status("200 OK").StatusCode(200).Body(volumesRedfishJSON).Build()
	testClient.CustomReturnForActions[http.MethodGet] = append(testClient.CustomReturnForActions[http.MethodGet], &volumesCollection)

	specificVolume := responseBuilder.Status("200 OK").StatusCode(200).Body(volumeRedfishJSON).Build()
	testClient.CustomReturnForActions[http.MethodGet] = append(testClient.CustomReturnForActions[http.MethodGet], &specificVolume)

	diags := createRedfishStorageVolume(service, d)
	if diags.HasError() || len(d.Id()) == 0 { //If there are errors or ID has not been set, test FAILS
		t.Errorf("FAILED - First test - Create a volume with volume_type included in controller OperationApplyTime and POST succeed")
	}

	//Second test - Create a volume with volume_type NOT INCLUDED in controller OperationApplyTime
	configMap = map[string]interface{}{
		"storage_controller_id": "RAID.Integrated.1-1",
		"volume_name":           "MyVol",
		"volume_type":           "Mirrored",
		"volume_disks":          []interface{}{"Physical Disk 0:1:0", "Physical Disk 0:1:1"},
		"settings_apply_time":   "MadeUpSettingApplyTime",
	}

	d = schema.TestResourceDataRaw(t, getResourceStorageVolumeSchema(), configMap)
	testClient = &common.TestClient{}
	responseBuilder = NewResponseBuilder()

	//Get mocked client
	service, err = setStorageMockedClient(testClient, responseBuilder)
	if err != nil {
		t.Errorf("Error when creating the mocked client: %v", err)
	}

	//Get operationApplyTimes
	volumesEmptyCollection = responseBuilder.Status("200 OK").StatusCode(200).Body(noVolumesRedfishJSON).Build()
	testClient.CustomReturnForActions[http.MethodGet] = append(testClient.CustomReturnForActions[http.MethodGet], &volumesEmptyCollection)

	diags = createRedfishStorageVolume(service, d)
	if !diags.HasError() || len(d.Id()) != 0 { //If there are errors or ID has not been set, test FAILS
		t.Errorf("FAILED - Second test - Create a volume with volume_type NOT INCLUDED in controller OperationApplyTime")
	}
}

func setStorageMockedClient(testClient *common.TestClient, responseBuilder *responseBuilder) (*gofish.Service, error) {
	testClient.CustomReturnForActions = make(map[string][]interface{})

	rootResponse := responseBuilder.Status("200 OK").StatusCode(200).Body(rootRedfishJSON).Build()
	testClient.CustomReturnForActions[http.MethodGet] = append(testClient.CustomReturnForActions[http.MethodGet], &rootResponse)

	//Get mocked service
	service, err := gofish.ServiceRoot(testClient)
	if err != nil {
		return nil, err
	}

	//service.Systems() will make 1 + N GET calls (where N is the number of systems, normally just one).
	//	- First one to get the system collection
	//	- The N following correspond to the number of systems embedded (Normally just one)
	systemsResponse := responseBuilder.Status("200 OK").StatusCode(200).Body(systemsRedfishJSON).Build()
	testClient.CustomReturnForActions[http.MethodGet] = append(testClient.CustomReturnForActions[http.MethodGet], &systemsResponse)

	embeddedSystemResponse := responseBuilder.Status("200 OK").StatusCode(200).Body(systemEmbeddedRedfishJSON).Build()
	testClient.CustomReturnForActions[http.MethodGet] = append(testClient.CustomReturnForActions[http.MethodGet], &embeddedSystemResponse)

	//This mocked storage has 3 different controllers
	storageResponse := responseBuilder.Status("200 OK").StatusCode(200).Body(storageRedfishJSON).Build()
	testClient.CustomReturnForActions[http.MethodGet] = append(testClient.CustomReturnForActions[http.MethodGet], &storageResponse)

	firstStorageResponse := responseBuilder.Status("200 OK").StatusCode(200).Body(storage1RedfishJSON).Build()
	testClient.CustomReturnForActions[http.MethodGet] = append(testClient.CustomReturnForActions[http.MethodGet], &firstStorageResponse)

	secondStorageResponse := responseBuilder.Status("200 OK").StatusCode(200).Body(storage2RedfishJSON).Build()
	testClient.CustomReturnForActions[http.MethodGet] = append(testClient.CustomReturnForActions[http.MethodGet], &secondStorageResponse)

	thirdStorageResponse := responseBuilder.Status("200 OK").StatusCode(200).Body(storage3RedfishJSON).Build()
	testClient.CustomReturnForActions[http.MethodGet] = append(testClient.CustomReturnForActions[http.MethodGet], &thirdStorageResponse)

	return service, nil
}
