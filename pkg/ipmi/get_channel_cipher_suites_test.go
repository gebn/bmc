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
				Channel: ChannelPresentInterface,
			},
			[]byte{0x0e, 0x00, 0x80},
		},
		{
			&GetChannelCipherSuitesReq{
				Channel:     ChannelPrimaryIPMB,
				PayloadType: PayloadTypeOEM,
				ListIndex:   63,
			},
			[]byte{0x00, 0x02, 0xbf},
		},
		// deliberately use values out of range to check truncation
		{
			&GetChannelCipherSuitesReq{
				Channel:     0xff,
				PayloadType: 0xff,
				ListIndex:   0x7f,
			},
			[]byte{0x0f, 0x3f, 0xbf},
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
			t.Errorf("serialize %v failed with %v, wanted %#v", test.layer, err, test.want)
		case err == nil && !bytes.Equal(got, test.want):
			t.Errorf("serialize %v = %#v, want %#v", test.layer, got, test.want)
		}
	}
}

func TestGetChannelCipherSuitesRspDecodeFromBytes(t *testing.T) {
	table := []struct {
		data []byte
		want *GetChannelCipherSuitesRsp
	}{
		// no record data, which could happen if the cipher suite records
		// happen to end on a 16-byte boundary
		{
			[]byte{
				0x00,
			},
			&GetChannelCipherSuitesRsp{
				BaseLayer: layers.BaseLayer{
					Contents: []byte{
						0x00,
					},
					Payload: []byte{},
				},
				Channel:                 ChannelPrimaryIPMB,
				CipherSuiteRecordsChunk: []byte{},
			},
		},
		// record data ending mid-way through a response
		{
			[]byte{
				0x01,
				0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07,
			},
			&GetChannelCipherSuitesRsp{
				BaseLayer: layers.BaseLayer{
					Contents: []byte{
						0x01,
						0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07,
					},
					Payload: []byte{},
				},
				Channel:                 1,
				CipherSuiteRecordsChunk: []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07},
			},
		},
		// a full response of data, which will be the case for all possibly up
		// to the last
		{
			[]byte{
				0x01,
				0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07,
				0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f,
			},
			&GetChannelCipherSuitesRsp{
				BaseLayer: layers.BaseLayer{
					Contents: []byte{
						0x01,
						0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07,
						0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f,
					},
					Payload: []byte{},
				},
				Channel: 1,
				CipherSuiteRecordsChunk: []byte{
					0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07,
					0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f,
				},
			},
		},
		// full response will trailing data
		{
			[]byte{
				0x01,
				0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07,
				0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f,
				0x00, 0x01, 0x02,
			},
			&GetChannelCipherSuitesRsp{
				BaseLayer: layers.BaseLayer{
					Contents: []byte{
						0x01,
						0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07,
						0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f,
					},
					Payload: []byte{
						0x00, 0x01, 0x02,
					},
				},
				Channel: 1,
				CipherSuiteRecordsChunk: []byte{
					0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07,
					0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f,
				},
			},
		},
	}
	layer := &GetChannelCipherSuitesRsp{}
	for _, test := range table {
		err := layer.DecodeFromBytes(test.data, gopacket.NilDecodeFeedback)
		switch {
		case err != nil && test.want != nil:
			t.Errorf("decode %#v failed with %v, wanted %v", test.data, err, test.want)
		case err == nil:
			if diff := cmp.Diff(test.want, layer); diff != "" {
				t.Errorf("decode %#v = %v, want %v: %v", test.data, layer, test.want, diff)
			}
		}
	}
}
