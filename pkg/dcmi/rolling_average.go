package dcmi

import (
	"time"
)

// secondsMultiplier parses the 2-bit time duration unit of a rolling average
// time period (see #5 of Table 6-3), returning a number to multiply the time
// duration with in order to provide the duration in seconds. Only bits 0 and 1
// of the input are interpreted.
func secondsMultiplier(unit uint8) int {
	switch unit {
	case 0:
		// 0b00: seconds
		return 1
	case 1:
		// 0b01: minutes
		return 60
	case 2:
		// 0b10: hours
		return 60 * 60
	default: // inc. 3
		// 0b11: days
		return 60 * 60 * 24
	}
}

// rollingAvgPeriodDuration turns the wire representation of a rolling average
// time period into a native duration type. Due to the encoding, the output will
// be a whole number of seconds, between 0 and 63 days. To convert the other
// way, use rollingAvgPeriodByte.
func rollingAvgPeriodDuration(b byte) time.Duration {
	value := int(b & 0x3f) // bottom 6 bits
	if value == 0 {
		return time.Duration(0)
	}
	unit := uint8(b >> 6) // top 2 bits
	seconds := value * secondsMultiplier(unit)
	return time.Second * time.Duration(seconds)
}

// rollingAvgPeriodByte turns a native duration into the wire representation of
// a rolling average time period. Only a subset of durations is representable;
// this function is best-effort. Rather than overflow, the max will be returned
// if the duration is too long. To convert the other way, use
// rollingAvgPeriodByte.
//
// Note that rollingAvgPeriodByte(rollingAvgPeriodDuration(b)) == b for any b,
// however rollingAvgPeriodDuration(rollingAvgPeriodByte(d)) == d is only true
// for some values of d.
func rollingAvgPeriodByte(d time.Duration) byte {
	// N.B. a given duration may have multiple possible representations, e.g.
	// 120 seconds == 2 minutes. The BMC may not understand these if it uses the
	// original byte as a key into a map, so we are very conservative and assume
	// anything >=1 of the next highest unit will be represented in that unit.
	switch {
	case d < time.Second*60:
		return byte(d.Seconds())
	case d < time.Minute*60:
		return byte(d.Minutes()) | 0x40
	case d < time.Hour*24:
		return byte(d.Hours()) | 0x80
	default:
		days := int(d.Hours() / 24)
		if days > 63 {
			days = 63
		}
		return byte(days) | 0xc0
	}
}
