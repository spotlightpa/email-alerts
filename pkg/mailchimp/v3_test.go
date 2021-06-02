package mailchimp_test

import (
	"context"
	"net/http"
	"os"
	"testing"

	"github.com/carlmjohnson/requests"
	"github.com/spotlightpa/email-alerts/pkg/mailchimp"
)

func TestV3(t *testing.T) {
	apiKey := os.Getenv("MC_TEST_API_KEY")
	listID := os.Getenv("MC_TEST_LISTID")
	email := os.Getenv("MC_TEST_EMAIL")
	interest := os.Getenv("MC_TEST_INTEREST")
	cl := http.Client{
		Transport: requests.Replay("testdata"),
	}
	if os.Getenv("TEST_RECORD") != "" {
		cl.Transport = requests.Record(nil, "testdata")
	}
	ctx := context.Background()
	v3 := mailchimp.NewV3(apiKey, listID, &cl)
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
