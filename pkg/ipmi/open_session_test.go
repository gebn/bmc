package ipmi

import (
	"bytes"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
)

func TestOpenSessionReqSerializeTo(t *testing.T) {
	table := []struct {
		layer *OpenSessionReq
		wire  []byte
	}{
		{
			&OpenSessionReq{
				Tag:               123,
				MaxPrivilegeLevel: PrivilegeLevelUser,
				SessionID:         0x03020401,
				AuthenticationPayload: AuthenticationPayload{
					Wildcard: true,
				},
				IntegrityPayload: IntegrityPayload{
					Algorithm: IntegrityAlgorithmHMACSHA196,
				},
				ConfidentialityPayload: ConfidentialityPayload{
					Algorithm: ConfidentialityAlgorithmAESCBC128,
				},
			},
			[]byte{0x7b, 0x02, 0x00, 0x00, 0x01, 0x04, 0x02, 0x03,
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x01, 0x00, 0x00, 0x08, 0x01, 0x00, 0x00, 0x00,
				0x02, 0x00, 0x00, 0x08, 0x01, 0x00, 0x00, 0x00},
		},
	}
	opts := gopacket.SerializeOptions{}
	for _, test := range table {
		sb := gopacket.NewSerializeBuffer()
		if err := test.layer.SerializeTo(sb, opts); err != nil {
			t.Errorf("serialize %v = error %v, want %v", test.layer, err,
				test.wire)
			continue
		}
		got := sb.Bytes()
		if !bytes.Equal(got, test.wire) {
			t.Errorf("serialize %v = %v, want %v", test.layer, got, test.wire)
		}
	}
}

func TestOpenSessionRspDecodeFromBytes(t *testing.T) {
	table := []struct {
		wire  []byte
		layer *OpenSessionRsp
	}{
		{
			[]byte{
				0x00,
				0x00,
				0x04,
				0x00,
				0xa4, 0xa3, 0xa2, 0xa0,
				0x9c, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x08, 0x01, 0x00, 0x00, 0x00,
				0x01, 0x00, 0x00, 0x08, 0x01, 0x00, 0x00, 0x00,
				0x02, 0x00, 0x00, 0x08, 0x01, 0x00, 0x00, 0x00},
			&OpenSessionRsp{
				BaseLayer: layers.BaseLayer{
					Contents: []byte{
						0x00, 0x00, 0x04, 0x00, 0xa4, 0xa3, 0xa2, 0xa0, 0x9c,
						0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x08, 0x01, 0x00,
						0x00, 0x00, 0x01, 0x00, 0x00, 0x08, 0x01, 0x00, 0x00,
						0x00, 0x02, 0x00, 0x00, 0x08, 0x01, 0x00, 0x00, 0x00},
				},
				Tag:                    0,
				Status:                 StatusCodeOK,
				MaxPrivilegeLevel:      PrivilegeLevelAdministrator,
				RemoteConsoleSessionID: 0xa0a2a3a4,
				ManagedSystemSessionID: 0x9c,
				AuthenticationPayload: AuthenticationPayload{
					Algorithm: AuthenticationAlgorithmHMACSHA1,
				},
				IntegrityPayload: IntegrityPayload{
					Algorithm: IntegrityAlgorithmHMACSHA196,
				},
				ConfidentialityPayload: ConfidentialityPayload{
					Algorithm: ConfidentialityAlgorithmAESCBC128,
				},
			},
		},
		{
			[]byte{
				0x01,
				0x11,
				0x00,
				0x01, 0x00, 0x00, 0x00},
			&OpenSessionRsp{
				BaseLayer: layers.BaseLayer{
					Contents: []byte{0x01, 0x11, 0x00, 0x01, 0x00, 0x00, 0x00},
				},
				Tag:                    1,
				Status:                 StatusCodeUnsupportedCipherSuite,
				RemoteConsoleSessionID: 0x00000001,
			},
		},
		{
			[]byte{0xc7},
			&OpenSessionRsp{
				BaseLayer: layers.BaseLayer{
					Contents: []byte{0xc7},
				},
				Status: StatusCodeInvalidRequestLength,
			},
		},
	}
	layer := &OpenSessionRsp{}
	for _, test := range table {
		err := layer.DecodeFromBytes(test.wire, gopacket.NilDecodeFeedback)
		switch {
		case err == nil && test.layer == nil:
			t.Errorf("decode %v succeeded with %v, wanted error", test.wire,
				layer)
		case err != nil && test.layer != nil:
			t.Errorf("decode %v failed with %v, wanted %v", test.wire, err,
				test.layer)
		case err == nil && test.layer != nil:
			if diff := cmp.Diff(test.layer, layer); diff != "" {
				t.Errorf("decode %v = %v, want %v: %v", test.wire, layer, test.layer, diff)
			}
		}
	}
}
