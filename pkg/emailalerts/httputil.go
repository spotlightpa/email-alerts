package emailalerts

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/getsentry/sentry-go"
)

func (app *appEnv) versionMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("SpotlightPA-App-Version", BuildVersion)
		h.ServeHTTP(w, r)
	})
}

func (app *appEnv) jsonResponse(ctx context.Context, statusCode int, w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	enc := json.NewEncoder(w)
	if err := enc.Encode(data); err != nil {
		app.logErr(ctx, err)
	}
}

func (app *appEnv) errorResponse(ctx context.Context, w http.ResponseWriter, err error) {
	app.logErr(ctx, err)
	code, errResp := errorResponseFrom(err)
	app.jsonResponse(ctx, code, w, errResp)
}

func (app *appEnv) logErr(ctx context.Context, err error) {
	if hub := sentry.GetHubFromContext(ctx); hub != nil {
		hub.CaptureException(err)
	} else {
		app.Printf("sentry not in context")
	}

	app.Printf("err: %v", err)
}

type statusErr struct {
	Cause      error
	StatusCode int
}

func withStatus(code int, err error) error {
	return &statusErr{err, code}
}

func (se *statusErr) Error() string {
	return fmt.Sprintf("[%d] %v", se.StatusCode, se.Cause)
}

func (se *statusErr) Unwrap() error {
	return se.Cause
}

func errorResponseFrom(err error) (status int, data interface{}) {
	if se := new(statusErr); errors.As(err, &se) {
		return se.StatusCode, struct {
			Error string `json:"error"`
		}{
			http.StatusText(se.StatusCode),
		}
	}
	return http.StatusInternalServerError, struct {
		Error string `json:"error"`
	}{
		http.StatusText(http.StatusInternalServerError),
	}
}
