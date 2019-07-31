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
			Name:    "Get DCMI Capabilities Info (Supported Capabilities) Response",
			Decoder: layerexts.BuildDecoder(&GetDCMICapabilitiesInfoSupportedCapabilitiesRsp{}),
		},
	)
	layerTypeGetDCMICapabilitiesInfoMandatoryPlatformAttrsRsp = gopacket.RegisterLayerType(
		2002,
		gopacket.LayerTypeMetadata{
			Name:    "Get DCMI Capabilities Info (Mandatory Platform Attributes) Response",
			Decoder: layerexts.BuildDecoder(&GetDCMICapabilitiesInfoMandatoryPlatformAttrsRsp{}),
		},
	)
	layerTypeGetDCMICapabilitiesInfoOptionalPlatformAttrsRsp = gopacket.RegisterLayerType(
		2003,
		gopacket.LayerTypeMetadata{
			Name:    "Get DCMI Capabilities Info (Optional Platform Attributes) Response",
			Decoder: layerexts.BuildDecoder(&GetDCMICapabilitiesInfoOptionalPlatformAttrsRsp{}),
		},
	)
	layerTypeGetDCMICapabilitiesInfoManageabilityAccessAttrsRsp = gopacket.RegisterLayerType(
		2004,
		gopacket.LayerTypeMetadata{
			Name:    "Get DCMI Capabilities Info (Manageability Access Attributes) Response",
			Decoder: layerexts.BuildDecoder(&GetDCMICapabilitiesInfoManageabilityAccessAttrsRsp{}),
		},
	)
	layerTypeGetDCMICapabilitiesInfoEnhancedSystemPowerStatisticsAttrsRsp = gopacket.RegisterLayerType(
		2005,
		gopacket.LayerTypeMetadata{
			Name:    "Get DCMI Capabilities Info (Enhanced System Power Statistics Attributes) Response",
			Decoder: layerexts.BuildDecoder(&GetDCMICapabilitiesInfoEnhancedSystemPowerStatisticsAttrsRsp{}),
		},
	)
)
