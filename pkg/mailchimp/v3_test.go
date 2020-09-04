package mailchimp_test

import (
	"context"
	"os"
	"testing"

	"github.com/spotlightpa/email-alerts/pkg/mailchimp"
)

func skipIfBlank(t *testing.T, msg string, ss ...string) {
	t.Helper()
	for _, s := range ss {
		if s == "" {
			t.Skip(msg)
		}
	}
}

func TestV3(t *testing.T) {
	apiKey := os.Getenv("MC_TEST_API_KEY")
	listID := os.Getenv("MC_TEST_LISTID")
	email := os.Getenv("MC_TEST_EMAIL")
	interest := os.Getenv("MC_TEST_INTEREST")

	skipIfBlank(t, "Missing MailChimp ENV vars",
		apiKey, listID, email, interest)

	ctx := context.Background()
	v3 := mailchimp.NewV3(apiKey, listID, nil)
	req := mailchimp.PutUserRequest{
		EmailAddress: email,
		StatusIfNew:  "subscribed",
		Interests: map[string]bool{
			interest: true,
		},
	}
	if err := v3.PutUser(ctx, &req); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
