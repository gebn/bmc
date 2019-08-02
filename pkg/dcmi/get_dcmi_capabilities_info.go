package dcmi

import (
	"encoding/binary"
	"fmt"
	"time"

	"github.com/gebn/bmc/pkg/ipmi"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
)

// CapabilitiesParameter is a kind of sub-command number within the Get DCMI
// Capabilities request. This seems to be an effort by the spec to keep packets
// as small as possible, but is a pain as we effectively have to send up to 5
// packets to discover all capabilities. To address this madness, this library
// does something equally mad and implements each parameter as its own command,
// as there is no way to know from a response packet alone which parameter was
// requested.
type CapabilitiesParameter uint8

// description returns the human-readable name of each parameter.
func (s CapabilitiesParameter) description() string {
	switch s {
	case 1:
		return "Supported DCMI Capabilities"
	case 2:
		return "Mandatory Platform Attributes"
	case 3:
		return "Optional Platform Attributes"
	case 4:
		return "Manageability Access Attributes"
	case 5:
		return "Enhanced System Power Statistics Attributes"
	default:
		return "Unknown"
	}
}

func (s CapabilitiesParameter) String() string {
	return fmt.Sprintf("%v(%v)", uint8(s), s.description())
}

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

	// Parameter specifies the type of attributes or capabilities desired. This
	// command returns one of several different possible pieces of information
	// depending on this value, e.g. supported DCMI capabilities, platform
	// attributes and access attributes. The response layer formats for the
	// different selector values are specified in Table 6-3.
	Parameter CapabilitiesParameter
}

func (*GetDCMICapabilitiesInfoReq) LayerType() gopacket.LayerType {
	return layerTypeGetDCMICapabilitiesInfoReq
}

func (g *GetDCMICapabilitiesInfoReq) SerializeTo(b gopacket.SerializeBuffer, _ gopacket.SerializeOptions) error {
	bytes, err := b.PrependBytes(1)
	if err != nil {
		return err
	}
	bytes[0] = uint8(g.Parameter)
	return nil
}

// getDCMICapabilitiesInfoRsp represents the header of the response to a Get
// DCMI Capabilities Info request, specified in 6.1. The rest of the header is
// dictated by the parameter specified in the request. Note this is not a layer.
type getDCMICapabilitiesInfoRspHeader struct {

	// MajorVersion gives the major version of DCMI spec conformance. This will
	// be 0x01 in all known implementations.
	MajorVersion uint8

	// MinorVersion gives the minor version of DCMI spec conformance. This will
	// be either 0x00, 0x01 or 0x05 in known implementations.
	MinorVersion uint8

	// Revision is the revision of the parameter data. This will be 0x01 for
	// DCMI v1.0, or 0x02 for DCMI v1.1 and v1.5. Note this does not correspond
	// to the revision of the overall spec implemented.
	Revision uint8
}

// Decode decodes the header and returns the remaining bytes (which could be
// empty), or nil if an error occurs.
func (g *getDCMICapabilitiesInfoRspHeader) Decode(data []byte, df gopacket.DecodeFeedback) ([]byte, error) {
	if len(data) < 3 {
		df.SetTruncated()
		return nil, fmt.Errorf("invalid response header, got length %v, need 3", len(data))
	}

	g.MajorVersion = uint8(data[0])
	g.MinorVersion = uint8(data[1])
	g.Revision = uint8(data[2])
	return data[3:], nil
}

