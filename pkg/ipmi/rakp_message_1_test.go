package ipmi

import (
	"bytes"
	"testing"

	"github.com/google/gopacket"
)

func TestRAKPMessage1SerializeTo(t *testing.T) {
	table := []struct {
		layer *RAKPMessage1
		wire  []byte
	}{
		{
			// empty username, role-based lookup
			&RAKPMessage1{
				Tag:                    0x22,
				ManagedSystemSessionID: 0x1020304,
				RemoteConsoleRandom: [16]byte{
					0x0, 0x1, 0x2, 0x3, 0x4, 0x5, 0x6, 0x7,
					0x8, 0x9, 0xa, 0xb, 0xc, 0xd, 0xe, 0xf},
				PrivilegeLevelLookup: true,
				MaxPrivilegeLevel:    PrivilegeLevelUser,
			},
			[]byte{
				0x22, 0x00, 0x00, 0x00, 0x04, 0x03, 0x02, 0x01, 0x0, 0x1, 0x2,
				0x3, 0x4, 0x5, 0x6, 0x7, 0x8, 0x9, 0xa, 0xb, 0xc, 0xd, 0xe, 0xf,
				0x02, 0x00, 0x00, 0x00},
		},
		{
			// non-empty username, username only lookup
			&RAKPMessage1{
				Tag:                    0x1,
				ManagedSystemSessionID: 0x4030201,
				RemoteConsoleRandom: [16]byte{
					0x0, 0x1, 0x2, 0x3, 0x4, 0x5, 0x6, 0x7,
					0x8, 0x9, 0xa, 0xb, 0xc, 0xd, 0xe, 0xf},
				MaxPrivilegeLevel: PrivilegeLevelAdministrator,
				Username:          "george",
			},
			[]byte{
				0x1, 0x00, 0x00, 0x00, 0x01, 0x02, 0x03, 0x04, 0x0, 0x1, 0x2,
				0x3, 0x4, 0x5, 0x6, 0x7, 0x8, 0x9, 0xa, 0xb, 0xc, 0xd, 0xe, 0xf,
				0x14, 0x00, 0x00, 0x06, 'g', 'e', 'o', 'r', 'g', 'e'},
		},
		{
			// non-empty username, role-based lookup
			&RAKPMessage1{
				Tag:                    0xff,
				ManagedSystemSessionID: 0x1040203,
				RemoteConsoleRandom: [16]byte{
					0x0, 0x1, 0x2, 0x3, 0x4, 0x5, 0x6, 0x7,
					0x8, 0x9, 0xa, 0xb, 0xc, 0xd, 0xe, 0xf},
				PrivilegeLevelLookup: true,
				MaxPrivilegeLevel:    PrivilegeLevelOperator,
				Username:             "george",
			},
			[]byte{
				0xff, 0x00, 0x00, 0x00, 0x03, 0x02, 0x04, 0x01, 0x0, 0x1, 0x2,
				0x3, 0x4, 0x5, 0x6, 0x7, 0x8, 0x9, 0xa, 0xb, 0xc, 0xd, 0xe, 0xf,
				0x03, 0x00, 0x00, 0x06, 'g', 'e', 'o', 'r', 'g', 'e'},
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

		var decoded RAKPMessage1
		if err := decoded.DecodeFromBytes(got, nil); err != nil {
			t.Errorf("decode %v = error %v", test.layer, err)
			continue
		}
		r := test.layer
		if r.Tag != decoded.Tag {
			t.Errorf("decode %v Tag = got %v, want %v", test.layer, decoded.Tag, r.Tag)
		}
		if r.ManagedSystemSessionID != decoded.ManagedSystemSessionID {
			t.Errorf("decode %v ManagedSystemSessionID = got %v, want %v", test.layer, decoded.ManagedSystemSessionID, r.ManagedSystemSessionID)
		}
		if r.RemoteConsoleRandom != decoded.RemoteConsoleRandom {
			t.Errorf("decode %v RemoteConsoleRandom = got %v, want %v", test.layer, decoded.RemoteConsoleRandom, r.RemoteConsoleRandom)
		}
		if r.MaxPrivilegeLevel != decoded.MaxPrivilegeLevel {
			t.Errorf("decode %v MaxPrivilegeLevel = got %v, want %v", test.layer, decoded.MaxPrivilegeLevel, r.MaxPrivilegeLevel)
		}
		if r.PrivilegeLevelLookup != decoded.PrivilegeLevelLookup {
			t.Errorf("decode %v PrivilegeLevelLookup = got %v, want %v", test.layer, decoded.PrivilegeLevelLookup, r.PrivilegeLevelLookup)
		}
		if r.Username != decoded.Username {
			t.Errorf("decode %v Username = got %v, want %v", test.layer, decoded.Username, r.Username)
		}
	}
}
