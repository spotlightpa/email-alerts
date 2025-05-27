package emailalerts

import (
	"net/http"
	"strings"
	"time"

	"github.com/earthboundkid/emailx/v2"
	"github.com/earthboundkid/resperr/v2"
	"github.com/spotlightpa/email-alerts/pkg/activecampaign"
	"github.com/spotlightpa/email-alerts/pkg/maxmind"
)

type ListAdd struct {
	EmailAddress string
	ContactID    activecampaign.ContactID
	ListID       activecampaign.ListID
	Status       activecampaign.Status
}

func (app *appEnv) postVerifySubscribe(w http.ResponseWriter, r *http.Request) http.Handler {
	app.Printf("start postVerifySubscribe")

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
		SignUpSource            string `json:"source"`
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
	emailAddress := emailx.Normalize(req.EmailAddress)

	if !app.kb.Verify(r.Context(), emailAddress) {
		err := resperr.New(http.StatusBadRequest,
			"Kickbox rejected %q", emailAddress)
		err = resperr.E{E: err,
			M: "There was a problem with the email address entered. Please check it and try again."}
		return app.replyErr(err)
	}

	ip := r.RemoteAddr
	val, err := app.maxcl.IPInsights(r.Context(), ip,
		"US", "CA", "UK", "PR")
	if err != nil {
		return app.replyErr(err)
	}
	if val == maxmind.ResultFailed {
		return app.replyErr(resperr.E{
			M: "Sorry, due to spam concerns, we are not accepting international subscribers at this time."})
	}
	app.l.Printf("should subscribe: email=%q", emailAddress)
	res, err := app.ac.FindContactByEmail(r.Context(), emailAddress)
	if err != nil {
		return app.replyErr(err)
	}

	shouldConfirm := false
	var contactID activecampaign.ContactID

	if len(res.Contacts) == 1 {
		contactID = res.Contacts[0].ID
		app.l.Printf("found user: email=%q id=%d", emailAddress, contactID)
	} else {
		if val == maxmind.ResultProvisional {
			shouldConfirm = true
		}
		app.l.Printf("not found; creating user: email=%q", emailAddress)
		if contactID, err = app.ac.CreateContact(r.Context(), activecampaign.Contact{
			Email:     emailAddress,
			FirstName: strings.TrimSpace(req.FirstName),
			LastName:  strings.TrimSpace(req.LastName),
			FieldValues: []activecampaign.FieldValue{
				{
					Field: activecampaign.SignUpSourceFieldID,
					Value: req.SignUpSource,
				}}}); err != nil {
			return app.replyErr(err)
		}
		app.Printf("created user email=%q id=%d", emailAddress, contactID)
	}
	if shouldConfirm {
		app.Printf("user email=%q id=%d needs to be confirmed", emailAddress, contactID)
		if err = app.ac.AddToAutomation(
			r.Context(), contactID, activecampaign.OptInTestAutomation,
		); err != nil {
			// Log and continue on error
			app.logErr(r.Context(), err)
		}
	}

	now := time.Now()
	var messages []string
	for listID, ok := range []bool{
		activecampaign.ListMaster:          true,
		activecampaign.ListPALocal:         req.PALocal == "1",
		activecampaign.ListPAPost:          req.PAPost == "1",
		activecampaign.ListInvestigator:    req.Investigator == "1",
		activecampaign.ListHowWeCare:       req.HowWeCare == "1",
		activecampaign.ListTalkOfTheTown:   req.TalkOfTheTown == "1" || req.StateCollege == "1",
		activecampaign.ListPennStateAlerts: req.PennStateAlerts == "1",
		activecampaign.ListBerksCounty:     req.BerksCounty == "1" || req.Berks == "1",
		activecampaign.ListBreakingNews:    req.BreakingNews == "1",
		activecampaign.ListWeekInReview:    req.WeekInReview == "1",
		activecampaign.ListEvents:          req.Events == "1",
	} {
		if !ok {
			continue
		}
		app.l.Printf("should to list %v", activecampaign.ListID(listID))
		sub := ListAdd{
			EmailAddress: emailAddress,
			ContactID:    contactID,
			ListID:       activecampaign.ListID(listID),
			Status:       activecampaign.StatusActive,
		}
		msg := Message{CreatedAt: now}
		if err := msg.Encode(sub); err != nil {
			return app.replyErr(err)
		}
		signed := app.signMessage(msg)
		messages = append(messages, signed)
	}
	return app.replyJSON(messages)
}

func (app *appEnv) postListAdd(w http.ResponseWriter, r *http.Request) http.Handler {
	app.Printf("start postListAdd")
	var req string
	if err := app.readJSON(r, &req); err != nil {
		return app.replyErr(err)
	}
	msg := app.unpackMessage(req)
	if msg == nil {
		return app.replyErr(resperr.E{M: "Message could not be verified."})
	}
	if !msg.ValidNow() {
		return app.replyErr(resperr.E{M: "Request expired. Try again."})
	}
	var sub ListAdd
	if err := msg.Decode(&sub); err != nil {
		return app.replyErr(resperr.E{M: "Message could not be parsed."})
	}
	app.Printf("subscribing %q (%d) to %v (%d) %v",
		sub.EmailAddress, sub.ContactID, sub.ListID, sub.ListID, sub.Status)
	if err := app.ac.AddToList(r.Context(), sub.ListID, sub.ContactID, sub.Status); err != nil {
		return app.replyErr(err)
	}
	return app.replyJSON("OK")
}
