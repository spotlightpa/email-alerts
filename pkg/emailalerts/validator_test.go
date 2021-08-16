package emailalerts_test

import (
	"fmt"

	"github.com/spotlightpa/email-alerts/pkg/emailalerts"
)

func ExampleValidator() {
	var v emailalerts.Validator
	v.Ensure(2 < 1, "heads", "Two are better than one.")
	v.Ensure(!true, "heads", "I win, tails you lose.")
	fmt.Println(v.Err())
	// Output:
	// validation error: heads=Two are better than one. heads=I win, tails you lose.
}
