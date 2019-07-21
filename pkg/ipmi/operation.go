package ipmi

import (
	"fmt"

	"github.com/gebn/bmc/pkg/iana"

	"github.com/google/gopacket"
)

// Operation uniquely identifies a command that the BMC can perform. This is not
// terminology defined in the specification; this exists to allow us to identify
// the payload type of a particular IPMI message, which contains this type.
type Operation struct {

	// Function is the network function code of the message. The command field
	// indicates the specific functionality desired within this function class.
	Function NetworkFunction

	// Command is the BMC function being requested, or the response.
	Command Command

	// Body is the defining body code. It is only relevant if the function is
	// Group, and is ignored otherwise.
	Body BodyCode

	// Enterprise is the enterprise number when the function is OEM/Group. It is
	// ignored otherwise.
	Enterprise iana.Enterprise
}

var (
	OperationGetDeviceIDReq = Operation{
		Function: NetworkFunctionAppReq,
		Command:  0x01,
	}
	OperationGetDeviceIDRsp = Operation{
		Function: NetworkFunctionAppRsp,
		Command:  0x01,
	}
	OperationGetSystemGUIDReq = Operation{
		Function: NetworkFunctionAppReq,
		Command:  0x37,
	}
	OperationGetSystemGUIDRsp = Operation{
		Function: NetworkFunctionAppRsp,
		Command:  0x37,
	}
	OperationGetChannelAuthenticationCapabilitiesReq = Operation{
		Function: NetworkFunctionAppReq,
		Command:  0x38,
	}
	OperationGetChannelAuthenticationCapabilitiesRsp = Operation{
		Function: NetworkFunctionAppRsp,
		Command:  0x38,
	}
	OperationCloseSessionReq = Operation{
		Function: NetworkFunctionAppReq,
		Command:  0x3c,
	}

	// operationLayerTypes tells us which layer comes next given a network
	// function and command. It should never be modified during runtime, as
	// there is no way to guarantee exclusive access.
	operationLayerTypes = map[Operation]gopacket.LayerType{
		OperationGetDeviceIDRsp:   LayerTypeGetDeviceIDRsp,
		OperationGetSystemGUIDRsp: LayerTypeGetSystemGUIDRsp,
		//OperationGetChannelAuthenticationCapabilitiesReq: LayerTypeGetChannelAuthenticationCapabilitiesReq,
		OperationGetChannelAuthenticationCapabilitiesRsp: LayerTypeGetChannelAuthenticationCapabilitiesRsp,
	}
)

func (o Operation) String() string {
	return fmt.Sprintf("%v, %v", o.Function, o.NextLayerType())
}

func (o Operation) NextLayerType() gopacket.LayerType {
	if layer, ok := operationLayerTypes[o]; ok {
		return layer
	}
	return gopacket.LayerTypePayload
}
