package dcmi

import (
	"encoding/binary"
	"fmt"
	"time"

	"github.com/gebn/bmc/pkg/ipmi"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
)

// GetDCMICapabilitiesInfoReq represents the Get DCMI Capabilities Info request,
// specified in 6.1. This is a session-less command, but the spec explicitly
// says it can be executed inside a session at any privilege level.
//
// This command is also recommended (not by the spec) for DCMI discovery.
// Although a DCMI flag is present in the RMCP Presence Pong message since DCMI
// 1.5 (2011), in practice, this alone is useless as large parts of DCMI (e.g.
// power management) are optional. Given sending this command to discover what
// is actually supported would be necessary anyway in most cases, it is advised
// to forget RMCP Presence Ping and just send this.
type GetDCMICapabilitiesInfoReq struct {
	layers.BaseLayer

	// can't actually parse what this is from the spec; leaving as the default
	// seems fine
	ParameterSelector uint8
}

func (*GetDCMICapabilitiesInfoReq) LayerType() gopacket.LayerType {
	return layerTypeGetDCMICapabilitiesInfoReq
}

func (g *GetDCMICapabilitiesInfoReq) SerializeTo(b gopacket.SerializeBuffer, _ gopacket.SerializeOptions) error {
	bytes, err := b.PrependBytes(1)
	if err != nil {
		return err
	}
	bytes[0] = g.ParameterSelector
	return nil
}

// GetDCMICapabilitiesInfoRsp represents the response to a Get DCMI Capabilities
// Info request, specified in 6.1.
type GetDCMICapabilitiesInfoRsp struct {
	layers.BaseLayer

	// Supported DCMI Capabilities

	// MajorVersion gives the major version of DCMI spec conformance. This will
	// be 0x01 in all known implementations.
	MajorVersion uint8

	// MinorVersion gives the minor version of DCMI spec conformance. This will
	// be either 0x01 or 0x05 in all known implementations.
	MinorVersion uint8

	// Revision is the parameter revision. This is always 0x02.
	Revision uint8

	// PowerManagement indicates whether the server supports the power
	// management platform capability.
	PowerManagement bool

	// OOBSecondaryLANChannelAvailable indicates whether an Out-of-Band
	// Secondary (second) LAN Channel is available.
	OOBSecondaryLANChannelAvailable bool

	// SerialTMODEAvailable indicates whether TMODE is available on the serial
	// port to the management controller.
	SerialTMODEAvailable bool

	// IBSystemInterfaceChannelAvailable indicates whether an in-band system
	// interface channel is available.
	IBSystemInterfaceChannelAvailable bool

	// Mandatory Platform Attributes

	// SELAutoRollover indicates whether SEL automatic rollover is enabled, also
	// known as SEL overwrite.
	SELAutoRollover bool

	// SELFlushOnRollover indicates whether, on rollover, the entire SEL is
	// flushed. This should be ignored in SELAutoRollover is false.
	SELFlushOnRollover bool

	// SELRecordLevelFlushOnRollover indicates whether individual SEL records
	// are flished upon rollover, as opposed to the entire SEL. This should be
	// ignored in SELAutoRollover is false.
	SELRecordLevelFlushOnRollover bool

	// SELMaxEntries contains the maximum number of SEL entries supported by the
	// system. The spec requires this value to be between 64 and 4096 inclusive.
	// It is a 12-bit uint on the wire, however as DCMI (like IPMI) uses
	// little-endian, the max representable value is 0xff0f, or 65295, rather
	// than 0x0fff, which would be 4095.
	SELMaxEntries uint16

	// TemperatureSamplingFrequency is the interval between successive
	// temperature samples. This will be a whole number of seconds between 0 and
	// 255.
	TemperatureSamplingFrequency time.Duration

	// Optional Platform Attributes

	// PowerManagementSlaveAddress gives the 7-bit I2C slave address of the
	// power management device on the IPMB.
	PowerManagementSlaveAddress ipmi.SlaveAddress

	// PowerManagementChannel is the channel number of the power management
	// controller.
	PowerManagementChannel ipmi.Channel

	// PowerManagementRevision is the power management controller device
	// revision.
	PowerManagementRevision uint8

	// Manageability Access Attributes

	// PrimaryLANOOBChannel is the primary LAN OOB channel number. This will
	// only be a valid channel number for systems supporting RMCP+. 0xff
	// indicates not supported; the Valid() method can be used to test for this.
	PrimaryLANOOBChannel ipmi.Channel

	// SecondaryLANOOBChannel is the secondary LAN OOB channel number. This may
	// be invalid on all systems, as a secondary channel is optional. 0xff
	// indicates not supported; the Valid() method can be used to test for this.
	SecondaryLANOOBChannel ipmi.Channel

	// SerialOOBChannel is the serial OOB TMODE channel number. This is
	// optional, and so may give an invalid channel number on all systems. 0xff
	// indicates not supported; use Valid() to test.
	SerialOOBChannel ipmi.Channel

	// Enhanced System Power Statistics (optional)

	// PowerRollingAvgTimePeriods returns the supported rolling average time
	// periods that can be requested with the Get Power Reading command. This
	// will be a whole number of seconds, minutes, hours or days, from 0 seconds
	// to 63 days. A value of 0 means the system supports obtaining the current
	// reading. This slice contains time periods in the order provided by the
	// BMC.
	PowerRollingAvgTimePeriods []time.Duration
}

