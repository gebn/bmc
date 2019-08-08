package bmc

import (
	"context"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// we care less about version here - distribution will follow connections
	// unless the user is treating different versions differently, in which case
	// they probably don't care about the break-down

	// we could add authentication, integrity and confidentiality labels to a
	// new algorithms counter, however that will remain static for a given fleet
	// - if people are interested in algorithm support, this is better
	// discovered via infrequent sweeps

	// we could time session establishment, however do we really care, provided
	// it succeeds? would also be a very sparse histogram

	// session re-opens must be tracked by the user of the library; we don't
	// have any visibility here (at least not currently)

	sessionOpenAttempts = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: namespace,
		Subsystem: "session",
		Name:      "open_attempts_total",
		Help:      "The number of times session establishment has begun.",
	})
	sessionOpenFailures = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: namespace,
		Subsystem: "session",
		Name:      "open_failures_total",
		Help: "The number of times session establishment did not produce " +
			"a usable session-based connection.",
	})
	sessionsOpen = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: namespace,
		Subsystem: "sessions",
		Name:      "open",
		Help: "The number of sessions currently established, including " +
			"those that have failed to close cleanly.",
	})
)

// Session is an established session-based IPMI v1.5 or 2.0 connection. More
// specifically, it is a multi-session connection, as the spec demands this of
// the LAN interface. Commands are sent in the context of the session, using
// the negotiated integrity and confidentiality algorithms.
type Session interface {
	Connection
	SessionCommands

	// ID returns our identifier for this session. Note the managed system and
	// remote console share the same identifier in IPMI v1.5, however each
	// chooses its own identifier in v2.0, so they likely differ.
	ID() uint32

	// Close closes the session by sending a Close Session command to the BMC.
	// As the underlying transport/socket is used but not managed by
	// connections, it is left open in case the user wants to continue issuing
	// session-less commands or establish a new session. It is envisaged that
	// this call is deferred immediately after successful session establishment.
	// If an error is returned, the session can be assumed to be closed.
	Close(context.Context) error
}
