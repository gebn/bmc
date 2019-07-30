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
	layerTypeGetDCMICapabilitiesInfoRsp = gopacket.RegisterLayerType(
		2001,
		gopacket.LayerTypeMetadata{
			Name:    "Get DCMI Capabilities Info Response",
			Decoder: layerexts.BuildDecoder(&GetDCMICapabilitiesInfoRsp{}),
		},
	)
)
