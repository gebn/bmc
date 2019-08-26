package ipmi

import (
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
)

func TestGetSDRRepositoryInfoRspDecodeFromBytes(t *testing.T) {
	tests := []struct {
		in   []byte
		want *GetSDRRepositoryInfoRsp
	}{
		// too short
		{
			make([]byte, 13),
			nil,
		},
		{
			[]byte{
				0x02,
				0xab, 0xba,
				0xcd, 0xdc,
				0x04, 0x03, 0x02, 0x01,
				0x01, 0x02, 0x03, 0x04,
				0x55,
			},
			&GetSDRRepositoryInfoRsp{
				BaseLayer: layers.BaseLayer{
					Contents: []byte{
						0x02,
						0xab, 0xba,
						0xcd, 0xdc,
						0x04, 0x03, 0x02, 0x01,
						0x01, 0x02, 0x03, 0x04,
						0x55,
					},
					Payload: []byte{},
				},
				Version:                          20,
				Records:                          47787,
				FreeSpace:                        56525,
				LastAddition:                     time.Unix(16909060, 0),
				LastErase:                        time.Unix(67305985, 0),
				Overflow:                         false,
				SupportsModalUpdate:              true,
				SupportsNonModalUpdate:           false,
				SupportsDelete:                   false,
				SupportsPartialAdd:               true,
				SupportsReserve:                  false,
				SupportsGetAllocationInformation: true,
			},
		},
		{
			[]byte{
				0x51,
				0x0f, 0xf0,
				0xf0, 0x0f,
				0x01, 0x02, 0x03, 0x04,
				0x04, 0x03, 0x02, 0x01,
				0xaa,
				0xff, // trailing
			},
			&GetSDRRepositoryInfoRsp{
				BaseLayer: layers.BaseLayer{
					Contents: []byte{
						0x51,
						0x0f, 0xf0,
						0xf0, 0x0f,
						0x01, 0x02, 0x03, 0x04,
						0x04, 0x03, 0x02, 0x01,
						0xaa,
					},
					Payload: []byte{0xff},
				},
				Version:                          15,
				Records:                          61455,
				FreeSpace:                        4080,
				LastAddition:                     time.Unix(67305985, 0),
				LastErase:                        time.Unix(16909060, 0),
				Overflow:                         true,
				SupportsModalUpdate:              false,
				SupportsNonModalUpdate:           true,
				SupportsDelete:                   true,
				SupportsPartialAdd:               false,
				SupportsReserve:                  true,
				SupportsGetAllocationInformation: false,
			},
		},
	}
	for _, test := range tests {
		rsp := &GetSDRRepositoryInfoRsp{}
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
