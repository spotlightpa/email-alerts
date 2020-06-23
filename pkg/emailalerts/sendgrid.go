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
	county := fipsToList[fips].Name
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
			To: []sendgrid.Address{{
				Name:  first + " " + last,
				Email: email,
			}},
			Substitutions: map[string]string{
				"first":  first,
				"last":   last,
				"county": county,
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
