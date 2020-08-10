package emailalerts

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/carlmjohnson/emailx"
	"github.com/carlmjohnson/resperr"
	"github.com/spotlightpa/email-alerts/pkg/httpjson"
	"github.com/spotlightpa/email-alerts/pkg/sendgrid"
)

func (app *appEnv) addContact(ctx context.Context, first, last, email string, fipsCodes []string) error {
	if !emailx.Valid(email) {
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
	if err := httpjson.Put(
		ctx, app.sg, sendgrid.EndpointAddContacts,
		data,
		nil,
	); err != nil {
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
			ID: AlertsUnsubGroupID,
		},
	}
	return httpjson.Post(ctx, app.sg, sendgrid.EndpointSend, data, nil)
}

type contactData struct {
	ID           string   `json:"id"`
	Email        string   `json:"email"`
	FirstName    string   `json:"first_name"`
	LastName     string   `json:"last_name"`
	FIPSCodes    []string `json:"fips_codes"`
	Unsubscribed bool     `json:"unsubscribed"`
}

func (app *appEnv) listSubscriptions(ctx context.Context, email string) (contact *contactData, err error) {
	if !emailx.Valid(email) {
		return nil, resperr.New(
			http.StatusBadRequest, "invalid email %q", email)
	}
	var searchResp sendgrid.SearchQueryResults
	if err = httpjson.Post(
		ctx, app.sg, sendgrid.EndpointSearchForUser,
		sendgrid.BuildSearchQuery(email),
		&searchResp,
	); err != nil {
		return nil, err
	}
	if n := len(searchResp.SearchResults); n == 0 {
		return nil, resperr.New(
			http.StatusNotFound,
			"could not find user %q", email)
	} else if n != 1 {
		return nil, fmt.Errorf(
			"wrong number of users found for email %q %d != 1",
			email, n)
	}

	unsub, err := app.isUnsubscribed(ctx, email)
	if err != nil {
		return nil, err
	}

	user := searchResp.SearchResults[0]
	contact = &contactData{
		ID:           user.ID,
		FirstName:    user.FirstName,
		LastName:     user.LastName,
		Email:        user.Email,
		FIPSCodes:    listIDsToFIPS(user.ListIDs),
		Unsubscribed: unsub,
	}
	return
}

const AlertsUnsubGroupID = 13641

func (app *appEnv) isUnsubscribed(ctx context.Context, email string) (unsubscribed bool, err error) {
	var unsub sendgrid.UnsubscribeGroupsResponse
	if err = httpjson.Get(
		ctx, app.sg,
		fmt.Sprintf(sendgrid.EndpointUnsubscribeGroupsByEmail, email),
		&unsub,
	); err != nil {
		return false, err
	}
	for _, group := range unsub.Suppressions {
		if group.ID == AlertsUnsubGroupID {
			return group.IsUnsubscribed, nil
		}
	}
	return false, fmt.Errorf("unsubscribe group not found")
}

func listIDsToFIPS(listIDs []string) []string {
	fipsCodes := make([]string, 0, len(listIDs))
	for _, listID := range listIDs {
		if fipsCode := listToFIPS[listID].FIPS; fipsCode != "" {
			fipsCodes = append(fipsCodes, fipsCode)
		}
	}
	return fipsCodes
}

func fipsCodesToListIDs(fipsCodes []string) []string {
	listIDs := make([]string, 0, len(fipsCodes))
	for _, code := range fipsCodes {
		id := fipsToList[code].ID
		if id == "" {
			continue
		}
		listIDs = append(listIDs, id)
	}
	return listIDs
}

func (app *appEnv) updateSubscriptions(ctx context.Context, user contactData) error {
	if user.Unsubscribed {
		return app.unsubscribe(ctx, user.Email)
	}

	var info sendgrid.UserInfo
	if err := httpjson.Get(
		ctx, app.sg,
		fmt.Sprintf(sendgrid.EndpointUserByID, user.ID),
		&info,
	); err != nil {
		return err
	}
	oldFIPSCodes := listIDsToFIPS(info.ListIDs)
	_, codesToRemove := symDiff(user.FIPSCodes, oldFIPSCodes)
	listIDsToRemove := fipsCodesToListIDs(codesToRemove)
	for _, listID := range listIDsToRemove {
		if err := httpjson.Delete(
			ctx, app.sg,
			fmt.Sprintf(sendgrid.EndpointRemoveUserFromList, listID, user.ID),
			nil,
			http.StatusAccepted,
		); err != nil {
			return err
		}
	}
	// Post this even if there are no IDs to update the username.
	// Posting all IDs, not just new ones because there may be stale data,
	// so we can't trust it.
	data := sendgrid.AddContactsRequest{
		ListIds: fipsCodesToListIDs(user.FIPSCodes),
		Contacts: []sendgrid.Contact{{
			FirstName: user.FirstName,
			LastName:  user.LastName,
			Email:     user.Email,
		}},
	}
	if err := httpjson.Put(
		ctx, app.sg, sendgrid.EndpointAddContacts, data, nil,
	); err != nil {
		return err
	}
	if err := httpjson.Delete(
		ctx, app.sg, fmt.Sprintf(
			sendgrid.EndpointRemoveFromUnsubscribeGroup,
			AlertsUnsubGroupID,
			user.Email,
		),
		nil,
		http.StatusNoContent,
	); err != nil {
		return err
	}

	return nil
}

func (app *appEnv) unsubscribe(ctx context.Context, email string) error {
	return httpjson.Post(
		ctx, app.sg,
		fmt.Sprintf(
			sendgrid.EndpointAddToUnsubscribeGroup,
			AlertsUnsubGroupID,
		),
		sendgrid.UnsubscribeGroupAddRequest{
			EmailAddresses: []string{email},
		},
		nil,
		http.StatusCreated,
	)
}
