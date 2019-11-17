package ipmi

import (
	"testing"
)

func TestConversionFactorsConvertReading(t *testing.T) {
	tests := []struct {
		cf   ConversionFactors
		raw  int16
		want float64
	}{
		{ConversionFactors{1, 0, 0, 0}, 40, 40},         // CPU temp
		{ConversionFactors{100, 0, 0, 0}, 128, 12800},   // fan speed
		{ConversionFactors{9, 171, 0, -3}, 181, 1.8},    // CPU voltage
		{ConversionFactors{7, 137, 0, -3}, 184, 1.425},  // DIMM voltage
		{ConversionFactors{51, 219, 0, -3}, 231, 12},    // 12V
		{ConversionFactors{31, 71, 0, -3}, 159, 5},      // 5VCC
		{ConversionFactors{15, 179, 0, -3}, 208, 3.299}, // 3.3VCC
		{ConversionFactors{1, 2, 3, 4}, 40, 20400000},
		{ConversionFactors{9, 27, 5, 2}, -33, 269970300},
	}
	for _, test := range tests {
		got := test.cf.ConvertReading(test.raw)
		if got != test.want {
			t.Errorf("%+v.ConvertReading(%v) = %v, want %v", test.cf,
				test.raw, got, test.want)
		}
	}
}
