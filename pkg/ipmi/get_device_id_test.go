package ipmi

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
)

func TestGetDecideIDRspDecodeFromBytes(t *testing.T) {
	tests := []struct {
		in   []byte
		want *GetDeviceIDRsp
	}{
		{
			[]byte{0x20, 0x81, 0x03, 0x45, 0x02, 0xbf, 0x4c, 0x1c, 0x00, 0x42,
				0x32, 0x01, 0x00, 0x00, 0x00},
			&GetDeviceIDRsp{
				BaseLayer: layers.BaseLayer{
					Contents: []byte{0x20, 0x81, 0x03, 0x45, 0x02, 0xbf, 0x4c,
						0x1c, 0x00, 0x42, 0x32, 0x01, 0x00, 0x00, 0x00},
				},
				ID:                               32,
				ProvidesSDRs:                     true,
				Revision:                         1,
				Available:                        true,
				MajorFirmwareRevision:            3,
				MinorFirmwareRevision:            45,
				MajorIPMIVersion:                 2,
				MinorIPMIVersion:                 0,
				SupportsChassisDevice:            true,
				SupportsBridgeDevice:             false,
				SupportsIPMBEventGeneratorDevice: true,
				SupportsIPMBEventReceiverDevice:  true,
				SupportsFRUInventoryDevice:       true,
				SupportsSELDevice:                true,
				SupportsSDRRepositoryDevice:      true,
				SupportsSensorDevice:             true,
				Manufacturer:                     7244,
				Product:                          12866,
				AuxiliaryFirmwareRevision:        [...]byte{0x01, 0x00, 0x00, 0x00},
			},
		},
		{
			[]byte{0x20, 0x7f, 0xff, 0x41, 0x51, 0xaa, 0xa2, 0x02, 0x00, 0x00,
				0x01, 0x00, 0x07, 0x28, 0x28},
			&GetDeviceIDRsp{
				BaseLayer: layers.BaseLayer{
					Contents: []byte{0x20, 0x7f, 0xff, 0x41, 0x51, 0xaa, 0xa2,
						0x02, 0x00, 0x00, 0x01, 0x00, 0x07, 0x28, 0x28},
				},
				ID:                               32,
				ProvidesSDRs:                     false,
				Revision:                         15,
				Available:                        false,
				MajorFirmwareRevision:            127,
				MinorFirmwareRevision:            41,
				MajorIPMIVersion:                 1,
				MinorIPMIVersion:                 5,
				SupportsChassisDevice:            true,
				SupportsBridgeDevice:             false,
				SupportsIPMBEventGeneratorDevice: true,
				SupportsIPMBEventReceiverDevice:  false,
				SupportsFRUInventoryDevice:       true,
				SupportsSELDevice:                false,
				SupportsSDRRepositoryDevice:      true,
				SupportsSensorDevice:             false,
				Manufacturer:                     674,
				Product:                          256,
				AuxiliaryFirmwareRevision:        [...]byte{0x00, 0x07, 0x28, 0x28},
			},
		},
		{
			[]byte{0x20, 0x01, 0x03, 0x72, 0x02, 0xbf, 0x7c, 0x2a, 0x00, 0x04,
				0x08, 0x00, 0x00, 0x00, 0x00},
			&GetDeviceIDRsp{
				BaseLayer: layers.BaseLayer{
					Contents: []byte{0x20, 0x01, 0x03, 0x72, 0x02, 0xbf, 0x7c,
						0x2a, 0x00, 0x04, 0x08, 0x00, 0x00, 0x00, 0x00},
				},
				ID:                               32,
				ProvidesSDRs:                     false,
				Revision:                         1,
				Available:                        true,
				MajorFirmwareRevision:            3,
				MinorFirmwareRevision:            72,
				MajorIPMIVersion:                 2,
				MinorIPMIVersion:                 0,
				SupportsChassisDevice:            true,
				SupportsBridgeDevice:             false,
				SupportsIPMBEventGeneratorDevice: true,
				SupportsIPMBEventReceiverDevice:  true,
				SupportsFRUInventoryDevice:       true,
				SupportsSELDevice:                true,
				SupportsSDRRepositoryDevice:      true,
				SupportsSensorDevice:             true,
				Manufacturer:                     10876,
				Product:                          2052,
				AuxiliaryFirmwareRevision:        [4]byte{},
			},
		},
	}
	for _, test := range tests {
		rsp := &GetDeviceIDRsp{}
		err := rsp.DecodeFromBytes(test.in, gopacket.NilDecodeFeedback)
		switch {
		case err == nil && test.want == nil:
			t.Errorf("expected error decoding %v, got none", test.in)
		case err == nil && test.want != nil:
			if diff := cmp.Diff(test.want, rsp); diff != "" {
				t.Errorf("decode %v = %v, want %v: %v", test.in, rsp, test.want, diff)
			}
		case err != nil && test.want != nil:
			t.Errorf("unexpected error: %v", err)
		}
	}
}
