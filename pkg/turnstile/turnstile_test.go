package turnstile_test

import (
	"net/http"
	"testing"

	"github.com/carlmjohnson/be"
	"github.com/carlmjohnson/requests/reqtest"
	"github.com/spotlightpa/email-alerts/pkg/turnstile"
)

func TestTurnstile(t *testing.T) {
	const testGoodKey = "1x0000000000000000000000000000000AA"
	const testBadKey = "2x0000000000000000000000000000000AA"
	tr := reqtest.Replay("testdata")
	{
		cl := turnstile.New(testGoodKey, &http.Client{Transport: tr})
		ok, err := cl.Validate(t.Context(), "abc123", "127.0.0.1")
		be.NilErr(t, err)
		be.True(t, ok)
	}
	{
		cl := turnstile.New(testBadKey, &http.Client{Transport: tr})
		ok, err := cl.Validate(t.Context(), "abc123", "127.0.0.1")
		be.NilErr(t, err)
		be.False(t, ok)
	}
}
