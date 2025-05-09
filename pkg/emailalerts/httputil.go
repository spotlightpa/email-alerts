package emailalerts

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"
	"strings"
	"time"

	"github.com/earthboundkid/mid"
	"github.com/earthboundkid/resperr/v2"
	"github.com/earthboundkid/versioninfo/v2"
	"github.com/getsentry/sentry-go"
)

func (app *appEnv) logReq(req *http.Request, res *http.Response, err error, duration time.Duration) {
	if err == nil {
		app.l.Printf("req.host=%q res.code=%d res.duration=%v",
			req.URL.Hostname(), res.StatusCode, duration)
	} else {
		app.l.Printf("req.host=%q err=%v res.duration=%v",
			req.URL.Hostname(), err, duration)
	}
}

func (app *appEnv) versionMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("SpotlightPA-App-Version", versioninfo.Revision)
		h.ServeHTTP(w, r)
	})
}

type jsonData struct {
	StatusCode int                 `json:"statuscode"`
	Status     string              `json:"status"`
	Data       any                 `json:"data,omitzero"`
	Error      string              `json:"error,omitzero"`
	Details    map[string][]string `json:"details,omitzero"`
}

func (app *appEnv) writeJSON(w http.ResponseWriter, r *http.Request, obj jsonData) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(obj.StatusCode)
	obj.Status = http.StatusText(obj.StatusCode)
	enc := json.NewEncoder(w)
	if err := enc.Encode(obj); err != nil {
		app.logErr(r.Context(), err)
	}
}

func (app *appEnv) replyJSON(data any) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		app.writeJSON(w, r, jsonData{
			Data:       data,
			StatusCode: http.StatusOK,
		})
	})
}

func (app *appEnv) replyErr(err error) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		app.logErr(r.Context(), err)
		app.writeJSON(w, r, jsonData{
			StatusCode: resperr.StatusCode(err),
			Error:      resperr.UserMessage(err),
			Details:    resperr.ValidationErrors(err),
		})
	})
}

func (app *appEnv) logErr(ctx context.Context, err error) {
	if hub := sentry.GetHubFromContext(ctx); hub != nil {
		hub.CaptureException(err)
	} else {
		app.Printf("sentry not in context")
	}

	app.Printf("err: %v", err)
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}

func try[T any](val T, err error) T {
	must(err)
	return val
}

func (app *appEnv) createToken(now time.Time) string {
	msg := Message{CreatedAt: now}
	return app.signMessage(msg)
}

func (app *appEnv) verifyToken(now time.Time, token string) bool {
	msg := app.unpackMessage(token)
	return msg != nil && msg.Body == "" && msg.ValidAt(now)
}

func timeoutMiddleware(timeout time.Duration) mid.Middleware {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx, stop := context.WithTimeout(r.Context(), timeout)
			defer stop()
			h.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func (app *appEnv) readJSON(r *http.Request, dst any) error {
	// Thanks to https://www.alexedwards.net/blog/how-to-properly-parse-a-json-request-body
	if ct := r.Header.Get("Content-Type"); ct != "" {
		value, _, _ := mime.ParseMediaType(ct)
		if value != "application/json" {
			return resperr.New(http.StatusUnsupportedMediaType,
				"request Content-Type must be application/json; got %s",
				ct)
		}
	}

	dec := json.NewDecoder(r.Body)
	err := dec.Decode(&dst)
	if err != nil {
		var (
			syntaxError        *json.SyntaxError
			unmarshalTypeError *json.UnmarshalTypeError
			maxBytesError      *http.MaxBytesError
		)

		switch {
		case errors.As(err, &syntaxError):
			return resperr.E{E: err, M: fmt.Sprintf("Request body contains badly-formed JSON (at position %d)", syntaxError.Offset)}

		case errors.Is(err, io.ErrUnexpectedEOF):
			return resperr.E{E: err, M: "Request body contains badly-formed JSON"}

		case errors.As(err, &unmarshalTypeError):
			return resperr.E{E: err, M: fmt.Sprintf("Request body contains an invalid value for the %q field (at position %d)", unmarshalTypeError.Field, unmarshalTypeError.Offset)}

		case strings.HasPrefix(err.Error(), "json: unknown field "):
			fieldName := strings.TrimPrefix(err.Error(), "json: unknown field ")
			return resperr.E{E: err, M: fmt.Sprintf("Request body contains unknown field %s", fieldName)}

		case errors.Is(err, io.EOF):
			return resperr.E{M: "Request body contains badly-formed JSON"}

		case errors.As(err, &maxBytesError):
			return resperr.New(http.StatusRequestEntityTooLarge,
				"request body exceeds max size %d: %w",
				maxBytesError.Limit, err)

		default:
			return resperr.New(http.StatusBadRequest, "readJSON: %w", err)
		}
	}

	var discard any
	if err := dec.Decode(&discard); !errors.Is(err, io.EOF) {
		return resperr.E{M: "Request body must only contain a single JSON object"}
	}

	return nil
}
