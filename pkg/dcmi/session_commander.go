package dcmi

import (
	"context"

	"github.com/gebn/bmc"
)

type sessionCommander struct {
	SessionlessCommands
	bmc.Session
}

func (s sessionCommander) GetPowerReading(ctx context.Context, r *GetPowerReadingReq) (*GetPowerReadingRsp, error) {
	cmd := &GetPowerReadingCmd{
		Req: *r,
	}
	if err := bmc.ValidateResponse(s.SendCommand(ctx, cmd)); err != nil {
		return nil, err
	}
	return &cmd.Rsp, nil
}

// NewSessionCommander wraps a session-based connection in a context that
// provides high-level access to DCMI commands. For convenience, this function
// accepts the Session interface, however DCMI is unlikely to work over IPMI
// v1.5. When sending repeated commands, it is recommended to use the
// SendCommand() method on the connection directly to reduce the number of
// allocations.
func NewSessionCommander(s bmc.Session) SessionCommands {
	return &sessionCommander{
		SessionlessCommands: NewSessionlessCommander(s),
		Session:             s,
	}
}
