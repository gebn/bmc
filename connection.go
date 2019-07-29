package bmc

import (
	"context"

	"github.com/gebn/bmc/pkg/ipmi"
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

	// SendCommand sends a command to the BMC, blocking until it receives a
	// response. This method will retry with the configured per-request timeout
	// until a valid response is received, or the context expires (whichever
	// happens first). A non-zero completion code is deemed to be a valid
	// response. If the final request fails with a transport error (including
	// timeout), a serialise/decode error occurs, or the message layer is
	// missing, the returned error will be non-nil, and the completion code must
	// be ignored.
	//
	// This method uses the response layer (if any) included in the command
	// interface for decoding the response. The caller should first check the
	// error, then the completion code, then assuming both indicate no error,
	// read the response layer if required. The ValidateResponse() function can
	// be used for the sake of brevity.
	//
	// This method must not allocate any memory, so is ideal in situations where
	// you intend to send the same command repeatedly, e.g. a Prometheus
	// exporter. If you don't need this performance, for the sake of one more
	// allocation per command, it is recommended to use the higher-level API,
	// e.g. GetDeviceID(), which wraps this.
	SendCommand(ctx context.Context, cmd ipmi.Command) (ipmi.CompletionCode, error)

	// Version returns the underlying IPMI version of the connection, either
	// "1.5" or "2.0". Note that even session-less connections use a session
	// wrapper, which has either the v1.5 or v2.0 format. This is provided for
	// informational and debugging purposes - branching based on this value is a
	// code smell.
	Version() string
}
