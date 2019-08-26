package ipmi

import (
	"encoding/binary"
	"fmt"

	"github.com/gebn/bmc/internal/pkg/bcd"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
)

// SDR represents a Sensor Data Record header, outlined at the beginning of 37
// and 43 of IPMI v1.5 and 2.0 respectively.
type SDR struct {
	layers.BaseLayer

	// ID is the current Record ID for the SDR. Note this may change whenever
	// the SDR Repository is modified. See RecordID documentation for more
	// details.
	ID RecordID

	// Version is the version number of the SDR specification. It is used with
	// the Type field to control how the record is parsed.
	Version uint8

	// Type indicates the variety of SDR. Confusingly, not all SDRs pertain to
	// sensors.
	Type RecordType
}

func (*SDR) LayerType() gopacket.LayerType {
	return LayerTypeSDR
}

func (s *SDR) CanDecode() gopacket.LayerClass {
	return s.LayerType()
}

func (*SDR) NextLayerType() gopacket.LayerType {
	return gopacket.LayerTypePayload
}

func (s *SDR) DecodeFromBytes(data []byte, df gopacket.DecodeFeedback) error {
	if len(data) < 5 {
		df.SetTruncated()
		return fmt.Errorf("SDR Header is always 5 bytes, got %v", len(data))
	}
	s.ID = RecordID(binary.LittleEndian.Uint16(data[0:2]))
	s.Version = bcd.Decode(data[2]&0xf)*10 + bcd.Decode(data[2]>>4)
	s.Type = RecordType(data[3])

	s.BaseLayer.Contents = data[:5]
	s.BaseLayer.Payload = data[5:]
	return nil
}