// GetDCMICapabilitiesInfoSupportedCapabilitiesRsp is returned when a Get DCMI
// Capabilities Info request is sent with parameter 1. The returned capabilities
// detail conformance to the spec for platform and manageability access.
type GetDCMICapabilitiesInfoSupportedCapabilitiesRsp struct {
	layers.BaseLayer
	getDCMICapabilitiesInfoRspHeader

	// Mandatory Platform Capabilities (v1.0). These fields were likely
	// removed as they allow implementations to treat mandatory commands as
	// optional.

	// TemperatureMonitor indicates whether the system supports the temperature
	// monitoring commands in Table 3-1 of v1.0. This is a v1.0-only field, but
	// will be forced to true for v1.1 and v1.5 systems for backwards
	// compatibility.
	TemperatureMonitor bool

	// ChassisPower indicates whether the system supports the chassis power
	// commands in Table 3-1 of v1.0. This is a v1.0-only field, but will be
	// forced to true for v1.1 and v1.5 systems for backwards compatibility.
	ChassisPower bool

	// SELLogging indicates whether the system supports the event logging
	// commands in Table 3-1 of v1.0. This is a v1.0-only field, but will be
	// forced to true for v1.1 and v1.5 systems for backwards compatibility.
	SELLogging bool

	// Identification indicates whether the system supports the identification
	// commands in Table 3-1 of v1.0. This is a v1.0-only field, but will be
	// forced to true for v1.1 and v1.5 systems for backwards compatibility.
	Identification bool

	// Optional Platform Capabilities

	// PowerManagement indicates whether the server supports the power
	// management platform capability.
	PowerManagement bool

	// Manageability Access Capabilities

	// VLANCapable indicates whether the system supports VLANs. This is a
	// v1.0-only field, and will be forced to true for v1.1 and v1.5 systems for
	// backwards compatibility.
	VLANCapable bool

	// SOLSupportes indicates whether the system supports serial-over-LAN. This
	// is a v1.0-only field, and will be forced to true for v1.1 and v1.5
	// systems for backwards compatibility.
	SOLSupported bool

	// OOBPrimaryLANChannelAvailable indicates whether an Out-of-Band Primary
	// LAN Channel is available. This is a v1.0-only field, and will be forced
	// to true for v1.1 and v1.5 systems for backwards compatibility.
	OOBPrimaryLANChannelAvailable bool

	// OOBSecondaryLANChannelAvailable indicates whether an Out-of-Band
	// Secondary (second) LAN Channel is available.
	OOBSecondaryLANChannelAvailable bool

	// SerialTMODEAvailable indicates whether TMODE is available on the serial
	// port to the management controller.
	SerialTMODEAvailable bool

	// IBKCSChannelAvailable indicates whether an in-band KCS channel is
	// available. This is a v1.0 field, forced to true for v1.1 and v1.5 for
	// backwards compatibility.
	IBKCSChannelAvailable bool

	// IBSystemInterfaceChannelAvailable indicates whether an in-band system
	// interface channel is available. This will always be false for v1.0, which
	// uses this bit for the KCS channel instead.
	IBSystemInterfaceChannelAvailable bool
}

func (*GetDCMICapabilitiesInfoSupportedCapabilitiesRsp) LayerType() gopacket.LayerType {
	return layerTypeGetDCMICapabilitiesInfoSupportedCapabilitiesRsp
}

func (g *GetDCMICapabilitiesInfoSupportedCapabilitiesRsp) CanDecode() gopacket.LayerClass {
	return g.LayerType()
}

func (*GetDCMICapabilitiesInfoSupportedCapabilitiesRsp) NextLayerType() gopacket.LayerType {
	return gopacket.LayerTypePayload
}

