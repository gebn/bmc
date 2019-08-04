package bmc

import (
	"context"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	sessionEstablishAttempts = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: "session",
			Name:      "establish_attempts_total",
			Help:      "The number of times session establishment has begun.",
		},
		[]string{"version"},
	)
	sessionEstablishFailures = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: "session",
			Name:      "establish_failures_total",
			Help:      "The number of times session establishment did not produce a usable session-based connection.",
		},
		[]string{"version"},
	)
	sessionEstablishSuccesses = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: "session",
			Name:      "establish_successes_total",
			Help:      "The number of times session establishment resulted in a usable session-based connection.",
		},
		// we can derive this metric from attempts - failures, so its purpose is
		// to provide the algorithms used. "integrity" and "confidentiality" are
		// always "none" for v1.5.
		[]string{"version", "authentication", "integrity", "confidentiality"},
	)
	sessionEstablishDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: namespace,
			Subsystem: "session",
			Name:      "establish_duration_seconds",
			Help: "The end-to-end time taken by the NewSession() method, both " +
				"when it succeeds and fails. Note, this does not include a " +
				"Get Channel Authentication Capabilities call, as that is not " +
				"mandatory.",
		},
		[]string{"version"},
	)
	sessionClose = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: "session",
			Name:      "close_total",
			Help:      "The number of attempts to close a session-based connection.",
		},
		[]string{"version"},
	)
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
