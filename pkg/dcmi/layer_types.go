package dcmi

import (
	"github.com/gebn/bmc/pkg/layerexts"

	"github.com/google/gopacket"
)

var (
	layerTypeGetDCMICapabilitiesInfoReq = gopacket.RegisterLayerType(
		2000,
		gopacket.LayerTypeMetadata{
			Name: "Get DCMI Capabilities Info Request",
		},
	)
	layerTypeGetDCMICapabilitiesInfoSupportedCapabilitiesRsp = gopacket.RegisterLayerType(
		2001,
		gopacket.LayerTypeMetadata{
			Name: "Get DCMI Capabilities Info (Supported Capabilities) Response",
			Decoder: layerexts.BuildDecoder(func() layerexts.LayerDecodingLayer {
				return &GetDCMICapabilitiesInfoSupportedCapabilitiesRsp{}
			}),
		},
	)
	layerTypeGetDCMICapabilitiesInfoMandatoryPlatformAttrsRsp = gopacket.RegisterLayerType(
		2002,
		gopacket.LayerTypeMetadata{
			Name: "Get DCMI Capabilities Info (Mandatory Platform Attributes) Response",
			Decoder: layerexts.BuildDecoder(func() layerexts.LayerDecodingLayer {
				return &GetDCMICapabilitiesInfoMandatoryPlatformAttrsRsp{}
			}),
		},
	)
	layerTypeGetDCMICapabilitiesInfoOptionalPlatformAttrsRsp = gopacket.RegisterLayerType(
		2003,
		gopacket.LayerTypeMetadata{
			Name: "Get DCMI Capabilities Info (Optional Platform Attributes) Response",
			Decoder: layerexts.BuildDecoder(func() layerexts.LayerDecodingLayer {
				return &GetDCMICapabilitiesInfoOptionalPlatformAttrsRsp{}
			}),
		},
	)
	layerTypeGetDCMICapabilitiesInfoManageabilityAccessAttrsRsp = gopacket.RegisterLayerType(
		2004,
		gopacket.LayerTypeMetadata{
			Name: "Get DCMI Capabilities Info (Manageability Access Attributes) Response",
			Decoder: layerexts.BuildDecoder(func() layerexts.LayerDecodingLayer {
				return &GetDCMICapabilitiesInfoManageabilityAccessAttrsRsp{}
			}),
		},
	)
	layerTypeGetDCMICapabilitiesInfoEnhancedSystemPowerStatisticsAttrsRsp = gopacket.RegisterLayerType(
		2005,
		gopacket.LayerTypeMetadata{
			Name: "Get DCMI Capabilities Info (Enhanced System Power Statistics Attributes) Response",
			Decoder: layerexts.BuildDecoder(func() layerexts.LayerDecodingLayer {
				return &GetDCMICapabilitiesInfoEnhancedSystemPowerStatisticsAttrsRsp{}
			}),
		},
	)
	layerTypeGetPowerReadingReq = gopacket.RegisterLayerType(
		2006,
		gopacket.LayerTypeMetadata{
			Name: "Get Power Reading Request",
		},
	)
	layerTypeGetPowerReadingRsp = gopacket.RegisterLayerType(
		2007,
		gopacket.LayerTypeMetadata{
			Name: "Get Power Reading Response",
			Decoder: layerexts.BuildDecoder(func() layerexts.LayerDecodingLayer {
				return &GetPowerReadingRsp{}
			}),
		},
	)
)
