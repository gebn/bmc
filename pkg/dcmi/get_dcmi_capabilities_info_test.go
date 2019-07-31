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

var (
	header = getDCMICapabilitiesInfoRspHeader{
		MajorVersion: 1,
		MinorVersion: 5,
		Revision:     2,
	}
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
				Parameter: 22,
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

func TestGetDCMICapabilitiesInfoRspHeaderDecode(t *testing.T) {
	tests := []struct {
		in        []byte
		want      *getDCMICapabilitiesInfoRspHeader // nil if error
		remaining []byte
	}{
		{
			make([]byte, 2), // too short
			nil,
			nil,
		},
		{
			[]byte{0x01, 0x01, 0x02},
			&getDCMICapabilitiesInfoRspHeader{
				MajorVersion: 1,
				MinorVersion: 1,
				Revision:     2,
			},
			[]byte{},
		},
		{
			[]byte{0x01, 0x05, 0x02},
			&header,
			[]byte{},
		},
		{
			[]byte{0x0f, 0xf0, 0x09, 0x01, 0x02, 0x03},
			&getDCMICapabilitiesInfoRspHeader{
				MajorVersion: 15,
				MinorVersion: 240,
				Revision:     9,
			},
			[]byte{0x01, 0x02, 0x03},
		},
	}
	header := &getDCMICapabilitiesInfoRspHeader{}
	for _, test := range tests {
		remaining, err := header.Decode(test.in, gopacket.NilDecodeFeedback)
		switch {
		case err == nil && test.want == nil:
			t.Errorf("decode %v succeeded with %v, wanted error", test.in,
				header)
		case err != nil && test.want != nil:
			t.Errorf("decode %v failed with %v, wanted %v", test.in, err,
				test.want)
		case !bytes.Equal(remaining, test.remaining):
			t.Errorf("decode %v failed: remaining %v, wanted %v", test.in,
				remaining, test.remaining)
		case err == nil && test.want != nil:
			if diff := cmp.Diff(test.want, header); diff != "" {
				t.Errorf("decode %v = %v, want %v: %v", test.in, header, test.want, diff)
			}
		}
	}
}

func TestGetDCMICapabilitiesInfoSupportedCapabilitiesRspDecodeToString(t *testing.T) {
	tests := []struct {
		in   []byte
		want *GetDCMICapabilitiesInfoSupportedCapabilitiesRsp
	}{
		{
			make([]byte, 5),
			nil,
		},
		{
			[]byte{
				0x01, 0x05, 0x02,
				0x00, 0x00, 0xf5,
				0x01, 0x02,
			},
			&GetDCMICapabilitiesInfoSupportedCapabilitiesRsp{
				BaseLayer: layers.BaseLayer{
					Contents: []byte{
						0x01, 0x05, 0x02,
						0x00, 0x00, 0xf5,
					},
					Payload: []byte{0x01, 0x02},
				},
				getDCMICapabilitiesInfoRspHeader:  header,
				PowerManagement:                   false,
				OOBSecondaryLANChannelAvailable:   true,
				SerialTMODEAvailable:              false,
				IBSystemInterfaceChannelAvailable: true,
			},
		},
		{
			[]byte{
				0x01, 0x05, 0x02,
				0xff, 0x01, 0xf2,
			},
			&GetDCMICapabilitiesInfoSupportedCapabilitiesRsp{
				BaseLayer: layers.BaseLayer{
					Contents: []byte{
						0x01, 0x05, 0x02,
						0xff, 0x01, 0xf2,
					},
					Payload: []byte{},
				},
				getDCMICapabilitiesInfoRspHeader:  header,
				PowerManagement:                   true,
				OOBSecondaryLANChannelAvailable:   false,
				SerialTMODEAvailable:              true,
				IBSystemInterfaceChannelAvailable: false,
			},
		},
	}
	layer := &GetDCMICapabilitiesInfoSupportedCapabilitiesRsp{}
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
			if diff := cmp.Diff(test.want, layer, cmp.AllowUnexported(*layer)); diff != "" {
				t.Errorf("decode %v = %v, want %v: %v", test.in, layer, test.want, diff)
			}
		}
	}
}

