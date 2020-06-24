package emailalerts

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strconv"
	"strings"

	"github.com/carlmjohnson/resperr"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

func (app *appEnv) routes() http.Handler {
	r := chi.NewRouter()
	if app.isLambda() {
		r.Use(middleware.RequestID)
	} else {
		r.Use(middleware.Recoverer)
	}
	r.Use(middleware.RealIP)
	r.Use(middleware.RequestLogger(&middleware.DefaultLogFormatter{Logger: app.l}))
	r.Use(app.versionMiddleware)
	r.Get("/api/healthcheck", app.ping)
	r.Get(`/api/healthcheck/{code:\d{3}}`, app.pingErr)
	r.Post(`/api/add-contact`, app.postAddContact)
	if app.isLambda() {
		r.NotFound(app.notFound)
	} else {
		r.NotFound(http.FileServer(http.Dir("./public")).ServeHTTP)
	}
	return r
}

// common errors
var (
	errBadRequest = resperr.WithStatusCode(nil, http.StatusBadRequest)
)

func (app *appEnv) notFound(w http.ResponseWriter, r *http.Request) {
	app.errorResponse(r.Context(), w, resperr.NotFound(r))
}

func (app *appEnv) ping(w http.ResponseWriter, r *http.Request) {
	app.Printf("start ping")
	w.Header().Set("Content-Type", "text/plain")
	w.Header().Set("Cache-Control", "public, max-age=60")
	b, err := httputil.DumpRequest(r, true)
	if err != nil {
		app.errorResponse(r.Context(), w, err)
		return
	}

	w.Write(b)
}

var errPing = fmt.Errorf("test ping")

func (app *appEnv) pingErr(w http.ResponseWriter, r *http.Request) {
	code := chi.URLParam(r, "code")
	statusCode, _ := strconv.Atoi(code)
	app.Printf("start pingErr %q", code)

	app.errorResponse(r.Context(), w, resperr.WithStatusCode(errPing, statusCode))
}

func (app *appEnv) postAddContact(w http.ResponseWriter, r *http.Request) {
	app.Printf("start postAddContact")
	email := r.FormValue("email")
	first := r.FormValue("first_name")
	last := r.FormValue("last_name")
	fipsCodes := r.PostForm["fips"]

	if err := app.addContact(r.Context(), first, last, email, fipsCodes); err != nil {
		app.logErr(r.Context(), err)

		sorryURL := validateRedirect(r.FormValue("redirect_sorry"), "/sorry.html")
		http.Redirect(w, r, sorryURL, http.StatusSeeOther)
		return
	}
	// Allow redirects to .spotlightpa.org
	dest := validateRedirect(r.FormValue("redirect"), "/thanks.html")
	http.Redirect(w, r, dest, http.StatusSeeOther)
}

func validateRedirect(formVal, fallback string) string {
	if u, err := url.Parse(formVal); err == nil {
		if u.Scheme == "https" && strings.HasSuffix(u.Host, ".spotlightpa.org") {
			return formVal
		}
	}
	return fallback
}
