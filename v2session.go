package bmc

import (
	"context"
	"encoding/hex"
	"fmt"
	"hash"
	"time"

	"github.com/gebn/bmc/pkg/ipmi"
	"github.com/gebn/bmc/pkg/layerexts"

	"github.com/cenkalti/backoff"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
)

type v2SessionRspLayers struct {
	getDeviceIDRspLayer      ipmi.GetDeviceIDRsp
	getChassisStatusRspLayer ipmi.GetChassisStatusRsp
}

// V2Session represents an established IPMI v2.0/RMCP+ session with a BMC.
type V2Session struct {
	v2ConnectionShared
	v2SessionRspLayers

	// LocalID is the remote console's session ID, used by the BMC to send us
	// packets.
	LocalID uint32

	// RemoteID is the managed system's session ID, used by us to send the BMC
	// packets.
	RemoteID uint32

	// AuthenticatedSequenceNumbers is the pair of sequence numbers for
	// authenticated packets.
	AuthenticatedSequenceNumbers sequenceNumbers

	// UnauthenticatedSequenceNumbers is the pair of sequence numbers for
	// unauthenticated packets, i.e. those without an auth code. We only send
	// unauthenticated packets to the BMC if IntegrityAlgorithmNone was
	// negotiated.
	UnauthenticatedSequenceNumbers sequenceNumbers

	// SIK is the session integrity key, whose creation is described in section
	// 13.31 of the spes. It is the result of applying the negotiated
	// authentication algorithm (which is usually, but may not be, an HMAC) to
	// some random numbers, the remote console's requested maximum privilege
	// level, and username. The SIK is then used to derive K_1 and K_2 (and
	// possibly more, but not for any algorithms in the spec) which are the keys
	// for the integrity algorithm and confidentiality algorithms respectively.
	SIK []byte

	// AuthenticationAlgorithm is the algorithm used to authenticate the user
	// during establishment of the session. Given the session is already
	// established, this will not be used further.
	AuthenticationAlgorithm ipmi.AuthenticationAlgorithm

	// IntegrityAlgorithm is the algorithm used to sign, or authenticate packets
	// sent by the managed system and remote console. This library authenticates
	// all packets it sends inside a session, provided IntegrityAlgorithmNone
	// was not negotiated.
	IntegrityAlgorithm ipmi.IntegrityAlgorithm

	// ConfidentialityAlgorithm is the algorithm used to encrypt and decrypt
	// packets sent by the managed system and remote console. This library
	// encrypts all packets it sends inside a session, provided
	// ConfidentialityAlgorithmNone was not negotiated.
	ConfidentialityAlgorithm ipmi.ConfidentialityAlgorithm

	// AdditionalKeyMaterialGenerator is the instance of the authentication
	// algorithm used during session establishment, loaded with the session
	// integrity key. It has no further use as far as the BMC is concerned by
	// the time we have this struct, however we keep it around to allow
	// providing K_n for information purposes.
	AdditionalKeyMaterialGenerator

	integrityAlgorithm hash.Hash

	// confidentialityLayer is used to send packets (we encrypt all outgoing
	// packets), and to decode incoming packets. It is created during session
	// establishment, and loaded with the right key material. The session
	// layer's ConfidentialityLayerType field is set to this layer's type, so it
	// returns this as the NextLayerType() of encrypted packets. When sending a
	// message, this layer's SerializeTo is called before adding the session
	// wrapper.
	confidentialityLayer layerexts.SerializableDecodingLayer

	// ARGH fix this duplication between here and v2sessionless!
	layers []gopacket.LayerType

	// parser is a packet decoder loaded with all relevant layer types, specific
	// to this session.
	parser *gopacket.DecodingLayerParser
}

// String returns a summary of the session's attributes on one line.
func (s *V2Session) String() string {
	return fmt.Sprintf("V2Session(Authentication: %v, Integrity: %v, Confidentiality: %v, LocalID: %v, RemoteID: %v, SIK: %v, K_1: %v, K_2: %v)",
		s.AuthenticationAlgorithm, s.IntegrityAlgorithm, s.ConfidentialityAlgorithm,
		s.LocalID, s.RemoteID,
		hex.EncodeToString(s.SIK),
		hex.EncodeToString(s.K(1)), hex.EncodeToString(s.K(2)))
}

func (s *V2Session) Version() string {
	return "2.0"
}

func (s *V2Session) ID() uint32 {
	return s.LocalID
}