func TestGetDCMICapabilitiesInfoMandatoryPlatformAttrsRspDecodeToString(t *testing.T) {
	tests := []struct {
		in   []byte
		want *GetDCMICapabilitiesInfoMandatoryPlatformAttrsRsp
	}{
		{
			make([]byte, 7),
			nil,
		},
		{
			[]byte{
				0x01, 0x05, 0x02,
				0xa5, 0x0a, 0x00, 0x00, 0x0f,
				0x03, 0x04,
			},
			&GetDCMICapabilitiesInfoMandatoryPlatformAttrsRsp{
				BaseLayer: layers.BaseLayer{
					Contents: []byte{
						0x01, 0x05, 0x02,
						0xa5, 0x0a, 0x00, 0x00, 0x0f,
					},
					Payload: []byte{0x03, 0x04},
				},
				getDCMICapabilitiesInfoRspHeader: header,
				SELAutoRollover:                  true,
				SELFlushOnRollover:               false,
				SELRecordLevelFlushOnRollover:    true,
				SELMaxEntries:                    2565,
				TemperatureSamplingFrequency:     time.Second * 15,
			},
		},
		{
			[]byte{
				0x01, 0x05, 0x02,
				0x4f, 0xff, 0xff, 0xff, 0xf0,
			},
			&GetDCMICapabilitiesInfoMandatoryPlatformAttrsRsp{
				BaseLayer: layers.BaseLayer{
					Contents: []byte{
						0x01, 0x05, 0x02,
						0x4f, 0xff, 0xff, 0xff, 0xf0,
					},
					Payload: []byte{},
				},
				getDCMICapabilitiesInfoRspHeader: header,
				SELAutoRollover:                  false,
				SELFlushOnRollover:               true,
				SELRecordLevelFlushOnRollover:    false,
				SELMaxEntries:                    65295,
				TemperatureSamplingFrequency:     time.Second * 240,
			},
		},
	}
	layer := &GetDCMICapabilitiesInfoMandatoryPlatformAttrsRsp{}
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
			if diff := cmp.Diff(test.want, layer, cmp.AllowUnexported(*layer)); diff != "" {
				t.Errorf("decode %v = %v, want %v: %v", test.in, layer, test.want, diff)
			}
		}
	}
}

func TestGetDCMICapabilitiesInfoOptionalPlatformAttrsRspDecodeToString(t *testing.T) {
	tests := []struct {
		in   []byte
		want *GetDCMICapabilitiesInfoOptionalPlatformAttrsRsp
	}{
		{
			make([]byte, 4),
			nil,
		},
		{
			[]byte{
				0x01, 0x05, 0x02,
				0x20, 0xf0,
				0x05, 0x06,
			},
			&GetDCMICapabilitiesInfoOptionalPlatformAttrsRsp{
				BaseLayer: layers.BaseLayer{
					Contents: []byte{
						0x01, 0x05, 0x02,
						0x20, 0xf0,
					},
					Payload: []byte{0x05, 0x06},
				},
				getDCMICapabilitiesInfoRspHeader: header,
				PowerManagementSlaveAddress:      ipmi.SlaveAddressBMC,
				PowerManagementChannel:           ipmi.ChannelSystemInterface,
				PowerManagementRevision:          0,
			},
		},
		{
			[]byte{
				0x01, 0x05, 0x02,
				0xff, 0x0f,
			},
			&GetDCMICapabilitiesInfoOptionalPlatformAttrsRsp{
				BaseLayer: layers.BaseLayer{
					Contents: []byte{
						0x01, 0x05, 0x02,
						0xff, 0x0f,
					},
					Payload: []byte{},
				},
				getDCMICapabilitiesInfoRspHeader: header,
				PowerManagementSlaveAddress:      ipmi.SlaveAddress(0x7f),
				PowerManagementChannel:           ipmi.ChannelPrimaryIPMB,
				PowerManagementRevision:          15,
			},
		},
	}
	layer := &GetDCMICapabilitiesInfoOptionalPlatformAttrsRsp{}
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
			if diff := cmp.Diff(test.want, layer, cmp.AllowUnexported(*layer)); diff != "" {
				t.Errorf("decode %v = %v, want %v: %v", test.in, layer, test.want, diff)
			}
		}
	}
}

