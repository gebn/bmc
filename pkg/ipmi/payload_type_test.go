package ipmi

import (
	"testing"
)

func TestPayloadTypeDescription(t *testing.T) {
	tests := []struct {
		in   PayloadType
		want string
	}{
		{PayloadTypeIPMI, "IPMI"},
		{0x20, "OEM0"},
		{0x27, "OEM7"},
		{0x28, "Unknown"},
	}
	for _, test := range tests {
		got := test.in.Description()
		if got != test.want {
			t.Errorf("PayloadType(%#v).Description() = %v, want %v",
				uint8(test.in), got, test.want)
		}
	}
}
