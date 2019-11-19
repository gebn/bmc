package dcmi

import (
	"context"
)

// SessionCommands represents the high-level API for commands that can be
// executed within a session.
type SessionCommands interface {
	SessionlessCommands

	GetPowerReading(context.Context, *GetPowerReadingReq) (*GetPowerReadingRsp, error)

	GetDCMISensorInfo(context.Context, *GetDCMISensorInfoReq) (*GetDCMISensorInfoRsp, error)
}