func (g *GetDCMICapabilitiesInfoSupportedCapabilitiesRsp) DecodeFromBytes(data []byte, df gopacket.DecodeFeedback) error {
	body, err := g.Decode(data, df)
	if err != nil {
		return err
	}

	minBodyLength := 3
	if len(body) < minBodyLength {
		df.SetTruncated()
		return fmt.Errorf("invalid capabilities response: need at least %v bytes, got %v",
			minBodyLength, len(body))
	}

	if g.MajorVersion == 1 && g.MinorVersion == 0 {
		g.TemperatureMonitor = body[0]&(1<<3) != 0
		g.ChassisPower = body[0]&(1<<2) != 0
		g.SELLogging = body[0]&(1<<1) != 0
		g.Identification = body[0]&1 != 0
	} else {
		g.TemperatureMonitor = true
		g.ChassisPower = true
		g.SELLogging = true
		g.Identification = true
	}

	g.PowerManagement = body[1]&1 != 0

	if g.MajorVersion == 1 && g.MinorVersion == 0 {
		g.VLANCapable = body[2]&(1<<5) != 0
		g.SOLSupported = body[2]&(1<<4) != 0
		g.OOBPrimaryLANChannelAvailable = body[2]&(1<<3) != 0
	} else {
		g.VLANCapable = true
		g.SOLSupported = true
		g.OOBPrimaryLANChannelAvailable = true
	}
	g.OOBSecondaryLANChannelAvailable = body[2]&(1<<2) != 0
	g.SerialTMODEAvailable = body[2]&(1<<1) != 0
	if g.MajorVersion == 1 && g.MinorVersion == 0 {
		g.IBKCSChannelAvailable = body[2]&1 != 0
		g.IBSystemInterfaceChannelAvailable = false
	} else {
		g.IBKCSChannelAvailable = true
		g.IBSystemInterfaceChannelAvailable = body[2]&1 != 0
	}

	g.Contents = data[:len(data)-len(body)+minBodyLength]
	g.Payload = body[minBodyLength:]

	return nil
}

// GetDCMICapabilitiesInfoMandatoryPlatformAttrsRsp is returned when a Get DCMI
// Capabilities Info request is sent with parameter 2.
type GetDCMICapabilitiesInfoMandatoryPlatformAttrsRsp struct {
	layers.BaseLayer
	getDCMICapabilitiesInfoRspHeader

	// SELAutoRollover indicates whether SEL automatic rollover is enabled, also
	// known as SEL overwrite.
	SELAutoRollover bool

	// SELFlushOnRollover indicates whether, on rollover, the entire SEL is
	// flushed. This should be ignored if SELAutoRollover is false, or if the
	// response is v1.0, where this is unspecified.
	SELFlushOnRollover bool

	// SELRecordLevelFlushOnRollover indicates whether individual SEL records
	// are flished upon rollover, as opposed to the entire SEL. This should be
	// ignored in SELAutoRollover is false, or if the response is v1.0, where
	// this is unspecified.
	SELRecordLevelFlushOnRollover bool

	// SELMaxEntries contains the maximum number of SEL entries supported by the
	// system. v1.0 gives no lower bound, v1.1 and v1.5 say 64. The max is 4096
	// for all implementations. It is a 12-bit uint on the wire, however as DCMI
	// (like IPMI) uses little-endian, the max representable value is 0xff0f, or
	// 65295, rather than 0x0fff, which would be 4095.
	SELMaxEntries uint16

	// AssetTagSupport indicates whether the system supports the asset tag
	// functions for a v1.0 system. This is mandatory in one place, recommended
	// in another, and removed from v1.1. It is set to true from v1.1 for
	// backwards compatibility.
	AssetTagSupport bool

	// DHCPHostNameSupport indicates whether the system publishes itself as a
	// DCMI controller when using DISCOVER mechanisms by setting option 12 (Host
	// Name) to equal "DCMI". This is recommended in v1.0, and mandatory from
	// v1.1, where it is forced to true.
	DHCPHostNameSupport bool

	// GUIDSupport indicates whether the system supports the system GUID
	// identification function. This is mandatory in the v1.0 spec, and the
	// field is removed from v1.1, where it is forced to true.
	GUIDSupport bool

	// BaseboardTemperature indicates whether at least one baseboard temperature
	// sensor is present. This is mandatory in v1.0, and the field is removed
	// from v1.1, where it is forced to true.
	BaseboardTemperature bool

	// ProcessorsTemperature indicates whether at least one processor
	// temperature sensor is present. This is mandatory in v1.0, and the field
	// is removed from v1.1, where it is forced to true.
	ProcessorsTemperature bool

	// InletTemperature indicates whether at least one baseboard temperature
	// sensor is present. This is mandatory in v1.0, and the field is removed
	// from v1.1, where it is forced to true.
	InletTemperature bool

	// TemperatureSamplingFrequency is the interval between successive
	// temperature samples. This will be a whole number of seconds between 0 and
	// 255. It will always be 0 for v1.0, where the field is not present.
	TemperatureSamplingFrequency time.Duration
}

