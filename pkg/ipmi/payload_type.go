package ipmi

// PayloadType identifies the layer immediately within the RMCP+ session
// wrapper. Values are specified in 13.27.3 of the IPMI v2.0 spec. This is a
// 6-bit uint on the wire.
type PayloadType uint8

const (
	// "standard" payload types

	PayloadTypeIPMI PayloadType = 0x0

	// PayloadTypeOEM means "check the OEM IANA and OEM payload ID to find out
	// what this actually is".
	PayloadTypeOEM PayloadType = 0x2

	// "session setup" payload types

	PayloadTypeOpenSessionReq PayloadType = 0x10
	PayloadTypeOpenSessionRsp PayloadType = 0x11
	PayloadTypeRAKPMessage1   PayloadType = 0x12
	PayloadTypeRAKPMessage2   PayloadType = 0x13
	PayloadTypeRAKPMessage3   PayloadType = 0x14
	PayloadTypeRAKPMessage4   PayloadType = 0x15
)

func (p PayloadType) String() string {
	switch p {
	case PayloadTypeIPMI:
		return "IPMI"
	case PayloadTypeOEM:
		return "OEM Explicit"
	case PayloadTypeOpenSessionReq:
		return "RMCP+ Open Session Request"
	case PayloadTypeOpenSessionRsp:
		return "RMCP+ Open Session Response"
	case PayloadTypeRAKPMessage1:
		return "RAKP Message 1"
	case PayloadTypeRAKPMessage2:
		return "RAKP Message 2"
	case PayloadTypeRAKPMessage3:
		return "RAKP Message 3"
	case PayloadTypeRAKPMessage4:
		return "RAKP Message 4"
	default:
		return "Unknown" // possibly OEM (0x20 through 0x27)
	}
}
