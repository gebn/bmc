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
	v10Header = getDCMICapabilitiesInfoRspHeader{
		MajorVersion: 1,
		MinorVersion: 0,
		Revision:     1,
	}
	v11Header = getDCMICapabilitiesInfoRspHeader{
		MajorVersion: 1,
		MinorVersion: 1,
		Revision:     2,
	}
	v15Header = getDCMICapabilitiesInfoRspHeader{
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
			[]byte{0x01, 0x00, 0x01},
			&v10Header,
			[]byte{},
		},
		{
			[]byte{0x01, 0x01, 0x02},
			&v11Header,
			[]byte{},
		},
		{
			[]byte{0x01, 0x05, 0x02},
			&v15Header,
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
				0x01, 0x00, 0x01,
				0x0a, 0x01, 0xd5,
			},
			&GetDCMICapabilitiesInfoSupportedCapabilitiesRsp{
				BaseLayer: layers.BaseLayer{
					Contents: []byte{
						0x01, 0x00, 0x01,
						0x0a, 0x01, 0xd5,
					},
					Payload: []byte{},
				},
				getDCMICapabilitiesInfoRspHeader:  v10Header,
				TemperatureMonitor:                true,
				ChassisPower:                      false,
				SELLogging:                        true,
				Identification:                    false,
				PowerManagement:                   true,
				VLANCapable:                       false,
				SOLSupported:                      true,
				OOBPrimaryLANChannelAvailable:     false,
				OOBSecondaryLANChannelAvailable:   true,
				SerialTMODEAvailable:              false,
				IBKCSChannelAvailable:             true,
				IBSystemInterfaceChannelAvailable: false,
			},
		},
		{
			[]byte{
				0x01, 0x00, 0x01,
				0x05, 0x00, 0x2a,
			},
			&GetDCMICapabilitiesInfoSupportedCapabilitiesRsp{
				BaseLayer: layers.BaseLayer{
					Contents: []byte{
						0x01, 0x00, 0x01,
						0x05, 0x00, 0x2a,
					},
					Payload: []byte{},
				},
				getDCMICapabilitiesInfoRspHeader:  v10Header,
				TemperatureMonitor:                false,
				ChassisPower:                      true,
				SELLogging:                        false,
				Identification:                    true,
				PowerManagement:                   false,
				VLANCapable:                       true,
				SOLSupported:                      false,
				OOBPrimaryLANChannelAvailable:     true,
				OOBSecondaryLANChannelAvailable:   false,
				SerialTMODEAvailable:              true,
				IBKCSChannelAvailable:             false,
				IBSystemInterfaceChannelAvailable: false,
			},
		},
		{
			[]byte{
				0x01, 0x01, 0x02,
				0x00, 0x00, 0x02,
			},
			&GetDCMICapabilitiesInfoSupportedCapabilitiesRsp{
				BaseLayer: layers.BaseLayer{
					Contents: []byte{
						0x01, 0x01, 0x02,
						0x00, 0x00, 0x02,
					},
					Payload: []byte{},
				},
				getDCMICapabilitiesInfoRspHeader:  v11Header,
				TemperatureMonitor:                true,
				ChassisPower:                      true,
				SELLogging:                        true,
				Identification:                    true,
				PowerManagement:                   false,
				VLANCapable:                       true,
				SOLSupported:                      true,
				OOBPrimaryLANChannelAvailable:     true,
				OOBSecondaryLANChannelAvailable:   false,
				SerialTMODEAvailable:              true,
				IBKCSChannelAvailable:             true,
				IBSystemInterfaceChannelAvailable: false,
			},
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
				getDCMICapabilitiesInfoRspHeader:  v15Header,
				TemperatureMonitor:                true,
				ChassisPower:                      true,
				SELLogging:                        true,
				Identification:                    true,
				PowerManagement:                   false,
				VLANCapable:                       true,
				SOLSupported:                      true,
				OOBPrimaryLANChannelAvailable:     true,
				OOBSecondaryLANChannelAvailable:   true,
				SerialTMODEAvailable:              false,
				IBKCSChannelAvailable:             true,
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
				getDCMICapabilitiesInfoRspHeader:  v15Header,
				TemperatureMonitor:                true,
				ChassisPower:                      true,
				SELLogging:                        true,
				Identification:                    true,
				PowerManagement:                   true,
				VLANCapable:                       true,
				SOLSupported:                      true,
				OOBPrimaryLANChannelAvailable:     true,
				OOBSecondaryLANChannelAvailable:   false,
				SerialTMODEAvailable:              true,
				IBKCSChannelAvailable:             true,
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
		// too short for v1.0
		{
			[]byte{
				0x01, 0x00, 0x01,
				0x00, 0x00, 0x00,
			},
			nil,
		},
		// too short for v1.1; this is 3 rather than 4 bytes as v1.1 responses
		// have been known to actually be v1.0 responses (SuperMicro) so we
		// interpret them as such
		{
			[]byte{
				0x01, 0x01, 0x01,
				0x00, 0x00, 0x00,
			},
			nil,
		},
		// too short for v1.5; this is 3 rather than 4 bytes as v1.1 responses
		// have been known to actually be v1.0 responses (SuperMicro) so we
		// interpret v1.5 assuming the same behaviour
		{
			[]byte{
				0x01, 0x05, 0x01,
				0x00, 0x00, 0x00,
			},
			nil,
		},
		{
			[]byte{
				0x01, 0x00, 0x01,
				0xf5, 0xaa, 0x05, 0x02,
			},
			&GetDCMICapabilitiesInfoMandatoryPlatformAttrsRsp{
				BaseLayer: layers.BaseLayer{
					Contents: []byte{
						0x01, 0x00, 0x01,
						0xf5, 0xaa, 0x05, 0x02,
					},
					Payload: []byte{},
				},
				getDCMICapabilitiesInfoRspHeader: v10Header,
				SELAutoRollover:                  true,
				SELFlushOnRollover:               false,
				SELRecordLevelFlushOnRollover:    false,
				SELMaxEntries:                    43525,
				AssetTagSupport:                  true,
				DHCPHostNameSupport:              false,
				GUIDSupport:                      true,
				BaseboardTemperature:             false,
				ProcessorsTemperature:            true,
				InletTemperature:                 false,
				TemperatureSamplingFrequency:     time.Duration(0),
			},
		},
		{
			[]byte{
				0x01, 0x00, 0x01,
				0x7a, 0x5a, 0x02, 0x05,
			},
			&GetDCMICapabilitiesInfoMandatoryPlatformAttrsRsp{
				BaseLayer: layers.BaseLayer{
					Contents: []byte{
						0x01, 0x00, 0x01,
						0x7a, 0x5a, 0x02, 0x05,
					},
					Payload: []byte{},
				},
				getDCMICapabilitiesInfoRspHeader: v10Header,
				SELAutoRollover:                  false,
				SELFlushOnRollover:               false,
				SELRecordLevelFlushOnRollover:    false,
				SELMaxEntries:                    23050,
				AssetTagSupport:                  false,
				DHCPHostNameSupport:              true,
				GUIDSupport:                      false,
				BaseboardTemperature:             true,
				ProcessorsTemperature:            false,
				InletTemperature:                 true,
				TemperatureSamplingFrequency:     time.Duration(0),
			},
		},
		// v1.0 response purporting to be v1.1
		{
			[]byte{
				0x01, 0x01, 0x02,
				0x7a, 0x5a, 0x02, 0x05,
				// we cannot handle trailing bytes, as the length is how we
				// identify v1.0; we could possibly look at whether bytes 3 and
				// 4 are null, however that's a level of heuristic too far
			},
			&GetDCMICapabilitiesInfoMandatoryPlatformAttrsRsp{
				BaseLayer: layers.BaseLayer{
					Contents: []byte{
						0x01, 0x01, 0x02,
						0x7a, 0x5a, 0x02, 0x05,
					},
					Payload: []byte{},
				},
				getDCMICapabilitiesInfoRspHeader: v11Header,
				SELAutoRollover:                  false,
				SELFlushOnRollover:               false,
				SELRecordLevelFlushOnRollover:    false,
				SELMaxEntries:                    23050,
				AssetTagSupport:                  false,
				DHCPHostNameSupport:              true,
				GUIDSupport:                      false,
				BaseboardTemperature:             true,
				ProcessorsTemperature:            false,
				InletTemperature:                 true,
				TemperatureSamplingFrequency:     time.Duration(0),
			},
		},
		{
			[]byte{
				0x01, 0x01, 0x02,
				0xa5, 0x0a, 0x00, 0x00, 0x0f,
				0x03, 0x04,
			},
			&GetDCMICapabilitiesInfoMandatoryPlatformAttrsRsp{
				BaseLayer: layers.BaseLayer{
					Contents: []byte{
						0x01, 0x01, 0x02,
						0xa5, 0x0a, 0x00, 0x00, 0x0f,
					},
					Payload: []byte{0x03, 0x04},
				},
				getDCMICapabilitiesInfoRspHeader: v11Header,
				SELAutoRollover:                  true,
				SELFlushOnRollover:               false,
				SELRecordLevelFlushOnRollover:    true,
				SELMaxEntries:                    2565,
				AssetTagSupport:                  true,
				DHCPHostNameSupport:              true,
				GUIDSupport:                      true,
				BaseboardTemperature:             true,
				ProcessorsTemperature:            true,
				InletTemperature:                 true,
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
				getDCMICapabilitiesInfoRspHeader: v15Header,
				SELAutoRollover:                  false,
				SELFlushOnRollover:               true,
				SELRecordLevelFlushOnRollover:    false,
				SELMaxEntries:                    65295,
				AssetTagSupport:                  true,
				DHCPHostNameSupport:              true,
				GUIDSupport:                      true,
				BaseboardTemperature:             true,
				ProcessorsTemperature:            true,
				InletTemperature:                 true,
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
				getDCMICapabilitiesInfoRspHeader: v15Header,
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
				getDCMICapabilitiesInfoRspHeader: v15Header,
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
				getDCMICapabilitiesInfoRspHeader: v15Header,
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
				getDCMICapabilitiesInfoRspHeader: v15Header,
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
				getDCMICapabilitiesInfoRspHeader: v15Header,
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
				getDCMICapabilitiesInfoRspHeader: v15Header,
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
				getDCMICapabilitiesInfoRspHeader: v15Header,
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
