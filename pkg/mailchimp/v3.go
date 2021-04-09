package mailchimp

import (
	"context"
	"crypto/md5"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"strings"

	"github.com/carlmjohnson/emailx"
	"github.com/carlmjohnson/resperr"
	"github.com/spotlightpa/email-alerts/pkg/httpjson"
)

func FlagVar(fs *flag.FlagSet) func(c *http.Client) V3 {
	apiKey := fs.String("mc-api-key", "", "`API key` for MailChimp newsletter archive")
	listID := fs.String("mc-list-id", "", "List `ID` for MailChimp newsletter archive")

	return func(c *http.Client) V3 {
		return NewV3(*apiKey, *listID, c)
	}
}

func V3Client(apiKey string, c *http.Client) *http.Client {
	if c == nil {
		c = http.DefaultClient
	}
	newClient := new(http.Client)
	*newClient = *c
	newClient.Transport = V3Transport(apiKey, newClient.Transport)
	return newClient
}

func V3Transport(apiKey string, rt http.RoundTripper) http.RoundTripper {
	return rtFunc(func(r *http.Request) (*http.Response, error) {
		if !strings.HasSuffix(r.URL.Host, "api.mailchimp.com") {
			return nil, fmt.Errorf("bad URL for MailChimp API: %v", r.URL)
		}
		if rt == nil {
			rt = http.DefaultTransport
		}
		newReq := r.Clone(r.Context())
		newReq.SetBasicAuth("", apiKey)
		return rt.RoundTrip(newReq)
	})
}

type rtFunc func(r *http.Request) (*http.Response, error)

func (rt rtFunc) RoundTrip(r *http.Request) (*http.Response, error) {
	return rt(r)
}

type V3 struct {
	cl         *http.Client
	datacenter string
	listID     string
}

func NewV3(apiKey, listID string, c *http.Client) V3 {
	cl := V3Client(apiKey, c)
	// API keys end with 123XYZ-us1, where us1 is the datacenter
	var datacenter string
	if n := strings.LastIndex(apiKey, "-"); n != -1 {
		datacenter = apiKey[n+1:]
	}
	return V3{
		cl:         cl,
		datacenter: datacenter,
		listID:     listID,
	}
}

func (v3 V3) PutUserURL(email string) string {
	return fmt.Sprintf(
		"https://%s.api.mailchimp.com/3.0/lists/%s/members/%s",
		v3.datacenter,
		v3.listID,
		SubscriberHash(email),
	)
}

func SubscriberHash(email string) string {
	email = emailx.Normalize(email)
	return fmt.Sprintf("%x", md5.Sum([]byte(email)))
}

func (v3 V3) PutUser(ctx context.Context, req *PutUserRequest) error {
	var resp PutUserResponse
	err := httpjson.Put(
		ctx,
		v3.cl,
		v3.PutUserURL(req.EmailAddress),
		req,
		&resp,
		http.StatusOK,
	)
	if err != nil {
		var code httpjson.UnexpectedStatusError
		if errors.As(err, &code) && code == http.StatusBadRequest {
			err = resperr.WithUserMessagef(
				resperr.New(
					http.StatusBadRequest, "bad address %q", req.EmailAddress),
				"Server rejected email address %q",
				req.EmailAddress,
			)
		} else {
			err = resperr.New(http.StatusBadGateway,
				"problem connecting to MailChimp: %w", err)
		}
	}
	return err
}
