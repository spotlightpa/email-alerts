package emailalerts

import (
	"net/http"
	"strings"

	"github.com/carlmjohnson/emailx"
	"github.com/carlmjohnson/resperr"
)

func validate(email, first, last string) error {
	if email == "" {
		err := resperr.New(http.StatusBadRequest, "no address")
		err = resperr.WithUserMessage(err, "No email address provided")
		return err
	}
	if !emailx.Valid(email) {
		err := resperr.New(http.StatusBadRequest,
			"invalid email %q", email)
		err = resperr.WithUserMessagef(err,
			"Invalid email address provided: %q.", email)
		return err
	}
	if strings.Contains(first, "://") {
		return resperr.New(http.StatusBadRequest,
			"invalid first name: %q", first)
	}
	if strings.Contains(last, "://") {
		return resperr.New(http.StatusBadRequest,
			"invalid last name: %q", first)
	}

	return nil
}
