package ipmi

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
)

func TestGetChassisStatusRspDecodeFromBytes(t *testing.T) {
	tests := []struct {
		in   []byte
		want *GetChassisStatusRsp
	}{
		{
			// too short
			[]byte{0x20},
			nil,
		},
		{
			// no front panel button capabilities
			[]byte{0x20, 0x00, 0x60},
			&GetChassisStatusRsp{
				BaseLayer: layers.BaseLayer{
					Contents: []byte{0x20, 0x00, 0x60},
					Payload:  []byte{}, // if nil, cmp.Diff reports difference
				},
				PowerRestorePolicy:   PowerRestorePolicyPriorState,
				ChassisIdentifyState: ChassisIdentifyStateIndefinite,
			},
		},
		{
			[]byte{0xaa, 0xaa, 0xaa, 0xaa},
			&GetChassisStatusRsp{
				BaseLayer: layers.BaseLayer{
					Contents: []byte{0xaa, 0xaa, 0xaa, 0xaa},
					Payload:  []byte{},
				},
				PowerRestorePolicy:          PowerRestorePolicyPriorState,
				PowerFault:                  true,
				PowerOverload:               true,
				LastPowerDownFault:          true,
				LastPowerDownOverload:       true,
				ChassisIdentifyState:        ChassisIdentifyStateUnknown,
				CoolingFault:                true,
				Lockout:                     true,
				StandbyButtonDisableAllowed: true,
				ResetButtonDisableAllowed:   true,
				StandbyButtonDisabled:       true,
				ResetButtonDisabled:         true,
			},
		},
		{
			[]byte{0x55, 0x55, 0x55, 0x55},
			&GetChassisStatusRsp{
				BaseLayer: layers.BaseLayer{
					Contents: []byte{0x55, 0x55, 0x55, 0x55},
					Payload:  []byte{},
				},
				PowerRestorePolicy:                      PowerRestorePolicyPowerOn,
				PowerControlFault:                       true,
				Interlock:                               true,
				PoweredOn:                               true,
				PoweredOnByIPMI:                         true,
				LastPowerDownInterlock:                  true,
				LastPowerDownSupplyFailure:              true,
				ChassisIdentifyState:                    ChassisIdentifyStateTemporary,
				DriveFault:                              true,
				Intrusion:                               true,
				DiagnosticInterruptButtonDisableAllowed: true,
				PowerOffButtonDisableAllowed:            true,
				DiagnosticInterruptButtonDisabled:       true,
				PowerOffButtonDisabled:                  true,
			},
		},
	}
	for _, test := range tests {
		rsp := &GetChassisStatusRsp{}
		err := rsp.DecodeFromBytes(test.in, gopacket.NilDecodeFeedback)
		switch {
		case err == nil && test.want == nil:
			t.Errorf("expected error decoding %v, got none", test.in)
		case err == nil && test.want != nil:
			if diff := cmp.Diff(test.want, rsp); diff != "" {
				t.Errorf("decode %v = %v, want %v: %v", test.in, rsp, test.want, diff)
			}
		case err != nil && test.want != nil:
			t.Errorf("unexpected error: %v", err)
		}
	}
}
