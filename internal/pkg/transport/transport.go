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

func New(addr string) (*transport, error) {
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

func (t *transport) Address() net.Addr {
	return t.fd.RemoteAddr()
}

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

func (t *transport) Close() error {
	return t.fd.Close()
}
