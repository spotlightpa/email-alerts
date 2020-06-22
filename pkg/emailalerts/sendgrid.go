package emailalerts

import (
	"context"
	"fmt"
	"strings"

	"github.com/spotlightpa/email-alerts/pkg/sendgrid"
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
	data = sendgrid.AddContactsRequest{
		ListIds: []string{id},
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
			Subject: fmt.Sprintf(
				"Welcome to the %s COVID-19 List from Spotlight PA",
				fipsToList[fips].Name,
			),
			To: []sendgrid.Address{{
				Name:  first + " " + last,
				Email: email,
			}},
		}},
		From: sendgrid.Address{
			Name:  "Spotlight PA",
			Email: "newsletters@spotlightpa.org",
		},
		ReplyTo: sendgrid.Address{
			Name:  "Spotlight PA",
			Email: "newsletters@spotlightpa.org",
		},
		Contents: []sendgrid.Content{{
			Type:  "text/plain",
			Value: "You have succesfully signed up for the list! :-)\n\n",
		}},
		UnsubGroup: sendgrid.UnsubGroup{
			ID: 13641,
		},
	}
	return postJSON(ctx, app.sg, sendgrid.SendURL, data)
}
