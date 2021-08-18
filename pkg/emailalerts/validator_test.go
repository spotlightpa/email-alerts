package emailalerts_test

import (
	"fmt"
	"testing"

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

func TestValidator(t *testing.T) {
	var v1 emailalerts.Validator
	v1.Ensure(2 < 1, "heads", "Two are better than one.")
	v1.Ensure(!true, "heads", "I win, tails you lose.")
	err := v1.Err()
	if err == nil {
		t.Fail()
	}
	fields := emailalerts.ValidationErrors(err)
	if len(fields) != 1 {
		t.Fatalf("got len %d", len(fields))
	}
	if len(fields["heads"]) != 2 {
		t.Fatalf("got len %d", len(fields))
	}
	var v2 emailalerts.Validator
	v2.Ensure(2 > 1, "heads", "One is the loneliest number.")
	v2.Ensure(true, "heads", "I win, tails you lose.")
	err = v2.Err()
	if err != nil {
		t.Fail()
	}
	fields = emailalerts.ValidationErrors(err)
	if len(fields) != 0 {
		t.Fatalf("got len %d", len(fields))
	}
}
