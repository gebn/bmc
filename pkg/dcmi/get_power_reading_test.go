package dcmi

import (
	"bytes"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/gopacket"
)

func TestGetPowerReadingReqSerializeTo(t *testing.T) {
	tests := []struct {
		in   *GetPowerReadingReq
		want []byte
	}{
		{
			&GetPowerReadingReq{
				Mode: SystemPowerStatisticsModeNormal,
			},
			[]byte{0x01, 0x00, 0x00},
		},
		{
			&GetPowerReadingReq{
				Mode:   SystemPowerStatisticsModeNormal,
				Period: time.Hour, // should be ignored
			},
			[]byte{0x01, 0x00, 0x00},
		},
		{
			&GetPowerReadingReq{
				Mode: SystemPowerStatisticsModeEnhanced,
			},
			[]byte{0x02, 0x00, 0x00},
		},
		{
			&GetPowerReadingReq{
				Mode:   SystemPowerStatisticsModeEnhanced,
				Period: time.Minute * 5,
			},
			[]byte{0x02, 0x45, 0x00},
		},
	}
	opts := gopacket.SerializeOptions{}
	for _, test := range tests {
		sb := gopacket.NewSerializeBuffer()
		if err := test.in.SerializeTo(sb, opts); err != nil {
			t.Errorf("serialize %v = error %v, want %v", test.in, err, test.want)
			continue
		}
		got := sb.Bytes()
		if !bytes.Equal(got, test.want) {
			t.Errorf("serialize %v = %v, want %v", test.in, got, test.want)
		}
	}
}

func TestGetPowerReadingRspDecodeFromBytes(t *testing.T) {
	tests := []struct {
		in   []byte
		want *GetPowerReadingRsp // nil if error
	}{
		{
			[]byte{
				0xae, 0x08,
				0x57, 0x04,
				0x05, 0x0d,
				0xd2, 0x04,
				0x73, 0xb6, 0x44, 0x5d,
				0xaa, 0xbb, 0xcc, 0xdd,
				1 << 6,
			},
			&GetPowerReadingRsp{
				Instantaneous: 2222,
				Min:           1111,
				Max:           3333,
				Avg:           1234,
				Timestamp:     time.Unix(1564784243, 0),
				Period:        time.Millisecond * 3721182122,
				Active:        true,
			},
		},
		{
			[]byte{
				0x7a, 0x00,
				0x50, 0x00,
				0x96, 0x00,
				0x78, 0x00,
				0x2b, 0xb8, 0x44, 0x5d,
				0xdd, 0xcc, 0xbb, 0xaa,
				^byte(1 << 6),
			},
			&GetPowerReadingRsp{
				Instantaneous: 122,
				Min:           80,
				Max:           150,
				Avg:           120,
				Timestamp:     time.Unix(1564784683, 0),
				Period:        time.Millisecond * 2864434397,
				Active:        false,
			},
		},
	}
	layer := &GetPowerReadingRsp{}
	for _, test := range tests {
		err := layer.DecodeFromBytes(test.in, gopacket.NilDecodeFeedback)
		switch {
		case err == nil && test.want == nil:
			t.Errorf("decode %v succeeded with %v, wanted error", test.in,
				layer)
		case err != nil && test.want != nil:
			t.Errorf("decode %v failed with %v, wanted %v", test.in, err,
				test.want)
		case err == nil && test.want != nil:
			if diff := cmp.Diff(test.want, layer); diff != "" {
				t.Errorf("decode %v = %v, want %v: %v", test.in, layer, test.want, diff)
			}
		}
	}
}
