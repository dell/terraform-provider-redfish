package dell

import (
	"encoding/json"

	"github.com/stmcginnis/gofish/redfish"
)

// Controller model to get controller data
type Controller struct {
	OdataContext                     string `json:"@odata.context"`
	OdataID                          string `json:"@odata.id"`
	OdataType                        string `json:"@odata.type"`
	AlarmState                       string `json:"AlarmState"`
	AutoConfigBehavior               string `json:"AutoConfigBehavior"`
	BootVirtualDiskFQDD              string `json:"BootVirtualDiskFQDD"`
	CacheSizeInMB                    int    `json:"CacheSizeInMB"`
	CachecadeCapability              string `json:"CachecadeCapability"`
	ConnectorCount                   int    `json:"ConnectorCount"`
	ControllerFirmwareVersion        string `json:"ControllerFirmwareVersion"`
	CurrentControllerMode            string `json:"CurrentControllerMode"`
	Description                      string `json:"Description"`
	Device                           string `json:"Device"`
	DeviceCardDataBusWidth           string `json:"DeviceCardDataBusWidth"`
	DeviceCardSlotLength             string `json:"DeviceCardSlotLength"`
	DeviceCardSlotType               string `json:"DeviceCardSlotType"`
	DriverVersion                    string `json:"DriverVersion"`
	EncryptionCapability             string `json:"EncryptionCapability"`
	EncryptionMode                   string `json:"EncryptionMode"`
	ID                               string `json:"Id"`
	KeyID                            string `json:"KeyID"`
	LastSystemInventoryTime          string `json:"LastSystemInventoryTime"`
	LastUpdateTime                   string `json:"LastUpdateTime"`
	MaxAvailablePCILinkSpeed         string `json:"MaxAvailablePCILinkSpeed"`
	MaxPossiblePCILinkSpeed          string `json:"MaxPossiblePCILinkSpeed"`
	Name                             string `json:"Name"`
	PCISlot                          string `json:"PCISlot"`
	PatrolReadState                  string `json:"PatrolReadState"`
	PersistentHotspare               string `json:"PersistentHotspare"`
	RealtimeCapability               string `json:"RealtimeCapability"`
	RollupStatus                     string `json:"RollupStatus"`
	SASAddress                       string `json:"SASAddress"`
	SecurityStatus                   string `json:"SecurityStatus"`
	SharedSlotAssignmentAllowed      string `json:"SharedSlotAssignmentAllowed"`
	SlicedVDCapability               string `json:"SlicedVDCapability"`
	SupportControllerBootMode        string `json:"SupportControllerBootMode"`
	SupportEnhancedAutoForeignImport string `json:"SupportEnhancedAutoForeignImport"`
	SupportRAID10UnevenSpans         string `json:"SupportRAID10UnevenSpans"`
	SupportsLKMtoSEKMTransition      string `json:"SupportsLKMtoSEKMTransition"`
	T10PICapability                  string `json:"T10PICapability"`
}

// ControllerBattery to get controller battery data
type ControllerBattery struct {
	OdataContext  string `json:"@odata.context"`
	OdataID       string `json:"@odata.id"`
	OdataType     string `json:"@odata.type"`
	Description   string `json:"Description"`
	Fqdd          string `json:"FQDD"`
	ID            string `json:"Id"`
	Name          string `json:"Name"`
	PrimaryStatus string `json:"PrimaryStatus"`
	RAIDState     string `json:"RAIDState"`
}

// StorageOEM to get storage oem data
type StorageOEM struct {
	OdataType             string                `json:"@odata.type"`
	DellController        Controller        `json:"DellController"`
	DellControllerBattery ControllerBattery `json:"DellControllerBattery"`
}

// UnmarshalJSON to unmarshal storage oem data
func (s *StorageOEM) UnmarshalJSON(data []byte) error {
	type temp StorageOEM
	type Dell struct {
		temp
	}
	var tempOEM struct {
		Dell Dell
	}

	err := json.Unmarshal(data, &tempOEM)
	if err != nil {
		return err
	}

	*s = StorageOEM(tempOEM.Dell.temp)
	return nil
}

// StorageExtended to extend the storage struct
type StorageExtended struct {
	Storage redfish.Storage
	OemData StorageOEM
}

// Storage utility function to extend the storage after unmarshalling
func Storage(storage *redfish.Storage) (*StorageExtended, error) {
	dellStorage := &StorageExtended{Storage: *storage, OemData: StorageOEM{}}
	var oemData StorageOEM
	err := json.Unmarshal(dellStorage.Storage.Oem, &oemData)
	if err != nil {
		return nil, err
	}
	dellStorage.OemData = oemData

	return dellStorage, nil
}
