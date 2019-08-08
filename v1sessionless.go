package bmc

import (
	"context"
	"fmt"

	"github.com/gebn/bmc/internal/pkg/transport"
	"github.com/gebn/bmc/pkg/ipmi"
)

var (
	v1ConnectionOpenAttempts = connectionOpenAttempts.WithLabelValues("1.5")
	v1ConnectionOpenFailures = connectionOpenFailures.WithLabelValues("1.5")
	v1ConnectionsOpen        = connectionsOpen.WithLabelValues("1.5")
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

func (s *V1Sessionless) SendCommand(ctx context.Context, c ipmi.Command) (ipmi.CompletionCode, error) {
	return ipmi.CompletionCodeUnspecified, fmt.Errorf("not implemented")
}

// TODO may want to implement getSessionChallenge()
// no activate session, as this does not use a null session ID, so is not
// technically a sessionless command according to the spec; need to think about
// how this command will be sent in this case - it certainly isn't inside a
// session, so may want to bend the rules
