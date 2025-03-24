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
	"time"

	"github.com/carlmjohnson/be"
	"github.com/carlmjohnson/requests"
	"github.com/carlmjohnson/requests/reqtest"
	"github.com/spotlightpa/email-alerts/pkg/activecampaign"
	"github.com/spotlightpa/email-alerts/pkg/kickbox"
)

func fixIP(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.RemoteAddr = "N/A"
		h.ServeHTTP(w, r)
	})
}

func TestEndToEndOld(t *testing.T) {
	t.Skip("TODO, fixme?")
	cl := http.Client{
		Transport: reqtest.Replay("testdata"),
	}
	if os.Getenv("TEST_RECORD") != "" {
		cl.Transport = reqtest.Record(nil, "testdata")
	}

	app := appEnv{
		l:  log.Default(),
		kb: kickbox.New("", log.Default()),
		ac: activecampaign.New("", "", nil),
	}

	srv := httptest.NewServer(fixIP(app.routes()))
	defer srv.Close()

	srv.Client().CheckRedirect = requests.NoFollow

	err := requests.
		New(reqtest.Server(srv)).
		Path("/api/subscribe-v2").
		BodyForm(url.Values{
			"EMAIL":                []string{"cjohnson@spotlightpa.org"},
			"FNAME":                []string{"Carlana"},
			"LNAME":                []string{"Johnson"},
			"investigator":         []string{"1"},
			"shibboleth":           []string{"!skcoR AP"},
			"shibboleth_timestamp": []string{time.Now().Add(-15 * time.Minute).Format(time.RFC3339)},
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
		New(reqtest.Server(srv)).
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

func TestEndToEnd(t *testing.T) {
	app := appEnv{
		l: log.Default(),
	}

	srv := httptest.NewServer(fixIP(app.routes()))
	defer srv.Close()

	var data struct {
		Data string
	}
	err := requests.
		New(reqtest.Server(srv)).
		Path("/api/token").
		ToJSON(&data).
		Fetch(t.Context())
	be.NilErr(t, err)
	be.In(t, ".", data.Data)
}
