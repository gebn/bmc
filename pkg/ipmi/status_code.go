package ipmi

// StatusCode represents an RMCP+ status code. A value of this type is contained
// in the RMCP+ Open Session Response and RAKP Messages 2, 3 and 4. This is the
// equivalent of an IPMI completion code. See section 13.24 for the full list of
// definitions.
type StatusCode uint8

const (
	// StatusCodeOK indicates successful completion, absent of error. This can
	// exist in all message types.
	StatusCodeOK StatusCode = 0x00

	// StatusCodeInsufficientResources indicates there were insufficient
	// resources to create a session. This can exist in all message types.
	StatusCodeInsufficientResources StatusCode = 0x01

	// StatusCodeInvalidSessionID indicates the managed system or remote console
	// does not recognise the session ID sent by the other end. In practice, the
	// remote console will likely be at fault. This can exist in all message
	// types.
	StatusCodeInvalidSessionID StatusCode = 0x02

	// StatusCodeUnauthorizedName is sent in RAKP Message 2 to indicate the
	// username was not found in the BMC's users table.
	StatusCodeUnauthorizedName StatusCode = 0x0d
)

func (s StatusCode) String() string {
	switch s {
	case StatusCodeOK:
		return "OK"
	case StatusCodeInsufficientResources:
		return "Insufficient resources"
	case StatusCodeInvalidSessionID:
		return "Invalid session ID"
	case StatusCodeUnauthorizedName:
		return "Unauthorized user"
	default:
		return "Unknown"
	}
}
