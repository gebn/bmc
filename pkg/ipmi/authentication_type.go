package ipmi

import (
	"fmt"
)

// AuthenticationType is used in the IPMI session header to indicate which
// authentication algorithm was used to sign the message. It is a 4-bit uint on
// the wire.
type AuthenticationType uint8

const (
	AuthenticationTypeNone     AuthenticationType = 0x0
	AuthenticationTypeMD2      AuthenticationType = 0x1
	AuthenticationTypeMD5      AuthenticationType = 0x2
	AuthenticationTypePassword AuthenticationType = 0x3
	AuthenticationTypeOEM      AuthenticationType = 0x5
	AuthenticationTypeRMCPPlus AuthenticationType = 0x6 // IPMI v2 only
)

func (t AuthenticationType) name() string {
	switch t {
	case AuthenticationTypeNone:
		return "None"
	case AuthenticationTypeMD2:
		return "MD2"
	case AuthenticationTypeMD5:
		return "MD5"
	case AuthenticationTypePassword:
		return "Password/Key"
	case AuthenticationTypeOEM:
		return "OEM"
	case AuthenticationTypeRMCPPlus:
		return "RMCP+"
	default:
		return "Unknown"
	}
}

func (t AuthenticationType) String() string {
	return fmt.Sprintf("%v(%v)", uint8(t), t.name())
}
