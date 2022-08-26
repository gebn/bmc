package ipmi

import (
	"fmt"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
)

// Get Channel Cipher Suites comamand is specified in 22.15 of IPMI v2.0, the
// command is used to look up what authentication, integrity and confidentiality
// algorithms are supported, The algorithms are used in combination as
// 'Cipher Suites'. This command only applies to implementations that support
// IPMI v2.0/RMCP+ sessions.

// CipherSuiteData represents Cipher Suites type and algorighm tag and its ID.
type CipherSuiteData uint8

// OEMIANAType represents IANA for the OEM or body that defined the Cipher Suite
// which includes 3 little endian bytes.
type OEMIANAType uint32

const (
	// StandardCipherSuite or OEMCipherSuite is a start of record byte, which is
	// specified in Table 22-19, IPMI2.0.
	StandardCipherSuite CipherSuiteData = 0xc0
	OEMCipherSuite      CipherSuiteData = 0xc1

	// ListAlgorithmByCipherSuite is a bit to list algorighms by Cipher Suite
	// instead of list all supported algorithms (Table 22-18, IPMI2.0).
	ListAlgorithmByCipherSuite CipherSuiteData = 0x80

	// AuthenticationAlgorithmTag a tag that indicate last 5 bits is
	// authentication algorighm number.
	AuthenticationAlgorithmTag CipherSuiteData = 0x00
	// IntegrityAlgorithmTag is a tag that indicate last 5 bits is integrity
	// algorighm number.
	IntegrityAlgorithmTag CipherSuiteData = 0x01
	// ConfidentialityAlgorithmTag is a tag that indicate last 5 bits is
	// confidentiality algorithm number.
	ConfidentialityAlgorithmTag CipherSuiteData = 0x02

	// CipherSuiteID_* is the IDs for Cipher Suite, specified in Table 22-20 of
	// IPMI2.0, 17 and 3 are most widely supported so far, they will be used for
	// detecting the best Cipher Suite.
	CipherSuiteID_3  CipherSuiteData = 3
	CipherSuiteID_17 CipherSuiteData = 17
)

// GetChannelCipherSuitesReq defines a Get Channel Cipher Suites request. Its format
// is specified in Table 22-18 of IPMV v2.0.
type GetChannelCipherSuitesReq struct {
	layers.BaseLayer

	// Channel number of current request, bits[7:4] are reserved, and bits[3:0]
	// are for channel number, and 0h-Bh, Fh are channel numbers, Eh retrieve
	// information for channel this request was issued on.
	Channel Channel

	// The Payload Type number is used to look up the Security Algorithm support
	// when establishing a separate session for a given payload type. Typically
	// the number is 00h (IPMI).
	PayloadType PayloadType

	// List index (00h-3Fh). 0h selects the first set of 16, 1h selects the next
	// set of 16, and so on.
	// When use 00h to get first set of algorithm numbers. The BMC returns 16
	// bytes at a time per index, starting from index 00h, until the list
	// data is exhausted, at which point it will 0 bytes or <16 bytes of list
	// data.
	ListIndex CipherSuiteData
}

func (*GetChannelCipherSuitesReq) LayerType() gopacket.LayerType {
	return LayerTypeGetChannelCipherSuitesReq
}

func (g *GetChannelCipherSuitesReq) SerializeTo(b gopacket.SerializeBuffer, _ gopacket.SerializeOptions) error {
	bytes, err := b.PrependBytes(3)
	if err != nil {
		return err
	}
	bytes[0] = uint8(g.Channel)
	bytes[1] = uint8(g.PayloadType)

	// Always list algorighms by Cipher Suite instead of list all supported
	// algorithms.
	bytes[2] = uint8(g.ListIndex | ListAlgorithmByCipherSuite)
	return nil
}

