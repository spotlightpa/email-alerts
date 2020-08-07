package emailalerts

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strconv"
	"strings"

	"github.com/carlmjohnson/resperr"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/cors"
	"github.com/spotlightpa/email-alerts/pkg/httpjson"
)

func (app *appEnv) routes() http.Handler {
	r := chi.NewRouter()
	origin := "https://*.spotlightpa.org"
	if app.isLambda() {
		r.Use(middleware.RequestID)
	} else {
		r.Use(middleware.Recoverer)
		origin = "*"
	}
	r.Use(middleware.RealIP)
	r.Use(middleware.RequestLogger(&middleware.DefaultLogFormatter{Logger: app.l}))
	r.Use(app.versionMiddleware)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins: []string{origin},
		AllowedHeaders: []string{"*"},
		AllowedMethods: []string{http.MethodGet, http.MethodPost},
		MaxAge:         300,
	}))
	r.Get("/api/healthcheck", app.ping)
	r.Get(`/api/healthcheck/{code:\d{3}}`, app.pingErr)
	r.Post(`/api/add-contact`, app.postAddContact)
	r.Get(`/api/list-subscriptions/{email}`, app.getListSubs)
	r.Post(`/api/update-subscriptions`, app.postUpdateSubs)
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
	app.replyErr(w, r, resperr.NotFound(r))
}

func (app *appEnv) ping(w http.ResponseWriter, r *http.Request) {
	app.Printf("start ping")
	w.Header().Set("Content-Type", "text/plain")
	w.Header().Set("Cache-Control", "public, max-age=60")
	b, err := httputil.DumpRequest(r, true)
	if err != nil {
		app.replyErr(w, r, err)
		return
	}

	w.Write(b)
}

var errPing = fmt.Errorf("test ping")

func (app *appEnv) pingErr(w http.ResponseWriter, r *http.Request) {
	code := chi.URLParam(r, "code")
	statusCode, _ := strconv.Atoi(code)
	app.Printf("start pingErr %q", code)

	app.replyErr(w, r, resperr.WithStatusCode(errPing, statusCode))
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

func (app *appEnv) getListSubs(w http.ResponseWriter, r *http.Request) {
	app.Printf("start getListSubs")
	emailB64 := chi.URLParamFromCtx(r.Context(), "email")
	if emailB64 == "" {
		app.replyErr(w, r, resperr.New(
			http.StatusBadRequest, "no parameters supplied",
		))
		return
	}
	emailBytes, err := base64.URLEncoding.DecodeString(emailB64)
	if err != nil {
		app.replyErr(w, r, resperr.New(
			http.StatusBadRequest, "could not decode request",
		))
		return
	}
	user, err := app.listSubscriptions(r.Context(), string(emailBytes))
	if err != nil {
		app.replyErr(w, r, err)
		return
	}
	app.replyJSON(w, r, http.StatusOK, user)
}

func (app *appEnv) postUpdateSubs(w http.ResponseWriter, r *http.Request) {
	app.Printf("start postUpdateSubs")

	var userData contactData
	if err := httpjson.DecodeRequest(w, r, &userData); err != nil {
		app.replyErr(w, r, resperr.New(
			http.StatusBadRequest, "could not decode request: %w", err,
		))
		return
	}
	if userData.Email == "" {
		app.replyErr(w, r, resperr.New(
			http.StatusBadRequest, "no email provided",
		))
		return
	}
	if err := app.updateSubscriptions(
		r.Context(), userData.FirstName, userData.LastName, userData.Email, userData.FIPSCodes,
	); err != nil {
		app.replyErr(w, r, err)
		return
	}
	app.replyJSON(w, r, http.StatusOK, "OK")
}
