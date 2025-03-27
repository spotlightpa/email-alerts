package emailalerts

import (
	"maps"
	"net/http"
	"slices"
	"strings"
	"time"

	"github.com/earthboundkid/emailx/v2"
	"github.com/earthboundkid/resperr/v2"
	"github.com/gorilla/schema"
	"github.com/spotlightpa/email-alerts/pkg/activecampaign"
)

func (app *appEnv) postSubscribeActiveCampaign(w http.ResponseWriter, r *http.Request) {
	app.Printf("start postSubscribeActiveCampaign")

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
		Events                  bool       `schema:"events"`
		Investigator            bool       `schema:"investigator"`
		PAPost                  bool       `schema:"papost"`
		BreakingNews            bool       `schema:"breaking_news"`
		PALocal                 bool       `schema:"palocal"`
		BerksCounty             bool       `schema:"berks_county"`
		Berks                   bool       `schema:"berks"`         // Alias for BerksCounty
		TalkOfTheTown           bool       `schema:"talkofthetown"` // Alias for StateCollege
		StateCollege            bool       `schema:"state_college"`
		WeekInReview            bool       `schema:"week_in_review"`
		PennStateAlerts         bool       `schema:"pennstatealert"`
		CentreCountyDocumenters bool       `schema:"centre_county_documenters"` // Alias for CentreDocumenters
		CentreDocumenters       bool       `schema:"centredocumenters"`
		HowWeCare               bool       `schema:"howwecare"` // Alias for care
		Care                    bool       `schema:"care"`
		Honeypot                bool       `schema:"contact"`
		Shibboleth              string     `schema:"shibboleth"`
		Timestamp               *time.Time `schema:"shibboleth_timestamp"`
		Turnstile               string     `schema:"cf-turnstile-response"`
	}
	if err := decoder.Decode(&req, r.PostForm); err != nil {
		app.redirectErr(w, r, err)
		return
	}
	if err := validate(req.EmailAddress, req.FirstName, req.LastName); err != nil {
		app.redirectErr(w, r, err)
		return
	}
	if req.Shibboleth != "!skcoR AP" || req.Timestamp == nil {
		err := resperr.New(http.StatusBadRequest,
			"missing shibboleth: %q", req.EmailAddress)
		err = resperr.E{E: err,
			M: "Too fast! Please leave the page open for 15 seconds before signing up."}
		app.redirectErr(w, r, err)
		return
	}
	if time.Since(*req.Timestamp).Abs() > 24*time.Hour {
		err := resperr.New(http.StatusBadRequest,
			"bad timestamp: %q", req.EmailAddress)
		err = resperr.E{E: err,
			M: "Page too old. Please reload the window and try again."}
		app.redirectErr(w, r, err)
		return
	}
	if req.Honeypot {
		err := resperr.New(http.StatusBadRequest,
			"checked honeypot %q", req.EmailAddress)
		err = resperr.E{E: err,
			M: "There was a problem with your request"}
		app.redirectErr(w, r, err)
		return
	}
	// ip, _, _ := strings.Cut(r.RemoteAddr, ":")
	// ok, err := app.tc.Validate(r.Context(), req.Turnstile, ip)
	// if err != nil {
	// 	err = resperr.E{E: err, S: http.StatusBadGateway,
	// 		M: "There was a problem connecting to the server."}
	// 	app.redirectErr(w, r, err)
	// 	return
	// }
	// if !ok {
	// 	err := resperr.New(http.StatusBadRequest,
	// 		"Turnstile rejected %q", req.EmailAddress)
	// 	err = resperr.E{E: err,
	// 		M: "There was a problem with your request."}
	// 	app.redirectErr(w, r, err)
	// 	return
	// }

	if !app.kb.Verify(r.Context(), req.EmailAddress) {
		err := resperr.New(http.StatusBadRequest,
			"Kickbox rejected %q", req.EmailAddress)
		err = resperr.E{E: err,
			M: "There was a problem with your request."}
		app.redirectErr(w, r, err)
		return
	}

	ip := r.RemoteAddr
	ok, err := app.maxcl.IPInCountry(r.Context(), ip,
		"US", "CA", "UK", "PR")
	if err != nil {
		app.redirectErr(w, r, err)
		return
	}
	if !ok {
		app.redirectErr(w, r, resperr.E{
			M: "Sorry, due to spam concerns, we are not accept international subscribers at this time."})
		return
	}
	app.l.Println("subscribing user", req.EmailAddress)

	interests := map[int]bool{
		1: true, // Master list
		3: req.PALocal,
		4: req.PAPost,
		5: req.Investigator,
		6: req.HowWeCare,
		7: req.TalkOfTheTown ||
			req.StateCollege,
		8: req.PennStateAlerts,
		9: req.BerksCounty ||
			req.Berks,
		10: req.BreakingNews,
		11: req.WeekInReview,
		13: req.Events,
	}
	maps.DeleteFunc(interests, func(k int, v bool) bool {
		return !v
	})

	if err := app.ac.CreateContact(r.Context(), activecampaign.Contact{
		Email:     emailx.Normalize(req.EmailAddress),
		FirstName: strings.TrimSpace(req.FirstName),
		LastName:  strings.TrimSpace(req.LastName),
	}); err != nil {
		app.redirectErr(w, r, err)
		return
	}

	app.l.Printf("subscribed: email=%q interests=%v", req.EmailAddress, interests)
	res, err := app.ac.FindContactByEmail(r.Context(), emailx.Normalize(req.EmailAddress))
	if err != nil {
		app.redirectErr(w, r, err)
		return
	}
	if len(res.Contacts) != 1 {
		err := resperr.New(http.StatusBadRequest,
			"Could not find user ID %q", req.EmailAddress)
		err = resperr.E{E: err,
			M: "There was a problem while processing your request. Please try again."}
		app.redirectErr(w, r, err)
		return
	}
	contactID := res.Contacts[0].ID
	app.l.Printf("found user: id=%d", contactID)

	for _, listID := range slices.Sorted(maps.Keys(interests)) {
		if err := app.ac.AddToList(r.Context(), listID, contactID); err != nil {
			app.redirectErr(w, r, err)
			return
		}
	}
	dest := validateRedirect(r.FormValue("redirect"), "/thanks.html")
	http.Redirect(w, r, dest, http.StatusSeeOther)
}