func (*GetDCMICapabilitiesInfoMandatoryPlatformAttrsRsp) LayerType() gopacket.LayerType {
	return layerTypeGetDCMICapabilitiesInfoMandatoryPlatformAttrsRsp
}

func (g *GetDCMICapabilitiesInfoMandatoryPlatformAttrsRsp) CanDecode() gopacket.LayerClass {
	return g.LayerType()
}

func (*GetDCMICapabilitiesInfoMandatoryPlatformAttrsRsp) NextLayerType() gopacket.LayerType {
	return gopacket.LayerTypePayload
}

func (g *GetDCMICapabilitiesInfoMandatoryPlatformAttrsRsp) DecodeFromBytes(data []byte, df gopacket.DecodeFeedback) error {
	body, err := g.Decode(data, df)
	if err != nil {
		return err
	}

	minBodyLength := 4
	if len(body) < minBodyLength {
		df.SetTruncated()
		return fmt.Errorf("invalid capabilities response: need at least %v bytes, got %v",
			minBodyLength, len(body))
	}

	// This should be 5 for v1.1 and v1.5, however there are implementations
	// (SuperMicro) known to say v1.1 in the header but give a v1.0 format body,
	// so we treat all 4 byte bodies as v1.0, and all other lengths as what the
	// BMC told us. Note this means trailing bytes are *not* handled correctly
	// for v1.0 (they are assumed to be v1.1/v1.5 responses).
	isVersion10 := len(body) == 4 || g.MajorVersion == 1 && g.MinorVersion == 0

	g.SELAutoRollover = body[0]&(1<<7) != 0
	if isVersion10 {
		g.SELFlushOnRollover = false
		g.SELRecordLevelFlushOnRollover = false
	} else {
		g.SELFlushOnRollover = body[0]&(1<<6) != 0
		g.SELRecordLevelFlushOnRollover = body[0]&(1<<5) != 0
	}
	g.SELMaxEntries = binary.LittleEndian.Uint16([]byte{body[0] & 0xf, body[1]})

	if isVersion10 {
		g.AssetTagSupport = body[2]&(1<<2) != 0
		g.DHCPHostNameSupport = body[2]&(1<<1) != 0
		g.GUIDSupport = body[2]&1 != 0

		g.BaseboardTemperature = body[3]&(1<<2) != 0
		g.ProcessorsTemperature = body[3]&(1<<1) != 0
		g.InletTemperature = body[3]&1 != 0
	} else {
		g.AssetTagSupport = true
		g.DHCPHostNameSupport = true
		g.GUIDSupport = true

		g.BaseboardTemperature = true
		g.ProcessorsTemperature = true
		g.InletTemperature = true
	}

	if isVersion10 {
		g.TemperatureSamplingFrequency = 0
	} else {
		g.TemperatureSamplingFrequency = time.Second * time.Duration(body[4])
	}

	bodySectionLength := 5
	if isVersion10 {
		bodySectionLength = 4
	}
	g.Contents = data[:len(data)-len(body)+bodySectionLength]
	g.Payload = body[bodySectionLength:]

	return nil
}

// GetDCMICapabilitiesInfoOptionalPlatformAttrsRsp is returned when a Get DCMI
// Capabilities Info request is sent with parameter 3.
type GetDCMICapabilitiesInfoOptionalPlatformAttrsRsp struct {
	layers.BaseLayer
	getDCMICapabilitiesInfoRspHeader

	// PowerManagementSlaveAddress gives the 7-bit I2C slave address of the
	// power management device on the IPMB.
	PowerManagementSlaveAddress ipmi.SlaveAddress

	// PowerManagementChannel is the channel number of the power management
	// controller.
	PowerManagementChannel ipmi.Channel

	// PowerManagementRevision is the power management controller device
	// revision.
	PowerManagementRevision uint8
}

