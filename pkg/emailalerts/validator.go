package emailalerts

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/carlmjohnson/resperr"
)

// Validator creates a map of fields to error messages.
type Validator url.Values

// Ensure adds the provided message to field if cond is not met.
// Ensure works with the zero value of Validator.
func (v *Validator) Ensure(cond bool, field, message string, a ...interface{}) {
	if cond {
		return
	}
	if *v == nil {
		*v = make(Validator)
	}
	(*url.Values)(v).Add(field, fmt.Sprintf(message, a...))
}

// Err transforms v to a ValidatorError if v is not empty.
// The error created shares the same underlying map reference as v.
func (v *Validator) Err() error {
	if len(*v) < 1 {
		return nil
	}
	return validatorErrors(*v)
}

// ValidationErrors returns any ValidationError found in err's error chain
// or an empty map.
func ValidationErrors(err error) url.Values {
	if ve := (ValidationError)(nil); err != nil && errors.As(err, &ve) {
		return ve.ValidationErrors()
	}
	return nil
}

type ValidationError interface {
	error
	ValidationErrors() url.Values
}

type validatorErrors url.Values

var _ ValidationError = validatorErrors{}
var _ resperr.StatusCoder = validatorErrors{}

func (ve validatorErrors) Error() string {
	s, _ := url.QueryUnescape(url.Values(ve).Encode())
	return fmt.Sprintf("validation error: %s", strings.ReplaceAll(s, "&", " "))
}

func (ve validatorErrors) ValidationErrors() url.Values {
	return url.Values(ve)
}

func (ve validatorErrors) StatusCode() int {
	return http.StatusBadRequest
}
