package bmc

import (
	"context"

	"github.com/gebn/bmc/pkg/ipmi"
	"github.com/gebn/bmc/pkg/layerexts"

	"github.com/google/gopacket"
)

// Connection is an IPMI v1.5 or v2.0 session-less, single-session or
// multi-session connection. The IPMI version and nature of the connection is
// fixed upon creation - if sending two messages, it will never be the case that
// one uses one wrapper format and the second another. It defines logical things
// that can be done once communication is established with a BMC. Note that this
// is *not* a transport in itself - hence why there is no Close() - but it
// abstracts over a transport to provide its functionality. This interface is
// always wrapped in something else that has a Close() to cleanly terminate the
// underlying connection.
type Connection interface {

	// SendMessage sends a command to the BMC, blocking until it receives a
	// response. If there is no command layer, this should be set to nil. This
	// method will retry with the configured per-request timeout until a valid
	// response is received, or the context expires (whichever happens first). A
	// non-zero completion code is deemed to be a valid response. If the final
	// request fails with a transport error or timeout, the error will be
	// non-nil, and the completion code must be ignored.
	//
	// This method does not return the response layer as it is assumed the
	// caller knows which layer to expect, and has it in the correct type. The
	// caller should first check the error, then the completion code, then
	// assuming both indicate no error, for the layer it expects. This method
	// will return an error if the response does not contain a message layer.
	//
	// This is a low-level method, used to implement the higher-level IPMI
	// commands. If possible, it is recommended to use those instead for
	// simplicity.
	SendMessage(ctx context.Context, op *ipmi.Operation, cmd gopacket.SerializableLayer) (layerexts.DecodedTypes, ipmi.CompletionCode, error)

	// Version returns the underlying IPMI version of the connection, either
	// "1.5" or "2.0". Note that even session-less connections use a session
	// wrapper, which has either the v1.5 or v2.0 format. This is provided for
	// informational and debugging purposes - branching based on this value is a
	// code smell.
	Version() string
}

// Sessionless is a session-less IPMI v1.5 or 2.0 connection. It enables the
// sending of commands to a BMC outside of the context of a session (however
// note that all such commands can also be validly sent inside a session, for
// example Get Channel Authentication Capabilities is commonly used as a form of
// keepalive). Creating a concrete session-less connection will require a
// transport in order to send bytes.
type Sessionless interface {
	Connection
	SessionlessCommands

	// NewSession() does not go here, as the sessionless interface fixes the
	// session layer, whereas inside a session, this must be manipulated.
	// Creating a session also requires access to a transport, which we
	// deliberately abstract away here.

	// there is no Close() here, as this represents things that can be done over
	// the session rather than the underlying transport, which is kept separate.
	// A session-less connection has no state on the remote console or the
	// managed system, so can simply be abandoned rather than closed.
}

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
