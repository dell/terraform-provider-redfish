package dell

import (
	"encoding/json"

	"github.com/stmcginnis/gofish/redfish"
)

// UpdateServiceExtended struct extends the gofish UpdateService and includes Dell OEM actions
type UpdateServiceExtended struct {
	*redfish.UpdateService
	// Actions will hold all UpdateService Dell OEM actions
	Actions UpdateServiceActions
}

// UpdateServiceActions contains Dell OEM actions
type UpdateServiceActions struct {
	// DellUpdateServiceTarget is the URL to be targetted for Dell's update
	DellUpdateServiceTarget string
	// DellUpdateServiceInstallUpon are the installing times
	DellUpdateServiceInstallUpon []string
}

// UnmarshalJSON unmarshals Dell update service object from raw JSON
func (u *UpdateServiceActions) UnmarshalJSON(data []byte) error {
	type DellUpdateService struct {
		InstallUpon []string `json:"InstallUpon@Redfish.AllowableValues"`
		Target      string
	}
	var t struct {
		DellUpdateService DellUpdateService `json:"DellUpdateService.v1_0_0#DellUpdateService.Install"`
	}

	err := json.Unmarshal(data, &t)
	if err != nil {
		return err
	}

	u.DellUpdateServiceTarget = t.DellUpdateService.Target
	u.DellUpdateServiceInstallUpon = t.DellUpdateService.InstallUpon

	return nil
}

// UpdateService returns a Dell.UpdateServiceExtended pointer given a redfish.UpdateService pointer from gofish library
// This is the wrapper that extracts and parses Dell UpdateService OEM actions
func UpdateService(updateService *redfish.UpdateService) (*UpdateServiceExtended, error) {
	dellUpdate := UpdateServiceExtended{UpdateService: updateService}
	var oemUpdateService UpdateServiceActions

	err := json.Unmarshal(dellUpdate.OemActions, &oemUpdateService)
	if err != nil {
		return nil, err
	}
	dellUpdate.Actions = oemUpdateService

	return &dellUpdate, nil
}
