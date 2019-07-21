package ipmi

import (
	"encoding/binary"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
)

// CloseSessionReq implements the Close Session command, specified in section
// 18.17 of v1.5 and 22.19 of v2.0. It immediately terminates the specified
// session. The sending user must be operating with administrator privileges to
// close any session other than the one this request is sent over.
type CloseSessionReq struct {
	layers.BaseLayer

	// ID is the ID of the session to close. In the case of IPMI v2.0, this is
	// the manage system's session ID - not the remote console's. If this is
	// null, the session handle is additionally sent. It must be non-null when
	// using IPMI v1.5 to ensure the Handle field is not sent.
	ID uint32

	// Handle is the session handle to close, only used if the ID is null. A
	// handle appears to be another kind of session ID, but scoped to the
	// current channel. Each new session receives an incremented handle number.
	// The handle can be obtained via Get Session Info if necessary. 0x00 is
	// reserved, however will be encoded and decoded. This field is only
	// specified in IPMI v2.0.
	Handle uint8
}

func (*CloseSessionReq) LayerType() gopacket.LayerType {
	return LayerTypeCloseSessionReq
}

func (c *CloseSessionReq) SerializeTo(b gopacket.SerializeBuffer, _ gopacket.SerializeOptions) error {
	length := 4
	if c.ID == 0 {
		length++
	}
	bytes, err := b.PrependBytes(length)
	if err != nil {
		return err
	}
	binary.LittleEndian.PutUint32(bytes[0:4], c.ID)
	if c.ID == 0 {
		bytes[4] = c.Handle
	}
	return nil
}
