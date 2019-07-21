package bcd

import (
	"testing"
)

func TestDecode(t *testing.T) {
	tests := []struct {
		in   byte
		want uint8
	}{
		{0x00, 0},
		{0x01, 1},
		{0x02, 2},
		{0x10, 10},
		{0x11, 11},
		{0x99, 99},
	}
	for _, test := range tests {
		if got := Decode(test.in); got != test.want {
			t.Errorf("Decode(%v) = %v, want %v", test.in, got, test.want)
		}
	}
}
