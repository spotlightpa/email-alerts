package emailalerts

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/carlmjohnson/resperr"
	"github.com/spotlightpa/email-alerts/pkg/httpjson"
	"github.com/spotlightpa/email-alerts/pkg/sendgrid"
)

func (app *appEnv) addContact(ctx context.Context, first, last, email string, fipsCodes []string) error {
	if !strings.Contains(email, "@") {
		return fmt.Errorf("invalid email: %q", email)
	}
	if len(fipsCodes) < 1 {
		return fmt.Errorf("no county mailing list selected")
	}
	ids := make([]string, 0, len(fipsCodes))
	counties := make([]string, 0, len(fipsCodes))
	for _, fips := range fipsCodes {
		id := fipsToList[fips].ID
		if id == "" {
			return fmt.Errorf("invalid fips: %q", fips)
		}
		ids = append(ids, id)
		counties = append(counties, fipsToList[fips].Name)
	}
	var data interface{}
	data = sendgrid.AddContactsRequest{
		ListIds: ids,
		Contacts: []sendgrid.Contact{{
			FirstName: first,
			LastName:  last,
			Email:     email,
		}},
	}
	if err := putJSON(ctx, app.sg, sendgrid.AddContactsURL, data); err != nil {
		return err
	}
	data = sendgrid.SendRequest{
		Personalizations: []sendgrid.Personalization{{
			To: []sendgrid.Address{{
				Name:  first + " " + last,
				Email: email,
			}},
			Substitutions: map[string]string{
				"first":  first,
				"last":   last,
				"county": strings.Join(counties, " / "),
			},
		}},
		From: sendgrid.Address{
			Name:  "Spotlight PA",
			Email: "newsletters@spotlightpa.org",
		},
		ReplyTo: sendgrid.Address{
			Name:  "Spotlight PA",
			Email: "newsletters@spotlightpa.org",
		},
		TemplateID: "d-375d4ead8f99430d9ed1a674dd40ffa0",
		UnsubGroup: sendgrid.UnsubGroup{
			ID: 13641,
		},
	}
	return postJSON(ctx, app.sg, sendgrid.SendURL, data)
}

func (app *appEnv) listSubscriptions(ctx context.Context, email string) (userID string, fipsCodes []string, err error) {

	var searchResp sendgrid.SearchQueryResults
	if err = httpjson.Post(
		ctx, app.sg, sendgrid.SearchForUserURL,
		sendgrid.BuildSearchQuery(email),
		&searchResp,
	); err != nil {
		return "", nil, err
	}
	if n := len(searchResp.SearchResults); n == 0 {
		return "", nil, resperr.New(
			http.StatusNotFound,
			"could not find user %q", email)
	} else if n != 1 {
		return "", nil, fmt.Errorf(
			"wrong number of users found for email %q %d != 1",
			email, n)
	}
	userID = searchResp.SearchResults[0].ID
	listIDs := searchResp.SearchResults[0].ListIDs
	fipsCodes = make([]string, 0, len(listIDs))
	for _, listID := range listIDs {
		if fipsCode := listToFIPS[listID].FIPS; fipsCode != "" {
			fipsCodes = append(fipsCodes, fipsCode)
		}
	}
	return
}

func (app *appEnv) updateSubscriptions(ctx context.Context, first, last, email string, fipsCodes []string) error {
	userID, currentCodes, err := app.listSubscriptions(ctx, email)
	if err != nil {
		return err
	}
	_, codesToRemove := symDiff(currentCodes, fipsCodes)
	for _, code := range codesToRemove {
		listID := fipsToList[code].ID
		if listID == "" {
			continue
		}
		if err = httpjson.Delete(
			ctx, app.sg,
			fmt.Sprintf(sendgrid.RemoveUserFromListURL, listID, userID),
			nil,
			http.StatusAccepted,
		); err != nil {
			return err
		}
	}
	listIDs := make([]string, 0, len(fipsCodes))
	for _, code := range fipsCodes {
		id := fipsToList[code].ID
		if id == "" {
			continue
		}
		listIDs = append(listIDs, id)
	}
	// Post this even if there are no IDs to update the username.
	// Posting all IDs, not just new ones because the search endpoint
	// returns stale data, so we can't trust it.
	data := sendgrid.AddContactsRequest{
		ListIds: listIDs,
		Contacts: []sendgrid.Contact{{
			FirstName: first,
			LastName:  last,
			Email:     email,
		}},
	}
	if err := httpjson.Put(
		ctx, app.sg, sendgrid.AddContactsURL, data, nil,
	); err != nil {
		return err
	}
	return nil
}
