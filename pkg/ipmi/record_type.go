package ipmi

import (
	"fmt"
)

// RecordType identifies the format of a Sensor Data Record. Although called
// SDRs, a sensor data record does not necessarily pertain to a sensor.
type RecordType uint8

const (
	RecordTypeFullSensor                        RecordType = 0x01
	RecordTypeCompactSensor                     RecordType = 0x02
	RecordTypeEventOnly                         RecordType = 0x03
	RecordTypeEntityAssociation                 RecordType = 0x08
	RecordTypeDeviceRelativeEntityAssociation   RecordType = 0x09
	RecordTypeGenericDeviceLocator              RecordType = 0x10
	RecordTypeFRUDeviceLocator                  RecordType = 0x11
	RecordTypeManagementControllerDeviceLocator RecordType = 0x12
	RecordTypeManagementControllerConfirmation  RecordType = 0x13
	RecordTypeBMCMessageChannelInfo             RecordType = 0x14
)

var (
	recordTypeDescriptions = map[RecordType]string{
		RecordTypeFullSensor:                        "Full Sensor Record",
		RecordTypeCompactSensor:                     "Compact Sensor Record",
		RecordTypeEventOnly:                         "Event-only Record",
		RecordTypeEntityAssociation:                 "Entity Association Record",
		RecordTypeDeviceRelativeEntityAssociation:   "Device-relative Entity Association Record",
		RecordTypeGenericDeviceLocator:              "Generic Device Locator Record",
		RecordTypeFRUDeviceLocator:                  "FRU Device Locator Record",
		RecordTypeManagementControllerDeviceLocator: "Management Controller Device Locator Record",
		RecordTypeManagementControllerConfirmation:  "Management Controller Confirmation Record",
		RecordTypeBMCMessageChannelInfo:             "BMC Message Channel Info Record",
	}
)

func (t RecordType) Description() string {
	if desc, ok := recordTypeDescriptions[t]; ok {
		return desc
	}
	return "Unknown"
}

func (t RecordType) String() string {
	return fmt.Sprintf("%#x(%v)", uint8(t), t.Description())
}
