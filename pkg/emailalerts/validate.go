package emailalerts

import (
	"strings"

	"github.com/carlmjohnson/emailx"
)

var forbiddenNames = []string{"://"}

func validate(email, first, last string) error {
	var v Validator
	v.Ensure(email != "", "EMAIL", "No email address provided.")
	v.Ensure(email == "" || emailx.Valid(email), "EMAIL",
		"Invalid email address provided: %q.", email)
	for _, s := range forbiddenNames {
		v.Ensure(!strings.Contains(first, s), "FNAME",
			"First name contains invalid characters %q", s)
		v.Ensure(!strings.Contains(last, s), "LNAME",
			"Last name contains invalid characters %q", s)
	}
	return v.Err()
}
