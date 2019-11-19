package dcmi

import (
	"bytes"
	"testing"

	"github.com/gebn/bmc/pkg/ipmi"

	"github.com/google/go-cmp/cmp"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
)

func TestGetDCMISensorInfoReqSerializeTo(t *testing.T) {
	tests := []struct {
		layer *GetDCMISensorInfoReq
		want  []byte
	}{
		{
			&GetDCMISensorInfoReq{
				Type:   ipmi.SensorTypeTemperature,
				Entity: ipmi.EntityIDAirInlet,
			},
			[]byte{0x01, 0x37, 0, 0},
		},
		{
			&GetDCMISensorInfoReq{
				Type:          ipmi.SensorTypeTemperature,
				Entity:        ipmi.EntityIDSystemBoard,
				InstanceStart: 8,
			},
			[]byte{0x01, 0x07, 0, 0x08},
		},
		{
			&GetDCMISensorInfoReq{
				Type:          ipmi.SensorTypeTemperature,
				Entity:        ipmi.EntityIDProcessor,
				Instance:      1,
				InstanceStart: 5, // should be ignored
			},
			[]byte{0x01, 0x03, 1, 0},
		},
	}
	opts := gopacket.SerializeOptions{}
	for _, test := range tests {
		sb := gopacket.NewSerializeBuffer()
		if err := test.layer.SerializeTo(sb, opts); err != nil {
			t.Errorf("serialize %v = error %v, want %v", test.layer, err, test.want)
			continue
		}
		got := sb.Bytes()
		if !bytes.Equal(got, test.want) {
			t.Errorf("serialize %v = %v, want %v", test.layer, got, test.want)
		}
	}
}

func TestGetDCMISensorInfoRspDecodeFromBytes(t *testing.T) {
	tests := []struct {
		encoded []byte
		want    *GetDCMISensorInfoRsp
	}{
		// too short
		{
			[]byte{0x0},
			nil,
		},

		// empty result
		{
			[]byte{0x0, 0x0},
			&GetDCMISensorInfoRsp{
				BaseLayer: layers.BaseLayer{
					Contents: []byte{0x0, 0x0},
					Payload:  []byte{},
				},
				Instances: 0,
				RecordIDs: nil,
			},
		},

		// packet 1 byte too short for specified number of recordIDs
		{
			[]byte{0x0, 0x2, 0x1, 0x2, 0x3},
			nil,
		},
		{
			[]byte{0x2, 0x1, 0xab, 0xba},
			&GetDCMISensorInfoRsp{
				BaseLayer: layers.BaseLayer{
					Contents: []byte{0x2, 0x1, 0xab, 0xba},
					Payload:  []byte{},
				},
				Instances: 2,
				RecordIDs: []ipmi.RecordID{0xbaab},
			},
		},
		{
			[]byte{0x9, 0x2, 0xf0, 0x0f, 0x0f, 0xf0, 0xff},
			&GetDCMISensorInfoRsp{
				BaseLayer: layers.BaseLayer{
					Contents: []byte{0x9, 0x2, 0xf0, 0x0f, 0x0f, 0xf0},
					Payload:  []byte{0xff},
				},
				Instances: 9,
				RecordIDs: []ipmi.RecordID{0xff0, 0xf00f},
			},
		},
	}
	layer := &GetDCMISensorInfoRsp{}
	for _, test := range tests {
		err := layer.DecodeFromBytes(test.encoded, gopacket.NilDecodeFeedback)
		switch {
		case err == nil && test.want == nil:
			t.Errorf("decode %v succeeded with %v, wanted error", test.encoded,
				layer)
		case err != nil && test.want != nil:
			t.Errorf("decode %v failed with %v, wanted %v", test.encoded, err,
				test.want)
		case err == nil && test.want != nil:
			if diff := cmp.Diff(test.want, layer, cmp.AllowUnexported(*layer)); diff != "" {
				t.Errorf("decode %v = %v, want %v: %v", test.encoded, layer, test.want, diff)
			}
		}
	}
}