func (app *appEnv) postSubscribeJSON(w http.ResponseWriter, r *http.Request) http.Handler {
	app.Printf("start postSubscribeJSON")

	var req struct {
		EmailAddress            string `json:"EMAIL"`
		FirstName               string `json:"FNAME"`
		LastName                string `json:"LNAME"`
		Events                  bool   `json:"events"`
		Investigator            bool   `json:"investigator"`
		PAPost                  bool   `json:"papost"`
		BreakingNews            bool   `json:"breaking_news"`
		PALocal                 bool   `json:"palocal"`
		BerksCounty             bool   `json:"berks_county"`
		Berks                   bool   `json:"berks"`         // Alias for BerksCounty
		TalkOfTheTown           bool   `json:"talkofthetown"` // Alias for StateCollege
		StateCollege            bool   `json:"state_college"`
		WeekInReview            bool   `json:"week_in_review"`
		PennStateAlerts         bool   `json:"pennstatealert"`
		CentreCountyDocumenters bool   `json:"centre_county_documenters"` // Alias for CentreDocumenters
		CentreDocumenters       bool   `json:"centredocumenters"`
		HowWeCare               bool   `json:"howwecare"` // Alias for care
		Care                    bool   `json:"care"`
		Honeypot                bool   `json:"contact"`
		Token                   string `json:"token"`
	}
	if err := app.readJSON(r, &req); err != nil {
		return app.replyErr(err)
	}
	if err := validate(req.EmailAddress, req.FirstName, req.LastName); err != nil {
		return app.replyErr(err)
	}
	if !app.verifyToken(time.Now(), req.Token) {
		return app.replyErr(resperr.New(http.StatusBadRequest, "invalid token %q", req.Token))
	}
	if !app.kb.Verify(r.Context(), req.EmailAddress) {
		err := resperr.New(http.StatusBadRequest,
			"Kickbox rejected %q", req.EmailAddress)
		err = resperr.E{E: err,
			M: "There was a problem with the email address entered. Please check it and try again."}
		return app.replyErr(err)
	}

	ip := r.RemoteAddr
	ok, err := app.maxcl.IPInCountry(r.Context(), ip,
		"US", "CA", "UK", "PR")
	if err != nil {
		return app.replyErr(err)
	}
	if !ok {
		return app.replyErr(resperr.E{
			M: "Sorry, due to spam concerns, we are not accept international subscribers at this time."})
	}
	app.l.Println("subscribing user", req.EmailAddress)

	interests := map[int]bool{
		1: true, // Master list
		3: req.PALocal,
		4: req.PAPost,
		5: req.Investigator,
		6: req.HowWeCare,
		7: req.TalkOfTheTown ||
			req.StateCollege,
		8: req.PennStateAlerts,
		9: req.BerksCounty ||
			req.Berks,
		10: req.BreakingNews,
		11: req.WeekInReview,
		13: req.Events,
	}
	maps.DeleteFunc(interests, func(k int, v bool) bool {
		return !v
	})

	if err := app.ac.CreateContact(r.Context(), activecampaign.Contact{
		Email:     emailx.Normalize(req.EmailAddress),
		FirstName: strings.TrimSpace(req.FirstName),
		LastName:  strings.TrimSpace(req.LastName),
	}); err != nil {
		return app.replyErr(err)
	}

	app.l.Printf("subscribed: email=%q interests=%v", req.EmailAddress, interests)
	res, err := app.ac.FindContactByEmail(r.Context(), emailx.Normalize(req.EmailAddress))
	if err != nil {
		return app.replyErr(err)
	}
	if len(res.Contacts) != 1 {
		err := resperr.New(http.StatusBadRequest,
			"Could not find user ID %q", req.EmailAddress)
		err = resperr.E{E: err,
			M: "There was a problem while processing your request. Please try again."}
		return app.replyErr(err)
	}
	contactID := res.Contacts[0].ID
	app.l.Printf("found user: id=%d", contactID)

	for _, listID := range slices.Sorted(maps.Keys(interests)) {
		if err := app.ac.AddToList(r.Context(), listID, contactID); err != nil {
			return app.replyErr(err)
		}
	}
	return app.replyJSON("ok")
}
