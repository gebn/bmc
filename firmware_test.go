package bmc

import (
	"testing"

	"github.com/gebn/bmc/pkg/iana"
	"github.com/gebn/bmc/pkg/ipmi"
)

func TestFirmwareVersion(t *testing.T) {
	tests := []struct {
		in   *ipmi.GetDeviceIDRsp
		want string
	}{
		{
			&ipmi.GetDeviceIDRsp{
				ID:                        33,
				Revision:                  1,
				MajorFirmwareRevision:     1,
				MinorFirmwareRevision:     20,
				Manufacturer:              iana.EnterpriseIntel,
				Product:                   73,
				AuxiliaryFirmwareRevision: [4]byte{0x01, 0x17, 0xa1, 0x16},
			},
			"01.20.5793",
		},
		{
			&ipmi.GetDeviceIDRsp{
				ID:                        32,
				Revision:                  1,
				MajorFirmwareRevision:     2,
				MinorFirmwareRevision:     41,
				Manufacturer:              iana.EnterpriseDell,
				Product:                   1,
				AuxiliaryFirmwareRevision: [4]byte{0x00, 0x07, 0x28, 0x28},
			},
			"2.41.40.40b07",
		},
		{
			&ipmi.GetDeviceIDRsp{
				ID:                        32,
				Revision:                  1,
				MajorFirmwareRevision:     2,
				MinorFirmwareRevision:     50,
				Manufacturer:              iana.EnterpriseDell,
				Product:                   1,
				AuxiliaryFirmwareRevision: [4]byte{0x00, 0x21, 0x32, 0x32},
			},
			"2.50.50.50b33",
		},
		{
			&ipmi.GetDeviceIDRsp{
				ID:                        32,
				Revision:                  1,
				MajorFirmwareRevision:     3,
				MinorFirmwareRevision:     15,
				Manufacturer:              iana.EnterpriseDell,
				Product:                   1,
				AuxiliaryFirmwareRevision: [4]byte{0x00, 0x01, 0x11, 0x0f},
			},
			"3.15.17.15b01",
		},
		{
			&ipmi.GetDeviceIDRsp{
				ID:                        32,
				Revision:                  1,
				MajorFirmwareRevision:     3,
				MinorFirmwareRevision:     45,
				Manufacturer:              iana.EnterpriseQuanta,
				Product:                   12866,
				AuxiliaryFirmwareRevision: [4]byte{0x01, 0x00, 0x00, 0x00},
			},
			"3.45.01",
		},
		{
			&ipmi.GetDeviceIDRsp{
				ID:                        32,
				Revision:                  1,
				MajorFirmwareRevision:     3,
				MinorFirmwareRevision:     72,
				Manufacturer:              iana.EnterpriseSuperMicro,
				Product:                   2052,
				AuxiliaryFirmwareRevision: [4]byte{},
			},
			"3.72",
		},
	}
	for _, test := range tests {
		if got := FirmwareVersion(test.in); got != test.want {
			t.Errorf("FirmwareVersion(%v) = %v, want %v", test.in, got,
				test.want)
		}
	}
}
