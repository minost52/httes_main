package types

import "fmt"

// Constants for custom error types and reasons
const (
	// Types
	ErrorProxy          = "proxyError"
	ErrorConn           = "connectionError"
	ErrorUnkown         = "unknownError"
	ErrorIntented       = "intentedError" // Errors for created intentionally
	ErrorDns            = "dnsError"
	ErrorParse          = "parseError"
	ErrorAddr           = "addressError"
	ErrorInvalidRequest = "invalidRequestError"

	// Reasons
	ReasonProxyFailed  = "proxy connection refused"
	ReasonProxyTimeout = "proxy timeout"
	ReasonConnTimeout  = "connection timeout"
	ReasonReadTimeout  = "read timeout"
	ReasonConnRefused  = "connection refused"

	// In gracefully stop, engine cancels the ongoing requests.
	// We can detect the canceled requests with the help of this.
	ReasonCtxCanceled = "context canceled"
)

// RequestError is our custom error struct created in the requester.Requester implementations.
type RequestError struct {
	Type   string
	Reason string
}

// Custom error message method of ScenarioError
func (e *RequestError) Error() string {
	return fmt.Sprintf("%s: %s", e.Type, e.Reason)
}

type ScenarioValidationError struct { // UnWrappable
	msg        string
	wrappedErr error
}

func (sc ScenarioValidationError) Error() string {
	return sc.msg
}

func (sc ScenarioValidationError) Unwrap() error {
	return sc.wrappedErr
}

type EnvironmentNotDefinedError struct { // UnWrappable
	msg        string
	wrappedErr error
}

func (sc EnvironmentNotDefinedError) Error() string {
	return sc.msg
}

func (sc EnvironmentNotDefinedError) Unwrap() error {
	return sc.wrappedErr
}

type CaptureConfigError struct { // UnWrappable
	msg        string
	wrappedErr error
}

func (sc CaptureConfigError) Error() string {
	return sc.msg
}

func (sc CaptureConfigError) Unwrap() error {
	return sc.wrappedErr
}
