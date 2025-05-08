package emailalerts

import (
	"maps"
	"net/http"
	"slices"
	"strings"
	"time"

	"github.com/earthboundkid/emailx/v2"
	"github.com/earthboundkid/resperr/v2"
	"github.com/spotlightpa/email-alerts/pkg/activecampaign"
)

func (app *appEnv) postSubscribeJSON(w http.ResponseWriter, r *http.Request) http.Handler {
	app.Printf("start postSubscribeJSON")

	var req struct {
		EmailAddress            string `json:"EMAIL"`
		FirstName               string `json:"FNAME"`
		LastName                string `json:"LNAME"`
		Events                  string `json:"events"`
		Investigator            string `json:"investigator"`
		PAPost                  string `json:"papost"`
		BreakingNews            string `json:"breaking_news"`
		PALocal                 string `json:"palocal"`
		BerksCounty             string `json:"berks_county"`
		Berks                   string `json:"berks"`         // Alias for BerksCounty
		TalkOfTheTown           string `json:"talkofthetown"` // Alias for StateCollege
		StateCollege            string `json:"state_college"`
		WeekInReview            string `json:"week_in_review"`
		PennStateAlerts         string `json:"pennstatealert"`
		CentreCountyDocumenters string `json:"centre_county_documenters"` // Alias for CentreDocumenters
		CentreDocumenters       string `json:"centredocumenters"`
		HowWeCare               string `json:"howwecare"` // Alias for care
		Care                    string `json:"care"`
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

	interests := map[activecampaign.ListID]bool{
		1: true, // Master list
		3: req.PALocal == "1",
		4: req.PAPost == "1",
		5: req.Investigator == "1",
		6: req.HowWeCare == "1",
		7: req.TalkOfTheTown == "1" ||
			req.StateCollege == "1",
		8: req.PennStateAlerts == "1",
		9: req.BerksCounty == "1" ||
			req.Berks == "1",
		10: req.BreakingNews == "1",
		11: req.WeekInReview == "1",
		13: req.Events == "1",
	}
	maps.DeleteFunc(interests, func(k activecampaign.ListID, v bool) bool {
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

	status := activecampaign.StatusActive
	for _, listID := range slices.Sorted(maps.Keys(interests)) {
		if err := app.ac.AddToList(r.Context(), listID, contactID, status); err != nil {
			return app.replyErr(err)
		}
	}
	return app.replyJSON("ok")
}
