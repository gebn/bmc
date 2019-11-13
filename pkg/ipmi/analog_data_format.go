package ipmi

import (
	"fmt"
)

// AnalogDataFormat represents the binary format of analog sensor readings and
// thresholds. It is specified in byte 21 of the Full Sensor Record table in
// 37.1 and 43.1 of v1.5 and v2.0 respectively. It is a 2-bit uint on the wire.
type AnalogDataFormat uint8

const (
	// AnalogDataFormatUnsigned indicates an unsigned analog sensor. It is also
	// used in the case
	// where the sensor provides neither analog readings nor thresholds.
	AnalogDataFormatUnsigned AnalogDataFormat = iota
	AnalogDataFormatOnesComplement
	AnalogDataFormatTwosComplement

	// AnalogDataFormatNotAnalog indicates the sensor does not have numeric
	// readings, only thresholds.
	AnalogDataFormatNotAnalog
)

var (
	analogDataFormatDescriptions = map[AnalogDataFormat]string{
		AnalogDataFormatUnsigned:       "Unsigned",
		AnalogDataFormatOnesComplement: "1's Complement",
		AnalogDataFormatTwosComplement: "2's Complement",
		AnalogDataFormatNotAnalog:      "No analog readings",
	}
)

func (f AnalogDataFormat) Description() string {
	if desc, ok := analogDataFormatDescriptions[f]; ok {
		return desc
	}
	return "Unknown"
}

func (f AnalogDataFormat) String() string {
	return fmt.Sprintf("%#v(%v)", uint8(f), f.Description())
}
