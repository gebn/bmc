package ipmi

import (
	"testing"
)

func TestParseAnalogDataFormatUnsigned(t *testing.T) {
	tests := []struct {
		in   byte
		want int16
	}{
		{0b00000000, 0},
		{0b00000001, 1},
		{0b10000000, 128},
		{0b11111111, 255},
	}
	for _, test := range tests {
		got := parseAnalogDataFormatUnsigned(test.in)
		if got != test.want {
			t.Errorf("parseAnalogDataFormatUnsigned(%#b) = %v, want %v",
				test.in, got, test.want)
		}
	}
}

func TestParseAnalogDataFormatOnesComplement(t *testing.T) {
	tests := []struct {
		in   byte
		want int16
	}{
		{0b00000000, 0},
		{0b00000001, 1},
		{0b10000000, -127},
		{0b11111111, 0},
	}
	for _, test := range tests {
		got := parseAnalogDataFormatOnesComplement(test.in)
		if got != test.want {
			t.Errorf("parseAnalogDataFormatOnesComplement(%#b) = %v, want %v",
				test.in, got, test.want)
		}
	}
}

func TestParseAnalogDataFormatTwosComplement(t *testing.T) {
	tests := []struct {
		in   byte
		want int16
	}{
		{0b00000000, 0},
		{0b00000001, 1},
		{0b10000000, -128},
		{0b11111111, -1},
	}
	for _, test := range tests {
		got := parseAnalogDataFormatTwosComplement(test.in)
		if got != test.want {
			t.Errorf("parseAnalogDataFormatTwosComplement(%#b) = %v, want %v",
				test.in, got, test.want)
		}
	}
}

func TestAnalogDataFormatParser(t *testing.T) {
	tests := []struct {
		adf  AnalogDataFormat
		err  bool
		in   byte
		want int16
	}{
		{AnalogDataFormatUnsigned, false, 0b01010101, 85},
		{AnalogDataFormatUnsigned, false, 0b10101010, 170},
		{AnalogDataFormatOnesComplement, false, 0b01010101, 85},
		{AnalogDataFormatOnesComplement, false, 0b10101010, -85},
		{AnalogDataFormatTwosComplement, false, 0b01010101, 85},
		{AnalogDataFormatTwosComplement, false, 0b10101010, -86},
		{AnalogDataFormatNotAnalog, true, 0, 0},
		{123, true, 0, 0},
	}
	for _, test := range tests {
		parser, err := test.adf.Parser()
		if err != nil && test.err == false {
			t.Errorf("%v.Parser() returned '%v', want parser", test.adf, err)
			continue
		}
		if err == nil && test.err == true {
			t.Errorf("%v.Parser() returned %v, want err", test.adf, parser)
			continue
		}

		if parser == nil {
			// passed - expected err, got one
			continue
		}

		parsed := parser.Parse(test.in)
		if parsed != test.want {
			t.Errorf("%v.Parse(%v) = %v, want %v",
				parser, test.in, parsed, test.want)
			continue
		}
	}
}
