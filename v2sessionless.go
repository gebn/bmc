package bmc

import (
	"context"
	"fmt"
	"time"

	"github.com/gebn/bmc/pkg/ipmi"
	"github.com/gebn/bmc/pkg/layerexts"

	"github.com/cenkalti/backoff"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
)

// v2SessionlessRspLayers contains all layers we expect to receive over a
// sessionless IPMI v2.0 connection. We build a DecodingLayerParser from this.
// The idea is this is a single allocation which we can efficiently embed in the
// sessionless connection struct, and then share with a V2Session if/when one is
// created. Because we only allow a single command to be sent on a socket at a
// time, this is safe, and saves some memory.
type v2SessionlessRspLayers struct {
	rmcpLayer            layers.RMCP
	sessionSelectorLayer ipmi.SessionSelector
	v2SessionLayer       ipmi.V2Session

	messageLayer                                 ipmi.Message
	getSystemGUIDRspLayer                        ipmi.GetSystemGUIDRsp
	getChannelAuthenticationCapabilitiesRspLayer ipmi.GetChannelAuthenticationCapabilitiesRsp

	payloadLayer gopacket.Payload // always the final layer, usually empty
}

// this separates these layers from being used in a V2 session - they will never
// be received inside a session
type v2SessionEstablishmentRspLayers struct {
	openSessionRspLayer ipmi.OpenSessionRsp
	rakpMessage2Layer   ipmi.RAKPMessage2
	rakpMessage4Layer   ipmi.RAKPMessage4
}

type v2ConnectionShared struct {
	v2SessionlessRspLayers

	transport transport

	// buffer is used to build all packets to send during this connection.
	// Reusing this drastically reduces the number of allocations we have to do
	// when building packets.
	buffer gopacket.SerializeBuffer
}

// V2Sessionless represents a session-less connection to a BMC using a "null"
// IPMI v2.0 session wrapper.
type V2Sessionless struct {
	v2ConnectionShared
	v2SessionEstablishmentRspLayers

	layers []gopacket.LayerType
	parser *gopacket.DecodingLayerParser
}

func newV2Sessionless(t transport) *V2Sessionless {
	s := &V2Sessionless{
		v2ConnectionShared: v2ConnectionShared{
			transport: t,
			buffer:    gopacket.NewSerializeBuffer(),
		},
	}
	s.parser = gopacket.NewDecodingLayerParser(
		s.rmcpLayer.LayerType(),
		&s.rmcpLayer,
		&s.sessionSelectorLayer,
		&s.v2SessionLayer,
		&s.messageLayer,
		&s.getSystemGUIDRspLayer,
		&s.getChannelAuthenticationCapabilitiesRspLayer,
		&s.payloadLayer,
		&s.openSessionRspLayer,
		&s.rakpMessage2Layer,
		&s.rakpMessage4Layer)
	return s
}

func (s *V2Sessionless) Version() string {
	return "2.0"
}

func (s *V2Sessionless) sendPayload(
	ctx context.Context,
	p *ipmi.Payload,
	ls ...gopacket.SerializableLayer,
) (layerexts.DecodedTypes, error) {
	s.rmcpLayer = layers.RMCP{
		Version:  layers.RMCPVersion1,
		Sequence: 0xFF, // do not send us an ACK
		Class:    layers.RMCPClassIPMI,
	}
	s.v2SessionLayer = ipmi.V2Session{
		Payload: *p,
	}

	// we can't mix direct arguments and slices when passing variadic args:
	// https://stackoverflow.com/a/18949245
	ls = append([]gopacket.SerializableLayer{
		&s.rmcpLayer,
		&s.v2SessionLayer,
	}, ls...)

	// we don't need to increment a sequence number between retries, so can
	// serialise this just once
	if err := gopacket.SerializeLayers(s.buffer, serializeOptions, ls...); err != nil {
		return nil, err
	}

	bytes := []byte(nil)
	retryable := func() error {
		requestCtx, cancel := context.WithTimeout(ctx, time.Second*2) // TODO make configurable
		defer cancel()
		resp, err := s.transport.Send(requestCtx, s.buffer.Bytes())
		bytes = resp
		return err
	}
	if err := backoff.Retry(retryable, backoff.NewExponentialBackOff()); err != nil {
		return nil, err
	}

	if err := s.parser.DecodeLayers(bytes, &s.layers); err != nil {
		return nil, err
	}

	return layerexts.DecodedTypes(s.layers), nil
}

func (s *V2Sessionless) SendMessage(
	ctx context.Context,
	op *ipmi.Operation,
	cmd gopacket.SerializableLayer,
) (layerexts.DecodedTypes, ipmi.CompletionCode, error) {
	s.messageLayer = ipmi.Message{
		Operation:     *op,
		RemoteAddress: ipmi.SlaveAddressBMC.Address(),
		RemoteLUN:     ipmi.LUNBMC,
		LocalAddress:  ipmi.SoftwareIDRemoteConsole1.Address(),
		Sequence:      1,
	}

	// allows passing nil as the final parameter for commands with no payload
	if cmd == nil {
		cmd = gopacket.Payload(nil)
	}

	layers, err := s.sendPayload(ctx, &ipmi.PayloadIPMI, &s.messageLayer, cmd)
	if err != nil {
		return layers, ipmi.CompletionCodeUnspecified, err
	}

	// ensure message layer returned so we have a completion code
	if err := layers.Contains(ipmi.LayerTypeMessage); err != nil {
		return layers, ipmi.CompletionCodeUnspecified, err
	}

	return layers, s.messageLayer.CompletionCode, nil
}

