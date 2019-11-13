package ipmi

import (
	"testing"
)

func TestEntityInstanceIsSystemRelative(t *testing.T) {
	tests := []struct {
		instance EntityInstance
		want     bool
	}{
		{0, true},
		{0x20, true},
		{0x5f, true},
		{0x60, false},
		{0x7f, false},
		{0xff, false},
	}
	for _, test := range tests {
		got := test.instance.IsSystemRelative()
		if got != test.want {
			t.Errorf("IsSystemRelative(%v) = %v, want %v", test.instance, got,
				test.want)
		}
	}
}

func TestEntityInstanceIsDeviceRelative(t *testing.T) {
	tests := []struct {
		instance EntityInstance
		want     bool
	}{
		{0, false},
		{0x20, false},
		{0x5f, false},
		{0x60, true},
		{0x65, true},
		{0x7f, true},
		{0xff, false},
	}
	for _, test := range tests {
		got := test.instance.IsDeviceRelative()
		if got != test.want {
			t.Errorf("IsDeviceRelative(%v) = %v, want %v", test.instance, got,
				test.want)
		}
	}
}
