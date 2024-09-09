package emailalerts

import (
	"maps"
	"net/http"
	"strings"
	"time"

	"github.com/carlmjohnson/emailx"
	"github.com/carlmjohnson/resperr"
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
			"Too fast! Please leave the page open for 15 seconds before signing up.")
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
		// TODO!!
		// "6137d9281f": req.BreakingNews,
		// "5c3b89e306": req.WeekInReview,
		// "650bf212f7": req.CentreCountyDocumenters ||
		// 	req.CentreDocumenters,
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

	app.l.Printf("subscribed: email=%q", req.EmailAddress)
	res, err := app.ac.FindContactByEmail(r.Context(), emailx.Normalize(req.EmailAddress))
	if err != nil {
		app.redirectErr(w, r, err)
		return
	}
	if len(res.Contacts) != 1 {
		err := resperr.New(http.StatusBadRequest,
			"Could not find user ID %q", req.EmailAddress)
		err = resperr.WithUserMessage(err,
			"There was a problem while processing your request. Please try again.")
		app.redirectErr(w, r, err)
		return
	}
	contactID := res.Contacts[0].ID
	app.l.Printf("found user: id=%d", contactID)

	for listID := range interests {
		if err := app.ac.AddToList(r.Context(), listID, contactID); err != nil {
			app.redirectErr(w, r, err)
			return
		}
	}
	dest := validateRedirect(r.FormValue("redirect"), "/thanks.html")
	http.Redirect(w, r, dest, http.StatusSeeOther)
}
