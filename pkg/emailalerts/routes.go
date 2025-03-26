package emailalerts

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/earthboundkid/mid"
	sentryhttp "github.com/getsentry/sentry-go/http"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jub0bs/cors"
)

func (app *appEnv) routes() http.Handler {
	srv := http.NewServeMux()
	srv.HandleFunc("GET /api/healthcheck", app.ping)
	srv.HandleFunc("POST /api/subscribe-v2", app.postSubscribeActiveCampaign)
	srv.Handle("GET /api/token", mid.Controller(app.getToken))
	srv.Handle("POST /api/subscribe-v3", mid.Controller(app.postSubscribeJSON))
	if app.isLambda() {
		srv.HandleFunc("/", app.notFound)
	} else {
		srv.Handle("/", http.FileServerFS(os.DirFS("./public")))
	}

	var stack mid.Stack
	stack.Push(sentryhttp.
		New(sentryhttp.Options{
			WaitForDelivery: true,
			Timeout:         5 * time.Second,
			Repanic:         !app.isLambda(),
		}).
		Handle)
	stack.PushIf(app.isLambda(), middleware.RequestID)
	stack.PushIf(!app.isLambda(), middleware.Recoverer)
	stack.Push(middleware.RealIP)
	stack.Push(middleware.RequestLogger(&middleware.DefaultLogFormatter{Logger: app.l}))
	stack.Push(timeoutMiddleware(9 * time.Second))
	stack.Push(app.versionMiddleware)
	stack.Push(must(cors.NewMiddleware(cors.Config{
		Origins:         []string{"*"},
		Methods:         []string{http.MethodGet, http.MethodPost},
		MaxAgeInSeconds: int((5 * time.Minute).Seconds()),
	})).Wrap)

	return stack.Handler(srv)
}

func (app *appEnv) notFound(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusNotFound)
	w.Write([]byte("HTTP Status 404: Not Found"))
}

func (app *appEnv) ping(w http.ResponseWriter, r *http.Request) {
	app.Printf("start ping")
	w.Header().Set("Content-Type", "text/plain")
	w.Header().Set("Cache-Control", "public, max-age=60")
	b, err := httputil.DumpRequest(r, true)
	if err != nil {
		app.replyErr(err).ServeHTTP(w, r)
		return
	}

	w.Write(b)
}

func validateRedirect(formVal, fallback string) string {
	if u, err := url.Parse(formVal); err == nil {
		if u.Scheme == "https" && strings.HasSuffix(u.Host, ".spotlightpa.org") {
			return formVal
		}
	}
	return fallback
}

func (app *appEnv) getToken(w http.ResponseWriter, r *http.Request) http.Handler {
	// TODO: Check IP first
	return app.replyJSON(app.createToken(time.Now()))
}
