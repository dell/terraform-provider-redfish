package redfish

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stmcginnis/gofish"
	"github.com/stmcginnis/gofish/common"
	"net/http"
	"testing"
)

func TestCreateRedfishUserAccount(t *testing.T) {
	var testClient *common.TestClient
	var responseBuilder *responseBuilder

	//Set mocked configmap
	configMap := map[string]interface{}{
		"username": "mike",
		"password": "test",
		"enabled":  true,
		"role_id":  "None",
	}

	//First test - Create user when there is space and patch succeed
	d := schema.TestResourceDataRaw(t, getResourceUserAccountSchema(), configMap)
	testClient = &common.TestClient{}
	responseBuilder = NewResponseBuilder()
	//Get mocked client
	service, err := setUserAccountMockedClient(testClient, responseBuilder)
	if err != nil {
		t.Errorf("Error when creating the mocked client: %v", err)
	}
	user1 := responseBuilder.Status("200 OK").StatusCode(200).Body(managerAccount1).Build()
	testClient.CustomReturnForActions[http.MethodGet] = append(testClient.CustomReturnForActions[http.MethodGet], &user1)
	user2 := responseBuilder.Status("200 OK").StatusCode(200).Body(managerAccount2).Build()
	testClient.CustomReturnForActions[http.MethodGet] = append(testClient.CustomReturnForActions[http.MethodGet], &user2)
	user3 := responseBuilder.Status("200 OK").StatusCode(200).Body(managerAccountEmpty).Build()
	testClient.CustomReturnForActions[http.MethodGet] = append(testClient.CustomReturnForActions[http.MethodGet], &user3)
	//Path will return success
	pathCall := responseBuilder.Status("200 OK").StatusCode(200).Body("").Build()
	testClient.CustomReturnForActions[http.MethodPatch] = append(testClient.CustomReturnForActions[http.MethodPatch], &pathCall)

	diags := createRedfishUserAccount(service, d)
	if diags.HasError() || len(d.Id()) == 0 { //If there are errors or ID has not been set, test FAILS
		t.Errorf("FAILED -First test - Create user when there is space and patch succeed")
	}

	//Second test - Create user when there is space and patch fails
	d = schema.TestResourceDataRaw(t, getResourceUserAccountSchema(), configMap)
	testClient = &common.TestClient{}
	responseBuilder = NewResponseBuilder()
	//Get mocked client
	service, err = setUserAccountMockedClient(testClient, responseBuilder)
	if err != nil {
		t.Errorf("Error when creating the mocked client: %v", err)
	}
	user1 = responseBuilder.Status("200 OK").StatusCode(200).Body(managerAccount1).Build()
	testClient.CustomReturnForActions[http.MethodGet] = append(testClient.CustomReturnForActions[http.MethodGet], &user1)
	user2 = responseBuilder.Status("200 OK").StatusCode(200).Body(managerAccount2).Build()
	testClient.CustomReturnForActions[http.MethodGet] = append(testClient.CustomReturnForActions[http.MethodGet], &user2)
	user3 = responseBuilder.Status("200 OK").StatusCode(200).Body(managerAccountEmpty).Build()
	testClient.CustomReturnForActions[http.MethodGet] = append(testClient.CustomReturnForActions[http.MethodGet], &user3)
	//Path will return Error
	pathCall = responseBuilder.Status("503 Error").StatusCode(503).Body("").Build()
	testClient.CustomReturnForActions[http.MethodPatch] = append(testClient.CustomReturnForActions[http.MethodPatch], &pathCall)

	diags = createRedfishUserAccount(service, d)
	if !diags.HasError() || len(d.Id()) > 0 {
		t.Errorf("FAILED - Second test - Create user when there is space and patch fails")
	}

	//Third test - Create user when there is no space
	d = schema.TestResourceDataRaw(t, getResourceUserAccountSchema(), configMap)
	testClient = &common.TestClient{}
	responseBuilder = NewResponseBuilder()
	//Get mocked client
	service, err = setUserAccountMockedClient(testClient, responseBuilder)
	if err != nil {
		t.Errorf("Error when creating the mocked client: %v", err)
	}
	user1 = responseBuilder.Status("200 OK").StatusCode(200).Body(managerAccount1).Build()
	testClient.CustomReturnForActions[http.MethodGet] = append(testClient.CustomReturnForActions[http.MethodGet], &user1)
	user2 = responseBuilder.Status("200 OK").StatusCode(200).Body(managerAccount2).Build()
	testClient.CustomReturnForActions[http.MethodGet] = append(testClient.CustomReturnForActions[http.MethodGet], &user2)
	user3 = responseBuilder.Status("200 OK").StatusCode(200).Body(managerAccountTest).Build()
	testClient.CustomReturnForActions[http.MethodGet] = append(testClient.CustomReturnForActions[http.MethodGet], &user3)

	diags = createRedfishUserAccount(service, d)
	if !diags.HasError() || len(d.Id()) > 0 {
		t.Errorf("FAILED - Third test - Create user when there is no space")
	}
}

