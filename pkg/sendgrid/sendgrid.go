package sendgrid

import (
	"bufio"
	"fmt"
	"net/http"
	"strings"
	"time"
)

const (
	AddContactsURL        = "https://api.sendgrid.com/v3/marketing/contacts"
	RemoveUserFromListURL = "https://api.sendgrid.com/v3/marketing/lists/%s/contacts?contact_ids=%s"
	SearchForUserURL      = "https://api.sendgrid.com/v3/marketing/contacts/search"
	SendURL               = "https://api.sendgrid.com/v3/mail/send"
	UserByIDURL           = "https://api.sendgrid.com/v3/marketing/contacts/%s"
)

type Logger interface {
	Printf(string, ...interface{})
}

func NewMockClient(l Logger) *http.Client {
	cl := &http.Client{}
	cl.Transport = mockTripper{l}
	return cl
}

type mockTripper struct {
	l Logger
}

func (mt mockTripper) RoundTrip(r *http.Request) (*http.Response, error) {
	mt.l.Printf("mocking Sendgrid call to %v", r.URL)
	buf := bufio.NewReader(strings.NewReader("HTTP/1.1 200 OK\r\n\r\n"))
	resp, _ := http.ReadResponse(buf, nil)

	return resp, nil
}

func NewClient(token string) *http.Client {
	cl := &http.Client{}
	cl.Transport = roundTripper{
		fmt.Sprintf("Bearer %s", token),
		http.DefaultTransport,
	}
	cl.Timeout = 8 * time.Second
	return cl
}

type roundTripper struct {
	auth  string
	trans http.RoundTripper
}

func (rt roundTripper) RoundTrip(r *http.Request) (*http.Response, error) {
	r2 := new(http.Request)
	*r2 = *r
	r2.Header = r.Header.Clone()
	r2.Header.Set("Authorization", rt.auth)
	// Make sure we don't send token somewhere else if it's accidentally used
	// as a regular client
	if r2.URL.Host != "api.sendgrid.com" {
		return nil, fmt.Errorf("bad URL for SendGrid client: %v", r2.URL)
	}
	return rt.trans.RoundTrip(r2)
}

type AddContactsRequest struct {
	ListIds  []string  `json:"list_ids"`
	Contacts []Contact `json:"contacts"`
}

type Contact struct {
	Email     string `json:"email"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

type SendRequest struct {
	Personalizations []Personalization `json:"personalizations"`
	From             Address           `json:"from"`
	ReplyTo          Address           `json:"reply_to"`
	TemplateID       string            `json:"template_id"`
	UnsubGroup       UnsubGroup        `json:"asm"`
}

type Address struct {
	Email string `json:"email"`
	Name  string `json:"name"`
}

type Personalization struct {
	To            []Address         `json:"to"`
	Substitutions map[string]string `json:"dynamic_template_data"`
}

type Content struct {
	Type  string `json:"type"`
	Value string `json:"value"`
}

type UnsubGroup struct {
	ID int `json:"group_id"`
}

func SGQLEscape(s string) string {
	return strings.ReplaceAll(
		strings.ReplaceAll(s, `\`, `\\`),
		`'`, `\'`)
}

func BuildSearchQuery(email string) interface{} {
	query := fmt.Sprintf("email = '%s'",
		strings.ToLower(SGQLEscape(email)))
	return struct {
		Query string `json:"query"`
	}{
		Query: query,
	}
}

type SearchQueryResults struct {
	SearchResults []UserInfo `json:"result"`
	ContactCount  int        `json:"contact_count"`
	Metadata      struct {
		Self string `json:"self"`
	} `json:"_metadata"`
}

type UserInfo struct {
	AddressLine1        string            `json:"address_line_1"`
	AddressLine2        string            `json:"address_line_2"`
	AlternateEmails     []string          `json:"alternate_emails"`
	City                string            `json:"city"`
	Country             string            `json:"country"`
	Email               string            `json:"email"`
	FirstName           string            `json:"first_name"`
	ID                  string            `json:"id"`
	LastName            string            `json:"last_name"`
	ListIDs             []string          `json:"list_ids"`
	PostalCode          string            `json:"postal_code"`
	StateProvinceRegion string            `json:"state_province_region"`
	PhoneNumber         string            `json:"phone_number"`
	Whatsapp            string            `json:"whatsapp"`
	Line                string            `json:"line"`
	Facebook            string            `json:"facebook"`
	UniqueName          string            `json:"unique_name"`
	CustomFields        map[string]string `json:"custom_fields"`
	CreatedAt           time.Time         `json:"created_at"`
	UpdatedAt           time.Time         `json:"updated_at"`
	Metadata            struct {
		Self string `json:"self"`
	} `json:"_metadata"`
}
