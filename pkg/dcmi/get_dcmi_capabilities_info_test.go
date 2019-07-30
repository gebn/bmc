package dcmi

import (
	"bytes"
	"testing"
	"time"

	"github.com/gebn/bmc/pkg/ipmi"

	"github.com/google/go-cmp/cmp"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
)

func TestGetDCMICapabilitiesInfoReqSerializeTo(t *testing.T) {
	tests := []struct {
		in   *GetDCMICapabilitiesInfoReq
		want []byte
	}{
		{
			&GetDCMICapabilitiesInfoReq{},
			[]byte{0x00},
		},
		{
			&GetDCMICapabilitiesInfoReq{
				ParameterSelector: 22,
			},
			[]byte{0x16},
		},
	}
	opts := gopacket.SerializeOptions{}
	for _, test := range tests {
		sb := gopacket.NewSerializeBuffer()
		if err := test.in.SerializeTo(sb, opts); err != nil {
			t.Errorf("serialize %v = error %v, want %v", test.in, err, test.want)
			continue
		}
		got := sb.Bytes()
		if !bytes.Equal(got, test.want) {
			t.Errorf("serialize %v = %v, want %v", test.in, got, test.want)
		}
	}
}

func TestGetDCMICapabilitiesInfoRspDecodeFromBytes(t *testing.T) {
	tests := []struct {
		in   []byte
		want *GetDCMICapabilitiesInfoRsp
	}{
		// too short
		{
			make([]byte, 15),
			nil,
		},

		// no rolling avg time periods (len == 16)
		{
			[]byte{
				0x01, 0x01, 0x02,
				0x00, 0x00, 0xf5,
				0xa5, 0x0a, 0x00, 0x00, 0x0f,
				0x20, 0xf0,
				0x01, 0xff, 0x03,
			},
			&GetDCMICapabilitiesInfoRsp{
				BaseLayer: layers.BaseLayer{
					Contents: []byte{
						0x01, 0x01, 0x02,
						0x00, 0x00, 0xf5,
						0xa5, 0x0a, 0x00, 0x00, 0x0f,
						0x20, 0xf0,
						0x01, 0xff, 0x03,
					},
				},
				MajorVersion:                      1,
				MinorVersion:                      1,
				Revision:                          2,
				PowerManagement:                   false,
				OOBSecondaryLANChannelAvailable:   true,
				SerialTMODEAvailable:              false,
				IBSystemInterfaceChannelAvailable: true,
				SELAutoRollover:                   true,
				SELFlushOnRollover:                false,
				SELRecordLevelFlushOnRollover:     true,
				SELMaxEntries:                     2565,
				TemperatureSamplingFrequency:      time.Second * 15,
				PowerManagementSlaveAddress:       ipmi.SlaveAddressBMC,
				PowerManagementChannel:            ipmi.ChannelSystemInterface,
				PowerManagementRevision:           0,
				PrimaryLANOOBChannel:              ipmi.Channel(1),
				SecondaryLANOOBChannel:            ipmi.Channel(0xff),
				SerialOOBChannel:                  ipmi.Channel(3),
				PowerRollingAvgTimePeriods:        nil,
			},
		},

		// 0 supported rolling avg time periods (len == 17, useless byte)
		{
			[]byte{
				0x01, 0x05, 0xff,
				0xff, 0x01, 0xf2,
				0x4f, 0xff, 0xff, 0xff, 0xf0,
				0xff, 0x0f,
				0xff, 0x02, 0xff,
				0x00,
			},
			&GetDCMICapabilitiesInfoRsp{
				BaseLayer: layers.BaseLayer{
					Contents: []byte{
						0x01, 0x05, 0xff,
						0xff, 0x01, 0xf2,
						0x4f, 0xff, 0xff, 0xff, 0xf0,
						0xff, 0x0f,
						0xff, 0x02, 0xff,
						0x00,
					},
					Payload: []byte{},
				},
				MajorVersion:                      1,
				MinorVersion:                      5,
				Revision:                          255,
				PowerManagement:                   true,
				OOBSecondaryLANChannelAvailable:   false,
				SerialTMODEAvailable:              true,
				IBSystemInterfaceChannelAvailable: false,
				SELAutoRollover:                   false,
				SELFlushOnRollover:                true,
				SELRecordLevelFlushOnRollover:     false,
				SELMaxEntries:                     65295,
				TemperatureSamplingFrequency:      time.Second * 240,
				PowerManagementSlaveAddress:       ipmi.SlaveAddress(0x7f),
				PowerManagementChannel:            ipmi.ChannelPrimaryIPMB,
				PowerManagementRevision:           15,
				PrimaryLANOOBChannel:              ipmi.Channel(0xff),
				SecondaryLANOOBChannel:            ipmi.Channel(2),
				SerialOOBChannel:                  ipmi.Channel(0xff),
				PowerRollingAvgTimePeriods:        []time.Duration{},
			},
		},

		// too short for indicated number of avg time periods (indicated: 1,
		// actual: 0)
		{
			[]byte{
				0x01, 0x05, 0xff,
				0xff, 0x01, 0xf2,
				0x4f, 0xff, 0xff, 0xff, 0xf0,
				0xff, 0x0f,
				0xff, 0x02, 0xff,
				0x01,
			},
			nil,
		},

		// too short for indicated number of avg time periods (indicated: 4,
		// actual: 3)
		{
			[]byte{
				0x01, 0x05, 0xff,
				0xff, 0x01, 0xf2,
				0x4f, 0xff, 0xff, 0xff, 0xf0,
				0xff, 0x0f,
				0xff, 0x02, 0xff,
				0x04, 0x01, 0x02, 0x03,
			},
			nil,
		},

		// 1 supported rolling avg time periods (len == 18)
		{
			[]byte{
				0x01, 0x05, 0xff,
				0xff, 0x01, 0xf2,
				0x4f, 0xff, 0xff, 0xff, 0xf0,
				0xff, 0x0f,
				0xff, 0x02, 0xff,
				0x01, 0x00,
			},
			&GetDCMICapabilitiesInfoRsp{
				BaseLayer: layers.BaseLayer{
					Contents: []byte{
						0x01, 0x05, 0xff,
						0xff, 0x01, 0xf2,
						0x4f, 0xff, 0xff, 0xff, 0xf0,
						0xff, 0x0f,
						0xff, 0x02, 0xff,
						0x01, 0x00,
					},
					Payload: []byte{},
				},
				MajorVersion:                      1,
				MinorVersion:                      5,
				Revision:                          255,
				PowerManagement:                   true,
				OOBSecondaryLANChannelAvailable:   false,
				SerialTMODEAvailable:              true,
				IBSystemInterfaceChannelAvailable: false,
				SELAutoRollover:                   false,
				SELFlushOnRollover:                true,
				SELRecordLevelFlushOnRollover:     false,
				SELMaxEntries:                     65295,
				TemperatureSamplingFrequency:      time.Second * 240,
				PowerManagementSlaveAddress:       ipmi.SlaveAddress(0x7f),
				PowerManagementChannel:            ipmi.ChannelPrimaryIPMB,
				PowerManagementRevision:           15,
				PrimaryLANOOBChannel:              ipmi.Channel(0xff),
				SecondaryLANOOBChannel:            ipmi.Channel(2),
				SerialOOBChannel:                  ipmi.Channel(0xff),
				PowerRollingAvgTimePeriods:        []time.Duration{0},
			},
		},

		// 5 supported rolling avg time periods (len == 22), with trailing bytes
		// (longer payload than required; looking for truncation to be correct)
		{
			[]byte{
				0x01, 0x01, 0x02,
				0x00, 0x00, 0xf5,
				0xa5, 0x0a, 0x00, 0x00, 0x0f,
				0x20, 0xf0,
				0x01, 0xff, 0x03,
				0x05, 0x2a, 0xd5, 0xb3, 0x4c, 0x27,
				0x00, 0x00, 0x00,
			},
			&GetDCMICapabilitiesInfoRsp{
				BaseLayer: layers.BaseLayer{
					Contents: []byte{
						0x01, 0x01, 0x02,
						0x00, 0x00, 0xf5,
						0xa5, 0x0a, 0x00, 0x00, 0x0f,
						0x20, 0xf0,
						0x01, 0xff, 0x03,
						0x05, 0x2a, 0xd5, 0xb3, 0x4c, 0x27,
					},
					Payload: []byte{0x00, 0x00, 0x00},
				},
				MajorVersion:                      1,
				MinorVersion:                      1,
				Revision:                          2,
				PowerManagement:                   false,
				OOBSecondaryLANChannelAvailable:   true,
				SerialTMODEAvailable:              false,
				IBSystemInterfaceChannelAvailable: true,
				SELAutoRollover:                   true,
				SELFlushOnRollover:                false,
				SELRecordLevelFlushOnRollover:     true,
				SELMaxEntries:                     2565,
				TemperatureSamplingFrequency:      time.Second * 15,
				PowerManagementSlaveAddress:       ipmi.SlaveAddressBMC,
				PowerManagementChannel:            ipmi.ChannelSystemInterface,
				PowerManagementRevision:           0,
				PrimaryLANOOBChannel:              ipmi.Channel(1),
				SecondaryLANOOBChannel:            ipmi.Channel(0xff),
				SerialOOBChannel:                  ipmi.Channel(3),
				PowerRollingAvgTimePeriods: []time.Duration{
					time.Second * 42,
					time.Hour * 24 * 21,
					time.Hour * 51,
					time.Minute * 12,
					time.Second * 39,
				},
			},
		},
	}
	layer := &GetDCMICapabilitiesInfoRsp{}
	for _, test := range tests {
		err := layer.DecodeFromBytes(test.in, gopacket.NilDecodeFeedback)
		switch {
		case err == nil && test.want == nil:
			t.Errorf("decode %v succeeded with %v, wanted error", test.in,
				layer)
		case err != nil && test.want != nil:
			t.Errorf("decode %v failed with %v, wanted %v", test.in, err,
				test.want)
		case err == nil && test.want != nil:
			if diff := cmp.Diff(test.want, layer); diff != "" {
				t.Errorf("decode %v = %v, want %v: %v", test.in, layer, test.want, diff)
			}
		}
	}
}
