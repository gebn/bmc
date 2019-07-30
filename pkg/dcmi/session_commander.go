package dcmi

import (
	"github.com/gebn/bmc"
)

type sessionCommander struct {
	SessionlessCommands
	bmc.Session
}

// SessionCommander wraps a session-based connection in a context that provides
// high-level access to DCMI commands. For convenience, this function accepts
// the Session interface, however DCMI is unlikely to work over IPMI v1.5. When
// sending repeated commands, it is recommended to use the SendCommand() method
// on the connection directly to reduce the number of allocations.
func SessionCommander(s bmc.Session) SessionCommands {
	return &sessionCommander{
		SessionlessCommands: SessionlessCommander(s),
		Session:             s,
	}
}
