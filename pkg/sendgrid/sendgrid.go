package sendgrid

import (
	"bufio"
	"fmt"
	"net/http"
	"strings"
	"time"
)

const (
	AddContactsURL = "https://api.sendgrid.com/v3/marketing/contacts"
	SendURL        = "https://api.sendgrid.com/v3/mail/send"
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
	cl.Timeout = 5 * time.Second
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
	r2.URL.Host = "api.sendgrid.com"
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
	Contents         []Content         `json:"content"`
}

type Address struct {
	Email string `json:"email"`
	Name  string `json:"name"`
}

type Personalization struct {
	To      []Address `json:"to"`
	Subject string    `json:"subject"`
}

type Content struct {
	Type  string `json:"type"`
	Value string `json:"value"`
}
