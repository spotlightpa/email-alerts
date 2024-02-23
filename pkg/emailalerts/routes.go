package emailalerts

import (
	"maps"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/carlmjohnson/emailx"
	"github.com/carlmjohnson/resperr"
	"github.com/earthboundkid/mid"
	sentryhttp "github.com/getsentry/sentry-go/http"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/gorilla/schema"
	"github.com/spotlightpa/email-alerts/pkg/mailchimp"
)

func (app *appEnv) routes() http.Handler {
	srv := http.NewServeMux()
	srv.HandleFunc("GET /api/healthcheck", app.ping)
	srv.HandleFunc("POST /api/subscribe", app.postSubscribeMailchimp)
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
	stack.Push(app.versionMiddleware)
	origin := "https://*.spotlightpa.org"
	if !app.isLambda() {
		origin = "*"
	}
	stack.Push(cors.Handler(cors.Options{
		AllowedOrigins: []string{origin},
		AllowedHeaders: []string{"*"},
		AllowedMethods: []string{http.MethodGet, http.MethodPost},
		MaxAge:         300,
	}))

	return stack.Handler(srv)
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
		EmailAddress            string     `schema:"EMAIL"`
		FirstName               string     `schema:"FNAME"`
		LastName                string     `schema:"LNAME"`
		Investigator            bool       `schema:"investigator"`
		PAPost                  bool       `schema:"papost"`
		BreakingNews            bool       `schema:"breaking_news"`
		PALocal                 bool       `schema:"palocal"`
		BerksCounty             bool       `schema:"berks_county"`
		TalkOfTheTown           bool       `schema:"talkofthetown"` // Alias for StateCollege
		StateCollege            bool       `schema:"state_college"`
		WeekInReview            bool       `schema:"week_in_review"`
		PennStateAlerts         bool       `schema:"pennstatealert"`
		CentreCountyDocumenters bool       `schema:"centre_county_documenters"` // Alias for CentreDocumenters
		CentreDocumenters       bool       `schema:"centredocumenters"`
		Honeypot                bool       `schema:"contact"`
		Shibboleth              string     `schema:"shibboleth"`
		Timestamp               *time.Time `schema:"shibboleth_timestamp"`
	}
	if err := decoder.Decode(&req, r.PostForm); err != nil {
		app.redirectErr(w, r, err)
		return
	}
	if err := validate(req.EmailAddress, req.FirstName, req.LastName); err != nil {
		app.redirectErr(w, r, err)
		return
	}
	if req.Shibboleth != "PA Rocks!" || req.Timestamp == nil {
		err := resperr.New(http.StatusBadRequest,
			"missing shibboleth: %q", req.EmailAddress)
		err = resperr.WithUserMessage(err,
			"JavaScript is required to sign up for a mailing list.")
		app.redirectErr(w, r, err)
		return
	}
	if time.Since(*req.Timestamp).Abs() > 24*time.Hour {
		err := resperr.New(http.StatusBadRequest,
			"bad timestamp: %q", req.EmailAddress)
		err = resperr.WithUserMessage(err,
			"Page too old. Please reload the window and try again.")
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
	if !app.kb.Verify(r.Context(), req.EmailAddress) {
		err := resperr.New(http.StatusBadRequest,
			"Kickbox rejected %q", req.EmailAddress)
		err = resperr.WithUserMessage(err,
			"There was a problem with your request")
		app.redirectErr(w, r, err)
		return
	}
	mergeFields := map[string]string{
		"FNAME": strings.TrimSpace(req.FirstName),
		"LNAME": strings.TrimSpace(req.LastName),
	}
	maps.DeleteFunc(mergeFields, func(k, v string) bool {
		return v == ""
	})

	interests := map[string]bool{
		"00612929b8": req.Investigator,
		"8dbf00ee98": req.PAPost,
		"6137d9281f": req.BreakingNews,
		"84cfce88c7": req.PALocal,
		"aa8800a947": req.BerksCounty,
		"ff98baba5f": req.TalkOfTheTown ||
			req.StateCollege,
		"5c3b89e306": req.WeekInReview,
		"062c085860": req.PennStateAlerts,
		"650bf212f7": req.CentreCountyDocumenters ||
			req.CentreDocumenters,
	}
	maps.DeleteFunc(interests, func(k string, v bool) bool {
		return !v
	})

	if err := app.mc.PutUser(r.Context(), &mailchimp.PutUserRequest{
		EmailAddress: emailx.Normalize(req.EmailAddress),
		StatusIfNew:  "subscribed",
		Status:       "subscribed",
		MergeFields:  mergeFields,
		Interests:    interests,
		IPOpt:        r.Header.Get("X-Forwarded-For"),
	}); err != nil {
		app.redirectErr(w, r, err)
		return
	}

	app.l.Printf("subscribed: email=%q", req.EmailAddress)

	dest := validateRedirect(r.FormValue("redirect"), "/thanks.html")
	http.Redirect(w, r, dest, http.StatusSeeOther)
}