func (*GetDCMICapabilitiesInfoOptionalPlatformAttrsRsp) LayerType() gopacket.LayerType {
	return layerTypeGetDCMICapabilitiesInfoOptionalPlatformAttrsRsp
}

func (g *GetDCMICapabilitiesInfoOptionalPlatformAttrsRsp) CanDecode() gopacket.LayerClass {
	return g.LayerType()
}

func (*GetDCMICapabilitiesInfoOptionalPlatformAttrsRsp) NextLayerType() gopacket.LayerType {
	return gopacket.LayerTypePayload
}

func (g *GetDCMICapabilitiesInfoOptionalPlatformAttrsRsp) DecodeFromBytes(data []byte, df gopacket.DecodeFeedback) error {
	body, err := g.Decode(data, df)
	if err != nil {
		return err
	}

	minBodyLength := 2
	if len(body) < minBodyLength {
		df.SetTruncated()
		return fmt.Errorf("invalid capabilities response: need at least %v bytes, got %v",
			minBodyLength, len(body))
	}

	g.PowerManagementSlaveAddress = ipmi.SlaveAddress(body[0] >> 1)
	g.PowerManagementChannel = ipmi.Channel(body[1] >> 4)
	g.PowerManagementRevision = uint8(body[1] & 0xf)

	g.Contents = data[:len(data)-len(body)+minBodyLength]
	g.Payload = body[minBodyLength:]

	return nil
}

// GetDCMICapabilitiesInfoManageabilityAccessAttrsRsp is returned when a Get
// DCMI Capabilities Info request is sent with parameter 4.
type GetDCMICapabilitiesInfoManageabilityAccessAttrsRsp struct {
	layers.BaseLayer
	getDCMICapabilitiesInfoRspHeader

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
}

func (*GetDCMICapabilitiesInfoManageabilityAccessAttrsRsp) LayerType() gopacket.LayerType {
	return layerTypeGetDCMICapabilitiesInfoManageabilityAccessAttrsRsp
}

func (g *GetDCMICapabilitiesInfoManageabilityAccessAttrsRsp) CanDecode() gopacket.LayerClass {
	return g.LayerType()
}

func (*GetDCMICapabilitiesInfoManageabilityAccessAttrsRsp) NextLayerType() gopacket.LayerType {
	return gopacket.LayerTypePayload
}

func (g *GetDCMICapabilitiesInfoManageabilityAccessAttrsRsp) DecodeFromBytes(data []byte, df gopacket.DecodeFeedback) error {
	body, err := g.Decode(data, df)
	if err != nil {
		return err
	}

	minBodyLength := 3
	if len(body) < minBodyLength {
		df.SetTruncated()
		return fmt.Errorf("invalid capabilities response: need at least %v bytes, got %v",
			minBodyLength, len(body))
	}

	g.PrimaryLANOOBChannel = ipmi.Channel(body[0])
	g.SecondaryLANOOBChannel = ipmi.Channel(body[1])
	g.SerialOOBChannel = ipmi.Channel(body[2])

	g.Contents = data[:len(data)-len(body)+minBodyLength]
	g.Payload = body[minBodyLength:]

	return nil
}

// GetDCMICapabilitiesInfoEnhancedSystemPowerStatisticsAttrsRsp is returned when
// a Get DCMI Capabilities Info request is sent with parameter 5. This is not
// supported by v1.0.
type GetDCMICapabilitiesInfoEnhancedSystemPowerStatisticsAttrsRsp struct {
	layers.BaseLayer
	getDCMICapabilitiesInfoRspHeader

	// PowerRollingAvgTimePeriods returns the supported rolling average time
	// periods that can be requested with the Get Power Reading command. This
	// will be a whole number of seconds, minutes, hours or days, from 0 seconds
	// to 63 days. A value of 0 means the system supports obtaining the current
	// reading. This slice contains time periods in the order provided by the
	// BMC.
	PowerRollingAvgTimePeriods []time.Duration
}