// first layer must be the session layer
func (s *V2Session) sendPayload(ctx context.Context, ls ...gopacket.SerializableLayer) (layerexts.DecodedTypes, error) {
	s.rmcpLayer = layers.RMCP{
		Version:  layers.RMCPVersion1,
		Sequence: 0xFF, // do not send us an ACK
		Class:    layers.RMCPClassIPMI,
	}

	// we can't mix direct arguments and slices when passing variadic args:
	// https://stackoverflow.com/a/18949245
	ls = append([]gopacket.SerializableLayer{
		&s.rmcpLayer,
	}, ls...)

	response := []byte(nil)
	terminalErr := error(nil)
	retryable := func() error {
		if err := ctx.Err(); err != nil {
			terminalErr = err
			return nil
		}
		// TODO handle AuthenticationAlgorithmNone properly
		s.AuthenticatedSequenceNumbers.Inbound++
		s.v2SessionLayer.Sequence = s.AuthenticatedSequenceNumbers.Inbound
		if err := gopacket.SerializeLayers(s.buffer, serializeOptions, ls...); err != nil {
			//&s.rmcpLayer,
			//&s.v2SessionLayer,
			//s.confidentialityLayer, // TODO handle ConfidentialityAlgorithmNone
			//&s.messageLayer,
			//cmd); err != nil {
			// this is not a retryable error
			terminalErr = err
			return nil
		}
		requestCtx, cancel := context.WithTimeout(ctx, time.Second*2) // TODO make configurable
		defer cancel()
		resp, err := s.transport.Send(requestCtx, s.buffer.Bytes())
		response = resp
		return err
	}
	if err := backoff.Retry(retryable, backoff.NewExponentialBackOff()); err != nil {
		return nil, err
	}
	if terminalErr != nil {
		return nil, terminalErr
	}
	if err := s.parser.DecodeLayers(response, &s.layers); err != nil {
		return nil, err
	}
	return layerexts.DecodedTypes(s.layers), nil
}

func (s *V2Session) SendMessage(ctx context.Context, op *ipmi.Operation, cmd gopacket.SerializableLayer) (layerexts.DecodedTypes, ipmi.CompletionCode, error) {

	// TODO fix duplication between here and V2Sessionless

	s.v2SessionLayer = ipmi.V2Session{
		Encrypted:                true,
		Authenticated:            true,
		ID:                       s.RemoteID,
		Payload:                  ipmi.PayloadIPMI,
		IntegrityAlgorithm:       s.integrityAlgorithm,
		ConfidentialityLayerType: s.confidentialityLayer.LayerType(),
	}
	s.messageLayer = ipmi.Message{
		Operation:     *op,
		RemoteAddress: ipmi.SlaveAddressBMC.Address(),
		RemoteLUN:     ipmi.LUNBMC,
		LocalAddress:  ipmi.SoftwareIDRemoteConsole1.Address(),
		Sequence:      1, // used at the session level
	}

	// allows passing nil as the final parameter for commands with no payload
	if cmd == nil {
		cmd = gopacket.Payload(nil)
	}

	layers, err := s.sendPayload(ctx, &s.v2SessionLayer, s.confidentialityLayer,
		&s.messageLayer, cmd)
	if err != nil {
		return layers, ipmi.CompletionCodeUnspecified, err
	}

	// ensure message layer returned so we have a completion code
	if err := layers.Contains(ipmi.LayerTypeMessage); err != nil {
		return layers, ipmi.CompletionCodeUnspecified, err
	}

	return layers, s.messageLayer.CompletionCode, nil
}

func (s *V2Session) GetSystemGUID(ctx context.Context) ([16]byte, error) {
	return getSystemGUID(ctx, s, &s.getSystemGUIDRspLayer)
}

func (s *V2Session) GetChannelAuthenticationCapabilities(ctx context.Context, r *ipmi.GetChannelAuthenticationCapabilitiesReq) (*ipmi.GetChannelAuthenticationCapabilitiesRsp, error) {
	return getChannelAuthenticationCapabilities(ctx, s, r,
		&s.getChannelAuthenticationCapabilitiesRspLayer)
}

func (s *V2Session) GetDeviceID(ctx context.Context) (*ipmi.GetDeviceIDRsp, error) {
	layers, code, err := s.SendMessage(ctx, &ipmi.OperationGetDeviceIDReq, nil)
	if err != nil {
		return nil, err
	}
	if err := validateCompletionCode(code); err != nil {
		return nil, err
	}
	if err := layers.InnermostEquals(ipmi.LayerTypeGetDeviceIDRsp); err != nil {
		return nil, err
	}
	return &s.getDeviceIDRspLayer, nil
}

func (s *V2Session) GetChassisStatus(ctx context.Context) (*ipmi.GetChassisStatusRsp, error) {
	layers, code, err := s.SendMessage(ctx, &ipmi.OperationGetChassisStatusReq, nil)
	if err != nil {
		return nil, err
	}
	if err := validateCompletionCode(code); err != nil {
		return nil, err
	}
	if err := layers.InnermostEquals(ipmi.LayerTypeGetChassisStatusRsp); err != nil {
		return nil, err
	}
	return &s.getChassisStatusRspLayer, nil
}

func (s *V2Session) ChassisControl(ctx context.Context, c ipmi.ChassisControl) error {
	_, code, err := s.SendMessage(ctx, &ipmi.OperationChassisControlReq,
		&ipmi.ChassisControlReq{
			ChassisControl: c,
		})
	if err != nil {
		return err
	}
	if err := validateCompletionCode(code); err != nil {
		return err
	}
	return nil
}

func (s *V2Session) closeSession(ctx context.Context) error {
	_, code, err := s.SendMessage(ctx, &ipmi.OperationCloseSessionReq,
		&ipmi.CloseSessionReq{
			ID: s.RemoteID,
		})
	if err != nil {
		return err
	}
	if err := validateCompletionCode(code); err != nil {
		return err
	}
	return nil
}

func (s *V2Session) Close(ctx context.Context) error {
	return s.closeSession(ctx)
}
