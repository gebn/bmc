package ipmi

import (
	"fmt"

	"github.com/gebn/bmc/pkg/iana"

	"github.com/google/gopacket"
)

// Payload contains RMCP+ session fields which, taken together, describe the
// format of a IPMI v2.0 session payload. V2Session embeds this type. This type
// does not appear in the specification.
type Payload struct {

	// PayloadType identifies the payload, e.g. an IPMI or RAKP message. When
	// this has a value of OEM (0x2), it must be used together with the
	// Enterprise and PayloadID fields to identify the format.
	PayloadType PayloadType

	// Enterprise is the IANA Enterprise Number of the OEM who describes the
	// payload. This field only exists on the wire if the payload type is OEM
	// explicit.
	Enterprise iana.Enterprise

	// PayloadID identifies the payload within the Enterprise when the payload
	// is OEM-defined. This field only exists on the wire if the payload type is
	// OEM explicit.
	PayloadID uint16
}

var (
	PayloadIPMI = Payload{
		PayloadType: PayloadTypeIPMI,
	}
	PayloadOpenSessionReq = Payload{
		PayloadType: PayloadTypeOpenSessionReq,
	}
	PayloadOpenSessionRsp = Payload{
		PayloadType: PayloadTypeOpenSessionRsp,
	}
	PayloadRAKPMessage1 = Payload{
		PayloadType: PayloadTypeRAKPMessage1,
	}
	PayloadRAKPMessage2 = Payload{
		PayloadType: PayloadTypeRAKPMessage2,
	}
	PayloadRAKPMessage3 = Payload{
		PayloadType: PayloadTypeRAKPMessage3,
	}
	PayloadRAKPMessage4 = Payload{
		PayloadType: PayloadTypeRAKPMessage4,
	}

	payloadLayerTypes = map[Payload]gopacket.LayerType{
		PayloadIPMI:           LayerTypeMessage,
		PayloadOpenSessionReq: LayerTypeOpenSessionReq,
		PayloadOpenSessionRsp: LayerTypeOpenSessionRsp,
		PayloadRAKPMessage1:   LayerTypeRAKPMessage1,
		PayloadRAKPMessage2:   LayerTypeRAKPMessage2,
		PayloadRAKPMessage3:   LayerTypeRAKPMessage3,
		PayloadRAKPMessage4:   LayerTypeRAKPMessage4,
	}
)

func (p Payload) NextLayerType() gopacket.LayerType {
	if layer, ok := payloadLayerTypes[p]; ok {
		return layer
	}
	return gopacket.LayerTypePayload
}

func (p Payload) String() string {
	switch p.PayloadType {
	case PayloadTypeOEM:
		return fmt.Sprintf("Payload(OEM, %v, %#x", p.Enterprise, p.PayloadID)
	default:
		return fmt.Sprintf("Payload(%v)", p.PayloadType)
	}
}

// RegisterOEMPayload adds or overrides how an IPMI v2.0 OEM payload is handled
// within a session. This is implemented via a map, so care must be taken to not
// call this function in parallel.
func RegisterOEMPayload(enterprise iana.Enterprise, payloadID uint16, LayerType gopacket.LayerType) {
	payload := Payload{
		PayloadType: PayloadTypeOEM,
		Enterprise:  enterprise,
		PayloadID:   payloadID,
	}
	payloadLayerTypes[payload] = LayerType
}