func TestReadRedfishUserAccount(t *testing.T) {
	var testClient *common.TestClient
	var responseBuilder *responseBuilder

	//Set mocked configmap
	configMap := map[string]interface{}{
		"username": "test",
		"password": "test",
		"enabled":  true,
		"role_id":  "None",
	}

	//First test - Read user with id that doesnt exists
	d := schema.TestResourceDataRaw(t, getResourceUserAccountSchema(), configMap)
	d.SetId("3") //State file would have 3 as id
	testClient = &common.TestClient{}
	responseBuilder = NewResponseBuilder()
	//Get mocked client
	service, err := setUserAccountMockedClient(testClient, responseBuilder)
	if err != nil {
		t.Errorf("Error when creating the mocked client: %v", err)
	}
	user1 := responseBuilder.Status("200 OK").StatusCode(200).Body(managerAccount1).Build()
	testClient.CustomReturnForActions[http.MethodGet] = append(testClient.CustomReturnForActions[http.MethodGet], &user1)
	user2 := responseBuilder.Status("200 OK").StatusCode(200).Body(managerAccount2).Build()
	testClient.CustomReturnForActions[http.MethodGet] = append(testClient.CustomReturnForActions[http.MethodGet], &user2)
	user3 := responseBuilder.Status("200 OK").StatusCode(200).Body(managerAccountEmpty).Build()
	testClient.CustomReturnForActions[http.MethodGet] = append(testClient.CustomReturnForActions[http.MethodGet], &user3)

	diags := readRedfishUserAccount(service, d)
	if diags.HasError() || len(d.Id()) > 0 {
		t.Errorf("FAILED - Read user with id that doesnt exists")
	}

	//Second test - Read user with id that exists
	d = schema.TestResourceDataRaw(t, getResourceUserAccountSchema(), configMap)
	d.SetId("3")
	testClient = &common.TestClient{}
	responseBuilder = NewResponseBuilder()
	//Get mocked client
	service, err = setUserAccountMockedClient(testClient, responseBuilder)
	if err != nil {
		t.Errorf("Error when creating the mocked client: %v", err)
	}
	user1 = responseBuilder.Status("200 OK").StatusCode(200).Body(managerAccount1).Build()
	testClient.CustomReturnForActions[http.MethodGet] = append(testClient.CustomReturnForActions[http.MethodGet], &user1)
	user2 = responseBuilder.Status("200 OK").StatusCode(200).Body(managerAccount2).Build()
	testClient.CustomReturnForActions[http.MethodGet] = append(testClient.CustomReturnForActions[http.MethodGet], &user2)
	user3 = responseBuilder.Status("200 OK").StatusCode(200).Body(managerAccountTest).Build()
	testClient.CustomReturnForActions[http.MethodGet] = append(testClient.CustomReturnForActions[http.MethodGet], &user3)

	diags = readRedfishUserAccount(service, d)
	if diags.HasError() || d.Get("username").(string) != "test" {
		t.Errorf("FAILED - Second test - Read user with id that exists")
	}
}

