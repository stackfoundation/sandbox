package errors

import "bytes"

// CompositeError An error that is a compsoite of others
type CompositeError struct {
	errors []error
}

func (e *CompositeError) Error() string {
	if e.errors != nil {
		var text bytes.Buffer

		for _, err := range e.errors {
			if err != nil {
				text.WriteString(err.Error())
				text.WriteString("\n")
			}
		}

		return text.String()
	}
	return ""
}

// NewCompositeError Create a new empty composite error
func NewCompositeError() *CompositeError {
	err := &CompositeError{
		errors: make([]error, 0, 2),
	}

	return err
}

// Append Append a new error onto this composite error
func (e *CompositeError) Append(err error) {
	if err != nil {
		e.errors = append(e.errors, err)
	}
}

// OrNilIfEmpty Return the composite error itself, or nil if it contains no errors
func (e *CompositeError) OrNilIfEmpty() error {
	if len(e.errors) > 0 {
		return e
	}

	return nil
}
