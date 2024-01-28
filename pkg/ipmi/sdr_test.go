package ipmi

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
)

func TestSDRDecodeFromBytes(t *testing.T) {
	tests := []struct {
		in   []byte
		want *SDR
	}{
		// too short
		{
			make([]byte, 4),
			nil,
		},
		{
			[]byte{
				0x0f, 0xf0,
				0x99,
				0x01,
				0x16,
			},
			&SDR{
				BaseLayer: layers.BaseLayer{
					Contents: []byte{
						0x0f, 0xf0,
						0x99,
						0x01,
						0x16,
					},
					Payload: []byte{},
				},
				ID:      61455,
				Version: 99,
				Type:    RecordTypeFullSensor,
				Length:  22,
			},
		},
		{
			[]byte{
				0xf0, 0x0f,
				0x51,
				0x02,
				0x20,
				0x01, 0x02,
			},
			&SDR{
				BaseLayer: layers.BaseLayer{
					Contents: []byte{
						0xf0, 0x0f,
						0x51,
						0x02,
						0x20,
					},
					Payload: []byte{0x01, 0x02},
				},
				ID:      4080,
				Version: 15,
				Type:    RecordTypeCompactSensor,
				Length:  32,
			},
		},
	}
	for _, test := range tests {
		rsp := &SDR{}
		err := rsp.DecodeFromBytes(test.in, gopacket.NilDecodeFeedback)
		switch {
		case err == nil && test.want == nil:
			t.Errorf("expected error decoding %v, got none", test.in)
		case err == nil && test.want != nil:
			if diff := cmp.Diff(test.want, rsp); diff != "" {
				t.Errorf("decode %v = %v, want %v: %v", test.in, rsp, test.want, diff)
			}
		case err != nil && test.want != nil:
			t.Errorf("unexpected error: %v", err)
		}
	}
}
