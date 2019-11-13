package ipmi

import (
	"encoding/hex"
	"testing"
)

type idStringTest struct {
	encoded  []byte
	chars    int
	decoded  string
	consumed int
	err      bool
}

func TestDecodeBCDPlus(t *testing.T) {
	tests := []idStringTest{
		{
			[]byte{0x12},
			4,
			"",
			0,
			true,
		},
		{
			[]byte{0x01, 0x23, 0x45, 0x67, 0x89, 0xab, 0xcd, 0xef},
			16,
			"0123456789 -.:,_",
			8,
			false,
		},
		{
			[]byte{0x22, 0xb5, 0x6d, 0xab, 0x34},
			10,
			"22-56: -34",
			5,
			false,
		},
		{
			[]byte{0x33, 0x30},
			3,
			"333",
			2,
			false,
		},
	}
	for _, test := range tests {
		decoded, consumed, err := decodeBCDPlus(test.encoded, test.chars)
		if (test.err && err == nil) || decoded != test.decoded || consumed != test.consumed {
			t.Errorf("decodeBCDPlus(%v, %v) gave (%v, %v, %v), wanted (%v, %v, %v)",
				hex.EncodeToString(test.encoded), test.chars,
				decoded, consumed, err,
				test.decoded, test.consumed, test.err)
		}
	}
}

func TestDecodePacked6BitAscii(t *testing.T) {
	tests := []idStringTest{
		{
			// data too short for chars
			[]byte{},
			2,
			"",
			0,
			true,
		},
		{
			// I    0x29    0b101001
			// P    0x30    0b110000
			// M    0x2d    0b101101
			// I    0x29    0b101001
			[]byte{0b00101001, 0b11011100, 0b10100110},
			4,
			"IPMI",
			3,
			false,
		},
		{
			// G    0x27    0b100111
			// E    0x25    0b100101
			// O    0x2f    0b101111
			// R    0x32    0b110010
			// G    0x27    0b100111
			// E    0x25    0b100101
			[]byte{0b01100111, 0b11111001, 0b11001010, 0b01100111, 0b00001001},
			6,
			"GEORGE",
			5,
			false,
		},
	}
	for _, test := range tests {
		decoded, consumed, err := decodePacked6BitAscii(test.encoded, test.chars)
		if (test.err && err == nil) || decoded != test.decoded || consumed != test.consumed {
			t.Errorf("decodePacked6BitAscii(%v, %v) gave (%v, %v, %v), wanted (%v, %v, %v)",
				hex.EncodeToString(test.encoded), test.chars,
				decoded, consumed, err,
				test.decoded, test.consumed, test.err)
		}
	}
}

func TestDecode8BitAsciiLatin1(t *testing.T) {
	tests := []idStringTest{
		{
			// must have >=2 bytes
			[]byte{0x01},
			1,
			"",
			0,
			true,
		},
		{
			// too short for chars
			[]byte{0x01, 0x02},
			3,
			"",
			0,
			true,
		},
		{
			[]byte(`:K;&e7-uN 8O3Fd0k?nECU\ctu3}"M5o`),
			32,
			`:K;&e7-uN 8O3Fd0k?nECU\ctu3}"M5o`,
			32,
			false,
		},
		{
			[]byte("HKB}_1P?%|;;drG"),
			10,
			"HKB}_1P?%|",
			10,
			false,
		},
	}
	for _, test := range tests {
		decoded, consumed, err := decode8BitAsciiLatin1(test.encoded, test.chars)
		if (test.err && err == nil) || decoded != test.decoded || consumed != test.consumed {
			t.Errorf("decode8BitAsciiLatin1(%v, %v) gave (%v, %v, %v), wanted (%v, %v, %v)",
				hex.EncodeToString(test.encoded), test.chars,
				decoded, consumed, err,
				test.decoded, test.consumed, test.err)
		}
	}
}
