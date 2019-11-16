package complement

import (
	"testing"
)

func TestOnes(t *testing.T) {
	tests := []struct {
		in   byte
		want int8
	}{
		{0b01111111, 127},
		{0b01111110, 126},
		{0b00000010, 2},
		{0b00000001, 1},
		{0b00000000, 0},
		{0b11111111, 0},
		{0b11111110, -1},
		{0b11111101, -2},
		{0b10000001, -126},
		{0b10000000, -127},
	}
	for _, test := range tests {
		got := Ones(test.in)
		if got != test.want {
			t.Errorf("Ones(%#b) = %v, want %v", test.in, got, test.want)
		}
	}
}
