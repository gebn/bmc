package bmc

import (
	"context"
	"fmt"
	"time"

	"github.com/gebn/bmc/internal/pkg/transport"
	"github.com/gebn/bmc/pkg/ipmi"
	"github.com/gebn/bmc/pkg/layerexts"

	"github.com/cenkalti/backoff"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
)

var (
	// these not only save a map lookup each open, but also register the labels
	v2ConnectionOpenAttempts = connectionOpenAttempts.WithLabelValues("2.0")
	v2ConnectionOpenFailures = connectionOpenFailures.WithLabelValues("2.0")
	v2ConnectionsOpen        = connectionsOpen.WithLabelValues("2.0")
)

// v2ConnectionLayers contains layers common to all v2.0 connections. Although
// these layers are common, both V2Sessionless and V2Session embed this as a
// value, so each gets a fresh set of layers. This uses a little more memory,
// but it means when a session is closed, its session layer doesn't have a
// dangling confidentiality layer etc. This is why this is not embedded in
// v2ConnectionShared.
type v2ConnectionLayers struct {
	rmcpLayer            layers.RMCP
	sessionSelectorLayer ipmi.SessionSelector
	v2SessionLayer       ipmi.V2Session
	messageLayer         ipmi.Message
}

// v2ConnectionShared contains fields that a session-less connection passes to
// sessions created from it. V2Sessionless embeds a value of this type, and
// V2Session embeds a pointer which is set to the V2Sessionless's value.
//
// Note that a given BMC only supports a single command at a time, which is what
// makes this possible - if a session is sending a command, the session-less
// connection it was initiated from cannot send concurrently.
type v2ConnectionShared struct {

	// transport is the underlying UDP socket for the connection.
	transport transport.Transport

	// buffer is used to build all packets to send during this connection.
	// Reusing this between sends drastically reduces the number of allocations
	// we have to do when building packets, and reusing it between session-less
	// and session-based connections reduces it a little further.
	buffer gopacket.SerializeBuffer

	// layers contains layer types decoded by the connection's
	// gopacket.DecodingLayerParser. Although this slice is shared, each
	// connection has its own DLP, as each session may have a different
	// confidentiality layer.
	layers []gopacket.LayerType

	// backoff saves allocating a backoff each request. We must call .Reset() to
	// reset this between requests.
	backoff backoff.BackOff
}

// V2Sessionless represents a session-less connection to a BMC using a "null"
// IPMI v2.0 session wrapper.
type V2Sessionless struct {
	v2ConnectionLayers
	v2ConnectionShared

	// decode parses the layers in v2ConnectionShared.
	decode gopacket.DecodingLayerFunc
}

func newV2Sessionless(t transport.Transport) *V2Sessionless {
	s := &V2Sessionless{
		v2ConnectionShared: v2ConnectionShared{
			transport: t,
			buffer:    gopacket.NewSerializeBuffer(),
			backoff:   backoff.NewExponentialBackOff(),
		},
	}
	dlc := gopacket.DecodingLayerContainer(gopacket.DecodingLayerArray(nil))
	dlc = dlc.Put(&s.rmcpLayer)
	dlc = dlc.Put(&s.sessionSelectorLayer)
	dlc = dlc.Put(&s.v2SessionLayer)
	dlc = dlc.Put(&s.messageLayer)
	s.decode = dlc.LayersDecoder(s.rmcpLayer.LayerType(), gopacket.NilDecodeFeedback)
	return s
}

func (s *V2Sessionless) Close() error {
	// we intercept this call purely to do the gauge bookkeeping
	if err := s.transport.Close(); err != nil {
		return err
	}
	v2ConnectionsOpen.Dec()
	return nil
}

func (s *V2Sessionless) Version() string {
	return "2.0"
}

