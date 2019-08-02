package dcmi

import (
	"fmt"
)

// SystemPowerStatisticsMode represents whether enhanced system power statistics
// are in use.
type SystemPowerStatisticsMode uint8

const (
	// SystemPowerStatisticsModeNormal means no rolling average time period can
	// be specified in the Get Power Reading command. This is the only option
	// supported by v1.0.
	SystemPowerStatisticsModeNormal SystemPowerStatisticsMode = 0x01

	// SystemPowerStatisticsModeEnhanced means one of the rolling average time
	// periods returned in the Get DCMI Capabilities Info response can be sent
	// to the BMC, and the corresponding readings will be returned. This is less
	// useful when constantly polling BMCs for the current reading.
	SystemPowerStatisticsModeEnhanced SystemPowerStatisticsMode = 0x02
)

// describe returns a human-friendly name for the mode.
func (s SystemPowerStatisticsMode) describe() string {
	switch s {
	case SystemPowerStatisticsModeNormal:
		return "Normal"
	case SystemPowerStatisticsModeEnhanced:
		return "Enhanced"
	default:
		return "Unknown"
	}
}

func (s SystemPowerStatisticsMode) String() string {
	return fmt.Sprintf("%v(%v)", uint8(s), s.describe())
}
