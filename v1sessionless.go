package bmc

import (
	"context"
	"fmt"

	"github.com/gebn/bmc/internal/pkg/transport"
	"github.com/gebn/bmc/pkg/ipmi"
	"github.com/gebn/bmc/pkg/layerexts"

	"github.com/google/gopacket"
)

// V1Sessionless represents a session-less connection to a BMC using a "null"
// IPMI v1.5 session wrapper.
type V1Sessionless struct {
	transport transport.Transport
	session   ipmi.V1Session
}

func (s *V1Sessionless) Version() string {
	return "1.5"
}

func (s *V1Sessionless) SendMessage(ctx context.Context, op *ipmi.Operation, cmd gopacket.SerializableLayer) (layerexts.DecodedTypes, ipmi.CompletionCode, error) {
	return nil, ipmi.CompletionCodeUnspecified, fmt.Errorf("not implemented")
}

func (s *V1Sessionless) GetSystemGUID(ctx context.Context) ([16]byte, error) {
	return [16]byte{}, fmt.Errorf("not implemented")
}

func (s *V1Sessionless) GetChannelAuthenticationCapabilities(ctx context.Context, r *ipmi.GetChannelAuthenticationCapabilitiesReq) (*ipmi.GetChannelAuthenticationCapabilitiesRsp, error) {
	return nil, fmt.Errorf("not implemented")
}

// TODO may want to implement getSessionChallenge()
// no activate session, as this does not use a null session ID, so is not
// technically a sessionless command according to the spec; need to think about
// how this command will be sent in this case - it certainly isn't inside a
// session, so may want to bend the rules
