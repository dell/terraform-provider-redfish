package redfish

import (
	"github.com/stmcginnis/gofish"
	"github.com/stmcginnis/gofish/common"
	"github.com/stmcginnis/gofish/redfish"
	"net/http"
	"testing"
)

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
		} else {
			if err == nil {
				t.Errorf("Test number %v passed when it was supposed to fail", v.noTest)
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
	cases := []struct {
		noTest           int
		postResponseCode int
		location         string
		shouldPass       bool
	}{
		{1, http.StatusAccepted, "/redfish/v1/TaskService/Tasks/JID_031156904278", true},
		{2, http.StatusNotFound, "", false},
		{3, http.StatusAccepted, "", false},
	}

	for _, v := range cases {
		testClient := &common.TestClient{}
		testClient.CustomReturnForActions = make(map[string][]interface{})
		responseBuilder := &responseBuilder{}
		postResponse := responseBuilder.Body("").Status("TEST").StatusCode(v.postResponseCode).Headers(map[string]string{"Location": v.location}).Build()
		testClient.CustomReturnForActions[http.MethodDelete] = append(testClient.CustomReturnForActions[http.MethodDelete], &postResponse)
		_, err := deleteVolume(testClient, "TEST")
		if v.shouldPass {
			if err != nil {
				t.Errorf("Test number %v failed %v", v.noTest, err)
			}
		} else {
			if err == nil {
				t.Errorf("Test number %v passed when it was supposed to fail", v.noTest)
			}
		}
	}
}

func TestGetDrives(t *testing.T) {
	/*
		Possible cases:
			- All drives passed exist
			- One or more drives does not exist
			- Errors with the Client (not sure if we can check that with the mocked client)
	*/
	cases := []struct {
		noTest     int
		drives     []string
		shouldPass bool
	}{
		{1, []string{"Physical Disk 0:1:0", "Physical Disk 0:1:1"}, true},
		{2, []string{"Physical Disk 0:2:0", "Physical Disk 0:2:1"}, false},
		{3, []string{"Physical Disk 0:1:0", "Physical Disk 0:1:1", "Physical Disk 0:1:2"}, false},
	}
	for _, v := range cases {
		storage, err := setStorageMockedClient("storage:drives", "")
		if err != nil {
			t.Errorf("There was an error with the mocked client")
		}
		_, err = getDrives(storage.(*redfish.Storage), v.drives)
		if v.shouldPass {
			if err != nil {
				t.Errorf("Test number %v failed %v", v.noTest, err)
			}
		} else {
			if err == nil {
				t.Errorf("Test number %v passed when it was supposed to fail", v.noTest)
			}
		}
	}
}

func TestCreateVolume(t *testing.T) {
	/*
		Possible cases:
			- The volume is created successfully (HTTP 202 Accepted)
			- The was an issue when creating the volume (HTTP return code different from 202 Accepted)
			- Errors with the Client (not sure if we can check that with the mocked client)
	*/
	cases := []struct {
		noTest           int
		postResponseCode int
		location         string
		shouldPass       bool
	}{
		{1, http.StatusAccepted, "/redfish/v1/TaskService/Tasks/JID_031156904278", true},
		{2, http.StatusNotFound, "", false},
		{3, http.StatusAccepted, "", false},
	}
	for _, v := range cases {
		testClient := &common.TestClient{}
		testClient.CustomReturnForActions = make(map[string][]interface{})
		responseBuilder := &responseBuilder{}
		postResponse := responseBuilder.Body("").Status("TEST").StatusCode(v.postResponseCode).Headers(map[string]string{"Location": v.location}).Build()
		testClient.CustomReturnForActions[http.MethodPost] = append(testClient.CustomReturnForActions[http.MethodPost], &postResponse)
		_, err := createVolume(testClient, "/redfish/v1/test", "Mirrored", "HelloWorld", []*redfish.Drive{{}, {}}, "Immediate")
		if v.shouldPass {
			if err != nil {
				t.Errorf("Test number %v failed %v", v.noTest, err)
			}
		} else {
			if err == nil {
				t.Errorf("Test number %v passed when it was supposed to fail", v.noTest)
			}
		}
	}
}

func TestGetVolumeID(t *testing.T) {
	/*
		Possible cases:
			- Volume exist
			- Volume does not exist
			- Errors with the Client (not sure if we can check that with the mocked client)
	*/
	cases := []struct {
		noTest         int
		volumeName     string
		responseOption string
		shouldPass     bool
	}{
		{1, "MyVol", "included", true},
		{2, "MyVol", "empty", false},
		{3, "MyOverpoweredVolume", "included", false},
		{4, "MyOverpoweredVolume", "empty", false},
	}

	for _, v := range cases {
		storage, err := setStorageMockedClient("storage:volumes", v.responseOption)
		if err != nil {
			t.Errorf("There was an error with the mocked client")
		}
		_, err = getVolumeID(storage.(*redfish.Storage), v.volumeName)
		if v.shouldPass {
			if err != nil {
				t.Errorf("Test number %v failed %v", v.noTest, err)
			}
		} else {
			if err == nil {
				t.Errorf("Test number %v passed when it was supposed to fail", v.noTest)
			}
		}
	}
}
