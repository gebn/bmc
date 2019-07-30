package dcmi

import (
	"context"
)

// SessionlessCommands contains the high-level API for commands that can be
// executed outside the context of a session.
type SessionlessCommands interface {
	GetDCMICapabilitiesInfo(context.Context) (*GetDCMICapabilitiesInfoRsp, error)
}
