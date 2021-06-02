package emailalerts

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"

	"github.com/carlmjohnson/requests"
	"github.com/spotlightpa/email-alerts/pkg/mailchimp"
)

func fixIP(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.RemoteAddr = "N/A"
		h.ServeHTTP(w, r)
	})
}

func TestEndToEnd(t *testing.T) {
	apikey := os.Getenv("TEST_MC_API_KEY")
	listid := os.Getenv("TEST_LIST_ID")
	cl := http.Client{
		Transport: requests.Replay("testdata"),
	}
	if os.Getenv("TEST_RECORD") != "" {
		cl.Transport = requests.Record(nil, "testdata")
	}

	mc := mailchimp.NewV3(apikey, listid, &cl)
	app := appEnv{
		l:  log.Default(),
		mc: mc,
	}

	srv := httptest.NewServer(fixIP(app.routes()))
	defer srv.Close()

	err := requests.URL(srv.URL).
		Client(&http.Client{
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
		}).
		Path("/api/subscribe").
		BodyForm(url.Values{
			"EMAIL":        []string{"cjohnson@spotlightpa.org"},
			"FNAME":        []string{"Carl"},
			"LNAME":        []string{"Johnson"},
			"investigator": []string{"1"},
		}).
		CheckStatus(http.StatusSeeOther).
		AddValidator(func(res *http.Response) error {
			if u, err := res.Location(); err != nil {
				return err
			} else if u.Path != "/thanks.html" {
				return fmt.Errorf("bad redirect: %v", u)
			}
			return nil
		}).
		Fetch(context.Background())
	if err != nil {
		t.Fatal(err)
	}
}
