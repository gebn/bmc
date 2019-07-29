package bmc

import (
	"github.com/gebn/bmc/pkg/ipmi"

	"github.com/google/gopacket/layers"
)

// v1SessionLayers represents layers common to all V1Session commands. They are
// allocated in a block here and reused for efficiency.
type v1SessionLayers struct {
	rmcpLayer            layers.RMCP
	sessionSelectorLayer ipmi.SessionSelector
	v1sessionLayer       ipmi.V1Session
	messageLayer         ipmi.Message
}
