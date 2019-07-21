package ipmi

import (
	"bytes"
	"testing"

	"github.com/google/gopacket"
)

func TestRAKPMessage3SerializeTo(t *testing.T) {
	table := []struct {
		layer *RAKPMessage3
		wire  []byte
	}{
		{
			// success
			&RAKPMessage3{
				Tag:                    0x01,
				Status:                 0x00,
				ManagedSystemSessionID: 0x4030201,
				AuthCode:               []byte{0x2, 0x1},
			},
			[]byte{0x01, 0x00, 0x00, 0x00, 0x01, 0x02, 0x03, 0x04, 0x2, 0x1},
		},
		{
			// failure
			&RAKPMessage3{
				Tag:                    0x00,
				Status:                 0x02,
				ManagedSystemSessionID: 0x1020304,
			},
			[]byte{0x00, 0x02, 0x00, 0x00, 0x4, 0x03, 0x02, 0x01},
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
