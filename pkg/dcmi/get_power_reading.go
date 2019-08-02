package dcmi

import (
	"encoding/binary"
	"fmt"
	"time"

	"github.com/gebn/bmc/pkg/ipmi"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
)

// GetPowerReadingReq implements the Get Power Reading command, specified in
// 6.6.1.
type GetPowerReadingReq struct {
	layers.BaseLayer

	// Mode indicates whether enhanced system power statistics are desired. Must
	// be SystemPowerStatisticsModeNormal for v1.0, in which case the BMC is in
	// control over the statistics reporting time period.
	Mode SystemPowerStatisticsMode

	// If Mode is SystemPowerStatisticsModeEnhanced, the rolling average time
	// period over which to retrieve statistics. Note that this cannot be
	// arbitrary - only a period returned in the PowerRollingAvgTimePeriods
	// field of Get DCMI Capabilities Info response can be used. If scraping the
	// power reading, this should be equal to the scrape interval.
	Period time.Duration
}

func (*GetPowerReadingReq) LayerType() gopacket.LayerType {
	return layerTypeGetPowerReadingReq
}

func (g *GetPowerReadingReq) SerializeTo(b gopacket.SerializeBuffer, _ gopacket.SerializeOptions) error {
	bytes, err := b.PrependBytes(3)
	if err != nil {
		return err
	}
	bytes[0] = uint8(g.Mode)
	switch g.Mode {
	case SystemPowerStatisticsModeEnhanced:
		bytes[1] = rollingAvgPeriodByte(g.Period)
	default:
		// v1.0; ignore g.Period even if set
		bytes[1] = 0x00
	}
	bytes[2] = 0x00
	return nil
}

// GetPowerReadingRsp represents the response to a Get Power Reading command,
// specified in 6.6.1. Be wary of interpreting this response without checking
// the timestamp and period (especially the latter if not using enhanced mode,
// where the remote console is in control of the period).
type GetPowerReadingRsp struct {
	layers.BaseLayer

	// Instantaneous gives the current power consumption in watts.
	Instantaneous uint16

	// Min gives the minimum power over the period in watts.
	Min uint16

	// Max gives the maximum power over the period in watts.
	Max uint16

	// Avg gives the average power over the period in watts.
	Avg uint16

	// Timestamp indicates when the power readings are for. If using enhanced
	// power statistics, this is the end of the averaging window.
	Timestamp time.Time

	// Period is the sampling period over which the controller is reporting
	// statistics. If SystemPowerStatisticsModeEnhanced was used, this will
	// equal the duration requested, otherwise it is up the the BMC.
	Period time.Duration

	// Active indicates whether power measurement is currently active.
	Active bool
}

func (*GetPowerReadingRsp) LayerType() gopacket.LayerType {
	return layerTypeGetPowerReadingRsp
}

func (g *GetPowerReadingRsp) CanDecode() gopacket.LayerClass {
	return g.LayerType()
}

func (*GetPowerReadingRsp) NextLayerType() gopacket.LayerType {
	return gopacket.LayerTypePayload
}

func (g *GetPowerReadingRsp) DecodeFromBytes(data []byte, df gopacket.DecodeFeedback) error {
	if len(data) < 17 {
		df.SetTruncated()
		return fmt.Errorf("power reading response must be 17 bytes, got %v", len(data))
	}

	g.Instantaneous = binary.LittleEndian.Uint16(data[0:2])
	g.Min = binary.LittleEndian.Uint16(data[2:4])
	g.Max = binary.LittleEndian.Uint16(data[4:6])
	g.Avg = binary.LittleEndian.Uint16(data[6:8])

	// TODO possibly handle unspecified (0xffffffff) and relative to system
	// startup (<0x20000000) numbers of seconds - unclear how prevalent these
	// are. Rules are in section 37 of IPMI v2.0. Don't forget tests.
	g.Timestamp = time.Unix(int64(binary.LittleEndian.Uint32(data[8:12])), 0)
	g.Period = time.Millisecond *
		time.Duration(binary.LittleEndian.Uint32(data[12:16]))
	g.Active = data[16]&(1<<6) != 0
	return nil
}

type GetPowerReadingCmd struct {
	Req GetPowerReadingReq
	Rsp GetPowerReadingRsp
}

// Name returns "Get Power Reading".
func (*GetPowerReadingCmd) Name() string {
	return "Get Power Reading"
}

func (*GetPowerReadingCmd) Operation() *ipmi.Operation {
	return &operationGetPowerReadingReq
}

func (c *GetPowerReadingCmd) Request() gopacket.SerializableLayer {
	return &c.Req
}

func (c *GetPowerReadingCmd) Response() gopacket.DecodingLayer {
	return &c.Rsp
}
