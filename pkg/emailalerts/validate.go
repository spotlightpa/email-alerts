package emailalerts

import (
	"net/http"
	"strings"

	"github.com/carlmjohnson/emailx"
	"github.com/carlmjohnson/resperr"
)

var forbiddenNames = []string{"://"}

func validate(email, first, last string) error {
	if email == "" {
		err := resperr.New(http.StatusBadRequest, "no address")
		err = resperr.WithUserMessage(err, "No email address provided")
		return err
	}
	if !emailx.Valid(email) {
		err := resperr.New(http.StatusBadRequest, "invalid email")
		err = resperr.WithUserMessagef(err,
			"Invalid email address provided: %q.", email)
		return err
	}
	for _, s := range forbiddenNames {
		if strings.Contains(first, s) || strings.Contains(last, s) {
			return resperr.New(http.StatusBadRequest, "invalid first/last name")
		}
	}

	return nil
}
