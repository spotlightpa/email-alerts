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

type sendGridSendRequest struct {
	Personalizations []personalization `json:"personalizations"`
	From             contact           `json:"from"`
	ReplyTo          contact           `json:"reply_to"`
	Contents         []content         `json:"content"`
}

type contact struct {
	Email string `json:"email"`
	Name  string `json:"name"`
}

type personalization struct {
	To      []contact `json:"to"`
	Subject string    `json:"subject"`
}

type content struct {
	Type  string `json:"type"`
	Value string `json:"value"`
}

const (
	sendGridAddContactsURL = "https://api.sendgrid.com/v3/marketing/contacts"
	sendGridSendURL        = "https://api.sendgrid.com/v3/mail/send"
)

func (app *appEnv) addContact(ctx context.Context, first, last, email, fips string) error {
	id := fipsToList[fips].ID
	if id == "" {
		return fmt.Errorf("invalid fips: %q", fips)
	}
	if !strings.Contains(email, "@") {
		return fmt.Errorf("invalid email: %q", email)
	}
	var data interface{}
	data = sendGridAddContactsRequest{
		ListIds: []string{id},
		Contacts: []sendGridContact{{
			FirstName: first,
			LastName:  last,
			Email:     email,
		}},
	}
	if err := putJSON(ctx, app.sg, sendGridAddContactsURL, data); err != nil {
		return err
	}
	data = sendGridSendRequest{
		Personalizations: []personalization{{
			Subject: fmt.Sprintf(
				"Welcome to the %s COVID-19 List from Spotlight PA",
				fipsToList[fips].Name,
			),
			To: []contact{{
				Name:  first + " " + last,
				Email: email,
			}},
		}},
		From: contact{
			Name:  "Spotlight PA",
			Email: "newsletters@spotlightpa.org",
		},
		ReplyTo: contact{
			Name:  "Spotlight PA",
			Email: "newsletters@spotlightpa.org",
		},
		Contents: []content{{
			Type:  "text/plain",
			Value: "You have succesfully signed up for the list! :-)",
		}},
	}
	return postJSON(ctx, app.sg, sendGridSendURL, data)
}
