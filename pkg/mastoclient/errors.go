package mastoclient

// InvalidSuspendType is returned when the provided suspend type is invalid
type InvalidSuspendType struct {
	Err          error
	typeProvided *string
	Msg          string
}

// Error returns the error message
func (e *InvalidSuspendType) Error() string {
	if e.Msg == "" {
		e.Msg = "invalid suspend type"
	}
	if e.typeProvided != nil {
		e.Msg += ": " + *e.typeProvided
	}
	if e.Err != nil {
		e.Msg += ": " + e.Err.Error()
	}
	return e.Msg
}

// NoAccessToken is returned when the Mastodon access token value is missing.
type NoAccessToken struct {
	Err error
	Msg string
}

// Error returns the error message
func (e *NoAccessToken) Error() string {
	if e.Msg == "" {
		e.Msg = "No access token. use WithAccessToken()"
	}
	if e.Err != nil {
		e.Msg += ": " + e.Err.Error()
	}
	return e.Msg
}

// NoInstance is returned when the Mastodon instance value is missing.
type NoInstance struct {
	Err error
	Msg string
}

// Error returns the error message
func (e *NoInstance) Error() string {
	if e.Msg == "" {
		e.Msg = "no instance. use WithInstance()"
	}
	if e.Err != nil {
		e.Msg += ": " + e.Err.Error()
	}
	return e.Msg
}

// PostFailed is retrned when an HTTP POST operation to Mastodon fails
type PostFailed struct {
	// Err is a proper error object
	Err error

	// Msg is a string to give addition context to the error message
	Msg string

	// Status is used to convey http status codes or other related data
	Status string
}

// Error returns the error message
func (e *PostFailed) Error() string {
	if e.Msg == "" {
		e.Msg = "post failed"
	}
	if e.Err != nil {
		e.Msg += ": " + e.Err.Error()
	}
	return e.Msg
}
