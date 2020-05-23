package ipmi

import (
	"bytes"
	"net"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
)

func TestGetSessionInfoReqSerializeTo(t *testing.T) {
	table := []struct {
		layer *GetSessionInfoReq
		want  []byte
	}{
		{
			&GetSessionInfoReq{},
			[]byte{0x00},
		},
		{
			&GetSessionInfoReq{
				Index: 1,
			},
			[]byte{0x01},
		},
		{
			&GetSessionInfoReq{
				Index:  SessionIndexHandle,
				Handle: 5,
			},
			[]byte{0xfe, 0x5},
		},
		{
			&GetSessionInfoReq{
				Index: SessionIndexID,
				ID:    22,
			},
			[]byte{0xff, 0x16, 0x0, 0x0, 0x0},
		},
		{
			&GetSessionInfoReq{
				Index:  5,
				Handle: 6,
				ID:     22,
			},
			[]byte{0x5},
		},
		{
			&GetSessionInfoReq{
				Handle: 6,
				ID:     22,
			},
			[]byte{0x0},
		},
	}
	for _, test := range table {
		sb := gopacket.NewSerializeBuffer()
		err := test.layer.SerializeTo(sb, gopacket.SerializeOptions{
			FixLengths: true,
		})
		got := sb.Bytes()

		switch {
		case err != nil && test.want != nil:
			t.Errorf("serialize %v failed with %v, wanted %v", test.layer, err, test.want)
		case err == nil && !bytes.Equal(got, test.want):
			t.Errorf("serialize %v = %v, want %v", test.layer, got, test.want)
		}
	}
}

func TestGetSessionInfoRspDecodeFromBytes(t *testing.T) {
	table := []struct {
		in   []byte
		want *GetSessionInfoRsp
	}{
		{
			[]byte{0x0, 01},
			nil, // should be at least 3 bytes
		},
		{
			[]byte{0x0, 0x02, 0x02},
			&GetSessionInfoRsp{
				BaseLayer: layers.BaseLayer{
					Contents: []byte{0x0, 0x02, 0x02},
					Payload:  []byte{},
				},
				Handle: 0,
				Max:    2,
				Active: 2,
			},
		},
		{
			[]byte{0x16, 0x08, 0x04},
			nil, // should be at least 6 bytes for active session
		},
		{
			[]byte{0x00, 0x10, 0x2, 0x2, 0x2, 0x11},
			&GetSessionInfoRsp{
				BaseLayer: layers.BaseLayer{
					Contents: []byte{0x00, 0x10, 0x2, 0x2, 0x2, 0x11},
					Payload:  []byte{},
				},
				Handle:         0, // Super Micro sends 0 but includes additional fields
				Max:            16,
				Active:         2,
				UserID:         2,
				PrivilegeLevel: PrivilegeLevelUser,
				IsIPMIv2:       true,
				Channel:        Channel(1),
			},
		},
		{
			[]byte{0x16, 0x08, 0x04, 0x1, 0x2, 0x11},
			&GetSessionInfoRsp{
				BaseLayer: layers.BaseLayer{
					Contents: []byte{0x16, 0x08, 0x04, 0x1, 0x2, 0x11},
					Payload:  []byte{},
				},
				Handle:         22,
				Max:            8,
				Active:         4,
				UserID:         1,
				PrivilegeLevel: PrivilegeLevelUser,
				IsIPMIv2:       true,
				Channel:        Channel(1),
			},
		},
		{
			[]byte{0x16, 0x08, 0x04, 0x1, 0x2, 0x11, 0x1},
			&GetSessionInfoRsp{
				BaseLayer: layers.BaseLayer{
					Contents: []byte{0x16, 0x08, 0x04, 0x1, 0x2, 0x11},
					Payload:  []byte{0x1},
				},
				Handle:         22,
				Max:            8,
				Active:         4,
				UserID:         1,
				PrivilegeLevel: PrivilegeLevelUser,
				IsIPMIv2:       true,
				Channel:        Channel(1),
			},
		},
		{
			[]byte{
				0xfd, 0xf, 0xe, 0x16, 0x4, 0x0f,
				0xa, 0x16, 0x1, 0x3,
				0xd3, 0xf5, 0xfb, 0xbf, 0x83, 0xed,
				0x55, 0xf6,
				0x0,
			},
			&GetSessionInfoRsp{
				BaseLayer: layers.BaseLayer{
					Contents: []byte{
						0xfd, 0xf, 0xe, 0x16, 0x4, 0xf,
						0xa, 0x16, 0x1, 0x3,
						0xd3, 0xf5, 0xfb, 0xbf, 0x83, 0xed,
						0x55, 0xf6,
					},
					Payload: []byte{0x0},
				},
				Handle:         253,
				Max:            15,
				Active:         14,
				UserID:         22,
				PrivilegeLevel: PrivilegeLevelAdministrator,
				IsIPMIv2:       false,
				Channel:        Channel(15),
				IP:             net.IPv4(10, 22, 1, 3),
				MAC:            net.HardwareAddr{0xd3, 0xf5, 0xfb, 0xbf, 0x83, 0xed},
				Port:           63061,
			},
		},
	}
	layer := &GetSessionInfoRsp{}
	for _, test := range table {
		err := layer.DecodeFromBytes(test.in, gopacket.NilDecodeFeedback)
		switch {
		case err == nil && test.want == nil:
			t.Errorf("expected error decoding %v, got none", test.in)
		case err == nil && test.want != nil:
			if diff := cmp.Diff(test.want, layer); diff != "" {
				t.Errorf("decode %v = %v, want %v: %v", test.in, layer, test.want, diff)
			}
		case err != nil && test.want != nil:
			t.Errorf("unexpected error: %v", err)
		}
	}
}