func (*GetDCMICapabilitiesInfoEnhancedSystemPowerStatisticsAttrsRsp) LayerType() gopacket.LayerType {
	return layerTypeGetDCMICapabilitiesInfoEnhancedSystemPowerStatisticsAttrsRsp
}

func (g *GetDCMICapabilitiesInfoEnhancedSystemPowerStatisticsAttrsRsp) CanDecode() gopacket.LayerClass {
	return g.LayerType()
}

func (*GetDCMICapabilitiesInfoEnhancedSystemPowerStatisticsAttrsRsp) NextLayerType() gopacket.LayerType {
	return gopacket.LayerTypePayload
}

func (g *GetDCMICapabilitiesInfoEnhancedSystemPowerStatisticsAttrsRsp) DecodeFromBytes(data []byte, df gopacket.DecodeFeedback) error {
	body, err := g.Decode(data, df)
	if err != nil {
		return err
	}

	minBodyLength := 1
	if len(body) < minBodyLength {
		df.SetTruncated()
		return fmt.Errorf("invalid capabilities response: need at least %v bytes, got %v",
			minBodyLength, len(body))
	}

	periods := int(body[0])
	switch {
	case periods == 0:
		g.PowerRollingAvgTimePeriods = g.PowerRollingAvgTimePeriods[:0]
	case len(body) < 1+periods:
		df.SetTruncated()
		return fmt.Errorf("managed system indicated %v supported rolling "+
			"average time periods, but only room for %v in payload of length "+
			"%v", periods, len(body)-1, len(body))
	default:
		g.PowerRollingAvgTimePeriods = make([]time.Duration, periods)
		for i := 0; i < periods; i++ {
			g.PowerRollingAvgTimePeriods[i] = rollingAvgPeriodDuration(body[1+i])
		}
	}

	g.Contents = data[:len(data)-len(body)+minBodyLength+periods]
	g.Payload = body[minBodyLength+periods:]

	return nil
}

type getDCMICapabilitiesInfoCmd GetDCMICapabilitiesInfoReq

func (*getDCMICapabilitiesInfoCmd) Operation() *ipmi.Operation {
	return &operationGetDCMICapabilitiesInfoReq
}

func (c *getDCMICapabilitiesInfoCmd) Request() gopacket.SerializableLayer {
	return (*GetDCMICapabilitiesInfoReq)(c)
}

type GetDCMICapabilitiesInfoSupportedCapabilitiesCmd struct {
	getDCMICapabilitiesInfoCmd
	Rsp GetDCMICapabilitiesInfoSupportedCapabilitiesRsp
}

func NewGetDCMICapabilitiesInfoSupportedCapabilitiesCmd() *GetDCMICapabilitiesInfoSupportedCapabilitiesCmd {
	return &GetDCMICapabilitiesInfoSupportedCapabilitiesCmd{
		getDCMICapabilitiesInfoCmd: getDCMICapabilitiesInfoCmd{
			Parameter: 1,
		},
	}
}

// Name returns "Get DCMI Capabilities Info (Supported Capabilities)".
func (*GetDCMICapabilitiesInfoSupportedCapabilitiesCmd) Name() string {
	return "Get DCMI Capabilities Info (Supported Capabilities)"
}

func (c *GetDCMICapabilitiesInfoSupportedCapabilitiesCmd) Response() gopacket.DecodingLayer {
	return &c.Rsp
}

type GetDCMICapabilitiesInfoMandatoryPlatformAttrsCmd struct {
	getDCMICapabilitiesInfoCmd
	Rsp GetDCMICapabilitiesInfoMandatoryPlatformAttrsRsp
}

func NewGetDCMICapabilitiesInfoMandatoryPlatformAttrsCmd() *GetDCMICapabilitiesInfoMandatoryPlatformAttrsCmd {
	return &GetDCMICapabilitiesInfoMandatoryPlatformAttrsCmd{
		getDCMICapabilitiesInfoCmd: getDCMICapabilitiesInfoCmd{
			Parameter: 2,
		},
	}
}

