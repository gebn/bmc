package ipmi

import (
	"bytes"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
)

func TestGetSensorReadingReqSerializeTo(t *testing.T) {
	tests := []struct {
		layer *GetSensorReadingReq
		want  []byte
	}{
		{
			&GetSensorReadingReq{
				Number: 0,
			},
			[]byte{0x00},
		},
		{
			&GetSensorReadingReq{
				Number: 22,
			},
			[]byte{0x16},
		},
		{
			&GetSensorReadingReq{
				Number: 254,
			},
			[]byte{0xfe},
		},
	}
	for _, test := range tests {
		sb := gopacket.NewSerializeBuffer()
		err := test.layer.SerializeTo(sb, gopacket.SerializeOptions{})
		got := sb.Bytes()

		switch {
		case err != nil && test.want != nil:
			t.Errorf("serialize %+v failed with %v, wanted %v", test.layer, err, test.want)
		case err == nil && !bytes.Equal(got, test.want):
			t.Errorf("serialize %+v = %v, want %v", test.layer, got, test.want)
		}
	}
}

func TestGetSensorReadingRspDecodeFromBytes(t *testing.T) {
	tests := []struct {
		in   []byte
		want *GetSensorReadingRsp
	}{
		{
			make([]byte, 2),
			nil,
		},
		{
			[]byte{0x16, 0b10100000, 0},
			&GetSensorReadingRsp{
				BaseLayer: layers.BaseLayer{
					Contents: []byte{0x16, 0b10100000, 0},
					Payload:  []byte{},
				},
				Reading:              22,
				EventMessagesEnabled: true,
				ScanningEnabled:      false,
				ReadingUnavailable:   true,
			},
		},
		{
			[]byte{0xff, 0b01011111, 0, 1, 2, 3},
			&GetSensorReadingRsp{
				BaseLayer: layers.BaseLayer{
					Contents: []byte{0xff, 0b01011111, 0, 1},
					Payload:  []byte{2, 3},
				},
				Reading:              255,
				EventMessagesEnabled: false,
				ScanningEnabled:      true,
				ReadingUnavailable:   false,
			},
		},
	}
	for _, test := range tests {
		rsp := &GetSensorReadingRsp{}
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