func (*GetDCMICapabilitiesInfoRsp) LayerType() gopacket.LayerType {
	return layerTypeGetDCMICapabilitiesInfoRsp
}

func (g *GetDCMICapabilitiesInfoRsp) CanDecode() gopacket.LayerClass {
	return g.LayerType()
}

func (*GetDCMICapabilitiesInfoRsp) NextLayerType() gopacket.LayerType {
	return gopacket.LayerTypePayload
}

func (g *GetDCMICapabilitiesInfoRsp) DecodeFromBytes(data []byte, df gopacket.DecodeFeedback) error {
	// field lengths: 3 + 3 + 5 + 2 + 3 + (1+)
	if len(data) < 16 {
		df.SetTruncated()
		return fmt.Errorf("invalid command response, got length %v, need at least 16", len(data))
	}

	g.MajorVersion = uint8(data[0])
	g.MinorVersion = uint8(data[1])
	g.Revision = uint8(data[2])

	g.PowerManagement = data[4]&1 != 0
	g.OOBSecondaryLANChannelAvailable = data[5]&(1<<2) != 0
	g.SerialTMODEAvailable = data[5]&(1<<1) != 0
	g.IBSystemInterfaceChannelAvailable = data[5]&1 != 0

	g.SELAutoRollover = data[6]&(1<<7) != 0
	g.SELFlushOnRollover = data[6]&(1<<6) != 0
	g.SELRecordLevelFlushOnRollover = data[6]&(1<<5) != 0
	g.SELMaxEntries = binary.LittleEndian.Uint16([]byte{data[6] & 0xf, data[7]})
	g.TemperatureSamplingFrequency = time.Second * time.Duration(data[10])

	g.PowerManagementSlaveAddress = ipmi.SlaveAddress(data[11] >> 1)
	g.PowerManagementChannel = ipmi.Channel(data[12] >> 4)
	g.PowerManagementRevision = uint8(data[12] & 0xf)

	g.PrimaryLANOOBChannel = ipmi.Channel(data[13])
	g.SecondaryLANOOBChannel = ipmi.Channel(data[14])
	g.SerialOOBChannel = ipmi.Channel(data[15])

	if len(data) > 16 {
		// supports enhanced system power stats
		periods := int(data[16])
		if len(data) < 17+periods {
			df.SetTruncated()
			return fmt.Errorf("managed system indicated %v supported rolling "+
				"average time periods, but only room for %v in payload of "+
				"length %v", periods, len(data)-17, len(data))
		}
		g.PowerRollingAvgTimePeriods = make([]time.Duration, periods)
		for i := 0; i < periods; i++ {
			unit := uint8(data[17+i] >> 6)  // top 2 bits
			value := int(data[17+i] & 0x3f) // bottom 6 bits
			seconds := value * secondsMultiplier(unit)
			g.PowerRollingAvgTimePeriods[i] = time.Second * time.Duration(seconds)
		}
		g.BaseLayer.Contents = data[:17+periods]
		g.BaseLayer.Payload = data[17+periods:]
	} else {
		g.PowerRollingAvgTimePeriods = g.PowerRollingAvgTimePeriods[:0]
		g.BaseLayer.Contents = data
		g.BaseLayer.Payload = nil
	}
	return nil
}

// secondsMultiplier parses the 2-bit time duration unit, returning a number to
// multiply the time duration with in order to provide the duration in seconds.
// Only the two LSBs of the input are interpreted.
func secondsMultiplier(unit uint8) int {
	switch unit {
	case 0:
		// 0b00: seconds
		return 1
	case 1:
		// 0b01: minutes
		return 60
	case 2:
		// 0b10: hours
		return 60 * 60
	default: // inc. 3
		// 0b11: days
		return 60 * 60 * 24
	}
}

type GetDCMICapabilitiesInfoCmd struct {
	Req GetDCMICapabilitiesInfoReq
	Rsp GetDCMICapabilitiesInfoRsp
}

// Name returns "Get DCMI Capabilities Info".
func (*GetDCMICapabilitiesInfoCmd) Name() string {
	return "Get DCMI Capabilities Info"
}

func (*GetDCMICapabilitiesInfoCmd) Operation() *ipmi.Operation {
	return &operationGetDCMICapabilitiesInfoReq
}

func (c *GetDCMICapabilitiesInfoCmd) Request() gopacket.SerializableLayer {
	return &c.Req
}

func (c *GetDCMICapabilitiesInfoCmd) Response() gopacket.DecodingLayer {
	return &c.Rsp
}
