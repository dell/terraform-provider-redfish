package redfish

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stmcginnis/gofish/common"
)

func TestReadRedfishVirtualMediaCollection(t *testing.T) {
	//FIRST TEST - Read VirtualMedia collection
	resultMap := map[string]string{
		"RemovableDisk": "",
	}
	configMap := map[string]interface{}{} //Empty map
	d := schema.TestResourceDataRaw(t, getDataSourceRedfishVirtualMediaSchema(), configMap)
	testClient := &common.TestClient{}
	responseBuilder := NewResponseBuilder()

	service, err := setVirtualMediaMockedClientForCreate(testClient, responseBuilder) //Function reused from resource_redfish_virtual_media_test.go
	if err != nil {
		t.Errorf("Error when creating the mocked client: %v", err)
	}

	diags := readRedfishVirtualMediaCollection(service, d)
	if diags.HasError() {
		t.Errorf("FIRST TEST - Read VirtualMedia collection failed")
	}

	//Perform checks
	vm := d.Get("virtual_media").([]interface{})
	for _, v := range vm {
		w := v.(map[string]interface{})
		if _, e := resultMap[w["id"].(string)]; !e {
			t.Errorf("FIRST TEST - Got value doesn't match with test value")
		}
	}
}
