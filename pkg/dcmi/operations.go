package dcmi

import (
	"github.com/gebn/bmc/pkg/ipmi"
)

var (
	operationGetDCMICapabilitiesInfoReq = ipmi.Operation{
		Function: ipmi.NetworkFunctionGroupReq,
		Body:     ipmi.BodyCodeDCMI,
		Command:  0x01,
	}
	operationGetPowerReadingReq = ipmi.Operation{
		Function: ipmi.NetworkFunctionGroupReq,
		Body:     ipmi.BodyCodeDCMI,
		Command:  0x02,
	}
)
