package redfish

import (
	"net/http"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stmcginnis/gofish/common"
)

func TestReadRedfishStorageCollection(t *testing.T) {
	//FIRST TEST - Read Storage Volume collection
	resultStorageMap := map[string]string{
		"RAID.Integrated.1-1": "",
	}
	resultDisksMap := map[string]string{
		"Physical Disk 0:1:0": "",
		"Physical Disk 0:1:1": "",
	}
	configMap := map[string]interface{}{} //Empty map
	d := schema.TestResourceDataRaw(t, getDataSourceRedfishStorageSchema(), configMap)
	testClient := &common.TestClient{}
	responseBuilder := NewResponseBuilder()

	service, err := setStorageMockedClient(testClient, responseBuilder) //Function reused from resource_redfish_storage_volume_test.go
	if err != nil {
		t.Errorf("Error when creating the mocked client: %v", err)
	}

	//Add two disks to testClient
	firstDiskResponse := responseBuilder.Status("200 OK").StatusCode(200).Body(drive1RedfishJSON).Build()
	testClient.CustomReturnForActions[http.MethodGet] = append(testClient.CustomReturnForActions[http.MethodGet], &firstDiskResponse)

	secondDiskResponse := responseBuilder.Status("200 OK").StatusCode(200).Body(drive2RedfishJSON).Build()
	testClient.CustomReturnForActions[http.MethodGet] = append(testClient.CustomReturnForActions[http.MethodGet], &secondDiskResponse)

	diags := readRedfishStorageCollection(service, d)
	if diags.HasError() {
		t.Errorf("FIRST TEST - Read Storage Volume collection failed")
	}

	//Check result
	storage := d.Get("storage").([]interface{})
	for _, v := range storage {
		w := v.(map[string]interface{})
		//Check controller
		if _, e := resultStorageMap[w["storage_controller_id"].(string)]; !e {
			t.Errorf("FIRST TEST - Read Storage Volume collection failed. Got storage different than expected")
		}
		//Check disks
		disks := w["drives"].([]interface{})
		for _, d := range disks {
			disk := d.(string)
			if _, e := resultDisksMap[disk]; !e {
				t.Errorf("FIRST TEST - Read Storage Volume collection failed. Got disk different than expected")
			}
		}
	}

}
