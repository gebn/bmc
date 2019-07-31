package dcmi

import (
	"testing"
	"time"
)

func TestSecondsMultiplier(t *testing.T) {
	tests := []struct {
		in   uint8
		want int
	}{
		{0, 1},
		{1, 60},
		{2, 3600},
		{3, 86400},
		// higher inputs unspecified
	}
	for _, test := range tests {
		if got := secondsMultiplier(test.in); got != test.want {
			t.Errorf("secondsMultiplier(%v) = %v, want %v", test.in, got,
				test.want)
		}
	}
}

func TestRollingAvgPeriod(t *testing.T) {
	tests := []struct {
		b byte
		d time.Duration
	}{
		{0x00, time.Duration(0)},
		{0x01, time.Second},
		{0x05, time.Second * 5},
		{0x0f, time.Second * 15},
		{0x1e, time.Second * 30},
		{0x2a, time.Second * 42},
		{0x41, time.Minute},
		{0x43, time.Minute * 3},
		{0x47, time.Minute * 7},
		{0x4f, time.Minute * 15},
		{0x5e, time.Minute * 30},
		{0x55, time.Minute * 21},
		{0x81, time.Hour},
		{0xc1, time.Hour * 24},
		{0xcc, time.Hour * 24 * 12},
		{0xff, time.Hour * 24 * 63},
	}
	for _, test := range tests {
		if got := rollingAvgPeriodDuration(test.b); got != test.d {
			t.Errorf("rollingAvgPeriodDuration(%v) = %v, want %v", test.b, got,
				test.d)
		}
		if got := rollingAvgPeriodByte(test.d); got != test.b {
			t.Errorf("rollingAvgPeriodByte(%v) = %#x, want %#x", test.d, got,
				test.b)
		}
	}
}
