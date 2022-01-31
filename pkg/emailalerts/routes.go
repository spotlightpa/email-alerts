package emailalerts

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strconv"
	"strings"

	"github.com/carlmjohnson/emailx"
	"github.com/carlmjohnson/resperr"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/gorilla/schema"
	"github.com/spotlightpa/email-alerts/pkg/mailchimp"
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
	r.Post(`/api/subscribe`, app.postSubscribeMailchimp)
	if app.isLambda() {
		r.NotFound(app.notFound)
	} else {
		r.NotFound(http.FileServer(http.Dir("./public")).ServeHTTP)
	}
	return r
}

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

func validateRedirect(formVal, fallback string) string {
	if u, err := url.Parse(formVal); err == nil {
		if u.Scheme == "https" && strings.HasSuffix(u.Host, ".spotlightpa.org") {
			return formVal
		}
	}
	return fallback
}

func (app *appEnv) postSubscribeMailchimp(w http.ResponseWriter, r *http.Request) {
	app.Printf("start postSubscribeMailchimp")

	if err := r.ParseForm(); err != nil {
		err = resperr.New(http.StatusBadRequest,
			"could not parse request: %w", err)
		app.redirectErr(w, r, err)
		return
	}

	decoder := schema.NewDecoder()
	decoder.IgnoreUnknownKeys(true)
	var req struct {
		EmailAddress string `schema:"EMAIL"`
		FirstName    string `schema:"FNAME"`
		LastName     string `schema:"LNAME"`
		Investigator bool   `schema:"investigator"`
		PAPost       bool   `schema:"papost"`
		BreakingNews bool   `schema:"breaking_news"`
		PALocal      bool   `schema:"palocal"`
		Honeypot     bool   `schema:"contact"`
	}
	if err := decoder.Decode(&req, r.PostForm); err != nil {
		app.redirectErr(w, r, err)
		return
	}
	if err := validate(req.EmailAddress, req.FirstName, req.LastName); err != nil {
		app.redirectErr(w, r, err)
		return
	}
	if req.Honeypot {
		err := resperr.New(http.StatusBadRequest,
			"checked honeypot %q", req.EmailAddress)
		err = resperr.WithUserMessage(err,
			"There was a problem with your request")
		app.redirectErr(w, r, err)
		return
	}
	mergeFields := map[string]string{
		"FNAME": strings.TrimSpace(req.FirstName),
		"LNAME": strings.TrimSpace(req.LastName),
	}
	removeBlank(mergeFields)

	interests := map[string]bool{
		"1839fa2e3f": req.Investigator,
		"eda85eb7dd": req.PAPost,
		"39b11b47d6": req.BreakingNews,
		"022f8229cc": req.PALocal,
	}
	removeFalse(interests)

	if err := app.mc.PutUser(r.Context(), &mailchimp.PutUserRequest{
		EmailAddress: emailx.Normalize(req.EmailAddress),
		StatusIfNew:  "subscribed",
		Status:       "subscribed",
		MergeFields:  mergeFields,
		Interests:    interests,
		IPOpt:        r.RemoteAddr,
	}); err != nil {
		app.redirectErr(w, r, err)
		return
	}
	dest := validateRedirect(r.FormValue("redirect"), "/thanks.html")
	http.Redirect(w, r, dest, http.StatusSeeOther)
}
