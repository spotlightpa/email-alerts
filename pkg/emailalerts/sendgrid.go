package emailalerts

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"
)

func (app *appEnv) setSendGrid(token string) error {
	// shallow copy
	app.sg = &*http.DefaultClient
	app.sg.Transport = sendGridRT{
		fmt.Sprintf("Bearer %s", token),
	}
	app.sg.Timeout = 5 * time.Second
	return nil
}

type sendGridRT struct {
	auth string
}

func (sg sendGridRT) RoundTrip(r *http.Request) (*http.Response, error) {
	r.Header.Set("Authorization", sg.auth)
	return http.DefaultTransport.RoundTrip(r)
}

type sendGridAddContactsRequest struct {
	ListIds  []string          `json:"list_ids"`
	Contacts []sendGridContact `json:"contacts"`
}

type sendGridContact struct {
	Email     string `json:"email"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

const sendGridAddContactsURL = "https://api.sendgrid.com/v3/marketing/contacts"

func (app *appEnv) addContact(ctx context.Context, first, last, email, fips string) error {
	id := fipsToList[fips].ID
	if id == "" {
		return fmt.Errorf("invalid fips: %q", fips)
	}
	if !strings.Contains(email, "@") {
		return fmt.Errorf("invalid email: %q", email)
	}
	data := sendGridAddContactsRequest{
		ListIds: []string{id},
		Contacts: []sendGridContact{{
			FirstName: first,
			LastName:  last,
			Email:     email,
		}},
	}
	return putJSON(ctx, app.sg, sendGridAddContactsURL, data)
}
