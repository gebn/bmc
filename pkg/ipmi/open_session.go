package ipmi

import (
	"encoding/binary"
	"fmt"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
)

// OpenSessionReq represents an RMCP+ Open Session Request message, specified in
// section 13.17.
type OpenSessionReq struct {
	layers.BaseLayer

	// Tag is copied into the BMC's Open Session Response message to help the
	// remote console match it up with this request. If the remote console
	// retries this message, it should increment this.
	Tag uint8

	// MaxPrivilegeLevel is the highest privilege level the remote console wants
	// the BMC to allow for the session. If this is 0x0 (Highest), the BMC will
	// give us the highest level it is willing to, given the cipher suites the
	// remote console indicated support for.
	MaxPrivilegeLevel PrivilegeLevel

	// SessionID is what the remote console wants the BMC to use to identify
	// packets sent to it for this session. This should not be null, to avoid
	// conflicting with out-of-session messaging. N.B. this is not what the
	// remote console uses to send packets to the BMC.
	SessionID uint32

	// AuthenticationPayloads is a slice of authentication payloads to include
	// in the request. This must have a length of at least 1. If a wildcard
	// payload is specified, it should be the only one in the slice.
	AuthenticationPayloads []AuthenticationPayload

	// IntegrityPayloads is a slice of integrity payloads to include in the
	// request. This must have a length of at least 1. If a wildcard payload is
	// specified, it should be the only one in the slice.
	IntegrityPayloads []IntegrityPayload

	// ConfidentialityPayloads is a slice of confidentiality payloads to include
	// in the request. This must have a length of at least 1. If a wildcard
	// payload is specified, it should be the only one in the slice.
	ConfidentialityPayloads []ConfidentialityPayload
}

func (*OpenSessionReq) LayerType() gopacket.LayerType {
	return LayerTypeOpenSessionReq
}

func (o *OpenSessionReq) SerializeTo(b gopacket.SerializeBuffer, opts gopacket.SerializeOptions) error {
	// We make no assumptions about algorithm payload lengths, however in
	// practice, all sorts of stuff would break on the response side if they
	// were not all 8 bytes. Fortunately, this does not look likely to ever
	// change.
	d, err := b.PrependBytes(8)
	if err != nil {
		return err
	}
	d[0] = o.Tag
	d[1] = uint8(o.MaxPrivilegeLevel) & 0xF
	d[2] = 0x00
	d[3] = 0x00
	binary.LittleEndian.PutUint32(d[4:8], o.SessionID)
	for _, p := range o.AuthenticationPayloads {
		if err := p.Serialise(b); err != nil {
			return err
		}
	}
	for _, p := range o.IntegrityPayloads {
		if err := p.Serialise(b); err != nil {
			return err
		}
	}
	for _, p := range o.ConfidentialityPayloads {
		if err := p.Serialise(b); err != nil {
			return err
		}
	}
	return nil
}

// OpenSessionRsp represents an RMCP+ Open Session Response message, specified
// in section 13.18. This is distinct from the RAKP messages, partly because
// even if a RAKP message fails, the open session request and response does not
// have to be repeated, as they are stateless.
type OpenSessionRsp struct {
	layers.BaseLayer

	// Tag is the tag passed in the request, to ease matching up the response.
	Tag uint8

	// Status is the RMCP+ status code indicating whether the BMC was able to
	// service the request.
	Status StatusCode

	// MaxPrivilegeLevel is the Maximum Privilege Level allowed for the
	// session based on the security algorithms that were proposed in the
	// request. It will be 0 if the status is not OK.
	MaxPrivilegeLevel PrivilegeLevel

	// RemoteConsoleSessionID is an echo of the session ID the remote console
	// asked the BMC to use in its request.
	RemoteConsoleSessionID uint32

	// if the Status is not OK, the packet effectively stops here

	// ManagedSystemSessionID is the session ID the BMC would like the remote
	// console to use when sending it messages for this session. This will not
	// be null, as that would conflict with out-of-session messaging.
	ManagedSystemSessionID uint32

	// AuthenticationPayload contains the authentication algorithm selected by
	// the managed system for the session. This should not be a wildcard or
	// none, but this is not validated.
	AuthenticationPayload AuthenticationPayload

	// IntegrityPayload contains the integrity algorithm selected by the managed
	// system for the session. This should not be a wildcard or none, but this
	// is not validated.
	IntegrityPayload IntegrityPayload

	// ConfidentialityPayload contains the confidentiality algorithm selected by
	// the managed system for the session. This should not be a wildcard or
	// none, but this is not validated.
	ConfidentialityPayload ConfidentialityPayload
}

func (*OpenSessionRsp) LayerType() gopacket.LayerType {
	return LayerTypeOpenSessionRsp
}

func (o *OpenSessionRsp) CanDecode() gopacket.LayerClass {
	return o.LayerType()
}

func (*OpenSessionRsp) NextLayerType() gopacket.LayerType {
	return gopacket.LayerTypePayload
}

func (o *OpenSessionRsp) DecodeFromBytes(data []byte, df gopacket.DecodeFeedback) error {
	if len(data) < 8 { // minimum in case of non-zero status code
		df.SetTruncated()
		return fmt.Errorf("RMCP+ Open Session Response must be at least 8 bytes, got %v", len(data))
	}
	o.Tag = uint8(data[0])
	o.Status = StatusCode(data[1])
	o.MaxPrivilegeLevel = PrivilegeLevel(data[2])
	// [3] reserved
	o.RemoteConsoleSessionID = binary.LittleEndian.Uint32(data[4:8])

	if o.Status == StatusCodeOK {
		if len(data) != 36 {
			df.SetTruncated()
			return fmt.Errorf("Success RMCP+ Open Session Response must be 36 bytes long, got %v", len(data))
		}
		o.BaseLayer.Contents = data[:36]
		o.ManagedSystemSessionID = binary.LittleEndian.Uint32(data[8:12])
		if _, err := o.AuthenticationPayload.Deserialise(data[12:20], df); err != nil {
			return err
		}
		if _, err := o.IntegrityPayload.Deserialise(data[20:28], df); err != nil {
			return err
		}
		if _, err := o.ConfidentialityPayload.Deserialise(data[28:36], df); err != nil {
			return err
		}
	} else {
		o.BaseLayer.Contents = data[:8]
		o.ManagedSystemSessionID = 0
		o.AuthenticationPayload = AuthenticationPayload{}
		o.IntegrityPayload = IntegrityPayload{}
		o.ConfidentialityPayload = ConfidentialityPayload{}
	}
	return nil
}

type OpenSessionPayload struct {
	Req OpenSessionReq
	Rsp OpenSessionRsp
}

// Descriptor returns PayloadDescriptorOpenSessionReq.
func (*OpenSessionPayload) Descriptor() *PayloadDescriptor {
	return &PayloadDescriptorOpenSessionReq
}

func (p *OpenSessionPayload) Request() gopacket.SerializableLayer {
	return &p.Req
}

func (p *OpenSessionPayload) Response() gopacket.DecodingLayer {
	return &p.Rsp
}