func TestUpdateRedfishUserAccount(t *testing.T) {
	var testClient *common.TestClient
	var responseBuilder *responseBuilder
	//Create fake context
	ctx := context.TODO()

	//Set mocked configmap
	configMap := map[string]interface{}{
		"username": "test2",
		"password": "test",
		"enabled":  true,
		"role_id":  "None",
	}

	//First test - Update user with different attributes than expected - PATCH succeed
	d := schema.TestResourceDataRaw(t, getResourceUserAccountSchema(), configMap)
	d.SetId("3")
	testClient = &common.TestClient{}
	responseBuilder = NewResponseBuilder()
	//Get mocked client
	service, err := setUserAccountMockedClient(testClient, responseBuilder)
	if err != nil {
		t.Errorf("Error when creating the mocked client: %v", err)
	}
	user1 := responseBuilder.Status("200 OK").StatusCode(200).Body(managerAccount1).Build()
	testClient.CustomReturnForActions[http.MethodGet] = append(testClient.CustomReturnForActions[http.MethodGet], &user1)
	user2 := responseBuilder.Status("200 OK").StatusCode(200).Body(managerAccount2).Build()
	testClient.CustomReturnForActions[http.MethodGet] = append(testClient.CustomReturnForActions[http.MethodGet], &user2)
	user3 := responseBuilder.Status("200 OK").StatusCode(200).Body(managerAccountTest).Build()
	testClient.CustomReturnForActions[http.MethodGet] = append(testClient.CustomReturnForActions[http.MethodGet], &user3)
	//Path will return success
	pathCall := responseBuilder.Status("200 OK").StatusCode(200).Body("").Build()
	testClient.CustomReturnForActions[http.MethodPatch] = append(testClient.CustomReturnForActions[http.MethodPatch], &pathCall)

	diags := updateRedfishUserAccount(ctx, service, d, nil)
	if diags.HasError() || d.Get("username").(string) != "test2" {
		t.Errorf("FAILED - First test - Update user with different attributes than expected")
	}

	//Second test - Update user with different attributes than expected - PATCH fails
	d = schema.TestResourceDataRaw(t, getResourceUserAccountSchema(), configMap)
	d.SetId("3")
	testClient = &common.TestClient{}
	responseBuilder = NewResponseBuilder()
	//Get mocked client
	service, err = setUserAccountMockedClient(testClient, responseBuilder)
	if err != nil {
		t.Errorf("Error when creating the mocked client: %v", err)
	}
	user1 = responseBuilder.Status("200 OK").StatusCode(200).Body(managerAccount1).Build()
	testClient.CustomReturnForActions[http.MethodGet] = append(testClient.CustomReturnForActions[http.MethodGet], &user1)
	user2 = responseBuilder.Status("200 OK").StatusCode(200).Body(managerAccount2).Build()
	testClient.CustomReturnForActions[http.MethodGet] = append(testClient.CustomReturnForActions[http.MethodGet], &user2)
	user3 = responseBuilder.Status("200 OK").StatusCode(200).Body(managerAccountTest).Build()
	testClient.CustomReturnForActions[http.MethodGet] = append(testClient.CustomReturnForActions[http.MethodGet], &user3)
	//Path will return Error
	pathCall = responseBuilder.Status("503 Error").StatusCode(503).Body("").Build()
	testClient.CustomReturnForActions[http.MethodPatch] = append(testClient.CustomReturnForActions[http.MethodPatch], &pathCall)

	diags = updateRedfishUserAccount(ctx, service, d, nil)
	if !diags.HasError() {
		t.Errorf("FAILED - Second test - Update user with different attributes than expected - PATCH fails")
	}
}

