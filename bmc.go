// Package bmc implements an IPMI v1.5/2.0 remote console. pkg/ipmi provides the
// layers; this package makes IPMI work in Go.
package bmc

import (
	"errors"
	"fmt"
	"strings"

	"github.com/gebn/bmc/internal/pkg/transport"
	"github.com/gebn/bmc/pkg/ipmi"

	"github.com/google/gopacket"
)

var (
	serializeOptions = gopacket.SerializeOptions{
		FixLengths:       true,
		ComputeChecksums: true,
	}

	namespace = "bmc"

	// SuccessfulEmptyResponse is special case error used to indicate the
	// scenario where we execute a command with a response, so we expect >0
	// bytes after the message, and get a 0x00 (Normal) completion code, but the
	// response packet ends after the message layer. This is known to happen
	// when we send DCMI Get Power Reading to a Super Micro motherboard where
	// the PSU does not support PMBus, and sending a Super Micro motherboard Get
	// Channel Authentication Capabilities inside a session.
	//
	// This indicates a comformance problem with the BMC (it is indeed an
	// error), however sometimes it is still useful, e.g. Get Channel
	// Authentication Capabilities is used as a keepalive inside a session:
	// getting a 0x00 with empty response is the most efficient way of doing
	// this, so it could be deliberate; it's just unexpected and needs
	// special-casing.
	SuccessfulEmptyResponse = errors.New("expected a response payload, but " +
		"the packet ended after the message layer")
)

// TODO need to implement v1 sending
//// Dial queries the BMC at the supplied IP[:port] (IPv6 must be enclosed in
//// square brackets) for IPMI v2.0 capability. If it supports IPMI v2.0, a
//// V2SessionlessTransport will be returned, otherwise a V1SessionlessTransport
//// will be returned. If you know the BMC's capabilities, or need a specific
//// feature (e.g. DCMI), use the DialV*() functions instead, which expose
//// additional information and functionality.
//func Dial(ctx context.Context, addr string) (SessionlessTransport, error) {
//	t, err := newTransport(addr)
//	if err != nil {
//		return nil, err
//	}
//	v1 := newV1SessionlessTransport(t)
//	capabilities, err := v1.GetChannelAuthenticationCapabilities(
//		ctx,
//		&ipmi.GetChannelAuthenticationCapabilitiesReq{
//			ExtendedData:      true,
//			Channel:           ipmi.ChannelPresentInterface,
//			MaxPrivilegeLevel: ipmi.PrivilegeLevelAdministrator,
//		},
//	)
//	if err != nil {
//		v1.Close()
//		return nil, err
//	}
//	if capabilities.SupportsV2 {
//		// prefer IPMI v2.0 if supported; reuse socket
//		return newV2SessionlessTransport(t), nil
//	}
//	// assume capabilities.SupportsV1 == true by virtue of getting here
//	return v1, nil
//}

// DialV1 establishes a new IPMI v1.5 connection with the supplied BMC. The
// address follows the same format as for Dial(). Use this if you know the BMC
// does not support IPMI v2.0. In general, if a BMC supports v2.0, that should
// be used over v1.5.
func DialV1(addr string) (*V1SessionlessTransport, error) {
	v1ConnectionOpenAttempts.Inc()
	t, err := newTransport(addr)
	if err != nil {
		v1ConnectionOpenFailures.Inc()
		return nil, err
	}
	v1ConnectionsOpen.Inc()
	return newV1SessionlessTransport(t), nil
}

func newV1SessionlessTransport(t transport.Transport) *V1SessionlessTransport {
	return &V1SessionlessTransport{
		Transport: t,
		V1Sessionless: V1Sessionless{
			transport: t,
		},
	}
}

// DialV2 establishes a new IPMI v2.0 connection with the supplied BMC. The
// address follows the same format as for Dial(). Use this if you know the BMC
// supports IPMI v2.0 and/or require DCMI functionality.
func DialV2(addr string) (*V2SessionlessTransport, error) {
	v2ConnectionOpenAttempts.Inc()
	t, err := newTransport(addr)
	if err != nil {
		v2ConnectionOpenFailures.Inc()
		return nil, err
	}
	v2ConnectionsOpen.Inc()
	return newV2SessionlessTransport(t), nil
}

func newV2SessionlessTransport(t transport.Transport) *V2SessionlessTransport {
	return &V2SessionlessTransport{
		Transport:     t,
		V2Sessionless: newV2Sessionless(t),
	}
}

func newTransport(addr string) (transport.Transport, error) {
	// default to port 623
	if !strings.Contains(addr, ":") || strings.HasSuffix(addr, "]") {
		addr = addr + ":623"
	}
	return transport.New(addr)
}

// ValidateResponse is a helper to remove some boilerplate error handling from
// SendCommand() calls. It ensures a non-nil error and normal completion code.
// If the error is non-nil, it is returned. If the completion code is
// non-normal, an error is returned containing the actual value.
func ValidateResponse(c ipmi.CompletionCode, err error) error {
	if err != nil {
		return err
	}
	if c != ipmi.CompletionCodeNormal {
		return fmt.Errorf("received non-normal completion code: %v", c)
	}
	return nil
}
