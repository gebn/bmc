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

func (t RecordType) Description() string {
	switch t {
	case RecordTypeFullSensor:
		return "Full Sensor Record"
	case RecordTypeCompactSensor:
		return "Compact Sensor Record"
	case RecordTypeEventOnly:
		return "Event-Only Record"
	case RecordTypeEntityAssociation:
		return "Entity Association Record"
	case RecordTypeDeviceRelativeEntityAssociation:
		return "Device-relative Entity Association Record"
	case RecordTypeGenericDeviceLocator:
		return "Generic Device Locator Record"
	case RecordTypeFRUDeviceLocator:
		return "FRU Device Locator Record"
	case RecordTypeManagementControllerDeviceLocator:
		return "Management Controller Device Locator Record"
	case RecordTypeManagementControllerConfirmation:
		return "Management Controller Confirmation Record"
	case RecordTypeBMCMessageChannelInfo:
		return "BMC Message Channel Info Record"
	default:
		return "Unknown"
	}
}

func (t RecordType) String() string {
	return fmt.Sprintf("%#x(%v)", uint8(t), t.Description())
}