func (s *V2Sessionless) sendPayload(ctx context.Context, p ipmi.Payload) error {
	s.rmcpLayer = layers.RMCP{
		Version:  layers.RMCPVersion1,
		Sequence: 0xFF, // do not send us an ACK
		Class:    layers.RMCPClassIPMI,
	}
	s.v2SessionLayer = ipmi.V2Session{
		PayloadDescriptor: *p.Descriptor(),
	}

	// we don't need to increment a sequence number between retries, so can
	// serialise this just once
	// N.B. no message layer as this is only used for RMCP+ session setup (see
	// ipmi.Payload interface for more details)
	if err := gopacket.SerializeLayers(s.buffer, serializeOptions,
		&s.rmcpLayer,
		// session selector only used when decoding
		&s.v2SessionLayer,
		p.Request()); err != nil {
		return err
	}

	if _, err := s.send(ctx); err != nil {
		return err
	}

	// makes it easier to work with
	types := layerexts.DecodedTypes(s.layers)
	if err := types.InnermostEquals(ipmi.LayerTypeV2Session); err != nil {
		return err
	}

	if err := p.Response().DecodeFromBytes(s.v2SessionLayer.LayerPayload(), gopacket.NilDecodeFeedback); err != nil {
		return err
	}
	return nil
}

// saves having to write two SerializeLayers calls in SendCommand
func serializableLayerOrEmpty(s gopacket.SerializableLayer) gopacket.SerializableLayer {
	if s == nil {
		return gopacket.Payload(nil)
	}
	return s
}

func (s *V2Sessionless) SendCommand(ctx context.Context, c ipmi.Command) (ipmi.CompletionCode, error) {
	s.rmcpLayer = layers.RMCP{
		Version:  layers.RMCPVersion1,
		Sequence: 0xFF, // do not send us an ACK
		Class:    layers.RMCPClassIPMI,
	}
	s.v2SessionLayer = ipmi.V2Session{
		PayloadDescriptor: ipmi.PayloadDescriptorIPMI,
	}
	s.messageLayer = ipmi.Message{
		Operation:     *c.Operation(),
		RemoteAddress: ipmi.SlaveAddressBMC.Address(),
		RemoteLUN:     ipmi.LUNBMC,
		LocalAddress:  ipmi.SoftwareIDRemoteConsole1.Address(),
		Sequence:      1,
	}

	// TODO increment metric with c.Name() label here, outside session

	// we don't need to increment a sequence number between retries, so can
	// serialise this just once
	if err := gopacket.SerializeLayers(s.buffer, serializeOptions,
		&s.rmcpLayer,
		// session selector only used when decoding
		&s.v2SessionLayer,
		&s.messageLayer,
		serializableLayerOrEmpty(c.Request())); err != nil {
		return ipmi.CompletionCodeUnspecified, err
	}

	if _, err := s.send(ctx); err != nil {
		return ipmi.CompletionCodeUnspecified, err
	}

	// makes it easier to work with
	types := layerexts.DecodedTypes(s.layers)
	if err := types.InnermostEquals(ipmi.LayerTypeMessage); err != nil {
		return ipmi.CompletionCodeUnspecified, err
	}

	if c.Response() != nil {
		if err := c.Response().DecodeFromBytes(s.messageLayer.LayerPayload(), gopacket.NilDecodeFeedback); err != nil {
			return ipmi.CompletionCodeUnspecified, err
		}
	}
	return s.messageLayer.CompletionCode, nil
}

func (s *V2Sessionless) send(ctx context.Context) (gopacket.LayerType, error) {
	response := []byte(nil)
	ctxErr := error(nil)
	retryable := func() error {
		if err := ctx.Err(); err != nil {
			ctxErr = err
			return nil
		}
		requestCtx, cancel := context.WithTimeout(ctx, time.Second*2) // TODO make configurable
		defer cancel()
		bytes, err := s.transport.Send(requestCtx, s.buffer.Bytes())
		response = bytes
		return err
	}
	s.backoff.Reset()
	if err := backoff.Retry(retryable, s.backoff); err != nil {
		return gopacket.LayerTypeZero, err
	}
	if ctxErr != nil {
		return gopacket.LayerTypeZero, ctxErr
	}

	return s.decode(response, &s.layers)
}

