package models

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// StorageDatasource is struct for storage data-source
type StorageDatasource struct {
	ID            types.String            `tfsdk:"id"`
	RedfishServer RedfishServer           `tfsdk:"redfish_server"`
	Storages      []StorageControllerData `tfsdk:"storage"`
}

// StorageControllerData is struct for data of a storage controller
type StorageControllerData struct {
	ID     types.String   `tfsdk:"storage_controller_id"`
	Drives []types.String `tfsdk:"drives"`
}
