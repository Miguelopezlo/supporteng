package respond

import "fmt"

var (
	NotFoundErr  = RequestError{404, "Not found"}
	ForbiddenErr = RequestError{403, "Forbidden"}
)

// Error is an error to return in an http response
type Error interface {
	StatusCode() int
	Error() string
}

type RequestError struct {
	Status  int    `json:"-"`
	Message string `json:"message"`
}

func (e RequestError) Error() string {
	return fmt.Sprintf("HTTP %d: %s", e.Status, e.Message)
}

func (e RequestError) StatusCode() int {
	return e.Status
}

func NewRequestError(status int, message string) error {
	return &RequestError{status, message}
}

var _ Error = RequestError{}