// GetChannelCipherSuitesRsp represents the response to a Get Channel Cipher Suites
// request.
type GetChannelCipherSuitesRsp struct {
	layers.BaseLayer

	// Channel number that the Authentication Algorithms are being returned
	// for. If the channel number in the request was set to Eh, this will return
	// the channel number for the channel that the request was received on.
	Channel Channel

	// ID is the Cipher Suite ID, a numeric way of identifying the Cipher Suite
	// on the platform. It’s used in commands and configuration parameters that
	// enable and disable Cipher Suites (Table 22-20, IPMI v2.0).
	ID CipherSuiteData

	// Type indicate the start of record, Standard or OEM Cipher Suite.
	Type CipherSuiteData

	// OEMIANA is Least significant byte first. 3-byte IANA for the OEM or body
	// that defined the Cipher Suite.
	OEMIANA OEMIANAType

	// ListDataExhausted indicate the list data is exhausted.
	ListDataExhausted bool

	// AuthenticationAlgorithms is a list to align with Session options, a
	// Cipher Suite is only allowed to utilize one authentication algorithm.
	AuthenticationAlgorithms []AuthenticationAlgorithm

	// IntegrityAlgorithms defines all supported interity algorithms.
	IntegrityAlgorithms []IntegrityAlgorithm

	// ConfidentialityAlgorithms defines all supported confidentiality algorithms.
	ConfidentialityAlgorithms []ConfidentialityAlgorithm
}

func (*GetChannelCipherSuitesRsp) LayerType() gopacket.LayerType {
	return LayerTypeGetChannelCipherSuitesRsp
}

func (g *GetChannelCipherSuitesRsp) CanDecode() gopacket.LayerClass {
	return g.LayerType()
}

func (*GetChannelCipherSuitesRsp) NextLayerType() gopacket.LayerType {
	return gopacket.LayerTypePayload
}

func (g *GetChannelCipherSuitesRsp) DecodeFromBytes(data []byte, df gopacket.DecodeFeedback) error {
	length := len(data)
	if length < 6 {
		df.SetTruncated()
		return fmt.Errorf("response must be at least 6 bytes for the cipher suite, got %v",
			length)
	}

	// Check the Cipher Suite type first, they have different data bytes.
	g.Type = CipherSuiteData(data[1])
	if g.Type != StandardCipherSuite && g.Type != OEMCipherSuite {
		return fmt.Errorf("unexpected cipher suite type, got %v", g.Type)
	}

	g.Channel = Channel(data[0])
	g.ID = CipherSuiteData(data[2])
	offset := 3

	// OEM Chipher Suite need 3 more bytes for OEM IANA.
	if g.Type == OEMCipherSuite {
		g.OEMIANA = OEMIANAType(uint32(data[3]) + uint32(data[4])<<8 + uint32(data[5])<<16)
		offset += 3
	}

	// The size of these algorithms is variable, detect them with its tag.
	for i := offset; i < length; i++ {
		switch CipherSuiteData(data[i] >> 6) {
		case AuthenticationAlgorithmTag:
			g.AuthenticationAlgorithms = append(g.AuthenticationAlgorithms, AuthenticationAlgorithm(data[i]&0x1f))
		case IntegrityAlgorithmTag:
			g.IntegrityAlgorithms = append(g.IntegrityAlgorithms, IntegrityAlgorithm(data[i]&0x1f))
		case ConfidentialityAlgorithmTag:
			g.ConfidentialityAlgorithms = append(g.ConfidentialityAlgorithms, ConfidentialityAlgorithm(data[i]&0x1f))
		}
	}

	// The number 17 is 16 data bytes plus a index.
	// The BMC returns sixteen (16) bytes at a time per index, starting from index
	// 00h, until the list data is exhausted, at which point it will 0 bytes or <16
	// bytes of list data.
	if length == 17 {
		g.ListDataExhausted = false
	} else {
		g.ListDataExhausted = true
	}

	g.BaseLayer.Contents = data[:length]
	g.BaseLayer.Payload = data[length:]
	return nil
}

type GetChannelCipherCmd struct {
	Req GetChannelCipherSuitesReq
	Rsp GetChannelCipherSuitesRsp
}

// Name returns "Get Channel Cipher Suite".
func (*GetChannelCipherCmd) Name() string {
	return "Get Channel Cipher Suite"
}

// Operation returns &OperationGetChannelCipherSuitesReq.
func (*GetChannelCipherCmd) Operation() *Operation {
	return &OperationGetChannelCipherSuitesReq
}

func (c *GetChannelCipherCmd) Request() gopacket.SerializableLayer {
	return &c.Req
}

func (c *GetChannelCipherCmd) Response() gopacket.DecodingLayer {
	return &c.Rsp
}
