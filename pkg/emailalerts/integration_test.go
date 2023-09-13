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

	"github.com/carlmjohnson/be"
	"github.com/carlmjohnson/requests"
	"github.com/spotlightpa/email-alerts/pkg/kickbox"
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
		kb: kickbox.New("", log.Default()),
	}

	srv := httptest.NewServer(fixIP(app.routes()))
	defer srv.Close()

	srv.Client().CheckRedirect = requests.NoFollow

	err := requests.
		New(requests.TestServerConfig(srv)).
		Path("/api/subscribe").
		BodyForm(url.Values{
			"EMAIL":        []string{"cjohnson@spotlightpa.org"},
			"FNAME":        []string{"Carl"},
			"LNAME":        []string{"Johnson"},
			"investigator": []string{"1"},
			"shibboleth":   []string{"PA Rocks!"},
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
	be.NilErr(t, err)

	err = requests.
		New(requests.TestServerConfig(srv)).
		Path("/api/subscribe").
		BodyForm(url.Values{
			"EMAIL":        []string{""},
			"FNAME":        []string{"http://buynow.com"},
			"LNAME":        []string{"https://viagra.com"},
			"investigator": []string{"1"},
		}).
		CheckStatus(http.StatusSeeOther).
		AddValidator(func(res *http.Response) error {
			if u, err := res.Location(); err != nil {
				return err
			} else if u.Path != "/sorry.html" ||
				u.RawQuery != "code=400&msg=Bad+Request&errors=%7B%22EMAIL%22%3A%5B%22No+email+address+provided.%22%5D%2C%22FNAME%22%3A%5B%22First+name+contains+invalid+characters+%5C%22%3A%2F%2F%5C%22%22%5D%2C%22LNAME%22%3A%5B%22Last+name+contains+invalid+characters+%5C%22%3A%2F%2F%5C%22%22%5D%7D" {
				return fmt.Errorf("bad redirect: %v", u)
			}
			return nil
		}).
		Fetch(context.Background())
	be.NilErr(t, err)
}
