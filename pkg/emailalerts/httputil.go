package emailalerts

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/earthboundkid/mid"
	"github.com/earthboundkid/resperr/v2"
	"github.com/earthboundkid/versioninfo/v2"
	"github.com/getsentry/sentry-go"
)

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
	Details    map[string][]string `json:"errors,omitzero"`
}

func (app *appEnv) writeJSON(obj jsonData) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(obj.StatusCode)
		obj.Status = http.StatusText(obj.StatusCode)
		enc := json.NewEncoder(w)
		if err := enc.Encode(obj); err != nil {
			app.logErr(r.Context(), err)
		}
	})
}

func (app *appEnv) replyJSON(data any) http.Handler {
	return app.writeJSON(jsonData{
		Data:       data,
		StatusCode: http.StatusOK,
	})
}

func (app *appEnv) replyErr(err error) http.Handler {
	return app.writeJSON(jsonData{
		StatusCode: resperr.StatusCode(err),
		Error:      resperr.UserMessage(err),
		Details:    resperr.ValidationErrors(err),
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

func (app *appEnv) redirectErr(w http.ResponseWriter, r *http.Request, err error) {
	app.logErr(r.Context(), err)

	sorryURL := validateRedirect(r.FormValue("redirect_sorry"), "/sorry.html")
	code := resperr.StatusCode(err)
	msg := resperr.UserMessage(err)
	sorryURL = fmt.Sprintf("%s?code=%d&msg=%s",
		sorryURL, code, url.QueryEscape(msg))
	if ve := resperr.ValidationErrors(err); len(ve) > 0 {
		b, _ := json.Marshal(ve)
		sorryURL += fmt.Sprintf("&errors=%s", url.QueryEscape(string(b)))
	}
	http.Redirect(w, r, sorryURL, http.StatusSeeOther)
}

func must[T any](val T, err error) T {
	if err != nil {
		panic(err)
	}
	return val
}

func (app *appEnv) createToken(now time.Time) string {
	encoding := base64.URLEncoding

	// Make a timestamp
	timestamp := must(now.MarshalBinary())
	token := make([]byte, 0,
		encoding.EncodedLen(len(timestamp))+1+encoding.EncodedLen(sha256.Size))
	// Encode the timestamp
	token = base64.URLEncoding.AppendEncode(token, timestamp)

	// Sign the encoded timestamp
	mac := hmac.New(sha256.New, []byte(app.signingSecret))
	mac.Write(token)
	rawSig := mac.Sum(nil)

	// Use a dot to separate the parts
	token = append(token, '.')
	// Append the encoded signature
	token = base64.URLEncoding.AppendEncode(token, rawSig)
	return string(token)
}

const validityWindow time.Duration = 5 * time.Minute

func (app *appEnv) verifyToken(now time.Time, token string) bool {
	// Split on the dot
	encodedTimestamp, encodedSig, ok := bytes.Cut([]byte(token), []byte{'.'})
	if !ok {
		return false
	}
	encoding := base64.URLEncoding
	// Decode the signature first
	sig, err := encoding.AppendDecode(nil, encodedSig)
	if err != nil {
		return false
	}
	// Check that the signature matches the encoded timestamp
	mac := hmac.New(sha256.New, []byte(app.signingSecret))
	mac.Write(encodedTimestamp)
	expectedSig := mac.Sum(nil)
	if !hmac.Equal(sig, expectedSig) {
		return false
	}
	// Decode the timestamp
	timestamp, err := encoding.AppendDecode(nil, encodedTimestamp)
	if err != nil {
		// This should not happen because it passed the signing check
		app.logErr(context.Background(), err)
		return false
	}
	var ts time.Time
	if err := ts.UnmarshalBinary(timestamp); err != nil {
		// This should not happen because it passed the signing check
		app.logErr(context.Background(), err)
		return false
	}
	// Check that it is inside the window
	// Prevents credential reuse
	return now.After(ts) && ts.Add(validityWindow).After(now)
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
	dec.DisallowUnknownFields()

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
