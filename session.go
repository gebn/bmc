package bmc

import (
	"context"
)

// Session is an established session-based IPMI v1.5 or 2.0 connection. More
// specifically, it is a multi-session connection, as the spec demands this of
// the LAN interface. Commands are sent in the context of the session, using
// the negotiated integrity and confidentiality algorithms.
type Session interface {
	Connection
	SessionCommands

	// ID returns our identifier for this session. Note the managed system and
	// remote console share the same identifier in IPMI v1.5, however each
	// chooses its own identifier in v2.0, so they likely differ.
	ID() uint32

	// Close closes the session by sending a Close Session command to the BMC.
	// As the underlying transport/socket is used but not managed by
	// connections, it is left open in case the user wants to continue issuing
	// session-less commands or establish a new session. It is envisaged that
	// this call is deferred immediately after successful session establishment.
	// If an error is returned, the session can be assumed to be closed.
	Close(context.Context) error
}
