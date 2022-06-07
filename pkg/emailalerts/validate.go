package emailalerts

import (
	"strings"

	"github.com/carlmjohnson/emailx"
	"github.com/carlmjohnson/resperr"
)

var forbiddenNames = []string{"://"}

func validate(email, first, last string) error {
	var v resperr.Validator
	v.Ensure("EMAIL", email != "", "No email address provided.")
	v.EnsureIf("EMAIL", emailx.Valid(email),
		"Invalid email address provided: %q.", email)
	for _, s := range forbiddenNames {
		v.Ensure("FNAME", !strings.Contains(first, s),
			"First name contains invalid characters %q", s)
		v.Ensure("LNAME", !strings.Contains(last, s),
			"Last name contains invalid characters %q", s)
	}
	return v.Err()
}