func TestDeleteRedfishUserAccount(t *testing.T) {
	var testClient *common.TestClient
	var responseBuilder *responseBuilder

	//Set mocked configmap
	configMap := map[string]interface{}{
		"username": "test",
		"password": "test",
		"enabled":  true,
		"role_id":  "None",
	}

	//First test - Delete user - PATCH Succeed
	d := schema.TestResourceDataRaw(t, getResourceUserAccountSchema(), configMap)
	d.SetId("3")
	testClient = &common.TestClient{}
	responseBuilder = NewResponseBuilder()
	//Get mocked client
	service, err := setUserAccountMockedClient(testClient, responseBuilder)
	if err != nil {
		t.Errorf("Error when creating the mocked client: %v", err)
	}
	user1 := responseBuilder.Status("200 OK").StatusCode(200).Body(managerAccount1).Build()
	testClient.CustomReturnForActions[http.MethodGet] = append(testClient.CustomReturnForActions[http.MethodGet], &user1)
	user2 := responseBuilder.Status("200 OK").StatusCode(200).Body(managerAccount2).Build()
	testClient.CustomReturnForActions[http.MethodGet] = append(testClient.CustomReturnForActions[http.MethodGet], &user2)
	user3 := responseBuilder.Status("200 OK").StatusCode(200).Body(managerAccountTest).Build()
	testClient.CustomReturnForActions[http.MethodGet] = append(testClient.CustomReturnForActions[http.MethodGet], &user3)
	//Path will return success
	pathCall := responseBuilder.Status("200 OK").StatusCode(200).Body("").Build()
	testClient.CustomReturnForActions[http.MethodPatch] = append(testClient.CustomReturnForActions[http.MethodPatch], &pathCall)
	diags := deleteRedfishUserAccount(service, d)
	if diags.HasError() || len(d.Id()) > 0 {
		t.Errorf("FAILED - First test - Delete user - PATCH Succeed")
	}

	//Second test - Delete user - PATCH fails
	d = schema.TestResourceDataRaw(t, getResourceUserAccountSchema(), configMap)
	d.SetId("3")
	testClient = &common.TestClient{}
	responseBuilder = NewResponseBuilder()
	//Get mocked client
	service, err = setUserAccountMockedClient(testClient, responseBuilder)
	if err != nil {
		t.Errorf("Error when creating the mocked client: %v", err)
	}
	user1 = responseBuilder.Status("200 OK").StatusCode(200).Body(managerAccount1).Build()
	testClient.CustomReturnForActions[http.MethodGet] = append(testClient.CustomReturnForActions[http.MethodGet], &user1)
	user2 = responseBuilder.Status("200 OK").StatusCode(200).Body(managerAccount2).Build()
	testClient.CustomReturnForActions[http.MethodGet] = append(testClient.CustomReturnForActions[http.MethodGet], &user2)
	user3 = responseBuilder.Status("200 OK").StatusCode(200).Body(managerAccountTest).Build()
	testClient.CustomReturnForActions[http.MethodGet] = append(testClient.CustomReturnForActions[http.MethodGet], &user3)
	//Path will return Error
	pathCall = responseBuilder.Status("503 Error").StatusCode(503).Body("").Build()
	testClient.CustomReturnForActions[http.MethodPatch] = append(testClient.CustomReturnForActions[http.MethodPatch], &pathCall)
	diags = deleteRedfishUserAccount(service, d)
	if !diags.HasError() {
		t.Errorf("FAILED - First test - Delete user - PATCH Succeed")
	}

}

func setUserAccountMockedClient(testClient *common.TestClient, responseBuilder *responseBuilder) (*gofish.Service, error) {
	testClient.CustomReturnForActions = make(map[string][]interface{})

	//Add rootResponse to map
	rootResponse := responseBuilder.Status("200 OK").StatusCode(200).Body(rootRedfishJSON).Build()
	testClient.CustomReturnForActions[http.MethodGet] = append(testClient.CustomReturnForActions[http.MethodGet], &rootResponse)

	//Get mocked service
	service, err := gofish.ServiceRoot(testClient)
	if err != nil {
		return nil, err
	}

	//Add accountServiceResponse
	accountServiceResponse := responseBuilder.Status("200 OK").StatusCode(200).Body(accountServiceRedfishJSON).Build()
	testClient.CustomReturnForActions[http.MethodGet] = append(testClient.CustomReturnForActions[http.MethodGet], &accountServiceResponse)

	//Add managerAccountCollection
	managerAccountResponse := responseBuilder.Status("200 OK").StatusCode(200).Body(managerAccountCollection).Build()
	testClient.CustomReturnForActions[http.MethodGet] = append(testClient.CustomReturnForActions[http.MethodGet], &managerAccountResponse)

	return service, nil
}
