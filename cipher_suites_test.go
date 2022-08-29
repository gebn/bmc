package bmc

import (
	"testing"

	"github.com/gebn/bmc/pkg/ipmi"

	"github.com/google/go-cmp/cmp"
)

func TestParseCipherSuiteRecordData(t *testing.T) {
	tests := []struct {
		in   []byte
		want []ipmi.CipherSuiteRecord // nil if error
	}{
		// missing auth algo
		{
			in: []byte{
				0xc0, 0x00,
			},
		},
		// integrity and confidentiality implicitly "none"
		{
			[]byte{
				0xc0, 0x00, 0x00,
			},
			[]ipmi.CipherSuiteRecord{
				{},
			},
		},
		// everything explicitly specified, still "none"
		{
			[]byte{
				0xc0, 0x00, 0x00, 0x40, 0x80,
			},
			[]ipmi.CipherSuiteRecord{
				{},
			},
		},
		// second record truncated after start of record
		{
			in: []byte{
				0xc0, 0x00, 0x00, 0x40, 0x80,
				0xc0,
			},
		},
		// second record truncated after cipher suite ID
		{
			in: []byte{
				0xc0, 0x00, 0x00, 0x40, 0x80,
				0xc0, 0x00,
			},
		},
		{
			[]byte{
				0xc0, 0x11, 0x03, 0x44, 0x81, // cipher suite 17
				0xc1, 0x16, 0x00, 0x01, 0x02, 0x01, 0x41, 0x81, // OEM equivalent of cipher suite 3
			},
			[]ipmi.CipherSuiteRecord{
				{
					CipherSuiteID: 17,
					CipherSuite:   ipmi.CipherSuite17,
				},
				{
					CipherSuiteID: 22,
					CipherSuite:   ipmi.CipherSuite3,
					Enterprise:    0x020100,
				},
			},
		},
	}
	for _, test := range tests {
		got, err := parseCipherSuiteRecordData(test.in)
		if err != nil && test.want != nil {
			t.Errorf("parseCipherSuiteRecordData(%v) failed with %v, wanted %v", test.in, err, test.want)
			continue
		}
		if err == nil && test.want == nil {
			t.Errorf("parseCipherSuiteRecordData(%v) returned %v, wanted error", test.in, got)
			continue
		}
		if diff := cmp.Diff(test.want, got); diff != "" {
			t.Errorf("parseCipherSuiteRecordData(%v) returned %v, want %v: %v", test.in, got, test.want, diff)
		}
	}
}
