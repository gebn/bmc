package ipmi

import (
	"bytes"
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
)

type dummyHasher []byte

func (dummyHasher) Write(_ []byte) (int, error) {
	return 0, errors.New("not writeable")
}

func (d dummyHasher) Sum(_ []byte) []byte {
	return []byte(d)
}

func (dummyHasher) Reset() {}

func (d dummyHasher) Size() int {
	return len(d)
}

func (d dummyHasher) BlockSize() int {
	return d.Size()
}

var hasher = dummyHasher([]byte{0x01, 0x02, 0x03, 0x04})

func TestV2SessionDecode(t *testing.T) {
	table := []struct {
		layer *V2Session
		wire  []byte
	}{
		{
			// not an IPMIv2 packet
			nil,
			[]byte{0x3, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0},
		},
		{
			// not encrypted, not authenticated, 1 byte payload
			&V2Session{
				BaseLayer: layers.BaseLayer{
					Contents: []byte{0x6, 0x0, 0x1, 0x2, 0x3, 0x4, 0x4, 0x3,
						0x2, 0x1, 0x1, 0x0},
					Payload: []byte{0x0},
				},
				ID:                 0x4030201,
				Sequence:           0x1020304,
				Length:             1,
				IntegrityAlgorithm: hasher,
			},
			[]byte{0x6, 0x0, 0x1, 0x2, 0x3, 0x4, 0x4, 0x3, 0x2, 0x1, 0x1, 0x0,
				0x0},
		},
		{
			// not encrypted, not authenticated, OEM payload, 1 byte payload
			&V2Session{
				BaseLayer: layers.BaseLayer{
					Contents: []byte{0x6, 0x20, 0x1, 0x2, 0x3, 0x4, 0x4, 0x3,
						0x2, 0x1, 0x1, 0x0},
					Payload: []byte{0x0},
				},
				PayloadDescriptor: PayloadDescriptor{
					PayloadType: 0x20, // first OEM reserved
				},
				ID:                 0x4030201,
				Sequence:           0x1020304,
				Length:             1,
				IntegrityAlgorithm: hasher,
			},
			[]byte{0x6, 0x20, 0x1, 0x2, 0x3, 0x4, 0x4, 0x3, 0x2, 0x1, 0x1, 0x0,
				0x0},
		},
		{
			// not encrypted, not authenticated, OEM, 1 byte payload
			&V2Session{
				BaseLayer: layers.BaseLayer{
					Contents: []byte{0x6, 0x2, 0xa2, 0x2, 0x0, 0x0, 0x1, 0x2,
						0x1, 0x2, 0x3, 0x4, 0x4, 0x3, 0x2, 0x1, 0x1, 0x0},
					Payload: []byte{0x0},
				},
				PayloadDescriptor: PayloadDescriptor{
					PayloadType: 0x2,
					Enterprise:  674,
					PayloadID:   0x201,
				},
				ID:                 0x4030201,
				Sequence:           0x1020304,
				Length:             1,
				IntegrityAlgorithm: hasher,
			},
			[]byte{0x6, 0x2, 0xa2, 0x2, 0x0, 0x0, 0x1, 0x2, 0x1, 0x2, 0x3, 0x4,
				0x4, 0x3, 0x2, 0x1, 0x1, 0x0, 0x0},
		},
		{
			// encrypted, not authenticated, IPMI, 1 byte payload
			&V2Session{
				BaseLayer: layers.BaseLayer{
					Contents: []byte{0x6, 1 << 7, 0x1, 0x2, 0x3, 0x4, 0x4, 0x3,
						0x2, 0x1, 0x1, 0x0},
					Payload: []byte{0x0},
				},
				Encrypted:          true,
				ID:                 0x4030201,
				Sequence:           0x1020304,
				Length:             1,
				IntegrityAlgorithm: hasher,
			},
			[]byte{0x6, 1 << 7, 0x1, 0x2, 0x3, 0x4, 0x4, 0x3, 0x2, 0x1, 0x1,
				0x0, 0x0},
		},
		{
			// not encrypted, authenticated, IPMI, 1 byte payload
			&V2Session{
				BaseLayer: layers.BaseLayer{
					Contents: []byte{0x6, 1 << 6, 0x1, 0x2, 0x3, 0x4, 0x4, 0x3,
						0x2, 0x1, 0x1, 0x0},
					Payload: []byte{0x0},
				},
				Authenticated:      true,
				ID:                 0x4030201,
				Sequence:           0x1020304,
				Length:             1,
				Pad:                1,
				Signature:          []byte{0x1, 0x2, 0x3, 0x4},
				IntegrityAlgorithm: hasher,
			},
			[]byte{0x6, 1 << 6, 0x1, 0x2, 0x3, 0x4, 0x4, 0x3, 0x2, 0x1, 0x1,
				0x0, 0x0, 0xff, 0x1, 0x07, 0x1, 0x2, 0x3, 0x4}, // arbitrary 4-byte auth code
		},
		{
			// encrypted, authenticated, IPMI, 1 byte payload
			&V2Session{
				BaseLayer: layers.BaseLayer{
					Contents: []byte{0x6, 0xc0, 0x1, 0x2, 0x3, 0x4, 0x4, 0x3,
						0x2, 0x1, 0x2, 0x0},
					Payload: []byte{0x0, 0x0},
				},
				Encrypted:          true,
				Authenticated:      true,
				ID:                 0x4030201,
				Sequence:           0x1020304,
				Length:             2,
				Pad:                0,
				Signature:          []byte{0x1, 0x2, 0x3, 0x4},
				IntegrityAlgorithm: hasher,
			},
			[]byte{0x6, 0xc0, 0x1, 0x2, 0x3, 0x4, 0x4, 0x3, 0x2, 0x1, 0x2,
				0x0, 0x0, 0x0, 0x0, 0x07, 0x1, 0x2, 0x3, 0x4}, // arbitrary 4-byte auth code
		},
	}
	session := &V2Session{
		IntegrityAlgorithm: dummyHasher([]byte{0x1, 0x2, 0x3, 0x4}),
	}
	for _, test := range table {
		if test.layer != nil {
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
		}

		decodeErr := session.DecodeFromBytes(test.wire, gopacket.NilDecodeFeedback)
		switch {
		case decodeErr == nil && test.layer == nil:
			t.Errorf("decode %v succeeded with %v, wanted error", test.wire,
				session)
		case decodeErr != nil && test.layer != nil:
			t.Errorf("decode %v failed with %v, wanted %v", test.wire, decodeErr,
				test.layer)
		case decodeErr == nil && test.layer != nil:
			if diff := cmp.Diff(test.layer, session); diff != "" {
				t.Errorf("decode %v = %v, want %v: %v", test.wire, session, test.layer, diff)
			}
		}
	}
}
