package bmc

import (
	"github.com/gebn/bmc/pkg/ipmi"

	"github.com/google/gopacket/layers"
)

type sessionlessRspLayers struct {
	rmcpLayer                                    layers.RMCP
	sessionSelectorLayer                         ipmi.SessionSelector
	messageLayer                                 ipmi.Message
	getSystemGUIDRspLayer                        ipmi.GetSystemGUIDRsp
	getChannelAuthenticationCapabilitiesRspLayer ipmi.GetChannelAuthenticationCapabilitiesRsp
}