func (s *V2Sessionless) GetSystemGUID(ctx context.Context) ([16]byte, error) {
	return getSystemGUID(ctx, s, &s.getSystemGUIDRspLayer)
}

func getSystemGUID(ctx context.Context, c connection, l *ipmi.GetSystemGUIDRsp) ([16]byte, error) {
	layers, code, err := c.SendMessage(ctx, &ipmi.OperationGetSystemGUIDReq, nil)
	if err != nil {
		return [16]byte{}, err
	}
	if err := validateCompletionCode(code); err != nil {
		return [16]byte{}, err
	}
	if err := layers.InnermostEquals(l.LayerType()); err != nil {
		return [16]byte{}, err
	}

	// we could return a google/uuid type, however that requires the BMC return
	// a valid GUID in network byte order, and the spec says it should be
	// treated as an opaque value. The user can interpret these bytes how they
	// wish.
	return l.GUID, nil
}

func (s *V2Sessionless) GetChannelAuthenticationCapabilities(
	ctx context.Context,
	r *ipmi.GetChannelAuthenticationCapabilitiesReq,
) (*ipmi.GetChannelAuthenticationCapabilitiesRsp, error) {
	return getChannelAuthenticationCapabilities(ctx, s, r,
		&s.getChannelAuthenticationCapabilitiesRspLayer)
}

func getChannelAuthenticationCapabilities(
	ctx context.Context,
	c connection,
	req *ipmi.GetChannelAuthenticationCapabilitiesReq,
	rsp *ipmi.GetChannelAuthenticationCapabilitiesRsp,
) (*ipmi.GetChannelAuthenticationCapabilitiesRsp, error) {
	// we could set req.ExtendedData here as we're guaranteed to be IPMI v2.0,
	// however let the user decide
	layers, code, err := c.SendMessage(ctx,
		&ipmi.OperationGetChannelAuthenticationCapabilitiesReq, req)
	if err != nil {
		return nil, err
	}
	if err := validateCompletionCode(code); err != nil {
		return nil, err
	}
	if err := layers.InnermostEquals(rsp.LayerType()); err != nil {
		return nil, err
	}
	return rsp, nil
}

func (s *V2Sessionless) openSession(
	ctx context.Context,
	r *ipmi.OpenSessionReq,
) (*ipmi.OpenSessionRsp, error) {
	layers, err := s.sendPayload(ctx, &ipmi.PayloadOpenSessionReq, r)
	if err != nil {
		return nil, err
	}
	if err := layers.InnermostEquals(ipmi.LayerTypeOpenSessionRsp); err != nil {
		return nil, err
	}
	if s.openSessionRspLayer.Tag != r.Tag {
		return nil, fmt.Errorf("tag mismatch; expected %v, got %v", r.Tag,
			s.openSessionRspLayer.Tag)
	}
	if s.openSessionRspLayer.Status != ipmi.StatusCodeOK {
		return nil, fmt.Errorf("managed system returned non-OK status: %v",
			s.openSessionRspLayer.Status)
	}
	return &s.openSessionRspLayer, nil
}

func (s *V2Sessionless) rakpMessage1(
	ctx context.Context,
	r *ipmi.RAKPMessage1,
) (*ipmi.RAKPMessage2, error) {
	layers, err := s.sendPayload(ctx, &ipmi.PayloadRAKPMessage1, r)
	if err != nil {
		return nil, err
	}
	if err := layers.InnermostEquals(ipmi.LayerTypeRAKPMessage2); err != nil {
		return nil, err
	}
	if s.rakpMessage2Layer.Tag != r.Tag {
		return nil, fmt.Errorf("tag mismatch; expected %v, got %v", r.Tag,
			s.openSessionRspLayer.Tag)
	}
	if s.rakpMessage2Layer.Status != ipmi.StatusCodeOK {
		return nil, fmt.Errorf("managed system returned non-OK status: %v",
			s.rakpMessage2Layer.Status)
	}
	return &s.rakpMessage2Layer, nil
}

func (s *V2Sessionless) rakpMessage3(
	ctx context.Context,
	r *ipmi.RAKPMessage3,
) (*ipmi.RAKPMessage4, error) {
	layers, err := s.sendPayload(ctx, &ipmi.PayloadRAKPMessage3, r)
	if err != nil {
		return nil, err
	}
	if err := layers.InnermostEquals(ipmi.LayerTypeRAKPMessage4); err != nil {
		return nil, err
	}
	if s.rakpMessage4Layer.Tag != r.Tag {
		return nil, fmt.Errorf("tag mismatch; expected %v, got %v", r.Tag,
			s.openSessionRspLayer.Tag)
	}
	if s.rakpMessage4Layer.Status != ipmi.StatusCodeOK {
		return nil, fmt.Errorf("managed system returned non-OK status: %v",
			s.rakpMessage4Layer.Status)
	}
	return &s.rakpMessage4Layer, nil
}
