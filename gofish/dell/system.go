package dell

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/stmcginnis/gofish/common"
	"github.com/stmcginnis/gofish/redfish"

	"github.com/stmcginnis/gofish"
)

const expand = "$expand=*($levels=1)"

func UnmarshalHttpResponse(resp *http.Response, v any) error {
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	return json.Unmarshal(body, v)
}

type Service struct {
	parent gofish.Service
	Params ServiceParams
}

type Link string

func (l *Link) UnmarshalJSON(b []byte) error {
	var s common.Link
	err := json.Unmarshal(b, &s)
	if err != nil {
		return err
	}
	*l = Link(s.String())
	return nil
}

type ServiceParams struct {
	ID                 string
	CertificateService Link
	Chassis            Link
	Managers           Link
	Tasks              Link
	StorageServices    Link
	StorageSystems     Link
	AccountService     Link
	EventService       Link
	PowerEquipment     Link
	Registries         Link
	Systems            Link
	CompositionService Link
	Fabrics            Link
	JobService         Link
	JSONSchemas        Link `json:"JsonSchemas"`
	ResourceBlocks     Link
	SessionService     Link
	TelemetryService   Link
	UpdateService      Link
	Links              struct {
		Sessions Link
	}
}

func GetService(c *gofish.APIClient) (Service, error) {
	resp, err := c.Get(common.DefaultServiceRoot)
	if err != nil {
		return Service{}, err
	}
	var ret Service
	if err := UnmarshalHttpResponse(resp, &ret); err != nil {
		return Service{}, err
	}
	ret.parent.SetClient(c)
	return ret, nil
}

func (root *Service) UnmarshalJSON(b []byte) error {
	errRoot := json.Unmarshal(b, &root.parent)
	errParams := json.Unmarshal(b, &root.Params)
	return errors.Join(errRoot, errParams)
}

type Collection struct {
	Members json.RawMessage
}

func (root *Service) GetSystems() ([]System, error) {
	var systemsCollection Collection
	// fmt.Sprintln("Starting with Getsystems")
	err := root.parent.Get(root.parent.GetClient(),
		fmt.Sprintf("%s?%s", string(root.Params.Systems), expand),
		&systemsCollection)
	if err != nil {
		return nil, err
	}
	var ret []System
	if err := json.Unmarshal(systemsCollection.Members, &ret); err != nil {
		return nil, err
	}
	for i := range ret {
		ret[i].parent.SetClient(root.parent.GetClient())
	}
	return ret, err
}

func (root *Service) GetSystem() (System, error) {
	systems, err := root.GetSystems()
	if err != nil {
		return System{}, err
	}
	if len(systems) == 0 {
		return System{}, errors.New("no systems found")
	}
	return systems[0], nil
}

type System struct {
	parent redfish.ComputerSystem
	Params SystemParams
}

type SystemParams struct {
	// Actions            CSActions
	Bios               Link
	Processors         Link
	Memory             Link
	EthernetInterfaces Link
	SimpleStorage      Link
	SecureBoot         Link
	Storage            Link
	NetworkInterfaces  Link
	LogServices        Link
	MemoryDomains      Link
	PCIeDevices        []Link
	PCIeFunctions      []Link
	VirtualMedia       Link
	// Links              CSLinks
	// Settings           common.Settings `json:"@Redfish.Settings"`
}

func (root *System) UnmarshalJSON(b []byte) error {
	errRoot := json.Unmarshal(b, &root.parent)
	errParams := json.Unmarshal(b, &root.Params)
	return errors.Join(errRoot, errParams)
}
