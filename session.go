package bmc

import (
	"github.com/gebn/bmc/pkg/ipmi"
)

type sessionRspLayers struct {
	getDeviceIDRspLayer      ipmi.GetDeviceIDRsp
	getChassisStatusRspLayer ipmi.GetChassisStatusRsp
}
