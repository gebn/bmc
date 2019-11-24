package dcmi

import (
	"encoding/binary"
	"fmt"

	"github.com/gebn/bmc/pkg/ipmi"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
)

// GetDCMISensorInfoReq represents the Get DCMI Sensor Info command, specified
// in 6.5.2 of DCMI v1.0, v1.1 and v1.5. This command is used to help system
// management software find sensors monitoring specific things the DCMI authors
// believe are particularly pertinent, that are difficult to identify using IPMI
// alone. Note, even though read-only, this has a minimum privilege level of
// Operator.
type GetDCMISensorInfoReq struct {
	layers.BaseLayer

	// Type is the kind of thing we are interested in finding sensors for. As of
	// DCMI v1.5, the only valid value here is temperature (0x01).
	Type ipmi.SensorType

	// Entity is the type of component we are interested in. As of DCMI v1.5,
	// this is inlet, CPU or baseboard temperature. In v1.0 and v1.1, 0x40, 0x41
	// and 0x42 respectively were used to refer to these, however v1.5
	// recommends using the IPMI entity IDs: 0x37, 0x03 and 0x07 respectively.
	// DCMI v1.5 promises to map the DCMI values to their IPMI equivalents,
	// encouraging use of the latter. In practice, try the IPMI entities first,
	// falling back on the DCMI ones if there is an empty response.
	Entity ipmi.EntityID

	// Instance specifies the instance of the SDR to retrieve. 0x00 indicates to
	// retrieve all instance associated with the entity.
	Instance ipmi.EntityInstance

	// InstanceStart is for use when Instance is 0x00. It is intended for when
	// there are >8 instances of a particular sensor, so they cannot all be
	// returned in a single response. This can be used to offset the start
	// record ID. It looks like the idea is to set this to the highest instance
	// received + 1. Note IPMI makes no guarantees about the instance space,
	// however it seems DCMI instances must be sequential and incrementing.
	InstanceStart uint8
}

func (*GetDCMISensorInfoReq) LayerType() gopacket.LayerType {
	return layerTypeGetDCMISensorInfoReq
}

func (g *GetDCMISensorInfoReq) SerializeTo(b gopacket.SerializeBuffer, _ gopacket.SerializeOptions) error {
	bytes, err := b.PrependBytes(4)
	if err != nil {
		return err
	}
	bytes[0] = uint8(g.Type)
	bytes[1] = uint8(g.Entity)
	bytes[2] = uint8(g.Instance)
	if g.Instance == 0 {
		bytes[3] = g.InstanceStart
	} else {
		bytes[3] = 0
	}
	return nil
}

// GetDCMISensorInfoRsp represents the BMC's response to a Get DCMI Sensor Info
// request. It is specified in 6.5.2 of DCMI v1.0, v1.1 and v1.5.
type GetDCMISensorInfoRsp struct {
	layers.BaseLayer

	// Instances gives the total number of instances of the requested entity. If
	// this is greater than the number of record IDs returned (and Instance was
	// not specified in the request), it is an invitation to issue a new request
	// with InstanceStart set.
	Instances uint8

	// RecordIDs contains the record IDs returned by the BMC. In DCMI v1.1 and
	// v1.5, these may include SDRs for both the DCMI and IPMI entities.
	RecordIDs []ipmi.RecordID
}

func (*GetDCMISensorInfoRsp) LayerType() gopacket.LayerType {
	return layerTypeGetDCMISensorInfoRsp
}

func (g *GetDCMISensorInfoRsp) CanDecode() gopacket.LayerClass {
	return g.LayerType()
}

func (*GetDCMISensorInfoRsp) NextLayerType() gopacket.LayerType {
	return gopacket.LayerTypePayload
}

func (g *GetDCMISensorInfoRsp) DecodeFromBytes(data []byte, df gopacket.DecodeFeedback) error {
	if len(data) < 2 {
		df.SetTruncated()
		return fmt.Errorf("expected at least 2 bytes, got %v", len(data))
	}

	g.Instances = data[0]

	recordIDs := int(data[1]) // it's a uint8, but this eliminates conversions
	// the spec says recordIDs <= 8, but we don't enforce this
	expectLength := 2 + recordIDs*2
	if len(data) < expectLength {
		return fmt.Errorf("expected %v bytes for %v record IDs, got %v",
			expectLength, recordIDs, len(data))
	}

	g.RecordIDs = g.RecordIDs[:0] // it would be nice to set this to len(recordIDs)
	for i := 0; i < recordIDs; i++ {
		offset := 2 + i*2
		recordID := ipmi.RecordID(binary.LittleEndian.Uint16(data[offset:]))
		g.RecordIDs = append(g.RecordIDs, recordID)
	}

	g.BaseLayer.Contents = data[:expectLength]
	g.BaseLayer.Payload = data[expectLength:]
	return nil
}

type GetDCMISensorInfoCmd struct {
	Req GetDCMISensorInfoReq
	Rsp GetDCMISensorInfoRsp
}

// Name returns "Get DCMI Sensor Info".
func (*GetDCMISensorInfoCmd) Name() string {
	return "Get DCMI Sensor Info"
}

func (*GetDCMISensorInfoCmd) Operation() *ipmi.Operation {
	return &operationGetDCMISensorInfoReq
}

func (g *GetDCMISensorInfoCmd) Request() gopacket.SerializableLayer {
	return &g.Req
}

func (g *GetDCMISensorInfoCmd) Response() gopacket.DecodingLayer {
	return &g.Rsp
}
