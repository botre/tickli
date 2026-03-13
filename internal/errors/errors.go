package errors

// NotFoundError indicates that a requested resource was not found.
type NotFoundError struct {
	Message string
}

func (e *NotFoundError) Error() string {
	return e.Message
}

// UsageError indicates invalid command-line usage.
type UsageError struct {
	Message string
}

func (e *UsageError) Error() string {
	return e.Message
}

// AuthError indicates an authentication or authorization failure.
type AuthError struct {
	Message string
}

func (e *AuthError) Error() string {
	return e.Message
}
