package emailalerts

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/getsentry/sentry-go"
	"golang.org/x/net/context/ctxhttp"
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

func doJSON(ctx context.Context, cl *http.Client, method, url string, data interface{}) error {
	blob, err := json.Marshal(data)
	if err != nil {
		return err
	}
	r := bytes.NewReader(blob)
	req, err := http.NewRequest(method, url, r)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	rsp, err := ctxhttp.Do(ctx, cl, req)
	if err != nil {
		return err
	}
	defer rsp.Body.Close()

	if rsp.StatusCode < 200 || rsp.StatusCode >= 300 {
		return fmt.Errorf("unexpected status: %d %s", rsp.StatusCode, rsp.Status)
	}

	// Drain connection
	_, err = io.Copy(ioutil.Discard, rsp.Body)
	return err
}

func postJSON(ctx context.Context, cl *http.Client, url string, data interface{}) error {
	return doJSON(ctx, cl, http.MethodPost, url, data)
}

func putJSON(ctx context.Context, cl *http.Client, url string, data interface{}) error {
	return doJSON(ctx, cl, http.MethodPut, url, data)
}
