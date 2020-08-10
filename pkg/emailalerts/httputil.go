package emailalerts

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/carlmjohnson/resperr"
	"github.com/getsentry/sentry-go"
)

func (app *appEnv) versionMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("SpotlightPA-App-Version", BuildVersion)
		h.ServeHTTP(w, r)
	})
}

func (app *appEnv) replyJSON(w http.ResponseWriter, r *http.Request, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	enc := json.NewEncoder(w)
	if err := enc.Encode(data); err != nil {
		app.logErr(r.Context(), err)
	}
}

func (app *appEnv) replyErr(w http.ResponseWriter, r *http.Request, err error) {
	app.logErr(r.Context(), err)
	code, errResp := errorResponseFrom(err)
	app.replyJSON(w, r, code, errResp)
}

func (app *appEnv) logErr(ctx context.Context, err error) {
	if hub := sentry.GetHubFromContext(ctx); hub != nil {
		hub.CaptureException(err)
	} else {
		app.Printf("sentry not in context")
	}

	app.Printf("err: %v", err)
}

func errorResponseFrom(err error) (status int, data interface{}) {
	code := resperr.StatusCode(err)
	msg := resperr.UserMessage(err)
	return code, struct {
		Error string `json:"error"`
	}{
		msg,
	}
}
