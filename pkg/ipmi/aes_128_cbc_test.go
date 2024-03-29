package ipmi

import (
	"bytes"
	"testing"

	"github.com/google/gopacket"
)

func TestAES128CBCDecodeFromBytes(t *testing.T) {
	table := []struct {
		key     [16]byte
		data    []byte // IV + encrypted message and trailer
		message []byte // decrypted
	}{
		// 7 bytes data (< block size)
		{
			key: [16]byte{
				0x0e, 0xd9, 0x8c, 0x34, 0xac, 0x8f, 0x34, 0xce, 0x4d, 0xd7,
				0xd9, 0x05, 0x12, 0xb0, 0xf9, 0x7a,
			},
			data: []byte{
				0x4d, 0x15, 0x80, 0x8c, 0x3e, 0xee, 0x67, 0xd6, 0x3b, 0x1c,
				0xb0, 0xd1, 0xae, 0x76, 0xdf, 0xcb,
				0xf3, 0x13, 0xa7, 0xbe, 0x62, 0x58, 0x14, 0xa4, 0x7d, 0xa0,
				0xf6, 0x6f, 0xdf, 0x21, 0xcc, 0xba,
			},
			message: []byte{
				0x7b, 0xec, 0x46, 0xd5, 0xbb, 0x90, 0xba,
			},
		},
		// 15 bytes of data encrypted with 0 bytes of padding (1 byte taken by
		// padding length)
		{
			key: [16]byte{
				0x6f, 0x9c, 0xad, 0xa3, 0x92, 0xa3, 0xbb, 0x12, 0x8d, 0xdb,
				0x49, 0x5f, 0xc8, 0x2a, 0x17, 0x21,
			},
			data: []byte{
				0x94, 0x1e, 0xf9, 0x18, 0xb0, 0x06, 0xd0, 0x84, 0x26, 0xa1,
				0xe2, 0x72, 0x22, 0x37, 0x0b, 0x0f,
				0x7b, 0x74, 0x2d, 0x86, 0x97, 0x42, 0xd8, 0x64, 0x25, 0x5f,
				0x4d, 0xad, 0x2e, 0x14, 0x6b, 0x23,
			},
			message: []byte{
				0xf1, 0xc7, 0xed, 0xfa, 0xc8, 0xf1, 0xa5, 0x40, 0xcd, 0xc4,
				0x3a, 0x3c, 0x9b, 0x30, 0x81,
			},
		},
		// 15 bytes of data encrypted with 16 bytes of padding.
		// This is not how AES128CBC.SerializeTo would encrypt this message, but
		// other implementations of AES in CBC mode may work this way (e.g.
		// OpenSSL).
		{
			key: [16]byte{
				0x12, 0xd4, 0x51, 0x8d, 0x94, 0x2e, 0x28, 0x78, 0x6a, 0x75,
				0x8b, 0xf5, 0xbe, 0x25, 0xaf, 0xf9,
			},

			data: []byte{
				0xa2, 0x56, 0x33, 0xf7, 0xe2, 0xb4, 0x12, 0x33, 0xb8, 0xb,
				0xfb, 0xde, 0x47, 0x66, 0xa8, 0x9e,
				0x7a, 0xb7, 0xca, 0x4b, 0x3d, 0xb7, 0x8a, 0xf9, 0xc9, 0x5,
				0xaf, 0x3, 0xac, 0xb4, 0xae, 0xc8, 0xdb, 0x37, 0xb8, 0x42,
				0x2b, 0x62, 0x44, 0x3f, 0x33, 0x29, 0x52, 0x85, 0xd2, 0x11,
				0x73, 0xa,
			},
			message: []byte{
				0x66, 0x66, 0x68, 0x36, 0x53, 0x42, 0x34, 0x35, 0x6b, 0x6d,
				0x2c, 0x30, 0x39, 0x2d, 0x33,
			},
		},
		// 16 bytes data (must flow onto second block for padding length)
		{
			key: [16]byte{
				0x02, 0x4a, 0x88, 0x40, 0xdd, 0x55, 0x04, 0xfb, 0xc9, 0x2e,
				0x9d, 0xff, 0x83, 0x58, 0xe4, 0x8d,
			},
			data: []byte{
				0x34, 0x0b, 0xd5, 0x65, 0xdf, 0x54, 0x66, 0xfd, 0xf0, 0x9c,
				0x73, 0x23, 0x9c, 0xfc, 0x00, 0xfe,
				0x57, 0x08, 0x00, 0xf6, 0x0d, 0x00, 0x05, 0x4a, 0x3f, 0xf3,
				0xc8, 0xba, 0x42, 0x51, 0x7c, 0xa1,
				0x31, 0x8f, 0x8d, 0x4a, 0xed, 0x81, 0x6b, 0xa9, 0x24, 0x4e,
				0x44, 0x8e, 0xed, 0x42, 0x7b, 0xda,
			},
			message: []byte{
				0x8a, 0xa9, 0xe3, 0x30, 0x62, 0xdf, 0x5b, 0x48, 0xa7, 0x0d,
				0x77, 0xa3, 0xb8, 0xa2, 0x22, 0x3a,
			},
		},
		// 35 bytes data (multiple blocks)
		{
			key: [16]byte{
				0x16, 0x27, 0xf9, 0x99, 0xcb, 0xe2, 0xf8, 0x62, 0x3e, 0x61,
				0xa3, 0xcc, 0xfe, 0x58, 0x9d, 0xc5,
			},
			data: []byte{
				0x92, 0x04, 0x37, 0x5f, 0x06, 0x5d, 0x00, 0xaa, 0xb3, 0xf9,
				0x59, 0xc6, 0x0b, 0xab, 0x07, 0x28,
				0xc8, 0xa0, 0xf7, 0x77, 0x34, 0xd6, 0xdc, 0xd0, 0xa5, 0xd1,
				0x39, 0x96, 0xc7, 0x34, 0x6a, 0x65,
				0xe5, 0xf7, 0x5e, 0xfc, 0xc7, 0x37, 0x2d, 0x01, 0x32, 0x83,
				0xc9, 0x18, 0x51, 0x03, 0xab, 0xa2,
				0xa6, 0xef, 0x87, 0x71, 0xc1, 0x2f, 0xef, 0x87, 0xc4, 0x38,
				0x60, 0xe9, 0x09, 0x50, 0x7e, 0xc6,
			},
			message: []byte{
				0x7a, 0x81, 0xd6, 0x53, 0x96, 0x09, 0x95, 0xe5, 0x4c, 0x0c,
				0x4b, 0xac, 0xd5, 0xa7, 0x1f, 0x8b,
				0xac, 0x9c, 0x4d, 0xc3, 0x31, 0xd0, 0xa0, 0x1b, 0x50, 0x25,
				0x8e, 0x3e, 0x5d, 0x5e, 0x52, 0xed,
				0x5e, 0x3f, 0x2e,
			},
		},
		// invalid padding
		{
			key: [16]byte{
				0x12, 0xd4, 0x51, 0x8d, 0x94, 0x2e, 0x28, 0x78, 0x6a, 0x75,
				0x8b, 0xf5, 0xbe, 0x25, 0xaf, 0xf9,
			},
			data: []byte{
				0x4e, 0x86, 0xa3, 0x8e, 0xcb, 0x0f, 0x1b, 0xe9, 0xac, 0x46,
				0x73, 0x76, 0xc8, 0x96, 0x04, 0x32,
				0x67, 0xaa, 0xa3, 0x84, 0x33, 0xe8, 0xcb, 0x63, 0x66, 0x46,
				0xce, 0x1f, 0x14, 0xf5, 0xaf, 0x16,
			},
			// nil message to indicate error (first pad byte 0x00, not 0x01)
			// plaintext:
			// 3a b8 39 8f 55 c3 30 b8 30 5f 00 01 02 03 04 05
		},
	}
	for _, test := range table {
		layer, err := NewAES128CBC(test.key)
		if err != nil {
			t.Errorf("error creating cipher with key %v: %v", test.key, err)
			continue
		}
		err = layer.DecodeFromBytes(test.data, gopacket.NilDecodeFeedback)
		switch {
		case err == nil && test.message == nil:
			t.Errorf("expected error decoding %v, got none", test.data)
		case err == nil && test.message != nil:
			if !bytes.Equal(layer.Payload, test.message) {
				t.Errorf("decode %v = %v, got %v", test.data, layer.Payload,
					test.message)
			}
		case err != nil && test.message != nil:
			t.Errorf("unexpected error: %v", err)
		}
	}
}
