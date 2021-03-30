package redfish

import (
	"context"
	"net/http"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stmcginnis/gofish"
	"github.com/stmcginnis/gofish/common"
)

func TestCreateRedfishVirtualMedia(t *testing.T) {
	var testClient *common.TestClient
	var responseBuilder *responseBuilder

	//Set mocked configmap
	configMap := map[string]interface{}{
		"virtual_media_id": "RemovableDisk",
		"image":            "http://my-mocked-server.org/centos7.iso",
	}

	//FIRST TEST - Create virtual media using VM that exists and post succeed
	d := schema.TestResourceDataRaw(t, getResourceRedfishVirtualMediaSchema(), configMap)
	testClient = &common.TestClient{}
	responseBuilder = NewResponseBuilder()
	service, err := setVirtualMediaMockedClientForCreate(testClient, responseBuilder)
	if err != nil {
		t.Errorf("Error when creating the mocked client: %v", err)
	}

	//Mocked POST response OK
	resultPostResponse := responseBuilder.Status("200 OK").StatusCode(200).Body("").Build()
	testClient.CustomReturnForActions[http.MethodPost] = append(testClient.CustomReturnForActions[http.MethodPost], &resultPostResponse)

	diags := createRedfishVirtualMedia(service, d)
	//Check if ID has been set
	if diags.HasError() || len(d.Id()) == 0 {
		t.Errorf("FAILED - First test - Create VirtualMedia and POST succeed")
	}

	//SECOND TEST - Try to create a virtual media using a VM that doesn't exist
	//Set mocked configmap
	configMap = map[string]interface{}{
		"virtual_media_id": "WhateverMedia",
		"image":            "http://my-mocked-server.org/centos7.iso",
	}
	d = schema.TestResourceDataRaw(t, getResourceRedfishVirtualMediaSchema(), configMap)
	testClient = &common.TestClient{}
	responseBuilder = NewResponseBuilder()
	service, err = setVirtualMediaMockedClientForCreate(testClient, responseBuilder)
	if err != nil {
		t.Errorf("Error when creating the mocked client: %v", err)
	}

	diags = createRedfishVirtualMedia(service, d)
	//Check if ID has been set
	if !diags.HasError() || len(d.Id()) > 0 {
		t.Errorf("FAILED - Second test - Try to create a virtual media using a VM that doesn't exist")
	}

	//THIRD TEST - Create virtual media using VM that exists and post failed
	configMap = map[string]interface{}{
		"virtual_media_id": "RemovableDisk",
		"image":            "http://my-mocked-server.org/centos7.iso",
	}
	d = schema.TestResourceDataRaw(t, getResourceRedfishVirtualMediaSchema(), configMap)
	testClient = &common.TestClient{}
	responseBuilder = NewResponseBuilder()
	service, err = setVirtualMediaMockedClientForCreate(testClient, responseBuilder)
	if err != nil {
		t.Errorf("Error when creating the mocked client: %v", err)
	}

	//Mocked POST response OK
	resultPostResponse = responseBuilder.Status("503 SERVER ERROR").StatusCode(503).Body("").Build()
	testClient.CustomReturnForActions[http.MethodPost] = append(testClient.CustomReturnForActions[http.MethodPost], &resultPostResponse)

	diags = createRedfishVirtualMedia(service, d)
	//Check if ID has been set
	if !diags.HasError() || len(d.Id()) > 0 {
		t.Errorf("FAILED - Third test - Create virtual media using VM that exists and post failed")
	}
}

func TestReadRedfishVirtualMedia(t *testing.T) {
	//FIRST TEST - Read VirtualMedia that is connected and exists in the state file
	configMap := map[string]interface{}{
		"virtual_media_id": "RemovableDisk",
		"image":            "http://my-mocked-server.org/centos7.iso",
	}
	d := schema.TestResourceDataRaw(t, getResourceRedfishVirtualMediaSchema(), configMap)
	testClient := &common.TestClient{}
	responseBuilder := NewResponseBuilder()
	service, err := setVirtualMediaMockedClient(testClient, responseBuilder, true)
	if err != nil {
		t.Errorf("Error when creating the mocked client: %v", err)
	}
	d.SetId("/redfish/v1/Managers/iDRAC.Embedded.1/VirtualMedia/RemovableDisk")
	diags := readRedfishVirtualMedia(service, d)
	//Check if ID is kept
	if diags.HasError() || len(d.Id()) == 0 {
		t.Errorf("FAILED - First test - Read VirtualMedia that is connected and exists in the state file")
	}

	//SECOND TEST - Read a VirtualMedia that is desconnected and exists in the state file
	testClient = &common.TestClient{}
	responseBuilder = NewResponseBuilder()
	service, err = setVirtualMediaMockedClient(testClient, responseBuilder, false)
	if err != nil {
		t.Errorf("Error when creating the mocked client: %v", err)
	}
	d.SetId("/redfish/v1/Managers/iDRAC.Embedded.1/VirtualMedia/RemovableDisk")
	diags = readRedfishVirtualMedia(service, d)
	//Check if ID is deleted
	if diags.HasError() || len(d.Id()) > 0 {
		t.Errorf("FAILED - Second test - Read a VirtualMedia that is desconnected and exists in the state file")
	}
}