// Name returns "Get DCMI Capabilities Info (Mandatory Platform Attributes)".
func (*GetDCMICapabilitiesInfoMandatoryPlatformAttrsCmd) Name() string {
	return "Get DCMI Capabilities Info (Mandatory Platform Attributes)"
}

func (c *GetDCMICapabilitiesInfoMandatoryPlatformAttrsCmd) Response() gopacket.DecodingLayer {
	return &c.Rsp
}

type GetDCMICapabilitiesInfoOptionalPlatformAttrsCmd struct {
	getDCMICapabilitiesInfoCmd
	Rsp GetDCMICapabilitiesInfoOptionalPlatformAttrsRsp
}

func NewGetDCMICapabilitiesInfoOptionalPlatformAttrsCmd() *GetDCMICapabilitiesInfoOptionalPlatformAttrsCmd {
	return &GetDCMICapabilitiesInfoOptionalPlatformAttrsCmd{
		getDCMICapabilitiesInfoCmd: getDCMICapabilitiesInfoCmd{
			Parameter: 3,
		},
	}
}

// Name returns "Get DCMI Capabilities Info (Optional Platform Attributes)".
func (*GetDCMICapabilitiesInfoOptionalPlatformAttrsCmd) Name() string {
	return "Get DCMI Capabilities Info (Optional Platform Attributes)"
}

func (c *GetDCMICapabilitiesInfoOptionalPlatformAttrsCmd) Response() gopacket.DecodingLayer {
	return &c.Rsp
}

type GetDCMICapabilitiesInfoManageabilityAccessAttrsCmd struct {
	getDCMICapabilitiesInfoCmd
	Rsp GetDCMICapabilitiesInfoManageabilityAccessAttrsRsp
}

func NewGetDCMICapabilitiesInfoManageabilityAccessAttrsCmd() *GetDCMICapabilitiesInfoManageabilityAccessAttrsCmd {
	return &GetDCMICapabilitiesInfoManageabilityAccessAttrsCmd{
		getDCMICapabilitiesInfoCmd: getDCMICapabilitiesInfoCmd{
			Parameter: 4,
		},
	}
}

// Name returns "Get DCMI Capabilities Info (Manageability Access Attributes)".
func (*GetDCMICapabilitiesInfoManageabilityAccessAttrsCmd) Name() string {
	return "Get DCMI Capabilities Info (Manageability Access Attributes)"
}

func (c *GetDCMICapabilitiesInfoManageabilityAccessAttrsCmd) Response() gopacket.DecodingLayer {
	return &c.Rsp
}

type GetDCMICapabilitiesInfoEnhancedSystemPowerStatisticsAttrsCmd struct {
	getDCMICapabilitiesInfoCmd
	Rsp GetDCMICapabilitiesInfoEnhancedSystemPowerStatisticsAttrsRsp
}

func NewGetDCMICapabilitiesInfoEnhancedSystemPowerStatisticsAttrsCmd() *GetDCMICapabilitiesInfoEnhancedSystemPowerStatisticsAttrsCmd {
	return &GetDCMICapabilitiesInfoEnhancedSystemPowerStatisticsAttrsCmd{
		getDCMICapabilitiesInfoCmd: getDCMICapabilitiesInfoCmd{
			Parameter: 5,
		},
	}
}

// Name returns "Get DCMI Capabilities Info (Enhanced System Power Statistics
// Attributes)".
func (*GetDCMICapabilitiesInfoEnhancedSystemPowerStatisticsAttrsCmd) Name() string {
	return "Get DCMI Capabilities Info (Enhanced System Power Statistics Attributes)"
}

func (c *GetDCMICapabilitiesInfoEnhancedSystemPowerStatisticsAttrsCmd) Response() gopacket.DecodingLayer {
	return &c.Rsp
}
