package ipmi

import (
	"bytes"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
)

func TestGetChannelAuthenticationCapabilitiesReqSerializeTo(t *testing.T) {
	table := []struct {
		layer *GetChannelAuthenticationCapabilitiesReq
		want  []byte
	}{
		{
			&GetChannelAuthenticationCapabilitiesReq{
				ExtendedData:      true,
				Channel:           ChannelPrimaryIPMB,
				MaxPrivilegeLevel: PrivilegeLevelAdministrator,
			},
			[]byte{0x80, 0x04},
		},
		{
			&GetChannelAuthenticationCapabilitiesReq{
				ExtendedData:      false,
				Channel:           ChannelPresentInterface,
				MaxPrivilegeLevel: PrivilegeLevelUser,
			},
			[]byte{0x0e, 0x02},
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

func TestGetChannelAuthenticationCapabilitiesRspDecodeFromBytes(t *testing.T) {
	table := []struct {
		data []byte
		want *GetChannelAuthenticationCapabilitiesRsp
	}{
		{
			[]byte{0x0, 0x15, 0x15, 0x1, 0x3, 0x2, 0x1, 0x22},
			&GetChannelAuthenticationCapabilitiesRsp{
				BaseLayer: layers.BaseLayer{
					Contents: []byte{0x0, 0x15, 0x15, 0x1, 0x3, 0x2, 0x1,
						0x22},
					Payload: []byte{},
				},
				Channel:                    ChannelPrimaryIPMB,
				ExtendedCapabilities:       false,
				AuthenticationTypeOEM:      false,
				AuthenticationTypePassword: true,
				AuthenticationTypeMD5:      true,
				AuthenticationTypeMD2:      false,
				AuthenticationTypeNone:     true,
				TwoKeyLogin:                false,
				PerMessageAuthentication:   true,
				UserLevelAuthentication:    false,
				NonNullUsernamesEnabled:    true,
				NullUsernamesEnabled:       false,
				AnonymousLoginEnabled:      true,
				SupportsV2:                 false,
				SupportsV1:                 true,
				OEM:                        66051,
				OEMData:                    0x22,
			},
		},
		{
			[]byte{0xe, 0xa2, 0x2a, 0x3, 0x1, 0x2, 0x3, 0xFF, 0x1},
			&GetChannelAuthenticationCapabilitiesRsp{
				BaseLayer: layers.BaseLayer{
					Contents: []byte{0xe, 0xa2, 0x2a, 0x3, 0x1, 0x2, 0x3,
						0xFF},
					Payload: []byte{0x1},
				},
				Channel:                    ChannelPresentInterface,
				ExtendedCapabilities:       true,
				AuthenticationTypeOEM:      true,
				AuthenticationTypePassword: false,
				AuthenticationTypeMD5:      false,
				AuthenticationTypeMD2:      true,
				AuthenticationTypeNone:     false,
				TwoKeyLogin:                true,
				PerMessageAuthentication:   false,
				UserLevelAuthentication:    true,
				NonNullUsernamesEnabled:    false,
				NullUsernamesEnabled:       true,
				AnonymousLoginEnabled:      false,
				SupportsV2:                 true,
				SupportsV1:                 true,
				OEM:                        197121,
				OEMData:                    0xFF,
			},
		},
	}
	layer := &GetChannelAuthenticationCapabilitiesRsp{}
	for _, test := range table {
		err := layer.DecodeFromBytes(test.data, gopacket.NilDecodeFeedback)
		switch {
		case err != nil && test.want != nil:
			t.Errorf("decode %v failed with %v, wanted %v", test.data, err, test.want)
		case err == nil:
			if diff := cmp.Diff(test.want, layer); diff != "" {
				t.Errorf("decode %v = %v, want %v: %v", test.data, layer, test.want, diff)
			}
		}
	}
}
