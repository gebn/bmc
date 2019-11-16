package ipmi

import (
	"testing"
)

func TestConvertAnalogDataFormatUnsigned(t *testing.T) {
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
		got := convertAnalogDataFormatUnsigned(test.in)
		if got != test.want {
			t.Errorf("convertAnalogDataFormatUnsigned(%#b) = %v, want %v",
				test.in, got, test.want)
		}
	}
}

func TestConvertAnalogDataFormatOnesComplement(t *testing.T) {
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
		got := convertAnalogDataFormatOnesComplement(test.in)
		if got != test.want {
			t.Errorf("convertAnalogDataFormatOnesComplement(%#b) = %v, want %v",
				test.in, got, test.want)
		}
	}
}

func TestConvertAnalogDataFormatTwosComplement(t *testing.T) {
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
		got := convertAnalogDataFormatTwosComplement(test.in)
		if got != test.want {
			t.Errorf("convertAnalogDataFormatTwosComplement(%#b) = %v, want %v",
				test.in, got, test.want)
		}
	}
}

func TestAnalogDataFormatConverter(t *testing.T) {
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
		converter, err := test.adf.Converter()
		if err != nil && test.err == false {
			t.Errorf("%v.Converter() returned '%v', want converter",
				test.adf, err)
			continue
		}
		if err == nil && test.err == true {
			t.Errorf("%v.Converter() returned %v, want err",
				test.adf, converter)
			continue
		}

		if converter == nil {
			// passed - expected err, got one
			continue
		}

		converted := converter.Convert(test.in)
		if converted != test.want {
			t.Errorf("%v.Convert(%v) = %v, want %v",
				converter, test.in, converted, test.want)
			continue
		}
	}
}
