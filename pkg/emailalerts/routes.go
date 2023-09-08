package emailalerts

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"

	"github.com/carlmjohnson/emailx"
	"github.com/carlmjohnson/resperr"
	"github.com/carlmjohnson/rootdown"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/gorilla/schema"
	"github.com/spotlightpa/email-alerts/pkg/mailchimp"
	"golang.org/x/exp/maps"
)

func (app *appEnv) routes() http.Handler {
	var mw rootdown.MiddlewareStack
	if app.isLambda() {
		mw.Push(middleware.RequestID)
	} else {
		mw.Push(middleware.Recoverer)
	}
	mw.Push(middleware.RealIP)
	mw.Push(middleware.RequestLogger(&middleware.DefaultLogFormatter{Logger: app.l}))
	mw.Push(app.versionMiddleware)
	origin := "https://*.spotlightpa.org"
	if !app.isLambda() {
		origin = "*"
	}
	mw.Push(cors.Handler(cors.Options{
		AllowedOrigins: []string{origin},
		AllowedHeaders: []string{"*"},
		AllowedMethods: []string{http.MethodGet, http.MethodPost},
		MaxAge:         300,
	}))

	var rr rootdown.Router
	rr.Get("/api/healthcheck", app.ping, mw...)
	rr.Post("/api/subscribe", app.postSubscribeMailchimp, mw...)
	if app.isLambda() {
		rr.NotFound(app.notFound, mw...)
	} else {
		_ = rr.Mount("", "", os.DirFS("./public"), mw...)
	}
	return &rr
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
		EmailAddress    string `schema:"EMAIL"`
		FirstName       string `schema:"FNAME"`
		LastName        string `schema:"LNAME"`
		Investigator    bool   `schema:"investigator"`
		PAPost          bool   `schema:"papost"`
		BreakingNews    bool   `schema:"breaking_news"`
		PALocal         bool   `schema:"palocal"`
		BerksCounty     bool   `schema:"berks_county"`
		TalkOfTheTown   bool   `schema:"talkofthetown"` // Alias for StateCollege
		StateCollege    bool   `schema:"state_college"`
		WeekInReview    bool   `schema:"week_in_review"`
		PennStateAlerts bool   `schema:"pennstatealert"`
		Honeypot        bool   `schema:"contact"`
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
		IPOpt:        r.RemoteAddr,
	}); err != nil {
		app.redirectErr(w, r, err)
		return
	}
	if req.StateCollege {
		if err := app.mc.UserTags(
			r.Context(),
			emailx.Normalize(req.EmailAddress),
			mailchimp.AddTag,
			"state_college",
		); err != nil {
			app.redirectErr(w, r, err)
			return
		}
	}
	dest := validateRedirect(r.FormValue("redirect"), "/thanks.html")
	http.Redirect(w, r, dest, http.StatusSeeOther)
}
