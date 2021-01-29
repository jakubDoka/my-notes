package core

import "fmt"

// Err is standard error with nesting capability
type Err struct {
	message string
	args    []interface{}
	err     error
}

// NErr ...
func NErr(message string) Err {
	return Err{message: message}
}

// Args sets error args that should be used when formatting
func (e Err) Args(args ...interface{}) error {
	e.args = args
	return e
}

// Wrap wraps error into caller and returns copy
func (e Err) Wrap(err error) error {
	e.err = err
	return e
}

//Unwrap unwraps the error if it is holding any
func (e Err) Unwrap() error {
	return e.err
}

func (e Err) Error() string {
	message := fmt.Sprintf(e.message, e.args...)
	if e.err == nil {
		return message
	}
	return message + ": " + e.err.Error()
}

// Is ...
func (e Err) Is(err error) bool {
	if val, ok := err.(Err); ok && val.message == e.message {
		if e.err == nil && val.err == nil {
			return true
		}

		if e, ok := e.err.(Err); ok {
			return e.Is(val.err)
		}
	}

	return false
}
