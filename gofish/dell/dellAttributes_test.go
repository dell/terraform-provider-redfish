package dell

import (
	"encoding/json"
	"strings"
	"testing"
)

var idracAttributes = `{
    "@Redfish.Settings": {
        "@odata.context": "/redfish/v1/$metadata#Settings.Settings",
        "@odata.type": "#Settings.v1_3_0.Settings",
        "SettingsObject": {
            "@odata.id": "/redfish/v1/Managers/iDRAC.Embedded.1/Oem/Dell/DellAttributes/iDRAC.Embedded.1/Settings"
        },
        "SupportedApplyTimes": [
            "Immediate",
            "AtMaintenanceWindowStart"
        ]
    },
    "@odata.context": "/redfish/v1/$metadata#DellAttributes.DellAttributes",
    "@odata.id": "/redfish/v1/Managers/iDRAC.Embedded.1/Oem/Dell/DellAttributes/iDRAC.Embedded.1",
    "@odata.type": "#DellAttributes.v1_0_0.DellAttributes",
    "AttributeRegistry": "ManagerAttributeRegistry.v1_0_0",
    "Attributes": {
        "SupportAssist.1.DefaultProtocolPort": 0,
        "SupportAssist.1.HostOSProxyPort": 1,
        "CurrentNIC.1.DedicatedNICScanTime": 5,
        "CurrentNIC.1.MTU": 1500,
        "CurrentNIC.1.NumberOfLOM": 4,
        "CurrentNIC.1.SharedNICScanTime": 30,
        "CurrentNIC.1.VLanID": 1,
        "CurrentNIC.1.VLanPriority": 0,
        "CurrentIPv6.1.IPV6NumOfExtAddress": 0,
        "CurrentIPv6.1.PrefixLength": 64,
        "TelemetryPSUMetrics.1.DevicePollFrequency": 60,
        "TelemetryPSUMetrics.1.ReportInterval": 60,
        "TelemetryPowerStatistics.1.DevicePollFrequency": 60,
        "TelemetryPowerStatistics.1.ReportInterval": 60
    },
    "Description": "This schema provides the oem attributes",
    "Id": "iDRACAttributes",
    "Name": "OEMAttributeRegistry"
}`

func TestDellAttributes(t *testing.T) {
	var dellAttributes DellAttributes

	err := json.NewDecoder(strings.NewReader(idracAttributes)).Decode(&dellAttributes)
	if err != nil {
		t.Fatal("couldn't decode idracAttributes")
	}

	assertField(t, dellAttributes.ID, "iDRACAttributes")
	assertField(t, dellAttributes.Name, "OEMAttributeRegistry")
	assertField(t, dellAttributes.Description, "This schema provides the oem attributes")
	assertField(t, string(dellAttributes.settingsObject), "/redfish/v1/Managers/iDRAC.Embedded.1/Oem/Dell/DellAttributes/iDRAC.Embedded.1/Settings")
	assertMapKeyValue(t, dellAttributes.Attributes.Int("CurrentNIC.1.SharedNICScanTime"), 30)
}
