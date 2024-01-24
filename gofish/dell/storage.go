package dell

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/stmcginnis/gofish/redfish"
)

type Storage struct {
	parent redfish.Storage
	Params StorageParams
}

type StorageParams struct {
	Controllers Link
	Drives      []Link
	Volumes     Link
}

func (root *Storage) UnmarshalJSON(b []byte) error {
	errRoot := json.Unmarshal(b, &root.parent)
	errParams := json.Unmarshal(b, &root.Params)
	return errors.Join(errRoot, errParams)
}

func (root *System) GetStorages() ([]Storage, error) {
	var storageCollection Collection
	err := root.parent.Get(root.parent.GetClient(),
		fmt.Sprintf("%s?%s", string(root.Params.Storage), expand),
		&storageCollection)
	if err != nil {
		return nil, err
	}
	var ret []Storage
	if err := json.Unmarshal(storageCollection.Members, &ret); err != nil {
		return nil, err
	}
	for i := range ret {
		ret[i].parent.SetClient(root.parent.GetClient())
	}
	return ret, err
}
