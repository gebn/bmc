// Package transport provides a context wrapper and error handling around a UDP
// connection.
package transport

import (
	"context"
	"fmt"
	"net"
)

// Transport encapsulates a UDP connection and a receive buffer. Unless
// specified otherwise, access must be serialised.
type Transport struct {
	fd *net.UDPConn

	// recvBuf is used for reading bytes off the wire. This means we do not
	// allocate any memory in the hot path, but causes a race condition if the
	// transport is used concurrently.
	recvBuf [512]byte
}

// New establishes a connection to a UDP endpoint. It is recommended to defer a
// call to Close() immediately after the error check.
func New(addr string) (*Transport, error) {
	raddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		return nil, err
	}

	c, err := net.DialUDP("udp", nil, raddr)
	if err != nil {
		return nil, err
	}
	return &Transport{
		fd: c,
	}, nil
}

// Address returns the remote IP:port of the endpoint.
func (t *Transport) Address() net.Addr {
	return t.fd.RemoteAddr()
}

// Send sends the supplied data to the remote host, blocking until it receives a
// reply packet, which is then returned. An error is returned if a transport
// error occurs or the context expires.
func (t *Transport) Send(ctx context.Context, b []byte) ([]byte, error) {
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
func (t *Transport) Close() error {
	return t.fd.Close()
}
