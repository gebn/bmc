package ipmi

import (
	"github.com/google/go-cmp/cmp"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"testing"
)

func TestReserveSDRRepositoryRspDecodeFomBytes(t *testing.T) {
	tests := []struct {
		in   []byte
		want *ReserveSDRRepositoryRsp
	}{
		{
			in: []byte{0x20, 0x58},
			want: &ReserveSDRRepositoryRsp{
				BaseLayer:     layers.BaseLayer{Contents: []byte{0x20, 0x58}},
				ReservationID: 22560,
			},
		},
	}
	for _, test := range tests {
		rsp := &ReserveSDRRepositoryRsp{}
		if err := rsp.DecodeFromBytes(test.in, gopacket.NilDecodeFeedback); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if diff := cmp.Diff(test.want, rsp); diff != "" {
			t.Errorf("decode %v = %v, want %v: %v", test.in, rsp, test.want, diff)
		}
	}
}
