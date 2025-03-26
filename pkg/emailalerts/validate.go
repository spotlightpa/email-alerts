package emailalerts

import (
	"strings"

	"github.com/earthboundkid/emailx/v2"
	"github.com/earthboundkid/resperr/v2"
)

var forbiddenNames = []string{`://`}

func validate(email, first, last string) error {
	var v resperr.Validator
	v.AddIf("EMAIL", email == "", "No email address provided.")
	v.AddIfUnset("EMAIL", !emailx.Valid(email),
		"Invalid email address provided: %q.", email)
	for _, s := range forbiddenNames {
		v.AddIf("FNAME", strings.Contains(first, s),
			"First name contains invalid characters %q", s)
		v.AddIf("LNAME", strings.Contains(last, s),
			"Last name contains invalid characters %q", s)
	}
	return v.Err()
}
