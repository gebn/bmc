package ipmi

import (
	"bytes"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
)

func TestMessage(t *testing.T) {
	table := []struct {
		layer   *Message
		payload []byte // content to put into the SerializeBuffer
		wire    []byte
	}{
		{ // app request
			&Message{
				BaseLayer: layers.BaseLayer{
					Contents: []byte{0x20, 0x18, 0xc8, 0x81, 0xbe, 0x38}, // final checksum omitted
					Payload:  []byte{},
				},
				Operation: Operation{
					Function: 0x6,
					Command:  0x38,
				},
				RemoteAddress: SlaveAddressBMC.Address(),
				RemoteLUN:     0x0,
				LocalAddress:  SoftwareIDRemoteConsole1.Address(),
				LocalLUN:      0x2,
				Sequence:      0x2f,
			},
			[]byte{},
			[]byte{0x20, 0x18, 0xc8, 0x81, 0xbe, 0x38, 0x89},
		},
		{ // group request
			&Message{
				BaseLayer: layers.BaseLayer{
					Contents: []byte{0x24, 0xb1, 0x2b, 0x23, 0xfe, 0x9f, 0xdc},
					Payload:  []byte{0x1, 0x2},
				},
				Operation: Operation{
					Function: 0x2c,
					Command:  0x9f,
					Body:     0xdc,
				},
				RemoteAddress: SlaveAddress(0x12).Address(),
				RemoteLUN:     0x1,
				LocalAddress:  SoftwareID(0x11).Address(),
				LocalLUN:      0x2,
				Sequence:      0x3f,
			},
			[]byte{0x1, 0x2},
			[]byte{0x24, 0xb1, 0x2b, 0x23, 0xfe, 0x9f, 0xdc, 0x1, 0x2, 0x61},
		},
		{ // oem request
			&Message{
				BaseLayer: layers.BaseLayer{
					Contents: []byte{0x4d, 0xba, 0xf9, 0x9f, 0x3, 0x98, 0x8e, 0xe8, 0x21},
					Payload:  []byte{0x2, 0x1},
				},
				Operation: Operation{
					Function:   0x2e,
					Command:    0x98,
					Enterprise: 2222222,
				},
				RemoteAddress: 0x4d,
				RemoteLUN:     0x2,
				LocalAddress:  0x9f,
				LocalLUN:      0x3,
				Sequence:      0x0,
			},
			[]byte{0x2, 0x1},
			[]byte{0x4d, 0xba, 0xf9, 0x9f, 0x3, 0x98, 0x8e, 0xe8, 0x21, 0x2, 0x1, 0x2c},
		},
		{ // app response
			&Message{
				BaseLayer: layers.BaseLayer{
					Contents: []byte{0x20, 0x1c, 0xc4, 0x81, 0xbe, 0x38, 0x22},
					Payload:  []byte{},
				},
				Operation: Operation{
					Function: 0x7,
					Command:  0x38,
				},
				RemoteAddress:  SlaveAddressBMC.Address(),
				RemoteLUN:      0x0,
				LocalAddress:   SoftwareIDRemoteConsole1.Address(),
				LocalLUN:       0x2,
				Sequence:       0x2f,
				CompletionCode: 0x22,
			},
			[]byte{},
			[]byte{0x20, 0x1c, 0xc4, 0x81, 0xbe, 0x38, 0x22, 0x67},
		},
		{ // group response
			&Message{
				BaseLayer: layers.BaseLayer{
					Contents: []byte{0x24, 0xb5, 0x27, 0x23, 0xfe, 0x9f, 0x0, 0xdc},
					Payload:  []byte{0x1, 0x2},
				},
				Operation: Operation{
					Function: 0x2d,
					Command:  0x9f,
					Body:     0xdc,
				},
				RemoteAddress:  SlaveAddress(0x12).Address(),
				RemoteLUN:      0x1,
				LocalAddress:   SoftwareID(0x11).Address(),
				LocalLUN:       0x2,
				Sequence:       0x3f,
				CompletionCode: 0x0,
			},
			[]byte{0x1, 0x2},
			[]byte{0x24, 0xb5, 0x27, 0x23, 0xfe, 0x9f, 0x0, 0xdc, 0x1, 0x2, 0x61},
		},
		{ // oem response
			&Message{
				BaseLayer: layers.BaseLayer{
					Contents: []byte{0x4d, 0xbe, 0xf5, 0x9f, 0x3, 0x98, 0xff, 0x8e, 0xe8, 0x21},
					Payload:  []byte{0x2, 0x1},
				},
				Operation: Operation{
					Function:   0x2f,
					Command:    0x98,
					Enterprise: 2222222,
				},
				RemoteAddress:  0x4d,
				RemoteLUN:      0x2,
				LocalAddress:   0x9f,
				LocalLUN:       0x3,
				Sequence:       0x0,
				CompletionCode: 0xff,
			},
			[]byte{0x2, 0x1},
			[]byte{0x4d, 0xbe, 0xf5, 0x9f, 0x3, 0x98, 0xff, 0x8e, 0xe8, 0x21, 0x2, 0x1, 0x2d},
		},
		{
			&Message{
				BaseLayer: layers.BaseLayer{
					Contents: []byte{0x20, 0x18, 0xc8, 0x81, 0, 0x37},
					Payload:  []byte{},
				},
				Operation: Operation{
					Function: 0x6,
					Command:  0x37,
				},
				RemoteAddress: SlaveAddressBMC.Address(),
				RemoteLUN:     0x0,
				LocalAddress:  SoftwareIDRemoteConsole1.Address(),
				LocalLUN:      0x0,
				Sequence:      0,
			},
			[]byte{},
			[]byte{0x20, 0x18, 0xc8, 0x81, 0, 0x37, 0x48},
		},
	}
	for _, test := range table {
		sb := gopacket.NewSerializeBuffer()
		payload, _ := sb.PrependBytes(len(test.payload))
		copy(payload, test.payload)
		serializeErr := test.layer.SerializeTo(sb, gopacket.SerializeOptions{
			ComputeChecksums: true,
		})
		got := sb.Bytes()

		switch {
		case serializeErr != nil:
			t.Errorf("serialize %v failed with %v, wanted %v", test.layer,
				serializeErr, test.wire)
		case !bytes.Equal(got, test.wire):
			t.Errorf("serialize %v = %v, want %v", test.layer, got, test.wire)
		}

		decoded := &Message{}
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

func TestChecksum(t *testing.T) {
	table := []struct {
		bytes []byte
		want  uint8
	}{
		{[]byte{0, 1, 2, 0xe}, 0xef},
		{[]byte{0x80, 0x15, 0x1, 0x8, 0x30, 0x33, 0x31, 0x35, 0x31, 0x30, 0x33, 0x30, 0x2, 0x9, 0x30, 0x33, 0x35, 0x31, 0x2d, 0x33, 0x32, 0x31, 0x30}, 0xe},
		{[]byte{0x20, 0x18}, 0xc8}, // taken from FreeIPMI example at https://www.gnu.org/software/freeipmi/freeipmi-design.txt
		{[]byte{0x20, 0x0, 0x37, 0x0, 0x43, 0x30, 0x30, 0x31, 0x4d, 0x53, 0xc, 0xc4, 0x7a, 0x37, 0x86, 0x5f, 0x0, 0x0, 0x0, 0x0}, 0xcf},
		{[]byte{0x81, 0, 0x38, 0xe, 0x4}, 0x35},
		{[]byte{0x81, 0, 0x38, 0x8e, 0x4}, 0xb5},
	}

	for _, test := range table {
		got := checksum(test.bytes)
		if got != test.want {
			t.Errorf("checksum(%#v) = %#x, want %#x", test.bytes, got, test.want)
		}
	}
}
