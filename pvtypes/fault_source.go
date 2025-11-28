package pvtypes

import (
	"strings"
)

// FaultSource identifies which side of the client-server relationship is
// responsible for a validation error, following HTTP status code semantics
// (4xx vs 5xx) but remaining protocol-agnostic.
//
// This allows domain logic to indicate fault attribution without coupling
// to HTTP-specific status codes. The HTTP layer maps FaultSource to
// appropriate status codes (typically 422 for ClientFault, 500 for ServerFaultSource).
//
// Other protocols can map similarly:
//   - gRPC: ClientFault → INVALID_ARGUMENT, ServerFaultSource → INTERNAL
//   - CLI: ClientFault → exit code 1, ServerFaultSource → exit code 2
//   - Message queues: ClientFault → NACK+requeue, ServerFaultSource → DLQ
type FaultSource int

const (
	// UnspecifiedFaultSource indicates fault source was not set (should not occur in production).
	UnspecifiedFaultSource FaultSource = iota

	// ClientFaultSource indicates the error originated from invalid client input.
	// Examples: wrong data type, constraint violation, malformed parameter.
	// The client must modify their request to resolve the error.
	// HTTP mapping: 4xx status codes (typically 422 Unprocessable Entity).
	ClientFaultSource

	// ServerFaultSource indicates the error originated from server misconfiguration or state.
	// Examples: invalid template syntax, missing required configuration.
	// The server administrator must fix the configuration to resolve the error.
	// HTTP mapping: 5xx status codes (typically 500 Internal Server Error).
	ServerFaultSource
)

// String returns the string representation of the fault source.
func (fs FaultSource) String() string {
	switch fs {
	case ClientFaultSource:
		return "Client fault"
	case ServerFaultSource:
		return "Server fault"
	default:
		return "Unspecified fault"
	}
}

// Slug returns the lowercase string identifier differentiating the fault source.
func (fs FaultSource) Slug() string {
	switch fs {
	case ClientFaultSource:
		return "client"
	case ServerFaultSource:
		return "server"
	default:
		return "unspecified_fault"
	}
}

// IsClientFault returns true if this is a client-side fault.
func (fs FaultSource) IsClientFault() bool {
	return fs == ClientFaultSource
}

// IsServerFault returns true if this is a server-side fault.
func (fs FaultSource) IsServerFault() bool {
	return fs == ServerFaultSource
}
func ParseFaultSource(source string) (fs FaultSource, err error) {
	source = strings.ToLower(source)
	switch source {
	case "client":
		fs = ClientFaultSource
	case "server":
		fs = ServerFaultSource
	default:
		source = strings.TrimSpace(source)
		switch {
		case strings.HasPrefix(source, "client"):
			fs = ClientFaultSource
		case strings.HasPrefix(source, "server"):
			fs = ServerFaultSource
		default:
			fs = UnspecifiedFaultSource
			err = NewErr(ErrInvalidFaultSource,
				"fault_source", source,
			)
		}
	}
	return fs, err
}
