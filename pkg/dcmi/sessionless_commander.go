package dcmi

import (
	"context"

	"github.com/gebn/bmc"
)

type sessionlessCommander struct {
	bmc.Sessionless
}

func (s sessionlessCommander) GetDCMICapabilitiesInfo(ctx context.Context) (*GetDCMICapabilitiesInfoRsp, error) {
	cmd := &GetDCMICapabilitiesInfoCmd{}
	// technically, DCMI uses a different set of codes, but the overlap is
	// great enough that this does not need to be accounted for
	if err := bmc.ValidateResponse(s.Sessionless.SendCommand(ctx, cmd)); err != nil {
		return nil, err
	}
	return &cmd.Rsp, nil
}

// SessionlessCommander wraps a session-less connection in a context that
// provides high-level access to DCMI commands. For convenience, this function
// accepts the Sessionless interface, however DCMI is unlikely to work over IPMI
// v1.5. When sending repeated commands, it is recommended to use the
// SendCommand() method on the connection directly to reduce the number of
// allocations.
func SessionlessCommander(s bmc.Sessionless) SessionlessCommands {
	return &sessionlessCommander{
		Sessionless: s,
	}
}
