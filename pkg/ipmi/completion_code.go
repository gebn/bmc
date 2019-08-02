package ipmi

import (
	"fmt"
)

// CompletionCode indicates whether a command executed successfully. It is
// analogous to a command status code. It is a 1 byte uint on the wire. Values
// are specified in Table 5-2 of the IPMI v2.0 spec.
//
// N.B. if the completion code is not 0, the rest of the response may be
// truncated, and if it is not, the remaining structure is OEM-dependent, so in
// practice the rest of the message should be uninterpreted.
type CompletionCode uint8

func (c CompletionCode) Description() string {
	switch c {
	case CompletionCodeNormal:
		return "Normal"
	case CompletionCodeInvalidSessionID:
		return "Invalid session ID"
	case CompletionCodeNodeBusy:
		return "Node Busy"
	case CompletionCodeUnrecognisedCommand:
		return "Unrecognised Command"
	case CompletionCodeRequestTruncated:
		return "Request Truncated"
	case CompletionCodeUnspecified:
		return "Unspecified error"
	default:
		return "Unknown"
	}
}

func (c CompletionCode) String() string {
	return fmt.Sprintf("%#x(%v)", uint8(c), c.Description())
}

const (
	CompletionCodeNormal CompletionCode = 0x0

	// CompletionCodeInvalidSessionID is returned by Close Session if the
	// specified session ID does not match one the BMC knows about. Untested as
	// to whether this is also returned if the used doesn't have the required
	// privileges.
	CompletionCodeInvalidSessionID CompletionCode = 0x87

	CompletionCodeNodeBusy            CompletionCode = 0xc0
	CompletionCodeUnrecognisedCommand CompletionCode = 0xc1

	// CompletionCodeRequestTruncated means the request ended prematurely. Did
	// you forget to add the final request data layer?
	CompletionCodeRequestTruncated CompletionCode = 0xc6

	CompletionCodeUnspecified CompletionCode = 0xff
)
