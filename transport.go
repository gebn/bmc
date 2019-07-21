package bmc

import (
	"context"
	"net"
	"strings"

	trans "github.com/gebn/bmc/internal/pkg/transport"
)

// transport defines an interface capable of sending and receiving data to and
// from a device. It logically represents a UDP socket, and is independent of
// IPMI. This is satisfied by *transport.Transport.
type transport interface {

	// Address returns the IP:port of the remote device. This will always have
	// the port, even if the address provided was missing it (we default to
	// 623).
	Address() net.Addr

	// Send encapsulates the provided data in a UDP packet and sends it to the
	// BMC's address. It then blocks until a packet is received, and returns the
	// data it contains. If the context expires before all of this is performed,
	// or there is a network error, the returned slice will be nil and the error
	// will be returned.
	Send(context.Context, []byte) ([]byte, error)

	// Close cleanly shuts down the underlying connection, returning any error
	// that occurs. It is envisaged that this call is deferred as soon as the
	// transport is successfully created.
	Close() error
}

func newTransport(addr string) (transport, error) {
	// default to port 623
	if !strings.Contains(addr, ":") || strings.HasSuffix(addr, "]") {
		addr = addr + ":623"
	}
	return trans.New(addr)
}
