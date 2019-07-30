package dcmi

// SessionCommands represents the high-level API for commands that can be
// executed within a session.
type SessionCommands interface {
	SessionlessCommands
}
