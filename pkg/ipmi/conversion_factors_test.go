package ipmi

import (
	"testing"
)

func TestConversionFactorsConvertReading(t *testing.T) {
	tests := []struct {
		cf   ConversionFactors
		raw  int16
		want int64
	}{
		{ConversionFactors{1, 0, 0, 0}, 40, 40},
		{ConversionFactors{1, 2, 3, 4}, 40, 20400000},
		{ConversionFactors{-1, 1, -8, 5}, 33, -3300000},
		{ConversionFactors{9, 27, 5, 2}, -33, 269970300},
		{ConversionFactors{-1, -1, -8, -8}, -33, 0},
	}
	for _, test := range tests {
		got := test.cf.ConvertReading(test.raw)
		if got != test.want {
			t.Errorf("%+v.ConvertReading(%v) = %v, want %v", test.cf,
				test.raw, got, test.want)
		}
	}
}
