package ipmi

import (
	"bytes"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
)

func TestGetChannelCipherSuitesReqSerializeTo(t *testing.T) {
	table := []struct {
		layer *GetChannelCipherSuitesReq
		want  []byte
	}{
		{
			&GetChannelCipherSuitesReq{
				Channel:     12,
				PayloadType: 34,
				ListIndex:   56,
			},
			[]byte{
				0x0c,
				0x22,
				0x38 | 0x80,
			},
		},
		{
			&GetChannelCipherSuitesReq{
				Channel:     56,
				PayloadType: 12,
				ListIndex:   34,
			},
			[]byte{
				0x38,
				0x0c,
				0x22 | 0x80,
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

func TestGetChannelCipherSuitesRspDecodeFromBytes(t *testing.T) {
	tests := []struct {
		in   []byte
		want *GetChannelCipherSuitesRsp
	}{
		// too short
		{
			make([]byte, 1),
			nil,
		},
		{
			[]byte{
				0x03, 0xc0, 0x11, 0x02, 0x42, 0x81,
			},
			&GetChannelCipherSuitesRsp{
				BaseLayer: layers.BaseLayer{
					Contents: []byte{0x03, 0xc0, 0x11, 0x02, 0x42, 0x81},
					Payload:  []byte{},
				},
				Channel:           3,
				ID:                17,
				Type:              0xc0,
				OEMIANA:           0,
				ListDataExhausted: true,
				AuthenticationAlgorithms: []AuthenticationAlgorithm{
					AuthenticationAlgorithmHMACMD5,
				},
				IntegrityAlgorithms: []IntegrityAlgorithm{
					IntegrityAlgorithmHMACMD5128,
				},
				ConfidentialityAlgorithms: []ConfidentialityAlgorithm{
					ConfidentialityAlgorithmAESCBC128,
				},
			},
		},
		{
			[]byte{
				0x01, 0xc1, 0x03, 0x01, 0x02, 0x03,
				0x01, 0x41, 0x82, 0x42, 0x81,
			},
			&GetChannelCipherSuitesRsp{
				BaseLayer: layers.BaseLayer{
					Contents: []byte{
						0x01, 0xc1, 0x03, 0x01, 0x02, 0x03,
						0x01, 0x41, 0x82, 0x42, 0x81,
					},
					Payload: []byte{},
				},
				Channel:           1,
				ID:                3,
				Type:              0xc1,
				OEMIANA:           0x030201,
				ListDataExhausted: true,
				AuthenticationAlgorithms: []AuthenticationAlgorithm{
					AuthenticationAlgorithmHMACSHA1,
				},
				IntegrityAlgorithms: []IntegrityAlgorithm{
					IntegrityAlgorithmHMACSHA196,
					IntegrityAlgorithmHMACMD5128,
				},
				ConfidentialityAlgorithms: []ConfidentialityAlgorithm{
					ConfidentialityAlgorithmXRC4128,
					ConfidentialityAlgorithmAESCBC128,
				},
			},
		},
	}
	for _, test := range tests {
		rsp := &GetChannelCipherSuitesRsp{}
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
