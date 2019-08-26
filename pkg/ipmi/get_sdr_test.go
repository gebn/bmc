package ipmi

import (
	"bytes"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
)

func TestGetSDRReqSerializeTo(t *testing.T) {
	table := []struct {
		layer *GetSDRReq
		want  []byte
	}{
		{
			&GetSDRReq{
				ReservationID: 12345,
				RecordID:      54321,
				Length:        22,
			},
			[]byte{
				0x39, 0x30,
				0x31, 0xd4,
				0x00,
				0x16,
			},
		},
		{
			&GetSDRReq{
				ReservationID: 54321,
				RecordID:      12345,
				Offset:        22,
				Length:        255,
			},
			[]byte{
				0x31, 0xd4,
				0x39, 0x30,
				0x16,
				0xff,
			},
		},
	}
	for _, test := range table {
		sb := gopacket.NewSerializeBuffer()
		err := test.layer.SerializeTo(sb, gopacket.SerializeOptions{})
		got := sb.Bytes()

		switch {
		case err != nil && test.want != nil:
			t.Errorf("serialize %v failed with %v, wanted %v", test.layer, err, test.want)
		case err == nil && !bytes.Equal(got, test.want):
			t.Errorf("serialize %v = %v, want %v", test.layer, got, test.want)
		}
	}
}

func TestGetSDRRspDecodeFromBytes(t *testing.T) {
	tests := []struct {
		in   []byte
		want *GetSDRRsp
	}{
		// too short
		{
			make([]byte, 1),
			nil,
		},
		{
			[]byte{
				0x0f, 0xf0,
			},
			&GetSDRRsp{
				BaseLayer: layers.BaseLayer{
					Contents: []byte{0x0f, 0xf0},
					Payload:  []byte{},
				},
				Next: 61455,
			},
		},
		{
			[]byte{
				0xf0, 0x0f,
				0x01, 0x02, 0x03,
			},
			&GetSDRRsp{
				BaseLayer: layers.BaseLayer{
					Contents: []byte{0xf0, 0x0f},
					Payload:  []byte{0x01, 0x02, 0x03},
				},
				Next: 4080,
			},
		},
	}
	for _, test := range tests {
		rsp := &GetSDRRsp{}
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