func TestUpdateRedfishVirtualMedia(t *testing.T) {
	//Create fake context
	ctx := context.TODO()

	//FIRST TEST - Unmount and mount virtual media
	configMap := map[string]interface{}{
		"virtual_media_id": "RemovableDisk",
		"image":            "http://my-mocked-server.org/centos7.iso",
	}
	d := schema.TestResourceDataRaw(t, getResourceRedfishVirtualMediaSchema(), configMap)
	testClient := &common.TestClient{}
	responseBuilder := NewResponseBuilder()
	service, err := setVirtualMediaMockedClient(testClient, responseBuilder, true)
	if err != nil {
		t.Errorf("Error when creating the mocked client: %v", err)
	}

	ejectMediaPostResponse := responseBuilder.Status("200 OK").StatusCode(200).Body("").Build()
	testClient.CustomReturnForActions[http.MethodPost] = append(testClient.CustomReturnForActions[http.MethodPost], &ejectMediaPostResponse)
	mountMediaPostResponse := responseBuilder.Status("200 OK").StatusCode(200).Body("").Build()
	testClient.CustomReturnForActions[http.MethodPost] = append(testClient.CustomReturnForActions[http.MethodPost], &mountMediaPostResponse)

	diags := updateRedfishVirtualMedia(ctx, service, d, nil)
	if diags.HasError() {
		t.Errorf("FAILED - Second test - Read a VirtualMedia that is desconnected and exists in the state file")
	}
}

func TestDeleteRedfishVirtualMedia(t *testing.T) {
	//FIRST TEST - Unmount virtual media
	configMap := map[string]interface{}{
		"virtual_media_id": "RemovableDisk",
		"image":            "http://my-mocked-server.org/centos7.iso",
	}
	d := schema.TestResourceDataRaw(t, getResourceRedfishVirtualMediaSchema(), configMap)
	testClient := &common.TestClient{}
	responseBuilder := NewResponseBuilder()
	service, err := setVirtualMediaMockedClient(testClient, responseBuilder, true)
	if err != nil {
		t.Errorf("Error when creating the mocked client: %v", err)
	}

	ejectMediaPostResponse := responseBuilder.Status("200 OK").StatusCode(200).Body("").Build()
	testClient.CustomReturnForActions[http.MethodPost] = append(testClient.CustomReturnForActions[http.MethodPost], &ejectMediaPostResponse)

	d.SetId("/redfish/v1/Managers/iDRAC.Embedded.1/VirtualMedia/RemovableDisk")
	diags := deleteRedfishVirtualMedia(service, d)

	if diags.HasError() {
		t.Errorf("FAILED - First test - Unmount virtual media")
	}

}

func setVirtualMediaMockedClientForCreate(testClient *common.TestClient, responseBuilder *responseBuilder) (*gofish.Service, error) {
	testClient.CustomReturnForActions = make(map[string][]interface{})
	//Add rootResponse to map
	rootResponse := responseBuilder.Status("200 OK").StatusCode(200).Body(rootRedfishJSON).Build()
	testClient.CustomReturnForActions[http.MethodGet] = append(testClient.CustomReturnForActions[http.MethodGet], &rootResponse)

	//Get mocked service
	service, err := gofish.ServiceRoot(testClient)
	if err != nil {
		return nil, err
	}

	//Add manager collection
	managerCollectionResponse := responseBuilder.Status("200 OK").StatusCode(200).Body(managerCollection).Build()
	testClient.CustomReturnForActions[http.MethodGet] = append(testClient.CustomReturnForActions[http.MethodGet], &managerCollectionResponse)

	//Add idrac manager
	idracManagerResponse := responseBuilder.Status("200 OK").StatusCode(200).Body(idracManager).Build()
	testClient.CustomReturnForActions[http.MethodGet] = append(testClient.CustomReturnForActions[http.MethodGet], &idracManagerResponse)

	//Add virtual media collection
	virtualMediaCollectionResponse := responseBuilder.Status("200 OK").StatusCode(200).Body(idracVirtualMediaCollection).Build()
	testClient.CustomReturnForActions[http.MethodGet] = append(testClient.CustomReturnForActions[http.MethodGet], &virtualMediaCollectionResponse)

	//Add virtual media removable disk
	removableDiskResponse := responseBuilder.Status("200 OK").StatusCode(200).Body(virtualMediaRemovableDiskDisconnected).Build()
	testClient.CustomReturnForActions[http.MethodGet] = append(testClient.CustomReturnForActions[http.MethodGet], &removableDiskResponse)

	return service, nil
}

func setVirtualMediaMockedClient(testClient *common.TestClient, responseBuilder *responseBuilder, connected bool) (*gofish.Service, error) {
	testClient.CustomReturnForActions = make(map[string][]interface{})

	//Add rootResponse to map
	rootResponse := responseBuilder.Status("200 OK").StatusCode(200).Body(rootRedfishJSON).Build()
	testClient.CustomReturnForActions[http.MethodGet] = append(testClient.CustomReturnForActions[http.MethodGet], &rootResponse)

	//Get mocked service
	service, err := gofish.ServiceRoot(testClient)
	if err != nil {
		return nil, err
	}

	if connected {
		removableDiskResponse := responseBuilder.Status("200 OK").StatusCode(200).Body(virtualMediaRemovableDiskConnected).Build()
		testClient.CustomReturnForActions[http.MethodGet] = append(testClient.CustomReturnForActions[http.MethodGet], &removableDiskResponse)
	} else {
		removableDiskResponse := responseBuilder.Status("200 OK").StatusCode(200).Body(virtualMediaRemovableDiskDisconnected).Build()
		testClient.CustomReturnForActions[http.MethodGet] = append(testClient.CustomReturnForActions[http.MethodGet], &removableDiskResponse)
	}

	return service, nil
}
