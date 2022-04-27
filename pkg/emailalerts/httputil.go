package emailalerts

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/carlmjohnson/resperr"
	"github.com/carlmjohnson/versioninfo"
	"github.com/getsentry/sentry-go"
)

func (app *appEnv) versionMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("SpotlightPA-App-Version", versioninfo.Revision)
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

func (app *appEnv) redirectErr(w http.ResponseWriter, r *http.Request, err error) {
	app.logErr(r.Context(), err)

	sorryURL := validateRedirect(r.FormValue("redirect_sorry"), "/sorry.html")
	code := resperr.StatusCode(err)
	msg := resperr.UserMessage(err)
	sorryURL = fmt.Sprintf("%s?code=%d&msg=%s",
		sorryURL, code, url.QueryEscape(msg))
	if ve := ValidationErrors(err); len(ve) > 0 {
		b, _ := json.Marshal(ve)
		sorryURL += fmt.Sprintf("&errors=%s", url.QueryEscape(string(b)))
	}
	http.Redirect(w, r, sorryURL, http.StatusSeeOther)
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