func TestGetDCMICapabilitiesInfoManageabilityAccessAttrsRspDecodeToString(t *testing.T) {
	tests := []struct {
		in   []byte
		want *GetDCMICapabilitiesInfoManageabilityAccessAttrsRsp
	}{
		{
			make([]byte, 5),
			nil,
		},
		{
			[]byte{
				0x01, 0x05, 0x02,
				0x01, 0xff, 0x03,
				0x07, 0x08,
			},
			&GetDCMICapabilitiesInfoManageabilityAccessAttrsRsp{
				BaseLayer: layers.BaseLayer{
					Contents: []byte{
						0x01, 0x05, 0x02,
						0x01, 0xff, 0x03,
					},
					Payload: []byte{0x07, 0x08},
				},
				getDCMICapabilitiesInfoRspHeader: header,
				PrimaryLANOOBChannel:             ipmi.Channel(1),
				SecondaryLANOOBChannel:           ipmi.Channel(0xff),
				SerialOOBChannel:                 ipmi.Channel(3),
			},
		},
		{
			[]byte{
				0x01, 0x05, 0x02,
				0xff, 0x02, 0xff,
			},
			&GetDCMICapabilitiesInfoManageabilityAccessAttrsRsp{
				BaseLayer: layers.BaseLayer{
					Contents: []byte{
						0x01, 0x05, 0x02,
						0xff, 0x02, 0xff,
					},
					Payload: []byte{},
				},
				getDCMICapabilitiesInfoRspHeader: header,
				PrimaryLANOOBChannel:             ipmi.Channel(0xff),
				SecondaryLANOOBChannel:           ipmi.Channel(2),
				SerialOOBChannel:                 ipmi.Channel(0xff),
			},
		},
	}
	layer := &GetDCMICapabilitiesInfoManageabilityAccessAttrsRsp{}
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
			if diff := cmp.Diff(test.want, layer, cmp.AllowUnexported(*layer)); diff != "" {
				t.Errorf("decode %v = %v, want %v: %v", test.in, layer, test.want, diff)
			}
		}
	}
}

func TestGetDCMICapabilitiesInfoEnhancedSystemPowerStatisticsAttrsRspDecodeToString(t *testing.T) {
	tests := []struct {
		in   []byte
		want *GetDCMICapabilitiesInfoEnhancedSystemPowerStatisticsAttrsRsp
	}{
		{
			make([]byte, 3),
			nil,
		},
		// wasted byte
		{
			[]byte{
				0x01, 0x05, 0x02,
				0x00,
				0x09, 0x0a,
			},
			&GetDCMICapabilitiesInfoEnhancedSystemPowerStatisticsAttrsRsp{
				BaseLayer: layers.BaseLayer{
					Contents: []byte{
						0x01, 0x05, 0x02,
						0x00,
					},
					Payload: []byte{0x09, 0x0a},
				},
				getDCMICapabilitiesInfoRspHeader: header,
				PowerRollingAvgTimePeriods:       nil,
			},
		},
		// too short for indicated number of avg time periods (indicated: 1,
		// actual: 0)
		{
			[]byte{
				0x01, 0x05, 0x02,
				0x01,
			},
			nil,
		},
		// too short for indicated number of avg time periods (indicated: 4,
		// actual: 3)
		{
			[]byte{
				0x01, 0x05, 0x02,
				0x04, 0x01, 0x02, 0x03,
			},
			nil,
		},
		// 1 supported rolling avg time periods
		{
			[]byte{
				0x01, 0x05, 0x02,
				0x01, 0x00,
			},
			&GetDCMICapabilitiesInfoEnhancedSystemPowerStatisticsAttrsRsp{
				BaseLayer: layers.BaseLayer{
					Contents: []byte{
						0x01, 0x05, 0x02,
						0x01, 0x00,
					},
					Payload: []byte{},
				},
				getDCMICapabilitiesInfoRspHeader: header,
				PowerRollingAvgTimePeriods:       []time.Duration{0},
			},
		},
		// 5 supported rolling avg time periods, with trailing bytes (longer
		// payload than required; looking for truncation to be correct)
		{
			[]byte{
				0x01, 0x05, 0x02,
				0x05, 0x2a, 0xd5, 0xb3, 0x4c, 0x27,
				0x00, 0x00,
			},
			&GetDCMICapabilitiesInfoEnhancedSystemPowerStatisticsAttrsRsp{
				BaseLayer: layers.BaseLayer{
					Contents: []byte{
						0x01, 0x05, 0x02,
						0x05, 0x2a, 0xd5, 0xb3, 0x4c, 0x27,
					},
					Payload: []byte{0x00, 0x00},
				},
				getDCMICapabilitiesInfoRspHeader: header,
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
	layer := &GetDCMICapabilitiesInfoEnhancedSystemPowerStatisticsAttrsRsp{}
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
			if diff := cmp.Diff(test.want, layer, cmp.AllowUnexported(*layer)); diff != "" {
				t.Errorf("decode %v = %v, want %v: %v", test.in, layer, test.want, diff)
			}
		}
	}
}
