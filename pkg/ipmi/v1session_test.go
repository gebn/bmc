package ipmi

import (
	"bytes"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
)

func TestV1Session(t *testing.T) {
	table := []struct {
		layer *V1Session
		wire  []byte
	}{
		{
			&V1Session{
				BaseLayer: layers.BaseLayer{
					Contents: []byte{0x0, 0x0, 0x0, 0x0, 0x40, 0x0, 0x0, 0x0,
						0x20, 0x3},
					Payload: []byte{0x0, 0x0, 0x0},
				},
				AuthType: AuthenticationTypeNone,
				Sequence: 1073741824,
				ID:       536870912,
				Length:   3,
			},
			[]byte{0x0, 0x0, 0x0, 0x0, 0x40, 0x0, 0x0, 0x0, 0x20, 0x3, 0, 0, 0},
		},
		{
			&V1Session{
				BaseLayer: layers.BaseLayer{
					Contents: []byte{0x6, 0xa3, 0x8, 0x0, 0x0, 0x62, 0x4, 0x0,
						0x0, 0x1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
					Payload: []byte{},
				},
				AuthType: AuthenticationTypeRMCPPlus,
				Sequence: 2211,
				ID:       1122,
				AuthCode: [16]byte{0x1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
				Length:   0,
			},
			[]byte{0x6, 0xa3, 0x8, 0x0, 0x0, 0x62, 0x4, 0x0, 0x0, 0x1,
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		},
	}
	for _, test := range table {
		sb := gopacket.NewSerializeBuffer()
		sb.PrependBytes(int(test.layer.Length))
		serializeErr := test.layer.SerializeTo(sb, gopacket.SerializeOptions{
			FixLengths: true,
		})
		got := sb.Bytes()

		switch {
		case serializeErr != nil:
			t.Errorf("serialize %v failed with %v, wanted %v", test.layer,
				serializeErr, test.wire)
		case !bytes.Equal(got, test.wire):
			t.Errorf("serialize %v = %v, want %v", test.layer, got, test.wire)
		}

		decoded := &V1Session{}
		decodeErr := decoded.DecodeFromBytes(got, gopacket.NilDecodeFeedback)
		switch {
		case decodeErr != nil:
			t.Errorf("decode %v failed with %v, wanted %v", got, decodeErr,
				test.layer)
		default:
			if diff := cmp.Diff(test.layer, decoded); diff != "" {
				t.Errorf("decode %v = %v, want %v: %v", got, decoded, test.layer, diff)
			}
		}
	}
}
