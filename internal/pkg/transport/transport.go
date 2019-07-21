// Package transport provides a context wrapper and error handling around a UDP
// connection.
package transport

import (
	"context"
	"fmt"
	"net"
)

type transport struct {
	fd *net.UDPConn

	// recvBuf is used for reading bytes off the wire. This means we do not
	// allocate any memory in the hot path, but causes a race condition if the
	// transport is used concurrently.
	recvBuf [512]byte
}

// Address returns the remote IP:port of the endpoint.
func (t *transport) Address() net.Addr {
	return t.fd.RemoteAddr()
}

// Send sends the supplied data to the remote host, blocking until it receives a
// reply packet, which is then returned. An error is returned if a transport
// error occurs or the context expires.
func (t *transport) Send(ctx context.Context, b []byte) ([]byte, error) {
	if deadline, ok := ctx.Deadline(); ok {
		if err := t.fd.SetWriteDeadline(deadline); err != nil {
			return nil, err
		}
	}
	n, err := t.fd.Write(b)
	if err != nil {
		return nil, err
	}
	if n != len(b) {
		return nil, fmt.Errorf("wrote incomplete message (%v/%v bytes)", n,
			len(b))
	}

	// read
	if deadline, ok := ctx.Deadline(); ok {
		if err := t.fd.SetReadDeadline(deadline); err != nil {
			return nil, err
		}
	}
	n, _, err = t.fd.ReadFromUDP(t.recvBuf[:])
	if err != nil {
		return nil, err
	}

	return t.recvBuf[:n], nil
}

// Close cleanly shuts down the transport, rendering it unusable.
func (t *transport) Close() error {
	return t.fd.Close()
}

// Transport defines an interface capable of sending and receiving data to and
// from a device. It logically represents a UDP socket and receive buffer.
// Unless specified otherwise, access must be serialised.
type Transport interface {

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

// New establishes a connection to a UDP endpoint. It is recommended to defer a
// call to Close() immediately after the error check.
func New(addr string) (Transport, error) {
	raddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		return nil, err
	}

	c, err := net.DialUDP("udp", nil, raddr)
	if err != nil {
		return nil, err
	}
	return &transport{
		fd: c,
	}, nil
}
