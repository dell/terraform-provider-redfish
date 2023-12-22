package dell

// Entity provides the common basis for dell and gofish objects
type Entity struct {
	ODataContext string `json:"@odata.context"`
	ODataID      string `json:"@odata.id"`
	ODataType    string `json:"@odata.type"`
	ID           string `json:"Id"`
	Name         string
	Description  string
}
