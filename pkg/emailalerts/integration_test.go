package emailalerts

import (
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/carlmjohnson/be"
	"github.com/carlmjohnson/requests"
	"github.com/carlmjohnson/requests/reqtest"
	"github.com/spotlightpa/email-alerts/pkg/activecampaign"
	"github.com/spotlightpa/email-alerts/pkg/kickbox"
	"github.com/spotlightpa/email-alerts/pkg/maxmind"
)

func fixIP(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.RemoteAddr = "N/A"
		h.ServeHTTP(w, r)
	})
}

func TestEndToEnd(t *testing.T) {
	cl := http.Client{}
	cl.Transport = reqtest.Replay("testdata")
	app := appEnv{
		l:             log.Default(),
		kb:            kickbox.New("", log.Default(), &cl),
		ac:            activecampaign.New("", "", &cl),
		maxcl:         maxmind.New("", "", &cl, log.Default()),
		signingSecret: "abc123",
	}

	srv := httptest.NewServer(fixIP(app.routes()))
	defer srv.Close()

	var token string
	rb := requests.New(reqtest.Server(srv))

	{ // Should fail if no JSON
		err := rb.Clone().
			Path("/api/verify-subscribe").
			BodyJSON(nil).
			CheckStatus(400).
			Fetch(t.Context())
		be.NilErr(t, err)
	}
	{ // Should fail if no token
		var res any
		err := rb.Clone().
			Path("/api/verify-subscribe").
			BodyJSON(map[string]any{
				"EMAIL": "x@y.com",
			}).
			CheckStatus(400).
			ToJSON(&res).
			Fetch(t.Context())
		be.NilErr(t, err)
	}
	{ // Get a token
		var data struct {
			Data string
		}
		err := rb.Clone().
			Path("/api/token").
			ToJSON(&data).
			Fetch(t.Context())
		be.NilErr(t, err)
		be.In(t, ".", data.Data)
		token = data.Data
	}
	{ // Use token for request
		var res any
		err := rb.Clone().
			Path("/api/verify-subscribe").
			BodyJSON(map[string]any{
				"EMAIL": "x@y.com",
				"token": token,
			}).
			CheckStatus(200).
			ToJSON(&res).
			Fetch(t.Context())
		be.NilErr(t, err)
	}
	{ // Should fail if no JSON
		err := rb.Clone().
			Path("/api/list-add").
			BodyJSON(nil).
			CheckStatus(400).
			Fetch(t.Context())
		be.NilErr(t, err)
	}
	{ // Should fail if no signature
		var res any
		err := rb.Clone().
			Path("/api/list-add").
			BodyJSON("abc.123").
			CheckStatus(400).
			ToJSON(&res).
			Fetch(t.Context())
		be.NilErr(t, err)
	}
	{ // fail future messages
		msg := Message{CreatedAt: time.Now().AddDate(1, 0, 0)}
		msg.Encode(ListAdd{
			EmailAddress: "john@doe.org",
		})
		signedMsg := app.signMessage(msg)
		var res any
		err := rb.Clone().
			Path("/api/list-add").
			BodyJSON(signedMsg).
			CheckStatus(400).
			ToJSON(&res).
			Fetch(t.Context())
		be.NilErr(t, err)
	}
	{ // fail old messages
		msg := Message{CreatedAt: time.Now().AddDate(0, 0, -1)}
		msg.Encode(ListAdd{
			EmailAddress: "john@doe.org",
		})
		signedMsg := app.signMessage(msg)
		var res any
		err := rb.Clone().
			Path("/api/list-add").
			BodyJSON(signedMsg).
			CheckStatus(400).
			ToJSON(&res).
			Fetch(t.Context())
		be.NilErr(t, err)
	}
	{ // accept a signed, current message
		msg := Message{CreatedAt: time.Now()}
		msg.Encode(ListAdd{
			EmailAddress: "john@doe.org",
		})
		signedMsg := app.signMessage(msg)
		var res any
		err := rb.Clone().
			Path("/api/list-add").
			BodyJSON(signedMsg).
			CheckStatus(200).
			ToJSON(&res).
			Fetch(t.Context())
		be.NilErr(t, err)
	}
}
