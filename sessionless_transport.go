package bmc

import (
	"context"

	"github.com/gebn/bmc/internal/pkg/transport"
)

// SessionlessTransport represents a session-less IPMI v1.5 or v2.0 LAN
// connection, its underlying transport, and a means of creating a new session
// using that transport. The IPMI version is fixed at creation time by the
// session-less connection; to obtain the version, call Version(). This is the
// type returned by Dial().
type SessionlessTransport interface {

	// Transport is the underlying socket, used to send and receive arbitrary
	// bytes, and get the address of the BMC. The Close() method of this
	// interface closes the transport, not the sessionless-connection (which
	// does not require closing).
	transport.Transport

	// Sessionless is the IPMI connection to the BMC, allowing the user to send
	// things at a higher level of abstraction than the transport alone
	// provides.
	Sessionless

	// NewSession opens a new session to the BMC using the underlying wrapper
	// format. This is generic as is works with both IPMI v1.5 and v2.0; for
	// more control over establishment, use DialV*() to obtain a
	// V1SessionlessTransport or V2SessionlessTransport. NewSession uses the
	// sessionless methods for establishment.
	NewSession(ctx context.Context, opts *SessionOpts) (Session, error)
}

// V1SessionlessTransport is a session-less connection to a BMC using an IPMI
// v1.5 session wrapper, along with its underlying transport. A pointer to this
// type is returned by DialV1().
type V1SessionlessTransport struct {
	transport.Transport
	V1Sessionless
}

func (s *V1SessionlessTransport) Close() error {
	// we intercept this call purely to do the bookkeeping. Even if the
	// connection fails to close, we regard it as such as there is nothing else
	// we can do. Note, it is essential to realise that Close() has no meaning
	// at the level of an abstract "connection", nor in the case of the
	// session-less connection.  Close() only exist for a session-based
	// connection. We cannot have the asymmetry of Close() on a session-less
	// closing the transport, and Close() on a session leaving it alone.
	defer v1ConnectionsOpen.Dec()
	return s.Transport.Close()
}

// V2SessionlessTransport is a session-less connection to a BMC using an IPMI
// v2.0/RMCP+ session wrapper, along with its underlying transport. A pointer to
// this type is returned by DialV2().
type V2SessionlessTransport struct {
	transport.Transport
	*V2Sessionless
}

func (s *V2SessionlessTransport) Close() error {
	defer v2ConnectionsOpen.Dec()
	return s.Transport.Close()
}
