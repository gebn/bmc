package ipmi

// SessionHandle uniquely identifies a session within the context of a given
// channel, as opposed to globally for the BMC which is the case for SessionID.
// A typical BMC keeps track of the last handle that was assigned, and
// increments it for new sessions. 0x00 is reserved.
type SessionHandle uint8
