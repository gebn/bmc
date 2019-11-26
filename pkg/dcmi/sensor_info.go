package dcmi

import (
	"context"

	"github.com/gebn/bmc"
	"github.com/gebn/bmc/pkg/ipmi"
)

var (
	ipmiSensorEntityIDs = []ipmi.EntityID{
		ipmi.EntityIDAirInlet,
		ipmi.EntityIDProcessor,
		ipmi.EntityIDSystemBoard,
	}
	dcmiSensorEntityIDs = []ipmi.EntityID{
		ipmi.EntityIDDCMIAirInlet,
		ipmi.EntityIDDCMIProcessor,
		ipmi.EntityIDDCMISystemBoard,
	}
)

// SensorInfo models the exhaustive result of the Get DCMI Sensor Info command,
// obtained through calling it possibly multiple times to enumerate all
// instances. Fields correspond to entity IDs defined in Table 6-6, 6-8 and 6-14
// of DCMI v1.0, v1.1 and v1.5 respectively.
type SensorInfo struct {
	Inlet     []ipmi.RecordID
	CPU       []ipmi.RecordID
	Baseboard []ipmi.RecordID
}

// sensorMap represents an intermediate result of asking the BMC for all sensors
// with either IPMI or DCMI entity IDs. This could be the type of SensorInfo,
// but a struct was considered a clearer, more concise and internally correct
// representation, e.g. this could insinuate all record IDs were returned, or
// sensors with a mixture of IPMI and DCMI entities could be returned.
type sensorMap map[ipmi.EntityID][]ipmi.RecordID

// CountRecordIDs returns the number of record IDs in a sensor map.
func (m sensorMap) CountRecordIDs() int {
	entries := 0
	for _, v := range m {
		entries += len(v)
	}
	return entries
}

// GetSensorInfo retrieves the RecordIDs of all inlet, CPU and baseboard
// temperatures for a system. It is an abstraction over Get DCMI Sensor Info,
// handling the case where the BMC cares about IPMI vs. DCMI entity IDs, and the
// case where there are >8 sensors, so the command has to be invoked many times.
// For practical use, this should always be preferred to calling Get DCMI Sensor
// Info manually. To avoid duplicate sensors, this method will only return
// either sensors obtained by passing IPMI EntityIDs, or those obtained by
// passing DCMI EntityIDs - never an intersection of the two.
func GetSensorInfo(ctx context.Context, s bmc.Session) (*SensorInfo, error) {
	cmd := &GetDCMISensorInfoCmd{
		Req: GetDCMISensorInfoReq{
			Type: ipmi.SensorTypeTemperature,
		},
	}

	// we optimistically assume DCMI v1.5, and send IPMI entity IDs first; the
	// BMC will map these if it can
	sensors, err := getSensorMap(ctx, s, cmd, ipmiSensorEntityIDs)

	// if we got *any* sensors, we assume the BMC knew what we meant with IPMI
	// entity IDs and returned what it had. The inverted error handling is to
	// ensure we fall back - some DCMI v1.0 and v1.1 BMCs may complain if we
	// give them an IPMI entity ID; we should still try DCMI IDs.
	if err == nil && sensors.CountRecordIDs() > 0 {
		return &SensorInfo{
			Inlet:     sensors[ipmi.EntityIDAirInlet],
			CPU:       sensors[ipmi.EntityIDProcessor],
			Baseboard: sensors[ipmi.EntityIDSystemBoard],
		}, nil
	}

	// fall back on older DCMI entity IDs; we will also end up here if the BMC
	// does not support the Get DCMI Sensor Info command
	sensors, err = getSensorMap(ctx, s, cmd, dcmiSensorEntityIDs)
	if err != nil {
		return nil, err
	}
	return &SensorInfo{
		Inlet:     sensors[ipmi.EntityIDDCMIAirInlet],
		CPU:       sensors[ipmi.EntityIDDCMIProcessor],
		Baseboard: sensors[ipmi.EntityIDDCMISystemBoard],
	}, nil
}

// getSensorMap retrieves all sensors for a given list of entities. These
// entities are assumed to all have the same type, equal to the value already
// set in cmd's request, which is not modified. Note that when querying for DCMI
// entity IDs, DCMI v1.1+ BMCs may return record IDs for SDRs with the
// corresponding IPMI entity IDs, e.g. querying for 0x40 may return an SDR for
// entity 0x37.
//
// Regardless of the number of sensors retrieved, the returned map will have
// length equal to the unique set of elements in entities, so avoid using len()
// on the map itself to check if anything was returned.
func getSensorMap(ctx context.Context, s bmc.Session, cmd *GetDCMISensorInfoCmd, entities []ipmi.EntityID) (sensorMap, error) {
	sensors := sensorMap{}
	for _, entityID := range entities {
		cmd.Req.Entity = entityID
		recordIDs, err := getEntityInstances(ctx, s, cmd)
		if err != nil {
			return nil, err
		}
		sensors[entityID] = recordIDs
	}
	return sensors, nil
}

// getEntityInstances retrieves all RecordIDs for a given EntityID. The
// sensor type and entity ID should be set on the input command before calling
// this function. This function mutates the Instance and InstanceStart fields in
// the request to enumerate all sensors. Depending on the version of DCMI, the
// record IDs returned may have a different, but compatible, entity ID.
func getEntityInstances(ctx context.Context, s bmc.Session, cmd *GetDCMISensorInfoCmd) ([]ipmi.RecordID, error) {
	recordIDs := []ipmi.RecordID{}
	totalInstances := len(recordIDs) + 1 // to ensure we enter the loop once

	// we always want to retrieve as many instances in each command as possible;
	// we don't touch this again
	cmd.Req.Instance = 0

	for len(recordIDs) < totalInstances {
		cmd.Req.InstanceStart = uint8(len(recordIDs) + 1)
		if err := bmc.ValidateResponse(s.SendCommand(ctx, cmd)); err != nil {
			return nil, err
		}

		// we need to update this to the real total after the first response,
		// but we keep it updated in case the number of sensors changes over
		// time (very unlikely)
		totalInstances = int(cmd.Rsp.Instances)

		// the backing array for the slice is overwritten each command, so we
		// must copy it - and we want a single slice of all recordIDs anyway
		for _, recordID := range cmd.Rsp.RecordIDs {
			recordIDs = append(recordIDs, recordID)
		}

		// this prevents looping forever, and the case where we allow retrieving
		// an unreasonable number of sensors (255 is the max specifiable in the
		// command response)
		if len(cmd.Rsp.RecordIDs) == 0 || len(recordIDs) == 255 {
			break
		}
	}
	return recordIDs, nil
}