func (s *V2Sessionless) GetSystemGUID(ctx context.Context) ([16]byte, error) {
	return getSystemGUID(ctx, s)
}

func getSystemGUID(ctx context.Context, c Connection) ([16]byte, error) {
	cmd := &ipmi.GetSystemGUIDCmd{}
	if err := ValidateResponse(c.SendCommand(ctx, cmd)); err != nil {
		return [16]byte{}, err
	}

	// we could return a google/uuid type, however that requires the BMC return
	// a valid GUID in network byte order, and the spec says it should be
	// treated as an opaque value. The user can interpret these bytes how they
	// wish.
	return cmd.Rsp.GUID, nil
}

func (s *V2Sessionless) GetChannelAuthenticationCapabilities(
	ctx context.Context,
	r *ipmi.GetChannelAuthenticationCapabilitiesReq,
) (*ipmi.GetChannelAuthenticationCapabilitiesRsp, error) {
	return getChannelAuthenticationCapabilities(ctx, s, r)
}

func getChannelAuthenticationCapabilities(
	ctx context.Context,
	c Connection,
	req *ipmi.GetChannelAuthenticationCapabilitiesReq,
) (*ipmi.GetChannelAuthenticationCapabilitiesRsp, error) {
	// we could set req.ExtendedData here as we're guaranteed to be IPMI v2.0,
	// however let the user decide
	cmd := &ipmi.GetChannelAuthenticationCapabilitiesCmd{
		Req: *req,
	}
	if err := ValidateResponse(c.SendCommand(ctx, cmd)); err != nil {
		return nil, err
	}
	return &cmd.Rsp, nil
}

func (s *V2Sessionless) openSession(ctx context.Context, r *ipmi.OpenSessionReq) (*ipmi.OpenSessionRsp, error) {
	// if we were being *really* aggressive, we could store these payloads in
	// the sessionless struct for reuse during any future session establishments
	payload := &ipmi.OpenSessionPayload{
		Req: *r,
	}
	if err := s.sendPayload(ctx, payload); err != nil {
		return nil, err
	}
	rsp := &payload.Rsp
	if rsp.Tag != r.Tag {
		return nil, fmt.Errorf("tag mismatch; expected %v, got %v", r.Tag,
			rsp.Tag)
	}
	if rsp.Status != ipmi.StatusCodeOK {
		return nil, fmt.Errorf("managed system returned non-OK status: %v",
			rsp.Status)
	}
	return rsp, nil
}

func (s *V2Sessionless) rakpMessage1(ctx context.Context, r *ipmi.RAKPMessage1) (*ipmi.RAKPMessage2, error) {
	payload := &ipmi.RAKPMessage1Payload{
		Req: *r,
	}
	if err := s.sendPayload(ctx, payload); err != nil {
		return nil, err
	}
	rsp := &payload.Rsp
	if rsp.Tag != r.Tag {
		return nil, fmt.Errorf("tag mismatch; expected %v, got %v", r.Tag,
			rsp.Tag)
	}
	if rsp.Status != ipmi.StatusCodeOK {
		return nil, fmt.Errorf("managed system returned non-OK status: %v",
			rsp.Status)
	}
	return rsp, nil
}

func (s *V2Sessionless) rakpMessage3(ctx context.Context, r *ipmi.RAKPMessage3) (*ipmi.RAKPMessage4, error) {
	payload := &ipmi.RAKPMessage3Payload{
		Req: *r,
	}
	if err := s.sendPayload(ctx, payload); err != nil {
		return nil, err
	}
	rsp := &payload.Rsp
	if rsp.Tag != r.Tag {
		return nil, fmt.Errorf("tag mismatch; expected %v, got %v", r.Tag,
			rsp.Tag)
	}
	if rsp.Status != ipmi.StatusCodeOK {
		return nil, fmt.Errorf("managed system returned non-OK status: %v",
			rsp.Status)
	}
	return rsp, nil
}
